# FieldMark Domain Model

This document describes the core business entities, their relationships, and state machines.

## Aggregates Overview

```mermaid
erDiagram
    PROJECT ||--o{ INSPECTION : "has"
    INSPECTION ||--o{ VIOLATION : "detects"
    VIOLATION ||--o{ CORRECTIVE_ACTION : "requires"

    PROJECT {
        uuid id PK
        string name
        string status
        int compliance_score
        timestamp created_at
    }

    INSPECTION {
        uuid id PK
        uuid project_id FK
        string status
        timestamp scheduled_at
        timestamp started_at
        timestamp completed_at
    }

    VIOLATION {
        uuid id PK
        uuid inspection_id FK
        string code
        string description
        string status
        uuid assigned_to
    }

    CORRECTIVE_ACTION {
        uuid id PK
        uuid violation_id FK
        string description
        string status
        timestamp submitted_at
        timestamp resolved_at
    }
```

## Project State Machine

```mermaid
stateDiagram-v2
    [*] --> Draft
    Draft --> Active : start
    Active --> OnHold : place_on_hold
    OnHold --> Active : resume
    Active --> Closed : close
    Closed --> [*]
    Active --> Completed : complete
    Completed --> Closed
```

**Valid transitions**: `start`, `complete`, `cancel`, `place_on_hold`, `resume`, `close`

## Inspection State Machine

```mermaid
stateDiagram-v2
    [*] --> Scheduled
    Scheduled --> InProgress : start
    InProgress --> Completed : complete
    InProgress --> Cancelled : cancel
    Completed --> [*]
    Cancelled --> [*]
```

## Violation + Corrective Action Flow

```mermaid
stateDiagram-v2
    [*] --> Open
    Open --> Assigned : assign
    Assigned --> InProgress : start_work
    InProgress --> Submitted : submit_corrective_action
    Submitted --> UnderReview : take_for_review
    UnderReview --> Approved : approve_resolution
    UnderReview --> Rejected : reject_resolution
    Rejected --> InProgress
    Approved --> Closed : close
    Closed --> Voided : void
    Voided --> [*]
```

## Key Invariants

- A `Violation` must have at least one `CorrectiveAction` before resolution.
- `compliance_score` on `Project` is recomputed after any state change affecting open violations.
- All state transitions append an `AuditEntry` in the same transaction.
- User references are opaque IDs (no FKs from `domain` to auth schemas).

See `domain/` tables in PostgreSQL for exact column definitions.
