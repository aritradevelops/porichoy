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
	return &tenantModel{
		ID:                  t.ID,
		ParentID:            t.ParentID,
		Name:                t.Name,
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

// Create persists t. t already carries CreatedBy/UpdatedBy (set by tenant.Service), so no
// actor.Actor is needed here (CODING_STANDARDS.md §4).
func (r *TenantRepository) Create(ctx context.Context, t *tenant.Tenant) error {
	_, err := r.db.NewInsert().Model(tenantToModel(t)).Exec(ctx)
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

// GetByID is the authorized, scope-filtered lookup (tenant.Repository). The tenant
// module's scope→filter mapping (AUTHORIZATION_MODEL.md §4) is exact-match: below root
// scope, a caller may only fetch their own act.TenantID, never another tenant — even a
// sibling or a descendant. A mismatch returns nil, nil (not an error), consistent with
// FindByID/RegisterDomain's existing "not found rather than forbidden" convention for
// this package, so a caller outside their tenant can't distinguish "doesn't exist" from
// "exists but isn't yours."
func (r *TenantRepository) GetByID(ctx context.Context, act actor.Actor, id uuid.UUID) (*tenant.Tenant, error) {
	if act.Scope != actor.ScopeRoot && id != act.TenantID {
		return nil, nil
	}
	return r.FindByID(ctx, id)
}

// Update persists t's current field values. t already carries UpdatedBy (set by
// tenant.Service), so no actor.Actor is needed here (CODING_STANDARDS.md §4).
func (r *TenantRepository) Update(ctx context.Context, t *tenant.Tenant) error {
	_, err := r.db.NewUpdate().Model(tenantToModel(t)).WherePK().Exec(ctx)
	return err
}

// SoftDelete marks id deleted by act.PrincipalID. There's no entity here to carry that
// value, unlike Create/Update, so it's a required parameter (CODING_STANDARDS.md §4).
func (r *TenantRepository) SoftDelete(ctx context.Context, act actor.Actor, id uuid.UUID) error {
	_, err := r.db.NewUpdate().
		Model((*tenantModel)(nil)).
		Set("deleted_at = ?", time.Now()).
		Set("deleted_by = ?", act.PrincipalID).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Exec(ctx)
	return err
}

// ListChildren returns the direct children of parentID — one level of the hierarchy, not
// the full subtree (tenant.Repository). Same exact-match scope filter as GetByID
// (AUTHORIZATION_MODEL.md §4): below root scope, a caller may only list children of their
// own act.TenantID. A mismatch returns an empty slice, not an error, mirroring GetByID's
// "not found" rather than "forbidden" framing.
func (r *TenantRepository) ListChildren(ctx context.Context, act actor.Actor, parentID uuid.UUID) ([]*tenant.Tenant, error) {
	if act.Scope != actor.ScopeRoot && parentID != act.TenantID {
		return nil, nil
	}
	var models []*tenantModel
	if err := r.db.NewSelect().Model(&models).
		Where("parent_id = ?", parentID).
		Where("deleted_at IS NULL").
		Scan(ctx); err != nil {
		return nil, err
	}
	result := make([]*tenant.Tenant, len(models))
	for i, m := range models {
		result[i] = tenantFromModel(m)
	}
	return result, nil
}
