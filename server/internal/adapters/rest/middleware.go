package rest

import (
	"strings"

	"github.com/aritradevelops/porichoy/server/internal/actor"
	"github.com/aritradevelops/porichoy/server/internal/apperror"
	"github.com/aritradevelops/porichoy/server/internal/identity"
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

// bearerPrefix is the Authorization header scheme Authentication looks for before falling
// back to authCookieName.
const bearerPrefix = "Bearer "

// authCookieName is the fallback location for the access token when it isn't sent as a
// bearer header — no response sets this cookie yet (Signup/Login only return tokens in the
// JSON body), this is forward-looking for a future browser-based flow that does.
const authCookieName = "access_token"

// extractToken returns the raw token from c: the Authorization header's Bearer scheme
// first, falling back to the authCookieName cookie. Returns "" if neither is present.
func extractToken(c *fiber.Ctx) string {
	if h := c.Get(fiber.HeaderAuthorization); strings.HasPrefix(h, bearerPrefix) {
		return strings.TrimPrefix(h, bearerPrefix)
	}
	return c.Cookies(authCookieName)
}

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

// Authentication verifies the caller's access token — extractToken's Authorization-header-
// then-cookie precedence — against the tenant TenantResolution already resolved, and stashes
// the authenticated principal's UserID in Locals for Authorization to build the Actor from.
// Replaces the former X-Debug-Principal-ID stand-in wholesale (that stub's own doc comment
// said to, not extend it) now that Signup/Login (internal/identity.Service) issue real,
// verifiable tokens.
func Authentication(identitySvc *identity.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := extractToken(c)
		if token == "" {
			return fail(c, apperror.New("identity.unauthenticated", fiber.StatusUnauthorized))
		}

		principalID, err := identitySvc.Authenticate(c.Context(), tenantFromLocals(c), token)
		if err != nil {
			return fail(c, err)
		}
		c.Locals(localsPrincipalID, principalID)
		return c.Next()
	}
}

// Authorization is a placeholder for real permission checking — internal/authorization's
// Role/RoleAssignment tables exist and are populated (Signup, the CLI seed), but nothing
// queries them at request time yet, so AUTHORIZATION_MODEL.md §2 steps 1-2, looking up the
// caller's permissions, have nowhere real to look. It does not reject any request. What it
// DOES do for real: build the
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

// tenantFromLocals extracts the *tenant.Tenant TenantResolution resolved. Unlike
// actorFromLocals, this is usable behind TenantResolution alone — signup runs before any
// principal/actor exists, so it only ever has this much of the chain.
func tenantFromLocals(c *fiber.Ctx) *tenant.Tenant {
	return c.Locals(localsTenant).(*tenant.Tenant)
}
