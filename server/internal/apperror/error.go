// Package apperror defines the error type domain and application code use to signal
// failures — identified by an i18n translation key rather than a hardcoded message
// (CODING_STANDARDS.md §8). Adapters resolve Key to locale-specific text at the
// response boundary; nothing in this package knows about message text or locales.
package apperror

// Error is a domain/application error identified by an i18n key.
type Error struct {
	Key string
}

// New creates an Error carrying the given i18n key.
func New(key string) *Error {
	return &Error{Key: key}
}

func (e *Error) Error() string {
	return e.Key
}
