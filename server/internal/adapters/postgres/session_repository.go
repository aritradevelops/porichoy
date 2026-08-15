package postgres

import (
	"context"
	"time"

	"github.com/aritradevelops/porichoy/server/internal/app"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// sessionModel is Bun's mapping of the sessions table (DATA_MODEL.md `sessions`) — kept
// separate from app.Session (CODING_STANDARDS.md §2).
type sessionModel struct {
	bun.BaseModel `bun:"table:sessions,alias:se"`

	ID               uuid.UUID `bun:"id,pk,type:uuid"`
	UserID           uuid.UUID `bun:"user_id,type:uuid"`
	AppID            uuid.UUID `bun:"app_id,type:uuid"`
	Aud              string    `bun:"aud"`
	RefreshTokenHash string    `bun:"refresh_token_hash"`
	DeviceLabel      *string   `bun:"device_label"`
	IPAddress        *string   `bun:"ip_address"`

	CreatedAt    time.Time `bun:"created_at"`
	LastActiveAt time.Time `bun:"last_active_at"`
	ExpiresAt    time.Time `bun:"expires_at"`

	DeletedAt *time.Time `bun:"deleted_at"`
	DeletedBy *uuid.UUID `bun:"deleted_by,type:uuid"`
}

func sessionToModel(s *app.Session) *sessionModel {
	return &sessionModel{
		ID:               s.ID,
		UserID:           s.UserID,
		AppID:            s.AppID,
		Aud:              s.Aud,
		RefreshTokenHash: s.RefreshTokenHash,
		DeviceLabel:      s.DeviceLabel,
		IPAddress:        s.IPAddress,
		CreatedAt:        s.CreatedAt,
		LastActiveAt:     s.LastActiveAt,
		ExpiresAt:        s.ExpiresAt,
		DeletedAt:        s.DeletedAt,
		DeletedBy:        s.DeletedBy,
	}
}

// SessionRepository implements app.SessionRepository using Postgres via Bun.
type SessionRepository struct {
	db bun.IDB
}

// NewSessionRepository builds a SessionRepository from an open Bun connection.
func NewSessionRepository(db *bun.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

var _ app.SessionRepository = (*SessionRepository)(nil)

// Create persists s (app.SessionRepository) — no actor.Actor, a session's creator is always
// its own UserID. Participates in an ambient transaction if ctx carries one (tx.go).
func (r *SessionRepository) Create(ctx context.Context, s *app.Session) error {
	_, err := dbFromContext(ctx, r.db).NewInsert().Model(sessionToModel(s)).Exec(ctx)
	return err
}
