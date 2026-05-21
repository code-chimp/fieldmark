# Epic List

## Epic 1: Walking Skeleton — Auth, Design System & Parity Foundation

Lay the cross-stack foundation: confirm the three native CLI skeletons, author the infrastructure-owned `domain.*` DDL, stand up the parity tooling (`tools/parity/` + `make parity`), bootstrap the design system (Basecoat pinned, semantic color tokens, Phase 1 custom components), wire framework-native authentication (.NET Identity with `dotnet_auth`, Django `auth` with `django_auth`, Go stub middleware per ADR-012), seed identical-UUID dev users, and render an identical role-aware login + empty home page with the working light/dark theme toggle and FlashRegion across all three stacks. After this epic `make up && make run-{net,django,go}` produces three stacks rendering byte-identical chrome and `make parity` runs clean.
**FRs covered:** FR1, FR2, FR3, FR4, FR5, FR6, FR7, FR8, FR55a (introduced at login form), FR58 (initial), FR60, FR61, FR62, FR63, FR64
**UX-DRs:** UX-DR1, UX-DR2, UX-DR3, UX-DR4, UX-DR5, UX-DR6, UX-DR7, UX-DR8, UX-DR14, UX-DR15, UX-DR31, UX-DR32, UX-DR33, UX-DR34, UX-DR35

## Epic 2: Project Lifecycle & Compliance Dashboard

Admin/Project Manager creates Projects with required metadata; authorized users view the Compliance Dashboard (portfolio tiles + AG Grid project list with server-side row model); any authorized user drills into the Project Detail anchor screen (header strip with ComplianceTile, tabbed content area, EntityRail). Project Manager places Projects on hold and resumes them; transitions write AuditEntries in the same transaction and OOB-swap `#compliance-tile` + `#audit-log` — first lighting up the three-region orchestration pattern on a simple transition before it carries the anchor demo in Epic 5. Reference data (Trade Types, Violation Categories, Compliance Rules) readable by Administrators. AG Grid endpoint contract, 409 + originating partial, POST-only state changes, and one-transaction discipline are established here as cross-cutting patterns reused by every subsequent epic.
**FRs covered:** FR9, FR10, FR11, FR12, FR13, FR36 (introduced), FR39, FR40, FR41, FR42, FR43, FR44, FR45, FR46, FR47, FR48 (projects grid), FR49, FR50, FR51, FR52, FR53, FR54, FR55, FR56, FR57
**UX-DRs:** UX-DR9, UX-DR10, UX-DR11, UX-DR12, UX-DR13, UX-DR16, UX-DR17, UX-DR18, UX-DR19, UX-DR20 (introduced), UX-DR21, UX-DR22, UX-DR23, UX-DR24, UX-DR25, UX-DR26, UX-DR27, UX-DR28, UX-DR29, UX-DR30

## Epic 3: Inspection Workflow & Violation Genesis

Compliance Officer schedules Inspections (trade, inspector, scheduled time); Inspector starts them and completes them with outcome (Pass / Fail / Conditional), notes, and findings; Compliance Officer can cancel a Scheduled Inspection with reason. When an Inspection is completed with Fail-class findings, the system automatically opens a Violation for each finding atomically in the same transaction. Inspections tab on Project Detail renders the list (AG Grid) and loads Inspection detail into EntityRail on row select. Server-side compliance rule evaluation is introduced; score recomputes on Inspection completion and Violation opening in the same transaction.
**FRs covered:** FR16, FR17, FR18, FR19, FR20, FR21, FR34, FR35 (inspection-completed / violation-opened paths), FR48 (inspections grid)

## Epic 4: Violation Lifecycle & Assignment

Site Supervisor sees a queue of Violations assigned to them, filterable by status and overdue flag. Compliance Officer assigns and reassigns Violations to Site Supervisors (reassignment while InProgress emits an audit entry as a self-transition). System marks Violations as overdue server-side once due date passes and status is non-terminal. Administrator can void any non-terminal Violation with a recorded reason; voided Violations do not affect compliance scoring and score recomputes on void. Resolved or Voided Violations cannot be reopened — the user path is to create a new Violation from a new Finding.
**FRs covered:** FR22, FR22a, FR23, FR24, FR25, FR26, FR27, FR35 (void path), FR48 (violations grid)

## Epic 5: Corrective Action Workflow — The Anchor Demo

The thesis-proving epic. Site Supervisor submits a Corrective Action on an InProgress Violation (Violation Open → InProgress on first submission); Compliance Officer takes the CA for review (Submitted → UnderReview); approves it (CA → Approved, Violation → Resolved, compliance score recomputed) atomically in one transaction — and in one HTTP response the primary `#violation-detail` partial swap, OOB `#compliance-tile` update, and OOB `#audit-log` row append all land in a single paint, across all three stacks. Rejection-with-notes keeps the Violation in InProgress and re-renders the Submit affordance so Pat can resubmit; multiple CAs accumulate, only the latest non-Rejected is eligible for approval, and the submitter cannot review their own CA. This is where the canonical three-region orchestration pattern (UX-DR20) and audit-as-receipt pattern (UX-DR23) are verified end-to-end.
**FRs covered:** FR28, FR29, FR30, FR31, FR32, FR33, FR35 (CA-approved/resolved paths), FR36 (anchor canonical OOB)
**UX-DRs:** UX-DR20 (canonical demonstration)

## Epic 6: Project Closure & Compliance Gate Enforcement

Project Manager attempts to close an Active Project: the closure gate (`OpenViolationGate` + `RequiredInspectionPerTrade`) is evaluated server-side on every render of the Project Detail Summary; the Close button is `absent | disabled-with-tooltip | present` per the affordance trichotomy, with the disabled tooltip text coming from the server-side `ClosureGateError` message (which is part of the cross-stack parity contract). On click-when-blocked the server returns HTTP 409 with the originating partial re-rendered showing current state and an inline alert — no modal, no toast, no URL change. Once the gate passes (because the user scheduled the required inspection and resolved the violations through Epics 3–5), closure succeeds and writes a `ProjectClosed` audit entry. Compliance rule parameters (severity weights, due-offset values) are persisted in reference data and consulted dynamically by the rules engine — changing a parameter changes evaluation behavior without code changes.
**FRs covered:** FR14, FR15, FR37, FR38

## Epic 7: Cross-Stack Parity Demonstration & Demo-Ready Quality

Delivers value to the Talk Audience persona (Journey 5). Final Playwright cross-stack E2E coverage of every MVP user workflow × 3 stacks with `@axe-core/playwright` embedded in every scenario (FR65); domain-method unit-test coverage closed out per stack in each stack's idiomatic framework (FR66); visual regression suite captures canonical screens (Compliance Dashboard, Project Detail, Violation Detail) × 3 stacks × 2 themes × 3 viewports (1280/1024/768/375); color-blindness simulation (deuteranopia, protanopia) and 200% browser zoom verification on canonical screens; cross-stack latency-divergence verification (≤ 50 ms p95); `pg_indexes` zero-diff and route-inventory zero-diff verified as the final cross-stack inventory check (FR58 final); canonical component example gallery in `fieldmark_shared/components/` complete and tested against each stack's wrappers; documented "demo run" recipe in the repository root README.
**FRs covered:** FR58 (final), FR59, FR65, FR66
**UX-DRs:** UX-DR36, UX-DR37, UX-DR38, UX-DR39, UX-DR40
