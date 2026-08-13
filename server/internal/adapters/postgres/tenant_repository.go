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

// tenantModel is Bun's mapping of the tenant table (DATA_MODEL.md `tenant`) — kept separate
// from tenant.Tenant so that package stays infra-agnostic (CODING_STANDARDS.md §2).
type tenantModel struct {
	bun.BaseModel `bun:"table:tenants,alias:t"`

	ID       uuid.UUID  `bun:"id,pk,type:uuid"`
	ParentID *uuid.UUID `bun:"parent_id,type:uuid"`
	Name     string     `bun:"name"`
	// Ancestors is a Postgres uuid[] column, indexed with GIN (migration
	// 20260813000001) — see tenant.Tenant.Ancestors for what it holds and why.
	Ancestors []uuid.UUID `bun:"ancestors,type:uuid[]"`

	LogoURL             string   `bun:"logo_url"`
	BrandImageURL       *string  `bun:"brand_image_url"`
	LoginLayout         string   `bun:"login_layout"`
	MFARequired         bool     `bun:"mfa_required"`
	EnabledLoginMethods []string `bun:"enabled_login_methods,type:jsonb"`
	AuditRetentionDays  int      `bun:"audit_retention_days"`

	CreatedAt time.Time  `bun:"created_at"`
	UpdatedAt time.Time  `bun:"updated_at"`
	CreatedBy *uuid.UUID `bun:"created_by,type:uuid"`
	UpdatedBy *uuid.UUID `bun:"updated_by,type:uuid"`

	DeletedAt *time.Time `bun:"deleted_at"`
	DeletedBy *uuid.UUID `bun:"deleted_by,type:uuid"`
}

func tenantFromModel(m *tenantModel) *tenant.Tenant {
	methods := make([]tenant.LoginMethod, len(m.EnabledLoginMethods))
	for i, v := range m.EnabledLoginMethods {
		methods[i] = tenant.LoginMethod(v)
	}
	return &tenant.Tenant{
		ID:                  m.ID,
		ParentID:            m.ParentID,
		Name:                m.Name,
		Ancestors:           m.Ancestors,
		LogoURL:             m.LogoURL,
		BrandImageURL:       m.BrandImageURL,
		LoginLayout:         tenant.LoginLayout(m.LoginLayout),
		MFARequired:         m.MFARequired,
		EnabledLoginMethods: methods,
		AuditRetentionDays:  m.AuditRetentionDays,
		CreatedAt:           m.CreatedAt,
		UpdatedAt:           m.UpdatedAt,
		CreatedBy:           m.CreatedBy,
		UpdatedBy:           m.UpdatedBy,
		DeletedAt:           m.DeletedAt,
		DeletedBy:           m.DeletedBy,
	}
}

func tenantToModel(t *tenant.Tenant) *tenantModel {
	methods := make([]string, len(t.EnabledLoginMethods))
	for i, v := range t.EnabledLoginMethods {
		methods[i] = string(v)
	}
	// Bun's array appender writes a nil slice as SQL NULL, not '{}' — coerce so the
	// column's NOT NULL constraint is satisfied regardless of whether the caller populated
	// Ancestors (root tenants and any hand-built fixture leave it nil).
	ancestors := t.Ancestors
	if ancestors == nil {
		ancestors = []uuid.UUID{}
	}
	return &tenantModel{
		ID:                  t.ID,
		ParentID:            t.ParentID,
		Name:                t.Name,
		Ancestors:           ancestors,
		LogoURL:             t.LogoURL,
		BrandImageURL:       t.BrandImageURL,
		LoginLayout:         string(t.LoginLayout),
		MFARequired:         t.MFARequired,
		EnabledLoginMethods: methods,
		AuditRetentionDays:  t.AuditRetentionDays,
		CreatedAt:           t.CreatedAt,
		UpdatedAt:           t.UpdatedAt,
		CreatedBy:           t.CreatedBy,
		UpdatedBy:           t.UpdatedBy,
		DeletedAt:           t.DeletedAt,
		DeletedBy:           t.DeletedBy,
	}
}

// TenantRepository implements tenant.Repository using Postgres via Bun.
type TenantRepository struct {
	db *bun.DB
}

// NewTenantRepository builds a TenantRepository from an open Bun connection.
func NewTenantRepository(db *bun.DB) *TenantRepository {
	return &TenantRepository{db: db}
}

var _ tenant.Repository = (*TenantRepository)(nil)

// withTenantScope restricts q to rows act is authorized to touch: itself, or any row whose
// ancestors contain act.TenantID (via the @> containment operator, so the GIN index on
// ancestors — migration 20260813000001 — is actually usable; `? = ANY(ancestors)` couldn't
// use it). Root passes through unrestricted; any scope other than root/tenant is denied
// outright — org/app/own aren't valid scopes for this module (AUTHORIZATION_MODEL.md §4).
//
// Generic over bun's own QueryBuilder abstraction so the exact same restriction composes
// onto a SelectQuery or UpdateQuery alike — this is bun's documented mechanism for sharing
// filters across query types, rather than a bespoke fetch-then-check-in-Go helper.
func withTenantScope[Q interface{ QueryBuilder() bun.QueryBuilder }](q Q, act actor.Actor) Q {
	qb := q.QueryBuilder()
	switch act.Scope {
	case actor.ScopeRoot:
		// no restriction
	case actor.ScopeTenant:
		qb = qb.Where("id = ? OR ancestors @> ARRAY[?]::uuid[]", act.TenantID, act.TenantID)
	default:
		qb = qb.Where("FALSE")
	}
	return qb.Unwrap().(Q)
}

