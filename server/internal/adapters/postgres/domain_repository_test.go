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

func newTestDomain(tenantID uuid.UUID, domain string) *tenant.TenantDomain {
	return &tenant.TenantDomain{
		ID:        uuid.New(),
		TenantID:  tenantID,
		Domain:    domain,
		CreatedAt: time.Now().UTC().Truncate(time.Microsecond),
	}
}

func TestDomainRepository_CreateAndFindByDomain(t *testing.T) {
	tenants := NewTenantRepository(testDB)
	domains := NewDomainRepository(testDB)
	ctx := context.Background()

	tt := newTestTenant("Brand C", nil)
	require.NoError(t, tenants.Create(ctx, tt))

	d := newTestDomain(tt.ID, "brand-c.example.com")
	require.NoError(t, domains.Create(ctx, d))

	got, err := domains.FindByDomain(ctx, "brand-c.example.com")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, tt.ID, got.TenantID)
}

func TestDomainRepository_FindByDomain_NotFound(t *testing.T) {
	domains := NewDomainRepository(testDB)

	got, err := domains.FindByDomain(context.Background(), "nowhere.example.com")

	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestDomainRepository_UniqueAmongActiveRows(t *testing.T) {
	tenants := NewTenantRepository(testDB)
	domains := NewDomainRepository(testDB)
	ctx := context.Background()

	tt := newTestTenant("Brand D", nil)
	require.NoError(t, tenants.Create(ctx, tt))

	first := newTestDomain(tt.ID, "brand-d.example.com")
	require.NoError(t, domains.Create(ctx, first))

	// A second active row for the same domain must violate the partial unique index.
	second := newTestDomain(tt.ID, "brand-d.example.com")
	require.Error(t, domains.Create(ctx, second))

	// But once the first is soft-deleted, the domain is free to be re-registered.
	act := actor.Actor{PrincipalID: uuid.New(), TenantID: tt.ID, Scope: actor.ScopeTenant}
	require.NoError(t, domains.SoftDelete(ctx, act, first.ID))

	third := newTestDomain(tt.ID, "brand-d.example.com")
	require.NoError(t, domains.Create(ctx, third))
}

func TestDomainRepository_ListByTenant(t *testing.T) {
	tenants := NewTenantRepository(testDB)
	domains := NewDomainRepository(testDB)
	ctx := context.Background()

	tt := newTestTenant("Brand E", nil)
	require.NoError(t, tenants.Create(ctx, tt))

	require.NoError(t, domains.Create(ctx, newTestDomain(tt.ID, "one.brand-e.example.com")))
	require.NoError(t, domains.Create(ctx, newTestDomain(tt.ID, "two.brand-e.example.com")))

	other := newTestTenant("Brand F", nil)
	require.NoError(t, tenants.Create(ctx, other))
	require.NoError(t, domains.Create(ctx, newTestDomain(other.ID, "brand-f.example.com")))

	act := actor.Actor{PrincipalID: uuid.New(), TenantID: tt.ID, Scope: actor.ScopeTenant}
	list, err := domains.ListByTenant(ctx, act, tt.ID)

	require.NoError(t, err)
	assert.Len(t, list, 2)
}
