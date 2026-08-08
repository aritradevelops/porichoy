# Technical Design: Porichoy

Companion to [PRD.md](./PRD.md). This document covers implementation-level decisions: stack,
architecture, data model, and operational concerns. Functional scope and product behavior
are defined in the PRD — this document assumes that scope and answers "how."

## 1. Architecture Overview

- **Style**: Hexagonal (ports & adapters). Core domain/business logic is isolated from
  delivery mechanisms and infrastructure. External interfaces — REST API, MCP server, etc.
  — are separate "ports" that all call into the same core logic, rather than duplicating
  business rules per interface.
- **Deployment shape**: Monolith. A single deployable Go application, backed by Postgres and
  Redis. Chosen for operational simplicity for self-hosters over independently-scaled
  services.
- **One unified UI, not a separate admin dashboard**: Porichoy ships a single default
  theme/SPA that serves every user-facing surface — login, signup, self-service account
  management (profile, MFA, password, sessions — PRD §9), the organization switcher (PRD
  §6), *and* administration. There is no distinct "admin dashboard" app. Admin
  functionality is simply shown or hidden within the same UI based on the logged-in user's
  permissions (the same `{module}:{action}@{scope}` model from §6 below) — a user with no
  admin permissions never sees admin options; a tenant superadmin sees the full surface.
- **Swappable theming**: the shipped UI is a default theme only. Because deployment is
  self-hosted, "the org's own backend" *is* the Porichoy instance itself — there is no
  separate proxy tier. An organization can keep the default theme as-is, or build and host
  its own custom theme against the same REST API / client SDKs, replacing the shipped UI
  entirely.
