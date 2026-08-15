package rest

import (
	"net/http"
	"time"

	"github.com/aritradevelops/porichoy/server/internal/apperror"
	"github.com/aritradevelops/porichoy/server/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// TenantHandlers holds the tenant module's REST handlers (CODING_STANDARDS.md §5) — thin
// callers of tenant.Service, one exported method per route.
type TenantHandlers struct {
	svc *tenant.Service
}

// NewTenantHandlers builds a TenantHandlers from its Service dependency.
func NewTenantHandlers(svc *tenant.Service) *TenantHandlers {
	return &TenantHandlers{svc: svc}
}

// tenantResponse is the tenants module's response DTO, snake_case per this adapter's field
// casing convention. Mirrors tenant.Tenant field-for-field except DeletedAt/DeletedBy,
// which a caller reaching this handler (a non-deleted row, per every Repository method's
// own `deleted_at IS NULL` filter) never has a reason to see.
type tenantResponse struct {
	ID                  uuid.UUID  `json:"id"`
	ParentID            *uuid.UUID `json:"parent_id"`
	Name                string     `json:"name"`
	LogoURL             string     `json:"logo_url"`
	BrandImageURL       *string    `json:"brand_image_url"`
	LoginLayout         string     `json:"login_layout"`
	MFARequired         bool       `json:"mfa_required"`
	EnabledLoginMethods []string   `json:"enabled_login_methods"`
	AuditRetentionDays  int        `json:"audit_retention_days"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

func toTenantResponse(t *tenant.Tenant) tenantResponse {
	methods := make([]string, len(t.EnabledLoginMethods))
	for i, m := range t.EnabledLoginMethods {
		methods[i] = string(m)
	}
	return tenantResponse{
		ID:                  t.ID,
		ParentID:            t.ParentID,
		Name:                t.Name,
		LogoURL:             t.LogoURL,
		BrandImageURL:       t.BrandImageURL,
		LoginLayout:         string(t.LoginLayout),
		MFARequired:         t.MFARequired,
		EnabledLoginMethods: methods,
		AuditRetentionDays:  t.AuditRetentionDays,
		CreatedAt:           t.CreatedAt,
		UpdatedAt:           t.UpdatedAt,
	}
}

// createTenantRequest is POST /api/v1/tenants/create's body. ParentID nil means "under the
// caller's own tenant" — tenant.Service.CreateTenant's own default (USER_JOURNEYS_
// ADMIN_TENANT_MANAGEMENT.md §2, MCP_TOOLS.md tenant_create), not re-implemented here.
type createTenantRequest struct {
	Name     string     `json:"name" validate:"required"`
	ParentID *uuid.UUID `json:"parent_id"`
}

// Create handles POST /api/v1/tenants/create (permission tenants:create).
func (h *TenantHandlers) Create(c *fiber.Ctx) error {
	var req createTenantRequest
	if err := bindAndValidate(c, &req); err != nil {
		return fail(c, err)
	}

	t, err := h.svc.CreateTenant(c.Context(), actorFromLocals(c), req.Name, req.ParentID)
	if err != nil {
		return fail(c, err)
	}
	return success(c, http.StatusCreated, toTenantResponse(t))
}

// Get handles GET /api/v1/tenants/get/:id (permission tenants:get). tenant.Service.GetTenant
// (and the exact-match scope filter beneath it, AUTHORIZATION_MODEL.md §4) returns nil, nil
// rather than an error for both "doesn't exist" and "exists but isn't yours" — this handler
// is where that becomes a 404 apperror, the one place in the chain that has to decide it's
// actually an error at all.
func (h *TenantHandlers) Get(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fail(c, apperror.New("validation.invalid_body", http.StatusBadRequest))
	}

	t, err := h.svc.GetTenant(c.Context(), actorFromLocals(c), id)
	if err != nil {
		return fail(c, err)
	}
	if t == nil {
		return fail(c, apperror.New("tenant.not_found", http.StatusNotFound))
	}
	return success(c, http.StatusOK, toTenantResponse(t))
}

// tenantListResponse is GET /api/v1/tenants/list's response — Total/Page/PageSize let a
// caller compute page count (ceil(Total / PageSize)) without a separate count request.
type tenantListResponse struct {
	Items    []tenantResponse `json:"items"`
	Total    int              `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
}

// List handles GET /api/v1/tenants/list?page=&page_size= (permission tenants:list) — one
// page of every tenant the caller is authorized to see (root: all; tenant scope: itself plus
// its subtree), backing the root-admin "all tenants" table
// (USER_JOURNEYS_ADMIN_TENANT_MANAGEMENT.md §2). Both query params are optional —
// tenant.Service.ListTenants normalizes missing/out-of-range values, this handler just parses
// whatever's given (or fiber's own zero-value default) straight through.
func (h *TenantHandlers) List(c *fiber.Ctx) error {
	params := tenant.ListParams{
		Page:     c.QueryInt("page", 1),
		PageSize: c.QueryInt("page_size", 0),
	}
	result, err := h.svc.ListTenants(c.Context(), actorFromLocals(c), params)
	if err != nil {
		return fail(c, err)
	}
	items := make([]tenantResponse, len(result.Items))
	for i, t := range result.Items {
		items[i] = toTenantResponse(t)
	}
	return success(c, http.StatusOK, tenantListResponse{
		Items:    items,
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
	})
}

// configureTenantRequest is POST /api/v1/tenants/configure/:id's body — mirrors
// tenant.Config field-for-field: every field is a pointer, nil means "leave unchanged"
// (partial update, USER_JOURNEYS_ADMIN_TENANT_MANAGEMENT.md §3), except
// EnabledLoginMethods, which (like tenant.Config's own field) uses nil-slice-vs-non-nil to
// mean the same thing since a slice has no separate pointer form.
type configureTenantRequest struct {
	LogoURL             *string  `json:"logo_url"`
	BrandImageURL       *string  `json:"brand_image_url"`
	LoginLayout         *string  `json:"login_layout" validate:"omitempty,oneof=centered split"`
	MFARequired         *bool    `json:"mfa_required"`
	EnabledLoginMethods []string `json:"enabled_login_methods" validate:"omitempty,dive,oneof=email_password email_otp phone_otp webauthn google apple"`
	AuditRetentionDays  *int     `json:"audit_retention_days" validate:"omitempty,min=1"`
}

// Configure handles POST /api/v1/tenants/configure/:id (permission tenants:configure).
func (h *TenantHandlers) Configure(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fail(c, apperror.New("validation.invalid_body", http.StatusBadRequest))
	}

	var req configureTenantRequest
	if err := bindAndValidate(c, &req); err != nil {
		return fail(c, err)
	}

	cfg := tenant.Config{
		LogoURL:            req.LogoURL,
		BrandImageURL:      req.BrandImageURL,
		MFARequired:        req.MFARequired,
		AuditRetentionDays: req.AuditRetentionDays,
	}
	if req.LoginLayout != nil {
		layout := tenant.LoginLayout(*req.LoginLayout)
		cfg.LoginLayout = &layout
	}
	if req.EnabledLoginMethods != nil {
		methods := make([]tenant.LoginMethod, len(req.EnabledLoginMethods))
		for i, m := range req.EnabledLoginMethods {
			methods[i] = tenant.LoginMethod(m)
		}
		cfg.EnabledLoginMethods = methods
	}

	t, err := h.svc.ConfigureTenant(c.Context(), actorFromLocals(c), id, cfg)
	if err != nil {
		return fail(c, err)
	}
	return success(c, http.StatusOK, toTenantResponse(t))
}
