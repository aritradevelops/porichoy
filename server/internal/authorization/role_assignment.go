package authorization

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// RoleAssignment binds a principal (a user or api_credential, undifferentiated,
// DATA_MODEL.md §0) to a Role — scope isn't stored here, it lives on each permission string
// within the role itself (AUTHORIZATION_MODEL.md §3, DATA_MODEL.md `role_assignments`). OrgID
// isn't modeled yet — organizations don't exist anywhere else in this codebase either.
type RoleAssignment struct {
	ID          uuid.UUID
	PrincipalID uuid.UUID
	RoleID      uuid.UUID

	CreatedAt time.Time
	CreatedBy *uuid.UUID

	DeletedAt *time.Time
	DeletedBy *uuid.UUID
}

// IsDeleted reports whether ra has been unassigned.
func (ra *RoleAssignment) IsDeleted() bool {
	return ra.DeletedAt != nil
}

// RoleAssignmentRepository persists RoleAssignments.
type RoleAssignmentRepository interface {
	// Create binds principalID to roleID. No actor.Actor: used both by the CLI seed
	// bootstrap and by identity.Service.Signup (pre-auth, granting the resolved app's
	// default_signup_role_id) — CreatedBy is nil in both cases (system auto-assignment,
	// DATA_MODEL.md `role_assignments.created_by`).
	Create(ctx context.Context, ra *RoleAssignment) error
	// ListByPrincipal returns every non-deleted RoleAssignment for principalID — backs
	// Service.EffectivePermissions.
	ListByPrincipal(ctx context.Context, principalID uuid.UUID) ([]*RoleAssignment, error)
}
