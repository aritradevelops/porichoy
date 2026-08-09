import { useState } from 'react'
import { GlobeIcon, PlugIcon } from 'lucide-react'
import { ListRow } from '@/components/list-row'
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

// Scope-adaptive (UI_PAGES.md §5): a root-scoped viewer sees every tenant on
// the deployment and picks one to drill into; a tenant-scoped admin lands
// directly on their own tenant's settings, no list. Same static-toggle
// pattern as dashboard-page.tsx, for the same reason (no lib/client yet to
// resolve a real scope, AUTHORIZATION_MODEL.md §2) — flip to 'tenant' to
// preview that variant.
type Scope = 'root' | 'tenant'
const viewerScope: Scope = 'root'

interface Tenant {
  id: string
  name: string
  domain: string
  status: 'active' | 'disabled'
  createdAt: string
  mfaRequired: boolean
  loginLayout: 'Centered' | 'Split'
  domains: { domain: string; verified: boolean }[]
  providerCredentials: { provider: string; configured: boolean }[]
}

// Static placeholder tenants — will come from the tenants list API once
// lib/client exists (UI_CODING_STANDARDS.md §13).
const tenants: Tenant[] = [
  {
    id: 'acme',
    name: 'Acme Corp',
    domain: 'acme.porichoy.example',
    status: 'active',
    createdAt: '2025-11-02',
    mfaRequired: true,
    loginLayout: 'Centered',
    domains: [
      { domain: 'login.acme.com', verified: true },
      { domain: 'auth.acme.io', verified: false },
    ],
    providerCredentials: [
      { provider: 'Google', configured: true },
      { provider: 'Apple', configured: false },
      { provider: 'Email OTP', configured: true },
    ],
  },
  {
    id: 'beta',
    name: 'Beta Inc',
    domain: 'beta.porichoy.example',
    status: 'active',
    createdAt: '2026-01-18',
    mfaRequired: false,
    loginLayout: 'Split',
    domains: [{ domain: 'id.betainc.com', verified: true }],
    providerCredentials: [{ provider: 'Google', configured: true }],
  },
  {
    id: 'northbridge',
    name: 'Northbridge Labs',
    domain: 'northbridge.porichoy.example',
    status: 'disabled',
    createdAt: '2026-04-30',
    mfaRequired: false,
    loginLayout: 'Centered',
    domains: [],
    providerCredentials: [],
  },
]

// "Their own tenant" for the tenant-scoped variant preview — a real
// implementation would resolve this from the caller's scope context instead.
const ownTenantId = 'acme'

