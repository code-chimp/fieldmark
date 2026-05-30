---
stepsCompleted: [1, 2, 3, 4, 5, 6, 7, 8]
lastStep: 8
status: 'complete'
completedAt: '2026-05-10'
deferredDecisions:
  - id: NET-MAPSTER
    summary: 'Resolved in step-04. Decision: REJECTED for MVP — manual projection preferred (low mapping volume, source-readability priority, cross-stack symmetry argument). Reopen conditions documented in §Core Architectural Decisions → NET-MAPSTER.'
    raisedAt: 'step-03-starter'
    resolvedAt: 'step-04-decisions'
    resolution: 'rejected-for-mvp'
  - id: CI-PIPELINE
    summary: 'CI/CD deferred per user feedback in step-04. PRD §Non-Goals already excludes CI/CD pipelines. Architectural-symmetry enforcement moves to Make targets + optional pre-commit hook. Revisit triggers documented in §Core Architectural Decisions → D18.'
    raisedAt: 'step-04-decisions'
    resolvedAt: 'step-04-decisions'
    resolution: 'deferred-to-post-thesis'
inputDocuments:
  - _bmad-output/planning-artifacts/prd/ (canonical, sharded — 12 sections + index)
  - _bmad-output/planning-artifacts/research/architecture-decisions.md (seed context — ADR-011 to ADR-014)
  - _bmad-output/planning-artifacts/research/domain-model.md (seed context — ERD, schema, state machines)
  - _bmad-output/planning-artifacts/research/domain-schema-ownership-primer.md (seed context)
  - _bmad-output/planning-artifacts/research/authentication-authorization-primer.md (seed context)
  - _bmad-output/planning-artifacts/research/playwright-e2e-philosophy.md (seed context)
  - _bmad-output/planning-artifacts/research/dotnet-reference.md (seed context)
  - _bmad-output/planning-artifacts/research/django-reference.md (seed context)
  - _bmad-output/planning-artifacts/research/fiber-reference.md (seed context)
  - _bmad-output/planning-artifacts/research/ux-guide.md (seed context)
  - _bmad-output/planning-artifacts/research/project-brief.md (seed context)
  - CLAUDE.md (root — observed-current-state, foundational but improvable)
  - FieldMark/ (.NET skeleton — observed layout)
  - fieldmark_py/ (Django skeleton — observed layout)
  - fieldmark-go/ (Go/Fiber skeleton — observed layout)
  - e2e/ (Playwright skeleton — observed layout)
inputDocumentsNote: 'Research files are seed context only — read once for understanding, not maintained going forward. This BMad-generated architecture document becomes the source of truth for architectural decisions, displacing research/architecture-decisions.md the same way the canonical PRD displaced research/prd.md. Skeleton stacks are observed-current-state; this document codifies original decisions and may identify gaps to revisit.'
workflowType: 'architecture'
project_name: 'FieldMark'
user_name: 'Tim'
date: '2026-05-09'
---

# Architecture Decision Document

_This document builds collaboratively through step-by-step discovery. Sections are appended as we work through each architectural decision together._

## Project Context Analysis

### Requirements Overview

**Functional Requirements (70 FRs across 14 categories, MVP + Growth).** The PRD's FR catalog is unusually architecture-shaped — many FRs read as direct architectural commitments rather than capability statements. The functional surface decomposes into:

- **Identity & access (FR1–FR8):** framework-local authentication; conceptual roles (Admin, Compliance Officer, Inspector, Site Supervisor, Executive); server-decided action affordance (FR6 — buttons absent vs. disabled vs. present is a server call).
- **Project lifecycle (FR9–FR15):** Active ↔ OnHold + Active → Closed gated by `can_close()` (closure-gate rules: no open violations + every assigned trade has ≥1 Pass/Conditional inspection).
- **Inspection workflow (FR16–FR21):** Scheduled → InProgress → Completed; Fail-class findings auto-spawn Violations atomically in the same transaction.
- **Violation lifecycle (FR22–FR27):** Open → InProgress → Resolved (terminal) or Voided (terminal). No reopen path. Due date computed at open time from severity, immutable thereafter (FR22a).
- **Corrective Action (FR28–FR33):** Submitted → UnderReview → Approved/Rejected. Submitter ≠ reviewer. Only the latest non-Rejected may be approved. Approval is the single canonical resolution path for a Violation.
- **Compliance Rules Engine & Scoring (FR34–FR38):** Server-only evaluation; per-project 0–100 score; same-transaction recomputation; ClosureGate vs. ScoringPenalty rules; configurable parameters via reference data.
- **Audit Trail (FR39–FR43):** AuditEntry per domain mutation, same-transaction write, append-only, before/after JSON, opaque actor UUID.
- **Dashboard / AG Grid / Cross-Cutting (FR44–FR59):** HTMX partial-refresh dashboard; AG Grid server-side row model on ≥2 views with `{rows, lastRow}` contract; HTTP 409 + originating partial on rule violation; identical routes/IDs/contracts/method-names across all three stacks.
- **Accessibility (FR60–FR64):** WCAG 2.1 AA; HTMX-specific concerns (focus management on swaps, `aria-live` for OOB, `hx-disabled-elt`).
- **Test discipline (FR65–FR66):** Playwright E2E across all three stacks; per-stack unit tests for domain invariants.
- **Growth (FR67–FR70):** Reference-data admin UI, executive trend dashboard, multi-stack parity test suite, runtime ComplianceRule parameter editing.

**Non-Functional Requirements (architecturally binding).** Mostly locked in §Success Criteria → Measurable Outcomes and §Architectural Constraints (PRD-Binding):

- **Performance:** ≤200ms p95 partial-swap (action → panel + tile + audit row, one round trip); ≤300ms p95 grid → detail panel; same-transaction score recomputation; cross-stack divergence > 50ms p95 = defect.
- **Cross-stack symmetry:** zero-diff `pg_indexes` snapshot; zero-diff route inventory; identical HTMX target IDs (`#project-detail`, `#compliance-tile`, `#violation-detail`, `#audit-log`); identical AG Grid contract; identical audit action strings; identical canonical method-name list.
- **Backend authority:** server-only rules/validation/authorization; client requests HTML; HTMX is sole interactivity beyond AG Grid.
- **Schema ownership:** `domain` schema is infrastructure-owned (`docker/postgres/init/`, hand-authored SQL); `dotnet_auth` / `django_auth` / `fiber_auth` are framework-local; no FKs from `domain.*` to any auth schema; user references are opaque UUIDs.
- **Testability:** real PostgreSQL only (Testcontainers / pytest-django / Go equivalent); SQLite forbidden in tests.
- **Auditability:** every domain mutation writes an AuditEntry in the same transaction; append-only at app level (revoked privileges in production).
- **Maintainability:** build-blocking lint/format (`dotnet format` + analyzers, `ruff`/`black`, `gofmt`/`golangci-lint`); architectural simplicity enforced (no CQRS / Repository / MediatR / AutoMapper / fat service layers / client-side state stores).
- **Accessibility:** WCAG 2.1 AA, axe-core/playwright in every E2E scenario.

**Scale & Complexity:**

- Primary domain: **Web application** (server-rendered MPA + HTMX + AG Grid islands; .NET Razor Pages, Django Templates, Go/Fiber `html/template`).
- Complexity level: **Medium-high.** Drivers: three-stack symmetry; four aggregate roots with non-trivial state machines; server-evaluated rules engine with closure gates and scoring; ADR-locked guardrails; infrastructure-owned shared schema across three frameworks; cross-stack parity gates in CI.
- Estimated architectural surfaces: 3 application stacks × {Web/Domain/Data layers per stack} + 1 infrastructure-owned PostgreSQL schema + 1 shared compiled CSS pipeline + 1 cross-stack E2E suite + Docker Compose harness ≈ **~14–16 distinct surfaces** with parity contracts between them.

### Technical Constraints & Dependencies

**Locked technology choices** (by ADR / PRD; not negotiable without amendment):

- **PostgreSQL 17** — sole datastore. SQLite forbidden in tests. Schemas: `domain` (infra-owned), `django_auth`, `dotnet_auth`, `fiber_auth`, `infra`. UUIDs generated in app code (not `gen_random_uuid()`). Enums stored as VARCHAR + CHECK constraints (not native PG ENUMs). All timestamps `TIMESTAMPTZ`.
- **HTMX 4.x** + **AG Grid Enterprise 35.x** — versions pinned identically across all stacks; mismatch is build-blocking. Enterprise is used to demonstrate the true Server-Side Row Model; the demo runs without a license key and the "unlicensed" watermark is an accepted, deliberate tradeoff.
- **Tailwind CSS v4** — single source `fieldmark_shared/src/fieldmark.css`; compiled `dist/` is committed and symlinked into all three apps.
- **.NET 10 / Razor Pages / EF Core** — `EFCore.NamingConventions` for `snake_case`; `ToTable("...", "domain")` fluent config; rich domain entities; no AutoMapper. Skeleton already enforces `Nullable=enable`, `TreatWarningsAsErrors=true`, `AnalysisMode=Recommended`, `EnforceCodeStyleInBuild=true` via `Directory.Build.props`.
- **Python 3.14+ / Django 6.x / Django ORM** — `Meta.managed = False` + `db_table = 'domain"."project'` for shared tables; `ruff` + `black` + `mypy` + `pytest-django`. Skeleton already pins these.
- **Go 1.26+ / Fiber v3** — `cmd/web` entry; `internal/{app,data,domain,web}` layered structure (thin `app` coordination layer is the only place explicit dependency wiring lives — must not contain business rules); explicit SQL via `pgx/v5`; `gofiber/template/html/v2`.
- **Playwright + axe-core/playwright** — sole E2E mechanism; runs against all three stacks; same scenario passes in all three. TypeScript-authored under `e2e/`.
- **Docker Compose** — sole local-dev orchestration; init scripts in `docker/postgres/init/` create schemas and `domain.*` DDL.

**Naming conventions** (all stacks conform; zero-diff is contract):

