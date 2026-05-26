# Getting Started with FieldMark

FieldMark is a construction compliance and inspection management system implemented across three parallel stacks (.NET, Django, Go) against a shared PostgreSQL database. Each stack is independent — you can work in one without touching the others.

## Prerequisites

| Tool | Version | Required for | Notes |
|------|---------|-------------|-------|
| [Docker Desktop](https://www.docker.com/products/docker-desktop/) | — | All stacks | Runs PostgreSQL 17 locally |
| [GNU Make](https://www.gnu.org/software/make/) | — | All stacks | Orchestrates dev commands (`make up`, `make seed`, etc.) |
| [.NET SDK](https://dotnet.microsoft.com/en-us/download) | 8.0+ | .NET stack | `dotnet --version` to check |
| [Python 3.12+](https://www.python.org/downloads/) | 3.12+ | Django stack | |
| [uv](https://docs.astral.sh/uv/) | — | Django stack | Python package manager (`pip install uv` or standalone) |
| [Go](https://go.dev/dl/) | 1.22+ | Go stack | `go version` to check |
| [Node.js 20+](https://nodejs.org/) | 20+ | CSS builds only | Tailwind's Oxide engine requires Node ≥ 20 |
| [pnpm](https://pnpm.io/) | 11.x | CSS builds only | `npm install -g pnpm` |

You do not need every language toolchain installed — pick the stack you want to work on.

### Windows / WSL notes

This project uses **symlinks** to share front-end assets across stacks. When cloning on Windows or WSL:

- **WSL (recommended):** Clone into WSL's native ext4 filesystem (`~/projects/fieldmark`), not `/mnt/c/`. Symlinks work natively there. Then enable them in Git:

  ```bash
  git config core.symlinks true
  git checkout -- FieldMark/FieldMark.Web/wwwroot/vendor/
  git checkout -- fieldmark_py/static/vendor/
  git checkout -- fieldmark-go/internal/web/static/vendor/
  ```

- **Native Windows:** Enable [Developer Mode](https://learn.microsoft.com/en-us/windows/apps/get-started/enable-your-device-for-development) (Settings → Privacy & Security → For Developers), then set `git config core.symlinks true` before cloning.

  If Developer Mode is not available, the symlinks will materialize as broken plain-text files. In that case you will need a workaround such as a post-clone script that copies the vendor directories.

After fixing symlinks, verify with `ls -la` — the vendor entries should show arrow targets like `vendor/htmx -> ../../../../fieldmark_shared/vendor/htmx`.

---

## 1. Start the Database

```bash
make up
```

This starts PostgreSQL 17 inside Docker on `localhost:5432` (user: `fieldmark`, password: `fieldmark`). The init scripts in `docker/postgres/init/` run automatically on first start, creating five schemas (`domain`, `django_auth`, `dotnet_auth`, `fiber_auth`, `infra`) and populating the domain reference data (trade types, violation categories, compliance rules).

Verify everything landed correctly:

```bash
./tools/verify-domain-schema.sh
```

You should see: `OK domain schema verified (5 schemas, 12 tables, ...)`.

If you need to reset the database later (e.g. after changing init scripts), run `make reset` — this destroys the volume and re-runs init.

---

## 2. Build Shared CSS (optional)

The compiled CSS (`dist/fieldmark.css`) is committed to the repo, so you can skip this step. Build it only if you are modifying styles:

```bash
cd fieldmark_shared
pnpm install        # first time only
pnpm run build
```

To watch for changes during development, run `pnpm run watch` alongside your dev server.

---

## 3. Set Up Your Stack

Each stack creates its own **auth tables** (user accounts, roles) in a dedicated schema and then **seeds** six dev users with known credentials. The auth tables are *not* created by the database init scripts — each framework does it on its own.

Pick the stack (or stacks) you want to run.

### .NET (Razor Pages + HTMX) — port :4000

```bash
# Create auth tables + seed dev users (one command — runs on every startup)
make run-net
```

This runs EF Core migrations (creates `dotnet_auth.users`, `dotnet_auth.roles`, etc.), then seeds roles and dev users. Once the server starts, you can log in at `http://localhost:4000/login`.

To seed without starting the full web server:

```bash
make seed-net
```

### Django (Templates + HTMX) — port :8000

```bash
# Create auth tables (Django migrations)
cd fieldmark_py && uv run python manage.py migrate

# Seed groups (conceptual roles)
uv run python manage.py seed_groups

# Seed dev users
uv run python manage.py seed_dev_users

# Start the dev server
uv run python manage.py runserver
```

Then visit `http://localhost:8000/login`.

Shortcut from the repo root:

```bash
make seed-django   # runs seed_dev_users (requires migrate + seed_groups first)
make run-django    # starts the dev server
```

### Go / Fiber (Templates + HTMX) — port :3000

```bash
# Create auth tables (fiber_auth.users, fiber_auth.user_roles)
cd fieldmark-go && go run ./cmd/migrate-fiber-auth

# Seed dev users
go run ./cmd/seed

# Start the dev server
go run ./cmd/web
```

Then visit `http://localhost:3000/login`.

Shortcut from the repo root:

```bash
make seed-go       # runs go run ./cmd/seed
make run-go        # starts the dev server
```

### All at once

```bash
make seed          # seeds all three stacks
```

This runs `seed-net && seed-django && seed-go`. You must have created auth tables for each stack first (migrations for .NET/Django, `migrate-fiber-auth` for Go).

---

## 4. Dev Accounts

All stacks use the same shared manifest (`docker/postgres/init/seed-uuids/dev-users.json`). Every account uses password `FieldMark!2026`.

| Username | Display Name | Role | Notes |
|----------|-------------|------|-------|
| `aisha` | Aisha Patel | **Admin** | Full system access |
| `marisol` | Marisol Vega | Compliance Officer | Manage compliance rules and scoring |
| `ravi` | Ravi Kumar | Inspector | Perform inspections and file violations |
| `pat` | Pat Smith | Site Supervisor | Oversee site work and resolve violations |
| `kenji` | Kenji Tanaka | Executive | Read-only dashboard and reporting |
| `testuser` | Test User | *(none)* | No assigned role; useful for testing authorization gates |

---

## 5. Verify Authentication

Open a browser to your chosen stack's login page. Sign in with any of the accounts above. After login you should see the home page with your name, role badge, and the application chrome (sidebar, theme toggle, avatar menu).

---

## Troubleshooting

**"404 Not Found" on vendor assets (CSS, JS, fonts):** Your symlinks are broken. See the Windows/WSL notes in [Prerequisites](#prerequisites). Run `ls -la` on your stack's vendor directory — if entries show as regular files instead of `->` symlink arrows, re-checkout or fix your symlink configuration.

**"Cannot connect to database":** Ensure PostgreSQL is running (`make up`) and the container is healthy (`docker ps` should show `fieldmark-local` up). On first start, init scripts take a few seconds.

**"psql: command not found":** The verify script requires `psql`. macOS: `brew install libpq && brew link --force libpq`. Linux: install `postgresql-client`. WSL: `sudo apt install postgresql-client`.

**"Role already exists" or "user already exists":** The seeders are idempotent — re-running them is safe. If you get errors, the table creation step (migrations or `migrate-fiber-auth`) may not have run first.

**Go: "relation fiber_auth.users does not exist":** You skipped `go run ./cmd/migrate-fiber-auth`. Run it before the seed or web server.

**Django: "no such table: auth_user":** You skipped `uv run python manage.py migrate`. Run it before `seed_groups` or `runserver`.

**".NET: Cannot find the fallback endpoint" or port conflicts:** Ensure nothing else is running on :4000. Use `dotnet run --project FieldMark.Web --urls "http://localhost:4001"` to pick a different port.

**CSS changes not showing:** Run `cd fieldmark_shared && pnpm run build` and hard-refresh the browser. If you changed `@source` paths in `src/fieldmark.css`, restart the build.
