# server

The Porichoy backend. Go, monolith, hexagonal (ports & adapters) architecture — see
[`wiki/TECHNICAL_DESIGN.md`](../wiki/TECHNICAL_DESIGN.md) §1–§2 for the full rationale.

```
cmd/server/          entrypoint (main package)
internal/
  tenant/            Tenant, DomainRegistry, TenantProviderCredential
  app/               App, Session
  identity/          User, Password, MFAMethod, ExternalIdentity, VerificationToken
  organization/      Organization, OrgMembership
  authorization/     Role, RoleAssignment, APICredential
  audit/             AuditLog
                     Each of the 6 packages above (wiki/CODING_STANDARDS.md §2–§4) holds
                     its entities, the repository interfaces (ports) it needs, AND a
                     Service exposing that context's use cases — e.g. organization.Service
                     .CreateOrganization composes organization.Repository +
                     authorization.Repository to create the org and assign its creator the
                     Owner role, in one place, once. This is what REST and MCP handlers
                     both call into. Cross-package references are always by ID (uuid.UUID)
                     or through a repository interface — never an embedded struct pointer.
  apperror/          the i18n-keyed error type (wiki/CODING_STANDARDS.md §2, §8) business
                     logic raises instead of a hardcoded message.
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
