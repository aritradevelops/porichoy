// Package crypto holds the concrete implementations of the domain packages' crypto-shaped
// ports (JWT signing today) — kept out of internal/app itself so that package stays
// infra-agnostic (CODING_STANDARDS.md §2), same reasoning as bun models living in
// internal/adapters/postgres rather than the domain packages.
package crypto

import (
	"errors"
	"time"

	"github.com/aritradevelops/porichoy/server/internal/app"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/lestrrat-go/jwx/v3/jwt"
)

// ErrUnsupportedSigningAlgorithm is returned by Issuer.Issue for any App.SigningAlgorithm
// other than HS256 — the only one implemented this pass (RS256/ES256/JWKS are future work,
// TECHNICAL_DESIGN §4).
var ErrUnsupportedSigningAlgorithm = errors.New("crypto: unsupported signing algorithm")

// Issuer implements app.TokenIssuer via lestrrat-go/jwx (CODING_STANDARDS.md §1).
type Issuer struct{}

// NewIssuer builds an Issuer. Stateless — HS256 needs no held configuration beyond what each
// call's *app.App already carries.
func NewIssuer() *Issuer {
	return &Issuer{}
}

var _ app.TokenIssuer = (*Issuer)(nil)

// Issue signs claims into a compact JWS using a's own signing config (app.TokenIssuer).
func (i *Issuer) Issue(a *app.App, claims app.Claims, ttl time.Duration) (string, error) {
	if a.SigningAlgorithm != app.SigningAlgorithmHS256 {
		return "", ErrUnsupportedSigningAlgorithm
	}

	now := time.Now()
	b := jwt.NewBuilder().
		Subject(claims.Subject).
		Audience([]string{claims.Audience}).
		IssuedAt(now).
		Expiration(now.Add(ttl)).
		JwtID(uuid.NewString())
	for k, v := range claims.Extra {
		b = b.Claim(k, v)
	}
	tok, err := b.Build()
	if err != nil {
		return "", err
	}

	signed, err := jwt.Sign(tok, jwt.WithKey(jwa.HS256(), a.SigningKeyConfig))
	if err != nil {
		return "", err
	}
	return string(signed), nil
}
