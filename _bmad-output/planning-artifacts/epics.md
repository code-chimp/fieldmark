---
stepsCompleted: ["step-01-validate-prerequisites", "step-02-design-epics", "step-03-create-stories", "step-04-final-validation"]
inputDocuments:
  - _bmad-output/planning-artifacts/prd/index.md
  - _bmad-output/planning-artifacts/prd/functional-requirements.md
  - _bmad-output/planning-artifacts/prd/non-functional-requirements.md
  - _bmad-output/planning-artifacts/prd/product-scope.md
  - _bmad-output/planning-artifacts/prd/user-journeys.md
  - _bmad-output/planning-artifacts/prd/architectural-constraints-prd-binding.md
  - _bmad-output/planning-artifacts/prd/web-app-specific-requirements.md
  - _bmad-output/planning-artifacts/prd/success-criteria.md
  - _bmad-output/planning-artifacts/architecture.md
  - _bmad-output/planning-artifacts/ux-design-specification.md
---

# FieldMark - Epic Breakdown

## Overview

This document provides the complete epic and story breakdown for FieldMark, decomposing the requirements from the PRD, UX Design Specification, and Architecture into implementable stories. FieldMark is delivered as three parallel stacks (.NET Razor Pages, Django, Go/Fiber) against a single infrastructure-owned PostgreSQL `domain` schema; **a story is not done until all three stacks pass it.**

## Requirements Inventory

### Functional Requirements

**Authentication & Identity**
- **FR1:** A user can authenticate against the application using credentials managed by the framework's native authentication system. Authentication is framework-local per ADR-012; cross-stack login session sharing is explicitly not required.
- **FR2:** The system can identify the authenticated user on every request and resolve their conceptual role(s) (Administrator, Compliance Officer, Inspector, Site Supervisor, Executive).
- **FR3:** A user can log out, terminating their session within the stack they are authenticated against.
- **FR4:** Unauthenticated users are redirected to the framework-local login page when attempting to access any business route.

**Authorization**
- **FR5:** The system can determine, for any given action on any given entity, whether the current user is authorized to perform it.
- **FR6:** The system never renders an action button (or other action affordance) for an action the current user is not authorized to perform.
- **FR7:** The system rejects any direct request (HTTP-level) to perform an action the user is not authorized for, regardless of UI state, returning an authorization-failure response.
- **FR8:** Role assignments per user are managed by the framework-native authorization machinery in each stack; the application reads them, it does not own a shared role schema.

**Project Management**
- **FR9:** A Project Manager or Administrator can create a new Project with required metadata (code, name, start date, optional target completion date, trade scope assignments, inspector assignments).
- **FR10:** A Project Manager, Compliance Officer, or Administrator can view a list of Projects, filtered and sorted by status and compliance score.
- **FR11:** Any authorized user can view a Project Detail page showing the project's current status, compliance score, assigned trades, inspections, violations, and audit log — all in a single rendered view with HTMX-driven tab swaps.
- **FR12:** A Project Manager can transition a Project from Active to OnHold with a recorded reason.
- **FR13:** A Project Manager can transition a Project from OnHold back to Active.
- **FR14:** A Project Manager can transition a Project from Active to Closed only when all closure-gate rules are satisfied; the system blocks closure with a rendered explanation when gates are unmet.
- **FR15:** The system displays the result of `can_close()` as the rendered state of the Close action (button absent for unauthorized users, disabled with explanation for users authorized but blocked, enabled when permitted).

**Inspection Management**
- **FR16:** A Compliance Officer or Administrator can schedule an Inspection on a Project, specifying trade type, inspector, and scheduled time.
- **FR17:** An Inspector can transition an Inspection from Scheduled to InProgress (start it).
- **FR18:** An Inspector can transition an Inspection from InProgress to Completed, supplying an outcome (Pass, Fail, or Conditional), notes, and zero or more findings.
- **FR19:** A Compliance Officer or Administrator can cancel a Scheduled Inspection with a recorded reason.
- **FR20:** When an Inspection is Completed with Fail-class findings, the system automatically opens a Violation for each finding, atomically within the same transaction.
- **FR21:** The system displays Inspection lists scoped to a Project and to the current user's role, filterable by status and date range.

**Violation Management**
- **FR22:** The system records a Violation's origin Finding, severity, due date, and current status.
- **FR22a:** The Violation's due date is computed at open time from its severity and is immutable thereafter; no UI or API path supports modifying due date after the Violation has been opened.
- **FR23:** A Compliance Officer can assign or reassign a Violation to a Site Supervisor; reassignment while in InProgress is a self-transition that emits an audit entry.
- **FR24:** A Site Supervisor or assigned user can view the Violations assigned to them, filterable by status and overdue flag.
- **FR25:** The system marks a Violation as overdue when its due date has passed and its status is non-terminal.
- **FR26:** An Administrator can void a Violation in any non-terminal state with a recorded reason; voided violations do not affect compliance scoring.
- **FR27:** The system prevents reopening of a Resolved or Voided Violation; users who require revisiting the underlying issue must create a new Violation from a new Finding.

**Corrective Action Management**
- **FR28:** A Site Supervisor (or assigned remediator) can submit a Corrective Action for a Violation in InProgress state, including a description and optional evidence reference.
- **FR29:** A Compliance Officer can take a submitted Corrective Action for review, transitioning its state to UnderReview.
- **FR30:** A Compliance Officer can approve a Corrective Action in UnderReview state, atomically transitioning the parent Violation to Resolved within the same transaction.
- **FR31:** A Compliance Officer can reject a Corrective Action in UnderReview state with review notes; the parent Violation remains in InProgress and a new Corrective Action may be submitted.
- **FR32:** The system prevents the submitter of a Corrective Action from being its reviewer.
- **FR33:** Only the latest non-Rejected Corrective Action of a Violation may be approved.

**Compliance Rules Engine & Scoring**
- **FR34:** The system evaluates compliance rules entirely server-side; clients never compute compliance state.
- **FR35:** The system recomputes a Project's compliance score on every state transition that affects it (Violation opened, resolved, voided; Corrective Action approved; Inspection completed) within the same transaction as the triggering write.
- **FR36:** The system surfaces the current compliance score of a Project on the Project Detail screen and updates it via HTMX out-of-band swap when a same-page action affects it.
- **FR37:** The system enforces the closure gate rules (`OpenViolationGate`, `RequiredInspectionPerTrade`) at the moment of attempted closure; closure is rejected with an inline explanation when any gate is unmet.
- **FR38:** Compliance rules and their parameters (severity weights, due-offset values) are persisted in reference data and evaluated dynamically; changing a parameter changes evaluation behavior without code changes. *(MVP persistence + evaluation; admin UI is Growth.)*

**Audit Trail**
- **FR39:** The system writes an AuditEntry for every domain mutation in the same database transaction as the change.
- **FR40:** Each AuditEntry records the actor (opaque user UUID per ADR-012), action, entity type and ID, the project it belongs to (when applicable), the before/after state as JSON, and any metadata (reason strings, role at time of action).
- **FR41:** AuditEntries are append-only; the system does not provide any UI or API path to update or delete an existing entry.
- **FR42:** Any authorized user can view a Project's audit log, ordered most-recent-first, accessible from the Project Detail screen.
- **FR43:** Executive-role users can view the audit log read-only, with no path to mutate any entity from that view.

**Dashboard & Reporting**
- **FR44:** A user can view the Compliance Dashboard showing portfolio-level aggregates: average compliance score, count of overdue violations by severity, count of Projects by lifecycle state.
- **FR45:** The Compliance Dashboard renders via HTMX-driven partial refresh; tile-level updates do not require full-page reloads.
- **FR46:** A user can drill from the Compliance Dashboard into a Project Detail screen via HTMX swap; navigation does not trigger a full-page reload.
- **FR47:** A user can filter and sort the Project list by status, compliance score, and ownership.

**Data Grid Interactions (AG Grid)**
- **FR48:** The system serves AG Grid data via a server-side row model on at least two views (Project list and one of: Inspection list, Violation list, Audit log).
- **FR49:** AG Grid endpoints return JSON conforming to the cross-stack contract `{ "rows": [...], "lastRow": N }`.
- **FR50:** Row selection in any AG Grid fires an HTMX request that loads the corresponding detail panel; the grid does not own the detail rendering.
- **FR51:** AG Grid configurations contain no business rules and no client-side row computation; filters and sorts that affect server data are passed to the server as request parameters.

**Reference Data Administration**
- **FR52:** An Administrator can view the catalog of Trade Types, Violation Categories, and Compliance Rules. *(Read access in MVP; full CRUD UI is Growth.)*
- **FR53:** Reference data is loaded from the database on application start and on changes; no application restart is required to pick up reference-data changes.

**Cross-Cutting System Behavior**
- **FR54:** State-changing actions in the UI use HTTP POST (or appropriate non-GET method); GET requests never mutate state.
- **FR55:** When a domain rule rejects an action, the system returns HTTP 409 with the originating partial re-rendered showing the error and the unchanged state.
- **FR55a:** When client-submitted data fails server-side input validation (malformed input, type error, missing required field — distinct from a domain rule violation), the system returns HTTP 422 with the originating form partial re-rendered showing field-level errors. Each invalid field carries `aria-invalid="true"` and `aria-describedby` linking to its inline error message; a top-level InlineAlert with `role="alert"` summarises the error count and links to the first invalid field. No state is mutated on 422.
- **FR56:** When an authorization check rejects an action, the system returns the appropriate authorization-failure response (HTTP 403 or framework-equivalent) without leaking information about the underlying entity state.
- **FR57:** All domain mutations occur within a transaction that includes the audit-entry write and any required compliance-score recomputation.
- **FR58:** Identical routes, HTMX target IDs, AG Grid endpoint contracts, audit action strings, and domain method names are present across all three stacks (modulo language casing); a diff in any of these is a defect.
- **FR59:** Domain rule violations, authorization failures, and validation errors are handled identically (in observable behavior) across all three stacks.

**Accessibility**
- **FR60:** All interactive controls are keyboard-operable with the expected activation keys; tab order matches visual order; focus is visible.
- **FR61:** Form errors are programmatically associated with their inputs via `aria-invalid` and `aria-describedby` and are announced to assistive technology.
- **FR62:** HTMX swaps that represent meaningful state changes shift focus to the swapped region (or to a documented focus-target element within it) so that screen-reader users perceive the change.
- **FR63:** Out-of-band swap targets (compliance score tile, notification badges) carry `aria-live` so updates are announced to assistive technology.
- **FR64:** Buttons that trigger HTMX requests visibly indicate their disabled state during the request to both visual and assistive-technology users.

**Test & Quality (Capability Requirements)**
- **FR65:** Every MVP user-facing workflow has a Playwright E2E scenario that runs against all three stacks and asserts equivalent observable behavior.
- **FR66:** Every domain method with state-transition logic has unit-test coverage proving its invariants in each stack's idiomatic test framework.

**Growth-Phase Capabilities (Out of MVP — for visibility, not for MVP planning)**
- **FR67:** *(Growth)* Administrator CRUD UI for reference data.
- **FR68:** *(Growth)* Executive portfolio-level trend visualizations.
- **FR69:** *(Growth)* Cross-stack parity comparison report.
- **FR70:** *(Growth)* Editable compliance rule parameters via UI.

### NonFunctional Requirements

