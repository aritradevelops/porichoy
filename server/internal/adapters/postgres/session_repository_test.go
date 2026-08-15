//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/aritradevelops/porichoy/server/internal/app"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestSessionRepository_Create(t *testing.T) {
	tenants := NewTenantRepository(testDB)
	apps := NewAppRepository(testDB)
	users := NewUserRepository(testDB)
	sessions := NewSessionRepository(testDB)
	ctx := context.Background()

	tt := mustCreateRoot(t, tenants, "Session Tenant")
	sysApp := newTestSystemApp(tt.ID)
	require.NoError(t, apps.CreateSystem(ctx, sysApp))
	u := newTestUser(tt.ID, "sess-"+uuid.NewString()+"@example.com")
	require.NoError(t, users.Create(ctx, u))

	_, refreshHash, err := app.NewRefreshToken()
	require.NoError(t, err)
	now := time.Now().UTC().Truncate(time.Microsecond)
	sess := &app.Session{
		ID:               uuid.New(),
		UserID:           u.ID,
		AppID:            sysApp.ID,
		Aud:              tt.ID.String(),
		RefreshTokenHash: refreshHash,
		CreatedAt:        now,
		LastActiveAt:     now,
		ExpiresAt:        now.Add(24 * time.Hour),
	}
	require.NoError(t, sessions.Create(ctx, sess))

	var count int
	err = testDB.NewSelect().Table("sessions").Where("id = ?", sess.ID).ColumnExpr("count(*)").Scan(ctx, &count)
	require.NoError(t, err)
	require.Equal(t, 1, count)
}
