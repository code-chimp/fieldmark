# Product Scope

## MVP — Minimum Viable Product

These are the features the talk depends on. Everything here is in scope across all three stacks.

- **Project lifecycle** — create, view, list, place on hold, resume, close (closure gated by compliance rules).
- **Inspection workflow** — schedule, start, complete with findings (Pass/Fail/Conditional), cancel.
- **Violation lifecycle** — auto-spawned from Fail-class findings; assign to supervisor; track to resolution or void.
- **Corrective Action workflow** — submit by Site Supervisor, take for review, approve or reject by Compliance Officer; only Compliance Officer–approved actions resolve a violation.
- **Compliance Rules Engine** — server-side rule evaluation, configurable thresholds, automatic recalculation of compliance score on every relevant state transition.
- **Compliance Dashboard** — portfolio-level overview with HTMX partial refresh and drill-down to project detail.
- **Project Detail (anchor screen)** — current truth: status, compliance score, inspections, violations, audit log, all in one rendered view; tabs are HTMX swaps, not client routes.
- **AG Grid integration on at least two views** — server-side row model only, row selection drives HTMX detail panel updates.
- **Per-project audit log** — immutable, append-only, written in the same transaction as the triggering domain mutation.
- **Role-based access control** — covering all four primary personas (PM, Compliance Officer, Inspector, Site Supervisor); implemented framework-locally per ADR-012.
- **Infrastructure-owned `domain` schema** — created by Docker init scripts; EF Core / Django ORM / Go SQL all map to it without owning it.
- **Cross-stack parity** — every story passes in all three stacks before it is considered done.

## Growth Features (Post-MVP)

In scope only if time permits, or as follow-on after the talk.

- Reference-data administration UI (TradeType, ViolationCategory, ComplianceRule). Django Admin handles this for free; .NET and Go get minimal Razor / Fiber admin pages matched to capability, not polish.
- Multi-stack parity test suite running identical scenarios across all three implementations and producing a parity report.
- Executive / Oversight read-only dashboard with trend visualizations.
- Configurable severity weights and due-offset overrides via ComplianceRule.parameters (the data model supports it; UI does not yet).

## Vision (Future)

Post-talk, post-reference-implementation. Direction, not commitment.

- Notification system (email, webhooks, in-app).
- Real file uploads for corrective action evidence (current implementation is placeholder).
- Regulatory rule configuration UI for non-developer admins.
- Time-series compliance score history (currently the audit log is the history).
- GIS / site mapping for JobSite visualization.
- Mobile-native or PWA layer (current scope is desktop-first responsive).
