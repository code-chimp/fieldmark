# FieldMark Project-Level E2E Playwright Harness

## Purpose

This document defines a **single, shared Playwright end-to-end (E2E) test project** for the FieldMark monorepo.

The goal is to provide one browser-based test harness that can validate equivalent user workflows across multiple FieldMark implementations, including:

- ASP.NET Core / Razor Pages
- Django
- Go / Fiber
- future comparison clients if added later

This harness is intentionally **black-box and behavior-first**:

- it tests the application through the browser
- it validates workflows and visible state
- it does not encode framework internals

This approach reinforces the central FieldMark thesis:

> One authoritative domain, multiple replaceable application projections.

---

## Why a Shared Root-Level Playwright Project

A single E2E project at the monorepo root is preferred because it:

- avoids duplicating test infrastructure in each backend
- validates equivalent behavior across all implementations
- makes apples-to-apples comparison possible
- keeps tests aligned to product behavior rather than framework details
- provides a single quality gate for shared workflows

The browser becomes the contract.

---

## Monorepo Placement

Recommended placement:

```text
fieldmark/
├── FieldMark/                 # .NET implementation
├── fieldmark_py/              # Django implementation
├── fieldmark_go/              # Fiber implementation
├── e2e/                       # Shared Playwright project
├── docker/
├── docker-compose.yml
├── docs/
└── ui/
```

The Playwright project should live at the repository root as a peer to all application implementations.

---

## High-Level Testing Strategy

The E2E harness should support two categories of tests:

### 1. Shared Behavioral Tests

These are the most important tests.

They validate workflows that should behave the same regardless of backend implementation.

Examples:

- dashboard loads and shows key metrics
- projects list is navigable
- project detail shows inspections and violations
- resolving a violation updates visible state
- audit log reflects workflow actions

### 2. Implementation-Specific Tests

These are secondary tests.

They validate behavior that is intentionally framework-specific.

Examples:

- Django admin login works
- .NET admin/config pages render
- Fiber-only implementation details if introduced later

Shared product workflows should remain the primary focus.

---

## Recommended Folder Layout for the Playwright Project

```text
e2e/
├── package.json
├── playwright.config.ts
├── tsconfig.json
├── .env.example
├── tests/
│   ├── shared/
│   │   ├── dashboard.spec.ts
│   │   ├── projects.spec.ts
│   │   ├── project-detail.spec.ts
│   │   ├── inspections.spec.ts
│   │   └── violations.spec.ts
│   ├── dotnet/
│   │   └── admin.spec.ts
│   ├── django/
│   │   └── admin.spec.ts
│   └── fiber/
│       └── smoke.spec.ts
├── fixtures/
│   ├── apps.ts
│   ├── auth.ts
│   └── pages.ts
├── helpers/
│   ├── navigation.ts
│   ├── selectors.ts
│   ├── workflows.ts
│   └── assertions.ts
└── reports/
```

This structure separates:

- shared tests
- backend-specific tests
- reusable fixtures and helpers

---

## Core Design Principle: Shared Selector Contract

A shared E2E harness becomes dramatically easier if all implementations expose stable selectors using `data-testid` attributes.

### Example shared selectors

```text
data-testid="dashboard-compliance-score"
data-testid="dashboard-open-violations"
data-testid="project-list-grid"
data-testid="project-row"
data-testid="project-detail"
data-testid="inspection-list"
data-testid="violation-list"
data-testid="resolve-violation-button"
data-testid="audit-log"
```

### Rule

All FieldMark implementations should adopt the same selector contract for shared UX surfaces.

This avoids coupling tests to framework-specific DOM structure, CSS classes, or incidental markup differences.

---

## Runtime Targeting Model

The Playwright project should support multiple application targets via configuration.

### Recommended backend targets

```text
dotnet  -> http://localhost:5000
django  -> http://localhost:8000
fiber   -> http://localhost:3000
```

### Execution modes

#### Mode A: single-target execution

Run the shared suite against one backend at a time.

Use this for:
- local debugging
- initial bring-up
- troubleshooting specific implementations

#### Mode B: multi-project execution

Define separate Playwright projects for:
- dotnet
- django
- fiber

Then run the same shared tests against all of them in one command.

Use this for:
- comparison testing
- regression detection
- CI later if desired

---

## Recommended Playwright Configuration Shape

### Conceptual projects

```text
projects:
- name: dotnet
  use.baseURL: http://localhost:5000

- name: django
  use.baseURL: http://localhost:8000

- name: fiber
  use.baseURL: http://localhost:3000
```

