# UI Coding Standards

Companion to [TECHNICAL_DESIGN.md](./TECHNICAL_DESIGN.md) (architecture, §1) and
[CODING_STANDARDS.md](./CODING_STANDARDS.md) (the `server/` equivalent of this document).
Applies to `ui/` specifically — the default theme SPA. Written before implementation starts,
mirroring how CODING_STANDARDS.md preceded `server/` implementation (CODING_STANDARDS.md §8).

## 1. Libraries

| Concern | Choice |
|---|---|
| Framework | React |
| Build tool | Vite |
| Language | TypeScript |
| Package manager | pnpm |
| Styling | Tailwind CSS |
| Component primitives | [shadcn/ui](https://ui.shadcn.com/) (Radix-based, generated into the repo, not an opaque dependency) |
| Server-state / data fetching | [TanStack Query](https://tanstack.com/query) |
| Client/UI state | React Context + `useReducer` |
| Routing | React Router |
| Forms | React Hook Form |
| Validation | Zod |
| i18n | react-i18next |
| Unit / component testing | Vitest + React Testing Library |
| Linting | ESLint |
| Formatting | Prettier |

> E2e testing (e.g. Playwright) is deliberately deferred, not skipped permanently — revisit
> once the auth/consent flow (§11, the highest-value flow to cover end-to-end) is far enough
> along to be worth it.

## 2. Architecture

- **Single app, not a workspace.** `ui/` ships one Vite app (the default theme) — no
  `apps/`/`packages/` split. The reusable pieces a self-hoster building a custom theme would
  want (§3) live as an internal folder within this same app, not a separately
  versioned/published package. Revisit if that folder's code outgrows "reference
  implementation to copy from."
- **`ui/`'s client lib is not `sdk/typescript`.** `sdk/typescript` (TECHNICAL_DESIGN.md §7.2)
  is for third-party *apps* integrating "Login with Porichoy" and calling Porichoy's
  management API from their own code. `ui/`'s internal client lib (§3) is a distinct thing:
  it talks to Porichoy's UI-facing endpoints (auth, self-service, admin) on behalf of the
  default theme itself, and doubles as the reference a self-hoster copies from when building
  a custom theme. The two are not layered on each other.
- **Feature-based folders**, mirroring the backend's bounded-context grouping
  (CODING_STANDARDS.md §3) rather than a generic layer split:
  ```
  src/
    features/
      auth/           # login, signup, MFA challenge, consent screen
      account/        # profile, MFA management, password, sessions
      organizations/  # org switcher, create/invite/leave
      admin/          # tenant/app/role management (lazy-loaded, §6)
    lib/
      client/         # API client + headless hooks/state machines (§3)
    components/       # shared, feature-agnostic UI primitives (shadcn/ui output lives here)
  ```
- Each `features/*` folder owns its own components and hooks, and calls into `lib/client` —
  it does not reach into another feature folder directly.

## 3. `lib/client` — the API / Custom-Theme Reference Layer

- **Scope: API client + headless logic, no styled components.** Typed wrappers over the
  UI-facing REST endpoints (auth, self-service, admin), plus framework-level building blocks
  like a `useAuthFlow` hook encoding the OAuth/consent step sequence (USER_JOURNEYS.md) as
  state, not markup. A custom theme brings its own components/design — this layer's job is
  only "talk to Porichoy correctly."
- Since it's not a separately published package (§2), a self-hoster building a custom theme
  is expected to read/copy from this folder rather than `npm install` it — document this
  explicitly in `ui/README.md` once it exists, so the intent isn't only implicit in the
  folder's existence.

## 4. State Management

- **Server state**: TanStack Query, for anything that comes from the REST API (profile,
  sessions, roles, tenant config, etc.) — caching, refetch-on-focus, and mutation/loading/
  error state come for free instead of hand-rolled per screen.
- **Client/UI state**: React Context + `useReducer` for state that isn't server data — active
  wizard step, modal open/closed, the in-progress org-switcher selection. No global state
  library; consistent with the backend's preference for explicit, framework-free wiring over
  DI/state frameworks (CODING_STANDARDS.md §2).

## 5. Styling & Components

- Tailwind CSS for layout/utility styling.
- shadcn/ui components, generated into `src/components/`, then customized in place — treated
  as owned source, not a dependency to upgrade.
- Per-tenant branding (logo, brand image, login layout, accent colors — DATA_MODEL.md
  `tenant`) is applied via **CSS custom properties set at runtime** from the fetched tenant
  config (resolved by origin, TECHNICAL_DESIGN.md §3.3), not baked in at build time — one
  build serves every tenant. Tailwind utilities reference these variables (e.g.
  `bg-[var(--tenant-accent)]`) rather than each tenant needing its own build.

## 6. Routing & Code Splitting

- React Router for all client-side routing.
- Admin routes (`features/admin/*`) are **lazy-loaded** via `React.lazy` — a user with no
  admin permissions never downloads that code, consistent with the permission-gated-
  visibility model (TECHNICAL_DESIGN.md §1) and the ui-ux-pro-max skill's performance
  priority (lazy loading, avoid shipping unused code).

## 7. Forms & Validation

- React Hook Form for all forms (uncontrolled inputs, no per-keystroke re-render cost).
- Zod schemas for validation — where a form's shape matches an API request/response, the
  same Zod schema is reused as the TypeScript source of truth for that shape rather than
  hand-duplicating a type and a validator.

## 8. Auth & Token Handling

- **Tokens live in HttpOnly, Secure, SameSite cookies set by Porichoy** — never in
  `localStorage`/`sessionStorage`, and never held only in a JS variable. Matches
  TECHNICAL_DESIGN.md §8's "secure/SameSite cookie flags on any session cookies it sets."
- Because tokens are cookie-based, state-changing requests need CSRF protection (also
  TECHNICAL_DESIGN.md §8) — `lib/client` (§3) is responsible for attaching whatever CSRF
  mechanism the backend expects (e.g. a double-submit header), so individual feature code
  never has to think about it per call.
- The UI is deployed behind the same reverse proxy as the API (§10) specifically so cookies
  stay same-site — a genuinely cross-origin deployment would need `SameSite=None` and revisit
  the CSRF approach.

## 9. i18n

- `react-i18next`, with the UI's **own translation keyspace** for its own copy (labels,
  buttons, help text) — separate from the backend's `DomainError` keys (CODING_STANDARDS.md
  §7).
