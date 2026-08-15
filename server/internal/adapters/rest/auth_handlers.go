package rest

import (
	"net/http"

	"github.com/aritradevelops/porichoy/server/internal/identity"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// AuthHandlers holds the auth module's REST handlers (CODING_STANDARDS.md §5) — thin callers
// of identity.Service.
type AuthHandlers struct {
	svc *identity.Service
}

// NewAuthHandlers builds an AuthHandlers from its Service dependency.
func NewAuthHandlers(svc *identity.Service) *AuthHandlers {
	return &AuthHandlers{svc: svc}
}

// signupRequest is POST /api/v1/auth/signup's body.
type signupRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// signupResponse is Signup's response DTO — RefreshToken is the raw value, returned exactly
// once; only its hash is persisted server-side (identity.Service.Signup).
type signupResponse struct {
	UserID       uuid.UUID `json:"user_id"`
	Email        string    `json:"email"`
	AccessToken  string    `json:"access_token"`
	IDToken      string    `json:"id_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
}

func toSignupResponse(r *identity.SignupResult) signupResponse {
	return signupResponse{
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
// Authentication/Authorization: there's no caller to authenticate yet, this endpoint creates
// one), same precedent as domains/resolve.
func (h *AuthHandlers) Signup(c *fiber.Ctx) error {
	var req signupRequest
	if err := bindAndValidate(c, &req); err != nil {
		return fail(c, err)
	}

	result, err := h.svc.Signup(c.Context(), tenantFromLocals(c), req.Email, req.Password)
	if err != nil {
		return fail(c, err)
	}
	return success(c, http.StatusCreated, toSignupResponse(result))
}
