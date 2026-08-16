// Mirrors the backend's own {module}:{action}@{scope} matching
// (server/internal/authorization/service.go's ResolveScope), but only needs a boolean "do I
// hold this at all" answer for nav-item visibility — scope *ranking* stays a backend-only
// concern (it drives query filtering, not UI visibility), never reimplemented here.
export function hasPermission(
  permissions: string[],
  moduleName: string,
  action: string,
): boolean {
  return permissions.some((permission) => {
    const [modAction, scope] = permission.split('@')
    if (!scope) return false
    const [mod, act] = modAction.split(':')
    return mod === moduleName && (act === action || act === '*')
  })
}
