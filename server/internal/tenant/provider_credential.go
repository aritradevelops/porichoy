package tenant

import (
	"context"
	"time"

	"github.com/aritradevelops/porichoy/server/internal/actor"
	"github.com/google/uuid"
)

// ProviderType identifies a pluggable per-tenant external provider: social login
// apps (PRD §5.1) or OTP delivery (TECHNICAL_DESIGN §5).
type ProviderType string

const (
	ProviderTypeGoogle   ProviderType = "google"
	ProviderTypeApple    ProviderType = "apple"
	ProviderTypeOTPEmail ProviderType = "otp_email"
	ProviderTypeOTPSMS   ProviderType = "otp_sms"
)

// ProviderCredential holds one tenant's credentials for one external provider —
// at most one row per (tenant, provider type). ConfigEncrypted is opaque
// ciphertext; encryption/decryption happens at the adapter layer via the
// application-level encryption port (TECHNICAL_DESIGN §8), not in this package.
type ProviderCredential struct {
	ID              uuid.UUID
	TenantID        uuid.UUID
	ProviderType    ProviderType
	ConfigEncrypted []byte

	CreatedAt time.Time
	UpdatedAt time.Time
	CreatedBy *uuid.UUID
	UpdatedBy *uuid.UUID

	DeletedAt *time.Time
	DeletedBy *uuid.UUID
}

// IsDeleted reports whether c has been soft-deleted.
func (c *ProviderCredential) IsDeleted() bool {
	return c.DeletedAt != nil
}

// ProviderCredentialRepository persists and retrieves ProviderCredentials.
type ProviderCredentialRepository interface {
	// Upsert creates or replaces the credential for c's (TenantID, ProviderType) pair.
	Upsert(ctx context.Context, c *ProviderCredential) error
	FindByTenantAndType(ctx context.Context, act actor.Actor, tenantID uuid.UUID, providerType ProviderType) (*ProviderCredential, error)
	ListByTenant(ctx context.Context, act actor.Actor, tenantID uuid.UUID) ([]*ProviderCredential, error)
	SoftDelete(ctx context.Context, act actor.Actor, id uuid.UUID) error
}
