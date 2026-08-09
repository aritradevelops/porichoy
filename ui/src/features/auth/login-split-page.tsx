import { ImageIcon } from 'lucide-react'
import { LoginForm } from './login-form'

// Split layout variant (UI_PAGES.md §9) — the tenant's uploaded brand image/
// illustration fills one half of the viewport, the same LoginForm
// (login-form.tsx, UI_IMPLEMENTATION_TODO Phase 6.2) fills the other. Which
// of this vs. the Centered variant (login-page.tsx) is used for a given app
// is an admin-time choice (Admin → Apps → Branding → login layout,
// branding-tab.tsx, Phase 3) resolved once lib/client exists — not a runtime
// branch, so this stays its own page/component rather than a variant prop on
// LoginPage. Routed at /login/split for now purely so the layout is
// reachable/inspectable; a real deployment only ever serves one variant for
// a given app, never both at distinct URLs simultaneously.
export function LoginSplitPage() {
  return (
    <div className="flex min-h-svh">
      <div className="bg-muted hidden border-r border-border lg:flex lg:w-1/2 lg:items-center lg:justify-center">
        {/* Tenant's uploaded brand image/illustration (tenant.brand_image_url)
            — static placeholder block, no real image asset or lib/client yet.
            Neutral bg-muted/border-border, matching every other stand-in for
            an uploaded asset in this codebase (Home's app-logo boxes, Apps
            list/detail, Branding tab's own logo box and login-layout preview
            panels) rather than a one-off accent treatment — accent stays
            reserved for actions/focus per UI_CODING_STANDARDS.md §5.1. Image/
            illustration only, no text or quote overlay (UI_PAGES.md §9's
            earlier design decision). */}
        <ImageIcon className="text-muted-foreground size-20" />
      </div>

      <div className="bg-background flex flex-1 items-center justify-center p-6">
        <div className="flex w-full max-w-sm flex-col gap-6">
          {/* Tenant logo placeholder — same as the Centered variant
              (login-page.tsx). */}
          <div className="bg-primary text-primary-foreground shadow-[var(--shadow-card)] mx-auto flex h-12 w-12 items-center justify-center rounded-lg text-lg font-semibold">
            P
          </div>

          <LoginForm />
        </div>
      </div>
    </div>
  )
}
