package postgres

import (
	"context"
	"time"

	"github.com/aritradevelops/porichoy/server/internal/authorization"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// roleAssignmentModel is Bun's mapping of the role_assignments table (DATA_MODEL.md
// `role_assignments`) — kept separate from authorization.RoleAssignment
// (CODING_STANDARDS.md §2).
type roleAssignmentModel struct {
	bun.BaseModel `bun:"table:role_assignments,alias:ra"`

	ID          uuid.UUID `bun:"id,pk,type:uuid"`
	PrincipalID uuid.UUID `bun:"principal_id,type:uuid"`
	RoleID      uuid.UUID `bun:"role_id,type:uuid"`

	CreatedAt time.Time  `bun:"created_at"`
	CreatedBy *uuid.UUID `bun:"created_by,type:uuid"`

	DeletedAt *time.Time `bun:"deleted_at"`
	DeletedBy *uuid.UUID `bun:"deleted_by,type:uuid"`
}

func roleAssignmentFromModel(m *roleAssignmentModel) *authorization.RoleAssignment {
	return &authorization.RoleAssignment{
		ID:          m.ID,
		PrincipalID: m.PrincipalID,
		RoleID:      m.RoleID,
		CreatedAt:   m.CreatedAt,
		CreatedBy:   m.CreatedBy,
		DeletedAt:   m.DeletedAt,
		DeletedBy:   m.DeletedBy,
	}
}

func roleAssignmentToModel(ra *authorization.RoleAssignment) *roleAssignmentModel {
	return &roleAssignmentModel{
		ID:          ra.ID,
		PrincipalID: ra.PrincipalID,
		RoleID:      ra.RoleID,
		CreatedAt:   ra.CreatedAt,
		CreatedBy:   ra.CreatedBy,
		DeletedAt:   ra.DeletedAt,
		DeletedBy:   ra.DeletedBy,
	}
}

// RoleAssignmentRepository implements authorization.RoleAssignmentRepository using Postgres
// via Bun.
type RoleAssignmentRepository struct {
	db bun.IDB
}

// NewRoleAssignmentRepository builds a RoleAssignmentRepository from an open Bun connection.
func NewRoleAssignmentRepository(db *bun.DB) *RoleAssignmentRepository {
	return &RoleAssignmentRepository{db: db}
}

var _ authorization.RoleAssignmentRepository = (*RoleAssignmentRepository)(nil)

// Create binds ra.PrincipalID to ra.RoleID (authorization.RoleAssignmentRepository) — no
// actor.Actor, used both by the CLI seed bootstrap and identity.Service.Signup (pre-auth).
// Participates in an ambient transaction if ctx carries one (tx.go).
func (r *RoleAssignmentRepository) Create(ctx context.Context, ra *authorization.RoleAssignment) error {
	_, err := dbFromContext(ctx, r.db).NewInsert().Model(roleAssignmentToModel(ra)).Exec(ctx)
	return err
}

// ListByPrincipal returns every non-deleted RoleAssignment for principalID
// (authorization.RoleAssignmentRepository) — backs Service.EffectivePermissions.
// Participates in an ambient transaction if ctx carries one (tx.go).
func (r *RoleAssignmentRepository) ListByPrincipal(ctx context.Context, principalID uuid.UUID) ([]*authorization.RoleAssignment, error) {
	var models []*roleAssignmentModel
	err := dbFromContext(ctx, r.db).NewSelect().Model(&models).
		Where("principal_id = ?", principalID).
		Where("deleted_at IS NULL").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	assignments := make([]*authorization.RoleAssignment, len(models))
	for i, m := range models {
		assignments[i] = roleAssignmentFromModel(m)
	}
	return assignments, nil
}
