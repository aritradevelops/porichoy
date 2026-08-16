package rest

import (
	"context"
	"time"

	"github.com/aritradevelops/porichoy/server/internal/app"
	"github.com/aritradevelops/porichoy/server/internal/authorization"
	"github.com/aritradevelops/porichoy/server/internal/identity"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// Local mocks of identity.Repository/PasswordRepository and its sibling collaborator ports
// (app.Repository/SessionRepository/TokenIssuer, authorization.RoleRepository/
// RoleAssignmentRepository/PermissionCache) — same reasoning as mocks_test.go's tenant
// mocks: the internal/identity and internal/authorization packages' own mocks are unexported
// and not importable from here.

type mockUserRepo struct{ mock.Mock }

func (m *mockUserRepo) Create(ctx context.Context, u *identity.User) error {
	return m.Called(ctx, u).Error(0)
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*identity.User, error) {
	args := m.Called(ctx, tenantID, email)
	u, _ := args.Get(0).(*identity.User)
	return u, args.Error(1)
}

func (m *mockUserRepo) FindByID(ctx context.Context, tenantID, id uuid.UUID) (*identity.User, error) {
	args := m.Called(ctx, tenantID, id)
	u, _ := args.Get(0).(*identity.User)
	return u, args.Error(1)
}

type mockPasswordRepo struct{ mock.Mock }

func (m *mockPasswordRepo) Create(ctx context.Context, p *identity.Password) error {
	return m.Called(ctx, p).Error(0)
}

func (m *mockPasswordRepo) FindByUserID(ctx context.Context, userID uuid.UUID) (*identity.Password, error) {
	args := m.Called(ctx, userID)
	p, _ := args.Get(0).(*identity.Password)
	return p, args.Error(1)
}

type mockIdentityAppRepo struct{ mock.Mock }

func (m *mockIdentityAppRepo) CreateSystem(ctx context.Context, a *app.App) error {
	return m.Called(ctx, a).Error(0)
}

func (m *mockIdentityAppRepo) FindSystemAppByTenant(ctx context.Context, tenantID uuid.UUID) (*app.App, error) {
	args := m.Called(ctx, tenantID)
	a, _ := args.Get(0).(*app.App)
	return a, args.Error(1)
}

func (m *mockIdentityAppRepo) SetDefaultSignupRole(ctx context.Context, appID, roleID uuid.UUID) error {
	return m.Called(ctx, appID, roleID).Error(0)
}

type mockSessionRepo struct{ mock.Mock }

func (m *mockSessionRepo) Create(ctx context.Context, s *app.Session) error {
	return m.Called(ctx, s).Error(0)
}

type mockRoleAssignmentRepo struct{ mock.Mock }

func (m *mockRoleAssignmentRepo) Create(ctx context.Context, ra *authorization.RoleAssignment) error {
	return m.Called(ctx, ra).Error(0)
}

func (m *mockRoleAssignmentRepo) ListByPrincipal(ctx context.Context, principalID uuid.UUID) ([]*authorization.RoleAssignment, error) {
	args := m.Called(ctx, principalID)
	assignments, _ := args.Get(0).([]*authorization.RoleAssignment)
	return assignments, args.Error(1)
}

// mockRoleRepo/mockPermissionCache back authorization.Service, which AuthHandlers now also
// depends on (to cache a freshly logged-in user's permissions) — same
// not-importable-from-here reasoning as the identity mocks above.

type mockRoleRepo struct{ mock.Mock }

func (m *mockRoleRepo) CreateSystem(ctx context.Context, r *authorization.Role) error {
	return m.Called(ctx, r).Error(0)
}

func (m *mockRoleRepo) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]*authorization.Role, error) {
	args := m.Called(ctx, ids)
	roles, _ := args.Get(0).([]*authorization.Role)
	return roles, args.Error(1)
}

type mockPermissionCache struct{ mock.Mock }

func (m *mockPermissionCache) SetUserPermissions(ctx context.Context, appID, userID uuid.UUID, permissions []string, ttl time.Duration) error {
	return m.Called(ctx, appID, userID, permissions, ttl).Error(0)
}

func (m *mockPermissionCache) GetUserPermissions(ctx context.Context, appID, userID uuid.UUID) ([]byte, error) {
	args := m.Called(ctx, appID, userID)
	raw, _ := args.Get(0).([]byte)
	return raw, args.Error(1)
}

type mockTokenIssuer struct{ mock.Mock }

func (m *mockTokenIssuer) Issue(a *app.App, claims app.Claims, ttl time.Duration) (string, error) {
	args := m.Called(a, claims, ttl)
	return args.String(0), args.Error(1)
}

func (m *mockTokenIssuer) Verify(a *app.App, tokenString string) (app.Claims, error) {
	args := m.Called(a, tokenString)
	c, _ := args.Get(0).(app.Claims)
	return c, args.Error(1)
}

// noopTxRunner just invokes fn directly — tests use mocked repositories with no real
// database underneath, so there's nothing to actually commit/roll back.
type noopTxRunner struct{}

func (noopTxRunner) RunInTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}
