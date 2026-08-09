import { LayoutGridIcon } from 'lucide-react'
import { Link } from 'react-router-dom'
import { getInitials } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'

// Personal app launcher (UI_PAGES.md §1) — not the metrics dashboard, that
// lives at Admin → Dashboard (features/admin/dashboard). Static placeholder
// apps; will come from the tenant's registered-apps API once lib/client
// exists (UI_CODING_STANDARDS.md §13). No organization context on a card —
// org selection happens inside the app's own authorization flow
// (USER_JOURNEYS_ORGANIZATIONS.md §3), not here.
interface HomeApp {
  id: string
  name: string
  description: string
  // Static relative-time display string for now; will be a real timestamp
  // (or null for "never signed in") once lib/client exists. Ordered here in
  // the recency order "Recently used" should display, since there's no real
  // timestamp yet to sort by.
  lastSignedIn: string | null
}

const apps: HomeApp[] = [
  {
    id: 'crm',
    name: 'Northbridge CRM',
    description: 'Customer relationship management for the sales team.',
    lastSignedIn: '2 hours ago',
  },
  {
    id: 'docuvault',
    name: 'DocuVault',
    description: 'Secure document storage and e-signatures.',
    lastSignedIn: '3 days ago',
  },
  {
    id: 'analytics',
    name: 'Analytics Suite',
    description: 'Product usage dashboards and reports.',
    lastSignedIn: '1 week ago',
  },
  {
    id: 'helpdesk',
    name: 'Helpdesk Pro',
    description: 'Support ticketing and customer chat.',
    lastSignedIn: null,
  },
  {
    id: 'teamsync',
    name: 'TeamSync',
    description: 'Internal chat and video meetings.',
    lastSignedIn: null,
  },
]

function AppCard({ app }: { app: HomeApp }) {
  return (
    <Card>
      <CardContent className="flex h-full flex-col gap-4">
        <div className="flex items-start gap-3">
          {/* App logo/icon placeholder — will render app.logo_url once
              lib/client exists (DATA_MODEL.md `apps.logo_url`); falls back to
              initials in the meantime. */}
          <div className="flex size-10 shrink-0 items-center justify-center rounded-md border border-border bg-muted text-sm font-semibold text-foreground">
            {getInitials(app.name)}
          </div>
          <div className="min-w-0">
            <p className="truncate text-sm font-semibold text-foreground">{app.name}</p>
            <p className="mt-0.5 line-clamp-2 text-xs text-muted-foreground">
              {app.description}
            </p>
          </div>
        </div>
        <p className="text-xs text-muted-foreground">
          {app.lastSignedIn ? `Last signed in ${app.lastSignedIn}` : 'Never signed in'}
        </p>
        <Button type="button" className="mt-auto w-full">
          Sign in
        </Button>
      </CardContent>
    </Card>
  )
}

function EmptyState() {
  return (
    <Card>
      <CardContent className="flex flex-col items-center gap-3 py-12 text-center">
        <div className="flex size-12 items-center justify-center rounded-md border border-border bg-muted text-muted-foreground">
          <LayoutGridIcon className="size-5" />
        </div>
        <div className="max-w-sm">
          <p className="text-sm font-medium text-foreground">No apps yet</p>
          <p className="mt-1 text-sm text-muted-foreground">
            Apps registered to your tenant will show up here once you&apos;re granted
            access to at least one.
          </p>
        </div>
        {/* Only shown to viewers who also hold admin permissions
            (UI_PAGES.md §1's empty-state note) — rendered unconditionally for
            now, same permission-gate-TODO pattern as app-shell.tsx, since
            there's no lib/client yet to check that. */}
        <Button asChild variant="outline" size="sm">
          <Link to="/admin/apps">Go to Admin → Apps</Link>
        </Button>
      </CardContent>
    </Card>
  )
}

function AppLauncher({ apps }: { apps: HomeApp[] }) {
  const recentlyUsed = apps.filter((app) => app.lastSignedIn !== null)
  const allApps = [...apps].sort((a, b) => a.name.localeCompare(b.name))

  return (
    <div className="flex flex-col gap-8">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight text-foreground">
          Welcome back
        </h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Sign in to any app you have access to.
        </p>
      </div>

      {apps.length === 0 ? (
        <EmptyState />
      ) : (
        <>
          {recentlyUsed.length > 0 && (
            <section className="flex flex-col gap-3">
              <h2 className="text-xs font-semibold tracking-[0.08em] text-muted-foreground uppercase">
                Recently used
              </h2>
              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
                {recentlyUsed.map((app) => (
                  <AppCard key={app.id} app={app} />
                ))}
              </div>
            </section>
          )}

          <section className="flex flex-col gap-3">
            <h2 className="text-xs font-semibold tracking-[0.08em] text-muted-foreground uppercase">
              All apps
            </h2>
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
              {allApps.map((app) => (
                <AppCard key={app.id} app={app} />
              ))}
            </div>
          </section>
        </>
      )}
    </div>
  )
}

export function HomePage() {
  return <AppLauncher apps={apps} />
}

// Not routed to — exported so the empty state (zero apps under the tenant)
// can be previewed without hand-editing the mock data above, e.g. by
// temporarily swapping <HomePage /> for <HomePageEmpty /> in App.tsx, or
// from a future home-page.test.tsx.
export function HomePageEmpty() {
  return <AppLauncher apps={[]} />
}
