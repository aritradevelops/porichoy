// Package config loads server configuration via koanf (CODING_STANDARDS.md §1): a set of
// hardcoded local-dev defaults, overridable by PORICHOY_-prefixed environment variables.
// No YAML/file source yet — env-only is enough for what this server needs so far
// (CODING_STANDARDS.md §1 names YAML+env as the eventual shape, not a day-one requirement).
package config

import (
	"strings"

	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/v2"
)

// Config is the server's full runtime configuration.
type Config struct {
	// Port the REST adapter's Fiber app listens on.
	Port string
	// DBDSN is a standard postgres:// connection string, passed as-is to
	// postgres.Open (internal/adapters/postgres/db.go).
	DBDSN string
	// MigrationsDir is passed as-is to postgres.Migrate. Relative to the process's
	// working directory (not this source file's location) — the default assumes
	// `go run ./cmd/server` is invoked from server/, matching how the module itself is
	// rooted; override via PORICHOY_MIGRATIONS_DIR for any other run location (e.g. a
	// container image that copies migrations/ somewhere else).
	MigrationsDir string
}

// defaults match server/deploy/docker-compose.yml's local Postgres service, so `go run
// ./cmd/server` works against `docker compose up` with zero configuration.
var defaults = map[string]any{
	"port":           "8080",
	"db_dsn":         "postgres://porichoy:porichoy@localhost:5432/porichoy?sslmode=disable",
	"migrations_dir": "migrations",
}

// Load reads defaults, then overlays PORICHOY_-prefixed environment variables —
// PORICHOY_PORT and PORICHOY_DB_DSN override "port"/"db_dsn" respectively.
func Load() (*Config, error) {
	k := koanf.New(".")

	if err := k.Load(confmap.Provider(defaults, "."), nil); err != nil {
		return nil, err
	}

	envProvider := env.Provider(".", env.Opt{
		Prefix: "PORICHOY_",
		TransformFunc: func(key, value string) (string, any) {
			key = strings.ToLower(strings.TrimPrefix(key, "PORICHOY_"))
			return key, value
		},
	})
	if err := k.Load(envProvider, nil); err != nil {
		return nil, err
	}

	return &Config{
		Port:          k.String("port"),
		DBDSN:         k.String("db_dsn"),
		MigrationsDir: k.String("migrations_dir"),
	}, nil
}
