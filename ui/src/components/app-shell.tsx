import { useState } from 'react'
import {
  Building2Icon,
  ChevronRightIcon,
  ChevronsUpDownIcon,
  FileClockIcon,
  HomeIcon,
  KeyRoundIcon,
  LayoutDashboardIcon,
  LayoutGridIcon,
  LogOutIcon,
  ScrollTextIcon,
  ShieldCheckIcon,
  ShieldIcon,
  UserIcon,
  UsersRoundIcon,
  type LucideIcon,
} from 'lucide-react'
import { NavLink, Outlet, useLocation, useNavigate } from 'react-router-dom'
import { cn, getInitials } from '@/lib/utils'
import { useAuth } from '@/lib/client/auth-context'
import { hasPermission } from '@/lib/client/permissions'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { ThemeToggle } from '@/components/theme-toggle'

interface NavItem {
  to: string
  label: string
  icon: LucideIcon
  end?: boolean
  // Set only for items backed by a permission the authorization model actually has a catalog
  // entry for today (AUTHORIZATION_MODEL.md's {module}:{action}@{scope}) — see adminNavItems'
  // own comment for why the rest of this list doesn't have one yet.
  permission?: { module: string; action: string }
}

// User section (ungrouped, top) — visible to every logged-in user regardless
// of permissions (UI_PAGES.md §0), so unlike the Administration items below,
// these don't need a future permission-gate comment.
const userNavItems: NavItem[] = [
  { to: '/', label: 'Home', icon: HomeIcon, end: true },
  { to: '/profile', label: 'Profile', icon: UserIcon },
  { to: '/security', label: 'Security', icon: ShieldCheckIcon },
]

// Administration group (UI_PAGES.md §0), permission-gated per item
// (AUTHORIZATION_MODEL.md's {module}:{action}@{scope} grants) via hasPermission below.
//
// Only Tenants maps to a permission any seeded role actually grants
// (cmd/seed/main.go's superAdminPermissions/tenantAdminPermissions include "tenants:*@...").
// Dashboard/Apps have no backing module or REST endpoint yet at all — gating them on an
// invented permission string would just hide them for every user, including root, since
// nothing could ever grant it. Gate for real once each module ships its own backend
// authorization; until then they stay unconditionally visible, same as before this pass.
const adminNavItems: NavItem[] = [
  { to: '/admin', label: 'Dashboard', icon: LayoutDashboardIcon, end: true },
  {
    to: '/admin/tenants',
    label: 'Tenants',
    icon: Building2Icon,
    permission: { module: 'tenants', action: 'list' },
  },
  { to: '/admin/apps', label: 'Apps', icon: LayoutGridIcon },
]

// Access Control's static sub-list (UI_PAGES.md §7) — same "no backing module yet" reasoning
// as Dashboard/Apps above; none of roles/permissions/policies has a REST endpoint or
// permission catalog entry today.
const accessControlSubItems: NavItem[] = [
  { to: '/admin/access-control/roles', label: 'Roles', icon: UsersRoundIcon },
  {
    to: '/admin/access-control/permissions',
    label: 'Permissions',
    icon: KeyRoundIcon,
  },
  { to: '/admin/access-control/policies', label: 'Policies', icon: ScrollTextIcon },
]

// No backing module yet — same reasoning as above.
const auditLogsItem: NavItem = {
  to: '/admin/audit-logs',
  label: 'Audit Logs',
  icon: FileClockIcon,
}

function visibleNavItems(items: NavItem[], permissions: string[]) {
  return items.filter(
    (item) =>
      !item.permission ||
      hasPermission(permissions, item.permission.module, item.permission.action),
  )
}

function navLinkClassName({ isActive }: { isActive: boolean }) {
  return cn(
    'flex items-center gap-2 rounded-md px-3 py-2 text-sm font-medium transition-colors',
    isActive
      ? 'bg-sidebar-primary text-sidebar-primary-foreground'
      : 'text-sidebar-foreground hover:bg-sidebar-accent',
  )
}

