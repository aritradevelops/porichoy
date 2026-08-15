package app

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
)

// Session is one issued refresh token's persisted record (DATA_MODEL.md `sessions`) —
// Postgres is the source of truth, so it's revocable by an admin and visible in a user's own
// "active sessions" list (PRD §9.4) regardless of cache state. Rotate-on-use/reuse-detection
// (TECHNICAL_DESIGN §4) isn't implemented yet — this pass only persists the row a later
// login/refresh endpoint will manage.
type Session struct {
	ID     uuid.UUID
	UserID uuid.UUID
	AppID  uuid.UUID
	// Aud mirrors the issued token's aud claim — the tenant ID for a system-app session
	// (TECHNICAL_DESIGN §3.5).
	Aud              string
	RefreshTokenHash string
	DeviceLabel      *string
	IPAddress        *string

	CreatedAt    time.Time
	LastActiveAt time.Time
	ExpiresAt    time.Time

	DeletedAt *time.Time
	DeletedBy *uuid.UUID
}

// IsDeleted reports whether s has been revoked.
func (s *Session) IsDeleted() bool {
	return s.DeletedAt != nil
}

// SessionRepository persists Sessions.
type SessionRepository interface {
	// Create persists s. No actor.Actor — a session's creator is always its own UserID
	// (DATA_MODEL.md `sessions`), and this is called from pre-auth flows (Signup) anyway.
	Create(ctx context.Context, s *Session) error
}

// NewRefreshToken generates a random refresh token and its storable hash. The raw value is
// returned to the caller exactly once and never persisted; only the hash is. Refresh tokens
// are already high-entropy random values, so a fast hash (SHA-256) is used rather than a
// slow/salted one (bcrypt) — unlike passwords, there's no low-entropy secret to protect
// against offline guessing.
func NewRefreshToken() (raw, hash string, err error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", err
	}
	raw = base64.RawURLEncoding.EncodeToString(buf)
	sum := sha256.Sum256([]byte(raw))
	return raw, hex.EncodeToString(sum[:]), nil
}
