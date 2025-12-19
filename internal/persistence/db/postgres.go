package db

import (
	"context"
	"fmt"
	"time"

	"github.com/aritradeveops/porichoy/internal/persistence/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

const wait = 30 * time.Second

// implements Database
type Postgres struct {
	uri string
	// conn *pgx.Conn
	pool *pgxpool.Pool
}

func NewPostgres(uri string) Database {
	return &Postgres{
		uri: uri,
	}
}

func (p *Postgres) Connect() error {
	connectionCtx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	// conn, err := pgx.Connect(connectionCtx, p.uri)
	pool, err := pgxpool.New(connectionCtx, p.uri)
	if err != nil {
		return fmt.Errorf("failed to connect to the database : %v", err)
	}
	p.pool = pool
	return nil
}
func (p *Postgres) Disconnect() error {
	if p.pool == nil {
		return NotInitializedErr("Postgres")
	}
	p.pool.Close()
	p.pool = nil
	return nil
}
func (p *Postgres) Health() error {
	if p.pool == nil {
		return NotInitializedErr("Postgres")
	}
	pingCtx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	err := p.pool.Ping(pingCtx)
	if err != nil {
		return fmt.Errorf("failed to ping the database: %v", err)
	}
	return nil
}

func (p *Postgres) Tx() (repository.DBTX, error) {
	if p.pool == nil {
		return nil, NotInitializedErr("Postgres")
	}
	return p.pool, nil
}
