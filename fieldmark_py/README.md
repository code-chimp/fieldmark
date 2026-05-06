# FieldMark — Django Project

The Python/Django implementation of FieldMark, built with Django Templates and HTMX.

## Architecture

This project follows a server-authoritative, domain-centric architecture. Business rules live on Django model methods — not in views, signals, middleware, or custom managers. Django is used as an adapter framework, not as an implicit design authority. Detailed rationale is in `docs/FieldMark_Django_Architecture_Reference.md` and `docs/architecture.md` at the repo root.

### Project Structure

```
fieldmark_py/
├── manage.py
├── pyproject.toml            Project config (uv)
├── uv.lock                   Dependency lock file
├── .python-version           Python 3.14
├── .ruff.toml                Linting config
├── mypy.ini                  Type checking config
│
├── fieldmark/                Django project package
│   ├── settings.py
│   ├── urls.py
│   ├── wsgi.py
│   ├── asgi.py
│   └── static/vendor/        HTMX, AG Grid, shared vendor assets
│
├── projects/                 Django app — project aggregate (planned)
├── inspections/              Django app — inspection aggregate (planned)
├── violations/               Django app — violation aggregate (planned)
├── audit/                    Django app — audit trail (planned)
├── compliance/               Django app — rules engine and scoring (planned)
├── reference/                Django app — trade types, categories, rules (planned)
└── grid/                     Django app — AG Grid JSON endpoints (planned)
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
| Interactivity | HTMX | 2.x |
| Data grids | AG Grid Community | 32.x |
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

**4. Run the development server:**

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
- Structured JSON logging via `structlog`.

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

PostgreSQL is the shared system of record. In the dual-stack setup, both stacks point at the same database but use distinct schemas (`fm_dotnet`, `fm_python`) to avoid migration collisions during parallel development.

Django migrations own Django framework tables, auth tables, and admin support tables. Domain schema authority is defined in `docs/domain-model.md`. Dual ownership of domain tables is prohibited.

## Parity

This implementation must remain structurally equivalent to the .NET stack at all times. A story is not done until both stacks pass it. A parity test passing on one stack and failing on the other is a build-blocking defect.

## Related Documentation

- [Root README](../README.md) — project overview, thesis, domain summary
- [.NET README](../FieldMark/README.md) — the parallel .NET implementation
- [Architecture](../docs/architecture.md) — full architecture with data flow patterns
- [Domain Model](../docs/domain-model.md) — entities, state machines, schema
- [Django Architecture Reference](../docs/FieldMark_Django_Architecture_Reference.md) — Django-specific guardrails
