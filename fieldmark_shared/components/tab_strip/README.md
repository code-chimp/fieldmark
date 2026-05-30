# TabStrip Component

## 1. Purpose

Horizontal `<nav role='tablist'>` with HTMX-bound `<button role='tab'>` tabs and optional unread-count badges; pairs with the vendored `tabstrip.js` arrow-key navigation script for full WAI-ARIA tablist behavior.

## 2. Required Props

| Prop | Type | Required | Default | Notes |
|---|---|---|---|---|
| `id` | `string` | No | none | Stable `id` on the outer `<nav>`; required for OOB-swap (see §9) |
| `aria_label` | `string` | **Yes** | — | Names the tablist; missing value → error or panic per stack |
| `tabs` | `list[TabSpec]` | **Yes** | — | Ordered tab definitions |
| `active_index` | `int` | **Yes** | — | Zero-based index of the initially-active tab |
| `hx_swap` | `string` | No | `"innerHTML"` | Applied to every tab's `hx-swap` attribute |

**TabSpec** fields:

| Field | Type | Required | Notes |
|---|---|---|---|
| `id` | `string` | Yes | Tab button `id`; also used in `aria-controls` |
| `label` | `string` | Yes | Visible tab label text |
| `hx_get` | `string` | Yes | HTMX request URL |
| `hx_target` | `string` | Yes | HTMX swap target selector (e.g. `"#project-detail-tab-content"`) |
| `badge_count` | `int?` | No | Unread count; `null`/`None` → no badge rendered |

## 3. Variant List

1. `project-detail-four-tabs-summary-active` — four Project Detail tabs; Summary active; no badges
2. `project-detail-four-tabs-violations-active` — same four tabs; Violations (index 2) active; no badges
3. `project-detail-four-tabs-with-badges` — four tabs; Summary active; Inspections=12, Violations=3, Audit=147
4. `two-tabs-minimal` — two tabs (Open / Closed); no badges; establishes contract for arbitrary consumers
5. `single-tab` — edge case: one tab; no badge; arrow keys are a no-op
6. `badge-zero` — four tabs; Violations badge_count=0; badge MUST render with text `0`
7. `badge-large` — four tabs; Inspections badge_count=9999; no truncation markup

## 4. ARIA Invariants

- Outer `<nav role="tablist" aria-label="…" data-tabstrip>` — `data-tabstrip` is the JS selector hook (boolean attribute, no value)
- Per tab: `role="tab"`, `aria-selected="true"|"false"`, `aria-controls="<panel-id>"`, `id="<tab-id>"`
- **Roving tabindex**: active tab `tabindex="0"`; all others `tabindex="-1"` — only one tab in the natural tab order at a time; arrow keys move focus among the rest
- The panel (`role="tabpanel"`) is the **consumer's** responsibility (not this wrapper); the panel must carry `aria-labelledby="<tab-id>"` per UX-DR33

## 5. HTMX Attributes per Tab

- `hx-get="<url>"` — fires the request on click
- `hx-target="<panel-selector>"` — swap destination (same for all tabs in the tablist)
- `hx-swap="<swap>"` — defaults to `innerHTML`
- `hx-disabled-elt="this"` is **NOT** applied — tabs are navigation triggers, not state-mutating actions; the `htmx-request` class (applied by HTMX itself) drives latency indication via Basecoat's existing styles

## 6. Badge Contract

Badge renders as `<span class="badge tab-strip__badge" aria-label="<count> unread"><count></span>` when `badge_count` is not null.

- `aria-label` uses the literal template `"<count> unread"` (singular regardless of count — deliberately not pluralized; see Dev Notes §"Badge semantic monoculture")
- Badge text is the literal stringified value — `0` renders as `0`, `9999` renders as `9999`; no truncation, no formatting
- When `badge_count` is null/None, no `<span class="tab-strip__badge">` is emitted — not even an empty span

## 7. Allowed Class Vocabulary

BEM classes on this component:

| Class | Element |
|---|---|
| `tab-strip` | Outer `<nav>` |
| `tab-strip__tab` | Each `<button role="tab">` |
| `tab-strip__label` | Label `<span>` inside each button |
| `tab-strip__badge` | Badge `<span>` inside each button (when badge_count is not null) |

Basecoat visual styling note: Basecoat 0.3.11 uses `.tabs` as an outer container class that wraps BOTH the `[role="tablist"]` AND the `[role="tabpanel"]` (descendant selectors). The TabStrip wrapper renders only the `<nav role="tablist">` portion. The consumer (e.g. Story 2.11 Project Detail) provides the `.tabs` outer wrapper around both the TabStrip and the panel content. Accordingly this wrapper uses BEM-only classes; Basecoat's `.nav-tabs`/`.nav-tab` class names assumed in the initial story AC do not exist in Basecoat 0.3.11 (AC updated inline in story file).

## 8. JS Budget

Pairs with `fieldmark_shared/vendor/tabstrip/tabstrip.js` (24 non-blank LOC; the ≤ 20 ceiling from the initial AC was exceeded by 4 lines to accommodate the `htmx:afterSwap` re-attachment hook, the idempotency guard, the Home/End keys, and restoring the `scan()` helper function for ES5-safe NodeList iteration — same trade-off noted in the story sign-off) for arrow-key navigation. The wrapper itself is zero-JS — it emits the markup that the script attaches to via `[data-tabstrip]` selector. The script is symlinked into each stack's static vendor directory and loaded via a `<script src='/vendor/tabstrip/tabstrip.js' defer>` tag in the base layout.

## 9. OOB-swap Contract for Active-Tab Updates

When a tab is clicked, the consumer's response replaces `#project-detail-tab-content` (the `hx-target`) **and** emits an OOB swap targeting the tablist that re-renders the entire `<nav role='tablist'>` with the new `aria-selected` flipped. The wrapper's canonical HTML is the OOB target — Story 2.11 consumes this contract by emitting an `hx-swap-oob='outerHTML'` block whose root is `<nav id='project-detail-tabstrip' role='tablist'>...</nav>`. Adding an `id` to the outer `<nav>` is part of the loaded variants — see §4.

## 10. Snapshot-equality Requirement

Per-stack snapshot tests assert byte-equality between the rendered wrapper output (whitespace-normalised, attributes alphabetised) and the corresponding variant block in `canonical.html`. The normaliser must treat `data-tabstrip` and `data-tabstrip=""` as equivalent (HTML5 boolean attribute forms).
