package authorization

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_CreateSystemRole(t *testing.T) {
	roles := &mockRoleRepo{}
	assignments := &mockRoleAssignmentRepo{}
	svc := NewService(roles, assignments, &mockPermissionCache{})
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
	svc := NewService(roles, assignments, &mockPermissionCache{})
	principalID, roleID := uuid.New(), uuid.New()
	assignments.On("Create", mock.Anything, mock.AnythingOfType("*authorization.RoleAssignment")).Return(nil)

	ra, err := svc.AssignSystemRole(context.Background(), principalID, roleID)

	require.NoError(t, err)
	require.Equal(t, principalID, ra.PrincipalID)
	require.Equal(t, roleID, ra.RoleID)
	require.Nil(t, ra.CreatedBy)
	assignments.AssertExpectations(t)
}

func TestService_EffectivePermissions_DedupesAcrossRoles(t *testing.T) {
	roles := &mockRoleRepo{}
	assignments := &mockRoleAssignmentRepo{}
	svc := NewService(roles, assignments, &mockPermissionCache{})
	principalID := uuid.New()
	role1ID, role2ID := uuid.New(), uuid.New()

	assignments.On("ListByPrincipal", mock.Anything, principalID).Return([]*RoleAssignment{
		{ID: uuid.New(), PrincipalID: principalID, RoleID: role1ID},
		{ID: uuid.New(), PrincipalID: principalID, RoleID: role2ID},
	}, nil)
	roles.On("FindByIDs", mock.Anything, []uuid.UUID{role1ID, role2ID}).Return([]*Role{
		{ID: role1ID, Permissions: []string{"tenants:*@root", "domains:*@root"}},
		{ID: role2ID, Permissions: []string{"domains:*@root", "provider_credentials:*@root"}},
	}, nil)

	got, err := svc.EffectivePermissions(context.Background(), principalID)

	require.NoError(t, err)
	require.Equal(t, []string{"tenants:*@root", "domains:*@root", "provider_credentials:*@root"}, got)
}

func TestService_EffectivePermissions_NoAssignments(t *testing.T) {
	roles := &mockRoleRepo{}
	assignments := &mockRoleAssignmentRepo{}
	svc := NewService(roles, assignments, &mockPermissionCache{})
	principalID := uuid.New()
	assignments.On("ListByPrincipal", mock.Anything, principalID).Return(nil, nil)

	got, err := svc.EffectivePermissions(context.Background(), principalID)

	require.NoError(t, err)
	require.Empty(t, got)
	roles.AssertNotCalled(t, "FindByIDs", mock.Anything, mock.Anything)
}

func TestService_CacheUserPermissions_OK(t *testing.T) {
	roles := &mockRoleRepo{}
	assignments := &mockRoleAssignmentRepo{}
	permCache := &mockPermissionCache{}
	svc := NewService(roles, assignments, permCache)
	tenantID, userID, roleID := uuid.New(), uuid.New(), uuid.New()

	assignments.On("ListByPrincipal", mock.Anything, userID).Return([]*RoleAssignment{
		{ID: uuid.New(), PrincipalID: userID, RoleID: roleID},
	}, nil)
	roles.On("FindByIDs", mock.Anything, []uuid.UUID{roleID}).Return([]*Role{
		{ID: roleID, Permissions: []string{"tenants:*@root"}},
	}, nil)
	permCache.On("SetUserPermissions", mock.Anything, tenantID, userID, []string{"tenants:*@root"}, time.Minute).Return(nil)

	err := svc.CacheUserPermissions(context.Background(), tenantID, userID, time.Minute)

	require.NoError(t, err)
	permCache.AssertExpectations(t)
}