- **Login/signup flow variants**: the default theme ships with 2 layout variants for the
  login/credential-entry screen, explicitly chosen per tenant (admin picks in the UI, not
  inferred from the image):
  1. **Centered** — tenant logo + login form centered on the screen.
  2. **Split** — tenant's brand background image on one side, login form on the other.
  Both variants render the **tenant's** own logo and brand image (configured at the tenant
  level, §2) on the actual login screen. This is distinct from an app's own logo, which is
  configured per-app and shown separately on the OAuth **consent screen** ("App X is
  requesting access...") after the user has already authenticated against the tenant.
- **Programmatic/automation access**: In addition to the REST API, an MCP server is exposed
  as another port onto the same core logic, authenticated via API credentials generated
  through the same UI's admin surface.
- **Future monetization**: No paid/hosted tier is being built now, but the hexagonal
  structure is deliberately kept clean (core domain logic vs. adapters) so a future
  hosted-specific adapter could be added later without reworking the core.

## 2. Tech Stack

| Layer | Choice |
|---|---|
| Backend language | Go |
| Primary datastore | PostgreSQL |
| Cache provider | Pluggable (port/adapter); default implementation is Redis, other providers can be adopted by self-hosters |
| UI | Single default theme (SPA) covering login, signup, self-service account management, org switcher, and permission-gated admin — no separate admin dashboard. Fully swappable for a custom theme built against the REST API/SDKs. |
| API style | REST (JSON), URL path versioning (e.g. `/v1/...`) |
| Client SDKs | Go, Python, TypeScript — OAuth flow integration + management API calls (see §7.2) |
| Programmatic access | MCP server, as an additional port onto the core logic |
| Packaging | Docker Compose (primary quick-start); Go binary is natural given the language, self-hosters may run Postgres/Redis separately |
| Instance-level config | Config file (YAML) with env var overrides (12-factor style) |
| DB migrations | Go migration library (golang-migrate/goose) |
| Metrics export | Prometheus (`/metrics` endpoint) |
| Health checks | `/healthz` (liveness), `/readyz` (readiness — DB/Redis reachability) |
| License | MIT |

## 3. Multi-Tenancy & Data Model

### 3.1 Tenant Isolation Strategy
- **Shared tables + `tenant_id` column** on every tenant-scoped row. Chosen over
  schema-per-tenant or database-per-tenant for operational simplicity at current scale
  (small/medium — a handful of brand tenants and their apps, not hyperscale multi-tenant
  SaaS).
- Application-level enforcement of `tenant_id` scoping on all queries. (Postgres Row-Level
  Security was considered as a stronger guarantee but not chosen for v1 — revisit if
  cross-tenant data leakage risk becomes a concern at scale.)

### 3.2 Tenant Hierarchy Storage
- **Adjacency list**: each `Tenant` row stores a `parent_id`. Supports the arbitrary-depth
  hierarchy (root → brand tenant → optional sub-tenants) from the PRD.
- Subtree/ancestor queries use recursive CTEs. Acceptable at current scale; revisit
  (materialized path or closure table) only if hierarchy queries become a measured
  bottleneck.

### 3.3 Tenant Resolution by Origin
- Separate from app registration (`client_id`/`client_secret`, per PRD §4.3): tenants
  register their allowed domain(s)/origin(s) in a **dedicated domain registry**, independent
  of individual app records.
- Incoming API requests are resolved to a tenant by matching the request's `Origin` header
  against this registry. This is how a self-hosted instance serving multiple tenants (e.g.
  multiple brands) determines tenant context for calls coming from each org's own
  custom-built login UI.

### 3.4 Core Entities (baseline)

| Entity | Notes |
|---|---|
| `Tenant` | Adjacency-list hierarchy (`parent_id`). Root tenant has no parent. |
| `DomainRegistry` | Maps a registered origin/domain → tenant. Used for tenant resolution (§3.3). |
| `App` | OAuth2/OIDC client registered under a tenant. Owns its own signing key config (§4). Every tenant gets one auto-provisioned, non-deletable **default system app** at creation time (§3.5) — otherwise a regular row in this table, admin-editable like any other app. |
| `User` | Identity, scoped to exactly one tenant. |
| `Password` | Separate table from `User`, only present for users on the email+password method. Not 1:1 — a user has many rows over time via soft delete (exactly one active, `deleted_at IS NULL`); this also serves as the password history for reuse checks (PRD §9.3), no separate history table. |
| `MFAMethod` | Per-user, supports multiple enrolled methods (TOTP, WebAuthn, etc.) per PRD §9.2. |
| `Organization` | End-customer company, scoped to a tenant, shared across all apps under that tenant. |
| `OrgMembership` | User ↔ Organization, supports multi-org membership. |
| `Role` | App-scoped (individual sign-up apps) or org-scoped (organization-enabled apps), per PRD §7.2. Permissions (string identifiers, `{module}:{action}@{scope}`) and policies (opaque JSON, client-app-interpreted) are embedded directly as arrays on the role — not separate entities — to avoid join fan-out on the frequent path of editing a role's grants (DATA_MODEL.md `role`). |
| `RoleAssignment` | Binds `{principal, role}` — a principal is a user *or* an API credential, undifferentiated (DATA_MODEL.md §0); resources can be created/modified/deleted via API credentials exactly as by a logged-in user. Scope isn't stored here, it lives on each permission string within the role itself. The one exception is an optional org reference, needed only to disambiguate a shared/base role reused across multiple orgs (DATA_MODEL.md `role_assignment`). |
| `Session` | Source of truth in Postgres; backs the "view/revoke active sessions" self-service feature (PRD §9.4) and admin-triggered revocation. |
| `AuditLog` | Every API call (PRD §11); written via a pluggable log-sink port (§8 below), default Postgres, swappable for an external sink. |

Entities marked with "TBD" characteristics are expected to be refined as implementation
proceeds — this table is a baseline, not a final schema. See
[DATA_MODEL.md](./DATA_MODEL.md) for the actual table definitions (columns, types, keys,
ER diagram).

### 3.5 Default System App

Every tenant's token model is otherwise app-scoped (§4) — but tenant-level operations
(direct login, self-service account management PRD §9, the tenant's own session/MCP admin
context) have no third-party `App` to issue tokens against. Rather than build a separate
signing/verification path for this case, each tenant gets one **default system app**,
auto-provisioned when the tenant is created:

- Regular `App` row — same per-app signing key config (JWT secret, key pair, or JWKS) as any
  other app, admin-editable through the same UI.
- **Non-deletable**, and excluded from the normal app list shown to admins (or clearly
  badged as system-owned) so it isn't mistaken for a real integration.
- **No OAuth redirect/consent flow** — since the browser and Porichoy are the same
  first-party service here (no external `redirect_uri`, nothing to consent to), direct
  tenant login authenticates the user and issues the ID token + access token pair straight
  away against this app's signing config, without the Authorization Code + PKCE dance used
  for real third-party apps.
- Tokens issued against it carry **`aud` = the tenant** (not a generic client_id), distinct
  from app-scoped tokens (`aud` = that specific app's client_id) issued when a real app
  completes its own OAuth flow. This is what a tenant session token *is* under the hood —
  the same JWT structure and verification pipeline as every other token, just scoped to the
  tenant itself rather than to a third-party integration.

**Security note**: third-party apps never receive or handle a tenant-audience token — they
only ever get the app-scoped token pair from their own OAuth exchange — so there's no path
for a tenant-audience token to leak into a third party's internal API. The only requirement
this places on implementation is standard JWT hygiene: every endpoint must validate `aud`
against what it expects (tenant-level endpoints require `aud` = tenant; app-facing endpoints
like the runtime permissions API require `aud` = the calling app's own client_id) rather than
accepting any validly-signed Porichoy token.

See [AUTHORIZATION_MODEL.md](./AUTHORIZATION_MODEL.md) §4 for how the system app's
non-visibility to app-scoped admins is actually enforced at the query-filter level.

## 4. Tokens & Signing

- **Per-app signing configuration**: unlike a single instance-wide signing key, each app —
  including the tenant's default system app (§3.5) — chooses its own algorithm (RS256,
  RS512, ES256, etc.) and key material — either a shared secret, a bring-your-own
  public/private key pair, or a JWKS reference. The admin dashboard also supports
  provisioning new key material directly.
- **Two-tier audience model**: every token's `aud` claim is either the tenant itself (for
  tokens issued against the default system app — direct login, self-service account
  management) or a specific app's client_id (for tokens issued when a real app completes its
  own OAuth flow). Verification always checks `aud` against what the calling endpoint
  expects — see §3.5 for the security rationale.
