# Story 1.8: Wire Django built-in `auth` to `django_auth` schema with conceptual-role Groups

Status: done

## Story

As an administrator using the Django stack,
I want framework-native authentication backed by the `django_auth` schema with role assignment via Groups,
So that Django's identity layer mirrors the .NET stack's isolation (Story 1.7), never touches `domain.*`, and is ready for the login/logout flow that lands in Story 1.11.

## Acceptance Criteria

1. **Django auth tables resolve into the `django_auth` schema.** All framework-managed tables created by `django.contrib.auth`, `django.contrib.sessions`, `django.contrib.contenttypes`, and `django.contrib.admin` land in the `django_auth` Postgres schema. The current `fieldmark_py/fieldmark/settings.py` already configures this via `OPTIONS["options"] = "-c search_path=django_auth,public"` on the default database connection — this is the **chosen mechanism** (a third valid option alongside a custom `DatabaseRouter` or per-model `db_table` overrides; the epic AC lists those two but the `search_path` mechanism is equivalent in effect and simpler to maintain). The dev's task on this AC is to **verify the resulting placement**, document the choice in `settings.py` next to the existing comment, and confirm the explicit table list lands in `django_auth` (see AC #2).

2. **After `migrate`, the eight required tables exist in `django_auth` and nowhere else.** `uv run python manage.py migrate` succeeds against a freshly-reset database (after `make reset`). Verified by `psql`:
   - `\dt django_auth.*` lists at minimum: `auth_user`, `auth_group`, `auth_permission`, `auth_user_groups`, `auth_user_user_permissions`, `auth_group_permissions`, `django_session`, `django_admin_log`.
   - The supporting `django_content_type` and `django_migrations` tables also land in `django_auth` (consequence of `search_path` — this is expected and correct; documented in settings).
   - `\dt domain.*` is unchanged from the `010_domain_tables.sql` canonical inventory (12 tables — Story 1.2).
   - `\dt public.*` returns zero rows — no auth table accidentally lands in `public`.

3. **No Django auth migration touches `domain.*`.** Verified two ways:
   - `SELECT app, name FROM django_auth.django_migrations ORDER BY app, name;` lists only the built-in apps' migrations (`admin`, `auth`, `contenttypes`, `sessions`). No row references a `domain.*` table.
   - `grep -rn 'domain' fieldmark_py/*/migrations/` returns zero matches outside of `__pycache__`. App migration folders (`projects/migrations/`, `inspections/migrations/`, etc.) remain empty (`__init__.py` only) — Story 1.8 does **not** introduce domain models, only auth wiring.

4. **Five canonical conceptual-role Groups are seeded idempotently.** A new management command `uv run python manage.py seed_groups` creates exactly these Django Groups in `django_auth.auth_group`:
   - `ADMIN`
   - `COMPLIANCE_OFFICER`
   - `INSPECTOR`
   - `SITE_SUPERVISOR`
   - `EXECUTIVE`

   Running the command a second time produces **no** duplicates, **no** errors, and prints a one-line "exists" notice per Group instead of "created". Verified by capturing the row IDs after the first run and comparing after the second (IDs unchanged, row count remains exactly 5).

5. **No login/logout views, no auth middleware wiring, no Django Admin re-styling.** This story wires the schema + Groups only. The login/logout HTML views and the unauthenticated-redirect contract are Story 1.11's scope. `MIDDLEWARE` continues to include the default Django auth/session middleware (it is already present); **no** new entries are added. `urls.py` is unchanged — no `path("login/", ...)` or `path("logout/", ...)` is added in this story. The `--dump-routes` output (`uv run python manage.py dump_routes`) is **byte-identical** to its HEAD-before-this-story output (the `dump_routes` command already excludes `/admin/` per `tools/management/commands/dump_routes.py`).

6. **`make parity` exits 0.** After this story lands, `make parity` from the repo root still exits 0. The route inventories of all three stacks remain identical (no Django-only auth routes are added). The `pg_indexes` snapshot for the `domain` schema (`tools/parity/canonical-pg-indexes.txt`) is unchanged — Django auth touches only `django_auth`, never `domain`.

