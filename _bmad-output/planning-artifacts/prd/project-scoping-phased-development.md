# Project Scoping & Phased Development

The Product Scope section above defines _what_ ships in MVP, Growth, and Vision phases. This section defines the _strategy_ behind that phasing, the resource posture, and the risks that could compress or expand it.

## MVP Strategy & Philosophy

**MVP approach: Demonstration MVP.** FieldMark's MVP is not a problem-solving MVP (the domain is simulated), an experience MVP (delight is a constraint, not the goal), or a revenue MVP (there is no revenue). It is a **demonstration MVP**: the minimum scope at which the architectural thesis can be falsified or confirmed.

The implication: scope reductions are evaluated against "does the architectural argument still hold?" rather than "do users still get value?" A feature is in MVP if and only if removing it would either (a) break the anchor workflow ("resolve a violation" with three-thing update), (b) compromise the three-stack symmetry argument, or (c) make the architectural rules visibly dishonest (e.g., shipping without an audit log makes the "audit-on-every-mutation" claim hollow).

**Validation path: build first, then decide what the artifact is for.** The architectural argument is not assumed; it is being proven to the author first. The MVP exists so the author can verify the smoothness, the symmetry, and the integration quality firsthand before deciding whether — and how — to share it externally. A talk is one possible output; a public reference implementation is another; a private synthesis exercise that informs future work is a third. The artifact's value does not depend on the talk happening.

**Resource Requirements & Posture**

