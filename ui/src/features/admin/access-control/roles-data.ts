// Shared static mock data for Roles (4.1) and Policies (4.3) — kept in its
// own module, same reasoning as features/admin/apps/apps-data.ts, so
// policies-page.tsx can reference a role by id without reaching into
// roles-page.tsx's internals. Will be replaced by the roles API once
// lib/client exists (UI_CODING_STANDARDS.md §13).

// AUTHORIZATION_MODEL.md §3's fixed, total scope order — reused as a literal
// union rather than a free-form string so a grant's scope is exhaustively
// checked against the ranking that actually exists.
export type Scope = 'own' | 'org' | 'app' | 'tenant' | 'root'

export const SCOPE_LABEL: Record<Scope, string> = {
  own: 'Own',
  org: 'Organization',
  app: 'App',
  tenant: 'Tenant',
  root: 'Root',
}

// {module}:{action}@{scope} — the exact shape AUTHORIZATION_MODEL.md §1's
// route convention derives from (`/api/v1/{module}/{action}` -> the grant
// without its scope; the scope is attached once resolved, §2-3).
export interface Grant {
  module: string
  action: string
  scope: Scope
}

export interface Role {
  id: string
  name: string
  // The role's own headline scope level (UI_PAGES.md §7.1's "scope level"
  // column) — not necessarily identical to every one of its grants' scopes
  // below, though in this mock data they line up for simplicity.
  scope: Scope
  assignmentCount: number
  grants: Grant[]
}

// Spans all five scope levels (own/org/app/tenant/root) so the Roles table's
// scope column actually demonstrates variety, per UI_IMPLEMENTATION_TODO's
// 4.1 instructions.
export const roles: Role[] = [
  {
    id: 'super-admin',
    name: 'Super Admin',
    scope: 'root',
    assignmentCount: 2,
    grants: [
      { module: 'users', action: '*', scope: 'root' },
      { module: 'tenants', action: '*', scope: 'root' },
      { module: 'roles', action: '*', scope: 'root' },
      { module: 'apps', action: '*', scope: 'root' },
      { module: 'audit_logs', action: 'list', scope: 'root' },
    ],
  },
  {
    id: 'tenant-admin',
    name: 'Tenant Admin',
    scope: 'tenant',
    assignmentCount: 6,
    grants: [
      { module: 'users', action: '*', scope: 'tenant' },
      { module: 'apps', action: '*', scope: 'tenant' },
      { module: 'roles', action: 'assign', scope: 'tenant' },
      { module: 'domains', action: 'verify', scope: 'tenant' },
      { module: 'audit_logs', action: 'list', scope: 'tenant' },
    ],
  },
  {
    id: 'app-admin',
    name: 'App Admin',
    scope: 'app',
    assignmentCount: 11,
    grants: [
      { module: 'apps', action: 'update', scope: 'app' },
      { module: 'users', action: 'list', scope: 'app' },
      { module: 'sessions', action: 'revoke', scope: 'app' },
    ],
  },
  {
    id: 'org-owner',
    name: 'Org Owner',
    scope: 'org',
    assignmentCount: 34,
    grants: [
      { module: 'org_memberships', action: 'invite', scope: 'org' },
      { module: 'org_memberships', action: 'remove', scope: 'org' },
      { module: 'roles', action: 'assign', scope: 'org' },
      { module: 'users', action: 'list', scope: 'org' },
    ],
  },
  {
    id: 'member',
    name: 'Member',
    scope: 'own',
    assignmentCount: 512,
    grants: [
      { module: 'users', action: 'update', scope: 'own' },
      { module: 'mfa_methods', action: '*', scope: 'own' },
      { module: 'sessions', action: 'list', scope: 'own' },
      { module: 'sessions', action: 'revoke', scope: 'own' },
    ],
  },
]

export function getRoleById(id: string): Role | undefined {
  return roles.find((role) => role.id === id)
}
