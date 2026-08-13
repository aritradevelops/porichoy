//go:build integration

package rest

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/aritradevelops/porichoy/server/internal/adapters/postgres"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/uptrace/bun"
)

// testDB is shared across every integration test in this package for one test-binary run —
// one container, not one per test (CODING_STANDARDS.md §6), same pattern as
// internal/adapters/postgres/main_test.go.
var testDB *bun.DB

func TestMain(m *testing.M) {
	os.Exit(runTests(m))
}

func runTests(m *testing.M) int {
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx, "postgres:16-alpine",
		tcpostgres.WithDatabase("porichoy_test"),
		tcpostgres.WithUsername("porichoy"),
		tcpostgres.WithPassword("porichoy"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		fmt.Println("start postgres container:", err)
		return 1
	}
	defer func() { _ = container.Terminate(ctx) }()

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		fmt.Println("get connection string:", err)
		return 1
	}

	testDB = postgres.Open(dsn)
	defer testDB.Close()

	if err := postgres.Migrate(testDB.DB, "../../../migrations"); err != nil {
		fmt.Println("run migrations:", err)
		return 1
	}

	return m.Run()
}
