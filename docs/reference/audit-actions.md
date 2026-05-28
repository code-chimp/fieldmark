# Audit Actions — Canonical Contract

> **Status:** live — populated by Story 2.2 (`domain.audit_entry` + `append_audit_entry()` helper), 2026-05-27.
> Scaffolded by Epic 1 retrospective action item A4 (2026-05-25).

This document is the **single source of truth** for the canonical set of audit-action strings emitted by `domain.audit_entry.action`. Each stack (.NET, Django, Go) implements a native enum/constants module whose values must match this list exactly. Per-stack conformance tests assert that alignment.

See the root [CLAUDE.md](../../CLAUDE.md) **Cross-Stack Architecture Principle** for why this lives as documentation rather than as shared code.

---

## Canonical Action List

Fifteen action strings total. Stored verbatim in `domain.audit_entry.action`. Adding or removing an action requires the **Change Procedure** at the bottom of this document.

| Action | Entity | Emitted when | Story that introduces emission | Notes |
|---|---|---|---|---|
| `ProjectCreated` | Project | `Project.create(...)` succeeds and the project row is written | Story 2.8 | `before_state` is `NULL`; `after_state` snapshots the new project |
| `ProjectPlacedOnHold` | Project | `project.place_on_hold(reason)` transitions `ACTIVE → ON_HOLD` | Story 2.12 | `metadata.reason` carries the user-supplied reason string |
| `ProjectResumed` | Project | `project.resume()` transitions `ON_HOLD → ACTIVE` | Story 2.12 | `metadata.reason` optional |
| `ProjectClosed` | Project | `project.close(actor)` transitions `ACTIVE → CLOSED` after all closure gates pass | Story 6.2 | `before_state.status = ACTIVE`, `after_state.status = CLOSED` |
| `InspectionScheduled` | Inspection | `project.schedule_inspection(trade, inspector, when)` writes a new `domain.inspection` row with `status=SCHEDULED` | Story 3.4 | `after_state` snapshots the scheduled inspection |
| `InspectionStarted` | Inspection | `inspection.start(actor)` transitions `SCHEDULED → IN_PROGRESS` | Story 3.8 | `after_state.started_at = now()` |
| `InspectionCompleted` | Inspection | `inspection.complete(outcome, notes, findings)` transitions `IN_PROGRESS → COMPLETED` | Story 3.9 | Same transaction may also emit one or more `ViolationOpened` per spawned violation |
| `InspectionCancelled` | Inspection | `inspection.cancel(reason)` transitions any pre-terminal state to `CANCELLED` | Story 3.10 | `metadata.reason` required |
| `ViolationOpened` | Violation | `Violation.open_from_finding(finding, severity, due_date)` writes a new `domain.violation` row | Story 3.9 | Always co-emitted with the parent `InspectionCompleted` in the same transaction |
| `ViolationAssigned` | Violation | `violation.assign(supervisor_id)` or a reassignment self-transition | Story 4.4 | `before_state.assigned_to`, `after_state.assigned_to`, `metadata.note` optional |
| `ViolationVoided` | Violation | `violation.void(actor, reason)` transitions a non-terminal violation to `VOIDED` | Story 4.7 | Same transaction recomputes project compliance score |
| `CorrectiveActionSubmitted` | CorrectiveAction | `violation.submit_corrective_action(...)` writes a CA with `status=SUBMITTED` and, if the violation was `OPEN`, transitions it to `IN_PROGRESS` | Story 5.2 | CA submission, not the violation transition, is the recorded event |
| `CorrectiveActionTakenForReview` | CorrectiveAction | A reviewer claims the CA, transitioning it `SUBMITTED → UNDER_REVIEW` | Story 5.4 | `before_state.status = SUBMITTED`, `after_state.status = UNDER_REVIEW` |
| `CorrectiveActionApproved` | Violation | `violation.approve_resolution(ca_id, reviewer)` — CA → `APPROVED`, Violation → `RESOLVED` | Story 5.5 | Anchor demo. Same transaction recomputes project compliance score. `entity_type = "Violation"` (the resolution is the violation-level event) |
| `CorrectiveActionRejected` | CorrectiveAction | `violation.reject_resolution(ca_id, reviewer, notes)` — CA → `REJECTED`, violation stays `IN_PROGRESS` | Story 5.6 | `metadata.notes_length` (review notes are stored on the CA, not the audit metadata) |

---

## Casing Convention

The persisted string in `domain.audit_entry.action` is **`PascalCase`, present-tense past-form** — e.g. `ProjectPlacedOnHold`, `CorrectiveActionTakenForReview`. Not `placeOnHold`, not `PROJECT_PLACED_ON_HOLD`, not `project.placed_on_hold`.

The convention is **binding across stacks** at the persisted-string level. Per-stack symbol names follow the host language's idiom — the conformance tests in AC6 prove that whatever the symbol is, it round-trips to the canonical PascalCase string:

