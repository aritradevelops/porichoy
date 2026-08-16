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

func TestService_CreateRootTenant(t *testing.T) {
	tenants := &mockTenantRepo{}
	tenants.On("CreateRoot", mock.Anything, mock.AnythingOfType("*tenant.Tenant")).Return(nil)
	svc := NewService(tenants, &mockDomainRepo{}, &mockCredentialRepo{})

	got, err := svc.CreateRootTenant(context.Background(), "Root")

	require.NoError(t, err)
	assert.Equal(t, "Root", got.Name)
	assert.True(t, got.IsRoot())
	assert.Equal(t, LoginLayoutCentered, got.LoginLayout)
	assert.Equal(t, []LoginMethod{LoginMethodEmailPassword}, got.EnabledLoginMethods)
	assert.Nil(t, got.CreatedBy)
	assert.Nil(t, got.UpdatedBy)
	tenants.AssertExpectations(t)
}

func TestService_CreateTenant(t *testing.T) {
	tenants := &mockTenantRepo{}
	tenants.On("Create", mock.Anything, mock.Anything, mock.AnythingOfType("*tenant.Tenant")).Return(nil)
	svc := NewService(tenants, &mockDomainRepo{}, &mockCredentialRepo{})

	act := newActor(t)
	parentID := newUUID(t)
	got, err := svc.CreateTenant(context.Background(), act, "Brand A", &parentID)

	require.NoError(t, err)
	assert.Equal(t, "Brand A", got.Name)
	assert.False(t, got.IsRoot())
	assert.Equal(t, &parentID, got.ParentID)
	assert.Equal(t, LoginLayoutCentered, got.LoginLayout)
	assert.Equal(t, &act.PrincipalID, got.CreatedBy)
	assert.Equal(t, &act.PrincipalID, got.UpdatedBy)
	tenants.AssertExpectations(t)
}

func TestService_CreateTenant_DefaultsParentToActorTenant(t *testing.T) {
	tenants := &mockTenantRepo{}
	tenants.On("Create", mock.Anything, mock.Anything, mock.AnythingOfType("*tenant.Tenant")).Return(nil)
	svc := NewService(tenants, &mockDomainRepo{}, &mockCredentialRepo{})

	act := newActor(t)
	got, err := svc.CreateTenant(context.Background(), act, "Brand A", nil)

	require.NoError(t, err)
	require.NotNil(t, got.ParentID)
	assert.Equal(t, act.TenantID, *got.ParentID)
}

func TestService_CreateTenant_ParentNotFound(t *testing.T) {
	tenants := &mockTenantRepo{}
	tenants.On("Create", mock.Anything, mock.Anything, mock.AnythingOfType("*tenant.Tenant")).
		Return(ErrTenantNotFound)
	svc := NewService(tenants, &mockDomainRepo{}, &mockCredentialRepo{})

	parentID := newUUID(t)
	_, err := svc.CreateTenant(context.Background(), newActor(t), "Brand A", &parentID)

	require.Error(t, err)
	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "tenant.not_found", appErr.Key)
}

func TestService_RegisterRootDomain_OK(t *testing.T) {
	domains := &mockDomainRepo{}
	domains.On("FindByDomain", mock.Anything, "root.example.com").Return(nil, nil)
	domains.On("CreateRoot", mock.Anything, mock.AnythingOfType("*tenant.TenantDomain")).Return(nil)
	svc := NewService(&mockTenantRepo{}, domains, &mockCredentialRepo{})

	tenantID := uuid.New()
	got, err := svc.RegisterRootDomain(context.Background(), tenantID, "root.example.com")

	require.NoError(t, err)
	assert.Equal(t, tenantID, got.TenantID)
	assert.Equal(t, "root.example.com", got.Domain)
	assert.Nil(t, got.CreatedBy)
	domains.AssertExpectations(t)
}

func TestService_RegisterRootDomain_AlreadyRegistered(t *testing.T) {
	domains := &mockDomainRepo{}
	domains.On("FindByDomain", mock.Anything, "root.example.com").
		Return(&TenantDomain{ID: uuid.New()}, nil)
	svc := NewService(&mockTenantRepo{}, domains, &mockCredentialRepo{})

	_, err := svc.RegisterRootDomain(context.Background(), uuid.New(), "root.example.com")

	require.Error(t, err)
	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "tenant.domain_already_registered", appErr.Key)
	domains.AssertNotCalled(t, "CreateRoot", mock.Anything, mock.Anything)
}

func TestService_RegisterDomain_AlreadyRegistered(t *testing.T) {
	domains := &mockDomainRepo{}
	domains.On("FindByDomain", mock.Anything, "brand-a.example.com").
		Return(&TenantDomain{ID: uuid.New()}, nil)
	svc := NewService(&mockTenantRepo{}, domains, &mockCredentialRepo{})

	_, err := svc.RegisterDomain(context.Background(), newActor(t), uuid.New(), "brand-a.example.com")

	require.Error(t, err)
	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "tenant.domain_already_registered", appErr.Key)
	domains.AssertNotCalled(t, "Create", mock.Anything, mock.Anything, mock.Anything)
}

