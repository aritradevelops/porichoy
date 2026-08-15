//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/aritradevelops/porichoy/server/internal/authorization"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func newTestSystemRole(tenantID, appID uuid.UUID, name string) *authorization.Role {
	now := time.Now().UTC().Truncate(time.Microsecond)
	return &authorization.Role{
		ID:          uuid.New(),
		TenantID:    tenantID,
		AppID:       &appID,
		Name:        name,
		IsSystem:    true,
		Permissions: []string{"tenants:*@root"},
		Policies:    []byte("[]"),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func TestRoleRepository_CreateSystem(t *testing.T) {
	tenants := NewTenantRepository(testDB)
	apps := NewAppRepository(testDB)
	roles := NewRoleRepository(testDB)
	ctx := context.Background()

	tt := mustCreateRoot(t, tenants, "Role Tenant")
	sysApp := newTestSystemApp(tt.ID)
	require.NoError(t, apps.CreateSystem(ctx, sysApp))

	role := newTestSystemRole(tt.ID, sysApp.ID, "Super Admin")
	require.NoError(t, roles.CreateSystem(ctx, role))

	var count int
	err := testDB.NewSelect().Table("roles").Where("id = ?", role.ID).ColumnExpr("count(*)").Scan(ctx, &count)
	require.NoError(t, err)
	require.Equal(t, 1, count)
}
