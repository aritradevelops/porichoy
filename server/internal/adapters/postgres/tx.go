package postgres

import (
	"context"

	"github.com/uptrace/bun"
)

// txKey is the unexported context key TxRunner stashes an in-flight bun.Tx under —
// unexported so only this file can read or write it.
type txKey struct{}

// TxRunner implements identity.TxRunner via a real Bun transaction — the mechanism
// identity.Service.Signup needs to commit its User+Password+RoleAssignment+Session write
// atomically across four repositories (CODING_STANDARDS.md §3 assigns the transactional
// boundary to the Service, but defers the exact mechanism to whenever it's actually needed —
// this is the first such case in this codebase).
type TxRunner struct {
	db *bun.DB
}

// NewTxRunner builds a TxRunner from an open Bun connection.
func NewTxRunner(db *bun.DB) *TxRunner {
	return &TxRunner{db: db}
}

// RunInTx runs fn inside a Bun transaction, stashing it in ctx so any repository call made
// through fn's ctx (via dbFromContext, below) transparently participates instead of opening
// its own connection.
//
// Reentrant: if ctx already carries a transaction (an outer RunInTx call, from this or
// another TxRunner instance sharing the same underlying connection — e.g. cmd/seed wrapping
// identity.Service.Signup's own internal RunInTx call), fn just joins it instead of starting
// a second, unrelated one — Postgres has no true nested transactions, and BEGINning again on
// an existing bun.Tx isn't what a caller composing multiple Service calls into one atomic
// bootstrap sequence wants.
func (r *TxRunner) RunInTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if _, ok := ctx.Value(txKey{}).(bun.Tx); ok {
		return fn(ctx)
	}
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		return fn(context.WithValue(ctx, txKey{}, tx))
	})
}

// dbFromContext returns the ambient bun.Tx stashed by TxRunner.RunInTx if ctx carries one,
// otherwise fallback (the repository's own *bun.DB). This is what lets the new app/identity/
// authorization repositories participate in identity.Service.Signup's transaction without
// their method signatures changing — every one of their write methods calls this first.
func dbFromContext(ctx context.Context, fallback bun.IDB) bun.IDB {
	if tx, ok := ctx.Value(txKey{}).(bun.Tx); ok {
		return tx
	}
	return fallback
}
