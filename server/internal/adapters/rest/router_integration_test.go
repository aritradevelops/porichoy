//go:build integration

package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aritradevelops/porichoy/server/internal/actor"
	"github.com/aritradevelops/porichoy/server/internal/adapters/crypto"
	"github.com/aritradevelops/porichoy/server/internal/adapters/postgres"
	"github.com/aritradevelops/porichoy/server/internal/app"
	"github.com/aritradevelops/porichoy/server/internal/identity"
	"github.com/aritradevelops/porichoy/server/internal/tenant"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// newIdentityService wires a real identity.Service from Postgres repositories — none of the
// tests in this file exercise signup itself (that's auth_handlers_test.go's job), but rest.New
// always requires one, same as tenantSvc.
func newIdentityService() *identity.Service {
	return identity.NewService(
		postgres.NewUserRepository(testDB),
		postgres.NewPasswordRepository(testDB),
		postgres.NewAppRepository(testDB),
		postgres.NewSessionRepository(testDB),
		postgres.NewRoleAssignmentRepository(testDB),
		crypto.NewIssuer(),
		postgres.NewTxRunner(testDB),
	)
}

// mustIssueTestToken creates tenantID's system app (every tenant this file seeds needs
// exactly one, per the idx_apps_tenant_system unique index — callers that need the app
// itself, e.g. to enable a login method against it, should create their own tenant instead
// of going through this) and mints a valid access token for a fresh random principal against
// it — every authenticated-route test needs one, now that Authentication verifies real
// signatures (internal/identity.Service.Authenticate) instead of trusting a debug header.
func mustIssueTestToken(t *testing.T, tenantID uuid.UUID) string {
	t.Helper()
	ctx := context.Background()
	sysApp, err := app.NewService(postgres.NewAppRepository(testDB)).CreateSystemApp(ctx, tenantID, "System")
	require.NoError(t, err)
	token, err := crypto.NewIssuer().Issue(sysApp, app.Claims{Subject: uuid.NewString(), Audience: tenantID.String()}, time.Hour)
	require.NoError(t, err)
	return token
}

// authHeader is doRequest's headers argument for a bearer-authenticated request.
func authHeader(token string) map[string]string {
	return map[string]string{"Authorization": "Bearer " + token}
}

// seedTenant inserts a tenant, one domain for it, and its system app (mustIssueTestToken)
// directly through the real Postgres repositories — bootstrapping a resolvable, authable
// Host is a precondition every authenticated route needs (TenantResolution+Authentication
// both run on every request in the "authed" group), and creating the very first tenant a
// domain can resolve to isn't itself a REST concern (root-tenant bootstrap is CLI-only,
// UI_PAGES.md/CODING_STANDARDS' own framing) — so tests seed it the same way
// internal/adapters/postgres's own integration tests build fixtures, then drive everything
// else through real HTTP requests against the real app.
func seedTenant(t *testing.T, domain string) (svc *tenant.Service, tenantID uuid.UUID, token string) {
	t.Helper()
	ctx := context.Background()
	tenants := postgres.NewTenantRepository(testDB)
	domains := postgres.NewDomainRepository(testDB)
	creds := postgres.NewProviderCredentialRepository(testDB)

	now := time.Now().UTC().Truncate(time.Microsecond)
	tt := &tenant.Tenant{
		ID:          uuid.New(),
		Name:        "Acme",
		LoginLayout: tenant.LoginLayoutCentered,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	require.NoError(t, tenants.CreateRoot(ctx, tt))

	d := &tenant.TenantDomain{ID: uuid.New(), TenantID: tt.ID, Domain: domain, CreatedAt: now}
	require.NoError(t, domains.Create(ctx, actor.Actor{Scope: actor.ScopeRoot}, d))

	return tenant.NewService(tenants, domains, creds), tt.ID, mustIssueTestToken(t, tt.ID)
}

// seedTenantTree seeds a 3-level chain root -> child -> grandchild, each with its own
// resolvable domain (via the same repos, so descendant-scope authorization — computed from
// the precomputed ancestors array, AUTHORIZATION_MODEL.md §4 — is exercised end-to-end
// through real HTTP requests instead of directly against the repository). Domain names are
// randomized per call (not just "tree-root.example.com" etc.) since multiple tests in this
// package share one Postgres container/test-run (CODING_STANDARDS.md §6) and domains are
// unique instance-wide — a fixed name would collide across tests. Returns rootDomain so
// callers can resolve against the root without hardcoding it themselves.
func seedTenantTree(t *testing.T) (svc *tenant.Service, rootID, childID, grandchildID uuid.UUID, rootDomain, rootToken string) {
	t.Helper()
	ctx := context.Background()
	tenants := postgres.NewTenantRepository(testDB)
	domains := postgres.NewDomainRepository(testDB)
	creds := postgres.NewProviderCredentialRepository(testDB)
	now := time.Now().UTC().Truncate(time.Microsecond)
	root := actor.Actor{Scope: actor.ScopeRoot}
	suffix := uuid.NewString()

	rootTenant := &tenant.Tenant{ID: uuid.New(), Name: "Tree Root", LoginLayout: tenant.LoginLayoutCentered, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, tenants.CreateRoot(ctx, rootTenant))
	rootDomain = suffix + "-root.example.com"
	require.NoError(t, domains.Create(ctx, root, &tenant.TenantDomain{ID: uuid.New(), TenantID: rootTenant.ID, Domain: rootDomain, CreatedAt: now}))

	childTenant := &tenant.Tenant{ID: uuid.New(), ParentID: &rootTenant.ID, Name: "Tree Child", LoginLayout: tenant.LoginLayoutCentered, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, tenants.Create(ctx, root, childTenant))
	require.NoError(t, domains.Create(ctx, root, &tenant.TenantDomain{ID: uuid.New(), TenantID: childTenant.ID, Domain: suffix + "-child.example.com", CreatedAt: now}))

	grandchildTenant := &tenant.Tenant{ID: uuid.New(), ParentID: &childTenant.ID, Name: "Tree Grandchild", LoginLayout: tenant.LoginLayoutCentered, CreatedAt: now, UpdatedAt: now}
	require.NoError(t, tenants.Create(ctx, root, grandchildTenant))
	require.NoError(t, domains.Create(ctx, root, &tenant.TenantDomain{ID: uuid.New(), TenantID: grandchildTenant.ID, Domain: suffix + "-grandchild.example.com", CreatedAt: now}))

	return tenant.NewService(tenants, domains, creds), rootTenant.ID, childTenant.ID, grandchildTenant.ID, rootDomain, mustIssueTestToken(t, rootTenant.ID)
}

func doRequest(t *testing.T, app interface {
	Test(*http.Request, ...int) (*http.Response, error)
}, method, host, target string, body any, headers map[string]string) (int, envelope) {
	t.Helper()
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		require.NoError(t, err)
		reader = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, "http://"+host+target, reader)
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	var env envelope
	require.NoError(t, json.Unmarshal(raw, &env))
	return resp.StatusCode, env
}

