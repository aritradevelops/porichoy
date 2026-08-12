package rest

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aritradevelops/porichoy/server/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const testHost = "acme.example.com"

// testApp wires a real Fiber app (rest.New) over mocked repositories, so every request
// exercises the actual middleware chain and handler code — only the DB is faked, matching
// how internal/tenant/service_test.go isolates Service from Postgres.
type testApp struct {
	app      *fiber.App
	tenants  *mockTenantRepo
	domains  *mockDomainRepo
	creds    *mockCredentialRepo
	tenantID uuid.UUID
}

// newTestApp builds a testApp with testHost already resolving to a fixture tenant
// (tenantID) — the one thing every authenticated-route test needs, since TenantResolution
// runs on every request in the "authed" group.
func newTestApp(t *testing.T) *testApp {
	t.Helper()
	tenants := &mockTenantRepo{}
	domains := &mockDomainRepo{}
	creds := &mockCredentialRepo{}

	tenantID := uuid.New()
	domains.On("FindByDomain", mock.Anything, testHost).
		Return(&tenant.TenantDomain{TenantID: tenantID}, nil)
	tenants.On("FindByID", mock.Anything, tenantID).
		Return(&tenant.Tenant{ID: tenantID, Name: "Acme"}, nil)

	svc := tenant.NewService(tenants, domains, creds)
	return &testApp{app: New(svc), tenants: tenants, domains: domains, creds: creds, tenantID: tenantID}
}

// do sends a request against ta.app and returns the decoded envelope alongside the raw
// status code, so tests can assert on both without repeating the boilerplate.
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

func TestTenants_Create_OK(t *testing.T) {
	ta := newTestApp(t)
	ta.tenants.On("Create", mock.Anything, mock.AnythingOfType("*tenant.Tenant")).Return(nil)

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
	ta.tenants.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
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

func TestDomains_Register_OK(t *testing.T) {
	ta := newTestApp(t)
	ta.domains.On("FindByDomain", mock.Anything, "new.example.com").Return(nil, nil)
	ta.domains.On("Create", mock.Anything, mock.AnythingOfType("*tenant.TenantDomain")).Return(nil)

	status, env := ta.do(t, http.MethodPost, "/api/v1/domains/register",
		registerDomainRequest{TenantID: ta.tenantID, Domain: "new.example.com"}, nil)

	require.Equal(t, http.StatusCreated, status)
	require.Nil(t, env.Error)
}

func TestDomains_Register_ForbiddenForOtherTenant(t *testing.T) {
	ta := newTestApp(t)
	otherTenant := uuid.New()

	status, env := ta.do(t, http.MethodPost, "/api/v1/domains/register",
		registerDomainRequest{TenantID: otherTenant, Domain: "new.example.com"}, nil)

	require.Equal(t, http.StatusForbidden, status)
	require.Equal(t, "domain.forbidden", env.Error.Key)
	ta.domains.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestDomains_Register_RootCanRegisterForAnyTenant(t *testing.T) {
	ta := newTestApp(t)
	otherTenant := uuid.New()
	ta.domains.On("FindByDomain", mock.Anything, "other.example.com").Return(nil, nil)
	ta.domains.On("Create", mock.Anything, mock.AnythingOfType("*tenant.TenantDomain")).Return(nil)

	status, env := ta.do(t, http.MethodPost, "/api/v1/domains/register",
		registerDomainRequest{TenantID: otherTenant, Domain: "other.example.com"},
		map[string]string{"X-Debug-Scope": "root"})

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
