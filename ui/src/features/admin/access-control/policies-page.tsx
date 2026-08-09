import { Badge } from '@/components/ui/badge'
import { Card, CardContent } from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { getRoleById } from './roles-data'

interface Policy {
  id: string
  name: string
  conditionSummary: string
  roleIds: string[]
}

// Policies are a NEW concept, not yet backed by a real data-model entity
// (UI_PAGES.md §11.2/§7.3 open item — no `policies` table exists in
// DATA_MODEL.md yet) — this whole page is mock-only until that lands, more so
// than the rest of Phase 4's static placeholder data. roleIds reference
// roles-data.ts (4.1) so "attached role(s)" reflects real role names rather
// than made-up strings.
const policies: Policy[] = [
  {
    id: 'business-hours',
    name: 'Business hours only',
    conditionSummary: 'Allow only between 09:00–18:00, tenant-local time',
    roleIds: ['tenant-admin'],
  },
  {
    id: 'office-ips',
    name: 'Office IPs only',
    conditionSummary: 'Caller IP must fall within 203.0.113.0/24',
    roleIds: ['super-admin', 'tenant-admin'],
  },
  {
    id: 'step-up-mfa',
    name: 'Step-up MFA required',
    conditionSummary: 'Caller must have completed MFA within the last 15 minutes',
    roleIds: ['tenant-admin', 'org-owner'],
  },
  {
    id: 'trusted-device',
    name: 'Trusted device only',
    conditionSummary: "Caller's device must be a previously-verified device",
    roleIds: ['app-admin'],
  },
]

// Conditional/attribute-based rules layered on top of a role's grants
// (UI_PAGES.md §7.3) — distinct from the role itself. Table: policy name,
// condition summary, attached role(s).
export function PoliciesPage() {
  return (
    <div className="flex flex-col gap-8">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight text-foreground">
          Policies
        </h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Conditional rules layered on top of a role's grants — not yet backed by a real
          entity (UI_PAGES.md §11.2), mock data only.
        </p>
      </div>

      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Policy</TableHead>
                <TableHead>Condition</TableHead>
                <TableHead>Attached to</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {policies.map((policy) => (
                <TableRow key={policy.id}>
                  <TableCell className="font-medium text-foreground">
                    {policy.name}
                  </TableCell>
                  <TableCell className="text-muted-foreground">
                    {policy.conditionSummary}
                  </TableCell>
                  <TableCell>
                    <div className="flex flex-wrap gap-1.5">
                      {policy.roleIds.map((roleId) => {
                        const role = getRoleById(roleId)
                        return (
                          <Badge key={roleId} variant="secondary">
                            {role?.name ?? roleId}
                          </Badge>
                        )
                      })}
                    </div>
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
