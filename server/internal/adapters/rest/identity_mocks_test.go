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
// (app.Repository/SessionRepository/TokenIssuer, authorization.RoleAssignmentRepository) —
// same reasoning as mocks_test.go's tenant mocks: the internal/identity package's own mocks
// are unexported and not importable from here.

type mockUserRepo struct{ mock.Mock }

func (m *mockUserRepo) Create(ctx context.Context, u *identity.User) error {
	return m.Called(ctx, u).Error(0)
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*identity.User, error) {
	args := m.Called(ctx, tenantID, email)
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

type mockTokenIssuer struct{ mock.Mock }

func (m *mockTokenIssuer) Issue(a *app.App, claims app.Claims, ttl time.Duration) (string, error) {
	args := m.Called(a, claims, ttl)
	return args.String(0), args.Error(1)
}

// noopTxRunner just invokes fn directly — tests use mocked repositories with no real
// database underneath, so there's nothing to actually commit/roll back.
type noopTxRunner struct{}

func (noopTxRunner) RunInTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}
