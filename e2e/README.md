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

## Agent Runbook: Sandbox-Safe Playwright

For Codex / Claude style agents running inside a managed sandbox:

1. Start the backend(s) first.
2. Run the narrowest useful Playwright command first:

```bash
pnpm exec playwright test --project=dotnet
pnpm exec playwright test tests/shared/project-transition-flow.spec.ts --project=django
```

3. If browser startup fails with a macOS permission error such as:
   - `Permission denied (1100)`
   - `MachPortRendezvousServer`
   - browser bootstrap / launch denial before any assertions run

   then the failure is probably **sandbox-related**, not a test failure. Re-run the same command with escalation / outside the sandbox.

4. When the spec mutates shared database state, prefer **per-project** runs over the full matrix and reset known fixture rows between runs.

### Important Notes

- The current Playwright config uses `devices["Desktop Chrome"]` as a preset. That does **not** mean the fix is “switch to Chrome vs Chromium”.
- In this repo, the recurring agent failure mode has been **browser launch permission**, and it still applies to standard Playwright Chromium launches.
- If one stack passes only when run alone, inspect shared DB state before changing selectors or app code.

## Selector policy

Prefer **`getByRole` / `getByLabel`** and parity-locked copy; use **`#project-detail`** and other HTMX IDs from root **CLAUDE.md**; use **`data-testid`** only when stacks diverge. Helpers live in `helpers/selectors.ts`.

## Layout

- `tests/shared/` — behavioral parity tests
- `tests/dotnet|django|fiber/` — backend-specific tests (admin, deferred flows)
