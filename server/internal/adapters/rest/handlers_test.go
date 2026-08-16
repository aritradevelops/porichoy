package rest

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aritradevelops/porichoy/server/internal/adapters/crypto"
	"github.com/aritradevelops/porichoy/server/internal/app"
	"github.com/aritradevelops/porichoy/server/internal/authorization"
	"github.com/aritradevelops/porichoy/server/internal/identity"
	"github.com/aritradevelops/porichoy/server/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const testHost = "acme.example.com"

// testTenantScopePermissions is what a tenant admin would have cached at Login — granted by
// default to newTestApp's fixture principal so every existing authed-route test keeps
// exercising the same tenant-scope behavior it did before Authenticate started doing a real
// cache-backed permission check (previously the X-Debug-Scope stub always defaulted to
// actor.ScopeTenant). Mirrors cmd/seed/main.go's tenantAdminPermissions.
var testTenantScopePermissions = []string{"tenants:*@tenant", "domains:*@tenant", "provider_credentials:*@tenant"}

// testApp wires a real Fiber app (rest.New) over mocked repositories, so every request
// exercises the actual middleware chain and handler code — only the DB is faked, matching
// how internal/tenant/service_test.go isolates Service from Postgres. tokens is a real
// crypto.Issuer, not a mock: Authenticate now verifies a real signature
// (internal/identity.Service.Authenticate), and every authed-route test needs that to
// succeed without restating JWT-mocking boilerplate — do() attaches the pre-minted valid
// token automatically. permCache is likewise wired with a real, tenant-scope permission set
// for principalID, since Authenticate now also does a real cache-backed permission check
// (authzSvc.ResolveScope) rather than trusting a debug header.
type testApp struct {
	app          *fiber.App
	tenants      *mockTenantRepo
	domains      *mockDomainRepo
	creds        *mockCredentialRepo
	users        *mockUserRepo
	passwords    *mockPasswordRepo
	identityApps *mockIdentityAppRepo
	sessions     *mockSessionRepo
	assignments  *mockRoleAssignmentRepo
	// tokens/roles are only populated by signupTestApp (auth_handlers_test.go) — Signup/
	// Login tests want to mock+assert on Issue/FindByIDs directly. newTestApp uses a real
	// crypto.Issuer instead (see token/principalID below) since its authed-route tests need
	// Authenticate's Verify call to actually succeed, which a bare mock can't do without
	// per-test setup; its roles is an unused stand-in, needed only because New requires an
	// *authorization.Service to exist at all.
	tokens      *mockTokenIssuer
	roles       *mockRoleRepo
	permCache   *mockPermissionCache
	appID       uuid.UUID
	tenantID    uuid.UUID
	principalID uuid.UUID
	token       string
}

// newTestApp builds a testApp with testHost already resolving to a fixture tenant
// (tenantID) with its own system app, a valid access token for principalID against that
// tenant, and principalID's cached permissions (testTenantScopePermissions) — the things
// every authenticated-route test needs, since TenantResolution+Authenticate both run on
// every request in the "authed" group.
func newTestApp(t *testing.T) *testApp {
	t.Helper()
	tenants := &mockTenantRepo{}
	domains := &mockDomainRepo{}
	creds := &mockCredentialRepo{}
	users := &mockUserRepo{}
	passwords := &mockPasswordRepo{}
	identityApps := &mockIdentityAppRepo{}
	sessions := &mockSessionRepo{}
	assignments := &mockRoleAssignmentRepo{}
	issuer := crypto.NewIssuer()

	tenantID := uuid.New()
	domains.On("FindByDomain", mock.Anything, testHost).
		Return(&tenant.TenantDomain{TenantID: tenantID}, nil)
	tenants.On("FindByID", mock.Anything, tenantID).
		Return(&tenant.Tenant{ID: tenantID, Name: "Acme"}, nil)

	sysApp := &app.App{
		ID:               uuid.New(),
		TenantID:         tenantID,
		IsSystem:         true,
		SigningAlgorithm: app.SigningAlgorithmHS256,
		SigningKeyConfig: []byte("test-fixture-signing-key-32-bytes!!"),
	}
	identityApps.On("FindSystemAppByTenant", mock.Anything, tenantID).Return(sysApp, nil)

	principalID := uuid.New()
	token, err := issuer.Issue(sysApp, app.Claims{Subject: principalID.String(), Audience: tenantID.String()}, time.Hour)
	require.NoError(t, err)

	roles := &mockRoleRepo{}
	permCache := &mockPermissionCache{}
	permissionsJSON, err := json.Marshal(testTenantScopePermissions)
	require.NoError(t, err)
	permCache.On("GetUserPermissions", mock.Anything, sysApp.ID, principalID).Return(permissionsJSON, nil)

	tenantSvc := tenant.NewService(tenants, domains, creds)
	identitySvc := identity.NewService(users, passwords, identityApps, sessions, assignments, issuer, noopTxRunner{})
	authzSvc := authorization.NewService(roles, assignments, permCache)
	return &testApp{
		app:          New(tenantSvc, identitySvc, authzSvc),
		tenants:      tenants,
		domains:      domains,
		creds:        creds,
		users:        users,
		passwords:    passwords,
		identityApps: identityApps,
		sessions:     sessions,
		assignments:  assignments,
		roles:        roles,
		permCache:    permCache,
		appID:        sysApp.ID,
		tenantID:     tenantID,
		principalID:  principalID,
		token:        token,
	}
}

