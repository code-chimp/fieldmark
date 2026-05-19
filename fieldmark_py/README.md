# FieldMark — Django Project

The Python/Django implementation of FieldMark, built with Django Templates and HTMX.

## Architecture

This project follows a server-authoritative, domain-centric architecture. Business rules live on Django model methods — not in views, signals, middleware, or custom managers. Django is used as an adapter framework, not as an implicit design authority. Detailed rationale is in `_bmad-output/planning-artifacts/research/django-reference.md` and `_bmad-output/planning-artifacts/research/architecture-decisions.md`.

### Project Structure

```
fieldmark_py/
├── manage.py
├── pyproject.toml            Project config (uv)
├── uv.lock                   Dependency lock file
├── pytest.ini                pytest config
├── mypy.ini                  Type checking config
│
├── fieldmark/                Django project package
│   ├── settings.py
│   ├── urls.py
│   ├── wsgi.py
│   └── asgi.py
│
├── static/                   Project-level static assets
│   └── vendor/               Vendored HTMX, AG Grid, shared CSS
├── templates/                Project-level base template
│   └── base.html
│
├── projects/                 Django app — project aggregate
├── inspections/              Django app — inspection aggregate
├── violations/               Django app — violation aggregate and CorrectiveAction
├── audit/                    Django app — audit trail
├── compliance/               Django app — rules engine and scoring
├── reference/                Django app — trade types, categories, rules
└── grid/                     Django app — AG Grid JSON endpoints
```

Apps represent bounded contexts, not technical layers. Each app owns its models, views, urls, forms, templates, and static assets.

### Conceptual Parity with .NET

| Concept | .NET | Django |
|---|---|---|
| Domain logic | FieldMark.Domain project | Model methods / domain modules per app |
| Persistence | EF Core (FieldMark.Data) | Django ORM |
| Composition root | Program.cs (FieldMark.Web) | Django settings + views |
| Admin tooling | Deferred | Django Admin (platform tooling, not product UX) |

Parity is conceptual, not idiomatic. Both stacks expose the same routes, HTMX target IDs, AG Grid endpoint contracts, and audit entry shapes.

## Tech Stack

| Layer | Choice | Version |
|---|---|---|
| Runtime | Python | 3.14+ |
| Web framework | Django | 6.x |
| ORM | Django ORM | bundled |
| Database driver | psycopg | 3.x |
| Database | PostgreSQL | 17 |
| Interactivity | HTMX | 4.x |
| Data grids | AG Grid Community | 35.x |
| Package manager | uv | latest |
| Linting | ruff | latest |
| Formatting | black | latest |
| Type checking | mypy + django-stubs | latest |

## Prerequisites

- [Python 3.14+](https://www.python.org/)
- [uv](https://docs.astral.sh/uv/) (Python package manager)
- PostgreSQL 17 running locally (see root `docker-compose.yml`)

## Getting Started

**1. Start PostgreSQL** (from the repo root):

```bash
docker compose up -d
```

**2. Install dependencies:**

```bash
cd fieldmark_py
uv sync
```

**3. Run migrations:**

```bash
uv run python manage.py migrate
```

**4. Seed conceptual-role Groups (idempotent — safe to re-run):**

```bash
uv run python manage.py seed_groups
```

**5. Run the development server:**

```bash
uv run python manage.py runserver
```

The app will be available at `http://localhost:8000`.

### Creating Migrations

When models change:

```bash
uv run python manage.py makemigrations
```

## Architectural Constraints

The following patterns are **explicitly rejected** and will not be introduced without an ADR amendment:

- Django signals for business logic
- Fat views containing domain rules
- Service layers duplicating model behavior
- Custom managers implementing workflows
- Cross-app side effects
- Implicit transactional coupling
- Client-side state management
- SQLite for tests (use real PostgreSQL)

### What belongs where

**Model methods** — state transition methods (`place_on_hold`, `close`, `approve_resolution`, etc.), domain invariants, `can_*` predicates, and typed exceptions. This is the only place state transitions occur.

**Views** — thin orchestrators: authorize, begin transaction (`transaction.atomic`), load aggregate, invoke model method, append audit entry, recompute compliance score, commit, render template. If a view is doing business logic, it belongs on the model.

**Forms** — input validation that mirrors entity invariants. Forms are not the source of truth; models enforce invariants independently.

**Django Admin** — platform tooling for reference data management. Not part of the product UX. Not required to mirror .NET UI. Over-customization is discouraged.

### Coding Standards

- Use `transaction.atomic` around any view that mutates state.
- Models use `TextChoices` for enum-like fields.
- Use `select_related` / `prefetch_related` aggressively to avoid N+1 queries.
- Class-based or function-based views are both acceptable — stay consistent within an app.
- Use Python's standard `logging` module with structured output (structlog may be added later).

## Request Flow

```
Browser
  │  HTMX request (hx-get / hx-post)
  ▼
Django view
  │  authorize (role check / permission decorator)
  │  load aggregate via ORM
  │  invoke model method
  │  persist (single transaction, includes audit write + recomputation)
  │  render template (partial or full page)
  ▼
Server-rendered HTML → HTMX swaps into DOM
```

For AG Grid views, JSON endpoints return paginated data using the server-side row model. Row selection triggers an HTMX request to load a detail panel.

### CSRF with HTMX

HTMX requests carry the CSRF token via `hx-headers` configured globally on `<body>`:

```html
<body hx-headers='{"X-CSRFToken": "{{ csrf_token }}"}'>
```

## Database & Migration Ownership

PostgreSQL is the shared system of record. All three stacks share a single database using schema-level isolation to avoid migration collisions:

| Schema | Owner |
|---|---|
| `domain` | Infrastructure SQL init scripts — not any framework |
| `django_auth` | This stack |
| `dotnet_auth` | .NET stack |
| `fiber_auth` | Go stack |

Django migrations are scoped to `django_auth` only — Django's built-in auth, admin, and session tables. The `domain` schema is created by SQL init scripts in `docker/postgres/init/` and is not Django's to create or evolve. Domain models must set `managed = False` in their `Meta` class so Django never attempts to create or alter those tables. Dual ownership of domain tables is prohibited.

## Parity

This implementation must remain structurally equivalent to the .NET and Go stacks at all times. A story is not done until all three stacks pass it. A parity test passing on one stack and failing on another is a build-blocking defect.

## Related Documentation

- [Root README](../README.md) — project overview, thesis, domain summary
- [.NET README](../FieldMark/README.md) — the parallel .NET/Razor Pages implementation
- [Go README](../fieldmark-go/README.md) — the parallel Go/Fiber implementation
- [Domain Model](../_bmad-output/planning-artifacts/research/domain-model.md) — entities, state machines, schema
- [Django Architecture Reference](../_bmad-output/planning-artifacts/research/django-reference.md) — Django-specific guardrails
- [Architecture Decisions](../_bmad-output/planning-artifacts/research/architecture-decisions.md) — ADRs and hard constraints
