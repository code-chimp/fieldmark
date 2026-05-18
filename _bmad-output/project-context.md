---
project_name: 'FieldMark'
user_name: 'Tim'
date: '2026-05-18'
sections_completed: ['technology_stack', 'language_specific_rules', 'framework_specific_rules', 'testing_rules', 'code_quality_style_rules', 'development_workflow_rules', 'critical_dont_miss_rules']
status: 'complete'
rule_count: 35
optimized_for_llm: true
existing_patterns_found: 10
---

# Project Context for AI Agents

_This file contains critical rules and patterns that AI agents must follow when implementing code in this project. Focus on unobvious details that agents might otherwise miss._

---

## Technology Stack & Versions

**Core Technologies:**
- .NET (ASP.NET Core), Django (Python), Go (Fiber) — parallel HTMX + server-rendered stacks (identical routes, HTMX targets, AG Grid contracts, audit strings).
- PostgreSQL 17 — shared `domain` schema (infrastructure SQL inits in `docker/postgres/init/`) + isolated auth schemas (`django_auth`, `dotnet_auth`, `fiber_auth`).
- HTMX + AG Grid (server-side row model) + Tailwind (via `fieldmark_shared/` pnpm build, symlinked to each stack).

**Key Dependencies & Notes:**
- Makefile orchestration (`make up/reset/run-*/test-*`); domain inits run once on empty volume (use `make reset` after schema changes).
- No client-side state (no Redux/Zustand); AG Grid is isolated island.
- Real PostgreSQL only in tests (no SQLite).

## Critical Implementation Rules

### Language-Specific Rules

**Cross-Stack Conventions:**
- Wire/DB casing is always `snake_case`; code uses language idiom (e.g., PascalCase in C#, snake_case in Python/Go).
- Domain exceptions from entity methods → HTTP 409 (Conflict) + error partial; validation errors → HTTP 422 + form partial with `aria-invalid`.

**C# (.NET) Rules:**
- Use EF Core directly (no generic repos/UoW); DbContext scoped to `dotnet_auth` only for auth.
- Manual projection to view models (no AutoMapper).

**Python (Django) Rules:**
- Django ORM direct; auth schema only (`django_auth`); domain queries via raw SQL or ORM on `domain` tables.
- Use `uv run` for tests/scripts per stack CLAUDE.md.

**Go Rules:**
- Direct SQL or ORM on domain; Fiber handlers thin, call domain methods.
- Follow stack-specific CLAUDE.md for build/test.

### Framework-Specific Rules

**HTMX Patterns (All Stacks):**
- Partials: single root element with stable ID (must match exactly across .NET/Django/Go).
- State changes: `<button hx-post>`, never `<a>` links.
- `hx-swap-oob` only for header tiles (`#compliance-tile`).
- Server decides button presence/absence vs disabled.
- Domain exceptions → HTTP 409 + partial with error; validation → HTTP 422 + form + `aria-invalid`.

**AG Grid Rules:**
- Server-side row model only. Contract: `{ "rows": [...], "lastRow": N }`.
- Row selection triggers HTMX detail load (no business logic in grid config).

**Shared Assets:**
- `fieldmark_shared/` is source of truth for Tailwind (src → dist committed), AG Grid, HTMX.
- Symlinked into each stack's vendor/static; rebuild with `make css` (skips unless deps installed).

**Canonical HTMX Target IDs:**
`#project-detail`, `#project-list`, `#violation-detail`, `#violation-list`, `#inspection-list`, `#audit-log`, `#compliance-tile` (OOB), `#corrective-action-form`, `#corrective-action-list`, `#flash-region`.

### Testing Rules

**Environment Requirements:**
- Real PostgreSQL 17 only in all tests (never SQLite).
- Use `make reset` after domain schema changes to re-init volume.

**Execution:**
- Per-stack: `make test-net`, `make test-django`, `make test-go`.
- Cross-stack verification: `make parity` (route/index diffs), `make e2e` (skips unless e2e/ deps installed).
- CSS: `make css` (conditional on fieldmark_shared/ deps).

**Patterns:**
- Tests must cover domain state transitions and canonical request flow.
- AuditEntry writes validated in same transaction.

### Code Quality & Style Rules

**Workflow & Commands:**
- Always start with `make help`, `make up/reset` for DB, `./tools/verify-domain-schema.sh`.
- Root Makefile is executable truth; stack CLIs invoked only via `make run-*` / `make test-*`.

**Documentation & Comments:**
- Progressive disclosure: high-level in CLAUDE.md/AGENTS.md; deep details in docs/.
- Stack-specific rules in each CLAUDE.md (e.g., dotnet csharpier, EF migrations scoped to auth only).

**Naming & Organization:**
- Domain aggregates first (Project, Inspection, Violation, CorrectiveAction).
- File naming and structure per stack idiom, but maintain symmetry in public contracts.

### Development Workflow Rules

**Daily Commands:**
- `make up` to start Postgres (inits schemas on first run).
- `make reset` to destroy volume + re-init after any `docker/postgres/init/` changes.
- Run stacks in separate terminals: `make run-net` (:5000), `make run-django` (:8000), `make run-go` (:3000).

**Verification:**
- `./tools/verify-domain-schema.sh` (requires psql) to confirm domain.* tables.
- `make parity` for cross-stack diffs (routes, indexes).

**Setup Gotchas:**
- psql required (macOS: `brew install libpq && brew link --force libpq`).
- pnpm in `fieldmark_shared/` and `e2e/` (conditional Makefile skips).

### Critical Don't-Miss Rules

**Hard Rules (Non-Negotiable):**
- Backend authority only: domain rules, transitions, validation, authorization — server only.
- Infrastructure-owned `domain` schema: created by SQL init scripts; frameworks touch only their auth schemas. Never dual-own domain tables.
- No fat service layers: handlers call entity methods directly.
- No repository, Unit-of-Work, CQRS, MediatR, or in-process buses.
- No client-side state stores (no Redux, Zustand, etc.).
- No AutoMapper: project to view models manually.
- AuditEntry writes are mandatory and in the same transaction.
- Stack symmetry: routes, HTMX targets, AG Grid contracts, audit strings, domain method names identical across stacks. Divergence = defect.

**Anti-Patterns to Avoid:**
- Modifying `domain` schema from framework migrations.
- Adding client state or frontend frameworks.
- Using SQLite in any tests.
- Placing business logic in handlers or services instead of entities.
- Diverging public contracts (routes/IDs/audit actions) between stacks.

**Edge Cases & Gotchas:**
- Postgres init scripts run only on first container start (empty volume) — always `make reset` after init changes.
- `make parity` catches symmetry breaks post-Story 1.3.

---

## Usage Guidelines

**For AI Agents:**

- Read this file before implementing any code.
- Follow ALL rules exactly as documented.
- When in doubt, prefer the more restrictive option.
- Update this file if new patterns emerge.

**For Humans:**

- Keep this file lean and focused on agent needs.
- Update when technology stack changes.
- Review quarterly for outdated rules.
- Remove rules that become obvious over time.

Last Updated: 2026-05-18