- **Team size:** one contributor (the author), augmented by AI agents (BMAD method, Claude Code).
- **Required skills:** .NET 10 / Razor Pages, Python 3.14+ / Django 6.x, Go 1.26+ / Fiber v3, PostgreSQL 17, HTMX 4.x, AG Grid Community 35.x, Tailwind CSS v4, Playwright. Author has working knowledge of all stacks; agents close the gaps.
- **Time budget:** elastic. Work happens in evening and weekend slices, around employment and family obligations. There is no external deadline. Throughput is also bounded by AI-assistance constraints — Claude Pro plan rate limits effectively cap the volume of agent-assisted work in any given window.
- **Pace expectation:** this is a project, not a sprint. Calendar time is not the constraint; _consistent forward motion across all three stacks_ is. A month with no progress is acceptable; a month where one stack pulls ahead and the others fall behind is not (it violates the parity discipline that makes the artifact's argument).
- **Infrastructure:** local Docker Compose (PostgreSQL 17). No production hosting required. Hosted demo environment is a Vision-phase question contingent on whether the artifact is shared externally.

## Risk Mitigation Strategy

**Technical Risks**

| Risk                                                                                  | Likelihood | Impact                                  | Mitigation                                                                                                                                                                                                            |
| ------------------------------------------------------------------------------------- | ---------- | --------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| HTMX cannot deliver the SPA-equivalent smoothness target on one or more workflows     | Medium     | High — invalidates thesis               | Build the anchor workflow (resolve a violation) end-to-end early in implementation, on at least one stack, and measure latency before committing to it as the artifact's anchor. Iterate or change anchors if needed. |
| AG Grid server-side row model has a hard incompatibility with one of the three stacks | Low        | High — breaks symmetry                  | Stand up a minimal AG Grid integration in all three stacks during scaffolding, before deep feature work. ADR amendment if a stack proves unviable.                                                                    |
| Cross-stack drift accumulates faster than parity gates can catch                      | Medium     | Medium — erodes the artifact's argument | Story-level parity gates (already established); E2E suite running against all three stacks; `pg_indexes` and route-inventory diff in CI.                                                                              |
| Tailwind v4 + AG Grid theming integration produces visual divergence across stacks    | Low        | Low — cosmetic                          | Single shared compiled CSS (`fieldmark_shared/dist/`) symlinked into all three apps; AG Grid theming is part of that compilation. Visual diff testing optional.                                                        |
| Auth implementations diverge on roles, blocking cross-stack scenario tests            | Medium     | Medium — blocks parity testing          | Conceptual roles fixed in PRD (Admin, Compliance Officer, Inspector, Site Supervisor, Executive); each stack maps natively per ADR-012; seed scripts use identical UUID values across stacks so audit trails compare. |

**Audience / Adoption Risks (conditional on external sharing)**

These risks apply only if the artifact is shared externally (talk, blog series, public reference implementation). They do not apply to the artifact's value as a personal synthesis exercise.

| Risk                                                                                                   | Likelihood | Impact | Mitigation                                                                                                                                                                                                          |
| ------------------------------------------------------------------------------------------------------ | ---------- | ------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| External audience finds the demo workflow contrived or unconvincing despite working architecture       | Medium     | Medium | Anchor on a workflow audiences recognize as enterprise-credible (violation resolution with three-thing update); rehearse with at least one external skeptic before any public delivery.                             |
| The "AG Grid is fine" position is rejected because pure-HATEOAS audience members find it too pragmatic | Low        | Medium | Pragmatic-islands section grounds the position in HTMX project's own documentation (json-enc, htmx.process, web components examples); not arguing against the project, arguing inside it.                           |
| Reference implementation gets cloned but no follow-on engagement                                       | Medium     | Low    | Engagement is a 6-month-out success metric, not a build criterion. Lack of engagement does not invalidate the architecture; it just affects reach.                                                                  |
| Three-stack symmetry argument is dismissed as "you only did this because you had time, no team would"  | Medium     | Medium | Reframe in any external presentation: three stacks exist to make the architectural argument _stack-independent_; a real team picks one. The point is variability isolation, not three-stack production deployments. |

**Resource & Continuity Risks**

These are the actual ongoing risks given the project's elastic timeline and solo-with-agents resourcing.

| Risk                                                                                                                       | Likelihood | Impact                                              | Mitigation                                                                                                                                                                                                                                                                                       |
| -------------------------------------------------------------------------------------------------------------------------- | ---------- | --------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Project stalls — life happens, weeks go by without progress, momentum is lost                                              | High       | Medium — natural consequence of a hobby project     | Accept it. Architectural state is documented well enough (this PRD, ADRs, domain model) that returning after a long break costs hours of re-orientation, not days. The PRD itself is part of the mitigation.                                                                                     |
| Stack divergence creeps in during long pauses — one stack is fresh in memory, another isn't                                | Medium     | Medium                                              | Always pause at a parity boundary, not mid-story. Resume by running the cross-stack diff (`pg_indexes`, routes, E2E) before writing new code.                                                                                                                                                    |
| Interest fades; the project becomes "a thing I started and didn't finish"                                                  | Medium     | Low — emotionally costly, not architecturally fatal | The artifact provides standalone value at multiple completion points: PRD alone (synthesis exercise), one stack working (architecture validated for the author), two stacks working (cross-stack symmetry argument viable), all three (full demonstration). Each is a legitimate stopping point. |
| Agent-assisted development produces architectural drift (CQRS sneaks in, repository pattern emerges, client state appears) | Medium     | High — invalidates thesis at code-review level      | PRD-binding architectural constraints are explicit and visible to agents. Code review treats violations as defects. Architectural smoke tests in CI where feasible (no `MediatR` package reference; no Redux-like imports).                                                                      |
| Claude Pro plan rate limits cap velocity below what the author wants in a given session                                    | High       | Low                                                 | Plan ahead — batch larger architectural questions into single sessions; use lower-cost models for routine edits. Rate-limited-out is a normal stopping point, not a setback.                                                                                                                     |
| External dependency (HTMX, AG Grid, Fiber, Django) ships a breaking change during a long pause                             | Low        | Medium — rework                                     | Versions pinned across all three stacks; upgrades are coordinated stories, not maintenance tasks.                                                                                                                                                                                                |

## Acceptable Stopping Points

This artifact has multiple legitimate completion points. Each is a real deliverable; none of them is failure if the project pauses there.

1. **PRD and planning artifacts complete (current target).** Synthesis is itself the deliverable. The author has converted a long-running set of intuitions into a coherent, examinable document.
2. **One stack reaches MVP, validated against the architectural rules.** The architectural argument is proven _for the author_; the smoothness target is verified.
3. **Two stacks reach MVP with parity discipline holding.** The three-stack symmetry argument is viable — variability is isolated, framework is the variable, architecture is the constant. Talk material exists if the author wants it.
4. **All three stacks reach MVP.** The full demonstration is in hand.
5. **Growth and Vision phases.** Continued investment, not required for the architectural argument to land.

The "scope compression order" below applies _if and when_ the author chooses to draw a line and ship something externally — not as a forced response to a deadline that doesn't exist.

## Scope Compression Order (If/When External Sharing Becomes the Goal)

If the artifact is brought to a public form and choices need to be made about what to include:

1. **Growth-phase items first** — admin UI for reference data, parity test suite, executive trend dashboard, configurable severity weights.
2. **Third-stack polish** — a public artifact with two stacks fully demonstrated and one stack documented but not feature-complete is intellectually honest and architecturally sufficient.
3. **Visual polish on non-anchor screens** — admin pages, list views that aren't on the demo path, edge-case error screens.
4. **Vision-phase items** — already out of MVP; mentioned for completeness.

The anchor workflow (compliance dashboard → project detail → resolve a violation, with all three things updating in one round trip) is not negotiable for any external presentation. If that doesn't work, what's being shown isn't the architecture being argued for.

## Epic Shaping Principles

Because this project does not run on time-boxed sprints, the natural unit of work is not "what fits in two weeks" but **"what produces something runnable that teaches us something."** Epics are sized by learning milestones, not calendar.

**An epic in FieldMark must:**

1. **Land in all three stacks before being called done.** The existing parity rule — a story (and by extension, an epic) is not done until all three stacks pass it.
2. **Produce something runnable end-to-end at the boundary.** Not a half-built layer waiting for the next epic. If you stopped at epic close, the artifact still works as far as it goes.
3. **Ship its tests with it.** E2E coverage (Playwright, against all three stacks) and unit coverage (per stack, idiomatic tools) for the behavior the epic adds. Tests are not a follow-on epic. There is no "we'll add tests later" backlog item — that backlog grows silently and is never serviced. Quality assurance evolves alongside the code, the way a traditional team would do it with a dedicated QA pairing on every story.
4. **Validate or invalidate something architecturally meaningful.** An epic that only adds features without answering an architectural question is doing horizontal layer work and should be re-shaped. The Anchor Workflow MVP epic, for example, exists specifically to falsify or confirm the smoothness target _early_, before deeper feature investment.
5. **Be pause-safe.** End at a state where stopping for a month doesn't leave broken builds, half-applied schema migrations, or one stack ahead of the others. Always pause at a parity boundary.

**Test-discipline corollary.** The "no E2E coverage" / "tests deferred" anti-pattern is forbidden. An epic that adds a workflow without the corresponding Playwright scenario covering it across all three stacks is not done. An epic that adds a domain method without unit tests proving its invariants is not done. This is the same discipline that would be enforced by a dedicated QA team in a traditional environment; in this project it's enforced by the epic definition itself.

**What this means for ordering.** Architectural-validation epics come early — Walking Skeleton, AG Grid Integration, Anchor Workflow MVP — because they answer the questions whose answers steer everything that follows. Pure feature-extension epics (additional violation states, more report views) come later, after the architectural floor is verified.

**What this does _not_ mean.** Epics are still allowed to be small. The Walking Skeleton epic is shorter than the Anchor Workflow epic. Sizing follows what it takes to produce a runnable, tested, parity-locked deliverable — not a fixed quantum. Some epics will be a weekend; some will span months of intermittent work. Both are fine as long as the epic boundary is honest.