// Shared by both scope variants once "in" a tenant (UI_PAGES.md §5): name/
// config, custom domains (`domain_registries`), SSO/social provider
// credentials (`tenant_provider_credentials`). No lib/client yet, so "Save
// changes" is visual-only.
function TenantSettings({ tenant }: { tenant: Tenant }) {
  return (
    <div className="flex flex-col gap-6">
      <Card>
        <CardHeader>
          <CardTitle>Tenant configuration</CardTitle>
          <CardDescription>Name, MFA policy, and login layout.</CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col gap-4">
          <Field className="max-w-sm">
            <FieldLabel htmlFor="tenant-name">Tenant name</FieldLabel>
            <Input id="tenant-name" defaultValue={tenant.name} />
          </Field>

          <div className="flex flex-wrap items-center gap-x-6 gap-y-3 text-sm">
            <div className="flex items-center gap-2">
              <span className="text-muted-foreground">Domain</span>
              <span className="font-mono text-xs text-foreground">{tenant.domain}</span>
            </div>
            <div className="flex items-center gap-2">
              <span className="text-muted-foreground">Status</span>
              <Badge variant={tenant.status === 'active' ? 'success' : 'outline'}>
                {tenant.status === 'active' ? 'Active' : 'Disabled'}
              </Badge>
            </div>
            <div className="flex items-center gap-2">
              <span className="text-muted-foreground">MFA</span>
              <Badge variant={tenant.mfaRequired ? 'secondary' : 'outline'}>
                {tenant.mfaRequired ? 'Required' : 'Optional'}
              </Badge>
            </div>
            <div className="flex items-center gap-2">
              <span className="text-muted-foreground">Login layout</span>
              <Badge variant="outline">{tenant.loginLayout}</Badge>
            </div>
          </div>

          <div>
            <Button type="button" size="sm">
              Save changes
            </Button>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Custom domains</CardTitle>
          <CardDescription>Origins that resolve to this tenant.</CardDescription>
        </CardHeader>
        <CardContent>
          {tenant.domains.length === 0 ? (
            <p className="text-sm text-muted-foreground">
              No custom domains registered.
            </p>
          ) : (
            tenant.domains.map((d) => (
              <ListRow
                key={d.domain}
                icon={GlobeIcon}
                title={d.domain}
                action={
                  <Badge variant={d.verified ? 'success' : 'warning'}>
                    {d.verified ? 'Verified' : 'Pending'}
                  </Badge>
                }
              />
            ))
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>SSO &amp; social providers</CardTitle>
          <CardDescription>Login providers configured for this tenant.</CardDescription>
        </CardHeader>
        <CardContent>
          {tenant.providerCredentials.length === 0 ? (
            <p className="text-sm text-muted-foreground">
              No provider credentials configured.
            </p>
          ) : (
            tenant.providerCredentials.map((p) => (
              <ListRow
                key={p.provider}
                icon={PlugIcon}
                title={p.provider}
                action={
                  <Badge variant={p.configured ? 'success' : 'outline'}>
                    {p.configured ? 'Configured' : 'Not configured'}
                  </Badge>
                }
              />
            ))
          )}
        </CardContent>
      </Card>
    </div>
  )
}

// Root variant: table of every tenant; selecting a row drills into its
// settings below (UI_PAGES.md §5). A real implementation would navigate to a
// per-tenant route instead of local state — this is the simplest way to
// demonstrate the drill-down without adding routing/state machinery Phase 2
// doesn't need yet.
function TenantsRootView() {
  const [selectedTenantId, setSelectedTenantId] = useState(tenants[0].id)
  const selectedTenant =
    tenants.find((tenant) => tenant.id === selectedTenantId) ?? tenants[0]

  return (
    <div className="flex flex-col gap-6">
      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Domain</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Created</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {tenants.map((tenant) => (
                <TableRow
                  key={tenant.id}
                  data-state={tenant.id === selectedTenantId ? 'selected' : undefined}
                  onClick={() => setSelectedTenantId(tenant.id)}
                  className="cursor-pointer"
                >
                  <TableCell className="font-medium text-foreground">
                    {tenant.name}
                  </TableCell>
                  <TableCell className="font-mono text-xs text-muted-foreground">
                    {tenant.domain}
                  </TableCell>
                  <TableCell>
                    <Badge variant={tenant.status === 'active' ? 'success' : 'outline'}>
                      {tenant.status === 'active' ? 'Active' : 'Disabled'}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-muted-foreground">
                    {tenant.createdAt}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      <TenantSettings tenant={selectedTenant} />
    </div>
  )
}

export function TenantsPage() {
  const ownTenant = tenants.find((tenant) => tenant.id === ownTenantId) ?? tenants[0]

  return (
    <div className="flex flex-col gap-8">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight text-foreground">
          Tenants
        </h1>
        <p className="mt-1 text-sm text-muted-foreground">
          {viewerScope === 'root'
            ? 'Every tenant on this deployment. Select one to view its settings.'
            : `Settings for ${ownTenant.name}.`}
        </p>
      </div>

      {viewerScope === 'root' ? (
        <TenantsRootView />
      ) : (
        <TenantSettings tenant={ownTenant} />
      )}
    </div>
  )
}
