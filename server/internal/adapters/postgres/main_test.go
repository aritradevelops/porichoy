//go:build integration

package postgres

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/uptrace/bun"
)

// testDB is shared across every test in this package for one test-binary run — one
// container, not one per test, per CODING_STANDARDS.md §6's "hermetic, per-test-run
// containers."
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

	testDB = Open(dsn)
	defer testDB.Close()

	if err := Migrate(testDB.DB, "../../../migrations"); err != nil {
		fmt.Println("run migrations:", err)
		return 1
	}

	return m.Run()
}
