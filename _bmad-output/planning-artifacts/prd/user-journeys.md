# User Journeys

## Journey 1 — Marisol, Compliance Officer (the anchor demo, happy path)

**Persona.** Marisol is a Compliance Officer responsible for inspection oversight across a portfolio of active projects. Pre-FieldMark, her workflow is a mix of email, spreadsheet trackers, and PDF inspection forms scanned into shared drives. State — what's open, what's overdue, what's been resolved — is nowhere authoritative; she rebuilds it weekly from inboxes.

**Opening scene.** Monday morning. She lands on the Compliance Dashboard. One tile: portfolio compliance score, currently 78%. Below it: a project list grid showing each project's status, score, and open-violation count. She filters to projects she owns, sorts by score ascending, and the worst is _Riverside Substation Upgrade_ at 62%.

**Rising action.** She clicks the project row. The detail panel swaps in via HTMX — no full-page reload. She lands on the Project Detail anchor screen: status (Active), score (62%), tabs for Summary / Inspections / Violations / Audit. She clicks Violations. A list renders, severity-tagged, with one Critical item flagged overdue. She opens it.

**Climax.** The violation detail shows a single corrective action submitted by the Site Supervisor two days ago, status `UnderReview` (she took it for review yesterday). Evidence text and notes are present. She reviews, decides it's adequate, and clicks **Approve**. In one round trip:

- Violation status flips Open/InProgress → Resolved.
- The compliance score tile (out-of-band swap in the page header) jumps from 62% → 71%.
- A new audit row appears in the Audit tab: `CorrectiveActionApproved` by her, with timestamp and the `before_state` / `after_state` JSON.

She didn't navigate. Nothing reloaded. The state she sees is the state the database now holds.

**Resolution.** She moves to the next worst-scoring project. Her workflow is now a single screen per project; truth and action are in the same view. The architecture has receded behind the work.

**Capabilities revealed.** Compliance dashboard with portfolio aggregation; project list grid (AG Grid); project detail anchor screen with HTMX tab swaps; violation detail with action buttons gated by role and state; corrective action review workflow; out-of-band compliance-score updates; transactional audit logging.

---

## Journey 2 — Pat, Site Supervisor (edge case — rejection and resubmission)

**Persona.** Pat runs the electrical crew at Riverside Substation Upgrade. Pat sees violations assigned to them and submits remediation work as corrective actions. Pat is not authorized to resolve violations themselves — that's the Compliance Officer's job.

