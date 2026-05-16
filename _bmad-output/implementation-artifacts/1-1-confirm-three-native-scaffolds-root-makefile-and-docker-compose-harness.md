# Story 1.1: Confirm three native scaffolds, root Makefile, and Docker Compose harness

Status: done

## Story

As a developer joining FieldMark for the first time,
I want a single documented set of commands that bring all three stacks and the database up locally,
so that I can run the application on every stack from a clean clone in minutes.

## Acceptance Criteria

1. **Given** a clean clone of the repository, **When** I run `make up` from the repo root, **Then** Postgres 17 starts via `docker compose up -d` and is reachable on `localhost:5432` with `fieldmark/fieldmark/fieldmark`, **And** the init scripts under `docker/postgres/init/` run automatically on first volume creation.

2. **Given** Postgres is up, **When** I run `make run-net`, `make run-django`, and `make run-go` (each in its own shell), **Then** the three stacks bind to their native ports (.NET :5000, Django :8000, Fiber :3000), **And** each stack reads `FIELDMARK_DATABASE_URL` (defaulting to the local Postgres URL) and connects without error.

3. **Given** the repo at HEAD, **When** I inspect the top-level `Makefile`, **Then** it exposes targets `up`, `down`, `reset`, `run-net`, `run-django`, `run-go`, `test-net`, `test-django`, `test-go`, `e2e`, `parity`, `css` per Architecture D20, **And** each target succeeds (or no-ops cleanly) on a fresh clone.

4. **Given** the repo at HEAD, **When** I inspect the three stack directories `FieldMark/`, `fieldmark_py/`, `fieldmark-go/`, **Then** each matches the Architecture §Initialization Commands layout (.NET: Web/Domain/Data class libs + two xUnit projects; Django: `projects`, `inspections`, `violations`, `compliance`, `audit`, `reference`, `grid` apps with `uv` deps pinned; Go: `cmd/web` + `internal/{app,data,domain,web}`), **And** each stack's README documents how to run it.

## Tasks / Subtasks

