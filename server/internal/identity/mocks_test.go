package identity

import (
	"context"
	"time"

	"github.com/aritradevelops/porichoy/server/internal/app"
	"github.com/aritradevelops/porichoy/server/internal/authorization"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type mockUserRepo struct{ mock.Mock }

func (m *mockUserRepo) Create(ctx context.Context, u *User) error {
	return m.Called(ctx, u).Error(0)
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*User, error) {
	args := m.Called(ctx, tenantID, email)
	u, _ := args.Get(0).(*User)
	return u, args.Error(1)
}

type mockPasswordRepo struct{ mock.Mock }

func (m *mockPasswordRepo) Create(ctx context.Context, p *Password) error {
	return m.Called(ctx, p).Error(0)
}

type mockAppRepo struct{ mock.Mock }

func (m *mockAppRepo) CreateSystem(ctx context.Context, a *app.App) error {
	return m.Called(ctx, a).Error(0)
}

func (m *mockAppRepo) FindSystemAppByTenant(ctx context.Context, tenantID uuid.UUID) (*app.App, error) {
	args := m.Called(ctx, tenantID)
	a, _ := args.Get(0).(*app.App)
	return a, args.Error(1)
}

func (m *mockAppRepo) SetDefaultSignupRole(ctx context.Context, appID, roleID uuid.UUID) error {
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

type mockTxRunner struct{}

// RunInTx just invokes fn directly with ctx unchanged — no real transaction needed for unit
// tests, which use mocked repositories with no actual database underneath.
func (mockTxRunner) RunInTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}
