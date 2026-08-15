//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/aritradevelops/porichoy/server/internal/app"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestSystemApp(tenantID uuid.UUID) *app.App {
	now := time.Now().UTC().Truncate(time.Microsecond)
	return &app.App{
		ID:                     uuid.New(),
		TenantID:               tenantID,
		Name:                   "System",
		ClientID:               uuid.NewString(),
		IsSystem:               true,
		SigningAlgorithm:       app.SigningAlgorithmHS256,
		SigningKeyConfig:       []byte("test-secret"),
		AccessTokenTTLSeconds:  app.DefaultAccessTokenTTLSeconds,
		IDTokenTTLSeconds:      app.DefaultIDTokenTTLSeconds,
		RefreshTokenTTLSeconds: app.DefaultRefreshTokenTTLSeconds,
		CreatedAt:              now,
		UpdatedAt:              now,
	}
}

func TestAppRepository_CreateSystemAndFindByTenant(t *testing.T) {
	tenants := NewTenantRepository(testDB)
	apps := NewAppRepository(testDB)
	ctx := context.Background()

	tt := mustCreateRoot(t, tenants, "App Tenant")
	sysApp := newTestSystemApp(tt.ID)
	require.NoError(t, apps.CreateSystem(ctx, sysApp))

	got, err := apps.FindSystemAppByTenant(ctx, tt.ID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, sysApp.ClientID, got.ClientID)
	assert.True(t, got.IsSystem)
	assert.Equal(t, app.SigningAlgorithmHS256, got.SigningAlgorithm)
	assert.Equal(t, sysApp.SigningKeyConfig, got.SigningKeyConfig)
}

func TestAppRepository_FindSystemAppByTenant_NotFound(t *testing.T) {
	tenants := NewTenantRepository(testDB)
	apps := NewAppRepository(testDB)
	ctx := context.Background()

	tt := mustCreateRoot(t, tenants, "No System App")

	got, err := apps.FindSystemAppByTenant(ctx, tt.ID)
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestAppRepository_SetDefaultSignupRole(t *testing.T) {
	tenants := NewTenantRepository(testDB)
	apps := NewAppRepository(testDB)
	roles := NewRoleRepository(testDB)
	ctx := context.Background()

	tt := mustCreateRoot(t, tenants, "Role Default Tenant")
	sysApp := newTestSystemApp(tt.ID)
	require.NoError(t, apps.CreateSystem(ctx, sysApp))

	role := newTestSystemRole(tt.ID, sysApp.ID, "User")
	require.NoError(t, roles.CreateSystem(ctx, role))

	require.NoError(t, apps.SetDefaultSignupRole(ctx, sysApp.ID, role.ID))

	got, err := apps.FindSystemAppByTenant(ctx, tt.ID)
	require.NoError(t, err)
	require.NotNil(t, got.DefaultSignupRoleID)
	assert.Equal(t, role.ID, *got.DefaultSignupRoleID)
}
