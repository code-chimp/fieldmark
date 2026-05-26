# Component Edge-Case Checklist

Every component story must address these nine edge categories as part of its AC list. Each entry has a canonical resolution proven in Story 1.14's hardening pass — implement that resolution unless there's a documented reason to deviate.

This list exists because Epic 1 surfaced all nine categories *after* their stories had shipped (the 2026-05-17 deferred-work block). Front-loading them in story ACs collapses an entire review round per component story.

---

## 1. Unknown enum / vocabulary values

**Failure mode:** Component receives a value not in its documented vocabulary (`badge-foobar`, `data-score-band="medium"`) and renders the default style silently — no warning, no visible signal that data is malformed.

**Canonical resolution:**
- Documented fallback style class (e.g. `.badge-unknown`) with neutral but visually distinct treatment (e.g. dashed border).
- Single server-side warning log on first encounter per request (`logger.warn("unknown badge token: %s", value)`).
- A unit test that asserts an unknown value renders the fallback class and emits the warning.

**Reference:** Story 1.14 AC2.4 / Task 4; resolved entries in [deferred-work.md](../../_bmad-output/implementation-artifacts/deferred-work.md).

---

## 2. Font load failure / blocked network

**Failure mode:** Brand fonts 404 or are blocked; layout shifts or text becomes unreadable; no test catches it.

**Canonical resolution:**
- `font-display: swap` on every `@font-face` declaration.
- Playwright visual-regression test that blocks font requests and asserts CLS ≤ 0.1.
- Test must run on a CI lane that supports `layout-shift` (Chromium); skip elsewhere with a graceful guard.

**Reference:** Story 1.14 AC1.3 / Task 3.

---

## 3. JavaScript fails to initialize

**Failure mode:** A component depends on a JS initialization marker (`[data-sidebar-initialized]`, `[data-tabstrip-initialized]`) and renders hidden or jumps when the marker is absent.

**Canonical resolution:**
- CSS default is the visible, degraded-but-functional state.
- JS adds the marker → progressive enhancement engages collapse/animation/etc.
- A Playwright test with `javaScriptEnabled: false` asserts the component is visible and functional.
- A CSS-contract test (no browser) asserts the PE rule exists in `dist/fieldmark.css`.

**Reference:** Story 1.14 AC2.5 / Task 5; sidebar pattern in [fieldmark_shared/CLAUDE.md](../../fieldmark_shared/CLAUDE.md) §"Sidebar progressive enhancement".

---

## 4. AG Grid empty / loading overlay states

**Failure mode:** Grid shows AG Grid's default unstyled overlay; visually inconsistent with the design system.

**Canonical resolution:**
- `.ag-overlay-loading-center` and `.ag-overlay-no-rows-center` styled to match Basecoat surfaces.
- Loading state visually distinct from empty state (spinner vs neutral text).
- Dark-mode overrides present.

**Reference:** Story 1.14 AC2.6 / Task 6; rules in `fieldmark_shared/src/_ag-grid.css`.

---

## 5. Stacking / queueing without bound

**Failure mode:** Toaster accumulates unlimited toasts; height grows unbounded; layout breaks.

**Canonical resolution:**
- CSS `max-height` cap (e.g. 5 × toast height + gap).
- `overflow-y: auto` so overflow scrolls rather than pushes layout.
- Future Epic-2+ stories may add a JS queue helper with a hard count limit — until then, the CSS cap is the contract.
- Playwright visual-regression test for stacked-toast overflow state.

**Reference:** Story 1.14 AC2.7 / Task 7.

---

## 6. Text overflow & special characters in user-visible strings

**Failure mode:** `data-tooltip` with HTML entities renders raw `&amp;`; long content clips silently with no ellipsis; right-edge tooltips overflow the viewport.

**Canonical resolution:**
- Entities decode to characters (server-side escape, not double-escape).
- `max-width` on tooltip pseudo-element.
- `white-space: normal; word-break: break-word; text-overflow: ellipsis` on overflow.
- Per-stack escaping conformance test that round-trips entities through the rendering pipeline.

**Reference:** Story 1.14 AC2.8 / Task 8.

---

## 7. Reduced-motion preference

**Failure mode:** Users with `prefers-reduced-motion: reduce` see abrupt unmanaged transitions; or worse, the transition is the affordance ("animating in" *is* the appearance signal).

**Canonical resolution:**
- Global `@media (prefers-reduced-motion: reduce)` rule in `_a11y.css` that disables all transitions/animations.
- Affordances do not rely on motion alone — instant state changes still communicate the change (e.g., visible focus ring, aria-expanded toggle).
- axe-core scan on the affected page still reports zero WCAG 2.1 AA violations.

**Reference:** Story 1.14 AC1.1 / Task 2.

---

## 8. Forced-colors / high-contrast mode

**Failure mode:** Status communicated by color alone (badge color, score band color) becomes invisible or indistinguishable in Windows High Contrast / `forced-colors: active`.

**Canonical resolution:**
- Color is never the sole information carrier — text label and/or icon always present.
- `@media (forced-colors: active)` block with `forced-color-adjust: auto; border: 1px solid ButtonText` so badges retain a visible boundary.
- Conformance: an axe-color-contrast lane (Epic 7) verifies the rule.

**Reference:** Story 1.14 AC1.2 / Task 2.

---

## 9. Empty / whitespace text input to derived values

**Failure mode:** A helper that derives initials, slugs, or display tokens from user-supplied strings returns blank when the input is empty or whitespace-only.

**Canonical resolution:**
- Documented deterministic fallback token (e.g., `"??"` for initials, `"unnamed"` for slugs).
- Unit test for both empty string and whitespace-only string.
- The fallback is **identical across stacks** — write it once in the canonical-example doc, then implement per stack against the same test fixture.

**Reference:** Story 1.13 review-finding patch on `AvatarInitials` blank-input case.

---

## How to Use This Checklist

**When authoring a component story (`bmad-create-story`):** for each category above that applies to the component, add a Given/When/Then AC block citing the category number. Do not assume a category doesn't apply without checking — the failure modes are subtle and the cost of a missed AC is one review round.

**When implementing a story (`bmad-dev-story` / `bmad-quick-dev`):** if a category is in the AC list, implement the canonical resolution. If you deviate, document the deviation in story dev notes with rationale and update the AC.

**When reviewing a story (`bmad-code-review`):** for any component being introduced or modified, walk the nine categories and flag any AC the story should have but doesn't. If the story shipped without an AC for an applicable category, that's a defer-to-followup, not a story blocker.

---

*Ratified by Epic 1 retrospective 2026-05-25. Update procedure: a new category requires an ADR amendment.*
