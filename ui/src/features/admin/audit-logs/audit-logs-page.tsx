import { useState } from 'react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Field, FieldLabel } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { SCOPE_LABEL, type Scope } from '@/features/admin/access-control/roles-data'

interface AuditLogEntry {
  id: string
  actor: string
  module: string
  action: string
  scope: Scope
  // Resolved target of the action within that scope — '—' when the scope
  // already fully identifies the subject (e.g. `own`, or a root-wide
  // system job with no single target).
  target: string
  ip: string
  timestamp: string // ISO 8601, UTC — sliced (not parsed as Date) for both
  // display and date-range filtering below, so there's no timezone drift
  // between what's shown and what's filtered.
}

// Every API call is audit logged (PRD §11) — UI, REST, or MCP alike, mixing
// admin/config changes (module:action pairs consistent with the vocabulary
// already used on the Roles/Permissions pages, Phase 4) with authentication
// events (login attempts, MFA challenges, password resets). Actors, tenants,
// and apps reuse the same names invented in Phases 2-3 (Acme Corp, Beta Inc,
// Northbridge Labs; Northbridge CRM, DocuVault, Helpdesk Pro) so the mock
// admin data reads as one consistent deployment rather than disconnected
// per-page fixtures. Spans ~2.5 weeks so the date-range filter has something
// real to demonstrate. Will come from the audit log API once lib/client
// exists (UI_CODING_STANDARDS.md §13).
const auditLogEntries: AuditLogEntry[] = [
  {
    id: 'a1',
    actor: 'Jane Doe',
    module: 'roles',
    action: 'assign',
    scope: 'tenant',
    target: 'Acme Corp',
    ip: '203.0.113.42',
    timestamp: '2026-08-09T09:12:04Z',
  },
  {
    id: 'a2',
    actor: 'Jane Doe',
    module: 'auth',
    action: 'login_success',
    scope: 'own',
    target: '—',
    ip: '203.0.113.42',
    timestamp: '2026-08-09T08:47:19Z',
  },
  {
    id: 'a3',
    actor: 'marco@acme.example',
    module: 'auth',
    action: 'login_failure',
    scope: 'own',
    target: '—',
    ip: '198.51.100.23',
    timestamp: '2026-08-08T22:03:55Z',
  },
  {
    id: 'a4',
    actor: 'Priya Nair',
    module: 'apps',
    action: 'create',
    scope: 'tenant',
    target: 'Beta Inc',
    ip: '198.51.100.7',
    timestamp: '2026-08-08T16:20:10Z',
  },
  {
    id: 'a5',
    actor: 'Marco Reyes',
    module: 'domains',
    action: 'verify',
    scope: 'tenant',
    target: 'Acme Corp',
    ip: '203.0.113.19',
    timestamp: '2026-08-08T11:05:33Z',
  },
  {
    id: 'a6',
    actor: 'System',
    module: 'sessions',
    action: 'expire',
    scope: 'app',
    target: 'DocuVault',
    ip: '—',
    timestamp: '2026-08-07T19:41:02Z',
  },
  {
    id: 'a7',
    actor: 'Jane Doe',
    module: 'apps',
    action: 'update',
    scope: 'app',
    target: 'Northbridge CRM',
    ip: '203.0.113.42',
    timestamp: '2026-08-07T14:18:47Z',
  },
  {
    id: 'a8',
    actor: 'Marco Reyes',
    module: 'auth',
    action: 'mfa_challenge',
    scope: 'own',
    target: '—',
    ip: '203.0.113.19',
    timestamp: '2026-08-06T10:52:29Z',
  },
  {
    id: 'a9',
    actor: 'Priya Nair',
    module: 'org_memberships',
    action: 'invite',
    scope: 'org',
    target: 'Marketing Org',
    ip: '198.51.100.7',
    timestamp: '2026-08-05T21:37:14Z',
  },
  {
    id: 'a10',
    actor: 'Jane Doe',
    module: 'tenants',
    action: 'update',
    scope: 'tenant',
    target: 'Acme Corp',
    ip: '203.0.113.42',
    timestamp: '2026-08-05T08:09:56Z',
  },
  {
    id: 'a11',
    actor: 'System',
    module: 'users',
    action: 'expire_draft',
    scope: 'tenant',
    target: 'Northbridge Labs',
    ip: '—',
    timestamp: '2026-08-04T17:26:38Z',
  },
  {
    id: 'a12',
    actor: 'Marco Reyes',
    module: 'auth',
    action: 'password_reset',
    scope: 'own',
    target: '—',
    ip: '203.0.113.19',
    timestamp: '2026-08-03T13:44:21Z',
  },
  {
    id: 'a13',
    actor: 'Priya Nair',
    module: 'apps',
    action: 'regenerate_secret',
    scope: 'app',
    target: 'Helpdesk Pro',
    ip: '198.51.100.7',
    timestamp: '2026-08-02T09:58:03Z',
  },
  {
    id: 'a14',
    actor: 'Jane Doe',
    module: 'roles',
    action: 'update',
    scope: 'tenant',
    target: 'Acme Corp',
    ip: '203.0.113.42',
    timestamp: '2026-07-31T15:12:47Z',
  },
  {
    id: 'a15',
    actor: 'unknown@baduser.example',
    module: 'auth',
    action: 'login_failure',
    scope: 'own',
    target: '—',
    ip: '192.0.2.88',
    timestamp: '2026-07-30T12:03:19Z',
  },
  {
    id: 'a16',
    actor: 'Marco Reyes',
    module: 'domains',
    action: 'verify',
    scope: 'tenant',
    target: 'Beta Inc',
    ip: '203.0.113.19',
    timestamp: '2026-07-28T09:30:52Z',
  },
  {
    id: 'a17',
    actor: 'System',
    module: 'audit_logs',
    action: 'purge',
    scope: 'root',
    target: '—',
    ip: '—',
    timestamp: '2026-07-25T18:47:06Z',
  },
  {
    id: 'a18',
    actor: 'Jane Doe',
    module: 'api_credentials',
    action: 'create',
    scope: 'tenant',
    target: 'Acme Corp',
    ip: '203.0.113.42',
    timestamp: '2026-07-23T11:15:38Z',
  },
]

