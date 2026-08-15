// Package authorization models roles and role assignments (DATA_MODEL.md §5) and implements
// the bootstrap use cases the CLI seed command needs (USER_JOURNEYS_ADMIN_TENANT_MANAGEMENT.md
// §1) — creating system roles and granting them. Real permission-check/scope-resolution logic
// (AUTHORIZATION_MODEL.md §2-3) is a separate, later piece of work; nothing here enforces
// anything yet.
package authorization

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Role carries a set of permissions and/or policies (AUTHORIZATION_MODEL.md §5), embedded
// directly as arrays rather than normalized into child tables (DATA_MODEL.md `roles`) — a
// role edit is a single-row update, and the runtime authorization check reads from a
// precomputed cache, never this table directly.
type Role struct {
	ID       uuid.UUID
	TenantID uuid.UUID
	// AppID is nil for an org-management system role (e.g. Owner); set for that app's own
	// business role catalog (PRD §7.2). OrgID isn't modeled yet — organizations don't exist
	// anywhere else in this codebase either.
	AppID       *uuid.UUID
	Name        string
	IsSystem    bool
	Permissions []string
	Policies    json.RawMessage

	CreatedAt time.Time
	UpdatedAt time.Time
	CreatedBy *uuid.UUID
	UpdatedBy *uuid.UUID

	DeletedAt *time.Time
	DeletedBy *uuid.UUID
}

// IsDeleted reports whether r has been soft-deleted.
func (r *Role) IsDeleted() bool {
	return r.DeletedAt != nil
}

// RoleRepository persists and retrieves Roles.
type RoleRepository interface {
	// CreateSystem persists a system-provisioned role (is_system=true) — the CLI seed's
	// bootstrap path, no actor.Actor since it runs before any principal exists.
	CreateSystem(ctx context.Context, r *Role) error
}