func TestIntegration_TenantLifecycle(t *testing.T) {
	svc, tenantID, token := seedTenant(t, "lifecycle.example.com")
	app := New(svc, newIdentityService())

	// Create a child tenant under the resolved one.
	status, env := doRequest(t, app, http.MethodPost, "lifecycle.example.com", "/api/v1/tenants/create",
		createTenantRequest{Name: "Child Brand"}, authHeader(token))
	require.Equal(t, http.StatusCreated, status)
	require.Nil(t, env.Error)
	created := env.Data.(map[string]any)
	require.Equal(t, "Child Brand", created["name"])

	// Get the resolved tenant itself back.
	status, env = doRequest(t, app, http.MethodGet, "lifecycle.example.com", "/api/v1/tenants/get/"+tenantID.String(), nil, authHeader(token))
	require.Equal(t, http.StatusOK, status)
	require.Nil(t, env.Error)
	got := env.Data.(map[string]any)
	require.Equal(t, "Acme", got["name"])

	// Configure it.
	status, env = doRequest(t, app, http.MethodPost, "lifecycle.example.com", "/api/v1/tenants/configure/"+tenantID.String(),
		configureTenantRequest{MFARequired: boolPtr(true)}, authHeader(token))
	require.Equal(t, http.StatusOK, status)
	require.Nil(t, env.Error)
	configured := env.Data.(map[string]any)
	require.Equal(t, true, configured["mfa_required"])

	// Register a new domain for it.
	status, env = doRequest(t, app, http.MethodPost, "lifecycle.example.com", "/api/v1/domains/register",
		registerDomainRequest{TenantID: tenantID, Domain: "lifecycle-2.example.com"}, authHeader(token))
	require.Equal(t, http.StatusCreated, status)
	require.Nil(t, env.Error)

	// The newly registered domain resolves to the same tenant via the public endpoint (no
	// auth needed — it runs before any principal/actor exists).
	status, env = doRequest(t, app, http.MethodGet, "lifecycle.example.com", "/api/v1/domains/resolve?domain=lifecycle-2.example.com", nil, nil)
	require.Equal(t, http.StatusOK, status)
	resolved := env.Data.(map[string]any)
	require.Equal(t, tenantID.String(), resolved["tenant_id"])
}

