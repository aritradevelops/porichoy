package tenant

import (
	"testing"

	"github.com/aritradevelops/porichoy/server/internal/actor"
	"github.com/google/uuid"
)

// newUUID generates a fresh random UUID for use as test fixture data.
func newUUID(t *testing.T) uuid.UUID {
	t.Helper()
	id, err := uuid.NewRandom()
	if err != nil {
		t.Fatalf("generate uuid: %v", err)
	}
	return id
}

// newActor returns a populated actor.Actor fixture for tests exercising authorized
// Service methods.
func newActor(t *testing.T) actor.Actor {
	t.Helper()
	return actor.Actor{
		PrincipalID: newUUID(t),
		TenantID:    newUUID(t),
		Scope:       actor.ScopeTenant,
	}
}
