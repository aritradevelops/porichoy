# server

The Porichoy backend. Go, monolith, hexagonal (ports & adapters) architecture — see
[`wiki/TECHNICAL_DESIGN.md`](../wiki/TECHNICAL_DESIGN.md) §1–§2 for the full rationale.

```
cmd/server/          entrypoint (main package)
internal/
  domain/            business logic — tenant, app, user, org, role, permission, session,
                     audit... the same "module" vocabulary used by the API route
                     convention and the {module}:{action}@{scope} permission format
                     (wiki/AUTHORIZATION_MODEL.md)
  ports/              interfaces: repositories, CachePort, SecretsPort, EncryptionPort,
                     LogSinkPort, TokenSigner, EmailPort, SMSPort
  adapters/
    rest/             /api/{version}/{module}/{action} handlers + router
    mcp/              MCP server (wiki/MCP_TOOLS.md)
    postgres/         repository implementations
    cache/redis/      default cache provider
    secrets/          env/file, vault, kms
    logsink/          postgres + external sink
    email/, sms/      OTP delivery providers
    crypto/           encryption-at-rest adapter
migrations/           SQL, golang-migrate/goose
config/               default YAML + schema
deploy/               docker-compose.yml, etc.
```

Schema: [`wiki/DATA_MODEL.md`](../wiki/DATA_MODEL.md). Journeys implemented here:
[`wiki/user-journeys/`](../wiki/user-journeys).

## Running

```
go run ./cmd/server
```

Not yet functional beyond the stub entrypoint — nothing under `internal/` is implemented.
