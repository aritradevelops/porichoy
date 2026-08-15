package postgres

import (
	"context"
	"time"

	"github.com/aritradevelops/porichoy/server/internal/identity"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// passwordModel is Bun's mapping of the passwords table (DATA_MODEL.md `passwords`) — kept
// separate from identity.Password (CODING_STANDARDS.md §2).
type passwordModel struct {
	bun.BaseModel `bun:"table:passwords,alias:pw"`

	ID           uuid.UUID `bun:"id,pk,type:uuid"`
	UserID       uuid.UUID `bun:"user_id,type:uuid"`
	PasswordHash string    `bun:"password_hash"`

	CreatedAt time.Time `bun:"created_at"`

	DeletedAt *time.Time `bun:"deleted_at"`
	DeletedBy *uuid.UUID `bun:"deleted_by,type:uuid"`
}

func passwordToModel(p *identity.Password) *passwordModel {
	return &passwordModel{
		ID:           p.ID,
		UserID:       p.UserID,
		PasswordHash: p.PasswordHash,
		CreatedAt:    p.CreatedAt,
		DeletedAt:    p.DeletedAt,
		DeletedBy:    p.DeletedBy,
	}
}

// PasswordRepository implements identity.PasswordRepository using Postgres via Bun.
type PasswordRepository struct {
	db bun.IDB
}

// NewPasswordRepository builds a PasswordRepository from an open Bun connection.
func NewPasswordRepository(db *bun.DB) *PasswordRepository {
	return &PasswordRepository{db: db}
}

var _ identity.PasswordRepository = (*PasswordRepository)(nil)

// Create persists p as the user's new active password (identity.PasswordRepository) — no
// actor.Actor. Participates in an ambient transaction if ctx carries one (tx.go). Uniqueness
// (at most one active password per user) is enforced by a partial unique index, not
// application code.
func (r *PasswordRepository) Create(ctx context.Context, p *identity.Password) error {
	_, err := dbFromContext(ctx, r.db).NewInsert().Model(passwordToModel(p)).Exec(ctx)
	return err
}
