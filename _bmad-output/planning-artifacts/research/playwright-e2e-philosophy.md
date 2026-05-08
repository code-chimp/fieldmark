# FieldMark Playwright E2E — philosophy primer

**Audience:** Test Architect, BMAD planning agents, and contributors aligning FieldMark’s browser automation with architecture constraints.

**Operational checklist:** [docs/FieldMark_Project_Level_Playwright_E2E_Harness.md](../../../docs/FieldMark_Project_Level_Playwright_E2E_Harness.md).  
**Harness implementation:** [e2e/README.md](../../../e2e/README.md).

---

## Purpose

FieldMark uses **one shared Playwright project** at the monorepo root (`e2e/`). The browser exercises **product-visible behavior** across **multiple backends** (.NET, Django, Fiber) against the **same domain**. The browser is the **behavior contract**—proof that alternative stacks remain interchangeable projections of one authoritative domain.

---

## Philosophy

- **Black-box:** Drive the real application through HTTP/HTML; do not assert framework internals (ORM wiring, DI graphs, handler signatures).
- **Behavior-first:** Focus on workflows and rendered state that matter to users and auditors (dashboards, lists, transitions, audit visibility).
- **Backend-neutral assertions:** Prefer locators that survive stack differences (see Locator strategy below).
- **Aligns with thesis:** *One authoritative domain, multiple replaceable application projections.*

---

## Test tiers

| Tier | Scope | Examples |
|------|--------|----------|
| **Shared behavioral** | Same user journeys — must behave equivalently on every stack | Dashboard smoke, project drill-down, violation lifecycle, audit trail visibility |
| **Implementation-specific** | Framework tooling or deliberate divergence | Django Admin, .NET Identity/admin surfaces, Fiber-only interim smoke |

**Primary investment** belongs in **shared** suites. **Admin** and platform tooling tests stay **secondary**, isolated under stack folders — they must not dilute parity coverage.

---

## Boundaries vs unit tests

- **Unit tests** prove domain rules and application logic **inside** each codebase (idiomatic pytest / xUnit / `go test`). See `_bmad-output/planning-artifacts/research/architecture-decisions.md` (**Unit testing & E2E boundaries**).
- **E2E tests** prove **cross-surface workflows** and **visible outcomes** after HTTP round-trips.
- **Do not duplicate** full unit scenarios in Playwright; avoid testing identical rule matrices twice unless a regression uniquely manifests in the browser.

---

## Targeting model

- Each backend runs on a distinct **base URL** (defaults: .NET `5000`, Django `8000`, Fiber `3000` — see harness doc).
- Playwright **projects** (`dotnet`, `django`, `fiber`) share **`tests/shared/**`** so the same spec runs **three times** with different `baseURL`.
- **Single-target runs** (`--project=fiber`) support local debugging without starting every stack.

---

## Locator strategy

Order of preference (details in `e2e/helpers/selectors.ts`):

1. **Accessibility / semantic locators** — `getByRole`, `getByLabel`, `getByPlaceholder`, `getByText` where copy and roles are **parity-locked** across stacks (matches Playwright’s recommended practice).
2. **Stable HTMX / layout IDs** required by root **`CLAUDE.md`** (`#project-detail`, `#compliance-tile`, `#violation-detail`, `#audit-log`) — cross-stack contract for swaps and panels.
3. **`data-testid`** — use **only** when markup or accessible names **differ by stack** but the behavior under test must remain comparable.

Avoid coupling tests to **CSS classes** or incidental DOM shape unless necessary.

---

## Authentication

Auth is **framework-local** (ADR-012). Shared flows must **not** assume a single login mechanism. Plan **per-stack login helpers** under `e2e/helpers/auth/` (or equivalent) when auth exists; product tests call the helper appropriate to the active Playwright project.

---

## Data

Early phases should rely on **predictable seeded domain data** (Postgres init / seed scripts) so lists and violations are stable. Prefer **idempotent** assumptions; introduce dedicated reset or isolation scripts only when parallel runs or CI demand it.

---

## Success criteria (harness-level)

- Shared smoke passes against **each** backend when that stack is running and seeded.
- Locator and URL contracts are **documented** (root `CLAUDE.md`, `e2e/helpers/selectors.ts`, harness doc).
- At least one **non-trivial shared workflow** (beyond smoke) validates parity across stacks when features exist.

---

## References

- [docs/FieldMark_Project_Level_Playwright_E2E_Harness.md](../../../docs/FieldMark_Project_Level_Playwright_E2E_Harness.md) — folder layout, scripts, ports, bring-up phases.
- [CLAUDE.md](../../../CLAUDE.md) — HTMX target IDs, three-stack symmetry, PostgreSQL rules.
- [architecture-decisions.md](architecture-decisions.md) — unit vs E2E boundaries.
- [e2e/README.md](../../../e2e/README.md) — how to run tests locally.

---

**Status:** Priming artifact for agentic Test Architect / planning sessions — keep concise; defer procedural detail to the harness doc and `e2e/README.md`.
