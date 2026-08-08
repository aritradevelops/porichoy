package tenant

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTenantDomain_IsDeleted(t *testing.T) {
	active := &TenantDomain{}
	assert.False(t, active.IsDeleted())

	deletedAt := time.Now()
	deleted := &TenantDomain{DeletedAt: &deletedAt}
	assert.True(t, deleted.IsDeleted())
}