- **NFR1 — Performance (locked targets):** HTMX partial-swap perceived latency ≤ 200 ms p95 (local dev); AG Grid row-select → detail panel ≤ 300 ms p95; no full-page reload on any state-changing action; cross-stack latency divergence > 50 ms p95 on the same scenario is a defect; compliance-score recomputation in same transaction as the triggering write (no follow-up request).
- **NFR2 — Security (server-authoritative):** Server-side authority for all rules, validation, and authorization on every request. Authentication framework-local per ADR-012 (no shared identity, no SSO, no third-party IdP). All client input treated as untrusted (UI button absence is *not* the auth mechanism). Framework-native salted password hashing. Framework-native CSRF on state-changing requests. Parameterized SQL only — no string-concatenated SQL.
- **NFR3 — Accessibility (WCAG 2.1 AA across all three stacks):** Enforced via `@axe-core/playwright` in every E2E scenario. HTMX-specific commitments: focus management on swaps, `aria-live` on OOB targets, `hx-disabled-elt` request-pending state, errors associated via `aria-describedby`.
- **NFR4 — Reliability (local-dev artifact):** Clean start from `docker compose up -d`; recovers from DB restart without data loss (transactions durable); no silent corruption on transaction abort (pre-mutation state preserved); zero external services beyond PostgreSQL (no broker, cache, email).
- **NFR5 — Data Integrity:** Every domain mutation occurs in one DB transaction including audit-entry write and compliance-score recomputation. `domain` schema enforces structural invariants via CHECK constraints (status enums, score range 0–100, severity ranges, completion-implies-outcome, voided-implies-reason) as defense-in-depth — primary enforcement remains on entity methods. Audit append-only (no UI/API update/delete). UUIDs generated in app code (not `gen_random_uuid()`). Timestamps stored as `TIMESTAMPTZ` (UTC).
- **NFR6 — Maintainability / Readability:** Code readability is a first-class quality attribute. Cross-stack rule is *structural symmetry, not naming-convention identity* (C# `PascalCase`, Python `snake_case`, Go `PascalCase`/`camelCase`, all mapping to `snake_case` DB and JSON). Canonical method-name list (`start`, `complete`, `cancel`, `place_on_hold`, `resume`, `close`, `assign`, `submit_corrective_action`, `approve_resolution`, `reject_resolution`, `void`) — divergence is a defect. Comments explain *why*. Per-stack auto-formatters + linters are build-blocking (`dotnet format` + analyzers; `ruff`+`black`+`mypy`; `gofmt`+`golangci-lint`).
- **NFR7 — Portability / Cross-Stack Compatibility:** The `domain` PostgreSQL schema is the contract; no framework owns it. HTMX 4.x and AG Grid Community 35.x versions pinned identically across stacks (mismatch is build-blocking). Compiled Tailwind CSS authored once in `fieldmark_shared/`, symlinked into all three apps. `pg_indexes` snapshot diff across stacks = zero. Route inventory diff across stacks = zero (modulo casing).
- **NFR8 — Observability (minimal, audit-centric):** AuditEntry is the primary observability mechanism for domain events. Framework-default HTTP request logging is sufficient. No metrics endpoint, no `/healthz`, no Prometheus, no tracing, no Sentry. Production observability is an explicit non-goal.
- **NFR9 — i18n: Not applicable.** All UI strings and audit action strings in English; locale-default formatting per browser. Translation infrastructure out of scope.
- **NFR10 — Browser Compatibility:** Last 2 stable versions of Chrome / Firefox / Safari / Edge. IE not supported. Mobile browsers best-effort. A feature working in Chromium but not in Safari is a defect.
- **NFR11 — Explicit Non-Goals (do not generate stories for these):** Multi-tenancy, horizontal scaling, HA/failover, production hosting, CI/CD beyond test runners, compliance certification, structured logging/log shipping, i18n/RTL, mobile-native apps, offline/PWA, real-time collaboration (WebSocket/SSE), file uploads (evidence_ref is a string placeholder), email/SMS/push, full-text search, PDF/CSV/report export, audit retention/purge policy.

### Additional Requirements

Technical & infrastructure requirements derived from the Architecture document that affect epic/story creation.

**Foundation / Starter (no third-party starter applies — native CLI scaffolding only):**
- Three skeleton projects scaffolded per Architecture §Initialization Commands: `FieldMark/` (.NET 10, Razor Pages + Domain/Data/Web class libs + xUnit), `fieldmark_py/` (Django 6, `uv`, app-per-aggregate), `fieldmark-go/` (Fiber v3 + `pgx/v5`, `cmd/web` + `internal/{domain,data,app,web}`). Skeletons already exist in repo — Story 1.1 confirms presence, not greenfield scaffolding.
- Cross-stack tooling repos: `e2e/` (Playwright + axe-core + biome), `fieldmark_shared/` (Tailwind v4 + Basecoat + vendored HTMX + AG Grid), `docker/` (Postgres 17 Compose harness).
- Top-level `Makefile` is the single source of truth for run/test/parity commands (D20).

**Database & data layer:**
- PostgreSQL 17, single instance, Docker Compose only.
- Schemas: `domain`, `django_auth`, `dotnet_auth`, `fiber_auth`, `infra` — created by SQL in `docker/postgres/init/001_schemas.sql`.
- Canonical `domain.*` DDL lives in `docker/postgres/init/010_domain_tables.sql` (hand-authored; **not** generated by any framework's migrations).
- Indexes inventory in `docker/postgres/init/020_domain_indexes.sql`; reference seed in `090_seed_reference.sql`; dev-users seed in `091_seed_dev_users.sql`.
- Same-UUID seed strategy: shared manifest `docker/postgres/init/seed-uuids/dev-users.json`; per-stack runners idempotently write to each stack's `*_auth` schema using identical UUIDs.
- Connection string standardized as `FIELDMARK_DATABASE_URL` (Postgres URL form) across all stacks; local default `postgresql://fieldmark:fieldmark@localhost:5432/fieldmark`.
- ORM/data conventions: EF Core 9.x (Npgsql provider + `EFCore.NamingConventions.UseSnakeCaseNamingConvention()`) with `FieldMark.Data/Configuration/` fluent configs; Django `Meta.managed = False` on `domain.*` models with `db_table = 'domain"."<table>'`; Go narrow per-aggregate `Store` interfaces in `internal/data/` with `pgx/v5` and explicit SQL.

**Authentication wiring:**
- .NET: ASP.NET Core Identity, schema target `dotnet_auth`, snake_case table mapping, password policy (length ≥ 10, digit + lowercase + uppercase required, non-alphanumeric not required); `modelBuilder.HasDefaultSchema("dotnet_auth")` on the Identity DbContext.
- Django: built-in `auth` with `django_auth` schema target via DB router or `db_table`; conceptual roles seeded as Django Groups (`ADMIN`, `COMPLIANCE_OFFICER`, `INSPECTOR`, `SITE_SUPERVISOR`, `EXECUTIVE`).
- Go / Fiber: **deferred** per ADR-012; MVP uses stub middleware injecting a configurable `actor_id` UUID. Real Go auth is a follow-on epic, not MVP scope.

**HTTP/HTMX/Grid contracts (canonical across all three stacks):**
- AG Grid endpoints: `POST /grid/projects`, `POST /grid/violations`, `POST /grid/inspections`, `POST /grid/audit/:projectId` — request body is AG Grid SSRM payload; response `{ "rows": [...], "lastRow": N }`.
- HTMX target ID inventory (full canonical list): `#compliance-tile`, `#project-detail`, `#project-list`, `#violation-detail`, `#violation-list`, `#inspection-list`, `#audit-log`, `#corrective-action-form`, `#corrective-action-list`, `#flash-region`. New target ID requires an ADR amendment.
- Canonical audit action strings: `ProjectCreated`, `ProjectClosed`, `ProjectPlacedOnHold`, `ProjectResumed`, `InspectionScheduled`, `InspectionStarted`, `InspectionCompleted`, `InspectionCancelled`, `ViolationOpened`, `ViolationAssigned`, `ViolationVoided`, `CorrectiveActionSubmitted`, `CorrectiveActionTakenForReview`, `CorrectiveActionApproved`, `CorrectiveActionRejected`. **`ProjectCreated` added by ADR amendment during epic planning** for forensic completeness on project creation. Stored verbatim; further additions are ADR amendments. (15 strings total.)
- Error rendering: single typed `DomainRuleException` per stack → handler re-renders originating partial with HTTP 409 + inline error + unchanged state. No global exception middleware for domain errors. Authorization failures bubble to framework-native 403 without entity-state leakage.

**Frontend / assets:**
- Tailwind v4 compiled in `fieldmark_shared/` from `src/fieldmark.css` → `dist/fieldmark.css` (committed). Symlinked into each app's static dir.
- Vendored HTMX 4.x and AG Grid Community 35.x under `fieldmark_shared/vendor/`; symlinked. No CDN.
- AG Grid Quartz theme + project overrides compiled into `fieldmark.css`.

**Cross-stack parity tooling (replaces CI for MVP per D18):**
- `tools/parity/` shell scripts: `dump-pg-indexes.sh`, `dump-routes-net.sh`, `dump-routes-django.sh`, `dump-routes-fiber.sh`, `diff-routes.sh`, `diff-pg-indexes.sh`.
- Each stack exposes a `--dump-routes` subcommand (`dotnet run --project FieldMark.Web -- --dump-routes`, `manage.py show_urls`, `go run ./cmd/web -dump-routes`).
- `make parity` runs both diff scripts; exits non-zero on any divergence.
- Optional pre-commit hook sample at `tools/git-hooks/pre-commit.sample` runs `make parity` on commits touching any stack.
- README copy in each stack instructs running `make parity` before committing routing/schema/HTMX-id changes.

**Configuration:**
- Env-var driven only — no committed `.env`, no secrets vault. Required: `FIELDMARK_DATABASE_URL`, `FIELDMARK_LOG_LEVEL` (default `info`).
- Connection pooling: .NET `AddDbContextPool<>` default 100; Django `CONN_MAX_AGE = 60`; Go `pgxpool` size = 4× CPU. No PgBouncer.

**Testing infrastructure:**
- Real PostgreSQL only — no SQLite in tests (Testcontainers for .NET; pytest-django; Go integration build tag).
- Single Playwright suite at `e2e/`, parallel projects per stack, `@axe-core/playwright` in every scenario; biome formats TS.

**Implementation sequence (per Architecture §Decision Impact):** (1) `010_domain_tables.sql` unblocks every stack; (2) `tools/parity/` + Makefile before code drifts; (3) EF Core fluent config + Project mapping (proof); (4) Django `Meta.managed=False` Project model (proof); (5) Go `ProjectStore` + pgx (proof); (6) Anchor Workflow MVP epic falsifies/confirms thesis on one stack first.

**Explicitly rejected for MVP (do not generate stories):** AutoMapper / Mapster, Repository or UoW abstractions, CQRS / MediatR / in-process buses, client-side state stores, SQLite in tests, fat service layer in .NET/Django (Go has a thin `app` coordination layer by design — wiring only, never business rules), CI/CD pipelines (parity tooling replaces CI for MVP; CI graduates only when artifact is shared externally or a second contributor joins).

### UX Design Requirements

Actionable design system, component, pattern, and accessibility work derived from the UX Design Specification. Each item is sized to generate at least one cross-stack story with testable acceptance criteria.

**Design System Foundation:**
- **UX-DR1: Basecoat adoption.** Adopt Basecoat (pinned exact pre-1.0 patch version, e.g., `0.3.11`) as the canonical component vocabulary. Import its CSS into `fieldmark_shared/src/fieldmark.css`; document the pin in `fieldmark_shared/package.json` and the architecture doc alongside HTMX and AG Grid pins.
- **UX-DR2: Semantic color tokens.** Define `--color-success`, `--color-warning`, `--color-danger`, `--color-info`, `--color-neutral` as CSS custom properties in light *and* dark variants. Each token must meet 4.5:1 contrast against both light (`neutral-50/100`) and dark (`neutral-900/950`) surfaces. Verified at design time and via axe-core at build time.
- **UX-DR3: Status badge color vocabulary.** Bind entity-state → token mappings for Project, Inspection (with outcomes), Violation (with severity overlay), CorrectiveAction, and Severity badges per the Step 8 tables. Adding a state requires an ADR amendment; color follows deterministically.
- **UX-DR4: Compliance score thresholds.** Categorical color mapping: ≥ 90 Healthy / 70–89 Watch / 50–69 Concern / < 50 Critical, paired with the numeric value (color never sole carrier).
- **UX-DR5: Light/Dark theme toggle (MVP).** Server-readable `fm_theme` cookie (`system|light|dark`, `SameSite=Lax`, `Max-Age=1y`); `data-theme` attribute on `<html>` driven server-side; tiny 5-line inline script (the only inline JS in the app) resolves `system` before first paint; HTMX `hx-post` to `/preferences/theme` returns `204` with `HX-Trigger: theme-changed`. Cross-stack convention.
- **UX-DR6: Typography & tabular numerals.** Self-hosted Inter (variable woff2) for UI; JetBrains Mono for monospace; no Google Fonts requests. Default body `text-sm` (14px). `font-feature-settings: "tnum"` applied to all updating numbers (score, timestamps, counts).
- **UX-DR7: Iconography.** Adopt Lucide as the icon library. Inline SVG. Decorative icons carry `aria-hidden="true"`; functional icons have an accessible name. No icon-only buttons without accessible names.
- **UX-DR8: Tailwind defaults + spacing scale.** No custom breakpoints; standard scale (`4/8/16/24/32`); container `max-w-screen-2xl` with `px-6` gutters collapsing to `px-4` at tablet.

**Custom Components (Phase 1 foundation; Phase 2 layout):**
- **UX-DR9: StatusBadge component** — entity-state badge with deterministic semantic color, severity bump for Critical/High, color always paired with text.
- **UX-DR10: ActionButton component** — encapsulates the affordance trichotomy `absent | disabled-with-tooltip | present` server-side; emits Basecoat button with `hx-post`, `hx-target`, `hx-disabled-elt="this"`, or `aria-disabled` + tooltip on disabled.
- **UX-DR11: ComplianceTile component** — `role="status"`, `aria-live="polite"`, `aria-atomic="true"`, threshold-derived color paired with numeric value and threshold word, `tnum` to prevent OOB-swap jitter. Used at `#compliance-tile`.
- **UX-DR12: AuditRow component** — action StatusBadge · actor · relative timestamp (`tnum`) · optional disclosure of before/after JSON snapshot (JetBrains Mono). Lives in `aria-live="polite"` parent.
- **UX-DR13: InlineAlert component** — `role="alert"` (danger/warning) or `role="status"` (info/success); used heavily for 409 in-place rendering.
- **UX-DR14: FlashRegion component** — page-level `aria-live="polite"` `#flash-region`; transient ~5s announcements; reserved for system-state notices, **not** for domain errors.
- **UX-DR15: ThemeToggle component** — header-strip icon button cycling System → Light → Dark; icon reflects resolved theme; `aria-label` describes current + next; keyboard activatable.
- **UX-DR16: TabStrip component** — `<nav role="tablist">` with `<button role="tab" aria-selected aria-controls hx-get hx-target>`; arrow-key navigation (~15 lines JS); tab-content region carries `role="tabpanel"` + `aria-labelledby`; no Basecoat tab JS.
- **UX-DR17: DashboardTile component** — single-question summary tile; `role="status"` when updates arrive via OOB swap; specialization for ComplianceTile.
- **UX-DR18: EntityRail component** — `<aside tabindex="-1" role="region" aria-label>` sticky right rail at `≥ 1280px`, un-fixes and stacks beneath list at `768–1279px`; holds `#violation-detail` / `#inspection-detail` / `#corrective-action-detail` slots; empty-state copy when no entity selected.
- **UX-DR19: AGGridPanel wrapper** — server-side row model only; row-click handler fires `htmx.ajax(...)` to load detail into EntityRail; Quartz theme + FieldMark overrides; documented axe disables per AG Grid version.

**UX Patterns (10 named conventions — pattern violation is a defect, not a stylistic preference):**
- **UX-DR20: Three-Region Round-Trip pattern.** State-changing POSTs affecting an entity, compliance score, and audit log MUST return one response containing primary partial + OOB `#compliance-tile` + OOB `#audit-log` row.
- **UX-DR21: Affordance Trichotomy pattern.** Every action control is `absent | disabled-with-tooltip | present`. Greyed-without-tooltip is forbidden.
- **UX-DR22: Errors Render In Place (409 + originating partial).** Domain rule violations return HTTP 409 with the originating partial re-rendered showing current state and an inline alert. OOB tile and audit log MUST NOT update on 409.
- **UX-DR23: Audit Row As Receipt.** Every state change appends an audit row in the same DB transaction; the rendered row OOB-swaps into `#audit-log` in the same HTTP response. The audit row is the user's receipt — no separate success toast.
- **UX-DR24: Anchor Screen With HTMX Tabs.** Project Detail never navigates; tabs are HTMX swaps targeting `#project-detail-tab-content`. EntityRail is independent of tab state.
- **UX-DR25: List + Detail Co-Presence.** Right-rail detail at desktop, stacked detail at tablet — never page-level navigation to a detail screen.
- **UX-DR26: Empty State With Next Action.** Empty states render a "next action" affordance using ActionButton's trichotomy logic.
- **UX-DR27: Latency-Triggered Indication.** `hx-disabled-elt="this"` on every HTMX-firing button; opacity-only 100ms fade applied only if swap takes ≥ 100ms; no skeletons, no staggered reveals; `prefers-reduced-motion: reduce` drops to 0ms.
- **UX-DR28: Server-Decided Filtering.** All filtering, sorting, and pagination — including AG Grid — pass parameters to the server; clients never compute filter results.
- **UX-DR29: Stable Target IDs As Identity.** All HTMX targets are stable, semantic ids — never class-based, never positional. Cross-stack parity is verified by these ids.

**Responsive & Accessibility:**
- **UX-DR30: Responsive collapse rules.** Three-tier desktop-first (≥1280 / 768–1279 / <768): header strip overflow at 1024px; dashboard tile row 4 → 2×2 → 1; AG Grid responsive hide → horizontal scroll; EntityRail un-fixes at `<1280px`; tab strip horizontal scroll at mobile. Tailwind defaults only — no custom breakpoints. Verified at 1280/1024/768/375.
- **UX-DR31: Three named focus-management conventions.** (a) Primary partial swap → `tabindex="-1"` on swapped root + `HX-Trigger`-driven focus; (b) OOB swap → focus stays at trigger, region is `aria-live="polite"`; (c) Tab content swap → focus moves to swapped tab panel root.
- **UX-DR32: Live-region politeness contract.** `polite` for `#compliance-tile`, `#audit-log`, `#flash-region`, EntityRail; `assertive` (via `role="alert"`) for InlineAlert danger/warning; `aria-atomic="true"` on `#compliance-tile`, `false` on `#audit-log`.
- **UX-DR33: Skip-link & landmark structure.** "Skip to main content" as first focusable element; one `<header>`, one `<nav aria-label="Main">`, one `<main id="main-content">`, optional `<aside>`, optional `<footer>`. Strict single `<h1>` per page, no heading-level skipping.
- **UX-DR34: Form validation announcement (422 contract).** Re-render form partial with top InlineAlert (`role="alert"`) containing error count + link to first invalid field; each invalid field carries `aria-invalid="true"` + `aria-describedby` to its message.
- **UX-DR35: Focus indicator & touch targets.** 2px focus ring at 2px offset in body text color via `:focus-visible` override; touch targets ≥ 44×44px under `(pointer: coarse)`.
- **UX-DR36: AG Grid accessibility configuration.** `tests/axe-config.json` documents per-rule AG Grid disables with version + rationale; reviewed each AG Grid upgrade.
- **UX-DR37: Color-blindness simulation & 200% zoom.** Playwright runs deuteranopia/protanopia filter on canonical screens; 200% page zoom must not break layout (WCAG 1.4.4).

**Quality Gates / Verification harness:**
- **UX-DR38: Cross-stack visual regression suite.** Playwright snapshots of canonical screens (Compliance Dashboard, Project Detail, Violation Detail) × 3 stacks × 2 themes × 3 viewports. Pixel divergence beyond a tight threshold is a defect.
- **UX-DR39: axe-core in every E2E scenario.** `@axe-core/playwright` embedded in every cross-stack scenario; new AA-level violations are build-blocking.
- **UX-DR40: Canonical components example gallery.** Storybook-style static HTML examples for every custom component live in `fieldmark_shared/components/`; per-stack wrappers are tested for byte-identical output against these.

### FR Coverage Map

| FR | Epic | Note |
|---|---|---|
| FR1 | Epic 1 | Framework-native authentication |
| FR2 | Epic 1 | Authenticated user + role resolution per request |
| FR3 | Epic 1 | Logout |
| FR4 | Epic 1 | Redirect unauthenticated requests to login |
| FR5 | Epic 1 | Authorization decision primitive (`authz.Can`) |
| FR6 | Epic 1 | Server-decided action-button absence (affordance trichotomy) |
| FR7 | Epic 1 | HTTP-level rejection of unauthorized requests |
| FR8 | Epic 1 | Framework-native role assignment plumbing |
| FR9 | Epic 2 | Create Project |
| FR10 | Epic 2 | View Project list (filter/sort by status, score) |
| FR11 | Epic 2 | Project Detail anchor screen (tabs + EntityRail) |
| FR12 | Epic 2 | Active → OnHold transition with reason |
| FR13 | Epic 2 | OnHold → Active transition |
| FR14 | Epic 6 | Active → Closed (only when closure gates satisfied) |
| FR15 | Epic 6 | `can_close()` rendered as affordance trichotomy state |
| FR16 | Epic 3 | Schedule Inspection |
| FR17 | Epic 3 | Start Inspection (Scheduled → InProgress) |
| FR18 | Epic 3 | Complete Inspection with outcome + findings |
| FR19 | Epic 3 | Cancel Scheduled Inspection with reason |
| FR20 | Epic 3 | Auto-open Violation on Fail finding (same transaction) |
| FR21 | Epic 3 | Inspection list scoped + filterable |
| FR22 | Epic 4 | Violation records origin Finding, severity, due, status |
| FR22a | Epic 4 | Due date computed at open time, immutable |
| FR23 | Epic 4 | Assign / reassign Violation |
| FR24 | Epic 4 | Site Supervisor's assigned-violation queue |
| FR25 | Epic 4 | Overdue marking |
| FR26 | Epic 4 | Void Violation with reason |
| FR27 | Epic 4 | No reopening of Resolved/Voided |
| FR28 | Epic 5 | Submit Corrective Action |
| FR29 | Epic 5 | Take CA for review (UnderReview) |
| FR30 | Epic 5 | Approve CA — anchor three-region demo |
| FR31 | Epic 5 | Reject CA with notes |
| FR32 | Epic 5 | Reviewer ≠ submitter enforcement |
| FR33 | Epic 5 | Only latest non-Rejected CA may be approved |
| FR34 | Epic 3 | Server-side rule evaluation (introduced) |
| FR35 | Epics 3, 4, 5 | Recompute on every affecting transition (E3 inspect/violation-open, E4 void, E5 CA-approved/resolved) |
| FR36 | Epics 2, 5 | ComplianceTile OOB swap (intro E2; anchor canonical E5) |
| FR37 | Epic 6 | Closure-gate enforcement |
| FR38 | Epic 6 | Compliance rule parameters persisted + dynamic eval |
| FR39 | Epic 2 | AuditEntry on every domain mutation, same transaction |
| FR40 | Epic 2 | AuditEntry payload (actor, action, entity, before/after, metadata) |
| FR41 | Epic 2 | Append-only — no update/delete path |
| FR42 | Epic 2 | View Project audit log |
| FR43 | Epic 2 | Executive read-only audit access |
| FR44 | Epic 2 | Compliance Dashboard portfolio aggregates |
| FR45 | Epic 2 | HTMX partial refresh on tiles |
| FR46 | Epic 2 | Dashboard → Project Detail drill via HTMX swap |
| FR47 | Epic 2 | Filter/sort Project list (status, score, ownership) |
| FR48 | Epics 2, 3, 4 | AG Grid SSRM on ≥2 views (E2 projects; E3 inspections; E4 violations/audit) |
| FR49 | Epic 2 | AG Grid wire contract `{ "rows": [...], "lastRow": N }` |
| FR50 | Epic 2 | Row-select → HTMX detail-panel load |
| FR51 | Epic 2 | No business rules / client computation inside grid configs |
| FR52 | Epic 2 | Administrator reads reference data catalog |
| FR53 | Epic 2 | Reference data loaded without app restart |
| FR54 | Epic 2 | POST-only state changes; never GET |
| FR55 | Epic 2 | HTTP 409 + originating partial on rule violation |
| FR55a | Epic 1 | HTTP 422 + field-level errors on input validation failure (introduced at login form) |
| FR56 | Epic 2 | 403 (or equivalent) without state leakage |
| FR57 | Epic 2 | One-transaction discipline (mutation + audit + score recompute) |
| FR58 | Epics 1, 7 | Cross-stack route/HTMX-id/grid/audit-string/method-name parity (E1 initial; E7 final inventory) |
| FR59 | Epic 7 | Identical observable behavior across stacks on every scenario |
| FR60 | Epic 1 | Keyboard operability, tab order, visible focus |
| FR61 | Epic 1 | `aria-invalid` + `aria-describedby` on form errors |
| FR62 | Epic 1 | Focus management on HTMX state-changing swaps |
| FR63 | Epic 1 | `aria-live` on OOB targets |
| FR64 | Epic 1 | Visible + AT-announced disabled state during HTMX request |
| FR65 | Epic 7 | Playwright cross-stack E2E for every MVP workflow |
| FR66 | Epic 7 | Domain method unit-test coverage per stack |
| FR67–FR70 | *(Growth — out of MVP)* | Not in any MVP epic |

**UX-DR coverage:**
- Epic 1: UX-DR1, UX-DR2, UX-DR3, UX-DR4, UX-DR5, UX-DR6, UX-DR7, UX-DR8, UX-DR14, UX-DR15, UX-DR31, UX-DR32, UX-DR33, UX-DR34, UX-DR35
- Epic 2: UX-DR9, UX-DR10, UX-DR11, UX-DR12, UX-DR13, UX-DR16, UX-DR17, UX-DR18, UX-DR19, UX-DR20 (introduced on hold/resume), UX-DR21, UX-DR22, UX-DR23, UX-DR24, UX-DR25, UX-DR26, UX-DR27, UX-DR28, UX-DR29, UX-DR30
- Epic 5: UX-DR20 (canonical anchor demo — three-region orchestration verified end-to-end)
- Epic 7: UX-DR36, UX-DR37, UX-DR38, UX-DR39, UX-DR40

## Epic List

### Epic 1: Walking Skeleton — Auth, Design System & Parity Foundation

Lay the cross-stack foundation: confirm the three native CLI skeletons, author the infrastructure-owned `domain.*` DDL, stand up the parity tooling (`tools/parity/` + `make parity`), bootstrap the design system (Basecoat pinned, semantic color tokens, Phase 1 custom components), wire framework-native authentication (.NET Identity with `dotnet_auth`, Django `auth` with `django_auth`, Go stub middleware per ADR-012), seed identical-UUID dev users, and render an identical role-aware login + empty home page with the working light/dark theme toggle and FlashRegion across all three stacks. After this epic `make up && make run-{net,django,go}` produces three stacks rendering byte-identical chrome and `make parity` runs clean.
**FRs covered:** FR1, FR2, FR3, FR4, FR5, FR6, FR7, FR8, FR55a (introduced at login form), FR58 (initial), FR60, FR61, FR62, FR63, FR64
**UX-DRs:** UX-DR1, UX-DR2, UX-DR3, UX-DR4, UX-DR5, UX-DR6, UX-DR7, UX-DR8, UX-DR14, UX-DR15, UX-DR31, UX-DR32, UX-DR33, UX-DR34, UX-DR35

### Epic 2: Project Lifecycle & Compliance Dashboard

Admin/Project Manager creates Projects with required metadata; authorized users view the Compliance Dashboard (portfolio tiles + AG Grid project list with server-side row model); any authorized user drills into the Project Detail anchor screen (header strip with ComplianceTile, tabbed content area, EntityRail). Project Manager places Projects on hold and resumes them; transitions write AuditEntries in the same transaction and OOB-swap `#compliance-tile` + `#audit-log` — first lighting up the three-region orchestration pattern on a simple transition before it carries the anchor demo in Epic 5. Reference data (Trade Types, Violation Categories, Compliance Rules) readable by Administrators. AG Grid endpoint contract, 409 + originating partial, POST-only state changes, and one-transaction discipline are established here as cross-cutting patterns reused by every subsequent epic.
**FRs covered:** FR9, FR10, FR11, FR12, FR13, FR36 (introduced), FR39, FR40, FR41, FR42, FR43, FR44, FR45, FR46, FR47, FR48 (projects grid), FR49, FR50, FR51, FR52, FR53, FR54, FR55, FR56, FR57
**UX-DRs:** UX-DR9, UX-DR10, UX-DR11, UX-DR12, UX-DR13, UX-DR16, UX-DR17, UX-DR18, UX-DR19, UX-DR20 (introduced), UX-DR21, UX-DR22, UX-DR23, UX-DR24, UX-DR25, UX-DR26, UX-DR27, UX-DR28, UX-DR29, UX-DR30

### Epic 3: Inspection Workflow & Violation Genesis

Compliance Officer schedules Inspections (trade, inspector, scheduled time); Inspector starts them and completes them with outcome (Pass / Fail / Conditional), notes, and findings; Compliance Officer can cancel a Scheduled Inspection with reason. When an Inspection is completed with Fail-class findings, the system automatically opens a Violation for each finding atomically in the same transaction. Inspections tab on Project Detail renders the list (AG Grid) and loads Inspection detail into EntityRail on row select. Server-side compliance rule evaluation is introduced; score recomputes on Inspection completion and Violation opening in the same transaction.
**FRs covered:** FR16, FR17, FR18, FR19, FR20, FR21, FR34, FR35 (inspection-completed / violation-opened paths), FR48 (inspections grid)

### Epic 4: Violation Lifecycle & Assignment

Site Supervisor sees a queue of Violations assigned to them, filterable by status and overdue flag. Compliance Officer assigns and reassigns Violations to Site Supervisors (reassignment while InProgress emits an audit entry as a self-transition). System marks Violations as overdue server-side once due date passes and status is non-terminal. Administrator can void any non-terminal Violation with a recorded reason; voided Violations do not affect compliance scoring and score recomputes on void. Resolved or Voided Violations cannot be reopened — the user path is to create a new Violation from a new Finding.
**FRs covered:** FR22, FR22a, FR23, FR24, FR25, FR26, FR27, FR35 (void path), FR48 (violations grid)

### Epic 5: Corrective Action Workflow — The Anchor Demo

The thesis-proving epic. Site Supervisor submits a Corrective Action on an InProgress Violation (Violation Open → InProgress on first submission); Compliance Officer takes the CA for review (Submitted → UnderReview); approves it (CA → Approved, Violation → Resolved, compliance score recomputed) atomically in one transaction — and in one HTTP response the primary `#violation-detail` partial swap, OOB `#compliance-tile` update, and OOB `#audit-log` row append all land in a single paint, across all three stacks. Rejection-with-notes keeps the Violation in InProgress and re-renders the Submit affordance so Diego can resubmit; multiple CAs accumulate, only the latest non-Rejected is eligible for approval, and the submitter cannot review their own CA. This is where the canonical three-region orchestration pattern (UX-DR20) and audit-as-receipt pattern (UX-DR23) are verified end-to-end.
**FRs covered:** FR28, FR29, FR30, FR31, FR32, FR33, FR35 (CA-approved/resolved paths), FR36 (anchor canonical OOB)
**UX-DRs:** UX-DR20 (canonical demonstration)

### Epic 6: Project Closure & Compliance Gate Enforcement

Project Manager attempts to close an Active Project: the closure gate (`OpenViolationGate` + `RequiredInspectionPerTrade`) is evaluated server-side on every render of the Project Detail Summary; the Close button is `absent | disabled-with-tooltip | present` per the affordance trichotomy, with the disabled tooltip text coming from the server-side `ClosureGateError` message (which is part of the cross-stack parity contract). On click-when-blocked the server returns HTTP 409 with the originating partial re-rendered showing current state and an inline alert — no modal, no toast, no URL change. Once the gate passes (because the user scheduled the required inspection and resolved the violations through Epics 3–5), closure succeeds and writes a `ProjectClosed` audit entry. Compliance rule parameters (severity weights, due-offset values) are persisted in reference data and consulted dynamically by the rules engine — changing a parameter changes evaluation behavior without code changes.
**FRs covered:** FR14, FR15, FR37, FR38

### Epic 7: Cross-Stack Parity Demonstration & Demo-Ready Quality

Delivers value to the Talk Audience persona (Journey 5). Final Playwright cross-stack E2E coverage of every MVP user workflow × 3 stacks with `@axe-core/playwright` embedded in every scenario (FR65); domain-method unit-test coverage closed out per stack in each stack's idiomatic framework (FR66); visual regression suite captures canonical screens (Compliance Dashboard, Project Detail, Violation Detail) × 3 stacks × 2 themes × 3 viewports (1280/1024/768/375); color-blindness simulation (deuteranopia, protanopia) and 200% browser zoom verification on canonical screens; cross-stack latency-divergence verification (≤ 50 ms p95); `pg_indexes` zero-diff and route-inventory zero-diff verified as the final cross-stack inventory check (FR58 final); canonical component example gallery in `fieldmark_shared/components/` complete and tested against each stack's wrappers; documented "demo run" recipe in the repository root README.
**FRs covered:** FR58 (final), FR59, FR65, FR66
**UX-DRs:** UX-DR36, UX-DR37, UX-DR38, UX-DR39, UX-DR40

## Epic 1: Walking Skeleton — Auth, Design System & Parity Foundation

Lay the cross-stack foundation so every later epic implements only its own domain delta. After Epic 1, `make up && make run-{net,django,go}` produces three stacks rendering byte-identical chrome at a role-aware empty home page, with the light/dark theme toggle, FlashRegion, and the affordance-trichotomy primitive in place. `make parity` runs clean.

### Story 1.1: Confirm three native scaffolds, root Makefile, and Docker Compose harness

As a developer joining FieldMark for the first time,
I want a single documented set of commands that bring all three stacks and the database up locally,
So that I can run the application on every stack from a clean clone in minutes.

**Acceptance Criteria:**

**Given** a clean clone of the repository
**When** I run `make up` from the repo root
**Then** Postgres 17 starts via `docker compose up -d` and is reachable on `localhost:5432` with `fieldmark/fieldmark/fieldmark`
**And** the init scripts under `docker/postgres/init/` run automatically on first volume creation.

**Given** Postgres is up
**When** I run `make run-net`, `make run-django`, and `make run-go` (each in its own shell)
**Then** the three stacks bind to their native ports (.NET :5000, Django :8000, Fiber :3000)
**And** each stack reads `FIELDMARK_DATABASE_URL` (defaulting to the local Postgres URL) and connects without error.

**Given** the repo at HEAD
**When** I inspect the top-level `Makefile`
**Then** it exposes targets `up`, `down`, `reset`, `run-net`, `run-django`, `run-go`, `test-net`, `test-django`, `test-go`, `e2e`, `parity`, `css` per Architecture D20
**And** each target succeeds (or no-ops cleanly) on a fresh clone.

**Given** the repo at HEAD
**When** I inspect the three stack directories `FieldMark/`, `fieldmark_py/`, `fieldmark-go/`
**Then** each matches the Architecture §Initialization Commands layout (`.NET`: Web/Domain/Data class libs + xUnit projects; Django: `projects`, `inspections`, `violations`, `compliance`, `audit`, `reference`, `grid` apps with `uv` deps pinned; Go: `cmd/web` + `internal/{app,data,domain,web}`)
**And** each stack's README documents how to run it.

---

### Story 1.2: Verify Postgres init scripts produce the canonical `domain.*` schema on a fresh volume

As a developer working across three stacks,
I want a single command that destroys and re-creates the database in a known canonical state,
So that any drift between framework mapping code and the infrastructure-owned schema surfaces immediately.

**Acceptance Criteria:**

**Given** a running database with arbitrary local state
**When** I run `make reset` (`docker compose down -v && docker compose up -d`)
**Then** the volume is destroyed and recreated
**And** `001_schemas.sql`, `010_domain_tables.sql`, and `020_domain_seed.sql` execute in order with no errors visible in `docker logs`.

**Given** the database has been initialized
**When** I connect with `psql` and run `\dn`
**Then** the schemas `domain`, `dotnet_auth`, `django_auth`, `fiber_auth`, `infra` are all present.

**Given** the database has been initialized
**When** I run `SELECT table_name FROM information_schema.tables WHERE table_schema='domain' ORDER BY table_name`
**Then** exactly 12 tables are returned: `audit_entry`, `compliance_rule`, `corrective_action`, `finding`, `inspection`, `job_site`, `project`, `project_inspector`, `project_trade_scope`, `trade_type`, `violation`, `violation_category`.

**Given** the database has been initialized
**When** I inspect `domain.trade_type`, `domain.violation_category`, and `domain.compliance_rule`
**Then** the reference rows from `020_domain_seed.sql` are present and identical to the file's `INSERT` statements (verified by row count + a `SELECT` sample).

**Given** the canonical DDL is owned by infrastructure (ADR-014)
**When** I grep each stack for tooling that could mutate the `domain` schema (`dotnet ef migrations add` against a DbContext whose `HasDefaultSchema` is `"domain"`, Django `makemigrations` against a `domain.*` model with `Meta.managed = True`, Go migration tools targeting `domain.*`)
**Then** zero matches are found
**And** each stack's README explicitly states that `domain.*` is infrastructure-owned and that framework migrations only apply to its `*_auth` schema.

---

### Story 1.3: Establish `tools/parity/` and `make parity` with per-stack `--dump-routes`

As an agent or developer modifying any of the three stacks,
I want a single local command that detects cross-stack drift on routes and database indexes,
So that I catch divergence before it reaches code review — without depending on CI.

**Acceptance Criteria:**

**Given** the repo at HEAD
**When** I inspect `tools/parity/`
**Then** the directory contains executable scripts `dump-pg-indexes.sh`, `dump-routes-net.sh`, `dump-routes-django.sh`, `dump-routes-fiber.sh`, `diff-routes.sh`, `diff-pg-indexes.sh` (per Architecture D19).

**Given** each stack
**When** I invoke its route-dump subcommand
**Then** .NET responds to `dotnet run --project FieldMark/FieldMark.Web -- --dump-routes`, Django responds to `manage.py show_urls` (or equivalent custom command), and Go responds to `go run ./cmd/web -dump-routes`
**And** each command writes a normalized line-per-route list (METHOD + path) to stdout, sorted, with language casing normalized to lowercase.

**Given** the database has been initialized and all three stacks are buildable
**When** I run `make parity` from the repo root
**Then** the script invokes `diff-routes.sh` (comparing all three route dumps) and `diff-pg-indexes.sh` (snapshotting `pg_indexes WHERE schemaname='domain'` against the canonical file)
**And** both diffs exit `0` (clean).

**Given** I intentionally add a route to one stack and not the others
**When** I run `make parity`
**Then** the command exits non-zero and prints the diff identifying the divergent route.

**Given** `tools/git-hooks/pre-commit.sample` is committed
**When** I read it
**Then** it shows how to opt in to running `make parity` on commits touching any of `FieldMark/`, `fieldmark_py/`, `fieldmark-go/`, or `docker/postgres/init/`.

---

### Story 1.4: Bootstrap design system foundation in `fieldmark_shared/`

As a developer styling any FieldMark screen on any stack,
I want one compiled CSS bundle with the Basecoat component vocabulary, semantic tokens, status-badge vocabulary, typography, and vendored JS,
So that I can render byte-identical markup across the three stacks without authoring per-stack CSS.

**Acceptance Criteria:**

**Given** the repo at HEAD
**When** I inspect `fieldmark_shared/package.json`
**Then** `tailwindcss@4.x` is pinned to an exact patch and `basecoat-css` is pinned to an exact pre-1.0 patch (e.g., `0.3.11`) — no `^` or `~` ranges (UX-DR1)
**And** the version pins are documented in `_bmad-output/planning-artifacts/architecture.md` alongside HTMX and AG Grid.

**Given** `fieldmark_shared/src/fieldmark.css`
**When** I read it
**Then** it imports Basecoat's CSS, the AG Grid Quartz theme, and declares the five semantic color tokens `--color-success`, `--color-warning`, `--color-danger`, `--color-info`, `--color-neutral` (UX-DR2) with both light and dark variants
**And** each token meets ≥ 4.5:1 contrast against `neutral-50/100` and `neutral-900/950`, with a one-line comment recording the contrast ratio at design time.

**Given** the same file
**When** I read it
**Then** the status-badge color vocabulary (UX-DR3) for Project, Inspection, Violation (with severity overlay), CorrectiveAction, and Severity is encoded as deterministic class-to-token mappings
**And** the compliance-score threshold mapping (UX-DR4) is encoded as a single CSS rule keyed on `data-score-band` (`healthy`, `watch`, `concern`, `critical`).

**Given** `fieldmark_shared/src/`
**When** I read its CSS
**Then** Inter and JetBrains Mono are referenced as `@font-face` declarations pointing to self-hosted woff2 files under `fieldmark_shared/vendor/fonts/` (UX-DR6)
**And** body default is `text-sm` (14px), `font-feature-settings: "tnum"` is applied via a `.tnum` utility to compliance score, timestamps, counts, and any DOM element with numeric updating values.

**Given** the spacing scale (UX-DR8)
**When** I read `fieldmark_shared/src/_layout.css`
**Then** it uses only Tailwind defaults — no custom breakpoints — and documents `max-w-screen-2xl` container + `px-6 → px-4` gutter collapse with a single comment per rule naming the collapse point it implements.

**Given** the vendored JS strategy (Architecture D15)
**When** I inspect `fieldmark_shared/vendor/`
**Then** `htmx/htmx.min.js and ag-grid/35.2.1/ag-grid-community.min.js are committed; each stack's vendor/ static dir has directory symlinks pointing here
**And** each stack's static directory symlinks `dist/fieldmark.css` and the vendor directory.

**Given** the design system is built
**When** I run `cd fieldmark_shared && npm run build` (alias `make css`)
**Then** `fieldmark_shared/dist/fieldmark.css` is produced
**And** the compiled file is committed (no build step required after clone).

---

### Story 1.5: Implement cross-stack base layout with skip-link, landmarks, and FlashRegion

As a screen-reader user landing on any FieldMark page on any stack,
I want a consistent landmark structure, a working skip-link, and a polite live region for system announcements,
So that I can navigate the application predictably regardless of which stack served the page.

**Acceptance Criteria:**

**Given** each stack
**When** I open the rendered base layout (Razor `_Layout.cshtml`, Django `base.html`, Go `layouts/base.tmpl`)
**Then** the document body's first focusable element is a "Skip to main content" link that targets `#main-content` (UX-DR33)
**And** the link is visually hidden until focused.

**Given** the same base layout
**When** I inspect the document structure
**Then** there is exactly one `<header>`, one `<nav aria-label="Main">`, one `<main id="main-content">`, an optional `<aside>` slot for EntityRail, and an optional `<footer>`
**And** there are no nested landmarks of the same role.

**Given** every page rendered by any stack
**When** I count `<h1>` elements
**Then** exactly one is present (the page title), and heading levels never skip (no `<h3>` without a prior `<h2>` in the same section) (UX-DR33).

**Given** the base layout
**When** I inspect it
**Then** `#flash-region` is present as a `<div id="flash-region" role="status" aria-live="polite" aria-atomic="false">` in page chrome (UX-DR14, UX-DR32)
**And** it is empty by default and renders any messages from a per-stack `flash_messages()` template helper.

**Given** focus styling (UX-DR35)
**When** I tab through any rendered page
**Then** the `:focus-visible` ring is 2px wide at 2px offset, in body text color
**And** touch targets render at ≥ 44×44px under `(pointer: coarse)` media query.

**Given** the three stacks
**When** I capture the rendered HTML of `/` on each stack
**Then** the chrome (header skeleton, nav skeleton, skip-link, FlashRegion, main slot, footer skeleton) is byte-identical modulo any per-stack server-rendered values (none expected at this story).

---

### Story 1.6: Implement ThemeToggle with cookie persistence per stack

As any user on any stack,
I want a single header-strip control that cycles System → Light → Dark with my preference remembered across sessions,
So that the application matches my environment without flashing the wrong theme on first paint.

**Acceptance Criteria:**

**Given** I land on any page with no prior preference
**When** the page renders
**Then** the server emits `<html data-theme="system">` and a 5-line inline `<script>` resolves `prefers-color-scheme` and sets `data-theme="light"` or `data-theme="dark"` before first paint (UX-DR5)
**And** that inline script is the only inline JavaScript in the application; its presence is documented in the architecture doc.

**Given** the ThemeToggle component renders in the header strip beside the user avatar slot
**When** I inspect it (UX-DR15)
**Then** it is a 36×36 icon button with `aria-label="Theme: <current>; activate to cycle"`
**And** the displayed Lucide icon (Sun / Moon / Monitor) reflects the *currently resolved* theme.

**Given** I click the ThemeToggle
**When** the click fires
**Then** an HTMX `hx-post` is sent to `/preferences/theme` with the cycled value (`system` → `light` → `dark` → `system`)
**And** the server sets `Set-Cookie: fm_theme=<value>; Path=/; SameSite=Lax; Max-Age=31536000` and returns HTTP `204` with `HX-Trigger: theme-changed`
**And** a small client-side listener (≤ 20 LOC, vendored as `theme-toggle.js`) updates `data-theme` on `<html>` immediately.

**Given** I refresh the page after setting a preference
**When** the page renders
**Then** the server reads the `fm_theme` cookie and emits the correct `data-theme` attribute before first paint
**And** no flash of wrong theme is visible.

**Given** the three stacks
**When** I capture the rendered Theme Toggle markup on each
**Then** the HTML is byte-identical for identical inputs
**And** the `/preferences/theme` endpoint exists at the same path on all three (verified by `make parity`).

**Given** the user activates the toggle by keyboard (Space or Enter)
**When** I observe in a screen reader
**Then** the cycle works and the `aria-label` value updates to describe the new current + next state.

---

### Story 1.7: Wire ASP.NET Core Identity to `dotnet_auth` schema with conceptual roles

As an administrator using the .NET stack,
I want framework-native authentication backed by the `dotnet_auth` schema with the canonical password policy,
So that user identity is owned by .NET and never leaks into the `domain.*` schema.

**Acceptance Criteria:**

**Given** the .NET solution
**When** I inspect `FieldMark.Data` and `FieldMark.Web`
**Then** an `AuthDbContext` is configured with `modelBuilder.HasDefaultSchema("dotnet_auth")` and `UseSnakeCaseNamingConvention()` (Architecture D6)
**And** all seven Identity tables (`users`, `roles`, `user_roles`, `role_claims`, `user_claims`, `user_logins`, `user_tokens`) are mapped into `dotnet_auth`.

**Given** the password policy
**When** I read the Identity options registration
**Then** `RequireDigit = true`, `RequireLowercase = true`, `RequireUppercase = true`, `RequireNonAlphanumeric = false`, `RequiredLength = 10` are set.

**Given** Identity migrations
**When** I list `FieldMark.Data/Migrations/Auth/`
**Then** initial migration files exist that create the seven `dotnet_auth.*` tables only
**And** no migration touches `domain.*` (verified by grep).

**Given** Identity is wired
**When** the application starts for the first time after `make reset`
**Then** the five canonical role records are seeded into `dotnet_auth.roles` with names `ADMIN`, `COMPLIANCE_OFFICER`, `INSPECTOR`, `SITE_SUPERVISOR`, `EXECUTIVE`
**And** seeding is idempotent (running the seeder twice produces the same state).

**Given** parity tooling
**When** I run `make parity`
**Then** the route inventory diff remains clean (no .NET-only auth routes break parity — Django and Go have equivalent endpoints).

---

### Story 1.8: Wire Django built-in `auth` to `django_auth` schema with conceptual-role Groups

As an administrator using the Django stack,
I want framework-native authentication backed by the `django_auth` schema with role assignment via Groups,
So that Django's identity layer mirrors the .NET stack's isolation and never touches `domain.*`.

**Acceptance Criteria:**

**Given** `fieldmark_py/fieldmark/settings.py`
**When** I read database routing
**Then** a `DatabaseRouter` is configured (or `db_table` overrides applied to Django auth models) so that `auth_user`, `auth_group`, `auth_permission`, `auth_user_groups`, `auth_user_user_permissions`, `auth_group_permissions`, `django_session`, `django_admin_log` all resolve into the `django_auth` schema (Architecture D7).

**Given** Django migrations
**When** I run `uv run python manage.py migrate`
**Then** auth tables are created in `django_auth` and no auth migration touches `domain.*` (verified by inspecting `django_migrations` table).

**Given** the auth schema is migrated
**When** the application starts (or a one-shot management command runs)
**Then** five Django Groups are present: `ADMIN`, `COMPLIANCE_OFFICER`, `INSPECTOR`, `SITE_SUPERVISOR`, `EXECUTIVE`
**And** the seeding management command is idempotent.

**Given** `make parity`
**When** I run it after Django auth is wired
**Then** route inventory diff stays clean and `pg_indexes` for `domain.*` shows zero changes from the canonical inventory.

---

### Story 1.9: Implement Go/Fiber stub authentication middleware

As a developer running the Go stack at MVP,
I want a stub authentication mechanism that injects a configurable user identity into the request context,
So that the Go stack can render role-aware pages and exercise the cross-stack parity contract while real auth remains deferred per ADR-012.

**Acceptance Criteria:**

**Given** `fieldmark-go/internal/web/auth/`
**When** I inspect it
**Then** a `StubAuthMiddleware` exists that reads a user identifier from (in order) the `X-FieldMark-Actor` header, the `FIELDMARK_STUB_ACTOR` env var, or falls back to an "anonymous" sentinel.

**Given** the middleware resolves a user id
**When** the request context is hydrated
**Then** the user's UUID, username, and resolved conceptual role are bound to `c.Locals("user", ...)` and accessible from any handler
**And** the middleware looks up the user from a small `fiber_auth.users` + `fiber_auth.user_roles` pair of tables it owns (seeded in Story 1.10).

**Given** a request arrives with no identity
**When** the handler is `[`Authorize required`]`
**Then** the middleware returns HTTP `302` to `/login` (which renders a user-switcher stub list, not a real form).

**Given** ADR-012 explicitly defers real Go auth
**When** I read `fieldmark-go/CLAUDE.md`
**Then** the stub strategy is documented along with the explicit deferral and what landing real auth would look like (epic-sized work, not MVP).

**Given** `make parity`
**When** I run it
**Then** the Go stack's route inventory matches .NET and Django modulo language casing — including the `/login` and `/logout` paths.

---

### Story 1.10: Author shared UUID dev-user manifest and per-stack idempotent seed runners

As a developer running cross-stack scenarios,
I want every stack's dev users to share identical UUIDs,
So that audit comparison and cross-stack E2E parity tests can assert on actor identity without translation tables.

**Acceptance Criteria:**

**Given** `docker/postgres/init/seed-uuids/dev-users.json`
**When** I read it
**Then** it contains exactly six users: Marisol (`COMPLIANCE_OFFICER`), Diego (`SITE_SUPERVISOR`), Aisha (`ADMIN`), an inspector "Ravi" (`INSPECTOR`), Kenji (`EXECUTIVE`), and a no-role test user
**And** each entry has a canonical UUID (UUIDv7 preferred), a username, a display name, an initial password, and a role.

**Given** the .NET seeder `FieldMark.Web/SeedData/DevUsers.cs`
**When** `make run-net` starts the application with an empty database
**Then** the seeder reads the JSON manifest and writes the six users to `dotnet_auth.users` with the manifest's UUIDs as primary keys, hashed via ASP.NET Core Identity's `IPasswordHasher`
**And** running the seeder twice produces no duplicates and no errors (idempotent).

**Given** the Django seeder `fieldmark_py/<app>/management/commands/seed_dev_users.py`
**When** I run `uv run python manage.py seed_dev_users`
**Then** the six users are written to `django_auth.auth_user` with the manifest UUIDs (stored in a `uuid` column or as `username=<uuid>` if Django auth's PK contract is incompatible — chosen approach documented in the command's docstring)
**And** users are assigned to their conceptual-role Group.

**Given** the Go seeder `fieldmark-go/cmd/seed/main.go`
**When** I run `go run ./cmd/seed`
**Then** the six users are written to `fiber_auth.users` + `fiber_auth.user_roles` with the manifest UUIDs.

**Given** all three seeders have run
**When** I query each stack's auth tables for `id`/`uuid` of `marisol`
**Then** the returned UUID is identical across all three stacks (verified by SQL spot-check).

**Given** `020_domain_seed.sql` already seeds reference data
**When** I inspect the per-stack seed runners
**Then** none of them write into `domain.*` (reference data ownership stays with infrastructure SQL).

---

### Story 1.11: Login, logout, and unauthenticated-redirect across all three stacks

As any FieldMark user,
I want to log in with my username and password on .NET and Django and to pick my actor on Go,
So that the application identifies me on every request and rejects access to business routes until I authenticate.

**Acceptance Criteria:**

**Given** I am unauthenticated on any stack
**When** I request any business route (e.g., `/`, `/projects`, `/dashboard`)
**Then** I am redirected to `/login` (FR4)
**And** the response is HTTP `302` (or framework-equivalent).

**Given** the .NET login page
**When** it renders
**Then** the form is built from Basecoat input components with `<label>`-associated inputs
**And** on validation failure each invalid field renders `aria-invalid="true"` + `aria-describedby` linking to its error message, and the form partial is re-rendered with HTTP `422` containing a top InlineAlert with `role="alert"` and a link to the first invalid field (UX-DR34, FR61).

**Given** the Django login page
**When** it renders
**Then** the same form contract holds (Basecoat markup, label association, 422 + `aria-invalid`/`aria-describedby` on failure) — byte-identical markup verified by snapshot.

**Given** the Go login page
**When** it renders
**Then** it presents a list of seeded users from `fiber_auth.users` styled as Basecoat buttons; clicking a user sets the `X-FieldMark-Actor` cookie and redirects to `/`
**And** the page is clearly labeled as a development stub per ADR-012.

**Given** I am authenticated on any stack
**When** I click Log Out
**Then** the session is terminated (FR3) and I am redirected to `/login`
**And** subsequent requests to business routes redirect to `/login` again.

**Given** I am authenticated
**When** any request is handled
**Then** the handler can resolve my UUID and conceptual role(s) via the per-stack equivalent of `currentUser` (FR2)
**And** an unauthorized direct request (e.g., POSTing to `/projects/:id/close` without role) returns HTTP `403` without leaking the entity state (FR7, FR56).

**Given** `make parity`
**When** I run it with all auth wired
**Then** routes `/login`, `/logout`, `/preferences/theme` exist on all three stacks and the diff is clean.

---

### Story 1.12: Implement `authz.Can` primitive and ActionButton trichotomy helper per stack

As a developer rendering an action affordance on any FieldMark screen,
I want a single template helper that decides `absent | disabled-with-tooltip | present` per the affordance trichotomy,
So that future epics can introduce action buttons without re-deciding the rendering rule per screen.

**Acceptance Criteria:**

**Given** each stack
**When** I inspect its authorization module
**Then** it exposes a function with the signature `Can(user, action: string, entity?) -> bool` (.NET: `DomainPolicies.Can(...)`, Django: `fieldmark.authz.can(...)`, Go: `authz.Can(...)`) (FR5)
**And** the function consults the user's conceptual role(s) and any entity-scope rules (e.g., assignment, ownership) — initially trivial since no entities exist yet.

**Given** the ActionButton template helper (UX-DR10, UX-DR21)
**When** I inspect each stack's wrapper (`Pages/Shared/_ActionButton.cshtml`, `templates/components/_action_button.html`, `internal/web/templates/components/action_button.tmpl`)
**Then** it accepts `permission: bool`, `state_allows: bool`, `label: string`, `hx_post: string`, `hx_target: string`, and optional `disabled_reason: string`
**And** it implements the trichotomy:
- `permission=false` → renders nothing
- `permission=true && state_allows=false` → renders a Basecoat `<button disabled aria-disabled="true">` with a tooltip carrying `disabled_reason` and `aria-describedby` linking to the tooltip
- `permission=true && state_allows=true` → renders a Basecoat `<button hx-post=... hx-target=... hx-swap=... hx-disabled-elt="this">` (UX-DR27, FR64).

**Given** identical inputs on all three stacks
**When** I render the same ActionButton invocation
**Then** the produced HTML is byte-identical (verified by a unit test per stack snapshotting against a canonical example in `fieldmark_shared/components/action_button.example.html`).

**Given** the ActionButton renders a disabled button
**When** I navigate by keyboard
**Then** the disabled button retains its place in the tab order, the tooltip is keyboard-reachable, and the `aria-describedby` association is announced by screen readers (UX-DR21).

**Given** Epic 1 has no live use sites for ActionButton
**When** I grep each stack's templates for usages
**Then** zero rendered call sites are found (the primitive exists for Epic 2 onward) — but the unit-test snapshots prove the helper renders correctly.

---

### Story 1.13: Render empty role-aware Home page identically across all three stacks

As any authenticated user on any stack,
I want to land on a clean Home page that reflects who I am and offers the theme toggle,
So that I can confirm I am logged in on the right stack with the right identity before the product features land.

**Acceptance Criteria:**

**Given** I am authenticated on any stack
**When** I navigate to `/`
**Then** I see a page with:
- a `<header>` containing the FieldMark wordmark (left), the ThemeToggle (right of avatar), and an avatar showing my initials (Story 1.6)
- a single `<h1>FieldMark</h1>` (UX-DR33)
- a role badge using the StatusBadge color vocabulary showing my resolved conceptual role
- an empty content slot with a placeholder string ("Your projects will appear here.")
- the FlashRegion (`#flash-region`) in chrome (Story 1.5).

**Given** I render `/` on each of the three stacks while logged in as the same user (same UUID, courtesy of Story 1.10)
**When** I capture the rendered HTML
**Then** the chrome and the role badge are byte-identical across stacks (Basecoat-classed markup; no per-stack class names).

**Given** I am unauthenticated
**When** I navigate to `/`
**Then** I am redirected to `/login` (FR4 — already covered in Story 1.11; reasserted here for the empty Home).

**Given** the page renders
**When** I run an axe-core scan
**Then** zero WCAG 2.1 AA violations are reported (UX-DR39 — applies to every rendered page; locked in here as the first instance).

**Given** I tab through the page
**When** I observe focus order
**Then** Skip-Link → ThemeToggle → Avatar Menu → Logout → page body, in that order, with the visible focus ring at every step (UX-DR35).

**Given** `make parity`
**When** I run it after Story 1.13 lands
**Then** route inventory and `pg_indexes` for `domain.*` are clean across all three stacks
**And** Epic 1 is complete.

## Epic 2: Project Lifecycle & Compliance Dashboard

Establish every cross-cutting pattern (canonical request flow, audit-in-same-transaction, AG Grid SSRM, three-region OOB orchestration, EntityRail + TabStrip layout, Phase-2 components) on the Project aggregate. After this epic an Admin/PM can create projects, the Compliance Dashboard renders the portfolio, the Project Detail anchor screen is live, and place-on-hold/resume transitions exercise the three-region pattern.

### Story 2.1: Map `domain.project` and supporting tables into each stack's data layer

As a developer building Project-related features in any stack,
I want each stack's data layer to read and write `domain.project`, `domain.job_site`, `domain.project_trade_scope`, and `domain.project_inspector` against the existing canonical DDL,
So that subsequent stories can implement Project behavior without inventing schema.

**Acceptance Criteria:**

**Given** the .NET stack
**When** I inspect `FieldMark.Data/Configuration/`
**Then** `ProjectConfiguration.cs`, `JobSiteConfiguration.cs`, `ProjectTradeScopeConfiguration.cs`, `ProjectInspectorConfiguration.cs` use `ToTable("<table>", "domain")` and the snake_case naming convention
**And** enum columns use `HasConversion<string>()` for the `SCREAMING_SNAKE_CASE` storage convention (per `domain-model.md` §9).

**Given** the Django stack
**When** I inspect `projects/models.py`
**Then** `Project`, `JobSite`, `ProjectTradeScope`, `ProjectInspector` declare `Meta.managed = False` and `db_table = 'domain"."<table>'`
**And** field types match the canonical DDL exactly.

**Given** the Go stack
**When** I inspect `internal/data/`
**Then** `ProjectStore` is a narrow interface with read methods only at this story (no writes yet) backed by a `pgx`-using implementation in `internal/data/projectstore.go`
**And** column lists in SQL match the canonical DDL.

**Given** all three mappings exist
**When** I run `make parity`
**Then** `pg_indexes` for `domain.*` shows zero diff against the canonical inventory.

**Given** each stack's domain unit-test project
**When** I run `make test-net`, `make test-django`, `make test-go`
**Then** a smoke test exists per stack that loads a seeded Project by ID and asserts every column maps round-trip.

---

### Story 2.2: Map `domain.audit_entry` and provide a per-stack `append_audit_entry()` helper

As a handler author across all three stacks,
I want a single helper that appends an AuditEntry within the current DB transaction,
So that FR39 (audit-on-every-mutation) is mechanically satisfied for every transition Epic 2+ introduces.

**Acceptance Criteria:**

**Given** each stack's data layer
**When** I inspect it
**Then** `domain.audit_entry` is mapped: .NET via `AuditEntryConfiguration.cs`; Django via `audit/models.py` with `Meta.managed = False`; Go via `AuditEntryStore` interface + pgx implementation in `internal/data/auditentrystore.go`.

**Given** any mutating handler
**When** it calls `append_audit_entry(actor_id, action, entity_type, entity_id, project_id?, before_state, after_state, metadata?)` (or the per-stack idiomatic equivalent)
**Then** the helper writes the row using the same `DbContext` / connection / `pgx.Tx` as the surrounding transaction (FR39, FR57)
**And** `before_state` / `after_state` are serialized as JSONB
**And** `action` is stored verbatim from a canonical-string enum (no inventing variants — FR40).

**Given** a handler aborts (rule violation / exception)
**When** the transaction rolls back
**Then** no audit entry is left orphaned (verified by a per-stack integration test against Testcontainers / pytest-django / Go integration build tag).

**Given** the canonical action string list from CLAUDE.md
**When** I inspect each stack's enum/constants
**Then** the same 14 strings are present with the same SCREAMING/PascalCase casing
**And** the cross-stack diff for audit action constants is clean.

---

### Story 2.3: Map reference data tables and expose a read API per stack

As an Administrator,
I want to view the catalog of Trade Types, Violation Categories, and Compliance Rules,
So that I can verify the canonical reference data is loaded and visible (FR52, FR53).

**Acceptance Criteria:**

**Given** each stack
**When** I inspect its `reference` module (.NET `Domain/Entities/Reference/` + `Data/Configuration/Reference*.cs`; Django `reference/` app; Go `internal/domain/reference.go` + `internal/data/referencestore.go`)
**Then** `TradeType`, `ViolationCategory`, `ComplianceRule` map to their `domain.*` tables read-only
**And** rows are loaded on first request (not at process start) and cached in process memory with a TTL or invalidation hook
**And** no application restart is required to pick up changes (FR53 — verified by `UPDATE domain.compliance_rule …` then re-fetching).

**Given** I am authenticated as Administrator
**When** I navigate to `/admin/reference`
**Then** three sections render: Trade Types, Violation Categories, Compliance Rules
**And** each is a Basecoat Table (not AG Grid — UX-DR carve-out) showing all rows
**And** no Create/Edit/Delete affordances are rendered (FR67 is Growth).

**Given** I am authenticated as any non-Administrator role
**When** I navigate to `/admin/reference`
**Then** I receive HTTP 403 without entity-state leakage (FR56).

---

### Story 2.4: Implement Phase-2 markup-only components — StatusBadge, InlineAlert, AuditRow, DashboardTile

As a developer rendering Project Detail or the Compliance Dashboard,
I want four small Basecoat-compliant wrapper templates per stack with byte-identical output,
So that I can compose subsequent stories without inventing markup per screen.

**Acceptance Criteria:**

**Given** each stack's `partials/components/` directory
**When** I inspect it
**Then** wrappers for **StatusBadge** (UX-DR9), **InlineAlert** (UX-DR13), **AuditRow** (UX-DR12), and **DashboardTile** (UX-DR17) exist as markup-only templates with no logic.

**Given** the StatusBadge wrapper
**When** I render it with each entity-state combination from the Step 8 vocabulary (Project, Inspection, Violation+severity, CorrectiveAction)
**Then** the produced markup matches the canonical example in `fieldmark_shared/components/status_badge/` byte-for-byte across all three stacks
**And** color is never the sole information carrier (text label always present).

**Given** the InlineAlert wrapper
**When** I render it with `severity ∈ {danger, warning, info, success}`
**Then** danger/warning render `role="alert"` and info/success render `role="status"` (UX-DR13)
**And** the icon is Lucide-sourced and paired with text.

**Given** the AuditRow wrapper
**When** I render it with an action + actor + timestamp + before/after JSON
**Then** it lives inside an `aria-live="polite"` parent
**And** the disclosure for before/after JSON has `aria-expanded` and uses JetBrains Mono for the snippet (UX-DR12).

**Given** the DashboardTile wrapper
**When** I render it with a label + value + optional secondary text
**Then** it renders as a Basecoat Card with uppercase label + `text-3xl font-bold tnum` value (UX-DR17)
**And** when `role="status"` is set it announces value changes politely.

**Given** the canonical example gallery at `fieldmark_shared/components/`
**When** I run the per-stack snapshot tests
**Then** each stack's wrapper produces output byte-identical to the canonical example (UX-DR40 initial coverage).

---

### Story 2.5: Implement ComplianceTile component and `#compliance-tile` OOB target

As a user of any FieldMark screen showing project compliance,
I want a tile that renders the project's compliance score with threshold-based color and announces changes politely to assistive technology,
So that the canonical anchor demo's OOB-update mechanism has its target in place.

**Acceptance Criteria:**

**Given** each stack's component wrapper
**When** I inspect it (UX-DR11)
**Then** ComplianceTile renders as `<section id="compliance-tile" role="status" aria-live="polite" aria-atomic="true">` containing an uppercase label, a `text-3xl font-bold tnum` numeric score, the semantic threshold word ("Healthy"/"Watch"/"Concern"/"Critical"), and the semantic color (UX-DR4 thresholds).

**Given** scores 0–100
**When** the tile renders each
**Then** the band-to-color mapping is exactly: ≥90 success, 70–89 warning (lighter), 50–69 warning (darker), <50 danger
**And** color is paired with both the numeric value and the threshold word.

**Given** a Project Detail or Compliance Dashboard render
**When** the page is served
**Then** `#compliance-tile` appears once per page (header strip on Project Detail; first dashboard tile on Compliance Dashboard rendered with id `#compliance-tile-portfolio`)
**And** Story 2.12 / 2.13's OOB swap can target `#compliance-tile` without inventing markup.

**Given** the three stacks
**When** I render ComplianceTile with identical inputs
**Then** the produced HTML is byte-identical (verified by snapshot against `fieldmark_shared/components/compliance_tile/`).

---

### Story 2.6: Implement EntityRail component with responsive collapse

As a Compliance Officer or Project Manager working on the Project Detail screen at desktop,
I want a sticky right-rail container that holds the currently selected entity's detail and collapses to stacked layout below tablet,
So that List+Detail co-presence works at desktop and falls back gracefully on smaller viewports.

**Acceptance Criteria:**

**Given** each stack's wrapper (UX-DR18)
**When** I inspect it
**Then** EntityRail renders as `<aside id="<entity>-detail" tabindex="-1" role="region" aria-label="<entity-type> detail">` with three slots: header strip (entity-type label + dismiss ×), body, action footer.

**Given** no entity is selected
**When** the rail renders
**Then** the empty state reads "Select an entity to see its detail here." with Basecoat Card styling and an `aria-label="Empty entity rail"`.

**Given** the viewport
**When** width ≥ 1280px
**Then** the rail is sticky on the right at the third of the content area
**And** at 768–1279px the rail un-fixes and renders stacked below the list
**And** at <768px the same stacked behavior holds (UX-DR30).

**Given** a row is selected in any list within Project Detail
**When** the HTMX request lands a partial into the rail
**Then** focus moves to the rail's root via `tabindex="-1"` and either `HX-Trigger`-driven focus script or autofocus (UX-DR31 primary-swap convention).

**Given** I switch tabs on Project Detail
**When** the tab content swaps
**Then** the rail's content is not cleared by the tab swap (rail is independent of tab content per UX-DR24).

---

### Story 2.7: Implement TabStrip component with arrow-key navigation

As any user of Project Detail,
I want a horizontal tab strip whose tabs swap content via HTMX with full ARIA tablist semantics,
So that I can move between Summary / Inspections / Violations / Audit without page navigation and with keyboard support.

**Acceptance Criteria:**

**Given** each stack's wrapper (UX-DR16)
**When** I inspect it
**Then** TabStrip renders as `<nav role="tablist">` containing `<button role="tab" aria-selected="<bool>" aria-controls="<panel-id>" hx-get="<url>" hx-target="<panel-id>" hx-swap="innerHTML">` per tab
**And** the corresponding tab-panel container carries `role="tabpanel"` + `aria-labelledby="<tab-id>"` (UX-DR33 strict heading + landmark rules).

**Given** a tab is active
**When** I inspect its rendering
**Then** `aria-selected="true"` is set, the underline indicator is applied, and the tab is bold per Basecoat
**And** after a tab click the active tab updates `aria-selected` via OOB swap returned in the same response.

**Given** focus is on a tab
**When** I press Left/Right arrow keys
**Then** focus cycles between tabs without activating them; Enter or Space activates the focused tab
**And** the keyboard JS is ≤ 15 LOC and vendored as `tabstrip.js` (UX-DR15-style budget reflected in Step 11 components budget).

**Given** the three stacks
**When** I render the TabStrip with identical inputs
**Then** the HTML output is byte-identical.

---

### Story 2.8: Project create form (PM/Admin)

As a Project Manager or Administrator,
I want to create a new Project with required metadata,
So that the application has Projects to manage, inspect, and report on.

**Acceptance Criteria:**

**Given** I am authenticated as `ADMIN` or another role with `project.create` permission (initially `ADMIN`)
**When** I navigate to `/projects/new`
**Then** a form renders using Basecoat inputs with `<label>` associations capturing: code (unique), name, start date, optional target completion date, trade scope assignments (multi-select from `domain.trade_type`), inspector assignments (multi-select from seeded inspector users) (FR9).

**Given** the form
**When** I submit invalid input (missing required fields, duplicate code, end-date before start-date)
**Then** the form re-renders with HTTP `422` containing a top InlineAlert with `role="alert"` and `aria-invalid="true"` + `aria-describedby` on each invalid field (UX-DR34, FR61).

**Given** the form
**When** I submit valid input
**Then** a single transaction performs: load reference rows → call `Project.create(...)` entity method → write `Project` to `domain.project` → write `project_trade_scope` and `project_inspector` rows → append `AuditEntry(action="ProjectCreated", actor=me, before=null, after=<project state>)` → commit (FR57)
**And** I am redirected via `HX-Redirect` (or framework-equivalent) to `/projects/<id>`.

**Given** I am not authorized to create projects
**When** I navigate to `/projects/new`
**Then** the link is `absent` from the Compliance Dashboard (FR6) and a direct request returns HTTP `403` without state leakage (FR7).

**Given** I am authenticated as any authorized role
**When** the route `POST /projects/` is called with GET method
**Then** the response is HTTP `405` (or framework-equivalent rejecting GET on mutating routes) (FR54).

**Given** `make parity`
**When** I run it after this story
**Then** the route inventory contains `GET /projects/new` and `POST /projects/` on all three stacks with diff clean.

Note: `ProjectCreated` has been added to the canonical audit action string list (now 15 strings total) by ADR amendment ratified during epic planning. The story may emit it directly.

---

### Story 2.9: Project list AG Grid with server-side row model

As a Project Manager, Compliance Officer, or Administrator,
I want to view the Projects list with server-side filtering, sorting, and pagination,
So that I can navigate the portfolio without client-side compute (FR48–FR51).

**Acceptance Criteria:**

**Given** each stack
**When** I inspect its grid handler
**Then** `POST /grid/projects` accepts an AG Grid SSRM request payload (filterModel, sortModel, startRow, endRow) and returns JSON `{ "rows": [...], "lastRow": N }` with `snake_case` keys (FR49, Architecture D10).

**Given** the response rows
**When** I inspect each
**Then** they contain `id`, `code`, `name`, `status`, `compliance_score`, `start_date`, `target_completion_date`, `pm_name` — projected manually from `domain.project` join `domain.user`-equivalent (NFR6 — no AutoMapper).

**Given** the Project list page at `/projects`
**When** it renders
**Then** the AGGridPanel wrapper (UX-DR19) initializes AG Grid in SSRM mode pointing at `POST /grid/projects`
**And** filtering by status and sorting by `compliance_score` work end-to-end against the server (FR47, FR51 — no client-side filter).

**Given** I click a row
**When** the row-click handler fires
**Then** `htmx.ajax("GET", "/projects/<id>", { target: "#project-detail" })` runs and loads Project Detail (FR50)
**And** the grid does not render the detail itself (FR51).

**Given** the empty state (no projects)
**When** the grid renders
**Then** a custom no-rows overlay reads "No projects yet — create one to get started" with an ActionButton present for users with `project.create` permission, absent otherwise (UX-DR26).

**Given** `make parity`
**When** I run it
**Then** `POST /grid/projects` exists on all three stacks identically.

---

### Story 2.10: Compliance Dashboard with portfolio tiles

As any authorized user landing on FieldMark,
I want a Compliance Dashboard showing portfolio aggregates and a Projects list,
So that I can see "where is risk?" at a glance and drill into a Project from there (FR44, FR46).

**Acceptance Criteria:**

**Given** I am authenticated and have `dashboard.view` permission
**When** I navigate to `/dashboard` (and the empty Home from Story 1.13 redirects here once Story 2.10 lands)
**Then** the page renders four tiles at the top: ComplianceTile (`#compliance-tile-portfolio`) showing the portfolio average compliance score, a DashboardTile for "Overdue Violations" broken down by severity sub-badges, a DashboardTile for "Active Projects" count, and a DashboardTile for "Inspections This Week" count (FR44).

**Given** the tile row
**When** the viewport collapses
**Then** at 1024–1279px the row reflows to 2×2; at <768px to single column (UX-DR30).

**Given** the tile row
**When** any tile's underlying value changes from a same-page action
**Then** the affected tile is OOB-swappable (`id` stable; `role="status"` set on tiles that update) — wiring of actual change events is deferred to the epics that introduce them (UX-DR45 not violated because tiles render correctly at every static state).

**Given** the dashboard
**When** the tile row is rendered
**Then** below it the Projects AG Grid (Story 2.9) fills the remaining viewport.

**Given** any tile, including the ComplianceTile, renders
**When** there is no data (no projects, no inspections)
**Then** the tile renders `—` rather than `0` to distinguish "empty" from "zero" (UX-DR17).

**Given** `make parity`
**When** I run it
**Then** `GET /dashboard` is present on all three stacks.

---

### Story 2.11: Project Detail anchor screen with header strip, TabStrip, and EntityRail

As any authorized user,
I want to view a Project's current truth on a single page with HTMX-driven tabs and a co-present detail rail,
So that I can read the status, score, inspections, violations, and audit log without navigation (FR11).

**Acceptance Criteria:**

**Given** I navigate to `/projects/<id>`
**When** the page renders
**Then** the page has: a header strip with the project's code + name + StatusBadge for current status + ComplianceTile (`#compliance-tile`, FR36); a TabStrip with tabs Summary / Inspections / Violations / Audit; a content region `#project-detail-tab-content` (the tab panel target); a sticky `EntityRail` on the right at ≥ 1280px (UX-DR18, UX-DR24).

**Given** the Summary tab is active (default)
**When** the content region renders
**Then** it shows the project's code, name, start date, target completion date, PM name, list of assigned trades, list of assigned inspectors, and a "Place on Hold" / "Resume" / "Close" ActionButton row using the affordance trichotomy (UX-DR10) — the actual transitions land in Stories 2.12 and Epic 6.

**Given** I click another tab
**When** HTMX fires `hx-get` for that tab's content endpoint
**Then** only `#project-detail-tab-content` swaps; the header strip, ComplianceTile, EntityRail, and FlashRegion are unaffected (UX-DR24)
**And** focus moves to the new tab-panel root (UX-DR31 tab-content convention).

**Given** I am Executive role (FR43)
**When** I navigate to `/projects/<id>`
**Then** no action buttons render anywhere on the page (UX-DR21 affordance trichotomy collapses to `absent` for all permission-gated actions).

**Given** the page at <1280px
**When** it renders
**Then** EntityRail un-fixes and stacks below the tab content (UX-DR30).

**Given** `make parity`
**When** I run it
**Then** `/projects/:id` route + tab endpoints (`/projects/:id/tabs/summary`, `/projects/:id/tabs/inspections`, `/projects/:id/tabs/violations`, `/projects/:id/tabs/audit`) exist on all three stacks.

---

### Story 2.12: Place-On-Hold and Resume transitions with three-region OOB orchestration

As a Project Manager,
I want to place an Active Project on hold and resume it back to Active, each with a recorded reason and an audit row visible immediately,
So that the canonical three-region orchestration pattern (UX-DR20) is verified before the anchor demo lands in Epic 5 (FR12, FR13).

**Acceptance Criteria:**

**Given** I am authorized (`project.place_on_hold` or `project.resume`)
**When** I click the corresponding ActionButton on Project Detail
**Then** an inline form expands requesting a reason; submission fires `POST /projects/<id>/place-on-hold` (or `/resume`) with the reason in the body (FR54 — POST only).

**Given** a valid request
**When** the handler runs the canonical flow
**Then** it: authorizes → begins transaction → loads `Project` aggregate → calls `project.place_on_hold(reason)` (or `.resume()`) → appends `AuditEntry(action="ProjectPlacedOnHold" or "ProjectResumed", actor, before, after, metadata={reason})` → commits (FR57, FR40)
**And** the response body contains the re-rendered `#project-detail` partial **plus** OOB `#compliance-tile` (re-rendered with the unchanged score — score isn't affected, but the OOB pattern is exercised) **plus** OOB `#audit-log` row prepended (UX-DR20, UX-DR23).

**Given** the request is unauthorized
**When** the handler resolves the authz check
**Then** HTTP `403` is returned without entity-state leakage (FR7, FR56)
**And** no `#compliance-tile` or `#audit-log` OOB updates are emitted (UX-DR22).

**Given** the entity method raises (e.g., already on hold)
**When** the exception bubbles to the handler
**Then** HTTP `409` is returned with the originating `#project-detail` partial re-rendered showing *current* state + inline InlineAlert (UX-DR22, FR55)
**And** `#compliance-tile` and `#audit-log` are **not** updated.

**Given** the project transitions
**When** I open the Audit tab
**Then** the new audit row is present at the top
**And** the action string is exactly `ProjectPlacedOnHold` or `ProjectResumed` (cross-stack-identical — Architecture canonical list).

**Given** an E2E Playwright scenario runs against all three stacks
**When** the place-on-hold action is exercised
**Then** all three stacks produce a single HTTP round trip (no follow-up requests) with the three-region updates in one paint
**And** the local-dev p95 timing is ≤ 200 ms per stack with cross-stack divergence ≤ 50 ms p95 (NFR1).

---

### Story 2.13: Project audit log tab

As any authorized user on Project Detail,
I want to view the project's full audit history most-recent-first,
So that I have forensic visibility into every domain mutation that touched the project (FR42, FR43).

**Acceptance Criteria:**

**Given** I navigate to the Audit tab on `/projects/<id>`
**When** the tab content swaps in
**Then** `#audit-log` renders as a list (Basecoat Table, not AG Grid — per UX-DR carve-out) of AuditRow components for every `domain.audit_entry` row where `project_id = <id>`, ordered `created_at DESC`.

**Given** I am Executive role (FR43)
**When** I view the Audit tab
**Then** every row renders as a read-only AuditRow with no action affordances (UX-DR21 collapses all actions to `absent`).

**Given** a single AuditEntry
**When** I expand its disclosure
**Then** the `before_state` and `after_state` JSONB columns render in JetBrains Mono with `aria-expanded` toggling correctly (UX-DR12).

**Given** the page first renders
**When** the audit log is initially fetched
**Then** at most 100 rows load on initial render (older rows behind a "Load more" affordance that fetches the next 100 via HTMX `hx-get` appending into `#audit-log`).

**Given** `make parity`
**When** I run it
**Then** the audit-log shape (HTML structure of AuditRow + the Load-more affordance) is byte-identical across stacks.

---

### Story 2.14: Reference data read pages for Administrator

As an Administrator,
I want to view the catalog of reference data,
So that I can confirm what's loaded without running SQL (FR52).

**Acceptance Criteria:**

**Given** Story 2.3 mapped the reference data
**When** I am authenticated as `ADMIN` and navigate to `/admin/reference/trade-types`, `/admin/reference/violation-categories`, `/admin/reference/compliance-rules`
**Then** each page renders a Basecoat Table listing rows with their canonical columns (e.g., `code`, `name`, `severity_weight` for compliance rules) — no CRUD affordances (FR67 is Growth, intentionally absent).

**Given** any non-Administrator
**When** they navigate to any `/admin/reference/*` route
**Then** HTTP `403` is returned (FR56) — links to these pages are absent from any rendered nav (FR6).

**Given** `make parity`
**When** I run it
**Then** the three reference routes exist on all three stacks and the diff is clean.

## Epic 3: Inspection Workflow & Violation Genesis

Compliance Officer schedules Inspections; Inspector starts and completes them with findings; Fail findings auto-open Violations in the same DB transaction. Server-side compliance rule evaluation lit up here for the first time; score recomputes on inspection completion and violation opening.

### Story 3.1: Map `domain.inspection` and `domain.finding` into each stack's data layer

As a developer building Inspection features,
I want each stack's data layer to read and write `domain.inspection` and `domain.finding` against the existing canonical DDL,
So that subsequent stories can implement Inspection state transitions.

**Acceptance Criteria:**

**Given** the .NET stack
**When** I inspect `FieldMark.Data/Configuration/`
**Then** `InspectionConfiguration.cs` and `FindingConfiguration.cs` map the tables to `domain` schema with snake_case + enum-string conventions.

**Given** the Django stack
**When** I inspect `inspections/models.py`
**Then** `Inspection` and `Finding` models declare `Meta.managed = False` and target `domain.inspection` / `domain.finding` (FR-mapping only — DDL untouched).

**Given** the Go stack
**When** I inspect `internal/data/inspectionstore.go`
**Then** `InspectionStore` interface with `pgx`-backed implementation reads Inspections and their Findings (eager-loaded via SQL JOIN).

**Given** all mappings exist
**When** I run `make parity`
**Then** `pg_indexes` for `domain.*` is unchanged
**And** per-stack domain unit tests load a seeded Inspection with its Findings and assert round-trip fidelity.

---

### Story 3.2: Map `domain.violation` into each stack's data layer (write-capable for auto-open)

As a developer implementing Story 3.9 (auto-open Violations on Fail finding),
I want each stack's data layer to INSERT into `domain.violation` within the inspection-completion transaction,
So that Fail findings spawn Violations atomically per FR20.

**Acceptance Criteria:**

**Given** each stack
**When** I inspect its data layer
**Then** `ViolationStore` (Go) / `ViolationConfiguration.cs` (.NET) / `violations/models.py` Meta.managed=False (Django) is in place with read + insert capability
**And** the inserted row carries the canonical fields: `finding_id`, `severity`, `due_date`, `status='OPEN'`, `created_at`.

**Given** the Violation aggregate has no entity methods at this story
**When** I grep each stack's domain layer
**Then** only `Violation.open_from_finding(finding, severity, due_date_offset_days)` exists (the auto-open factory); assign/void/etc. land in Epic 4.

**Given** `make parity`
**When** I run it
**Then** `pg_indexes` is unchanged and no routes touch `/violations/*` yet (route parity stays clean).

---

### Story 3.3: Implement compliance rule engine and scoring helper

As a developer implementing any state transition that affects compliance,
I want a single per-stack pure function that recomputes a Project's compliance score from its current Violations and Inspections,
So that FR34 (server-side rule evaluation) and FR35 (same-transaction recompute) are mechanically satisfied by every handler that needs them.

**Acceptance Criteria:**

**Given** each stack
**When** I inspect its compliance module (.NET `Domain/Entities/Project.RecomputeComplianceScore`; Django `compliance/scoring.py`; Go `internal/domain/compliance_rules.go`)
**Then** a function `recompute_compliance_score(project, violations, completed_inspections, rules) -> int` exists that returns a value in `[0, 100]` (`CHECK` constraint already enforces this) (FR34).

**Given** the function
**When** it is called
**Then** it reads severity weights and due-offset values from the reference data (`domain.compliance_rule`), not from hardcoded constants (FR38 dynamic eval).

**Given** identical input fixtures (same Project + same Violations + same Inspections + same rule parameters)
**When** the function runs on each of the three stacks
**Then** the returned score is identical across stacks (cross-stack determinism — verified by a shared test fixture).

**Given** unit tests
**When** I run `make test-net`, `make test-django`, `make test-go`
**Then** each stack has unit tests covering: all-Pass Inspections + zero Violations → 100; one Open Critical Violation → < 50; one Resolved Violation does not lower score; Voided Violations do not affect score (FR26 invariant).

**Given** the function does not own transactional concerns
**When** it is called from a handler
**Then** the handler is responsible for persisting the new score within its surrounding transaction (FR57).

---

### Story 3.4: Schedule Inspection (Compliance Officer / Administrator)

As a Compliance Officer or Administrator,
I want to schedule an Inspection on a Project for a specific trade, inspector, and scheduled time,
So that Inspections exist for inspectors to perform (FR16).

**Acceptance Criteria:**

**Given** I am authorized (`inspection.schedule`)
**When** I click "Schedule Inspection" on the Inspections tab of Project Detail
**Then** an inline form expands (`#corrective-action-form`-style pattern — separate id `#inspection-schedule-form`) capturing trade (select from project's assigned trades), inspector (select from project's assigned inspectors), scheduled time (datetime).

**Given** a valid submission
**When** the handler runs `POST /projects/<id>/inspections/`
**Then** the canonical flow executes: authorize → begin txn → load Project → call `project.schedule_inspection(trade, inspector, when)` (entity factory) → write `domain.inspection` row with `status='SCHEDULED'` → append `AuditEntry(action="InspectionScheduled", ...)` → commit
**And** the response body re-renders `#inspection-list` plus OOB `#audit-log` row (no `#compliance-tile` update — scheduling doesn't affect score).

**Given** an unauthorized request
**When** the handler resolves authz
**Then** HTTP `403` without state leakage (FR7, FR56).

**Given** invalid input (trade not in project scope, inspector not assigned, scheduled-time in the past)
**When** the entity method raises
**Then** HTTP `409` with the originating form partial + InlineAlert (UX-DR22, FR55).

**Given** `make parity`
**When** I run it
**Then** `POST /projects/:id/inspections/` exists on all three stacks.

---

### Story 3.5: Inspection list AG Grid with SSRM endpoint

As any authorized user on the Inspections tab of Project Detail,
I want to see inspections with server-side filter/sort/pagination,
So that I can navigate inspection history at scale (FR48).

**Acceptance Criteria:**

**Given** the Inspections tab is active on `/projects/<id>`
**When** the tab content renders
**Then** AGGridPanel initializes against `POST /grid/inspections` with the project id passed as a request parameter
**And** the response is `{ "rows": [...], "lastRow": N }` with rows containing `id`, `trade_name`, `inspector_name`, `scheduled_at`, `status`, `outcome`, `completed_at` (FR49).

**Given** I click a row
**When** the row-click handler fires
**Then** `htmx.ajax("GET", "/inspections/<id>", { target: "#inspection-detail" })` loads inspection detail into EntityRail (FR50)
**And** the grid does not own the detail rendering (FR51).

**Given** `make parity`
**When** I run it
**Then** `POST /grid/inspections` exists identically on all three stacks.

---

### Story 3.6: Inspection list filters and date-range scoping

As an Inspector,
I want to filter Inspections by status and date range and see only inspections assigned to me by default,
So that my queue is meaningful (FR21).

**Acceptance Criteria:**

**Given** I am Inspector role
**When** I land on the Inspections tab
**Then** the default filter `inspector_id = me` is applied server-side
**And** the grid's filter UI exposes status (`SCHEDULED`/`IN_PROGRESS`/`COMPLETED`/`CANCELLED`), outcome (`PASS`/`FAIL`/`CONDITIONAL`/`NULL`), and a date range on `scheduled_at`.

**Given** I am Compliance Officer or Administrator
**When** I land on the Inspections tab
**Then** the default filter shows all inspections for the project
**And** the inspector filter is exposed as an additional dropdown.

**Given** the AG Grid filter model
**When** filters are applied
**Then** they are forwarded to the server in the SSRM payload and applied in SQL — never client-side (FR51, UX-DR28).

---

### Story 3.7: Inspection detail rendered in EntityRail

As any authorized user viewing the Inspections tab,
I want clicking a row to load the Inspection's details into the EntityRail,
So that I see the Inspection's findings and available actions without leaving the tab.

**Acceptance Criteria:**

**Given** I select a row in the Inspections AG Grid
**When** HTMX fires `GET /inspections/<id>`
**Then** the response is an HTML partial rooted at `<section id="inspection-detail" tabindex="-1" role="region" aria-label="Inspection detail">` that loads into the EntityRail
**And** focus moves to `#inspection-detail` (UX-DR31 primary-swap convention).

**Given** the rendered inspection detail
**When** I inspect it
**Then** it shows: trade name, inspector name, scheduled at, status StatusBadge, outcome StatusBadge (or `—` if not completed), notes, a Findings list (one card per finding with severity + description), and an ActionButton row using the trichotomy for "Start Inspection", "Complete Inspection", "Cancel Inspection" — actual transitions land in Stories 3.8 / 3.9 / 3.10.

**Given** at <1280px
**When** the rail collapses
**Then** the inspection detail renders stacked below the list per UX-DR30.

---

### Story 3.8: Start Inspection (Scheduled → InProgress)

As an Inspector,
I want to start an Inspection assigned to me,
So that I can record findings against it (FR17).

**Acceptance Criteria:**

**Given** I am Inspector and the Inspection is in `SCHEDULED` state and assigned to me
**When** I click "Start" on the Inspection detail
**Then** `POST /inspections/<id>/start` runs the canonical flow: authorize (`inspection.start` + assignment scope check) → begin txn → load Inspection → call `inspection.start(actor)` → status → `IN_PROGRESS`, `started_at = now()` → append `AuditEntry(action="InspectionStarted", ...)` → commit
**And** the response body re-renders `#inspection-detail` plus OOB `#audit-log` row (no score recompute yet — completion is what affects score).

**Given** the Inspection is not in `SCHEDULED` state
**When** I click "Start"
**Then** the ActionButton is `disabled-with-tooltip` (UX-DR21) — direct request via DevTools returns HTTP `409` with the originating partial + InlineAlert.

**Given** I am not the assigned inspector
**When** the Start button considers my permission
**Then** the button is `absent` (FR6) — direct request returns HTTP `403`.

---

### Story 3.9: Complete Inspection with findings and auto-open Violations atomically

As an Inspector,
I want to complete an InProgress Inspection with an outcome and zero or more findings,
So that Fail findings automatically open Violations and the project's compliance score updates in the same transaction (FR18, FR20, FR35).

**Acceptance Criteria:**

**Given** the Inspection is in `IN_PROGRESS` and I am the assigned inspector
**When** I open the "Complete" form on Inspection detail
**Then** I see outcome radio (`PASS` / `FAIL` / `CONDITIONAL`), notes textarea, and a repeating finding sub-form (severity select from `domain.violation_category.default_severity`, description, optional category).

**Given** a valid submission
**When** `POST /inspections/<id>/complete` runs
**Then** the canonical flow executes inside one transaction (FR57): authorize → load Inspection (with Findings) → call `inspection.complete(outcome, notes, findings)` → for each `FAIL`-class Finding: call `Violation.open_from_finding(finding, severity, due_date_offset_from_rule)` → INSERT each new `domain.violation` row → append `AuditEntry(action="InspectionCompleted", ...)` plus one `AuditEntry(action="ViolationOpened", entity=violation_id, ...)` per spawned Violation (FR40) → recompute Project compliance score via Story 3.3's function → UPDATE `domain.project.compliance_score` → commit.

**Given** the request succeeds
**When** the response is returned
**Then** the body contains `#inspection-detail` re-rendered (status `COMPLETED`, outcome badge, no Complete button) **plus** OOB `#compliance-tile` with the new score **plus** OOB `#audit-log` with the new rows appended (multiple rows possible if violations were spawned) **plus** OOB `#violation-list` re-rendered if the user is on the Violations tab — three-region orchestration pattern fully exercised (UX-DR20).

**Given** the entity method raises (e.g., Inspection not InProgress, outcome+findings inconsistent — `PASS` outcome with `FAIL` findings)
**When** the exception bubbles
**Then** HTTP `409` with the originating partial + InlineAlert; no DB state changed; no OOB updates emitted (UX-DR22).

**Given** the transaction aborts partway (e.g., DB connection loss after inserting some Violations)
**When** the handler resolves
**Then** the entire transaction rolls back — no orphan Violations, no orphan AuditEntries, no partial score update (FR57, NFR5).

**Given** an E2E Playwright scenario runs against all three stacks
**When** complete-with-fail-findings is exercised
**Then** the three-region paint happens in a single HTTP response on all three stacks with cross-stack timing divergence ≤ 50 ms p95 (NFR1).

---

### Story 3.10: Cancel Scheduled Inspection with reason

As a Compliance Officer or Administrator,
I want to cancel a Scheduled Inspection with a recorded reason,
So that abandoned schedules don't clutter the queue (FR19).

**Acceptance Criteria:**

**Given** the Inspection is in `SCHEDULED` and I am authorized (`inspection.cancel`)
**When** I click "Cancel" on Inspection detail
**Then** an inline form expands capturing a reason; `POST /inspections/<id>/cancel` runs the canonical flow → status → `CANCELLED`, `cancelled_at`, `cancellation_reason` set → append `AuditEntry(action="InspectionCancelled", metadata={reason})` → commit
**And** the response re-renders `#inspection-detail` + OOB `#audit-log` row (no score impact).

**Given** the Inspection is not `SCHEDULED` (already `IN_PROGRESS` / `COMPLETED` / `CANCELLED`)
**When** the entity method raises
**Then** HTTP `409` with originating partial + InlineAlert (UX-DR22)
**And** the Cancel button is `disabled-with-tooltip` whenever the precondition fails (UX-DR21).

## Epic 4: Violation Lifecycle & Assignment

Site Supervisors see their queue; Compliance Officers assign and reassign; overdue marking lights up; Administrators can void; reopening of terminal states is blocked.

### Story 4.1: Extend Violation data layer with write operations (assign, void, reassign)

As a developer implementing Violation workflow,
I want each stack's data layer to support UPDATE on `domain.violation` for assignment, reassignment, and void,
So that Epic 4's entity methods have persistence.

**Acceptance Criteria:**

**Given** each stack
**When** I inspect its Violation data path
**Then** it supports updating `assigned_to`, `status`, `voided_at`, `void_reason`
**And** `Violation` aggregate methods `assign(supervisor_id)`, `void(reason, actor)`, `reject_resubmission()` exist on the entity (entity methods own the rules; data layer is dumb).

**Given** the canonical DDL CHECK constraints
**When** I attempt to violate them (e.g., voided without reason)
**Then** the DB rejects the write (defense-in-depth — NFR5) and the entity method's pre-check raises before reaching the DB anyway.

---

### Story 4.2: Violation list AG Grid with SSRM endpoint

As any authorized user on the Violations tab of Project Detail,
I want a server-side row-model grid of the project's Violations,
So that I can browse Violations with filter and sort (FR48).

**Acceptance Criteria:**

**Given** the Violations tab is active
**When** the tab content renders
**Then** AGGridPanel initializes against `POST /grid/violations` with `project_id` as a request parameter
**And** rows contain `id`, `category_name`, `severity`, `status`, `assigned_to_name`, `due_date`, `is_overdue` (bool — see Story 4.6), `created_at`.

**Given** I click a row
**When** the row-click handler fires
**Then** `GET /violations/<id>` loads the Violation detail into `#violation-detail` in the EntityRail (FR50)
**And** the grid does not render the detail itself.

**Given** `make parity`
**When** I run it
**Then** `POST /grid/violations` exists on all three stacks identically.

---

### Story 4.3: Violation detail rendered in EntityRail

As any authorized user,
I want to see a selected Violation's detail in the EntityRail,
So that I can view severity, status, assignment, and audit context co-present with the list.

**Acceptance Criteria:**

**Given** I select a row in the Violations grid
**When** HTMX fires `GET /violations/<id>`
**Then** the response is a partial rooted at `<section id="violation-detail" tabindex="-1" role="region" aria-label="Violation detail">` loaded into the EntityRail
**And** focus moves to `#violation-detail` (UX-DR31).

**Given** the rendered Violation detail
**When** I inspect it
**Then** it shows: Severity StatusBadge (with Critical/High bump per UX-DR3), Status StatusBadge, the origin Finding excerpt, due date with `is_overdue` styling if applicable, assignee name, action button row using the trichotomy (`Assign`, `Void`, plus CA-related buttons for Epic 5).

**Given** I am Executive role (FR43)
**When** I view the rail
**Then** all action buttons are `absent` (UX-DR21).

---

### Story 4.4: Assign and reassign Violation to Site Supervisor

As a Compliance Officer,
I want to assign or reassign a Violation to a Site Supervisor,
So that a remediator owns the Violation and the audit trail records each reassignment (FR23).

**Acceptance Criteria:**

**Given** I am authorized (`violation.assign`)
**When** I click "Assign" or "Reassign" on the Violation detail
**Then** an inline form expands with a select of users in the `SITE_SUPERVISOR` role and an optional note; `POST /violations/<id>/assign` runs the canonical flow → call `violation.assign(supervisor_id)` → if the Violation was `OPEN` and is now in scope of a remediator, status stays `OPEN` (assignment doesn't auto-transition); reassignment while in `IN_PROGRESS` is a self-transition that emits an audit entry → append `AuditEntry(action="ViolationAssigned", before={assigned_to: old}, after={assigned_to: new}, metadata={note?})` → commit (FR23).

**Given** the response
**When** it returns
**Then** `#violation-detail` is re-rendered with the new assignee plus OOB `#audit-log` row (no `#compliance-tile` update — assignment doesn't affect score).

**Given** the Violation is in a terminal state (`RESOLVED` / `VOIDED`)
**When** the entity method is called
**Then** it raises and the handler returns HTTP `409` with originating partial + InlineAlert.

---

### Story 4.5: Site Supervisor's assigned-violation queue

As a Site Supervisor,
I want to see Violations assigned to me filterable by status and overdue flag,
So that I know what I need to work on (FR24).

**Acceptance Criteria:**

**Given** I am Site Supervisor and navigate to `/violations/mine`
**When** the page renders
**Then** an AGGridPanel renders a grid filtered server-side by `assigned_to = me`
**And** filter UI exposes status and "Overdue only" toggle (consults Story 4.6's overdue marking).

**Given** the empty queue
**When** I have no assigned violations
**Then** the no-rows overlay reads "No violations assigned to you. Good work." (UX-DR26).

**Given** I am any non-supervisor role
**When** I navigate to `/violations/mine`
**Then** I am redirected to `/dashboard` (the page is gated by role — FR6 collapses the link to absent in nav for non-supervisors).

---

### Story 4.6: Server-rendered overdue marking

As a user viewing any Violation list or detail,
I want overdue Violations visually marked,
So that urgency is communicated without color being the sole carrier (FR25).

**Acceptance Criteria:**

**Given** a Violation with `due_date < now()` and `status NOT IN ('RESOLVED', 'VOIDED')`
**When** any view that includes it renders
**Then** the server computes `is_overdue = true` and emits an "Overdue" text badge alongside the status StatusBadge with `--color-danger` styling — color paired with text (NFR3, WCAG 1.4.1).

**Given** a Violation not yet due, or in terminal state
**When** the same view renders
**Then** no Overdue badge is rendered and `is_overdue = false` in the AG Grid row payload.

**Given** the overdue computation
**When** I grep client code for it
**Then** zero client-side overdue calculation exists — `is_overdue` is always server-computed (FR34, UX-DR28).

**Given** the system time advances past a Violation's `due_date`
**When** the next render occurs
**Then** the Violation begins displaying as overdue without any background job or app restart (purely render-time computation).

---

### Story 4.7: Void Violation (Administrator) with score recompute

As an Administrator,
I want to void a non-terminal Violation with a recorded reason,
So that mistakenly opened or no-longer-applicable Violations can be removed from compliance scoring (FR26, FR35).

**Acceptance Criteria:**

**Given** the Violation is in a non-terminal state and I am authorized (`violation.void`)
**When** I click "Void" on the Violation detail
**Then** an inline form expands capturing a reason; `POST /violations/<id>/void` runs the canonical flow → call `violation.void(reason, actor)` → status → `VOIDED`, `void_reason`, `voided_at` set → append `AuditEntry(action="ViolationVoided", metadata={reason})` → recompute Project compliance score via Story 3.3 → UPDATE `domain.project.compliance_score` → commit (FR57).

**Given** the response
**When** it returns
**Then** `#violation-detail` is re-rendered with terminal state plus OOB `#compliance-tile` with the new score plus OOB `#audit-log` row (UX-DR20 three-region pattern).

**Given** the Violation is already in a terminal state
**When** the entity method is called
**Then** it raises and the handler returns HTTP `409` (UX-DR22) — Void button is `disabled-with-tooltip` whenever the precondition fails (UX-DR21).

**Given** the Voided Violation
**When** the compliance score is recomputed
**Then** it is excluded from scoring inputs (FR26 invariant — verified by Story 3.3's unit tests).

---

### Story 4.8: Block reopening of Resolved or Voided Violations

As a user who would otherwise try to re-open a terminal Violation,
I want the system to prevent it,
So that audit history remains a faithful append-only ledger and users follow the "new finding → new violation" path (FR27).

**Acceptance Criteria:**

**Given** a Violation in `RESOLVED` or `VOIDED` state
**When** I view it on Violation detail
**Then** no transition affordance returns the Violation to `OPEN` or `IN_PROGRESS` — all such ActionButtons are `absent` (UX-DR21, FR6).

**Given** a direct HTTP-level attempt to call any reopening endpoint (which does not exist)
**When** the request is made
**Then** the server returns HTTP `404` (no such route) (FR7).

**Given** I want to revisit the underlying issue
**When** I navigate from the Resolved/Voided Violation detail
**Then** a contextual note in the Violation detail says "To revisit this issue, complete an Inspection with a new Fail finding."

## Epic 5: Corrective Action Workflow — The Anchor Demo

The thesis-proving epic. The canonical Marisol Approve flow lands as three-region OOB orchestration in a single HTTP response across all three stacks.

### Story 5.1: Map `domain.corrective_action` writes into each stack's data layer

As a developer implementing Corrective Action workflow,
I want write-capable CA mapping per stack against the existing canonical DDL,
So that submission, take-for-review, approval, and rejection persist correctly.

**Acceptance Criteria:**

**Given** each stack
**When** I inspect its data path
**Then** CRUD on `domain.corrective_action` is mapped (.NET `CorrectiveActionConfiguration.cs`; Django `violations/models.py` with `Meta.managed=False`; Go `CorrectiveActionStore` + pgx)
**And** CA is treated as part of the Violation aggregate (Violation methods own CA transitions per the canonical method-name list — `submit_corrective_action`, `approve_resolution`, `reject_resolution`).

**Given** the canonical DDL
**When** I inspect `domain.corrective_action`
**Then** the foreign key to `domain.violation` exists, status enum is in place, `submitted_at`, `reviewed_at`, `reviewer_id`, `review_notes` columns exist
**And** the Violation aggregate enforces invariants in code; DB CHECK constraints are defense-in-depth.

---

### Story 5.2: Submit Corrective Action (Site Supervisor)

As a Site Supervisor,
I want to submit a Corrective Action for a Violation assigned to me,
So that my remediation work enters the review pipeline (FR28).

**Acceptance Criteria:**

**Given** I am assigned to a Violation in `OPEN` or `IN_PROGRESS` state
**When** I click "Submit Corrective Action" on the Violation detail
**Then** an inline form expands at `#corrective-action-form` capturing description (textarea, required) and optional evidence_ref (string placeholder per NFR11 — no file upload).

**Given** a valid submission
**When** `POST /violations/<id>/corrective-actions/` runs
**Then** the canonical flow executes: authorize (`corrective_action.submit` + assignment scope) → begin txn → load Violation aggregate → call `violation.submit_corrective_action(description, evidence_ref, actor)` → INSERT `domain.corrective_action` with status `SUBMITTED` → if Violation was `OPEN`, transition to `IN_PROGRESS` (entity rule) → append `AuditEntry(action="CorrectiveActionSubmitted")` → commit (FR57).

**Given** the response
**When** it returns
**Then** `#violation-detail` is re-rendered (status `IN_PROGRESS`, Submit button gone, CA card appears at `#corrective-action-list`) plus OOB `#audit-log` row — no `#compliance-tile` update (submission doesn't change score).

**Given** the same submitter immediately submits a second CA while one is non-terminal
**When** the entity rule fires
**Then** the handler returns HTTP `409` with InlineAlert "A corrective action is already in flight" (UX-DR22) — Submit button is `disabled-with-tooltip` whenever a non-terminal CA already exists for this Violation.

---

### Story 5.3: Render Corrective Action list within Violation detail

As any user viewing a Violation,
I want to see all Corrective Actions submitted against it with the latest non-Rejected highlighted,
So that I understand the remediation history (supporting FR33 — only latest non-Rejected may be approved).

**Acceptance Criteria:**

**Given** a Violation with CAs
**When** I view Violation detail
**Then** `#corrective-action-list` renders one card per CA ordered submission-DESC, each showing: status StatusBadge, submitter name, submitted_at, description (truncated with disclosure), evidence_ref if present, and (if reviewed) reviewer name + reviewed_at + review_notes
**And** the latest non-Rejected CA is visually highlighted (e.g., a left-border accent in the semantic color of its status).

**Given** there are no CAs yet
**When** the list renders
**Then** an empty state reads "No corrective actions submitted yet" with a Submit ActionButton for authorized assignees (UX-DR26).

---

### Story 5.4: Take Corrective Action for review (Submitted → UnderReview)

As a Compliance Officer,
I want to take a Submitted CA for review,
So that it enters my review queue and signals to the submitter that work has started (FR29).

**Acceptance Criteria:**

**Given** I am authorized (`corrective_action.take_for_review`) and the CA is `SUBMITTED` (Stating FR32 — and I am NOT the submitter)
**When** I click "Take for Review" on the CA card
**Then** `POST /violations/<id>/corrective-actions/<ca_id>/take` runs the canonical flow → call `violation.take_corrective_action_for_review(ca_id, actor)` → CA status `UNDER_REVIEW`, `reviewer_id = actor`, `reviewed_at = now()` (timestamp tentative — finalized on approve/reject) → append `AuditEntry(action="CorrectiveActionTakenForReview", metadata={reviewer_id})` → commit.

**Given** the response
**When** it returns
**Then** `#violation-detail` re-renders showing the CA in `UNDER_REVIEW` plus OOB `#audit-log` row.

**Given** I am the submitter
**When** "Take for Review" is rendered
**Then** the button is `absent` (FR32 — reviewer ≠ submitter; permission collapses to `false`).

---

### Story 5.5: Approve Corrective Action — canonical anchor three-region demo

As Marisol the Compliance Officer,
I want to approve a Corrective Action and see the Violation resolve, the compliance tile update, and a new audit row appear in one paint,
So that I have my receipt without refresh, modal, or follow-up request (FR30, FR32, FR33, FR35, FR36, UX-DR20, UX-DR23).

**Acceptance Criteria:**

**Given** the CA is `UNDER_REVIEW`, I am the reviewer (FR32 — must NOT be the submitter), and the CA is the latest non-Rejected CA of the Violation (FR33)
**When** I click "Approve" on the CA card
**Then** `POST /violations/<id>/corrective-actions/<ca_id>/approve` runs the canonical flow inside one transaction (FR57): authorize → begin txn → load Violation aggregate (with CAs) → call `violation.approve_resolution(ca_id, reviewer)` → CA status `APPROVED`, Violation status `RESOLVED`, `resolved_at = now()` → append `AuditEntry(action="CorrectiveActionApproved", actor, before={violation: IN_PROGRESS, ca: UNDER_REVIEW}, after={violation: RESOLVED, ca: APPROVED})` → recompute Project compliance score via Story 3.3 → UPDATE `domain.project.compliance_score` → commit (FR35, FR57).

**Given** the response
**When** it returns
**Then** the response body contains:
- the primary `#violation-detail` partial re-rendered (Resolved badge, no Approve/Reject buttons, CA card showing Approved by Marisol at the timestamp), **plus**
- OOB swap for `#compliance-tile` with the new score (UX-DR11, FR36), **plus**
- OOB swap appending a new audit row to `#audit-log` (UX-DR12, UX-DR23) — all in a single HTTP response (UX-DR20).

**Given** I am the submitter of the CA
**When** the Approve button is evaluated
**Then** the button is `absent` (FR32) and a direct request returns HTTP `403`.

**Given** the CA is not the latest non-Rejected for this Violation (a newer CA exists)
**When** the entity method is called
**Then** it raises and the handler returns HTTP `409` with originating partial + InlineAlert "A more recent corrective action exists" (FR33, UX-DR22) — Approve button is `disabled-with-tooltip` whenever this precondition fails.

**Given** a Playwright cross-stack E2E scenario for the anchor demo
**When** it runs against .NET, Django, and Go
**Then** each stack produces exactly one HTTP round trip (no follow-up requests), zero full-page reloads (`page.on('load')` not fired), and the three OOB regions all update in the same paint
**And** local-dev p95 timing is ≤ 200 ms per stack with cross-stack divergence ≤ 50 ms p95 (NFR1)
**And** focus moves to `#violation-detail` after the swap (UX-DR31)
**And** `aria-live` regions announce the score change and new audit row (UX-DR32).

---

### Story 5.6: Reject Corrective Action with notes (UnderReview → Rejected)

As a Compliance Officer reviewing a CA,
I want to reject the CA with review notes,
So that the Site Supervisor can resubmit and the Violation remains in remediation (FR31).

**Acceptance Criteria:**

**Given** the CA is `UNDER_REVIEW` and I am the reviewer (not the submitter)
**When** I click "Reject" on the CA card
**Then** an inline form expands at `#corrective-action-form` capturing review_notes (textarea, required); `POST /violations/<id>/corrective-actions/<ca_id>/reject` runs the canonical flow → call `violation.reject_resolution(ca_id, reviewer, notes)` → CA status `REJECTED`, `reviewed_at = now()`, `review_notes` set → Violation remains in `IN_PROGRESS` (FR31) → append `AuditEntry(action="CorrectiveActionRejected", metadata={notes_redacted_to_length})` → commit.

**Given** the response
**When** it returns
**Then** `#violation-detail` re-renders showing CA Rejected with the review notes visible AND the "Submit Corrective Action" button BACK (FR31 — Violation stays InProgress and new submission is allowed) plus OOB `#audit-log` row — no `#compliance-tile` update (no resolution).

**Given** the Site Supervisor returns to the Violation
**When** they view it
**Then** they see the rejected CA's review notes inline and the Submit button rendered, enabling resubmission per FR31's cycle.

**Given** multiple CAs accumulate
**When** the Violation is viewed
**Then** the CA list shows all of them, ordered DESC, with their statuses; FR33's invariant (only latest non-Rejected may be approved) holds across the cycle.

---

### Story 5.7: Cross-stack anchor-demo E2E and timing parity

As the talk-audience persona (Journey 5),
I want a Playwright scenario that runs the anchor demo end-to-end on all three stacks with timing assertions,
So that the thesis is mechanically verifiable in CI / locally before the talk.

**Acceptance Criteria:**

**Given** `e2e/tests/anchor-demo.spec.ts`
**When** I run `make e2e`
**Then** the test runs against three Playwright projects (`.NET`, `Django`, `Go`) executing the same scenario: log in as Marisol → navigate to a project with a pre-seeded Violation having a `UNDER_REVIEW` CA → click Approve → assert the three-region update happened in one HTTP request → assert `page.on('load')` did not fire → assert `#violation-detail` has Resolved badge → assert `#compliance-tile` shows the new score → assert `#audit-log` has the new row at the top.

**Given** the timing assertions
**When** the test measures
**Then** each stack's local-dev p95 across 20 runs is ≤ 200 ms and cross-stack divergence is ≤ 50 ms p95 (NFR1).

**Given** `@axe-core/playwright` runs in the same scenario
**When** the page settles after the swap
**Then** zero new WCAG 2.1 AA violations are reported (UX-DR39).

**Given** the test fixture seeds the same UUID across stacks (per Story 1.10)
**When** the audit row's actor is read from each stack's `#audit-log`
**Then** the rendered actor name is "Marisol Cervantes" on all three stacks with identical formatting.

## Epic 6: Project Closure & Compliance Gate Enforcement

Aisha's denial-then-recovery flow. Closure gate evaluated server-side every render; the Close button trichotomy is the FR15 contract; dynamic rule parameters complete FR38.

### Story 6.1: Implement closure-gate rules `OpenViolationGate` and `RequiredInspectionPerTrade`

As a developer implementing project closure,
I want two pure server-side rule functions that say "is the project closeable, and if not why",
So that FR37 (closure-gate enforcement) is reusable from both the render path (button trichotomy) and the handler path (transition guard).

**Acceptance Criteria:**

**Given** each stack
**When** I inspect its compliance module
**Then** two functions exist with the signature `evaluate(project, violations, inspections) -> GateResult` (.NET `Domain/Entities/Reference/ComplianceRule.cs` strategy classes; Django `compliance/rules.py`; Go `internal/domain/compliance_rules.go`):
- `OpenViolationGate`: passes if there are zero `OPEN` or `IN_PROGRESS` Violations on the project.
- `RequiredInspectionPerTrade`: passes if every Trade in the project's `project_trade_scope` has at least one `COMPLETED` Inspection with a non-`FAIL` outcome.

**Given** `GateResult`
**When** I inspect it
**Then** it carries `is_passed: bool`, `gate_name: string`, `explanation: string` (human-readable, in the cross-stack-identical canonical wording — UX-DR45 not violated because the strings are server-rendered consistently).

**Given** unit tests
**When** I run `make test-{net,django,go}`
**Then** each stack covers: all gates pass; one gate fails; multiple gates fail (handler renders all explanations).

---

### Story 6.2: Implement `Project.close()` entity method with `ClosureGateError`

As a developer implementing the close transition,
I want a single entity method that evaluates all closure gates and either transitions the Project to Closed or raises a typed exception with the explanation,
So that FR14 and the affordance trichotomy (FR15) share one source of truth.

**Acceptance Criteria:**

**Given** each stack
**When** I inspect the Project aggregate
**Then** `project.close(actor)` evaluates Story 6.1's gates; if all pass, sets status `CLOSED`, `closed_at = now()`; if any fail, raises a typed `ClosureGateError` carrying the list of failed `GateResult`s.

**Given** `ClosureGateError`
**When** I inspect it
**Then** the message text is canonical across stacks (the same input projects produce the same error message verbatim on all three stacks — verified by a cross-stack test fixture).

**Given** `project.can_close()` (a query method)
**When** it is called
**Then** it returns the same gate evaluations as `close()` would without performing the transition — used by the render path in Story 6.3.

---

### Story 6.3: Close ActionButton trichotomy and closure flow

As a Project Manager,
I want the Close button to reflect server-decided gate state and to perform closure or render an inline explanation,
So that FR15's affordance trichotomy is the canonical FR6/FR55 demonstration on a non-trivial precondition.

**Acceptance Criteria:**

**Given** the Project Detail Summary tab renders
**When** the Close button is considered
**Then** the trichotomy resolves server-side using `project.can_close()`:
- non-PM/non-Admin → `absent` (FR6)
- PM/Admin + any gate fails → `disabled-with-tooltip` (tooltip text = the gate failure explanations joined; `aria-describedby` linked) (FR15, UX-DR21)
- PM/Admin + all gates pass → `present` enabled (FR15)

**Given** I am authorized and gates pass
**When** I click Close
**Then** `POST /projects/<id>/close` runs the canonical flow → call `project.close(actor)` → status `CLOSED`, `closed_at` set → append `AuditEntry(action="ProjectClosed", before={status: ACTIVE}, after={status: CLOSED})` → commit
**And** the response re-renders `#project-detail` (status Closed, no further actions) plus OOB `#audit-log` row (no `#compliance-tile` change at this transition; score is whatever it was at close time).

**Given** I am authorized but gates fail
**When** I click Close anyway via DevTools/curl
**Then** the entity method raises `ClosureGateError` and the handler returns HTTP `409` with the originating `#project-detail` Summary partial re-rendered showing current state and an InlineAlert containing the explanation (UX-DR22, FR55)
**And** the Close button re-renders as `disabled-with-tooltip` per the same trichotomy
**And** `#compliance-tile` and `#audit-log` are **not** updated (UX-DR22).

**Given** a Closed Project
**When** I view its Project Detail
**Then** the affordance trichotomy collapses to `absent` for all state-changing actions (no Re-open path — Closed is terminal in MVP).

---

### Story 6.4: Aisha's denial-then-recovery user flow end-to-end

As Aisha the Project Manager,
I want to attempt closure on a non-closeable project, see the inline explanation, fix the underlying gap (schedule + complete inspection / resolve violation), and successfully close on the next attempt,
So that Journey 3 from the PRD works as designed end-to-end on all three stacks.

**Acceptance Criteria:**

**Given** a seeded project missing a required `Plumbing` inspection
**When** I click Close
**Then** HTTP `409` returns with the originating partial + InlineAlert reading the canonical text "Closure blocked: trade Plumbing has no completed inspection" (UX-DR22) — exact text identical across stacks.

**Given** I navigate to the Inspections tab, schedule a Plumbing inspection, complete it with Pass outcome
**When** I return to the Summary tab
**Then** `project.can_close()` re-evaluates on the new render and the Close button is rendered `present` enabled.

**Given** I click Close
**When** the transition succeeds
**Then** the project moves to `CLOSED`, `ProjectClosed` audit row appears at the top of the audit log (verified in the Audit tab).

**Given** an E2E Playwright scenario for Journey 3
**When** it runs against all three stacks
**Then** the denial path and recovery path both pass with identical observable behavior; cross-stack timing divergence remains ≤ 50 ms p95 (NFR1).

---

### Story 6.5: Verify dynamic compliance rule parameter changes take effect without code change

As an operator (or talk demo presenter),
I want to update a compliance rule parameter and see scoring behavior change on the next request,
So that FR38 (dynamic eval) is mechanically demonstrated.

**Acceptance Criteria:**

**Given** a Project with a known compliance score derived from current rule parameters
**When** I run `UPDATE domain.compliance_rule SET severity_weight = <new_value> WHERE rule_name = 'CriticalViolationWeight'` (in psql)
**Then** the next request that triggers a recompute (e.g., voiding a Violation, completing an Inspection) produces a new score reflecting the updated weight — verified by a Playwright scenario or scripted test.

**Given** the change
**When** I do NOT restart any application server
**Then** the new value is picked up because reference data is reloaded per Story 2.3's invalidation hook (FR53, FR38).

**Given** the reference-data invalidation strategy
**When** I read each stack's `reference` module
**Then** the strategy is documented (TTL, version stamp, or explicit invalidate call) and behaves identically across stacks under the same UPDATE.

## Epic 7: Cross-Stack Parity Demonstration & Demo-Ready Quality

Final epic. Locks in the artifact's persuasive purpose: every MVP scenario runs identically on all three stacks under a published harness; visual regression, axe, latency, route/index diff, and component byte-equivalence all pass clean.

### Story 7.1: Full Playwright cross-stack E2E suite for every MVP workflow

As the talk-audience persona,
I want a Playwright suite that runs every MVP user-facing workflow against all three stacks with axe-core embedded,
So that the cross-stack thesis is mechanically verifiable in one command (FR65, UX-DR39).

**Acceptance Criteria:**

**Given** `e2e/tests/`
**When** I inspect it
**Then** one spec exists per MVP user journey (Journey 1 — anchor demo; Journey 2 — Diego reject/resubmit cycle; Journey 3 — Aisha closure denial+recovery; Journey 4 — Kenji read-only browse) plus per-feature specs (project create, place-on-hold, resume, schedule inspection, complete inspection, assign violation, void violation, theme toggle, login/logout, reference-data read).

**Given** `playwright.config.ts`
**When** I inspect it
**Then** three parallel projects (`.NET` @ :5000, `Django` @ :8000, `Go` @ :3000) are configured; `make e2e` runs all specs against all three.

**Given** any spec
**When** it executes against any stack
**Then** an `@axe-core/playwright` scan runs at every meaningful render and fails the spec on any new WCAG 2.1 AA violation.

**Given** the suite completes
**When** I inspect the report
**Then** every spec passes on all three stacks (no per-stack skips).

---

### Story 7.2: Cross-stack visual regression suite

As a maintainer or demo presenter,
I want pixel-snapshot tests of the canonical screens across stacks × themes × viewports,
So that cross-stack divergence is caught at the rendering level — and Basecoat upgrades have a gate (UX-DR38).

**Acceptance Criteria:**

**Given** `e2e/tests/visual/`
**When** I inspect it
**Then** snapshots cover: Compliance Dashboard, Project Detail (Summary, Inspections tab, Violations tab, Audit tab), Violation Detail, Login page, Reference Data pages
**And** each screen is captured per stack (3) × per theme (light, dark — 2) × per viewport (1280, 1024, 768, 375 — 4) = 24 snapshots per screen.

**Given** the first run produces baselines
**When** they are reviewed and committed
**Then** subsequent runs compare against them with a tight pixel-difference threshold (≤ 0.1% for layout-stable regions; documented exception list in `e2e/visual/exceptions.md`).

**Given** any cross-stack pixel divergence beyond the threshold
**When** the suite runs
**Then** it fails the build with a diff image attached (UX-DR38).

**Given** the Basecoat version pinned in `fieldmark_shared/package.json`
**When** the version is bumped
**Then** the suite must be re-baselined as a coordinated three-stack story — the gating mechanism is the suite itself.

---

### Story 7.3: Color-blindness simulation and 200% browser-zoom verification

As an accessibility-conscious user,
I want assurance that status badges remain distinguishable under deuteranopia/protanopia and that the layout holds at 200% browser zoom,
So that WCAG 1.4.1 and 1.4.4 are mechanically verified (UX-DR37).

**Acceptance Criteria:**

**Given** `e2e/tests/a11y-color.spec.ts`
**When** it runs
**Then** Playwright applies a color-vision filter (deuteranopia, then protanopia) to canonical screens and asserts that status badges are distinguishable by their text labels (the test reads each badge's accessible text and confirms uniqueness within a list).

**Given** `e2e/tests/a11y-zoom.spec.ts`
**When** it runs
**Then** Playwright sets viewport zoom to 200% on canonical screens and asserts: no horizontal scrolling appears (outside the AG Grid acknowledged exception), no content is clipped, no interactive element loses focusable area below 44×44px.

**Given** both specs
**When** the suite runs
**Then** they execute against all three stacks (any per-stack divergence is a defect).

---

### Story 7.4: Cross-stack latency-divergence verification

As the project's NFR1 author,
I want each MVP scenario to assert a per-stack p95 ≤ 200 ms and a cross-stack divergence ≤ 50 ms p95,
So that the performance contract is mechanically tested, not asserted (NFR1).

**Acceptance Criteria:**

**Given** every E2E spec
**When** it measures action→repaint timing for the canonical interactions (Approve CA, Place On Hold, Resume, Complete Inspection, Close Project, Reject CA)
**Then** it captures p95 across N≥20 runs per stack
**And** asserts each stack's p95 ≤ 200 ms locally.

**Given** the same scenario across stacks
**When** the cross-stack divergence is computed (max(p95) − min(p95))
**Then** the test asserts the difference is ≤ 50 ms; otherwise it fails.

**Given** AG Grid row→detail interactions
**When** measured
**Then** p95 ≤ 300 ms per stack with the same cross-stack divergence rule.

---

### Story 7.5: Domain method unit-test coverage closure

As the project's FR66 author,
I want every state-transition entity method on every aggregate to have unit tests proving its invariants in each stack,
So that the domain-rule contract is enforced at the language level — not only via E2E.

**Acceptance Criteria:**

**Given** the canonical method list (`start`, `complete`, `cancel`, `place_on_hold`, `resume`, `close`, `assign`, `submit_corrective_action`, `approve_resolution`, `reject_resolution`, `void`)
**When** I inspect each stack's domain test project (`FieldMark.Tests.Domain/`, per-app `tests/test_*_state.py`, `internal/domain/*_test.go`)
**Then** every method has positive tests (happy path) and negative tests (precondition violations) — at least one test per entity rule documented in `domain-model.md`.

**Given** the cross-stack rule "method names are canonical"
**When** I `grep` each stack
**Then** the same method names appear with idiomatic casing (.NET PascalCase, Python snake_case, Go PascalCase exported / camelCase unexported) — any divergence fails Story 7.6's parity check.

**Given** `make test-{net,django,go}`
**When** each runs
**Then** all domain unit tests pass on each stack, and per-stack coverage tooling (where supported) reports ≥ 90% line coverage on the `domain/` packages.

---

### Story 7.6: Final cross-stack inventory check — routes, indexes, audit strings

As the artifact's stack-symmetry author,
I want a single command that asserts zero diff on routes, `pg_indexes`, audit action constants, HTMX target IDs, and AG Grid endpoint contracts,
So that FR58 (final) and NFR7 are mechanically gated (FR58, FR59).

**Acceptance Criteria:**

**Given** `make parity`
**When** I run it at the end of every MVP-finishing PR
**Then** it executes: `tools/parity/diff-routes.sh`, `tools/parity/diff-pg-indexes.sh`, `tools/parity/diff-audit-actions.sh` (new — greps each stack's audit-action enum/constants and compares to the canonical 14-string list), `tools/parity/diff-target-ids.sh` (new — greps each stack's templates for HTMX `id="..."` and asserts only the canonical inventory appears), `tools/parity/diff-grid-endpoints.sh` (new — asserts the four `/grid/*` endpoints respond with the contract shape on all three stacks).

**Given** any divergence
**When** the script runs
**Then** it exits non-zero with a human-readable diff identifying which stack diverged and on which contract.

**Given** the README at repo root
**When** I read the "Demo Run" section
**Then** it documents the order: `make reset && make run-{net,django,go} && make parity && make e2e` as the smoke-test recipe.

---

### Story 7.7: Canonical component example gallery with per-stack snapshot tests

As a developer adding or modifying any custom component,
I want a canonical static-HTML gallery of every component and per-stack tests that the wrapper output is byte-identical,
So that component drift across stacks is caught at unit-test time, not in E2E (UX-DR40).

**Acceptance Criteria:**

**Given** `fieldmark_shared/components/`
**When** I list it
**Then** each of StatusBadge, ActionButton, ComplianceTile, AuditRow, EntityRail, DashboardTile, InlineAlert, ThemeToggle, AGGridPanel, TabStrip, FlashRegion has its own subdirectory with `<component>.example.html` files showing every state combination.

**Given** each stack's component test suite
**When** it runs
**Then** for each canonical example, it renders the per-stack wrapper with the same inputs and asserts byte-equivalence against the example HTML (whitespace-normalized).

**Given** any wrapper produces non-byte-identical output
**When** the test runs
**Then** it fails with a diff identifying the divergent component and stack.

---

### Story 7.8: AG Grid axe ruleset, manual SR test recipe, and Demo Run documentation

As the artifact's accessibility-stewardship author,
I want AG Grid axe disables documented per rule with rationale, a manual screen-reader test recipe, and a one-line "Demo Run" recipe in the README,
So that the artifact's accessibility posture is honest and reproducible (UX-DR36, NFR3).

**Acceptance Criteria:**

**Given** `tests/axe-config.json` (or per-stack equivalent referenced by `@axe-core/playwright`)
**When** I read it
**Then** every disabled AG Grid axe rule is listed with: rule id, AG Grid version, rationale, review-on-upgrade flag
**And** disables are reviewed each AG Grid upgrade per UX-DR36.

**Given** `_bmad-output/planning-artifacts/manual-a11y-recipe.md` (new)
**When** I read it
**Then** it documents: how to run VoiceOver on Safari/macOS and NVDA on Firefox/Windows against the canonical anchor demo; what to assert at each step; what acceptable observed behavior looks like; cadence (per major milestone, quarterly minimum).

**Given** the repo root `README.md`
**When** I read its "Demo Run" section
**Then** it documents the one-line recipe (`make reset && make run-{net,django,go}` in three terminals, then `make parity && make e2e` to verify, then navigate to each stack's URL to demo) and a brief talking-track outline pointing at Stories 5.5 (anchor demo), 6.4 (denial+recovery), and the cross-stack parity assertion.
