package app

import "time"

// Claims is the set of inputs TokenIssuer builds a signed JWT from. Subject/Audience become
// the token's sub/aud claims (TECHNICAL_DESIGN §3.5/§4); Extra is merged in as additional
// top-level claims (e.g. an ID token's email/email_verified).
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
