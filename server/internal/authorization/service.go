package authorization

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Service implements the authorization module's bootstrap use cases — role and
// role-assignment creation for the CLI seed command. It is not the runtime authorization
// engine (AUTHORIZATION_MODEL.md §2-3) — that doesn't exist yet.
type Service struct {
	roles       RoleRepository
	assignments RoleAssignmentRepository
}

// NewService wires a Service from its repository dependencies.
func NewService(roles RoleRepository, assignments RoleAssignmentRepository) *Service {
	return &Service{roles: roles, assignments: assignments}
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
