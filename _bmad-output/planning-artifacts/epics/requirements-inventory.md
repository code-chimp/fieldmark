# Requirements Inventory

## Functional Requirements

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

## NonFunctional Requirements

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

## Additional Requirements

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

## UX Design Requirements

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

## FR Coverage Map

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
