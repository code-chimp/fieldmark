# FieldMark Documentation Index

Welcome to the FieldMark project documentation. This folder contains the canonical, published documentation for developers, contributors, and operators.

## Documentation Structure

- [Overview](overview.md) — Friendly introduction for new developers and post-presentation readers.
- [Getting Started](getting-started.md) — Quick start, prerequisites, running the application, and development setup.
- [Architecture](architecture.md) — System overview, three-stack design, canonical patterns, domain model, request flow, and invariants.
- [Domain Model](domain-model.md) — Aggregates, state machines, relationships (with Mermaid diagrams).
- [Request Flow](request-flow.md) — Canonical mutating request lifecycle (with sequence diagram).
- [Hard Rules](hard-rules.md) — Non-negotiable constraints across all stacks (backend authority, schema ownership, no fat services, etc.).

## Related References

- Root [CLAUDE.md](../CLAUDE.md) — AI/agent guidance with links to these docs (progressive disclosure).
- Stack-specific CLAUDE.md files in `FieldMark/`, `fieldmark_py/`, and `fieldmark-go/`.
- `_bmad-output/planning-artifacts/` — Historical planning artifacts (not maintained post-kickoff).

Use the docs here for day-to-day reference. They are designed for progressive disclosure: high-level summaries stay in CLAUDE.md; deep details live here.
