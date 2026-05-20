# Story 1.10: Author shared UUID dev-user manifest and per-stack idempotent seed runners

Status: ready-for-dev

## Story

As a developer running cross-stack scenarios,
I want every stack's dev users to share identical UUIDs sourced from one canonical manifest,
So that audit comparison and cross-stack E2E parity tests can assert on actor identity (and on `domain.audit_entry.actor_id` rows) without translation tables.

## Acceptance Criteria

1. **The canonical manifest exists at `docker/postgres/init/seed-uuids/dev-users.json` and contains exactly six entries.** Each entry is a JSON object with the following keys (Architecture §Important Gaps item 8 / line 1450 — finalize the recommended shape):
   - `id`            — string, canonical user UUID. UUIDv7 preferred (time-ordered, debug-friendly). Pre-generated and committed; the file is the source of truth.
   - `username`      — string, lowercase ASCII, unique. Used as the per-stack login handle and the `X-FieldMark-Actor` cookie/header value (Story 1.9).
   - `display_name`  — string, human-readable display label (e.g., `"Marisol Vega"`).
   - `password`      — string, initial password meeting the .NET Identity policy from Architecture D6 (≥ 10 chars, ≥ 1 digit, ≥ 1 lowercase, ≥ 1 uppercase; non-alphanumeric optional). Same plaintext used across all three stacks at first seed so the dev can sign in with the same credentials on every stack.
   - `role`          — string, exactly one of the five canonical conceptual roles (`ADMIN`, `COMPLIANCE_OFFICER`, `INSPECTOR`, `SITE_SUPERVISOR`, `EXECUTIVE`) — **or** `null` for the no-role test user.

   The six entries are:

   | username | display_name | role |
   |---|---|---|
   | `marisol` | Marisol Vega | `COMPLIANCE_OFFICER` |
   | `pat` | Pat Smith | `SITE_SUPERVISOR` |
   | `aisha` | Aisha Patel | `ADMIN` |
   | `ravi` | Ravi Kumar | `INSPECTOR` |
   | `kenji` | Kenji Tanaka | `EXECUTIVE` |
   | `testuser` | Test User (no role) | `null` |

   Display names listed above are recommendations; the dev may adjust to taste, but the **set of six personas, usernames, and role assignments is fixed** (the names `marisol`, `pat`, `aisha`, `ravi`, `kenji` are referenced in epic AC, Story 1.9, Story 5.5, Story 6.4, and `domain-model.md` persona narratives — changing them would orphan downstream references).

2. **The manifest is valid JSON, schema-stable, and easy to read.** Specifically:
   - File is pretty-printed with two-space indent (matches `fieldmark_shared/package.json` and `e2e/package.json` conventions).
   - Top-level shape is `{"users": [ {...}, {...}, ... ]}` — an object with a `users` array, not a bare array. This leaves room for additive top-level metadata (e.g., `schema_version`) without breaking parsers.
   - File begins with a top-level `"$schema"` key referencing a sibling `dev-users.schema.json` (JSON Schema draft 2020-12). The schema file is committed alongside and validates: required keys per entry, `id` matches the UUID regex `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, `role` is `null` or one of the five canonical names, `username` matches `^[a-z][a-z0-9_-]{1,63}$`.
   - The six entries' `id` UUIDs are **distinct, lowercase, and dash-formatted**. Generate UUIDv7s via any reliable tool (e.g., `uuidv7-cli`, online UUIDv7 generator, or a one-shot `python -c "import uuid_extensions; print(uuid_extensions.uuid7())"`). Once generated and committed, **never rotate them** without coordinating across all three seeders simultaneously — these IDs become the canonical `domain.audit_entry.actor_id` values from Epic 2 onward.

3. **The .NET seeder `FieldMark/FieldMark.Web/SeedData/DevUsersSeeder.cs` exists and runs as part of the application startup path.** Specifically:
   - Lives in the existing `SeedData/` directory created by Story 1.7 (`RoleSeeder.cs` is its sibling).
   - Exposes a static async method `SeedAsync(IServiceProvider services, IWebHostEnvironment env, CancellationToken ct)` returning `Task`.
   - Reads the manifest from disk. Path resolution: combine `env.ContentRootPath` with `"../../docker/postgres/init/seed-uuids/dev-users.json"` (the .NET project content root is `FieldMark/FieldMark.Web/`, three levels deep relative to the repo root's `docker/`). Use `Path.GetFullPath` to normalize. If the file is missing, throw an `InvalidOperationException` with a clear message — do not silently no-op.
   - For each manifest entry:
     - Resolve the user via `UserManager<IdentityUser<Guid>>.FindByIdAsync(entry.Id.ToString())`. If present, **skip create**; if absent, call `userManager.CreateAsync(new IdentityUser<Guid> { Id = entry.Id, UserName = entry.Username, NormalizedUserName = entry.Username.ToUpperInvariant(), Email = $"{entry.Username}@fieldmark.local", NormalizedEmail = ..., EmailConfirmed = true, SecurityStamp = Guid.NewGuid().ToString() }, entry.Password)`. `CreateAsync` invokes `IPasswordHasher<IdentityUser<Guid>>` automatically — the seeder must not call the hasher directly (Architecture D6: framework-native password hashing).
     - If `role` is non-null: call `userManager.AddToRoleAsync(user, entry.Role)` if `!await userManager.IsInRoleAsync(user, entry.Role)`. The role records were seeded by Story 1.7's `RoleSeeder` and must exist; if `AddToRoleAsync` returns a failed `IdentityResult`, throw with the joined error description.
   - Idempotence: re-running the seeder against an already-seeded database produces zero new rows, zero errors, and zero duplicate `user_roles`. Verified by capturing `SELECT COUNT(*) FROM dotnet_auth.users` and `SELECT COUNT(*) FROM dotnet_auth.user_roles` before and after a second run — identical counts.
   - Wired into `Program.cs` immediately after the Story 1.7 `RoleSeeder.SeedAsync(...)` invocation, inside the same `IServiceScope` block. The dev-users seeder depends on the role records existing — the call order is **mandatory** (roles first, users second).

4. **The Django seeder `fieldmark_py/projects/management/commands/seed_dev_users.py` exists and is idempotent.** Specifically:
   - Architecture line 1140 places this command under `projects/management/commands/` (**not** `tools/`). Story 1.8 reserved this location explicitly. Do **not** put it under `tools/management/commands/`.
   - Invoked as `uv run python manage.py seed_dev_users` (no flags required).
   - Reads `docker/postgres/init/seed-uuids/dev-users.json` resolved relative to `settings.BASE_DIR` (which is `fieldmark_py/`). Path: `settings.BASE_DIR.parent / "docker" / "postgres" / "init" / "seed-uuids" / "dev-users.json"`. If the file is missing, raise `CommandError` with a clear message.
   - **UUID storage approach — chosen and documented:** Django's `auth_user.id` is a `BIGSERIAL` AutoField and the project does not use a custom user model (Story 1.8 D7). The manifest UUIDs therefore **cannot** be `auth_user` primary keys. The chosen approach is **a side table `django_auth.dev_user_uuid` mapping `user_id → uuid`** (one row per user, FK to `auth_user.id` with `ON DELETE CASCADE`, `UNIQUE(uuid)`). Rationale: (a) preserves Django's idiomatic integer PK and out-of-the-box admin/auth ergonomics, (b) keeps human-readable `username` values, (c) provides the UUID lookup that future `domain.audit_entry.actor_id` writes need, (d) the side-table lives in `django_auth` (framework-local — ADR-012). **Do not** use `username=<uuid>` (destroys readable usernames) and **do not** add a `uuid` column to `auth_user` via a custom migration (alters framework table beyond what's necessary). The command's module docstring records this choice and why; `fieldmark_py/CLAUDE.md` gains an `## Authentication / User UUIDs` note pointing at the side table.
   - The side table is created via a Django migration in a small new app — recommendation: add it to the existing `tools` app (`tools/migrations/0001_dev_user_uuid.py`) since it's cross-cutting and `tools` already exists (Story 1.3 / Story 1.8). Set `Meta.managed = True` and use `db_table = 'django_auth"."dev_user_uuid'` so the migration creates the table in `django_auth`. Verify via `\dt django_auth.dev_user_uuid` after `make reset && uv run python manage.py migrate`.
   - For each manifest entry, the command:
     - Uses `User.objects.update_or_create(username=entry.username, defaults={"first_name": entry.display_name.split()[0], "last_name": " ".join(entry.display_name.split()[1:]), "email": f"{entry.username}@fieldmark.local", "is_active": True})`.
     - Calls `user.set_password(entry.password)` and `user.save()` — `set_password` invokes Django's configured `PASSWORD_HASHERS` (PBKDF2 by default) per PRD §NFR Security.
     - Calls `DevUserUuid.objects.update_or_create(user_id=user.pk, defaults={"uuid": entry.id})` to record the canonical UUID mapping.
     - If `role` is non-null: `user.groups.set([Group.objects.get(name=entry.role)])` — clearing prior groups so re-running with a changed role updates cleanly. The five Groups are seeded by Story 1.8's `seed_groups` and must exist; if a `Group.DoesNotExist` is raised, surface it as a `CommandError` with a hint to run `seed_groups` first.
   - Idempotence: second run produces zero new `auth_user` rows, zero new `dev_user_uuid` rows, zero new `auth_user_groups` rows. Verified by row-count comparison.
   - The command prints one INFO-level log line per user (`created` / `updated`) and a final summary (`seed_dev_users: 6 users, 0 errors`).

