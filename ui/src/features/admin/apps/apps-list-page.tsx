import { useNavigate } from 'react-router-dom'
import { getInitials } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { apps } from './apps-data'

// Registered apps (UI_PAGES.md §6). Flat list — unlike tenants-page.tsx's
// root variant, selecting a row here navigates to a real route (3.2) rather
// than drilling in place, so there's no local "selected row" state; row/
// column conventions (font-mono technical identifiers, success/outline
// status badges) otherwise match tenants-page.tsx for consistency between
// the two table-based admin lists.
export function AppsListPage() {
  const navigate = useNavigate()

  return (
    <div className="flex flex-col gap-8">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight text-foreground">
            Apps
          </h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Apps registered under this tenant. Select one to configure branding and
            OAuth.
          </p>
        </div>
        {/* Non-functional until lib/client exists to actually create an app. */}
        <Button type="button">New app</Button>
      </div>

      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Type</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Created</TableHead>
                <TableHead>Client ID</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {apps.map((app) => (
                <TableRow
                  key={app.id}
                  onClick={() => navigate(`/admin/apps/${app.id}`)}
                  className="cursor-pointer"
                >
                  <TableCell>
                    <div className="flex items-center gap-3">
                      {/* App logo placeholder — will render app.logo_url once
                          lib/client exists (DATA_MODEL.md `apps.logo_url`). */}
                      <div className="flex size-8 shrink-0 items-center justify-center rounded-md border border-border bg-muted text-xs font-semibold text-foreground">
                        {getInitials(app.name)}
                      </div>
                      <span className="font-medium text-foreground">{app.name}</span>
                    </div>
                  </TableCell>
                  <TableCell>
                    <Badge variant="outline">
                      {app.supportsOrganizations
                        ? 'Organizations'
                        : 'Individual sign-up'}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    <Badge variant={app.status === 'active' ? 'success' : 'outline'}>
                      {app.status === 'active' ? 'Active' : 'Disabled'}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-muted-foreground">
                    {app.createdAt}
                  </TableCell>
                  <TableCell className="font-mono text-xs text-muted-foreground">
                    {app.clientId}
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
