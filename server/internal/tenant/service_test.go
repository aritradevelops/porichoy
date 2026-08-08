package tenant

import (
	"context"
	"testing"

	"github.com/aritradevelops/porichoy/server/internal/apperror"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_CreateTenant(t *testing.T) {
	tenants := &mockTenantRepo{}
	tenants.On("Create", mock.Anything, mock.AnythingOfType("*tenant.Tenant")).Return(nil)
	svc := NewService(tenants, &mockDomainRepo{}, &mockCredentialRepo{})

	actor := uuid.New()
	got, err := svc.CreateTenant(context.Background(), "Brand A", nil, &actor)

	require.NoError(t, err)
	assert.Equal(t, "Brand A", got.Name)
	assert.True(t, got.IsRoot())
	assert.Equal(t, LoginLayoutCentered, got.LoginLayout)
	assert.Equal(t, &actor, got.CreatedBy)
	tenants.AssertExpectations(t)
}

func TestService_RegisterDomain_AlreadyRegistered(t *testing.T) {
	domains := &mockDomainRepo{}
	domains.On("FindByDomain", mock.Anything, "brand-a.example.com").
		Return(&TenantDomain{ID: uuid.New()}, nil)
	svc := NewService(&mockTenantRepo{}, domains, &mockCredentialRepo{})

	_, err := svc.RegisterDomain(context.Background(), uuid.New(), "brand-a.example.com", nil)

	require.Error(t, err)
	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "tenant.domain_already_registered", appErr.Key)
	domains.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestService_RegisterDomain_OK(t *testing.T) {
	domains := &mockDomainRepo{}
	domains.On("FindByDomain", mock.Anything, "brand-a.example.com").Return(nil, nil)
	domains.On("Create", mock.Anything, mock.AnythingOfType("*tenant.TenantDomain")).Return(nil)
	svc := NewService(&mockTenantRepo{}, domains, &mockCredentialRepo{})

	tenantID := uuid.New()
	got, err := svc.RegisterDomain(context.Background(), tenantID, "brand-a.example.com", nil)

	require.NoError(t, err)
	assert.Equal(t, tenantID, got.TenantID)
	assert.Equal(t, "brand-a.example.com", got.Domain)
}

func TestService_ResolveTenantByDomain_NotFound(t *testing.T) {
	domains := &mockDomainRepo{}
	domains.On("FindByDomain", mock.Anything, "unknown.example.com").Return(nil, nil)
	svc := NewService(&mockTenantRepo{}, domains, &mockCredentialRepo{})

	got, err := svc.ResolveTenantByDomain(context.Background(), "unknown.example.com")

	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestService_ResolveTenantByDomain_Found(t *testing.T) {
	tenantID := uuid.New()
	domains := &mockDomainRepo{}
	domains.On("FindByDomain", mock.Anything, "brand-a.example.com").
		Return(&TenantDomain{TenantID: tenantID}, nil)
	tenants := &mockTenantRepo{}
	tenants.On("GetByID", mock.Anything, tenantID).
		Return(&Tenant{ID: tenantID, Name: "Brand A"}, nil)
	svc := NewService(tenants, domains, &mockCredentialRepo{})

	got, err := svc.ResolveTenantByDomain(context.Background(), "brand-a.example.com")

	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "Brand A", got.Name)
}

func TestService_ConfigureTenant_NotFound(t *testing.T) {
	tenants := &mockTenantRepo{}
	tenants.On("GetByID", mock.Anything, mock.Anything).Return(nil, nil)
	svc := NewService(tenants, &mockDomainRepo{}, &mockCredentialRepo{})

	_, err := svc.ConfigureTenant(context.Background(), uuid.New(), Config{}, nil)

	require.Error(t, err)
	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "tenant.not_found", appErr.Key)
}

func TestService_ConfigureTenant_PartialUpdate(t *testing.T) {
	tenantID := uuid.New()
	existing := &Tenant{
		ID:          tenantID,
		Name:        "Brand A",
		LogoURL:     "https://old-logo",
		LoginLayout: LoginLayoutCentered,
	}
	tenants := &mockTenantRepo{}
	tenants.On("GetByID", mock.Anything, tenantID).Return(existing, nil)
	tenants.On("Update", mock.Anything, mock.AnythingOfType("*tenant.Tenant")).Return(nil)
	svc := NewService(tenants, &mockDomainRepo{}, &mockCredentialRepo{})

	newLogo := "https://new-logo"
	mfaOn := true
	got, err := svc.ConfigureTenant(context.Background(), tenantID, Config{
		LogoURL:     &newLogo,
		MFARequired: &mfaOn,
	}, nil)

	require.NoError(t, err)
	assert.Equal(t, "https://new-logo", got.LogoURL)
	assert.True(t, got.MFARequired)
	// Untouched fields stay as they were.
	assert.Equal(t, "Brand A", got.Name)
	assert.Equal(t, LoginLayoutCentered, got.LoginLayout)
}

func TestService_SetProviderCredential_CreatesWhenAbsent(t *testing.T) {
	creds := &mockCredentialRepo{}
	creds.On("FindByTenantAndType", mock.Anything, mock.Anything, ProviderTypeGoogle).
		Return(nil, nil)
	creds.On("Upsert", mock.Anything, mock.AnythingOfType("*tenant.ProviderCredential")).Return(nil)
	svc := NewService(&mockTenantRepo{}, &mockDomainRepo{}, creds)

	tenantID := uuid.New()
	got, err := svc.SetProviderCredential(context.Background(), tenantID, ProviderTypeGoogle, []byte("ciphertext"), nil)

	require.NoError(t, err)
	assert.Equal(t, tenantID, got.TenantID)
	assert.Equal(t, ProviderTypeGoogle, got.ProviderType)
}

func TestService_SetProviderCredential_ReplacesWhenPresent(t *testing.T) {
	existing := &ProviderCredential{
		ID:              uuid.New(),
		ProviderType:    ProviderTypeGoogle,
		ConfigEncrypted: []byte("old"),
	}
	creds := &mockCredentialRepo{}
	creds.On("FindByTenantAndType", mock.Anything, mock.Anything, ProviderTypeGoogle).
		Return(existing, nil)
	creds.On("Upsert", mock.Anything, existing).Return(nil)
	svc := NewService(&mockTenantRepo{}, &mockDomainRepo{}, creds)

	got, err := svc.SetProviderCredential(context.Background(), uuid.New(), ProviderTypeGoogle, []byte("new"), nil)

	require.NoError(t, err)
	assert.Equal(t, existing.ID, got.ID)
	assert.Equal(t, []byte("new"), got.ConfigEncrypted)
}
