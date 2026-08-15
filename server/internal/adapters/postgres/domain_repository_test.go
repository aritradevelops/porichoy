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

func TestDomainRepository_CreateRoot(t *testing.T) {
	tenants := NewTenantRepository(testDB)
	domains := NewDomainRepository(testDB)
	ctx := context.Background()

	tt := newTestTenant("Root Domain Tenant", nil)
	require.NoError(t, tenants.CreateRoot(ctx, tt))

	d := newTestDomain(tt.ID, "root-domain-"+uuid.NewString()+".example.com")
	require.NoError(t, domains.CreateRoot(ctx, d))

	got, err := domains.FindByDomain(ctx, d.Domain)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, tt.ID, got.TenantID)
}

func TestDomainRepository_CreateAndFindByDomain(t *testing.T) {
	tenants := NewTenantRepository(testDB)
	domains := NewDomainRepository(testDB)
	ctx := context.Background()
	root := rootActor()

	tt := newTestTenant("Brand C", nil)
	require.NoError(t, tenants.CreateRoot(ctx, tt))

	d := newTestDomain(tt.ID, "brand-c.example.com")
	require.NoError(t, domains.Create(ctx, root, d))

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
	root := rootActor()

	tt := newTestTenant("Brand D", nil)
	require.NoError(t, tenants.CreateRoot(ctx, tt))

	first := newTestDomain(tt.ID, "brand-d.example.com")
	require.NoError(t, domains.Create(ctx, root, first))

	// A second active row for the same domain must violate the partial unique index.
	second := newTestDomain(tt.ID, "brand-d.example.com")
	require.Error(t, domains.Create(ctx, root, second))

	// But once the first is soft-deleted, the domain is free to be re-registered.
	act := actor.Actor{PrincipalID: uuid.New(), TenantID: tt.ID, Scope: actor.ScopeTenant}
	require.NoError(t, domains.SoftDelete(ctx, act, first.ID))

	third := newTestDomain(tt.ID, "brand-d.example.com")
	require.NoError(t, domains.Create(ctx, root, third))
}