- **Tokens issued at login** (per PRD §5.2): an **ID token** (client-side user identification,
  standard OIDC claims) and an **access token** (used by the client app to call Porichoy
  APIs on the user's behalf, including the runtime permissions API). The same pair is issued
  for a direct tenant login, scoped to the default system app instead of a third-party app.
- **Token lifetimes**: configurable **per app**, not per tenant — consistent with signing
  config already being per-app (this section). Every app, including the tenant's default
  system app (§3.5), sets its own access/ID/refresh token TTLs; a tenant-level login session
  simply uses whatever TTLs are configured on that tenant's default system app. Sane defaults
  (e.g. access/ID token ~15 minutes, refresh token ~30 days) — exact defaults to be finalized
  during implementation.
- **Refresh tokens**: rotate-on-use, with reuse detection — if an already-used refresh token
  is replayed, the entire token family is revoked.
- **Sessions**: Postgres is the source of truth (not just Redis), so that admin-triggered
  revocation and the self-service "view active sessions" feature (PRD §9.4) remain accurate
  even across cache eviction/restarts.

## 5. Authentication Methods — Implementation

- **Password hashing**: bcrypt.
- **Password reuse**: on change, the new password is hashed and checked against the user's
  last 4 `Password` rows (active + soft-deleted, by `created_at`) — a match is rejected. On
  success, the current row is soft-deleted (`deleted_at`/`deleted_by`) and a new active row
  is inserted; nothing is hard-deleted or pruned (DATA_MODEL.md `password`).
- **WebAuthn**: supports both passwordless primary login and use as an MFA factor, chosen
  per tenant/user configuration — consistent with the multi-method model in PRD §5.1.
- **TOTP secrets**: encrypted at rest using a single instance-wide encryption key (KMS-backed
  where available). Confirmed — kept intentionally simpler than the per-app signing-key
  pattern in §4; accepted as a deliberate inconsistency rather than revisited.
- **Social login (Google/Apple)**: per-tenant OAuth app credentials — each tenant registers
  its own client ID/secret with the provider, consistent with the per-app secrets model.
- **OTP delivery (email/SMS)**: pluggable, per-tenant configured — each tenant supplies its
  own provider credentials (e.g. SES, Twilio) via the admin UI. Detailed provider
  integration and recovery flows remain deferred to a future document per PRD §5.5.
- **Rate limiting** (login attempts, OTP requests): enforced in-app via the cache provider
  (sliding window/token bucket; default Redis, per §2), configurable per tenant.

## 6. Permissions Engine

- **Evaluation model**: precomputed and cached. A user's effective permissions and policies
  are materialized in the cache provider (default Redis, per §2) when roles/assignments
  change, rather than recomputed live on every call — the runtime permissions API reads from
  this cache.
- **Latency target**: no hard SLA set — design for correctness first, revisit once real usage
  patterns exist. Deliberately left open rather than a placeholder gap.
- **Roles vs. Permissions vs. Policies**: a role can carry permissions (string identifiers),
  policies (opaque JSON, app-defined semantics), or both — the app/org chooses how granular
  or freeform their authorization model is. Porichoy does not interpret policy content;
  it stores and returns it via the runtime permissions API for the client app to evaluate.
  Both are embedded as arrays directly on the role row, not normalized child tables
  (DATA_MODEL.md `role`) — deliberate, to keep role edits (the frequent operation) a
  single-row update.
- **Route convention, scope resolution, and query filtering**: see
  [AUTHORIZATION_MODEL.md](./AUTHORIZATION_MODEL.md) for how routes map to
  `{module}:{action}` checks, how the highest applicable scope is resolved
  (`own < org < app < tenant < root`), and how that scope becomes a per-module query filter.

## 7. API Design

- **Style**: REST (JSON) across the board — admin/management API, auth/self-service API
  (called by custom-built login UIs), and the runtime permissions API.
- **Versioning**: URL path versioning (`/v1/...`).
- **MCP**: positioned as a first-class management interface, equally important to the UI and
  REST API (PRD §8) — not just an automation bolt-on. Authenticated via API credentials
  generated through the same UI's admin surface, routed to the same core business logic as
  the REST port. Only a curated subset of management routes are exposed via MCP
  (availability), and every exposed call is still checked against the caller's
  `{module}:{action}@{scope}` permissions (authorization) — both gates apply independently,
  per PRD §8. See [MCP_TOOLS.md](./MCP_TOOLS.md) for the concrete tool list.
- **Management API auth model**: human admins (via the shipped UI) authenticate with their
  normal session/access token. Automation — scripts, CI, the MCP port — authenticates with
  separate API keys carrying their own rate-limit tier, distinct from end-user auth rate
  limiting.

### 7.2 Client SDKs

- Officially published SDKs for **Go, Python, and TypeScript**, wrapping the REST API.
- Cover two things:
  1. **OAuth2/OIDC flow integration** — lets a third-party app (an `App` under a tenant, per
     §3.4) implement "Login with Porichoy" without hand-rolling the Authorization Code +
     PKCE exchange, token refresh, etc.
  2. **Management operations** — convenience wrappers for role/permission administration:
     assigning roles, managing roles, creating permissions, creating roles, and similar.
- All management calls are authenticated with the caller's **access token**; Porichoy's
  server validates the token and checks the caller's permissions (via the same
  `{module}:{action}@{scope}` model, §6) before performing the action — the SDK itself
  enforces nothing, it's a thin client over the REST API.

