//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/aritradevelops/porichoy/server/internal/identity"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestPasswordRepository_Create(t *testing.T) {
	tenants := NewTenantRepository(testDB)
	users := NewUserRepository(testDB)
	passwords := NewPasswordRepository(testDB)
	ctx := context.Background()

	tt := mustCreateRoot(t, tenants, "Password Tenant")
	u := newTestUser(tt.ID, "pw-"+uuid.NewString()+"@example.com")
	require.NoError(t, users.Create(ctx, u))

	p := &identity.Password{
		ID:           uuid.New(),
		UserID:       u.ID,
		PasswordHash: "bcrypt-hash",
		CreatedAt:    time.Now().UTC().Truncate(time.Microsecond),
	}
	require.NoError(t, passwords.Create(ctx, p))
}

func TestPasswordRepository_Create_SecondActivePasswordViolatesUniqueness(t *testing.T) {
	tenants := NewTenantRepository(testDB)
	users := NewUserRepository(testDB)
	passwords := NewPasswordRepository(testDB)
	ctx := context.Background()

	tt := mustCreateRoot(t, tenants, "Password Uniqueness Tenant")
	u := newTestUser(tt.ID, "pw2-"+uuid.NewString()+"@example.com")
	require.NoError(t, users.Create(ctx, u))

	now := time.Now().UTC().Truncate(time.Microsecond)
	require.NoError(t, passwords.Create(ctx, &identity.Password{
		ID: uuid.New(), UserID: u.ID, PasswordHash: "hash-1", CreatedAt: now,
	}))

	err := passwords.Create(ctx, &identity.Password{
		ID: uuid.New(), UserID: u.ID, PasswordHash: "hash-2", CreatedAt: now,
	})
	require.Error(t, err, "a second active password for the same user must violate the partial unique index")
}

func TestPasswordRepository_FindByUserID_ReturnsActivePassword(t *testing.T) {
	tenants := NewTenantRepository(testDB)
	users := NewUserRepository(testDB)
	passwords := NewPasswordRepository(testDB)
	ctx := context.Background()

	tt := mustCreateRoot(t, tenants, "Password Find Tenant")
	u := newTestUser(tt.ID, "pw-find-"+uuid.NewString()+"@example.com")
	require.NoError(t, users.Create(ctx, u))
	require.NoError(t, passwords.Create(ctx, &identity.Password{
		ID: uuid.New(), UserID: u.ID, PasswordHash: "bcrypt-hash", CreatedAt: time.Now().UTC().Truncate(time.Microsecond),
	}))

	got, err := passwords.FindByUserID(ctx, u.ID)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, "bcrypt-hash", got.PasswordHash)
}

func TestPasswordRepository_FindByUserID_NoneReturnsNil(t *testing.T) {
	passwords := NewPasswordRepository(testDB)

	got, err := passwords.FindByUserID(context.Background(), uuid.New())
	require.NoError(t, err)
	require.Nil(t, got)
}
