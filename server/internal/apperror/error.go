// Package apperror defines the error type domain and application code use to signal
// failures — identified by an i18n translation key rather than a hardcoded message
// (CODING_STANDARDS.md §8). Adapters resolve Key to locale-specific text at the
// response boundary; nothing in this package knows about message text or locales.
package apperror

// Error is a domain/application error identified by an i18n key, plus the HTTP status
// it maps to at the REST boundary (CODING_STANDARDS.md §5's response envelope). Status
// travels with the error rather than living in a REST-adapter-local lookup table, so it
// can't drift out of sync as new error keys get added — every apperror.New call site
// states its own status up front. Status is a plain int (not an http.Status* constant
// import) to keep this package free of any transport-layer dependency.
type Error struct {
	Key    string
	Status int
}

// New creates an Error carrying the given i18n key and HTTP status.
func New(key string, status int) *Error {
	return &Error{Key: key, Status: status}
}

func (e *Error) Error() string {
	return e.Key
}
