package tenant

import (
	"context"

	"github.com/aritradevelops/porichoy/server/internal/actor"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type mockTenantRepo struct{ mock.Mock }

func (m *mockTenantRepo) CreateRoot(ctx context.Context, t *Tenant) error {
	return m.Called(ctx, t).Error(0)
}

func (m *mockTenantRepo) Create(ctx context.Context, act actor.Actor, t *Tenant) error {
	return m.Called(ctx, act, t).Error(0)
}

func (m *mockTenantRepo) FindByID(ctx context.Context, id uuid.UUID) (*Tenant, error) {
	args := m.Called(ctx, id)
	t, _ := args.Get(0).(*Tenant)
	return t, args.Error(1)
}

func (m *mockTenantRepo) GetByID(ctx context.Context, act actor.Actor, id uuid.UUID) (*Tenant, error) {
	args := m.Called(ctx, act, id)
	t, _ := args.Get(0).(*Tenant)
	return t, args.Error(1)
}

func (m *mockTenantRepo) Update(ctx context.Context, t *Tenant) error {
	return m.Called(ctx, t).Error(0)
}

func (m *mockTenantRepo) SoftDelete(ctx context.Context, act actor.Actor, id uuid.UUID) error {
	return m.Called(ctx, act, id).Error(0)
}

func (m *mockTenantRepo) ListChildren(ctx context.Context, act actor.Actor, parentID uuid.UUID) ([]*Tenant, error) {
	args := m.Called(ctx, act, parentID)
	list, _ := args.Get(0).([]*Tenant)
	return list, args.Error(1)
}

func (m *mockTenantRepo) List(ctx context.Context, act actor.Actor, params ListParams) (ListResult, error) {
	args := m.Called(ctx, act, params)
	result, _ := args.Get(0).(ListResult)
	return result, args.Error(1)
}

type mockDomainRepo struct{ mock.Mock }

func (m *mockDomainRepo) Create(ctx context.Context, act actor.Actor, d *TenantDomain) error {
	return m.Called(ctx, act, d).Error(0)
}

func (m *mockDomainRepo) CreateRoot(ctx context.Context, d *TenantDomain) error {
	return m.Called(ctx, d).Error(0)
}

func (m *mockDomainRepo) FindByDomain(ctx context.Context, domain string) (*TenantDomain, error) {
	args := m.Called(ctx, domain)
	d, _ := args.Get(0).(*TenantDomain)
	return d, args.Error(1)
}

func (m *mockDomainRepo) ListByTenant(ctx context.Context, act actor.Actor, tenantID uuid.UUID) ([]*TenantDomain, error) {
	args := m.Called(ctx, act, tenantID)
	list, _ := args.Get(0).([]*TenantDomain)
	return list, args.Error(1)
}

func (m *mockDomainRepo) SoftDelete(ctx context.Context, act actor.Actor, id uuid.UUID) error {
	return m.Called(ctx, act, id).Error(0)
}

type mockCredentialRepo struct{ mock.Mock }

func (m *mockCredentialRepo) Upsert(ctx context.Context, act actor.Actor, c *ProviderCredential) error {
	return m.Called(ctx, act, c).Error(0)
}

func (m *mockCredentialRepo) FindByTenantAndType(ctx context.Context, act actor.Actor, tenantID uuid.UUID, providerType ProviderType) (*ProviderCredential, error) {
	args := m.Called(ctx, act, tenantID, providerType)
	c, _ := args.Get(0).(*ProviderCredential)
	return c, args.Error(1)
}

func (m *mockCredentialRepo) ListByTenant(ctx context.Context, act actor.Actor, tenantID uuid.UUID) ([]*ProviderCredential, error) {
	args := m.Called(ctx, act, tenantID)
	list, _ := args.Get(0).([]*ProviderCredential)
	return list, args.Error(1)
}

func (m *mockCredentialRepo) SoftDelete(ctx context.Context, act actor.Actor, id uuid.UUID) error {
	return m.Called(ctx, act, id).Error(0)
}
