import { ActivityIcon } from 'lucide-react'
import { Link } from 'react-router-dom'
import { CartesianGrid, Line, LineChart, XAxis } from 'recharts'
import { ListRow } from '@/components/list-row'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
  type ChartConfig,
} from '@/components/ui/chart'

// Tenant-wide metrics (PRD §10). Scope-adaptive: a root-scoped viewer sees
// platform-wide aggregates across every tenant, a tenant-scoped admin sees
// only their own tenant — same page, content narrows per the caller's
// resolved scope (AUTHORIZATION_MODEL.md §2's `own < org < app < tenant <
// root` ranking). There's no lib/client yet to resolve a real scope, so both
// variants are built behind this static toggle — flip to 'root' to preview
// the platform-wide variant. tenants-page.tsx uses the same toggle pattern
// for consistency between the two scope-adaptive admin pages.
type Scope = 'root' | 'tenant'
const viewerScope: Scope = 'tenant'

interface Stat {
  label: string
  value: string
  delta: string
  positive: boolean
}

const statsByScope: Record<Scope, Stat[]> = {
  root: [
    { label: 'MAU', value: '48,210', delta: '+6%', positive: true },
    { label: 'MFA adoption', value: '57%', delta: '+3%', positive: true },
    { label: 'Login success rate', value: '98.2%', delta: '−0.4%', positive: false },
  ],
  tenant: [
    { label: 'MAU', value: '8,940', delta: '+4%', positive: true },
    { label: 'MFA adoption', value: '62%', delta: '−2%', positive: false },
    { label: 'Login success rate', value: '99.1%', delta: '+0.2%', positive: true },
  ],
}

// New signups over time, up to 1 year of history (PRD §10). Static monthly
// series — will come from the metrics API once lib/client exists.
const signupSeriesByScope: Record<Scope, { month: string; signups: number }[]> = {
  root: [
    { month: 'Sep', signups: 3120 },
    { month: 'Oct', signups: 3340 },
    { month: 'Nov', signups: 3580 },
    { month: 'Dec', signups: 3410 },
    { month: 'Jan', signups: 3720 },
    { month: 'Feb', signups: 3890 },
    { month: 'Mar', signups: 4050 },
    { month: 'Apr', signups: 4210 },
    { month: 'May', signups: 4180 },
    { month: 'Jun', signups: 4460 },
    { month: 'Jul', signups: 4620 },
    { month: 'Aug', signups: 4890 },
  ],
  tenant: [
    { month: 'Sep', signups: 210 },
    { month: 'Oct', signups: 236 },
    { month: 'Nov', signups: 248 },
    { month: 'Dec', signups: 229 },
    { month: 'Jan', signups: 262 },
    { month: 'Feb', signups: 271 },
    { month: 'Mar', signups: 288 },
    { month: 'Apr', signups: 301 },
    { month: 'May', signups: 294 },
    { month: 'Jun', signups: 318 },
    { month: 'Jul', signups: 332 },
    { month: 'Aug', signups: 356 },
  ],
}

const chartConfig = {
  signups: {
    label: 'New signups',
    color: 'var(--accent)',
  },
} satisfies ChartConfig

interface AuditEntry {
  id: string
  actor: string
  action: string
  tenant?: string
  timestamp: string
}

// N most recent Audit Log entries (actor, action, timestamp) — same shape
// Audit Logs itself will use once it's real (Phase 5, UI_PAGES.md §8).
const recentActivityByScope: Record<Scope, AuditEntry[]> = {
  root: [
    {
      id: 'a1',
      actor: 'Jane Doe',
      action: 'roles:assign',
      tenant: 'Acme Corp',
      timestamp: '2 hours ago',
    },
    {
      id: 'a2',
      actor: 'Priya Nair',
      action: 'apps:create',
      tenant: 'Beta Inc',
      timestamp: '5 hours ago',
    },
    {
      id: 'a3',
      actor: 'System',
      action: 'tenants:create',
      tenant: 'Northbridge Labs',
      timestamp: '1 day ago',
    },
    {
      id: 'a4',
      actor: 'Marco Reyes',
      action: 'domains:verify',
      tenant: 'Acme Corp',
      timestamp: '2 days ago',
    },
  ],
  tenant: [
    { id: 'a1', actor: 'Jane Doe', action: 'roles:assign', timestamp: '2 hours ago' },
    { id: 'a2', actor: 'Jane Doe', action: 'apps:update', timestamp: '1 day ago' },
    {
      id: 'a3',
      actor: 'Marco Reyes',
      action: 'domains:verify',
      timestamp: '2 days ago',
    },
    {
      id: 'a4',
      actor: 'Marco Reyes',
      action: 'mfa_methods:remove',
      timestamp: '4 days ago',
    },
  ],
}

export function DashboardPage() {
  const stats = statsByScope[viewerScope]
  const signups = signupSeriesByScope[viewerScope]
  const recentActivity = recentActivityByScope[viewerScope]

  return (
    <div className="flex flex-col gap-8">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight text-foreground">
          {viewerScope === 'root' ? 'Platform overview' : 'Welcome back'}
        </h1>
        <p className="mt-1 text-sm text-muted-foreground">
          {viewerScope === 'root'
            ? 'Aggregates across every tenant on this deployment.'
            : "Here's what's happening with Acme Corp."}
        </p>
      </div>

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
        {stats.map((stat) => (
          <Card key={stat.label}>
            <CardContent className="flex flex-col gap-1.5">
              <p className="text-sm text-muted-foreground">{stat.label}</p>
              <div className="flex items-baseline gap-2">
                <span className="text-3xl font-semibold text-foreground">
                  {stat.value}
                </span>
                <span
                  className={
                    stat.positive
                      ? 'text-sm font-medium text-success'
                      : 'text-sm font-medium text-destructive'
                  }
                >
                  {stat.delta}
                </span>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Signups over time</CardTitle>
          <CardDescription>
            {viewerScope === 'root'
              ? 'New users across every tenant, last 12 months.'
              : 'New users in Acme Corp, last 12 months.'}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <ChartContainer config={chartConfig} className="aspect-auto h-64 w-full">
            <LineChart data={signups} margin={{ left: 12, right: 12 }}>
              <CartesianGrid vertical={false} />
              <XAxis dataKey="month" tickLine={false} axisLine={false} tickMargin={8} />
              <ChartTooltip content={<ChartTooltipContent />} />
              <Line
                dataKey="signups"
                type="monotone"
                stroke="var(--color-signups)"
                strokeWidth={2}
                dot={false}
              />
            </LineChart>
          </ChartContainer>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Recent activity</CardTitle>
          <CardDescription>
            {viewerScope === 'root'
              ? 'Latest audit log entries across every tenant.'
              : 'Latest audit log entries.'}
          </CardDescription>
          <CardAction>
            <Button asChild variant="ghost" size="sm">
              <Link to="/admin/audit-logs">View all</Link>
            </Button>
          </CardAction>
        </CardHeader>
        <CardContent>
          {recentActivity.map((entry) => (
            <ListRow
              key={entry.id}
              icon={ActivityIcon}
              title={entry.actor}
              subtext={
                entry.tenant
                  ? `${entry.action} · ${entry.tenant} · ${entry.timestamp}`
                  : `${entry.action} · ${entry.timestamp}`
              }
            />
          ))}
        </CardContent>
      </Card>
    </div>
  )
}
