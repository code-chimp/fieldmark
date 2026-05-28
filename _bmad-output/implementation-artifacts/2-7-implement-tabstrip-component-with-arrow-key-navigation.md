# Story 2.7: Implement TabStrip component with arrow-key navigation

Status: ready-for-dev

Epic: 2 — Project Lifecycle & Compliance Dashboard
Source AC: [_bmad-output/planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md](../planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md) §Story 2.7
Depends on: Story 2.4 (per-component directory convention + snapshot harness), Story 2.5 (Go sibling-`.go`-file precedent), Story 2.6 (slot-pass-through grep-guard scoping pattern, e2e Playwright suite location). This story introduces the first **JS-bearing** component in Epic 2 (~15 lines of arrow-key navigation logic), vendored under `fieldmark_shared/vendor/tabstrip/` and symlinked into the three stacks — same pattern Story 1.6 used for `theme-toggle.js`.

## Story

As any keyboard or screen-reader user of the Project Detail anchor screen (Story 2.11) — Compliance Officer, Project Manager, Site Supervisor, Administrator — and as the developer composing future tabbed surfaces (Story 2.13 Audit tab is the first non-2.11 consumer),
I want a markup-only **TabStrip** wrapper per stack that renders a `<nav role="tablist">` containing `<button role="tab" aria-selected aria-controls hx-get hx-target hx-swap>` per tab — with optional unread-count badges after the label — paired with one tiny shared JavaScript file (`tabstrip.js`, ≤ 20 LOC) implementing Left/Right arrow-key cycling between tabs (focus moves; Enter / Space activates),
So that Story 2.11 (Project Detail) can compose Summary / Inspections / Violations / Audit tabs without inventing markup, the active-tab `aria-selected` flip-via-OOB-swap pattern (UX-DR Pattern 5) has its target shape locked in, full WAI-ARIA tablist semantics ship from the first tabbed surface, and arrow-key navigation across tabs satisfies WCAG 2.1.1 (keyboard) and 2.4.3 (focus order) without per-stack divergence.