func TestDomainRepository_ListByTenant(t *testing.T) {
	tenants := NewTenantRepository(testDB)
	domains := NewDomainRepository(testDB)
	ctx := context.Background()
	root := rootActor()

	tt := newTestTenant("Brand E", nil)
	require.NoError(t, tenants.CreateRoot(ctx, tt))

	require.NoError(t, domains.Create(ctx, root, newTestDomain(tt.ID, "one.brand-e.example.com")))
	require.NoError(t, domains.Create(ctx, root, newTestDomain(tt.ID, "two.brand-e.example.com")))

	other := newTestTenant("Brand F", nil)
	require.NoError(t, tenants.CreateRoot(ctx, other))
	require.NoError(t, domains.Create(ctx, root, newTestDomain(other.ID, "brand-f.example.com")))

	act := actor.Actor{PrincipalID: uuid.New(), TenantID: tt.ID, Scope: actor.ScopeTenant}
	list, err := domains.ListByTenant(ctx, act, tt.ID)

	require.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestDomainRepository_Create_SucceedsForOwnAndDescendantTenant(t *testing.T) {
	tenants := NewTenantRepository(testDB)
	domains := NewDomainRepository(testDB)
	ctx := context.Background()
	root := rootActor()

	self := mustCreateRoot(t, tenants, "Domain Self")
	child := mustCreateChild(t, tenants, root, "Domain Child", self.ID)

	act := actor.Actor{PrincipalID: uuid.New(), TenantID: self.ID, Scope: actor.ScopeTenant}

	require.NoError(t, domains.Create(ctx, act, newTestDomain(self.ID, "own.domain-scope.example.com")))
	require.NoError(t, domains.Create(ctx, act, newTestDomain(child.ID, "descendant.domain-scope.example.com")))
}

func TestDomainRepository_Create_DeniedForUnrelatedTenant(t *testing.T) {
	tenants := NewTenantRepository(testDB)
	domains := NewDomainRepository(testDB)
	ctx := context.Background()

	self := mustCreateRoot(t, tenants, "Domain Self 2")
	unrelated := mustCreateRoot(t, tenants, "Domain Unrelated")

	act := actor.Actor{PrincipalID: uuid.New(), TenantID: self.ID, Scope: actor.ScopeTenant}
	err := domains.Create(ctx, act, newTestDomain(unrelated.ID, "unrelated.domain-scope.example.com"))

	require.ErrorIs(t, err, tenant.ErrTenantNotFound)
}

func TestDomainRepository_Create_NonTenantScopeDenied(t *testing.T) {
	tenants := NewTenantRepository(testDB)
	domains := NewDomainRepository(testDB)
	ctx := context.Background()

	self := mustCreateRoot(t, tenants, "Domain Self 3")

	act := actor.Actor{PrincipalID: uuid.New(), TenantID: self.ID, Scope: actor.ScopeOrg}
	err := domains.Create(ctx, act, newTestDomain(self.ID, "org-scope.domain-scope.example.com"))

	require.ErrorIs(t, err, tenant.ErrTenantNotFound)
}

func TestDomainRepository_ListByTenant_DeniedForUnrelatedTenantReturnsEmpty(t *testing.T) {
	tenants := NewTenantRepository(testDB)
	domains := NewDomainRepository(testDB)
	ctx := context.Background()
	root := rootActor()

	self := mustCreateRoot(t, tenants, "Domain Self 4")
	unrelated := mustCreateRoot(t, tenants, "Domain Unrelated 2")
	require.NoError(t, domains.Create(ctx, root, newTestDomain(unrelated.ID, "hidden.domain-scope.example.com")))

	act := actor.Actor{PrincipalID: uuid.New(), TenantID: self.ID, Scope: actor.ScopeTenant}
	list, err := domains.ListByTenant(ctx, act, unrelated.ID)

	require.NoError(t, err)
	assert.Empty(t, list)
}

func TestDomainRepository_SoftDelete_DeniedOutsideSubtreeIsNoop(t *testing.T) {
	tenants := NewTenantRepository(testDB)
	domains := NewDomainRepository(testDB)
	ctx := context.Background()
	root := rootActor()

	self := mustCreateRoot(t, tenants, "Domain Self 5")
	unrelated := mustCreateRoot(t, tenants, "Domain Unrelated 3")
	d := newTestDomain(unrelated.ID, "survives.domain-scope.example.com")
	require.NoError(t, domains.Create(ctx, root, d))

	act := actor.Actor{PrincipalID: uuid.New(), TenantID: self.ID, Scope: actor.ScopeTenant}
	require.NoError(t, domains.SoftDelete(ctx, act, d.ID))

	got, err := domains.FindByDomain(ctx, "survives.domain-scope.example.com")
	require.NoError(t, err)
	require.NotNil(t, got, "domain must survive a denied delete attempt")
}

func TestDomainRepository_SoftDelete_SucceedsForDescendantTenant(t *testing.T) {
	tenants := NewTenantRepository(testDB)
	domains := NewDomainRepository(testDB)
	ctx := context.Background()
	root := rootActor()

	self := mustCreateRoot(t, tenants, "Domain Self 6")
	child := mustCreateChild(t, tenants, root, "Domain Child 2", self.ID)
	d := newTestDomain(child.ID, "deleted.domain-scope.example.com")
	require.NoError(t, domains.Create(ctx, root, d))

	act := actor.Actor{PrincipalID: uuid.New(), TenantID: self.ID, Scope: actor.ScopeTenant}
	require.NoError(t, domains.SoftDelete(ctx, act, d.ID))

	got, err := domains.FindByDomain(ctx, "deleted.domain-scope.example.com")
	require.NoError(t, err)
	assert.Nil(t, got)
}
