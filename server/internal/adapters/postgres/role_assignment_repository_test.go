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
