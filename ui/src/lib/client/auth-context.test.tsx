import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen, waitFor } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { AuthProvider, useAuth } from './auth-context'
import { meQueryKey } from './auth'
import { setAccessToken } from './token-store'

function jsonResponse(status: number, body: unknown) {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  })
}

function Probe() {
  const { user } = useAuth()
  return <p>{user ? user.email : 'no-user'}</p>
}

describe('AuthProvider', () => {
  beforeEach(() => {
    setAccessToken(null)
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('drops cached /auth/me data when the token is cleared outside of logout() (e.g. a 401)', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        jsonResponse(200, {
          data: {
            user_id: 'u1',
            email: 'someone@example.com',
            display_name: null,
            tenant_id: 't1',
            tenant_name: 'Acme',
            permissions: [],
          },
          error: null,
        }),
      ),
    )
    const queryClient = new QueryClient()
    render(
      <QueryClientProvider client={queryClient}>
        <AuthProvider>
          <Probe />
        </AuthProvider>
      </QueryClientProvider>,
    )

    setAccessToken('a-real-token')
    await screen.findByText('someone@example.com')
    expect(queryClient.getQueryData(meQueryKey)).toBeDefined()

    // Simulates http.ts's own 401 handling, which clears the token-store directly rather than
    // going through AuthProvider's logout() — the cached /auth/me data must be dropped here
    // too, not just on an explicit sign-out, or a later session on the same tab could briefly
    // render the previous session's identity/permissions.
    setAccessToken(null)

    await waitFor(() => {
      expect(queryClient.getQueryData(meQueryKey)).toBeUndefined()
    })
    expect(await screen.findByText('no-user')).toBeInTheDocument()
  })
})