// CreateRoot persists the root tenant (tenant.Repository). Pre-authentication bootstrap
// only — no actor.Actor exists yet, t.ParentID is always nil, t.Ancestors is always empty.
func (r *TenantRepository) CreateRoot(ctx context.Context, t *tenant.Tenant) error {
	_, err := r.db.NewInsert().Model(tenantToModel(t)).Exec(ctx)
	return err
}

// Create persists a child tenant under t.ParentID (tenant.Repository). Runs one scoped
// lookup of the parent — reusing withTenantScope, the same filter GetByID/ListChildren use
// — which does double duty as both the authorization check (root: any parent; tenant scope:
// parent must be act's own tenant or a descendant of it) and the source data for
// t.Ancestors. A missing or out-of-scope parent both surface as tenant.ErrTenantNotFound,
// preserving the "not found, not forbidden" convention used everywhere else here.
func (r *TenantRepository) Create(ctx context.Context, act actor.Actor, t *tenant.Tenant) error {
	parent := new(tenantModel)
	q := r.db.NewSelect().Model(parent).
		Where("id = ?", *t.ParentID).
		Where("deleted_at IS NULL")
	err := withTenantScope(q, act).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return tenant.ErrTenantNotFound
	}
	if err != nil {
		return err
	}

	t.Ancestors = append(append([]uuid.UUID{}, parent.Ancestors...), parent.ID)
	_, err = r.db.NewInsert().Model(tenantToModel(t)).Exec(ctx)
	return err
}

// FindByID is the pre-authentication, unscoped lookup (tenant.Repository) — used by
// tenant.Service.ResolveTenantByDomain, which runs before an actor.Actor can exist.
func (r *TenantRepository) FindByID(ctx context.Context, id uuid.UUID) (*tenant.Tenant, error) {
	m := new(tenantModel)
	err := r.db.NewSelect().Model(m).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return tenantFromModel(m), nil
}

// GetByID is the authorized, scope-filtered lookup (tenant.Repository). The tenant module's
// scope→filter mapping (AUTHORIZATION_MODEL.md §4) is descendant-based, via withTenantScope:
// below root scope, a caller may fetch their own act.TenantID or any descendant of it. A
// mismatch returns nil, nil (not an error) — the existing "not found rather than forbidden"
// convention for this package, so a caller outside their subtree can't distinguish "doesn't
// exist" from "exists but isn't yours".
func (r *TenantRepository) GetByID(ctx context.Context, act actor.Actor, id uuid.UUID) (*tenant.Tenant, error) {
	m := new(tenantModel)
	q := r.db.NewSelect().Model(m).
		Where("id = ?", id).
		Where("deleted_at IS NULL")
	err := withTenantScope(q, act).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return tenantFromModel(m), nil
}

// Update persists t's current field values. t already carries UpdatedBy (set by
// tenant.Service), so no actor.Actor is needed here (CODING_STANDARDS.md §4) — by the time
// Update is called, t was already fetched (and thus authorized) via GetByID.
func (r *TenantRepository) Update(ctx context.Context, t *tenant.Tenant) error {
	_, err := r.db.NewUpdate().Model(tenantToModel(t)).WherePK().Exec(ctx)
	return err
}

// SoftDelete marks id deleted by act.PrincipalID (tenant.Repository), scope-filtered the
// same way as GetByID — a mismatch or missing row updates nothing (RowsAffected == 0) and
// returns nil, mirroring the "not found" framing used throughout this package. On an actual
// delete, also strips id out of every other tenant's ancestors array in the same
// transaction, per AUTHORIZATION_MODEL.md §4 — this does not cascade-delete descendants or
// touch their parent_id, only the stale ancestor reference.
func (r *TenantRepository) SoftDelete(ctx context.Context, act actor.Actor, id uuid.UUID) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		q := tx.NewUpdate().
			Model((*tenantModel)(nil)).
			Set("deleted_at = ?", time.Now()).
			Set("deleted_by = ?", act.PrincipalID).
			Where("id = ?", id).
			Where("deleted_at IS NULL")
		res, err := withTenantScope(q, act).Exec(ctx)
		if err != nil {
			return err
		}
		n, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if n == 0 {
			return nil
		}

		_, err = tx.NewUpdate().
			Table("tenants").
			Set("ancestors = array_remove(ancestors, ?)", id).
			Where("ancestors @> ARRAY[?]::uuid[]", id).
			Exec(ctx)
		return err
	})
}

// ListChildren returns the direct children of parentID — one level of the hierarchy, not
// the full subtree (tenant.Repository). Scope-filtered the same way as GetByID, but applied
// to the children rows themselves rather than parentID: a child's ancestors always includes
// its parent's own ancestors plus the parent's ID (see Create), so "is act.TenantID an
// ancestor-or-self of this child" already captures "is this child under a parent the actor
// can reach" — no separate parent lookup needed.
func (r *TenantRepository) ListChildren(ctx context.Context, act actor.Actor, parentID uuid.UUID) ([]*tenant.Tenant, error) {
	var models []*tenantModel
	q := r.db.NewSelect().Model(&models).
		Where("parent_id = ?", parentID).
		Where("deleted_at IS NULL")
	if err := withTenantScope(q, act).Scan(ctx); err != nil {
		return nil, err
	}
	result := make([]*tenant.Tenant, len(models))
	for i, m := range models {
		result[i] = tenantFromModel(m)
	}
	return result, nil
}
