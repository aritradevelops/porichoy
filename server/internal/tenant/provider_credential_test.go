package tenant

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestProviderCredential_IsDeleted(t *testing.T) {
	active := &ProviderCredential{}
	assert.False(t, active.IsDeleted())

	deletedAt := time.Now()
	deleted := &ProviderCredential{DeletedAt: &deletedAt}
	assert.True(t, deleted.IsDeleted())
}
