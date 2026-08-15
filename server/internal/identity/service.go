package identity

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/aritradevelops/porichoy/server/internal/app"
	"github.com/aritradevelops/porichoy/server/internal/apperror"
	"github.com/aritradevelops/porichoy/server/internal/authorization"
	"github.com/aritradevelops/porichoy/server/internal/tenant"
	"github.com/google/uuid"
)

// Service implements the identity module's use cases — signup and the CLI seed's root-user
// bootstrap. It composes sibling contexts' repository interfaces directly
// (CODING_STANDARDS.md §3: a Service may depend on sibling contexts' repository ports, never
// their Services), rather than introducing a separate orchestration layer.
type Service struct {
	users           Repository
	passwords       PasswordRepository
	apps            app.Repository
	sessions        app.SessionRepository
	roleAssignments authorization.RoleAssignmentRepository
	tokens          app.TokenIssuer
	tx              TxRunner
}

// NewService wires a Service from its repository/port dependencies.
func NewService(
	users Repository,
	passwords PasswordRepository,
	apps app.Repository,
	sessions app.SessionRepository,
	roleAssignments authorization.RoleAssignmentRepository,
	tokens app.TokenIssuer,
	tx TxRunner,
) *Service {
	return &Service{
		users:           users,
		passwords:       passwords,
		apps:            apps,
		sessions:        sessions,
		roleAssignments: roleAssignments,
		tokens:          tokens,
		tx:              tx,
	}
}

// SignupResult is Signup's outcome — the created User plus its first issued token pair.
// RefreshToken is the raw value, returned exactly once; only its hash is persisted
// (app.NewRefreshToken).
type SignupResult struct {
	User                  *User
	AccessToken           string
	IDToken               string
	RefreshToken          string
	AccessTokenTTLSeconds int
}

// Signup creates a new user under t via email+password, against t's default system app
// (TECHNICAL_DESIGN §3.5), and issues its first session + token pair. No actor.Actor —
// pre-authentication; t is the tenant TenantResolution middleware already resolved (which
// already excludes a soft-deleted tenant, so Signup doesn't re-check that itself).
func (s *Service) Signup(ctx context.Context, t *tenant.Tenant, email, password string) (*SignupResult, error) {
	email = normalizeEmail(email)
	if !hasEmailPasswordMethod(t) {
		return nil, apperror.New("identity.login_method_disabled", http.StatusBadRequest)
	}

	sysApp, err := s.apps.FindSystemAppByTenant(ctx, t.ID)
	if err != nil {
		return nil, err
	}
	if sysApp == nil {
		return nil, apperror.New("identity.system_app_not_found", http.StatusInternalServerError)
	}

	existing, err := s.users.FindByEmail(ctx, t.ID, email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, apperror.New("identity.email_already_registered", http.StatusConflict)
	}

	hash, err := HashPassword(password)
	if err != nil {
		if errors.Is(err, ErrPasswordTooLong) {
			return nil, apperror.New("identity.password_too_long", http.StatusBadRequest)
		}
		return nil, err
	}

	now := time.Now()
	u := &User{
		ID:            uuid.New(),
		TenantID:      t.ID,
		Status:        UserStatusActive,
		Email:         &email,
		EmailVerified: false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	rawRefresh, refreshHash, err := app.NewRefreshToken()
	if err != nil {
		return nil, err
	}
	sess := &app.Session{
		ID:               uuid.New(),
		UserID:           u.ID,
		AppID:            sysApp.ID,
		Aud:              t.ID.String(),
		RefreshTokenHash: refreshHash,
		CreatedAt:        now,
		LastActiveAt:     now,
		ExpiresAt:        now.Add(time.Duration(sysApp.RefreshTokenTTLSeconds) * time.Second),
	}

	err = s.tx.RunInTx(ctx, func(ctx context.Context) error {
		if err := s.users.Create(ctx, u); err != nil {
			return err
		}
		if err := s.passwords.Create(ctx, &Password{
			ID:           uuid.New(),
			UserID:       u.ID,
			PasswordHash: hash,
			CreatedAt:    now,
		}); err != nil {
			return err
		}
		if sysApp.DefaultSignupRoleID != nil {
			if err := s.roleAssignments.Create(ctx, &authorization.RoleAssignment{
				ID:          uuid.New(),
				PrincipalID: u.ID,
				RoleID:      *sysApp.DefaultSignupRoleID,
				CreatedAt:   now,
			}); err != nil {
				return err
			}
		}
		return s.sessions.Create(ctx, sess)
	})
	if err != nil {
		return nil, err
	}

	claims := app.Claims{Subject: u.ID.String(), Audience: t.ID.String()}
	accessToken, err := s.tokens.Issue(sysApp, claims, time.Duration(sysApp.AccessTokenTTLSeconds)*time.Second)
	if err != nil {
		return nil, err
	}
	idClaims := claims
	idClaims.Extra = map[string]any{"email": email, "email_verified": u.EmailVerified}
	idToken, err := s.tokens.Issue(sysApp, idClaims, time.Duration(sysApp.IDTokenTTLSeconds)*time.Second)
	if err != nil {
		return nil, err
	}

	return &SignupResult{
		User:                  u,
		AccessToken:           accessToken,
		IDToken:               idToken,
		RefreshToken:          rawRefresh,
		AccessTokenTTLSeconds: sysApp.AccessTokenTTLSeconds,
	}, nil
}

// CreateRootUser creates the CLI seed's root superadmin — User + Password only, no
// session/tokens (they log in via the UI afterward, once a login endpoint exists). Skips
// Signup's enabled-login-methods check since bootstrap predates tenant configuration.
// EmailVerified is true — operator-entered via the seed CLI, no verification flow to run
// anyway. No actor.Actor — mirrors tenant.Service.CreateRootTenant.
func (s *Service) CreateRootUser(ctx context.Context, tenantID uuid.UUID, email, password string) (*User, error) {
	email = normalizeEmail(email)
	hash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	u := &User{
		ID:            uuid.New(),
		TenantID:      tenantID,
		Status:        UserStatusActive,
		Email:         &email,
		EmailVerified: true,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	err = s.tx.RunInTx(ctx, func(ctx context.Context) error {
		if err := s.users.Create(ctx, u); err != nil {
			return err
		}
		return s.passwords.Create(ctx, &Password{
			ID:           uuid.New(),
			UserID:       u.ID,
			PasswordHash: hash,
			CreatedAt:    now,
		})
	})
	if err != nil {
		return nil, err
	}
	return u, nil
}

func hasEmailPasswordMethod(t *tenant.Tenant) bool {
	return slices.Contains(t.EnabledLoginMethods, tenant.LoginMethodEmailPassword)
}

// normalizeEmail lowercases and trims email once, at the single point every path (Signup,
// CreateRootUser) enters the system — so the duplicate-email lookup, the persisted value, and
// the partial unique index (idx_users_tenant_email) all agree on the same canonical form.
// Without this, "user@example.com" and "User@Example.com " would be treated as distinct
// accounts by every layer.
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
