# Persistence Schema (Domain Layer)

This document describes the canonical database schema for the `domain` schema in PostgreSQL. The schema is owned by infrastructure SQL init scripts (`docker/postgres/init/`), not by any application framework.

## Schema Ownership

| Schema       | Owner                          | Purpose                          |
|--------------|--------------------------------|----------------------------------|
| `domain`     | Infrastructure SQL             | Business entities and aggregates |
| `django_auth`, `dotnet_auth`, `fiber_auth` | Framework-local migrations | Authentication and authorization |
| `infra`      | Infrastructure                 | Shared infrastructure objects    |

**Key rules:**
- No foreign keys from `domain.*` tables into any `*_auth` schema.
- User references are stored as opaque `uuid` values only.
- All tables use `snake_case` naming.

## Entity Relationship Diagram

```mermaid
erDiagram
    PROJECT ||--o{ INSPECTION : "has many"
    INSPECTION ||--o{ VIOLATION : "detects"
    VIOLATION ||--o{ CORRECTIVE_ACTION : "requires"
    PROJECT ||--o{ PROJECT_TRADE_SCOPE : "scopes"
    PROJECT ||--o{ PROJECT_INSPECTOR : "assigns"

    PROJECT {
        uuid id PK
        string name
        string code
        string status
        int compliance_score
        timestamp created_at
        timestamp updated_at
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
        timestamp due_at
    }

    CORRECTIVE_ACTION {
        uuid id PK
        uuid violation_id FK
        string description
        string status
        timestamp submitted_at
        timestamp resolved_at
    }

    AUDIT_ENTRY {
        uuid id PK
        string action
        uuid actor_id
        jsonb before_state
        jsonb after_state
        timestamp created_at
    }

    PROJECT_TRADE_SCOPE {
        uuid project_id PK,FK
        uuid trade_type_id PK,FK
    }

    PROJECT_INSPECTOR {
        uuid project_id PK,FK
        uuid inspector_id PK,FK
    }

    TRADE_TYPE {
        uuid id PK
        string name
    }

    VIOLATION_CATEGORY {
        uuid id PK
        string name
    }

    COMPLIANCE_RULE {
        uuid id PK
        string name
        jsonb parameters
    }
```

## Core Tables Summary

### Primary Aggregates
- `domain.project`
- `domain.inspection`
- `domain.violation`
- `domain.corrective_action`

### Supporting Tables
- `domain.audit_entry` — append-only audit trail (written in same transaction as every mutation)
- `domain.trade_type`, `domain.violation_category`, `domain.compliance_rule` — reference data
- Junction tables: `domain.project_trade_scope`, `domain.project_inspector`

## Naming & Type Conventions

- **Tables & columns**: `snake_case`
- **Primary keys**: `uuid` (generated in application code)
- **Timestamps**: `timestamptz`
- **Enums**: stored as `varchar` + `CHECK` constraints (not native PostgreSQL enums)
- **JSON columns**: `jsonb` for flexible state snapshots in `audit_entry`

## Invariants Enforced at the Database Level

- `compliance_score` on `Project` is recomputed after relevant state changes.
- Every mutating operation writes exactly one `AuditEntry` row.
- A `Violation` must have at least one `CorrectiveAction` before it can be resolved.

## Related Documentation

- [Domain Model](domain-model.md) — State machines and business invariants
- [Architecture](architecture.md) — Request flow and schema ownership rules
- [Hard Rules](hard-rules.md) — Non-negotiable persistence constraints

See the init scripts in `docker/postgres/init/` for the authoritative DDL.
