# Product Requirements Document (PRD)

## Product Name
**Construction Compliance & Inspection Management System (CCIMS)**

## Purpose
The Construction Compliance & Inspection Management System (CCIMS) is a web-based application designed to manage construction project compliance, inspections, violations, and corrective actions. The system demonstrates that a modern, highly interactive, enterprise-grade user experience can be achieved with a server-driven architecture using Razor Pages + HTMX, without the cognitive and architectural overhead of a full SPA.

The product will serve multiple purposes:
- A **realistic demo application** for architecture and UX exploration
- A **comparison baseline** for Angular, React, and Python (Flask/Django) implementations
- A **teaching tool** illustrating server-owned workflows, validation, and UI orchestration

---

## Goals & Success Criteria

### Primary Goals
- Demonstrate SPA-level reactivity using server-rendered HTML and HTMX
- Showcase integration of rich third-party JavaScript controls (AG Grid)
- Centralize domain logic, workflows, and state management on the backend
- Provide a non-trivial, realistic domain aligned with construction and electrical industries

### Success Metrics
- Minimal client-side state management logic
- No duplicated business rules between client and server
- Clear mapping between HTTP requests and user interactions
- Ability to compare implementation complexity across UI stacks

---

## Target Users

### Primary Personas
- **Project Manager** – monitors compliance status and inspections across projects
- **Compliance Officer / Inspector** – performs inspections and records violations
- **Site Supervisor** – resolves violations and submits corrective actions
- **Executive / Oversight Role** – views risk, compliance, and trends via dashboards

---

## Core Use Cases

1. View compliance status across projects
2. Schedule, perform, and record inspections
3. Capture violations and corrective actions
4. Track compliance changes over time
5. Enforce regulatory and process workflows
6. Provide real-time compliance dashboards

---

## Functional Scope

### Projects
- Create, view, and manage construction projects
- Assign job sites, trade types, and inspectors
- Maintain project lifecycle states (Active, On Hold, Closed)

### Inspections
- Schedule inspections per project and trade
- Record inspection outcomes (Pass, Fail, Conditional)
- Capture inspection metadata (date, inspector, notes)

### Violations
- Record violations discovered during inspections
- Categorize violations by code, severity, and trade
- Associate corrective actions and deadlines
- Track violation lifecycle (Open → In Progress → Resolved)

### Corrective Actions
- Submit remediation evidence (notes, attachments placeholder)
- Validate prerequisites for resolution
- Enforce rule-based completion requirements

### Compliance Rules Engine
- Server-side rule evaluation (e.g., required inspections, code requirements)
- Configurable thresholds and project scoring
- Automatic recalculation of compliance metrics

### Audit & History
- Immutable inspection and violation history
- Activity log per project showing state transitions

---

## Dashboard & Reporting (Key "Wow" Area)

### Compliance Dashboard
- Overall compliance score per project
- Number of open / overdue violations
- Inspections due and recently completed
- Risk indicators by trade or subsystem

### Interactive Behavior
- HTMX-driven partial updates for dashboard components
- Drill-down navigation without full-page reloads
- Lazy-loaded sections using `hx-trigger="revealed"`

---

## Data Grid Requirements (AG Grid)

AG Grid will be used for data-heavy views:
- Project inspection lists
- Violation listings
- Audit/history views

### Features
- Server-side row model
- Pagination, sorting, filtering
- Row selection triggering HTMX detail panel updates
- Minimal JavaScript configuration (no client-side business rules)

---

## Non-Functional Requirements

### Performance
- Fast first paint via server-rendered HTML
- Incremental updates via partial HTML swaps

### Scalability
- Stateless HTTP request handling
- Horizontal scaling compatible

### Security
- Role-based access control (RBAC)
- Server-enforced validation and transitions
- No trust in client-submitted state

### Accessibility
- Semantic HTML
- Keyboard-navigable workflows
- Graceful degradation when JavaScript is unavailable

---

## Technical Architecture (High-Level)

### Backend
- ASP.NET Core
- Razor Pages for UI
- HTMX for interactivity
- Entity Framework Core for persistence
- Minimal APIs for AG Grid JSON endpoints

### Frontend
- HTML over the wire
- HTMX attributes for interaction modeling
- AG Grid as a JS island

### Data Persistence
- Relational database
- EF Core aggregates and navigation
- Audit and history tables

---

## Comparison Implementations (Optional)

The same domain and workflows may be implemented using:
- Angular + AG Grid
- React + AG Grid
- Flask or Django + HTMX

These versions are intended for architectural comparison, not feature divergence.

---

## Out of Scope (Initial Phase)
- Mobile-native apps
- Offline-first behavior
- Real-time collaboration
- File uploads (may be stubbed)

---

## Future Enhancements
- Regulatory rule configuration UI
- Notification system (email/webhooks)
- GIS / site mapping integration
- Advanced analytics

---

## Epic Breakdown (Initial)

1. Project Setup & Domain Modeling
2. Compliance Rules Engine
3. Inspections & Violations Workflow
4. Dashboard & Reporting
5. AG Grid Integration
6. Security & Roles
7. Comparative UI Implementations

---

## Key Product Demonstration Messages

- The backend orchestrates workflows and state
- HTML can be reactive without SPAs
- Rich JS controls do not require client-owned applications
- Reduced cognitive load improves long-term maintainability

---

*This PRD is intended as a foundation for architecture design, UX exploration, and sprint/epic planning.*
