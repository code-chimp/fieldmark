# Story 1.1: Confirm three native scaffolds, root Makefile, and Docker Compose harness

Status: ready-for-dev

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

- [ ] **Task 1: Audit the three native scaffolds against Architecture §Initialization Commands (AC: #4)**
  - [ ] 1.1 Read [FieldMark/FieldMark.sln](FieldMark/FieldMark.sln) and confirm four projects exist: `FieldMark.Web`, `FieldMark.Domain`, `FieldMark.Data`, `FieldMark.Tests.Domain`, `FieldMark.Tests.Integration` (5 projects total — the spec lists "4-project solution" plus 2 test projects). Confirm `Directory.Build.props` enforces `<Nullable>enable</Nullable>`, `<TreatWarningsAsErrors>true</TreatWarningsAsErrors>`, `<AnalysisMode>Recommended</AnalysisMode>`, `<EnforceCodeStyleInBuild>true</EnforceCodeStyleInBuild>`.
  - [ ] 1.2 Read [fieldmark_py/pyproject.toml](fieldmark_py/pyproject.toml) and confirm Python 3.14, `django>=6.0`, `psycopg[binary]>=3.3` are pinned; dev deps include `ruff`, `black`, `mypy`, `django-stubs`, `pytest`, `pytest-django`. Confirm the seven Django apps exist as directories: `projects/`, `inspections/`, `violations/`, `compliance/`, `audit/`, `reference/`, `grid/`.
  - [ ] 1.3 Read [fieldmark-go/go.mod](fieldmark-go/go.mod) and confirm `github.com/gofiber/fiber/v3`, `github.com/gofiber/template/html/v2`, `github.com/jackc/pgx/v5` are present. Confirm the layered package layout: `cmd/web/`, `internal/{app,data,domain,web}/`. The existing [fieldmark-go/Makefile](fieldmark-go/Makefile) stays as the stack-local Go quality gate — do not delete or modify it.
  - [ ] 1.4 If anything is missing or diverges from the Architecture spec, **stop and surface the divergence** rather than silently fixing — divergence from §Initialization Commands is an ADR amendment, not a story task.

- [ ] **Task 2: Audit `docker-compose.yml` and the init-script directory (AC: #1)**
  - [ ] 2.1 Read [docker-compose.yml](docker-compose.yml) and confirm it has: Postgres 17 image, port `5432:5432`, named volume `postgres_data`, env vars `POSTGRES_DB/USER/PASSWORD=fieldmark/fieldmark/fieldmark`, init-scripts volume mount `./docker/postgres/init:/docker-entrypoint-initdb.d`.
  - [ ] 2.2 Confirm [docker/postgres/init/001_schemas.sql](docker/postgres/init/001_schemas.sql), [docker/postgres/init/010_domain_tables.sql](docker/postgres/init/010_domain_tables.sql), [docker/postgres/init/020_domain_seed.sql](docker/postgres/init/020_domain_seed.sql) all exist (already verified in Story 1.2's preconditions — this task confirms the docker-compose path mounts them).

- [ ] **Task 3: Author the root `Makefile` per Architecture D20 (AC: #2, #3)**
  - [ ] 3.1 Create `/Users/timothygoshinski/work/lab/htmx/fieldmark/Makefile` (POSIX make, no GNU-only extensions, BSD-compatible — works on macOS default `make` and Linux `gmake`). Use **tabs for recipe indentation** (Makefile syntax requirement — most common breakage on fresh-clone runs).
  - [ ] 3.2 Declare `.PHONY:` for every non-file target. Implement targets in this exact order (matches Story AC text and D20):
    - `up` — `docker compose up -d`
    - `down` — `docker compose down`
    - `reset` — `docker compose down -v && docker compose up -d` (destroys volume; re-runs init scripts)
    - `run-net` — `cd FieldMark && dotnet run --project FieldMark.Web`
    - `run-django` — `cd fieldmark_py && uv run python manage.py runserver`
    - `run-go` — `cd fieldmark-go && go run ./cmd/web`
    - `test-net` — `cd FieldMark && dotnet test`
    - `test-django` — `cd fieldmark_py && uv run pytest`
    - `test-go` — `cd fieldmark-go && go test ./...`
    - `e2e` — `cd e2e && pnpm test` (target may no-op cleanly until Story 7.1 lands the `e2e/` package — see Task 3.4)
    - `parity` — `tools/parity/diff-routes.sh && tools/parity/diff-pg-indexes.sh` (no-ops cleanly until Story 1.3 lands the scripts — see Task 3.4)
    - `css` — `pnpm --filter fieldmark_style build` (no-ops cleanly until Story 1.4 lands `fieldmark_style/` — see Task 3.4)
  - [ ] 3.3 Add a default `help` target (and make it the default goal) that prints each target with a one-line description (use `## comment` after target name + a single awk recipe to extract — standard pattern).
  - [ ] 3.4 **Important: targets for future stories must no-op cleanly, not error.** Wrap the not-yet-implemented targets with an existence guard:
    ```makefile
    e2e:
    	@if [ -d e2e ] && [ -f e2e/package.json ]; then \
    		cd e2e && pnpm test; \
    	else \
    		echo "(skip) e2e/ not yet scaffolded — lands in Story 7.1"; \
    	fi
    ```
    Same pattern for `parity` (skip if `tools/parity/diff-routes.sh` is absent) and `css` (skip if `fieldmark_style/package.json` is absent). This keeps AC #3's "succeeds (or no-ops cleanly) on a fresh clone" honest.
  - [ ] 3.5 Verify the file's recipe lines use **literal TAB characters**, not spaces. After writing, run `cat -A Makefile | head -40` and confirm every recipe line starts with `^I` (tab indicator).

- [ ] **Task 4: Document the workflow in repo-root `README.md` (AC: #4)**
  - [ ] 4.1 Read the existing [README.md](README.md) at repo root. If it does not exist or lacks a "Getting Started" section, add one. If it exists, update only what's needed — do not rewrite unrelated content.
  - [ ] 4.2 The Getting Started section must document, in order:
    1. Prerequisites: Docker Desktop / Docker Engine, .NET 10 SDK, Python 3.14 + `uv`, Go 1.23+, `pnpm` (for later stories' CSS build / e2e).
    2. `make up` — start Postgres.
    3. `make run-net` / `make run-django` / `make run-go` in three terminals.
    4. Default URLs: .NET http://localhost:5000, Django http://localhost:8000, Go http://localhost:3000.
    5. `make reset` to nuke the DB volume and re-run init scripts.
    6. Link to per-stack READMEs for stack-specific dev instructions: [FieldMark/README.md](FieldMark/README.md), [fieldmark_py/README.md](fieldmark_py/README.md), [fieldmark-go/README.md](fieldmark-go/README.md).
    7. Pointer to [CLAUDE.md](CLAUDE.md) for architectural rules and the per-stack `CLAUDE.md` files.
  - [ ] 4.3 Verify each stack's existing `README.md` already documents the run command for that stack. If any is missing, add it (one-line addition) — but **do not rewrite** existing stack READMEs.

- [ ] **Task 5: End-to-end smoke verification (AC: #1, #2, #3)**
  - [ ] 5.1 From a clean shell at repo root: `make up`. Wait ~5 seconds. Verify with `docker ps` that the `fieldmark-local` container is `healthy` (or at least `Up`).
  - [ ] 5.2 Connect with `psql postgres://fieldmark:fieldmark@localhost:5432/fieldmark -c "\dn"` and confirm the five schemas (`domain`, `dotnet_auth`, `django_auth`, `fiber_auth`, `infra`) are present. (Story 1.2 owns the deeper DDL verification; this is a quick smoke for AC #1.)
  - [ ] 5.3 `make run-net` in one terminal — confirm it builds and binds to :5000 without DB connection errors. `Ctrl-C` to stop.
  - [ ] 5.4 `make run-django` in another terminal — confirm `uv run python manage.py runserver` binds to :8000 without errors. `Ctrl-C` to stop. (Note: `runserver` defaults to :8000; if a port flag is needed, document it in the Makefile.)
  - [ ] 5.5 `make run-go` in a third terminal — confirm `go run ./cmd/web` binds to :3000. `Ctrl-C` to stop. (If the entrypoint binds to a different port today, surface that as a divergence — port assignments are part of the Architecture spec and `make parity` will eventually depend on them.)
  - [ ] 5.6 `make reset` — confirm the volume is destroyed and recreated; the init scripts run again.
  - [ ] 5.7 `make e2e`, `make parity`, `make css` — confirm each prints its `(skip)` message and exits `0` per Task 3.4.

- [ ] **Task 6: Cross-stack parity sanity check (AC: #3)**
  - [ ] 6.1 The `parity` target no-ops at this story (Story 1.3 lands the actual scripts). Just confirm Task 3.4's skip message fires.
  - [ ] 6.2 Manually confirm the three stack run commands work simultaneously (run all three at once — they bind to different ports, share the same DB, and do not collide).

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

### Completion Notes List

- Ultimate context engine analysis completed — comprehensive developer guide created.
- Primary deliverable is the root `Makefile`; all other tasks are audit + documentation.
- Future-story targets (`e2e`, `parity`, `css`) must no-op cleanly per Task 3.4 — this is not a TODO, it's a structural requirement of AC #3.

### File List

*(Populated by dev agent upon completion.)*
