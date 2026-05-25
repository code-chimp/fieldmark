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

- [.NET / FieldMark/CLAUDE.md](../../FieldMark/CLAUDE.md)
- [Django / fieldmark_py/CLAUDE.md](../../fieldmark_py/CLAUDE.md)
- [Go / fieldmark-go/CLAUDE.md](../../fieldmark-go/CLAUDE.md)

## Shared Assets

CSS and vendor libs (AG Grid, HTMX) live in `fieldmark_shared/` and are symlinked into each stack's static/vendor.

## CSS Pipeline

Shared CSS lives in `fieldmark_shared/`. All `pnpm` commands run from that directory.

```bash
cd fieldmark_shared
pnpm install           # first-time setup (pnpm only — npm/yarn are blocked)
pnpm run build         # compile + optimize (development)
pnpm run build:prod    # compile, minify, optimize (production)
pnpm run build:raw     # Tailwind compile only — skips optimize-css; use when debugging the pipeline itself
pnpm run watch         # watch mode; run alongside app dev servers
```

### Two-step build

1. **Tailwind** compiles `src/fieldmark.css` (with `@source` directives scanning all three stacks) → `dist/fieldmark.css`.
2. **`scripts/optimize-css.mjs`** runs LightningCSS over the output to merge duplicate selectors emitted by Tailwind v4 + Basecoat, then removes consecutive `content: var(--tw-content)` duplicates. Result is lower byte count. The script writes to a `.tmp` file then atomically renames to avoid partial writes.

`build:raw` skips step 2 — useful when you suspect `optimize-css.mjs` is causing a problem and want the raw Tailwind output.

### Pre-build sanity checks (`prebuild`)

Two scripts run automatically before every `pnpm run build`:

- **`scripts/check-sources.mjs`** — resolves each `@source "..."` glob in `src/fieldmark.css` and exits non-zero if any glob matches zero files. This catches silent failures when a stack directory is renamed.
- **`scripts/check-basecoat-classes.mjs`** — greps `node_modules/basecoat-css/dist/basecoat.css` for the pinned class names (`.btn`, `.badge`, `.alert`, `.field`, `.toaster`, `.toast`, `.sidebar`). Exits non-zero listing missing classes. Catches Basecoat class renames on version bumps.

### pnpm-only guard

`fieldmark_shared/package.json` includes `"packageManager": "pnpm@11.0.8"` and a `preinstall` script that exits non-zero with a clear message if run with npm or yarn. Running `npm install` in that directory will fail fast with `Use pnpm`.

### Basecoat upgrade procedure

See [docs/how-to/basecoat-upgrade-checklist.md](../how-to/basecoat-upgrade-checklist.md) for the step-by-step procedure and the rationale for the pinned class smoke test.

### `optimize-css.mjs` failure modes

- **Missing input** → exits 1, message to stderr: `cannot read input`
- **Directory as input** → exits 1: `not a regular file`
- **Empty input** → exits 1: `empty` / `0 bytes`
- **LightningCSS warning of type `error` or `unsupported`** → prints all warnings to stderr, then exits 1
- **Write failure** → `.tmp` file is cleaned up; original output is not overwritten

## Next Steps

- Review [Architecture](../explanation/architecture.md)
- Read root [CLAUDE.md](../../CLAUDE.md) for agent rules
- Read [fieldmark_shared/CLAUDE.md](../../fieldmark_shared/CLAUDE.md) for full CSS pipeline, symlink topology, and pinned dependency table