## 8. Security

- **Encryption at rest**: an application-level encryption port/adapter (consistent with the
  hexagonal architecture) encrypts sensitive fields (phone numbers, TOTP secrets, provider
  API keys) before they reach Postgres, rather than relying on database-level encryption.
- **Secrets management**: env vars/config file by default, with a pluggable secrets-provider
  port so production deployments can swap in Vault/KMS without changing core logic.
- **CSRF/XSS responsibility split**: Porichoy's responsibility ends at its own API
  surface — CORS scoped to registered origins per the domain registry (§3.3), CSRF
  protection on state-changing calls, secure/SameSite cookie flags on any session cookies it
  sets. The client SDKs (§7.2) are thin API wrappers with no security opinions baked in;
  XSS/CSRF hardening of a custom-built UI is entirely the integrator's responsibility.
- **Compliance posture**: GDPR-aware by design (e.g. user data should be exportable/
  deletable) even though formal compliance reporting is out of PRD scope for v1.

## 9. Audit Logging

- **Scope: every API call** is logged, including auth events (login attempts, logouts, MFA
  challenges, password/OTP resets) and admin/configuration changes. Matches PRD.md §10.
- **Storage**: pluggable log-sink port/adapter — default implementation writes to Postgres,
  swappable for an external sink (e.g. a SIEM) in production, consistent with the
  ports/adapters architecture.
- **Retention**: configurable per tenant, default 90 days.
- **Access**: governed by the same `{module}:{action}@{scope}` permission model (e.g. an
  `audit:read@{scope}` permission) as everything else in §6.
- **Write path**: async/buffered — API requests are not blocked on audit log writes. Requests
  publish to a buffer/queue in front of the log-sink port, decoupling request latency from
  log volume, which matters given every API call (not just auth events) is logged.

## 10. Deployment & Operations

- **Packaging**: Docker Compose for quick-start self-hosting; a Kubernetes/Helm path and a
  standalone Go binary are both natural extensions but Docker Compose is the primary
  supported path for v1.
- **Configuration**: YAML config file for structured instance-level settings, with env var
  overrides for containerized/secret-injection workflows.
- **Migrations**: managed via a Go migration library (golang-migrate or goose), run as part
  of deploy/startup.
- **Scale target**: small/medium for now — designed for a handful of brand tenants and their
  apps at today's scale, not hyperscale multi-tenant SaaS. Revisit tenant-isolation and
  permission-caching choices (§3.1, §6) if this changes materially.

## 11. Observability

- **Metrics**: Prometheus-format `/metrics` endpoint.
- **Health checks**: `/healthz` (liveness) and `/readyz` (readiness — verifies DB/Redis
  connectivity), needed for both Docker Compose healthchecks and any future k8s probes.

## 12. Open Source Project

- **License**: MIT.
- **Structural intent**: no paid/hosted tier is being built now, but the hexagonal
  architecture is intentionally kept clean (core domain logic separated from
  delivery/infrastructure adapters) so that a future hosted-specific adapter could be
  layered on later without reworking the core — not over-engineered for this today, just
  kept in mind.

## 13. Open Items (TBD)

None outstanding — all items previously flagged in this document have been resolved and are
reflected inline above. Revisit this section if new open questions come up during
implementation.