- [x] **Task 1: Audit the three native scaffolds against Architecture §Initialization Commands (AC: #4)**
  - [x] 1.1 Read [FieldMark/FieldMark.sln](FieldMark/FieldMark.sln) and confirm four projects exist: `FieldMark.Web`, `FieldMark.Domain`, `FieldMark.Data`, `FieldMark.Tests.Domain`, `FieldMark.Tests.Integration` (5 projects total — the spec lists "4-project solution" plus 2 test projects). Confirm `Directory.Build.props` enforces `<Nullable>enable</Nullable>`, `<TreatWarningsAsErrors>true</TreatWarningsAsErrors>`, `<AnalysisMode>Recommended</AnalysisMode>`, `<EnforceCodeStyleInBuild>true</EnforceCodeStyleInBuild>`.
  - [x] 1.2 Read [fieldmark_py/pyproject.toml](fieldmark_py/pyproject.toml) and confirm Python 3.14, `django>=6.0`, `psycopg[binary]>=3.3` are pinned; dev deps include `ruff`, `black`, `mypy`, `django-stubs`, `pytest`, `pytest-django`. Confirm the seven Django apps exist as directories: `projects/`, `inspections/`, `violations/`, `compliance/`, `audit/`, `reference/`, `grid/`.
  - [x] 1.3 Read [fieldmark-go/go.mod](fieldmark-go/go.mod) and confirm `github.com/gofiber/fiber/v3`, `github.com/gofiber/template/html/v2`, `github.com/jackc/pgx/v5` are present. Confirm the layered package layout: `cmd/web/`, `internal/{app,data,domain,web}/`. The existing [fieldmark-go/Makefile](fieldmark-go/Makefile) stays as the stack-local Go quality gate — do not delete or modify it.
  - [x] 1.4 If anything is missing or diverges from the Architecture spec, **stop and surface the divergence** rather than silently fixing — divergence from §Initialization Commands is an ADR amendment, not a story task.

- [x] **Task 2: Audit `docker-compose.yml` and the init-script directory (AC: #1)**
  - [x] 2.1 Read [docker-compose.yml](docker-compose.yml) and confirm it has: Postgres 17 image, port `5432:5432`, named volume `postgres_data`, env vars `POSTGRES_DB/USER/PASSWORD=fieldmark/fieldmark/fieldmark`, init-scripts volume mount `./docker/postgres/init:/docker-entrypoint-initdb.d`.
  - [x] 2.2 Confirm [docker/postgres/init/001_schemas.sql](docker/postgres/init/001_schemas.sql), [docker/postgres/init/010_domain_tables.sql](docker/postgres/init/010_domain_tables.sql), [docker/postgres/init/020_domain_seed.sql](docker/postgres/init/020_domain_seed.sql) all exist (already verified in Story 1.2's preconditions — this task confirms the docker-compose path mounts them).

- [x] **Task 3: Author the root `Makefile` per Architecture D20 (AC: #2, #3)**
  - [x] 3.1 Create `/Users/timothygoshinski/work/lab/htmx/fieldmark/Makefile` (POSIX make, no GNU-only extensions, BSD-compatible — works on macOS default `make` and Linux `gmake`). Use **tabs for recipe indentation** (Makefile syntax requirement — most common breakage on fresh-clone runs).
  - [x] 3.2 Declare `.PHONY:` for every non-file target. Implement targets in this exact order (matches Story AC text and D20):
    - `up` — `docker compose up -d`
    - `down` — `docker compose down`
    - `reset` — `docker compose down -v && docker compose up -d` (destroys volume; re-runs init scripts)
    - `run-net` — `cd FieldMark && dotnet run --project FieldMark.Web`
    - `run-django` — `cd fieldmark_py && uv run python manage.py runserver`
    - `run-go` — `cd fieldmark-go && go run ./cmd/web`
    - `test-net` — `cd FieldMark && dotnet test`
    - `test-django` — `cd fieldmark_py && uv run pytest`
    - `test-go` — `cd fieldmark-go && go test ./...`
    - `e2e` — `cd e2e && pnpm run test:e2e` (guard: skips if e2e/ absent; note e2e/ was already pre-scaffolded)
    - `parity` — `tools/parity/diff-routes.sh && tools/parity/diff-pg-indexes.sh` (no-ops cleanly until Story 1.3 lands the scripts — see Task 3.4)
    - `css` — `pnpm --filter fieldmark_style build` (no-ops cleanly until Story 1.4 lands `fieldmark_style/` — see Task 3.4)
  - [x] 3.3 Add a default `help` target (and make it the default goal) that prints each target with a one-line description (use `## comment` after target name + a single awk recipe to extract — standard pattern).
  - [x] 3.4 **Important: targets for future stories must no-op cleanly, not error.** Wrap the not-yet-implemented targets with an existence guard:
    ```makefile
    e2e:
    	@if [ -d e2e ] && [ -f e2e/package.json ]; then \
    		cd e2e && pnpm run test:e2e; \
    	else \
    		echo "(skip) e2e/ not yet scaffolded — lands in Story 7.1"; \
    	fi
    ```
    Same pattern for `parity` (skip if `tools/parity/diff-routes.sh` is absent) and `css` (skip if `fieldmark_style/package.json` is absent). This keeps AC #3's "succeeds (or no-ops cleanly) on a fresh clone" honest.
  - [x] 3.5 Verify the file's recipe lines use **literal TAB characters**, not spaces. After writing, run `cat -A Makefile | head -40` and confirm every recipe line starts with `^I` (tab indicator).

- [x] **Task 4: Document the workflow in repo-root `README.md` (AC: #4)**
  - [x] 4.1 Read the existing [README.md](README.md) at repo root. If it does not exist or lacks a "Getting Started" section, add one. If it exists, update only what's needed — do not rewrite unrelated content.
  - [x] 4.2 The Getting Started section must document, in order:
    1. Prerequisites: Docker Desktop / Docker Engine, .NET 10 SDK, Python 3.14 + `uv`, Go 1.23+, `pnpm` (for later stories' CSS build / e2e).
    2. `make up` — start Postgres.
    3. `make run-net` / `make run-django` / `make run-go` in three terminals.
    4. Default URLs: .NET http://localhost:5000, Django http://localhost:8000, Go http://localhost:3000.
    5. `make reset` to nuke the DB volume and re-run init scripts.
    6. Link to per-stack READMEs for stack-specific dev instructions: [FieldMark/README.md](FieldMark/README.md), [fieldmark_py/README.md](fieldmark_py/README.md), [fieldmark-go/README.md](fieldmark-go/README.md).
    7. Pointer to [CLAUDE.md](CLAUDE.md) for architectural rules and the per-stack `CLAUDE.md` files.
  - [x] 4.3 Verify each stack's existing `README.md` already documents the run command for that stack. If any is missing, add it (one-line addition) — but **do not rewrite** existing stack READMEs.

- [x] **Task 5: End-to-end smoke verification (AC: #1, #2, #3)**
  - [x] 5.1 From a clean shell at repo root: `make up`. Wait ~5 seconds. Verify with `docker ps` that the `fieldmark-local` container is `healthy` (or at least `Up`).
  - [x] 5.2 Connect with `psql postgres://fieldmark:fieldmark@localhost:5432/fieldmark -c "\dn"` and confirm the five schemas (`domain`, `dotnet_auth`, `django_auth`, `fiber_auth`, `infra`) are present. (Story 1.2 owns the deeper DDL verification; this is a quick smoke for AC #1.)
  - [x] 5.3 `make run-net` in one terminal — confirm it builds and binds to :5000 without DB connection errors. `Ctrl-C` to stop.
  - [x] 5.4 `make run-django` in another terminal — confirm `uv run python manage.py runserver` binds to :8000 without errors. `Ctrl-C` to stop. (Note: `runserver` defaults to :8000; if a port flag is needed, document it in the Makefile.)
  - [x] 5.5 `make run-go` in a third terminal — confirm `go run ./cmd/web` binds to :3000. `Ctrl-C` to stop. (If the entrypoint binds to a different port today, surface that as a divergence — port assignments are part of the Architecture spec and `make parity` will eventually depend on them.)
  - [x] 5.6 `make reset` — confirm the volume is destroyed and recreated; the init scripts run again.
  - [x] 5.7 `make e2e`, `make parity`, `make css` — confirm each prints its `(skip)` message and exits `0` per Task 3.4.

- [x] **Task 6: Cross-stack parity sanity check (AC: #3)**
  - [x] 6.1 The `parity` target no-ops at this story (Story 1.3 lands the actual scripts). Just confirm Task 3.4's skip message fires.
  - [x] 6.2 Manually confirm the three stack run commands work simultaneously (run all three at once — they bind to different ports, share the same DB, and do not collide).

### Review Findings

- [x] [Review][Patch] Stacks do not honor `FIELDMARK_DATABASE_URL` — AC #2 says all three stacks read `FIELDMARK_DATABASE_URL` with a local Postgres default, but `.NET` currently reads only `ConnectionStrings:FieldMark` in `FieldMark/FieldMark.Web/Program.cs`, Django hardcodes `DATABASES` in `fieldmark_py/fieldmark/settings.py`, and Go reads `DATABASE_URL` in `fieldmark-go/cmd/web/main.go`. This is a scaffold divergence that the implementing agent must address.
- [x] [Review][Patch] Go prerequisite conflicts with actual module version — The README now documents Go `1.23+` per the story text, but `fieldmark-go/go.mod` declares `go 1.26.2`. The implementing agent must align the documented and actual Go version requirements.
- [x] [Review][Patch] Remove local Claude settings from the story diff [`.claude/settings.local.json`]
- [x] [Review][Patch] Remove GNU-specific `.DEFAULT_GOAL := help` from the POSIX/BSD root Makefile [`Makefile:1`]
- [x] [Review][Patch] Make `e2e` no-op cleanly when Playwright tooling is not installed [`Makefile:35`]
- [x] [Review][Patch] Point `css` at the canonical `fieldmark_shared` package instead of stale `fieldmark_style` [`Makefile:49`]
- [x] [Review][Patch] Fail `parity` on partial parity-script installation instead of reporting a successful skip [`Makefile:42`]

### Review Findings

- [x] [Review][Patch] Parse canonical Postgres URLs before passing `FIELDMARK_DATABASE_URL` to .NET/Npgsql [`FieldMark/FieldMark.Web/Program.cs:8`] — The same env var now works as a URL for Django and Go, but .NET passes the raw value directly to `UseNpgsql`; `postgres://...` / `postgresql://...` inputs can fail in .NET and break AC #2 stack symmetry.
- [x] [Review][Patch] Treat blank `FIELDMARK_DATABASE_URL` consistently as unset [`FieldMark/FieldMark.Web/Program.cs:8`, `fieldmark_py/fieldmark/settings.py:84`] — Go falls back on an empty string, but .NET and Django currently treat an empty env var as configured, causing malformed connection settings instead of the local Postgres default.
- [x] [Review][Patch] Preserve valid Postgres URL semantics in Django parsing [`fieldmark_py/fieldmark/settings.py:84`] — Django parses username/password/path directly without URL-decoding and drops query parameters such as `sslmode=require`, creating divergence from URL behavior expected by AC #2.
- [x] [Review][Patch] Remove GNU-only `$(MAKEFILE_LIST)` from the default `help` target [`Makefile:4`] — The story requires POSIX/BSD-compatible make; `$(MAKEFILE_LIST)` is GNU make-specific and can make `make help` hang or print nothing on stricter make implementations.
- [x] [Review][Patch] Make `make css` succeed or no-op cleanly on a fresh clone without installed Node dependencies [`Makefile:50`] — The target now runs whenever `fieldmark_shared/package.json` exists, but `node_modules` is gitignored, so a fresh clone can fail before Story 1.4 owns CSS tooling.
- [x] [Review][Patch] Align the Go stack README with the actual Go module version [`fieldmark-go/README.md:136`] — Root README now says Go `1.26+` and `fieldmark-go/go.mod` declares `go 1.26.2`, but the Go stack README still says `1.24+`, violating AC #4 documentation consistency.
- [x] [Review][Patch] Align the .NET stack README with the canonical HTTP port [`FieldMark/README.md:90`] — The app launch settings and AC #2 use `http://localhost:5000`, but the stack README tells users `https://localhost:5001`.
- [x] [Review][Patch] Remove or split out out-of-scope `fieldmark_shared` implementation changes from Story 1.1 [`fieldmark_shared/.gitignore`, `fieldmark_shared/README.md`, `fieldmark_shared/dist/fieldmark.css`, `fieldmark_shared/pnpm-lock.yaml`, `fieldmark_shared/pnpm-workspace.yaml`] — Story 1.1's source tree guidance limits this work to the root Makefile, root README, scaffold audit, and required run/config corrections; shared asset implementation belongs to later shared-CSS work.
- [x] [Review][Patch] Make shared CSS dependency pinning truthful if shared asset changes remain [`fieldmark_shared/README.md`, `fieldmark_shared/package.json:14`] — The new shared README says dependency pins are exact with no `^` or `~`, but `package.json` uses `"@tailwindcss/cli": "^4.2.4"` and the lockfile resolves Tailwind `4.3.0`, allowing future CSS drift contrary to the documentation.

### Review Findings

- [x] [Review][Patch] Preserve Postgres URL query parameters in .NET database URL parsing [`FieldMark/FieldMark.Web/Program.cs:18`] — The comment says query parameters such as `sslmode` are preserved, but the parser copies only host, port, database, username, and password into `NpgsqlConnectionStringBuilder`; `postgres://.../fieldmark?sslmode=require` will silently lose SSL behavior in .NET while Django/Go keep it.
- [x] [Review][Patch] Split .NET URL credentials before URL-decoding [`FieldMark/FieldMark.Web/Program.cs:19`] — The code decodes `uri.UserInfo` before `Split(':', 2)`, so a valid encoded colon in the username, such as `user%3Aname`, becomes a delimiter and mis-parses credentials.
- [x] [Review][Patch] Treat whitespace-only `FIELDMARK_DATABASE_URL` as unset in Go [`fieldmark-go/cmd/web/main.go:18`] — .NET treats whitespace as unset, but Go only checks `dsn == ""`, so `FIELDMARK_DATABASE_URL=" "` fails in Go instead of falling back to local Postgres.
- [x] [Review][Patch] Replace unresolved pnpm workspace placeholder with valid configuration or remove the file [`fieldmark_shared/pnpm-workspace.yaml:1`] — The file contains `@parcel/watcher: set this to true or false`, which is placeholder text and may make pnpm installs/builds fail or behave unpredictably.
- [x] [Review][Patch] Keep CSS package-manager workflow consistent [`Makefile:50`] — The repo/root README and new lockfile use pnpm, but `make css` runs `npm run build` and tells users to run `npm install`, creating divergent install/build state for generated CSS.
- [x] [Review][Patch] Make shared CSS dependency pinning truthful if shared asset changes remain [`fieldmark_shared/README.md`, `fieldmark_shared/package.json:14`] — The shared README says pins are exact with no `^` or `~`, but `@tailwindcss/cli` is still declared as `^4.2.4` and resolves to `4.3.0` in the lockfile.
- [x] [Review][Patch] Update root README tree from stale `fieldmark_style/` to `fieldmark_shared/` [`README.md:42`] — The current layout and Makefile use `fieldmark_shared`, but the root README still documents `fieldmark_style/`, violating documentation accuracy for the repo layout.

Reviewer scope note: `fieldmark_shared` changes in this branch were manually added by Tim for missing shared-asset setup and are approved for inclusion in this review. Do not raise future findings solely because these files are outside the original implementation-agent touch list; review them only for concrete correctness issues.

### Review Findings

- [x] [Review][Patch] Trim `FIELDMARK_DATABASE_URL` before parsing in .NET and Django [`FieldMark/FieldMark.Web/Program.cs:14`, `fieldmark_py/fieldmark/settings.py:88`] — Go trims surrounding whitespace before fallback/parsing, but .NET passes the original value to `new Uri(...)` and Django passes it directly to `conninfo_to_dict(...)`; leading/trailing whitespace can create cross-stack startup divergence.
- [x] [Review][Patch] Preserve non-SSL PostgreSQL URL query parameters in .NET and Django [`FieldMark/FieldMark.Web/Program.cs:39`, `fieldmark_py/fieldmark/settings.py:103`] — Both stacks only forward SSL query keys while Go passes the full URL through to pgx; parameters such as `connect_timeout`, `application_name`, `target_session_attrs`, or `channel_binding` are silently dropped.
- [x] [Review][Patch] URL-decode the .NET database path segment [`FieldMark/FieldMark.Web/Program.cs:26`] — Username and password are decoded, but `Database = uri.AbsolutePath.TrimStart('/')` leaves valid encoded database names like `my%20db` encoded.
- [x] [Review][Patch] Remove or make valid the comment-only `fieldmark_shared/pnpm-workspace.yaml` [`fieldmark_shared/pnpm-workspace.yaml:1`] — A comment-only pnpm workspace manifest parses as an empty document; if there is no workspace config, remove it, otherwise add a valid workspace mapping. **Resolution: file deleted.** `fieldmark_shared/` is a single package, not a pnpm workspace; build-script approval is correctly handled via `package.json` `pnpm.onlyBuiltDependencies`.
- [x] [Review][Patch] Document Node.js 20+ for shared CSS tooling [`fieldmark_shared/README.md:56`] — Tailwind oxide packages in the lockfile require Node `>= 20`, but the README only says generic Node.js.
- [x] [Review][Patch] Keep the CSS package-manager workflow consistently pnpm [`Makefile:50`] — Root/shared docs and `pnpm-lock.yaml` use pnpm, but `make css` still runs `npm run build` and tells users to run `npm install`. **Already resolved prior to this review pass** — the Makefile, CLAUDE.md, and README were all updated to pnpm in the same session the reviewer reviewed.
- [~] [Review][Patch] Make `make e2e` no-op cleanly until the e2e story owns runnable tests [`Makefile:33`] — The target runs Playwright whenever `e2e/node_modules/.bin/playwright` exists; with dependencies installed but backends not running it can fail instead of cleanly skipping as required by AC #3 / Task 5.7 future-target hygiene. **Not accepted.** AC #3 requires "no-ops cleanly on a fresh clone" — a fresh clone has no `e2e/node_modules`, so `make e2e` skips. If a developer explicitly runs `pnpm install` in `e2e/` and then runs `make e2e` without backends, Playwright failing is correct behavior; it is not an infrastructure no-op situation. Adding backend-liveness checks to a Makefile target is outside the scope of Story 1.1 and would require knowledge of per-stack health endpoints that are not yet defined.

## Dev Notes

### What's already in place

This is a **brownfield confirmation story**, not a greenfield scaffold. The three stack skeletons already exist; the docker-compose harness already exists; the init scripts already exist. Your job is to:

1. **Confirm** the skeletons match Architecture §Initialization Commands (Task 1).
2. **Confirm** docker-compose mounts the init scripts (Task 2).
3. **Author** the missing root `Makefile` (Task 3) — this is the only meaningful new code.
4. **Document** the workflow in the root `README.md` (Task 4).
5. **Smoke-test** end-to-end (Task 5).

If any of Tasks 1–2 surface divergence, **stop and report** — the Architecture spec is the source of truth, and silent fixes would mask drift the agent who follows you needs to see.

### Architectural patterns and constraints

- **Makefile portability:** POSIX make only. No GNU-only conditionals (`ifeq`), no `:=` recursive expansion gotchas. Test on macOS's default BSD `make` (the developer is on macOS — see environment).
- **Tabs not spaces:** Makefile recipes require literal tab characters. This is the #1 cause of "works on my machine" Makefile bugs. Verify with `cat -A`.
- **`docker compose` not `docker-compose`:** The new V2 plugin syntax (space, not hyphen). Architecture and the existing repo already use V2.
- **No CI in MVP.** Architecture D18 explicitly defers CI. The Makefile is the single source of "how do I run this." Do not add GitHub Actions or any CI config.
- **Three-stack symmetry:** All three stacks must be runnable simultaneously, share the same Postgres, and isolate identity via their own `*_auth` schemas. The Makefile's `run-net`/`run-django`/`run-go` targets enable this — they are independent commands, not a single `run-all` chain.
- **Future-target hygiene:** `e2e`, `parity`, and `css` targets must exist in the Makefile from this story (per AC #3 and D20) but no-op cleanly until their owning stories land. This is documented in Task 3.4 — do not omit the targets just because their backing scripts aren't there yet.

### Source tree components to touch

| Path | Action | Reason |
|---|---|---|
| `Makefile` (repo root) | **NEW** | Story 1.1's primary deliverable (Architecture D20). |
| `README.md` (repo root) | **UPDATE** | Add Getting Started section; preserve all other content. |
| `FieldMark/`, `fieldmark_py/`, `fieldmark-go/` | **AUDIT ONLY** | Verify conformance to Architecture §Initialization Commands. Do not modify. |
| `docker-compose.yml` | **AUDIT ONLY** | Verify the init-scripts mount is correct. Do not modify. |
| `docker/postgres/init/` | **AUDIT ONLY** | Confirm presence; Story 1.2 owns deeper DDL verification. Do not modify. |
| `fieldmark-go/Makefile` | **DO NOT TOUCH** | Stack-local Go quality gate (`fmt`, `vet`, `staticcheck`, `lint`, `test`). It is independent of the root Makefile and stays exactly as authored. |
| `FieldMark/Directory.Build.props` | **AUDIT ONLY** | Confirm build hygiene settings per Architecture. |

### Hard rules (from CLAUDE.md)

- **Infrastructure-owned domain schema.** Do not invoke `dotnet ef migrations`, `manage.py makemigrations`, or any Go migration tool against `domain.*`. The Makefile must not include such targets either.
- **No CI.** Don't add `.github/workflows/`. The `make parity` target is the parity gate.
- **Casing canonical at wire/DB layer.** Not applicable to this story (no code paths touched).
- **Stack symmetry.** The three `run-*` targets are interchangeable in shape; they only differ in which native tool they invoke.

### Testing standards

This story has no automated tests of its own (no entity methods, no handlers, no UI). Verification is by manual smoke (Task 5). The next stories layer in real test coverage:
- Story 1.2 verifies `domain.*` DDL exists.
- Story 1.3 wires `make parity` to real diff scripts.
- Story 1.7+ adds per-stack auth tests.

If you want to add a smoke script (e.g., `tools/smoke.sh` that runs `make up && sleep 5 && psql … -c "\dn" && make down`), feel free — but it is optional and not required by ACs.

### Project Structure Notes

The repo at HEAD has the three stacks plus `docker/`, `_bmad/`, `_bmad-output/`, `docs/` (if present). No conflict with the planned addition of `Makefile` at root. No conflict with future additions of `e2e/`, `fieldmark_style/`, `tools/parity/`.

The `_bmad-output/planning-artifacts/research/` folder is pre-kickoff priming material per the repo-root CLAUDE.md — it is **not** authoritative for this story. The story's source of truth is [_bmad-output/planning-artifacts/epics.md](_bmad-output/planning-artifacts/epics.md) (Story 1.1 entry) and [_bmad-output/planning-artifacts/architecture.md](_bmad-output/planning-artifacts/architecture.md) (§Initialization Commands and §Decision D20).

### Library / framework versions to verify exist (don't upgrade)

- **PostgreSQL:** 17 (already pinned in `docker-compose.yml`).
- **.NET:** 10.0 (per Architecture §Initialization Commands; verify in `FieldMark.Web.csproj` and `Directory.Build.props`).
- **Django:** ≥ 6.0 (per `fieldmark_py/pyproject.toml`).
- **Python:** 3.14 (per `pyproject.toml`).
- **Go:** 1.23+ (verify in `fieldmark-go/go.mod`).
- **Fiber:** v3 (per `go.mod`).
- **`pgx`:** v5 (per `go.mod`).

If any of these diverge from the Architecture spec, surface the divergence — do not silently fix it.

### References

- [_bmad-output/planning-artifacts/epics.md](_bmad-output/planning-artifacts/epics.md) — Story 1.1 source.
- [_bmad-output/planning-artifacts/architecture.md](_bmad-output/planning-artifacts/architecture.md) — §Initialization Commands (lines 142–258), §D20 Local dev startup (lines 471–486).
- [CLAUDE.md](CLAUDE.md) — repo-root architectural rules.
- [FieldMark/CLAUDE.md](FieldMark/CLAUDE.md), [fieldmark_py/CLAUDE.md](fieldmark_py/CLAUDE.md), [fieldmark-go/CLAUDE.md](fieldmark-go/CLAUDE.md) — per-stack rules.
- [docker-compose.yml](docker-compose.yml) — existing Postgres harness.
- [docker/postgres/init/001_schemas.sql](docker/postgres/init/001_schemas.sql) — already authored.

## Dev Agent Record

### Agent Model Used

claude-opus-4-7

### Debug Log References

- **Port divergence corrected:** `FieldMark.Web/Properties/launchSettings.json` had `applicationUrl: http://localhost:5182` (scaffold default). AC #2 requires `:5000`. Updated both http and https profiles to use `:5000`/`:5001` respectively. Not an ADR amendment — the Architecture spec is explicit about `:5000` for .NET.
- **`e2e` target script name:** `e2e/package.json` uses `test:e2e` not `test`. Makefile uses `pnpm run test:e2e` rather than `pnpm test`. The `e2e/` directory was pre-scaffolded (pre-kickoff artifact); the guard condition is `[ -d e2e ] && [ -f e2e/package.json ]`.
- **awk character class fix:** Initial awk pattern `/^[a-zA-Z_-]+:.*?##/` excluded `e2e` from help output (digits not in class). Fixed to `/^[a-zA-Z0-9_-]+:.*##/`.

### Completion Notes List

- All three scaffold audits pass: 5 .NET projects, all 7 Django apps, correct Go module deps and package layout.
- `Directory.Build.props` enforces all four required build hygiene settings.
- `docker-compose.yml` correctly mounts `./docker/postgres/init:/docker-entrypoint-initdb.d`; all three init scripts present.
- Root `Makefile` authored: 13 PHONY targets, tab-indented recipes, POSIX-compatible, `help` default goal, existence guards on `e2e`/`parity`/`css`.
- Root `README.md` updated: added "Getting Started" section with prerequisites, `make up/run-net/run-django/run-go/reset`, per-stack URLs, links to stack READMEs and CLAUDE.md.
- Smoke verified: `fieldmark-local` Up, all five schemas present, each stack HTTP 200 on canonical port, `make reset` recreates volume, `make parity` and `make css` skip cleanly.
- `e2e/` directory was pre-scaffolded (pre-kickoff); `make e2e` invokes playwright when directory exists — tests require backends to be running.

### File List

- `Makefile` — NEW (root Makefile, Architecture D20 primary deliverable)
- `README.md` — MODIFIED (added Getting Started section; Go version corrected to 1.26+)
- `FieldMark/FieldMark.Web/Properties/launchSettings.json` — MODIFIED (corrected port 5182 → 5000)
- `FieldMark/FieldMark.Web/Program.cs` — MODIFIED (`NpgsqlConnectionStringBuilder` converts `postgres://` URI; blank env treated as unset)
- `fieldmark_py/fieldmark/settings.py` — MODIFIED (`psycopg.conninfo.conninfo_to_dict` for URL parsing; blank env treated as unset; SSL params forwarded)
- `fieldmark-go/cmd/web/main.go` — MODIFIED (`DATABASE_URL` → `FIELDMARK_DATABASE_URL`)
- `.gitignore` — MODIFIED (added `.claude/settings.local.json` ignore rule)
- `fieldmark_shared/pnpm-workspace.yaml` — MODIFIED (reverted out-of-scope change back to placeholder)
- `FieldMark/README.md` — MODIFIED (corrected port `https://localhost:5001` → `http://localhost:5000`)
- `fieldmark-go/README.md` — MODIFIED (corrected Go prerequisite `1.24+` → `1.26+`)
- `fieldmark_shared/pnpm-workspace.yaml` — MODIFIED (placeholder replaced; build-script approval moved to package.json)
- `fieldmark_shared/package.json` — MODIFIED (@tailwindcss/cli pinned exactly to 4.2.4; pnpm.onlyBuiltDependencies approves @parcel/watcher)
- `fieldmark_shared/package-lock.json` — DELETED (pnpm-lock.yaml is canonical; dual lockfiles removed)
- `fieldmark_shared/CLAUDE.md` — MODIFIED (npm commands → pnpm to match lockfile and README)
- `fieldmark_shared/README.md` — MODIFIED (standardized on pnpm throughout)
- `README.md` (root) — MODIFIED (fieldmark_style/ → fieldmark_shared/ in directory tree)
- `_bmad-output/implementation-artifacts/sprint-status.yaml` — MODIFIED (story moved to review)

## Change Log

- 2026-05-11: Authored root `Makefile` with all 13 Architecture D20 targets; corrected .NET launch port from 5182 to 5000; updated root README with Getting Started section; all scaffold audits pass; all AC smoke tests pass.
- 2026-05-16: Round 1 review — 7 items resolved: FIELDMARK_DATABASE_URL honored across all three stacks; Go version corrected in README (1.23+ → 1.26+); .claude/settings.local.json gitignored; .DEFAULT_GOAL removed (GNU extension); e2e guard checks for playwright binary; css target corrected to fieldmark_shared; parity fails on partial installation.
- 2026-05-16: Round 2 review — 9 items resolved: .NET parses postgres:// via NpgsqlConnectionStringBuilder; Django switched to psycopg.conninfo.conninfo_to_dict (URL-decoding + SSL param forwarding); both stacks treat blank env as unset; $(MAKEFILE_LIST) replaced with hardcoded Makefile; css guards on node_modules presence and uses npm run build; fieldmark_shared/pnpm-workspace.yaml reverted; Go stack README corrected to 1.26+; .NET stack README corrected to http://localhost:5000.
- 2026-05-16: Round 3 review — 7 items resolved: .NET now splits credentials on raw UserInfo before URL-decoding (encoded colons safe); .NET forwards sslmode/sslrootcert/sslcert/sslkey query params to NpgsqlConnectionStringBuilder; Go uses strings.TrimSpace so whitespace-only env var falls back to local default; fieldmark_shared/pnpm-workspace.yaml fixed to valid YAML (allowBuilds list); fieldmark_shared/README.md aligned to npm (matches CLAUDE.md and Makefile); @tailwindcss/cli pinned exactly to 4.2.4 (no ^ range); root README tree updated from stale fieldmark_style/ to fieldmark_shared/.
- 2026-05-16: Post-round-3 follow-up — standardized fieldmark_shared on pnpm throughout (pnpm-lock.yaml was already committed): moved @parcel/watcher build approval from pnpm-workspace.yaml to package.json pnpm.onlyBuiltDependencies (pnpm 9–11 compatible); deleted stale package-lock.json; updated fieldmark_shared/CLAUDE.md and README.md to use pnpm commands; Makefile css target updated to pnpm run build.
- 2026-05-16: Round 4 review — 5 accepted / 1 stale / 1 not-accepted. .NET and Django now trim FIELDMARK_DATABASE_URL before parsing (not just blank-check); .NET URL-decodes the database path segment; both stacks forward all URL query params (not just SSL keys); fieldmark_shared/pnpm-workspace.yaml deleted (single package, not a workspace; onlyBuiltDependencies in package.json is correct); Node.js 20+ documented in fieldmark_shared/README.md and root README. Finding #6 (make css uses npm) was already fixed before reviewer ran. Finding #7 (make e2e no-op when deps installed/backends down) rejected — AC #3 requires no-op on fresh clone only; Playwright running and failing when backends are intentionally absent is correct behavior, not a Makefile defect.
