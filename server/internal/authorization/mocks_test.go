package authorization

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type mockRoleRepo struct{ mock.Mock }

func (m *mockRoleRepo) CreateSystem(ctx context.Context, r *Role) error {
	return m.Called(ctx, r).Error(0)
}

type mockRoleAssignmentRepo struct{ mock.Mock }

func (m *mockRoleAssignmentRepo) Create(ctx context.Context, ra *RoleAssignment) error {
	return m.Called(ctx, ra).Error(0)
}
