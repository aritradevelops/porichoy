// Package tenant models the tenant hierarchy and its per-tenant configuration
// (DATA_MODEL.md §1: Tenant, DomainRegistry, TenantProviderCredential), and implements
// the tenant-management use cases (Service, in service.go) that the REST and MCP
// adapters call into (CODING_STANDARDS.md §3) — creating tenants, registering domains,
// and configuring tenant-level settings (USER_JOURNEYS_ADMIN_TENANT_MANAGEMENT.md §2–§3,
// MCP_TOOLS.md §4).
package tenant

import (
	"context"
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

// Repository persists and retrieves Tenants.
//
// Create and Update take no actor.Actor — the *Tenant being persisted already carries
// CreatedBy/UpdatedBy (set by the Service), so it would be redundant. SoftDelete and
// ListChildren are authorized-only operations with no entity to carry that information, so
// they take one (CODING_STANDARDS.md §4).
type Repository interface {
	Create(ctx context.Context, t *Tenant) error
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
}
