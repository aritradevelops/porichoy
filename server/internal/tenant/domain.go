package tenant

import (
	"context"
	"time"

	"github.com/aritradevelops/porichoy/server/internal/actor"
	"github.com/google/uuid"
)

// TenantDomain maps a registered origin/domain to the tenant it belongs to
// (TECHNICAL_DESIGN §3.3). This is how a self-hosted instance serving multiple
// tenants resolves which tenant an incoming request's Origin header belongs to —
// the highest-traffic lookup in the system, which is why it's a dedicated table
// rather than a column on Tenant (a tenant can register more than one domain).
type TenantDomain struct {
	ID       uuid.UUID
	TenantID uuid.UUID
	Domain   string

	CreatedAt time.Time
	CreatedBy *uuid.UUID

	DeletedAt *time.Time
	DeletedBy *uuid.UUID
}

// IsDeleted reports whether d has been soft-deleted.
func (d *TenantDomain) IsDeleted() bool {
	return d.DeletedAt != nil
}

// DomainRepository persists and retrieves TenantDomains.
type DomainRepository interface {
	Create(ctx context.Context, d *TenantDomain) error
	// FindByDomain resolves the tenant a given origin belongs to. Returns nil,
	// nil if no tenant has registered this domain.
	//
	// This is a global, unique-key lookup — used both pre-authentication (tenant
	// resolution, TECHNICAL_DESIGN.md §3.3, which runs before an actor.Actor can even
	// exist) and inside authorized flows (RegisterDomain's uniqueness check). Domain
	// uniqueness is instance-wide, never scoped, so this method never takes an Actor,
	// in either context.
	FindByDomain(ctx context.Context, domain string) (*TenantDomain, error)
	ListByTenant(ctx context.Context, act actor.Actor, tenantID uuid.UUID) ([]*TenantDomain, error)
	SoftDelete(ctx context.Context, act actor.Actor, id uuid.UUID) error
}
