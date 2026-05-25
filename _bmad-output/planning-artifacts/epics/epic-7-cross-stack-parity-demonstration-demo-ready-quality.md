# Epic 7: Cross-Stack Parity Demonstration & Demo-Ready Quality

Final epic. Locks in the artifact's persuasive purpose: every MVP scenario runs identically on all three stacks under a published harness; visual regression, axe, latency, route/index diff, and component byte-equivalence all pass clean.

## Story 7.1: Full Playwright cross-stack E2E suite for every MVP workflow

As the talk-audience persona,
I want a Playwright suite that runs every MVP user-facing workflow against all three stacks with axe-core embedded,
So that the cross-stack thesis is mechanically verifiable in one command (FR65, UX-DR39).

**Acceptance Criteria:**

**Given** `e2e/tests/`
**When** I inspect it
**Then** one spec exists per MVP user journey (Journey 1 — anchor demo; Journey 2 — Pat reject/resubmit cycle; Journey 3 — Aisha closure denial+recovery; Journey 4 — Kenji read-only browse) plus per-feature specs (project create, place-on-hold, resume, schedule inspection, complete inspection, assign violation, void violation, theme toggle, login/logout, reference-data read).

**Given** `playwright.config.ts`
**When** I inspect it
**Then** three parallel projects (`.NET` @ :5000, `Django` @ :8000, `Go` @ :3000) are configured; `make e2e` runs all specs against all three.

**Given** any spec
**When** it executes against any stack
**Then** an `@axe-core/playwright` scan runs at every meaningful render and fails the spec on any new WCAG 2.1 AA violation.

**Given** the suite completes
**When** I inspect the report
**Then** every spec passes on all three stacks (no per-stack skips).

---

## Story 7.2: Cross-stack visual regression suite

As a maintainer or demo presenter,
I want pixel-snapshot tests of the canonical screens across stacks × themes × viewports,
So that cross-stack divergence is caught at the rendering level — and Basecoat upgrades have a gate (UX-DR38).

**Acceptance Criteria:**

**Given** `e2e/tests/visual/`
**When** I inspect it
**Then** snapshots cover: Compliance Dashboard, Project Detail (Summary, Inspections tab, Violations tab, Audit tab), Violation Detail, Login page, Reference Data pages
**And** each screen is captured per stack (3) × per theme (light, dark — 2) × per viewport (1280, 1024, 768, 375 — 4) = 24 snapshots per screen.

**Given** the first run produces baselines
**When** they are reviewed and committed
**Then** subsequent runs compare against them with a tight pixel-difference threshold (≤ 0.1% for layout-stable regions; documented exception list in `e2e/visual/exceptions.md`).

**Given** any cross-stack pixel divergence beyond the threshold
**When** the suite runs
**Then** it fails the build with a diff image attached (UX-DR38).

**Given** the Basecoat version pinned in `fieldmark_shared/package.json`
**When** the version is bumped
**Then** the suite must be re-baselined as a coordinated three-stack story — the gating mechanism is the suite itself.

---

## Story 7.3: Color-blindness simulation and 200% browser-zoom verification

As an accessibility-conscious user,
I want assurance that status badges remain distinguishable under deuteranopia/protanopia and that the layout holds at 200% browser zoom,
So that WCAG 1.4.1 and 1.4.4 are mechanically verified (UX-DR37).

**Acceptance Criteria:**

