package tenant

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTenant_IsRoot(t *testing.T) {
	root := &Tenant{ID: newUUID(t)}
	assert.True(t, root.IsRoot())

	childParent := newUUID(t)
	child := &Tenant{ID: newUUID(t), ParentID: &childParent}
	assert.False(t, child.IsRoot())
}

func TestTenant_IsDeleted(t *testing.T) {
	active := &Tenant{}
	assert.False(t, active.IsDeleted())

	deletedAt := time.Now()
	deleted := &Tenant{DeletedAt: &deletedAt}
	assert.True(t, deleted.IsDeleted())
}
