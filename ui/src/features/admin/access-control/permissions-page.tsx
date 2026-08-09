import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'

interface Permission {
  module: string
  action: string
  description: string
}

// Hand-derived representative sample of {module}:{action} pairs
// (AUTHORIZATION_MODEL.md §1's route convention) — not exhaustive, not the
// literal route table, just believable given the modules this codebase
// implies (users, roles, apps, tenants, domains, org_memberships,
// audit_logs, api_credentials). Will be generated from the real route table
// once lib/client exists (UI_CODING_STANDARDS.md §13).
const permissions: Permission[] = [
  {
    module: 'users',
    action: 'list',
    description: 'List users within the resolved scope.',
  },
  { module: 'users', action: 'read', description: "Read a single user's profile." },
  { module: 'users', action: 'update', description: "Update a user's profile fields." },
  { module: 'users', action: 'delete', description: 'Soft-delete a user account.' },
  {
    module: 'roles',
    action: 'list',
    description: 'List roles within the resolved scope.',
  },
  { module: 'roles', action: 'create', description: 'Create a new role.' },
  { module: 'roles', action: 'update', description: "Update a role's grants." },
  { module: 'roles', action: 'assign', description: 'Assign a role to a user.' },
  { module: 'apps', action: 'list', description: 'List registered apps.' },
  { module: 'apps', action: 'create', description: 'Register a new app.' },
  {
    module: 'apps',
    action: 'update',
    description: "Update an app's branding or OAuth configuration.",
  },
  {
    module: 'apps',
    action: 'regenerate_secret',
    description: "Rotate an app's client secret.",
  },
  { module: 'tenants', action: 'list', description: 'List tenants.' },
  { module: 'tenants', action: 'create', description: 'Create a new tenant.' },
  {
    module: 'tenants',
    action: 'update',
    description: "Update a tenant's configuration.",
  },
  {
    module: 'domains',
    action: 'list',
    description: "List a tenant's registered domains.",
  },
  {
    module: 'domains',
    action: 'verify',
    description: 'Verify ownership of a registered domain.',
  },
  {
    module: 'org_memberships',
    action: 'list',
    description: "List an organization's members.",
  },
  {
    module: 'org_memberships',
    action: 'invite',
    description: 'Invite a user to an organization.',
  },
  {
    module: 'org_memberships',
    action: 'remove',
    description: 'Remove a member from an organization.',
  },
  { module: 'audit_logs', action: 'list', description: 'List audit log entries.' },
  { module: 'api_credentials', action: 'list', description: 'List API credentials.' },
  {
    module: 'api_credentials',
    action: 'create',
    description: 'Create a new API credential.',
  },
  {
    module: 'api_credentials',
    action: 'revoke',
    description: 'Revoke an API credential.',
  },
]

// Read-only reference of every {module}:{action} pair the platform defines
// (UI_PAGES.md §7.2) — derived from routes/code, not admin-configurable, so
// deliberately no create/edit/delete affordance anywhere on this page. Not a
// gap: don't add one.
export function PermissionsPage() {
  return (
    <div className="flex flex-col gap-8">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight text-foreground">
          Permissions
        </h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Every permission the platform defines. Read-only — derived from routes and
          code, not something you configure here.
        </p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Reference</CardTitle>
          <CardDescription>
            Combine with a scope (AUTHORIZATION_MODEL.md §3) to form a full grant, e.g.
            users:list@tenant.
          </CardDescription>
        </CardHeader>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Permission</TableHead>
                <TableHead>Module</TableHead>
                <TableHead>Action</TableHead>
                <TableHead>Description</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {permissions.map((permission) => (
                <TableRow key={`${permission.module}:${permission.action}`}>
                  <TableCell className="font-mono text-xs text-foreground">
                    {permission.module}:{permission.action}
                  </TableCell>
                  <TableCell className="text-muted-foreground">
                    {permission.module}
                  </TableCell>
                  <TableCell className="text-muted-foreground">
                    {permission.action}
                  </TableCell>
                  <TableCell className="text-muted-foreground">
                    {permission.description}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  )
}
