package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/aritradevelops/porichoy/server/internal/authorization"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// roleModel is Bun's mapping of the roles table (DATA_MODEL.md `roles`) — kept separate from
// authorization.Role (CODING_STANDARDS.md §2).
type roleModel struct {
	bun.BaseModel `bun:"table:roles,alias:ro"`

	ID          uuid.UUID       `bun:"id,pk,type:uuid"`
	TenantID    uuid.UUID       `bun:"tenant_id,type:uuid"`
	AppID       *uuid.UUID      `bun:"app_id,type:uuid"`
	Name        string          `bun:"name"`
	IsSystem    bool            `bun:"is_system"`
	Permissions []string        `bun:"permissions,type:text[]"`
	Policies    json.RawMessage `bun:"policies,type:jsonb"`

	CreatedAt time.Time  `bun:"created_at"`
	UpdatedAt time.Time  `bun:"updated_at"`
	CreatedBy *uuid.UUID `bun:"created_by,type:uuid"`
	UpdatedBy *uuid.UUID `bun:"updated_by,type:uuid"`

	DeletedAt *time.Time `bun:"deleted_at"`
	DeletedBy *uuid.UUID `bun:"deleted_by,type:uuid"`
}

func roleFromModel(m *roleModel) *authorization.Role {
	return &authorization.Role{
		ID:          m.ID,
		TenantID:    m.TenantID,
		AppID:       m.AppID,
		Name:        m.Name,
		IsSystem:    m.IsSystem,
		Permissions: m.Permissions,
		Policies:    m.Policies,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		CreatedBy:   m.CreatedBy,
		UpdatedBy:   m.UpdatedBy,
		DeletedAt:   m.DeletedAt,
		DeletedBy:   m.DeletedBy,
	}
}

func roleToModel(r *authorization.Role) *roleModel {
	perms := r.Permissions
	if perms == nil {
		perms = []string{}
	}
	policies := r.Policies
	if policies == nil {
		policies = json.RawMessage("[]")
	}
	return &roleModel{
		ID:          r.ID,
		TenantID:    r.TenantID,
		AppID:       r.AppID,
		Name:        r.Name,
		IsSystem:    r.IsSystem,
		Permissions: perms,
		Policies:    policies,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
		CreatedBy:   r.CreatedBy,
		UpdatedBy:   r.UpdatedBy,
		DeletedAt:   r.DeletedAt,
		DeletedBy:   r.DeletedBy,
	}
}

// RoleRepository implements authorization.RoleRepository using Postgres via Bun.
type RoleRepository struct {
	db bun.IDB
}

// NewRoleRepository builds a RoleRepository from an open Bun connection.
func NewRoleRepository(db *bun.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

var _ authorization.RoleRepository = (*RoleRepository)(nil)

// CreateSystem persists r's system-provisioned role row (authorization.RoleRepository) — no
// actor.Actor, the CLI seed's bootstrap path. Participates in an ambient transaction if ctx
// carries one (tx.go).
func (r *RoleRepository) CreateSystem(ctx context.Context, role *authorization.Role) error {
	_, err := dbFromContext(ctx, r.db).NewInsert().Model(roleToModel(role)).Exec(ctx)
	return err
}

// FindByIDs returns every non-deleted Role among ids (authorization.RoleRepository) —
// backs Service.EffectivePermissions. Participates in an ambient transaction if ctx carries
// one (tx.go).
func (r *RoleRepository) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]*authorization.Role, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var models []*roleModel
	err := dbFromContext(ctx, r.db).NewSelect().Model(&models).
		Where("id IN (?)", bun.List(ids)).
		Where("deleted_at IS NULL").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	roles := make([]*authorization.Role, len(models))
	for i, m := range models {
		roles[i] = roleFromModel(m)
	}
	return roles, nil
}