5. **The Go seeder `fieldmark-go/cmd/seed/main.go` exists and is idempotent.** Specifically:
   - Invoked as `go run ./cmd/seed` from `fieldmark-go/` (no flags). Documented in `fieldmark-go/README.md` Getting Started after the Story 1.9 `migrate-fiber-auth` step.
   - Reads `FIELDMARK_DATABASE_URL` (defaulting to `postgres://fieldmark:fieldmark@localhost:5432/fieldmark` — same default as `cmd/web/main.go` and `cmd/migrate-fiber-auth/main.go`).
   - Manifest path resolution: walks up from the executable's working directory (typically `fieldmark-go/`) to find `../docker/postgres/init/seed-uuids/dev-users.json`. Use `filepath.Abs` + a small `findManifest()` helper that ascends until it finds the file or fails with a clear error. **Do not** embed the manifest into the Go binary — the manifest is shared cross-stack and must live as a single source of truth in `docker/postgres/init/seed-uuids/`.
   - Opens a `pgxpool.Pool` (Story 1.9 standardized on pgxpool), starts a single transaction, and for each manifest entry executes:

     ```sql
     INSERT INTO fiber_auth.users (id, username, display_name)
         VALUES ($1, $2, $3)
         ON CONFLICT (id) DO UPDATE
            SET username     = EXCLUDED.username,
                display_name = EXCLUDED.display_name;

     -- Only if role is non-null:
     INSERT INTO fiber_auth.user_roles (user_id, role)
         VALUES ($1, $2)
         ON CONFLICT (user_id, role) DO NOTHING;

     -- Defensively prune any stale role assignment that no longer matches the manifest:
     DELETE FROM fiber_auth.user_roles
       WHERE user_id = $1
         AND role <> $2;
     ```

     (For the no-role test user, only the `INSERT INTO fiber_auth.users` runs; both `user_roles` statements are skipped.)
   - **No password storage in `fiber_auth.users`.** The Go stack uses stub auth (ADR-012, Story 1.9): identity is asserted via the `X-FieldMark-Actor` cookie/header carrying a username, not a password. The manifest's `password` field is read by the Go seeder but **not** persisted. Document this in the command's package comment so a future engineer doesn't assume password storage is missing accidentally — it's deliberately absent.
   - On successful commit: log `"seed_dev_users: 6 users, 0 errors"`, exit 0. On any error: roll back, print wrapped error to stderr, exit non-zero.
   - Idempotence: second run produces zero new rows in `fiber_auth.users` and `fiber_auth.user_roles`. Verified by row-count comparison.

