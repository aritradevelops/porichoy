import { lazy, Suspense } from 'react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import { AppShell } from '@/components/app-shell'
import { ThemeProvider } from '@/components/theme-provider'
import { AuthProvider } from '@/lib/client/auth-context'
import { RequireAuth, RedirectIfAuthenticated } from '@/lib/client/require-auth'
import { LoginPage } from '@/features/auth/login-page'
import { LoginSplitPage } from '@/features/auth/login-split-page'
import { HomePage } from '@/features/home/home-page'
import { ProfilePage } from '@/features/account/profile-page'
import { SecurityPage } from '@/features/account/security-page'

// features/admin/* is React.lazy-loaded (UI_CODING_STANDARDS.md §6) — a user
// with no admin permissions never downloads it. Every /admin/* route now has
// real content (Phases 2-5) and its own lazy import below.
const DashboardPage = lazy(() =>
  import('@/features/admin/dashboard/dashboard-page').then((m) => ({
    default: m.DashboardPage,
  })),
)
const TenantsPage = lazy(() =>
  import('@/features/admin/tenants/tenants-page').then((m) => ({
    default: m.TenantsPage,
  })),
)
const AppsListPage = lazy(() =>
  import('@/features/admin/apps/apps-list-page').then((m) => ({
    default: m.AppsListPage,
  })),
)
const AppDetailPage = lazy(() =>
  import('@/features/admin/apps/app-detail-page').then((m) => ({
    default: m.AppDetailPage,
  })),
)
const RolesPage = lazy(() =>
  import('@/features/admin/access-control/roles-page').then((m) => ({
    default: m.RolesPage,
  })),
)
const PermissionsPage = lazy(() =>
  import('@/features/admin/access-control/permissions-page').then((m) => ({
    default: m.PermissionsPage,
  })),
)
const PoliciesPage = lazy(() =>
  import('@/features/admin/access-control/policies-page').then((m) => ({
    default: m.PoliciesPage,
  })),
)
const AuditLogsPage = lazy(() =>
  import('@/features/admin/audit-logs/audit-logs-page').then((m) => ({
    default: m.AuditLogsPage,
  })),
)

function AdminRouteFallback() {
  return <p className="text-sm text-muted-foreground">Loading…</p>
}

const queryClient = new QueryClient()

// Route tree, matching UI_PAGES.md §0's nav structure.
//
// AuthProvider must sit inside QueryClientProvider (it uses TanStack Query for /auth/me) and
// outside BrowserRouter's Routes (RequireAuth/RedirectIfAuthenticated need Router context to
// read the current location and redirect).
export function App() {
  return (
    <ThemeProvider>
      <QueryClientProvider client={queryClient}>
        <AuthProvider>
          <BrowserRouter>
            <Routes>
              <Route
                path="/login"
                element={
                  <RedirectIfAuthenticated>
                    <LoginPage />
                  </RedirectIfAuthenticated>
                }
              />
              {/* Split login layout preview (UI_PAGES.md §9, Phase 6.2) — not a
                real deployment route, just a way to reach the variant for
                inspection. A real deployment resolves exactly one of
                Centered/Split for a given app's login_layout, never both. */}
              <Route
                path="/login/split"
                element={
                  <RedirectIfAuthenticated>
                    <LoginSplitPage />
                  </RedirectIfAuthenticated>
                }
              />
              <Route
                element={
                  <RequireAuth>
                    <AppShell />
                  </RequireAuth>
                }
              >
                {/* User section (UI_PAGES.md §0, UI_IMPLEMENTATION_TODO Phase 1). */}
                <Route path="/" element={<HomePage />} />
                <Route path="/profile" element={<ProfilePage />} />
                <Route path="/security" element={<SecurityPage />} />

                {/* Administration section (UI_PAGES.md §0, UI_IMPLEMENTATION_TODO
                  Phases 2-5) — every route here is a real, lazy-loaded page. */}
                <Route
                  path="/admin"
                  element={
                    <Suspense fallback={<AdminRouteFallback />}>
                      <DashboardPage />
                    </Suspense>
                  }
                />
                <Route
                  path="/admin/tenants"
                  element={
                    <Suspense fallback={<AdminRouteFallback />}>
                      <TenantsPage />
                    </Suspense>
                  }
                />
                <Route
                  path="/admin/apps"
                  element={
                    <Suspense fallback={<AdminRouteFallback />}>
                      <AppsListPage />
                    </Suspense>
                  }
                />
                <Route
                  path="/admin/apps/:appId"
                  element={
                    <Suspense fallback={<AdminRouteFallback />}>
                      <AppDetailPage />
                    </Suspense>
                  }
                />
                <Route
                  path="/admin/access-control/roles"
                  element={
                    <Suspense fallback={<AdminRouteFallback />}>
                      <RolesPage />
                    </Suspense>
                  }
                />
                <Route
                  path="/admin/access-control/permissions"
                  element={
                    <Suspense fallback={<AdminRouteFallback />}>
                      <PermissionsPage />
                    </Suspense>
                  }
                />
                <Route
                  path="/admin/access-control/policies"
                  element={
                    <Suspense fallback={<AdminRouteFallback />}>
                      <PoliciesPage />
                    </Suspense>
                  }
                />
                <Route
                  path="/admin/audit-logs"
                  element={
                    <Suspense fallback={<AdminRouteFallback />}>
                      <AuditLogsPage />
                    </Suspense>
                  }
                />
              </Route>
              <Route path="*" element={<Navigate to="/" replace />} />
            </Routes>
          </BrowserRouter>
        </AuthProvider>
      </QueryClientProvider>
    </ThemeProvider>
  )
}
