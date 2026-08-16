package tenant

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/aritradevelops/porichoy/server/internal/actor"
	"github.com/aritradevelops/porichoy/server/internal/apperror"
	"github.com/google/uuid"
)

// Service implements the tenant use cases.
type Service struct {
	tenants Repository
	domains DomainRepository
	creds   ProviderCredentialRepository
}

// NewService wires a Service from its domain repository dependencies.
func NewService(
	tenants Repository,
	domains DomainRepository,
	creds ProviderCredentialRepository,
) *Service {
	return &Service{tenants: tenants, domains: domains, creds: creds}
}

// CreateRootTenant creates the root tenant. This is the CLI seed command's bootstrap step
// (USER_JOURNEYS_ADMIN_TENANT_MANAGEMENT.md §1) and only ever runs once, before any
// principal or actor.Actor exists at all — so, unlike CreateTenant, it takes no actor.Actor.
// ParentID, CreatedBy, and UpdatedBy are all nil (system-generated, DATA_MODEL.md §0).
// EnabledLoginMethods defaults to email+password — the only method identity.Service actually
// implements this pass — so the seeded root superadmin can log in immediately afterward;
// without this, ConfigureTenant (the only other way to set it) is itself an authenticated
// route nothing could reach yet.
func (s *Service) CreateRootTenant(ctx context.Context, name string) (*Tenant, error) {
	now := time.Now()
	t := &Tenant{
		ID:                  uuid.New(),
		ParentID:            nil,
		Name:                name,
		LoginLayout:         LoginLayoutCentered, // sane default until configured via §3
		EnabledLoginMethods: []LoginMethod{LoginMethodEmailPassword},
		CreatedAt:           now,
		UpdatedAt:           now,
	}
	if err := s.tenants.CreateRoot(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

// CreateTenant creates a brand/child tenant under an existing one
// (USER_JOURNEYS_ADMIN_TENANT_MANAGEMENT.md §2) — an authorized operation, unlike root-tenant
// bootstrap (CreateRootTenant). If parentID is nil, it defaults to the caller's own tenant
// (MCP_TOOLS.md tenant_create).
func (s *Service) CreateTenant(ctx context.Context, act actor.Actor, name string, parentID *uuid.UUID) (*Tenant, error) {
	if parentID == nil {
		parentID = &act.TenantID
	}
	now := time.Now()
	t := &Tenant{
		ID:          uuid.New(),
		ParentID:    parentID,
		Name:        name,
		LoginLayout: LoginLayoutCentered, // sane default until configured via §3
		CreatedAt:   now,
		UpdatedAt:   now,
		CreatedBy:   &act.PrincipalID,
		UpdatedBy:   &act.PrincipalID,
	}
	if err := s.tenants.Create(ctx, act, t); err != nil {
		if errors.Is(err, ErrTenantNotFound) {
			return nil, apperror.New("tenant.not_found", http.StatusNotFound)
		}
		return nil, err
	}
	return t, nil
}

// GetTenant fetches a tenant by ID.
func (s *Service) GetTenant(ctx context.Context, act actor.Actor, id uuid.UUID) (*Tenant, error) {
	return s.tenants.GetByID(ctx, act, id)
}

// Pagination bounds for ListTenants — kept small since the REST layer has no other
// pagination endpoint yet to share a convention with (AUTHORIZATION_MODEL.md has no
// documented default; this is the first list-with-many-rows route in this bounded context).
const (
	defaultTenantListPageSize = 20
	maxTenantListPageSize     = 100
)

// ListTenants returns one page of every tenant act is authorized to see — every tenant for
// root scope, act's own tenant plus its subtree for tenant scope (Repository.List).
// Normalizes params here (Page clamped to >=1; PageSize clamped to
// [1, maxTenantListPageSize], defaulting to defaultTenantListPageSize when <= 0) so neither
// the REST layer nor the repository has to.
func (s *Service) ListTenants(ctx context.Context, act actor.Actor, params ListParams) (ListResult, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = defaultTenantListPageSize
	}
	if params.PageSize > maxTenantListPageSize {
		params.PageSize = maxTenantListPageSize
	}
	return s.tenants.List(ctx, act, params)
}

// ResolveTenantByDomain resolves which tenant owns the given origin — the tenant-resolution
// lookup every incoming request goes through (TECHNICAL_DESIGN §3.3). Returns nil, nil if
// no tenant has registered this domain, rather than an error — "not found" is an expected
// outcome here, not a failure.
func (s *Service) ResolveTenantByDomain(ctx context.Context, domain string) (*Tenant, error) {
	d, err := s.domains.FindByDomain(ctx, domain)
	if err != nil {
		return nil, err
	}
	if d == nil {
		return nil, nil
	}
	return s.tenants.FindByID(ctx, d.TenantID)
}

// RegisterRootDomain registers domain for tenantID with no actor.Actor — the CLI seed's
// bootstrap path (USER_JOURNEYS_ADMIN_TENANT_MANAGEMENT.md §1), registering the root
// tenant's first domain in the same transaction that creates the tenant itself. Without
// this, a fresh instance would have no domain any Host header could ever resolve to, and
// RegisterDomain itself is only reachable behind an authenticated route — a deadlock only
// bootstrap-time code can break.
func (s *Service) RegisterRootDomain(ctx context.Context, tenantID uuid.UUID, domain string) (*TenantDomain, error) {
	existing, err := s.domains.FindByDomain(ctx, domain)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, apperror.New("tenant.domain_already_registered", http.StatusConflict)
	}

	d := &TenantDomain{
		ID:        uuid.New(),
		TenantID:  tenantID,
		Domain:    domain,
		CreatedAt: time.Now(),
	}
	if err := s.domains.CreateRoot(ctx, d); err != nil {
		return nil, err
	}
	return d, nil
}

