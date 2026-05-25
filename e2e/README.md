# FieldMark E2E (Playwright)

Shared browser tests for **.NET**, **Django**, and **Fiber** stacks. Same specs under `tests/shared/` run once per backend Playwright **project** with different `baseURL`. See [_bmad-output/planning-artifacts/research/playwright-e2e-philosophy.md](../_bmad-output/planning-artifacts/research/playwright-e2e-philosophy.md) for the guiding philosophy.

## Prerequisites

- Node.js and **pnpm** (or npm) in `e2e/`
- At least one FieldMark backend running locally:
  - **dotnet** — default `http://localhost:4000`
  - **django** — default `http://localhost:8000`
  - **fiber** — default `http://localhost:3000`
- **PostgreSQL** when the stack under test requires it (see root `CLAUDE.md`)

Install browsers once:

```bash
pnpm exec playwright install
```

Optional: copy `.env.example` to `.env` and adjust URLs.

## Commands

From **`e2e/`**:

| Script | Purpose |
|--------|---------|
| `pnpm run test:e2e` | All backend projects (matrix) |
| `pnpm run test:e2e:dotnet` | Project `dotnet` only |
| `pnpm run test:e2e:django` | Project `django` only |
| `pnpm run test:e2e:fiber` | Project `fiber` only |
| `pnpm run test:e2e:ui` | Playwright UI mode |
| `pnpm run test:e2e:codegen` | Record locators (pass URL if needed, e.g. `pnpm exec playwright codegen http://localhost:3000`) |

Single-target debugging avoids failures from backends that are not running; use `--project=fiber` etc.

## Selector policy

Prefer **`getByRole` / `getByLabel`** and parity-locked copy; use **`#project-detail`** and other HTMX IDs from root **CLAUDE.md**; use **`data-testid`** only when stacks diverge. Helpers live in `helpers/selectors.ts`.

## Layout

- `tests/shared/` — behavioral parity tests
- `tests/dotnet|django|fiber/` — backend-specific tests (admin, deferred flows)
