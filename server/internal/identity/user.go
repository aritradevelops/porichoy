// Package identity models user accounts and their passwords (DATA_MODEL.md §3), and
// implements signup (USER_JOURNEYS_ADMIN_TENANT_MANAGEMENT.md, the email+password path) and
// the CLI seed command's root-superadmin bootstrap.
package identity

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// UserStatus is a User's activation state (DATA_MODEL.md `users.status`).
type UserStatus string

const (
	// UserStatusDraft is an invited-but-not-yet-activated account (org-invite journey) —
	// not produced by anything in this pass.
	UserStatusDraft  UserStatus = "draft"
	UserStatusActive UserStatus = "active"
)

// User is an identity scoped to exactly one tenant (DATA_MODEL.md §3, PRD §4.2).
type User struct {
	ID       uuid.UUID
	TenantID uuid.UUID
	Status   UserStatus
	// DraftExpiresAt is only meaningful for UserStatusDraft — always nil this pass, every
	// user created here is UserStatusActive from the start.
	DraftExpiresAt    *time.Time
	DisplayName       *string
	ProfilePictureURL *string
	Email             *string
	EmailVerified     bool
	PendingEmail      *string
	Phone             *string
	PhoneVerified     bool
	PendingPhone      *string

	CreatedAt time.Time
	UpdatedAt time.Time
	CreatedBy *uuid.UUID
	UpdatedBy *uuid.UUID

	DeletedAt *time.Time
	DeletedBy *uuid.UUID
}

// IsDeleted reports whether u has been soft-deleted.
func (u *User) IsDeleted() bool {
	return u.DeletedAt != nil
}

// Repository persists and retrieves Users.
type Repository interface {
	// Create persists u. No actor.Actor — both call sites this pass (Signup, the CLI seed's
	// CreateRootUser) are pre-authentication.
	Create(ctx context.Context, u *User) error
	// FindByEmail is a pre-authentication, tenant-scoped lookup — used by Signup's
	// duplicate-email check, which runs before an actor.Actor can exist. Returns nil, nil if
	// no active user has this email within tenantID.
	FindByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*User, error)
}
