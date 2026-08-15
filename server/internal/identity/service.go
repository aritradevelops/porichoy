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

// Service implements the identity module's use cases — signup, login (email+password only;
// other tenant.LoginMethod values have no Service method yet), and the CLI seed's root-user
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

// AuthResult is the outcome of any flow that ends in an issued session — Signup and Login
// alike: the User plus its fresh token pair. RefreshToken is the raw value, returned exactly
// once; only its hash is persisted (app.NewRefreshToken).
type AuthResult struct {
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
func (s *Service) Signup(ctx context.Context, t *tenant.Tenant, email, password string) (*AuthResult, error) {
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

	accessToken, idToken, err := s.issueTokenPair(sysApp, t, u)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		User:                  u,
		AccessToken:           accessToken,
		IDToken:               idToken,
		RefreshToken:          rawRefresh,
		AccessTokenTTLSeconds: sysApp.AccessTokenTTLSeconds,
	}, nil
}

// Login authenticates an existing user by email+password against t's default system app and
// issues a fresh session + token pair — same token-issuance shape as Signup, but never
// creates a User. The email-not-found, wrong-password, and inactive-account cases all
// collapse to the same identity.invalid_credentials error so a caller can't distinguish
// "no such account" from "wrong password" or probe account status.
func (s *Service) Login(ctx context.Context, t *tenant.Tenant, email, password string) (*AuthResult, error) {
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

	u, err := s.users.FindByEmail(ctx, t.ID, email)
	if err != nil {
		return nil, err
	}

	// pwd stays nil for an unknown/inactive user — VerifyPassword still runs below, against
	// dummyPasswordHash, so this path costs the same as a real wrong-password attempt
	// (see dummyPasswordHash's doc comment).
	var pwd *Password
	if u != nil && u.Status == UserStatusActive {
		pwd, err = s.passwords.FindByUserID(ctx, u.ID)
		if err != nil {
			return nil, err
		}
	}
	compareHash := dummyPasswordHash
	if pwd != nil {
		compareHash = pwd.PasswordHash
	}
	if !VerifyPassword(compareHash, password) || pwd == nil {
		return nil, apperror.New("identity.invalid_credentials", http.StatusUnauthorized)
	}

	now := time.Now()
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
	if err := s.sessions.Create(ctx, sess); err != nil {
		return nil, err
	}

	accessToken, idToken, err := s.issueTokenPair(sysApp, t, u)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		User:                  u,
		AccessToken:           accessToken,
		IDToken:               idToken,
		RefreshToken:          rawRefresh,
		AccessTokenTTLSeconds: sysApp.AccessTokenTTLSeconds,
	}, nil
}

// issueTokenPair issues a fresh access + ID token for u against sysApp/t — the shared final
// step of Signup and Login, once a session already exists to back them.
func (s *Service) issueTokenPair(sysApp *app.App, t *tenant.Tenant, u *User) (accessToken, idToken string, err error) {
	claims := app.Claims{Subject: u.ID.String(), Audience: t.ID.String()}
	accessToken, err = s.tokens.Issue(sysApp, claims, time.Duration(sysApp.AccessTokenTTLSeconds)*time.Second)
	if err != nil {
		return "", "", err
	}
	idClaims := claims
	idClaims.Extra = map[string]any{"email": *u.Email, "email_verified": u.EmailVerified}
	idToken, err = s.tokens.Issue(sysApp, idClaims, time.Duration(sysApp.IDTokenTTLSeconds)*time.Second)
	if err != nil {
		return "", "", err
	}
	return accessToken, idToken, nil
}

// CreateRootUser creates the CLI seed's root superadmin — User + Password only, no
// session/tokens (they log in via the UI afterward, through Login). Skips
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
