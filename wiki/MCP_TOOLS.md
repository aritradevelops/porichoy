# MCP Tool Surface

**Status: draft, first pass** — scoped specifically to automating the tenant/app/role setup
flow (USER_JOURNEYS_ADMIN_TENANT_MANAGEMENT.md §2–§5), since that's the most repetitive,
documentation-heavy part of administering Porichoy today. Other tool categories (metrics
queries, audit log queries, member/invite management, etc.) aren't covered in this pass.

Companion to [PRD.md](./PRD.md) §8 (Management Interfaces),
[TECHNICAL_DESIGN.md](./TECHNICAL_DESIGN.md) §7 (API Design),
[AUTHORIZATION_MODEL.md](./AUTHORIZATION_MODEL.md), and
[user-journeys/USER_JOURNEYS_ADMIN_TENANT_MANAGEMENT.md](./user-journeys/USER_JOURNEYS_ADMIN_TENANT_MANAGEMENT.md).

## 1. The Problem This Solves

Setting up a new brand tenant today means reading through several docs and clicking
through several UI screens in a specific order: create tenant → register domain →
configure branding/login methods/MFA → register app(s) → define roles → designate a
default signup role. It's a repeated, mechanical process, not a creative one — a good fit
for an agent to walk an admin through conversationally from a terminal, asking the
necessary questions along the way, instead of the admin needing to already know the right
order and every required field.

## 2. Design Choice: Granular Tools, Not One Wizard

The setup "flow" isn't implemented as a single composite tool. Each tool below maps to one
step from the admin journey doc, mirroring the same `{module}:{action}` vocabulary as the
REST API (AUTHORIZATION_MODEL §1) — the *flow* emerges from an agent calling several of
these in sequence during a conversation, asking the human for whatever input each step
needs, rather than one rigid scripted tool with a fixed question order.

> **Flagging the alternative rather than silently picking**: a single `setup_tenant` mega-tool
> (taking one large structured payload) was considered and rejected here — it would be less
> flexible for an agent to interleave with clarifying questions, and harder to resume
> mid-flow if the admin changes their mind partway through. Revisit if granular tools prove
> unwieldy in practice.

## 3. Authentication Ordering (Chicken-and-Egg)

MCP calls authenticate via `api_credentials` keys (TECHNICAL_DESIGN §7) — but the very first
bootstrap step (creating the root tenant itself, USER_JOURNEYS_ADMIN_TENANT_MANAGEMENT.md
§1) happens before any tenant, user, or API credential exists at all. So:

1. The CLI seed command (§1, unchanged) creates the root tenant + internal app + root
   superadmin — still CLI-only, not MCP.
2. The root superadmin logs into the UI once with those seeded credentials and generates an
   API credential from the admin surface (TECHNICAL_DESIGN §7).
3. **Only from that point on** can the conversational, MCP-driven setup flow below be used
   — for creating brand tenants, apps, and roles.

> **Open — nice-to-have, not required.** The CLI seed command could optionally also output
> a starter API credential directly, skipping step 2's manual UI trip. Not decided; the
> three-step ordering above works without it.

## 4. Tools

Each tool is subject to the same two gates as any MCP call (PRD §8): it must be part of
this curated subset (all of the below are, by design, since this pass exists specifically
to expose them), and the caller's held permissions must actually grant the action
(AUTHORIZATION_MODEL §2–§3) — an API credential with insufficient permissions gets the same
403 a REST caller would.

### `tenant_create`
Maps to USER_JOURNEYS_ADMIN_TENANT_MANAGEMENT.md §2.
- **Input**: `name`, `parent_tenant_id` (defaults to the caller's own tenant).
- **Output**: new `tenant_id`.

### `tenant_register_domain`
Maps to §2 (domain registry step).
- **Input**: `tenant_id`, `domain`.
- **Output**: confirmation.

### `tenant_configure`
Maps to §3. Partial-update — an agent can call this multiple times as the conversation
gathers more answers, rather than needing everything up front.
- **Input**: `tenant_id`, and any of: `logo_url`, `brand_image_url`, `login_layout`
  (`centered`/`split`), `enabled_login_methods` (array), `mfa_required` (boolean),
  `audit_retention_days`.
- **Output**: confirmation + current effective settings.

### `tenant_set_provider_credential`
Maps to §3 (provider credentials sub-step) — separate from `tenant_configure` since it's
repeatable (one call per provider) and handles sensitive values.
- **Input**: `tenant_id`, `provider_type` (`google`/`apple`/`otp_email`/`otp_sms`),
  `config` (provider-specific credentials).
- **Output**: confirmation (never echoes the credential back).

### `app_register`
Maps to §4. Takes everything needed to fully stand up a new app in one call — signing key
and token TTLs can be auto-provisioned with sane defaults (TECHNICAL_DESIGN §4) if not
specified, so a minimal call only needs `name` and `supports_organizations`.
- **Input**: `tenant_id`, `name`, `redirect_uris`, `logo_url` (optional),
  `supports_organizations` (boolean), `signing_algorithm` (optional),
  `access_token_ttl_seconds`/`id_token_ttl_seconds`/`refresh_token_ttl_seconds` (optional).
- **Output**: `app_id`, `client_id`, `client_secret` (shown once).

### `role_create`
Maps to §5. Since permissions/policies are embedded arrays, not separate objects
(DATA_MODEL.md `roles`), this one call is the entire "define a role" step — no separate
"create permission" tool exists.
- **Input**: `app_id`, `org_id` (optional, for an org-specific customization),
  `name`, `permissions` (array of strings), `policies` (array of objects, optional),
  `set_as_default_signup` (boolean, optional — folds the §5 "designate default" step into
  role creation itself for convenience during initial setup).
- **Output**: `role_id`.

## 5. Not Yet Covered

Deliberately out of this first pass, consistent with
USER_JOURNEYS_ADMIN_TENANT_MANAGEMENT.md §7:
- Changing an app's default signup role *after* initial setup (only creation-time is
  covered, via `role_create`'s flag).
- Editing/rotating an app's signing key after creation.
- Deleting/deactivating a tenant, app, or role.
- Any tool outside the setup flow (metrics, audit log queries, member/invite management,
  org role customization) — separate future passes.

## 6. Open Items

1. **CLI seed → starter API credential** (§3) — nice-to-have shortcut, not decided.
2. **Composite vs. granular tools** (§2) — granular chosen for this pass; revisit if it
   proves unwieldy for an agent to orchestrate in practice.
