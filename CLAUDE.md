# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Repository Is

FieldMark is a construction compliance and inspection management system implemented **across three parallel stacks** — .NET (Razor Pages + HTMX), Django (Templates + HTMX), and Go (Fiber + HTMX) — against a shared PostgreSQL 17 database. It is a teaching artifact demonstrating server-authoritative architecture as an alternative to SPAs. A story is never done until all three stacks pass it.

See [docs/tutorials/getting-started.md](docs/tutorials/getting-started.md) for infrastructure, quick start, and setup. See [docs/explanation/architecture.md](docs/explanation/architecture.md) for full architecture, patterns, and domain details. CSS pipeline (two-step build, pre-build checks, pnpm guard, Basecoat upgrade procedure): [fieldmark_shared/CLAUDE.md](fieldmark_shared/CLAUDE.md) and [docs/tutorials/getting-started.md#css-pipeline](docs/tutorials/getting-started.md).

The `docs/` folder is organized by the Diátaxis framework: `tutorials/` (learning-oriented), `how-to/` (problem-oriented), `reference/` (information-oriented), `explanation/` (understanding-oriented). Pick the quadrant that matches the reader's need.

## Hard Rules (all stacks)

See [docs/reference/hard-rules.md](docs/reference/hard-rules.md). Stack-specific rules live in each project's own `CLAUDE.md`.

## Key Reference Documents

- [docs/explanation/architecture.md](docs/explanation/architecture.md) and [docs/tutorials/getting-started.md](docs/tutorials/getting-started.md) — canonical docs (progressive disclosure from this file).
- `_bmad-output/planning-artifacts/architecture.md` — architectural source of truth (decisions, patterns, structure, validation)
- `_bmad-output/planning-artifacts/prd/` — product requirements source of truth (sharded; index at `prd/index.md`)

The `_bmad-output/planning-artifacts/research/` folder contains pre-kickoff priming material. It is not maintained and is not authoritative.
