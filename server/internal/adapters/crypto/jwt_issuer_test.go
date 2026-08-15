package crypto

import (
	"testing"
	"time"

	"github.com/aritradevelops/porichoy/server/internal/app"
	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/lestrrat-go/jwx/v3/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIssuer_Issue_ProducesAVerifiableToken(t *testing.T) {
	secret := []byte("test-secret-at-least-32-bytes-long!")
	a := &app.App{SigningAlgorithm: app.SigningAlgorithmHS256, SigningKeyConfig: secret}
	issuer := NewIssuer()

	signed, err := issuer.Issue(a, app.Claims{
		Subject:  "user-1",
		Audience: "tenant-1",
		Extra:    map[string]any{"email": "a@example.com", "email_verified": false},
	}, 15*time.Minute)
	require.NoError(t, err)
	require.NotEmpty(t, signed)

	tok, err := jwt.Parse([]byte(signed), jwt.WithKey(jwa.HS256(), secret))
	require.NoError(t, err)
	subject, ok := tok.Subject()
	require.True(t, ok)
	assert.Equal(t, "user-1", subject)
	audience, ok := tok.Audience()
	require.True(t, ok)
	assert.Equal(t, []string{"tenant-1"}, audience)
	var email string
	require.NoError(t, tok.Get("email", &email))
	assert.Equal(t, "a@example.com", email)
}

func TestIssuer_Issue_RejectsWrongKey(t *testing.T) {
	a := &app.App{SigningAlgorithm: app.SigningAlgorithmHS256, SigningKeyConfig: []byte("right-secret-at-least-32-bytes!!")}
	issuer := NewIssuer()

	signed, err := issuer.Issue(a, app.Claims{Subject: "user-1", Audience: "tenant-1"}, time.Minute)
	require.NoError(t, err)

	_, err = jwt.Parse([]byte(signed), jwt.WithKey(jwa.HS256(), []byte("wrong-secret-at-least-32-bytes!!")))
	assert.Error(t, err)
}

func TestIssuer_Issue_UnsupportedAlgorithm(t *testing.T) {
	a := &app.App{SigningAlgorithm: "RS256"}
	issuer := NewIssuer()

	_, err := issuer.Issue(a, app.Claims{Subject: "user-1", Audience: "tenant-1"}, time.Minute)

	require.ErrorIs(t, err, ErrUnsupportedSigningAlgorithm)
}

func TestIssuer_Verify_RoundTripsIssue(t *testing.T) {
	a := &app.App{SigningAlgorithm: app.SigningAlgorithmHS256, SigningKeyConfig: []byte("test-secret-at-least-32-bytes-long!")}
	issuer := NewIssuer()
	signed, err := issuer.Issue(a, app.Claims{Subject: "user-1", Audience: "tenant-1"}, time.Hour)
	require.NoError(t, err)

	claims, err := issuer.Verify(a, signed)

	require.NoError(t, err)
	assert.Equal(t, "user-1", claims.Subject)
	assert.Equal(t, "tenant-1", claims.Audience)
}

func TestIssuer_Verify_RejectsWrongKey(t *testing.T) {
	a := &app.App{SigningAlgorithm: app.SigningAlgorithmHS256, SigningKeyConfig: []byte("right-secret-at-least-32-bytes!!")}
	issuer := NewIssuer()
	signed, err := issuer.Issue(a, app.Claims{Subject: "user-1", Audience: "tenant-1"}, time.Hour)
	require.NoError(t, err)

	wrongKeyApp := &app.App{SigningAlgorithm: app.SigningAlgorithmHS256, SigningKeyConfig: []byte("wrong-secret-at-least-32-bytes!!")}
	_, err = issuer.Verify(wrongKeyApp, signed)

	assert.Error(t, err)
}

func TestIssuer_Verify_RejectsExpiredToken(t *testing.T) {
	a := &app.App{SigningAlgorithm: app.SigningAlgorithmHS256, SigningKeyConfig: []byte("test-secret-at-least-32-bytes-long!")}
	issuer := NewIssuer()
	signed, err := issuer.Issue(a, app.Claims{Subject: "user-1", Audience: "tenant-1"}, -time.Minute)
	require.NoError(t, err)

	_, err = issuer.Verify(a, signed)

	assert.Error(t, err)
}

func TestIssuer_Verify_UnsupportedAlgorithm(t *testing.T) {
	a := &app.App{SigningAlgorithm: "RS256"}
	issuer := NewIssuer()

	_, err := issuer.Verify(a, "irrelevant")

	require.ErrorIs(t, err, ErrUnsupportedSigningAlgorithm)
}
