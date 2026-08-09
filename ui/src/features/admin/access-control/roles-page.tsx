import { useState } from 'react'
import { KeyRoundIcon } from 'lucide-react'
import { ListRow } from '@/components/list-row'
import { Badge } from '@/components/ui/badge'
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
import { roles, SCOPE_LABEL } from './roles-data'

// In-place, local-state drill-down (like tenants-page.tsx's root variant),
// not a real per-item route (like apps-list-page.tsx -> app-detail-page.tsx,
// UI_PAGES.md §6). A role's detail here is a single read-only grants list —
// no tabs, no independently-editable config surface — so it's closer in
// spirit to Tenants' inline settings preview than to an App's bookmarkable
// multi-tab detail page; the lighter pattern fits a permissions reference
// list better than a dedicated route would. UI_PAGES.md §7.1.
export function RolesPage() {
  const [selectedRoleId, setSelectedRoleId] = useState(roles[0].id)
  const selectedRole = roles.find((role) => role.id === selectedRoleId) ?? roles[0]

  return (
    <div className="flex flex-col gap-8">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight text-foreground">Roles</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Select a role to see the grants it carries.
        </p>
      </div>

      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Scope</TableHead>
                <TableHead>Assignments</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {roles.map((role) => (
                <TableRow
                  key={role.id}
                  data-state={role.id === selectedRoleId ? 'selected' : undefined}
                  onClick={() => setSelectedRoleId(role.id)}
                  className="cursor-pointer"
                >
                  <TableCell className="font-medium text-foreground">
                    {role.name}
                  </TableCell>
                  <TableCell>
                    <Badge variant="outline">{SCOPE_LABEL[role.scope]}</Badge>
                  </TableCell>
                  <TableCell className="text-muted-foreground">
                    {role.assignmentCount.toLocaleString()}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>{selectedRole.name}</CardTitle>
          <CardDescription>
            {selectedRole.grants.length}{' '}
            {selectedRole.grants.length === 1 ? 'grant' : 'grants'}.
          </CardDescription>
        </CardHeader>
        <CardContent>
          {selectedRole.grants.map((grant) => (
            <ListRow
              key={`${grant.module}:${grant.action}@${grant.scope}`}
              icon={KeyRoundIcon}
              title={
                <span className="font-mono text-xs">
                  {grant.module}:{grant.action}@{grant.scope}
                </span>
              }
            />
          ))}
        </CardContent>
      </Card>
    </div>
  )
}
