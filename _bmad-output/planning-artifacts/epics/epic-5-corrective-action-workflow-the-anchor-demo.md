# Epic 5: Corrective Action Workflow — The Anchor Demo

The thesis-proving epic. The canonical Marisol Approve flow lands as three-region OOB orchestration in a single HTTP response across all three stacks.

## Story 5.1: Map `domain.corrective_action` writes into each stack's data layer

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

## Story 5.2: Submit Corrective Action (Site Supervisor)

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

## Story 5.3: Render Corrective Action list within Violation detail

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

## Story 5.4: Take Corrective Action for review (Submitted → UnderReview)

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

## Story 5.5: Approve Corrective Action — canonical anchor three-region demo

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

## Story 5.6: Reject Corrective Action with notes (UnderReview → Rejected)

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

## Story 5.7: Cross-stack anchor-demo E2E and timing parity

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
