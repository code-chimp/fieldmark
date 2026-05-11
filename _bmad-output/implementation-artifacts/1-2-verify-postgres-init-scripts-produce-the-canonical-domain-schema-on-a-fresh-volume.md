# Story 1.2: Verify Postgres init scripts produce the canonical `domain.*` schema on a fresh volume

Status: ready-for-dev

## Story

As a developer working across three stacks,
I want a single command that destroys and re-creates the database in a known canonical state,
so that any drift between framework mapping code and the infrastructure-owned schema surfaces immediately.

## Acceptance Criteria

1. **Given** a running database with arbitrary local state, **When** I run `make reset` (`docker compose down -v && docker compose up -d`), **Then** the volume is destroyed and recreated, **And** `001_schemas.sql`, `010_domain_tables.sql`, and `020_domain_seed.sql` execute in order with no errors visible in `docker logs`.

2. **Given** the database has been initialized, **When** I connect with `psql` and run `\dn`, **Then** the schemas `domain`, `dotnet_auth`, `django_auth`, `fiber_auth`, `infra` are all present.

3. **Given** the database has been initialized, **When** I run `SELECT table_name FROM information_schema.tables WHERE table_schema='domain' ORDER BY table_name`, **Then** exactly 12 tables are returned: `audit_entry`, `compliance_rule`, `corrective_action`, `finding`, `inspection`, `job_site`, `project`, `project_inspector`, `project_trade_scope`, `trade_type`, `violation`, `violation_category`.

4. **Given** the database has been initialized, **When** I inspect `domain.trade_type`, `domain.violation_category`, and `domain.compliance_rule`, **Then** the reference rows from `020_domain_seed.sql` are present and identical to the file's `INSERT` statements (verified by row count + a `SELECT` sample).

5. **Given** the canonical DDL is owned by infrastructure (ADR-014), **When** I grep each stack for tooling that could mutate the `domain` schema (`dotnet ef migrations add` against a DbContext whose `HasDefaultSchema` is `"domain"`, Django `makemigrations` against a `domain.*` model with `Meta.managed = True`, Go migration tools targeting `domain.*`), **Then** zero matches are found, **And** each stack's README explicitly states that `domain.*` is infrastructure-owned and that framework migrations only apply to its `*_auth` schema.

## Tasks / Subtasks

