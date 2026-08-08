# Porichoy

An open source, self-hosted Identity Provider (IdP) — multi-tenant, OAuth2/OIDC, and
MCP-first: every management capability is a first-class citizen of the UI, the REST API,
*and* an MCP server, not an afterthought bolted onto one of them.

Full documentation lives in [`wiki/`](./wiki), starting with:

- [`wiki/PRD.md`](./wiki/PRD.md) — product requirements
- [`wiki/TECHNICAL_DESIGN.md`](./wiki/TECHNICAL_DESIGN.md) — architecture and implementation decisions
- [`wiki/DATA_MODEL.md`](./wiki/DATA_MODEL.md) — database schema
- [`wiki/AUTHORIZATION_MODEL.md`](./wiki/AUTHORIZATION_MODEL.md) — permission/scope model
- [`wiki/MCP_TOOLS.md`](./wiki/MCP_TOOLS.md) — MCP tool surface
- [`wiki/user-journeys/`](./wiki/user-journeys) — end-user and admin journeys

## Layout

This is a monorepo:

- [`server/`](./server) — the Go backend (hexagonal architecture: `internal/domain`,
  `internal/ports`, `internal/adapters/*`).
- [`sdk/`](./sdk) — official client SDKs (Go, Python, TypeScript).
- [`ui/`](./ui) — the default theme SPA (login, signup, self-service account management,
  org switcher, and permission-gated admin — one unified UI, no separate admin dashboard).
- [`wiki/`](./wiki) — all project documentation.

## Status

Early stage — documentation and architecture are drafted (see `wiki/`), implementation is
just getting started.

## License

MIT — see [LICENSE](./LICENSE).
