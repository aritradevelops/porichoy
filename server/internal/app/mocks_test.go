package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type mockAppRepo struct{ mock.Mock }

func (m *mockAppRepo) CreateSystem(ctx context.Context, a *App) error {
	return m.Called(ctx, a).Error(0)
}

func (m *mockAppRepo) FindSystemAppByTenant(ctx context.Context, tenantID uuid.UUID) (*App, error) {
	args := m.Called(ctx, tenantID)
	a, _ := args.Get(0).(*App)
	return a, args.Error(1)
}

func (m *mockAppRepo) SetDefaultSignupRole(ctx context.Context, appID, roleID uuid.UUID) error {
	return m.Called(ctx, appID, roleID).Error(0)
}
