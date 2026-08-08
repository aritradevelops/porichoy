package tenant

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type mockTenantRepo struct{ mock.Mock }

func (m *mockTenantRepo) Create(ctx context.Context, t *Tenant) error {
	return m.Called(ctx, t).Error(0)
}

func (m *mockTenantRepo) GetByID(ctx context.Context, id uuid.UUID) (*Tenant, error) {
	args := m.Called(ctx, id)
	t, _ := args.Get(0).(*Tenant)
	return t, args.Error(1)
}

func (m *mockTenantRepo) Update(ctx context.Context, t *Tenant) error {
	return m.Called(ctx, t).Error(0)
}

func (m *mockTenantRepo) SoftDelete(ctx context.Context, id uuid.UUID, deletedBy *uuid.UUID) error {
	return m.Called(ctx, id, deletedBy).Error(0)
}

func (m *mockTenantRepo) ListChildren(ctx context.Context, parentID uuid.UUID) ([]*Tenant, error) {
	args := m.Called(ctx, parentID)
	list, _ := args.Get(0).([]*Tenant)
	return list, args.Error(1)
}

type mockDomainRepo struct{ mock.Mock }

func (m *mockDomainRepo) Create(ctx context.Context, d *TenantDomain) error {
	return m.Called(ctx, d).Error(0)
}

func (m *mockDomainRepo) FindByDomain(ctx context.Context, domain string) (*TenantDomain, error) {
	args := m.Called(ctx, domain)
	d, _ := args.Get(0).(*TenantDomain)
	return d, args.Error(1)
}

func (m *mockDomainRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*TenantDomain, error) {
	args := m.Called(ctx, tenantID)
	list, _ := args.Get(0).([]*TenantDomain)
	return list, args.Error(1)
}

func (m *mockDomainRepo) SoftDelete(ctx context.Context, id uuid.UUID, deletedBy *uuid.UUID) error {
	return m.Called(ctx, id, deletedBy).Error(0)
}

type mockCredentialRepo struct{ mock.Mock }

func (m *mockCredentialRepo) Upsert(ctx context.Context, c *ProviderCredential) error {
	return m.Called(ctx, c).Error(0)
}

func (m *mockCredentialRepo) FindByTenantAndType(ctx context.Context, tenantID uuid.UUID, providerType ProviderType) (*ProviderCredential, error) {
	args := m.Called(ctx, tenantID, providerType)
	c, _ := args.Get(0).(*ProviderCredential)
	return c, args.Error(1)
}

func (m *mockCredentialRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*ProviderCredential, error) {
	args := m.Called(ctx, tenantID)
	list, _ := args.Get(0).([]*ProviderCredential)
	return list, args.Error(1)
}

func (m *mockCredentialRepo) SoftDelete(ctx context.Context, id uuid.UUID, deletedBy *uuid.UUID) error {
	return m.Called(ctx, id, deletedBy).Error(0)
}
