//go:build integration

package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aritradevelops/porichoy/server/internal/identity"
	"github.com/aritradevelops/porichoy/server/internal/tenant"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTxRunner_CommitsOnSuccess(t *testing.T) {
	tenants := NewTenantRepository(testDB)
	users := NewUserRepository(testDB)
	passwords := NewPasswordRepository(testDB)
	tx := NewTxRunner(testDB)
	ctx := context.Background()

	tt := mustCreateRoot(t, tenants, "Tx Commit Tenant")
	email := "tx-commit-" + uuid.NewString() + "@example.com"
	u := newTestUser(tt.ID, email)

	err := tx.RunInTx(ctx, func(ctx context.Context) error {
		if err := users.Create(ctx, u); err != nil {
			return err
		}
		return passwords.Create(ctx, &identity.Password{
			ID:           uuid.New(),
			UserID:       u.ID,
			PasswordHash: "bcrypt-hash",
			CreatedAt:    time.Now().UTC().Truncate(time.Microsecond),
		})
	})
	require.NoError(t, err)

	got, err := users.FindByEmail(ctx, tt.ID, email)
	require.NoError(t, err)
	assert.NotNil(t, got, "user must be visible after a committed transaction")
}

func TestTxRunner_RollsBackOnError(t *testing.T) {
	tenants := NewTenantRepository(testDB)
	users := NewUserRepository(testDB)
	tx := NewTxRunner(testDB)
	ctx := context.Background()

	tt := mustCreateRoot(t, tenants, "Tx Rollback Tenant")
	email := "tx-rollback-" + uuid.NewString() + "@example.com"
	u := newTestUser(tt.ID, email)
	sentinel := errors.New("boom")

	err := tx.RunInTx(ctx, func(ctx context.Context) error {
		if err := users.Create(ctx, u); err != nil {
			return err
		}
		return sentinel
	})
	require.ErrorIs(t, err, sentinel)

	got, err := users.FindByEmail(ctx, tt.ID, email)
	require.NoError(t, err)
	assert.Nil(t, got, "user must not be visible after a rolled-back transaction")
}

// TestTxRunner_Reentrant_NestedCallJoinsOuterTransaction proves RunInTx composes the way
// cmd/seed relies on: it wraps tenant.Service.CreateRootTenant (-> TenantRepository.CreateRoot)
// and identity.Service.CreateRootUser (which internally calls its own TxRunner.RunInTx) in one
// outer RunInTx. If the nested call actually joins the outer transaction rather than opening
// a second, independent one, a failure inside the nested call must roll back the outer write
// too — that's the behavior asserted here, not just "no error".
func TestTxRunner_Reentrant_NestedCallJoinsOuterTransaction(t *testing.T) {
	tenants := NewTenantRepository(testDB)
	outerTx := NewTxRunner(testDB)
	innerTx := NewTxRunner(testDB) // a second instance, same underlying *bun.DB — mirrors how
	// cmd/seed's outer TxRunner and identity.Service's own internal TxRunner are two separate
	// values sharing one connection pool.
	ctx := context.Background()
	sentinel := errors.New("boom")

	tenantName := "Reentrant Tenant " + uuid.NewString()
	var createdID uuid.UUID
	err := outerTx.RunInTx(ctx, func(ctx context.Context) error {
		tt := &tenant.Tenant{
			ID:          uuid.New(),
			Name:        tenantName,
			LoginLayout: tenant.LoginLayoutCentered,
			CreatedAt:   time.Now().UTC().Truncate(time.Microsecond),
			UpdatedAt:   time.Now().UTC().Truncate(time.Microsecond),
		}
		createdID = tt.ID
		if err := tenants.CreateRoot(ctx, tt); err != nil {
			return err
		}
		// Simulates identity.Service.CreateRootUser's own internal RunInTx call, made with
		// a *different* TxRunner instance than the outer one — this must join the outer
		// transaction (via ctx), not start an unrelated second one.
		return innerTx.RunInTx(ctx, func(ctx context.Context) error {
			return sentinel
		})
	})

	require.ErrorIs(t, err, sentinel)
	got, findErr := tenants.FindByID(ctx, createdID)
	require.NoError(t, findErr)
	assert.Nil(t, got, "the outer write must be rolled back when the nested call fails — proves it joined the same transaction rather than committing independently")
}
