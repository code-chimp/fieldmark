# CLAUDE.md — Go Stack

This file provides guidance to Claude Code (claude.ai/code) when working in the `fieldmark-go/` Go project. Read alongside the root `CLAUDE.md`.

## Commands

Run from the `fieldmark-go/` directory:

```bash
go run ./cmd/web/           # run the application
go build -o fieldmark-go ./cmd/web/   # build the binary
go test ./...               # run all tests
go test ./... -v            # verbose test output
go test ./... -cover        # test with coverage
go vet ./...                # static analysis
```

## Project Structure

```
fieldmark-go/
├── cmd/
│   └── web/
│       └── main.go          # composition root — wires all dependencies, starts Fiber
├── internal/
│   ├── domain/
│   │   ├── entities/        # Project, Inspection, Violation, CorrectiveAction
│   │   ├── valueobjects/    # typed wrappers (ProjectCode, Severity, etc.)
│   │   ├── enums/           # domain enum constants
│   │   └── errors/          # typed domain errors (InvalidStateTransition, etc.)
│   ├── app/
│   │   ├── services/        # use-case orchestration; delegates to domain
│   │   ├── ports/           # persistence interfaces the app layer depends on
│   │   └── dto/             # data transfer objects crossing app ↔ web boundary
│   ├── data/
│   │   └── postgres/
│   │       ├── db.go        # connection setup
│   │       ├── stores/      # ProjectStore, InspectionStore, ViolationStore, etc.
│   │       └── models/      # scan targets for SQL query results
│   └── web/
│       ├── handlers/        # Fiber route handlers — thin; no business logic
│       ├── middleware/       # auth, request-id, logging
│       ├── templates/
│       │   ├── layouts/     # base.html
│       │   ├── pages/       # full route surfaces
│       │   ├── partials/    # shared markup (nav, header, footer)
│       │   └── fragments/   # HTMX swap targets (compliance_tile, violation_row, etc.)
│       └── static/
│           ├── js/
│           └── vendor/      # htmx/, ag-grid/35.2.1/, fieldmark.css (symlink → fieldmark_style/dist/)
```

## Layer Responsibilities

### `internal/domain`
State-transition methods, `can_*` predicates, domain invariants, and typed errors. Must not import Fiber, database drivers, or anything outside the standard library.

### `internal/app`
Use-case orchestration. Coordinates domain calls and persistence via port interfaces. Must not import Fiber or render HTML.

### `internal/data/postgres`
Postgres connection, store implementations, and query logic. Implements the port interfaces defined in `internal/app/ports/`. Must not contain business rules.

### `internal/web`
Fiber route definitions, handlers, middleware, and template rendering. Must not contain business rules or query logic beyond calling `app/services`.

## Dependency Direction (hard rule)

```
web → app → domain
data/postgres → app, domain
```

- `domain` has zero outbound imports beyond the standard library.
- `app` depends on `domain` and its own port interfaces — never on concrete `data` types.
- `web` never reaches into `data` directly.
- `fiber.Ctx` must not escape the `web` layer.

## Database

This stack reads and writes to the shared `domain` schema and will eventually own `fiber_auth` for authentication.

**The `domain` schema is not created or migrated by this stack.** It is created by SQL init scripts in `docker/postgres/init/`. Stores in `internal/data/postgres/stores/` query `domain.*` tables via explicit SQL — no ORM, no auto-migration.

Domain tables store user identifiers as opaque values (`created_by_user_id`, etc.). Do not foreign-key `fiber_auth` tables from domain tables.

### Local connection

```
Host:     localhost
Port:     5432
Database: fieldmark
User:     fieldmark
Password: fieldmark
```

## Authentication

`fiber_auth` schema exists in the database but Fiber authentication is **deferred by design**. Do not scaffold auth middleware or user models until:
- Domain schema is stable
- Feature work explicitly requires it

## Hard Rules

- No business rules in handlers or middleware.
- `fiber.Ctx` stays in `internal/web` — never passes to `app` or `domain`.
- No persistence details in `domain` or `app`.
- No generic repository abstractions (`Repository[T]`, `Store[T]`, etc.). Stores are domain-specific and narrow.
- No auto-migration of the `domain` schema — ever.
- HTML responses are the default. JSON only for AG Grid data endpoints.
- No client-side state stores or workflow orchestration.
- Tests use real PostgreSQL. No mocking the database.

## Agent Behaviour Rules

- Scaffold structure before adding logic. Folders and interfaces before implementations.
- If a handler contains an `if` branch that evaluates a business rule, move the rule to the entity.
- If a solution requires explaining the layer it lives in, it is probably in the wrong layer.
- Do not add port interfaces speculatively — define them when a concrete dependency needs inverting.
- Prefer explicit SQL over any query-builder abstraction unless the abstraction is already present.

## Reference

- `_bmad-output/planning-artifacts/research/fiber-reference.md` — full Go/Fiber guardrails (authoritative)
- `_bmad-output/planning-artifacts/research/architecture-decisions.md` — ADRs and hard constraints
- `docs/FieldMark_Fiber_Architecture_and_Standup_Guide.md` — standup guide and architecture narrative
