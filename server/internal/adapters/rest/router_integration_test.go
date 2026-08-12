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

	"github.com/aritradevelops/porichoy/server/internal/adapters/postgres"
	"github.com/aritradevelops/porichoy/server/internal/tenant"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// seedTenant inserts a tenant and one domain for it directly through the real Postgres
// repositories — bootstrapping a resolvable Host is a precondition every authenticated
// route needs (TenantResolution runs on every request in the "authed" group), and creating
// the very first tenant a domain can resolve to isn't itself a REST concern (root-tenant
// bootstrap is CLI-only, UI_PAGES.md/CODING_STANDARDS' own framing) — so tests seed it the
// same way internal/adapters/postgres's own integration tests build fixtures, then drive
// everything else through real HTTP requests against the real app.
func seedTenant(t *testing.T, domain string) (svc *tenant.Service, tenantID uuid.UUID) {
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
	require.NoError(t, tenants.Create(ctx, tt))

	d := &tenant.TenantDomain{ID: uuid.New(), TenantID: tt.ID, Domain: domain, CreatedAt: now}
	require.NoError(t, domains.Create(ctx, d))

	return tenant.NewService(tenants, domains, creds), tt.ID
}

func doRequest(t *testing.T, app interface {
	Test(*http.Request, ...int) (*http.Response, error)
}, method, host, target string, body any) (int, envelope) {
	t.Helper()
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		require.NoError(t, err)
		reader = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, "http://"+host+target, reader)
	req.Header.Set("Content-Type", "application/json")

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
	svc, tenantID := seedTenant(t, "lifecycle.example.com")
	app := New(svc)

	// Create a child tenant under the resolved one.
	status, env := doRequest(t, app, http.MethodPost, "lifecycle.example.com", "/api/v1/tenants/create",
		createTenantRequest{Name: "Child Brand"})
	require.Equal(t, http.StatusCreated, status)
	require.Nil(t, env.Error)
	created := env.Data.(map[string]any)
	require.Equal(t, "Child Brand", created["name"])

	// Get the resolved tenant itself back.
	status, env = doRequest(t, app, http.MethodGet, "lifecycle.example.com", "/api/v1/tenants/get/"+tenantID.String(), nil)
	require.Equal(t, http.StatusOK, status)
	require.Nil(t, env.Error)
	got := env.Data.(map[string]any)
	require.Equal(t, "Acme", got["name"])

	// Configure it.
	status, env = doRequest(t, app, http.MethodPost, "lifecycle.example.com", "/api/v1/tenants/configure/"+tenantID.String(),
		configureTenantRequest{MFARequired: boolPtr(true)})
	require.Equal(t, http.StatusOK, status)
	require.Nil(t, env.Error)
	configured := env.Data.(map[string]any)
	require.Equal(t, true, configured["mfa_required"])

	// Register a new domain for it.
	status, env = doRequest(t, app, http.MethodPost, "lifecycle.example.com", "/api/v1/domains/register",
		registerDomainRequest{TenantID: tenantID, Domain: "lifecycle-2.example.com"})
	require.Equal(t, http.StatusCreated, status)
	require.Nil(t, env.Error)

	// The newly registered domain resolves to the same tenant via the public endpoint.
	status, env = doRequest(t, app, http.MethodGet, "lifecycle.example.com", "/api/v1/domains/resolve?domain=lifecycle-2.example.com", nil)
	require.Equal(t, http.StatusOK, status)
	resolved := env.Data.(map[string]any)
	require.Equal(t, tenantID.String(), resolved["tenant_id"])
}

func TestIntegration_ExactMatchScope_CannotReadOtherTenant(t *testing.T) {
	svc, _ := seedTenant(t, "scope-a.example.com")
	_, otherTenantID := seedTenant(t, "scope-b.example.com")
	app := New(svc)

	// scope-a's resolved actor tries to fetch scope-b's tenant by ID — must come back
	// as a 404, not the other tenant's data (AUTHORIZATION_MODEL.md §4's exact-match rule).
	status, env := doRequest(t, app, http.MethodGet, "scope-a.example.com", "/api/v1/tenants/get/"+otherTenantID.String(), nil)
	require.Equal(t, http.StatusNotFound, status)
	require.Equal(t, "tenant.not_found", env.Error.Key)
}

func boolPtr(b bool) *bool { return &b }
