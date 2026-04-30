# FieldMark Domain Data Model – ERD

## Purpose
This document describes the **foundational domain schema** for FieldMark: Construction Compliance & Inspection Management System.  
It is intended as a **data‑layer design reference** and as priming context for agentic systems.  

The schema favors:
- Backend authority
- Explicit workflows
- Auditability
- Cross‑stack parity (.NET + Django)

The model is intentionally **conservative and extensible**, avoiding premature normalization while remaining enterprise‑credible.

---

## Core Domain Concepts

- **Project** – A construction job or site under compliance oversight
- **Inspection** – A scheduled or performed inspection event
- **InspectionItem** – Individual checks within an inspection
- **Violation** – A failed or non‑compliant finding
- **CorrectiveAction** – Required remediation steps
- **ComplianceSnapshot** – Materialized compliance calculations
- **AuditEvent** – Immutable audit trail record

---

## Entity Relationship Diagram (Mermaid)

```mermaid
erDiagram
    PROJECT {
        uuid id PK
        string name
        string description
        string status
        datetime start_date
        datetime end_date
        datetime created_at
    }

    INSPECTION {
        uuid id PK
        uuid project_id FK
        string inspection_type
        string status
        datetime scheduled_at
        datetime performed_at
        datetime created_at
    }

    INSPECTION_ITEM {
        uuid id PK
        uuid inspection_id FK
        string code
        string description
        string result
        string notes
    }

    VIOLATION {
        uuid id PK
        uuid inspection_item_id FK
        string severity
        string status
        string code
        string description
        datetime identified_at
        datetime resolved_at
    }

    CORRECTIVE_ACTION {
        uuid id PK
        uuid violation_id FK
        string action_type
        string description
        string status
        datetime due_date
        datetime completed_at
    }

    COMPLIANCE_SNAPSHOT {
        uuid id PK
        uuid project_id FK
        decimal compliance_score
        datetime calculated_at
    }

    AUDIT_EVENT {
        uuid id PK
        string entity_type
        uuid entity_id
        string action
        string performed_by
        datetime performed_at
    }

    PROJECT ||--o{ INSPECTION : has
    INSPECTION ||--o{ INSPECTION_ITEM : contains
    INSPECTION_ITEM ||--o{ VIOLATION : results_in
    VIOLATION ||--o{ CORRECTIVE_ACTION : requires
    PROJECT ||--o{ COMPLIANCE_SNAPSHOT : generates
    PROJECT ||--o{ AUDIT_EVENT : audited
    INSPECTION ||--o{ AUDIT_EVENT : audited
    VIOLATION ||--o{ AUDIT_EVENT : audited
```

---

## Design Notes

### UUID Primary Keys
- All entities use UUIDs
- Enables cross‑stack parity and offline‑safe identifiers

### Explicit State Fields
- `status` fields are explicit, enumerable, and backend‑controlled
- No implicit state transitions

### Compliance as a Snapshot
- Compliance score is materialized, not derived live
- Enables auditability and historical analysis

### Audit Events
- Audit is append‑only
- No foreign‑key enforcement to allow schema evolution

---

## Out‑of‑Scope (Deliberate)

- User / Identity schema
- Roles / permissions
- Attachments / blobs
- Notification delivery
- Historical versioning tables

These may be layered later without disturbing the core model.

---

## Status

This ERD defines the **baseline domain data contract** for FieldMark and is suitable for:
- EF Core migration ownership
- Django model mapping
- Agentic schema reasoning
- Architecture reviews
