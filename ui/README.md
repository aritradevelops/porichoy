# ui

The default theme SPA — the one first-party UI Porichoy ships. Covers login, signup,
self-service account management, the organization switcher, and permission-gated admin, all
in one unified surface (no separate admin dashboard). See
[`wiki/TECHNICAL_DESIGN.md`](../wiki/TECHNICAL_DESIGN.md) §1–§2 for the full rationale, and
[`wiki/user-journeys/`](../wiki/user-journeys) for the flows it needs to implement.

Fully swappable — self-hosters can replace this with their own custom theme built against
the REST API / `sdk/typescript`, per the same docs.

Not yet implemented — frontend framework choice hasn't been made yet.
