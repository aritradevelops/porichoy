# UI Page Design Specs

Companion to [PRD.md](./PRD.md), [DATA_MODEL.md](./DATA_MODEL.md), and
[UI_CODING_STANDARDS.md](./UI_CODING_STANDARDS.md) §5.1 (Approved Visual Language — palette,
radius, density, component classes referenced below assume that token set). Catalogs what
each screen of the unified account/admin UI contains — layout and elements only, not
interaction flows (those belong in [user-journeys/](./user-journeys/)). Every nav item and
action button's visibility is permission-gated per the `{module}:{action}@{scope}` model
(AUTHORIZATION_MODEL.md) — a user simply doesn't see what they can't do.

## 0. Navigation Structure

One sidebar (256px, Flush treatment), two sections stacked, matching the `.nav-item` /
`.nav-group-label` / `.nav-sub` pattern already in the shell mockup:

- **User section** (ungrouped, top): Home, Profile, Security — visible to every logged-in
  user regardless of permissions.
- **Administration** (group label, below): Dashboard, Tenants, Apps, Access Control, Audit
  Logs. Access Control expands into a static sidebar sub-list: Roles, Permissions, Policies.
  Apps does **not** get a static sidebar sub-list — its Branding/OAuth Configuration are
  page-level tabs inside a specific app's detail view (§7), since that content is per-app,
  not fixed (see Open Items §11.4 if a static submenu was actually intended here instead).

## 1. Home (`/`)

Personal app launcher — not a metrics dashboard (that moved to Admin → Dashboard, §5).

- Page heading (e.g. "Welcome back").
- App cards grouped into two sections:
  - **Recently used** — sorted by last-signed-in, descending.
  - **All apps** — alphabetical.
- Each card: app logo/icon, app name, short description, "Last signed in {relative time}" /
  "Never signed in", primary "Sign in" button.
- No organization context shown on the card — org selection happens inside the app's own
  authorization flow (USER_JOURNEYS_ORGANIZATIONS.md §3), not here.
- Empty state (zero apps under this tenant): explanatory message, plus — only if the viewer
  also holds admin permissions — a link into Admin → Apps.

## 2. Profile (`/profile`)

