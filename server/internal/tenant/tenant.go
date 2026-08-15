// Package tenant models the tenant hierarchy and its per-tenant configuration
// (DATA_MODEL.md §1: Tenant, DomainRegistry, TenantProviderCredential), and implements
// the tenant-management use cases (Service, in service.go) that the REST and MCP
// adapters call into (CODING_STANDARDS.md §3) — creating tenants, registering domains,
// and configuring tenant-level settings (USER_JOURNEYS_ADMIN_TENANT_MANAGEMENT.md §2–§3,
// MCP_TOOLS.md §4).
package tenant

import (
	"context"
	"errors"
	"time"

	"github.com/aritradevelops/porichoy/server/internal/actor"
	"github.com/google/uuid"
)

// LoginLayout is the visual layout of a tenant's login screen (TECHNICAL_DESIGN §1).
type LoginLayout string

const (
	LoginLayoutCentered LoginLayout = "centered"
	LoginLayoutSplit    LoginLayout = "split"
)

// LoginMethod is a login mechanism a tenant can enable (PRD §5.1).
type LoginMethod string

const (
	LoginMethodEmailPassword LoginMethod = "email_password"
	LoginMethodEmailOTP      LoginMethod = "email_otp"
	LoginMethodPhoneOTP      LoginMethod = "phone_otp"
	LoginMethodWebAuthn      LoginMethod = "webauthn"
	LoginMethodGoogle        LoginMethod = "google"
	LoginMethodApple         LoginMethod = "apple"
)

// Tenant is an authentication realm (AUTHORIZATION_MODEL.md), arranged in an
// arbitrary-depth hierarchy via ParentID. The root tenant has a nil ParentID.
//
// Branding and auth-policy settings are folded directly in rather than split into
// their own tables — see DATA_MODEL.md `tenant` for why. Token TTLs are deliberately
// not here; those live on App (DATA_MODEL.md `app`).
type Tenant struct {
	ID       uuid.UUID
	ParentID *uuid.UUID
	Name     string
	// Ancestors is the tenant's full chain of ancestor IDs, precomputed at creation time as
	// the parent's own Ancestors plus the parent's ID (no recursive query). This is what
	// makes descendant-scope authorization (AUTHORIZATION_MODEL.md §4) a single containment
	// check — "is act.TenantID in t.Ancestors" — instead of a recursive tree walk. Empty for
	// the root tenant.
	Ancestors []uuid.UUID

	LogoURL             string
	BrandImageURL       *string // only used when LoginLayout is LoginLayoutSplit
	LoginLayout         LoginLayout
	MFARequired         bool
	EnabledLoginMethods []LoginMethod
	AuditRetentionDays  int

	CreatedAt time.Time
	UpdatedAt time.Time
	// CreatedBy/UpdatedBy/DeletedBy are principals (a user or an api_credential,
	// undifferentiated — DATA_MODEL.md §0). Nil for the root tenant, which is
	// system-bootstrapped rather than created by anyone.
	CreatedBy *uuid.UUID
	UpdatedBy *uuid.UUID

	DeletedAt *time.Time
	DeletedBy *uuid.UUID
}

// IsRoot reports whether t is the root tenant.
func (t *Tenant) IsRoot() bool {
	return t.ParentID == nil
}

// IsDeleted reports whether t has been soft-deleted.
func (t *Tenant) IsDeleted() bool {
	return t.DeletedAt != nil
}

// ErrTenantNotFound is returned whenever a referenced tenant either doesn't exist or exists
// but falls outside act's authorized subtree (AUTHORIZATION_MODEL.md §4) — both cases return
// the same sentinel, so a caller can't distinguish "doesn't exist" from "exists but isn't
// yours", consistent with this package's "not found, not forbidden" convention. Used by
// Repository.Create (the referenced tenant is t.ParentID) and, in the same way, by
// DomainRepository.Create and ProviderCredentialRepository.Upsert (the referenced tenant is
// the one the domain/credential is being created for).
var ErrTenantNotFound = errors.New("tenant: not found")

// Repository persists and retrieves Tenants.
//
// Update takes no actor.Actor — the *Tenant being persisted already carries UpdatedBy (set
// by the Service), so it would be redundant, and by the time Update is called the entity was
// already fetched (and thus authorized) via GetByID. SoftDelete and ListChildren are
// authorized-only operations with no entity to carry that information, so they take one
// (CODING_STANDARDS.md §4). Create also takes one now: unlike Update, there's no existing row
// to have already been authorized against, and Create needs act to both authorize against
// t.ParentID and compute t.Ancestors from the parent (see Create's own doc).
type Repository interface {
	// CreateRoot persists the root tenant. Pre-authentication bootstrap only — no
	// actor.Actor exists yet at this point (CreateRootTenant), t.ParentID is always nil, and
	// t.Ancestors is always empty.
	CreateRoot(ctx context.Context, t *Tenant) error
	// Create persists a child tenant under t.ParentID (always non-nil — Service.CreateTenant
	// defaults it to act.TenantID before calling this). Checks act is authorized to create
	// under that parent (AUTHORIZATION_MODEL.md §4: root may create under any parent; tenant
	// scope only under itself or a descendant), returning ErrTenantNotFound otherwise, and
	// computes t.Ancestors from the parent's own Ancestors plus the parent's ID.
	Create(ctx context.Context, act actor.Actor, t *Tenant) error
	// FindByID is a pre-authentication, unscoped lookup by primary key — used by
	// Service.ResolveTenantByDomain (TECHNICAL_DESIGN.md §3.3), which runs before an
	// actor.Actor can exist at all. Authorized code should use GetByID instead.
	FindByID(ctx context.Context, id uuid.UUID) (*Tenant, error)
	// GetByID is the authorized, scope-filtered lookup by primary key.
	GetByID(ctx context.Context, act actor.Actor, id uuid.UUID) (*Tenant, error)
	Update(ctx context.Context, t *Tenant) error
	SoftDelete(ctx context.Context, act actor.Actor, id uuid.UUID) error
	// ListChildren returns the direct children of parentID — one level of the
	// hierarchy, not the full subtree.
	ListChildren(ctx context.Context, act actor.Actor, parentID uuid.UUID) ([]*Tenant, error)
	// List returns one page of every tenant act is authorized to see, flat rather than
	// scoped to one parent's direct children: for root scope, every tenant in the system;
	// for tenant scope, act's own tenant plus its full subtree (AUTHORIZATION_MODEL.md §4).
	// This backs the root-admin "all tenants" listing
	// (USER_JOURNEYS_ADMIN_TENANT_MANAGEMENT.md §2) — distinct from ListChildren's one-level
	// drill-down. params is expected already normalized (Service.ListTenants does this) —
	// Page >= 1, PageSize within bounds.
	List(ctx context.Context, act actor.Actor, params ListParams) (ListResult, error)
}

// ListParams paginates Repository.List. Page is 1-indexed.
type ListParams struct {
	Page     int
	PageSize int
}

// ListResult is Repository.List's paginated result: Items is this page's rows, Total is the
// count of every row act is authorized to see across all pages (for computing page count:
// ceil(Total / PageSize)), and Page/PageSize echo back the params actually used to produce
// this result.
type ListResult struct {
	Items    []*Tenant
	Total    int
	Page     int
	PageSize int
}