**Scope boundary:** this story produces (a) `fieldmark_shared/components/tab_strip/canonical.html` (variant-delimited fixture, same convention Stories 2.4 / 2.5 / 2.6 use) + `README.md` (contract), (b) one markup-only wrapper template per stack in idiomatic locations, (c) per-stack snapshot tests asserting byte-equality, (d) **the vendored `fieldmark_shared/vendor/tabstrip/tabstrip.js` file** (≤ 20 LOC, no build step, no transpile) + the three symlinks into each stack's static vendor directory + base-layout `<script>` tags loading it, (e) **one Playwright behavior test** exercising arrow-key navigation against a fixture page (this is one of three documented exceptions to "no client-side a11y tests" — UX [ux-design-specification.md:1252](../planning-artifacts/ux-design-specification.md)), (f) appended row in `docs/reference/component-canonical-examples.md` Component Index, (g) appended entry in the JS budget table inside the wrapper's README (cross-referenced to UX-DR §"JS budget per component" at [ux-design-specification.md:932–939](../planning-artifacts/ux-design-specification.md)). **Out of scope:** any consumer page that actually mounts the TabStrip with real tab definitions (Story 2.11 Project Detail mounts the four Summary/Inspections/Violations/Audit tabs; Story 2.13 mounts the Audit tab content), the **OOB-swap response shape** that updates `aria-selected` on the active tab after a click (consumer story owns the server response composition — this story only documents the *target shape* that the OOB swap must produce), any `hx-push-url` URL-syncing for tab state (UX explicitly prohibits this unless deliberately enabled per consumer — out of scope this story), the tab-panel content itself (`#project-detail-tab-content` markup with `role="tabpanel"` is the consumer's responsibility — this story documents the contract surface but does not land the panel wrapper), CSS for the active-tab underline indicator (Basecoat already provides; do not add custom CSS this story).

## Acceptance Criteria

### AC1 — Canonical-example directory under `fieldmark_shared/components/`

**Given** the per-component-directory convention introduced by Stories 2.4 / 2.5 / 2.6
**When** I inspect `fieldmark_shared/components/`
**Then** a new sub-directory exists with this exact layout:

```
fieldmark_shared/components/tab_strip/
├── canonical.html
└── README.md
```

**And** `canonical.html` follows the same variant-delimited format as `action_button.example.html` (Story 1.12) — a leading HTML comment block documenting `fixture:` inputs, then `<!-- variant: <name> (inputs: ...) -->` delimiters separating each variant's expected output.

**And** `canonical.html` contains **exactly seven variant blocks**:

1. `project-detail-four-tabs-summary-active` — the canonical Project Detail tablist with four tabs `[Summary, Inspections, Violations, Audit]`; `active_index=0` (Summary active). All four tabs render with `hx-get` URLs (placeholder canonical strings `/projects/__ID__/summary`, `/projects/__ID__/inspections`, `/projects/__ID__/violations`, `/projects/__ID__/audit`); `hx-target="#project-detail-tab-content"`, `hx-swap="innerHTML"`. **No badges.** Summary tab has `aria-selected="true"`; others have `aria-selected="false"`.
2. `project-detail-four-tabs-violations-active` — same four tabs, `active_index=2` (Violations active). Shows that the `aria-selected="true"` shifts.
3. `project-detail-four-tabs-with-badges` — same four tabs, `active_index=0`, **with badges** on Inspections (`count=12`), Violations (`count=3`), Audit (`count=147`). Summary has no badge. Demonstrates the optional badge slot.
4. `two-tabs-minimal` — minimum-viable variant: two tabs `[Open, Closed]`, no badges, `active_index=0`, `hx-get="/__tab__/open"` / `/__tab__/closed`, `hx-target="#__panel__"`. Establishes the contract for arbitrary tablist consumers (not just Project Detail).
5. `single-tab` — **edge case** — one tab, `active_index=0`, no badge. This is unusual but must render correctly (a tablist with one tab is degenerate but valid ARIA; the arrow-key script must handle a single-tab tablist without erroring — see AC4 §single-tab-handling).
6. `badge-zero` — same as variant #3 but with `count=0` on the Violations tab. The badge MUST still render (the consumer page is responsible for *whether* a zero-count badge appears; the wrapper renders whatever count it receives). The badge text is the literal `0`. Per-stack test asserts the badge element is present in the rendered output for `count=0` and is absent only when `badge=None` / `null`.
7. `badge-large` — `count=9999` on one tab to verify no width-related markup is emitted (no `max-width` inline style, no truncation HTML — width is a CSS concern handled by Basecoat).

**And** `README.md` documents the contract in this fixed order: (1) Purpose (one sentence — "Horizontal `<nav role='tablist'>` with HTMX-bound `<button role='tab'>` tabs and optional unread-count badges; pairs with the vendored `tabstrip.js` arrow-key navigation script for full WAI-ARIA tablist behavior."), (2) Required props with types (`tabs: list of TabSpec`, where `TabSpec = { id: string, label: string, hx_get: string, hx_target: string, badge_count: int? }`; `active_index: int`; `hx_swap: string = "innerHTML"` — defaults documented), (3) Variant list (the seven names), (4) ARIA invariants (`<nav role="tablist">` outer; per tab `role="tab"`, `aria-selected="true"|"false"`, `aria-controls="<panel-id>"`, `id="<tab-id>"`; the panel — owned by consumer, not by this wrapper — must carry `role="tabpanel"` + `aria-labelledby="<tab-id>"` per UX-DR33 at [ux-design-specification.md:1217](../planning-artifacts/ux-design-specification.md)), (5) HTMX attributes per tab (`hx-get="<url>"`, `hx-target="<panel-id>"`, `hx-swap="<swap>"` — defaults `innerHTML`; `hx-disabled-elt="this"` is NOT applied here because tabs are navigation triggers, not state-mutating actions — UX Pattern 8 latency indication still applies via the `htmx-request` class which Basecoat handles), (6) Badge contract (badge renders as `<span class="badge tab-strip__badge" aria-label="<count> unread">{{count}}</span>` when `badge_count` is not null; the `aria-label` is "<count> unread" because the convention in MVP is unread-count badges — if a future consumer needs a different semantic, the wrapper accepts an optional `badge_aria_template` prop but this story does not introduce that prop; see Dev Notes §"Badge semantic monoculture"), (7) Allowed class vocabulary (`tab-strip`, `tab-strip__tab`, `tab-strip__label`, `tab-strip__badge` — BEM, matching prior story precedent; Basecoat `nav-tabs` and `nav-tab` classes are also applied for visual styling — verify the Basecoat 0.3.11 class names by reading the existing `dist/fieldmark.css` for `.nav-tabs` rules before authoring `canonical.html`; if Basecoat uses different class names, follow Basecoat's naming and document it), (8) JS budget (one-line: "Pairs with `fieldmark_shared/vendor/tabstrip/tabstrip.js` (≤ 20 LOC) for arrow-key navigation. The wrapper itself is zero-JS — it emits the markup that the script attaches to via `[data-tabstrip]` selector. The script is symlinked into each stack's static vendor directory and loaded via a `<script src='/vendor/tabstrip/tabstrip.js' defer>` tag in the base layout."), (9) OOB-swap contract for active-tab updates (one-line: "When a tab is clicked, the consumer's response replaces `#project-detail-tab-content` (the `hx-target`) **and** emits an OOB swap targeting the tablist that re-renders the entire `<nav role='tablist'>` with the new `aria-selected` flipped. The wrapper's canonical HTML is the OOB target — Story 2.11 consumes this contract by emitting an `hx-swap-oob='outerHTML'` block whose root is `<nav id='project-detail-tabstrip' role='tablist'>...</nav>`. Adding an `id` to the outer `<nav>` is part of the loaded variants — see AC2."), (10) Snapshot-equality requirement (standard line).

**And** `docs/reference/component-canonical-examples.md` is MODIFIED — one new row appended to the Component Index table with columns: `TabStrip`, fixture path, README path, three wrapper paths, three test paths.

### AC2 — TabStrip markup contract (UX-DR §"TabStrip", UX Pattern 5 — Anchor Screen With HTMX Tabs, UX-DR33 — strict heading + landmark rules)

**Given** the TabStrip contract at [ux-design-specification.md:910–916](../planning-artifacts/ux-design-specification.md) and UX Pattern 5 at [ux-design-specification.md:1032–1047](../planning-artifacts/ux-design-specification.md) and the OOB-swap requirement at line 1040
**When** I inspect `fieldmark_shared/components/tab_strip/canonical.html` for the `project-detail-four-tabs-summary-active` variant
**Then** the rendered HTML is **exactly** this shape (whitespace-normalized; attribute order alphabetized for the byte-equality comparison; per-tab id derived from the tab spec):

```html
<nav
  aria-label="Project Detail Tabs"
  class="nav-tabs tab-strip"
  data-tabstrip
  id="project-detail-tabstrip"
  role="tablist"
>
  <button
    aria-controls="project-detail-tab-content"
    aria-selected="true"
    class="nav-tab tab-strip__tab"
    hx-get="/projects/__ID__/summary"
    hx-swap="innerHTML"
    hx-target="#project-detail-tab-content"
    id="tab-summary"
    role="tab"
    tabindex="0"
    type="button"
  ><span class="tab-strip__label">Summary</span></button>
  <button
    aria-controls="project-detail-tab-content"
    aria-selected="false"
    class="nav-tab tab-strip__tab"
    hx-get="/projects/__ID__/inspections"
    hx-swap="innerHTML"
    hx-target="#project-detail-tab-content"
    id="tab-inspections"
    role="tab"
    tabindex="-1"
    type="button"
  ><span class="tab-strip__label">Inspections</span></button>
  <button
    aria-controls="project-detail-tab-content"
    aria-selected="false"
    class="nav-tab tab-strip__tab"
    hx-get="/projects/__ID__/violations"
    hx-swap="innerHTML"
    hx-target="#project-detail-tab-content"
    id="tab-violations"
    role="tab"
    tabindex="-1"
    type="button"
  ><span class="tab-strip__label">Violations</span></button>
  <button
    aria-controls="project-detail-tab-content"
    aria-selected="false"
    class="nav-tab tab-strip__tab"
    hx-get="/projects/__ID__/audit"
    hx-swap="innerHTML"
    hx-target="#project-detail-tab-content"
    id="tab-audit"
    role="tab"
    tabindex="-1"
    type="button"
  ><span class="tab-strip__label">Audit</span></button>
</nav>
```

**Key invariants embedded in the markup contract:**

- **The outer `<nav>` carries an `id`** (`project-detail-tabstrip` in the canonical fixtures). The wrapper accepts an `id` prop; if the caller passes none, the wrapper does **not** invent one. The OOB swap pattern requires the `id` to target the whole strip — see AC1 §contract-item-9.
- **`aria-label` on `<nav>`** names the tablist. Defaults to `"<wrapper-supplied label>"` — wrapper takes an `aria_label` prop. Required (no default) — the consumer must supply a semantic name. A per-stack test asserts that omitting `aria_label` triggers a clear error (or for Go where `string` is the zero value, a runtime panic with a descriptive message is acceptable — match the existing wrapper precedent for required props).
- **The `data-tabstrip` attribute** on the outer `<nav>` is the **selector hook for the arrow-key JS** (AC4). It is **value-less** (`data-tabstrip` with no `=""`), which is the HTML5 idiomatic form for boolean data attributes. Per-stack templating engines may render it as `data-tabstrip=""` or `data-tabstrip` — these are equivalent in HTML5. The snapshot-test normalizer must treat both as equal. **Verify the Story 1.11 normalizer handles value-less attributes** — if it canonicalizes them to `data-tabstrip=""`, all good; if not, extend the normalizer (do not work around by emitting `="true"` etc.).
- **`tabindex="0"` on the active tab; `tabindex="-1"` on inactive tabs** — this is the **roving tabindex pattern** required by WAI-ARIA tablist (only one tab is in the natural tab order; arrow-keys move focus among the others). The wrapper computes this from `active_index` — a per-stack helper (similar to Story 2.5 band resolver) emits `tabindex="0"` for the active tab and `tabindex="-1"` for the rest. A per-stack unit test verifies the tabindex distribution for active_index=0, active_index=2, and active_index=N-1 (last tab).
- **`type="button"`** is mandatory on every `<button>` — without it, the browser default is `type="submit"`, which would submit any enclosing form on Enter. The TabStrip is never inside a `<form>` in canonical FieldMark usage, but the defense-in-depth is cheap; per-stack snapshot tests assert `type="button"` is present.
- **`aria-controls`** points to a **single** panel id — `project-detail-tab-content` for Project Detail. **All four tabs control the same panel** because HTMX swaps replace the panel content in place; there is one panel, not four. This is the FieldMark idiom — distinct from a static-HTML tablist with four sibling panels (which would have four `aria-controls` values). The README documents this distinction.
- **No `<a href>` — all tabs are `<button>`.** UX Pattern 5 prohibits page navigation on tab switch; using `<a>` would tempt browsers into honoring the href (when JS fails to intercept) and would also surface in the address bar. `<button>` is the unambiguous semantic. The link-rather-than-button question is closed.
- **Label inside `<span class="tab-strip__label">`, not directly inside `<button>`.** Reason: the badge (when present) needs to sit *outside* the label span but *inside* the button, so the button's accessible name combines label + badge via the `<span>` grouping. See AC3 for the badge variants.

**And** the `project-detail-four-tabs-violations-active` variant differs only in:
- `aria-selected` values shift (Violations becomes `"true"`, Summary becomes `"false"`),
- `tabindex` values shift correspondingly (Violations becomes `0`, others `-1`).

The id, label text, hx-get URLs, and DOM order of the tabs are identical across the `summary-active` and `violations-active` variants.

**And** the `two-tabs-minimal` and `single-tab` variants follow the same skeletal pattern with fewer `<button>` elements. They carry a different outer `id` (matching the variant's named example) and a different `aria-label`.

### AC3 — Badge rendering contract (variants 3, 6, 7)

**Given** the badge contract in AC1 §contract-item-6
**When** I inspect the `project-detail-four-tabs-with-badges` variant in `canonical.html`
**Then** each tab with `badge_count` not null renders the badge **inside** the `<button>`, **after** the `<span class="tab-strip__label">`, with this exact shape:

```html
<button ...>
  <span class="tab-strip__label">Inspections</span>
  <span aria-label="12 unread" class="badge tab-strip__badge">12</span>
</button>
```

**And** when `badge_count` is null, no `<span class="tab-strip__badge">` is emitted at all (not an empty `<span>`, not a `<span>` with `display:none`). The footer-omission test pattern from Story 2.6 §AC4 applies — assert by absence.

**And** the badge `aria-label` uses the literal template `"<count> unread"` — note the singular form for `count=1` (`"1 unread"`, NOT `"1 unreads"` and NOT `"1 unread item"`). This story does NOT pluralize-correctly — see Dev Notes §"Badge semantic monoculture" for the explicit "deliberately not pluralizing" decision and the deferred-work entry for future i18n / pluralization.

**And** the badge text is the **literal `badge_count` value** stringified. For `count=0`, the badge renders with `0`. For `count=9999`, the badge renders with `9999`. No truncation (no "99+"); no formatting (no thousands separator). The wrapper does not own truncation — consumer can pre-format the count if it wants `"99+"`.

**And** per-stack tests for badge rendering:
1. `badge_count=12` → badge present with text `12` and `aria-label="12 unread"`.
2. `badge_count=0` → badge present with text `0` and `aria-label="0 unread"`. (Variant 6.)
3. `badge_count=null` → no badge element in rendered output.
4. `badge_count=9999` → badge present with text `9999`; no truncation markup. (Variant 7.)
5. `badge_count=-1` → **edge case** — the wrapper renders the negative number verbatim. A per-stack unit test asserts this is not silently coerced to `0` or `null`; if a consumer passes a negative count, that's the consumer's bug (a deferred-work entry could be added if a reviewer judges this surface needs hardening — recommended NOT this story to avoid scope creep).

### AC4 — Arrow-key navigation script (`fieldmark_shared/vendor/tabstrip/tabstrip.js`, ≤ 20 LOC)

**Given** the JS budget at [ux-design-specification.md:932–939](../planning-artifacts/ux-design-specification.md) ("TabStrip: ~15 lines for arrow-key navigation") and the epic AC ("the keyboard JS is ≤ 15 LOC and vendored as `tabstrip.js`")
**When** I inspect `fieldmark_shared/vendor/tabstrip/tabstrip.js`
**Then** the file is a self-contained ES5-compatible IIFE following the **Story 1.6 `theme-toggle.js` precedent** (see [fieldmark_shared/vendor/theme-toggle/theme-toggle.js](../../fieldmark_shared/vendor/theme-toggle/theme-toggle.js)). No build step, no transpile, no module syntax, no `let`/`const` (use `var`), no arrow functions, no `addEventListener` options object (use third-arg boolean only if needed), no template literals. ES5 ensures broad browser compatibility without requiring a polyfill layer — the same posture established by `theme-toggle.js`. The script is **≤ 20 LOC** of non-comment, non-blank lines (the AC text says "≤ 15 LOC" — the per-stack snapshot tests will count; if the script needs the additional 5 lines to handle the edge cases below cleanly, the README JS-budget line records the actual final count).

**Behavior contract (executable specification — the test in AC6 verifies all of these):**

1. **Selector hook.** The script attaches behavior to every `<nav role="tablist" data-tabstrip>` on the page. It uses `document.querySelectorAll('nav[data-tabstrip]')` at script-load time (the `defer` on the `<script>` tag means DOMContentLoaded has fired). It does **not** re-scan after HTMX swaps — see Re-attachment after OOB swap below.
2. **Arrow keys move focus, not activation.**
   - `ArrowLeft` / `ArrowRight` → focus moves to the previous / next `<button role="tab">` in the tablist's child order. Focus wraps: left-arrow from the first tab moves focus to the last; right-arrow from the last moves to the first. `tabindex` is updated to keep the roving-tabindex pattern correct — the focused tab gets `tabindex="0"` and the previously-focused tab gets `tabindex="-1"`.
   - **`Home` → focus the first tab.** **`End` → focus the last tab.** (WAI-ARIA tablist best practice — adds two more lines but the cost is trivial and the accessibility win is clear.)
3. **Enter / Space activate the focused tab.** The script calls `.click()` on the focused tab — this fires HTMX's bound click handler, sending the `hx-get` request. The script does **not** call `preventDefault()` for Enter / Space on a `<button>` because the browser default is already to fire `click`; explicit `.click()` is for `Space` which the browser converts to click on keyup, not keydown — making sure both keys land the click handler before any consumer-side keydown handler interferes.
4. **Re-attachment after OOB swap.** When the consumer's response replaces the `<nav>` via `hx-swap-oob`, the new node has no event listeners. The script listens for the global `htmx:afterSwap` event on `document.body` and re-runs the attachment logic against any newly-arrived `nav[data-tabstrip]`. This is **the one piece of complexity** the 15-LOC budget pays for — without it, arrow-keys stop working after the first tab click. The re-attachment must be idempotent (calling it twice on the same `<nav>` must not double-bind keydown listeners — track attachment via a `_tabstripBound` property on the node or by removing existing listeners before adding).
5. **Single-tab tablist.** For a tablist with exactly one `<button role="tab">`, arrow keys are a no-op (focus stays where it is). The script does not error. Variant 5 (`single-tab`) in `canonical.html` exists specifically to make this case observable.
6. **Disabled tabs.** Out of scope — no MVP TabStrip variant has disabled tabs. If a future consumer needs to disable a tab, the script's behavior is undefined and the README does not document it. Do not preemptively code for disabled tabs.
7. **Pointer events untouched.** Mouse click on a tab fires the HTMX request via HTMX's own bound handler — the script does not intercept clicks. It listens only on `keydown`.

**Concrete script shape (reference; the implementer may rearrange as long as the behavior contract holds):**

```js
(function () {
  'use strict';
  function focusTab(tabs, i) {
    tabs.forEach(function (t, j) { t.setAttribute('tabindex', j === i ? '0' : '-1'); });
    tabs[i].focus();
  }
  function attach(strip) {
    if (strip._tabstripBound) return;
    strip._tabstripBound = true;
    var tabs = Array.prototype.slice.call(strip.querySelectorAll('button[role="tab"]'));
    strip.addEventListener('keydown', function (e) {
      var i = tabs.indexOf(document.activeElement);
      if (i < 0) return;
      if (e.key === 'ArrowLeft') { focusTab(tabs, (i - 1 + tabs.length) % tabs.length); e.preventDefault(); }
      else if (e.key === 'ArrowRight') { focusTab(tabs, (i + 1) % tabs.length); e.preventDefault(); }
      else if (e.key === 'Home') { focusTab(tabs, 0); e.preventDefault(); }
      else if (e.key === 'End') { focusTab(tabs, tabs.length - 1); e.preventDefault(); }
      else if (e.key === 'Enter' || e.key === ' ') { tabs[i].click(); e.preventDefault(); }
    });
  }
  function scan() {
    document.querySelectorAll('nav[data-tabstrip]').forEach(attach);
  }
  scan();
  document.body.addEventListener('htmx:afterSwap', scan);
})();
```

This is 24 non-blank lines as written — the AC says "≤ 20" (the buffer over UX-DR's "~15"); refactor inline if necessary to hit 20, but **prefer clarity over byte-counting** as long as the 20-line ceiling holds. The README records the actual final count.

**And** the file is committed at `fieldmark_shared/vendor/tabstrip/tabstrip.js`. Per [fieldmark_shared/CLAUDE.md](../../fieldmark_shared/CLAUDE.md) §"How the Vendor JS Symlinks Work", the `vendor/tabstrip/` directory is symlinked into each stack:

| Stack | Symlink |
|---|---|
| .NET | `FieldMark/FieldMark.Web/wwwroot/vendor/tabstrip` → `../../../../fieldmark_shared/vendor/tabstrip` |
| Django | `fieldmark_py/static/vendor/tabstrip` → `../../../fieldmark_shared/vendor/tabstrip` |
| Go/Fiber | `fieldmark-go/internal/web/static/vendor/tabstrip` → `../../../../../fieldmark_shared/vendor/tabstrip` |

Each base layout (introduced by Story 1.5, modified here) adds **one** `<script src="/vendor/tabstrip/tabstrip.js" defer></script>` tag — placed alongside the existing `htmx.min.js` and `theme-toggle.js` script tags. The `defer` attribute is mandatory (the script depends on `document.body` existing when it runs); do not use `async`. Update the three base layout files per stack:

- **.NET:** `FieldMark/FieldMark.Web/Pages/Shared/_Layout.cshtml` — add the script tag in the same section as `theme-toggle.js`.
- **Django:** `fieldmark_py/templates/base.html` — same.
- **Go:** `fieldmark-go/internal/web/templates/base.html` (or wherever the Story 1.5 base layout lives — search for the existing `theme-toggle.js` tag and add alongside).

Update [fieldmark_shared/CLAUDE.md](../../fieldmark_shared/CLAUDE.md) §"How the Vendor JS Symlinks Work" — append a `tabstrip` column to the symlink table (or add a row noting the new symlink in the explanatory text — whichever the existing doc convention prefers; do not invent a new section). This documentation edit is mandatory; future contributors must see `tabstrip/` as part of the vendor set.

### AC5 — Per-stack wrapper templates

**Per-stack wrapper paths** (idiomatic per stack — no shared template, no symlinked partial):

- **.NET (Razor partial):** `FieldMark/FieldMark.Web/Pages/Shared/Components/_TabStrip.cshtml` (NEW). Invoked as `<partial name="Shared/Components/_TabStrip" model="@(new TabStripViewModel(id, ariaLabel, tabs, activeIndex))" />`. View model lives in-file: `record TabStripViewModel(string Id, string AriaLabel, IReadOnlyList<TabSpec> Tabs, int ActiveIndex); record TabSpec(string Id, string Label, string HxGet, string HxTarget, int? BadgeCount);`. The `hx_swap` default is hard-coded to `"innerHTML"` in the template body (it is not part of the prop set — the wrapper is opinionated; if a future consumer needs a different swap, expose the prop then). The roving-tabindex computation is a one-line `Model.Tabs.Select((t, i) => i == Model.ActiveIndex ? "0" : "-1")` projection or equivalent — keep the template body itself free of per-tab `@if` cascades.
- **Django (template include):** `fieldmark_py/templates/components/_tab_strip.html` (NEW). Invoked as `{% include "components/_tab_strip.html" with id="project-detail-tabstrip" aria_label="Project Detail Tabs" tabs=tab_list active_index=0 %}`. `tab_list` is a list of dicts: `[{"id": "tab-summary", "label": "Summary", "hx_get": "...", "hx_target": "...", "badge_count": None}, ...]`. The template iterates `{% for tab in tabs %}` and uses `{% if forloop.counter0 == active_index %}…{% endif %}` only at the tabindex / aria-selected boundary — keep the body free of nested conditionals.
- **Go (`html/template` `{{define}}` block):** `fieldmark-go/internal/web/templates/components/tab_strip.html` (NEW). Defines `{{define "tab_strip"}}…{{end}}`. Args struct + helper in sibling `tab_strip.go` (NEW; follows Story 2.5 / 2.6 sibling-`.go`-file precedent): `type TabStripArgs struct { ID, AriaLabel string; Tabs []TabSpec; ActiveIndex int }`; `type TabSpec struct { ID, Label, HxGet, HxTarget string; BadgeCount *int }` (pointer for the null-vs-zero distinction, same idiom as Story 2.5's `*int`). A helper function `tabTabindex(activeIndex, i int) string` returns `"0"` or `"-1"`; register it in the template-function map (`safeHTML` was added by Story 2.6 in the same map — add alongside). The template body iterates with `{{range $i, $t := .Tabs}}` and uses `{{tabTabindex $.ActiveIndex $i}}` inline.

**And** each wrapper's top-of-file comment references `docs/reference/component-canonical-examples.md` (matching prior wrapper precedent).

**And** none of the wrappers introduce slot pass-through — `label` and `aria_label` strings are framework-escaped (auto-escape in force). The grep guard from Stories 2.4 / 2.5 (`Html.Raw` / `|safe` / `template.HTML(` all forbidden) applies to TabStrip wrappers with **no exceptions** — there are no slot props in this component. A per-stack lint asserts zero occurrences of those tokens in the three new wrapper files.

### AC6 — Playwright behavior test for arrow-key navigation (the explicit exception to "no client-side a11y tests")

**Given** UX §"No client-side tests for accessibility patterns" at [ux-design-specification.md:1252](../planning-artifacts/ux-design-specification.md) ("The exceptions (TabStrip arrow-key navigation, ThemeToggle cycle, FlashRegion auto-dismiss) have small dedicated tests")
**When** I inspect the e2e suite
**Then** a new spec exists at `e2e/tests/shared/tabstrip-keyboard-navigation.spec.ts` (or the verified equivalent location per Story 2.6's resolution of the e2e directory layout) with these test cases — load against a fixture page (reuse the Story 2.6 fixture-page mechanism: a `/_test/render-partial/tab-strip` endpoint on .NET or a debug-gated Django view; do **not** introduce a production route). The fixture renders the `project-detail-four-tabs-summary-active` variant:

1. **Initial focus order.** Focus the first tab (`tab-summary`) via `page.keyboard.press('Tab')` from the body. Assert `document.activeElement.id === 'tab-summary'`.
2. **Right-arrow cycles forward.** From `tab-summary`, press `ArrowRight`. Assert focus moves to `tab-inspections`, and that `document.activeElement.getAttribute('tabindex') === '0'` while `tab-summary.getAttribute('tabindex') === '-1'`.
3. **Right-arrow wraps from last to first.** From `tab-audit`, press `ArrowRight`. Assert focus moves to `tab-summary`.
4. **Left-arrow cycles backward.** From `tab-summary`, press `ArrowLeft`. Assert focus moves to `tab-audit` (wrap).
5. **Home / End.** From any tab, press `Home`. Assert focus moves to `tab-summary`. Press `End`. Assert focus moves to `tab-audit`.
6. **Enter activates.** Focus `tab-inspections`. Mock the HTMX request (intercept `/projects/__ID__/inspections` to return a 200 with a stub body — Playwright `page.route()`). Press `Enter`. Assert the request was fired (intercept hit count == 1) and that the response body landed in `#project-detail-tab-content`.
7. **Space activates.** Same as #6 but with `Space`. Assert one request, one swap.
8. **Single-tab no-op.** Load the `single-tab` variant. Focus the tab. Press `ArrowRight`. Assert focus did not change and no JS error occurred (use `page.on('pageerror')` to assert).
9. **OOB-swap re-attachment.** Render the `summary-active` variant; click `tab-inspections`. The mocked response includes an `hx-swap-oob='outerHTML:#project-detail-tabstrip'` block carrying the `inspections-active` shape. After the swap settles, focus the new active tab and press `ArrowRight` — assert focus moves to the next tab (proves the `htmx:afterSwap` listener re-attached behavior to the new node).

**And** the test sets `page.on('console', ...)` to capture any console errors; assert that the test ends with zero console errors from the page (one of the recurring Epic 1 regression vectors was silent JS errors caught only by users).

**And** the test file's top-of-file comment cites:
- UX §"No client-side tests for accessibility patterns" exception list at [ux-design-specification.md:1252](../planning-artifacts/ux-design-specification.md),
- UX-DR §"TabStrip" component spec at [ux-design-specification.md:910–916](../planning-artifacts/ux-design-specification.md),
- WAI-ARIA Authoring Practices "Tabs" pattern (https://www.w3.org/WAI/ARIA/apg/patterns/tabs/) — this URL is documentation-only; do not link to it from production code.

### AC7 — Component edge-case checklist coverage (per [component-edge-case-checklist.md](../../docs/reference/component-edge-case-checklist.md))

Walked the nine categories — only the **applicable** ones get an AC:

**Given** category 3 (JavaScript fails to initialize)
**When** `tabstrip.js` fails to load (404, CSP block, JS disabled)
**Then** the CSS-default state of the tablist is "tabs visible, first tab visually marked active, all clickable via mouse and Enter" — i.e., the tabs work as plain HTMX buttons without arrow-key cycling. The wrapper does **NOT** depend on the script for visibility, layout, or basic activation. Per Story 1.14's "no JS init marker" rule and the [component-edge-case-checklist.md §3](../../docs/reference/component-edge-case-checklist.md) canonical resolution. A Playwright test with `javaScriptEnabled: false` (matching the Story 1.14 sidebar precedent) asserts: (a) all four tabs visible, (b) clicking the second tab via mouse fires a navigation request (HTMX is loaded separately and works without `tabstrip.js`; if HTMX is also disabled, the `<button>`s simply don't fire and the page is fully degraded — this is acceptable per progressive-enhancement), (c) no layout shift or hidden content.

**Given** category 6 (text overflow & special characters in user-visible strings)
**When** the wrapper receives a tab `label` with XSS-prone characters (`<`, `>`, `&`, `"`, `'`)
**Then** framework default auto-escaping converts them to entities — a per-stack XSS-payload round-trip test: `label="<script>alert(1)</script>"` produces `&lt;script&gt;alert(1)&lt;/script&gt;` inside `<span class="tab-strip__label">`. The same applies to `aria_label` and `id` props (defense in depth — `id` should be a known canonical string but is escaped anyway). Long labels (~100 chars) clip via the existing `.truncate` utility — add `class="tab-strip__label truncate"` if the test author judges long-label rendering is a real concern; otherwise omit for this story (a Project Detail tab label is always one of Summary / Inspections / Violations / Audit per UX, so the surface is theoretical). Defer the long-label-test to Epic 7 visual regression.

**Given** category 7 (reduced-motion preference)
**When** the active-tab underline indicator transitions between tabs
**Then** the global `@media (prefers-reduced-motion: reduce)` rule in `_a11y.css` (Story 1.14) handles transition suppression. The TabStrip introduces no new transitions or animations. The underline indicator is the Basecoat `.nav-tab[aria-selected="true"]` styling, which is a CSS class swap (instant), not an animation. Verify Basecoat does not include an active-indicator transition; if it does, the Story 1.14 reduced-motion rule already covers it.

**Given** category 8 (forced-colors / high-contrast mode)
**When** I render the TabStrip in `forced-colors: active`
**Then** the active tab's `aria-selected="true"` is communicated by both ARIA semantics (screen reader will say "selected") AND by the visible indicator. The Story 1.14 `_a11y.css` rule applies `forced-color-adjust: auto; border: 1px solid ButtonText` on Basecoat surface elements; verify the rule covers `.nav-tab`. If it does not, **extend the existing rule** to include `.nav-tab` (single-character edit to the selector list) — do not create a new file. The active-tab indicator (Basecoat-default underline) inherits forced-colors automatically because `border-bottom` is a UA-respecting property.

**Given** category 9 (empty / whitespace text input to derived values)
**When** the wrapper receives an empty or whitespace-only tab `label`
**Then** the wrapper renders the label span with the literal empty / whitespace content. A per-stack unit test asserts `label=""` and `label="   "` do not crash. The wrapper does not derive — same disposition as Story 2.6 §AC8.

**Given** categories 1 (unknown enum), 2 (font load), 4 (AG Grid), 5 (stacking)
**When** I evaluate against this story's deliverables
**Then** they are **N/A**:
- **1:** No vocabulary props.
- **2:** No new font references.
- **4:** Not AG Grid.
- **5:** No queueing.

### AC8 — Security-defaults checklist coverage (per [security-defaults.md](../../docs/reference/security-defaults.md))

Walked the seven categories — only the **applicable** ones get an AC:

**Given** category 3 (allowlist validation on writes) — **adapted for output-escaping on read**
**When** the wrapper renders any string prop (`id`, `aria_label`, tab `id`, tab `label`, tab `hx_get`, tab `hx_target`)
**Then** framework default auto-escaping is in force — verified by AC7 §category-6 XSS round-trip test extended to cover `hx_get` and `hx_target` (a payload like `hx_get="javascript:alert(1)"` produces literal `hx-get="javascript:alert(1)"` in the rendered HTML, **with the `javascript:` literal preserved**; HTMX's own URL validation is what rejects bad schemes at request-execution time, not the wrapper's escaping). The wrapper does NOT validate that `hx_get` is a relative URL — that is the consumer's responsibility (and HTMX defaults).

**Given** category 6 (CSRF posture)
**When** the wrapper emits `hx-get` on each tab
**Then** GET requests do not require CSRF tokens in any of the three stacks. The wrapper does not need to thread a CSRF token through. (If a future consumer ever switches to `hx-post` for tab-state mutations — extremely unlikely; tabs are read-only navigation — the consumer is responsible for the token. The wrapper accepts `hx_get` only, not arbitrary `hx_method`; this is by design.)

**Given** categories 1 (open-redirect), 2 (cookie attributes), 4 (dynamic RegExp), 5 (filesystem writes), 7 (stub-auth warnings)
**When** I evaluate against this story's deliverables
**Then** they are **N/A** — no redirects, no cookies, no regex on user input, no filesystem writes (the `pnpm run build` build is unchanged — no new CSS rules this story), no auth changes.

### AC9 — Cross-stack architecture principle three-deliverable check (root [CLAUDE.md](../../CLAUDE.md))

This story introduces **two cross-stack contract surfaces**:

1. The TabStrip markup shape (the canonical `<nav role="tablist" data-tabstrip>` + roving-tabindex pattern).
2. The arrow-key JS contract (the vendored `tabstrip.js` and the `data-tabstrip` selector hook).

Both produce all three deliverables:

| Deliverable | TabStrip markup | Arrow-key JS |
|---|---|---|
| Documentation contract | `fieldmark_shared/components/tab_strip/README.md` + Component Index row | README §JS-budget item + cross-references to `theme-toggle.js` vendoring precedent + `fieldmark_shared/CLAUDE.md` symlink-table update |
| Native implementation per stack | Three wrapper templates (AC5) | Single shared file (`tabstrip.js`) symlinked into three stacks + one `<script>` tag per stack's base layout |
| Per-stack conformance test | Snapshot tests over seven variants + tabindex / aria-selected unit tests | Playwright behavior test (AC6) — single host stack covers the cross-stack invariant because the script is symlinked, not duplicated |

**And** the vendored `tabstrip.js` is the **second** shared JS dependency in `fieldmark_shared/vendor/` after `theme-toggle.js`. This continues the established pattern, does not introduce a new pattern, and does not violate the Cross-Stack Architecture Principle — vendor JS symlinks are an explicit exception to "no shared code across stacks" per [fieldmark_shared/CLAUDE.md](../../fieldmark_shared/CLAUDE.md) §"Purpose" (vendor JS lives in `fieldmark_shared/vendor/` and is symlinked into each stack's static directory).

**And** there is **no new file** in `fieldmark_shared/src/` (no new CSS this story — Basecoat's existing `.nav-tabs` / `.nav-tab` classes carry the visuals; verify by inspecting `dist/fieldmark.css` before authoring the canonical).

### AC10 — `make parity` clean, no new production routes introduced

**Given** all wrappers + the vendored JS + the snapshot tests + the Playwright test land
**When** I run `make parity` from the repo root
**Then** route diff equals the Story 2.6 baseline. The Playwright fixture page reuses the Story 2.6 / 2.4 debug-gated test-render endpoint (no new route). `pg_indexes` diff: zero (no DB changes).

### AC11 — Build, type, lint, and test gates green on every stack

- **.NET:** `cd FieldMark && dotnet csharpier check . && dotnet build && dotnet test && dotnet test FieldMark.Tests.Integration/FieldMark.Tests.Integration.csproj` — clean. Snapshot tests pass with one `[Theory]` row per variant; tabindex / aria-selected unit tests pass.
- **Django:** `cd fieldmark_py && uv run ruff check . && uv run mypy . && uv run pytest && uv run pytest -m integration` — clean.
- **Go:** `cd fieldmark-go && make check && go test ./... && go test -tags=integration ./...` — clean. The `tabTabindex` template function is registered and exercised.
- **`fieldmark_shared`:** `cd fieldmark_shared && pnpm install && pnpm run build` — clean. **No** changes to `src/` this story, so `dist/fieldmark.css` should be unchanged byte-for-byte (verify with `git diff dist/fieldmark.css` — empty diff is the success criterion; if `pnpm run build` rewrites `dist/fieldmark.css` despite no source change, investigate and fix before commit). The vendored `tabstrip.js` is committed but is not built — pure source.
- **E2E:** the new Playwright spec (AC6) passes against the chosen host stack. JS-disabled progressive-enhancement spec (AC7 §category-3) passes. The other two stacks consume the same symlinked `tabstrip.js` so testing one host stack covers the cross-stack invariant.
- From repo root: `make parity` exits 0 (AC10) and `make test-all` exits 0.

## Tasks / Subtasks

- [ ] **Task 1: Author canonical example + README in `fieldmark_shared/`** (AC: #1, #2, #3, #9)
  - [ ] 1.1 Inspect `dist/fieldmark.css` for `.nav-tabs` / `.nav-tab` class names — record actual Basecoat 0.3.11 class vocabulary before authoring `canonical.html`. If different from what AC2 expects, adjust the AC and the canonical fixture (do not invent class names that don't exist in the compiled bundle).
  - [ ] 1.2 Create `fieldmark_shared/components/tab_strip/canonical.html` with seven variant blocks per AC1 (`project-detail-four-tabs-summary-active`, `project-detail-four-tabs-violations-active`, `project-detail-four-tabs-with-badges`, `two-tabs-minimal`, `single-tab`, `badge-zero`, `badge-large`).
  - [ ] 1.3 Create `fieldmark_shared/components/tab_strip/README.md` per AC1 §contract-fixed-order — ten sections.
  - [ ] 1.4 Append one row to `docs/reference/component-canonical-examples.md` Component Index.

- [ ] **Task 2: Vendor `tabstrip.js` + symlinks + base-layout script tags** (AC: #4, #9, #11)
  - [ ] 2.1 Create `fieldmark_shared/vendor/tabstrip/tabstrip.js` with the IIFE per AC4 §behavior-contract and §script-shape-reference. ES5-only; ≤ 20 LOC non-blank.
  - [ ] 2.2 Create the three symlinks per AC4 §symlink-table:
        - `FieldMark/FieldMark.Web/wwwroot/vendor/tabstrip` → `../../../../fieldmark_shared/vendor/tabstrip`
        - `fieldmark_py/static/vendor/tabstrip` → `../../../fieldmark_shared/vendor/tabstrip`
        - `fieldmark-go/internal/web/static/vendor/tabstrip` → `../../../../../fieldmark_shared/vendor/tabstrip`
  - [ ] 2.3 Add `<script src="/vendor/tabstrip/tabstrip.js" defer></script>` to each stack's base layout file (.NET `_Layout.cshtml`, Django `base.html`, Go `base.html` — locate by grepping for the existing `theme-toggle.js` tag and add immediately alongside).
  - [ ] 2.4 Update [fieldmark_shared/CLAUDE.md](../../fieldmark_shared/CLAUDE.md) §"How the Vendor JS Symlinks Work" — append the `tabstrip` row / column.
  - [ ] 2.5 Update the JS budget table in [ux-design-specification.md §"JS budget per component"](../../_bmad-output/planning-artifacts/ux-design-specification.md) ONLY if the actual line count differs from "~15" — record the final count in the wrapper README's JS budget line; do NOT modify the UX spec this story (the spec said "~15", which is a soft target; the README is authoritative for the actual count).

- [ ] **Task 3: .NET wrapper + tests** (AC: #2, #3, #5, #7, #8, #11)
  - [ ] 3.1 Create `FieldMark/FieldMark.Web/Pages/Shared/Components/_TabStrip.cshtml` with in-file `TabStripViewModel` + `TabSpec` records per AC5. Compute roving tabindex inline via projection or helper static method.
  - [ ] 3.2 Top-of-file comment references `docs/reference/component-canonical-examples.md`.
  - [ ] 3.3 Create `FieldMark/FieldMark.Tests.Web/Components/TabStripSnapshotTests.cs` (or Integration per Story 2.4 host decision). One `[Theory]` row per variant.
  - [ ] 3.4 Create `TabStripBehaviorTests.cs` — unit tests for: tabindex distribution at active_index=0 / 2 / N-1, aria-selected distribution, badge rendering (5 cases: count=12, 0, null, 9999, -1), required-prop missing (assert clear error / throw), `type="button"` present on every button.
  - [ ] 3.5 XSS round-trip tests for `label`, `aria_label`, `hx_get`, `hx_target`.
  - [ ] 3.6 Grep guard — assert zero occurrences of `Html.Raw` in `_TabStrip.cshtml`.
  - [ ] 3.7 Run `dotnet csharpier check . && dotnet build && dotnet test && dotnet test FieldMark.Tests.Integration/` — clean.

- [ ] **Task 4: Django wrapper + tests** (AC: #2, #3, #5, #7, #8, #11)
  - [ ] 4.1 Create `fieldmark_py/templates/components/_tab_strip.html` per AC5. Body iterates with `{% for tab in tabs %}`; uses `forloop.counter0` for active-index comparison.
  - [ ] 4.2 Top-of-file comment references the canonical-examples doc.
  - [ ] 4.3 Create `fieldmark_py/components/tests/test_tab_strip_snapshot.py` — `@pytest.mark.parametrize` over seven variants.
  - [ ] 4.4 Create `fieldmark_py/components/tests/test_tab_strip_behavior.py` — same five-case badge + tabindex / aria-selected + required-prop tests as AC5 §.NET.
  - [ ] 4.5 XSS round-trip tests.
  - [ ] 4.6 Grep guard — assert zero `|safe` in `_tab_strip.html`.
  - [ ] 4.7 Run `uv run ruff check . && uv run mypy . && uv run pytest && uv run pytest -m integration` — clean.

- [ ] **Task 5: Go wrapper + tests** (AC: #2, #3, #5, #7, #8, #11)
  - [ ] 5.1 Create `fieldmark-go/internal/web/templates/components/tab_strip.go` with `TabStripArgs` + `TabSpec` + `func tabTabindex(activeIndex, i int) string`. Register `tabTabindex` in the template function map alongside the `safeHTML` registration from Story 2.6.
  - [ ] 5.2 Create `fieldmark-go/internal/web/templates/components/tab_strip.html` (`{{define "tab_strip"}}…{{end}}`). Uses `{{tabTabindex $.ActiveIndex $i}}` and `{{if .BadgeCount}}…{{end}}` for the badge conditional.
  - [ ] 5.3 Top-of-file comments reference the canonical-examples doc.
  - [ ] 5.4 Create `tab_strip_test.go` — snapshot table-driven tests + behavior tests (tabindex / aria-selected / badge / required-prop / button-type).
  - [ ] 5.5 XSS round-trip tests for all string props.
  - [ ] 5.6 Grep guard — assert zero `template.HTML(` in `tab_strip.go` and `tab_strip.html`. (The args struct does not need `template.HTML` fields — all values are plain `string` / `int`.)
  - [ ] 5.7 Run `make check && go test ./... && go test -tags=integration ./...` — clean.

- [ ] **Task 6: Playwright behavior test** (AC: #6, #7-cat3, #11)
  - [ ] 6.1 Locate the e2e suite directory (verify per Story 2.6 resolution).
  - [ ] 6.2 Author the fixture page using the Story 2.4 / 2.6 debug-gated test-render endpoint. The fixture renders the `project-detail-four-tabs-summary-active` variant (plus the `single-tab` variant for case #8 — either as a separate fixture page or a query-string variant selector on the existing endpoint).
  - [ ] 6.3 Create `e2e/tests/shared/tabstrip-keyboard-navigation.spec.ts` with the nine test cases per AC6. Top-of-file comment cites UX spec.
  - [ ] 6.4 Create a parallel `e2e/tests/shared/tabstrip-no-js.spec.ts` (or extend the existing Story 1.14 sidebar-no-js spec) — Playwright with `javaScriptEnabled: false`, asserts AC7 §category-3 behavior (all tabs visible, mouse click fires navigation, no layout shift, no console errors).
  - [ ] 6.5 Run both specs against the host stack. All cases pass.

- [ ] **Task 7: Cross-stack verification + parity** (AC: #9, #10, #11)
  - [ ] 7.1 Run `make parity` — route diff equals Story 2.6 baseline.
  - [ ] 7.2 Run `make test-all` — green.
  - [ ] 7.3 Confirm each wrapper file's top-of-file comment references the canonical-examples doc.
  - [ ] 7.4 Verify the Component Index row for TabStrip is correctly populated.
  - [ ] 7.5 Verify all three symlinks resolve and the base-layout `<script>` tags load `tabstrip.js` over HTTP at `/vendor/tabstrip/tabstrip.js` in each stack (smoke test in dev: open a dev-server page, check Network panel for the 200 response on the JS file).
  - [ ] 7.6 Confirm `git diff fieldmark_shared/dist/fieldmark.css` is empty after `pnpm run build` — this story makes no CSS changes; the compiled bundle must be byte-identical.

- [ ] **Task 8: Story sign-off** (AC: all)
  - [ ] 8.1 Populate the Sign-off block below; flip sprint-status to `review`.

## Dev Notes

### Critical context (read before writing code)

- **The wrapper is markup-only; the JS is shared.** This is the first Epic 2 component where the markup-only wrapper *pairs* with a small shared JS file. The pattern is established by Story 1.6's `theme-toggle.js` — copy that posture exactly: ES5 IIFE, vendored under `fieldmark_shared/vendor/`, symlinked into all three stacks, loaded via a single base-layout `<script defer>` tag. **Do not** put the keyboard logic in each stack's template; **do not** transpile or build the JS; **do not** introduce a JS module system. If you find yourself reaching for `import` or `let` / `const`, you are deviating from the established posture — stop and re-read the theme-toggle precedent.
- **The roving tabindex pattern is non-negotiable.** WAI-ARIA tablist requires exactly one tab in the natural tab order (`tabindex="0"`) and the others out (`tabindex="-1"`). Arrow keys move focus among the others; Tab moves focus *out* of the tablist entirely (to the next focusable element, which is the panel root or whatever follows). A common bug is to give all tabs `tabindex="0"` (then Tab navigates *between tabs*, which is wrong and prolongs keyboard nav) or to give all tabs `tabindex="-1"` (then Tab skips the tablist entirely, which is also wrong). The wrapper computes this from `active_index`; the JS updates it on arrow-key focus moves. Both ends of the contract must be correct — a unit test asserts the wrapper-side distribution; the Playwright test asserts the JS-side distribution after arrow keys.
- **The OOB-swap target is the entire `<nav>`, not individual tabs.** When a consumer's response updates the active tab after a click, the canonical approach is to re-render the whole tablist with the new `aria-selected` / `tabindex` distribution and OOB-swap it via `hx-swap-oob='outerHTML:#project-detail-tabstrip'`. This requires the `<nav>` to have an `id` — see AC2 §key-invariants. This story does not implement the server-response composition (that's Story 2.11), but it does **document** the contract surface in the README so Story 2.11 has a clear target.
- **`data-tabstrip` is the boolean selector hook.** The JS attaches to `nav[data-tabstrip]`, not to a class — classes are styling concerns, data attributes are behavior hooks. This decouples the JS from any future Basecoat class-name change. If Basecoat 0.4 renames `.nav-tabs` to `.tablist`, the JS keeps working because it targets `data-tabstrip`. The same posture is used in Story 1.14's sidebar (`[data-sidebar-initialized]` is the JS-engaged marker).
- **`htmx:afterSwap` re-attachment must be idempotent.** Without idempotency, every HTMX swap binds another `keydown` listener to the new `<nav>`, and after a few tab clicks the user is firing N click events per Enter press. The reference implementation uses a `_tabstripBound` property; a more elegant approach is `WeakSet` (ES6, which we cannot use in ES5) or `Map` (ES6 likewise). Stick with the `_*` property convention from the theme-toggle precedent — it's not pretty, but it's ES5 and observable in dev tools.
- **`htmx:afterSwap` fires even for non-tabstrip swaps.** Every HTMX swap in the app triggers the script's `scan()`. This is intentional and cheap — `querySelectorAll('nav[data-tabstrip]')` against a typical page is a microsecond operation. Do not try to optimize by checking the event target — it changes the failure mode (if the OOB-swap target is not the event's direct target, the optimization misses it). Trust `scan()` to be a no-op when there's nothing new to attach (the `_tabstripBound` guard handles that).
- **Sequencing with prior stories.** This story uses the per-component directory convention, the variant-delimiter parser, the snapshot-test harness, the path-walker, the `Pages/Shared/Components/` Razor sub-directory, the Story 2.4 partial-render scaffold endpoint, the Story 2.5 Go sibling-`.go`-file pattern, the Story 2.6 template-function-map registration site, and the Story 2.6 e2e fixture-page mechanism. **Block on those** — if any primitive is missing at implementation time, flag it as a prior-story review patch, not a re-implementation here.
- **Basecoat class verification first.** AC2's canonical markup assumes `class="nav-tabs"` on the outer `<nav>` and `class="nav-tab"` on each button. **Before authoring `canonical.html`, grep `fieldmark_shared/dist/fieldmark.css` for these selectors** — confirm they exist in Basecoat 0.3.11. If Basecoat uses different names (e.g., `tab-list` / `tab`), update both the AC text in this file and the canonical markup. Do not invent class names that the compiled bundle doesn't carry. Update the AC inline with a note ("Basecoat 0.3.11 uses `<actual-class>` — adjusted from initial spec").

### Component-specific notes

- **Badge semantic monoculture.** The badge `aria-label="<count> unread"` assumes the MVP convention that all tab-strip badges are unread counts (Inspections badge = unread inspections; Violations badge = unread violations). This is true for the four Project Detail tabs in scope. **If a future consumer wants a different semantic** (e.g., a "high priority count" badge), the wrapper needs an additional `badge_aria_template` prop — but that's deferred. Record a deferred-work entry: "TabStrip future: support non-unread badge semantics via per-tab `badge_aria_template` prop. Currently hard-coded to '<count> unread'." Append to `_bmad-output/implementation-artifacts/deferred-work.md` at sign-off.
- **No pluralization of "unread".** `aria-label="1 unread"` is grammatically correct (treating "unread" as a noun-phrase elision of "unread items"). `aria-label="1 unreads"` would be wrong. This is the WCAG-compliant minimum; full English pluralization is i18n work (out of scope per PRD).
- **No keyboard behaviors beyond arrow / Home / End / Enter / Space.** The script does NOT bind Escape, Tab, or any other key. Tab leaves the tablist via browser default (the next focusable element is whatever comes after the `</nav>` in DOM order — for Project Detail that's the panel content; for the `single-tab` variant the next focusable depends on the page layout). The Playwright test verifies this behavior in case #1 (initial focus order).
- **No `aria-orientation` on the tablist.** ARIA 1.2 says tablists default to `aria-orientation="horizontal"` for `<nav role="tablist">`; horizontal is the FieldMark choice. Explicit `aria-orientation="horizontal"` would be redundant but harmless — *do not add it* (minimal-markup posture); a per-stack test could assert its absence if a future linter wants to enforce. Vertical tablists are out of scope.
- **The Inspections / Violations / Audit tabs' badge counts come from the server.** Story 2.11 wires the actual counts. This story just renders whatever count the consumer passes. The badge is **not** updated live via OOB swap from other actions — the count is only refreshed when the page re-renders or a tab swap response includes the OOB-updated tablist. This is consistent with the canonical-stale-acceptable posture (UX line 1041 — rail is independent; the strip is updated only on the OOB swap responses that consumers compose).
- **Why `<button>` not `<a>`.** Repeated for emphasis (this is a common review item): tabs are not navigation in the page-load sense — they are HTMX swaps. `<a href>` would tempt browsers into following the href when JS / HTMX fails; `<button>` produces no fallback navigation, which is the desired "do nothing if JS is dead" behavior (the page is still readable; the active tab's content is still rendered server-side as the initial state). If a future consumer ever wants tab links to be deep-linkable URLs, that's an `hx-push-url` story, not a `<a>` story.

### Edge cases (per [component-edge-case-checklist.md](../../docs/reference/component-edge-case-checklist.md))

Walked the nine categories — see AC7. Categories **3**, **6**, **7**, **8**, **9** apply; **1**, **2**, **4**, **5** are N/A.

### Security defaults (per [security-defaults.md](../../docs/reference/security-defaults.md))

Walked the seven categories — see AC8. Category **3** applies (output-escape on read); **6** is addressed by design (GET-only — no CSRF concern); **1**, **2**, **4**, **5**, **7** are N/A.

### Cross-stack contract three-deliverable check

Two contract surfaces (markup shape + JS contract) — both have all three deliverables per AC9. See AC9 table.

### Files this story modifies vs creates

| File | New / Modified | Purpose |
|---|---|---|
| `fieldmark_shared/components/tab_strip/canonical.html` | NEW | seven variant blocks |
| `fieldmark_shared/components/tab_strip/README.md` | NEW | contract |
| `fieldmark_shared/vendor/tabstrip/tabstrip.js` | NEW | ≤20-LOC ES5 IIFE |
| `FieldMark/FieldMark.Web/wwwroot/vendor/tabstrip` | NEW (symlink) | → `../../../../fieldmark_shared/vendor/tabstrip` |
| `fieldmark_py/static/vendor/tabstrip` | NEW (symlink) | → `../../../fieldmark_shared/vendor/tabstrip` |
| `fieldmark-go/internal/web/static/vendor/tabstrip` | NEW (symlink) | → `../../../../../fieldmark_shared/vendor/tabstrip` |
| `FieldMark/FieldMark.Web/Pages/Shared/_Layout.cshtml` | MODIFY | add `<script src="/vendor/tabstrip/tabstrip.js" defer></script>` |
| `fieldmark_py/templates/base.html` | MODIFY | same |
| `fieldmark-go/internal/web/templates/base.html` (or equivalent Story 1.5 base path) | MODIFY | same |
| `fieldmark_shared/CLAUDE.md` | MODIFY | append `tabstrip` to vendor-symlink table |
| `docs/reference/component-canonical-examples.md` | MODIFY | append one row to Component Index |
| `FieldMark/FieldMark.Web/Pages/Shared/Components/_TabStrip.cshtml` | NEW | wrapper |
| `FieldMark/FieldMark.Tests.{Web,Integration}/Components/TabStripSnapshotTests.cs` | NEW | snapshot tests |
| `FieldMark/FieldMark.Tests.{Web,Integration}/Components/TabStripBehaviorTests.cs` | NEW | tabindex / aria / badge / required-prop tests |
| `fieldmark_py/templates/components/_tab_strip.html` | NEW | wrapper |
| `fieldmark_py/components/tests/test_tab_strip_snapshot.py` | NEW | snapshot tests |
| `fieldmark_py/components/tests/test_tab_strip_behavior.py` | NEW | behavior tests |
| `fieldmark-go/internal/web/templates/components/tab_strip.go` | NEW | args + `tabTabindex` helper |
| `fieldmark-go/internal/web/templates/components/tab_strip.html` | NEW | `{{define "tab_strip"}}` wrapper |
| `fieldmark-go/internal/web/templates/components/tab_strip_test.go` | NEW | snapshot + behavior tests |
| `fieldmark-go/internal/web/templates/templates.go` (or function-map registration site) | MODIFY | register `tabTabindex` template function |
| `e2e/tests/shared/tabstrip-keyboard-navigation.spec.ts` | NEW | nine-case Playwright spec |
| `e2e/tests/shared/tabstrip-no-js.spec.ts` | NEW (or extend existing sidebar-no-js spec) | JS-disabled progressive-enhancement spec |
| `_bmad-output/implementation-artifacts/deferred-work.md` | MODIFY | new entry: "Story 2.7-followup — TabStrip badge semantic monoculture; add `badge_aria_template` prop when a non-unread-count consumer lands. Currently hard-coded to '<count> unread'." |

Anything outside this list — Project Detail page, the OOB-swap server response composition, the tab-panel `<div role="tabpanel">` markup, AG Grid, route registration, any DB change, CSS edits — is out of scope. Resist the urge.

### Files to read fully before editing

- [_bmad-output/planning-artifacts/ux-design-specification.md:910–916](../planning-artifacts/ux-design-specification.md) — TabStrip UX-DR spec.
- [_bmad-output/planning-artifacts/ux-design-specification.md:1032–1047](../planning-artifacts/ux-design-specification.md) — UX Pattern 5 (Anchor Screen With HTMX Tabs); binding for the OOB-swap contract.
- [_bmad-output/planning-artifacts/ux-design-specification.md:932–939](../planning-artifacts/ux-design-specification.md) — JS budget per component (TabStrip ~15 LOC).
- [_bmad-output/planning-artifacts/ux-design-specification.md:1216–1230](../planning-artifacts/ux-design-specification.md) — landmark structure + focus management + live-region politeness (binding for AC2).
- [_bmad-output/planning-artifacts/ux-design-specification.md:1252](../planning-artifacts/ux-design-specification.md) — "no client-side a11y tests" exception list (TabStrip arrow-key navigation is named).
- [_bmad-output/planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md](../planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md) §Story 2.7 — epic AC source.
- [docs/reference/component-edge-case-checklist.md](../../docs/reference/component-edge-case-checklist.md) — nine-category walkthrough; binding for AC7.
- [docs/reference/security-defaults.md](../../docs/reference/security-defaults.md) — seven-category walkthrough; binding for AC8.
- [fieldmark_shared/CLAUDE.md](../../fieldmark_shared/CLAUDE.md) §"How the Vendor JS Symlinks Work" — vendor-symlink precedent; §"Snapshot-test pipeline" — normalization. Read fully; this story extends both.
- [fieldmark_shared/vendor/theme-toggle/theme-toggle.js](../../fieldmark_shared/vendor/theme-toggle/theme-toggle.js) — **the ES5 IIFE precedent** — copy the posture exactly.
- [fieldmark_shared/dist/fieldmark.css](../../fieldmark_shared/dist/fieldmark.css) — search for `nav-tabs` / `nav-tab` (or whatever Basecoat 0.3.11 actually uses) to verify class names before authoring `canonical.html`.
- [_bmad-output/implementation-artifacts/2-4-implement-phase-2-markup-only-components-statusbadge-inlinealert-auditrow-dashboardtile.md](2-4-implement-phase-2-markup-only-components-statusbadge-inlinealert-auditrow-dashboardtile.md) — per-component-directory convention, snapshot harness, host decision.
- [_bmad-output/implementation-artifacts/2-5-implement-compliancetile-component-and-compliance-tile-oob-target.md](2-5-implement-compliancetile-component-and-compliance-tile-oob-target.md) — Go sibling-`.go`-file precedent.
- [_bmad-output/implementation-artifacts/2-6-implement-entityrail-component-with-responsive-collapse.md](2-6-implement-entityrail-component-with-responsive-collapse.md) — Go template-function-map registration site (`safeHTML`), e2e suite path resolution, Playwright fixture-page mechanism.
- WAI-ARIA Authoring Practices "Tabs" pattern — referenced in AC6 §test-citations but not a code dependency. Apply the roving-tabindex + arrow-key + Home/End conventions documented there.
- Stack rules: [FieldMark/CLAUDE.md](../../FieldMark/CLAUDE.md), [fieldmark_py/CLAUDE.md](../../fieldmark_py/CLAUDE.md), [fieldmark-go/CLAUDE.md](../../fieldmark-go/CLAUDE.md).
- Root cross-stack invariants: [CLAUDE.md](../../CLAUDE.md) §Cross-Stack Architecture Principle.

### Project Structure Notes

- The Razor `Pages/Shared/Components/` sub-directory exists (created by Story 2.4).
- The Django `templates/components/` and `components/tests/` directories exist (Stories 2.4 / 2.5 / 2.6).
- The Go `internal/web/templates/components/` directory exists; the function-map registration site was located by Story 2.6 (`safeHTML`). Add `tabTabindex` to the same registration.
- The e2e suite location was resolved by Story 2.6. Reuse.
- The base layout files were introduced by Story 1.5. Search by grepping for the existing `theme-toggle.js` script tag — the new TabStrip tag sits in the same vicinity.

### References

- AC source: [_bmad-output/planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md](../planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md) §Story 2.7
- UX-DR TabStrip spec: [ux-design-specification.md:910–916](../planning-artifacts/ux-design-specification.md)
- UX Pattern 5: [ux-design-specification.md:1032–1047](../planning-artifacts/ux-design-specification.md)
- JS budget per component: [ux-design-specification.md:932–939](../planning-artifacts/ux-design-specification.md)
- Focus management on HTMX swaps: [ux-design-specification.md:1218–1221](../planning-artifacts/ux-design-specification.md)
- "No client-side a11y tests" exceptions: [ux-design-specification.md:1252](../planning-artifacts/ux-design-specification.md)
- Component edge-case checklist: [docs/reference/component-edge-case-checklist.md](../../docs/reference/component-edge-case-checklist.md)
- Security defaults checklist: [docs/reference/security-defaults.md](../../docs/reference/security-defaults.md)
- Snapshot-test pipeline: [fieldmark_shared/CLAUDE.md](../../fieldmark_shared/CLAUDE.md) §"Snapshot-test pipeline"
- Vendor-symlink precedent: [fieldmark_shared/CLAUDE.md](../../fieldmark_shared/CLAUDE.md) §"How the Vendor JS Symlinks Work"
- ES5 IIFE precedent: [fieldmark_shared/vendor/theme-toggle/theme-toggle.js](../../fieldmark_shared/vendor/theme-toggle/theme-toggle.js)
- Per-component-directory convention: [Story 2.4](2-4-implement-phase-2-markup-only-components-statusbadge-inlinealert-auditrow-dashboardtile.md)
- Go sibling-`.go`-file precedent: [Story 2.5](2-5-implement-compliancetile-component-and-compliance-tile-oob-target.md)
- Go template-function-map site, e2e fixture-page: [Story 2.6](2-6-implement-entityrail-component-with-responsive-collapse.md)
- Cross-Stack Architecture Principle: root [CLAUDE.md](../../CLAUDE.md)
- Stack rules: [FieldMark/CLAUDE.md](../../FieldMark/CLAUDE.md), [fieldmark_py/CLAUDE.md](../../fieldmark_py/CLAUDE.md), [fieldmark-go/CLAUDE.md](../../fieldmark-go/CLAUDE.md)
- WAI-ARIA Authoring Practices "Tabs" pattern (documentation reference, not code dependency)

## Dev Agent Record

### Agent Model Used

_to be populated by dev-story_

### Debug Log References

### Completion Notes List

### File List

## Sign-off

| Field | Value |
|---|---|
| Final review date | _pending_ |
| Total review rounds | 0 |
| Final reviewer verdict | _pending — story created, status `ready-for-dev`_ |
| Deferred-work entries | _one new — "Story 2.7-followup — TabStrip badge semantic monoculture; add `badge_aria_template` prop when a non-unread-count consumer lands." Per Dev Notes §"Badge semantic monoculture"._ |
| Dev-notes divergences from epic AC | The epic AC says the keyboard JS is "≤ 15 LOC"; this story raises the ceiling to **≤ 20 LOC** to accommodate the `htmx:afterSwap` re-attachment hook (~3 extra lines beyond the bare arrow-key handler), the idempotency guard (~1 line), and the Home/End keys (~2 lines beyond Left/Right alone). The 20-line ceiling is recorded in AC4 and AC11; the wrapper README records the actual final count. Rationale: 15 LOC without the re-attachment is broken-in-practice (arrow keys stop working after the first tab click); shipping the broken-in-practice version to meet a soft AC number would be a regression. The epic AC is honored in spirit (the JS is tiny and vendored); the byte ceiling is +5 over the literal text. |

### Review Findings

_to be populated by code-review_
