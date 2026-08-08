# ui

The default theme SPA — the one first-party UI Porichoy ships. Covers login, signup,
self-service account management, the organization switcher, and permission-gated admin, all
in one unified surface (no separate admin dashboard). See
[`wiki/TECHNICAL_DESIGN.md`](../wiki/TECHNICAL_DESIGN.md) §1–§2 for the full rationale, and
[`wiki/user-journeys/`](../wiki/user-journeys) for the flows it needs to implement.

Fully swappable — self-hosters can replace this with their own custom theme built against
the REST API. Not against `sdk/typescript`, which targets third-party *apps* integrating
"Login with Porichoy," not this UI or a custom theme replacing it — instead, `src/lib/client`
(this app's own thin API client + headless auth-flow hooks, no styled components) is the
reference a custom theme is expected to read/copy from. See
[`wiki/UI_CODING_STANDARDS.md`](../wiki/UI_CODING_STANDARDS.md) §3.

Stack: React + Vite + TypeScript, Tailwind + shadcn/ui, TanStack Query, React Router, React
Hook Form + Zod. Full rationale and conventions in
[`wiki/UI_CODING_STANDARDS.md`](../wiki/UI_CODING_STANDARDS.md).

Not yet implemented — this document (and UI_CODING_STANDARDS.md) precede code, same as
CODING_STANDARDS.md preceded `server/`.
