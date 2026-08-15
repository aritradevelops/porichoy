package identity

import "context"

// TxRunner lets Signup's multi-repository write (User + Password + RoleAssignment + Session,
// spanning three packages' repositories) commit atomically — CODING_STANDARDS.md §3 assigns
// a Service its transactional boundary but defers the exact mechanism to whenever it's
// actually needed; this is that moment. fn's ctx carries whatever ambient transaction state
// the concrete implementation attaches — repositories read it back transparently
// (internal/adapters/postgres/tx.go), so no repository method signature changes.
type TxRunner interface {
	RunInTx(ctx context.Context, fn func(ctx context.Context) error) error
}
