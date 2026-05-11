# Non-Functional Requirements

Several quality attributes are already locked in earlier sections with specific targets and are referenced here rather than duplicated. The categories below cover what those sections don't, and explicitly mark categories that do not apply.

## Performance

Locked in §Success Criteria → Measurable Outcomes. Summary:

- HTMX partial-swap perceived latency ≤ 200 ms p95 (local dev, including the action → updated panel + tile + audit row round trip).
- AG Grid row selection → detail panel ≤ 300 ms p95.
- No full-page reload on any state-changing action.
- Cross-stack latency divergence > 50 ms p95 on the same scenario is a defect.
- Server-side compliance score recomputation occurs in the same transaction as the triggering write; no follow-up request is acceptable.

Performance is measured on local development hardware; production-grade latency targets are out of scope for an artifact that does not have a production deployment.

## Security

- Server-side authority for all domain rules, validation, and authorization (per §Architectural Constraints — Backend Authority).
- Authentication is framework-local per ADR-012; each stack uses its native authentication system. No shared identity backend, no cross-stack SSO, no third-party identity provider.
- Authorization checks run on every request that mutates state; UI-rendered absence of an action button is _not_ the authorization mechanism, only its observable surface.
- All client-submitted data is treated as untrusted; any field, parameter, or header may be tampered with via DevTools, an intercepting proxy (Burp Suite, mitmproxy), or a replay tool (Postman, Insomnia). Server-side validation is the only enforcement layer.
- Passwords (where applicable per stack) are stored using the framework's native salted hashing; plaintext storage is forbidden.
- CSRF protection is enabled per stack-native conventions for state-changing requests.
- SQL queries use parameterized statements or ORM-managed parameter binding; string-concatenated SQL is forbidden.
- The audit trail (FR39–FR43) provides forensic visibility for any domain mutation: who did what, to what, when, with full before/after state.

Security is sufficient for a credible enterprise-shaped demonstration; FieldMark does not target formal certification (SOC 2, ISO 27001, FedRAMP) and produces no compliance documentation.

## Accessibility

Locked in §Web App Specific Requirements — Accessibility, with capabilities expressed as FR60–FR64. Summary:

- Target: WCAG 2.1 Level AA conformance across all three stacks.
- Enforcement: `@axe-core/playwright` accessibility scans embedded in every E2E scenario; runtime scanning catches the common defects that no static linter exists for in Razor / Django templates / Go `html/template`.
- HTMX-specific concerns explicitly committed to: focus management on swaps, `aria-live` on out-of-band swap targets, `hx-disabled-elt` for request-pending state, error association via `aria-describedby`.

## Reliability & Availability

This is a local-development teaching artifact. There is no production deployment, no uptime SLA, and no disaster-recovery posture. The honest bar:

- The application starts cleanly from a fresh `docker compose up -d` followed by stack-specific run commands documented in each stack's `README.md`.
- The application recovers from a database restart without data loss (transactions are durable) and without requiring application restart beyond standard connection-pool reconnection.
- The application does not silently corrupt data on transaction abort; failed mutations leave the database in the pre-mutation state.
- The application does not require external services beyond PostgreSQL; no message broker, no cache, no email service is needed for MVP.

Hosted demo availability, monitoring, alerting, and incident response are explicit non-goals (see Non-Goals below).

## Data Integrity

- Every domain mutation occurs within a database transaction that includes the corresponding audit-entry write and any compliance-score recomputation (FR39, FR57).
- The `domain` schema enforces structural invariants via CHECK constraints (status enums, score range 0–100, severity ranges, completion-implies-outcome, voided-implies-reason, etc.) per the canonical DDL in `domain-model.md` §8. CHECK constraints are defense-in-depth; primary enforcement remains on the entity methods.
- Audit entries are append-only; no UI or API path supports update or delete (FR41). In production deployments (out of scope for MVP), append-only would be enforced at the database privilege level via revoked UPDATE/DELETE on `domain.audit_entry`.
- UUIDs are generated in application code, not via `gen_random_uuid()`, so behavior is identical across stacks.
- All timestamps are stored as `TIMESTAMPTZ` (UTC); local rendering is a presentation-layer concern.

## Maintainability & Readability

This is a teaching artifact. Code readability is a first-class quality attribute, not a soft preference.

