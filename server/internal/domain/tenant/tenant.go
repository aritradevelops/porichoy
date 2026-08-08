// Package tenant models the tenant hierarchy and its per-tenant configuration
// (DATA_MODEL.md §1: Tenant, DomainRegistry, TenantProviderCredential).
package tenant

import (
	"context"
	"time"

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
type Repository interface {
	Create(ctx context.Context, t *Tenant) error
	GetByID(ctx context.Context, id uuid.UUID) (*Tenant, error)
	Update(ctx context.Context, t *Tenant) error
	SoftDelete(ctx context.Context, id uuid.UUID, deletedBy *uuid.UUID) error
	// ListChildren returns the direct children of parentID — one level of the
	// hierarchy, not the full subtree.
	ListChildren(ctx context.Context, parentID uuid.UUID) ([]*Tenant, error)
}
