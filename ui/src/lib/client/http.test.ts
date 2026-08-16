import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { apiFetch, ApiError } from './http'
import { getAccessToken, setAccessToken } from './token-store'

function jsonResponse(status: number, body: unknown) {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  })
}

describe('apiFetch', () => {
  beforeEach(() => {
    setAccessToken(null)
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('returns the unwrapped data on success', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(jsonResponse(200, { data: { ok: true }, error: null })),
    )

    const result = await apiFetch<{ ok: boolean }>('/auth/me')

    expect(result).toEqual({ ok: true })
  })

  it('attaches the current access token as a bearer header', async () => {
    setAccessToken('a-real-token')
    const fetchMock = vi
      .fn()
      .mockResolvedValue(jsonResponse(200, { data: {}, error: null }))
    vi.stubGlobal('fetch', fetchMock)

    await apiFetch('/auth/me')

    const [, init] = fetchMock.mock.calls[0]
    expect((init.headers as Record<string, string>).Authorization).toBe(
      'Bearer a-real-token',
    )
  })

  it('throws ApiError carrying the backend error key/message', async () => {
    vi.stubGlobal(
      'fetch',
      vi
        .fn()
        .mockResolvedValue(
          jsonResponse(401, {
            data: null,
            error: { key: 'identity.unauthenticated', message: 'Nope.' },
          }),
        ),
    )

    await expect(apiFetch('/auth/me')).rejects.toMatchObject({
      status: 401,
      key: 'identity.unauthenticated',
      message: 'Nope.',
    })
    expect(await apiFetch('/auth/me').catch((e) => e)).toBeInstanceOf(ApiError)
  })

  it('clears the access token on a 401 response', async () => {
    setAccessToken('a-real-token')
    vi.stubGlobal(
      'fetch',
      vi
        .fn()
        .mockResolvedValue(
          jsonResponse(401, {
            data: null,
            error: { key: 'identity.unauthenticated', message: 'Nope.' },
          }),
        ),
    )

    await apiFetch('/auth/me').catch(() => undefined)

    expect(getAccessToken()).toBeNull()
  })

  it('throws a network ApiError when fetch itself rejects', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new TypeError('Failed to fetch')))

    await expect(apiFetch('/auth/me')).rejects.toMatchObject({
      status: 0,
      key: 'network.error',
    })
  })
})
