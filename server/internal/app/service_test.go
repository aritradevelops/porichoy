package app

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_CreateSystemApp(t *testing.T) {
	repo := &mockAppRepo{}
	svc := NewService(repo)
	tenantID := uuid.New()
	repo.On("CreateSystem", mock.Anything, mock.AnythingOfType("*app.App")).Return(nil)

	a, err := svc.CreateSystemApp(context.Background(), tenantID, "System")

	require.NoError(t, err)
	require.Equal(t, tenantID, a.TenantID)
	require.True(t, a.IsSystem)
	require.Equal(t, SigningAlgorithmHS256, a.SigningAlgorithm)
	require.Len(t, a.SigningKeyConfig, 32)
	require.NotEmpty(t, a.ClientID)
	require.Equal(t, DefaultAccessTokenTTLSeconds, a.AccessTokenTTLSeconds)
	require.Equal(t, DefaultIDTokenTTLSeconds, a.IDTokenTTLSeconds)
	require.Equal(t, DefaultRefreshTokenTTLSeconds, a.RefreshTokenTTLSeconds)
	repo.AssertExpectations(t)
}

func TestService_SetDefaultSignupRole(t *testing.T) {
	repo := &mockAppRepo{}
	svc := NewService(repo)
	appID, roleID := uuid.New(), uuid.New()
	repo.On("SetDefaultSignupRole", mock.Anything, appID, roleID).Return(nil)

	err := svc.SetDefaultSignupRole(context.Background(), appID, roleID)

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestNewRefreshToken(t *testing.T) {
	raw1, hash1, err := NewRefreshToken()
	require.NoError(t, err)
	require.NotEmpty(t, raw1)
	require.Len(t, hash1, 64) // hex-encoded sha256

	raw2, hash2, err := NewRefreshToken()
	require.NoError(t, err)
	require.NotEqual(t, raw1, raw2)
	require.NotEqual(t, hash1, hash2)
}
