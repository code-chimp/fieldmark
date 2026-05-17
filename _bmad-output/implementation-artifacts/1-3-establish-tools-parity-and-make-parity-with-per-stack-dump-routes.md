# Story 1.3: Establish `tools/parity/` and `make parity` with per-stack `--dump-routes`

Status: done

## Story

As an agent or developer modifying any of the three stacks,
I want a single local command that detects cross-stack drift on routes and database indexes,
so that I catch divergence before it reaches code review — without depending on CI.

## Acceptance Criteria

1. **Given** the repo at HEAD **When** I inspect `tools/parity/` **Then** the directory contains executable scripts `dump-pg-indexes.sh`, `dump-routes-net.sh`, `dump-routes-django.sh`, `dump-routes-fiber.sh`, `diff-routes.sh`, `diff-pg-indexes.sh` (per Architecture D19).

2. **Given** each stack **When** I invoke its route-dump subcommand **Then** .NET responds to `dotnet run --project FieldMark/FieldMark.Web -- --dump-routes`, Django responds to `manage.py dump_routes` (custom management command), and Go responds to `go run ./cmd/web -dump-routes` **And** each command writes a normalized line-per-route list (`METHOD path`) to stdout, sorted, with all casing normalized to lowercase.

3. **Given** the database has been initialized and all three stacks are buildable **When** I run `make parity` from the repo root **Then** the script invokes `diff-routes.sh` (comparing all three route dumps) and `diff-pg-indexes.sh` (snapshotting `pg_indexes WHERE schemaname='domain'` against the canonical file) **And** both diffs exit `0` (clean).

4. **Given** I intentionally add a route to one stack and not the others **When** I run `make parity` **Then** the command exits non-zero and prints the diff identifying the divergent route.

5. **Given** `tools/git-hooks/pre-commit.sample` is committed **When** I read it **Then** it shows how to opt in to running `make parity` on commits touching any of `FieldMark/`, `fieldmark_py/`, `fieldmark-go/`, or `docker/postgres/init/`.

## Tasks / Subtasks

