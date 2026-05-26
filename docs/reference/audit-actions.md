# Audit Actions — Canonical Contract

> **Status:** skeleton (scaffolded by Epic 1 retrospective action item A4, 2026-05-25).
> Content is populated by **Story 2.2** (`domain.audit_entry` + `append_audit_entry()` helper).

This document is the **single source of truth** for the canonical set of audit-action strings emitted by `domain.audit_entry.action`. Each stack (.NET, Django, Go) implements a native enum/constants module whose values must match this list exactly. Per-stack conformance tests assert that alignment.

See the root [CLAUDE.md](../../CLAUDE.md) **Cross-Stack Architecture Principle** for why this lives as documentation rather than as shared code.

---

## Canonical Action List

TODO (Story 2.2): populate the canonical list of audit-action strings. Initial scope is the 14 strings ratified at Epic 1 close, plus `ProjectCreated` added by ADR amendment during Epic 2 planning (15 total). Each entry should record: action string, when emitted, the entity type it applies to, and a brief description.

| Action | Entity | Emitted when | Notes |
|---|---|---|---|
| `ProjectCreated` | Project | TODO | TODO |
| TODO | TODO | TODO | TODO |

---

## Casing Convention

TODO (Story 2.2): document the SCREAMING/PascalCase rule and how it maps across stacks (C# `enum` PascalCase, Python module-level constants, Go `const` block).

---

## Per-Stack Native Implementations

Each stack owns its own enum/constants module. The top-of-file comment must reference this document.

- **.NET** — TODO: path under `FieldMark.Domain/`.
- **Django** — TODO: path under `fieldmark_py/audit/`.
- **Go** — TODO: path under `fieldmark-go/internal/domain/`.

---

## Conformance Test Contract

Each stack ships a conformance test that:

1. Reads the canonical list (parsed from this document, or from a checked-in fixture derived from it).
2. Asserts the stack's native enum/constants set matches exactly — no extras, no missing entries.

TODO (Story 2.2): define the fixture format (if any) and the per-stack test location.

---

## Change Procedure

TODO (Story 2.2): document the ADR-amendment procedure for adding/removing audit actions (Epic 2 planning amendment for `ProjectCreated` is the precedent).
