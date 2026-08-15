package postgres

import (
	"context"
	"database/sql"
	"errors"
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

func passwordFromModel(m *passwordModel) *identity.Password {
	return &identity.Password{
		ID:           m.ID,
		UserID:       m.UserID,
		PasswordHash: m.PasswordHash,
		CreatedAt:    m.CreatedAt,
		DeletedAt:    m.DeletedAt,
		DeletedBy:    m.DeletedBy,
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

// FindByUserID returns userID's currently active password (identity.PasswordRepository) —
// backs Login's credential check. Returns nil, nil if none exists. Participates in an
// ambient transaction if ctx carries one (tx.go).
func (r *PasswordRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*identity.Password, error) {
	m := new(passwordModel)
	err := dbFromContext(ctx, r.db).NewSelect().Model(m).
		Where("user_id = ?", userID).
		Where("deleted_at IS NULL").
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return passwordFromModel(m), nil
}
