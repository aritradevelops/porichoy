# PRD: Porichoy — Open Source Identity Provider

## 1. Overview

Porichoy is an open source, self-hosted Identity Provider (IdP) that lets an organization
run centralized authentication, authorization, and user management for multiple internal
brands and the applications under them, while exposing standard OAuth2/OIDC so those
applications can integrate without owning their own auth stack.

Porichoy is **MCP-first**: every management capability (tenant, app, role, and
permission administration) is available not just through a UI and REST API, but as a
first-class MCP server — administrators and their tools/agents can manage the IdP
conversationally, not only by clicking through a dashboard.

## 2. Goals

- Provide a single, self-hosted authentication system that multiple brands and apps can
  share, without forcing shared user identities across brands.
- Let each brand configure its own login methods, MFA policy, and permission model
  independently of other brands.
- Support both individual end-user apps and B2B-style apps with customer organizations.
- Give brand/app operators visibility into signup, activity, and security metrics.
- Let end users manage their own account (profile, credentials, MFA, sessions) without
  admin involvement.
- Make MCP a first-class, equally-important interface alongside the UI and REST API — not a
  bolt-on — so administration can happen through whatever tool (dashboard, script, or
  AI agent) fits the operator's workflow.

## 3. Core Concepts & Terminology

| Term | Definition |
|---|---|
| **Root tenant** | The top of the hierarchy. Managed by root-level superadmins. Owns creation/deletion of brand tenants and system-wide config. |
| **Tenant** | A node in an arbitrarily deep hierarchy (e.g. Root → Brand A → ...). A tenant is the actual **authentication realm** — users log in *to a tenant*, not to an app. |
| **Brand tenant** | A tenant representing a business unit (e.g. Brand A, Brand B) that owns one or more apps. |
| **App** | An OAuth2/OIDC **client** registered under a tenant. Apps do not perform authentication themselves — they redirect to their tenant for login and receive tokens back. |
| **User (identity)** | An account that exists at the tenant level. A user's identity is **isolated per brand tenant** — the same person needs separate accounts to use apps under different brands. |
| **Organization** | An end-customer company using an app (e.g. "Acme Corp" using Brand A App 1). Distinct from the tenant hierarchy. Shared across all apps within the same tenant — i.e. "Acme Corp" is a single org record usable by any app under Brand A. |
| **Role** | A named bundle of permissions. Scope of role authorship depends on app type (see §7). |
| **Permission** | Format: `{module}:{action}@{scope}` — defines what action a role grants on what module, and at what scope (e.g. org, app, tenant) it applies. |

## 4. Tenant & Multi-Tenancy

### 4.1 Hierarchy
- Arbitrary-depth tenant tree, rooted at a single root tenant (e.g. Root → Brand A → [future sub-divisions] → Apps).
- Root-level superadmins create/delete brand tenants and configure system-wide settings.
- Each tenant node can own apps directly and/or contain child tenants.

### 4.2 Identity Isolation
- A user account belongs to exactly one tenant.
- No automatic identity sharing across sibling tenants (e.g. a Brand A user and a Brand B
  user with the same email are unrelated accounts).

### 4.3 App Registration
- Self-service dashboard: a tenant admin registers an app under their tenant and receives
  OAuth2/OIDC client credentials (client_id/secret, redirect URIs, etc.).
- Apps do not configure their own login methods — see §5.1.

## 5. Authentication

### 5.1 Login Method Configuration
- Login methods are configured **only at the tenant level**: email + password, email + OTP,
  phone number + OTP, WebAuthn, Google login, Apple login.
- All apps under a tenant inherit the tenant's configured methods exactly — **no per-app
  override**. If a tenant enables only Google login, every app under it only offers Google
  login.
- A tenant may enable multiple methods simultaneously; the end user chooses which one to
  use at signup/login (standard "sign in with..." selector).

### 5.2 Protocol
- Full OAuth2 + OIDC support: Authorization Code + PKCE, Client Credentials grant, Refresh
  Tokens, standard OIDC discovery and JWKS endpoints.
- On completion of the OAuth flow, the app receives two tokens:
  - **ID token** — used by the client app to authenticate/identify the user locally (who is
    this user, per standard OIDC claims).
  - **Access token** — used by the client app to make requests to Porichoy on the user's
    behalf, including calls to the runtime permissions API (§7.3). This is a single
    general-purpose access token, not a separate token per API.