- Database/wire format: `snake_case` (canonical — Django-natural; .NET via `UseSnakeCaseNamingConvention()`; Go via explicit SQL).
- Code identifiers: idiomatic per stack (C# `PascalCase`, Python `snake_case`, Go `PascalCase` exported / `camelCase` unexported).
- Enum values: `SCREAMING_SNAKE_CASE` strings on wire and at rest.
- Domain method names: canonical list `start / complete / cancel / place_on_hold / resume / close / assign / submit_corrective_action / approve_resolution / reject_resolution / void`.

**Skeleton-state observations** (to convert from incidental to contractual in later steps):

- `Directory.Build.props` enforces nullable + warnings-as-errors + `EnforceCodeStyleInBuild` — already aligned with PRD's Maintainability NFR.
- `pyproject.toml` pins ruff + black + mypy + pytest-django — already aligned.
- `go.mod` pins Fiber v3 + pgx/v5 + `gofiber/template/html/v2` — already aligned.
- `docker/postgres/init/` exists; presence of `001_schemas.sql` and `010_domain_tables.sql` to be confirmed in a later step.
- `e2e/` Playwright skeleton scaffolded with biome, fixtures, helpers, tests, `playwright.config.ts` — cross-stack scenario coverage status to be confirmed.

### Cross-Cutting Concerns Identified

- **Transaction discipline.** Every mutating handler in every stack must follow the canonical sequence: authorize → BEGIN → load aggregate → call entity method → append AuditEntry → recompute compliance score (when applicable) → COMMIT → render partial. Architecture document needs this as an enforceable shape with examples per stack, not a prose convention.
- **Aggregate-root encapsulation.** Cross-aggregate writes happen at the request-handler level, not via orchestration services. Architecture specifies how aggregate boundaries are honored across .NET (DbContext + entity methods), Django (model methods), and Go (explicit SQL + thin app coordinator).
- **Schema-source-of-truth.** `domain.*` DDL is the contract. EF Core fluent configs, Django `managed = False` models, and Go SQL queries are consumers. Architecture specifies the change-management workflow when DDL evolves.
- **HTMX swap semantics.** Partials need exactly one root element with a stable, cross-stack-identical `id`; `hx-swap-oob` is allowed only for header tiles and must be documented at every use site. Architecture provides a shared partial-naming and target-ID inventory.
- **AG Grid as island.** Server-side row model only; row selection fires HTMX detail-panel loads; no client-side row computation. Architecture specifies the JSON endpoint contract once, then binds each stack to it.
- **Authorization expression.** Conceptual roles map to native authorization machinery in each stack (ASP.NET Core policies/roles, Django auth groups, Fiber middleware vs. `fiber_auth`) — and the architecture specifies how role-based action-button rendering is implemented uniformly so FR6 holds in observable behavior across stacks.
- **Cross-stack diff tooling.** PRD requires `pg_indexes` zero-diff and route-inventory zero-diff. Architecture specifies how and where these diffs run (CI step, local Make target, or both) so they're enforced, not aspirational.
- **Skeleton gaps to surface as steps progress.** Per direction to codify originals + identify gaps: (a) `docker/postgres/init/010_domain_tables.sql` presence; (b) per-stack `domain.*` mapping for at least one aggregate; (c) cross-stack Playwright scenario stubs; (d) CI configuration. These steer later step questions.

## Starter Template Evaluation

### Primary Technology Domain

**Web application** — server-rendered MPA with HTMX-driven partial updates and AG Grid as a scoped JavaScript island. Implemented three times in parallel (.NET / Django / Go-Fiber) against one shared infrastructure-owned PostgreSQL domain schema.

### Starter Options Considered

**No third-party starter template applies.** FieldMark's architectural posture (no SPA, no client-side state stores, no CQRS / MediatR / Repository abstractions, no AutoMapper, infrastructure-owned shared domain schema across three frameworks, cross-stack symmetry as a defect class) is too constrained for any general-purpose starter (T3, RedwoodJS, Blitz, Next.js+adapters) to align with — and any framework-specific community template (`Razor.Templates.AspNetCore`, `cookiecutter-django`, Fiber community boilerplates) would smuggle in patterns the ADRs forbid (typically a Repository layer, a Service layer, AutoMapper, a JS bundler for static assets, or auth scaffolding that assumes a single canonical user model).

**Decision: native framework CLI scaffolding only**, with hand-authored architectural overlays. Each stack's scaffolding command is documented below alongside the architectural decisions it established (or failed to establish, which we then layered on).

### Selected Starter: Native CLI scaffolding per stack

**Rationale for Selection:**

- ADR-011 (no CQRS, no Repository, no Clean Architecture) makes most opinionated starters unusable.
- ADR-014 (infrastructure-owned `domain` schema) means no framework migration tooling should generate the canonical tables — disqualifying any starter that bakes in `dotnet ef migrations` / `manage.py migrate` as the schema authority for shared business data.
- The cross-stack-symmetry rule (PRD §Architectural Constraints → Stack Symmetry) requires identical routes, target IDs, and method names across three frameworks. No starter targets all three.
- Lean foundations also keep "what's there and why" auditable for the talk audience — the artifact's persuasive purpose depends on the architecture being readable in source.

### Initialization Commands (per stack)

#### .NET — `FieldMark/`

```bash
# From repository root
dotnet new sln -n FieldMark -o FieldMark
cd FieldMark

dotnet new webapp -n FieldMark.Web -f net10.0          # Razor Pages app
dotnet new classlib -n FieldMark.Domain -f net10.0     # Pure domain (entities, exceptions, value objects)
dotnet new classlib -n FieldMark.Data -f net10.0       # EF Core DbContext + fluent configuration

dotnet new xunit -n FieldMark.Tests.Domain -f net10.0
dotnet new xunit -n FieldMark.Tests.Integration -f net10.0   # Testcontainers for real Postgres

dotnet sln add **/*.csproj

dotnet add FieldMark.Web reference FieldMark.Data FieldMark.Domain
dotnet add FieldMark.Data reference FieldMark.Domain

# Solution-wide build hygiene via Directory.Build.props (already committed):
#   <Nullable>enable</Nullable>
#   <ImplicitUsings>enable</ImplicitUsings>
#   <TreatWarningsAsErrors>true</TreatWarningsAsErrors>
#   <AnalysisMode>Recommended</AnalysisMode>
#   <EnforceCodeStyleInBuild>true</EnforceCodeStyleInBuild>
```

**Architectural decisions established by this scaffolding:**

- 4-project solution: Web (Razor Pages) → Data (EF Core mapping only) → Domain (pure, no framework dependencies); two test projects scoped to domain-rule unit tests and integration tests against real Postgres via Testcontainers. No "Application" or "Services" project (ADR-011).
- Build hygiene: nullable reference types enforced, all warnings are errors, `EnforceCodeStyleInBuild` makes IDE style rules fire in `dotnet build` / CI not just in the IDE.
- Razor Pages over MVC for the simpler page-handler model (one page = one URL = one handler, matches HTMX's request shape better).

#### Django — `fieldmark_py/`

```bash
# From repository root
uv init fieldmark_py --python 3.14
cd fieldmark_py
uv add 'django>=6.0' 'psycopg[binary]>=3.3'
uv add --dev 'ruff' 'black' 'mypy' 'django-stubs' 'pytest' 'pytest-django'

uv run django-admin startproject fieldmark .             # fieldmark/{settings,urls,wsgi,asgi}.py + manage.py

# One Django app per aggregate / functional area
uv run python manage.py startapp projects
uv run python manage.py startapp inspections
uv run python manage.py startapp violations
uv run python manage.py startapp compliance              # Rules engine + scoring
uv run python manage.py startapp audit
uv run python manage.py startapp reference               # TradeType, ViolationCategory, ComplianceRule
uv run python manage.py startapp grid                    # AG Grid endpoint helpers

mkdir -p templates static
```

**Architectural decisions established by this scaffolding:**

- App-per-aggregate layout: `projects`, `inspections`, `violations`, `compliance`, `audit`, `reference`, `grid` — each Django app maps to one functional concern; cross-app imports are allowed only for entity types and signal connections, not for orchestration.
- `uv` for dependency management and venv (Python 3.14+); `ruff` + `black` + `mypy` + `pytest-django` already pinned in `pyproject.toml` and enforced by build.
- Domain models will use `Meta.managed = False` and an explicit cross-schema `db_table = 'domain"."project'` (ADR-014). Standard `manage.py startapp` produces a managed model by default — this is an override applied per-app for shared-domain models. Auth-schema models (`django_auth.*`) remain framework-managed.

#### Go / Fiber — `fieldmark-go/`

```bash
# From repository root
mkdir fieldmark-go && cd fieldmark-go
go mod init github.com/code-chimp/fieldmark-go

go get github.com/gofiber/fiber/v3
go get github.com/gofiber/template/html/v2
go get github.com/jackc/pgx/v5

# Layered internal package layout (the only stack with an explicit `app` coordination layer)
mkdir -p cmd/web internal/{app,data,domain,web} tools

# cmd/web/main.go        — entry point: wire Fiber + DB pool + template engine, mount routes
# internal/domain/       — entities, value objects, state-transition methods, invariants
# internal/data/         — explicit SQL via pgx; small narrow Store interfaces
# internal/app/          — thin coordination layer: dependency wiring only, NO business rules
# internal/web/          — Fiber handlers, HTMX partials, AG Grid endpoint marshalling
```

**Architectural decisions established by this scaffolding:**

- Standard Go layout (`cmd/` for binaries, `internal/` for non-importable application code).
- Layer split: `domain` (pure) ← `data` (SQL) ← `app` (wiring) ← `web` (handlers). Dependency direction is one-way; `fiber.Ctx` does not escape the `web` package (per ADR-011 spirit applied to Go).
- The `app` package is the only legitimate place for explicit dependency wiring across the three stacks. It must remain a coordinator — no business rules — per the PRD's note that diverging Go from .NET/Django here is acceptable only because Go lacks the DI ergonomics those stacks have.
- `pgx/v5` chosen over `database/sql` + driver for explicit, fast, and type-aware Postgres access; matches the "explicit SQL against `domain.*`" rule (ADR-011).

#### Cross-stack — `e2e/`, `fieldmark_shared/`, `docker/`

```bash
# Playwright E2E (TypeScript + biome)
mkdir e2e && cd e2e
pnpm init
pnpm add -D '@playwright/test' '@axe-core/playwright' typescript '@biomejs/biome'
pnpm dlx playwright install
# Authored: playwright.config.ts (parallel projects per stack), fixtures/, helpers/, tests/

# Tailwind v4 (single source of CSS for all three stacks)
cd ../fieldmark_shared
pnpm init
pnpm add -D 'tailwindcss@4'
# src/fieldmark.css → dist/fieldmark.css (committed); symlinked into each app's static dir

# Docker Compose harness
cd ..
# docker-compose.yml: postgres:17 only
# docker/postgres/init/: hand-authored SQL — 001_schemas.sql + 010_domain_tables.sql + seed scripts
```

**Architectural decisions established by this scaffolding:**

- One Playwright suite, three parallel projects (one per stack); same scenarios run against all three; `@axe-core/playwright` embedded in every scenario.
- Tailwind compiled CSS committed to repo; no per-stack CSS authoring; symlinks into each app's static directory. CSS authoring lives once.
- `docker compose up -d` is the only local-dev command that touches infrastructure. Postgres init scripts run automatically on first volume creation; `docker compose down -v && docker compose up -d` is the documented re-init sequence.

### What This Scaffolding Does *Not* Provide (and must be hand-authored)

Architectural concerns the BMad architecture document needs to fill in over the remaining steps — *gaps* flagged for resolution:

- **`docker/postgres/init/010_domain_tables.sql`** — the canonical `domain.*` DDL (ERD in `research/domain-model.md` §8 has a sketch). Status: presence to confirm.
- **EF Core fluent configuration for `domain.*` mapping** — `ToTable("project", "domain")` + `UseSnakeCaseNamingConvention()` + value converters for enum-to-string. Status: skeleton has `FieldMark.Data/Configuration/` but coverage to confirm.
- **Django shared-domain models with `Meta.managed = False`** — separate from auth models in each app. Status: per-app domain models to confirm.
- **Go `Store` interfaces and pgx implementations for `domain.*`** — narrow per-aggregate (`ProjectStore`, `ViolationStore`, etc.). Status: `internal/data/` to inspect.
- **Cross-stack route inventory + diff tooling** — the `pg_indexes`-zero-diff and route-inventory-zero-diff PRD requirements have no scaffolding implementation yet.
- **Authentication wiring per stack** — Django gets it for free; .NET requires `dotnet_auth` schema + ASP.NET Core Identity wiring; Go currently has no auth wiring (deferred per ADR-012).
- **Seed scripts using identical UUIDs across stacks** — referenced in `domain-model.md` §3.11 but implementation status to confirm.

### Deferred Decisions Raised at this Step

- **NET-MAPSTER** — User has flagged Mapster as acceptable in the .NET project *if it simplifies architecture*. AutoMapper remains forbidden (licensing change + low ROI for this domain depth). The architect agent will weigh inclusion / exclusion in the .NET-specific decisions step (typically: are there enough entity → view-model boundaries to justify a mapper, or does the project's small surface mean manual projection is clearer?). Default position: manual projection unless a concrete pain point is identified.

**Note:** Project initialization using these commands is the historical record (skeletons already exist). For any future stack rework, the commands above are the canonical starting points. Future stories should not change scaffolding commands silently — any change to the foundational layout is an ADR amendment.

## Core Architectural Decisions

### Decision Priority Analysis

**Critical (block implementation):** EF Core driver/version, Postgres init-script ordering, ASP.NET Core Identity schema config, AG Grid endpoint URL convention, partial-naming convention per stack, cross-stack diff tooling location.

**Important (shape architecture):** Mapster decision (NET-MAPSTER → resolved as REJECTED for MVP), HTMX/AG Grid asset loading strategy, runtime config conventions, same-UUID seed strategy.

**Deferred:** Go/Fiber authentication wiring (ADR-012); production hosting; secrets management (no production target); **CI/CD pipeline (deferred to post-thesis-validation per user direction in this step; revisit triggers below)**.

### Data Architecture

**Already Locked:**

| Concern | Decision | Source |
|---|---|---|
| Database | PostgreSQL 17, single instance | PRD §Architectural Constraints |
| Schemas | `domain`, `django_auth`, `dotnet_auth`, `fiber_auth`, `infra` | ADR-013 |
| `domain.*` ownership | Infrastructure SQL (`docker/postgres/init/`) | ADR-014 |
| `*_auth.*` ownership | Framework-local migrations only | ADR-012 |
| Naming convention (db/wire) | `snake_case` canonical | `domain-model.md` §9 |
| Enum storage | `VARCHAR + CHECK`, `SCREAMING_SNAKE_CASE` strings | `domain-model.md` §9 |
| UUID generation | App code (not `gen_random_uuid()`) | `domain-model.md` §8 |
| Timestamps | `TIMESTAMPTZ` (UTC) | `domain-model.md` §8 |
| ORM pattern | Rich domain entities, no Repository | ADR-011 |
| Caching | None (no Redis, no in-process) | PRD §Non-Goals |

**Open Decisions Resolved:**

- **D1 — EF Core driver and naming-convention package:** `Npgsql.EntityFrameworkCore.PostgreSQL` (latest 9.x for .NET 10) + `EFCore.NamingConventions`. The Npgsql provider is the de-facto Postgres EF Core driver; `EFCore.NamingConventions` provides the global `UseSnakeCaseNamingConvention()` hook.
- **D2 — Postgres init script ordering:** numeric prefixes with 10-spacing for insertion room.
  ```
  docker/postgres/init/001_schemas.sql        # CREATE SCHEMA domain, *_auth, infra; GRANT
  docker/postgres/init/010_domain_tables.sql  # All domain.* DDL (ADR-014 canonical)
  docker/postgres/init/020_domain_indexes.sql # Cross-stack index inventory
  docker/postgres/init/090_seed_reference.sql # TradeType, ViolationCategory, ComplianceRule
  docker/postgres/init/091_seed_dev_users.sql # Generated by per-stack seed runners; identical UUIDs
  ```
  Status: `001_schemas.sql` exists; remainder needs confirmation/authoring.
- **D3 — Connection pooling:** framework-native pools, default sizing.
  - .NET: `AddDbContextPool<>`; max pool size 100 default.
  - Django: `psycopg` pooling via `CONN_MAX_AGE = 60`.
  - Go: `pgxpool` (already pulled via `pgx/v5`); pool size = 4× CPU.
  No PgBouncer for MVP.
- **D4 — Auth-schema migrations:** locked by ADR-012; framework-native, scoped to the matching `*_auth` schema. Documented usage:
  - .NET: `dotnet ef migrations add <name> --output-dir Migrations/Auth` against a separate `AuthDbContext` whose schema target is `dotnet_auth`.
  - Django: built-in `auth` app's migrations, with tables targeted at `django_auth` via the project's DB router or `db_table` overrides.
  - Go: deferred until auth is wired.
- **D5 — Connection string standardization:** single env var `FIELDMARK_DATABASE_URL` across all stacks (Postgres URL form: `postgresql://user:pass@host:port/db`). Each stack parses it natively. Local default exposes Postgres on `localhost:5432` with `fieldmark/fieldmark/fieldmark`.

### Authentication & Security

**Already Locked:**

| Concern | Decision | Source |
|---|---|---|
| Authentication strategy | Framework-local | ADR-012 |
| Authorization model | Domain-driven, native per stack | ADR-012 |
| Roles | `ADMIN`, `COMPLIANCE_OFFICER`, `INSPECTOR`, `SITE_SUPERVISOR`, `EXECUTIVE` | `domain-model.md` §3.12 |
| Server-side authority | All rules / validation / authorization | ADR-011, PRD |
| CSRF protection | Framework-native | PRD §NFR Security |
| Password hashing | Framework-native salted | PRD §NFR Security |
| SQL parameterization | Required; no string-concatenated SQL | PRD §NFR Security |
| Action-button rendering | Server-decided (absent / disabled / enabled) | FR6 |
| User refs in `domain.*` | Opaque UUID, no FK | ADR-012 |

**Open Decisions Resolved:**

- **D6 — ASP.NET Core Identity configuration:** ASP.NET Core Identity with schema target `dotnet_auth`, snake_case table mapping, password rules: `RequireDigit = true`, `RequireLowercase = true`, `RequireUppercase = true`, `RequireNonAlphanumeric = false`, `RequiredLength = 10`. Tables: `dotnet_auth.users`, `dotnet_auth.roles`, `dotnet_auth.user_roles`, `dotnet_auth.role_claims`, `dotnet_auth.user_claims`, `dotnet_auth.user_logins`, `dotnet_auth.user_tokens`. Configure via `modelBuilder.HasDefaultSchema("dotnet_auth")` on the Identity DbContext.
- **D7 — Django auth:** built-in `auth` system, no custom user model. Conceptual roles map to Django Groups. Tables in `django_auth` via DB router. Seed five Groups (`ADMIN`, `COMPLIANCE_OFFICER`, `INSPECTOR`, `SITE_SUPERVISOR`, `EXECUTIVE`) on first migration.
- **D8 — Go/Fiber auth:** **Deferred** to post-anchor-workflow (ADR-012 explicitly allows). Anchor Workflow MVP epic does not require Go-stack auth; Go scenarios run with stub middleware injecting a configurable `actor_id` UUID. Real auth is its own follow-on epic.
- **D9 — Same-UUID seed strategy:** per-stack seed runners reading a shared UUID manifest:
  ```
  docker/postgres/init/seed-uuids/dev-users.json   # canonical UUIDs + role assignments per username
  FieldMark/FieldMark.Web/SeedData/DevUsers.cs     # reads JSON, writes to dotnet_auth via Identity
  fieldmark_py/.../management/commands/seed_dev_users.py    # reads JSON, writes via Django ORM to django_auth
  fieldmark-go/cmd/seed/main.go                    # reads JSON, writes via SQL to fiber_auth (when auth lands)
  ```
  Each runner is idempotent. `domain.audit_entry.actor_id` for any seeded domain rows uses the same UUIDs from the manifest, so cross-stack audit comparison works.

### API & Communication Patterns

**Already Locked:**

| Concern | Decision | Source |
|---|---|---|
| Primary wire format | HTML partials via HTMX | PRD §Architectural Constraints |
| AG Grid wire format | JSON `{rows, lastRow}` (snake_case) | FR49, PRD |
| Rule-violation response | HTTP 409 + originating partial | FR55 |
| Authorization failure | HTTP 403 (or stack-equivalent) | FR56 |
| State-change methods | POST only; never GET | FR54 |
| HTMX target IDs (initial four) | `#project-detail`, `#compliance-tile`, `#violation-detail`, `#audit-log` | PRD §Architectural Constraints |
| GraphQL / general REST | Out of scope | PRD §Non-Goals |
| Rate limiting | Out of scope | PRD §Non-Goals |

**Open Decisions Resolved:**

- **D10 — AG Grid endpoint URL convention:** cross-stack-identical paths under `/grid/`:
  ```
  POST /grid/projects        # body: AG Grid SSRM request payload; response: {rows, lastRow}
  POST /grid/violations      # same shape
  POST /grid/inspections     # same shape
  POST /grid/audit/:projectId  # project-scoped audit log grid
  ```
  POST not GET — AG Grid SSRM payloads carry filter/sort objects that don't fit URLs. Routes registered identically in all three stacks; `make parity` (D18/D19) enforces.
- **D11 — HTMX target ID inventory (full canonical list, extending PRD's four):**
  ```
  #compliance-tile             — header-level OOB target
  #project-detail              — main detail panel
  #project-list                — dashboard project grid container
  #violation-detail            — main violation panel (within project detail)
  #violation-list              — violation tab content
  #inspection-list             — inspection tab content
  #audit-log                   — audit tab content
  #corrective-action-form      — submit/edit corrective action form
  #corrective-action-list      — list of CAs for a violation
  #flash-region                — aria-live region for non-OOB transient announcements
  ```
  Any new target ID requires an ADR amendment; this inventory is the contract.
- **D12 — Partial-naming convention per stack:** the *target ID inside the partial* is what's shared; file naming is idiomatic per stack.
  - .NET: `Pages/Shared/_ProjectDetail.cshtml` (Razor `_PascalCase.cshtml`).
  - Django: `templates/projects/_project_detail.html` (underscore prefix, `snake_case`).
  - Go: `internal/web/templates/projects/_project_detail.html`.
- **D13 — Error rendering pattern:** handlers catch a single typed `DomainRuleException` (per stack) and re-render the originating partial with an inline error message and unchanged state, returning HTTP 409. No global exception middleware for domain errors — they are expected outcomes. Authorization failures bubble to framework-native middleware (403) without entity-state leakage.

### Frontend Architecture

**Already Locked:**

| Concern | Decision | Source |
|---|---|---|
| Interactivity | HTMX 4.x | PRD §Architectural Constraints |
| JS islands | AG Grid Enterprise 35.x (true SSRM; unlicensed-watermark demo tradeoff) | PRD, FR48 |
| Client state stores | Forbidden | PRD §Forbidden Patterns |
| Routing | Server-driven (HTMX swaps) | PRD §Architectural Constraints |
| Styling | Tailwind v4, single compiled CSS | PRD §Web App Specific Requirements |
| Performance budgets | 200ms p95 swap; 300ms p95 grid→panel | PRD §Success Criteria |
| State-change methods | `<button hx-post>` only | PRD §Architectural Constraints |
| OOB swaps | Header tiles only; documented at every use site | PRD §Architectural Constraints |

**Open Decisions Resolved:**

- **D14 — AG Grid theming:** AG Grid Quartz theme compiled into `fieldmark_shared/dist/fieldmark.css` as part of the same Tailwind compile pass. Theme variables overridden in `fieldmark_shared/src/ag-grid-overrides.css` to align colors/spacing with the Tailwind palette.
- **D15 — Asset loading:** vendor locally; no CDN.
  - HTMX: `fieldmark_shared/vendor/htmx/htmx.min.js` (committed). Pinned: `4.0.0-beta2`.
  - AG Grid Enterprise 35.x: `fieldmark_shared/vendor/ag-grid/35.3.0/ag-grid-enterprise.min.js` (committed; the Enterprise UMD bundle includes Community). Pinned: `35.3.0`. No license key is set — the "unlicensed" watermark is an accepted demo tradeoff for showing Enterprise features (true SSRM).
  - Basecoat CSS component library: installed via `pnpm` in `fieldmark_shared/`. Pinned: `basecoat-css@0.3.11` (exact; no `^` or `~`). Pre-1.0 — treat minor bumps as breaking.
  - Symlinked into each app's `vendor/` static dir.
  Vendoring makes the version-pinning rule auditable (you can't audit a CDN URL pinned to "@latest"; you can audit a committed file).
- **D16 — Tailwind compilation trigger:** manual via `cd fieldmark_shared && npm run build` (npm script in `fieldmark_shared/package.json`). Compiled `dist/` is committed; CSS authoring is rare. Each stack's `CLAUDE.md` documents that CSS edits require a rebuild + commit. No watcher needed.

- **D20 — Single inline `<script>` exception (UX-DR5):** The application forbids inline JavaScript **with one deliberate exception**: a 5-line IIFE placed in `<head>` (after `<meta name="viewport">`, before the stylesheet `<link>`) that resolves the `system` theme preference before first paint.

  ```html
  <script>
  (function(){var d=document.documentElement,t=d.getAttribute('data-theme');
  if(t!=='system')return;
  d.setAttribute('data-theme',window.matchMedia('(prefers-color-scheme: dark)').matches?'dark':'light');})();
  </script>
  ```

  **Why it must be inline and blocking:** the script must run synchronously before the browser parses any stylesheet. An external `<script>` tag — even with `<link rel="preload">` — cannot guarantee zero-flash on first paint because resource loading is async. The `defer` attribute is prohibited here; a non-deferred external script would block HTML parsing entirely (worse). Inline is the only mechanism that is both synchronous and does not block the parser beyond its own 5 lines of execution.

  **What it does:** reads the server-rendered `data-theme` attribute on `<html>`. If the value is `"system"`, replaces it with `"light"` or `"dark"` via `matchMedia`. No-op for `"light"` or `"dark"` (which the server emitted directly from the `fm_theme` cookie). The cookie is HTTP-readable (no `HttpOnly`) so the client listener (`theme-toggle.js`) can also read it post-click.

  **This is the only inline JavaScript in the application.** Any future inline `<script>` must go through an architectural decision. The `theme-toggle.js` listener (`fieldmark_shared/vendor/theme-toggle/theme-toggle.js`) is loaded as an external `<script>` after HTMX and handles post-click DOM updates.

### Infrastructure & Deployment

**Already Locked:**

| Concern | Decision | Source |
|---|---|---|
| Hosting | Local development only | PRD §Non-Goals |
| Container | Docker Compose for Postgres | Project root CLAUDE.md |
| Production observability | Out of scope | PRD §Non-Goals |
| Scaling | Out of scope | PRD §Non-Goals |
| Audit log | Primary observability for domain events | FR39–FR43 |

**Open Decisions Resolved:**

- **D17 — Per-stack runtime configuration:** each stack reads from environment variables only — no `.env` checked in, no secrets vault for MVP.
  - Required env vars: `FIELDMARK_DATABASE_URL`, `FIELDMARK_LOG_LEVEL` (default `info`).
  - .NET: `IConfiguration` env var binding; Django: `os.environ` in `settings.py`; Go: `os.Getenv` wrapped in a small `internal/app/config` package.
  - Local dev defaults documented in each stack's `README.md`.

- **D18 — CI configuration: DEFERRED to post-thesis-validation.**
  Per user feedback in this step (and reinforced by PRD §Non-Goals which already excludes CI/CD pipelines), GitHub Actions is dropped from MVP scope. Architectural-symmetry enforcement moves to **local discipline**:
  1. **Make targets** — `make parity` runs the diff scripts (D19). Any developer or agent runs this in seconds before committing.
  2. **Optional pre-commit hook** — `.git/hooks/pre-commit` runs `make parity` on commits touching any of the three stacks. Hook is documented and provided as a sample (e.g., `tools/git-hooks/pre-commit.sample`); developers opt in by copying or symlinking it. Personal-discipline tool, not a forced gate.
  3. **README copy** — each stack's `README.md` documents: *"Before committing changes that touch routing, schema mapping, or HTMX target IDs, run `make parity` from repo root."*

  **CI graduates to scope when one of these triggers fires:**
  - The artifact is shared externally (talk, blog, public reference) — at that point external contributors can't be trusted to remember discipline.
  - A second contributor joins.
  - A drift bug escapes locally and ships to a stack-symmetry comparison the author cares about.

  This converts CI from "MVP infrastructure" to "post-thesis-validation epic when the audience changes." Honest scoping for a teaching artifact.

- **D19 — Cross-stack diff tooling location:** shell scripts under `tools/parity/` at repo root, callable from `make` (and from the optional pre-commit hook).
  ```
  tools/parity/dump-pg-indexes.sh        # connects to Postgres, dumps pg_indexes for domain.* sorted
  tools/parity/dump-routes-net.sh        # invokes `dotnet run --project FieldMark.Web -- --dump-routes`
  tools/parity/dump-routes-django.sh     # invokes `manage.py show_urls` (django-extensions) or custom command
  tools/parity/dump-routes-fiber.sh      # invokes `go run ./cmd/web -dump-routes`
  tools/parity/diff-routes.sh            # runs all three dumpers, normalizes casing, diffs, exits non-zero on diff
  tools/parity/diff-pg-indexes.sh        # dumps once per stack-mapped DbContext, asserts identical
  ```
  Each route-dump command is implemented in its own stack (one extra ~20-line Program.cs / management command / Go subcommand). The shell scripts are the cross-stack glue.

- **D20 — Local dev startup procedure:** top-level `Makefile` with stack-specific targets.
  ```makefile
  make up                # docker compose up -d
  make down              # docker compose down
  make reset             # docker compose down -v && docker compose up -d  (reseed Postgres)
  make run-net           # dotnet run --project FieldMark/FieldMark.Web
  make run-django        # cd fieldmark_py && uv run python manage.py runserver
  make run-go            # cd fieldmark-go && go run ./cmd/web
  make test-net          # dotnet test FieldMark/
  make test-django       # cd fieldmark_py && uv run pytest
  make test-go           # cd fieldmark-go && go test ./...
  make e2e               # cd e2e && pnpm test
  make parity            # tools/parity/diff-routes.sh && tools/parity/diff-pg-indexes.sh
  make css               # cd fieldmark_shared && npm run build
  ```
  Single source of "how do I run this" for newcomers and agents.

### NET-MAPSTER (resolved)

**Decision: REJECTED for MVP. Use manual projection.**

**Rationale:**
- Mapping volume is *low*. The domain has 4 aggregates plus reference data and audit. View models for HTMX partials are typically projected directly inside a LINQ query (`Select(p => new ProjectDetailVm { ... })`) — adding a mapper to project from `Project` to `ProjectDetailVm` doesn't save lines and adds a hop a reader has to follow.
- Source-readability is a stated quality attribute (PRD §Maintainability + the talk-audience purpose). Manual `new ViewModel { Foo = entity.Foo }` is the most readable mapping in any C# codebase.
- Cross-stack symmetry argument: Django and Go don't have an equivalent mapper. If .NET uses Mapster while the other two use direct projection, the "framework is the variable, architecture is the constant" thesis weakens — readers will wonder why .NET needs the abstraction.
- The view models are small. A `ProjectDetailVm` for HTMX rendering needs ~10 fields. The math doesn't favor a code-gen mapper at this volume.

**License-clean and zero-allocation are real Mapster strengths** — they're why we considered it. They aren't enough to overcome the readability and symmetry costs at this domain depth. AutoMapper remains forbidden (separate concern: licensing change + ROI).

**Reopen if:**
- A single endpoint accumulates more than ~3 distinct view-model shapes from the same entity.
- Duplication across LINQ projections becomes painful in code review.
- Entity → JSON DTO mapping for AG Grid endpoints starts producing the same projection in two handlers.

### Decision Impact Analysis

**Implementation sequence (highest leverage first):**

1. `docker/postgres/init/010_domain_tables.sql` — unblocks every stack's data layer.
2. `tools/parity/` scripts + Makefile — establishes the cross-stack diff contract before code drifts.
3. EF Core fluent config + `domain.*` mapping for Project (one aggregate as proof) — proves the mapping pattern.
4. Django `Meta.managed = False` model for Project — same proof on Django.
5. Go `ProjectStore` + pgx implementation — same proof on Go.
6. Anchor Workflow MVP epic — falsifies/confirms the smoothness target on at least one stack.

**Cross-component dependencies:**
- D9 (same-UUID seed) depends on D6 / D7 (auth-schema configs) being settled.
- D14 (AG Grid theming) depends on D15 (vendoring) so the theme file is local.
- D18 (CI) — explicitly skipped; D19 (parity tools) stands alone as the enforcement mechanism.
- The parity Makefile + optional pre-commit hook (D18 / D19 / D20) form a cohesive local-discipline triangle that does the same job CI would, with no infrastructure overhead.

## Implementation Patterns & Consistency Rules

### Pattern Categories Defined

**Conflict-point taxonomy.** AI agents working on FieldMark could diverge on ~30 dimensions — naming, layering, error rendering, audit-entry composition, partial-view contracts, AG Grid endpoint shape, transaction discipline, role-gated rendering, view model construction, test layout. The patterns below close those off explicitly.

**Two principles govern everything:**

1. **Canonical (cross-stack-identical):** routes, HTTP methods, target IDs, JSON wire format, audit action strings, entity method names, transaction discipline, error-response semantics. A diff across stacks is a defect.
2. **Idiomatic (per-stack):** code casing, file organization within a stack, framework-specific config, dependency injection style. Forcing identical idioms across C# / Python / Go is a worse defect than the cross-stack diff itself — the project's whole thesis is *the architecture is the constant; the framework is the variable*.

**Rule of thumb:** it's canonical if it's observable on the wire or in the database; it's idiomatic if it lives only in the source.

### Naming Patterns

**Database (canonical, all stacks identical):**

| Concern | Convention | Example |
|---|---|---|
| Schema | lower `snake_case` | `domain`, `dotnet_auth`, `django_auth`, `fiber_auth`, `infra` |
| Table | singular, `snake_case` | `domain.project`, `domain.violation`, `domain.audit_entry` |
| Column | `snake_case` | `compliance_score`, `started_at`, `inspector_id` |
| Foreign-key column | `<entity>_id` (no `fk_` prefix) | `project_id`, `inspector_id` |
| Index | `idx_<table>_<columns_or_purpose>` | `idx_violation_project_status`, `idx_violation_due` |
| Partial-index name | same convention; condition documented in DDL | `idx_violation_due` (`WHERE status IN ('Open','InProgress')`) |
| Constraint name | `<table>_<purpose>_<kind>` | `project_compliance_score_check`, `finding_spawned_violation_fk` |

Junction tables: `<entity_a>_<entity_b>` with composite PK — `domain.project_trade_scope`, `domain.project_inspector`. Never pluralize.

**JSON wire format (canonical):** `snake_case` field names. AG Grid endpoint contract example:

```json
{
  "rows": [
    {
      "id": "f9e4...",
      "code": "RIVERSIDE-01",
      "name": "Riverside Substation Upgrade",
      "status": "ACTIVE",
      "compliance_score": 71,
      "open_violation_count": 1
    }
  ],
  "last_row": 247
}
```

**Enum values (canonical, on wire and at rest):** `SCREAMING_SNAKE_CASE`. Examples: `OPEN`, `IN_PROGRESS`, `RESOLVED`, `VOIDED`, `COMPLIANCE_OFFICER`. Never lowercase, never PascalCase on the wire.

**Routes (canonical):** lowercase, kebab-case path segments, plural collection nouns, singular detail when scoped by id.

| Pattern | Example |
|---|---|
| Collection | `GET /projects` |
| Detail | `GET /projects/:id` |
| State change | `POST /projects/:id/close`, `POST /violations/:id/assign` |
| HTMX partial fetch | `GET /projects/:id/inspections` (returns `<div id="inspection-list">…</div>`) |
| AG Grid endpoint | `POST /grid/projects` |
| Auth | `/login`, `/logout`, `/account/...` |

Identical across stacks modulo language casing of route-binding syntax. `make parity` enforces.

**HTMX target IDs (canonical):** see Step 4 D11 for the full inventory. Adding a target ID is an ADR amendment.

**Audit action strings (canonical):** PascalCase, present-tense past-form action — `ProjectClosed`, `ProjectPlacedOnHold`, `ProjectResumed`, `InspectionStarted`, `InspectionCompleted`, `InspectionCancelled`, `ViolationOpened`, `ViolationAssigned`, `ViolationVoided`, `CorrectiveActionSubmitted`, `CorrectiveActionTakenForReview`, `CorrectiveActionApproved`, `CorrectiveActionRejected`. Stored verbatim in `domain.audit_entry.action`. Adding an action requires an ADR amendment.

**Code naming (idiomatic per stack):**

| Concern | .NET (C#) | Django (Python) | Go (Fiber) |
|---|---|---|---|
| Type | `PascalCase` | `PascalCase` | `PascalCase` (exported) |
| Method | `PascalCase` | `snake_case` | `PascalCase` (exported) / `camelCase` (unexported) |
| Field/property | `PascalCase` | `snake_case` | `PascalCase` (exported) |
| Local var | `camelCase` | `snake_case` | `camelCase` |
| File | `PascalCase.cs` | `snake_case.py` | `snake_case.go` |
| Constant | `PascalCase` | `SCREAMING_SNAKE_CASE` | `PascalCase` (exported) |

**Domain method names (canonical semantics; idiomatic casing):** `start / complete / cancel / place_on_hold / resume / close / assign / submit_corrective_action / approve_resolution / reject_resolution / void`. Casing translates per stack:

```
canonical                    .NET / Go              Django / Python
─────────                    ──────────              ───────────────
place_on_hold                PlaceOnHold             place_on_hold
submit_corrective_action     SubmitCorrectiveAction  submit_corrective_action
approve_resolution           ApproveResolution       approve_resolution
```

Adding a method outside this list is a defect; adding one *to* this list requires an ADR amendment.

### Structure Patterns

**Per-stack project layout (idiomatic, codified):**

```
.NET — FieldMark/
├── FieldMark.sln
├── Directory.Build.props                # solution-wide build hygiene (already committed)
├── FieldMark.Domain/                    # pure: entities, value objects, exceptions, state machines
│   ├── Entities/
│   ├── ValueObjects/
│   └── Exceptions/                      # DomainRuleException + subclasses
├── FieldMark.Data/                      # EF Core mapping ONLY (no business logic)
│   ├── Context/                         # FieldMarkDbContext, AuthDbContext (separate; dotnet_auth)
│   └── Configuration/                   # IEntityTypeConfiguration<T> per aggregate
├── FieldMark.Web/                       # Razor Pages, HTMX partials, handlers
│   ├── Pages/
│   │   ├── Shared/                      # _Layout, _ProjectDetail, _ViolationDetail, _ComplianceTile
│   │   ├── Projects/, Violations/       # page handlers
│   │   └── Grid/                        # AG Grid endpoint handlers
│   ├── Authorization/                   # ASP.NET Core policies + handlers
│   ├── Program.cs                       # composition root
│   └── wwwroot/                         # static assets (symlinked CSS + JS)
├── FieldMark.Tests.Domain/              # xUnit — state-machine and invariant tests
└── FieldMark.Tests.Integration/         # Testcontainers + real Postgres

Django — fieldmark_py/
├── manage.py, pyproject.toml, pytest.ini, mypy.ini
├── fieldmark/                           # project package: settings, urls, wsgi, asgi
├── projects/                            # one Django app per aggregate / functional area
│   ├── models.py                        # Project (Meta.managed=False, db_table='domain"."project')
│   ├── views.py                         # workflow handlers
│   ├── urls.py
│   ├── templates/projects/_project_detail.html
│   └── tests/
├── inspections/, violations/, compliance/, audit/, reference/, grid/
├── templates/                           # global templates (_layout, _compliance_tile)
├── static/                              # static assets (symlinked CSS + JS)
└── (auth)                               # built-in; tables in django_auth via DB router

Go — fieldmark-go/
├── go.mod, Makefile
├── cmd/web/main.go                      # entry: wire Fiber + pgxpool + templates, mount routes
├── cmd/seed/main.go                     # dev seed runner
├── internal/
│   ├── domain/                          # entities, state machines, invariants (pure)
│   │   ├── project.go, violation.go, …
│   │   └── errors.go                    # DomainRuleError types
│   ├── data/                            # explicit SQL, narrow Store interfaces, pgx
│   │   ├── projectstore.go, violationstore.go, …
│   │   └── tx.go                        # transaction helpers
│   ├── app/                             # thin coordinator: dependency wiring ONLY
│   │   ├── deps.go                      # Deps struct (DB pool, stores, etc.)
│   │   └── config.go                    # env-var parsing
│   └── web/                             # Fiber handlers, partials, AG Grid marshalling
│       ├── handlers/, templates/, auth/
│       └── routes.go
└── tools/                               # internal Go utilities (route dumper subcommand etc.)
```

**Test location:**
- .NET: separate `*.Tests.*` projects (xUnit). Domain tests in `FieldMark.Tests.Domain`; integration in `FieldMark.Tests.Integration`.
- Django: per-app `tests/` directory; pytest discovers via `pytest-django`. `pytest.mark.django_db` for transactional tests against real Postgres.
- Go: co-located `*_test.go`; integration tests build-tagged with `//go:build integration`.
- E2E: top-level `e2e/tests/` (Playwright + TypeScript), not per-stack.

### Format Patterns

**API response shape** — there is no "API" wrapper. FieldMark returns either:
1. **HTML partial** for HTMX requests (the partial has exactly one root element with a stable canonical id, e.g. `<div id="violation-detail">…</div>`).
2. **JSON `{rows, lastRow}`** for AG Grid endpoints (snake_case fields).

No `{data: …, error: …}` envelope. No pagination object. AG Grid manages pagination via `lastRow`. HTMX partials are unwrapped — the partial *is* the response body.

**Error response shape:**
- Domain rule violation → HTTP 409 + originating partial re-rendered with inline error message and unchanged entity state. The partial uses `aria-describedby` to associate the message with the relevant control.
- Authorization failure → HTTP 403 (or framework-equivalent) with no entity state in the response body.
- Validation error (form-level) → HTTP 422 + originating partial with field-level errors via `aria-invalid` and `aria-describedby`.
- Server error (uncaught) → HTTP 500 + framework's default error page in development; minimal generic page in any other environment.

**Date/time format:**
- Storage: `TIMESTAMPTZ` UTC.
- Wire (JSON): ISO 8601 with `Z` suffix — `"2026-05-09T14:23:01Z"`. Never local time on the wire.
- UI rendering: locale-default in the browser. Not the server's job.

**Boolean representation:** `true` / `false` (JSON), `BOOLEAN` (Postgres), native types in code. Never `1` / `0` in JSON; never `"yes"` / `"no"`.

**Null handling:** explicit `null` in JSON for optional fields (don't omit). Per-stack: C# nullable reference types are enforced; Python uses `Optional[T]`; Go uses pointer types or `pgtype.*` for nullable DB values.

### Communication Patterns

**No event bus, no message queue, no in-process eventing.** All cross-aggregate effects happen synchronously inside the same database transaction at the request handler. PRD §Forbidden Patterns explicitly lists in-process buses; codified here so an agent doesn't reach for `MediatR` "just for events."

**Audit entries are the only "events"** — they record what happened, in the same transaction as what happened, and are queryable via the audit log. Anything that needs to know "did X just happen?" reads `domain.audit_entry`, not a subscription.

**State management on the client: NONE.** PR-review anti-patterns:
- ❌ `Alpine.data()` storing business state.
- ❌ `localStorage` / `sessionStorage` writes for anything other than transient UI preferences (e.g., grid column widths).
- ❌ JavaScript variables holding "the current project" or "the user's role."
- ❌ AG Grid `getRowData()` used to compute filters or aggregates client-side.

The server is consulted for every view of state. HTMX swaps are the state synchronization mechanism.

### Process Patterns

**The Canonical Request Flow** — the most important pattern in this document; identical across all three stacks.

Every mutating handler does exactly this, in this order:

```
1. Authorize          — role check + ownership check where applicable
2. BEGIN              — open a transaction
3. Load aggregate     — fetch the aggregate root by id
4. Call entity method — domain logic; raises typed exception on rule violation
5. Append AuditEntry  — same transaction, opaque actor_id, action string, before/after JSON
6. Recompute score    — Project.recompute_compliance_score() if relevant
7. COMMIT
8. Render partial     — HTMX partial back to caller
```

Code stubs per stack (canonical handler — "Approve a corrective action"):

**.NET — `Pages/Violations/Detail.cshtml.cs`:**

```csharp
public async Task<IActionResult> OnPostApproveAsync(Guid id, Guid actionId)
{
    if (!_authz.Can(User, "violation.approve_resolution", id))
        return Forbid();                                                // 1. Authorize

    await using var tx = await _db.Database.BeginTransactionAsync();    // 2. BEGIN
    var violation = await _db.Violations
        .Include(v => v.CorrectiveActions)
        .FirstOrDefaultAsync(v => v.Id == id);
    if (violation is null) return NotFound();                           // 3. Load

    try
    {
        var action = violation.CorrectiveActions.Single(a => a.Id == actionId);
        violation.ApproveResolution(reviewer: CurrentUser, action);     // 4. Entity method
    }
    catch (DomainRuleException ex)
    {
        return new PartialViewResult                                    // 4b. Rule violation → 409 + partial
        {
            ViewName = "_ViolationDetail",
            ViewData = new ViewDataDictionary<ViolationDetailVm>(ViewData, ToVm(violation, error: ex.Message)),
            StatusCode = StatusCodes.Status409Conflict
        };
    }

    _db.AuditEntries.Add(new AuditEntry(                                // 5. Audit
        actor: CurrentUser.Id,
        action: "CorrectiveActionApproved",
        entityType: "Violation",
        entityId: id,
        projectId: violation.ProjectId,
        beforeState: snapshotBefore,
        afterState: snapshotAfter));

    var project = await _db.Projects.FindAsync(violation.ProjectId);
    project!.RecomputeComplianceScore();                                // 6. Score

    await _db.SaveChangesAsync();
    await tx.CommitAsync();                                             // 7. COMMIT

    return Partial("_ViolationDetail", ToVm(violation));                // 8. Render
}
```

**Django — `violations/views.py`:**

```python
@require_POST
def approve_corrective_action(request, violation_id, action_id):
    if not authz.can(request.user, "violation.approve_resolution", violation_id):
        return HttpResponseForbidden()                                  # 1. Authorize

    with transaction.atomic():                                          # 2. BEGIN ... 7. COMMIT (atomic)
        try:
            violation = Violation.objects.select_for_update().get(pk=violation_id)   # 3. Load
        except Violation.DoesNotExist:
            return HttpResponseNotFound()

        action = violation.corrective_actions.get(pk=action_id)
        try:
            violation.approve_resolution(reviewer=request.user, action=action)       # 4. Entity method
        except DomainRuleError as exc:                                  # 4b. Rule violation
            return render(request, "violations/_violation_detail.html",
                          {"violation": to_vm(violation, error=str(exc))},
                          status=409)

        AuditEntry.objects.create(                                      # 5. Audit
            actor_id=request.user.id, action="CorrectiveActionApproved",
            entity_type="Violation", entity_id=violation_id,
            project_id=violation.project_id,
            before_state=snapshot_before, after_state=snapshot_after,
        )

        violation.project.recompute_compliance_score()                  # 6. Score
        violation.project.save(update_fields=["compliance_score", "updated_at"])

    return render(request, "violations/_violation_detail.html",         # 8. Render
                  {"violation": to_vm(violation)})
```

**Go / Fiber — `internal/web/handlers/violations.go`:**

```go
func (h *ViolationHandlers) ApproveCorrectiveAction(c fiber.Ctx) error {
    violationID, _ := uuid.Parse(c.Params("id"))
    actionID, _ := uuid.Parse(c.Params("actionId"))

    actor := app.ActorFromCtx(c)
    if !h.authz.Can(actor, "violation.approve_resolution", violationID) {
        return c.SendStatus(fiber.StatusForbidden)                       // 1. Authorize
    }

    return h.deps.WithTx(c.Context(), func(tx pgx.Tx) error {            // 2. BEGIN ... 7. COMMIT
        violation, err := h.deps.Violations.LoadForUpdate(tx, violationID)   // 3. Load
        if err != nil { return err }

        action, _ := violation.FindCorrectiveAction(actionID)
        if err := violation.ApproveResolution(actor, action); err != nil {   // 4. Entity method
            var ruleErr *domain.RuleError
            if errors.As(err, &ruleErr) {                                // 4b. Rule violation
                vm := toVm(violation, ruleErr.Error())
                c.Status(fiber.StatusConflict)
                return c.Render("violations/_violation_detail", vm)
            }
            return err
        }

        if err := h.deps.AuditEntries.Append(tx, &domain.AuditEntry{     // 5. Audit
            ActorID:     actor.ID,
            Action:      "CorrectiveActionApproved",
            EntityType:  "Violation",
            EntityID:    violationID,
            ProjectID:   &violation.ProjectID,
            BeforeState: snapshotBefore,
            AfterState:  snapshotAfter,
        }); err != nil { return err }

        if err := h.deps.Projects.RecomputeScore(tx, violation.ProjectID); err != nil { return err }   // 6. Score

        return c.Render("violations/_violation_detail", toVm(violation, ""))   // 8. Render
    })
}
```

The shape is identical across stacks: same 8 steps, same order, same observable behavior. The only stack-level differences are dependency injection style and template syntax. **A handler doing anything outside these 8 steps is doing the wrong thing — the logic belongs on the entity.**

**Read-only (query) handler shape:**

1. Authorize (role check; row-level where applicable).
2. Load aggregate(s).
3. Project to view model.
4. Render partial.

No transaction needed for pure reads (Postgres default isolation is sufficient). No audit entry for reads (PRD doesn't require read auditing).

**AG Grid endpoint shape:**

1. Authorize (collection-level access).
2. Parse SSRM payload (start row, end row, sort, filter, group keys).
3. Translate to SQL with parameterized binding (each stack has a thin helper that converts SSRM payload → SQL `WHERE` / `ORDER BY`).
4. Query, return `{rows, lastRow}` JSON.

No business logic in the grid handler. The grid is a data tap; selection fires HTMX detail-panel loads via the row's `hx-get` attribute.

**Authorization expression pattern (canonical):**

Every "can the current user X" check goes through a single `IAuthorizationService` (or stack equivalent). The handler asks `authz.Can(user, action, entityId?)`. Implementation per stack — ASP.NET Core authorization handlers, Django decorators or a thin `authz.can()` function reading Django Groups, Go middleware reading `fiber_auth.user_role`. The handler never does role math itself.

**Action-button rendering pattern:** view models carry computed `can_*` booleans. Templates conditionally render based on those booleans. Three states explicitly:

```
absent   — user lacks permission (button is not in the rendered HTML at all)
disabled — user has permission but rule blocks (button is rendered with `disabled` attr + tooltip)
enabled  — user has permission and rule allows
```

Computed at render time, in the same handler that returns the partial. Never on the client.

### Enforcement Guidelines

**All AI agents working in this repo MUST:**

1. **Follow the Canonical Request Flow** for any mutating handler (8 steps, in order). Adding work outside those steps is a defect.
2. **Never invent an audit action string.** Use the canonical inventory; new strings require an ADR amendment.
3. **Map to existing `domain.*` tables only.** Never write a `CREATE TABLE` in a framework migration for `domain.*`. Schema additions go to `docker/postgres/init/` as a numbered SQL file.
4. **Render `can_*` booleans on the server.** Never put role logic in HTMX templates beyond a single conditional read of a precomputed value.
5. **Return HTTP 409 + originating partial** for domain rule violations. Never a JSON error body. Never an HTTP 200 with an embedded "success: false" flag.
6. **Use the canonical method-name list.** New transition methods require an ADR amendment.
7. **Honor stack idiomatic casing.** Don't apply `PascalCase` method names to Python or `snake_case` to C# in a quest for "consistency." Cross-stack consistency lives at the wire and DB layer; per-stack consistency lives at the source.
8. **Run `make parity` before committing** anything that touches routing, schema mapping, or HTMX target IDs. (Optional pre-commit hook automates this.)
9. **Reject pattern drift in PR review.** Architectural deviations are defects, not stylistic notes.

**Pattern-update process:** changes to canonical patterns require an ADR amendment recorded in this architecture document. Idiomatic per-stack patterns can evolve via PR + this document's update without ADR ceremony.

### Anti-Patterns (will fail review)

- ❌ Service classes that hold domain logic ("`ViolationService.ApproveResolution`" instead of `violation.ApproveResolution()`).
- ❌ Repository abstractions over EF Core / Django ORM / pgx.
- ❌ MediatR / in-process message buses.
- ❌ AutoMapper.
- ❌ Mapster (rejected at step-04; reopen conditions documented).
- ❌ Audit entries written outside the mutation transaction ("we'll log it after the commit").
- ❌ Compliance score recomputed by a background job ("eventually consistent score") — same-transaction or not at all.
- ❌ Client-side state stores (Redux, NgRx, Pinia, Zustand, Signals, Alpine `$store`).
- ❌ JSON API endpoints for HTMX (HTMX returns HTML; the only JSON endpoint pattern is AG Grid).
- ❌ Custom HTTP status codes ("499 means rule violation"). Use 409 / 403 / 422 per the standard.
- ❌ Generic exception handlers that swallow `DomainRuleException` — the partial-with-error rendering is the contract.
- ❌ Hand-rolled select grids when AG Grid would do — and conversely, AG Grid for simple list views that don't need server-side row model (use HTMX-rendered tables instead).

## Project Structure & Boundaries

### Complete Repository Directory Structure

```
fieldmark/                                          # repo root
├── README.md                                       # project overview, "how to run all three stacks"
├── CLAUDE.md                                       # observed-current-state architectural summary (cross-stack)
├── LICENSE
├── Makefile                                        # top-level orchestration (D20)
├── docker-compose.yml                              # Postgres 17 only; no per-stack containers
├── .gitignore
│
├── docker/
│   └── postgres/
│       └── init/
│           ├── 001_schemas.sql                     # CREATE SCHEMA domain, *_auth, infra; GRANT
│           ├── 010_domain_tables.sql               # all domain.* DDL (ADR-014 canonical)
│           ├── 020_domain_indexes.sql              # canonical index inventory
│           ├── 090_seed_reference.sql              # TradeType, ViolationCategory, ComplianceRule
│           └── seed-uuids/
│               └── dev-users.json                  # canonical user UUIDs (per-stack seeders consume)
│
├── tools/
│   ├── parity/                                     # cross-stack diff (D19)
│   │   ├── dump-pg-indexes.sh
│   │   ├── dump-routes-net.sh
│   │   ├── dump-routes-django.sh
│   │   ├── dump-routes-fiber.sh
│   │   ├── diff-routes.sh
│   │   └── diff-pg-indexes.sh
│   └── git-hooks/
│       └── pre-commit.sample                       # opt-in: runs `make parity`
│
├── fieldmark_shared/                                # shared front-end assets for all three stacks
│   ├── package.json
│   ├── src/
│   │   └── fieldmark.css                           # Tailwind v4 source; @source scans all three stacks
│   ├── dist/
│   │   └── fieldmark.css                           # COMMITTED build output (D16)
│   ├── vendor/                                     # locally vendored JS libs (D15); dir-symlinked into each stack
│   │   ├── htmx/
│   │   │   └── htmx.min.js
│   │   └── ag-grid/
│   │       └── 35.3.0/
│   │           └── ag-grid-enterprise.min.js
│   └── CLAUDE.md                                   # how to rebuild CSS; vendor symlink paths
│
├── e2e/                                            # cross-stack Playwright + axe-core
│   ├── package.json
│   ├── playwright.config.ts                        # 3 projects: net, django, fiber (parallel)
│   ├── biome.json
│   ├── tsconfig.json
│   ├── fixtures/
│   │   ├── seed-data.ts                            # reads dev-users.json + seeds domain rows for tests
│   │   └── auth-helpers.ts                         # logs in as a given role per-stack
│   ├── helpers/
│   │   ├── stack-config.ts                         # base URLs per project (.NET :4000, Django :8000, Fiber :3000)
│   │   ├── htmx-helpers.ts                         # waitForSwap, expectTargetSwapped
│   │   └── a11y.ts                                 # axe-core wrapper invoked in every test
│   ├── tests/
│   │   ├── anchor-resolve-violation.spec.ts        # the anchor demo workflow (Journey 1)
│   │   ├── corrective-action-rejection.spec.ts    # Journey 2
│   │   ├── project-closure-gate.spec.ts            # Journey 3
│   │   ├── executive-readonly.spec.ts              # Journey 4
│   │   ├── grid-row-selection.spec.ts              # AG Grid → HTMX detail panel
│   │   └── audit-trail.spec.ts                     # FR39–FR43 visibility
│   └── README.md
│
├── docs/
│   └── README.md                                   # links to architecture.md, PRD, ADRs
│
├── _bmad-output/
│   └── planning-artifacts/
│       ├── prd/                                    # canonical sharded PRD
│       │   └── ...
│       ├── architecture.md                         # this document
│       ├── prd-validation-report.md
│       └── research/                               # priming context only (not maintained)
│           └── ...
│
├── FieldMark/                                      # .NET 10 stack
│   ├── FieldMark.sln
│   ├── Directory.Build.props                       # solution-wide build hygiene
│   ├── README.md
│   ├── dotnet-tools.json
│   ├── FieldMark.Domain/
│   │   ├── FieldMark.Domain.csproj
│   │   ├── Entities/
│   │   │   ├── Project.cs                          # state machine, Close/PlaceOnHold/Resume + RecomputeComplianceScore
│   │   │   ├── JobSite.cs
│   │   │   ├── Inspection.cs                       # state machine, Start/Complete/Cancel
│   │   │   ├── Finding.cs                          # value object
│   │   │   ├── Violation.cs                        # state machine, Assign/ApproveResolution/Void
│   │   │   ├── CorrectiveAction.cs                 # TakeForReview/Approve/Reject
│   │   │   ├── AuditEntry.cs                       # write-once value object
│   │   │   └── Reference/                          # TradeType, ViolationCategory, ComplianceRule
│   │   ├── ValueObjects/
│   │   │   ├── ProjectStatus.cs                    # enum (Active/OnHold/Closed)
│   │   │   ├── ViolationStatus.cs, Severity.cs, ...
│   │   │   └── Role.cs                             # ADMIN/COMPLIANCE_OFFICER/INSPECTOR/SITE_SUPERVISOR/EXECUTIVE
│   │   └── Exceptions/
│   │       ├── DomainRuleException.cs              # base; never caught by global middleware
│   │       └── (subclasses: ClosureBlockedException, AlreadyResolvedException, ...)
│   │
│   ├── FieldMark.Data/
│   │   ├── FieldMark.Data.csproj
│   │   ├── Context/
│   │   │   ├── FieldMarkDbContext.cs               # maps ALL of domain.*; HasDefaultSchema("domain")
│   │   │   └── AuthDbContext.cs                    # ASP.NET Core Identity; HasDefaultSchema("dotnet_auth")
│   │   ├── Configuration/
│   │   │   ├── ProjectConfiguration.cs             # IEntityTypeConfiguration<Project> — ToTable("project","domain"), enum value converters
│   │   │   ├── ViolationConfiguration.cs
│   │   │   └── ... (one per aggregate + reference data)
│   │   └── Migrations/                             # auth migrations ONLY (dotnet_auth scope)
│   │       └── Auth/
│   │
│   ├── FieldMark.Web/
│   │   ├── FieldMark.Web.csproj
│   │   ├── Program.cs                              # composition root: AddDbContextPool, snake_case JSON, MapRazorPages
│   │   ├── appsettings.json                        # placeholder; FIELDMARK_DATABASE_URL via env
│   │   ├── appsettings.Development.json
│   │   ├── Authorization/
│   │   │   ├── DomainPolicies.cs                   # registers all "violation.approve_resolution" etc. policies
│   │   │   └── DomainAuthorizationHandler.cs       # native ASP.NET Core authorization
│   │   ├── Pages/
│   │   │   ├── Shared/
│   │   │   │   ├── _Layout.cshtml
│   │   │   │   ├── _ComplianceTile.cshtml          # OOB swap target
│   │   │   │   ├── _ProjectDetail.cshtml           # id="project-detail"
│   │   │   │   ├── _ProjectList.cshtml             # id="project-list"
│   │   │   │   ├── _ViolationDetail.cshtml         # id="violation-detail"
│   │   │   │   ├── _ViolationList.cshtml
│   │   │   │   ├── _InspectionList.cshtml
│   │   │   │   ├── _AuditLog.cshtml
│   │   │   │   ├── _CorrectiveActionForm.cshtml
│   │   │   │   ├── _CorrectiveActionList.cshtml
│   │   │   │   └── _FlashRegion.cshtml             # aria-live announcer
│   │   │   ├── Index.cshtml(.cs)                   # dashboard
│   │   │   ├── Projects/
│   │   │   │   ├── Index.cshtml(.cs)               # GET /projects
│   │   │   │   ├── Detail.cshtml(.cs)              # GET /projects/:id, POST close/place-on-hold/resume
│   │   │   │   └── ...
│   │   │   ├── Inspections/
│   │   │   ├── Violations/
│   │   │   │   └── Detail.cshtml(.cs)              # POST approve/reject corrective action
│   │   │   ├── Account/                            # framework-native auth UI (Identity)
│   │   │   └── Grid/                               # POST /grid/projects, /grid/violations, ...
│   │   │       ├── ProjectsGrid.cshtml.cs
│   │   │       ├── ViolationsGrid.cshtml.cs
│   │   │       └── ...
│   │   ├── ViewModels/
│   │   │   ├── ProjectDetailVm.cs                  # carries can_* booleans
│   │   │   ├── ViolationDetailVm.cs
│   │   │   └── ...
│   │   ├── SeedData/
│   │   │   └── DevUsers.cs                         # reads dev-users.json, seeds dotnet_auth
│   │   ├── Tools/
│   │   │   └── DumpRoutes.cs                       # `dotnet run -- --dump-routes` for parity tooling
│   │   └── wwwroot/
│   │       └── vendor/                             # all symlinks → ../../../../fieldmark_shared/…
│   │           ├── fieldmark.css                   # → fieldmark_shared/dist/fieldmark.css
│   │           ├── ag-grid                         # → fieldmark_shared/vendor/ag-grid  (dir symlink)
│   │           └── htmx                            # → fieldmark_shared/vendor/htmx     (dir symlink)
│   │
│   ├── FieldMark.Tests.Domain/                     # xUnit; pure state-machine and invariant tests
│   │   ├── ProjectTests.cs                         # tests Close/PlaceOnHold/Resume invariants
│   │   ├── ViolationTests.cs
│   │   ├── CorrectiveActionTests.cs
│   │   ├── ComplianceScoreTests.cs                 # algorithm correctness
│   │   └── ClosureGateTests.cs
│   │
│   └── FieldMark.Tests.Integration/                # Testcontainers + real Postgres
│       ├── Fixtures/
│       │   └── PostgresFixture.cs                  # spins up postgres:17 container with init scripts
│       ├── HandlerTests/
│       │   ├── ApproveCorrectiveActionTests.cs     # full request flow incl. audit + score
│       │   ├── CloseProjectTests.cs                # 409 + partial on closure-gate failure
│       │   └── ...
│       └── ParityTests/
│           └── PgIndexesTests.cs                   # asserts mapped indexes match DDL
│
├── fieldmark_py/                                   # Django 6.x stack (Python 3.14+)
│   ├── pyproject.toml
│   ├── uv.lock
│   ├── manage.py
│   ├── pytest.ini
│   ├── mypy.ini
│   ├── README.md
│   ├── main.py                                     # convenience entry; `uv run python main.py`
│   ├── fieldmark/                                  # project package
│   │   ├── __init__.py
│   │   ├── settings.py                             # DB router, INSTALLED_APPS, MIDDLEWARE
│   │   ├── urls.py                                 # composes per-app urls
│   │   ├── asgi.py, wsgi.py
│   │   ├── routers.py                              # DB router targeting auth tables to django_auth
│   │   ├── authz.py                                # `authz.can(user, action, entity_id)` — reads Groups
│   │   └── domain_db.py                            # connection wiring; SCHEMA-aware
│   │
│   ├── projects/                                   # one Django app per aggregate
│   │   ├── __init__.py
│   │   ├── apps.py
│   │   ├── models.py                               # class Project — Meta.managed=False, db_table='domain"."project'
│   │   ├── domain.py                               # state-transition methods (place_on_hold, close, ...)
│   │   ├── views.py                                # GET /projects, GET /projects/:id, POST close, ...
│   │   ├── urls.py
│   │   ├── view_models.py                          # to_vm(project, error=None) builds dict with can_* bools
│   │   ├── forms.py                                # form-level validation (NOT business rules)
│   │   ├── templates/projects/
│   │   │   ├── _project_detail.html                # id="project-detail"
│   │   │   ├── _project_list.html                  # id="project-list"
│   │   │   └── _compliance_tile.html               # OOB swap target (aria-live)
│   │   ├── management/commands/
│   │   │   └── seed_dev_users.py                   # reads dev-users.json, seeds django_auth
│   │   └── tests/
│   │       ├── test_project_state.py
│   │       └── test_close_handler.py               # @pytest.mark.django_db, hits real Postgres
│   │
│   ├── inspections/                                # same shape per app: models, domain, views, urls, view_models, templates, tests
│   ├── violations/
│   ├── compliance/                                 # rules engine + scoring; no domain entities of its own
│   │   ├── rules.py                                # OpenViolationGate, RequiredInspectionPerTrade, scoring weights
│   │   ├── scoring.py                              # recompute_compliance_score(project)
│   │   └── tests/
│   ├── audit/
│   │   ├── models.py                               # AuditEntry — Meta.managed=False, db_table='domain"."audit_entry'
│   │   └── append.py                               # `append(actor, action, entity, project_id, before, after)` helper
│   ├── reference/                                  # TradeType, ViolationCategory, ComplianceRule
│   ├── grid/                                       # AG Grid endpoint helpers
│   │   ├── ssrm.py                                 # parses SSRM payload → SQL where/order_by
│   │   ├── views.py                                # POST /grid/projects, /grid/violations, ...
│   │   └── tests/
│   │
│   ├── templates/                                  # global templates
│   │   ├── _base.html                              # layout
│   │   └── _flash_region.html                      # id="flash-region"
│   ├── static/                                     # symlinks to fieldmark_shared/vendor/
│   └── tools/
│       └── dump_routes.py                          # management command: emits route inventory
│
├── fieldmark-go/                                   # Go 1.26+ / Fiber v3 stack
│   ├── go.mod
│   ├── go.sum
│   ├── Makefile
│   ├── README.md
│   ├── cmd/
│   │   ├── web/
│   │   │   └── main.go                             # entry: parse env, build pgxpool, build template engine, mount routes
│   │   ├── seed/
│   │   │   └── main.go                             # reads dev-users.json (when fiber_auth lands)
│   │   └── tools/
│   │       └── dumproutes.go                       # `go run ./cmd/tools/dumproutes` emits route inventory
│   ├── internal/
│   │   ├── domain/                                 # PURE — no Fiber, no pgx, no framework
│   │   │   ├── project.go                          # Project struct + Close, PlaceOnHold, Resume, RecomputeComplianceScore
│   │   │   ├── inspection.go
│   │   │   ├── violation.go
│   │   │   ├── corrective_action.go
│   │   │   ├── audit_entry.go
│   │   │   ├── reference.go                        # TradeType, ViolationCategory, ComplianceRule
│   │   │   ├── role.go                             # role enum
│   │   │   ├── errors.go                           # *RuleError types
│   │   │   ├── compliance_rules.go                 # OpenViolationGate, RequiredInspectionPerTrade, scoring weights
│   │   │   └── ... (per-aggregate state-machine tests live next to source)
│   │   ├── data/                                   # explicit SQL via pgx; narrow Store interfaces
│   │   │   ├── tx.go                               # WithTx helper (begins, commits, rolls back)
│   │   │   ├── projectstore.go                     # type ProjectStore interface { Load, LoadForUpdate, Save, ... }
│   │   │   ├── violationstore.go
│   │   │   ├── correctiveactionstore.go
│   │   │   ├── auditentrystore.go                  # Append(tx, entry)
│   │   │   ├── inspectionstore.go
│   │   │   ├── referencestore.go
│   │   │   └── integration_test.go                 # //go:build integration — real Postgres
│   │   ├── app/                                    # THIN coordinator — wiring ONLY
│   │   │   ├── deps.go                             # Deps struct: DB pool, all Stores, Authz
│   │   │   ├── config.go                           # env-var parsing (FIELDMARK_DATABASE_URL, FIELDMARK_LOG_LEVEL)
│   │   │   └── actor.go                            # ActorFromCtx — extracts user id from request context
│   │   └── web/
│   │       ├── routes.go                           # central route registration (used by main.go AND dumproutes)
│   │       ├── handlers/
│   │       │   ├── projects.go                     # GET /projects, GET /projects/:id, POST close, ...
│   │       │   ├── violations.go                   # POST approve/reject corrective action
│   │       │   ├── inspections.go
│   │       │   ├── corrective_actions.go
│   │       │   └── grid.go                         # POST /grid/projects, /grid/violations, ...
│   │       ├── viewmodels/                         # per-aggregate view models with can_* booleans
│   │       │   ├── project.go, violation.go, ...
│   │       ├── auth/                               # MVP STUB middleware — injects fixed actor_id (ADR-012 deferral)
│   │       │   └── stub.go
│   │       ├── ssrm/                               # AG Grid SSRM payload → SQL where/order_by
│   │       │   └── parser.go
│   │       ├── templates/                          # html/template files
│   │       │   ├── _layout.html
│   │       │   ├── _compliance_tile.html
│   │       │   ├── projects/_project_detail.html
│   │       │   └── ...
│   │       └── static/                             # symlinks to fieldmark_shared/vendor/
│   └── tools/                                      # internal Go utilities (linting wrappers, etc.)
│
└── (no top-level CI directory — D18 deferred)
```

### Architectural Boundaries

**HTTP boundary (per stack):**
- Inbound HTTP → framework routing → handler/view/page-handler.
- `fiber.Ctx` (Go) does not escape `internal/web/`; converted to plain Go types at the handler boundary.
- `HttpContext` (.NET) and Django's `HttpRequest` similarly stay inside their respective Web layer.

**Domain boundary:**
- `FieldMark.Domain` (C#), each Django app's `domain.py` + `models.py` (Python), `internal/domain/` (Go) are *pure* — no framework imports, no DB libraries, no HTTP libraries.
- Domain code raises typed `DomainRuleException` / `DomainRuleError` / equivalent on rule violation. Handlers translate to HTTP 409 + originating partial.
- The full domain layer is covered by the unit-test project per stack (no DB, no HTTP — pure logic tests).

**Data-access boundary:**
- `FieldMark.Data` (.NET) — only EF Core fluent configuration and DbContext; no business logic. Single composition direction: `Web → Data → Domain`.
- Django models (`models.py` per app) — `Meta.managed = False` for shared domain tables; ORM is used directly from views (no Repository pattern).
- Go `internal/data/` — narrow `Store` interfaces (one per aggregate). All SQL parameterized; `*RowsAffected` errors propagated up.

**Authentication / authorization boundary:**
- Each stack owns its own `*_auth` schema; `domain.*` references users by opaque UUID.
- Conceptual roles (`ADMIN`, `COMPLIANCE_OFFICER`, `INSPECTOR`, `SITE_SUPERVISOR`, `EXECUTIVE`) implemented natively per stack.
- Single `authz.Can(user, action, entityId?)` (or stack-equivalent) call site; handlers never do role math.

**Frontend boundary:**
- Browser receives HTML. HTMX swaps regions identified by canonical target IDs.
- AG Grid is the only JS island. AG Grid receives `{rows, lastRow}` JSON; row selection fires HTMX detail-panel loads via `hx-get`.
- No client-side state. No service workers. No client-side routing.

**Cross-stack boundary:**
- The shared `domain.*` schema and the canonical wire formats (HTMX target IDs, AG Grid contract, audit action strings, JSON field naming) are the contracts. Anything observable is canonical; anything else is idiomatic per stack. Enforced via `make parity`.

### Requirements to Structure Mapping

**Functional Requirements → Directories**

| FR Category | .NET location | Django location | Go location |
|---|---|---|---|
| FR1–FR4: Authentication | `Pages/Account/`, ASP.NET Identity | `django.contrib.auth` + custom views | `internal/web/auth/` (stub for MVP) |
| FR5–FR8: Authorization | `Authorization/DomainPolicies.cs` | `fieldmark/authz.py` | `internal/app/` + `internal/web/auth/` |
| FR9–FR15: Project lifecycle | `Domain/Entities/Project.cs`, `Pages/Projects/` | `projects/domain.py`, `projects/views.py` | `internal/domain/project.go`, `internal/web/handlers/projects.go` |
| FR16–FR21: Inspection workflow | `Domain/Entities/Inspection.cs`, `Pages/Inspections/` | `inspections/` app | `internal/domain/inspection.go`, `internal/web/handlers/inspections.go` |
| FR22–FR27: Violation lifecycle | `Domain/Entities/Violation.cs`, `Pages/Violations/` | `violations/` app | `internal/domain/violation.go`, `internal/web/handlers/violations.go` |
| FR28–FR33: Corrective Action | `Domain/Entities/CorrectiveAction.cs`, `Pages/Violations/Detail.cshtml.cs` | `violations/` (CA is within Violation aggregate) | `internal/domain/corrective_action.go` |
| FR34–FR38: Compliance rules engine + scoring | `Domain/Entities/Project.RecomputeComplianceScore`, `Domain/Entities/Reference/ComplianceRule.cs` | `compliance/rules.py`, `compliance/scoring.py` | `internal/domain/compliance_rules.go` |
| FR39–FR43: Audit trail | `Domain/Entities/AuditEntry.cs`, `Data/Configuration/AuditEntryConfiguration.cs` | `audit/models.py`, `audit/append.py` | `internal/domain/audit_entry.go`, `internal/data/auditentrystore.go` |
| FR44–FR47: Dashboard | `Pages/Index.cshtml(.cs)` + `_ProjectList.cshtml` | `projects/views.py` (dashboard view) + `_project_list.html` | `internal/web/handlers/dashboard.go` |
| FR48–FR51: AG Grid integration | `Pages/Grid/*Grid.cshtml.cs` | `grid/views.py`, `grid/ssrm.py` | `internal/web/handlers/grid.go`, `internal/web/ssrm/parser.go` |
| FR52–FR53: Reference data | `Domain/Entities/Reference/`, `Data/Configuration/Reference*.cs` | `reference/` app | `internal/domain/reference.go`, `internal/data/referencestore.go` |
| FR54–FR59: Cross-cutting (POST-only state change, 409 + partial, identical contracts) | enforced in every handler in `Pages/` | enforced in every view across apps | enforced in every handler in `internal/web/handlers/` |
| FR60–FR64: Accessibility | partials' `aria-*` attrs across all `_*.cshtml`; `_FlashRegion.cshtml` for `aria-live` | partials' `aria-*` attrs across all `_*.html`; `_flash_region.html` | same; in `internal/web/templates/` |
| FR65: Playwright E2E | covered cross-stack in `e2e/tests/` | covered cross-stack in `e2e/tests/` | covered cross-stack in `e2e/tests/` |
| FR66: Domain unit tests | `FieldMark.Tests.Domain/` | per-app `tests/test_*_state.py` | `internal/domain/*_test.go` |
| FR67–FR70: Growth phase | not built in MVP — future epics | same | same |

**Cross-cutting concerns:**

- **Canonical request flow** — implemented in every mutating handler in every stack. No central abstraction; the pattern is the contract (Step 5).
- **Audit-on-every-mutation** — `AuditEntry` write is a step inside every mutating handler, in the same transaction.
- **Compliance score recomputation** — same: a step in mutating handlers, in the same transaction, where applicable.
- **Authorization** — single call site (`authz.Can`) per handler; implementation is stack-native.
- **Error rendering** — typed domain exception caught at the handler, partial re-rendered with HTTP 409.
- **Cross-stack parity** — enforced by `tools/parity/` shell scripts driven by `make parity`.

### Integration Points

**Internal communication:**
- All cross-aggregate effects happen synchronously inside the same DB transaction at the request handler. There is no internal RPC, no queue, no event bus.
- AG Grid → server: `POST /grid/<resource>` with SSRM payload; server returns `{rows, lastRow}` JSON.
- HTMX → server: `GET` (read partial) or `POST` (state change); server returns HTML partial.

**External integrations:**
- **PostgreSQL** — the only external dependency. Connected via `FIELDMARK_DATABASE_URL` env var. Each stack uses its native driver (Npgsql / psycopg / pgx).
- No third-party APIs, no SaaS integrations, no SMTP, no S3, no payment gateways. Per PRD §Non-Goals.

**Data flow (canonical mutating request — "Approve a corrective action"):**

```
Browser
  └─> POST /violations/:id/approve?actionId=:aid (HTMX hx-post)
         └─> [Stack-specific handler]
                ├─ 1. Authorize via authz.Can(user, "violation.approve_resolution", id)
                ├─ 2. BEGIN transaction
                ├─ 3. Load Violation aggregate (with CorrectiveActions)
                ├─ 4. violation.ApproveResolution(reviewer, action) — entity method
                │       └─ throws DomainRuleException → 4b. render partial w/ HTTP 409
                ├─ 5. Append AuditEntry { action: "CorrectiveActionApproved", before, after }
                ├─ 6. Project.RecomputeComplianceScore() — same transaction
                ├─ 7. COMMIT
                └─ 8. Render _ViolationDetail partial → response body
                          + (in same response) hx-swap-oob _ComplianceTile (project's new score)
                          + (in same response) hx-swap-oob _AuditLog (new audit row prepended)
Browser
  └─> HTMX swaps in:
         ├─ #violation-detail   — re-rendered partial with new status
         ├─ #compliance-tile    — OOB swap shows new score
         └─ #audit-log          — OOB swap shows new audit row
```

The "three things update in one round trip" anchor demo (PRD Journey 1) corresponds to OOB swaps for `#compliance-tile` and `#audit-log` returned alongside the primary `#violation-detail` partial in a single HTTP response.

### File Organization Patterns

**Configuration files:**
- Repo-root: `Makefile`, `docker-compose.yml`, `CLAUDE.md`, `README.md`.
- Per stack: framework-native config (`appsettings.json` / `settings.py` / env-via-`internal/app/config.go`).
- Env vars are the only source of secrets/config; no committed `.env` files.

**Source organization:**
- .NET: `Domain` (pure) → `Data` (mapping) → `Web` (handlers + UI). Solution-wide build hygiene in `Directory.Build.props`.
- Django: app-per-aggregate; each app contains `models / domain / views / view_models / urls / templates / tests`.
- Go: layered `cmd/web` → `internal/{domain, data, app, web}`; pure direction from `domain` outward.

**Test organization:**
- Domain tests adjacent to or in dedicated test projects per stack.
- Integration tests against real Postgres (Testcontainers / pytest-django / Go integration build tag).
- Cross-stack E2E tests in top-level `e2e/`.

**Asset organization:**
- All CSS authored once in `fieldmark_shared/src/`; compiled `dist/fieldmark.css` is committed; symlinked into each stack's `vendor/` static directory.
- All vendored JS (HTMX, AG Grid) lives in `fieldmark_shared/vendor/` with version-pinned subdirectories; directory-symlinked into each stack's `vendor/` static directory.
- No per-stack JS or CSS build pipelines. Tailwind compile is the only CSS build.

### Development Workflow Integration

**Development server structure:**
- `make up` — start Postgres (one container, init scripts run on first volume creation).
- `make run-net` / `make run-django` / `make run-go` — start one stack on its native port (.NET :4000, Django :8000, Fiber :3000).
- All three can run simultaneously; they share the same Postgres database (different `*_auth` schemas isolate identity).

**Build process structure:**
- .NET: `dotnet build FieldMark/` — runs analyzers, treats warnings as errors.
- Django: `uv sync` — restores deps; tests via `uv run pytest`.
- Go: `go build ./...` — builds all packages; tests via `go test ./...`.
- CSS: `cd fieldmark_shared && npm run build` — compiles Tailwind into committed `dist/`.

**Deployment structure:** explicitly out of scope (PRD §Non-Goals). The repo is local-development only. If/when deployment becomes a goal, it triggers a new ADR and a new architecture section.

## Architecture Validation Results

### Coherence Validation ✅

**Decision Compatibility** — all technology choices, ADRs, and patterns work together without contradiction. Spot checks:

- `EFCore.NamingConventions` (D1) + `domain-model.md` §9 `snake_case` canonical wire format → coherent. .NET emits the same column names Django reads natively.
- `Meta.managed = False` (D7 / Django patterns) + ADR-014 (infra-owned `domain.*`) → coherent. Django ORM consumes; never migrates.
- `pgx/v5` + explicit SQL (Go) + ADR-011 (no Repository) → coherent. Narrow `Store` interfaces are persistence boundaries, not abstractions.
- HTTP 409 + originating partial (D13) + HTMX swap semantics + `aria-describedby` for inline error (FR61) → coherent. Same partial re-renders with the error region populated; `aria-live` flash region announces the change.
- `make parity` (D18 / D19 / D20) + `pg_indexes` zero-diff (PRD §Success Criteria) + route-inventory zero-diff → coherent. The local-discipline triangle replaces CI for MVP.
- Vendored JS (D15) + version-pinning rule (PRD §Architectural Constraints) → coherent. Pinned files are auditable.
- Mapster rejection + ADR-011 (no AutoMapper, no fat service layers) + manual projection in LINQ/Django/Go → coherent. The .NET stack stays at the same source-density level as Django and Go.

**Pattern Consistency** — implementation patterns reinforce architectural decisions:

- The Canonical Request Flow (Step 5) is implementable in all three stacks with identical step ordering and observable behavior. Code stubs verified per stack.
- Domain method names use canonical semantics with idiomatic casing per stack — codified in the casing table in Step 5.
- Audit action strings are PascalCase verbatim across stacks (canonical wire format), not idiomatic per stack.
- Error rendering pattern (typed exception → 409 + partial) is identical in all three stack code stubs.
- `can_*` boolean rendering pattern is server-side only in all three stacks; templates do conditional rendering, not role math.

**Structure Alignment** — directory layouts support the chosen patterns:

- .NET 4-project solution honors ADR-011 (no Application/Service project; Web → Data → Domain).
- Django app-per-aggregate honors aggregate boundaries; cross-app imports limited to entity types and signals.
- Go `internal/{domain,data,app,web}` honors the one-way dependency direction; `app` is wiring-only as PRD §Forbidden Patterns specifies.
- E2E suite is top-level (not per-stack), matching the cross-stack-parity test discipline.

### Requirements Coverage Validation ✅

**Functional Requirements (70 FRs):** every FR maps to a specific architectural element. The §Project Structure → Requirements to Structure Mapping table is the explicit coverage matrix. Spot checks of harder cases:

- **FR6** (server decides whether to render an action button) → covered by the `can_*` boolean rendering pattern (Step 5) implemented in view models per stack.
- **FR15** (closure action: absent / disabled / enabled) → covered by the `can_close()` predicate on `Project` entity (`domain-model.md` §3.1) returning a tri-state surfaced on the view model.
- **FR20** (Fail-class findings auto-spawn Violations atomically) → covered by `Inspection.Complete()` calling `Violation` constructor inside the same transaction; verified by the Canonical Request Flow shape.
- **FR22a** (due_at immutable post-open) → covered by domain method on `Violation` setting `due_at` only at open-time (`domain-model.md` §3.7) with no setter exposed; defense-in-depth via DDL constraint optional.
- **FR32** (submitter ≠ reviewer for CorrectiveAction) → covered by entity-method invariant + `CHECK (submitted_by_id <> reviewed_by_id OR reviewed_by_id IS NULL)` in DDL.
- **FR48–FR51** (AG Grid endpoints) → covered by `/grid/<resource>` endpoint convention (D10), `internal/web/ssrm/parser.go` (Go) / `grid/ssrm.py` (Django) / `Pages/Grid/` (.NET) handlers, and `{rows, lastRow}` JSON contract.
- **FR55** (HTTP 409 + originating partial on rule violation) → covered by the typed-exception pattern in all three stack code stubs.
- **FR58–FR59** (cross-stack identical contracts) → covered by `tools/parity/diff-routes.sh` and `tools/parity/diff-pg-indexes.sh` (D19), invoked via `make parity`.
- **FR60–FR64** (accessibility: keyboard, ARIA, focus management on swaps, aria-live OOB, hx-disabled-elt) → covered by per-partial ARIA conventions; `_FlashRegion.cshtml` / `_flash_region.html` for non-OOB announcements; HTMX OOB swap convention restricted to header tiles with `aria-live`.
- **FR65** (Playwright cross-stack) → covered by `e2e/playwright.config.ts` 3-project parallel setup; same scenarios run against each stack.
- **FR66** (per-stack domain unit tests) → covered by `FieldMark.Tests.Domain/` (xUnit), per-app `tests/test_*_state.py` (pytest), `internal/domain/*_test.go` (go test).

**Non-Functional Requirements:** every binding NFR maps to a concrete enforcement mechanism.

- **Performance (200ms p95 / 300ms p95)** — enforced via the same-transaction Canonical Request Flow (no follow-up requests) and AG Grid SSRM (server-side row model, no client-side compute). Measurable via Playwright-recorded timings + axe-core scans.
- **Cross-stack symmetry** — enforced via `make parity` (D19).
- **Backend authority** — enforced via Domain layer purity (no framework imports), `can_*` server-rendered booleans, single `authz.Can` call site.
- **Schema ownership** — enforced via `docker/postgres/init/` being the only source of `domain.*` DDL; framework migrations scoped to `*_auth` only.
- **Testability (real Postgres)** — enforced via Testcontainers (.NET), pytest-django (Django), build-tagged integration tests (Go).
- **Auditability** — enforced via Step 5 of the Canonical Request Flow (audit-write in same transaction); reinforced by `domain-model.md` §3.10 append-only invariant.
- **Maintainability (build-blocking lint/format)** — enforced via `Directory.Build.props` (.NET), `pyproject.toml` ruff/black/mypy config (Django), `golangci-lint` (Go) — all already pinned in skeleton state.
- **Accessibility** — enforced via `@axe-core/playwright` in every E2E scenario; HTMX-specific patterns (focus management, `aria-live`, `hx-disabled-elt`) codified in Step 5.

### Implementation Readiness Validation ✅

**Decision completeness:** all D1–D20 decisions resolved and documented. NET-MAPSTER and CI-PIPELINE deferred decisions resolved with explicit reasoning and reopen criteria.

**Structure completeness:** complete directory tree defined down to file-level granularity for all three stacks plus shared infrastructure. Symlink relationships (CSS, vendor JS) explicit.

**Pattern completeness:** Canonical Request Flow has working code stubs in all three stacks. Naming patterns are tabulated for canonical (cross-stack) and idiomatic (per-stack) dimensions. Anti-patterns enumerated with explicit "will fail review" labeling.

### Gap Analysis Results

The architecture document is complete; the gaps below are **implementation gaps** (work-to-do) and **specification refinements** (worth tightening before story creation). None block the architecture's role as the source of truth.

**Critical Gaps:** 0

No architectural decisions are missing. Every FR has a defined home; every NFR has a defined enforcement mechanism.

**Important Gaps (worth resolving in early implementation epics):**

1. **`docker/postgres/init/010_domain_tables.sql` and `020_domain_indexes.sql`** — specified as the canonical schema source but not yet authored. The `research/domain-model.md` §8 DDL is the seed; authoring as the first implementation story is the highest-leverage move. Status: *specification complete; implementation pending.*
2. **`tools/parity/` shell scripts and per-stack route-dump subcommands** — specified (D19) but not authored. Without these, the cross-stack symmetry rule is aspirational. Should be the second implementation story (concurrent with the domain DDL).
3. **`Makefile` at repo root** — specified (D20) but not present. Single-command developer experience depends on it.
4. **`fieldmark_shared/vendor/` populated with HTMX and AG Grid 35.x** — ✅ resolved. `vendor/htmx/htmx.min.js` and `vendor/ag-grid/35.3.0/ag-grid-enterprise.min.js` (Enterprise bundle, includes Community) are committed; directory-symlinked into all three stacks' `vendor/` static directories.
5. **AuditEntry `before_state` / `after_state` JSON shape convention** — `domain-model.md` §3.10 specifies the columns are `jsonb` but doesn't specify what the snapshot includes. Recommendation: per-entity snapshot of *mutable* fields only (status, score, due_at, etc.) — not full entity state, not derived fields. Document this as a §6.1 addendum to the canonical request flow.
6. **HTMX OOB multi-partial response mechanism per stack** — Step 6's data-flow diagram shows three partials returned in one response (`#violation-detail` primary + `#compliance-tile` OOB + `#audit-log` OOB). The mechanism for composing this response differs per stack:
    - .NET: concatenated partial views with explicit `hx-swap-oob="true"` attributes on the OOB partials.
    - Django: a single composed template that includes the primary partial and the OOB partials inline.
    - Go: same as Django but using `html/template` composition.
    Should be codified as a "Composing OOB Responses" subsection in Implementation Patterns; right now it's implicit.
7. **`FIELDMARK_LOG_LEVEL` enumeration** — D17 mentions it without specifying allowed values. Recommendation: `debug`, `info`, `warn`, `error` (Go-style), each mapping to the framework-native level. Document this in the architecture doc.
8. **`dev-users.json` seed manifest schema** — D9 specifies the file exists but doesn't specify its shape. Recommendation: array of objects `{id: UUID, username, password, roles: [string]}`. Document this and commit a sample.
9. **`select_for_update()` discipline (Django, Go)** — the canonical request flow's Step 3 ("Load aggregate") should specify that mutating handlers use row-level locking to prevent lost updates under concurrent writes. The Django code stub uses `select_for_update()` but the .NET and Go stubs don't show their equivalents (`FOR UPDATE` in EF Core; `LoadForUpdate` is named in the Go stub). Should be codified.

**Nice-to-have Gaps (post-anchor-workflow):**

10. **Logging structure** — currently "framework-native HTTP request logging is sufficient." A future story could specify a structured-log convention if observability becomes a goal.
11. **Health-check endpoint** — currently out of scope (PRD Non-Goals). When deployment becomes a goal, add `/healthz`.
12. **Connection pool tuning per stack** — D3 specifies defaults; tuning to match real load is post-thesis.
13. **OpenTelemetry / distributed tracing** — explicitly out of scope per PRD. Vision-phase only.

### Validation Issues Addressed

The Important Gaps above are not "issues" per se — they are work items that fall naturally to early implementation stories. None are architectural gaps. They are flagged here so the BMad story-creation step (when reached) can sequence them.

The remaining 4 deferred items in the architecture document (NET-MAPSTER resolved, CI-PIPELINE deferred, Go-auth deferred per ADR-012, hosting/observability/secrets out of scope per PRD) all have explicit reopen criteria.

### Architecture Completeness Checklist

**Requirements Analysis**

- [x] Project context thoroughly analyzed
- [x] Scale and complexity assessed
- [x] Technical constraints identified
- [x] Cross-cutting concerns mapped

**Architectural Decisions**

- [x] Critical decisions documented with versions
- [x] Technology stack fully specified
- [x] Integration patterns defined
- [x] Performance considerations addressed

**Implementation Patterns**

- [x] Naming conventions established
- [x] Structure patterns defined
- [x] Communication patterns specified
- [x] Process patterns documented

**Project Structure**

- [x] Complete directory structure defined
- [x] Component boundaries established
- [x] Integration points mapped
- [x] Requirements to structure mapping complete

**16/16 checklist items confirmed.**

### Architecture Readiness Assessment

**Overall Status:** **READY FOR IMPLEMENTATION** (all 16 checklist items `[x]`; zero Critical Gaps; 9 Important Gaps that are work items, not architectural gaps).

**Confidence Level:** **High.** The PRD pre-resolved most architectural questions; this document codifies, fills gaps, and adds the per-stack pattern stubs that downstream agents need. The skeleton-state observations confirm that the architecture aligns with what's already been built.

**Key Strengths:**

- **Architectural binding lifted to PRD** at validation time (PRD §Architectural Constraints) means the architecture document is reinforcing rather than originating most decisions — agents working from either document will arrive at the same place.
- **Canonical Request Flow with working code stubs in all three stacks** is the single most important pattern; it's expressed in source-readable form rather than as prose.
- **Cross-stack vs. per-stack distinction codified** prevents the most common "consistency creep" failure mode where agents force PascalCase onto Python or snake_case onto C#.
- **Local-discipline triangle (Make + parity scripts + optional pre-commit hook)** replaces CI without losing the symmetry-enforcement function.
- **Comprehensive anti-patterns list** with "will fail review" labeling makes PR review unambiguous.
- **Gap surfacing at architecture time** rather than discovery at story-creation time — the 9 Important Gaps are sequenceable as the first implementation stories.

**Areas for Future Enhancement:**

- A "Composing OOB Responses" subsection in Implementation Patterns once the first multi-partial response is implemented in any stack — codify the actual mechanism used.
- An explicit "AuditEntry snapshot shape" subsection once the first audit write lands — codify the chosen field-set convention.
- A dedicated "Concurrency & locking" subsection if/when read-modify-write conflicts surface during implementation.
- CI workflow + cross-stack diff in CI when external-sharing triggers fire (D18 reopen criteria).

### Implementation Handoff

**AI Agent Guidelines:**

- Treat this architecture document and the canonical PRD (`_bmad-output/planning-artifacts/prd/`) as joint sources of truth.
- Follow the Canonical Request Flow exactly (8 steps, in order) for every mutating handler in every stack.
- Use the Requirements-to-Structure mapping table to locate where an FR's code lives.
- Check the Anti-Patterns list before introducing any abstraction (services, mappers, repositories, event buses, client state).
- Run `make parity` before committing any change to routing, schema mapping, or HTMX target IDs.
- Diverge from canonical-vs-idiomatic naming patterns only with explicit ADR amendment.
- When `domain-model.md` §8 DDL and this document agree on a column type or constraint, that's binding. When they disagree, prefer this document and update `domain-model.md` (or, if `domain-model.md` is in `research/`, update this document only).

**First Implementation Priorities (in order):**

1. **`docker/postgres/init/010_domain_tables.sql` + `020_domain_indexes.sql`** — author the canonical DDL based on `domain-model.md` §8. Unblocks every stack's data layer.
2. **`tools/parity/` scripts + repo-root `Makefile`** — establish the cross-stack diff contract before any drift can occur.
3. **`fieldmark_shared/vendor/` populated** — ✅ resolved. See gap item 4 above.
4. **One aggregate end-to-end in one stack (recommend Project + .NET, since the .NET skeleton is most complete)** — proves the full Canonical Request Flow works, including audit writes and compliance score recomputation.
5. **The same aggregate in the other two stacks** — proves cross-stack parity is achievable.
6. **Anchor Workflow MVP epic** — corrective-action approval with three-thing OOB swap. Falsifies or confirms the smoothness target on at least one stack.
