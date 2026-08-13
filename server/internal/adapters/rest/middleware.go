package rest

import (
	"github.com/aritradevelops/porichoy/server/internal/actor"
	"github.com/aritradevelops/porichoy/server/internal/apperror"
	"github.com/aritradevelops/porichoy/server/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// Fiber Locals keys. These are the only place an Actor (and the values it's built from)
// cross from Fiber-land into domain code (CODING_STANDARDS.md §5) — every handler that
// needs one extracts it immediately and passes it on as an explicit Go parameter.
const (
	localsTenant      = "tenant"
	localsPrincipalID = "principal_id"
	localsActor       = "actor"
)

// devPrincipalID is the fallback caller identity when no debug header is supplied — lets
// the stubbed auth chain (below) produce a usable Actor out of the box, not just when a
// caller remembers to pass headers.
var devPrincipalID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

// TenantResolution is the first stage of the chain (CODING_STANDARDS.md §5,
// AUTHORIZATION_MODEL.md's model): resolves which tenant the request's Host belongs to
// (TECHNICAL_DESIGN §3.3) via tenant.Service.ResolveTenantByDomain, and stashes it in
// Locals for Authorization to build the Actor's TenantID from. This is real, not stubbed —
// Service.ResolveTenantByDomain already exists and needs no auth/permission system to work.
// A request against an unregistered Host is rejected here, before authentication or
// authorization even run.
func TenantResolution(svc *tenant.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		t, err := svc.ResolveTenantByDomain(c.Context(), c.Hostname())
		if err != nil {
			return fail(c, err)
		}
		if t == nil {
			return fail(c, apperror.New("tenant.unresolved_domain", fiber.StatusNotFound))
		}
		c.Locals(localsTenant, t)
		return c.Next()
	}
}

// Authentication is a placeholder for real session/token verification
// (internal/identity — user login, session issuance — doesn't exist in this repo yet).
// For now it trusts an optional X-Debug-Principal-ID header, falling back to a fixed dev
// principal when absent, so the rest of the chain (and every downstream handler) has a
// real, working Actor to build without needing the identity subsystem to exist first.
// This is explicitly a stand-in, not a security boundary — replace wholesale once real
// authentication lands, don't extend it.
func Authentication() fiber.Handler {
	return func(c *fiber.Ctx) error {
		principalID := devPrincipalID
		if h := c.Get("X-Debug-Principal-ID"); h != "" {
			if id, err := uuid.Parse(h); err == nil {
				principalID = id
			}
		}
		c.Locals(localsPrincipalID, principalID)
		return c.Next()
	}
}

// Authorization is a placeholder for real permission checking
// (internal/authorization — Role, RoleAssignment — doesn't exist in this repo yet, so
// AUTHORIZATION_MODEL.md §2 steps 1-2, looking up the caller's permissions, have nothing
// to look up against). It does not reject any request. What it DOES do for real: build the
// actor.Actor every downstream Service/Repository call needs, from what TenantResolution
// and Authentication already resolved, plus an optional X-Debug-Scope header (defaults to
// actor.ScopeTenant — the most common real-world case) so different scope levels are
// exercisable without a real permission system. Replace wholesale, don't extend, once
// internal/authorization exists and real {module}:{action}@{scope} lookups are possible.
func Authorization() fiber.Handler {
	return func(c *fiber.Ctx) error {
		t, _ := c.Locals(localsTenant).(*tenant.Tenant)
		principalID, _ := c.Locals(localsPrincipalID).(uuid.UUID)

		scope := actor.ScopeTenant
		if h := c.Get("X-Debug-Scope"); h != "" {
			scope = actor.Scope(h)
		}

		c.Locals(localsActor, actor.Actor{
			PrincipalID: principalID,
			TenantID:    t.ID,
			Scope:       scope,
		})
		return c.Next()
	}
}

// actorFromLocals extracts the Actor Authorization built. Only ever called from within a
// route registered behind TenantResolution+Authentication+Authorization, so the type
// assertion is always expected to succeed — a handler wired up wrong (missing the chain)
// is a programming error, not a runtime condition to recover from gracefully.
func actorFromLocals(c *fiber.Ctx) actor.Actor {
	return c.Locals(localsActor).(actor.Actor)
}