- Architectural simplicity is enforced by the §Architectural Constraints section: no CQRS, no repositories, no mediator patterns, no anemic domain models, no fat service layers, no client-side state stores. Removing accidental complexity is the architectural product, not just a code-style preference.
- Each stack uses idiomatic patterns for its ecosystem; the cross-stack rule is _structural symmetry_, not _naming-convention identity_. C# is `PascalCase`, Python is `snake_case`, Go is `PascalCase` exported / `camelCase` unexported, all mapping to the canonical `snake_case` database and JSON wire format per `domain-model.md` §9.
- Domain methods follow the canonical name list (`start`, `complete`, `cancel`, `place_on_hold`, `resume`, `close`, `assign`, `submit_corrective_action`, `approve_resolution`, `reject_resolution`, `void`); divergence from this list across stacks is a defect.
- Comments in code explain _why_, not _what_. Code that requires explanatory comments to understand its mechanism should be rewritten until it doesn't.
- Documentation lives in this PRD, the ADRs, the domain model, the stack reference docs, and per-stack `CLAUDE.md` / `README.md` files. There is no separate wiki or external knowledge base.
- Readability is enforced, not asserted: each stack runs its idiomatic auto-formatter and linter as part of its standard build (`dotnet format` + analyzers for .NET, `ruff` and `black` for Django, `gofmt` and `golangci-lint` for Go). Lint or format violations are build-blocking. PR review treats unresolved comments, dead code, and architectural-rule violations as defects, not stylistic notes; each stack documents its specific enforcement configuration in its `CLAUDE.md`.

## Portability & Cross-Stack Compatibility

- The `domain` PostgreSQL schema is the contract. EF Core, Django ORM, and Go data access map to it; none owns it.
- HTMX 4.x and AG Grid Community 35.x versions are pinned identically across all three stacks. Mismatch is build-blocking.
- Compiled Tailwind CSS (`fieldmark_shared/dist/fieldmark.css`) is symlinked into all three apps; CSS is authored once.
- A `pg_indexes` snapshot diff across the three stacks must produce zero differences.
- Route inventory diff across the three stacks must produce zero differences (modulo language casing).

## Observability

For a local-development teaching artifact, observability is intentionally minimal and centers on what's load-bearing for the architectural argument:

- **Audit log** (FR39–FR43) is the primary observability mechanism for _domain_ events. It is the system of record for "what changed and why."
- Standard framework-provided HTTP request logging is sufficient for _application_ events. No structured logging library, no log shipping, no log aggregation is required.
- No metrics endpoint, no `/healthz`, no Prometheus exposition, no tracing instrumentation is required for MVP.
- Error reporting in development is via the framework's default error page or stderr; no Sentry or equivalent is wired in.

Production-grade observability is an explicit non-goal. If FieldMark is ever deployed beyond local development, observability becomes a new ADR and is a Vision-phase concern.

## Internationalization & Localization

**Not applicable.** All UI strings, error messages, and audit action strings are in English. Date and number formatting is locale-default per browser. Translation infrastructure (gettext, .resx, ICU MessageFormat) is explicitly out of scope. If a future contributor wants to add i18n, it is a new ADR and a new story sequence.

## Browser Compatibility

Locked in §Web App Specific Requirements — Browser Support Matrix. Summary:

- Last 2 stable versions of evergreen browsers (Chrome, Firefox, Safari, Edge).
- Internet Explorer not supported.
- Mobile browsers best-effort, not a primary target.
- A feature working in Chromium but not in Safari is a defect, not an acceptable tradeoff.

## Non-Goals (Explicit Exclusions)

These are explicitly _not_ requirements. Listing them prevents later requirement creep and makes the artifact's scope honest.

- **Scalability:** No multi-tenancy, no horizontal scaling design, no load-test SLAs, no capacity planning. The artifact is single-instance, single-database, demo-scale.
- **High availability:** No uptime SLA, no failover, no replication topology, no health-check endpoints, no graceful-degradation modes.
- **Production hosting:** No CI/CD pipeline beyond what runs the test suites, no infrastructure-as-code, no cloud deployment, no production secrets management.
- **Compliance certification:** No SOC 2, no ISO 27001, no FedRAMP, no HIPAA, no PCI-DSS. The "compliance" in the domain name refers to construction compliance (the simulated business domain), not regulatory compliance of the application itself.
- **Production-grade observability:** No structured logging, no log shipping, no metrics, no distributed tracing, no APM.
- **Internationalization:** No i18n infrastructure, no translation, no RTL layout support.
- **Mobile-native applications:** No iOS app, no Android app, no React Native, no PWA install prompt.
- **Offline-first behavior:** No service worker, no IndexedDB persistence, no offline queue.
- **Real-time multi-user collaboration:** No WebSocket-driven shared editing, no presence indicators, no SSE-driven push.
- **File uploads:** No real file storage; `evidence_ref` on Corrective Actions is a string placeholder.
- **Notification system:** No email, no SMS, no in-app push, no webhook delivery.
- **Search:** No full-text search, no Elasticsearch, no PostgreSQL `tsvector` indexing. Filtering and sorting in AG Grid is the only "find data" capability.
- **Reporting / data export:** No PDF export, no CSV export, no scheduled report generation.
- **Audit log retention policy enforcement:** Audit entries are append-only; no automated purge, no cold storage, no retention windows.
