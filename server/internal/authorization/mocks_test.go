package authorization

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type mockRoleRepo struct{ mock.Mock }

func (m *mockRoleRepo) CreateSystem(ctx context.Context, r *Role) error {
	return m.Called(ctx, r).Error(0)
}

func (m *mockRoleRepo) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]*Role, error) {
	args := m.Called(ctx, ids)
	roles, _ := args.Get(0).([]*Role)
	return roles, args.Error(1)
}

type mockRoleAssignmentRepo struct{ mock.Mock }

func (m *mockRoleAssignmentRepo) Create(ctx context.Context, ra *RoleAssignment) error {
	return m.Called(ctx, ra).Error(0)
}

func (m *mockRoleAssignmentRepo) ListByPrincipal(ctx context.Context, principalID uuid.UUID) ([]*RoleAssignment, error) {
	args := m.Called(ctx, principalID)
	assignments, _ := args.Get(0).([]*RoleAssignment)
	return assignments, args.Error(1)
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
