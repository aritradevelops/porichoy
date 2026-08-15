package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/aritradevelops/porichoy/server/internal/app"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// appModel is Bun's mapping of the apps table (DATA_MODEL.md `apps`) — kept separate from
// app.App (CODING_STANDARDS.md §2).
type appModel struct {
	bun.BaseModel `bun:"table:apps,alias:ap"`

	ID                     uuid.UUID  `bun:"id,pk,type:uuid"`
	TenantID               uuid.UUID  `bun:"tenant_id,type:uuid"`
	Name                   string     `bun:"name"`
	ClientID               string     `bun:"client_id"`
	ClientSecretHash       *string    `bun:"client_secret_hash"`
	RedirectURIs           []string   `bun:"redirect_uris,type:jsonb"`
	LogoURL                *string    `bun:"logo_url"`
	IsSystem               bool       `bun:"is_system"`
	SupportsOrganizations  bool       `bun:"supports_organizations"`
	DefaultSignupRoleID    *uuid.UUID `bun:"default_signup_role_id,type:uuid"`
	SigningAlgorithm       string     `bun:"signing_algorithm"`
	SigningKeyConfig       []byte     `bun:"signing_key_config_encrypted"`
	AccessTokenTTLSeconds  int        `bun:"access_token_ttl_seconds"`
	IDTokenTTLSeconds      int        `bun:"id_token_ttl_seconds"`
	RefreshTokenTTLSeconds int        `bun:"refresh_token_ttl_seconds"`

	CreatedAt time.Time  `bun:"created_at"`
	UpdatedAt time.Time  `bun:"updated_at"`
	CreatedBy *uuid.UUID `bun:"created_by,type:uuid"`
	UpdatedBy *uuid.UUID `bun:"updated_by,type:uuid"`

	DeletedAt *time.Time `bun:"deleted_at"`
	DeletedBy *uuid.UUID `bun:"deleted_by,type:uuid"`
}

func appFromModel(m *appModel) *app.App {
	return &app.App{
		ID:                     m.ID,
		TenantID:               m.TenantID,
		Name:                   m.Name,
		ClientID:               m.ClientID,
		ClientSecretHash:       m.ClientSecretHash,
		RedirectURIs:           m.RedirectURIs,
		LogoURL:                m.LogoURL,
		IsSystem:               m.IsSystem,
		SupportsOrganizations:  m.SupportsOrganizations,
		DefaultSignupRoleID:    m.DefaultSignupRoleID,
		SigningAlgorithm:       app.SigningAlgorithm(m.SigningAlgorithm),
		SigningKeyConfig:       m.SigningKeyConfig,
		AccessTokenTTLSeconds:  m.AccessTokenTTLSeconds,
		IDTokenTTLSeconds:      m.IDTokenTTLSeconds,
		RefreshTokenTTLSeconds: m.RefreshTokenTTLSeconds,
		CreatedAt:              m.CreatedAt,
		UpdatedAt:              m.UpdatedAt,
		CreatedBy:              m.CreatedBy,
		UpdatedBy:              m.UpdatedBy,
		DeletedAt:              m.DeletedAt,
		DeletedBy:              m.DeletedBy,
	}
}

func appToModel(a *app.App) *appModel {
	uris := a.RedirectURIs
	if uris == nil {
		uris = []string{}
	}
	return &appModel{
		ID:                     a.ID,
		TenantID:               a.TenantID,
		Name:                   a.Name,
		ClientID:               a.ClientID,
		ClientSecretHash:       a.ClientSecretHash,
		RedirectURIs:           uris,
		LogoURL:                a.LogoURL,
		IsSystem:               a.IsSystem,
		SupportsOrganizations:  a.SupportsOrganizations,
		DefaultSignupRoleID:    a.DefaultSignupRoleID,
		SigningAlgorithm:       string(a.SigningAlgorithm),
		SigningKeyConfig:       a.SigningKeyConfig,
		AccessTokenTTLSeconds:  a.AccessTokenTTLSeconds,
		IDTokenTTLSeconds:      a.IDTokenTTLSeconds,
		RefreshTokenTTLSeconds: a.RefreshTokenTTLSeconds,
		CreatedAt:              a.CreatedAt,
		UpdatedAt:              a.UpdatedAt,
		CreatedBy:              a.CreatedBy,
		UpdatedBy:              a.UpdatedBy,
		DeletedAt:              a.DeletedAt,
		DeletedBy:              a.DeletedBy,
	}
}

// AppRepository implements app.Repository using Postgres via Bun.
type AppRepository struct {
	db bun.IDB
}

// NewAppRepository builds an AppRepository from an open Bun connection.
func NewAppRepository(db *bun.DB) *AppRepository {
	return &AppRepository{db: db}
}

var _ app.Repository = (*AppRepository)(nil)

// CreateSystem persists a's system app row (app.Repository) — no actor.Actor, the CLI seed's
// bootstrap path. Participates in an ambient transaction if ctx carries one (tx.go).
func (r *AppRepository) CreateSystem(ctx context.Context, a *app.App) error {
	_, err := dbFromContext(ctx, r.db).NewInsert().Model(appToModel(a)).Exec(ctx)
	return err
}

// FindSystemAppByTenant is a pre-authentication, unscoped lookup of tenantID's default system
// app (app.Repository) — used by identity.Service.Signup. Returns nil, nil if the tenant has
// none (root-tenant-only scope this pass).
func (r *AppRepository) FindSystemAppByTenant(ctx context.Context, tenantID uuid.UUID) (*app.App, error) {
	m := new(appModel)
	err := dbFromContext(ctx, r.db).NewSelect().Model(m).
		Where("tenant_id = ?", tenantID).
		Where("is_system").
		Where("deleted_at IS NULL").
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return appFromModel(m), nil
}

// SetDefaultSignupRole sets appID's default_signup_role_id (app.Repository) — a single-column
// update, no actor.Actor.
func (r *AppRepository) SetDefaultSignupRole(ctx context.Context, appID, roleID uuid.UUID) error {
	_, err := dbFromContext(ctx, r.db).NewUpdate().
		Model((*appModel)(nil)).
		Set("default_signup_role_id = ?", roleID).
		Where("id = ?", appID).
		Exec(ctx)
	return err
}