- Session/token lifetime, TTL configuration, and refresh token rotation mechanics are
  deferred to technical design (not specified in this PRD beyond "must exist and be
  revocable by admins").

### 5.3 Account Linking
- If a user signs up/logs in via two different enabled methods using the same **verified**
  email address, the accounts are automatically linked/treated as one account.

### 5.4 Multi-Factor Authentication (MFA)
- MFA is a tenant-level policy: a tenant admin can require MFA for all users of that tenant.
- MFA enrollment and management is otherwise self-service (see §9.2).

### 5.5 OTP Delivery & Account Recovery
- Out of scope for this PRD. OTP provider selection (email/SMS), rate limiting, OTP expiry,
  forgot-password, and lost-MFA-device recovery flows are deferred to a separate technical
  design document.

## 6. Organizations

- An **organization** represents an end-customer company using one or more apps under a
  tenant (e.g. "Acme Corp" using Brand A App 1 and Brand A App 2).
- Organizations are created **self-service** by end users; the creator becomes the org's
  initial owner/admin.
- An organization is scoped to the tenant, not to a single app — created once, usable by
  any app under that tenant.
- A single user can be a member of multiple organizations and switch between them (workspace-
  switcher style UX), similar to Slack.
- Apps that don't support organizations operate purely on individual sign-ups (see §7).

## 7. Permissions, Roles & Policies

### 7.1 Permission Model
- Every permission follows the format **`{module}:{action}@{scope}`**.
  - `module`: the resource/feature area being governed (e.g. `billing`, `users`, `metrics`).
  - `action`: the operation being permitted (e.g. `read`, `write`, `assign`).
  - `scope`: the level at which the permission applies (e.g. a specific org, an app, a
    tenant) — this is what determines whether a given admin can act only within their org,
    across an entire app, or across an entire tenant.
- This single primitive governs **all** authorization decisions in the system, including:
  who can assign roles to other users, and who can view metrics at what scope.

### 7.2 Role Authorship (differs by app type)
- **Individual sign-up apps** (no organizations): roles are **app-scoped** — defined and
  owned by the app owner (the brand). End users are assigned these roles directly by
  whoever holds the relevant `roles:assign@app` permission.
- **Organization-enabled apps**: the app owner defines a base set of roles; **organization
  owners can customize/extend roles** within their own organization. Roles in this context
  are **organization-scoped**.
- Role assignment itself is not tied to a fixed persona (e.g. "org admin") — it is gated by
  whoever holds the corresponding `roles:assign@{scope}` permission, consistent with §7.1.

### 7.3 Enforcement
- Client apps do not receive full role/permission data embedded in tokens. Instead, apps
  call a **runtime permissions API**, authenticated with the user's **access token** (see
  §5.2), to fetch the user's current roles/permissions for authorization decisions.

## 8. Management Interfaces

Porichoy exposes its management capabilities (tenant, app, role, permission
administration, and more) through three parallel, equally-important interfaces — none is
positioned as primary:

- **UI** — the default shipped theme's permission-gated admin surface (detailed in the
  technical design).
- **REST API** — the underlying API the UI itself is built on.
- **MCP server** — a first-class interface for conversational/agent-driven administration,
  covering the same underlying operations as the REST API.

MCP is **not** a 1:1 mirror of the full REST management API. What's callable via MCP is
governed by two independent gates:
1. **Availability** — only a curated subset of management routes are exposed through MCP at
   all; not every REST route has an MCP equivalent.
2. **Authorization** — for whatever is exposed, the same `{module}:{action}@{scope}`
   permission model (§7.1) applies. An MCP caller can only perform actions their credentials
   actually grant, identical to how UI/REST callers are authorized.

See [MCP_TOOLS.md](./MCP_TOOLS.md) for the concrete tool surface — the first pass covers
the tenant/app/role setup flow specifically, as the most repetitive, documentation-heavy
part of administering Porichoy today.

## 9. Self-Service Account Management

Users manage their own account within their tenant (this is tenant-scoped, not per-app,
consistent with identity living at the tenant level).

### 9.1 Profile
- Editable fields: display name, profile picture (DP), email address, phone number.
- Changing email or phone number requires re-verification of the new value before it takes
  effect.

### 9.2 MFA Self-Management
- Users can enroll and remove multiple MFA methods themselves (e.g. TOTP app + WebAuthn key
  simultaneously).
- If the tenant force-enables MFA, the user cannot remove their last remaining MFA method —
  at least one must always stay active.

### 9.3 Password Management
- Available only to users whose account uses the email + password login method.
- Changing password requires re-entering the current password.
- New password must not match any of the user's last 4 passwords.

### 9.4 Session Management
- Users can view a list of their active sessions/devices and revoke (log out) individual
  sessions remotely.

## 10. Metrics & Analytics

- Dashboard-only in v1 — no CSV export or metrics API.
- Metrics tracked:
  - New user signups over time (up to 1 year of history).
  - Active users (DAU/MAU).
  - Login success/failure rates, broken down by login method.
  - MFA adoption rate.
- Access to metrics is governed by the same `{module}:{action}@{scope}` permission model
  (§7.1) — e.g. an org-scoped permission shows only that org's data; an app- or tenant-
  scoped permission shows data across all orgs/apps within that scope.

## 11. Audit Logging

- In scope for v1: **every API call** is audit logged, including authentication events
  (login attempts, logouts, MFA challenges, password/OTP resets) as well as admin/
  configuration changes (role edits, tenant/app/org config changes) and any other API
  activity — regardless of whether the call came via UI, REST, or MCP.

## 12. Out of Scope (v1)

- Billing / subscription / seat management.
- Metrics export (CSV/API).
- Detailed OTP delivery and account recovery flow design (separate technical design doc).
- Detailed session/token TTL and refresh rotation design (left to engineering/technical
  design).
- Self-service account deletion/deactivation.

## 13. Open Questions

None outstanding — all ambiguities identified during review have been resolved and are
reflected above. Revisit this PRD if scope changes during implementation.
