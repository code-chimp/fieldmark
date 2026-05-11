# Success Criteria

## User Success

FieldMark has two distinct user populations and the criteria differ for each.

**Audience success — mid-to-senior engineers in the talk.**

- A measurable shift from "HTMX is fine for small things" to "HTMX is viable for our application" in post-talk feedback, surveys, or conversations.
- The "resolve a violation" anchor demo produces visible audience reaction: the moment compliance score, violation status, and audit log all update in one round trip is recognized as the differentiator without needing to be explained.
- At least one architectural objection a skeptic walks in with — duplicated rules, security via client validation, "AG Grid forces a SPA," "you can't do real interactivity without React" — gets answered by the demo running in front of them, not by argument.
- The repository is cloned and inspected by attendees after the talk; questions in follow-ups are about applying the pattern to their own stack, not about whether the pattern is real.

**In-app persona success — workflow completion as a credibility test.**

- Project Manager can navigate from compliance dashboard → project detail → violation → audit log without a full-page reload.
- Compliance Officer can complete an inspection, spawn violations from findings, and approve a corrective action with state transitions visibly enforced by the server.
- Site Supervisor can submit a corrective action and see status reflect immediately when a Compliance Officer reviews it.
- Each persona's primary workflow finishes in ≤ 3 HTMX exchanges where the brief's UX guide implies a single conceptual action.
- The system never renders a button for an action the user is not authorized to take, never accepts a state transition the domain rules forbid, and always returns the user to a coherent rendered state on rule violation (HTTP 409 with the originating partial).

## Business Success

Reframed for a teaching artifact: **Talk and reference-implementation success.**

- The talk is delivered with the live demo working without fallback to slides for any anchor workflow.
- The repository continues to receive engagement (clones, references, comparison forks) for at least 6 months post-talk as a usable reference implementation.
- A subsequent SPA implementation (Angular or React) of the same domain — built by the author or a contributor — produces a defensible architectural-cost comparison.
- Three-stack parity holds: a contributor familiar with one stack can reason about the others by analogy, not by re-learning patterns.

## Technical Success

The architectural thesis is operationalized as a set of falsifiable counts and constraints. These are non-negotiable; missing any of them invalidates the demonstration.

| Metric                                                                                    | Target                                                                      |
| ----------------------------------------------------------------------------------------- | --------------------------------------------------------------------------- |
| Hand-written client-side JavaScript files                                                 | ≤ 5, all narrowly scoped (AG Grid wiring, minimal UX glue)                  |
| Business rules duplicated between client and server                                       | 0                                                                           |
| Lines of state-management code (Redux/NgRx/Pinia equivalents)                             | 0                                                                           |
| HTTP requests directly traceable to a user interaction                                    | 100%                                                                        |
| Architectural delta across .NET, Django, Go implementations                               | Limited to language idioms and framework syntax; zero structural divergence |
| Routes / HTMX target IDs / AG Grid contracts / audit action strings / domain method names | Identical across all three stacks (modulo casing conventions)               |
| Domain rules, validation, transitions, authorization                                      | Server-side only; no client-side authority                                  |
| Audit entry per domain mutation                                                           | 100%, written in the same transaction as the change                         |
| Compliance score recomputation                                                            | Server-side, in the same transaction as triggering write                    |
| Database tests using SQLite or other Postgres substitutes                                 | 0 — real PostgreSQL only (Testcontainers / pytest-django)                   |
| Framework migrations touching the `domain` schema                                         | 0 — domain is infrastructure-owned (ADR-014)                                |
| Foreign keys from `domain.*` to any auth schema                                           | 0 (ADR-012)                                                                 |

## Measurable Outcomes

**Interaction smoothness — the SPA-equivalence target, made falsifiable.**

| Outcome                                                                                                         | Target                                              |
| --------------------------------------------------------------------------------------------------------------- | --------------------------------------------------- |
| Full-page reload on a state-changing action (resolve violation, complete inspection, approve corrective action) | Never                                               |
| HTMX partial-swap perceived latency on local dev (action → updated panel + tile + audit row)                    | ≤ 200 ms p95                                        |
| AG Grid row selection → detail panel rendered                                                                   | ≤ 300 ms p95                                        |
| Compliance score tile updates after a state transition affecting it                                             | Same round trip as the action; no follow-up request |
| HTMX out-of-band swaps documented at every use site                                                             | 100%                                                |

**Cross-stack symmetry — measured, not asserted.**

| Outcome                                               | Target                                                 |
| ----------------------------------------------------- | ------------------------------------------------------ |
| `pg_indexes` snapshot diff across the three stacks    | Zero differences                                       |
| Route inventory diff across the three stacks          | Zero differences (modulo language casing)              |
| Playwright E2E suite passing against all three stacks | 100% of scenarios                                      |
| Story-level parity violations merged to main          | 0 (a story is not done until all three stacks pass it) |
