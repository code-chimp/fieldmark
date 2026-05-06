# FieldMark Docker Compose & Postgres Init Strategy

## Purpose

This document defines the **Docker Compose and PostgreSQL initialization strategy** for FieldMark.

Its goals are:

- deterministic local setup
- strong separation of concerns
- low developer friction
- architectural clarity

This document accompanies the Authentication & Authorization Strategy and should be implemented alongside it.

---

## Guiding Principle

> **Schemas are infrastructure, not application code.**

Therefore:

- schemas are created by Postgres init scripts
- frameworks migrate only their own tables
- no framework implicitly creates infrastructure

---

## Why Init Scripts Are Required

Using Postgres init scripts:

- guarantees schema existence before migrations
- removes ordering ambiguity
- simplifies onboarding
- mirrors real enterprise practices

This is especially important in a **multi‑framework demo**.

---

## Recommended Docker Layout

```text
docker/
├── postgres/
│   └── init/
│       └── 001_schemas.sql
├── docker-compose.yml
```

Postgres will automatically execute any `.sql` files in this directory on first startup.

---

## Schema Initialization Script

The init script should create **only structural invariants**.

Example: `001_schemas.sql`

```sql
CREATE SCHEMA IF NOT EXISTS domain;
CREATE SCHEMA IF NOT EXISTS django_auth;
CREATE SCHEMA IF NOT EXISTS dotnet_auth;
CREATE SCHEMA IF NOT EXISTS fiber_auth;
CREATE SCHEMA IF NOT EXISTS infra;
```

Optional role and permission setup may also live here.

---

## Docker Compose Configuration

Postgres service must mount the init directory:

```yaml
volumes:
  - ./docker/postgres/init:/docker-entrypoint-initdb.d
```

This ensures schemas are present before any backend connects.

---

## Django Configuration Expectations

After schema initialization:

- Django migrations assume `django_auth` exists
- Django never creates schemas
- Django writes only to `django_auth` and `domain`

If schemas are missing, startup should fail loudly.

---

## Developer Experience Benefits

This approach ensures:

- `docker compose up` works everywhere
- no manual SQL steps
- clear ownership boundaries
- easy addition of new backends

Future users can inspect the init script to understand database structure immediately.

---

## CI / Automation Notes

The same init scripts should be used:

- locally
- in CI containers
- in demo deployment environments

Infrastructure should never be re‑declared in application migrations.

---

## Status

Accepted – Docker Compose & Postgres Init Strategy
