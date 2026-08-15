package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
)

// Service implements the app module's bootstrap use cases — system app creation and
// configuration for the CLI seed command.
type Service struct {
	apps Repository
}

// NewService wires a Service from its repository dependency.
func NewService(apps Repository) *Service {
	return &Service{apps: apps}
}

// CreateSystemApp creates tenantID's default system app (TECHNICAL_DESIGN §3.5) — a random
// HS256 signing secret and sane default token TTLs, is_system=true. Used by the CLI seed
// bootstrap, no actor.Actor.
func (s *Service) CreateSystemApp(ctx context.Context, tenantID uuid.UUID, name string) (*App, error) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return nil, err
	}
	clientID, err := randomHex(16)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	a := &App{
		ID:                     uuid.New(),
		TenantID:               tenantID,
		Name:                   name,
		ClientID:               clientID,
		IsSystem:               true,
		SigningAlgorithm:       SigningAlgorithmHS256,
		SigningKeyConfig:       secret,
		AccessTokenTTLSeconds:  DefaultAccessTokenTTLSeconds,
		IDTokenTTLSeconds:      DefaultIDTokenTTLSeconds,
		RefreshTokenTTLSeconds: DefaultRefreshTokenTTLSeconds,
		CreatedAt:              now,
		UpdatedAt:              now,
	}
	if err := s.apps.CreateSystem(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

// SetDefaultSignupRole sets appID's default_signup_role_id (DATA_MODEL.md `apps`) — the role
// auto-assigned to every new signup against this app. Used by the CLI seed bootstrap.
func (s *Service) SetDefaultSignupRole(ctx context.Context, appID, roleID uuid.UUID) error {
	return s.apps.SetDefaultSignupRole(ctx, appID, roleID)
}

func randomHex(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
