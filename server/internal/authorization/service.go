package authorization

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/aritradevelops/porichoy/server/internal/actor"
	"github.com/aritradevelops/porichoy/server/internal/apperror"
	"github.com/google/uuid"
)

// Service implements the authorization module's bootstrap use cases — role and
// role-assignment creation for the CLI seed command — resolving/caching a principal's
// effective permissions at login, and the runtime permission check itself
// (AUTHORIZATION_MODEL.md §2), which the real Authenticate middleware
// (internal/adapters/rest) calls on every authed request.
type Service struct {
	roles       RoleRepository
	assignments RoleAssignmentRepository
	cache       PermissionCache
}

// NewService wires a Service from its repository/port dependencies.
func NewService(roles RoleRepository, assignments RoleAssignmentRepository, cache PermissionCache) *Service {
	return &Service{roles: roles, assignments: assignments, cache: cache}
}

// CreateSystemRole creates a system-provisioned role (is_system=true) under tenantID, scoped
// to appID's business role catalog. Used by the CLI seed bootstrap
// (USER_JOURNEYS_ADMIN_TENANT_MANAGEMENT.md §1) — no actor.Actor, pre-authentication.
func (s *Service) CreateSystemRole(ctx context.Context, tenantID, appID uuid.UUID, name string, permissions []string) (*Role, error) {
	now := time.Now()
	r := &Role{
		ID:          uuid.New(),
		TenantID:    tenantID,
		AppID:       &appID,
		Name:        name,
		IsSystem:    true,
		Permissions: permissions,
		Policies:    []byte("[]"),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.roles.CreateSystem(ctx, r); err != nil {
		return nil, err
	}
	return r, nil
}

// AssignSystemRole binds principalID to roleID with no recorded granter (system
// auto-assignment) — used by the CLI seed bootstrap to grant the root superadmin Super
// Admin.
func (s *Service) AssignSystemRole(ctx context.Context, principalID, roleID uuid.UUID) (*RoleAssignment, error) {
	ra := &RoleAssignment{
		ID:          uuid.New(),
		PrincipalID: principalID,
		RoleID:      roleID,
		CreatedAt:   time.Now(),
	}
	if err := s.assignments.Create(ctx, ra); err != nil {
		return nil, err
	}
	return ra, nil
}

// EffectivePermissions returns the deduplicated union of every permission string granted to
// principalID across all its RoleAssignments — first-seen order (by RoleAssignment, then by
// each Role's own Permissions order), not sorted.
func (s *Service) EffectivePermissions(ctx context.Context, principalID uuid.UUID) ([]string, error) {
	assignments, err := s.assignments.ListByPrincipal(ctx, principalID)
	if err != nil {
		return nil, err
	}
	if len(assignments) == 0 {
		return []string{}, nil
	}

	roleIDs := make([]uuid.UUID, len(assignments))
	for i, a := range assignments {
		roleIDs[i] = a.RoleID
	}
	roles, err := s.roles.FindByIDs(ctx, roleIDs)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]struct{})
	permissions := []string{}
	for _, role := range roles {
		for _, p := range role.Permissions {
			if _, ok := seen[p]; ok {
				continue
			}
			seen[p] = struct{}{}
			permissions = append(permissions, p)
		}
	}
	return permissions, nil
}

// CacheUserPermissions resolves userID's EffectivePermissions and materializes them in the
// cache (TECHNICAL_DESIGN.md §6) under appID+userID, expiring after ttl. Called once at
// Login (internal/adapters/rest.AuthHandlers.Login) — not on every request, and not yet on
// every role/assignment change.
func (s *Service) CacheUserPermissions(ctx context.Context, appID, userID uuid.UUID, ttl time.Duration) error {
	permissions, err := s.EffectivePermissions(ctx, userID)
	if err != nil {
		return err
	}
	return s.cache.SetUserPermissions(ctx, appID, userID, permissions, ttl)
}

// ErrForbidden is returned by ResolveScope when no cached permission matches — either no
// cache entry exists at all (never logged in, or it expired), or one exists but nothing in
// it matches this module+action at any scope.
var ErrForbidden = apperror.New("authorization.forbidden", http.StatusForbidden)

// ResolveScope is the runtime permission check (AUTHORIZATION_MODEL.md §2): looks up
// userID's cached permissions for appID, finds every one matching
// {module}:{action}@* — including a {module}:*@* wildcard action — and returns the broadest
// matching scope (§3). Returns ErrForbidden if the cache has nothing for this principal, or
// nothing in it matches.
func (s *Service) ResolveScope(ctx context.Context, appID, userID uuid.UUID, module, action string) (actor.Scope, error) {
	raw, err := s.cache.GetUserPermissions(ctx, appID, userID)
	if err != nil {
		return "", err
	}
	if raw == nil {
		return "", ErrForbidden
	}

	var permissions []string
	if err := json.Unmarshal(raw, &permissions); err != nil {
		return "", err
	}

	var best actor.Scope
	matched := false
	for _, p := range permissions {
		permModule, permAction, scope, ok := parsePermission(p)
		if !ok || permModule != module || (permAction != action && permAction != "*") {
			continue
		}
		if !matched || actor.Scope(scope).AtLeast(best) {
			best = actor.Scope(scope)
			matched = true
		}
	}
	if !matched {
		return "", ErrForbidden
	}
	return best, nil
}

// parsePermission splits a "module:action@scope" permission string into its three parts —
// ok is false if either separator is missing.
func parsePermission(p string) (module, action, scope string, ok bool) {
	moduleAction, scope, ok := strings.Cut(p, "@")
	if !ok {
		return "", "", "", false
	}
	module, action, ok = strings.Cut(moduleAction, ":")
	return module, action, scope, ok
}
