You are helping finalize pre-kickoff planning artifacts for FieldMark — a construction
compliance and inspection management system implemented across three backend stacks
(.NET Razor Pages, Django, Go Fiber) against a shared PostgreSQL database. This is a
teaching/demo project, not a production system.

---

## Repository layout (key paths)

```
fieldmark/
├── FieldMark/                         .NET solution
├── fieldmark_py/                      Django project
├── fieldmark-go/                      Go Fiber project (bare skeleton only)
├── docker-compose.yml                 currently brings up Postgres only; no init mounts yet
├── docker/                            directory exists but is empty
├── docs/                              NEW strategy docs (read these first — see below)
├── _bmad-output/planning-artifacts/
│   └── research/                      existing planning artifacts (source of truth)
│       ├── project-brief.md
│       ├── prd.md
│       ├── domain-model.md
│       ├── architecture-decisions.md  currently has ADR-011 + constraints only
│       ├── ux-guide.md
│       ├── dotnet-reference.md
│       └── django-reference.md
└── README.md
```

---

## What changed since the research folder was last updated

Five new strategy documents were authored in `docs/` that expand the project scope
and introduce three new architectural decisions. You must read all five before touching
anything:

- `docs/FieldMark_Authentication_Authorization_Strategy.md`
- `docs/FieldMark_Docker_Init_and_Compose_Strategy.md`
- `docs/FieldMark_Shared_Domain_Schema_Ownership_Strategy.md`
- `docs/FieldMark_Fiber_Architecture_and_Standup_Guide.md`
- `docs/FieldMark_ADR_Addendum_Auth_Init_Domain.md`

The three new ADRs (012, 013, 014) reverse or supersede assumptions baked into the
existing research docs. The most important changes:

1. **Go Fiber is now a third stack** alongside .NET and Django.

2. **The shared domain schema is infrastructure-owned** (ADR-014). It is NOT created
   or evolved by EF Core or Django migrations. It is created by PostgreSQL init scripts
   mounted into the Docker container. EF Core and Django map to `domain.*` tables but
   do not own them. This directly contradicts language in `dotnet-reference.md` and
   the Django stack's CLAUDE.md that says EF Core owns the domain schema.

3. **Authentication is framework-local** (ADR-012). Each stack has its own Postgres
   schema for auth: `django_auth`, `dotnet_auth`, `fiber_auth`. The `domain` schema
   stores user references as opaque identifiers — no foreign keys to any auth table.

4. **Postgres schemas are created by init scripts** (ADR-013), not by application
   migrations. The `docker/postgres/init/` directory must contain SQL that creates:
   `domain`, `django_auth`, `dotnet_auth`, `fiber_auth`, `infra`.

---

## Your two tasks

### Task 1 — Update the research artifacts

Integrate the new strategy docs into the existing research folder. Specifically:

**`architecture-decisions.md`** — append ADR-012, ADR-013, and ADR-014 from the ADR
addendum doc as Part 1 additions, following the same format as ADR-011. The existing
Part 2 constraints section should also gain a new constraint bullet confirming that
domain schema is infrastructure-owned (not framework-migrated) and that auth schemas
are framework-local.

**`dotnet-reference.md`** — remove or correct the statement that EF Core owns the
domain schema. EF Core maps to `domain.*` tables using explicit fluent configuration
but does not create or migrate them. EF Core migrations are scoped to `dotnet_auth`
only. Note that `UseSnakeCaseNamingConvention()` and explicit `ToTable()` calls are
required to align with the infrastructure-defined schema.

**`django-reference.md`** — same correction: Django migrations own only `django_auth`.
Django models that represent domain tables must use `managed = False` (or equivalent
discipline) so Django never attempts to create or alter those tables.

**`project-brief.md`** — update the two-stack references to three-stack, update the
Appendix A inventory table to reference a new `fiber-reference.md` file (see below),
and note the infrastructure-owned domain schema in the technical considerations section.

**`domain-model.md`** — the schema sketch in §8 should reflect the `domain` schema
prefix (e.g., `domain.projects`, `domain.violations`) on all shared tables. Verify
the naming conventions section (§9) still holds and is consistent with the new
infrastructure ownership model.

**New file: `fiber-reference.md`** — create this as a peer to `dotnet-reference.md`
and `django-reference.md`. Draw from `docs/FieldMark_Fiber_Architecture_and_Standup_Guide.md`.
It should cover: project layout, layer responsibilities, dependency direction rules,
persistence approach (explicit SQL against `domain.*`), auth deferral (`fiber_auth`
exists in the DB but Fiber auth is not yet implemented), HTMX/template strategy, and
agent guardrails.

### Task 2 — Wire up the Docker Compose Postgres init

Create the infrastructure that ADR-013 requires:

**`docker/postgres/init/001_schemas.sql`** — SQL init script that creates all required
schemas:

```sql
CREATE SCHEMA IF NOT EXISTS domain;
CREATE SCHEMA IF NOT EXISTS django_auth;
CREATE SCHEMA IF NOT EXISTS dotnet_auth;
CREATE SCHEMA IF NOT EXISTS fiber_auth;
CREATE SCHEMA IF NOT EXISTS infra;
```

**`docker-compose.yml`** — update the existing file to mount the init directory into
the Postgres container:

```yaml
volumes:
  - ./docker/postgres/init:/docker-entrypoint-initdb.d
```

Important: the init script only runs on first container startup (when the data volume
is empty). If the container has already been started, the developer must destroy the
volume and restart for schemas to be created. Note this in a comment in the compose file.

---

## Conventions to preserve

- All file names in `research/` use kebab-case.
- Internal cross-references use relative paths (no `docs/` prefix — files reference
  siblings directly, e.g., `domain-model.md` not `docs/domain-model.md`).
- ADR format: Status, Context, Decision, Consequences, Alternatives Rejected.
- The `archive/` subdirectory inside `research/` has been deleted; do not recreate it.
- Do not touch `docs/` files — they are raw strategy input, not planning artifacts.
- Do not modify any stack source code.

---

## Definition of done

- All five `docs/` strategy documents are fully reflected in the research artifacts.
- `architecture-decisions.md` contains ADR-011 through ADR-014.
- `dotnet-reference.md` and `django-reference.md` no longer claim ORM ownership of
  the domain schema.
- `fiber-reference.md` exists and is referenced from `project-brief.md` Appendix A.
- `docker/postgres/init/001_schemas.sql` exists and creates all five schemas.
- `docker-compose.yml` mounts the init directory.
- No planning artifact references the old `docs/` paths or the deleted archive.
