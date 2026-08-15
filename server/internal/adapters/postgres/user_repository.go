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

// userModel is Bun's mapping of the users table (DATA_MODEL.md `users`) — kept separate from
// identity.User (CODING_STANDARDS.md §2).
type userModel struct {
	bun.BaseModel `bun:"table:users,alias:us"`

	ID                uuid.UUID  `bun:"id,pk,type:uuid"`
	TenantID          uuid.UUID  `bun:"tenant_id,type:uuid"`
	Status            string     `bun:"status"`
	DraftExpiresAt    *time.Time `bun:"draft_expires_at"`
	DisplayName       *string    `bun:"display_name"`
	ProfilePictureURL *string    `bun:"profile_picture_url"`
	Email             *string    `bun:"email"`
	EmailVerified     bool       `bun:"email_verified"`
	PendingEmail      *string    `bun:"pending_email"`
	Phone             *string    `bun:"phone"`
	PhoneVerified     bool       `bun:"phone_verified"`
	PendingPhone      *string    `bun:"pending_phone"`

	CreatedAt time.Time  `bun:"created_at"`
	UpdatedAt time.Time  `bun:"updated_at"`
	CreatedBy *uuid.UUID `bun:"created_by,type:uuid"`
	UpdatedBy *uuid.UUID `bun:"updated_by,type:uuid"`

	DeletedAt *time.Time `bun:"deleted_at"`
	DeletedBy *uuid.UUID `bun:"deleted_by,type:uuid"`
}

func userFromModel(m *userModel) *identity.User {
	return &identity.User{
		ID:                m.ID,
		TenantID:          m.TenantID,
		Status:            identity.UserStatus(m.Status),
		DraftExpiresAt:    m.DraftExpiresAt,
		DisplayName:       m.DisplayName,
		ProfilePictureURL: m.ProfilePictureURL,
		Email:             m.Email,
		EmailVerified:     m.EmailVerified,
		PendingEmail:      m.PendingEmail,
		Phone:             m.Phone,
		PhoneVerified:     m.PhoneVerified,
		PendingPhone:      m.PendingPhone,
		CreatedAt:         m.CreatedAt,
		UpdatedAt:         m.UpdatedAt,
		CreatedBy:         m.CreatedBy,
		UpdatedBy:         m.UpdatedBy,
		DeletedAt:         m.DeletedAt,
		DeletedBy:         m.DeletedBy,
	}
}

func userToModel(u *identity.User) *userModel {
	return &userModel{
		ID:                u.ID,
		TenantID:          u.TenantID,
		Status:            string(u.Status),
		DraftExpiresAt:    u.DraftExpiresAt,
		DisplayName:       u.DisplayName,
		ProfilePictureURL: u.ProfilePictureURL,
		Email:             u.Email,
		EmailVerified:     u.EmailVerified,
		PendingEmail:      u.PendingEmail,
		Phone:             u.Phone,
		PhoneVerified:     u.PhoneVerified,
		PendingPhone:      u.PendingPhone,
		CreatedAt:         u.CreatedAt,
		UpdatedAt:         u.UpdatedAt,
		CreatedBy:         u.CreatedBy,
		UpdatedBy:         u.UpdatedBy,
		DeletedAt:         u.DeletedAt,
		DeletedBy:         u.DeletedBy,
	}
}

// UserRepository implements identity.Repository using Postgres via Bun.
type UserRepository struct {
	db bun.IDB
}

// NewUserRepository builds a UserRepository from an open Bun connection.
func NewUserRepository(db *bun.DB) *UserRepository {
	return &UserRepository{db: db}
}

var _ identity.Repository = (*UserRepository)(nil)

// Create persists u (identity.Repository) — no actor.Actor, both call sites (Signup, the CLI
// seed's CreateRootUser) are pre-authentication. Participates in an ambient transaction if
// ctx carries one (tx.go).
func (r *UserRepository) Create(ctx context.Context, u *identity.User) error {
	_, err := dbFromContext(ctx, r.db).NewInsert().Model(userToModel(u)).Exec(ctx)
	return err
}

// FindByEmail is a pre-authentication, tenant-scoped lookup (identity.Repository) — backs
// Signup's duplicate-email check. Returns nil, nil if no active user has this email within
// tenantID.
func (r *UserRepository) FindByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*identity.User, error) {
	m := new(userModel)
	err := dbFromContext(ctx, r.db).NewSelect().Model(m).
		Where("tenant_id = ?", tenantID).
		Where("email = ?", email).
		Where("deleted_at IS NULL").
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return userFromModel(m), nil
}
