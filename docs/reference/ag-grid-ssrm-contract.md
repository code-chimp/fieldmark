# AG Grid Server-Side Row Model — Wire Format Contract

> **Status:** skeleton (scaffolded by Epic 1 retrospective action item A4, 2026-05-25).
> Content is populated by **Story 2.9** (Project list AG Grid with server-side row model).

This document is the **single source of truth** for the AG Grid SSRM wire format used by every grid endpoint in FieldMark (`POST /grid/projects`, `POST /grid/inspections`, `POST /grid/violations`, etc.). Each stack implements the contract natively against its framework; per-stack conformance tests assert alignment.

See the root [CLAUDE.md](../../CLAUDE.md) **Cross-Stack Architecture Principle** for why this lives as documentation rather than as a shared codec.

---

## Request Shape

TODO (Story 2.9): document the full SSRM request payload.

```jsonc
{
  // TODO: filterModel — allowed operators per column type
  "filterModel": {},
  // TODO: sortModel — allowed sort directions, multi-column rules
  "sortModel": [],
  // TODO: pagination — startRow / endRow semantics, max page size
  "startRow": 0,
  "endRow": 100
}
```

- **Key casing:** snake_case throughout.
- **Allowed filter operators:** TODO (per column type).
- **Sort direction values:** TODO.
- **Pagination bounds:** TODO (max `endRow - startRow`).

---

## Response Shape

TODO (Story 2.9): document the response envelope.

```jsonc
{
  "rows": [
    // TODO: row projection rules — manual projection per NFR6 (no AutoMapper),
    // snake_case keys, primitive types only (no nested domain objects unless explicitly contracted).
  ],
  "lastRow": 0  // TODO: semantics — total row count when known, -1 when unknown, etc.
}
```

---

## Row Projection Rules

TODO (Story 2.9): codify the rule that rows are manually projected per NFR6 (no AutoMapper, no generic mappers). Document the canonical projected columns for `POST /grid/projects` as the first worked example.

---

## Error Behaviour

TODO (Story 2.9): document the response when the request is malformed (invalid filter operator, out-of-bounds pagination, unknown sort column).

---

## Per-Stack Native Implementations

- **.NET** — TODO: handler location, EF Core projection pattern.
- **Django** — TODO: view location, ORM/raw SQL projection pattern.
- **Go** — TODO: handler location, `pgx` projection pattern.

No shared codec, no generated stubs.

---

## Conformance Test Contract

Each stack ships a conformance test that:

1. Issues a canonical SSRM request fixture (`tests/fixtures/ssrm-canonical-request.json` or derived from this document) against the stack's grid endpoint.
2. Asserts the response shape, key casing, and `lastRow` semantics match the documented contract exactly.

TODO (Story 2.9): define the canonical-request fixture and per-stack test locations.
