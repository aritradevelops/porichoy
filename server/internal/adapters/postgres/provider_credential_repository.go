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

// providerCredentialModel is Bun's mapping of the tenant_provider_credential table
// (DATA_MODEL.md `tenant_provider_credential`) — kept separate from
// tenant.ProviderCredential (CODING_STANDARDS.md §2).
type providerCredentialModel struct {
	bun.BaseModel `bun:"table:tenant_provider_credentials,alias:tpc"`

	ID              uuid.UUID `bun:"id,pk,type:uuid"`
	TenantID        uuid.UUID `bun:"tenant_id,type:uuid"`
	ProviderType    string    `bun:"provider_type"`
	ConfigEncrypted []byte    `bun:"config_encrypted"`

	CreatedAt time.Time  `bun:"created_at"`
	UpdatedAt time.Time  `bun:"updated_at"`
	CreatedBy *uuid.UUID `bun:"created_by,type:uuid"`
	UpdatedBy *uuid.UUID `bun:"updated_by,type:uuid"`

	DeletedAt *time.Time `bun:"deleted_at"`
	DeletedBy *uuid.UUID `bun:"deleted_by,type:uuid"`
}

func credentialFromModel(m *providerCredentialModel) *tenant.ProviderCredential {
	return &tenant.ProviderCredential{
		ID:              m.ID,
		TenantID:        m.TenantID,
		ProviderType:    tenant.ProviderType(m.ProviderType),
		ConfigEncrypted: m.ConfigEncrypted,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
		CreatedBy:       m.CreatedBy,
		UpdatedBy:       m.UpdatedBy,
		DeletedAt:       m.DeletedAt,
		DeletedBy:       m.DeletedBy,
	}
}

func credentialToModel(c *tenant.ProviderCredential) *providerCredentialModel {
	return &providerCredentialModel{
		ID:              c.ID,
		TenantID:        c.TenantID,
		ProviderType:    string(c.ProviderType),
		ConfigEncrypted: c.ConfigEncrypted,
		CreatedAt:       c.CreatedAt,
		UpdatedAt:       c.UpdatedAt,
		CreatedBy:       c.CreatedBy,
		UpdatedBy:       c.UpdatedBy,
		DeletedAt:       c.DeletedAt,
		DeletedBy:       c.DeletedBy,
	}
}

// ProviderCredentialRepository implements tenant.ProviderCredentialRepository using
// Postgres via Bun.
type ProviderCredentialRepository struct {
	db *bun.DB
}

// NewProviderCredentialRepository builds a ProviderCredentialRepository from an open Bun
// connection.
func NewProviderCredentialRepository(db *bun.DB) *ProviderCredentialRepository {
	return &ProviderCredentialRepository{db: db}
}

var _ tenant.ProviderCredentialRepository = (*ProviderCredentialRepository)(nil)

// Upsert creates or replaces the credential for c's (TenantID, ProviderType) pair
// (tenant.ProviderCredentialRepository) — a real INSERT ... ON CONFLICT DO UPDATE, so it
// behaves correctly regardless of whether tenant.Service.SetProviderCredential is creating
// (c.ID freshly generated) or updating (c.ID from the row it fetched) in place. c already
// carries CreatedBy/UpdatedBy (set by the Service), so no actor.Actor is needed here
// (CODING_STANDARDS.md §4).
func (r *ProviderCredentialRepository) Upsert(ctx context.Context, c *tenant.ProviderCredential) error {
	_, err := r.db.NewInsert().
		Model(credentialToModel(c)).
		On("CONFLICT (id) DO UPDATE").
		Set("config_encrypted = EXCLUDED.config_encrypted").
		Set("updated_at = EXCLUDED.updated_at").
		Set("updated_by = EXCLUDED.updated_by").
		Exec(ctx)
	return err
}

// FindByTenantAndType looks up tenantID's active credential for providerType
// (tenant.ProviderCredentialRepository).
func (r *ProviderCredentialRepository) FindByTenantAndType(ctx context.Context, act actor.Actor, tenantID uuid.UUID, providerType tenant.ProviderType) (*tenant.ProviderCredential, error) {
	_ = act
	m := new(providerCredentialModel)
	err := r.db.NewSelect().Model(m).
		Where("tenant_id = ?", tenantID).
		Where("provider_type = ?", string(providerType)).
		Where("deleted_at IS NULL").
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return credentialFromModel(m), nil
}

// ListByTenant lists tenantID's active provider credentials
// (tenant.ProviderCredentialRepository).
func (r *ProviderCredentialRepository) ListByTenant(ctx context.Context, act actor.Actor, tenantID uuid.UUID) ([]*tenant.ProviderCredential, error) {
	_ = act
	var models []*providerCredentialModel
	if err := r.db.NewSelect().Model(&models).
		Where("tenant_id = ?", tenantID).
		Where("deleted_at IS NULL").
		Scan(ctx); err != nil {
		return nil, err
	}
	result := make([]*tenant.ProviderCredential, len(models))
	for i, m := range models {
		result[i] = credentialFromModel(m)
	}
	return result, nil
}

// SoftDelete removes id, freeing its (tenant, provider type) pair for reuse (the partial
// unique index on tenant_provider_credential only covers non-deleted rows).
func (r *ProviderCredentialRepository) SoftDelete(ctx context.Context, act actor.Actor, id uuid.UUID) error {
	_, err := r.db.NewUpdate().
		Model((*providerCredentialModel)(nil)).
		Set("deleted_at = ?", time.Now()).
		Set("deleted_by = ?", act.PrincipalID).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Exec(ctx)
	return err
}
