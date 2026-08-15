// Package app models OAuth2/OIDC client apps and their sessions (DATA_MODEL.md §2), and
// implements the bootstrap use cases the CLI seed command needs
// (USER_JOURNEYS_ADMIN_TENANT_MANAGEMENT.md §1) — creating a tenant's default system app
// (TECHNICAL_DESIGN §3.5) and configuring its default signup role.
package app

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// SigningAlgorithm is a JWT signing algorithm an App can use (TECHNICAL_DESIGN §4). Only
// HS256 is implemented this pass — RS256/ES256/JWKS are future work.
type SigningAlgorithm string

const SigningAlgorithmHS256 SigningAlgorithm = "HS256"

// Sane defaults for a newly-created system app's token lifetimes (TECHNICAL_DESIGN §4) —
// exact values to be finalized as real usage patterns emerge.
const (
	DefaultAccessTokenTTLSeconds  = 900
	DefaultIDTokenTTLSeconds      = 900
	DefaultRefreshTokenTTLSeconds = 2592000
)

// App is an OAuth2/OIDC client under a tenant (DATA_MODEL.md §2). Every tenant gets one
// auto-provisioned, non-deletable default system app (TECHNICAL_DESIGN §3.5) for tenant-level
// operations (direct login, self-service account management) that have no third-party App to
// issue tokens against — tokens issued against it carry aud = the tenant, not a client_id.
type App struct {
	ID                    uuid.UUID
	TenantID              uuid.UUID
	Name                  string
	ClientID              string
	ClientSecretHash      *string
	RedirectURIs          []string
	LogoURL               *string
	IsSystem              bool
	SupportsOrganizations bool
	// DefaultSignupRoleID is the role auto-assigned to a new user of this app
	// (DATA_MODEL.md `apps.default_signup_role_id`) — applies uniformly to individual and
	// organization signup.
	DefaultSignupRoleID *uuid.UUID

	SigningAlgorithm SigningAlgorithm
	// SigningKeyConfig holds the raw HS256 secret, unencrypted, for now — no encryption port
	// exists yet (TECHNICAL_DESIGN §8); same column/shape the future encrypted form will use.
	SigningKeyConfig []byte

	AccessTokenTTLSeconds  int
	IDTokenTTLSeconds      int
	RefreshTokenTTLSeconds int

	CreatedAt time.Time
	UpdatedAt time.Time
	CreatedBy *uuid.UUID
	UpdatedBy *uuid.UUID

	DeletedAt *time.Time
	DeletedBy *uuid.UUID
}

// IsDeleted reports whether a has been soft-deleted.
func (a *App) IsDeleted() bool {
	return a.DeletedAt != nil
}

// Repository persists and retrieves Apps.
type Repository interface {
	// CreateSystem persists a's system app row (is_system=true) — the CLI seed's bootstrap
	// path, no actor.Actor.
	CreateSystem(ctx context.Context, a *App) error
	// FindSystemAppByTenant is a pre-authentication, unscoped lookup of tenantID's default
	// system app — used by identity.Service.Signup, which runs before an actor.Actor can
	// exist. Returns nil, nil if the tenant has no system app (a child tenant created via
	// tenant.Service.CreateTenant doesn't provision one this pass — root-tenant-only for
	// now, AUTHORIZATION_MODEL.md's scope for this iteration).
	FindSystemAppByTenant(ctx context.Context, tenantID uuid.UUID) (*App, error)
	// SetDefaultSignupRole sets appID's default_signup_role_id. No actor.Actor — CLI seed
	// bootstrap only, for now.
	SetDefaultSignupRole(ctx context.Context, appID, roleID uuid.UUID) error
}
