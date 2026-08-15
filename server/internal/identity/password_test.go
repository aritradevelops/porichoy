package identity

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashPassword_RejectsOverBcryptLimit(t *testing.T) {
	tooLong := strings.Repeat("a", MaxPasswordBytes+1)

	_, err := HashPassword(tooLong)

	require.ErrorIs(t, err, ErrPasswordTooLong)
}

func TestHashPassword_AcceptsAtBcryptLimit(t *testing.T) {
	exact := strings.Repeat("a", MaxPasswordBytes)

	hash, err := HashPassword(exact)

	require.NoError(t, err)
	assert.True(t, VerifyPassword(hash, exact))
}
