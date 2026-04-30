# FieldMark UX Planning Guide

## Purpose
This document provides a **UX-first planning skeleton** for FieldMark: Construction Compliance & Inspection Management System. It is intended to seed collaborative UX design sessions and align designers with core architectural principles: backend authority, workflow clarity, and enterprise realism.

The guidance here is **technology-agnostic** and applies equally to Razor Pages + HTMX and Django + HTMX implementations.

---

## UX Design Principles (Anchor These First)

1. **The system owns truth** – the UI reflects authoritative state.
2. **Every screen answers one question** – avoid multipurpose dashboards.
3. **State is visible** – no hidden transitions or implicit client assumptions.
4. **Progress is explicit** – workflows advance only when allowed by rules.
5. **Enterprise calm** – minimal animation, predictable layouts, readable density.

---

## UX Planning Chart (Screen Inventory)

| Area | Screen | Purpose | Primary Interactions | Notes |
|-----|-------|---------|----------------------|-------|
| Dashboard | Compliance Overview | Portfolio-level status | Drill-down, filters | HTMX partial refresh |
| Projects | Project List | Navigate system of record | Select project | AG Grid |
| Projects | Project Detail | Anchor screen | Inspect, act, review | Most important screen |
| Inspections | Inspection List | Track inspections | View, schedule | Grid + detail swap |
| Inspections | Inspection Workflow | Execute inspection | Step progression | Linear, server-driven |
| Violations | Violation List | Monitor risk | Resolve, escalate | Rule-gated actions |
| Violations | Violation Detail | Corrective actions | Evidence, close | Authority moment |
| Audit | Activity Log | Compliance proof | Read-only | Immutable |
| Admin | Reference Data | Configure rules/codes | CRUD | Minimal UX |

---

## Navigation Model (Conceptual)

```
Dashboard
   │
   ├── Projects
   │      └── Project Detail
   │             ├── Inspections
   │             │      └── Inspection Workflow
   │             ├── Violations
   │             │      └── Violation Detail
   │             └── Audit Log
   │
   └── Admin (Role-Gated)
```

Navigation is **state-driven**, not exploratory.

---

## Key Screen Wireframe (Conceptual)

### Project Detail (Anchor Screen)

```
┌────────────────────────────────────────────┐
│ Project: Riverside Substation Upgrade       │
│ Status: Active | Compliance: 82%            │
└────────────────────────────────────────────┘

[ Summary ] [ Inspections ] [ Violations ] [ Audit ]

--------------------------------------------------
| Summary Panel                                  |
| - Trade scopes                                 |
| - Key dates                                   |
--------------------------------------------------
| Inspections (Grid)                             |
--------------------------------------------------
| Violations (Severity tagged)                   |
--------------------------------------------------
```

Design intent: one page shows **current truth**, not editable drafts.

---

## UX Sequence Diagram (Core Demo Flow)

### Scenario: Resolving a Violation

```
User
 │
 │ Click "Resolve Violation"
 │──────────────────────────────▶ UI
 │                               │
 │                               │ POST /violations/{id}/resolve
 │                               │──────────────────────────────▶ Backend
 │                               │                               │
 │                               │                               │ Validate rules & prerequisites
 │                               │                               │ Update domain state
 │                               │                               │ Recalculate compliance
 │                               │                               │
 │                               │◀──────────────────────────────│ HTML Partial
 │                               │
 │ UI Updates:
 │ - Violation status
 │ - Compliance score tile
 │ - Audit log entry
 │
 ◀────────────────────────────── User sees authoritative state
```

No client-side orchestration. No duplicated logic.

---

## Django vs .NET Admin UX Alignment

### Principle
**Administrative UX is a platform concern, not a product experience.**

### Django
- Use Django Admin for:
  - Inspection codes
  - Violation categories
  - Rule metadata
- Position as backend tooling

### .NET
- Minimal Razor Pages for:
  - Reference data
  - Configuration
- Match capability, not UI polish

Explicitly de-scope admin UX from primary demo narrative.

---

## UX Risks & Mitigations

| Risk | Mitigation |
|----|-----------|
| Over-design | Focus on states, not visuals |
| SPA mimicry | Keep interactions HTTP-visible |
| Hidden authority | Always show why actions are allowed/blocked |
| Demo overload | Pre-script one workflow |

---

## UX Deliverables for Kickoff Session

- Screen inventory (this chart)
- One detailed Project Detail wireframe
- One workflow sequence (inspection or violation)
- Admin capability matrix

These are sufficient to guide architecture, implementation, and demos.

---

## Summary

FieldMark’s UX should feel calm, authoritative, and inevitable. The interface does not negotiate with users—it reflects decisions already made by the system. The strongest UX success criterion is not delight, but **trust**.