**Given** `e2e/tests/a11y-color.spec.ts`
**When** it runs
**Then** Playwright applies a color-vision filter (deuteranopia, then protanopia) to canonical screens and asserts that status badges are distinguishable by their text labels (the test reads each badge's accessible text and confirms uniqueness within a list).

**Given** `e2e/tests/a11y-zoom.spec.ts`
**When** it runs
**Then** Playwright sets viewport zoom to 200% on canonical screens and asserts: no horizontal scrolling appears (outside the AG Grid acknowledged exception), no content is clipped, no interactive element loses focusable area below 44×44px.

**Given** both specs
**When** the suite runs
**Then** they execute against all three stacks (any per-stack divergence is a defect).

---

## Story 7.4: Cross-stack latency-divergence verification

As the project's NFR1 author,
I want each MVP scenario to assert a per-stack p95 ≤ 200 ms and a cross-stack divergence ≤ 50 ms p95,
So that the performance contract is mechanically tested, not asserted (NFR1).

**Acceptance Criteria:**

**Given** every E2E spec
**When** it measures action→repaint timing for the canonical interactions (Approve CA, Place On Hold, Resume, Complete Inspection, Close Project, Reject CA)
**Then** it captures p95 across N≥20 runs per stack
**And** asserts each stack's p95 ≤ 200 ms locally.

**Given** the same scenario across stacks
**When** the cross-stack divergence is computed (max(p95) − min(p95))
**Then** the test asserts the difference is ≤ 50 ms; otherwise it fails.

**Given** AG Grid row→detail interactions
**When** measured
**Then** p95 ≤ 300 ms per stack with the same cross-stack divergence rule.

---

## Story 7.5: Domain method unit-test coverage closure

As the project's FR66 author,
I want every state-transition entity method on every aggregate to have unit tests proving its invariants in each stack,
So that the domain-rule contract is enforced at the language level — not only via E2E.

**Acceptance Criteria:**

**Given** the canonical method list (`start`, `complete`, `cancel`, `place_on_hold`, `resume`, `close`, `assign`, `submit_corrective_action`, `approve_resolution`, `reject_resolution`, `void`)
**When** I inspect each stack's domain test project (`FieldMark.Tests.Domain/`, per-app `tests/test_*_state.py`, `internal/domain/*_test.go`)
**Then** every method has positive tests (happy path) and negative tests (precondition violations) — at least one test per entity rule documented in `domain-model.md`.

**Given** the cross-stack rule "method names are canonical"
**When** I `grep` each stack
**Then** the same method names appear with idiomatic casing (.NET PascalCase, Python snake_case, Go PascalCase exported / camelCase unexported) — any divergence fails Story 7.6's parity check.

**Given** `make test-{net,django,go}`
**When** each runs
**Then** all domain unit tests pass on each stack, and per-stack coverage tooling (where supported) reports ≥ 90% line coverage on the `domain/` packages.

---

## Story 7.6: Final cross-stack inventory check — routes, indexes, audit strings

As the artifact's stack-symmetry author,
I want a single command that asserts zero diff on routes, `pg_indexes`, audit action constants, HTMX target IDs, and AG Grid endpoint contracts,
So that FR58 (final) and NFR7 are mechanically gated (FR58, FR59).

**Acceptance Criteria:**

**Given** `make parity`
**When** I run it at the end of every MVP-finishing PR
**Then** it executes: `tools/parity/diff-routes.sh`, `tools/parity/diff-pg-indexes.sh`, `tools/parity/diff-audit-actions.sh` (new — greps each stack's audit-action enum/constants and compares to the canonical 14-string list), `tools/parity/diff-target-ids.sh` (new — greps each stack's templates for HTMX `id="..."` and asserts only the canonical inventory appears), `tools/parity/diff-grid-endpoints.sh` (new — asserts the four `/grid/*` endpoints respond with the contract shape on all three stacks).

**Given** any divergence
**When** the script runs
**Then** it exits non-zero with a human-readable diff identifying which stack diverged and on which contract.

**Given** the README at repo root
**When** I read the "Demo Run" section
**Then** it documents the order: `make reset && make run-{net,django,go} && make parity && make e2e` as the smoke-test recipe.

---

## Story 7.7: Canonical component example gallery with per-stack snapshot tests

As a developer adding or modifying any custom component,
I want a canonical static-HTML gallery of every component and per-stack tests that the wrapper output is byte-identical,
So that component drift across stacks is caught at unit-test time, not in E2E (UX-DR40).

**Acceptance Criteria:**

**Given** `fieldmark_shared/components/`
**When** I list it
**Then** each of StatusBadge, ActionButton, ComplianceTile, AuditRow, EntityRail, DashboardTile, InlineAlert, ThemeToggle, AGGridPanel, TabStrip, FlashRegion has its own subdirectory with `<component>.example.html` files showing every state combination.

**Given** each stack's component test suite
**When** it runs
**Then** for each canonical example, it renders the per-stack wrapper with the same inputs and asserts byte-equivalence against the example HTML (whitespace-normalized).

**Given** any wrapper produces non-byte-identical output
**When** the test runs
**Then** it fails with a diff identifying the divergent component and stack.

---

## Story 7.8: AG Grid axe ruleset, manual SR test recipe, and Demo Run documentation

As the artifact's accessibility-stewardship author,
I want AG Grid axe disables documented per rule with rationale, a manual screen-reader test recipe, and a one-line "Demo Run" recipe in the README,
So that the artifact's accessibility posture is honest and reproducible (UX-DR36, NFR3).

**Acceptance Criteria:**

**Given** `tests/axe-config.json` (or per-stack equivalent referenced by `@axe-core/playwright`)
**When** I read it
**Then** every disabled AG Grid axe rule is listed with: rule id, AG Grid version, rationale, review-on-upgrade flag
**And** disables are reviewed each AG Grid upgrade per UX-DR36.

**Given** `_bmad-output/planning-artifacts/manual-a11y-recipe.md` (new)
**When** I read it
**Then** it documents: how to run VoiceOver on Safari/macOS and NVDA on Firefox/Windows against the canonical anchor demo; what to assert at each step; what acceptable observed behavior looks like; cadence (per major milestone, quarterly minimum).

**Given** the repo root `README.md`
**When** I read its "Demo Run" section
**Then** it documents the one-line recipe (`make reset && make run-{net,django,go}` in three terminals, then `make parity && make e2e` to verify, then navigate to each stack's URL to demo) and a brief talking-track outline pointing at Stories 5.5 (anchor demo), 6.4 (denial+recovery), and the cross-stack parity assertion.