### Recommended defaults

- headless by default
- trace on failure
- screenshot on failure
- video retained on failure only
- retries enabled only in CI if later added

The root-level harness should remain simple and deterministic.

---

## Example Shared Test Flows

The following workflows are recommended as the initial shared E2E contract.

### 1. Dashboard Smoke Test

Validate that:
- the dashboard route loads
- summary metrics render
- key widgets are visible
- no client-side bootstrapping failure occurs

### 2. Project List Navigation

Validate that:
- the project list renders
- at least one project row is visible
- clicking a project navigates or drills into project detail

### 3. Project Detail Integrity

Validate that:
- project detail loads
- inspections section is present
- violations section is present
- compliance score is visible

### 4. Violation Resolution Flow

Validate that:
- a violation can be opened
- the resolve action is visible when allowed
- the action updates the visible status
- audit log updates after the action
- compliance indicators change if expected

### 5. Audit Visibility

Validate that:
- audit log fragment or section is visible
- at least one domain action is recorded
- ordering is newest-first if expected

These tests are strong because they validate FieldMark’s key architecture message:

- the backend owns workflow
- state changes are visible
- the UI reflects authoritative server decisions

---

## Handling Framework-Specific Authentication in E2E

Authentication is framework-local, so E2E tests must account for that.

### Recommended strategy

- shared product tests should prefer pre-authenticated fixtures or test users
- login helpers should be backend-specific
- the shared test flow should not assume one common login system

### Suggested helper approach

```text
helpers/auth/
- login-dotnet.ts
- login-django.ts
- login-fiber.ts
```

Each helper performs whatever login steps are needed for that backend.

Product workflow tests should call a backend-appropriate login helper, not inline login steps repeatedly.

---

## Admin Testing Guidance

Admin functionality should not be part of the primary shared product workflow suite.

Reason:
- Django admin is framework tooling
- .NET admin pages are custom product-adjacent tooling
- Fiber may not yet have equivalent screens

Therefore:

- keep admin tests separate
- treat them as backend-specific
- do not force artificial parity in the shared suite

---

## Data Management Strategy for E2E

Recommended initial approach:

- use a known local database state
- seed predictable demo records
- keep tests idempotent where possible

### Good first approach

- bootstrap a small seed dataset in the shared domain schema
- ensure each backend renders the same sample projects, inspections, and violations
- do not require test-created data at first

Later, if needed, add:
- test data reset scripts
- isolated test user setup
- fixture-level cleanup

---

## Suggested Initial NPM Scripts

Examples of the scripts the root Playwright project should expose:

```text
npm run test:e2e:dotnet
npm run test:e2e:django
npm run test:e2e:fiber
npm run test:e2e:all
npm run test:e2e:ui
npm run test:e2e:codegen
```

These scripts make the harness easy for humans and agents to use consistently.

---

## Bring-Up Sequence for the E2E Harness

### Phase 1

- create the root `e2e/` project
- install Playwright
- add one smoke test against one backend

### Phase 2

- define and adopt shared `data-testid` attributes
- add shared navigation and assertion helpers

### Phase 3

- add Playwright projects for Django, .NET, and Fiber
- run shared tests against all three

### Phase 4

- add backend-specific auth helpers
- add secondary admin and platform tests

This sequence keeps the test harness from becoming over-engineered too early.

---

## Guardrails for Human and Agent Contributors

All contributors must follow these rules when working on the E2E project:

1. Prefer shared behavioral tests over implementation-specific tests
2. Use `data-testid` selectors, not CSS class selectors
3. Avoid coupling tests to framework-specific DOM structure
4. Keep login helpers backend-specific and reusable
5. Do not duplicate workflows across target folders unless truly necessary
6. Keep admin testing separate from shared product workflow testing
7. Favor deterministic local test data over ad hoc creation flows early on

---

## Success Criteria

The Playwright harness is considered successfully established when:

- the root `e2e/` project runs independently of any one backend repository
- one shared smoke test can run against each backend
- selector conventions are documented and adopted
- at least one shared product workflow is validated across all implementations

At that point, the harness becomes a reusable quality layer for the entire FieldMark monorepo.

---

## Final Guidance

The shared Playwright harness should be viewed as:

- a product behavior contract
- a backend-neutral test surface
- a proof that the same domain workflows can be expressed by multiple stacks

This makes it one of the strongest architectural supporting artifacts in the whole repository.

---

## Status

Drafted – Project-level Playwright E2E harness strategy
