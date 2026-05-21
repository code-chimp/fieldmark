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

This stack reads and writes to the shared `domain` schema and owns the framework-local `fiber_auth` schema for stub authentication (see §Authentication).

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

The Go stack uses a **stub authentication middleware** (ADR-012 explicit deferral). Real auth — sessions, password hashing, CSRF, login forms, user management UI — is an epic-sized follow-on, not MVP scope.

**Where it lives:** `internal/web/auth/` (middleware + lookup). The hydrated principal type is `app.Actor` in `internal/app/actor.go`.

**How identity is resolved (in order):**
1. `X-FieldMark-Actor` cookie (set by Story 1.11's /login user-switcher).
2. `X-FieldMark-Actor` HTTP header (for scripted tests and ad-hoc curl).
3. `FIELDMARK_STUB_ACTOR` environment variable (deployment-fixed fallback).
4. Otherwise: anonymous (sentinel `app.Anonymous()`).

The resolved username is looked up against `fiber_auth.users` joined to `fiber_auth.user_roles`. On miss or DB error, the request binds the anonymous actor and continues — the application stays navigable while the developer investigates.

**Schema ownership:** `fiber_auth.users` and `fiber_auth.user_roles` are framework-local (ADR-012), Go-owned, defined in `internal/data/postgres/migrations/fiber_auth/001_initial.sql`. Apply with `go run ./cmd/migrate-fiber-auth` after `make reset`. **Never** colocate this DDL with `domain.*` init scripts — `domain` is infrastructure-owned (ADR-014).

**Dev user seeding:** `cmd/seed/main.go` populates `fiber_auth.users` and `fiber_auth.user_roles` from the shared manifest at `docker/postgres/init/seed-uuids/dev-users.json`. Run with `go run ./cmd/seed` (or `make seed-go` from the repo root). The manifest's `password` field is **intentionally not persisted** — the Go stack uses stub auth and has no password storage. This is a deliberate design choice (documented in the seeder's package comment), not an oversight. When the deferred real-auth epic lands, `fiber_auth.users` will gain a `password_hash` column and the seeder will grow to hash and persist.

**Story 1.11 shipped:**
- `/login` — user-switcher stub: lists seeded users as Basecoat buttons; sets the `X-FieldMark-Actor` cookie on submit.
- `/logout` — clears the cookie and redirects to `/login`.
- `auth.RequireAuth()` is mounted on all business routes; anonymous requests redirect to `/login`.
- `auth.RequireRole(role)` is available for future role-gated routes (`internal/web/auth/authz.go`).

**Replacing the stub with real auth is out of MVP scope.** Do not grow the stub incrementally — when real auth lands, it lands as a coherent epic (session tables, password hashing via `golang.org/x/crypto/bcrypt`, CSRF middleware, real login form, registration/management UI, password reset flow). Until then: this is the stub posture.

## Authorization

The single Go-side authorization decision primitive is `auth.Can` in `internal/web/auth/authz.go`. Signature:

```go
auth.Can(actor *app.Actor, action string, entityID uuid.UUID) bool
```

**Rules:**
- Handlers call `auth.Can`; view models carry the result as a `bool` field — templates must never call `Can` directly.
- Role names are defined in `internal/domain/role.go` as `domain.RoleAdmin`, `domain.RoleComplianceOfficer`, etc. (`type Role string` consts). Hard-coded role-name string literals elsewhere are a defect.
- Actions are registered at composition time via `auth.RegisterAction(action, roles...)`. Typically called from `cmd/web/main.go` or a per-aggregate `init()` in `internal/web/handlers/`. Story 1.12 ships the map empty — Epic 1 has no live action affordances.
- Entity-scope rules are deferred to Epic 2+ and will wire into `evaluateEntityScope` inside `authz.go`.

**ActionButton template:** `internal/web/templates/components/action_button.html`, defined as `{{ define "action_button" }}`. Invoked via `{{ template "action_button" .ActionButton }}` where `.ActionButton` is a `viewmodels.ActionButtonVM` carrying pre-computed `Permission` (from `Can`) and `StateAllows` (from the entity's `can_*` predicate). The template handles the trichotomy internally.

Canonical snapshot reference: `fieldmark_shared/components/action_button.example.html`.

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

## Home page

The Home page lives at `internal/web/templates/pages/home.html` and is served by the `/` handler in `cmd/web/main.go` (via `renderHomeContext`).

**This page is intentionally empty in Epic 1.** It renders `<h1>FieldMark</h1>`, the role badge, and a placeholder paragraph only. Story 2.10 replaces it with the real Compliance Dashboard.

**Chrome composition order (AC #2, Story 1.13 — all three stacks must match):**
`<a class="fm-wordmark">` → `<div class="ml-auto flex items-center gap-3">` containing theme-toggle then avatar menu. Any new chrome control added to any stack must be added to all three in the same commit (FR58).

**Role → badge-token mapping** (locked in Story 1.13; source of truth is `internal/domain/role.go`):

| Role | Token | Label |
|---|---|---|
| `ADMIN` | `danger` | Admin |
| `COMPLIANCE_OFFICER` | `info` | Compliance Officer |
| `INSPECTOR` | `warning` | Inspector |
| `SITE_SUPERVISOR` | `neutral` | Site Supervisor |
| `EXECUTIVE` | `success` | Executive |

The badge `<span class="badge badge-{token}" role="status">{label}</span>` is the first cross-stack visual proof of identity. Never hard-code tokens or labels outside `role.go`.
