# FieldMark documentation

This folder is intended to become the **root of the project’s published documentation** once documentation work proceeds beyond pre-planning. Treat this **`README.md` as permanent**: it describes how to read `docs/` as the library grows and how it differs from other documentation-adjacent folders in the repository.

## What stays vs what is temporary

| Role | Location | Notes |
|------|----------|--------|
| **Permanent** | This file (`docs/README.md`) | Landing page for `docs/`; updated as the doc set matures. |
| **Permanent (future)** | Additional canonical docs added here after kickoff | Product, architecture, operations, and onboarding material maintained as first-class documentation. |
| **Temporary** | Other Markdown files currently in `docs/` | Pre-planning research and drafts; **not** the long-term source of truth. They may be replaced, moved, or superseded by formal docs after kickoff. |

Do not assume naming or content of files beside this README will remain stable until the documentation structure is formally adopted.

## Current contents (inventory)

Alongside this README, `docs/` presently holds working notes and drafts—for example architecture narratives, ADR addenda, and shared-domain discussions. **Unit testing expectations for kickoff are captured in `_bmad-output/planning-artifacts/research/`** (`architecture-decisions.md` boundaries plus each stack `*-reference.md`). Older files here helped exploration; they **do not** replace formal specs elsewhere unless promoted explicitly.

## Authoritative planning references (today)

While this folder holds exploratory material, **planning-facing authority** for agents and implementation alignment lives under:

**`_bmad-output/planning-artifacts/research/`**

Use that tree for domain model, references per stack, architecture decisions, UX guidance, and related planning artifacts until documentation is reorganized post-kickoff.

## Repository context

FieldMark is a **construction compliance and inspection** system implemented as **three parallel stacks** (.NET + HTMX, Django + HTMX, Go/Fiber + HTMX) against a **shared PostgreSQL** database. Stack symmetry (routes, HTMX targets, AG Grid contracts, audit shapes, domain method names) is a **hard constraint**. Root **`CLAUDE.md`** summarizes repository-wide rules; each stack has its own **`CLAUDE.md`** under its project directory.

---

*Draft — revise section headings and links once the canonical documentation set exists.*