- Backend error responses (`{ error: { key, message } }`, CODING_STANDARDS.md §4) are mapped
  through a small lookup in the UI's i18n setup rather than displaying `error.message`
  directly — keeps the UI able to localize a backend error into any locale it supports,
  independent of whatever locale the backend happened to resolve.

## 10. Deployment & Config

- **Served separately** (nginx/CDN) in front of the API, on the same origin/reverse-proxy as
  the API (§8) — not embedded into the Go binary. Self-hosters can swap in their own custom
  theme build here without touching the backend deployment at all.
- **Config**: build-time env vars via Vite (`import.meta.env`) — one build per Porichoy
  deployment (which still serves every tenant on it, §5), not a runtime-fetched config file.
  Revisit if a single prebuilt release artifact ever needs to be deployed to multiple
  environments without rebuilding.

## 11. Testing

- Unit/component tests: Vitest + React Testing Library, colocated `*.test.tsx` alongside the
  code they test — same colocation convention as the backend (CODING_STANDARDS.md §5).
- E2e: deferred (§1) — revisit once the auth/consent flow (USER_JOURNEYS.md) is far enough
  along to be worth covering end-to-end.

## 12. Linting, Formatting & Commenting

- ESLint + Prettier.
- Commenting conventions are inherited from CODING_STANDARDS.md §6 unchanged — intention over
  mechanics, minimal, exported/public API still gets a one-line doc comment (JSDoc
  equivalent).

## 13. Build Order

Following the backend's precedent (CODING_STANDARDS.md §8: domain models before adapters) —
for `ui/`: `lib/client` (§3) is established first (API client + headless flow hooks,
unstyled), before any `features/*` screen is built on top of it. This lets the auth flow's
state machine be written and unit-tested independent of any component markup.

## 14. Open Items

1. **CSRF mechanism specifics** (§8) — which exact scheme (double-submit cookie, custom
   header check) isn't decided yet; deferred to whoever implements the backend's CSRF
   handling first, since the UI side just needs to match it.
2. **`lib/client` extraction** (§2/§3) — stays an internal folder for v1; revisit turning it
   into a real published package if custom-theme adoption makes that worthwhile.
