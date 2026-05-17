# Getting Started with FieldMark

FieldMark is a construction compliance and inspection management system with three parallel HTMX stacks against shared PostgreSQL.

## Prerequisites

- Docker Desktop (for PostgreSQL 17)
- .NET SDK, Python/Django, Go toolchain (for respective stacks)
- Node.js (for shared Tailwind build if modifying CSS)

## Quick Start

```bash
docker compose up -d    # Starts PostgreSQL 17 on localhost:5432 (user: fieldmark)
```

Postgres init scripts in `docker/postgres/init/` create schemas: `domain`, `django_auth`, `dotnet_auth`, `fiber_auth`, `infra`.

If volume issues: `docker compose down -v && docker compose up -d`.

## Stack-Specific Setup

See each stack's `CLAUDE.md` for build/run/test commands:

- [.NET / FieldMark/CLAUDE.md](../FieldMark/CLAUDE.md)
- [Django / fieldmark_py/CLAUDE.md](../fieldmark_py/CLAUDE.md)
- [Go / fieldmark-go/CLAUDE.md](../fieldmark-go/CLAUDE.md)

## Shared Assets

CSS and vendor libs (AG Grid, HTMX) live in `fieldmark_shared/` and are symlinked into each stack's static/vendor.

## Next Steps

- Review [Architecture](architecture.md)
- Read root [CLAUDE.md](../CLAUDE.md) for agent rules
