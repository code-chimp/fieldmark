# Functional Requirements

This section is the **capability contract** for FieldMark. Every feature implemented across the three stacks must trace back to a requirement listed here. A capability not on this list will not exist in the product. Capabilities required only for Growth or Vision phases are explicitly marked.

## Authentication & Identity

- **FR1:** A user can authenticate against the application using credentials managed by the framework's native authentication system. Authentication is framework-local per ADR-012; cross-stack login session sharing is explicitly not required.
- **FR2:** The system can identify the authenticated user on every request and resolve their conceptual role(s) (Administrator, Compliance Officer, Inspector, Site Supervisor, Executive).
- **FR3:** A user can log out, terminating their session within the stack they are authenticated against.
- **FR4:** Unauthenticated users are redirected to the framework-local login page when attempting to access any business route.

## Authorization

- **FR5:** The system can determine, for any given action on any given entity, whether the current user is authorized to perform it.
- **FR6:** The system never renders an action button (or other action affordance) for an action the current user is not authorized to perform.
- **FR7:** The system rejects any direct request (HTTP-level) to perform an action the user is not authorized for, regardless of UI state, returning an authorization-failure response.
- **FR8:** Role assignments per user are managed by the framework-native authorization machinery in each stack; the application reads them, it does not own a shared role schema.

## Project Management

- **FR9:** A Project Manager or Administrator can create a new Project with required metadata (code, name, start date, optional target completion date, trade scope assignments, inspector assignments).
- **FR10:** A Project Manager, Compliance Officer, or Administrator can view a list of Projects, filtered and sorted by status and compliance score.
- **FR11:** Any authorized user can view a Project Detail page showing the project's current status, compliance score, assigned trades, inspections, violations, and audit log — all in a single rendered view with HTMX-driven tab swaps.
- **FR12:** A Project Manager can transition a Project from Active to OnHold with a recorded reason.
- **FR13:** A Project Manager can transition a Project from OnHold back to Active.
- **FR14:** A Project Manager can transition a Project from Active to Closed only when all closure-gate rules are satisfied; the system blocks closure with a rendered explanation when gates are unmet.
- **FR15:** The system displays the result of `can_close()` as the rendered state of the Close action (button absent for unauthorized users, disabled with explanation for users authorized but blocked, enabled when permitted).

## Inspection Management

- **FR16:** A Compliance Officer or Administrator can schedule an Inspection on a Project, specifying trade type, inspector, and scheduled time.
- **FR17:** An Inspector can transition an Inspection from Scheduled to InProgress (start it).
- **FR18:** An Inspector can transition an Inspection from InProgress to Completed, supplying an outcome (Pass, Fail, or Conditional), notes, and zero or more findings.
- **FR19:** A Compliance Officer or Administrator can cancel a Scheduled Inspection with a recorded reason.
- **FR20:** When an Inspection is Completed with Fail-class findings, the system automatically opens a Violation for each finding, atomically within the same transaction.
- **FR21:** The system displays Inspection lists scoped to a Project and to the current user's role, filterable by status and date range.

## Violation Management

- **FR22:** The system records a Violation's origin Finding, severity, due date, and current status.
- **FR22a:** The Violation's due date is computed at open time from its severity and is immutable thereafter; no UI or API path supports modifying due date after the Violation has been opened.
- **FR23:** A Compliance Officer can assign or reassign a Violation to a Site Supervisor; reassignment while in InProgress is a self-transition that emits an audit entry.
- **FR24:** A Site Supervisor or assigned user can view the Violations assigned to them, filterable by status and overdue flag.
- **FR25:** The system marks a Violation as overdue when its due date has passed and its status is non-terminal.
- **FR26:** An Administrator can void a Violation in any non-terminal state with a recorded reason; voided violations do not affect compliance scoring.
- **FR27:** The system prevents reopening of a Resolved or Voided Violation; users who require revisiting the underlying issue must create a new Violation from a new Finding.

## Corrective Action Management

- **FR28:** A Site Supervisor (or assigned remediator) can submit a Corrective Action for a Violation in InProgress state, including a description and optional evidence reference.
- **FR29:** A Compliance Officer can take a submitted Corrective Action for review, transitioning its state to UnderReview.
- **FR30:** A Compliance Officer can approve a Corrective Action in UnderReview state, atomically transitioning the parent Violation to Resolved within the same transaction.
- **FR31:** A Compliance Officer can reject a Corrective Action in UnderReview state with review notes; the parent Violation remains in InProgress and a new Corrective Action may be submitted.
- **FR32:** The system prevents the submitter of a Corrective Action from being its reviewer.
- **FR33:** Only the latest non-Rejected Corrective Action of a Violation may be approved.

## Compliance Rules Engine & Scoring

- **FR34:** The system evaluates compliance rules entirely server-side; clients never compute compliance state.
- **FR35:** The system recomputes a Project's compliance score on every state transition that affects it (Violation opened, resolved, voided; Corrective Action approved; Inspection completed) within the same transaction as the triggering write.
- **FR36:** The system surfaces the current compliance score of a Project on the Project Detail screen and updates it via HTMX out-of-band swap when a same-page action affects it.
- **FR37:** The system enforces the closure gate rules (`OpenViolationGate`, `RequiredInspectionPerTrade`) at the moment of attempted closure; closure is rejected with an inline explanation when any gate is unmet.
- **FR38:** Compliance rules and their parameters (severity weights, due-offset values) are persisted in reference data and evaluated dynamically; changing a parameter changes evaluation behavior without code changes. _(The persistence and evaluation are MVP; the admin UI to edit them is Growth-phase per FR67.)_

