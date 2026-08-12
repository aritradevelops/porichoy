//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/aritradevelops/porichoy/server/internal/actor"
	"github.com/aritradevelops/porichoy/server/internal/tenant"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestTenant(name string, parentID *uuid.UUID) *tenant.Tenant {
	now := time.Now().UTC().Truncate(time.Microsecond)
	return &tenant.Tenant{
		ID:                  uuid.New(),
		ParentID:            parentID,
		Name:                name,
		LoginLayout:         tenant.LoginLayoutCentered,
		EnabledLoginMethods: []tenant.LoginMethod{tenant.LoginMethodEmailPassword, tenant.LoginMethodGoogle},
		AuditRetentionDays:  90,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
}

func TestTenantRepository_CreateAndFindByID(t *testing.T) {
	repo := NewTenantRepository(testDB)
	ctx := context.Background()

	tt := newTestTenant("Brand A", nil)
	require.NoError(t, repo.Create(ctx, tt))

	got, err := repo.FindByID(ctx, tt.ID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "Brand A", got.Name)
	assert.True(t, got.IsRoot())
	assert.Equal(t, tenant.LoginLayoutCentered, got.LoginLayout)
	assert.ElementsMatch(t, tt.EnabledLoginMethods, got.EnabledLoginMethods)
	assert.Equal(t, 90, got.AuditRetentionDays)
}

func TestTenantRepository_FindByID_NotFound(t *testing.T) {
	repo := NewTenantRepository(testDB)

	got, err := repo.FindByID(context.Background(), uuid.New())

	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestTenantRepository_GetByID_AuthorizedLookup(t *testing.T) {
	repo := NewTenantRepository(testDB)
	ctx := context.Background()

	tt := newTestTenant("Brand B", nil)
	require.NoError(t, repo.Create(ctx, tt))

	act := actor.Actor{PrincipalID: uuid.New(), TenantID: tt.ID, Scope: actor.ScopeTenant}
	got, err := repo.GetByID(ctx, act, tt.ID)

	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "Brand B", got.Name)
}

func TestTenantRepository_Update(t *testing.T) {
	repo := NewTenantRepository(testDB)
	ctx := context.Background()

	tt := newTestTenant("Original", nil)
	require.NoError(t, repo.Create(ctx, tt))

	tt.Name = "Renamed"
	tt.MFARequired = true
	require.NoError(t, repo.Update(ctx, tt))

	got, err := repo.FindByID(ctx, tt.ID)
	require.NoError(t, err)
	assert.Equal(t, "Renamed", got.Name)
	assert.True(t, got.MFARequired)
}

func TestTenantRepository_SoftDelete_ExcludesFromFindByID(t *testing.T) {
	repo := NewTenantRepository(testDB)
	ctx := context.Background()

	tt := newTestTenant("To Delete", nil)
	require.NoError(t, repo.Create(ctx, tt))

	act := actor.Actor{PrincipalID: uuid.New(), TenantID: tt.ID, Scope: actor.ScopeRoot}
	require.NoError(t, repo.SoftDelete(ctx, act, tt.ID))

	got, err := repo.FindByID(ctx, tt.ID)
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestTenantRepository_ListChildren(t *testing.T) {
	repo := NewTenantRepository(testDB)
	ctx := context.Background()

	parent := newTestTenant("Parent", nil)
	require.NoError(t, repo.Create(ctx, parent))

	child := newTestTenant("Child", &parent.ID)
	require.NoError(t, repo.Create(ctx, child))

	unrelated := newTestTenant("Unrelated", nil)
	require.NoError(t, repo.Create(ctx, unrelated))

	act := actor.Actor{PrincipalID: uuid.New(), TenantID: parent.ID, Scope: actor.ScopeRoot}
	children, err := repo.ListChildren(ctx, act, parent.ID)

	require.NoError(t, err)
	require.Len(t, children, 1)
	assert.Equal(t, "Child", children[0].Name)
}