**Opening scene.** He gets a notification (out of MVP scope, but he's been told verbally) that a Critical violation has been opened against his crew's work: `RULE_GROUNDING_REQUIRED` on a panel installed yesterday. He logs in, navigates to his assigned violations, and opens the one in question.

**Rising action.** The violation detail screen shows Open status, a 2-day due window (Critical), and a **Submit Corrective Action** button. Pat clicks it, fills in the description ("Re-grounded panel per spec; verified continuity to bus bar; photo attached as reference"), and submits. The page swaps to show the submission with status `Submitted`. The violation has moved Open → InProgress; the assigned-to field shows Pat.

**Climax — the rejection.** Marisol takes the action for review (status → `UnderReview`), reads it, and rejects it: the description references a photo, but evidence_ref is empty in MVP, and her review notes say "Need explicit measurement of ground resistance, not just continuity check." Pat refreshes — or, more accurately, the screen polls / Pat navigates back — and sees:

- Corrective action status: `Rejected`.
- Review notes from Marisol visible.
- Violation status: still `InProgress` (rejection does not revert it to Open).
- A new audit row: `CorrectiveActionRejected`.
- The **Submit Corrective Action** button is _back_. He can submit a new one.

**Resolution.** Pat runs the resistance test, submits a second corrective action with the measurement data in the description. This one passes review. The violation resolves. The next screen shows it as terminal — there is no reopen path.

**Capabilities revealed.** Role-gated action buttons (server decides what renders); submit/review/reject corrective action workflow; rejection-does-not-revert-violation invariant; audit log captures both rejection and resubmission; multiple corrective actions per violation with only the latest non-Rejected eligible for approval.

---

## Journey 3 — Aisha, Project Manager (portfolio drill-down, closure attempt)

**Persona.** Aisha owns a portfolio of six active projects. She doesn't perform inspections or resolve violations herself — she monitors compliance status and manages project lifecycle (start, hold, close).

**Opening scene.** Her milestone target for _Riverside Substation Upgrade_ is end of month. Compliance score is now 71% (post-Journey 1). She wants to close the project. She opens the project detail and clicks **Close Project**.

**Rising action — closure denial.** The server evaluates the closure gate: one trade scope (Plumbing) has zero Completed inspections with outcome ∈ {Pass, Conditional}. The request returns HTTP 409. The Project Detail panel re-renders with the closure rejected, an explanation visible inline ("Cannot close: required inspection missing for Plumbing"), and the project status unchanged. The Close button is now visibly disabled, with tooltip explaining why. No client-side validation was involved — the server made the call and rendered the result.

**Climax.** She schedules the missing Plumbing inspection from the Inspections tab. A few days later, the inspection is performed, completed with outcome Pass, and she returns to the project. Now the Close button is enabled (server decision, embedded in the rendered HTML). She clicks it. The project transitions Active → Closed. The audit log records `ProjectClosed`. The project disappears from her active-portfolio view.

**Resolution.** She never had to learn what closure rules existed. The system told her, in the rendered UI, what was missing and when she could proceed. The closure rule is server code; the UI is its projection.

**Capabilities revealed.** Closure gate evaluation as part of `can_close`; HTTP 409 + originating partial as the standard rule-violation response; absent vs. disabled vs. present buttons as a server decision; inspection scheduling and lifecycle; audit entry for project lifecycle transitions.

---

## Journey 4 — Kenji, Executive Oversight (read-only)

**Persona.** Kenji is a VP responsible for portfolio-level risk. He doesn't act on violations or schedule inspections. He wants a single screen that tells him "where is the risk, and is it getting worse."

**Opening scene.** He logs in. The Compliance Dashboard renders portfolio-level aggregates: average compliance score, count of overdue violations by severity, count of projects in each lifecycle state.

**Rising action.** He filters by project status (Active only) and sorts the project grid by score ascending. The bottom three projects — the ones likely to need his attention — are surfaced. He clicks one to drill into Project Detail. Every action button on that screen is absent for his role; the page is information, not affordance.

**Resolution.** He has no path to mutate state. The audit log on the project detail is visible to him — read-only proves nothing has been hidden. He closes the tab.

**Capabilities revealed.** Read-only role rendering — no action buttons present at all for Executive; dashboard aggregations across the portfolio; cross-project filtering and sorting in AG Grid; full audit visibility without write capability.

---

## Journey 5 — The Talk Audience (the meta-persona)

**Persona.** A skeptical mid-to-senior engineer. They've built React or Angular apps for years. They've encountered HTMX in smaller-scale demos and aren't yet convinced the architecture holds up under enterprise complexity. They are at this talk because someone they respect told them to come.

**Opening scene.** Slide one. The thesis. They're not buying it. They've heard versions of "the SPA isn't necessary" argument before — Stimulus, Turbo, Phoenix LiveView — and remain unconvinced until a non-trivial application makes the case in front of them.

**Rising action.** The presenter switches to FieldMark. They see the dashboard. They see the project detail. So far it's nothing they couldn't have done with a SPA. Maybe faster, who knows. Then the presenter clicks **Approve** on a corrective action. Three things update in one visible round trip — status, score tile, audit row. No spinner. No flicker. They lean forward slightly.

**Climax — the architecture reveal.** The presenter opens DevTools. There is no JavaScript orchestrating that update. Just the HTMX swap. The presenter then opens the .NET, Django, and Go source side-by-side and shows the same handler shape, the same HTMX target IDs, the same audit row written in the same transaction. Three stacks. One architecture. Zero client state.

**Resolution.** They don't convert in the room. They ask a question about AG Grid (because of course they do — they use Kendo at work). The presenter shows the JSON endpoint, the row-select-fires-HTMX detail load, and walks through how the grid is an island, not a wedge. The audience member nods, takes a photo of the repository URL, and clones it that night to see if it holds up to their own scrutiny.

**Capabilities revealed.** None new in this journey — but every capability shown above is now load-bearing on the talk landing.

---

## Reference-Data Admin (out of primary journey scope)

The Reference-Data Admin manages TradeType, ViolationCategory, and ComplianceRule records. Per the UX guide, admin UX is a platform concern, not a product experience: Django uses Django Admin; .NET and Go provide minimal Razor / Fiber pages matched to capability, not polish. No narrative journey is mapped because no demo time is allocated to admin workflows. The capability exists; it is not on the talk path.

## Journey Requirements Summary

| Journey                                               | Primary capability cluster                                                                                                        |
| ----------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------- |
| 1 — Marisol, Compliance Officer (anchor demo)         | Dashboard, Project Detail, violation detail, corrective action approval, OOB compliance score swap, audit log                     |
| 2 — Pat, Site Supervisor (rejection / resubmission) | Role-gated action buttons, corrective action submit/review/reject, rejection-doesn't-revert invariant, multi-action per violation |
| 3 — Aisha, Project Manager (closure denial)           | Closure gate evaluation, HTTP 409 with originating partial, server-decided button states, inspection lifecycle                    |
| 4 — Kenji, Executive (read-only)                      | Role-based rendering with action buttons absent, portfolio aggregation, full audit visibility without write                       |
| 5 — Talk Audience (meta)                              | The architectural reveal — interaction smoothness, three-stack symmetry, AG Grid as island                                        |
| Reference-Data Admin                                  | Reference-data CRUD; minimal UX; not on demo path                                                                                 |

These journeys collectively reveal the full MVP capability set defined in §Product Scope. No new capabilities are introduced here that aren't already locked in scope.
