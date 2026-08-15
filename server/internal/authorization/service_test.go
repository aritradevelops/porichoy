package authorization

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_CreateSystemRole(t *testing.T) {
	roles := &mockRoleRepo{}
	assignments := &mockRoleAssignmentRepo{}
	svc := NewService(roles, assignments)
	tenantID, appID := uuid.New(), uuid.New()
	roles.On("CreateSystem", mock.Anything, mock.AnythingOfType("*authorization.Role")).Return(nil)

	r, err := svc.CreateSystemRole(context.Background(), tenantID, appID, "Super Admin", []string{"tenants:*@root"})

	require.NoError(t, err)
	require.Equal(t, tenantID, r.TenantID)
	require.Equal(t, appID, *r.AppID)
	require.Equal(t, "Super Admin", r.Name)
	require.True(t, r.IsSystem)
	require.Equal(t, []string{"tenants:*@root"}, r.Permissions)
	roles.AssertExpectations(t)
}

func TestService_AssignSystemRole(t *testing.T) {
	roles := &mockRoleRepo{}
	assignments := &mockRoleAssignmentRepo{}
	svc := NewService(roles, assignments)
	principalID, roleID := uuid.New(), uuid.New()
	assignments.On("Create", mock.Anything, mock.AnythingOfType("*authorization.RoleAssignment")).Return(nil)

	ra, err := svc.AssignSystemRole(context.Background(), principalID, roleID)

	require.NoError(t, err)
	require.Equal(t, principalID, ra.PrincipalID)
	require.Equal(t, roleID, ra.RoleID)
	require.Nil(t, ra.CreatedBy)
	assignments.AssertExpectations(t)
}
