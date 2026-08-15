package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/aritradevelops/porichoy/server/internal/actor"
	"github.com/aritradevelops/porichoy/server/internal/tenant"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// domainModel is Bun's mapping of the domain_registry table (DATA_MODEL.md
// `domain_registry`) — kept separate from tenant.TenantDomain (CODING_STANDARDS.md §2).
type domainModel struct {
	bun.BaseModel `bun:"table:domain_registries,alias:dr"`

	ID       uuid.UUID `bun:"id,pk,type:uuid"`
	TenantID uuid.UUID `bun:"tenant_id,type:uuid"`
	Domain   string    `bun:"domain"`

	CreatedAt time.Time  `bun:"created_at"`
	CreatedBy *uuid.UUID `bun:"created_by,type:uuid"`

	DeletedAt *time.Time `bun:"deleted_at"`
	DeletedBy *uuid.UUID `bun:"deleted_by,type:uuid"`
}

func domainFromModel(m *domainModel) *tenant.TenantDomain {
	return &tenant.TenantDomain{
		ID:        m.ID,
		TenantID:  m.TenantID,
		Domain:    m.Domain,
		CreatedAt: m.CreatedAt,
		CreatedBy: m.CreatedBy,
		DeletedAt: m.DeletedAt,
		DeletedBy: m.DeletedBy,
	}
}

func domainToModel(d *tenant.TenantDomain) *domainModel {
	return &domainModel{
		ID:        d.ID,
		TenantID:  d.TenantID,
		Domain:    d.Domain,
		CreatedAt: d.CreatedAt,
		CreatedBy: d.CreatedBy,
		DeletedAt: d.DeletedAt,
		DeletedBy: d.DeletedBy,
	}
}

// DomainRepository implements tenant.DomainRepository using Postgres via Bun.
type DomainRepository struct {
	db bun.IDB
}

// NewDomainRepository builds a DomainRepository from an open Bun connection.
func NewDomainRepository(db *bun.DB) *DomainRepository {
	return &DomainRepository{db: db}
}

var _ tenant.DomainRepository = (*DomainRepository)(nil)

// Create persists d, after checking act is authorized to create a domain for d.TenantID
// (tenant.DomainRepository) — root may create for any tenant; tenant scope only for itself
// or a descendant (AUTHORIZATION_MODEL.md §4), via tenantAccessible (scope.go). Returns
// tenant.ErrTenantNotFound if not, same "not found, not forbidden" framing used throughout
// this package. The domain's uniqueness (among active rows) is enforced by a partial unique
// index, not application code — a violation surfaces as a plain Postgres error here;
// tenant.Service.RegisterDomain's own FindByDomain check exists to turn that into a clean
// apperror.Error before this call is even made.
func (r *DomainRepository) Create(ctx context.Context, act actor.Actor, d *tenant.TenantDomain) error {
	ok, err := tenantAccessible(ctx, r.db, act, d.TenantID)
	if err != nil {
		return err
	}
	if !ok {
		return tenant.ErrTenantNotFound
	}
	_, err = r.db.NewInsert().Model(domainToModel(d)).Exec(ctx)
	return err
}

// CreateRoot persists d with no authorization check (tenant.DomainRepository) — the CLI
// seed's bootstrap path, registering the root tenant's first domain before any actor.Actor
// exists. Participates in an ambient transaction if ctx carries one (tx.go), same as
// TenantRepository.CreateRoot.
func (r *DomainRepository) CreateRoot(ctx context.Context, d *tenant.TenantDomain) error {
	_, err := dbFromContext(ctx, r.db).NewInsert().Model(domainToModel(d)).Exec(ctx)
	return err
}

// FindByDomain resolves the tenant a given origin belongs to (tenant.DomainRepository) — a
// global, unique-key lookup with no actor.Actor, ever (CODING_STANDARDS.md §4): it backs
// both the pre-authentication tenant-resolution path and the uniqueness check inside the
// authorized RegisterDomain flow, and domain uniqueness is instance-wide, never scoped.
func (r *DomainRepository) FindByDomain(ctx context.Context, domain string) (*tenant.TenantDomain, error) {
	m := new(domainModel)
	err := r.db.NewSelect().Model(m).
		Where("domain = ?", domain).
		Where("deleted_at IS NULL").
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return domainFromModel(m), nil
}

// ListByTenant lists tenantID's registered domains (tenant.DomainRepository), after checking
// act is authorized against tenantID via tenantAccessible (scope.go) — same descendant-access
// rule as Create. Returns nil, nil (not an error) if not, matching GetByID's framing.
func (r *DomainRepository) ListByTenant(ctx context.Context, act actor.Actor, tenantID uuid.UUID) ([]*tenant.TenantDomain, error) {
	ok, err := tenantAccessible(ctx, r.db, act, tenantID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	var models []*domainModel
	if err := r.db.NewSelect().Model(&models).
		Where("tenant_id = ?", tenantID).
		Where("deleted_at IS NULL").
		Scan(ctx); err != nil {
		return nil, err
	}
	result := make([]*tenant.TenantDomain, len(models))
	for i, m := range models {
		result[i] = domainFromModel(m)
	}
	return result, nil
}

// SoftDelete unregisters id, freeing its domain value for reuse (the partial unique index
// on domain_registry only covers non-deleted rows, per the migration's own comment).
// SoftDelete takes only id, not the domain's tenant_id, so it fetches the row first to learn
// it, then checks act's authorization against that tenant via tenantAccessible (scope.go).
// A missing row or a denied check both no-op silently (nil error), same framing as
// TenantRepository.SoftDelete.
func (r *DomainRepository) SoftDelete(ctx context.Context, act actor.Actor, id uuid.UUID) error {
	m := new(domainModel)
	err := r.db.NewSelect().Model(m).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}

	ok, err := tenantAccessible(ctx, r.db, act, m.TenantID)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	_, err = r.db.NewUpdate().
		Model((*domainModel)(nil)).
		Set("deleted_at = ?", time.Now()).
		Set("deleted_by = ?", act.PrincipalID).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Exec(ctx)
	return err
}
