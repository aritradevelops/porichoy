# UI Implementation TODO

Ordered, hand-off-one-at-a-time checklist for building the pages specified in
[UI_PAGES.md](./UI_PAGES.md). Each item is scoped to be a single implementation session.

**Scope of every item below**: layout/elements only, matching what `login-page.tsx` and
`dashboard-page.tsx` already establish — static/placeholder data, no `lib/client` wiring
(that's a separate future pass per UI_CODING_STANDARDS.md §13's build order), permission
gates stubbed as "always visible" with a comment noting they should gate on real permissions
once `lib/client` exists (same pattern already used in `app-shell.tsx`). Don't add real
form submission, API calls, or state persistence — buttons/inputs can be visually complete
and non-functional.

**Order matters**: Phase 0 must land before any page in Phases 1–6, since it changes the
shell/routes every page mounts into. Phases 1–6 can otherwise be done in any order.

## Phase 0 — Shared groundwork (blocks everything else)

- [x] **0.1 Add missing shadcn primitives.** `table`, `tabs`, `badge` (for status/verification
      pills), `chart` (for the Dashboard trend line — pulls in `recharts`, not yet a
      dependency). Generate via `npx shadcn add <name>` into `src/components/ui/`, then
      restyle to match UI_CODING_STANDARDS.md §5.1 (sharp radius, Mono palette, Bold icon
      weight) rather than leaving shadcn defaults, same treatment already given to
      `avatar.tsx`/`dropdown-menu.tsx`.
- [x] **0.2 Build a reusable list-row component**, mirroring the shell mockup's `.list-row`
      (leading icon, title, subtext, right-aligned action). Used by Security's MFA methods
      and Active sessions (Phase 1), and Admin's Tenants/Apps tables (Phases 2–3).
- [x] **0.3 Restructure `app-shell.tsx` nav**: split into an ungrouped User section (Home,
      Profile, Security) followed by an "Administration" group-label section (Dashboard,
      Tenants, Apps, Access Control, Audit Logs). Access Control expands into a static
      `nav-sub` (Roles, Permissions, Policies) via the chevron-toggle pattern already
      sketched in the theme-preview artifact. Pick lucide icons for each new item.
- [x] **0.4 Update `App.tsx` route tree**: `/` → Home, `/profile`, `/security`, `/admin`
      (Dashboard), `/admin/tenants`, `/admin/apps`, `/admin/apps/:appId`,
      `/admin/access-control/roles`, `/admin/access-control/permissions`,
      `/admin/access-control/policies`, `/admin/audit-logs`. Move the current
      `DashboardPage` metrics content over to the new `/admin` route (see 2.1). `/login`
      stays as-is.

  **Done.** Built by a two-agent implement/review loop and independently verified: new
  primitives in `components/ui/{table,tabs,badge,chart}.tsx`, `components/list-row.tsx`,
  `app-shell.tsx` restructured, `App.tsx` route tree updated, `dashboard-page.tsx` moved to
  `features/admin/dashboard/`. Also fixed along the way: a global Bold icon stroke-width
  (`svg { stroke-width: 2.2 }` in `index.css`) and `tracking-[0.08em]` on group-label/table-
  header text to match the mockup's letter-spacing. One pre-existing, out-of-scope issue
  was flagged but left untouched: `card.tsx` uses `rounded-xl`, which isn't remapped by
  `index.css`'s `@theme inline` block (only xs/sm/md/lg/pill are), so it renders at
  Tailwind's default 12px instead of the Sharp scale's 6px ceiling — worth a follow-up.

## Phase 1 — User section

- [x] **1.1 Home** — `features/home/home-page.tsx`. UI_PAGES.md §1. App-card grid split into
      "Recently used" / "All apps"; each card has logo, name, description, last-signed-in,
      Sign in button. Empty state with conditional admin-link. Static list of 3–5 fake apps.
- [x] **1.2 Profile** — `features/account/profile-page.tsx`. UI_PAGES.md §2. Profile card,
      inline-editable fields (avatar, display name, email, phone) with verification pills.
      Inline edit can be visual-only (click → input appears, no save wiring yet).
