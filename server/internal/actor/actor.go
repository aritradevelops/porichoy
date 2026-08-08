// Package actor carries the resolved, authorized caller of a request: who (a user or
// api_credential, undifferentiated per DATA_MODEL.md §0), within which tenant/app/org, and
// at what authorization scope (AUTHORIZATION_MODEL.md §2-3).
//
// It has no dependencies beyond uuid, and is imported by whichever bounded-context packages
// (CODING_STANDARDS.md §4) have authorized-only operations — it is not itself a bounded
// context (CODING_STANDARDS.md §2).
//
// An Actor is a required value, never a pointer: there is no anonymous or system Actor.
// It is built by the authorization middleware once that exists, and passed explicitly
// (never via context.Context.Value) only into Service/Repository methods that back an
// authorized route or MCP tool. Methods backing public or pre-authentication routes (e.g.
// tenant resolution by domain, TECHNICAL_DESIGN.md §3.3) take no Actor parameter at all.
// System-generated actions (no caller at all — e.g. the CLI seed command bootstrapping the
// root tenant, USER_JOURNEYS_ADMIN_TENANT_MANAGEMENT.md §1) get their own actor-less methods
// instead of a nil Actor.
package actor

import "github.com/google/uuid"

// Scope is the breadth of a permission, narrowest to broadest (AUTHORIZATION_MODEL.md §3).
type Scope string

const (
	ScopeOwn    Scope = "own"
	ScopeOrg    Scope = "org"
	ScopeApp    Scope = "app"
	ScopeTenant Scope = "tenant"
	ScopeRoot   Scope = "root"
)

var scopeRank = map[Scope]int{
	ScopeOwn:    0,
	ScopeOrg:    1,
	ScopeApp:    2,
	ScopeTenant: 3,
	ScopeRoot:   4,
}

// AtLeast reports whether s is at least as broad as other, per AUTHORIZATION_MODEL.md §3's
// fixed ordering (own < org < app < tenant < root).
func (s Scope) AtLeast(other Scope) bool {
	return scopeRank[s] >= scopeRank[other]
}

// Actor is the resolved caller for one authorized request or operation.
type Actor struct {
	// PrincipalID is the caller: a user or api_credential, undifferentiated
	// (DATA_MODEL.md §0). This is the value stored on created_by/updated_by/deleted_by
	// columns when this Actor performs a write.
	PrincipalID uuid.UUID
	// TenantID is resolved via the domain registry (TECHNICAL_DESIGN.md §3.3).
	TenantID uuid.UUID
	// AppID is nil outside an app-scoped call.
	AppID *uuid.UUID
	// OrgID is nil outside an org-scoped call.
	OrgID *uuid.UUID
	// Scope is the highest resolved scope for the module:action being performed
	// (AUTHORIZATION_MODEL.md §2-3).
	Scope Scope
}