func TestService_RegisterDomain_OK(t *testing.T) {
	domains := &mockDomainRepo{}
	domains.On("FindByDomain", mock.Anything, "brand-a.example.com").Return(nil, nil)
	domains.On("Create", mock.Anything, mock.Anything, mock.AnythingOfType("*tenant.TenantDomain")).Return(nil)
	svc := NewService(&mockTenantRepo{}, domains, &mockCredentialRepo{})

	act := newActor(t)
	tenantID := uuid.New()
	got, err := svc.RegisterDomain(context.Background(), act, tenantID, "brand-a.example.com")

	require.NoError(t, err)
	assert.Equal(t, tenantID, got.TenantID)
	assert.Equal(t, "brand-a.example.com", got.Domain)
	assert.Equal(t, &act.PrincipalID, got.CreatedBy)
}

func TestService_RegisterDomain_TenantNotFound(t *testing.T) {
	domains := &mockDomainRepo{}
	domains.On("FindByDomain", mock.Anything, "brand-a.example.com").Return(nil, nil)
	domains.On("Create", mock.Anything, mock.Anything, mock.AnythingOfType("*tenant.TenantDomain")).
		Return(ErrTenantNotFound)
	svc := NewService(&mockTenantRepo{}, domains, &mockCredentialRepo{})

	_, err := svc.RegisterDomain(context.Background(), newActor(t), uuid.New(), "brand-a.example.com")

	require.Error(t, err)
	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "tenant.not_found", appErr.Key)
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
	tenants.On("FindByID", mock.Anything, tenantID).
		Return(&Tenant{ID: tenantID, Name: "Brand A"}, nil)
	svc := NewService(tenants, domains, &mockCredentialRepo{})

	got, err := svc.ResolveTenantByDomain(context.Background(), "brand-a.example.com")

	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "Brand A", got.Name)
}

func TestService_ConfigureTenant_NotFound(t *testing.T) {
	tenants := &mockTenantRepo{}
	tenants.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
	svc := NewService(tenants, &mockDomainRepo{}, &mockCredentialRepo{})

	_, err := svc.ConfigureTenant(context.Background(), newActor(t), uuid.New(), Config{})

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
	tenants.On("GetByID", mock.Anything, mock.Anything, tenantID).Return(existing, nil)
	tenants.On("Update", mock.Anything, mock.AnythingOfType("*tenant.Tenant")).Return(nil)
	svc := NewService(tenants, &mockDomainRepo{}, &mockCredentialRepo{})

	act := newActor(t)
	newLogo := "https://new-logo"
	mfaOn := true
	got, err := svc.ConfigureTenant(context.Background(), act, tenantID, Config{
		LogoURL:     &newLogo,
		MFARequired: &mfaOn,
	})

	require.NoError(t, err)
	assert.Equal(t, "https://new-logo", got.LogoURL)
	assert.True(t, got.MFARequired)
	assert.Equal(t, &act.PrincipalID, got.UpdatedBy)
	// Untouched fields stay as they were.
	assert.Equal(t, "Brand A", got.Name)
	assert.Equal(t, LoginLayoutCentered, got.LoginLayout)
}

func TestService_SetProviderCredential_CreatesWhenAbsent(t *testing.T) {
	creds := &mockCredentialRepo{}
	creds.On("FindByTenantAndType", mock.Anything, mock.Anything, mock.Anything, ProviderTypeGoogle).
		Return(nil, nil)
	creds.On("Upsert", mock.Anything, mock.Anything, mock.AnythingOfType("*tenant.ProviderCredential")).Return(nil)
	svc := NewService(&mockTenantRepo{}, &mockDomainRepo{}, creds)

	act := newActor(t)
	tenantID := uuid.New()
	got, err := svc.SetProviderCredential(context.Background(), act, tenantID, ProviderTypeGoogle, []byte("ciphertext"))

	require.NoError(t, err)
	assert.Equal(t, tenantID, got.TenantID)
	assert.Equal(t, ProviderTypeGoogle, got.ProviderType)
	assert.Equal(t, &act.PrincipalID, got.CreatedBy)
}

func TestService_SetProviderCredential_TenantNotFound(t *testing.T) {
	creds := &mockCredentialRepo{}
	creds.On("FindByTenantAndType", mock.Anything, mock.Anything, mock.Anything, ProviderTypeGoogle).
		Return(nil, nil)
	creds.On("Upsert", mock.Anything, mock.Anything, mock.AnythingOfType("*tenant.ProviderCredential")).
		Return(ErrTenantNotFound)
	svc := NewService(&mockTenantRepo{}, &mockDomainRepo{}, creds)

	_, err := svc.SetProviderCredential(context.Background(), newActor(t), uuid.New(), ProviderTypeGoogle, []byte("ciphertext"))

	require.Error(t, err)
	var appErr *apperror.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "tenant.not_found", appErr.Key)
}

func TestService_SetProviderCredential_ReplacesWhenPresent(t *testing.T) {
	existing := &ProviderCredential{
		ID:              uuid.New(),
		ProviderType:    ProviderTypeGoogle,
		ConfigEncrypted: []byte("old"),
	}
	creds := &mockCredentialRepo{}
	creds.On("FindByTenantAndType", mock.Anything, mock.Anything, mock.Anything, ProviderTypeGoogle).
		Return(existing, nil)
	creds.On("Upsert", mock.Anything, mock.Anything, existing).Return(nil)
	svc := NewService(&mockTenantRepo{}, &mockDomainRepo{}, creds)

	got, err := svc.SetProviderCredential(context.Background(), newActor(t), uuid.New(), ProviderTypeGoogle, []byte("new"))

	require.NoError(t, err)
	assert.Equal(t, existing.ID, got.ID)
	assert.Equal(t, []byte("new"), got.ConfigEncrypted)
}