Identity fields only; nothing password/session-related lives here (that's Security, §3).

- Single profile card (avatar + name + email, matching the existing shell mockup's
  `.profile-row` pattern), each field independently editable **inline** (click a field to
  edit just that one — no single "Edit profile" form).
- Fields (PRD §9.1): avatar/DP, display name, email address, phone number.
- Email/phone show a verification-state pill; changing either requires re-verification
  before the new value takes effect (PRD §9.1).

## 3. Security (`/security`)

Three sections, fixed order:

1. **Password**
   - Only shown for accounts using the email+password login method (PRD §9.3).
   - "Change password" action, plus a "Last changed {relative time}" line.
2. **MFA methods**
   - `.list-row` per method (icon, method name, "added {date} · last used {relative}"),
     a primary/default badge on one method, per-row "Remove" action.
   - When the tenant force-enables MFA and only one method remains, that method's Remove
     button is disabled with an explanatory tooltip (PRD §9.2 — can't drop below one).
   - Section-level "Add method" action.
3. **Active sessions**
   - `.list-row` per session: device/browser name+icon, approximate location, last-active
     time, a "This device" badge on the current session, per-row "Log out" action — except
     the current-device row itself, which gets no Log out action (ending your own active
     session from inside a list item is confusing UX and redundant with the sign-out
     control elsewhere in the shell; matches how most session-management UIs handle this
     row).

## 4. Admin → Dashboard (`/admin`)

Tenant-wide metrics (PRD §10). Scope-adaptive: a root-scoped viewer sees platform-wide
aggregates across every tenant; a tenant-scoped admin sees only their own tenant — same
page, content narrows per the caller's resolved scope (AUTHORIZATION_MODEL.md §2).

- Top row: three stat-cards — MAU, MFA adoption %, login success rate (each with a
  delta-vs-prior-period indicator, matching the existing `dashboard-page.tsx` pattern).
- One trend chart: signups over time (up to 1 year of history, PRD §10).
- Recent-activity card: the N most recent Audit Log entries (actor, action, timestamp),
  with a "View all" link into Audit Logs (§9).

## 5. Admin → Tenants (`/admin/tenants`)

Scope-adaptive:

- **Root**: table of every tenant on the deployment (name, domain, status, created date);
  selecting a row opens that tenant's own settings below.
- **Tenant-scoped admin**: lands directly on their own tenant's settings, no list.

Tenant settings content: tenant name/config, custom domains list (`domain_registries` —
domain, verification status), SSO/social provider credentials list
(`tenant_provider_credentials` — provider, configured status).

## 6. Admin → Apps (`/admin/apps`)

- **List**: table of registered apps — name/logo, type (org-enabled vs. individual
  sign-up, PRD §4.3), status (active/disabled), created date, client ID.
- Selecting a row opens that app's detail view, with two tabs:
  - **Branding**: logo upload, display name, login layout choice (Centered / Split — §10).
  - **OAuth Configuration**: client ID/secret (+ regenerate), redirect URIs, allowed grant
    types/scopes, a **read-only** mirror of the tenant's configured login methods
    (email+password, OTP, WebAuthn, Google, Apple) with a link to where they're actually
    configured, organizations-enabled toggle (PRD §4.3), token/session TTL settings (see
    Open Items §11.3 — PRD §12 currently defers the detailed TTL/rotation design to
    engineering).
    - **Correction**: login methods are configured **only at the tenant level** (PRD §5.1 —
      "no per-app override"); this page must not offer them as editable per-app toggles.
      The original draft of this spec incorrectly implied per-app toggles — caught during
      Phase 3 implementation and fixed here.

## 7. Admin → Access Control

Static sidebar sub-list: Roles, Permissions, Policies.

### 7.1 Roles (`/admin/access-control/roles`)

- Table: role name, scope level, number of assignments.
- Detail view per role: its `{module}:{action}@{scope}` grants.

### 7.2 Permissions (`/admin/access-control/permissions`)

- Read-only reference table of every `{module}:{action}` pair the platform defines
  (AUTHORIZATION_MODEL.md §1) — not admin-editable, since it's derived from routes/code,
  not configuration.

### 7.3 Policies (`/admin/access-control/policies`)

A **new** concept — a Policy is a conditional/attribute-based rule (e.g. business-hours-only,
office-IPs-only) layered on top of a role's grants, distinct from the role itself.

- Table: policy name, condition summary, which role(s) it's attached to.
- Not yet backed by a real entity — see Open Items §11.2.

## 8. Admin → Audit Logs (`/admin/audit-logs`)

- Columns: actor, action (`{module}:{action}`), module, resolved scope/target, IP address,
  timestamp.
- Filter: date range. (Actor/module/result filters were discussed but not confirmed in
  scope for this pass — see Open Items §11.5.)

## 9. Login (`/login`)

- **Centered** (existing, built): social buttons (Google, Apple) plus new placeholder
  buttons for OTP ("Send code") and WebAuthn ("Use security key") — all disabled until
  `lib/client` exists to fetch which methods a tenant actually enabled — then email/password
  form, forgot-password link, sign-up link.
- **Split** (new): tenant's uploaded brand image/illustration fills one half (static image
  only, no text/quote overlay), the same form content fills the other half.
- Which layout renders is a per-app admin-time choice (Admin → Apps → Branding → login
  layout, §6), not runtime-detected.

## 10. Open Items / Flags for Follow-Up

1. **Routing change needed**: `App.tsx`'s `/` currently renders the old metrics
   `DashboardPage`. This doc reassigns `/` → Home (app launcher, §1) and moves metrics to
   `/admin` (§4) — needs a route/component reshuffle pass.
2. **Policies is a new entity**, not present in DATA_MODEL.md's Authorization section (§5).
   Needs a data-model addition (e.g. a `policies` table plus its attachment to `roles`)
   before §7.3 can be real rather than a static mock.
3. **OAuth Configuration's token/session TTL settings** (§6) exposes a knob that PRD §12
   currently defers to engineering as an open design question — worth confirming the admin
   UI should expose this now vs. leaving it out until that design lands.
4. **Apps vs. Access Control submenu mechanics differ**: Access Control's three children are
   static sidebar entries (§0), while Apps' Branding/OAuth Configuration are per-app page
   tabs, not sidebar entries — this is an interpretation based on which content is
   per-instance vs. fixed, not something explicitly confirmed. Flag if a static sidebar
   submenu was actually intended for Apps too.
5. **Audit Logs filters** (§8): only date range was confirmed; actor/module/result filters
   were discussed as options but not selected — revisit once the page exists and log volume
   is known.
6. **"Finance" → "Tenants" correction**: an early pass at this nav used "Finance" for what
   turned out to be the "Tenants" admin section (voice-transcription artifact) — noted here
   so it doesn't resurface confused with PRD §12's billing-is-out-of-scope note, which still
   stands unchanged.
