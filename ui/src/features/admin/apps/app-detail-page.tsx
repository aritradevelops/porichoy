import { ArrowLeftIcon } from 'lucide-react'
import { Link, useParams } from 'react-router-dom'
import { getInitials } from '@/lib/utils'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { BrandingTab } from './branding-tab'
import { getAppById } from './apps-data'
import { OAuthConfigTab } from './oauth-config-tab'

// App detail shell (UI_PAGES.md §6) — looks the app up from the same static
// mock data apps-list-page.tsx renders, by the :appId route param, and wraps
// the Branding (3.3) / OAuth Configuration (3.4) tabs using the tabs.tsx
// primitive from Phase 0.
export function AppDetailPage() {
  const { appId } = useParams<{ appId: string }>()
  const app = getAppById(appId)

  if (!app) {
    return (
      <div className="flex flex-col gap-2">
        <h1 className="text-2xl font-semibold tracking-tight text-foreground">
          App not found
        </h1>
        <p className="text-sm text-muted-foreground">
          No app matches &quot;{appId}&quot;.{' '}
          <Link
            to="/admin/apps"
            className="text-primary underline-offset-4 hover:underline"
          >
            Back to Apps
          </Link>
          .
        </p>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-6">
      <Link
        to="/admin/apps"
        className="inline-flex w-fit items-center gap-1.5 text-sm text-muted-foreground transition-colors hover:text-foreground"
      >
        <ArrowLeftIcon className="size-4" />
        Back to Apps
      </Link>

      <div className="flex items-center gap-4">
        <div className="flex size-12 shrink-0 items-center justify-center rounded-md border border-border bg-muted text-base font-semibold text-foreground">
          {getInitials(app.name)}
        </div>
        <div>
          <h1 className="text-2xl font-semibold tracking-tight text-foreground">
            {app.name}
          </h1>
          <p className="text-sm text-muted-foreground">
            Client ID: <span className="font-mono text-xs">{app.clientId}</span>
          </p>
        </div>
      </div>

      <Tabs defaultValue="branding">
        <TabsList>
          <TabsTrigger value="branding">Branding</TabsTrigger>
          <TabsTrigger value="oauth">OAuth Configuration</TabsTrigger>
        </TabsList>
        <TabsContent value="branding" className="mt-4">
          <BrandingTab app={app} />
        </TabsContent>
        <TabsContent value="oauth" className="mt-4">
          <OAuthConfigTab app={app} />
        </TabsContent>
      </Tabs>
    </div>
  )
}
