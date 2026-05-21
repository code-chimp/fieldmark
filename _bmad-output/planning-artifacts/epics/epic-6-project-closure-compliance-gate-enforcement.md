# Epic 6: Project Closure & Compliance Gate Enforcement

Aisha's denial-then-recovery flow. Closure gate evaluated server-side every render; the Close button trichotomy is the FR15 contract; dynamic rule parameters complete FR38.

## Story 6.1: Implement closure-gate rules `OpenViolationGate` and `RequiredInspectionPerTrade`

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

## Story 6.2: Implement `Project.close()` entity method with `ClosureGateError`

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

## Story 6.3: Close ActionButton trichotomy and closure flow

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

## Story 6.4: Aisha's denial-then-recovery user flow end-to-end

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

## Story 6.5: Verify dynamic compliance rule parameter changes take effect without code change

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
