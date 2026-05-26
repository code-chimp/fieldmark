# AGENTS.md

Compact guidance for OpenCode agents in FieldMark (multi-stack HTMX reference impl).

## Essential Commands

Always start here:

```bash
make help          # list all targets
make up            # start Postgres 17 (first-run inits schemas)
make reset         # destroy volume + re-init (after schema changes)
./tools/verify-domain-schema.sh   # confirm domain.* (requires psql)
```

Run stacks (separate terminals):

```bash
make run-net       # :4000
make run-django    # :8000
make run-go        # :3000
```

Tests & verification:

```bash
make test-net
make test-django
make test-go
make e2e           # skips unless e2e/ deps installed
make parity        # cross-stack route/index diffs (Story 1.3+)
make css           # Tailwind build (skips unless fieldmark_shared/ deps)
```

## Key Architecture (Verified from Makefile + docker-compose + README)

- **Shared DB, isolated schemas**: `domain` owned by SQL init scripts in `docker/postgres/init/` (run once on empty volume). Auth schemas (`django_auth` etc.) owned by their stacks. Never dual-own domain tables.
- **Init quirk**: Postgres init scripts execute only on first container start (empty volume). Use `make reset` after changing `docker/postgres/init/`.
- **Orchestration**: Root Makefile is source of truth for dev flow. Stack CLIs (dotnet/uv/go) are invoked via `make run-*` / `make test-*`.
- **Symmetry**: Routes, HTMX targets, AG Grid contracts, audit strings, domain methods identical across .NET/Django/Go. Divergence = defect.
- **No client state**: HTMX + server-rendered only; AG Grid is island. No Redux etc. in any stack.
- **Domain first**: Business rules on entities; handlers thin. See docs/ for canonical request flow.

## Setup Gotchas

- `psql` required for verify script (macOS: `brew install libpq && brew link --force libpq`).
- pnpm in `fieldmark_shared/` and `e2e/` for CSS/e2e (conditional skips in Makefile).
- Each stack has own CLAUDE.md + README for framework-specific commands (e.g. dotnet csharpier, EF migrations scoped to dotnet_auth only, uv run pytest).
- Pre-kickoff: e2e/, parity/ tools, full CI not yet scaffolded.

## References (Progressive Disclosure)

- [docs/README.md](docs/README.md) — index to architecture, getting-started, hard-rules.
- Root [CLAUDE.md](CLAUDE.md) — cross-stack rules (now slim).
- Stack CLAUDE.md files for .NET/Django/Go specifics.
- `docker-compose.yml` + `Makefile` — executable truth for infra/dev.
- `_bmad-output/planning-artifacts/` — historical (not authoritative).

Prefer Makefile + docs/ over guessing paths or commands.
