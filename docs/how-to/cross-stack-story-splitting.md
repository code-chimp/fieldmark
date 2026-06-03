# How-To: Cross-Stack Story Splitting (a/b/c)

**Status:** Ratified 2026-06-03 via Correct Course (`sprint-change-proposal-2026-06-03.md`).
**Audience:** Anyone authoring epics/stories (`bmad-create-epics-and-stories`, `bmad-create-story`) or implementing them (`bmad-dev-story`).

## Why this exists

FieldMark ships every feature across three parallel stacks (.NET, Django, Go). Through Epic 2, large cross-stack stories produced:

- **Context bloat** — one story loaded .NET + Django + Go source + 3 test suites at once.
- **Oversized diffs** — reviewers had to chunk a 3-stack diff, losing cross-cutting context.
- **Same-shape churn** — a bug fixed in stack #1 reappeared in stacks #2/#3 *within the same story* (Epic 1 retro, "Same-shape bugs repeating across stacks"). Stories 2-4 (7 review rounds), 2-9 (5+), and 2-10 (4) are the evidence.

The Epic 1 retro already prescribed the countermeasure as an in-story discipline ("canonical example first, then implement per stack"). This policy promotes it to a **story boundary** so each chunk is reviewer-sized, single-stack, and the shape is reviewed *once* in the reference before porting.

## The split model

A split story-group has a parent number `N.M` and three sub-stories:

| Sub-story | Scope | Gate to pass before "done" |
|---|---|---|
| **`N.Ma` — Reference (.NET)** | Establish the flow: routes, data shapes, manual projections, markup/templates, entity methods. Author or update the contract doc (`docs/reference/*` or `docs/how-to/*`). | `.NET` suite green; contract doc landed. |
| **`N.Mb` — Port (Django + Go)** | Idiomatic re-implementation in both remaining stacks against `N.Ma` + the contract doc. No new design decisions — divergence from the reference is a defect, not a choice. | `make test-django` + `make test-go` green. |
| **`N.Mc` — Parity & DoD** | The three-stack invariant gate. | `make parity` clean; per-stack cross-stack conformance tests; byte-identical snapshots; E2E (where applicable); NFR1 timing (local-dev p95 ≤ 200 ms, divergence ≤ 50 ms p95). |

### Definition of done & the hard rule

Root `CLAUDE.md` states *"a story is never done until all three stacks pass it."* For a split group, that invariant is **relocated to `N.Mc`**. `N.Ma` and `N.Mb` are legitimately marked `done` in a deliberately stack-divergent state. `make parity` will be **red** on `N.Ma`/`N.Mb` by design — CI and dev workflow must not treat that as failure; parity is asserted only at `N.Mc`.

### Dependency chain (non-negotiable)

`N.Ma → N.Mb → N.Mc` is a hard sequence. **Do not start `N.Mb` until `N.Ma` is reviewed clean.** Porting against an unreviewed reference reintroduces exactly the same-shape churn this policy exists to kill.

## When to split vs. stay unified

Split is gated on **type AND size** — both must hold.

### Split into a/b/c when the story is BOTH:

- **Behavioral / UI-integration** — involves an HTTP handler + route, markup/template rendering with a snapshot/byte-identical requirement, OOB orchestration, or AG Grid SSRM; **and**
- **Large** — multi-region paint, AG Grid, a complex/multi-field form, a multi-entity atomic transaction, or otherwise the kind of diff a reviewer would have to chunk.

### Stay unified when the story is ANY of:

- **Pure data-layer mapping** (EF config / Django model / Go store) whose only assertion is schema round-trip.
- **A small single transition** (~one entity method + one POST route + minimal markup, e.g. start/cancel).
- **A pure cross-stack-deterministic function** whose determinism test *is its own parity proof* (e.g. a scoring helper) — apply reference-first *inside* the single story instead.

When in doubt, ask: *"Would a reviewer have to chunk this diff?"* If no, keep it unified.

## Naming & tracking

- Sub-story IDs: `N.Ma`, `N.Mb`, `N.Mc` (keep the parent number for traceability to the original intent).
- `sprint-status.yaml`: one entry per sub-story, all starting `backlog`.
- The epic file keeps the parent story's narrative ("As a… I want… So that…") once, then lists the a/b/c scope + ACs under it.

## Worked example — "Complete Inspection with auto-open Violations" (Epic 3, Story 3.9)

- **3.9a (.NET reference)** — `inspection.complete()` + `Violation.open_from_finding()`, the single atomic transaction, score recompute, the three-region OOB response, and the `docs/how-to/three-region-oob-orchestration.md` reference. `make test-net` green.
- **3.9b (port)** — idiomatic Django + Go implementations against 3.9a and the contract doc. `make test-django` + `make test-go` green.
- **3.9c (parity & DoD)** — `make parity` clean, the per-stack three-region conformance test, byte-identical partial snapshots, the cross-stack E2E Playwright scenario, and NFR1 timing. The three-stack invariant is satisfied here.

## Scope of application

- **Epic 2:** closed under the old model; not retroactively re-split. Story 2-13 ships unified (dev already in progress).
- **Epic 3:** re-broken in the epic file now (this change).
- **Epics 4–6:** the policy is applied lazily at `create-story` time — do not re-break their epic files speculatively.
