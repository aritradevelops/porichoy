package identity

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// MaxPasswordBytes is bcrypt's own input limit (golang.org/x/crypto/bcrypt silently rejects
// anything longer with bcrypt.ErrPasswordTooLong rather than truncating) — checked explicitly
// so a caller gets a clean, expected error instead of an opaque bcrypt failure surfacing as a
// generic 500 at the REST boundary.
const MaxPasswordBytes = 72

// ErrPasswordTooLong is returned by HashPassword when plain exceeds MaxPasswordBytes.
var ErrPasswordTooLong = errors.New("identity: password too long")

// Password is one historical entry in a user's password history (DATA_MODEL.md `passwords`)
// — not 1:1 with User, a user accumulates many rows over time via soft-delete, of which
// exactly one (deleted_at IS NULL) is active. Only present for users on the email+password
// method.
type Password struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	PasswordHash string

	CreatedAt time.Time

	DeletedAt *time.Time
	DeletedBy *uuid.UUID
}

// IsDeleted reports whether p has been superseded by a newer password.
func (p *Password) IsDeleted() bool {
	return p.DeletedAt != nil
}

// PasswordRepository persists Passwords.
type PasswordRepository interface {
	// Create persists p as the user's new active password. No actor.Actor — same reasoning
	// as Repository.Create.
	Create(ctx context.Context, p *Password) error
	// FindByUserID returns the user's currently active password (deleted_at IS NULL), or
	// nil, nil if none exists — backs Login's credential check. No actor.Actor — same
	// pre-authentication reasoning as Repository.FindByEmail.
	FindByUserID(ctx context.Context, userID uuid.UUID) (*Password, error)
}

// HashPassword hashes plain with bcrypt (TECHNICAL_DESIGN §5) — a plain function, not a port:
// bcrypt is a fixed, single-implementation choice (unlike JWT signing, which genuinely varies
// per app), so a swappable interface would be structure with no second implementation to
// justify it.
func HashPassword(plain string) (string, error) {
	if len(plain) > MaxPasswordBytes {
		return "", ErrPasswordTooLong
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword reports whether plain matches hash.
func VerifyPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

// dummyPasswordHash is a valid bcrypt hash of no real password. Login compares against it
// (instead of short-circuiting) whenever no real Password row exists to check, so the
// bcrypt cost is paid on every attempt — without this, "no such account" would return
// measurably faster than "wrong password", letting a timing side-channel reveal which
// emails are registered.
var dummyPasswordHash = func() string {
	hash, err := bcrypt.GenerateFromPassword([]byte("porichoy-login-timing-safety"), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return string(hash)
}()
