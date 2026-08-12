# Coding Standards

Engineering conventions for the `server/` codebase — how code actually gets written, as
opposed to [TECHNICAL_DESIGN.md](./TECHNICAL_DESIGN.md) (what gets built) and
[DATA_MODEL.md](./DATA_MODEL.md) (the schema). Applies to `server/` specifically; SDKs and
the UI will get their own conventions once those are underway.

## 1. Libraries (see TECHNICAL_DESIGN.md §2 for the full stack table)

| Concern | Choice |
|---|---|
| HTTP framework | [Fiber](https://gofiber.io/) |
| ORM / SQL layer | [Bun](https://bun.uptrace.dev/), via `bun/driver/pgdriver` (Bun's own pure-Go Postgres driver, no cgo) |
| Migrations | [goose](https://github.com/pressly/goose) |
| JWT | [lestrrat-go/jwx](https://github.com/lestrrat-go/jwx) — chosen over golang-jwt for its fuller JOSE/JWK support, relevant given per-app JWKS is planned (TECHNICAL_DESIGN §4) |
| Structured logging | [zerolog](https://github.com/rs/zerolog) |
| Config loading | [koanf](https://github.com/knadh/koanf) (YAML + env var overrides) |
| Request validation | [go-playground/validator](https://github.com/go-playground/validator) |
| UUID generation | [google/uuid](https://github.com/google/uuid), application-generated (not DB-generated) |
| Test assertions/mocking | [testify](https://github.com/stretchr/testify) (`assert`/`require` + `mock`) |
| Integration test infra | [testcontainers-go](https://golang.testcontainers.org/) |
| i18n | exact package not pinned yet (`go-i18n` is the likely default) — see §8 |

> **Interpreting "burn" as "Bun"** — flagging this rather than silently assuming, since the
> library name was given verbally and Bun is the closest real Go ORM name to what was said.
> Correct this if a different library was actually meant.

## 2. Architecture Patterns

- **No domain/application split — one package per bounded context, holding everything.**
  Entities, repository interfaces (ports), *and* the `Service` implementing that context's
  use cases all live together — e.g. `internal/tenant` defines `Tenant`, `Repository`, and
  `Service`. This was tried as two parallel package trees (`internal/domain/tenant` +
  `internal/application/tenant`) and reverted — the same-name-different-directory split
  forced an import alias (`domaintenant "…/internal/domain/tenant"`) in every file that
  needed both halves, for no real benefit. See §3 for what replaced it, and §10 for the
  scaffold history.
- **Bun models are separate from entities.** These packages have zero knowledge of Bun or
  Postgres. `internal/adapters/postgres` defines its own Bun-tagged model structs per entity
  and maps to/from the domain entity in the repository implementation. More mapping code per
  repository, but keeps entities fully infra-agnostic.
- **Dependency injection is manual.** Explicit constructor calls wired up in `main.go` (or a
  small bootstrap package under `cmd/server`) — no DI framework/codegen.
- **Domain errors carry their i18n key via a custom type.** `apperror.Error`
  (`internal/apperror`), carrying an i18n key (§8) — e.g.
  `apperror.New("tenant.domain_already_registered")`. HTTP handlers `errors.As` it out to
  build the response envelope (§5). Business logic never touches HTTP or i18n message text
  directly, just the key.
- **`internal/actor` is a small, dependency-free exception to "no domain/application
  split."** It isn't a bounded context — it holds one thing, the `Actor` struct (who's
  calling, their tenant/app/org, and their resolved authorization scope,
  AUTHORIZATION_MODEL.md §2), and depends on nothing but `uuid`. It deliberately doesn't
  live inside `internal/authorization` (§4's table) — that package is its own bounded
  context (Role/RoleAssignment/APICredential) and shouldn't become a mandatory kernel
  dependency for every other package just because it happens to also own scope-ranking
  logic. Whichever bounded-context packages have authorized-only operations import
  `internal/actor` for this one parameter type; see §4 for exactly when a method takes one.

## 3. Package Structure

Each bounded context is **one package**, holding its entities, the repository interfaces
(ports) it needs, and a `Service` exposing that context's use cases — no separate layer for
any of these.

- **`Service` is what REST and MCP handlers both call into.** The actual use cases — e.g.
  "create an organization and auto-assign its creator the Owner role"
  (USER_JOURNEYS_ORGANIZATIONS.md §1) — live here exactly once. Neither adapter
  re-implements this logic; both are thin callers of the same `Service` methods, so behavior
  can't drift between "created via the UI" and "created via an MCP tool call"
  (MCP_TOOLS.md).
- **A context's `Service` can depend on sibling contexts' repository interfaces directly** —
  e.g. `organization.Service` holds both `organization.Repository` and
  `authorization.Repository` to implement org creation + Owner-role assignment as one use
  case. This is exactly the cross-context dependency the old application-layer split existed
  to isolate; it turns out to just be an ordinary Go import (`organization` importing
  `authorization`), no separate layer needed to contain it, and no naming collision either —
  `authorization.Repository` reads cleanly from within `organization`'s own code, unlike the
  old `domaintenant.Repository` situation.
- **Services don't depend on sibling Services**, only on repository ports (their own
  context's and, per above, other contexts' when a use case genuinely spans both). Keeps the
  dependency graph a simple star rather than a web.
- **Depending on `internal/actor` isn't the same kind of cross-context dependency as the
  point above.** That's about one bounded context calling into another's business logic
  through a repository interface. `actor.Actor` carries no behavior, just data — every
  package that accepts one as a parameter doesn't turn the star into a web, it's just
  consuming a shared, inert value type (§2).
- **Handlers stay thin.** `internal/adapters/rest` and `internal/adapters/mcp` parse/validate
  the incoming request, call one `Service` method, and serialize the result — no business
  logic in the handler itself.
- **Transactional boundaries are owned by `Service` methods** — one needing multiple
  repository calls to succeed or fail together (e.g. the org row *and* the role assignment)
  coordinates that, even though the exact Bun transaction mechanics are an adapter-level
  detail decided when that code gets written.
- The 6 bounded-context packages are listed in §4's table.

## 4. Domain Modeling Conventions

- **Nullable fields use pointers.** `*string`, `*time.Time`, etc. — not `sql.Null*` types
  (those are a database/sql-layer concern, and domain entities have zero knowledge of the
  database, per §2) and not a generic `Optional[T]` wrapper. A nil pointer *is* the "no
  value" representation.
- **Enums are typed string constants**, not plain `string` fields — e.g.
  ```go
  type LoginLayout string
  const (
      LoginLayoutCentered LoginLayout = "centered"
      LoginLayoutSplit    LoginLayout = "split"
  )
  ```
  Applies to every `enum(...)` column in DATA_MODEL.md (`login_layout`, `provider_type`,
  the various `status` fields, `mfa_method.type`, etc.) — compile-time safety over a bare
  string, even though nothing enforces this at the database layer (DATA_MODEL.md §0 already
  established the DB itself doesn't constrain these).
- **`role.policies` is `json.RawMessage`**, not `map[string]any` or a custom struct.
  Porichoy never interprets policy content (PRD §7.1, DATA_MODEL.md `roles`) — keeping it as
  opaque raw bytes enforces that at the type level; there's no accidental temptation to
  introspect fields that aren't ours to interpret.
- **Packages are grouped by bounded context, not one-per-table.** DATA_MODEL.md's own
  §1–§6 grouping maps directly onto this — 6 packages instead of 18, each living directly
  under `internal/` (e.g. `internal/tenant`, sitting alongside `internal/adapters`):

  | Package | DATA_MODEL.md entities |
  |---|---|
  | `internal/tenant` | Tenant, DomainRegistry, TenantProviderCredential |
  | `internal/app` | App, Session |
  | `internal/identity` | User, Password, MFAMethod, ExternalIdentity, VerificationToken |
  | `internal/organization` | Organization, OrgMembership |
  | `internal/authorization` | Role, RoleAssignment, APICredential |
  | `internal/audit` | AuditLog |

  `sessions` sits under `app` (matching DATA_MODEL.md §2) even though it's arguably an
  identity concern — following the doc's existing grouping for consistency rather than
  re-deriving a boundary. `api_credentials` sits under `authorization`, consistent with it
  being a principal that holds role assignments the same way a user does (DATA_MODEL.md §0).
  Flagging both as judgment calls, not asked-for specifics.
- **Cross-entity references are always by ID, never by embedded struct pointer.** A `Role`
  holds `AppID uuid.UUID`, not `App *app.App`. This is what makes the grouping above work
  without import cycles — e.g. `role.AppID` and `app.DefaultSignupRoleID` reference each
  other's tables in DATA_MODEL.md, but since both are plain `uuid.UUID` fields, neither the
  `authorization` package nor the `app` package needs to import the other's *entity* type.
  Services do reach across contexts (§3) — but through a repository *interface*, still never
  an embedded struct.
- **A `Repository` method takes `act actor.Actor` only when the operation is authorized-only
  *and* doesn't already receive a full entity carrying `CreatedBy`/`UpdatedBy`.** Not
  blanket-added to every method:
  - `Create`/`Update`/`Upsert`-style methods take the entity itself, which the `Service`
    already stamped with the caller's id — adding `Actor` too would be redundant.
  - Methods with no entity to carry that (`SoftDelete`, list/get-style lookups that are only
    ever reached via an authorized route) take `act actor.Actor` — this is what a Postgres
    adapter implementation uses to build the scope-based query filter
    (AUTHORIZATION_MODEL.md §4).
  - Public/pre-authentication lookups take **no `Actor` at all**, ever — not even a nil or
    zero-value one. `internal/tenant`'s `Repository.FindByID` and `DomainRepository.FindByDomain`
    are the concrete examples: both back `Service.ResolveTenantByDomain`
    (TECHNICAL_DESIGN.md §3.3), which runs *before* authentication, so no `Actor` can exist
    yet. Where the same underlying data also needs an authorized, scope-filtered lookup
    (e.g. an admin fetching a tenant by id), that's a separate method (`GetByID`) — not the
    same method with an optional/fake actor.
  - System-generated actions with no caller at all (the CLI seed command bootstrapping the
    root tenant, USER_JOURNEYS_ADMIN_TENANT_MANAGEMENT.md §1) get their own dedicated method
    too (`Service.CreateRootTenant`, distinct from the authorized `Service.CreateTenant`) —
    consistent with `actor.Actor` never being nil/optional (§2).

## 5. API Layer

- **Routes are registered explicitly**, one per module/action, even though they follow the
  uniform `/api/{version}/{module}/{action}/{id?}` pattern (AUTHORIZATION_MODEL §1) — no
  dynamic dispatch/registry-lookup indirection. Easier to grep a URL straight to its handler.
- **Middleware is chained, one concern per middleware**: tenant resolution → authentication
  → authorization (AUTHORIZATION_MODEL §2), as three separate, independently testable Fiber
  middleware functions in that order — not folded into fewer, larger ones. The authorization
  middleware is what builds an `actor.Actor` (`internal/actor`, §2) from the resolved
  tenant/authenticated principal/permission lookup, and hands it to the handler via Fiber
  `Locals` — the *only* place an `Actor` crosses from Fiber-land into domain code. The
  handler extracts it immediately and passes it on as an explicit Go parameter from there;
  domain and adapter code never touches Fiber's request object directly, consistent with
  entities having zero knowledge of the delivery mechanism (§2). Routes that don't run this
  full chain (public/pre-authentication ones, e.g. tenant resolution itself) call `Service`
  methods that simply don't take an `Actor` parameter (§4) — there's nothing to hand off.
- **Response envelope**, for every success and error response alike:
  ```json
  { "data": ..., "error": { "key": "...", "message": "..." } }
  ```
  `data` is populated on success (`error` is `null`); `error` is populated on failure
  (`data` is `null`). `error.key` is the i18n key from the `DomainError` (§2); `error.message`
  is that key resolved to text for the request's locale (§8).
- Handlers here (and the MCP equivalents in `internal/adapters/mcp`) are thin callers of
  each bounded context's `Service` (§3) — routing/transport concerns only, no business
  logic.

## 6. Testing

Both unit and integration tests are required — this isn't optional or "nice to have."

- **Unit tests**: standard Go convention — `_test.go` files alongside the code they test.
  Cover both entity logic and `Service` methods within each bounded-context package,
  mocking repository ports (including sibling contexts' where a `Service` depends on one,
  §3) with `testify/mock` fakes.
- **Integration tests**: exercise real Postgres/Redis via `testcontainers-go` — hermetic,
  per-test-run containers, no dependency on `server/deploy/docker-compose.yml` already being
  up. This matters especially for anything touching the scope-based query filtering
  (AUTHORIZATION_MODEL §4) and the soft-delete/principal patterns (DATA_MODEL §0), where a
  mocked repository could hide a real query bug.
- **Convention**: gate integration tests behind a build tag (`//go:build integration`) so
  `go test ./...` stays fast and infra-free by default; CI runs a separate
  `go test -tags=integration ./...` pass.

## 7. Commenting

- **Comments capture intention, not mechanics.** Explain *why* something is done a
  particular way — a non-obvious constraint, a workaround, a decision that would otherwise
  look arbitrary — not what the code already says by being read. A comment that just
  restates the line below it should be deleted, not written.
- **Keep it minimal.** No multi-paragraph comment blocks. If a comment needs several
  sentences to land, that's usually a sign the code itself should be clearer (better naming,
  smaller function), not that it needs more prose next to it.
- **Simple language.** Plain, direct sentences — this is documentation for the next person
  reading the code, not a design essay.
- Exported identifiers still get a standard Go doc comment (the idiomatic
  `// FuncName does X` form, one line where possible) since that's what godoc/IDE tooling
  surfaces — this isn't in tension with "minimal," a one-liner satisfies both.

## 8. Error Messages (i18n)

Error messages are never hardcoded strings in application code. Each error carries a
**translation key** (via the `DomainError` type, §2), and the human-readable message is
resolved from a translation file at response time (based on requested locale — e.g.
`Accept-Language`), not baked into the Go source.

- Application code returns/raises a `DomainError` identified by its key (e.g.
  `auth.invalid_credentials`, `tenant.domain_already_registered`) — never the literal
  user-facing string.
- Translation files map each key to locale-specific message text. English is presumably the
  only shipped locale initially, but the mechanism doesn't hardcode that assumption.
- This also gives every error a stable, greppable identifier independent of its wording —
  useful for the client SDKs (TECHNICAL_DESIGN §7.2) to handle specific error cases
  programmatically rather than string-matching messages.

> **Open — exact locale-resolution mechanics** (which header, fallback locale, where
> translation files live in the repo) not decided yet.

## 9. Postgres Adapter

- **One repository struct per interface, in `internal/adapters/postgres`** — e.g.
  `TenantRepository` implements `tenant.Repository`. Each file also declares a compile-time
  check (`var _ tenant.Repository = (*TenantRepository)(nil)`) so an interface drift is a
  build failure, not a runtime surprise.
- **Explicit soft delete, not Bun's built-in soft-delete feature.** Bun can auto-manage a
  `deleted_at`-tagged field, but that only covers one column — this schema always pairs
  `deleted_at` with `deleted_by` (DATA_MODEL.md §0), so `SoftDelete` methods hand-write the
  `UPDATE ... SET deleted_at = ?, deleted_by = ?` themselves, and every read query hand-adds
  `WHERE deleted_at IS NULL`. Consistent with this codebase's general preference for explicit
  code over framework magic (§2 — manual DI, no codegen).
- **Migrations are goose SQL files, timestamp-named** (`goose create`'s default format,
  `YYYYMMDDHHMMSS_description.sql`), one file per bounded-context's initial table batch —
  e.g. `tenants`, `domain_registries`, and `tenant_provider_credentials`
  (`internal/tenant`'s three tables) ship together in one migration, not three.
- **Table names are always plural**, Go entity/package names always singular (e.g. `tenants`
  the table, but `Tenant`/`internal/tenant` the Go type and package) — the standard
  ORM-adjacent convention (a row is one `Tenant`; the table holding many of them is
  `tenants`).
- **Integration tests share one `TestMain`-managed container per package**, not one per test
  function — still "hermetic, per-test-run" per §6, just scoped to the test binary rather
  than each individual test case, since spinning a fresh Postgres per test is unnecessary
  overhead here.

## 10. Build Order

Initial implementation order: **entities and repository interfaces first, then each
context's `Service`, before wiring up any HTTP handler, Postgres adapter, or auth flow** —
per the grouped packages in §4. Each bounded-context package is established as a stable,
directly testable unit before anything adapter-level is built on top of it, even though
nothing is runnable end-to-end until the adapters follow.

## 11. Scaffold History

`server/internal/ports/` (created early on, holding nothing but empty per-concern
directories) was removed once §2 was decided — port interfaces live with their entities,
not centralized.

A separate `server/internal/domain/` + `server/internal/application/` split (one package
tree for entities+ports, a parallel tree for `Service`s) was then built, then reverted in
favor of the single-package-per-context structure in §2/§3 — the parallel trees forced an
import alias in every file needing both an entity and its `Service` (`internal/domain/tenant`
and `internal/application/tenant` are both package `tenant`, imported into the same file).
The two trees were flattened into 6 packages living directly under `internal/`
(`internal/tenant`, `internal/app`, `internal/identity`, `internal/organization`,
`internal/authorization`, `internal/audit`), matching §4's table.
