package tenant

import (
	"testing"

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
