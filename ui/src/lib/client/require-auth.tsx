import type { ReactElement } from 'react'
import { Navigate, useLocation, type Location } from 'react-router-dom'
import { useAuth } from './auth-context'

// Redirects to /login when unauthenticated, preserving the attempted location in route state
// so login-form.tsx can send the user back where they were headed.
export function RequireAuth({ children }: { children: ReactElement }) {
  const { accessToken } = useAuth()
  const location = useLocation()

  if (!accessToken) {
    return <Navigate to="/login" state={{ from: location }} replace />
  }
  return children
}

// Sends an already-authenticated user away from /login rather than showing it again.
export function RedirectIfAuthenticated({ children }: { children: ReactElement }) {
  const { accessToken } = useAuth()
  return accessToken ? <Navigate to="/" replace /> : children
}

export type LocationWithFrom = Location & { state?: { from?: Location } }
