// Maps an ApiError's key to locale text (UI_CODING_STANDARDS.md §9: backend error responses
// are mapped through a lookup here rather than displaying error.message directly, so the UI
// can localize a backend error independently of whatever locale the backend itself resolved).
// Single-locale placeholder, same stand-in shape as the backend's own envelope.go `messages`
// map — not the real translation-file mechanism, which is still an open item.
const messages: Record<string, string> = {
  'identity.invalid_credentials': 'Incorrect email or password.',
  'identity.login_method_disabled': "This tenant doesn't allow email/password sign-in.",
  'identity.unauthenticated': 'You must be signed in to do that.',
  'tenant.unresolved_domain': 'No tenant is registered for this domain.',
  'validation.invalid_body': 'The request body is malformed.',
  'validation.failed': 'The request failed validation.',
  'network.error': 'Could not reach the server. Check your connection and try again.',
}

// Falls back to fallback (the backend's own message text) for a key with no local entry,
// rather than a blank or raw key — mirrors envelope.go's own "never leave a response
// message-less" reasoning, just with a friendlier fallback than the raw key itself.
export function resolveErrorMessage(key: string, fallback: string): string {
  return messages[key] ?? fallback
}
