package authorization

import (
	"context"
	"testing"
	"time"

	"github.com/aritradevelops/porichoy/server/internal/actor"
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
	appID, userID, roleID := uuid.New(), uuid.New(), uuid.New()

	assignments.On("ListByPrincipal", mock.Anything, userID).Return([]*RoleAssignment{
		{ID: uuid.New(), PrincipalID: userID, RoleID: roleID},
	}, nil)
	roles.On("FindByIDs", mock.Anything, []uuid.UUID{roleID}).Return([]*Role{
		{ID: roleID, Permissions: []string{"tenants:*@root"}},
	}, nil)
	permCache.On("SetUserPermissions", mock.Anything, appID, userID, []string{"tenants:*@root"}, time.Minute).Return(nil)

	err := svc.CacheUserPermissions(context.Background(), appID, userID, time.Minute)

	require.NoError(t, err)
	permCache.AssertExpectations(t)
}

func TestService_CachedPermissions_OK(t *testing.T) {
	roles := &mockRoleRepo{}
	assignments := &mockRoleAssignmentRepo{}
	permCache := &mockPermissionCache{}
	svc := NewService(roles, assignments, permCache)
	appID, userID := uuid.New(), uuid.New()

	permCache.On("GetUserPermissions", mock.Anything, appID, userID).
		Return([]byte(`["tenants:*@root","domains:register@tenant"]`), nil)

	got, err := svc.CachedPermissions(context.Background(), appID, userID)

	require.NoError(t, err)
	require.Equal(t, []string{"tenants:*@root", "domains:register@tenant"}, got)
}

func TestService_CachedPermissions_NoCacheEntry(t *testing.T) {
	roles := &mockRoleRepo{}
	assignments := &mockRoleAssignmentRepo{}
	permCache := &mockPermissionCache{}
	svc := NewService(roles, assignments, permCache)
	appID, userID := uuid.New(), uuid.New()

	// Unlike ResolveScope, a cache miss here is legitimate empty state, not ErrForbidden — a
	// validly authenticated caller can simply hold zero grants.
	permCache.On("GetUserPermissions", mock.Anything, appID, userID).Return(nil, nil)

	got, err := svc.CachedPermissions(context.Background(), appID, userID)

	require.NoError(t, err)
	require.Equal(t, []string{}, got)
}

func TestService_ResolveScope_PicksBroadestMatchingScope(t *testing.T) {
	roles := &mockRoleRepo{}
	assignments := &mockRoleAssignmentRepo{}
	permCache := &mockPermissionCache{}
	svc := NewService(roles, assignments, permCache)
	appID, userID := uuid.New(), uuid.New()

	permCache.On("GetUserPermissions", mock.Anything, appID, userID).
		Return([]byte(`["tenants:create@own","tenants:create@tenant","domains:*@root"]`), nil)

	scope, err := svc.ResolveScope(context.Background(), appID, userID, "tenants", "create")

	require.NoError(t, err)
	require.Equal(t, actor.ScopeTenant, scope)
}

func TestService_ResolveScope_MatchesWildcardAction(t *testing.T) {
	roles := &mockRoleRepo{}
	assignments := &mockRoleAssignmentRepo{}
	permCache := &mockPermissionCache{}
	svc := NewService(roles, assignments, permCache)
	appID, userID := uuid.New(), uuid.New()

	permCache.On("GetUserPermissions", mock.Anything, appID, userID).
		Return([]byte(`["tenants:*@root"]`), nil)

	scope, err := svc.ResolveScope(context.Background(), appID, userID, "tenants", "configure")

	require.NoError(t, err)
	require.Equal(t, actor.ScopeRoot, scope)
}

func TestService_ResolveScope_NoCacheEntry(t *testing.T) {
	roles := &mockRoleRepo{}
	assignments := &mockRoleAssignmentRepo{}
	permCache := &mockPermissionCache{}
	svc := NewService(roles, assignments, permCache)
	appID, userID := uuid.New(), uuid.New()

	permCache.On("GetUserPermissions", mock.Anything, appID, userID).Return(nil, nil)

	_, err := svc.ResolveScope(context.Background(), appID, userID, "tenants", "create")

	require.ErrorIs(t, err, ErrForbidden)
}

func TestService_ResolveScope_NoMatchingModule(t *testing.T) {
	roles := &mockRoleRepo{}
	assignments := &mockRoleAssignmentRepo{}
	permCache := &mockPermissionCache{}
	svc := NewService(roles, assignments, permCache)
	appID, userID := uuid.New(), uuid.New()

	permCache.On("GetUserPermissions", mock.Anything, appID, userID).
		Return([]byte(`["domains:*@root"]`), nil)

	_, err := svc.ResolveScope(context.Background(), appID, userID, "tenants", "create")

	require.ErrorIs(t, err, ErrForbidden)
}

func TestService_ResolveScope_DoesNotMatchPartialModuleName(t *testing.T) {
	roles := &mockRoleRepo{}
	assignments := &mockRoleAssignmentRepo{}
	permCache := &mockPermissionCache{}
	svc := NewService(roles, assignments, permCache)
	appID, userID := uuid.New(), uuid.New()

	// "tenants_extra" must not satisfy a "tenants" lookup — proves the match is anchored to
	// the whole module segment (via the surrounding quotes), not a substring search.
	permCache.On("GetUserPermissions", mock.Anything, appID, userID).
		Return([]byte(`["tenants_extra:*@root"]`), nil)

	_, err := svc.ResolveScope(context.Background(), appID, userID, "tenants", "create")

	require.ErrorIs(t, err, ErrForbidden)
}

func TestService_ResolveScope_ModuleWithRegexMetacharactersIsTreatedLiterally(t *testing.T) {
	roles := &mockRoleRepo{}
	assignments := &mockRoleAssignmentRepo{}
	permCache := &mockPermissionCache{}
	svc := NewService(roles, assignments, permCache)
	appID, userID := uuid.New(), uuid.New()

	// module/action come straight from the URL path (moduleActionFromPath) — a caller could
	// send a path segment containing regex metacharacters. ".*" must not act as a wildcard
	// glob over module names here; QuoteMeta should make it match only the literal string
	// ".*", which isn't present, so this must still be forbidden rather than matching
	// "tenants" through an unintended wildcard.
	permCache.On("GetUserPermissions", mock.Anything, appID, userID).
		Return([]byte(`["tenants:*@root"]`), nil)

	_, err := svc.ResolveScope(context.Background(), appID, userID, ".*", "create")

	require.ErrorIs(t, err, ErrForbidden)
}
