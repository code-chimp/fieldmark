# docker/postgres/init

This directory contains the PostgreSQL initialization scripts and shared seed data consumed by all three FieldMark stacks.

## Init Scripts

Scripts run in filename order on a fresh Postgres volume (via `docker-entrypoint-initdb.d`). Only `.sql` files at the **top level** of this directory run automatically; subdirectory contents are ignored by the init mechanism.

| File | Purpose |
|---|---|
| `001_schemas.sql` | Create `domain`, `dotnet_auth`, `django_auth`, `fiber_auth` schemas |
| `010_domain_tables.sql` | Domain entity tables (projects, inspections, violations, corrective actions, audit log) |
| `020_domain_seed.sql` | Reference/lookup data: TradeType, ViolationCategory, ComplianceRule |

## Dev User Manifest

`seed-uuids/dev-users.json` is the single source of truth for the six FieldMark development personas. Each per-stack seeder reads this file at runtime and writes users into its own `*_auth` schema. **The init scripts do not read this file** — it exists here because `docker/postgres/init/` is the natural home for Postgres-adjacent shared seed data, and because all three stack seeders resolve it relative to the repo root.

### The Six Personas

| username | display_name | role |
|---|---|---|
| `marisol` | Marisol Vega | `COMPLIANCE_OFFICER` |
| `pat` | Pat Smith | `SITE_SUPERVISOR` |
| `aisha` | Aisha Patel | `ADMIN` |
| `ravi` | Ravi Kumar | `INSPECTOR` |
| `kenji` | Kenji Tanaka | `EXECUTIVE` |
| `testuser` | Test User | *(no role)* |

These usernames are referenced in epic acceptance criteria, Story 1.9, Story 5.5, Story 6.4, and `domain-model.md` persona narratives. **Do not rename them** without searching and updating every reference.

### Password Policy

All six users share the same plaintext password at first seed (`FieldMark!2026`). The policy comes from .NET Identity (Architecture D6): ≥ 10 characters, ≥ 1 digit, ≥ 1 lowercase, ≥ 1 uppercase. This password is for the development environment only and is rotated post-MVP.

Each stack hashes using its own framework-native hasher:
- **.NET**: `IPasswordHasher<IdentityUser<Guid>>` (PBKDF2 + HMAC-SHA256, 100k iterations)
- **Django**: `set_password()` via `PASSWORD_HASHERS` (PBKDF2 with SHA-256)
- **Go**: stub auth (ADR-012) — the `password` field is read but **not** persisted

### UUID Rotation Rule

The `id` values in `dev-users.json` are **committed once and never rotated** without coordinating all three stack seeders simultaneously. These IDs become the canonical `domain.audit_entry.actor_id` values from Epic 2 onward. A rotation without coordination would break cross-stack audit comparison.

### Running the Seeders

After the per-stack auth schemas are in place (Story 1.7 for .NET, Story 1.8 for Django, Story 1.9 for Go), seed all three stacks with:

```sh
make seed
```

Or individually:

```sh
make seed-net      # dotnet run --project FieldMark.Web -- --seed-dev-users
make seed-django   # uv run python manage.py seed_dev_users
make seed-go       # go run ./cmd/seed
```

All seeders are idempotent — re-running produces zero new rows.
