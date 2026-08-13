package rest

import (
	"context"

	"github.com/aritradevelops/porichoy/server/internal/actor"
	"github.com/aritradevelops/porichoy/server/internal/tenant"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// Local mocks of tenant.Repository/DomainRepository/ProviderCredentialRepository — the
// equivalents in internal/tenant/mocks_test.go are unexported and package-internal, so not
// importable from here. Same testify/mock shape, kept in sync by hand since there's no
// shared test-support package for it (a small, deliberate duplication rather than
// introducing one for two files).

type mockTenantRepo struct{ mock.Mock }

func (m *mockTenantRepo) CreateRoot(ctx context.Context, t *tenant.Tenant) error {
	return m.Called(ctx, t).Error(0)
}

func (m *mockTenantRepo) Create(ctx context.Context, act actor.Actor, t *tenant.Tenant) error {
	return m.Called(ctx, act, t).Error(0)
}

func (m *mockTenantRepo) FindByID(ctx context.Context, id uuid.UUID) (*tenant.Tenant, error) {
	args := m.Called(ctx, id)
	t, _ := args.Get(0).(*tenant.Tenant)
	return t, args.Error(1)
}

func (m *mockTenantRepo) GetByID(ctx context.Context, act actor.Actor, id uuid.UUID) (*tenant.Tenant, error) {
	args := m.Called(ctx, act, id)
	t, _ := args.Get(0).(*tenant.Tenant)
	return t, args.Error(1)
}

func (m *mockTenantRepo) Update(ctx context.Context, t *tenant.Tenant) error {
	return m.Called(ctx, t).Error(0)
}

func (m *mockTenantRepo) SoftDelete(ctx context.Context, act actor.Actor, id uuid.UUID) error {
	return m.Called(ctx, act, id).Error(0)
}

func (m *mockTenantRepo) ListChildren(ctx context.Context, act actor.Actor, parentID uuid.UUID) ([]*tenant.Tenant, error) {
	args := m.Called(ctx, act, parentID)
	list, _ := args.Get(0).([]*tenant.Tenant)
	return list, args.Error(1)
}

type mockDomainRepo struct{ mock.Mock }

func (m *mockDomainRepo) Create(ctx context.Context, act actor.Actor, d *tenant.TenantDomain) error {
	return m.Called(ctx, act, d).Error(0)
}

func (m *mockDomainRepo) FindByDomain(ctx context.Context, domain string) (*tenant.TenantDomain, error) {
	args := m.Called(ctx, domain)
	d, _ := args.Get(0).(*tenant.TenantDomain)
	return d, args.Error(1)
}

func (m *mockDomainRepo) ListByTenant(ctx context.Context, act actor.Actor, tenantID uuid.UUID) ([]*tenant.TenantDomain, error) {
	args := m.Called(ctx, act, tenantID)
	list, _ := args.Get(0).([]*tenant.TenantDomain)
	return list, args.Error(1)
}

func (m *mockDomainRepo) SoftDelete(ctx context.Context, act actor.Actor, id uuid.UUID) error {
	return m.Called(ctx, act, id).Error(0)
}

type mockCredentialRepo struct{ mock.Mock }

func (m *mockCredentialRepo) Upsert(ctx context.Context, act actor.Actor, c *tenant.ProviderCredential) error {
	return m.Called(ctx, act, c).Error(0)
}

func (m *mockCredentialRepo) FindByTenantAndType(ctx context.Context, act actor.Actor, tenantID uuid.UUID, providerType tenant.ProviderType) (*tenant.ProviderCredential, error) {
	args := m.Called(ctx, act, tenantID, providerType)
	c, _ := args.Get(0).(*tenant.ProviderCredential)
	return c, args.Error(1)
}

func (m *mockCredentialRepo) ListByTenant(ctx context.Context, act actor.Actor, tenantID uuid.UUID) ([]*tenant.ProviderCredential, error) {
	args := m.Called(ctx, act, tenantID)
	list, _ := args.Get(0).([]*tenant.ProviderCredential)
	return list, args.Error(1)
}

func (m *mockCredentialRepo) SoftDelete(ctx context.Context, act actor.Actor, id uuid.UUID) error {
	return m.Called(ctx, act, id).Error(0)
}