- [ ] **Task 1: Precondition check — Story 1.1 must be `done` (AC: #1)**
  - [ ] 1.1 Confirm the root `Makefile` exists at repo root and exposes the `reset` target. If not, halt and fix Story 1.1 first — `make reset` is the entry point for AC #1.
  - [ ] 1.2 Confirm `docker compose` V2 is installed (`docker compose version` returns a V2 string). On macOS this ships with Docker Desktop.

- [ ] **Task 2: Inventory the init scripts as the canonical specification (AC: #1, #3, #4)**
  - [ ] 2.1 Read [docker/postgres/init/001_schemas.sql](docker/postgres/init/001_schemas.sql) (22 lines). Confirm it creates exactly five schemas: `domain`, `django_auth`, `dotnet_auth`, `fiber_auth`, `infra`. Do **not** modify this file.
  - [ ] 2.2 Read [docker/postgres/init/010_domain_tables.sql](docker/postgres/init/010_domain_tables.sql) (223 lines). Inventory and record:
    - 12 `CREATE TABLE` statements: `trade_type`, `violation_category`, `compliance_rule`, `project`, `job_site`, `project_trade_scope`, `project_inspector`, `inspection`, `finding`, `violation`, `corrective_action`, `audit_entry`.
    - 5 `CREATE INDEX` / `CREATE UNIQUE INDEX` statements (the canonical index inventory for `domain.*`). Note their names — they are the spec against which `make parity`'s `pg_indexes` zero-diff check will assert in Story 1.3.
    - At least one `ALTER TABLE ... ADD CONSTRAINT` for the forward reference `finding.spawned_violation_id → domain.violation(id)` (created after `domain.violation` is defined).
    - All enum-like columns use `VARCHAR + CHECK`, not Postgres `ENUM` types (per ADR-014).
    - All timestamps are `TIMESTAMPTZ`.
    - No foreign keys from `domain.*` to any `*_auth.*` schema (ADR-012 — user references are opaque UUIDs).
    Do **not** modify this file.
  - [ ] 2.3 Read [docker/postgres/init/020_domain_seed.sql](docker/postgres/init/020_domain_seed.sql) (145 lines). Inventory and record:
    - 1 `INSERT INTO domain.trade_type` statement with 4 rows (Electrical, Plumbing, HVAC, Structural — or similar canonical trades).
    - 1 `INSERT INTO domain.violation_category` statement covering each severity level.
    - 1 `INSERT INTO domain.compliance_rule` statement with the four canonical MVP rules.
    - Hardcoded UUIDs in every row (not `gen_random_uuid()`) — they must be stable across `docker compose down -v && docker compose up -d` cycles.
    - No `ON CONFLICT DO NOTHING` clauses (the file's header explicitly forbids them — a script running twice means the container was not recreated cleanly).
    Do **not** modify this file.

- [ ] **Task 3: Author the post-reset verification script `tools/verify-domain-schema.sh` (AC: #1, #2, #3, #4)**
  - [ ] 3.1 Create `tools/verify-domain-schema.sh` (executable, `chmod +x`). The script connects to `postgresql://fieldmark:fieldmark@localhost:5432/fieldmark` and runs four assertions in order, exiting non-zero on any failure:
    1. **Schemas:** `psql -c "\dn"` returns exactly 5 user schemas (filter out built-in `public`, `pg_*`, `information_schema`): `django_auth`, `domain`, `dotnet_auth`, `fiber_auth`, `infra`.
    2. **Tables:** `psql -tAc "SELECT table_name FROM information_schema.tables WHERE table_schema='domain' ORDER BY table_name"` returns exactly the 12 expected names (compare to a hard-coded array — order matters; ORDER BY ensures determinism).
    3. **Reference seed counts:** `SELECT count(*) FROM domain.trade_type` returns ≥ 4; `SELECT count(*) FROM domain.violation_category` returns ≥ 4 (covers each severity level); `SELECT count(*) FROM domain.compliance_rule` returns exactly 4 (per the seed file's "four canonical MVP rules" comment).
    4. **Spot-check seed content:** `SELECT code FROM domain.trade_type WHERE code IN ('PLUMBING','ELECTRICAL') AND active = true` returns 2 rows. (Adjust codes to whatever the seed file actually uses — Task 2.3's inventory is the source of truth.)
  - [ ] 3.2 The script prints `OK domain schema verified (5 schemas, 12 tables, N reference rows)` on success and a precise diff on failure. Use `set -euo pipefail` at the top.
  - [ ] 3.3 The script must be idempotent and side-effect-free (read-only queries only).
  - [ ] 3.4 Document the script's usage in the repo-root README under the "Verifying the database" subsection of "Getting Started": `./tools/verify-domain-schema.sh` after `make up` / `make reset`.

- [ ] **Task 4: End-to-end fresh-volume verification (AC: #1, #2, #3, #4)**
  - [ ] 4.1 From a clean shell: `make reset`. Verify in `docker logs fieldmark-local` that all three init scripts ran ("CREATE SCHEMA", "CREATE TABLE", "INSERT 0 N" lines visible) and that no `ERROR:` lines appear.
  - [ ] 4.2 Run `./tools/verify-domain-schema.sh`. Confirm it exits `0` with the success banner.
  - [ ] 4.3 Manually run `psql postgres://fieldmark:fieldmark@localhost:5432/fieldmark -c "\dn"` to confirm five user schemas.
  - [ ] 4.4 Manually run `psql ... -c "\dt domain.*"` to confirm 12 tables.
  - [ ] 4.5 Manually run `psql ... -c "SELECT code, name, active FROM domain.trade_type ORDER BY code"` and confirm rows match the seed file verbatim.

- [ ] **Task 5: Confirm no framework migration tools target `domain.*` (AC: #5)**
  - [ ] 5.1 **.NET:** `grep -rn 'HasDefaultSchema.*"domain"' FieldMark/` should return zero matches. The `AuthDbContext` (when it lands in Story 1.7) will use `HasDefaultSchema("dotnet_auth")` — never `"domain"`. If `FieldMark.Data/Configuration/` contains EF Core fluent configs that `ToTable("<name>", "domain")` (these arrive in Story 2.1+), that is correct — they *map* to existing tables, they do not migrate them. Distinguish "map" (allowed, via `ToTable` + `HasDefaultSchema` on the correct context) from "migrate" (forbidden — would appear as a migration file under `FieldMark.Data/Migrations/` targeting `domain.*`).
  - [ ] 5.2 **Django:** `grep -rn "managed = True" fieldmark_py/` for any model targeting `domain.*` should return zero matches. All `domain.*` model classes (when they land in Stories 2.1, 3.1, etc.) will declare `class Meta: managed = False` + `db_table = 'domain"."<table>'`. Also confirm no migration file under any app's `migrations/` directory creates or alters a `domain.*` table — search with `grep -rln "schema=.domain.\|domain\.\|'domain'" fieldmark_py/*/migrations/ 2>/dev/null` and audit any matches. Django's built-in `auth` migrations targeting `django_auth` are fine.
  - [ ] 5.3 **Go:** `find fieldmark-go -name '*.go' | xargs grep -l 'CREATE TABLE\|ALTER TABLE\|DROP TABLE' 2>/dev/null` should return zero matches in production code. If a Go file embeds migration SQL strings touching `domain.*`, that is a defect — Go reads `domain.*` via `pgx`, never migrates it. (Test fixture SQL is fine if scoped to test schemas; flag any production-path matches.)
  - [ ] 5.4 Confirm each stack's README has explicit text stating that `domain.*` is infrastructure-owned. Required wording (or equivalent): *"The `domain` schema is created and migrated by `docker/postgres/init/`. Framework migrations in this stack apply only to the `<stack>_auth` schema. Running `<framework migration tool>` against `domain.*` is a defect."* If any of the three READMEs lacks this, add a "Schema ownership" subsection — keep it brief (3–5 lines).

- [ ] **Task 6: Document the verification workflow in repo-root README**
  - [ ] 6.1 Add a "Verifying the database" subsection to the Getting Started section authored in Story 1.1's Task 4. Content:
    ```
    After `make up` or `make reset`, verify the canonical schema:

        ./tools/verify-domain-schema.sh

    Expected output:
        OK domain schema verified (5 schemas, 12 tables, N reference rows)

    Non-zero exit = schema drift. Investigate before running any stack.
    ```
  - [ ] 6.2 Cross-link to Story 1.3's `make parity` (which adds index-diff coverage on top of schema presence) when that story lands — leave a `<!-- TODO: link from Story 1.3 -->` comment as a hook, but no link yet.

## Dev Notes

### Brownfield posture

**The init scripts already exist and are authoritative.** Your job is **verification**, not modification. The only new code in this story is `tools/verify-domain-schema.sh` + minor README updates.

If anything in Task 2's inventory diverges from what the story expects (12 tables, 5 indexes, 4 trade types, 4 compliance rules), **stop and surface the divergence** rather than fixing the SQL files. The DDL is infrastructure-owned per ADR-014; changes to it require an ADR amendment.

### Architectural patterns and constraints

- **ADR-014 — infrastructure-owned `domain` schema.** No framework owns or migrates this schema. The verification script enforces this at runtime; Task 5 enforces it at the codebase level.
- **ADR-013 — schema isolation.** Each framework has its own auth schema (`django_auth`, `dotnet_auth`, `fiber_auth`). Framework migrations target only their own auth schema.
- **ADR-012 — opaque user references.** `domain.*` rows reference users only via UUID columns; no FK to any `*_auth.*` schema. The verification script does not check this directly (it would require introspecting every column), but Task 2.2 confirms the comment header asserts it.
- **No SQLite in tests** (per repo-root CLAUDE.md). The verification script connects to the real Postgres 17 container — there is no in-memory shortcut.
- **`docker compose` not `docker-compose`.** V2 plugin syntax throughout.
- **`set -euo pipefail`** in every shell script you author. Failure to set this is a primary cause of silent script failures.

### Source tree components to touch

| Path | Action | Reason |
|---|---|---|
| `tools/verify-domain-schema.sh` | **NEW** | Story 1.2's primary deliverable. |
| `README.md` (repo root) | **UPDATE** | Add "Verifying the database" subsection. Preserve all content authored in Story 1.1. |
| `FieldMark/README.md` | **UPDATE (small)** | Add "Schema ownership" subsection if missing. |
| `fieldmark_py/README.md` | **UPDATE (small)** | Same. |
| `fieldmark-go/README.md` | **UPDATE (small)** | Same. |
| `docker-compose.yml` | **DO NOT TOUCH** | Already correct. |
| `docker/postgres/init/*.sql` | **DO NOT TOUCH** | Infrastructure-owned per ADR-014. |
| `Makefile` (repo root) | **DO NOT TOUCH** | Story 1.1's deliverable; `reset` target already in place. |

### Files this story reads (must read before authoring the verification script)

- [docker/postgres/init/001_schemas.sql](docker/postgres/init/001_schemas.sql) — 22 lines
- [docker/postgres/init/010_domain_tables.sql](docker/postgres/init/010_domain_tables.sql) — 223 lines (12 tables, 5 indexes, forward-reference ALTER TABLE)
- [docker/postgres/init/020_domain_seed.sql](docker/postgres/init/020_domain_seed.sql) — 145 lines (trade_type, violation_category, compliance_rule)
- [docker-compose.yml](docker-compose.yml) — for the connection string parameters
- [FieldMark/README.md](FieldMark/README.md), [fieldmark_py/README.md](fieldmark_py/README.md), [fieldmark-go/README.md](fieldmark-go/README.md) — for the README updates in Task 5.4

### Hard rules (from CLAUDE.md)

- **Infrastructure-owned domain schema.** Verification only; do not edit.
- **No CI in MVP.** The verification script is a local-discipline tool, not a CI gate. (Story 7.6 closes parity verification as the final demo gate.)
- **Stack symmetry.** All three READMEs receive equivalent "Schema ownership" updates — diverging text is itself a drift defect.

### Testing standards

- The verification script is its own test. There is no separate unit test for it — it's a one-shot smoke.
- Run the script three times during development: (a) on a fresh `make reset`, (b) after intentionally `DROP TABLE domain.audit_entry` (must exit non-zero with a clean message), (c) after `make reset` again (must exit `0`). The middle case validates the script's failure mode.
- Do **not** add a Postgres healthcheck or `wait-for-it` loop — `docker compose up -d` returns when the container is up, and the init scripts run synchronously before the container reports ready on the first start. If timing flakes appear in practice, that's a Story 1.1 follow-up, not Story 1.2.

### Previous Story Intelligence (Story 1.1)

- Story 1.1 authored the root `Makefile` with `up`, `down`, `reset` targets. `make reset` is `docker compose down -v && docker compose up -d` — destroys the volume so init scripts re-run.
- Story 1.1 also documented Getting Started in the repo-root `README.md`. Task 6 here **extends** that section — do not rewrite it.
- Story 1.1's Task 3.4 established a "future-target no-op skip" pattern in the Makefile for `e2e`, `parity`, `css`. This story does **not** add new Makefile targets — the verification script is invoked directly. (A future story could add `make verify-db` if desired; not in scope here.)
- Story 1.1 confirmed the three stack scaffolds match Architecture §Initialization Commands. This story does not re-audit them.

### Project Structure Notes

- The `tools/` directory does not yet exist at repo root (Story 1.3 introduces `tools/parity/`). This story creates `tools/verify-domain-schema.sh` as the first inhabitant of `tools/`. The directory does **not** need a README at this point — Story 1.3 will populate it further.
- No conflict with `_bmad-output/`, `docs/`, or the three stack directories.

### Library / framework requirements

- **`psql`** — must be available on the developer's PATH. On macOS, `brew install libpq` followed by `brew link --force libpq` is the standard install; or `brew install postgresql@17` which includes it. The repo-root README should mention this prerequisite (add to the Story 1.1 prerequisites list if missing).
- **`docker compose`** V2 — already a prerequisite from Story 1.1.
- **No new framework dependencies.** This story is shell + SQL only.

### References

- [_bmad-output/planning-artifacts/epics.md](_bmad-output/planning-artifacts/epics.md) — Story 1.2 source.
- [_bmad-output/planning-artifacts/architecture.md](_bmad-output/planning-artifacts/architecture.md) — §Data Architecture (lines 288–326), §D2 Postgres init script ordering.
- [CLAUDE.md](CLAUDE.md) — Database Schema Ownership table; Hard Rules ("Infrastructure-owned domain schema").
- [docker/postgres/init/001_schemas.sql](docker/postgres/init/001_schemas.sql), [docker/postgres/init/010_domain_tables.sql](docker/postgres/init/010_domain_tables.sql), [docker/postgres/init/020_domain_seed.sql](docker/postgres/init/020_domain_seed.sql) — canonical specification.
- [_bmad-output/implementation-artifacts/1-1-confirm-three-native-scaffolds-root-makefile-and-docker-compose-harness.md](_bmad-output/implementation-artifacts/1-1-confirm-three-native-scaffolds-root-makefile-and-docker-compose-harness.md) — Story 1.1 (prerequisite).

## Dev Agent Record

### Agent Model Used

claude-opus-4-7

### Debug Log References

### Completion Notes List

- Ultimate context engine analysis completed — comprehensive developer guide created.
- Primary deliverable is `tools/verify-domain-schema.sh`; all SQL files are AUDIT-ONLY.
- Cross-stack README updates enforce ADR-014 at the documentation level; Task 5 enforces at the codebase level.
- Story 1.1 is a precondition — `make reset` must work before this story's AC #1 can be verified.

### File List

*(Populated by dev agent upon completion.)*
