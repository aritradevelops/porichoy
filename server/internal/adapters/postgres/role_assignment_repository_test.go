//go:build integration

package postgres

import (
	"context"
	"testing"

	"github.com/aritradevelops/porichoy/server/internal/authorization"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestRoleAssignmentRepository_Create(t *testing.T) {
	tenants := NewTenantRepository(testDB)
	apps := NewAppRepository(testDB)
	roles := NewRoleRepository(testDB)
	users := NewUserRepository(testDB)
	assignments := NewRoleAssignmentRepository(testDB)
	ctx := context.Background()

	tt := mustCreateRoot(t, tenants, "Assignment Tenant")
	sysApp := newTestSystemApp(tt.ID)
	require.NoError(t, apps.CreateSystem(ctx, sysApp))
	role := newTestSystemRole(tt.ID, sysApp.ID, "Super Admin")
	require.NoError(t, roles.CreateSystem(ctx, role))
	u := newTestUser(tt.ID, "assign-"+uuid.NewString()+"@example.com")
	require.NoError(t, users.Create(ctx, u))

	ra := &authorization.RoleAssignment{
		ID:          uuid.New(),
		PrincipalID: u.ID,
		RoleID:      role.ID,
	}
	require.NoError(t, assignments.Create(ctx, ra))

	var count int
	err := testDB.NewSelect().Table("role_assignments").Where("id = ?", ra.ID).ColumnExpr("count(*)").Scan(ctx, &count)
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

func TestRoleAssignmentRepository_ListByPrincipal(t *testing.T) {
	tenants := NewTenantRepository(testDB)
	apps := NewAppRepository(testDB)
	roles := NewRoleRepository(testDB)
	users := NewUserRepository(testDB)
	assignments := NewRoleAssignmentRepository(testDB)
	ctx := context.Background()

	tt := mustCreateRoot(t, tenants, "ListByPrincipal Tenant")
	sysApp := newTestSystemApp(tt.ID)
	require.NoError(t, apps.CreateSystem(ctx, sysApp))
	role1 := newTestSystemRole(tt.ID, sysApp.ID, "Super Admin")
	require.NoError(t, roles.CreateSystem(ctx, role1))
	role2 := newTestSystemRole(tt.ID, sysApp.ID, "Tenant Admin")
	require.NoError(t, roles.CreateSystem(ctx, role2))
	u := newTestUser(tt.ID, "list-by-principal-"+uuid.NewString()+"@example.com")
	require.NoError(t, users.Create(ctx, u))
	other := newTestUser(tt.ID, "other-"+uuid.NewString()+"@example.com")
	require.NoError(t, users.Create(ctx, other))

	require.NoError(t, assignments.Create(ctx, &authorization.RoleAssignment{ID: uuid.New(), PrincipalID: u.ID, RoleID: role1.ID}))
	require.NoError(t, assignments.Create(ctx, &authorization.RoleAssignment{ID: uuid.New(), PrincipalID: u.ID, RoleID: role2.ID}))
	// Assigned to a different principal — proves the lookup filters by principal, not just
	// "return everything".
	require.NoError(t, assignments.Create(ctx, &authorization.RoleAssignment{ID: uuid.New(), PrincipalID: other.ID, RoleID: role1.ID}))

	got, err := assignments.ListByPrincipal(ctx, u.ID)

	require.NoError(t, err)
	require.Len(t, got, 2)
	gotRoleIDs := []uuid.UUID{got[0].RoleID, got[1].RoleID}
	require.ElementsMatch(t, []uuid.UUID{role1.ID, role2.ID}, gotRoleIDs)
}

func TestRoleAssignmentRepository_ListByPrincipal_None(t *testing.T) {
	assignments := NewRoleAssignmentRepository(testDB)

	got, err := assignments.ListByPrincipal(context.Background(), uuid.New())

	require.NoError(t, err)
	require.Empty(t, got)
}
