# Sprint Change Proposal — Cross-Stack Story Splitting (a/b/c)

- **Date:** 2026-06-03
- **Author:** John (Product Manager)
- **Requested by:** Tim
- **Trigger:** Mid-Epic-2 process friction (context bloat, oversized diffs, review churn)
- **Scope classification:** **Moderate** (backlog reorganization + process policy; no PRD/MVP change)
- **Mode:** Incremental

---

## Section 1 — Issue Summary

After implementing most of Epic 2, a recurring delivery-process friction was identified: building each feature in all three stacks (.NET, Django, Go) **simultaneously within one story** drives:

1. **Context bloat** — a single story session loads .NET + Django + Go source plus three test suites.
2. **Oversized commit diffs** — reviewers must chunk a 3-stack diff, which itself loses cross-cutting context.
3. **Same-shape review churn** — a bug fixed in stack #1 reappears in stacks #2/#3 *within the same story*.

**Evidence (Epic 2 review history, `sprint-status.yaml`):**

| Story | Review intensity |
|---|---|
| 2-4 (Phase-2 components) | **7 rounds** |
| 2-9 (AG Grid SSRM) | **5+ reruns** |
| 2-10 (Compliance dashboard) | **4 reruns** |
| 2-8 / 2-11 / 2-12 | 3–5 rounds each |

