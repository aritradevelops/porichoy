package rest

import (
	"github.com/aritradevelops/porichoy/server/internal/tenant"
	"github.com/gofiber/fiber/v2"
)

// New builds the Fiber app and registers every route (CODING_STANDARDS.md §5 — routes are
// registered explicitly, one per module/action, no dynamic dispatch). Takes tenant.Service
// directly rather than a Postgres-specific type, so this adapter stays swappable
// (CODING_STANDARDS.md §2) — cmd/server/main.go is the only place that knows Postgres is
// the backing store.
func New(tenantSvc *tenant.Service) *fiber.App {
	app := fiber.New()

	tenants := NewTenantHandlers(tenantSvc)
	domains := NewDomainHandlers(tenantSvc)

	api := app.Group("/api/v1")

	// Public/pre-authentication routes — don't run the TenantResolution/Authentication/
	// Authorization chain (CODING_STANDARDS.md §5's own example of this: "tenant
	// resolution itself"). domains/resolve is the same shape of route: nothing to resolve
	// a tenant *from* yet, since the caller is asking which tenant a domain belongs to in
	// the first place.
	api.Get("/domains/resolve", domains.Resolve)

	// Authenticated/authorized routes — full three-stage chain (CODING_STANDARDS.md §5).
	authed := api.Group("", TenantResolution(tenantSvc), Authentication(), Authorization())
	authed.Post("/tenants/create", tenants.Create)
	authed.Get("/tenants/get/:id", tenants.Get)
	authed.Post("/tenants/configure/:id", tenants.Configure)
	authed.Post("/domains/register", domains.Register)

	return app
}
