# Authorization Model: Routes, Scope Resolution & Filtering

Companion to [PRD.md](./PRD.md) §7 (Permissions, Roles & Policies) and
[TECHNICAL_DESIGN.md](./TECHNICAL_DESIGN.md) §3.5 (Default System App) and §6 (Permissions
Engine). Describes the concrete mechanism behind the `{module}:{action}@{scope}` permission
model: how a route maps to a permission check, how scope is resolved, and how that scope
becomes a query filter.

## 1. Route Convention

All API routes follow: **`/api/{version}/{module}/{action}/{id?}`**

The module and action are read directly off the URL — this is deliberately RPC-style rather
than verb-inferred REST, so the route itself always names the exact permission being
exercised (e.g. `/api/v1/users/list` → `users:list`, `/api/v1/roles/assign/42` →
`roles:assign`).

## 2. Middleware Permission Check

For every request:

1. Extract `{module}` and `{action}` from the route.
2. Look up every permission the caller (user or API credential) holds matching
   `{module}:{action}@*` — i.e. across all scopes they have for this exact module+action
   (or a matching wildcard, §5).
3. Resolve **the highest scope** among the matches (§3).
4. Attach the resolved scope, and its associated identifier (current tenant_id / app_id /
   org_id / caller's own id), to the request context for the handler to use.
5. If no matching permission exists at any scope, reject with 403.

## 3. Scope Ranking

Scopes form a fixed, total order, narrowest to broadest:

```
own < org < app < tenant < root
```

- **`own`**: the caller acting on a single record that belongs to them.
- **`org`**: everything within one organization.
- **`app`**: everything within one app (across all its organizations).
- **`tenant`**: everything within one tenant (across all its apps/organizations).
- **`root`**: everything, across every tenant — no filter applied at all.

"Take the highest scope" means: among all matching permissions for the current
`{module}:{action}`, pick the one highest in this ordering. This assumes the request already
has a resolved "current" tenant/app/org context (e.g. tenant resolved via the domain
registry, TECHNICAL_DESIGN §3.3) — scope resolution decides *how wide a filter* to apply
within that context, not which of several unrelated tenants/orgs to pick.

## 4. Per-Module Scope → Filter Mapping

Each module defines its own mapping from resolved scope to a query filter — this is where
the flexibility comes in; the same scope name can mean a different column/condition
depending on the resource.

**Default mapping** (applies unless a module overrides it):

| Scope | Default filter |
|---|---|
| `own` | `created_by = :caller_id` |
| `org` | `org_id = :org_id` |
| `app` | `app_id = :app_id` |
| `tenant` | `tenant_id = :tenant_id` |
| `root` | *(no filter)* |

**Per-module overrides**, applied where the default doesn't fit the resource's semantics:

- **`users` module, `own` scope**: `user_id = :caller_id` instead of `created_by` — a user
  record isn't "created by" itself in the usual sense, its owner *is* the record. This is
  also what makes self-service account management (PRD §9) fall out of the same mechanism
  for free: a logged-in user managing their own profile/MFA/sessions is just exercising
  `users:*@own` (§5) against their own row — no special-cased self-service logic needed at
  the authorization layer.
- **`apps` module**: the `system = true` row (the tenant's default system app,
  TECHNICAL_DESIGN §3.5) is only included in results when the resolved scope is `tenant` or
  `root`. At `app` scope, it's excluded outright — an app-level admin (someone scoped to a
  single specific app) can never see the system app, even incidentally. Only tenant admins
  and root admins can.

Additional modules will define their own mappings as they're built; this table is expected
to grow, not be exhaustive from day one.

## 5. Wildcard Actions

A permission's action position may be `*`, meaning "every action on this module at this
scope": **`{module}:*@{scope}`**.

- Example: `users:*@tenant` grants every action (read, list, create, update, delete, assign,
  etc.) on the `users` module, scoped to the tenant.
- Scoped to the **action position only** for now — module-level or fully wildcarded
  permissions (e.g. `*:*@root`) are not supported. A broad admin role is built by granting
  the module-wildcard permission explicitly for each module it needs, not by a single
  catch-all string.