The Epic 1 retrospective (2026-05-25) already named the root cause — *"Same-shape bugs repeating across stacks… Reviewer typically catches stack #2 after stack #1 is fixed"* — and prescribed the countermeasure as an in-story discipline (team agreement #4: *"Canonical example first, then implement per stack"*). This proposal promotes that discipline from an in-story convention to a **story boundary**.

---

## Section 2 — Impact Analysis

**Epic Impact**
- **Epic 2:** Effectively closed under the old model (12/14 `done`, 2-12 in `review`, 2-13 `backlog` with dev in progress). **Not retroactively re-split** — zero value, pure churn. 2-13 ships unified.
- **Epic 3:** Re-broken now (the imminent epic) as the worked proof of the new model.
- **Epics 4–6:** Policy applied **lazily** at `create-story` time. Epic files not re-broken now (speculative; they will drift before reached).

**Story Impact (Epic 3)** — type+size gate applied:

| Story | Disposition | Rationale |
|---|---|---|
| 3.1, 3.2 | Unified | Pure data-layer mapping |
| 3.3 | Unified | Pure deterministic function; its cross-stack determinism test is its own parity proof |
| **3.4** | Split a/b/c | Inline form + OOB orchestration |
| **3.5** | Split a/b/c | AG Grid SSRM (the 2-9 pain pattern) |
| 3.6 | Unified | Small extension of 3.5's grid |
| **3.7** | Split a/b/c | Markup-heavy partial (the 2-4 snapshot-churn pattern) |
| 3.8 | Unified | Small single transition |
| **3.9** | Split a/b/c | Marquee: multi-region paint + multi-entity atomic txn |
| 3.10 | Unified | Small single transition |

Net Epic 3: **6 unified + 4 split (×3) = 18 stories** (from 10).

**Artifact Conflicts**
- **PRD:** None. No scope/MVP change.
- **Architecture:** None to patterns/schema. One **wording** corollary added to root `CLAUDE.md` (relocating the "all three stacks pass" invariant to the `.c` story for split groups).
- **UX:** None.
- **Process/tooling:** New policy doc; `make parity` is asserted only on `.c` stories (legitimately red on `.a`/`.b`).

---

## Section 3 — Recommended Approach

**Direct Adjustment** (Option 1): introduce a tiered story-splitting policy applied forward, preserving all completed work. Rollback (Option 2) and MVP review (Option 3) are not applicable — nothing to revert, scope unchanged.

**The split model:**

| Sub-story | Scope | Gate |
|---|---|---|
| `.a` Reference (.NET) | Prove flow, data shapes, markup, contract doc | `.NET` suite green |
| `.b` Port (Django + Go) | Idiomatic port against `.a` + contract doc | Django + Go suites green |
| `.c` Parity & DoD | `make parity`, conformance tests, snapshots, E2E, NFR timing | **Three-stack invariant gate** |

**Key design decisions:**
- The *"a story is never done until all three stacks pass it"* invariant is **relocated to `.c`** — `.a`/`.b` are legitimately `done` in a deliberately stack-divergent state.
- `.a → .b → .c` is a **hard dependency chain**: `.b` does not start until `.a` is reviewed clean (porting against an unreviewed reference reintroduces the churn).
- Split is gated on **type AND size**: behavioral/UI-integration *and* large. Small single transitions and pure data-layer stories stay unified.

**Honest trade-off:** review-round *count* may not drop (1 story × 5 rounds ≈ 3 stories × ~2 rounds), but **total effort and context per session drop**, and the same-shape bug is paid once (in the reviewed reference) instead of three times. Success is measured by context-per-session and reviewer wall-clock, not round count. Story-management overhead rises (~1.8× story files in Epic 3) — accepted, and bounded by the size gate that spares trivial stories.

---

## Section 4 — Detailed Change Proposals

### 4.1 Stories / Epics
- **Epic 3 epic file** re-broken per the table in Section 2. Split stories carry a canonical AC set (the contract) plus scoped `.a`/`.b`/`.c` sub-sections. Unified stories unchanged.
- **`sprint-status.yaml`** Epic 3 block replaced: 18 entries, all `backlog`. Epics 1/2 untouched.

### 4.2 Architecture / Conventions
- **Root `CLAUDE.md`** — added the **Split-story corollary** under the Cross-Stack Architecture Principle, relocating the invariant + `make parity` assertion to the `.c` story and pointing to the policy doc.

### 4.3 Documentation
- **New: `docs/how-to/cross-stack-story-splitting.md`** — source of truth for the tier definition (type+size gate), a/b/c structure, gate placement, the hard dependency chain, naming convention (`N.Ma/N.Mb/N.Mc`), and a worked example (3.9). Followed by `bmad-create-story` and `bmad-create-epics-and-stories`.

---

## Section 5 — Implementation Handoff

**Scope:** Moderate → Product Owner / Developer coordination.

**Completed in this change (artifacts written):**
- [x] `CLAUDE.md` — split-story corollary
- [x] `docs/how-to/cross-stack-story-splitting.md` — policy doc
- [x] `epic-3-…​.md` — Epic 3 re-break
- [x] `sprint-status.yaml` — Epic 3 18-entry structure
- [x] `sprint-change-proposal-2026-06-03.md` — this document

**Next steps (Developer / story-creation workflow):**
1. Start Epic 3 with **3-1 → 3-2 → 3-3** (unified) as before.
2. For the first split group (**3-4a → 3-4b → 3-4c**), run `bmad-create-story` per sub-story, honoring the `.a → .b → .c` dependency chain. **Do not begin 3-4b until 3-4a is reviewed clean.**
3. Confirm CI/dev treats `make parity` as authoritative only on `.c` stories.
4. Carry a note into the Epic 2 retrospective and the Epic 3 retrospective: measure context-per-session and reviewer wall-clock against the Epic 2 baseline to validate the policy.

**Success criteria:** Reduced per-session context and reviewer wall-clock on Epic 3 split stories vs. the Epic 2 baseline; same-shape bugs caught in `.a` review rather than recurring in `.b`/`.c`.

---

## Addendum (2026-06-03) — Epic 3 reordered after approval

After this proposal was approved, Epic 3 was **reordered so infrastructure precedes the transitions that render into it**, and story numbers were re-assigned to ascend with build order (the IDs had no implementation history yet, so the renumber was near-free). The Section 2 disposition table above uses the **pre-reorder** numbers; the authoritative post-reorder mapping is:

| Build order | Final ID | Pre-reorder ID | Story | Disposition |
|---|---|---|---|---|
| 1–3 | 3.1, 3.2, 3.3 | (unchanged) | Data maps + scoring helper | Unified |
| 4 | **3.4** | 3.5 | Inspection list AG Grid SSRM | Split a/b/c |
| 5 | **3.5** | 3.6 | Inspection list filters | Unified |
| 6 | **3.6** | 3.7 | Inspection detail in EntityRail | Split a/b/c |
| 7 | **3.7** | 3.4 | Schedule Inspection | Split a/b/c |
| 8 | 3.8 | (unchanged) | Start Inspection | Unified |
| 9 | 3.9 | (unchanged) | Complete + auto-open Violations | Split a/b/c |
| 10 | 3.10 | (unchanged) | Cancel Inspection | Unified |

**Rationale:** the list (3.4) and detail rail (3.6) create the `#inspection-list` / `#inspection-detail` targets that the transitions (3.7 schedule, 3.8 start, 3.9 complete, 3.10 cancel) re-render — so each transition has a real DOM target and is DOM-verifiable rather than emit-per-contract. The first split story to build is now **3.4a** (Inspection list AG Grid, .NET reference). The drafted reference story for Schedule is **`3-7a-schedule-inspection-dotnet-reference.md`** (`ready-for-dev`).

---

*Generated by `bmad-correct-course` on 2026-06-03.*
