package authorization

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// PermissionCache materializes a principal's effective permissions, keyed by the app they
// were resolved within — an app owns its own role/permission catalog (DATA_MODEL.md `roles`,
// TECHNICAL_DESIGN.md §3.5) — for fast lookup at request time (TECHNICAL_DESIGN.md §6).
// SetUserPermissions populates it at Login; GetUserPermissions backs the real runtime
// permission check (AUTHORIZATION_MODEL.md §2, internal/adapters/rest.Authenticate). An
// interface rather than a concrete type so the default Redis implementation
// (internal/adapters/cache) can be swapped by a self-hoster (TECHNICAL_DESIGN.md §2).
type PermissionCache interface {
	// SetUserPermissions stores permissions (module:action@scope strings,
	// AUTHORIZATION_MODEL.md §3) for userID within appID, expiring after ttl.
	SetUserPermissions(ctx context.Context, appID, userID uuid.UUID, permissions []string, ttl time.Duration) error
	// GetUserPermissions returns the raw JSON-array-encoded permissions previously stored by
	// SetUserPermissions for appID+userID, or nil, nil if no entry exists (never logged in,
	// or the entry expired) — a miss is "no permissions", not an error; the caller
	// (Service.ResolveScope) is the one that turns that into a 403.
	GetUserPermissions(ctx context.Context, appID, userID uuid.UUID) ([]byte, error)
}