function formatTimestamp(iso: string) {
  return iso.replace('T', ' ').replace('Z', '').slice(0, 16)
}

// Actor, action ({module}:{action}), module, resolved scope/target, IP
// address, timestamp (UI_PAGES.md §8). Date-range filter only — actor/module/
// result filters were discussed but explicitly left out of scope for this
// pass (UI_PAGES.md §10 open item 5); don't add them here.
export function AuditLogsPage() {
  const [fromDate, setFromDate] = useState('')
  const [toDate, setToDate] = useState('')

  const filteredEntries = auditLogEntries.filter((entry) => {
    const date = entry.timestamp.slice(0, 10)
    if (fromDate && date < fromDate) return false
    if (toDate && date > toDate) return false
    return true
  })

  return (
    <div className="flex flex-col gap-8">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight text-foreground">
          Audit Logs
        </h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Every API call — UI, REST, or MCP alike — including authentication events and
          admin/configuration changes (PRD §11).
        </p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Date range</CardTitle>
          <CardDescription>
            {filteredEntries.length} of {auditLogEntries.length} entries.
          </CardDescription>
        </CardHeader>
        <CardContent className="flex flex-wrap items-end gap-4">
          <Field className="w-40">
            <FieldLabel htmlFor="audit-from">From</FieldLabel>
            <Input
              id="audit-from"
              type="date"
              value={fromDate}
              onChange={(e) => setFromDate(e.target.value)}
            />
          </Field>
          <Field className="w-40">
            <FieldLabel htmlFor="audit-to">To</FieldLabel>
            <Input
              id="audit-to"
              type="date"
              value={toDate}
              onChange={(e) => setToDate(e.target.value)}
            />
          </Field>
          <Button
            type="button"
            variant="ghost"
            size="sm"
            disabled={!fromDate && !toDate}
            onClick={() => {
              setFromDate('')
              setToDate('')
            }}
          >
            Reset
          </Button>
        </CardContent>
      </Card>

      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Actor</TableHead>
                <TableHead>Action</TableHead>
                <TableHead>Module</TableHead>
                <TableHead>Scope / target</TableHead>
                <TableHead>IP address</TableHead>
                <TableHead>Timestamp</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredEntries.length === 0 ? (
                <TableRow>
                  <TableCell
                    colSpan={6}
                    className="py-8 text-center text-muted-foreground"
                  >
                    No entries in this date range.
                  </TableCell>
                </TableRow>
              ) : (
                filteredEntries.map((entry) => (
                  <TableRow key={entry.id}>
                    <TableCell className="font-medium text-foreground">
                      {entry.actor}
                    </TableCell>
                    <TableCell className="font-mono text-xs text-foreground">
                      {entry.module}:{entry.action}
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {entry.module}
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <Badge variant="outline">{SCOPE_LABEL[entry.scope]}</Badge>
                        <span className="text-xs text-muted-foreground">
                          {entry.target}
                        </span>
                      </div>
                    </TableCell>
                    <TableCell className="font-mono text-xs text-muted-foreground">
                      {entry.ip}
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {formatTimestamp(entry.timestamp)}
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  )
}
