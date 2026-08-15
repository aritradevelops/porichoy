package rest

import (
	"net/http"
	"testing"

	"github.com/aritradevelops/porichoy/server/internal/app"
	"github.com/aritradevelops/porichoy/server/internal/identity"
	"github.com/aritradevelops/porichoy/server/internal/tenant"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// signupTestApp mirrors newTestApp but lets each test control the resolved tenant's
// EnabledLoginMethods — signup's behavior branches on that field, which newTestApp's shared
// fixture doesn't set (the tenant/domain handler tests never need to).
func signupTestApp(t *testing.T, loginMethods ...tenant.LoginMethod) *testApp {
	t.Helper()
	tenants := &mockTenantRepo{}
	domains := &mockDomainRepo{}
	creds := &mockCredentialRepo{}
	users := &mockUserRepo{}
	passwords := &mockPasswordRepo{}
	identityApps := &mockIdentityAppRepo{}
	sessions := &mockSessionRepo{}
	assignments := &mockRoleAssignmentRepo{}
	tokens := &mockTokenIssuer{}

	tenantID := uuid.New()
	domains.On("FindByDomain", mock.Anything, testHost).
		Return(&tenant.TenantDomain{TenantID: tenantID}, nil)
	tenants.On("FindByID", mock.Anything, tenantID).
		Return(&tenant.Tenant{ID: tenantID, Name: "Acme", EnabledLoginMethods: loginMethods}, nil)

	tenantSvc := tenant.NewService(tenants, domains, creds)
	identitySvc := identity.NewService(users, passwords, identityApps, sessions, assignments, tokens, noopTxRunner{})
	return &testApp{
		app:          New(tenantSvc, identitySvc),
		tenants:      tenants,
		domains:      domains,
		creds:        creds,
		users:        users,
		passwords:    passwords,
		identityApps: identityApps,
		sessions:     sessions,
		assignments:  assignments,
		tokens:       tokens,
		tenantID:     tenantID,
	}
}

func testSystemAppFixture(tenantID uuid.UUID) *app.App {
	return &app.App{
		ID:                     uuid.New(),
		TenantID:               tenantID,
		IsSystem:               true,
		SigningAlgorithm:       app.SigningAlgorithmHS256,
		AccessTokenTTLSeconds:  900,
		IDTokenTTLSeconds:      900,
		RefreshTokenTTLSeconds: 2592000,
	}
}

func TestAuth_Signup_OK(t *testing.T) {
	ta := signupTestApp(t, tenant.LoginMethodEmailPassword)
	sysApp := testSystemAppFixture(ta.tenantID)

	ta.identityApps.On("FindSystemAppByTenant", mock.Anything, ta.tenantID).Return(sysApp, nil)
	ta.users.On("FindByEmail", mock.Anything, ta.tenantID, "new@example.com").Return(nil, nil)
	ta.users.On("Create", mock.Anything, mock.AnythingOfType("*identity.User")).Return(nil)
	ta.passwords.On("Create", mock.Anything, mock.AnythingOfType("*identity.Password")).Return(nil)
	ta.sessions.On("Create", mock.Anything, mock.AnythingOfType("*app.Session")).Return(nil)
	ta.tokens.On("Issue", sysApp, mock.Anything, mock.Anything).Return("signed-token", nil)

	status, env := ta.do(t, http.MethodPost, "/api/v1/auth/signup",
		signupRequest{Email: "new@example.com", Password: "hunter222"}, nil)

	require.Equal(t, http.StatusCreated, status)
	require.Nil(t, env.Error)
	data := env.Data.(map[string]any)
	require.Equal(t, "new@example.com", data["email"])
	require.Equal(t, "signed-token", data["access_token"])
	require.Equal(t, "Bearer", data["token_type"])
	ta.assignments.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestAuth_Signup_ValidationError(t *testing.T) {
	ta := signupTestApp(t, tenant.LoginMethodEmailPassword)

	status, env := ta.do(t, http.MethodPost, "/api/v1/auth/signup",
		signupRequest{Email: "not-an-email", Password: "short"}, nil)

	require.Equal(t, http.StatusBadRequest, status)
	require.Equal(t, "validation.failed", env.Error.Key)
	ta.identityApps.AssertNotCalled(t, "FindSystemAppByTenant", mock.Anything, mock.Anything)
}

func TestAuth_Signup_LoginMethodDisabled(t *testing.T) {
	ta := signupTestApp(t, tenant.LoginMethodGoogle)

	status, env := ta.do(t, http.MethodPost, "/api/v1/auth/signup",
		signupRequest{Email: "new@example.com", Password: "hunter222"}, nil)

	require.Equal(t, http.StatusBadRequest, status)
	require.Equal(t, "identity.login_method_disabled", env.Error.Key)
}

func TestAuth_Signup_SystemAppNotFound(t *testing.T) {
	ta := signupTestApp(t, tenant.LoginMethodEmailPassword)
	ta.identityApps.On("FindSystemAppByTenant", mock.Anything, ta.tenantID).Return(nil, nil)

	status, env := ta.do(t, http.MethodPost, "/api/v1/auth/signup",
		signupRequest{Email: "new@example.com", Password: "hunter222"}, nil)

	require.Equal(t, http.StatusInternalServerError, status)
	require.Equal(t, "identity.system_app_not_found", env.Error.Key)
}

func TestAuth_Signup_EmailAlreadyRegistered(t *testing.T) {
	ta := signupTestApp(t, tenant.LoginMethodEmailPassword)
	sysApp := testSystemAppFixture(ta.tenantID)
	ta.identityApps.On("FindSystemAppByTenant", mock.Anything, ta.tenantID).Return(sysApp, nil)
	existingEmail := "existing@example.com"
	ta.users.On("FindByEmail", mock.Anything, ta.tenantID, existingEmail).
		Return(&identity.User{ID: uuid.New(), Email: &existingEmail}, nil)

	status, env := ta.do(t, http.MethodPost, "/api/v1/auth/signup",
		signupRequest{Email: existingEmail, Password: "hunter222"}, nil)

	require.Equal(t, http.StatusConflict, status)
	require.Equal(t, "identity.email_already_registered", env.Error.Key)
}

func TestAuth_Signup_AssignsDefaultSignupRole(t *testing.T) {
	ta := signupTestApp(t, tenant.LoginMethodEmailPassword)
	sysApp := testSystemAppFixture(ta.tenantID)
	roleID := uuid.New()
	sysApp.DefaultSignupRoleID = &roleID

	ta.identityApps.On("FindSystemAppByTenant", mock.Anything, ta.tenantID).Return(sysApp, nil)
	ta.users.On("FindByEmail", mock.Anything, ta.tenantID, "new@example.com").Return(nil, nil)
	ta.users.On("Create", mock.Anything, mock.AnythingOfType("*identity.User")).Return(nil)
	ta.passwords.On("Create", mock.Anything, mock.AnythingOfType("*identity.Password")).Return(nil)
	ta.assignments.On("Create", mock.Anything, mock.AnythingOfType("*authorization.RoleAssignment")).Return(nil)
	ta.sessions.On("Create", mock.Anything, mock.AnythingOfType("*app.Session")).Return(nil)
	ta.tokens.On("Issue", sysApp, mock.Anything, mock.Anything).Return("signed-token", nil)

	status, env := ta.do(t, http.MethodPost, "/api/v1/auth/signup",
		signupRequest{Email: "new@example.com", Password: "hunter222"}, nil)

	require.Equal(t, http.StatusCreated, status)
	require.Nil(t, env.Error)
	ta.assignments.AssertExpectations(t)
}
