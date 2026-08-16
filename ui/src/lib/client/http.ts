import { getAccessToken, setAccessToken } from './token-store'

// ApiError carries the backend's own {key, message} envelope error (envelope.go) straight
// through — key is what a future i18n lookup (UI_CODING_STANDARDS.md §9) would key off of,
// message is the English fallback text shown today.
export class ApiError extends Error {
  status: number
  key: string

  constructor(status: number, key: string, message: string) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.key = key
  }
}

interface Envelope<T> {
  data: T
  error: { key: string; message: string } | null
}

// apiFetch is the one place every request attaches the current Bearer token and unwraps the
// backend's {data, error} envelope (envelope.go) — every feature call goes through this, never
// a bare fetch.
export async function apiFetch<T>(
  path: string,
  init?: Omit<RequestInit, 'body'> & { body?: unknown },
): Promise<T> {
  const token = getAccessToken()

  let res: Response
  try {
    res = await fetch(`/api/v1${path}`, {
      ...init,
      headers: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
        ...init?.headers,
      },
      body: init?.body !== undefined ? JSON.stringify(init.body) : undefined,
    })
  } catch {
    throw new ApiError(0, 'network.error', 'Could not reach the server.')
  }

  if (res.status === 401) {
    // Self-healing: an expired/invalid token is discarded the instant any request discovers
    // it, from any call site — auth-context.tsx's AuthProvider reacts via
    // token-store.subscribeAccessToken, so this is the one place session teardown on token
    // rejection needs to be wired, not something every future call site repeats.
    setAccessToken(null)
  }

  let envelope: Envelope<T>
  try {
    envelope = await res.json()
  } catch {
    throw new ApiError(
      res.status,
      'network.error',
      'Unexpected response from the server.',
    )
  }

  if (envelope.error) {
    throw new ApiError(res.status, envelope.error.key, envelope.error.message)
  }
  return envelope.data
}