7. **Build, type, lint, and test pipelines stay green.** From `fieldmark_py/`:
   - `uv sync` — clean.
   - `uv run ruff check .` — zero issues.
   - `uv run mypy .` — zero errors (existing baseline).
   - `uv run pytest` — existing tests still pass (a new test asserting the seed command's idempotence is added per Task 5 below).

8. **CLAUDE.md and README authentication sections reflect the new state.** `fieldmark_py/CLAUDE.md` gains an `## Authentication` section that documents the `search_path` mechanism, the migration scope (`django_auth` only — never `domain`), the seed command, and notes that login/logout lands in Story 1.11. `fieldmark_py/README.md`'s "Getting Started" gains a step for `seed_groups` after `migrate`.

## Tasks / Subtasks

- [x] Task 1: Verify and document the existing `search_path` mechanism in `settings.py` (AC: #1, #2)
  - [x] 1.1 Read `fieldmark_py/fieldmark/settings.py` lines 86–117. Confirm `OPTIONS["options"] = "-c search_path=django_auth,public"` is present on `DATABASES["default"]`.
  - [x] 1.2 Expand the existing inline comment to be explicit about the architectural choice: this is the chosen mechanism (vs. a custom `DatabaseRouter` or per-model `db_table` overrides — Architecture D7) because it (a) requires no per-app boilerplate and (b) automatically captures sessions / admin / contenttypes / migrations tables in the same schema as auth. Keep the comment ≤ 6 lines.
  - [x] 1.3 Do **not** add a `fieldmark/routers.py` file even though the architecture's directory-structure diagram (line 1122) shows one — the `search_path` approach supersedes it. Note this divergence explicitly in the Authentication section of `CLAUDE.md` (Task 6).

- [x] Task 2: Apply migrations against a fresh database and verify schema placement (AC: #2, #3)
  - [x] 2.1 From repo root: `make reset` — destroys volume, re-runs `001_schemas.sql`, `010_domain_tables.sql`, `020_domain_seed.sql`.
  - [x] 2.2 Wait for Postgres to be reachable (a few seconds), then from `fieldmark_py/`: `uv run python manage.py migrate`. Output should show 18+ migrations applied across `contenttypes`, `auth`, `admin`, `sessions`.
  - [x] 2.3 Connect with `psql -h localhost -U fieldmark -d fieldmark` (password `fieldmark`) and run:
    - `\dt django_auth.*` — assert the eight required tables (AC #2) plus `django_content_type` and `django_migrations` are present.
    - `\dt domain.*` — assert exactly 12 tables (the canonical inventory from Story 1.2).
    - `\dt public.*` — assert zero rows.
    - `SELECT DISTINCT app FROM django_auth.django_migrations ORDER BY app;` — assert only `admin`, `auth`, `contenttypes`, `sessions` are listed.
  - [x] 2.4 Run `grep -rn 'domain' fieldmark_py/projects/migrations/ fieldmark_py/inspections/migrations/ fieldmark_py/violations/migrations/ fieldmark_py/audit/migrations/ fieldmark_py/compliance/migrations/ fieldmark_py/reference/migrations/ fieldmark_py/grid/migrations/` — assert zero matches outside `__pycache__`.

- [x] Task 3: Create the `seed_groups` management command (AC: #4)
  - [x] 3.1 Create `fieldmark_py/tools/management/commands/seed_groups.py`. Use the `tools` app (already in `INSTALLED_APPS`; already hosts `dump_routes.py`) — this is a cross-cutting bootstrap command, not aggregate-specific. Do **not** put it under `projects/` (that location is reserved for `seed_dev_users.py` in Story 1.10, per Architecture line 1140).
  - [x] 3.2 Implement the command using `django.contrib.auth.models.Group` and `Group.objects.get_or_create(name=...)`. The exact pattern:

    ```python
    """
    Management command: seed_groups

    Seeds the five canonical conceptual-role Groups required by FieldMark's
    authorization model (Architecture D7). Idempotent — safe to re-run.
    """

    from django.contrib.auth.models import Group
    from django.core.management.base import BaseCommand

    CANONICAL_GROUPS = (
        "ADMIN",
        "COMPLIANCE_OFFICER",
        "INSPECTOR",
        "SITE_SUPERVISOR",
        "EXECUTIVE",
    )


    class Command(BaseCommand):
        help = "Seed canonical conceptual-role Groups (idempotent)"

        def handle(self, *args, **options):
            for name in CANONICAL_GROUPS:
                _, created = Group.objects.get_or_create(name=name)
                verb = "created" if created else "exists"
                self.stdout.write(f"{verb}: {name}")
    ```

  - [x] 3.3 Do **not** wire this command into `apps.py:ready()` or a post-`migrate` signal. Django explicitly warns against doing database work in `ready()` (it runs at import time, can fire during test discovery, and breaks `--check`-style invocations). The command must be invoked manually after `migrate` — that invocation is added to the README in Task 6.
  - [x] 3.4 Do **not** introduce a top-level constant or enum elsewhere (e.g., `fieldmark/roles.py`) for the role names. Story 1.12 (`authz.Can` primitive) is the right place to introduce a typed `Role` value object across the stack. At Story 1.8 the names live as a tuple inside this command — mirrors the .NET stack's Story 1.7 decision to keep role names in `RoleSeeder.cs` until Story 1.12.

- [x] Task 4: Run the seed command and verify idempotence (AC: #4)
  - [x] 4.1 From `fieldmark_py/`: `uv run python manage.py seed_groups`. Expected stdout: five `created: <NAME>` lines.
  - [x] 4.2 Re-run the same command. Expected stdout: five `exists: <NAME>` lines, exit 0.
  - [x] 4.3 Verify via `psql`: `SELECT id, name FROM django_auth.auth_group ORDER BY name;` — exactly five rows, names match the canonical list, IDs unchanged between runs (capture before and after).

- [x] Task 5: Add a pytest test for the seed command (AC: #4, #7)
  - [x] 5.1 Create `fieldmark_py/tools/tests/__init__.py` (empty) and `fieldmark_py/tools/tests/test_seed_groups.py`.
  - [x] 5.2 Add the `tools` app to `pytest.ini`'s `testpaths` list (line 5–12). The current list omits `tools` — add it after `grid`.
  - [x] 5.3 Test pattern (uses `pytest-django`'s real-Postgres fixture per the project's testing rule — no SQLite):

    ```python
    """Integration tests for the seed_groups management command."""

    import pytest
    from django.contrib.auth.models import Group
    from django.core.management import call_command

    CANONICAL = {"ADMIN", "COMPLIANCE_OFFICER", "INSPECTOR", "SITE_SUPERVISOR", "EXECUTIVE"}


    @pytest.mark.django_db
    def test_seed_groups_creates_five_canonical_groups():
        call_command("seed_groups")
        names = set(Group.objects.values_list("name", flat=True))
        assert CANONICAL <= names
        assert Group.objects.filter(name__in=CANONICAL).count() == 5


    @pytest.mark.django_db
    def test_seed_groups_is_idempotent():
        call_command("seed_groups")
        ids_first = {g.id for g in Group.objects.filter(name__in=CANONICAL)}
        call_command("seed_groups")
        ids_second = {g.id for g in Group.objects.filter(name__in=CANONICAL)}
        assert ids_first == ids_second
        assert Group.objects.filter(name__in=CANONICAL).count() == 5
    ```

  - [x] 5.4 Run `uv run pytest tools/tests/test_seed_groups.py -v`. Both tests pass.

- [x] Task 6: Update `fieldmark_py/CLAUDE.md` and `fieldmark_py/README.md` (AC: #8)
  - [x] 6.1 Add a new `## Authentication` section to `fieldmark_py/CLAUDE.md` (place it after the `## Hard Rules (Django-specific)` section). Content covers:
    - Django's built-in `auth`/`sessions`/`admin`/`contenttypes` are the framework-native auth source. No custom user model.
    - The `django_auth` schema is the target for all framework-managed tables. The mechanism is **`OPTIONS["options"] = "-c search_path=django_auth,public"`** on the default DATABASES entry — not a custom `DatabaseRouter`, not `db_table` overrides. State that the architecture's reference to a `routers.py` (directory structure line 1122) was superseded by the simpler `search_path` approach, and that this divergence is intentional.
    - Domain models are `Meta.managed = False` with `db_table = 'domain"."<table>'` so Django never CREATEs `domain.*` tables (re-state ADR-014).
    - Conceptual roles map to Django Groups. The five canonical group names are: `ADMIN`, `COMPLIANCE_OFFICER`, `INSPECTOR`, `SITE_SUPERVISOR`, `EXECUTIVE`. They are seeded via `python manage.py seed_groups` (idempotent). The command lives in `tools/management/commands/seed_groups.py`.
    - Login, logout, and the unauthenticated-redirect contract land in Story 1.11.
  - [x] 6.2 Update `fieldmark_py/README.md` "Getting Started" (step 3 at line 89–93). After `uv run python manage.py migrate`, add a step:
    ```bash
    uv run python manage.py seed_groups
    ```
    with a one-line explanation that this seeds the five conceptual-role Groups (idempotent — safe to re-run).
  - [x] 6.3 The existing README "Database & Migration Ownership" section (line 168–179) is correct as written. No edit needed there.

- [x] Task 7: Verify parity, lint, and type checks (AC: #6, #7)
  - [x] 7.1 From repo root: `make parity` — exits 0. Routes still identical; `pg_indexes` for `domain.*` unchanged.
  - [x] 7.2 From `fieldmark_py/`: `uv run python manage.py dump_routes` — output unchanged from HEAD-before-this-story (capture both and diff).
  - [x] 7.3 From `fieldmark_py/`: `uv run ruff check .` — zero issues.
  - [x] 7.4 From `fieldmark_py/`: `uv run mypy .` — zero errors.
  - [x] 7.5 From `fieldmark_py/`: `uv run pytest` — all tests pass, including the two new tests from Task 5.

## Dev Notes

### Brownfield posture — what exists today (read before writing anything)

State of the Django stack at HEAD of this branch:

- `fieldmark_py/fieldmark/settings.py` (lines 36–61) already has `django.contrib.auth`, `django.contrib.sessions`, `django.contrib.admin`, `django.contrib.contenttypes`, and `django.contrib.messages` in `INSTALLED_APPS`, plus the corresponding middleware. **You do not add or remove apps in this story.**
- `fieldmark_py/fieldmark/settings.py` (lines 99–117) already configures `DATABASES["default"]["OPTIONS"]["options"] = "-c search_path=django_auth,public"`. The comment on lines 108–111 already explains it. This is the AC #1 mechanism, in place. Your job is to verify it works end-to-end and expand the comment slightly (Task 1.2).
- No `routers.py` file exists in `fieldmark/`. Don't create one (Task 1.3).
- All app `models.py` files (`projects/models.py`, `inspections/models.py`, etc.) are essentially empty (`# Create your models here.`). Domain models land in Epic 2+. This is **safe** for Story 1.8 — there is nothing for Django to accidentally migrate into `domain.*`.
- All app `migrations/` folders contain only `__init__.py`. Same reason.
- `fieldmark_py/tools/` is an existing INSTALLED_APP that hosts `tools/management/commands/dump_routes.py` (Story 1.3). It has no `tests/` folder yet — Task 5 creates one and adds it to `pytest.ini`'s `testpaths`.
- `fieldmark_py/pytest.ini` lists `projects`, `inspections`, `violations`, `audit`, `compliance`, `reference`, `grid` as `testpaths` — **`tools` is missing**. Task 5.2 adds it.
- `fieldmark_py/pyproject.toml` has `pytest-django>=4.12.0` in the dev deps. `@pytest.mark.django_db` is available out of the box.
- The `dotnet_auth` schema has been populated by Story 1.7 (in `review` status). The `django_auth` schema currently has zero tables (verified at HEAD). Story 1.8 populates it.
- `tools/parity/canonical-pg-indexes.txt` filters `WHERE schemaname='domain'` — auth schemas are intentionally out of scope. Don't edit this file.

### Why `search_path` instead of a `DatabaseRouter`

The architecture document's directory structure (line 1122) describes a `fieldmark/routers.py` that targets auth tables at `django_auth`. The implementation chose a simpler equivalent — set `search_path=django_auth,public` on the DB connection's `OPTIONS`. This produces the same outcome (all unqualified `CREATE TABLE`s land in `django_auth`) with strictly less code.

Trade-offs:

| Approach | Lines of code | Per-app boilerplate | Covers `admin`/`sessions`/`contenttypes`/`migrations`? |
|---|---|---|---|
| `db_table` overrides on each auth model | ~40+ | Yes (every framework model) | No — can't override Django's `django_migrations` table |
| Custom `DatabaseRouter` | ~30 | No, but new file | Yes |
| `search_path` on connection | 1 setting line | No | Yes (automatic — Postgres feature) |

The `search_path` approach is the cleanest of the three, but it diverges from the architecture diagram. Document the divergence in `CLAUDE.md` (Task 6.1) so a later reader doesn't add a redundant `routers.py` thinking it's missing.

The `search_path` approach is also why `django_content_type` and `django_migrations` end up in `django_auth` — they are unqualified `CREATE TABLE`s emitted by Django itself, so they follow the same routing. This is intentional and correct: it keeps the `public` schema empty and makes the schema isolation contract crisp ("every framework-owned table is in a `*_auth` schema").

### Why a management command, not a startup hook or migration data op

Three rejected alternatives:

1. **Calling the seeder from `AppConfig.ready()`** — Django explicitly warns against DB work in `ready()` (it runs at import time, including during `manage.py check`, `makemigrations`, test discovery, etc.). It would also race with `migrate` itself on a fresh database.
2. **A data migration in `tools/migrations/0001_seed_groups.py`** — would couple seed data to schema history. Re-running the seed (e.g., after a manual `DELETE` to test idempotence) would require a `--fake` dance. Also, `tools` has no models, so creating a migrations folder there is awkward.
3. **A `post_migrate` signal handler** — would violate the project's `## Hard Rules (Django-specific)` in `fieldmark_py/CLAUDE.md`: "No Django signals — not for business logic, not for side effects, not ever without an ADR."

A management command invoked manually after `migrate` is what the architecture spec already prescribes for the related `seed_dev_users.py` (Architecture line 1140 / Story 1.10). Story 1.8 follows the same pattern.

### Why Groups (not Permissions, not a custom Role model)

Django's `Group` is the canonical mechanism for "user has role X". Permissions are finer-grained (model-level CRUD bits) and not what FieldMark's conceptual roles model — roles are coarse role labels that the `authz.Can` primitive (Story 1.12) consults to decide whether actions are permitted on specific entities. The mapping is therefore Group-name → role; no custom `Role` model is needed at this story or any future story.

Django's `AbstractUser` is **not** subclassed. The architecture explicitly says "built-in `auth` system, no custom user model" (Architecture D7) — once you subclass, every later migration becomes harder and the conceptual-parity story with the .NET `IdentityUser<Guid>` weakens. We accept Django's default `User` model (integer PK, username/email auth). The cross-stack same-UUID work happens at the `domain.audit_entry.actor_id` layer (opaque UUID, ADR-012) and via the dev-user manifest (Story 1.10), not by changing the user PK type.

### Migration application sequence on a fresh database

After `make reset`, the first `migrate` run does this in order (Django decides dependency order — listed for orientation, not for hand-execution):

1. `contenttypes.0001_initial` → creates `django_auth.django_content_type`
2. `auth.0001_initial` → creates `django_auth.auth_user`, `auth_group`, `auth_permission`, `auth_user_groups`, `auth_user_user_permissions`, `auth_group_permissions`
3. `admin.0001_initial` → creates `django_auth.django_admin_log`
4. `sessions.0001_initial` → creates `django_auth.django_session`
5. A series of `auth.0002` through `auth.0012` alter migrations (column-length tweaks, validator updates) — all target the already-created tables in `django_auth`.

`django_auth.django_migrations` is the bookkeeping table; it is created implicitly before any migration runs.

You do **not** need to run `makemigrations` for any FieldMark app in this story — no models are being added.

### Idempotent seeding pattern (exact code, copy-ready)

See Task 3.2 for the canonical implementation. The key property is `Group.objects.get_or_create(name=name)` — Django uses a `SELECT … WHERE name=…` followed by `INSERT` if absent, all wrapped by Django's atomic save. `Group.name` has a `unique=True` constraint, so a race condition between two concurrent `seed_groups` runs would surface as an `IntegrityError` on one of them — that is acceptable behavior for a dev-time bootstrap command (it won't happen in normal use).

Print the `created`/`exists` distinction so the developer can visually confirm idempotence on the second run. Do **not** silently no-op — invisible side effects are harder to debug than verbose ones.

### `--dump-routes` invariant

`tools/management/commands/dump_routes.py` (Story 1.3) does not touch the database — it walks `get_resolver()` purely against `urls.py`. It already excludes `/admin/` (lines 50–54). Story 1.8 adds **no** routes, so the dump output is byte-identical to HEAD-before-this-story. Verify this in Task 7.2 by capturing both outputs and diffing.

`make parity` (Story 1.3) compares three dumps + `pg_indexes` for `domain.*`. Story 1.8 changes neither, so parity stays clean.

### Anti-patterns that must NOT slip in

- ❌ Subclassing `AbstractUser` or `AbstractBaseUser`. Architecture D7 forbids a custom user model.
- ❌ Adding a `post_migrate` signal handler to seed Groups. Violates `fieldmark_py/CLAUDE.md` "No Django signals — not ever without an ADR."
- ❌ Adding `path("login/", ...)` or `path("logout/", ...)` to `fieldmark/urls.py`. That is Story 1.11's scope; doing it here breaks the parity invariant (.NET and Go don't have these routes yet) and pre-empts the Basecoat login form design.
- ❌ Creating a `fieldmark/routers.py` because the architecture's directory diagram shows one. The `search_path` approach supersedes it; document the divergence in `CLAUDE.md`, don't add the file.
- ❌ Putting `seed_groups` under `projects/management/commands/`. That namespace is reserved for `seed_dev_users.py` (Story 1.10). Story 1.8's command is cross-cutting bootstrap — it lives under `tools/`.
- ❌ Introducing a `fieldmark/roles.py` constants module or a `Role` enum in `compliance/` to share role names. Story 1.12 designs the `authz.Can` primitive across the stack; introducing a leaked constant now would create churn.
- ❌ Editing `tools/parity/canonical-pg-indexes.txt`. It snapshots `domain.*` indexes only; auth schemas are intentionally out of scope.
- ❌ Setting `Group.objects.create(...)` (would raise `IntegrityError` on the second run) instead of `get_or_create(...)`.
- ❌ Calling `seed_groups` from `AppConfig.ready()`. Same reason as the signals ban: implicit DB work at import time.
- ❌ Editing `AUTH_PASSWORD_VALIDATORS` to mirror the .NET stack's `RequireDigit=true, RequiredLength=10, …` policy from Story 1.7. The epic AC for Story 1.8 does **not** require password-policy parity, and PRD §NFR Security only requires "framework-native salted" hashing. Story 1.11 (login/logout) is the natural place to revisit policy if cross-stack parity becomes user-facing.
- ❌ Using SQLite for the new pytest tests. The project rule (root `CLAUDE.md` → `docs/hard-rules.md`) is real PostgreSQL via `pytest-django`'s `django_db` fixture. The connection uses the same `FIELDMARK_DATABASE_URL`.

### Project Structure Notes

Files this story adds or modifies:

- **New:** `fieldmark_py/tools/management/commands/seed_groups.py`
- **New:** `fieldmark_py/tools/tests/__init__.py`
- **New:** `fieldmark_py/tools/tests/test_seed_groups.py`
- **Update:** `fieldmark_py/fieldmark/settings.py` — expand the inline comment around lines 108–111 (no functional change).
- **Update:** `fieldmark_py/pytest.ini` — add `tools` to `testpaths`.
- **Update:** `fieldmark_py/CLAUDE.md` — add a new `## Authentication` section.
- **Update:** `fieldmark_py/README.md` — add `seed_groups` to "Getting Started".

No file in `fieldmark/`, in any aggregate app, or in `docker/` is otherwise modified. All file locations align with the architecture's `fieldmark_py/` tree (lines 1109–1166) modulo the deliberate `routers.py` divergence (documented).

### Testing Standards

Per Architecture §Testing and the Django stack `CLAUDE.md`:

- **Real PostgreSQL only** — never SQLite. `@pytest.mark.django_db` on the new tests uses the configured `DATABASES["default"]`, which points at the running local Postgres via `FIELDMARK_DATABASE_URL`.
- The two new tests (Task 5.3) cover (a) the command creates five Groups and (b) re-running is idempotent. Together they satisfy AC #4's verification surface.
- The `pytest-django` test runner will create and tear down a test database. `search_path` carries over from `settings.py` to the test DB, so auth tables in the test DB also land in `django_auth` — no special handling required.
- **No login-flow tests in this story.** Login/logout integration tests land in Story 1.11. Adding speculative integration tests now would lock in a Razor-style assumption before Django's view shape is designed.
- **No domain-model tests.** No domain models exist yet on the Django side.

### Previous Story Intelligence

**Story 1.7 (.NET — currently in `review`)** is the structural analog of this story for the .NET stack. Lessons that transfer directly:

- **Cross-stack symmetry argument is real.** The .NET stack picked a separate `AuthDbContext` to keep migration ownership clean (`Migrations/Auth/` only). Django achieves the same property differently: built-in `auth` migrations are framework-owned and don't touch `domain.*` because (a) `search_path` puts them in `django_auth` and (b) no `domain.*` models exist that Django could try to migrate. The destination is the same; the mechanism is per-stack idiomatic.
- **Idempotence is a hard requirement, not a nice-to-have.** Story 1.7 used `RoleManager.RoleExistsAsync(name)` to short-circuit. Django uses `Group.objects.get_or_create(name=name)`. Both produce the "five rows, no duplicates on re-run" property.
- **No UI in this story.** Story 1.7 explicitly used `AddIdentityCore` (not `AddDefaultIdentity`) to avoid scaffolding the `/Identity/Account/*` Razor pages. Django's equivalent is "don't add `path("login/", ...)` to `urls.py` and don't add `LoginRequiredMiddleware`." The parity invariant is binding for both stacks.
- **Role names live in the seeder, not in domain.** Story 1.7 deliberately did not create a `Role` enum in `FieldMark.Domain` (Domain has zero outbound references). Story 1.8 makes the same call for Django — no `fieldmark/roles.py` constants module. Story 1.12 designs the typed `Role` value object across all stacks at once.
- **`--dump-routes` early-return is sacred.** Story 1.3 fixed both .NET and Go to not touch the DB before the dump-routes flag check. The Django `dump_routes` command walks `urls.py` and does not touch the DB at all — already correct, but do not regress it (don't import a model at command-import time).

**Story 1.4 (design system — in `review`)** is unrelated to this story's surface. No direct dependency.

### Git Intelligence

Recent commits (most relevant to this story):

- `d03f0fe feat: e1s3 establish tools parity` — added `tools/management/commands/dump_routes.py` (the command Story 1.8's `seed_groups.py` will sit next to in the same package). Established the `tools` app as the home for cross-cutting management commands.
- `cbf47e9 feat: e1s2 verified sql init scripts` — confirmed `docker/postgres/init/001_schemas.sql` creates the `django_auth` schema. Story 1.8 populates it.
- `a6fac88 feat: e1s1 confirm scaffolds` — the Django scaffold this story extends. INSTALLED_APPS, MIDDLEWARE, and the search_path-on-OPTIONS setup all ship from this commit; Story 1.8 verifies them end-to-end and adds the seed command.

No prior commit has added Django Groups, run `migrate` against `django_auth`, or added a management command other than `dump_routes`. This story is the first.

### Latest Technical Information

- **Django 6.0.4** is in use (`pyproject.toml` line 8). The `django.contrib.auth` API surface used by this story (`Group.objects.get_or_create`, `BaseCommand`, `call_command`) is stable since Django 1.x — no version-specific footguns.
- **psycopg 3.3.3+** (`pyproject.toml` line 9). The `OPTIONS["options"]` mechanism for setting `search_path` is supported identically in psycopg 2 and 3.
- **Python 3.14+** is required (`pyproject.toml` line 6). The type hints used in the seed command (`list[str]`, etc.) are Python 3.9+ syntax — no issue.
- **pytest-django 4.12+** is pinned. `@pytest.mark.django_db` does the right thing against real Postgres via `DATABASES["default"]`.
- No new packages need to be added. `django.contrib.auth` is bundled with Django; no `dj-rest-auth`, no `django-allauth`, no `djoser` — Story 1.11 will use Django's bundled `LoginView`/`LogoutView` (or a hand-authored equivalent) without third-party packages.
- **Postgres 17** is the target. `search_path` semantics have been stable since PostgreSQL 8.x — no version concerns.

### References

- [Architecture: Authentication & Security → D7](_bmad-output/planning-artifacts/architecture.md#authentication--security) — Django auth via built-in app; tables in `django_auth`; Groups for conceptual roles; seed five Groups on first migration.
- [Architecture: Schema ownership](_bmad-output/planning-artifacts/architecture.md#core-architectural-decisions) — `django_auth` schema is framework-local; no FKs from `domain.*` to auth schemas.
- [Architecture: Repository Directory Structure → `fieldmark_py/`](_bmad-output/planning-artifacts/architecture.md#complete-repository-directory-structure) — file locations; the prescribed `routers.py` is intentionally superseded by `search_path` (documented in `CLAUDE.md` per Task 6.1).
- [Architecture: Architectural Boundaries → Authentication / authorization](_bmad-output/planning-artifacts/architecture.md#architectural-boundaries) — opaque UUID refs in `domain.*`; auth implementation is per-stack-idiomatic.
- [docs/hard-rules.md](docs/hard-rules.md) — backend authority, infrastructure-owned domain schema, real PostgreSQL in tests, no service layers.
- [fieldmark_py/CLAUDE.md](fieldmark_py/CLAUDE.md) — Django-specific rules; no signals; no business logic in views/middleware; migrations scoped to `django_auth` only.
- [PRD FR1–FR8 — Authentication & Authorization](_bmad-output/planning-artifacts/prd/functional-requirements.md) — framework-local authentication; conceptual roles.
- [PRD architectural-constraints-prd-binding.md (ADR-012)](_bmad-output/planning-artifacts/prd/architectural-constraints-prd-binding.md) — schema isolation contract; per-stack auth ownership.
- [Story 1.3 implementation artifact](_bmad-output/implementation-artifacts/1-3-establish-tools-parity-and-make-parity-with-per-stack-dump-routes.md) — `--dump-routes` invariants; `make parity` contract.
- [Story 1.7 implementation artifact](_bmad-output/implementation-artifacts/1-7-wire-asp-net-core-identity-to-dotnet-auth-schema-with-conceptual-roles.md) — structural analog (.NET counterpart); idempotent role seeding pattern; no UI in this story.
- [docker/postgres/init/001_schemas.sql](docker/postgres/init/001_schemas.sql) — `django_auth` schema already created.

## Dev Agent Record

### Agent Model Used

claude-sonnet-4-6

### Debug Log References

No blockers encountered. All tasks executed cleanly in a single pass.

### Completion Notes List

- **Task 1**: Confirmed `search_path=django_auth,public` already present in `settings.py`. Expanded the OPTIONS comment from 3 lines to 7 lines covering Architecture D7, the `routers.py` supersession, and the Story 1.11 scope boundary.
- **Task 2**: `make reset` + `migrate` confirmed 10 tables in `django_auth` (8 required + `django_content_type` + `django_migrations`), 12 tables in `domain.*`, 0 tables in `public.*`, and only framework apps in `django_migrations`. Zero `domain` grep matches in app migrations.
- **Task 3**: Created `seed_groups.py` in `tools/management/commands/` using `Group.objects.get_or_create`. No `apps.py` wiring, no `roles.py` module.
- **Task 4**: First run produced 5 `created:` lines; second run produced 5 `exists:` lines. psql confirmed 5 rows with stable IDs.
- **Task 5**: Created `tools/tests/__init__.py` and `test_seed_groups.py`. Added `tools` to `pytest.ini` testpaths. Both tests pass against real PostgreSQL.
- **Task 6**: Added `## Authentication` section to `fieldmark_py/CLAUDE.md` documenting the `search_path` mechanism, `routers.py` divergence, ADR-014 domain isolation, role Groups, and Story 1.11 scope. Added `seed_groups` step (step 4) to README "Getting Started".
- **Task 7**: `make parity` exits 0 (4 routes, 21 indexes). `dump_routes` output unchanged (4 routes, no auth routes added). `ruff check .` clean. `mypy .` 0 errors. `pytest` 2 passed.

### File List

- **New:** `fieldmark_py/tools/management/commands/seed_groups.py`
- **New:** `fieldmark_py/tools/tests/__init__.py`
- **New:** `fieldmark_py/tools/tests/test_seed_groups.py`
- **Updated:** `fieldmark_py/fieldmark/settings.py` — expanded OPTIONS comment (no functional change)
- **Updated:** `fieldmark_py/pytest.ini` — added `tools` to testpaths
- **Updated:** `fieldmark_py/CLAUDE.md` — added `## Authentication` section
- **Updated:** `fieldmark_py/README.md` — added `seed_groups` step to Getting Started

## Change Log

- 2026-05-19: Story 1.8 implemented — verified `django_auth` schema placement via `search_path`, created idempotent `seed_groups` management command, added pytest integration tests (real PostgreSQL), updated CLAUDE.md and README. All ACs satisfied; `make parity` exits 0.
