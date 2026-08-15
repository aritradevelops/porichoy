//go:build integration

package cache

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestRedisCache_SetUserPermissions_StoresJSONStringifiedArray(t *testing.T) {
	c := NewRedisCache(testClient)
	ctx := context.Background()
	tenantID, userID := uuid.New(), uuid.New()
	permissions := []string{"tenants:*@root", "domains:*@root"}

	err := c.SetUserPermissions(ctx, tenantID, userID, permissions, time.Hour)
	require.NoError(t, err)

	raw, err := testClient.Get(ctx, permissionsKey(tenantID, userID)).Result()
	require.NoError(t, err)

	var got []string
	require.NoError(t, json.Unmarshal([]byte(raw), &got))
	require.Equal(t, permissions, got)

	ttl, err := testClient.TTL(ctx, permissionsKey(tenantID, userID)).Result()
	require.NoError(t, err)
	require.Greater(t, ttl, 55*time.Minute)
	require.LessOrEqual(t, ttl, time.Hour)
}

func TestRedisCache_SetUserPermissions_NilBecomesEmptyArray(t *testing.T) {
	c := NewRedisCache(testClient)
	ctx := context.Background()
	tenantID, userID := uuid.New(), uuid.New()

	err := c.SetUserPermissions(ctx, tenantID, userID, nil, time.Hour)
	require.NoError(t, err)

	raw, err := testClient.Get(ctx, permissionsKey(tenantID, userID)).Result()
	require.NoError(t, err)
	require.Equal(t, "[]", raw)
}

func TestRedisCache_SetUserPermissions_ScopedByTenant(t *testing.T) {
	c := NewRedisCache(testClient)
	ctx := context.Background()
	userID := uuid.New()
	tenantA, tenantB := uuid.New(), uuid.New()

	require.NoError(t, c.SetUserPermissions(ctx, tenantA, userID, []string{"tenants:*@root"}, time.Hour))
	require.NoError(t, c.SetUserPermissions(ctx, tenantB, userID, []string{"domains:*@tenant"}, time.Hour))

	rawA, err := testClient.Get(ctx, permissionsKey(tenantA, userID)).Result()
	require.NoError(t, err)
	rawB, err := testClient.Get(ctx, permissionsKey(tenantB, userID)).Result()
	require.NoError(t, err)
	require.NotEqual(t, rawA, rawB)
}
