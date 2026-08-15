//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/aritradevelops/porichoy/server/internal/identity"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestUser(tenantID uuid.UUID, email string) *identity.User {
	now := time.Now().UTC().Truncate(time.Microsecond)
	return &identity.User{
		ID:            uuid.New(),
		TenantID:      tenantID,
		Status:        identity.UserStatusActive,
		Email:         &email,
		EmailVerified: false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func TestUserRepository_CreateAndFindByEmail(t *testing.T) {
	tenants := NewTenantRepository(testDB)
	users := NewUserRepository(testDB)
	ctx := context.Background()

	tt := mustCreateRoot(t, tenants, "User Tenant")
	email := "a-" + uuid.NewString() + "@example.com"
	u := newTestUser(tt.ID, email)
	require.NoError(t, users.Create(ctx, u))

	got, err := users.FindByEmail(ctx, tt.ID, email)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, u.ID, got.ID)
	assert.Equal(t, email, *got.Email)
}

func TestUserRepository_FindByEmail_NotFound(t *testing.T) {
	tenants := NewTenantRepository(testDB)
	users := NewUserRepository(testDB)
	ctx := context.Background()

	tt := mustCreateRoot(t, tenants, "No User Tenant")

	got, err := users.FindByEmail(ctx, tt.ID, "nobody@example.com")
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestUserRepository_FindByEmail_TenantScoped(t *testing.T) {
	tenants := NewTenantRepository(testDB)
	users := NewUserRepository(testDB)
	ctx := context.Background()

	tenantA := mustCreateRoot(t, tenants, "Tenant A")
	tenantB := mustCreateRoot(t, tenants, "Tenant B")
	email := "shared-" + uuid.NewString() + "@example.com"
	require.NoError(t, users.Create(ctx, newTestUser(tenantA.ID, email)))

	gotInB, err := users.FindByEmail(ctx, tenantB.ID, email)
	require.NoError(t, err)
	assert.Nil(t, gotInB, "same email in a different tenant must not be found")
}
