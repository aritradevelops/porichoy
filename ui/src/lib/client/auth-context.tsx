import { createContext, useContext, useEffect, useState, type ReactNode } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { getAccessToken, setAccessToken, subscribeAccessToken } from './token-store'
import { meQueryKey, useMeQuery, type AuthResponse, type MeResponse } from './auth'

interface AuthContextValue {
  accessToken: string | null
  user: MeResponse | undefined
  isLoading: boolean
  login: (auth: AuthResponse) => void
  logout: () => void
}

const AuthContext = createContext<AuthContextValue | null>(null)

// Client/UI session state via React Context (UI_CODING_STANDARDS.md §4), orchestrating the
// module-level token-store (the one non-React state apiFetch's imperative calls read from)
// with TanStack Query's cache for /auth/me.
//
// Bearer token, in-memory only — this pass's decided scope, and a deliberate, documented
// divergence from UI_CODING_STANDARDS.md §8's eventual HttpOnly-cookie design: the backend
// doesn't set that cookie yet, and the CSRF mechanism a cookie flow would need is still an
// open item there. A page reload always starts unauthenticated (no persistence attempted) —
// acceptable since there's no refresh-token endpoint yet either, so nothing meaningful would
// be restored anyway.
export function AuthProvider({ children }: { children: ReactNode }) {
  const [accessToken, setToken] = useState<string | null>(getAccessToken)
  const queryClient = useQueryClient()
  const meQuery = useMeQuery(!!accessToken)

  // Single source of truth for the React-visible session state: fires on login/logout below,
  // and on http.ts's own 401 handling clearing the token-store directly — so an
  // expired/rejected token tears down the session (and triggers RequireAuth's redirect) no
  // matter which call site discovered it.
  useEffect(() => subscribeAccessToken(() => setToken(getAccessToken())), [])

  function login(auth: AuthResponse) {
    setAccessToken(auth.access_token)
    queryClient.invalidateQueries({ queryKey: meQueryKey })
  }

  function logout() {
    setAccessToken(null)
    queryClient.removeQueries({ queryKey: meQueryKey })
  }

  return (
    <AuthContext.Provider
      value={{
        accessToken,
        user: meQuery.data,
        isLoading: !!accessToken && meQuery.isLoading,
        login,
        logout,
      }}
    >
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within an AuthProvider')
  return ctx
}