// RegisterDomain registers a new allowed origin for tenantID. Domains are unique across
// the whole instance, not just within one tenant (DATA_MODEL.md `domain_registry`) — this
// check exists to return a clean DomainError instead of a raw constraint violation; the
// database-level unique constraint remains the actual invariant enforcer.
func (s *Service) RegisterDomain(ctx context.Context, act actor.Actor, tenantID uuid.UUID, domain string) (*TenantDomain, error) {
	existing, err := s.domains.FindByDomain(ctx, domain)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, apperror.New("tenant.domain_already_registered", http.StatusConflict)
	}

	d := &TenantDomain{
		ID:        uuid.New(),
		TenantID:  tenantID,
		Domain:    domain,
		CreatedAt: time.Now(),
		CreatedBy: &act.PrincipalID,
	}
	if err := s.domains.Create(ctx, act, d); err != nil {
		if errors.Is(err, ErrTenantNotFound) {
			return nil, apperror.New("tenant.not_found", http.StatusNotFound)
		}
		return nil, err
	}
	return d, nil
}

// Config is a partial update to a tenant's settings (USER_JOURNEYS_ADMIN_TENANT_MANAGEMENT.md
// §3) — nil fields are left unchanged.
type Config struct {
	LogoURL             *string
	BrandImageURL       *string
	LoginLayout         *LoginLayout
	MFARequired         *bool
	EnabledLoginMethods []LoginMethod // nil means unchanged; non-nil replaces wholesale
	AuditRetentionDays  *int
}

// ConfigureTenant applies a partial update to tenantID's settings.
func (s *Service) ConfigureTenant(ctx context.Context, act actor.Actor, tenantID uuid.UUID, cfg Config) (*Tenant, error) {
	t, err := s.tenants.GetByID(ctx, act, tenantID)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, apperror.New("tenant.not_found", http.StatusNotFound)
	}

	if cfg.LogoURL != nil {
		t.LogoURL = *cfg.LogoURL
	}
	if cfg.BrandImageURL != nil {
		t.BrandImageURL = cfg.BrandImageURL
	}
	if cfg.LoginLayout != nil {
		t.LoginLayout = *cfg.LoginLayout
	}
	if cfg.MFARequired != nil {
		t.MFARequired = *cfg.MFARequired
	}
	if cfg.EnabledLoginMethods != nil {
		t.EnabledLoginMethods = cfg.EnabledLoginMethods
	}
	if cfg.AuditRetentionDays != nil {
		t.AuditRetentionDays = *cfg.AuditRetentionDays
	}
	t.UpdatedAt = time.Now()
	t.UpdatedBy = &act.PrincipalID

	if err := s.tenants.Update(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

// SetProviderCredential creates or replaces tenantID's credential for the given provider
// (MCP_TOOLS.md tenant_set_provider_credential).
//
// configEncrypted is expected already encrypted — this method doesn't perform encryption
// itself. TECHNICAL_DESIGN §8 calls for an application-level encryption port that this
// Service should eventually own calling (encrypt-before-persist belongs in orchestration,
// not scattered across callers); that port doesn't exist yet, so for now the caller is
// responsible. Flagging rather than silently deferring the decision.
func (s *Service) SetProviderCredential(
	ctx context.Context,
	act actor.Actor,
	tenantID uuid.UUID,
	providerType ProviderType,
	configEncrypted []byte,
) (*ProviderCredential, error) {
	existing, err := s.creds.FindByTenantAndType(ctx, act, tenantID, providerType)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	if existing != nil {
		existing.ConfigEncrypted = configEncrypted
		existing.UpdatedAt = now
		existing.UpdatedBy = &act.PrincipalID
		if err := s.creds.Upsert(ctx, act, existing); err != nil {
			if errors.Is(err, ErrTenantNotFound) {
				return nil, apperror.New("tenant.not_found", http.StatusNotFound)
			}
			return nil, err
		}
		return existing, nil
	}

	c := &ProviderCredential{
		ID:              uuid.New(),
		TenantID:        tenantID,
		ProviderType:    providerType,
		ConfigEncrypted: configEncrypted,
		CreatedAt:       now,
		UpdatedAt:       now,
		CreatedBy:       &act.PrincipalID,
		UpdatedBy:       &act.PrincipalID,
	}
	if err := s.creds.Upsert(ctx, act, c); err != nil {
		if errors.Is(err, ErrTenantNotFound) {
			return nil, apperror.New("tenant.not_found", http.StatusNotFound)
		}
		return nil, err
	}
	return c, nil
}