// do sends a request against ta.app and returns the decoded envelope alongside the raw
// status code, so tests can assert on both without repeating the boilerplate. Every request
// carries ta.token as a bearer credential by default — pass an explicit "Authorization"
// entry in headers (even "") to exercise a different/missing-auth case.
func (ta *testApp) do(t *testing.T, method, target string, body any, headers map[string]string) (int, envelope) {
	t.Helper()
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		require.NoError(t, err)
		reader = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, "http://"+testHost+target, reader)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ta.token)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := ta.app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	var env envelope
	require.NoError(t, json.Unmarshal(raw, &env))
	return resp.StatusCode, env
}

func TestTenantResolution_UnregisteredHost(t *testing.T) {
	ta := newTestApp(t)
	req := httptest.NewRequest(http.MethodGet, "http://unregistered.example.com/api/v1/tenants/get/"+uuid.NewString(), nil)
	ta.domains.On("FindByDomain", mock.Anything, "unregistered.example.com").Return(nil, nil)

	resp, err := ta.app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// A request's Host header carries a port whenever the client connects to anything other than
// the default 80/443 (any local dev setup, or a non-standard-port deployment) — this must
// still resolve against the plain-hostname domain registered in domain_registry.
func TestTenantResolution_HostHeaderWithPort(t *testing.T) {
	ta := newTestApp(t)
	ta.tenants.On("GetByID", mock.Anything, mock.Anything, ta.tenantID).
		Return(&tenant.Tenant{ID: ta.tenantID, Name: "Acme"}, nil)
	req := httptest.NewRequest(http.MethodGet, "http://"+testHost+":5173/api/v1/tenants/get/"+ta.tenantID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+ta.token)

	resp, err := ta.app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAuthentication_NoToken(t *testing.T) {
	ta := newTestApp(t)

	status, env := ta.do(t, http.MethodGet, "/api/v1/tenants/get/"+ta.tenantID.String(), nil,
		map[string]string{"Authorization": ""})

	require.Equal(t, http.StatusUnauthorized, status)
	require.Equal(t, "identity.unauthenticated", env.Error.Key)
}

func TestAuthentication_MalformedToken(t *testing.T) {
	ta := newTestApp(t)

	status, env := ta.do(t, http.MethodGet, "/api/v1/tenants/get/"+ta.tenantID.String(), nil,
		map[string]string{"Authorization": "Bearer not-a-real-token"})

	require.Equal(t, http.StatusUnauthorized, status)
	require.Equal(t, "identity.unauthenticated", env.Error.Key)
}

func TestAuthentication_TokenSignedByAnotherApp(t *testing.T) {
	ta := newTestApp(t)
	wrongApp := &app.App{SigningAlgorithm: app.SigningAlgorithmHS256, SigningKeyConfig: []byte("a-completely-different-signing-key")}
	forged, err := crypto.NewIssuer().Issue(wrongApp, app.Claims{Subject: uuid.NewString(), Audience: ta.tenantID.String()}, time.Hour)
	require.NoError(t, err)

	status, env := ta.do(t, http.MethodGet, "/api/v1/tenants/get/"+ta.tenantID.String(), nil,
		map[string]string{"Authorization": "Bearer " + forged})

	require.Equal(t, http.StatusUnauthorized, status)
	require.Equal(t, "identity.unauthenticated", env.Error.Key)
}

func TestAuthentication_TokenForAnotherTenant(t *testing.T) {
	ta := newTestApp(t)
	sysApp := &app.App{SigningAlgorithm: app.SigningAlgorithmHS256, SigningKeyConfig: []byte("test-fixture-signing-key-32-bytes!!")}
	tokenForOtherTenant, err := crypto.NewIssuer().Issue(sysApp, app.Claims{Subject: uuid.NewString(), Audience: uuid.NewString()}, time.Hour)
	require.NoError(t, err)

	status, env := ta.do(t, http.MethodGet, "/api/v1/tenants/get/"+ta.tenantID.String(), nil,
		map[string]string{"Authorization": "Bearer " + tokenForOtherTenant})

	require.Equal(t, http.StatusUnauthorized, status)
	require.Equal(t, "identity.unauthenticated", env.Error.Key)
}

func TestAuthentication_TokenFromCookie(t *testing.T) {
	ta := newTestApp(t)
	ta.tenants.On("GetByID", mock.Anything, mock.Anything, ta.tenantID).
		Return(&tenant.Tenant{ID: ta.tenantID, Name: "Acme"}, nil)

	req := httptest.NewRequest(http.MethodGet, "http://"+testHost+"/api/v1/tenants/get/"+ta.tenantID.String(), nil)
	req.Header.Set("Cookie", "access_token="+ta.token)
	resp, err := ta.app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAuthorization_NoCachedPermissions(t *testing.T) {
	ta := newTestApp(t)
	// Overrides newTestApp's default granted-permissions mock with a cache miss — the
	// principal authenticated fine, but nothing was ever cached for them (never logged in,
	// or the entry expired).
	ta.permCache.ExpectedCalls = nil
	ta.permCache.On("GetUserPermissions", mock.Anything, ta.appID, ta.principalID).Return(nil, nil)

	status, env := ta.do(t, http.MethodGet, "/api/v1/tenants/get/"+ta.tenantID.String(), nil, nil)

	require.Equal(t, http.StatusForbidden, status)
	require.Equal(t, "authorization.forbidden", env.Error.Key)
}

func TestAuthorization_NoMatchingPermission(t *testing.T) {
	ta := newTestApp(t)
	// Cached permissions exist, but none of them are for the "tenants" module this route
	// exercises — proves the module+action match, not just "has some cache entry".
	ta.permCache.ExpectedCalls = nil
	permissionsJSON, err := json.Marshal([]string{"domains:*@root"})
	require.NoError(t, err)
	ta.permCache.On("GetUserPermissions", mock.Anything, ta.appID, ta.principalID).Return(permissionsJSON, nil)

	status, env := ta.do(t, http.MethodGet, "/api/v1/tenants/get/"+ta.tenantID.String(), nil, nil)

	require.Equal(t, http.StatusForbidden, status)
	require.Equal(t, "authorization.forbidden", env.Error.Key)
}

func TestTenants_Create_OK(t *testing.T) {
	ta := newTestApp(t)
	ta.tenants.On("Create", mock.Anything, mock.Anything, mock.AnythingOfType("*tenant.Tenant")).Return(nil)

	status, env := ta.do(t, http.MethodPost, "/api/v1/tenants/create", createTenantRequest{Name: "Brand A"}, nil)

	require.Equal(t, http.StatusCreated, status)
	require.Nil(t, env.Error)
	require.NotNil(t, env.Data)
}

func TestTenants_Create_ValidationError(t *testing.T) {
	ta := newTestApp(t)

	status, env := ta.do(t, http.MethodPost, "/api/v1/tenants/create", createTenantRequest{}, nil)

	require.Equal(t, http.StatusBadRequest, status)
	require.NotNil(t, env.Error)
	require.Equal(t, "validation.failed", env.Error.Key)
	ta.tenants.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
}

func TestTenants_Get_OK(t *testing.T) {
	ta := newTestApp(t)
	ta.tenants.On("GetByID", mock.Anything, mock.Anything, ta.tenantID).
		Return(&tenant.Tenant{ID: ta.tenantID, Name: "Acme"}, nil)

	status, env := ta.do(t, http.MethodGet, "/api/v1/tenants/get/"+ta.tenantID.String(), nil, nil)

	require.Equal(t, http.StatusOK, status)
	require.Nil(t, env.Error)
	data := env.Data.(map[string]any)
	require.Equal(t, "Acme", data["name"])
}

func TestTenants_Get_NotFound(t *testing.T) {
	ta := newTestApp(t)
	other := uuid.New()
	ta.tenants.On("GetByID", mock.Anything, mock.Anything, other).Return(nil, nil)

	status, env := ta.do(t, http.MethodGet, "/api/v1/tenants/get/"+other.String(), nil, nil)

	require.Equal(t, http.StatusNotFound, status)
	require.NotNil(t, env.Error)
	require.Equal(t, "tenant.not_found", env.Error.Key)
}

func TestTenants_Configure_OK(t *testing.T) {
	ta := newTestApp(t)
	existing := &tenant.Tenant{ID: ta.tenantID, Name: "Acme", LoginLayout: tenant.LoginLayoutCentered}
	ta.tenants.On("GetByID", mock.Anything, mock.Anything, ta.tenantID).Return(existing, nil)
	ta.tenants.On("Update", mock.Anything, mock.AnythingOfType("*tenant.Tenant")).Return(nil)

	newLogo := "https://new-logo"
	status, env := ta.do(t, http.MethodPost, "/api/v1/tenants/configure/"+ta.tenantID.String(),
		configureTenantRequest{LogoURL: &newLogo}, nil)

	require.Equal(t, http.StatusOK, status)
	require.Nil(t, env.Error)
	data := env.Data.(map[string]any)
	require.Equal(t, "https://new-logo", data["logo_url"])
}

func TestTenants_Configure_InvalidLoginLayout(t *testing.T) {
	ta := newTestApp(t)
	bad := "diagonal"

	status, env := ta.do(t, http.MethodPost, "/api/v1/tenants/configure/"+ta.tenantID.String(),
		configureTenantRequest{LoginLayout: &bad}, nil)

	require.Equal(t, http.StatusBadRequest, status)
	require.Equal(t, "validation.failed", env.Error.Key)
	ta.tenants.AssertNotCalled(t, "GetByID", mock.Anything, mock.Anything, mock.Anything)
}

func TestTenants_List_OK(t *testing.T) {
	ta := newTestApp(t)
	other := uuid.New()
	ta.tenants.On("List", mock.Anything, mock.Anything, tenant.ListParams{Page: 1, PageSize: 20}).
		Return(tenant.ListResult{
			Items:    []*tenant.Tenant{{ID: ta.tenantID, Name: "Acme"}, {ID: other, Name: "Other"}},
			Total:    2,
			Page:     1,
			PageSize: 20,
		}, nil)

	status, env := ta.do(t, http.MethodGet, "/api/v1/tenants/list", nil, nil)

	require.Equal(t, http.StatusOK, status)
	require.Nil(t, env.Error)
	data := env.Data.(map[string]any)
	require.Len(t, data["items"], 2)
	require.Equal(t, float64(2), data["total"])
	require.Equal(t, float64(1), data["page"])
	require.Equal(t, float64(20), data["page_size"])
}

func TestTenants_List_PassesPageParamsThrough(t *testing.T) {
	ta := newTestApp(t)
	ta.tenants.On("List", mock.Anything, mock.Anything, tenant.ListParams{Page: 2, PageSize: 5}).
		Return(tenant.ListResult{Page: 2, PageSize: 5}, nil)

	status, env := ta.do(t, http.MethodGet, "/api/v1/tenants/list?page=2&page_size=5", nil, nil)

	require.Equal(t, http.StatusOK, status)
	require.Nil(t, env.Error)
}

func TestDomains_Register_OK(t *testing.T) {
	ta := newTestApp(t)
	ta.domains.On("FindByDomain", mock.Anything, "new.example.com").Return(nil, nil)
	ta.domains.On("Create", mock.Anything, mock.Anything, mock.AnythingOfType("*tenant.TenantDomain")).Return(nil)

	status, env := ta.do(t, http.MethodPost, "/api/v1/domains/register",
		registerDomainRequest{TenantID: ta.tenantID, Domain: "new.example.com"}, nil)

	require.Equal(t, http.StatusCreated, status)
	require.Nil(t, env.Error)
}

func TestDomains_Resolve_OK(t *testing.T) {
	ta := newTestApp(t)
	// A distinct tenant from ta.tenantID (the testHost fixture) — reusing ta.tenantID
	// would collide with newTestApp's own FindByID("Acme") expectation, since testify/mock
	// matches .On() calls in registration order, not last-registered-wins.
	brandID := uuid.New()
	ta.domains.On("FindByDomain", mock.Anything, "brand-a.example.com").
		Return(&tenant.TenantDomain{TenantID: brandID}, nil)
	ta.tenants.On("FindByID", mock.Anything, brandID).
		Return(&tenant.Tenant{ID: brandID, Name: "Brand A"}, nil)

	req := httptest.NewRequest(http.MethodGet, "http://"+testHost+"/api/v1/domains/resolve?domain=brand-a.example.com", nil)
	resp, err := ta.app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	raw, _ := io.ReadAll(resp.Body)
	var env envelope
	require.NoError(t, json.Unmarshal(raw, &env))
	data := env.Data.(map[string]any)
	require.Equal(t, "Brand A", data["name"])
}

func TestDomains_Resolve_NotFound(t *testing.T) {
	ta := newTestApp(t)
	ta.domains.On("FindByDomain", mock.Anything, "unknown.example.com").Return(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "http://"+testHost+"/api/v1/domains/resolve?domain=unknown.example.com", nil)
	resp, err := ta.app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}
