package app

import "time"

// Claims is the set of inputs TokenIssuer builds a signed JWT from, and TokenVerifier
// recovers from a previously issued one. Subject/Audience are the token's sub/aud claims
// (TECHNICAL_DESIGN §3.5/§4); Extra is only meaningful on Issue (e.g. an ID token's
// email/email_verified) — Verify doesn't attempt to recover it, nothing that calls Verify
// needs anything beyond Subject/Audience.
type Claims struct {
	Subject  string
	Audience string
	Extra    map[string]any
}

// TokenIssuer signs Claims into a JWT using one App's own signing config
// (SigningAlgorithm/SigningKeyConfig) — an interface, not a plain function, so a future
// RS256/JWKS implementation (TECHNICAL_DESIGN §4) swaps in later without touching any caller.
// Only HS256 is implemented this pass (internal/adapters/crypto).
type TokenIssuer interface {
	Issue(a *App, claims Claims, ttl time.Duration) (string, error)
}

// TokenVerifier verifies and decodes a token previously issued for a — checks the signature
// and standard time-based claims (exp/nbf) using a's own signing config, the inverse of
// TokenIssuer.Issue. Backs the real Authentication middleware
// (internal/adapters/rest/middleware.go, via identity.Service.Authenticate).
type TokenVerifier interface {
	Verify(a *App, tokenString string) (Claims, error)
}

// TokenService is the combined port identity.Service depends on: every App-scoped flow this
// pass needs both directions (Signup/Login issue, Authenticate verifies), and the one
// concrete implementation (internal/adapters/crypto.Issuer) provides both, so there's no
// caller yet that wants just one half.
type TokenService interface {
	TokenIssuer
	TokenVerifier
}
