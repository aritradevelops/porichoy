# server

The Porichoy backend. Go, monolith, hexagonal (ports & adapters) architecture — see
[`wiki/TECHNICAL_DESIGN.md`](../wiki/TECHNICAL_DESIGN.md) §1–§2 for the full rationale.

```
cmd/server/          entrypoint (main package)
internal/
  domain/            business logic, grouped by bounded context (wiki/CODING_STANDARDS.md §4):
                       tenant/          Tenant, DomainRegistry, TenantProviderCredential
                       app/             App, Session
                       identity/        User, Password, MFAMethod, ExternalIdentity,
                                        VerificationToken
                       organization/    Organization, OrgMembership
                       authorization/   Role, RoleAssignment, APICredential
                       audit/           AuditLog
                     Each package defines its own entities AND the port interfaces it needs
                     (e.g. identity/ defines User + UserRepository) — no separate ports/
                     package. Cross-package references are always by ID (uuid.UUID), never
                     an embedded struct pointer.
  application/       cross-context use cases (wiki/CODING_STANDARDS.md §3), same 6-package
                     grouping as domain/ above. What REST and MCP handlers actually call —
                     e.g. application/organization.Service.CreateOrganization composes
                     organization.Repository + authorization.Repository to create the org
                     and assign its creator the Owner role, in one place, once.
  adapters/
    rest/             /api/{version}/{module}/{action} handlers + router (Fiber)
    mcp/              MCP server (wiki/MCP_TOOLS.md)
    postgres/         repository implementations (Bun)
    cache/redis/      default cache provider
    secrets/          env/file, vault, kms
    logsink/          postgres + external sink
    email/, sms/      OTP delivery providers
    crypto/           encryption-at-rest adapter
migrations/           SQL, goose
config/               default YAML + schema
deploy/               docker-compose.yml, etc.
```

Schema: [`wiki/DATA_MODEL.md`](../wiki/DATA_MODEL.md). Journeys implemented here:
[`wiki/user-journeys/`](../wiki/user-journeys). Testing/commenting/i18n conventions:
[`wiki/CODING_STANDARDS.md`](../wiki/CODING_STANDARDS.md).

## Running

```
go run ./cmd/server
```

Not yet functional beyond the stub entrypoint — nothing under `internal/` is implemented.