6. **After all three seeders have run, the same UUID resolves to the same person on every stack.** Specifically — picking `marisol` as the canonical spot-check:

   ```sql
   -- .NET
   SELECT id, user_name FROM dotnet_auth.users WHERE user_name = 'marisol';
   -- Django
   SELECT u.uuid, au.username
     FROM django_auth.dev_user_uuid u
     JOIN django_auth.auth_user au ON au.id = u.user_id
    WHERE au.username = 'marisol';
   -- Go
   SELECT id, username FROM fiber_auth.users WHERE username = 'marisol';
   ```

   All three queries return the **identical UUID string** (case-insensitive — Postgres' `uuid` type normalizes case). Repeat the spot-check for `pat`, `aisha`, `ravi`, `kenji`, `testuser` — six matches, three stacks, zero mismatches. Capture the proof in the story's Completion Notes.

7. **No seeder writes to `domain.*` schema.** This is the cross-cutting invariant the epic AC names explicitly. Verify by:
   - `grep -rn 'domain\.' FieldMark/FieldMark.Web/SeedData/DevUsersSeeder.cs` — zero matches.
   - `grep -rn 'domain\.' fieldmark_py/projects/management/commands/seed_dev_users.py` — zero matches.
   - `grep -rn 'domain\.' fieldmark-go/cmd/seed/main.go` — zero matches.
   - `SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='domain' AND table_name LIKE '%user%'` returns 0 (no `domain.user`-shaped tables exist; `domain.audit_entry.actor_id` is the only UUID column that *references* users, and it has no FK — ADR-012).
   - Reference seed data (TradeType, ViolationCategory, ComplianceRule) remains owned by `020_domain_seed.sql` — none of the new seeders touches those tables.

8. **`make parity` exits 0 after this story lands.** The route inventory diff stays clean (the new seeders register **zero** new routes — they are command-line invocations, not HTTP handlers). The `tools/parity/canonical-pg-indexes.txt` snapshot for `domain.*` is unchanged — Story 1.10 touches only `dotnet_auth.users` / `dotnet_auth.user_roles`, `django_auth.auth_user` / `django_auth.auth_user_groups` / `django_auth.dev_user_uuid`, and `fiber_auth.users` / `fiber_auth.user_roles`. No `domain` DDL or DML.

9. **A single make target `make seed` runs all three seeders in sequence.** Add to the root `Makefile`:

   ```makefile
   .PHONY: seed seed-net seed-django seed-go

   seed: seed-net seed-django seed-go ## Seed dev users into all three stacks' auth schemas
   	@echo "✓ All three stacks seeded from docker/postgres/init/seed-uuids/dev-users.json"

   seed-net:
   	cd FieldMark && dotnet run --project FieldMark.Web -- --seed-dev-users

   seed-django:
   	cd fieldmark_py && uv run python manage.py seed_dev_users

   seed-go:
   	cd fieldmark-go && go run ./cmd/seed
   ```

   The `--seed-dev-users` flag on the .NET side is a new `Program.cs` short-circuit: when present, the application runs `RoleSeeder.SeedAsync` followed by `DevUsersSeeder.SeedAsync` and exits 0 instead of starting the web server. (The existing `RoleSeeder` call inside the running web app's startup path also still runs — `--seed-dev-users` is just an alternative entry that runs both seeders without binding to `:5000`.) Reasons for the flag: (a) `make seed` should not require a free port, (b) parallel symmetry with Django's `manage.py seed_dev_users` and Go's `go run ./cmd/seed`, (c) lets the dev re-seed without restarting a running web app.

10. **Each stack's README "Getting Started" gains a `seed` step at the appropriate point.** For all three stacks: after the existing "Run the application / Apply auth migrations" step, document the seed invocation. The root `README.md` references `make seed` once as the canonical cross-stack convenience. The Getting Started note explains that `make seed` requires all three stacks' auth schemas to exist first (Story 1.7's migrations applied for .NET, Story 1.8's `migrate` for Django, Story 1.9's `migrate-fiber-auth` for Go).

11. **Build / test / lint gates stay green on each stack.**
    - .NET: `cd FieldMark && dotnet build && dotnet test` — clean.
    - Django: `cd fieldmark_py && uv run ruff check . && uv run mypy . && uv run python -m pytest` — clean. Tests include at minimum: `seed_dev_users` command produces 6 users, second run produces 0 new rows (use `@pytest.mark.django_db` per `fieldmark_py/CLAUDE.md`; **no SQLite** — real Postgres).
    - Go: from `fieldmark-go/`, `make fmt-check && make vet && make staticcheck && make test` — clean. Tests cover the pure-Go manifest parsing (`parseManifest` taking an `io.Reader`); DB-touching coverage is integration-tagged (`//go:build integration`) per the Story 1.9 testing posture.

## Tasks / Subtasks

- [ ] Task 1: Generate and commit the canonical manifest (AC: #1, #2)
  - [ ] 1.1 Generate six UUIDv7 strings. Recommended tools:
    - macOS / Linux with Python: `uv tool run --from uuid-extensions python -c "from uuid_extensions import uuid7; print(uuid7())"` six times.
    - Or any UUIDv7 generator — the values are committed once and never regenerated.
  - [ ] 1.2 Create `docker/postgres/init/seed-uuids/` (does not exist yet — verify via `ls docker/postgres/init/` before creating).
  - [ ] 1.3 Author `docker/postgres/init/seed-uuids/dev-users.json` with the six entries per AC #1. Use the recommended display names unless you have a strong preference. Pick a password that satisfies the Identity policy (≥ 10 chars, digit, lowercase, uppercase) — e.g., `FieldMark!2026` — and use the **same plaintext for every user** at first seed; this is dev-only and is rotated post-MVP. Example skeleton:

    ```json
    {
      "$schema": "./dev-users.schema.json",
      "users": [
        {
          "id":           "01923456-7890-7abc-def0-123456789abc",
          "username":     "marisol",
          "display_name": "Marisol Vega",
          "password":     "FieldMark!2026",
          "role":         "COMPLIANCE_OFFICER"
        }
      ]
    }
    ```

  - [ ] 1.4 Author `docker/postgres/init/seed-uuids/dev-users.schema.json` (JSON Schema draft 2020-12) validating: `users` array of 6 objects, `id` UUID regex, `username` ASCII regex, `role` enum + nullable. Keep it minimal — about 40 lines.
  - [ ] 1.5 Add a `## Dev User Manifest` section to `docker/postgres/init/README.md` (create the file if missing) documenting: the manifest's purpose, the six personas, the password policy origin (.NET Identity), the rotation rule ("never rotate IDs once committed without coordinating all three seeders"), and a one-line reminder that this file is **read** by per-stack seeders but **not** applied directly to Postgres by the init scripts.

- [ ] Task 2: Implement the .NET dev-users seeder (AC: #3, #9)
  - [ ] 2.1 Confirm Story 1.7 has landed (or is at least merged enough to provide): `FieldMark/FieldMark.Web/SeedData/RoleSeeder.cs`, `FieldMark/FieldMark.Data/Context/AuthDbContext.cs`, ASP.NET Core Identity registered in `Program.cs` with `IdentityUser<Guid>` / `IdentityRole<Guid>`. If 1.7 has not yet merged to this branch, **stop and surface the dependency** — do not invent the Identity wiring.
  - [ ] 2.2 Create `FieldMark/FieldMark.Web/SeedData/DevUsersSeeder.cs`. Sketch:

    ```csharp
    using System.Text.Json;
    using FieldMark.Web.SeedData;
    using Microsoft.AspNetCore.Hosting;
    using Microsoft.AspNetCore.Identity;
    using Microsoft.Extensions.DependencyInjection;

    namespace FieldMark.Web.SeedData;

    public static class DevUsersSeeder
    {
        private sealed record ManifestEntry(
            Guid Id,
            string Username,
            string DisplayName,
            string Password,
            string? Role);

        private sealed record Manifest(List<ManifestEntry> Users);

        public static async Task SeedAsync(IServiceProvider services, IWebHostEnvironment env, CancellationToken ct)
        {
            var manifestPath = Path.GetFullPath(
                Path.Combine(env.ContentRootPath, "..", "..", "docker", "postgres", "init", "seed-uuids", "dev-users.json"));

            if (!File.Exists(manifestPath))
                throw new InvalidOperationException($"DevUsersSeeder: manifest not found at {manifestPath}");

            var json = await File.ReadAllTextAsync(manifestPath, ct);
            var manifest = JsonSerializer.Deserialize<Manifest>(json, new JsonSerializerOptions
            {
                PropertyNamingPolicy = JsonNamingPolicy.SnakeCaseLower,
                ReadCommentHandling  = JsonCommentHandling.Skip,
            }) ?? throw new InvalidOperationException("DevUsersSeeder: manifest parse returned null");

            var userManager = services.GetRequiredService<UserManager<IdentityUser<Guid>>>();

            foreach (var entry in manifest.Users)
            {
                var existing = await userManager.FindByIdAsync(entry.Id.ToString());
                if (existing is null)
                {
                    var user = new IdentityUser<Guid>
                    {
                        Id                   = entry.Id,
                        UserName             = entry.Username,
                        NormalizedUserName   = entry.Username.ToUpperInvariant(),
                        Email                = $"{entry.Username}@fieldmark.local",
                        NormalizedEmail      = $"{entry.Username}@fieldmark.local".ToUpperInvariant(),
                        EmailConfirmed       = true,
                        SecurityStamp        = Guid.NewGuid().ToString(),
                    };
                    var create = await userManager.CreateAsync(user, entry.Password);
                    if (!create.Succeeded)
                        throw new InvalidOperationException(
                            $"DevUsersSeeder: CreateAsync failed for {entry.Username}: {string.Join("; ", create.Errors.Select(e => e.Description))}");
                    existing = user;
                }

                if (entry.Role is not null && !await userManager.IsInRoleAsync(existing, entry.Role))
                {
                    var add = await userManager.AddToRoleAsync(existing, entry.Role);
                    if (!add.Succeeded)
                        throw new InvalidOperationException(
                            $"DevUsersSeeder: AddToRoleAsync({entry.Role}) failed for {entry.Username}: {string.Join("; ", add.Errors.Select(e => e.Description))}");
                }
            }
        }
    }
    ```

  - [ ] 2.3 Update `Program.cs`: inside the existing `using (var scope = app.Services.CreateScope()) { await RoleSeeder.SeedAsync(scope.ServiceProvider, ct); }` block (introduced by Story 1.7), add a second line: `await DevUsersSeeder.SeedAsync(scope.ServiceProvider, env, CancellationToken.None);`. The order is **roles first, then users** — `AddToRoleAsync` requires roles to exist.
  - [ ] 2.4 Add the `--seed-dev-users` command-line flag (AC #9). Sketch — placed near the top of `Program.cs` after `var builder = WebApplication.CreateBuilder(args);` and the DB connection-string setup, but **before** route registration:

    ```csharp
    if (args.Contains("--seed-dev-users"))
    {
        builder.Services.AddLogging();
        // ... register AuthDbContext + Identity here (factor into a shared method
        //     called from both the seed path and the web path)
        var app = builder.Build();
        using (var scope = app.Services.CreateScope())
        {
            await RoleSeeder.SeedAsync(scope.ServiceProvider, CancellationToken.None);
            await DevUsersSeeder.SeedAsync(scope.ServiceProvider, app.Environment, CancellationToken.None);
        }
        Console.WriteLine("✓ Roles and dev users seeded.");
        return;
    }
    ```

    Refactor the Identity/DbContext registration into a private static helper (`RegisterIdentityAndAuthDbContext(WebApplicationBuilder)`) so both paths share it without duplication. The flag must short-circuit **before** the web server binds to its listener port.
  - [ ] 2.5 Run `cd FieldMark && dotnet build` — clean. Then end-to-end:
    - `make reset` (from repo root) — destroys volume.
    - `cd FieldMark && dotnet ef database update --project FieldMark.Data --startup-project FieldMark.Web --context AuthDbContext` (Story 1.7's auth migrations).
    - `cd FieldMark && dotnet run --project FieldMark.Web -- --seed-dev-users` — expects clean exit with the success line.
    - Re-run the same command — expects clean exit with the same success line, **no exceptions**, and zero new rows (verify via psql).

- [ ] Task 3: Implement the Django side-table migration + seeder (AC: #4, #7)
  - [ ] 3.1 Confirm Story 1.8 has landed: `search_path=django_auth,public` in `settings.py`, Django `auth` migrations applied, `seed_groups` command exists and the five Groups are in `django_auth.auth_group`. If not, stop and surface the dependency.
  - [ ] 3.2 Create `fieldmark_py/tools/models.py` with the side-table model (if a `tools/models.py` already exists with other content, append):

    ```python
    """Cross-cutting tools models.

    DevUserUuid maps Django auth_user.id → the canonical UUID from
    docker/postgres/init/seed-uuids/dev-users.json. This lets domain.audit_entry
    rows reference users by UUID across all three stacks (cross-stack audit
    parity, ADR-012). Lives in django_auth schema because the side table is
    framework-local, never referenced from domain.*.
    """

    import uuid

    from django.contrib.auth import get_user_model
    from django.db import models


    class DevUserUuid(models.Model):
        user    = models.OneToOneField(get_user_model(), on_delete=models.CASCADE, related_name="dev_uuid")
        uuid    = models.UUIDField(unique=True, default=uuid.uuid4, editable=False)

        class Meta:
            db_table = 'django_auth"."dev_user_uuid'
            verbose_name = "dev user UUID mapping"
    ```

    Note the deliberate `'django_auth"."dev_user_uuid'` quoting — this is the Django idiom for cross-schema `db_table` (mirrors `domain-model.md` patterns referenced from `fieldmark_py/CLAUDE.md`).

  - [ ] 3.3 Generate the migration: `cd fieldmark_py && uv run python manage.py makemigrations tools`. Verify the migration file lands at `fieldmark_py/tools/migrations/0001_initial.py` (or next-available number) and only creates `django_auth.dev_user_uuid`. Run `uv run python manage.py migrate tools` and verify via psql: `\dt django_auth.dev_user_uuid` exists.
  - [ ] 3.4 Create `fieldmark_py/projects/management/commands/seed_dev_users.py`. Sketch:

    ```python
    """Seed dev users from the shared UUID manifest.

    Reads docker/postgres/init/seed-uuids/dev-users.json (UTF-8 JSON, schema in
    docker/postgres/init/seed-uuids/dev-users.schema.json) and writes six users
    into django_auth.auth_user, mapping each to its canonical UUID via the
    tools.DevUserUuid side table. Role assignment uses Django Groups
    (seeded by tools.management.commands.seed_groups in Story 1.8).

    Idempotence: update_or_create + Group.set(...) — re-running with no
    manifest changes produces zero net mutations. Side-table UUID approach
    documented in Story 1.10 AC #4 (auth_user.id is BIGSERIAL; the canonical
    UUID lives in django_auth.dev_user_uuid).
    """

    from __future__ import annotations

    import json
    from pathlib import Path

    from django.conf import settings
    from django.contrib.auth import get_user_model
    from django.contrib.auth.models import Group
    from django.core.management.base import BaseCommand, CommandError
    from django.db import transaction

    from tools.models import DevUserUuid

    User = get_user_model()


    class Command(BaseCommand):
        help = "Seed dev users from docker/postgres/init/seed-uuids/dev-users.json (idempotent)."

        def handle(self, *args, **options) -> None:
            manifest_path: Path = settings.BASE_DIR.parent / "docker" / "postgres" / "init" / "seed-uuids" / "dev-users.json"
            if not manifest_path.exists():
                raise CommandError(f"seed_dev_users: manifest not found at {manifest_path}")

            data = json.loads(manifest_path.read_text(encoding="utf-8"))
            entries = data["users"]

            created_count = 0
            updated_count = 0

            with transaction.atomic():
                for entry in entries:
                    first, *rest = entry["display_name"].split(" ", 1)
                    last = rest[0] if rest else ""

                    user, was_created = User.objects.update_or_create(
                        username=entry["username"],
                        defaults={
                            "first_name": first,
                            "last_name":  last,
                            "email":      f"{entry['username']}@fieldmark.local",
                            "is_active":  True,
                        },
                    )
                    # set_password invokes Django's configured PASSWORD_HASHERS (PBKDF2 by
                    # default). Always call save() after set_password.
                    user.set_password(entry["password"])
                    user.save()

                    DevUserUuid.objects.update_or_create(
                        user=user,
                        defaults={"uuid": entry["id"]},
                    )

                    if entry["role"]:
                        try:
                            group = Group.objects.get(name=entry["role"])
                        except Group.DoesNotExist as exc:
                            raise CommandError(
                                f"seed_dev_users: Group '{entry['role']}' missing. "
                                f"Run `manage.py seed_groups` first (Story 1.8)."
                            ) from exc
                        user.groups.set([group])
                    else:
                        user.groups.clear()

                    if was_created:
                        created_count += 1
                        self.stdout.write(f"  + created {entry['username']} ({entry['role'] or 'no-role'})")
                    else:
                        updated_count += 1
                        self.stdout.write(f"  ~ updated {entry['username']} ({entry['role'] or 'no-role'})")

            self.stdout.write(self.style.SUCCESS(
                f"seed_dev_users: {len(entries)} users ({created_count} created, {updated_count} updated)"
            ))
    ```

  - [ ] 3.5 Add a pytest covering create + idempotence at `fieldmark_py/projects/tests/test_seed_dev_users.py`:

    ```python
    import pytest
    from django.core.management import call_command

    from django.contrib.auth import get_user_model
    from django.contrib.auth.models import Group
    from tools.models import DevUserUuid

    User = get_user_model()


    @pytest.fixture
    def groups_seeded(db):
        for name in ("ADMIN", "COMPLIANCE_OFFICER", "INSPECTOR", "SITE_SUPERVISOR", "EXECUTIVE"):
            Group.objects.get_or_create(name=name)


    @pytest.mark.django_db
    def test_seed_dev_users_creates_six_users(groups_seeded):
        call_command("seed_dev_users")
        assert User.objects.count() == 6
        assert DevUserUuid.objects.count() == 6
        # The five roled users are in their groups; testuser has none.
        assert User.objects.get(username="marisol").groups.filter(name="COMPLIANCE_OFFICER").exists()
        assert User.objects.get(username="testuser").groups.count() == 0


    @pytest.mark.django_db
    def test_seed_dev_users_is_idempotent(groups_seeded):
        call_command("seed_dev_users")
        baseline_user_ids = set(User.objects.values_list("id", flat=True))
        baseline_uuid_ids = set(DevUserUuid.objects.values_list("user_id", flat=True))

        call_command("seed_dev_users")

        assert set(User.objects.values_list("id", flat=True)) == baseline_user_ids
        assert set(DevUserUuid.objects.values_list("user_id", flat=True)) == baseline_uuid_ids
    ```

    `pytest-django` is already configured (Story 1.8 added it); `@pytest.mark.django_db` hits the real test database — no SQLite (root `CLAUDE.md` → `docs/hard-rules.md`).

  - [ ] 3.6 Run end-to-end:
    - `make reset` (from repo root).
    - `cd fieldmark_py && uv run python manage.py migrate` (Story 1.8's auth + Task 3.3's `dev_user_uuid` migration).
    - `uv run python manage.py seed_groups` (Story 1.8).
    - `uv run python manage.py seed_dev_users` — six users created.
    - Re-run `seed_dev_users` — zero new rows.
    - Verify in psql: `SELECT au.username, d.uuid FROM django_auth.auth_user au JOIN django_auth.dev_user_uuid d ON d.user_id = au.id ORDER BY au.username;` — six rows, UUIDs match the manifest.

- [ ] Task 4: Implement the Go dev-users seeder (AC: #5, #7)
  - [ ] 4.1 Confirm Story 1.9 has landed: `fiber_auth.users` and `fiber_auth.user_roles` exist after `cmd/migrate-fiber-auth`. If not, stop and surface the dependency.
  - [ ] 4.2 Create `fieldmark-go/cmd/seed/main.go`. Sketch:

    ```go
    // Command seed reads docker/postgres/init/seed-uuids/dev-users.json and
    // writes six users into fiber_auth.users + fiber_auth.user_roles using the
    // canonical UUIDs from the manifest. Idempotent via ON CONFLICT.
    //
    // No password storage: the Go stack uses stub auth (ADR-012, Story 1.9) —
    // identity is asserted via the X-FieldMark-Actor cookie/header carrying a
    // username, not a password. The manifest's "password" field is read and
    // discarded by this seeder; .NET and Django persist it via their own
    // framework's password hasher.
    package main

    import (
        "context"
        "encoding/json"
        "errors"
        "fmt"
        "log"
        "os"
        "path/filepath"
        "strings"

        "github.com/jackc/pgx/v5/pgxpool"
    )

    type manifestEntry struct {
        ID          string  `json:"id"`
        Username    string  `json:"username"`
        DisplayName string  `json:"display_name"`
        Password    string  `json:"password"` // intentionally unused (see file header)
        Role        *string `json:"role"`     // null for the no-role test user
    }

    type manifest struct {
        Users []manifestEntry `json:"users"`
    }

    func main() {
        if err := run(); err != nil {
            fmt.Fprintln(os.Stderr, "seed:", err)
            os.Exit(1)
        }
    }

    func run() error {
        path, err := findManifest()
        if err != nil {
            return fmt.Errorf("find manifest: %w", err)
        }

        m, err := parseManifest(path)
        if err != nil {
            return fmt.Errorf("parse manifest %s: %w", path, err)
        }

        dsn := strings.TrimSpace(os.Getenv("FIELDMARK_DATABASE_URL"))
        if dsn == "" {
            dsn = "postgres://fieldmark:fieldmark@localhost:5432/fieldmark"
        }

        ctx := context.Background()
        pool, err := pgxpool.New(ctx, dsn)
        if err != nil {
            return fmt.Errorf("pgxpool: %w", err)
        }
        defer pool.Close()

        tx, err := pool.Begin(ctx)
        if err != nil {
            return fmt.Errorf("begin: %w", err)
        }
        defer func() { _ = tx.Rollback(ctx) }()

        for _, e := range m.Users {
            if _, err := tx.Exec(ctx, `
                INSERT INTO fiber_auth.users (id, username, display_name)
                    VALUES ($1, $2, $3)
                    ON CONFLICT (id) DO UPDATE
                       SET username     = EXCLUDED.username,
                           display_name = EXCLUDED.display_name
            `, e.ID, e.Username, e.DisplayName); err != nil {
                return fmt.Errorf("upsert user %s: %w", e.Username, err)
            }

            if e.Role == nil {
                if _, err := tx.Exec(ctx,
                    `DELETE FROM fiber_auth.user_roles WHERE user_id = $1`, e.ID,
                ); err != nil {
                    return fmt.Errorf("clear roles for %s: %w", e.Username, err)
                }
                continue
            }

            if _, err := tx.Exec(ctx, `
                INSERT INTO fiber_auth.user_roles (user_id, role)
                    VALUES ($1, $2)
                    ON CONFLICT (user_id, role) DO NOTHING
            `, e.ID, *e.Role); err != nil {
                return fmt.Errorf("upsert role for %s: %w", e.Username, err)
            }

            if _, err := tx.Exec(ctx,
                `DELETE FROM fiber_auth.user_roles WHERE user_id = $1 AND role <> $2`,
                e.ID, *e.Role,
            ); err != nil {
                return fmt.Errorf("prune roles for %s: %w", e.Username, err)
            }
        }

        if err := tx.Commit(ctx); err != nil {
            return fmt.Errorf("commit: %w", err)
        }

        log.Printf("seed_dev_users: %d users, 0 errors", len(m.Users))
        return nil
    }

    func parseManifest(path string) (*manifest, error) {
        f, err := os.Open(path)
        if err != nil {
            return nil, err
        }
        defer f.Close()
        var m manifest
        if err := json.NewDecoder(f).Decode(&m); err != nil {
            return nil, err
        }
        if len(m.Users) == 0 {
            return nil, errors.New("manifest has zero users")
        }
        return &m, nil
    }

    // findManifest walks up from the current working directory looking for
    // docker/postgres/init/seed-uuids/dev-users.json. Returns the absolute path
    // or an error if not found within 5 ancestor levels.
    func findManifest() (string, error) {
        cwd, err := os.Getwd()
        if err != nil {
            return "", err
        }
        cur := cwd
        for range 5 {
            candidate := filepath.Join(cur, "docker", "postgres", "init", "seed-uuids", "dev-users.json")
            if _, err := os.Stat(candidate); err == nil {
                return filepath.Abs(candidate)
            }
            parent := filepath.Dir(cur)
            if parent == cur {
                break
            }
            cur = parent
        }
        return "", fmt.Errorf("dev-users.json not found within 5 ancestors of %s", cwd)
    }
    ```

  - [ ] 4.3 Add a unit test for `parseManifest` at `fieldmark-go/cmd/seed/main_test.go` (pure-Go, no DB):

    ```go
    package main

    import (
        "os"
        "path/filepath"
        "testing"
    )

    func TestParseManifest_RoundTrip(t *testing.T) {
        dir := t.TempDir()
        path := filepath.Join(dir, "dev-users.json")
        if err := os.WriteFile(path, []byte(`
            {"users":[
              {"id":"01923456-7890-7abc-def0-123456789abc","username":"marisol",
               "display_name":"Marisol Vega","password":"x","role":"COMPLIANCE_OFFICER"},
              {"id":"01923456-7890-7abc-def0-123456789def","username":"testuser",
               "display_name":"Test User","password":"x","role":null}
            ]}`), 0o600); err != nil {
            t.Fatal(err)
        }
        m, err := parseManifest(path)
        if err != nil {
            t.Fatal(err)
        }
        if len(m.Users) != 2 {
            t.Fatalf("want 2 users, got %d", len(m.Users))
        }
        if m.Users[1].Role != nil {
            t.Fatalf("testuser role should be nil, got %v", *m.Users[1].Role)
        }
    }

    func TestParseManifest_RejectsEmpty(t *testing.T) {
        dir := t.TempDir()
        path := filepath.Join(dir, "dev-users.json")
        if err := os.WriteFile(path, []byte(`{"users":[]}`), 0o600); err != nil {
            t.Fatal(err)
        }
        if _, err := parseManifest(path); err == nil {
            t.Fatal("want error for empty users array, got nil")
        }
    }
    ```

  - [ ] 4.4 Run end-to-end:
    - `make reset` (from repo root).
    - `cd fieldmark-go && go run ./cmd/migrate-fiber-auth` (Story 1.9).
    - `go run ./cmd/seed` — six users created. Log line: `seed_dev_users: 6 users, 0 errors`.
    - Re-run `go run ./cmd/seed` — same log line, zero new rows (verify via psql).
    - `make test` (from `fieldmark-go/`) — all tests pass including the new `parseManifest` tests.

- [ ] Task 5: Add `make seed` and update READMEs (AC: #9, #10)
  - [ ] 5.1 Append the `seed` target block (AC #9 sketch) to the root `Makefile`. Confirm `make seed` invokes all three seeders in order and prints the trailing success line. The dependency chain is `seed: seed-net seed-django seed-go` so any single failure halts the chain (`make` default behavior).
  - [ ] 5.2 Update `README.md` (repo root): add a `make seed` step after `make reset` + per-stack migrate steps. One-line note: "Seeds the six dev users into all three stacks' auth schemas, with identical UUIDs sourced from `docker/postgres/init/seed-uuids/dev-users.json`."
  - [ ] 5.3 Update `FieldMark/README.md`: add a `dotnet run --project FieldMark.Web -- --seed-dev-users` step after the Story 1.7 auth-migration step. Cross-reference `make seed` as the canonical cross-stack convenience.
  - [ ] 5.4 Update `fieldmark_py/README.md`: add a `uv run python manage.py seed_dev_users` step after `seed_groups`. Note the side-table dependency (`tools/migrations`) is applied by the regular `migrate` step.
  - [ ] 5.5 Update `fieldmark-go/README.md`: add a `go run ./cmd/seed` step after the Story 1.9 `migrate-fiber-auth` step. One-line note: "Reads the shared `dev-users.json` manifest and writes six users into `fiber_auth.users` + `fiber_auth.user_roles`. Idempotent."

- [ ] Task 6: Update each stack's CLAUDE.md (AC: #3, #4, #5)
  - [ ] 6.1 `FieldMark/CLAUDE.md` — add to (or create) an `## Authentication / Dev User Seeding` section: the seeder lives at `FieldMark/FieldMark.Web/SeedData/DevUsersSeeder.cs`, runs on every web-app startup (after `RoleSeeder`), and is also invocable via `dotnet run --project FieldMark.Web -- --seed-dev-users`. Identity hashing uses `IPasswordHasher` (framework-native); the manifest plaintext password is for dev-environment first-login only.
  - [ ] 6.2 `fieldmark_py/CLAUDE.md` — add an `## Authentication / User UUIDs` section pointing at `tools.models.DevUserUuid` and the chosen side-table approach (rationale per AC #4). Note that `domain.audit_entry.actor_id` writes (Epic 2+) must look up the UUID via `request.user.dev_uuid.uuid` rather than the integer `request.user.pk`.
  - [ ] 6.3 `fieldmark-go/CLAUDE.md` — extend the existing `## Authentication` section (rewritten in Story 1.9): note that `cmd/seed` populates `fiber_auth.users` / `fiber_auth.user_roles` from the shared manifest, and that the manifest's `password` field is intentionally not persisted on the Go side (stub posture).

- [ ] Task 7: Cross-stack UUID parity verification (AC: #6, #7, #8)
  - [ ] 7.1 With all three stacks seeded, run the six spot-check queries (AC #6) against psql. Capture the output in the story's Completion Notes — six matches, three stacks each.
  - [ ] 7.2 Run the `grep` invariants (AC #7) — three zero-match runs. Capture the commands and zero-match outputs.
  - [ ] 7.3 Run `make parity` from the repo root — exits 0. Capture stdout.
  - [ ] 7.4 Run `make seed` twice in a row from the repo root — both runs complete cleanly; second run produces zero new rows on every stack (psql row-count check before/after).

- [ ] Task 8: Final QA gates (AC: #11)
  - [ ] 8.1 .NET: `cd FieldMark && dotnet build && dotnet test` — clean.
  - [ ] 8.2 Django: `cd fieldmark_py && uv run ruff check . && uv run mypy . && uv run python -m pytest` — clean. New tests added under `projects/tests/test_seed_dev_users.py` pass.
  - [ ] 8.3 Go: `cd fieldmark-go && make fmt-check && make vet && make staticcheck && make test` — clean. New tests under `cmd/seed/main_test.go` pass.
  - [ ] 8.4 `make parity` (repo root) — exits 0.

## Dev Notes

### Brownfield posture — what exists today (read before writing anything)

State of the three stacks at the head of this branch, with respect to dev-user seeding:

- **`docker/postgres/init/seed-uuids/`** — does not exist yet. Story 1.10 creates the directory, the manifest JSON, and the JSON Schema file.
- **`docker/postgres/init/`** — currently contains `001_schemas.sql`, `010_domain_tables.sql`, `020_domain_seed.sql`. The architecture line 314 mentions `091_seed_dev_users.sql` as "generated by per-stack seed runners; identical UUIDs" — **this line is internally inconsistent with line 947 (`seed-uuids/dev-users.json` as the manifest) and with ADR-012 (auth is framework-local)**. The correct mental model is the line-947 / line 270 model: per-stack seeders read the JSON manifest. **Do not** author `091_seed_dev_users.sql`. The story implements per-stack seeders; line 314 is a stale residue from an earlier architecture draft.
- **`docker/postgres/init/020_domain_seed.sql`** — owns reference data (TradeType, ViolationCategory, ComplianceRule). Story 1.10 must **not** touch this file; AC #7 enforces.
- **`FieldMark/FieldMark.Web/SeedData/`** — does not exist at HEAD of this branch. Story 1.7 introduces it with `RoleSeeder.cs` (currently in `review` status). Story 1.10 adds `DevUsersSeeder.cs` alongside it. **If 1.7 hasn't merged yet when this story runs, the seeder file location and the `Program.cs` wiring point are still correct — but the `RoleSeeder.SeedAsync` reference and the `IdentityUser<Guid>`/`IdentityRole<Guid>` types depend on Story 1.7's wiring.** Surface a clear dependency stop if 1.7 hasn't landed.
- **`FieldMark/FieldMark.Web/Program.cs`** — currently registers Razor Pages and configures `FieldMarkDbContext`. It does **not** yet register Identity or `AuthDbContext` — those come from Story 1.7. Story 1.10's `Program.cs` modifications are predicated on Story 1.7's having landed.
- **`fieldmark_py/projects/`** — exists with `models.py`, `views.py`, `admin.py`, `tests.py`, `migrations/`. It does **not** have a `management/commands/` directory yet. Story 1.10 creates `projects/management/__init__.py`, `projects/management/commands/__init__.py`, and `projects/management/commands/seed_dev_users.py`. Mirror the existing pattern at `fieldmark_py/tools/management/commands/dump_routes.py` (Story 1.3) for `__init__.py` placement.
- **`fieldmark_py/tools/`** — exists with `apps.py`, `__init__.py`, `management/commands/dump_routes.py`. It does **not** have `models.py` or `migrations/` yet. Story 1.10 creates both for the `DevUserUuid` side table.
- **`fieldmark_py/projects/tests/`** — does not exist; only `projects/tests.py` (the default Django stub) exists. Convert to a package: create `projects/tests/__init__.py` and migrate the existing empty stub if any. Place `test_seed_dev_users.py` here.
- **`fieldmark-go/cmd/`** — currently contains only `cmd/web/`. Story 1.9 adds `cmd/migrate-fiber-auth/`. Story 1.10 adds `cmd/seed/`. All three are siblings.
- **`fieldmark-go/internal/data/postgres/db.go`** — Story 1.9 upgrades to `pgxpool.Pool`. Story 1.10's seeder opens its **own** pool inside `cmd/seed/main.go` (the seeder is a separate process from the web app) — do not import `postgres.Connect` from `internal/data/postgres/` and do not share a pool with the web app.
- **Architecture line 1175–1176** specifies `fieldmark-go/cmd/seed/main.go` with the comment "reads dev-users.json (when fiber_auth lands)". `fiber_auth` lands in Story 1.9; Story 1.10 implements the seeder.
- **Story dependencies — 1.7, 1.8, 1.9 are upstream:**
  - Story 1.7 (.NET) is currently `review` — provides `AuthDbContext`, `RoleSeeder`, Identity wiring.
  - Story 1.8 (Django) is `ready-for-dev` — provides `search_path=django_auth,public`, `seed_groups`, the five `auth_group` rows.
  - Story 1.9 (Go) is `ready-for-dev` — provides `fiber_auth.users` + `fiber_auth.user_roles` schema via `cmd/migrate-fiber-auth`.

  Before Story 1.10 implementation begins, the dev should confirm all three have merged to the working branch. If any has not, the dev should pause and either (a) implement only the stacks whose dependencies have landed, marking the others as "blocked pending Story 1.X", or (b) wait for all three to merge. **Do not invent the upstream wiring**.

### Why JSON, not SQL or YAML

Three serialization options were considered:

- **SQL (e.g., `091_seed_dev_users.sql`).** Tempting because the existing `020_domain_seed.sql` is SQL and the file would `INSERT` directly. Rejected for three reasons: (a) violates ADR-012 — `*_auth` schemas are framework-local; running SQL DML against them from `docker/postgres/init/` would place auth seeding in infrastructure tooling, contradicting D9's per-stack-seeder model; (b) password hashing differs per stack (.NET `IPasswordHasher`, Django `PBKDF2`, Go stub — no password); SQL would have to bake in pre-hashed values, which then can't be re-hashed if the stack swaps its algorithm; (c) the side-table UUID approach in Django requires Django ORM idioms (`update_or_create`) that pure SQL handles awkwardly.
- **YAML.** Rejected: introduces a YAML parser dependency in three languages (`YamlDotNet` for .NET, `PyYAML` already a Django transitive dep but only just, `yaml.v3` for Go). JSON is in every stack's standard library and adds zero new deps.
- **JSON.** Chosen. Standard library parsing in all three stacks. JSON Schema validation is widely supported. Pretty-printed, six entries, ~50 lines — trivially diff-able.

### Why a Django side-table for UUIDs (not a custom user model)

Django offers four ways to give `auth_user` a UUID:

1. **Custom user model with `id = UUIDField(primary_key=True)`.** Most idiomatic, but requires committing to it **before any migrations land** — Story 1.8 has already migrated `auth_user` with its default integer PK, so this option is closed without a destructive reset. Chosen approach must be additive.
2. **Add a `uuid` column to `auth_user` via a custom migration.** Possible (because `auth_user` lives in `django_auth`, which is framework-local), but mutates the framework table beyond its standard shape. Future Django `auth` migrations (e.g., `auth.0013_*` if Django releases one) would have to coexist with the custom column; while unlikely to conflict, it's a footgun.
3. **Side table `django_auth.dev_user_uuid(user_id, uuid)` with FK to `auth_user.id`.** Adds one row per user, leaves `auth_user` untouched, lives in the same framework-local schema (clean ADR-012 story), provides a clear lookup contract. **Chosen.**
4. **Store the UUID as `username` (i.e., `username = <UUID>`).** Destroys human-readable usernames, breaks the `/login` user-switcher narrative (Story 1.11), and forces all log lines and audit text to display UUIDs. Rejected immediately.

The side table is exactly two columns and one row per user — `(user_id BIGINT FK, uuid UUID UNIQUE)`. Lookup at audit-write time is one `JOIN` away from the user object. Domain audit-entry code (Epic 2+) will look up the UUID via `request.user.dev_uuid.uuid` (Django reverse-accessor via `related_name="dev_uuid"`).

### Why the manifest is `{"users": [...]}`, not a bare array

JSON Schema and parser-friendly: a top-level object lets the file grow additive metadata (`schema_version`, `last_rotated_at`) without breaking parsers that key off `[0]`. Cost: one extra layer in each parser. Benefit: a clean evolutionary path if Story 1.X needs to add `groups` (multi-role) or `permissions` (special-case overrides) without breaking the cross-stack seeders' shape contract.

### Why .NET wires the seeder into startup, but Django and Go don't

.NET's `Program.cs` already runs `RoleSeeder.SeedAsync` on every web-app startup (Story 1.7). Wiring `DevUsersSeeder.SeedAsync` into the same path is symmetric, idempotent, and means "running the web app once" is enough to seed both roles and users — the developer experience is "no extra step" on .NET.

Django and Go take a different path: `manage.py` commands and `cmd/seed` are explicit invocations. Reasons: (a) Django doesn't have a "startup hook" idiom comparable to .NET's `IServiceScope`; (b) Go's `cmd/web/main.go` is the server entry point and starting a server should not perform DDL/DML side effects on every launch — that pattern was already rejected for `cmd/migrate-fiber-auth` (Story 1.9 Dev Notes "Why a manual `cmd/migrate-fiber-auth` instead of startup auto-migration"); the same rationale applies here. The cross-stack `make seed` target re-establishes parity at the developer's invocation: one command, three stacks.

The `--seed-dev-users` flag on the .NET side (AC #9 / Task 2.4) lets `make seed` invoke .NET without binding to `:5000` — but the **startup** path on .NET still runs the seeder, because the web-app's first launch otherwise leaves the database in a half-bootstrapped state (roles seeded by `RoleSeeder`, users missing) until `make seed` runs. The dual-path on .NET is a deliberate choice for ergonomics; Django and Go are explicit-only.

### Why the Go seeder doesn't persist the manifest's `password` field

The Go stack uses stub auth (ADR-012, Story 1.9). Identity is asserted via the `X-FieldMark-Actor` cookie/header carrying a username — no password verification happens server-side. Persisting password hashes in `fiber_auth.users` would be wasted storage and a misleading signal to future readers that the Go stack does real auth. The seeder reads the `password` field (so it can validate the manifest schema with one parser shape across all three stacks) and discards it.

When the deferred real-auth epic for Go lands (post-MVP), `fiber_auth.users` gains a `password_hash` column and the seeder grows to hash + persist. Until then: stub posture, no password storage.

### Why UUIDv7 (not UUIDv4 or sequential IDs)

The recommendation is "UUIDv7 preferred" — not mandatory. The rationale:

- **Time-ordered.** UUIDv7's first 48 bits encode the Unix millisecond timestamp. Sorting `id`s sorts users by creation order, which is useful in audit-log paging when actor identity is a tiebreaker.
- **Debug-friendly.** A UUIDv7 string starting with `0192...` immediately reveals "this was generated in early 2026"; a UUIDv4 reveals nothing. When debugging cross-stack audit divergence, the timestamp prefix helps quickly correlate rows.
- **Postgres-compatible.** UUIDv7 stores in the same `uuid` column type as UUIDv4 (16 bytes); no index changes needed.

The dev may use UUIDv4 if the chosen generator doesn't support v7 — the manifest is six rows, generated once, and never rotated; the trade-off is small. The contract is "canonical UUID per user, identical across all three stacks" — version 4 vs. version 7 is a quality-of-life choice, not a correctness one.

### Why architecture line 314's `091_seed_dev_users.sql` is wrong

Architecture line 314:

```
docker/postgres/init/091_seed_dev_users.sql # Generated by per-stack seed runners; identical UUIDs
```

This line is inconsistent with the rest of the architecture document and with ADR-012:

- Line 270 (Cross-Cutting Concerns): "Seed scripts using identical UUIDs across stacks — referenced in `domain-model.md` §3.11 but implementation status to confirm."
- Line 349–356 (D9 — Same-UUID seed strategy): explicitly per-stack runners reading `seed-uuids/dev-users.json`.
- Line 947–948 (Repository Directory Structure): `seed-uuids/dev-users.json # canonical user UUIDs (per-stack seeders consume)`.
- ADR-012 (Authentication & Security): `*_auth` schemas are framework-local; the infrastructure init scripts own `domain.*` only.

A `091_seed_dev_users.sql` would have to `INSERT INTO dotnet_auth.users`, `django_auth.auth_user`, `fiber_auth.users` from a SQL file at infrastructure init time — which contradicts framework-local ownership. The line is residual from an earlier draft where the manifest was envisioned as a SQL file. **Story 1.10 follows the per-stack-seeder model.** A small note correcting the line could be added to the story's Completion Notes for the architecture maintainer.

### Anti-patterns that must NOT slip in

- ❌ Authoring `docker/postgres/init/091_seed_dev_users.sql`. Violates ADR-012 — see "Why architecture line 314" above.
- ❌ Writing rows into `domain.*` from any seeder. The epic AC explicitly forbids this; AC #7 is the verification gate. Reference data stays in `020_domain_seed.sql`; dev users stay in `*_auth`.
- ❌ Modifying `auth_user`'s integer primary key (Django) — adds maintenance risk against future framework migrations. Use the side table.
- ❌ Modifying or rewriting `Identity` tables' shapes (.NET) — `IdentityUser<Guid>` is enough. Don't subclass `IdentityUser` to add fields just to carry the manifest's `display_name`; the `UserName` + `Email` slots suffice for MVP, and `display_name` can be reconstructed from manifest + first name / last name on the Django side. Resist the urge to add `DisplayName` columns; defer to Epic 2 when a UI surfaces the value.
- ❌ Persisting the manifest's `password` plaintext anywhere. .NET and Django hash via framework-native hashers. The Go seeder discards it.
- ❌ Auto-rotating the manifest's UUIDs on re-run. The IDs are committed once and **never** rotate; this is the cross-stack identity contract. Rotation is a manual coordinated change across all three seeders + a `make reset`.
- ❌ Adding a CLI argument to any seeder to "seed only N users" or "seed a single user." The manifest is the contract — six users, all six seeded, no partial states. If selective seeding becomes useful, it's a future story with its own design.
- ❌ Sharing the .NET seeder's `IServiceScope` across both `RoleSeeder` and `DevUsersSeeder` invocations from the `--seed-dev-users` flag path differently than the startup path. The two paths must be functionally equivalent — same registrations, same scope, same order.
- ❌ Putting the Django seeder under `tools/management/commands/seed_dev_users.py`. Architecture line 1140 says `projects/management/commands/seed_dev_users.py`. Story 1.8 reserved this location. The `tools/` namespace holds *cross-cutting* commands (`dump_routes`, `seed_groups`); `seed_dev_users` is conceptually closer to the projects aggregate (where the personas first appear in user journeys, per the PRD).

  *Counter-consideration acknowledged:* one could argue `seed_dev_users` is also cross-cutting. The architecture spec wins — line 1140 is explicit, and Story 1.8 already coded the boundary expecting this placement. If a future story wants to move it, that's an ADR-level discussion.

- ❌ Embedding the manifest into any compiled binary (Go `//go:embed`, .NET embedded resource, Django `apps.py` constants). The manifest is shared across stacks and must live in `docker/postgres/init/seed-uuids/` as the single source of truth; each seeder reads it at runtime via filesystem.
- ❌ Committing a different password for one stack than another. The manifest's `password` field is shared. On .NET and Django, the plaintext is hashed by framework-native hashers; the developer should be able to log in with the same credentials on both stacks. On Go, the field is ignored (stub auth).
- ❌ Using SQLite or any in-memory database for the Django test that calls `seed_dev_users`. Hard rule: real Postgres for all DB-touching tests. `@pytest.mark.django_db` against the configured test database is the canonical pattern.
- ❌ Calling Django's `seed_dev_users` from a Django signal, `AppConfig.ready()`, or any startup hook. Hard rule (`fieldmark_py/CLAUDE.md`): no Django signals. The command is run explicitly.
- ❌ Hard-coding the manifest path in any seeder. Resolve relative to each stack's project root (`env.ContentRootPath` for .NET, `settings.BASE_DIR.parent` for Django, `findManifest()` for Go). The same manifest must be discoverable from all three working directories.

### Project Structure Notes

Files this story adds or modifies:

**Manifest (new — shared across stacks):**
- **New:** `docker/postgres/init/seed-uuids/dev-users.json`
- **New:** `docker/postgres/init/seed-uuids/dev-users.schema.json`
- **New (or update):** `docker/postgres/init/README.md` — `## Dev User Manifest` section.

**.NET:**
- **New:** `FieldMark/FieldMark.Web/SeedData/DevUsersSeeder.cs`
- **Update:** `FieldMark/FieldMark.Web/Program.cs` — invoke `DevUsersSeeder.SeedAsync` after `RoleSeeder.SeedAsync`; add `--seed-dev-users` flag short-circuit; refactor Identity/DbContext registration into a private helper for the dual-path symmetry.
- **Update:** `FieldMark/README.md` — Getting Started seed step.
- **Update:** `FieldMark/CLAUDE.md` — `## Authentication / Dev User Seeding` section.

**Django:**
- **New:** `fieldmark_py/projects/management/__init__.py` (empty)
- **New:** `fieldmark_py/projects/management/commands/__init__.py` (empty)
- **New:** `fieldmark_py/projects/management/commands/seed_dev_users.py`
- **New:** `fieldmark_py/projects/tests/__init__.py` (empty — converts existing `tests.py` stub into package)
- **New:** `fieldmark_py/projects/tests/test_seed_dev_users.py`
- **Delete (if empty):** `fieldmark_py/projects/tests.py` — replaced by the tests package.
- **New:** `fieldmark_py/tools/models.py`
- **New:** `fieldmark_py/tools/migrations/__init__.py` (if not present)
- **New:** `fieldmark_py/tools/migrations/0001_initial.py` (or next-available number — generated by `makemigrations`).
- **Update:** `fieldmark_py/README.md` — Getting Started seed step.
- **Update:** `fieldmark_py/CLAUDE.md` — `## Authentication / User UUIDs` section.

**Go:**
- **New:** `fieldmark-go/cmd/seed/main.go`
- **New:** `fieldmark-go/cmd/seed/main_test.go`
- **Update:** `fieldmark-go/README.md` — Getting Started seed step.
- **Update:** `fieldmark-go/CLAUDE.md` — extend `## Authentication` section with seeder note.

**Root:**
- **Update:** `Makefile` — `seed`, `seed-net`, `seed-django`, `seed-go` targets.
- **Update:** `README.md` — `make seed` step in Getting Started.

No file in `domain.*` init scripts (`010_domain_tables.sql`, `020_domain_seed.sql`), `fieldmark_shared/`, `e2e/`, `tools/parity/`, or any non-listed location is modified. All locations align with Architecture lines 939–1230 (Repository Directory Structure).

### Testing Standards

Per root `CLAUDE.md` → `docs/hard-rules.md` and each stack's CLAUDE.md:

- **No SQLite in tests** — real PostgreSQL only. Django tests use `@pytest.mark.django_db` against the configured test database; Go DB-touching tests are `//go:build integration` tagged and deferred (the seeder's pure-Go parser is the only thing unit-tested here); .NET tests use `Microsoft.EntityFrameworkCore.InMemory` only for non-Identity DbContexts where applicable (the `DevUsersSeeder` test, if added, should hit real Postgres via the integration project — but this story does **not** require a .NET test, because the seeder is already exercised by the cross-stack UUID-parity spot-check (AC #6) and the idempotence verification (AC #11)).
- **Pytest fixture for Group seeding** — the Django seeder depends on the five `auth_group` rows existing. The fixture pattern shown in Task 3.5 (`groups_seeded`) seeds them inline rather than depending on `seed_groups` having been called separately; this keeps the test self-contained.
- **Go test pattern** — standard library `testing` only (Story 1.9 testing standards carry forward). No testify, no gomock. Pure-Go tests cover `parseManifest`; DB-touching tests deferred.
- **No test of the full three-stack roundtrip** — the cross-stack UUID-parity spot-check is a manual SQL verification (AC #6, captured in Completion Notes), not an automated test. An E2E test that asserts "marisol's UUID is identical across stacks" is reasonable for Epic 7 (the Playwright suite), not for this story.

### Previous Story Intelligence

**Story 1.7 (.NET — `review`):**
- Established `SeedData/` directory and `RoleSeeder.SeedAsync(IServiceProvider, CancellationToken)` signature. `DevUsersSeeder` follows the same pattern: static class, async method, scope-passed via `IServiceProvider`.
- Established Identity wiring: `IdentityUser<Guid>` and `IdentityRole<Guid>` with `Guid` PKs for `uuid` columns at rest (Architecture §Schema ownership "UUIDs generated in app code"). The manifest's `id` field deserializes directly into `Guid` (matching JSON case-insensitively).
- Established `AuthDbContext` wired with `HasDefaultSchema("dotnet_auth")` and `UseSnakeCaseNamingConvention()`. The seeder writes through `UserManager`, never against `AuthDbContext` directly — this is the .NET Identity idiom and Story 1.7 honored it.
- Idempotence pattern: existence-check then create (`RoleManager.RoleExistsAsync` → `CreateAsync`). Story 1.10 mirrors with `UserManager.FindByIdAsync` → `CreateAsync`.

**Story 1.8 (Django — `ready-for-dev`):**
- Established `search_path=django_auth,public` on the default DB connection — all unqualified `CREATE TABLE`s land in `django_auth`. The `tools/migrations/0001_initial.py` (Task 3.3) creates `django_auth.dev_user_uuid` via this mechanism (the explicit `db_table = 'django_auth"."dev_user_uuid'` quoting is belt-and-suspenders — `search_path` already routes there, but the qualified `db_table` makes the intent explicit and survives a future `search_path` change).
- Established `tools` app as the home for cross-cutting management commands (`dump_routes`, `seed_groups`). The `DevUserUuid` model also lives in `tools` (cross-cutting — not specific to any aggregate).
- Established `seed_groups` invocation idiom: `Group.objects.get_or_create(name=...)`. Story 1.10's seeder calls `Group.objects.get(name=...)` (assumes groups exist) and surfaces a `CommandError` if not — pushing the dev to run `seed_groups` first rather than silently creating groups in the wrong command.

**Story 1.9 (Go — `ready-for-dev`):**
- Established `fiber_auth.users` (`id uuid PK, username varchar(64) UNIQUE, display_name varchar(128), created_at timestamptz`) and `fiber_auth.user_roles` (`user_id uuid FK, role varchar(64), PRIMARY KEY (user_id, role), CHECK (role IN (5 names))`). Story 1.10's seeder is the **first writer** against these tables.
- Established the manual-migration pattern (`cmd/migrate-fiber-auth`). Story 1.10's `cmd/seed/main.go` follows the same shape: small `main.go`, transaction-bracketed exec, `pgxpool.New` + `pool.Close()`.
- Noted that multi-role support is post-MVP (Story 1.12). Story 1.10 honors this: each manifest entry has exactly one `role` (or `null`); the seeder inserts at most one row into `fiber_auth.user_roles` per user.
- Noted `INSERT ... ON CONFLICT DO NOTHING` as the SQL idempotence pattern. Story 1.10 adopts it (`ON CONFLICT (id) DO UPDATE` for the `users` table to keep `username`/`display_name` in sync, and `ON CONFLICT (user_id, role) DO NOTHING` for `user_roles`).

**Story 1.4 (design system — `review`)** and **Story 1.5 / 1.6** are unrelated to this story's surface.

### Git Intelligence

Recent commits and their relevance to Story 1.10:

- `d03f0fe feat: e1s3 establish tools parity` — established the `make parity` contract that AC #8 verifies. Story 1.10 adds zero routes and touches zero `domain.*` indexes, so the parity surface is undisturbed.
- `cbf47e9 feat: e1s2 verified sql init scripts` — confirmed `docker/postgres/init/` ordering (001 → 010 → 020). Story 1.10 adds files under `seed-uuids/` (a subdirectory, no execution order) — the init scripts ignore it because `docker-entrypoint-initdb.d` only runs `.sql` files at the top level, not subdirectory contents. (Verify by inspecting `docker-compose.yml`'s `volumes:` mount — it should mount `init/` flat; if it mounts recursively and tries to run `*.json`, this AC needs to address that. Empirically: Postgres' init mechanism runs `*.sql` and `*.sh`/`*.sql.gz`; `.json` files are ignored.)
- `a6fac88 feat: e1s1 confirm scaffolds` — confirmed the three native scaffolds. The seeder files Story 1.10 creates all sit in scaffolded locations.

No prior commit has authored a dev-user manifest or any seeder code. Story 1.10 is the first.

### Latest Technical Information

- **JSON in .NET 10:** `System.Text.Json.JsonSerializer` with `JsonNamingPolicy.SnakeCaseLower` (added in .NET 8) is the canonical snake-case JSON pattern. The manifest is snake-case (matching cross-stack DB convention); deserialization to the C# record (PascalCase property names) needs the naming policy set explicitly. The `record` syntax with positional parameters is the cleanest binding shape.
- **.NET Identity password hashing:** `UserManager.CreateAsync(user, plaintextPassword)` invokes `IPasswordHasher<TUser>` internally (default: `PasswordHasher<TUser>` with PBKDF2 + HMAC-SHA256, 100,000 iterations as of .NET 10). The seeder must not call the hasher directly.
- **Django 6.0 `update_or_create`:** transactional by default since Django 4.1; the `with transaction.atomic():` wrapper in Task 3.4 is defensive (ensures all six seeds are one transaction).
- **Django `PASSWORD_HASHERS`:** default is PBKDF2 with SHA-256 (PRD §NFR Security: "framework-native salted"). `user.set_password(plaintext)` followed by `user.save()` is the canonical idiom; `User.objects.create_user(...)` is an alternative that combines both but doesn't compose cleanly with `update_or_create`.
- **`pgxpool` ON CONFLICT:** Postgres' `ON CONFLICT (col) DO UPDATE SET ... = EXCLUDED.*` is the canonical upsert. Story 1.9 documents pgx v5 patterns; the seeder reuses them.
- **PostgreSQL 17 UUID type:** stable; accepts case-insensitive UUID strings; stores normalized internally. UUIDv7 is just a 16-byte value with timestamp ordering in the upper bits — Postgres treats it identically to UUIDv4.
- **No new dependencies needed:**
  - .NET: `System.Text.Json` and `Microsoft.AspNetCore.Identity` are already referenced via Story 1.7.
  - Django: `django.contrib.auth`, `django.db`, `json` (stdlib) all present.
  - Go: `encoding/json`, `path/filepath`, `os` (stdlib); `pgxpool` already present from Story 1.9.

### References

- [Architecture: Authentication & Security → D9 (same-UUID seed strategy)](_bmad-output/planning-artifacts/architecture.md#authentication--security) — the canonical specification this story implements. Lines 349–356.
- [Architecture: Repository Directory Structure](_bmad-output/planning-artifacts/architecture.md#complete-repository-directory-structure) — file locations: `docker/postgres/init/seed-uuids/dev-users.json` (line 947), `FieldMark/FieldMark.Web/SeedData/DevUsers.cs` (line 1083), `fieldmark_py/projects/management/commands/seed_dev_users.py` (line 1140), `fieldmark-go/cmd/seed/main.go` (line 1175).
- [Architecture: Gap Analysis item 8](_bmad-output/planning-artifacts/architecture.md#gap-analysis-results) — manifest shape recommendation (`{id, username, password, roles}`). Story 1.10 finalizes the shape (with `display_name` added and `roles` → `role` singular per the per-stack-auth idiom).
- [Architecture: Cross-Cutting Concerns line 270](_bmad-output/planning-artifacts/architecture.md#cross-cutting-concerns-identified) — "Seed scripts using identical UUIDs across stacks — referenced in `domain-model.md` §3.11 but implementation status to confirm." This story confirms it.
- [PRD §NFR Security](_bmad-output/planning-artifacts/prd/non-functional-requirements.md) — "Password hashing: Framework-native salted." Justifies per-stack hasher choice.
- [PRD architectural-constraints-prd-binding.md — Authentication & Authorization](_bmad-output/planning-artifacts/prd/architectural-constraints-prd-binding.md) — ADR-012 schema isolation; opaque UUIDs; auth ownership per stack.
- [domain-model.md §3.12 — Conceptual Roles](_bmad-output/planning-artifacts/research/domain-model.md) — the five canonical role names. Personas (Marisol, Pat, Aisha, Ravi, Kenji) appear in user journeys throughout this document.
- [docs/hard-rules.md](docs/hard-rules.md) — backend authority, infrastructure-owned domain schema, real Postgres in tests, no service layers.
- [FieldMark/CLAUDE.md](FieldMark/CLAUDE.md) — .NET-specific rules; what belongs in `FieldMark.Web` vs. `FieldMark.Data`.
- [fieldmark_py/CLAUDE.md](fieldmark_py/CLAUDE.md) — Django-specific rules; no signals; migrations scoped to `django_auth` only.
- [fieldmark-go/CLAUDE.md](fieldmark-go/CLAUDE.md) — Go-specific rules; `fiber.Ctx` stays in web layer (irrelevant to `cmd/seed/`); explicit SQL preferred over query builders.
- [Story 1.7 implementation artifact](_bmad-output/implementation-artifacts/1-7-wire-asp-net-core-identity-to-dotnet-auth-schema-with-conceptual-roles.md) — .NET dependency: `RoleSeeder`, `AuthDbContext`, Identity wiring.
- [Story 1.8 implementation artifact](_bmad-output/implementation-artifacts/1-8-wire-django-built-in-auth-to-django-auth-schema-with-conceptual-role-groups.md) — Django dependency: `search_path`, `seed_groups`, five `auth_group` rows.
- [Story 1.9 implementation artifact](_bmad-output/implementation-artifacts/1-9-implement-go-fiber-stub-authentication-middleware.md) — Go dependency: `fiber_auth.users` / `fiber_auth.user_roles` schema, `cmd/migrate-fiber-auth` invocation idiom.
- [Story 1.11 (forthcoming)](_bmad-output/planning-artifacts/epics.md#story-111-login-logout-and-unauthenticated-redirect-across-all-three-stacks) — downstream consumer: the user-switcher on Go reads `fiber_auth.users` populated by this story; .NET / Django `/login` forms authenticate against the users populated by this story.

## Dev Agent Record

### Agent Model Used

_(populated by dev agent)_

### Debug Log References

### Completion Notes List

### File List