// Sidebar (256px) + slim top bar, sidebar treatment: Flush — full-height,
// edge-to-edge, 1px right border (UI_CODING_STANDARDS.md §5.1).
export function AppShell() {
  const { pathname } = useLocation()
  const navigate = useNavigate()
  const accessControlActive = pathname.startsWith('/admin/access-control')
  const [accessControlOpen, setAccessControlOpen] = useState(accessControlActive)
  const { user, logout } = useAuth()
  const permissions = user?.permissions ?? []

  return (
    <div className="bg-background flex min-h-svh">
      <aside className="bg-sidebar border-sidebar-border flex w-(--sidebar-width) shrink-0 flex-col border-r">
        {/* Org switcher — the trigger label shows the caller's real tenant (from /auth/me),
            but the dropdown's item list stays a placeholder: Organizations aren't a built
            feature yet (no orgs endpoint, no features/organizations/), so listing real orgs
            here would misrepresent scope this pass doesn't cover. */}
        <div className="border-sidebar-border flex h-(--topbar-height) items-center border-b px-4">
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <button
                type="button"
                className="hover:bg-sidebar-accent flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left transition-colors"
              >
                <div className="bg-primary text-primary-foreground flex h-6 w-6 shrink-0 items-center justify-center rounded-md text-xs font-semibold">
                  P
                </div>
                <span className="text-sidebar-foreground flex-1 truncate text-sm font-medium">
                  {user?.tenant_name ?? 'Loading…'}
                </span>
                <ChevronsUpDownIcon className="text-muted-foreground size-4 shrink-0" />
              </button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="start" className="w-56">
              <DropdownMenuItem>Acme Corp</DropdownMenuItem>
              <DropdownMenuItem>Beta Inc</DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem>Create organization</DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>

        <nav className="flex flex-1 flex-col gap-1 overflow-y-auto p-3">
          {userNavItems.map(({ to, label, icon: Icon, end }) => (
            <NavLink key={to} to={to} end={end} className={navLinkClassName}>
              <Icon className="size-4" />
              {label}
            </NavLink>
          ))}

          <div className="text-muted-foreground px-3 pt-4 pb-1 text-[11px] font-semibold tracking-[0.08em] uppercase">
            Administration
          </div>

          {visibleNavItems(adminNavItems, permissions).map(
            ({ to, label, icon: Icon, end }) => (
              <NavLink key={to} to={to} end={end} className={navLinkClassName}>
                <Icon className="size-4" />
                {label}
              </NavLink>
            ),
          )}

          <button
            type="button"
            onClick={() => setAccessControlOpen((open) => !open)}
            aria-expanded={accessControlOpen}
            className={cn(
              'flex w-full items-center gap-2 rounded-md px-3 py-2 text-left text-sm font-medium transition-colors',
              accessControlActive
                ? 'bg-sidebar-primary text-sidebar-primary-foreground'
                : 'text-sidebar-foreground hover:bg-sidebar-accent',
            )}
          >
            <ShieldIcon className="size-4" />
            <span className="flex-1">Access Control</span>
            <ChevronRightIcon
              className={cn(
                'size-4 shrink-0 transition-transform duration-150',
                accessControlOpen && 'rotate-90',
              )}
            />
          </button>
          {accessControlOpen && (
            <div className="flex flex-col gap-1 pl-6">
              {visibleNavItems(accessControlSubItems, permissions).map(
                ({ to, label, icon: Icon }) => (
                  <NavLink key={to} to={to} className={navLinkClassName}>
                    <Icon className="size-4" />
                    {label}
                  </NavLink>
                ),
              )}
            </div>
          )}

          <NavLink to={auditLogsItem.to} className={navLinkClassName}>
            <auditLogsItem.icon className="size-4" />
            {auditLogsItem.label}
          </NavLink>
        </nav>

        {/* User menu — real identity from /auth/me, sourced via useAuth(). */}
        <div className="border-sidebar-border border-t p-3">
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <button
                type="button"
                className="hover:bg-sidebar-accent flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left transition-colors"
              >
                <Avatar className="size-7">
                  <AvatarFallback className="text-xs">
                    {getInitials(user?.display_name ?? user?.email ?? '')}
                  </AvatarFallback>
                </Avatar>
                <div className="flex-1 truncate">
                  <p className="text-sidebar-foreground truncate text-sm font-medium">
                    {user?.display_name ?? user?.email ?? 'Loading…'}
                  </p>
                  <p className="text-muted-foreground truncate text-xs">
                    {user?.email ?? ''}
                  </p>
                </div>
              </button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="start" className="w-56">
              <DropdownMenuItem asChild>
                <NavLink to="/profile">
                  <UserIcon /> Profile
                </NavLink>
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem
                variant="destructive"
                onClick={() => {
                  logout()
                  navigate('/login')
                }}
              >
                <LogOutIcon /> Sign out
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </aside>

      <div className="flex flex-1 flex-col">
        <header className="border-border flex h-(--topbar-height) shrink-0 items-center justify-end border-b px-6">
          <ThemeToggle />
        </header>
        <main className="flex-1 overflow-y-auto p-8">
          <div className="mx-auto max-w-(--content-max-width)">
            <Outlet />
          </div>
        </main>
      </div>
    </div>
  )
}
