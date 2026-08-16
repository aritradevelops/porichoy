package postgres

import (
	"context"

	"github.com/aritradevelops/porichoy/server/internal/actor"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// tenantAccessible reports whether tenantID exists, isn't soft-deleted, and is within act's
// authorized subtree (AUTHORIZATION_MODEL.md §4) — itself, or a descendant of it, via the
// same tenants.ancestors containment check GetByID uses (withTenantScope, tenant_repository.go),
// just as a boolean existence check against one specific ID rather than a filter over many
// rows. Used by DomainRepository.Create/ProviderCredentialRepository.Upsert to authorize
// against the tenant a domain/credential is being created for, and by their
// ListByTenant/FindByTenantAndType/SoftDelete to authorize the tenantID they're already
// given — domain_registries and tenant_provider_credentials don't carry their own ancestors
// column, so the check always goes through this one tenants-table lookup rather than a
// subquery/join against those tables directly.
func tenantAccessible(ctx context.Context, db bun.IDB, act actor.Actor, tenantID uuid.UUID) (bool, error) {
	return withTenantScope(
		db.NewSelect().Model((*tenantModel)(nil)).Where("id = ?", tenantID).Where("deleted_at IS NULL"),
		act,
	).Exists(ctx)
}