- [x] **1.3 Security** — `features/account/security-page.tsx`. UI_PAGES.md §3. Three
      sections in order: Password (action + last-changed), MFA methods (list-row + primary
      badge + disabled-last-method state), Active sessions (list-row + "this device" badge).
      Static placeholder data throughout.

  **Done.** Verified independently. Home uses a Card grid (not list-row — doesn't fit
  multi-line launcher cards) with a separately-exported `HomePageEmpty` for previewing the
  empty state. Profile's `EditableField` is genuinely per-field (own `isEditing` state each),
  using `Badge` for verification pills. Security reuses `ListRow`/`Badge` throughout; MFA
  Remove uses `variant="destructive"` (caught by review — was `ghost`); current-device
  session row intentionally omits a Log out action (my call, not in the original spec —
  matches how most session-management UIs handle the row for your own active session).

## Phase 2 — Admin: Dashboard & Tenants

- [x] **2.1 Admin Dashboard** — `features/admin/dashboard/dashboard-page.tsx`. UI_PAGES.md
      §4. Migrate the existing `dashboard-page.tsx` stat-cards here, add one signups-over-time
      trend chart (static series, needs 0.1's `chart` primitive) and a recent-activity card
      (static rows, "View all" link to Audit Logs). Comment both scope variants (root
      platform-wide vs. tenant-scoped) even though only one renders for now.
- [x] **2.2 Admin Tenants** — `features/admin/tenants/tenants-page.tsx`. UI_PAGES.md §5.
      Build both variants with a comment on which a real scope check would pick: root
      variant (table of all tenants), tenant variant (settings form + domains list +
      provider-credentials list).

  **Done.** Verified independently. Both pages share an identical `Scope = 'root' |
  'tenant'` toggle pattern with cross-referencing comments (flagged as worth hoisting into a
  shared type if a third scope-adaptive admin page shows up, e.g. Apps). Dashboard's chart
  uses `chart.tsx` with the semantic `--accent` token; Tenants' table uses `table.tsx`,
  reusing its native `data-state="selected"` styling for click-to-select drill-down.
  Extending Dashboard pulled `recharts` in and pushed the main bundle chunk past Vite's
  500KB warning, so `App.tsx` now `React.lazy`/`Suspense`-loads the Dashboard and Tenants
  routes specifically (UI_CODING_STANDARDS.md §6) — verified via a real `vite build` that
  the split actually brings the main chunk back under the threshold.

## Phase 3 — Admin: Apps

- [x] **3.1 Apps list** — `features/admin/apps/apps-list-page.tsx`. UI_PAGES.md §6. Table:
      name/logo, type, status, created date, client ID. Static rows, "New app" primary
      action (non-functional).
- [x] **3.2 App detail shell** — `features/admin/apps/app-detail-page.tsx`. Tab bar
      (Branding / OAuth Configuration, needs 0.1's `tabs` primitive) wrapping 3.3/3.4;
      `:appId` route param just selects from static mock data.
- [x] **3.3 Branding tab** — `features/admin/apps/branding-tab.tsx`. Logo upload control
      (non-functional), display name field, login-layout choice (Centered/Split) as a
      segmented control or radio-card pair.
- [x] **3.4 OAuth Configuration tab** — `features/admin/apps/oauth-config-tab.tsx`. Client
      ID/secret + regenerate button, redirect URIs list, grant types/scopes, login-method
      toggles, organizations-enabled toggle, token/session TTL fields.

  **Done.** Verified independently. Apps list uses real-route navigation
  (`navigate(/admin/apps/:id)`), not the local-state drill-down Tenants uses — intentional
  divergence, documented inline, since Tenants' pattern doesn't fit a flat list. OAuth
  Configuration is sectioned into 6 Cards. **Spec bug caught and fixed during this phase**:
  UI_PAGES.md §6 originally implied per-app editable login-method toggles, which directly
  contradicts PRD §5.1 ("no per-app override" — tenant-level only). Resolved as a disabled
  read-only mirror of the tenant's configured methods with a link to Admin → Tenants; §6 has
  been corrected to document this as the permanent design, not a one-off judgment call.

## Phase 4 — Admin: Access Control

- [x] **4.1 Roles** — `features/admin/access-control/roles-page.tsx`. UI_PAGES.md §7.1.
      Table (name, scope, # assignments) → detail panel/drawer showing that role's
      `{module}:{action}@{scope}` grants.
- [x] **4.2 Permissions** — `features/admin/access-control/permissions-page.tsx`.
      UI_PAGES.md §7.2. Read-only table of `{module}:{action}` pairs — hand-derive a
      representative static list from AUTHORIZATION_MODEL.md §1's route convention.
- [x] **4.3 Policies** — `features/admin/access-control/policies-page.tsx`. UI_PAGES.md
      §7.3. Table: policy name, condition summary, attached role(s). Comment that this is
      mock-only until the data model gains a `policies` entity (UI_PAGES.md §10.2).

  **Done.** Verified independently. Roles' mock data spans all five AUTHORIZATION_MODEL.md
  §3 scope levels; drill-down uses Tenants' local-state pattern (reasoned as a better fit
  than Apps' real-route pattern, since a role's detail is a single read-only grants list,
  no tabs). Permissions has zero action affordances anywhere — confirmed deliberate per
  §7.2, not a gap. Policies' "attached roles" resolve to real role names via a shared
  `getRoleById` helper rather than duplicated strings.

## Phase 5 — Admin: Audit Logs

- [x] **5.1 Audit Logs** — `features/admin/audit-logs/audit-logs-page.tsx`. UI_PAGES.md §8.
      Table: actor, action, module, scope/target, IP, timestamp. Date-range filter control
      (visual only, non-functional).

  **Done.** Verified independently. All six columns present and correctly ordered; date
  range filter is genuinely functional (client-side, filters the 18 mock rows spanning
  ~2.5 weeks) with an empty state when a range matches nothing; deliberately no actor/
  module/result filters, matching the explicit out-of-scope note in §10.5. Reuses
  `Scope`/`SCOPE_LABEL` from Phase 4's roles-data.ts rather than redefining it. This closes
  out every `/admin/*` route — only Phase 6 (Login) remains.

## Phase 6 — Login

- [x] **6.1 Login method placeholders** — extend the existing
      `features/auth/login-page.tsx`. UI_PAGES.md §9. Add disabled OTP ("Send code") and
      WebAuthn ("Use security key") buttons alongside the existing disabled Google/Apple
      ones.
- [x] **6.2 Split login layout** — new component (e.g. a `variant` prop on
      `login-page.tsx`, or a sibling `login-split-page.tsx` — decide based on how much
      markup ends up shared). Brand image fills one half, the same form fills the other.

  **Done.** Verified independently, including live in-browser at both `/login` and
  `/login/split` (the latter a preview-only route — real deployments resolve exactly one
  variant per app, never both). The sign-in card was extracted verbatim into a shared
  `login-form.tsx`, used as-is by both layouts; existing `login-page.test.tsx` still passes
  unchanged, confirming no regression. OTP/WebAuthn buttons form a 2x2 grid with Google/
  Apple. Split's brand-image half uses the same neutral `bg-muted`/`border-border`
  placeholder convention as every other "stand-in for an uploaded asset" in the codebase
  (caught and fixed during review — it originally used an accent gradient), with no text/
  quote overlay per spec.

---

**All six phases complete.** Every page in [UI_PAGES.md](./UI_PAGES.md) now has a real,
design-system-consistent implementation. Nothing has been committed — see git status for
the full diff before deciding what to commit.

## Handoff notes (apply to every item)

- Follow the tone/format already set by `dashboard-page.tsx` and `login-page.tsx`: static
  data, comments citing the PRD/DATA_MODEL section behind each field, and a note on what's
  deferred to `lib/client`.
- Do one item per session/PR — don't batch multiple pages into a single change, so each is
  easy to review and hand off independently.
- If an item surfaces a design question UI_PAGES.md didn't answer, stop and ask rather than
  guessing — add the answer back into UI_PAGES.md §10 once resolved.
