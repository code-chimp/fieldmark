# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Repository Is

FieldMark is a construction compliance and inspection management system implemented **across three parallel stacks** — .NET (Razor Pages + HTMX), Django (Templates + HTMX), and Go (Fiber + HTMX) — against a shared PostgreSQL 17 database. It is a teaching artifact demonstrating server-authoritative architecture as an alternative to SPAs. A story is never done until all three stacks pass it.

See [docs/getting-started.md](docs/getting-started.md) for infrastructure, quick start, and setup. See [docs/architecture.md](docs/architecture.md) for full architecture, patterns, and domain details. CSS pipeline (two-step build, pre-build checks, pnpm guard, Basecoat upgrade procedure): [fieldmark_shared/CLAUDE.md](fieldmark_shared/CLAUDE.md) and [docs/getting-started.md#css-pipeline](docs/getting-started.md).

## Hard Rules (all stacks)

See [docs/hard-rules.md](docs/hard-rules.md). Stack-specific rules live in each project's own `CLAUDE.md`.

## Key Reference Documents

- [docs/architecture.md](docs/architecture.md) and [docs/getting-started.md](docs/getting-started.md) — canonical docs (progressive disclosure from this file).
- `_bmad-output/planning-artifacts/architecture.md` — architectural source of truth (decisions, patterns, structure, validation)
- `_bmad-output/planning-artifacts/prd/` — product requirements source of truth (sharded; index at `prd/index.md`)

The `_bmad-output/planning-artifacts/research/` folder contains pre-kickoff priming material. It is not maintained and is not authoritative.