## Audit Trail

- **FR39:** The system writes an AuditEntry for every domain mutation in the same database transaction as the change.
- **FR40:** Each AuditEntry records the actor (opaque user UUID per ADR-012), action, entity type and ID, the project it belongs to (when applicable), the before/after state as JSON, and any metadata (reason strings, role at time of action).
- **FR41:** AuditEntries are append-only; the system does not provide any UI or API path to update or delete an existing entry.
- **FR42:** Any authorized user can view a Project's audit log, ordered most-recent-first, accessible from the Project Detail screen.
- **FR43:** Executive-role users can view the audit log read-only, with no path to mutate any entity from that view.

## Dashboard & Reporting

- **FR44:** A user can view the Compliance Dashboard showing portfolio-level aggregates: average compliance score, count of overdue violations by severity, count of Projects by lifecycle state.
- **FR45:** The Compliance Dashboard renders via HTMX-driven partial refresh; tile-level updates do not require full-page reloads.
- **FR46:** A user can drill from the Compliance Dashboard into a Project Detail screen via HTMX swap; navigation does not trigger a full-page reload.
- **FR47:** A user can filter and sort the Project list by status, compliance score, and ownership.

## Data Grid Interactions (AG Grid)

- **FR48:** The system serves AG Grid data via a server-side row model on at least two views (Project list and one of: Inspection list, Violation list, Audit log).
- **FR49:** AG Grid endpoints return JSON conforming to the cross-stack contract `{ "rows": [...], "lastRow": N }`.
- **FR50:** Row selection in any AG Grid fires an HTMX request that loads the corresponding detail panel; the grid does not own the detail rendering.
- **FR51:** AG Grid configurations contain no business rules and no client-side row computation; filters and sorts that affect server data are passed to the server as request parameters.

## Reference Data Administration

- **FR52:** An Administrator can view the catalog of Trade Types, Violation Categories, and Compliance Rules. _(Read access in MVP; full CRUD UI in Growth phase per FR67.)_
- **FR53:** Reference data is loaded from the database on application start and on changes; no application restart is required to pick up reference-data changes.

## Cross-Cutting System Behavior

- **FR54:** State-changing actions in the UI use HTTP POST (or appropriate non-GET method); GET requests never mutate state.
- **FR55:** When a domain rule rejects an action, the system returns HTTP 409 with the originating partial re-rendered showing the error and the unchanged state.
- **FR55a:** When client-submitted data fails server-side input validation (malformed input, type error, missing required field — distinct from a domain rule violation), the system returns HTTP 422 with the originating form partial re-rendered showing field-level errors. Each invalid field carries `aria-invalid="true"` and `aria-describedby` linking to its inline error message; a top-level InlineAlert with `role="alert"` summarises the error count and links to the first invalid field. No state is mutated on 422.
- **FR56:** When an authorization check rejects an action, the system returns the appropriate authorization-failure response (HTTP 403 or framework-equivalent) without leaking information about the underlying entity state.
- **FR57:** All domain mutations occur within a transaction that includes the audit-entry write and any required compliance-score recomputation.
- **FR58:** Identical routes, HTMX target IDs, AG Grid endpoint contracts, audit action strings, and domain method names are present across all three stacks (modulo language casing); a diff in any of these is a defect.
- **FR59:** Domain rule violations, authorization failures, and validation errors are handled identically (in observable behavior) across all three stacks.

## Accessibility

- **FR60:** All interactive controls are keyboard-operable with the expected activation keys; tab order matches visual order; focus is visible.
- **FR61:** Form errors are programmatically associated with their inputs via `aria-invalid` and `aria-describedby` and are announced to assistive technology.
- **FR62:** HTMX swaps that represent meaningful state changes shift focus to the swapped region (or to a documented focus-target element within it) so that screen-reader users perceive the change.
- **FR63:** Out-of-band swap targets (compliance score tile, notification badges) carry `aria-live` so updates are announced to assistive technology.
- **FR64:** Buttons that trigger HTMX requests visibly indicate their disabled state during the request to both visual and assistive-technology users.

## Test & Quality (Capability Requirements, not implementation)

- **FR65:** Every MVP user-facing workflow has a Playwright E2E scenario that runs against all three stacks and asserts equivalent observable behavior.
- **FR66:** Every domain method with state-transition logic has unit-test coverage proving its invariants in each stack's idiomatic test framework.

## Growth-Phase Capabilities (Out of MVP)

- **FR67:** _(Growth)_ An Administrator can create, update, and deactivate reference-data records (Trade Types, Violation Categories, Compliance Rules) through a dedicated administration UI.
- **FR68:** _(Growth)_ An Executive can view portfolio-level trend visualizations of compliance score over time.
- **FR69:** _(Growth)_ The system runs an identical multi-stack parity test suite that produces a comparison report.
- **FR70:** _(Growth)_ An Administrator can edit Compliance Rule parameters (severity weights, due offsets) through a UI without code changes; the change takes effect on next score recomputation.
