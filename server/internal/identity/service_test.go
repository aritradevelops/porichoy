package identity

import (
	"context"
	"strings"
	"testing"

	"github.com/aritradevelops/porichoy/server/internal/app"
	"github.com/aritradevelops/porichoy/server/internal/apperror"
	"github.com/aritradevelops/porichoy/server/internal/tenant"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func assertAppErrorKey(t *testing.T, err error, key string) {
	t.Helper()
	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, key, appErr.Key)
}

func newTestService() (*Service, *mockUserRepo, *mockPasswordRepo, *mockAppRepo, *mockSessionRepo, *mockRoleAssignmentRepo, *mockTokenIssuer) {
	users := &mockUserRepo{}
	passwords := &mockPasswordRepo{}
	apps := &mockAppRepo{}
	sessions := &mockSessionRepo{}
	assignments := &mockRoleAssignmentRepo{}
	tokens := &mockTokenIssuer{}
	svc := NewService(users, passwords, apps, sessions, assignments, tokens, mockTxRunner{})
	return svc, users, passwords, apps, sessions, assignments, tokens
}

func testTenant(methods ...tenant.LoginMethod) *tenant.Tenant {
	return &tenant.Tenant{ID: uuid.New(), EnabledLoginMethods: methods}
}

func testSystemApp(tenantID uuid.UUID) *app.App {
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

func TestService_Signup_OK(t *testing.T) {
	svc, users, passwords, apps, sessions, assignments, tokens := newTestService()
	tt := testTenant(tenant.LoginMethodEmailPassword)
	sysApp := testSystemApp(tt.ID)
	roleID := uuid.New()
	sysApp.DefaultSignupRoleID = &roleID

	apps.On("FindSystemAppByTenant", mock.Anything, tt.ID).Return(sysApp, nil)
	users.On("FindByEmail", mock.Anything, tt.ID, "a@example.com").Return(nil, nil)
	users.On("Create", mock.Anything, mock.AnythingOfType("*identity.User")).Return(nil)
	passwords.On("Create", mock.Anything, mock.AnythingOfType("*identity.Password")).Return(nil)
	assignments.On("Create", mock.Anything, mock.AnythingOfType("*authorization.RoleAssignment")).Return(nil)
	sessions.On("Create", mock.Anything, mock.AnythingOfType("*app.Session")).Return(nil)
	tokens.On("Issue", sysApp, mock.Anything, mock.Anything).Return("signed-token", nil)

	result, err := svc.Signup(context.Background(), tt, "a@example.com", "hunter22")

	require.NoError(t, err)
	require.Equal(t, "a@example.com", *result.User.Email)
	require.False(t, result.User.EmailVerified)
	require.Equal(t, UserStatusActive, result.User.Status)
	require.Equal(t, "signed-token", result.AccessToken)
	require.Equal(t, "signed-token", result.IDToken)
	require.NotEmpty(t, result.RefreshToken)
	users.AssertExpectations(t)
	passwords.AssertExpectations(t)
	assignments.AssertExpectations(t)
	sessions.AssertExpectations(t)
}

func TestService_Signup_NormalizesEmail(t *testing.T) {
	svc, users, passwords, apps, sessions, _, tokens := newTestService()
	tt := testTenant(tenant.LoginMethodEmailPassword)
	sysApp := testSystemApp(tt.ID)

	apps.On("FindSystemAppByTenant", mock.Anything, tt.ID).Return(sysApp, nil)
	// The lookup and the persisted value must both use the normalized (lowercased, trimmed)
	// form — otherwise "User@Example.com " and "user@example.com" would be treated as
	// distinct accounts by the duplicate-email check and the partial unique index alike.
	users.On("FindByEmail", mock.Anything, tt.ID, "user@example.com").Return(nil, nil)
	users.On("Create", mock.Anything, mock.MatchedBy(func(u *User) bool {
		return *u.Email == "user@example.com"
	})).Return(nil)
	passwords.On("Create", mock.Anything, mock.AnythingOfType("*identity.Password")).Return(nil)
	sessions.On("Create", mock.Anything, mock.AnythingOfType("*app.Session")).Return(nil)
	tokens.On("Issue", sysApp, mock.Anything, mock.Anything).Return("signed-token", nil)

	result, err := svc.Signup(context.Background(), tt, " User@Example.com ", "hunter22")

	require.NoError(t, err)
	require.Equal(t, "user@example.com", *result.User.Email)
	users.AssertExpectations(t)
}

func TestService_Signup_PasswordTooLong(t *testing.T) {
	svc, users, _, apps, _, _, _ := newTestService()
	tt := testTenant(tenant.LoginMethodEmailPassword)
	sysApp := testSystemApp(tt.ID)
	apps.On("FindSystemAppByTenant", mock.Anything, tt.ID).Return(sysApp, nil)
	users.On("FindByEmail", mock.Anything, tt.ID, "a@example.com").Return(nil, nil)

	_, err := svc.Signup(context.Background(), tt, "a@example.com", strings.Repeat("a", MaxPasswordBytes+1))

	assertAppErrorKey(t, err, "identity.password_too_long")
}

func TestService_Signup_NoDefaultRole_SkipsAssignment(t *testing.T) {
	svc, users, passwords, apps, sessions, assignments, tokens := newTestService()
	tt := testTenant(tenant.LoginMethodEmailPassword)
	sysApp := testSystemApp(tt.ID) // DefaultSignupRoleID left nil

	apps.On("FindSystemAppByTenant", mock.Anything, tt.ID).Return(sysApp, nil)
	users.On("FindByEmail", mock.Anything, tt.ID, "a@example.com").Return(nil, nil)
	users.On("Create", mock.Anything, mock.AnythingOfType("*identity.User")).Return(nil)
	passwords.On("Create", mock.Anything, mock.AnythingOfType("*identity.Password")).Return(nil)
	sessions.On("Create", mock.Anything, mock.AnythingOfType("*app.Session")).Return(nil)
	tokens.On("Issue", sysApp, mock.Anything, mock.Anything).Return("signed-token", nil)

	_, err := svc.Signup(context.Background(), tt, "a@example.com", "hunter22")

	require.NoError(t, err)
	assignments.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestService_Signup_LoginMethodDisabled(t *testing.T) {
	svc, users, _, apps, _, _, _ := newTestService()
	tt := testTenant(tenant.LoginMethodGoogle) // no email_password

	_, err := svc.Signup(context.Background(), tt, "a@example.com", "hunter22")

	assertAppErrorKey(t, err, "identity.login_method_disabled")
	apps.AssertNotCalled(t, "FindSystemAppByTenant", mock.Anything, mock.Anything)
	users.AssertNotCalled(t, "FindByEmail", mock.Anything, mock.Anything, mock.Anything)
}

func TestService_Signup_SystemAppNotFound(t *testing.T) {
	svc, _, _, apps, _, _, _ := newTestService()
	tt := testTenant(tenant.LoginMethodEmailPassword)
	apps.On("FindSystemAppByTenant", mock.Anything, tt.ID).Return(nil, nil)

	_, err := svc.Signup(context.Background(), tt, "a@example.com", "hunter22")

	assertAppErrorKey(t, err, "identity.system_app_not_found")
}

func TestService_Signup_EmailAlreadyRegistered(t *testing.T) {
	svc, users, _, apps, _, _, _ := newTestService()
	tt := testTenant(tenant.LoginMethodEmailPassword)
	sysApp := testSystemApp(tt.ID)
	apps.On("FindSystemAppByTenant", mock.Anything, tt.ID).Return(sysApp, nil)
	users.On("FindByEmail", mock.Anything, tt.ID, "a@example.com").
		Return(&User{ID: uuid.New(), Email: strPtr("a@example.com")}, nil)

	_, err := svc.Signup(context.Background(), tt, "a@example.com", "hunter22")

	assertAppErrorKey(t, err, "identity.email_already_registered")
}

func TestService_CreateRootUser(t *testing.T) {
	svc, users, passwords, _, _, _, _ := newTestService()
	tenantID := uuid.New()
	users.On("Create", mock.Anything, mock.AnythingOfType("*identity.User")).Return(nil)
	passwords.On("Create", mock.Anything, mock.AnythingOfType("*identity.Password")).Return(nil)

	u, err := svc.CreateRootUser(context.Background(), tenantID, "root@example.com", "hunter22")

	require.NoError(t, err)
	require.Equal(t, tenantID, u.TenantID)
	require.True(t, u.EmailVerified)
	require.Equal(t, UserStatusActive, u.Status)
}

func strPtr(s string) *string { return &s }