| Stack | Symbol form | Example |
|---|---|---|
| .NET (C#) | `enum` member, PascalCase | `AuditAction.ProjectPlacedOnHold` |
| Django (Python) | `TextChoices` class member, `SCREAMING_SNAKE_CASE` symbol with PascalCase value | `AuditAction.PROJECT_PLACED_ON_HOLD = "ProjectPlacedOnHold", "ProjectPlacedOnHold"` |
| Go | `const` of typed `AuditAction string`, exported PascalCase with type prefix | `AuditActionProjectPlacedOnHold AuditAction = "ProjectPlacedOnHold"` |

---

## Per-Stack Native Implementations

Each stack owns its own enum/constants module. The top-of-file comment must reference this document.

- **.NET** — [`FieldMark/FieldMark.Domain/ValueObjects/AuditAction.cs`](../../FieldMark/FieldMark.Domain/ValueObjects/AuditAction.cs)
- **Django** — [`fieldmark_py/audit/actions.py`](../../fieldmark_py/audit/actions.py)
- **Go** — [`fieldmark-go/internal/domain/enums/audit_action.go`](../../fieldmark-go/internal/domain/enums/audit_action.go)

The corresponding `append_audit_entry()` helpers:

- **.NET** — [`FieldMark/FieldMark.Data/Auditing/AuditAppender.cs`](../../FieldMark/FieldMark.Data/Auditing/AuditAppender.cs) (interface in [`IAuditAppender.cs`](../../FieldMark/FieldMark.Data/Auditing/IAuditAppender.cs))
- **Django** — [`fieldmark_py/audit/append.py`](../../fieldmark_py/audit/append.py)
- **Go** — [`fieldmark-go/internal/data/postgres/auditentrystore.go`](../../fieldmark-go/internal/data/postgres/auditentrystore.go)

---

## Conformance Test Contract

A checked-in JSON fixture lives alongside this document at [`audit-actions.json`](audit-actions.json). It mirrors the action list above, in the **same order** (not alphabetical — story-flow order, matching the table).

Each stack ships a unit test that:

1. Walks up from the test working directory to locate `docs/reference/audit-actions.json` at the repo root.
2. Parses the `actions` array.
3. Extracts the stack's native action set.
4. Asserts set equality — no extras, no missing. On failure, prints the symmetric diff (`expected ∖ actual` and `actual ∖ expected`) so the developer sees both directions at once.

The test does not require a database.

Per-stack test paths:

| Stack | Path |
|---|---|
| .NET | [`FieldMark/FieldMark.Tests.Domain/ValueObjects/AuditActionConformanceTests.cs`](../../FieldMark/FieldMark.Tests.Domain/ValueObjects/AuditActionConformanceTests.cs) |
| Django | [`fieldmark_py/audit/tests/test_action_conformance.py`](../../fieldmark_py/audit/tests/test_action_conformance.py) |
| Go | [`fieldmark-go/internal/domain/enums/audit_action_conformance_test.go`](../../fieldmark-go/internal/domain/enums/audit_action_conformance_test.go) |

---

## Change Procedure

Adding or removing an audit action is a four-step change:

1. **ADR amendment** — add an `## ADR Amendment` block to the relevant epic file *and* this document, citing the FR or PRD requirement that drives the change. Precedent: the `ProjectCreated` amendment recorded in [epic-2 §Story 2.8 note](../../_bmad-output/planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md).
2. **Update the canonical list** above and the [`audit-actions.json`](audit-actions.json) fixture in the same commit. The two MUST remain in the same order — story-flow order, not alphabetical. A reviewer should be able to diff doc and fixture line-by-line.
3. **Add the symbol to each stack's native enum/constants module**, then re-run the three conformance tests (`.NET AuditActionConformanceTests`, Django `test_action_conformance.py`, Go `audit_action_conformance_test.go`). They must pass without modification beyond the new symbol.
4. **Run `make parity`** from the repo root. `pg_indexes` parity won't catch action-string drift directly (action strings live in row data, not schema), so the conformance tests are the primary gate; `make parity` remains the schema-level gate.

### Reconciliation note (Story 2.2)

[architecture.md:603](../../_bmad-output/planning-artifacts/architecture.md) enumerates **13** PascalCase action strings literally. The Epic 1 retrospective scaffolded this document with the assertion "14 strings ratified at Epic 1 close, plus `ProjectCreated` added by ADR amendment during Epic 2 planning (15 total)" — leaving the 14th unspecified.

Story 2.2 resolved this by cross-walking PRD FR9–FR15 (project lifecycle) and Epic 3/4/5/6 story emissions. The missing 14th string is **`InspectionScheduled`** — emitted by Story 3.4 (`POST /projects/<id>/inspections/schedule`, [epic-3 §Story 3.4](../../_bmad-output/planning-artifacts/epics/epic-3-inspection-workflow-violation-genesis.md#story-3.4)) but omitted from the architecture-line-603 enumeration.

Rejected candidates and why:
- `ViolationResolved` — Epic 5 Story 5.5 shows the resolution is captured by `CorrectiveActionApproved` (which carries the violation `IN_PROGRESS → RESOLVED` transition in `after_state`); a separate `ViolationResolved` emission would double-count the event and contradict FR40.
- `ViolationInProgress` — Epic 5 Story 5.2 shows the violation `OPEN → IN_PROGRESS` transition is implicit in `CorrectiveActionSubmitted`; no separate emission.
- `ProjectClosureBlocked` — closure-gate failures raise `ClosureBlockedException` and never write an audit row (the transaction rolls back); not an audit action.

`InspectionScheduled` is therefore the only flow-driven emission referenced by an epic story but absent from architecture-line-603. This change has been propagated into the canonical list above; the architecture document should be amended in the next architecture-update pass.
