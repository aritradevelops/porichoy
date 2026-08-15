package rest

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/aritradevelops/porichoy/server/internal/app"
	"github.com/aritradevelops/porichoy/server/internal/authorization"
	"github.com/aritradevelops/porichoy/server/internal/identity"
	"github.com/aritradevelops/porichoy/server/internal/tenant"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// signupTestApp mirrors newTestApp but lets each test control the resolved tenant's
// EnabledLoginMethods — signup's behavior branches on that field, which newTestApp's shared
// fixture doesn't set (the tenant/domain handler tests never need to). roles/permCache back
// authorization.Service, wired up with a permissive default (no role assignments -> empty
// permissions cached) so tests that don't care about permission-caching specifically don't
// need to restate it — override via ta.assignments/ta.roles/ta.permCache when a test does.
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
	roles := &mockRoleRepo{}
	permCache := &mockPermissionCache{}

	tenantID := uuid.New()
	domains.On("FindByDomain", mock.Anything, testHost).
		Return(&tenant.TenantDomain{TenantID: tenantID}, nil)
	tenants.On("FindByID", mock.Anything, tenantID).
		Return(&tenant.Tenant{ID: tenantID, Name: "Acme", EnabledLoginMethods: loginMethods}, nil)
	assignments.On("ListByPrincipal", mock.Anything, mock.Anything).Return(nil, nil).Maybe()
	permCache.On("SetUserPermissions", mock.Anything, tenantID, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()

	tenantSvc := tenant.NewService(tenants, domains, creds)
	identitySvc := identity.NewService(users, passwords, identityApps, sessions, assignments, tokens, noopTxRunner{})
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
		tokens:       tokens,
		roles:        roles,
		permCache:    permCache,
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

func TestAuth_Login_OK(t *testing.T) {
	ta := signupTestApp(t, tenant.LoginMethodEmailPassword)
	sysApp := testSystemAppFixture(ta.tenantID)
	hash, err := identity.HashPassword("hunter222")
	require.NoError(t, err)
	email := "existing@example.com"
	u := &identity.User{ID: uuid.New(), TenantID: ta.tenantID, Status: identity.UserStatusActive, Email: &email}

	ta.identityApps.On("FindSystemAppByTenant", mock.Anything, ta.tenantID).Return(sysApp, nil)
	ta.users.On("FindByEmail", mock.Anything, ta.tenantID, email).Return(u, nil)
	ta.passwords.On("FindByUserID", mock.Anything, u.ID).
		Return(&identity.Password{ID: uuid.New(), UserID: u.ID, PasswordHash: hash}, nil)
	ta.sessions.On("Create", mock.Anything, mock.AnythingOfType("*app.Session")).Return(nil)
	ta.tokens.On("Issue", sysApp, mock.Anything, mock.Anything).Return("signed-token", nil)

	status, env := ta.do(t, http.MethodPost, "/api/v1/auth/login",
		loginRequest{Email: email, Password: "hunter222"}, nil)

	require.Equal(t, http.StatusOK, status)
	require.Nil(t, env.Error)
	data := env.Data.(map[string]any)
	require.Equal(t, email, data["email"])
	require.Equal(t, "signed-token", data["access_token"])
}

func TestAuth_Login_CachesEffectivePermissions(t *testing.T) {
	ta := signupTestApp(t, tenant.LoginMethodEmailPassword)
	sysApp := testSystemAppFixture(ta.tenantID)
	sysApp.AccessTokenTTLSeconds = 900
	hash, err := identity.HashPassword("hunter222")
	require.NoError(t, err)
	email := "existing@example.com"
	u := &identity.User{ID: uuid.New(), TenantID: ta.tenantID, Status: identity.UserStatusActive, Email: &email}
	roleID := uuid.New()

	ta.identityApps.On("FindSystemAppByTenant", mock.Anything, ta.tenantID).Return(sysApp, nil)
	ta.users.On("FindByEmail", mock.Anything, ta.tenantID, email).Return(u, nil)
	ta.passwords.On("FindByUserID", mock.Anything, u.ID).
		Return(&identity.Password{ID: uuid.New(), UserID: u.ID, PasswordHash: hash}, nil)
	ta.sessions.On("Create", mock.Anything, mock.AnythingOfType("*app.Session")).Return(nil)
	ta.tokens.On("Issue", sysApp, mock.Anything, mock.Anything).Return("signed-token", nil)
	// Replaces signupTestApp's default "no assignments"/"accept any cache write" mocks
	// (registered first, so a looser mock.Anything expectation would otherwise win over a
	// more specific one added afterward — testify matches in registration order, not by
	// specificity) with this test's own, so CacheUserPermissions resolves and stores a real,
	// non-empty permission set.
	ta.assignments.ExpectedCalls = nil
	ta.permCache.ExpectedCalls = nil
	ta.assignments.On("ListByPrincipal", mock.Anything, u.ID).Return([]*authorization.RoleAssignment{
		{ID: uuid.New(), PrincipalID: u.ID, RoleID: roleID},
	}, nil)
	ta.roles.On("FindByIDs", mock.Anything, []uuid.UUID{roleID}).Return([]*authorization.Role{
		{ID: roleID, Permissions: []string{"tenants:*@tenant"}},
	}, nil)
	ta.permCache.On("SetUserPermissions", mock.Anything, ta.tenantID, u.ID, []string{"tenants:*@tenant"}, 900*time.Second).
		Return(nil)

	status, env := ta.do(t, http.MethodPost, "/api/v1/auth/login",
		loginRequest{Email: email, Password: "hunter222"}, nil)

	require.Equal(t, http.StatusOK, status)
	require.Nil(t, env.Error)
	ta.permCache.AssertExpectations(t)
}

func TestAuth_Login_SucceedsEvenWhenCacheWriteFails(t *testing.T) {
	ta := signupTestApp(t, tenant.LoginMethodEmailPassword)
	sysApp := testSystemAppFixture(ta.tenantID)
	hash, err := identity.HashPassword("hunter222")
	require.NoError(t, err)
	email := "existing@example.com"
	u := &identity.User{ID: uuid.New(), TenantID: ta.tenantID, Status: identity.UserStatusActive, Email: &email}

	ta.identityApps.On("FindSystemAppByTenant", mock.Anything, ta.tenantID).Return(sysApp, nil)
	ta.users.On("FindByEmail", mock.Anything, ta.tenantID, email).Return(u, nil)
	ta.passwords.On("FindByUserID", mock.Anything, u.ID).
		Return(&identity.Password{ID: uuid.New(), UserID: u.ID, PasswordHash: hash}, nil)
	ta.sessions.On("Create", mock.Anything, mock.AnythingOfType("*app.Session")).Return(nil)
	ta.tokens.On("Issue", sysApp, mock.Anything, mock.Anything).Return("signed-token", nil)
	// Overrides signupTestApp's default cache-write mock to fail — Login must still succeed
	// (the "best-effort" decision: nothing reads this cache yet, so a Redis outage shouldn't
	// take login down with it).
	ta.permCache.ExpectedCalls = nil
	ta.permCache.On("SetUserPermissions", mock.Anything, ta.tenantID, u.ID, mock.Anything, mock.Anything).
		Return(errors.New("redis: connection refused"))

	status, env := ta.do(t, http.MethodPost, "/api/v1/auth/login",
		loginRequest{Email: email, Password: "hunter222"}, nil)

	require.Equal(t, http.StatusOK, status)
	require.Nil(t, env.Error)
}

func TestAuth_Login_WrongPassword(t *testing.T) {
	ta := signupTestApp(t, tenant.LoginMethodEmailPassword)
	sysApp := testSystemAppFixture(ta.tenantID)
	hash, err := identity.HashPassword("hunter222")
	require.NoError(t, err)
	email := "existing@example.com"
	u := &identity.User{ID: uuid.New(), TenantID: ta.tenantID, Status: identity.UserStatusActive, Email: &email}

	ta.identityApps.On("FindSystemAppByTenant", mock.Anything, ta.tenantID).Return(sysApp, nil)
	ta.users.On("FindByEmail", mock.Anything, ta.tenantID, email).Return(u, nil)
	ta.passwords.On("FindByUserID", mock.Anything, u.ID).
		Return(&identity.Password{ID: uuid.New(), UserID: u.ID, PasswordHash: hash}, nil)

	status, env := ta.do(t, http.MethodPost, "/api/v1/auth/login",
		loginRequest{Email: email, Password: "wrong-password"}, nil)

	require.Equal(t, http.StatusUnauthorized, status)
	require.Equal(t, "identity.invalid_credentials", env.Error.Key)
}

func TestAuth_Login_UnknownEmail(t *testing.T) {
	ta := signupTestApp(t, tenant.LoginMethodEmailPassword)
	sysApp := testSystemAppFixture(ta.tenantID)
	ta.identityApps.On("FindSystemAppByTenant", mock.Anything, ta.tenantID).Return(sysApp, nil)
	ta.users.On("FindByEmail", mock.Anything, ta.tenantID, "nobody@example.com").Return(nil, nil)

	status, env := ta.do(t, http.MethodPost, "/api/v1/auth/login",
		loginRequest{Email: "nobody@example.com", Password: "hunter222"}, nil)

	require.Equal(t, http.StatusUnauthorized, status)
	require.Equal(t, "identity.invalid_credentials", env.Error.Key)
}

func TestAuth_Login_ValidationError(t *testing.T) {
	ta := signupTestApp(t, tenant.LoginMethodEmailPassword)

	status, env := ta.do(t, http.MethodPost, "/api/v1/auth/login",
		loginRequest{Email: "not-an-email", Password: ""}, nil)

	require.Equal(t, http.StatusBadRequest, status)
	require.Equal(t, "validation.failed", env.Error.Key)
	ta.identityApps.AssertNotCalled(t, "FindSystemAppByTenant", mock.Anything, mock.Anything)
}

func TestAuth_Login_LoginMethodDisabled(t *testing.T) {
	ta := signupTestApp(t, tenant.LoginMethodGoogle)

	status, env := ta.do(t, http.MethodPost, "/api/v1/auth/login",
		loginRequest{Email: "a@example.com", Password: "hunter222"}, nil)

	require.Equal(t, http.StatusBadRequest, status)
	require.Equal(t, "identity.login_method_disabled", env.Error.Key)
}
