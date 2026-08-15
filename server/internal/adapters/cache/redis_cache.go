// Package cache holds the concrete implementations of the domain packages' cache-shaped
// ports — kept out of internal/authorization itself so that package stays infra-agnostic
// (CODING_STANDARDS.md §2), same reasoning as internal/adapters/crypto for JWT signing.
package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/aritradevelops/porichoy/server/internal/authorization"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// RedisCache implements authorization.PermissionCache via Redis — the default cache provider
// (TECHNICAL_DESIGN.md §2).
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache builds a RedisCache from an open client.
func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

var _ authorization.PermissionCache = (*RedisCache)(nil)

// permissionsKey namespaces a principal's cached permissions by app — permission strings are
// meaningless outside the app they were resolved within (AUTHORIZATION_MODEL.md §3, an app
// owns its own role catalog), so the same userID under two different apps gets two
// independent entries. userID-then-appID ordering matches this project's original Node
// implementation (application/src/middlewares/auth.middleware.ts's `permissions_${user.id}_${user.app_id}`).
func permissionsKey(appID, userID uuid.UUID) string {
	return fmt.Sprintf("permissions_%s_%s", userID, appID)
}

// SetUserPermissions stores permissions as a single JSON-array-stringified value under
// appID+userID (authorization.PermissionCache), expiring after ttl.
func (c *RedisCache) SetUserPermissions(ctx context.Context, appID, userID uuid.UUID, permissions []string, ttl time.Duration) error {
	if permissions == nil {
		permissions = []string{}
	}
	encoded, err := json.Marshal(permissions)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, permissionsKey(appID, userID), encoded, ttl).Err()
}

// GetUserPermissions returns the raw JSON-array-encoded permissions previously stored for
// appID+userID (authorization.PermissionCache), or nil, nil on a cache miss.
func (c *RedisCache) GetUserPermissions(ctx context.Context, appID, userID uuid.UUID) ([]byte, error) {
	raw, err := c.client.Get(ctx, permissionsKey(appID, userID)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return raw, nil
}
