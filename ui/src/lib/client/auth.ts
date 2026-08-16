import { useMutation, useQuery } from '@tanstack/react-query'
import { apiFetch } from './http'

// DTOs mirror the backend's JSON field names verbatim (auth_handlers.go's authResponse/
// meResponse) — no client-side renaming, so the shape here is provably the wire shape.
export interface LoginRequest {
  email: string
  password: string
}

export interface AuthResponse {
  user_id: string
  email: string
  access_token: string
  id_token: string
  refresh_token: string
  token_type: string
  expires_in: number
}

export interface MeResponse {
  user_id: string
  email: string
  display_name: string | null
  tenant_id: string
  tenant_name: string
  permissions: string[]
}

export function useLoginMutation() {
  return useMutation({
    mutationFn: (body: LoginRequest) =>
      apiFetch<AuthResponse>('/auth/login', { method: 'POST', body }),
  })
}

export const meQueryKey = ['auth', 'me'] as const

// enabled is passed in rather than read internally so the caller (auth-context.tsx) decides
// when a token exists worth checking — this hook has no opinion on session state itself.
export function useMeQuery(enabled: boolean) {
  return useQuery({
    queryKey: meQueryKey,
    queryFn: () => apiFetch<MeResponse>('/auth/me'),
    enabled,
    retry: false, // a 401 here means "not logged in", not a transient failure worth retrying
    staleTime: Infinity, // identity/permissions only change via a fresh login this pass
  })
}