func TestIntegration_DescendantScope_ReachesDescendantButNotSibling(t *testing.T) {
	svc, _, _, grandchildID, rootDomain, rootToken := seedTenantTree(t)
	app := New(svc, newIdentityService())

	// tree-root's resolved actor (tenant scope) can reach a grandchild it didn't directly
	// create — AUTHORIZATION_MODEL.md §4's descendant-access rule, driven by the
	// precomputed ancestors array rather than an exact match.
	status, env := doRequest(t, app, http.MethodGet, rootDomain, "/api/v1/tenants/get/"+grandchildID.String(), nil, authHeader(rootToken))
	require.Equal(t, http.StatusOK, status)
	require.Nil(t, env.Error)

	// A tenant outside tree-root's subtree entirely stays unreachable — siblings still
	// can't see each other. Still authenticated as tree-root's own principal/token —
	// authorization (descendant-scope), not authentication, is what denies this.
	_, siblingID, _ := seedTenant(t, "unrelated-sibling-"+uuid.NewString()+".example.com")
	status, env = doRequest(t, app, http.MethodGet, rootDomain, "/api/v1/tenants/get/"+siblingID.String(), nil, authHeader(rootToken))
	require.Equal(t, http.StatusNotFound, status)
	require.Equal(t, "tenant.not_found", env.Error.Key)
}

func TestIntegration_DomainsRegister_DescendantScope(t *testing.T) {
	svc, _, _, grandchildID, rootDomain, rootToken := seedTenantTree(t)
	app := New(svc, newIdentityService())

	// tree-root's resolved actor (tenant scope) can register a domain for the grandchild
	// tenant it didn't directly create — same descendant-access rule as the tenants module,
	// now enforced inside DomainRepository.Create (AUTHORIZATION_MODEL.md §4).
	status, env := doRequest(t, app, http.MethodPost, rootDomain, "/api/v1/domains/register",
		registerDomainRequest{TenantID: grandchildID, Domain: "new-for-grandchild-" + uuid.NewString() + ".example.com"}, authHeader(rootToken))
	require.Equal(t, http.StatusCreated, status)
	require.Nil(t, env.Error)

	// A tenant outside tree-root's subtree stays unreachable — 404, not 403, consistent
	// with the "not found, not forbidden" framing used throughout this bounded context.
	_, siblingID, _ := seedTenant(t, "domains-sibling-"+uuid.NewString()+".example.com")
	status, env = doRequest(t, app, http.MethodPost, rootDomain, "/api/v1/domains/register",
		registerDomainRequest{TenantID: siblingID, Domain: "should-fail-" + uuid.NewString() + ".example.com"}, authHeader(rootToken))
	require.Equal(t, http.StatusNotFound, status)
	require.Equal(t, "tenant.not_found", env.Error.Key)
}

func TestIntegration_SignupThenLogin(t *testing.T) {
	domain := "login-flow-" + uuid.NewString() + ".example.com"
	// seedTenant already creates the tenant's one-and-only system app
	// (idx_apps_tenant_system) — only EnabledLoginMethods needs configuring here.
	svc, tenantID, _ := seedTenant(t, domain)
	root := actor.Actor{Scope: actor.ScopeRoot}
	_, err := svc.ConfigureTenant(context.Background(), root, tenantID, tenant.Config{
		EnabledLoginMethods: []tenant.LoginMethod{tenant.LoginMethodEmailPassword},
	})
	require.NoError(t, err)

	fiberApp := New(svc, newIdentityService())
	email := "login-e2e-" + uuid.NewString() + "@example.com"

	// Sign up, then log in with the same credentials — both public routes, no token needed.
	status, env := doRequest(t, fiberApp, http.MethodPost, domain, "/api/v1/auth/signup",
		signupRequest{Email: email, Password: "hunter222"}, nil)
	require.Equal(t, http.StatusCreated, status)
	require.Nil(t, env.Error)

	status, env = doRequest(t, fiberApp, http.MethodPost, domain, "/api/v1/auth/login",
		loginRequest{Email: email, Password: "hunter222"}, nil)
	require.Equal(t, http.StatusOK, status)
	require.Nil(t, env.Error)
	data := env.Data.(map[string]any)
	require.Equal(t, email, data["email"])
	require.NotEmpty(t, data["access_token"])
	loginToken := data["access_token"].(string)

	// The wrong password is rejected, without revealing that the account exists.
	status, env = doRequest(t, fiberApp, http.MethodPost, domain, "/api/v1/auth/login",
		loginRequest{Email: email, Password: "wrong-password"}, nil)
	require.Equal(t, http.StatusUnauthorized, status)
	require.Equal(t, "identity.invalid_credentials", env.Error.Key)

	// The freshly logged-in user's own access token authenticates a real, authed route —
	// proof the whole Login -> Authentication loop actually closes end-to-end.
	status, env = doRequest(t, fiberApp, http.MethodGet, domain, "/api/v1/tenants/get/"+tenantID.String(), nil, authHeader(loginToken))
	require.Equal(t, http.StatusOK, status)
	require.Nil(t, env.Error)
}

func boolPtr(b bool) *bool { return &b }
