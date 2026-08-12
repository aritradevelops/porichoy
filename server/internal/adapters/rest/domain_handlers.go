package rest

import (
	"net/http"
	"time"

	"github.com/aritradevelops/porichoy/server/internal/actor"
	"github.com/aritradevelops/porichoy/server/internal/apperror"
	"github.com/aritradevelops/porichoy/server/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// DomainHandlers holds the domains module's REST handlers (CODING_STANDARDS.md §5) —
// deliberately its own module (not folded under tenants), so domain registration gets its
// own {module}:{action}@{scope} permission namespace (AUTHORIZATION_MODEL.md §1) separate
// from tenant configuration itself.
type DomainHandlers struct {
	svc *tenant.Service
}

// NewDomainHandlers builds a DomainHandlers from its Service dependency.
func NewDomainHandlers(svc *tenant.Service) *DomainHandlers {
	return &DomainHandlers{svc: svc}
}

type domainResponse struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenant_id"`
	Domain    string    `json:"domain"`
	CreatedAt time.Time `json:"created_at"`
}

func toDomainResponse(d *tenant.TenantDomain) domainResponse {
	return domainResponse{ID: d.ID, TenantID: d.TenantID, Domain: d.Domain, CreatedAt: d.CreatedAt}
}

// registerDomainRequest is POST /api/v1/domains/register's body. TenantID is required
// (tenant.Service.RegisterDomain takes it as an explicit parameter, not implied from the
// Actor) — Register below enforces it can only ever be the caller's own tenant, short of
// root scope (the same exact-match philosophy as the tenants module,
// AUTHORIZATION_MODEL.md §4).
type registerDomainRequest struct {
	TenantID uuid.UUID `json:"tenant_id" validate:"required"`
	Domain   string    `json:"domain" validate:"required"`
}

// Register handles POST /api/v1/domains/register (permission domains:register). Unlike
// the tenants module's GetByID/ListChildren, tenant.Service.RegisterDomain itself performs
// no scope check (it trusts its caller) — so this handler is where exact-match enforcement
// actually happens for this one action, mirroring the repository-layer filter the tenants
// module uses for reads.
func (h *DomainHandlers) Register(c *fiber.Ctx) error {
	var req registerDomainRequest
	if err := bindAndValidate(c, &req); err != nil {
		return fail(c, err)
	}

	act := actorFromLocals(c)
	if act.Scope != actor.ScopeRoot && req.TenantID != act.TenantID {
		return fail(c, apperror.New("domain.forbidden", http.StatusForbidden))
	}

	d, err := h.svc.RegisterDomain(c.Context(), act, req.TenantID, req.Domain)
	if err != nil {
		return fail(c, err)
	}
	return success(c, http.StatusCreated, toDomainResponse(d))
}

type resolveDomainResponse struct {
	TenantID uuid.UUID `json:"tenant_id"`
	Name     string    `json:"name"`
}

// Resolve handles GET /api/v1/domains/resolve?domain=... — public, actor-less, registered
// outside the TenantResolution/Authentication/Authorization chain (CODING_STANDARDS.md §5's
// own example of a route that doesn't run it). Resolves an arbitrary domain given as a
// query param to its tenant; unrelated to (and doesn't run) TenantResolution's lookup of
// the *current* request's own Host.
func (h *DomainHandlers) Resolve(c *fiber.Ctx) error {
	domain := c.Query("domain")
	if domain == "" {
		return fail(c, apperror.New("validation.failed", http.StatusBadRequest))
	}

	t, err := h.svc.ResolveTenantByDomain(c.Context(), domain)
	if err != nil {
		return fail(c, err)
	}
	if t == nil {
		return fail(c, apperror.New("domain.not_found", http.StatusNotFound))
	}
	return success(c, http.StatusOK, resolveDomainResponse{TenantID: t.ID, Name: t.Name})
}