- [x] Task 1: Create per-stack route-dump subcommands (AC: #2)
  - [x] 1.1: .NET — `FieldMark.Web/Tools/DumpRoutes.cs` + arg parsing in `Program.cs`
  - [x] 1.2: Django — `fieldmark_py/tools/management/commands/dump_routes.py` management command
  - [x] 1.3: Go — `-dump-routes` flag on `fieldmark-go/cmd/web/main.go`
- [x] Task 2: Create `tools/parity/` shell scripts (AC: #1, #3, #4)
  - [x] 2.1: `dump-routes-net.sh`
  - [x] 2.2: `dump-routes-django.sh`
  - [x] 2.3: `dump-routes-fiber.sh`
  - [x] 2.4: `dump-pg-indexes.sh`
  - [x] 2.5: `diff-routes.sh`
  - [x] 2.6: `diff-pg-indexes.sh`
- [x] Task 3: Create `tools/git-hooks/pre-commit.sample` (AC: #5)
- [x] Task 4: Verify `make parity` exits 0 with all three stacks at HEAD
- [x] Task 5: Verify `make parity` exits non-zero after intentional drift

## Dev Notes

### Brownfield Posture

The root `Makefile` already has a wired `parity` target (lines 40–48) that looks for `tools/parity/diff-routes.sh` and `tools/parity/diff-pg-indexes.sh`. **Do not modify the Makefile** — it is already correct. Your job is to create the scripts it expects and the per-stack route-dump subcommands they invoke.

`tools/verify-domain-schema.sh` already exists from Story 1.2. The `tools/` directory is present but `tools/parity/` and `tools/git-hooks/` do not yet exist.

### Architectural Constraints

- **Architecture D19** — parity tooling lives in `tools/parity/` at repo root. Shell scripts only.
- **Architecture D18** — CI is deferred. These are local-discipline tools, not CI gates.
- **Architecture D20** — `make parity` orchestrates `diff-routes.sh && diff-pg-indexes.sh`.
- `set -euo pipefail` in every shell script (established pattern from `tools/verify-domain-schema.sh`).
- All scripts must be `chmod +x`.

### Route Dump Output Format

All three dump subcommands produce **identical normalized output**: one line per route, `METHOD /path`, sorted, all lowercase. Example:

```
get /
get /fragments/compliance-tile
get /privacy
```

Routes that are framework internals (e.g., Django's `/admin/`, .NET's `/Error`) are **excluded** from the dump — only application routes participate in parity.

Static-asset routes (`/static/**`, `_framework/**`) are also excluded.

### Per-Stack Route Dump Implementation

#### .NET — `FieldMark.Web/Tools/DumpRoutes.cs`

**Approach:** Check `args` for `--dump-routes` in `Program.cs` BEFORE `app.Run()`. If present, use `EndpointDataSource` to enumerate mapped endpoints, filter to Razor Pages (exclude Error), format, and write to stdout, then `Environment.Exit(0)`.

**Current state of `Program.cs`:** Lines 1–75 (read above). Insert the `--dump-routes` check after `app.MapRazorPages().WithStaticAssets()` but before `app.Run()`.

**Key implementation detail:** ASP.NET's `EndpointDataSource` from `app.Services.GetRequiredService<EndpointDataSource>()` enumerates all mapped endpoints. For Razor Pages, filter on `RouteEndpoint` with `PageActionDescriptor` or just filter by route pattern. Extract HTTP methods from `IHttpMethodMetadata` — pages respond to GET (and POST if they have an `OnPost` handler).

**Current .NET routes (from Pages/):** `/ (Index)`, `/Privacy`, `/Error` — exclude `/Error`.

#### Django — Management Command

**Location:** Create a Django app-level or project-level management command. Since this is a cross-app tool, put it in a new `tools` app: `fieldmark_py/tools/management/commands/dump_routes.py`.

**Approach:** The management command walks `urlpatterns` recursively, resolves all URL patterns, extracts the HTTP methods (default GET for standard views), formats, and writes to stdout. Use `django.urls.resolvers.get_resolver()` to get the root resolver.

**Must register `tools` in `INSTALLED_APPS`** in `fieldmark_py/fieldmark/settings.py`.

**Current Django routes:** Only `admin/` — which is excluded from parity. Output will be empty initially, which is correct (matching the other stacks' application routes only).

**Important:** Django 6.0 is in use (`pyproject.toml` / `settings.py`). The URL resolver API is stable.

#### Go — `fieldmark-go/cmd/tools/dumproutes/main.go`

**Approach:** Create a standalone tool that constructs the same Fiber app route registration as `cmd/web/main.go` (without starting the listener), then calls `app.GetRoutes()` to enumerate all routes.

**Alternative (simpler):** Add a `-dump-routes` flag to `cmd/web/main.go` itself. After route registration, if the flag is set, iterate `app.GetRoutes()`, filter out `/static/*` and middleware-only entries, format, print, and `os.Exit(0)`.

**Recommended: the flag approach on `cmd/web/main.go`** — this keeps route registration in one place and avoids duplication. The epics AC explicitly says `go run ./cmd/web -dump-routes`.

**Fiber v3 API:** `app.GetRoutes(filterUse ...bool)` returns `[]fiber.Route`. Each route has `Method` and `Path` fields. Call `app.GetRoutes(true)` to filter out middleware-only routes.

**Current Go routes:** `GET /` and `GET /fragments/compliance-tile` — plus `GET /static/*` (excluded).

### `diff-routes.sh` Implementation

1. Run all three dump scripts, capturing output to temp files.
2. Compare the three files pairwise (or diff all three against a single canonical).
3. If any diff is non-empty, print the diff and exit non-zero.
4. Clean up temp files on exit (trap).

### `diff-pg-indexes.sh` Implementation

1. Connect to Postgres using the same DSN pattern as `verify-domain-schema.sh`: `postgresql://fieldmark:fieldmark@localhost:5432/fieldmark`.
2. Query: `SELECT indexname, indexdef FROM pg_indexes WHERE schemaname = 'domain' ORDER BY indexname`.
3. Compare against a canonical snapshot file `tools/parity/canonical-pg-indexes.txt` (committed).
4. If diff, print and exit non-zero.

**Canonical indexes from `010_domain_tables.sql`:**
- `idx_audit_entity` — `domain.audit_entry(entity_type, entity_id)`
- `idx_audit_project` — `domain.audit_entry(project_id, created_at DESC)`
- `idx_inspection_project_status` — `domain.inspection(project_id, status)`
- `idx_violation_due` — `domain.violation(due_date) WHERE status IN ('OPEN','IN_PROGRESS')`
- `idx_violation_project_status` — `domain.violation(project_id, status)`

Plus any implicit primary key indexes created by PostgreSQL on each table.

**Important:** The canonical file should capture the FULL output of the query (both custom indexes and PK indexes) so drift in either direction is caught.

### `tools/git-hooks/pre-commit.sample`

A sample pre-commit hook that:
1. Checks if any staged files match `FieldMark/**`, `fieldmark_py/**`, `fieldmark-go/**`, or `docker/postgres/init/**`.
2. If matched, runs `make parity`.
3. If `make parity` fails, aborts the commit.

Document in the file header how to install: `cp tools/git-hooks/pre-commit.sample .git/hooks/pre-commit`.

### Source Tree — Files to Create

| File | Type | Description |
|------|------|-------------|
| `tools/parity/dump-routes-net.sh` | NEW | Invokes `dotnet run --project FieldMark/FieldMark.Web -- --dump-routes` |
| `tools/parity/dump-routes-django.sh` | NEW | Invokes `cd fieldmark_py && uv run python manage.py dump_routes` |
| `tools/parity/dump-routes-fiber.sh` | NEW | Invokes `cd fieldmark-go && go run ./cmd/web -dump-routes` |
| `tools/parity/dump-pg-indexes.sh` | NEW | Dumps `pg_indexes WHERE schemaname='domain'` |
| `tools/parity/diff-routes.sh` | NEW | Compares all three route dumps |
| `tools/parity/diff-pg-indexes.sh` | NEW | Compares live pg_indexes against canonical file |
| `tools/parity/canonical-pg-indexes.txt` | NEW | Committed snapshot of domain indexes |
| `tools/git-hooks/pre-commit.sample` | NEW | Opt-in parity check on commit |
| `FieldMark/FieldMark.Web/Tools/DumpRoutes.cs` | NEW | .NET route dump logic |
| `FieldMark/FieldMark.Web/Program.cs` | UPDATE | Add `--dump-routes` arg check before `app.Run()` |
| `fieldmark_py/tools/__init__.py` | NEW | Django `tools` app init |
| `fieldmark_py/tools/apps.py` | NEW | Django `tools` app config |
| `fieldmark_py/tools/management/__init__.py` | NEW | Management commands package |
| `fieldmark_py/tools/management/commands/__init__.py` | NEW | Commands package |
| `fieldmark_py/tools/management/commands/dump_routes.py` | NEW | Django route dump command |
| `fieldmark_py/fieldmark/settings.py` | UPDATE | Add `"tools"` to `INSTALLED_APPS` |
| `fieldmark-go/cmd/web/main.go` | UPDATE | Add `-dump-routes` flag handling |

### Files to READ Before Implementation

- `tools/verify-domain-schema.sh` — establishes shell script patterns (DSN, `set -euo pipefail`, check/report structure)
- `FieldMark/FieldMark.Web/Program.cs` — current .NET entry point; insert point for `--dump-routes`
- `fieldmark-go/cmd/web/main.go` — current Go entry point; insert point for `-dump-routes` flag
- `fieldmark_py/fieldmark/settings.py` — where to add `tools` to `INSTALLED_APPS`
- `fieldmark_py/fieldmark/urls.py` — current URL config (only `admin/`)
- `docker/postgres/init/010_domain_tables.sql` — index definitions for canonical snapshot

### Hard Rules

- **Do not modify the root Makefile.** The `parity` target is already wired correctly.
- **Do not modify `docker/postgres/init/` files.** They are infrastructure-owned and read-only for this story.
- **All shell scripts must start with `#!/usr/bin/env bash` and `set -euo pipefail`.**
- **Route dumps exclude framework internals** (Django admin, .NET Error page, static assets).
- **Output format is strict:** `method /path` per line, sorted, lowercase, no trailing slashes, no query params.
- **Stack symmetry:** all three stacks must produce the same route list for `make parity` to pass.

### Testing Strategy

1. Run `make parity` from repo root — must exit 0 with all three stacks at HEAD.
2. Manually add a test route to one stack (e.g., `app.Get("/test-drift", ...)` in Go), run `make parity` — must exit non-zero with a clear diff.
3. Remove the test route, run `make parity` again — must exit 0.
4. Run each dump script individually and verify the output format is correct.
5. Run `diff-pg-indexes.sh` alone and verify it matches the canonical snapshot.

### Previous Story Intelligence (Story 1.2)

- `tools/verify-domain-schema.sh` establishes the shell script conventions: `set -euo pipefail`, DSN string `postgresql://fieldmark:fieldmark@localhost:5432/fieldmark`, check/pass/fail reporting pattern.
- Story 1.2 created the first file in `tools/`. Story 1.3 adds `tools/parity/` and `tools/git-hooks/` as siblings.
- The dev agent for 1.2 used `psql -tAq` for quiet, tuple-only output — reuse this pattern in `dump-pg-indexes.sh`.
- Story 1.2's README updates documented `psql` from `libpq` (Homebrew) as a prerequisite — `dump-pg-indexes.sh` can rely on this being available.

### Git Intelligence

- Commit `a6fac88` (Story 1.1) created the root `Makefile` with the `parity` target placeholder.
- The `tools/` directory exists as untracked (Story 1.2 work not yet committed to this branch).
- The Go stack has a compiled `web` binary in the repo — this should be in `.gitignore` but is not this story's concern.

### Project Structure Notes

- `tools/parity/` is a peer of `tools/verify-domain-schema.sh` — flat structure under `tools/`.
- `tools/git-hooks/` is a new directory at the same level.
- The Django `tools` app follows the established pattern of Django apps in `fieldmark_py/` — it needs `__init__.py`, `apps.py`, and the `management/commands/` directory tree.
- The .NET `Tools/` directory under `FieldMark.Web/` follows the architecture doc's stated convention.
- Go's route dump uses a flag on the existing `cmd/web/main.go` entry point per the epics AC.

### References

- [Source: architecture.md §D18] — CI deferred; local discipline via `make parity`
- [Source: architecture.md §D19] — `tools/parity/` script inventory and per-stack dump commands
- [Source: architecture.md §D20] — Makefile target definitions
- [Source: epics.md §Story 1.3] — acceptance criteria and route-dump output contract
- [Source: root CLAUDE.md §Infrastructure] — database connection defaults
- [Source: FieldMark/CLAUDE.md] — .NET project structure, `Tools/DumpRoutes.cs` convention
- [Source: fieldmark_py/CLAUDE.md] — Django project structure, management command conventions
- [Source: fieldmark-go/CLAUDE.md] — Go project structure, `cmd/tools/dumproutes` convention

## Dev Agent Record

### Agent Model Used

claude-opus-4-6

### Completion Notes

- Root Makefile already had `parity` target wired — not modified.
- Canonical route set for walking skeleton baseline: `get /`, `get /fragments/compliance-tile`, `get /privacy`. All three stacks brought to identical output; this required adding stub routes/views to Django and a `/privacy` route to Go, and a `Pages/Fragments/ComplianceTile` Razor Page to .NET.
- Go route dump uses `-dump-routes` flag on `cmd/web/main.go` (not standalone tool), per Dev Notes override of task description.
- `dotnet run` emits "Using launch settings..." and "Building..." to stdout; `dump-routes-net.sh` filters to only `METHOD /path` lines via grep.
- Django's `dump_routes` command walks `get_resolver()` recursively, excludes `/admin`, normalizes trailing slashes.
- Canonical pg_indexes snapshot generated from live DB (21 indexes: 5 custom, 16 PK/unique).
- `make parity` exits 0 at HEAD; exits non-zero with a clear diff when `/test-drift` added to Go only (verified and reverted).
- All pre-existing tests pass: .NET (2 passed), Go (no test files), Django (0 collected — no existing tests).

**Review pass (2026-05-17, claude-sonnet-4-6):**
- ✅ Resolved [Decision]: Go DB connect moved after -dump-routes flag; `os.Exit(0)` replaced with `return`
- ✅ Resolved [Decision]: Removed blanket `2>/dev/null` from dump-routes-*.sh; real errors now surface on stderr
- ✅ Resolved [Decision]: Hardcoded GET deferred (correct — stacks are GET-only now)
- ✅ Resolved [Decision]: Executable bits confirmed 100755 in git index — no action needed
- ✅ Resolved [Patch]: Privacy.cshtml false positive — file already present from scaffold
- ✅ Resolved [Patch]: dump-pg-indexes.sh and diff-pg-indexes.sh now use `${FIELDMARK_DATABASE_URL:-...}` DSN fallback
- ✅ Resolved [Patch]: diff-routes.sh — removed redundant 3rd comparison; dropped unused vars; mktemp guarded
- ✅ Resolved [Patch]: Program.cs `Environment.Exit(0)` → `return` for proper host disposal
- ✅ Resolved [Patch]: ToolsConfig gained `default_auto_field = "django.db.models.BigAutoField"`
- ✅ Resolved [Patch]: diff-pg-indexes.sh mktemp guarded with explicit error message
- ✅ Bonus: ruff I001 fixed — merged split `from django.urls import` lines in dump_routes.py
- `make parity` still exits 0; .NET (2/2 passed), ruff clean, Go vet clean

### File List

- `tools/parity/dump-routes-net.sh` — new
- `tools/parity/dump-routes-django.sh` — new
- `tools/parity/dump-routes-fiber.sh` — new
- `tools/parity/dump-pg-indexes.sh` — new
- `tools/parity/diff-routes.sh` — new
- `tools/parity/diff-pg-indexes.sh` — new
- `tools/parity/canonical-pg-indexes.txt` — new
- `tools/git-hooks/pre-commit.sample` — new
- `FieldMark/FieldMark.Web/Tools/DumpRoutes.cs` — new
- `FieldMark/FieldMark.Web/Pages/Fragments/ComplianceTile.cshtml` — new (stub; brings .NET to route parity)
- `FieldMark/FieldMark.Web/Pages/Fragments/ComplianceTile.cshtml.cs` — new
- `FieldMark/FieldMark.Web/Program.cs` — update
- `fieldmark_py/tools/__init__.py` — new
- `fieldmark_py/tools/apps.py` — new
- `fieldmark_py/tools/management/__init__.py` — new
- `fieldmark_py/tools/management/commands/__init__.py` — new
- `fieldmark_py/tools/management/commands/dump_routes.py` — new
- `fieldmark_py/fieldmark/settings.py` — update
- `fieldmark_py/fieldmark/views.py` — new (home, privacy, compliance_tile stubs)
- `fieldmark_py/fieldmark/urls.py` — update
- `fieldmark_py/templates/pages/home.html` — new (stub)
- `fieldmark_py/templates/pages/privacy.html` — new (stub)
- `fieldmark_py/templates/fragments/compliance_tile.html` — new (stub)
- `fieldmark-go/cmd/web/main.go` — update
- `fieldmark-go/internal/web/templates/pages/privacy.html` — new (stub)

*(Review-pass additions — 2026-05-17)*
- `fieldmark-go/cmd/web/main.go` — update (DB connect moved after flag check; `os.Exit` → `return`)
- `FieldMark/FieldMark.Web/Program.cs` — update (`Environment.Exit(0)` → `return`)
- `tools/parity/dump-pg-indexes.sh` — update (FIELDMARK_DATABASE_URL fallback)
- `tools/parity/diff-pg-indexes.sh` — update (FIELDMARK_DATABASE_URL fallback + mktemp guard)
- `tools/parity/diff-routes.sh` — update (removed redundant 3rd comparison, mktemp guard, dropped unused vars)
- `tools/parity/dump-routes-net.sh` — update (removed 2>/dev/null)
- `tools/parity/dump-routes-django.sh` — update (removed 2>/dev/null)
- `tools/parity/dump-routes-fiber.sh` — update (removed 2>/dev/null)
- `fieldmark_py/tools/apps.py` — update (added default_auto_field)
- `fieldmark_py/tools/management/commands/dump_routes.py` — update (merged split import, ruff clean)

## Change Log

- 2026-05-17: Story implemented (claude-sonnet-4-6). [see original entry]
- 2026-05-17: Addressed code review findings (claude-sonnet-4-6). All 6 patches applied; 4 decisions resolved; 3 deferred confirmed. Bonus ruff I001 fix in dump_routes.py. `make parity` exits 0; all tests pass; ruff clean. Created `tools/parity/` with six shell scripts and canonical pg_indexes snapshot. Created per-stack route dump commands (--dump-routes on .NET, dump_routes management command on Django, -dump-routes flag on Go). Added Django `tools` app, home/privacy/compliance-tile views and templates. Added Go /privacy route and stub template. Added .NET ComplianceTile fragment Razor Page. All three stacks produce identical route dump (`get /`, `get /fragments/compliance-tile`, `get /privacy`). `make parity` exits 0 at HEAD, exits non-zero with intentional drift.

### Review Findings (2026-05-17 code review)

**decision-needed**
- [x] [Review][Decision] Go dump requires live DB — FIXED: DB connect moved after flag check; `os.Exit` replaced with `return`
- [x] [Review][Decision] All dump scripts suppress stderr with 2>/dev/null — FIXED: blanket suppression removed from all three dump-routes-*.sh wrappers; dotnet stdout noise already filtered by grep
- [x] [Review][Decision] Hardcoded GET for every route — DEFERRED: current stacks have GET-only routes; adding method awareness is out of scope for Story 1.3
- [x] [Review][Decision] Missing executable bit — FALSE POSITIVE: all scripts confirmed 100755 in git index; pre-commit.sample already executable

**patch**
- [x] [Review][Patch] .NET missing /privacy page — FALSE POSITIVE: Privacy.cshtml and Privacy.cshtml.cs already present from scaffold at Pages/Privacy.cshtml; route correctly emitted in dump
- [x] [Review][Patch] Dump-pg-indexes.sh hardcodes DSN — FIXED: uses `${FIELDMARK_DATABASE_URL:-postgresql://...}` fallback pattern
- [x] [Review][Patch] diff-routes.sh redundant pairwise compares — FIXED: removed third Django vs Fiber comparison (transitivity suffices); cleaned up unused NET_OUT/DJANGO_OUT/FIBER_OUT variables
- [x] [Review][Patch] Environment.Exit(0) in .NET bypasses host disposal — FIXED: replaced with `return` in Program.cs top-level statement
- [x] [Review][Patch] Django ToolsConfig missing default_auto_field — FIXED: added `default_auto_field = "django.db.models.BigAutoField"`
- [x] [Review][Patch] mktemp failure not guarded — FIXED: both diff-routes.sh and diff-pg-indexes.sh now guard mktemp with `|| { echo "ERROR: mktemp failed" >&2; exit 1; }`

**defer**
- [x] [Review][Defer] Django URL cycle detection and pattern None guard — pre-existing URLconf risk, not introduced here
- [x] [Review][Defer] uv / go / psql presence assumed on PATH — env prerequisite, not code defect
- [x] [Review][Defer] canonical-pg-indexes bootstrap requires manual first run — documented workflow, not a bug

**dismissed**: 4 (nits on sprint-status, minor style, unreachable code comments, system-check warnings)

**Also fixed during review pass:** ruff I001 — `dump_routes.py` had split `from django.urls import` lines; merged into single import.
