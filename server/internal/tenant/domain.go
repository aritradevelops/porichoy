package tenant

import (
	"context"
	"time"

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
	FindByDomain(ctx context.Context, domain string) (*TenantDomain, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*TenantDomain, error)
	SoftDelete(ctx context.Context, id uuid.UUID, deletedBy *uuid.UUID) error
}
