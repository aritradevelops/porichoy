//go:build integration

package cache

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

// testClient is shared across every test in this package for one test-binary run — one
// container, not one per test, per CODING_STANDARDS.md §6's "hermetic, per-test-run
// containers."
var testClient *redis.Client

func TestMain(m *testing.M) {
	os.Exit(runTests(m))
}

func runTests(m *testing.M) int {
	ctx := context.Background()

	container, err := tcredis.Run(ctx, "redis:7-alpine")
	if err != nil {
		fmt.Println("start redis container:", err)
		return 1
	}
	defer func() { _ = testcontainers.TerminateContainer(container) }()

	connStr, err := container.ConnectionString(ctx)
	if err != nil {
		fmt.Println("get redis connection string:", err)
		return 1
	}
	opts, err := redis.ParseURL(connStr)
	if err != nil {
		fmt.Println("parse redis connection string:", err)
		return 1
	}

	testClient = redis.NewClient(opts)
	defer testClient.Close()

	return m.Run()
}
