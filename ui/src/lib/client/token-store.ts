// The single in-memory holder of the current Bearer access token (decided scope: no cookies,
// no localStorage/sessionStorage — see auth-context.tsx's own doc comment for why). Lives
// outside React because http.ts's apiFetch is called from TanStack Query's queryFn/mutationFn,
// which run without hook access — something module-scoped has to hold "what does the next
// fetch call send". auth-context.tsx's AuthProvider is the only place that treats this as a
// source of truth for rendering, via subscribeAccessToken.
type Listener = () => void

let token: string | null = null
const listeners = new Set<Listener>()

export function getAccessToken() {
  return token
}

export function setAccessToken(next: string | null) {
  token = next
  listeners.forEach((listener) => listener())
}

export function subscribeAccessToken(listener: Listener) {
  listeners.add(listener)
  return () => {
    listeners.delete(listener)
  }
}
