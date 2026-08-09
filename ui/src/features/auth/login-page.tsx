import { LoginForm } from './login-form'

// Centered layout variant (TECHNICAL_DESIGN §1, UI_PAGES.md §9) — tenant logo
// + LoginForm centered on screen over a dotted background. The actual
// sign-in Card (social/challenge buttons, email/password fields, links)
// lives in login-form.tsx (UI_IMPLEMENTATION_TODO Phase 6.2), shared as-is
// with the Split variant (login-split-page.tsx) — this page only supplies
// the chrome around it. Which variant renders for a given app is an
// admin-time choice (Admin → Apps → Branding → login layout,
// branding-tab.tsx, Phase 3), not something decided per-request — so Split
// is a separate page/route, not a runtime branch inside this one.
export function LoginPage() {
  return (
    <div
      className="bg-background flex min-h-svh items-center justify-center p-6"
      style={{
        backgroundImage: 'radial-gradient(var(--border) 1.5px, transparent 1.5px)',
        backgroundSize: '28px 28px',
        backgroundPosition: '-14px -14px',
      }}
    >
      <div className="flex w-full max-w-sm flex-col gap-6">
        {/* Tenant logo placeholder — will render tenant.logo_url once tenant
            config is fetched by lib/client. */}
        <div className="bg-primary text-primary-foreground shadow-[var(--shadow-card)] mx-auto flex h-12 w-12 items-center justify-center rounded-lg text-lg font-semibold">
          P
        </div>

        <LoginForm />
      </div>
    </div>
  )
}
