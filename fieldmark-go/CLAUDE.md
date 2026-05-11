# CLAUDE.md — Go Stack

This file provides guidance to Claude Code (claude.ai/code) when working in the `fieldmark-go/` Go project. Read alongside the root `CLAUDE.md`.

## Commands

Run from the `fieldmark-go/` directory.

### Application

```bash
go run ./cmd/web/
go build -o fieldmark-go ./cmd/web/
```

### QA

Tool versions are pinned in `go.mod` / `go.sum` via the root `tool (...)` block. Run CLI tools without global installs:

```bash
go tool golang.org/x/tools/cmd/goimports -w .
go tool honnef.co/go/tools/cmd/staticcheck ./...
```

**Makefile** (POSIX `make`: Git Bash, WSL, Linux, macOS):

```bash
make fmt           # gofmt + goimports (write)
make fmt-check     # fail if formatting needed (CI)
make vet           # go vet ./...
make staticcheck   # staticcheck ./...
make test          # go test ./...
make check         # fmt-check + vet + staticcheck + test
make lint          # golangci-lint run ./... (needs binary on PATH)
make check-all     # make check + lint
```

Without `make`, run the same steps manually using `gofmt`, `go vet ./...`, `go tool …`, and `go test ./...`.

**golangci-lint** is optional aggregate linting (`.golangci.yml`). Install the [v2 binary](https://golangci-lint.run/welcome/install/) separately; it is not a Go module dependency.

### IDE

GoLand and VS Code Go integration typically run **gofmt**, **go vet**, and **staticcheck** on save or in the analysis pass—align editor settings with the checklist rather than adding alternate linters.


## Project Structure

Flat layered layout. Each layer is a single package; no nested sub-packages by concept (entities/, valueobjects/, etc.) — Go's package boundary is what enforces architectural separation.

```
fieldmark-go/
├── cmd/
│   ├── web/main.go          # entry: parse env, build pgxpool, build template engine, mount routes
│   ├── seed/main.go         # dev seed runner (when fiber_auth lands)
│   └── tools/dumproutes.go  # `go run ./cmd/tools/dumproutes` — emits route inventory for parity tooling
└── internal/
    ├── domain/              # PURE — Project, Inspection, Violation, CorrectiveAction, AuditEntry, reference data,
    │                        #         state-transition methods, can_* predicates, *RuleError types,
    │                        #         compliance rule + scoring code. No Fiber. No pgx. Standard library only.
    ├── data/                # explicit SQL via pgx; narrow Store interfaces (ProjectStore, ViolationStore, etc.).
    │                        # No business rules.
    ├── app/                 # THIN coordinator — dependency wiring ONLY (Deps struct, env config, ActorFromCtx).
    │                        # MUST NOT contain business rules or use-case orchestration.
    └── web/                 # Fiber handlers, html/template rendering, viewmodels, ssrm parser, auth middleware.
                             # `fiber.Ctx` must not escape this package.
```

## Layer Responsibilities

- **`internal/domain`** — entities and behavior. Zero outbound imports beyond the standard library.
- **`internal/data`** — Postgres access. Narrow per-aggregate `Store` interfaces with concrete pgx implementations. No business rules. No generic `Repository[T]`.
- **`internal/app`** — wiring only. The `Deps` struct holds the DB pool, the Stores, and the authz checker. This is where `main.go` composes the application graph. **Not** a service layer; **no use-case orchestration**; **no business rules**. Handlers in `web` call domain methods directly with stores from `Deps`.
- **`internal/web`** — Fiber handlers, html/template files, view models with `can_*` booleans, AG Grid SSRM payload parser, auth middleware (stub for MVP per ADR-012).

## Dependency Direction (hard rule)

```
web → app → domain
web → data → domain
```

- `domain` has zero outbound imports beyond the standard library.
- `app` is a Deps container; it imports `domain` and `data` for type wiring only.
- `web` reaches into `data` through `app.Deps`, not directly to data implementations.
- `fiber.Ctx` must not escape the `web` package.

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

## Hard Rules (Go-specific)

Root `CLAUDE.md` covers cross-stack rules (no client-side state, no fat service layers, real PostgreSQL in tests, infrastructure-owned `domain` schema). The Go-specific rules are:

- No business rules in handlers, middleware, or `internal/app/`.
- `fiber.Ctx` stays in `internal/web` — never passes to `app` or `domain`.
- No persistence details in `domain` or `app`.
- No generic repository abstractions (`Repository[T]`, `Store[T]`, etc.). Stores are per-aggregate and narrow.
- HTML responses are the default. JSON only for AG Grid data endpoints.

## Agent Behaviour Rules

- Scaffold structure before adding logic. Folders and interfaces before implementations.
- If a handler contains an `if` branch that evaluates a business rule, move the rule to the entity.
- If a solution requires explaining the layer it lives in, it is probably in the wrong layer.
- Do not add port interfaces speculatively — define them when a concrete dependency needs inverting.
- Prefer explicit SQL over any query-builder abstraction unless the abstraction is already present.

## Reference

- `_bmad-output/planning-artifacts/architecture.md` — architectural source of truth (canonical request flow with Go/Fiber code stub, decisions, patterns)
- `_bmad-output/planning-artifacts/prd/` — capability source of truth
- Root `CLAUDE.md` — cross-stack rules and canonical inventories (audit actions, HTMX target IDs, method names)
