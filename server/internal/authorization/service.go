package authorization

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Service implements the authorization module's bootstrap use cases — role and
// role-assignment creation for the CLI seed command — plus resolving and caching a
// principal's effective permissions at login. It is not the runtime authorization engine
// (AUTHORIZATION_MODEL.md §2-3) — that doesn't exist yet; nothing reads the cache this
// populates.
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
// cache (TECHNICAL_DESIGN.md §6) under tenantID+userID, expiring after ttl. Called once at
// Login (internal/adapters/rest.AuthHandlers.Login) — not on every request, and not yet on
// every role/assignment change.
func (s *Service) CacheUserPermissions(ctx context.Context, tenantID, userID uuid.UUID, ttl time.Duration) error {
	permissions, err := s.EffectivePermissions(ctx, userID)
	if err != nil {
		return err
	}
	return s.cache.SetUserPermissions(ctx, tenantID, userID, permissions, ttl)
}
