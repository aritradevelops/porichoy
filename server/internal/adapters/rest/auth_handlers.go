package rest

import (
	"log"
	"net/http"
	"time"

	"github.com/aritradevelops/porichoy/server/internal/authorization"
	"github.com/aritradevelops/porichoy/server/internal/identity"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// AuthHandlers holds the auth module's REST handlers (CODING_STANDARDS.md §5) — thin callers
// of identity.Service, plus authorization.Service for caching a freshly logged-in user's
// permissions.
type AuthHandlers struct {
	svc   *identity.Service
	authz *authorization.Service
}

// NewAuthHandlers builds an AuthHandlers from its Service dependencies.
func NewAuthHandlers(svc *identity.Service, authz *authorization.Service) *AuthHandlers {
	return &AuthHandlers{svc: svc, authz: authz}
}

// signupRequest is POST /api/v1/auth/signup's body.
type signupRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// authResponse is the response DTO for both Signup and Login — RefreshToken is the raw
// value, returned exactly once; only its hash is persisted server-side
// (identity.Service.Signup/Login).
type authResponse struct {
	UserID       uuid.UUID `json:"user_id"`
	Email        string    `json:"email"`
	AccessToken  string    `json:"access_token"`
	IDToken      string    `json:"id_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
}

func toAuthResponse(r *identity.AuthResult) authResponse {
	return authResponse{
		UserID:       r.User.ID,
		Email:        *r.User.Email,
		AccessToken:  r.AccessToken,
		IDToken:      r.IDToken,
		RefreshToken: r.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    r.AccessTokenTTLSeconds,
	}
}

// Signup handles POST /api/v1/auth/signup — registered behind TenantResolution only (no
// Authenticate: there's no caller to authenticate yet, this endpoint creates one), same
// precedent as domains/resolve.
func (h *AuthHandlers) Signup(c *fiber.Ctx) error {
	var req signupRequest
	if err := bindAndValidate(c, &req); err != nil {
		return fail(c, err)
	}

	result, err := h.svc.Signup(c.Context(), tenantFromLocals(c), req.Email, req.Password)
	if err != nil {
		return fail(c, err)
	}
	return success(c, http.StatusCreated, toAuthResponse(result))
}

// loginRequest is POST /api/v1/auth/login's body — email+password only; the only method
// identity.Service.Login implements this pass.
type loginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// Login handles POST /api/v1/auth/login — same "TenantResolution only" reasoning as Signup:
// this endpoint is how a caller *becomes* authenticated, so there's no principal yet to run
// Authenticate against.
func (h *AuthHandlers) Login(c *fiber.Ctx) error {
	var req loginRequest
	if err := bindAndValidate(c, &req); err != nil {
		return fail(c, err)
	}

	t := tenantFromLocals(c)
	result, err := h.svc.Login(c.Context(), t, req.Email, req.Password)
	if err != nil {
		return fail(c, err)
	}

	// Materializes the user's effective permissions in the cache (TECHNICAL_DESIGN.md §6),
	// keyed by the same AppID Authenticate resolves at request time, TTL matched to the
	// access token this same response carries so the two expire together. Best-effort: a
	// cache-write failure is logged but doesn't fail an otherwise-valid login — the tradeoff
	// is that this user simply can't pass Authenticate's permission check again until they
	// log in a second time to repopulate it.
	ttl := time.Duration(result.AccessTokenTTLSeconds) * time.Second
	if err := h.authz.CacheUserPermissions(c.Context(), result.AppID, result.User.ID, ttl); err != nil {
		log.Printf("cache permissions for user %s: %v", result.User.ID, err)
	}

	return success(c, http.StatusOK, toAuthResponse(result))
}
