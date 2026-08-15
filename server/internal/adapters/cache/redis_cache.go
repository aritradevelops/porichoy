// Package cache holds the concrete implementations of the domain packages' cache-shaped
// ports — kept out of internal/authorization itself so that package stays infra-agnostic
// (CODING_STANDARDS.md §2), same reasoning as internal/adapters/crypto for JWT signing.
package cache

import (
	"context"
	"encoding/json"
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

// permissionsKey namespaces a principal's cached permissions by tenant — permission strings
// are meaningless outside the tenant they were resolved within (AUTHORIZATION_MODEL.md §3),
// so the same userID under two different tenants gets two independent entries.
func permissionsKey(tenantID, userID uuid.UUID) string {
	return fmt.Sprintf("permissions:%s:%s", tenantID, userID)
}

// SetUserPermissions stores permissions as a single JSON-array-stringified value under
// tenantID+userID (authorization.PermissionCache), expiring after ttl.
func (c *RedisCache) SetUserPermissions(ctx context.Context, tenantID, userID uuid.UUID, permissions []string, ttl time.Duration) error {
	if permissions == nil {
		permissions = []string{}
	}
	encoded, err := json.Marshal(permissions)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, permissionsKey(tenantID, userID), encoded, ttl).Err()
}
