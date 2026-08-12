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

func newTestCredential(tenantID uuid.UUID, providerType tenant.ProviderType, ciphertext string) *tenant.ProviderCredential {
	now := time.Now().UTC().Truncate(time.Microsecond)
	return &tenant.ProviderCredential{
		ID:              uuid.New(),
		TenantID:        tenantID,
		ProviderType:    providerType,
		ConfigEncrypted: []byte(ciphertext),
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

func TestProviderCredentialRepository_UpsertCreatesThenUpdates(t *testing.T) {
	tenants := NewTenantRepository(testDB)
	creds := NewProviderCredentialRepository(testDB)
	ctx := context.Background()

	tt := newTestTenant("Brand G", nil)
	require.NoError(t, tenants.Create(ctx, tt))

	c := newTestCredential(tt.ID, tenant.ProviderTypeGoogle, "old-ciphertext")
	require.NoError(t, creds.Upsert(ctx, c))

	act := actor.Actor{PrincipalID: uuid.New(), TenantID: tt.ID, Scope: actor.ScopeTenant}
	got, err := creds.FindByTenantAndType(ctx, act, tt.ID, tenant.ProviderTypeGoogle)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, []byte("old-ciphertext"), got.ConfigEncrypted)

	// Same ID, new ciphertext — Upsert must update in place, not insert a second row.
	got.ConfigEncrypted = []byte("new-ciphertext")
	got.UpdatedAt = time.Now().UTC().Truncate(time.Microsecond)
	require.NoError(t, creds.Upsert(ctx, got))

	list, err := creds.ListByTenant(ctx, act, tt.ID)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, []byte("new-ciphertext"), list[0].ConfigEncrypted)
}

func TestProviderCredentialRepository_FindByTenantAndType_NotFound(t *testing.T) {
	tenants := NewTenantRepository(testDB)
	creds := NewProviderCredentialRepository(testDB)
	ctx := context.Background()

	tt := newTestTenant("Brand H", nil)
	require.NoError(t, tenants.Create(ctx, tt))

	act := actor.Actor{PrincipalID: uuid.New(), TenantID: tt.ID, Scope: actor.ScopeTenant}
	got, err := creds.FindByTenantAndType(ctx, act, tt.ID, tenant.ProviderTypeApple)

	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestProviderCredentialRepository_SoftDeleteThenReAdd(t *testing.T) {
	tenants := NewTenantRepository(testDB)
	creds := NewProviderCredentialRepository(testDB)
	ctx := context.Background()

	tt := newTestTenant("Brand I", nil)
	require.NoError(t, tenants.Create(ctx, tt))

	c := newTestCredential(tt.ID, tenant.ProviderTypeOTPEmail, "ciphertext-1")
	require.NoError(t, creds.Upsert(ctx, c))

	act := actor.Actor{PrincipalID: uuid.New(), TenantID: tt.ID, Scope: actor.ScopeTenant}
	require.NoError(t, creds.SoftDelete(ctx, act, c.ID))

	got, err := creds.FindByTenantAndType(ctx, act, tt.ID, tenant.ProviderTypeOTPEmail)
	require.NoError(t, err)
	assert.Nil(t, got)

	// The partial unique index on (tenant_id, provider_type) only covers active rows, so a
	// new credential for the same pair must be insertable after the old one is deleted.
	replacement := newTestCredential(tt.ID, tenant.ProviderTypeOTPEmail, "ciphertext-2")
	require.NoError(t, creds.Upsert(ctx, replacement))
}
