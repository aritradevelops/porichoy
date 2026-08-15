package authorization

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// PermissionCache materializes a principal's effective permissions for fast lookup
// (TECHNICAL_DESIGN.md §6: "A user's effective permissions and policies are materialized in
// the cache provider... when roles/assignments change"). This pass only populates it at
// Login (internal/adapters/rest.AuthHandlers.Login); nothing reads it back yet, and nothing
// recomputes it on a role/assignment change — the runtime permission check itself is
// separate, later work. An interface rather than a concrete type so the default Redis
// implementation (internal/adapters/cache) can be swapped by a self-hoster
// (TECHNICAL_DESIGN.md §2).
type PermissionCache interface {
	// SetUserPermissions stores permissions (module:action@scope strings,
	// AUTHORIZATION_MODEL.md §3) for userID within tenantID, expiring after ttl.
	SetUserPermissions(ctx context.Context, tenantID, userID uuid.UUID, permissions []string, ttl time.Duration) error
}
