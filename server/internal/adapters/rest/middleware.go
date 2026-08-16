package rest

import (
	"strings"

	"github.com/aritradevelops/porichoy/server/internal/actor"
	"github.com/aritradevelops/porichoy/server/internal/apperror"
	"github.com/aritradevelops/porichoy/server/internal/authorization"
	"github.com/aritradevelops/porichoy/server/internal/identity"
	"github.com/aritradevelops/porichoy/server/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// Fiber Locals keys. These are the only place an Actor crosses from Fiber-land into domain
// code (CODING_STANDARDS.md §5) — every handler that needs one extracts it immediately and
// passes it on as an explicit Go parameter.
const (
	localsTenant = "tenant"
	localsActor  = "actor"
)

// bearerPrefix is the Authorization header scheme Authenticate looks for before falling
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
// Locals for Authenticate to build the Actor's TenantID from. This is real, not stubbed —
// Service.ResolveTenantByDomain already exists and needs no auth/permission system to work.
// A request against an unregistered Host is rejected here, before authentication or
// authorization even run.
func TenantResolution(svc *tenant.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		t, err := svc.ResolveTenantByDomain(c.Context(), hostnameWithoutPort(c))
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

// hostnameWithoutPort returns c.Hostname() with any ":port" suffix stripped —
// fiber.Ctx.Hostname() returns the raw Host header value verbatim (fasthttp's
// URI().Host()), port included when the client sent one, so a request to
// "admin.example.com:5173" would otherwise fail to match the registered domain
// "admin.example.com". Domains in this system are DNS hostnames, never IP literals, so a
// plain split on the last colon is safe (no IPv6 bracket-notation to account for).
func hostnameWithoutPort(c *fiber.Ctx) string {
	host := c.Hostname()
	if i := strings.LastIndexByte(host, ':'); i != -1 {
		return host[:i]
	}
	return host
}

// Authenticate is the merged authentication+authorization stage — this repo's real
// implementation of what the legacy Node service's authMiddleware
// (application/src/middlewares/auth.middleware.ts) does in one pass, and the wholesale
// replacement AUTHORIZATION_MODEL.md §2 and the former separate Authentication/Authorization
// stubs' own doc comments both called for: verify the caller's token, then resolve their
// cached effective permissions for this route's {module}:{action} and build the Actor from
// the broadest matching scope. Runs behind TenantResolution.
//
//  1. Verify: extractToken's Authorization-header-then-cookie precedence, against the tenant
//     TenantResolution already resolved (identitySvc.Authenticate) — 401
//     identity.unauthenticated if missing/invalid.
//  2. Extract {module}/{action} from the route path (AUTHORIZATION_MODEL.md §1).
//  3. Look up the principal's permissions cached at Login, keyed by the app Authenticate
//     resolved (not the debug header this replaces), and resolve the broadest scope matching
//     this module+action (authzSvc.ResolveScope) — 403 authorization.forbidden if none does.
//  4. Build the actor.Actor every downstream Service/Repository call needs.
func Authenticate(identitySvc *identity.Service, authzSvc *authorization.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		principalID, appID, err := verifyToken(c, identitySvc)
		if err != nil {
			return fail(c, err)
		}

		module, action, ok := moduleActionFromPath(c.Path())
		if !ok {
			return fail(c, authorization.ErrForbidden)
		}
		scope, err := authzSvc.ResolveScope(c.Context(), appID, principalID, module, action)
		if err != nil {
			return fail(c, err)
		}

		c.Locals(localsActor, actor.Actor{
			PrincipalID: principalID,
			TenantID:    tenantFromLocals(c).ID,
			AppID:       &appID,
			Scope:       scope,
		})
		return c.Next()
	}
}

// verifyToken extracts and verifies the caller's token — the shared first step of Authenticate
// and AuthenticateOnly (identity.Service.Authenticate itself, against the tenant
// TenantResolution already resolved).
func verifyToken(c *fiber.Ctx, identitySvc *identity.Service) (principalID, appID uuid.UUID, err error) {
	token := extractToken(c)
	if token == "" {
		return uuid.Nil, uuid.Nil, apperror.New("identity.unauthenticated", fiber.StatusUnauthorized)
	}
	return identitySvc.Authenticate(c.Context(), tenantFromLocals(c), token)
}

// AuthenticateOnly verifies the caller's token (verifyToken, same as Authenticate) but runs no
// {module}:{action}@{scope} permission check — for routes that are authenticated-only, not a
// permission-gated business operation. Today that's exactly one route, GET /api/v1/auth/me:
// reading your own identity/permission list can't be permission-gated without a chicken-and-egg
// problem (checking a permission to learn what permissions you have), and running it through
// ResolveScope would 403 a validly authenticated caller who simply holds no grants yet —
// precisely the state that endpoint exists to report truthfully.
//
// The Actor this builds has Scope hardcoded to actor.ScopeOwn — not a resolved grant, just
// documents that whatever a route behind this middleware does, it acts on the caller's own
// record only (identitySvc.Authenticate resolves principalID off the verified token's own
// subject claim, never a request-supplied parameter).
func AuthenticateOnly(identitySvc *identity.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		principalID, appID, err := verifyToken(c, identitySvc)
		if err != nil {
			return fail(c, err)
		}

		c.Locals(localsActor, actor.Actor{
			PrincipalID: principalID,
			TenantID:    tenantFromLocals(c).ID,
			AppID:       &appID,
			Scope:       actor.ScopeOwn,
		})
		return c.Next()
	}
}

// moduleActionFromPath extracts {module} and {action} from a route path shaped
// /api/{version}/{module}/{action}[/...] (AUTHORIZATION_MODEL.md §1) — ok is false if path
// has too few segments to contain both (a programming error in route registration, not a
// real request shape, since every authed route is registered by this codebase itself).
func moduleActionFromPath(path string) (module, action string, ok bool) {
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) < 4 {
		return "", "", false
	}
	return parts[2], parts[3], true
}

// actorFromLocals extracts the Actor Authenticate built. Only ever called from within a
// route registered behind TenantResolution+Authenticate, so the type assertion is always
// expected to succeed — a handler wired up wrong (missing the chain) is a programming error,
// not a runtime condition to recover from gracefully.
func actorFromLocals(c *fiber.Ctx) actor.Actor {
	return c.Locals(localsActor).(actor.Actor)
}

// tenantFromLocals extracts the *tenant.Tenant TenantResolution resolved. Unlike
// actorFromLocals, this is usable behind TenantResolution alone — signup runs before any
// principal/actor exists, so it only ever has this much of the chain.
func tenantFromLocals(c *fiber.Ctx) *tenant.Tenant {
	return c.Locals(localsTenant).(*tenant.Tenant)
}
