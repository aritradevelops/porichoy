// Package postgres is the default Postgres adapter (TECHNICAL_DESIGN.md §2) for the
// bounded-context repository interfaces defined under internal/*. Each repository here maps
// its own Bun-tagged model to/from the corresponding domain entity — the domain packages
// stay infra-agnostic (CODING_STANDARDS.md §2).
package postgres

import (
	"database/sql"

	"github.com/pressly/goose/v3"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

// Open connects to Postgres via Bun's pure-Go driver (CODING_STANDARDS.md §1 —
// bun/driver/pgdriver, no cgo). dsn is a standard postgres:// connection string.
func Open(dsn string) *bun.DB {
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	return bun.NewDB(sqldb, pgdialect.New())
}

// Migrate runs every pending goose migration in dir against db (CODING_STANDARDS.md §1/§2).
func Migrate(db *sql.DB, dir string) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	return goose.Up(db, dir)
}
