# Story 2.6: Implement EntityRail component with responsive collapse

Status: ready-for-dev

Epic: 2 — Project Lifecycle & Compliance Dashboard
Source AC: [_bmad-output/planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md](../planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md) §Story 2.6
Depends on: Story 2.4 (per-component directory convention + snapshot harness + path-walker + `Pages/Shared/Components/` sub-directory), Story 2.5 (refined band-resolver pattern is precedent for the optional-content conditional inside templates). This story extends both — adds a fifth component directory under `fieldmark_shared/components/`, lands a single new layout-CSS rule for the responsive collapse, and introduces the `<aside>` landmark wrapper that becomes the canonical HTMX target shape for `#violation-detail`, `#inspection-detail`, `#corrective-action-detail`.

## Story

As a Compliance Officer or Project Manager working on the Project Detail anchor screen (Story 2.11) at desktop, and as the developer composing the Violations / Inspections tabs (Stories 4.3 / 3.7) and the right-rail consumer pages,
I want a markup-only **EntityRail** wrapper per stack — byte-identical HTML across .NET, Django, and Go — that renders an `<aside id="<rail-id>" tabindex="-1" role="region" aria-label="<entity-type> detail" aria-live="polite">` container with three named slots (header strip, body, action footer), an empty state ("Select an entity to see its detail here."), and a single shared layout-CSS rule that makes the rail sticky on the right at ≥ 1280px and stacked beneath the list at < 1280px,
So that downstream stories (2.11 Project Detail, 3.7 Inspection Detail, 4.3 Violation Detail, 5.3 Corrective Action list inside violation detail) can compose List+Detail Co-Presence (UX Pattern 6) without inventing markup, focus management after an HTMX swap is well-defined (`tabindex="-1"` is in place), the rail survives tab-content swaps (rail is independent of `#project-detail-tab-content`), and the responsive collapse rule is centralized in `_layout.css` rather than per-stack templates.

**Scope boundary:** this story produces (a) `fieldmark_shared/components/entity_rail/canonical.html` (variant-delimited fixture, same convention Stories 2.4 / 2.5 introduce) + `README.md` (contract), (b) one markup-only wrapper template per stack in idiomatic locations, (c) per-stack snapshot tests asserting byte-equality, (d) **one new CSS rule block** in `fieldmark_shared/src/_layout.css` implementing the responsive collapse (the only `fieldmark_shared/src/` edit this story makes), (e) appended row in `docs/reference/component-canonical-examples.md` Component Index, (f) **one Playwright responsive-viewport test** in the existing e2e suite asserting the collapse rule at the three documented viewport widths (1280, 1024, 375). **Out of scope:** any consumer page that actually populates the rail (Project Detail Story 2.11, Violation Detail Story 4.3, Inspection Detail Story 3.7, CA list Story 5.3 — they will use this wrapper and supply their own partial), the focus-on-swap *invocation* mechanism (this story lands `tabindex="-1"` on the rail so focus *can* land there; the actual `.focus()` call after an HTMX swap is the consumer story's responsibility — either via `autofocus` on the inserted partial root or `HX-Trigger`-driven script per UX-DR §"Focus management on HTMX swaps" line 1219), any JavaScript (zero-JS component — the focus mechanism is consumer-owned and uses either HTML's native `autofocus` or the existing base-layout `HX-Trigger` listener landed in Story 1.5), the dismiss-× behavior wiring (the wrapper emits the button markup in the loaded variant as a slot placeholder; the actual `hx-get` to "clear the rail" is consumer-owned in Story 2.11+), and any AG Grid row-click wiring (that's Story 2.9 — the rail wrapper is the *target*, not the trigger).

## Acceptance Criteria

### AC1 — Canonical-example directory under `fieldmark_shared/components/`

**Given** the per-component-directory convention introduced by Story 2.4 and extended by Story 2.5
**When** I inspect `fieldmark_shared/components/`
**Then** a new sub-directory exists with this exact layout:

```
fieldmark_shared/components/entity_rail/
├── canonical.html
└── README.md
```

**And** `canonical.html` follows the same variant-delimited format as `action_button.example.html` (Story 1.12) — a leading HTML comment block documenting `fixture:` inputs, then `<!-- variant: <name> (inputs: ...) -->` delimiters separating each variant's expected output.

**And** `canonical.html` contains **exactly six variant blocks**:

1. `empty-violation` — `id="violation-detail", entity_type_label="Violation", entity_loaded=false` → empty-state shell with the "Select an entity to see its detail here." card.
2. `empty-inspection` — `id="inspection-detail", entity_type_label="Inspection", entity_loaded=false` → identical shell with the violation-specific aria-label replaced by `"Inspection detail"`.
3. `empty-corrective-action` — `id="corrective-action-detail", entity_type_label="Corrective Action", entity_loaded=false` → identical shell with `aria-label="Corrective Action detail"`.
4. `loaded-shell-violation` — `id="violation-detail", entity_type_label="Violation", entity_loaded=true, body_slot="__BODY__", footer_slot="__FOOTER__"` → the **loaded shell** with the three slots populated by literal sentinel strings (`__BODY__`, `__FOOTER__`). The sentinels prove the slots are wired without forcing the canonical fixture to carry real entity content. The header strip shows the entity-type label + the dismiss-× button markup (a `<button>` with `aria-label="Close <entity-type> detail"` — see AC3 §dismiss-button-shape).
5. `loaded-shell-inspection` — same as #4 with `entity_type_label="Inspection"`, `id="inspection-detail"`.
6. `loaded-shell-corrective-action` — same as #4 with `entity_type_label="Corrective Action"`, `id="corrective-action-detail"`.

**And** `README.md` documents the contract in this fixed order (matching Stories 2.4 / 2.5): (1) Purpose (one sentence — "Right-rail container for the currently-selected entity's detail; the canonical `#violation-detail` / `#inspection-detail` / `#corrective-action-detail` HTMX target shape with stable focus surface and responsive collapse"), (2) Required props with types (`id: string` — the stable HTMX target id; `entity_type_label: string` — used in `aria-label` and the loaded-shell header; `entity_loaded: bool` — selects empty vs loaded variant; `body_slot: string?` — caller-supplied inner HTML for the body slot when `entity_loaded=true`; `footer_slot: string?` — caller-supplied action-footer markup when `entity_loaded=true`), (3) Variant list (the six names), (4) ARIA invariants (`role="region"`, `aria-label="<entity-type-label> detail"` for loaded and `aria-label="Empty entity rail"` for empty; `tabindex="-1"` always; `aria-live="polite"` per UX-DR §"Live region politeness" at [ux-design-specification.md:1223](../planning-artifacts/ux-design-specification.md)), (5) Slot contract (header strip is wrapper-owned for both states — empty renders empty-state card, loaded renders entity-type label + dismiss-× button; body slot is caller-owned via `body_slot`; footer slot is caller-owned via `footer_slot` and is omitted when null), (6) Allowed class vocabulary (`entity-rail`, `entity-rail__header`, `entity-rail__dismiss`, `entity-rail__body`, `entity-rail__footer`, `entity-rail__empty` — BEM, matching Story 2.4 / 2.5 precedent; Basecoat `card` class on the empty-state inner card), (7) Snapshot-equality requirement ("Per-stack wrappers MUST render output byte-equal to the matching variant block in `canonical.html` after the standard normalization defined in `fieldmark_shared/CLAUDE.md` §'Snapshot-test pipeline'. Caller-supplied `body_slot` and `footer_slot` strings pass through verbatim — the wrapper does not escape, transform, or wrap them. The caller is responsible for the safety of HTML they place in the slots; the consumer page is the trust boundary, not the wrapper."), (8) Responsive-collapse contract (one-line: "Sticky right rail at ≥ 1280px; un-fixes and stacks at < 1280px. Rule lives in `fieldmark_shared/src/_layout.css` under `/* EntityRail responsive collapse */`."), (9) Focus-on-swap invariant (one-line: "Wrapper emits `tabindex='-1'` so consumer-supplied focus-after-swap (`autofocus` on inserted partial root, or `HX-Trigger` event handler) lands focus on the rail root. The wrapper itself does not invoke `.focus()` — that is a consumer concern.").

**And** `docs/reference/component-canonical-examples.md` (Story 2.4) is MODIFIED — one new row appended to the Component Index table with columns: `EntityRail`, fixture path, README path, three wrapper paths, three test paths.

### AC2 — EntityRail empty-state markup contract (UX-DR §"EntityRail", UX Pattern 6 — List + Detail Co-Presence, UX Pattern 7 — Empty State With Next Action)

**Given** the EntityRail contract at [ux-design-specification.md:871–877](../planning-artifacts/ux-design-specification.md) and the empty-state-with-next-action rule at [ux-design-specification.md:1066–1076](../planning-artifacts/ux-design-specification.md)
**When** I inspect `fieldmark_shared/components/entity_rail/canonical.html` for the `empty-violation` variant
**Then** the rendered HTML is **exactly** this shape (whitespace-normalized; attribute order alphabetized for the byte-equality comparison):

```html
<aside
  aria-label="Violation detail"
  aria-live="polite"
  class="entity-rail entity-rail--empty"
  id="violation-detail"
  role="region"
  tabindex="-1"
>
  <div class="card entity-rail__empty" aria-label="Empty entity rail">
    <p>Select an entity to see its detail here.</p>
  </div>
</aside>
```

**And** the three empty-state variants differ only in:
- the `id` attribute value (`violation-detail` / `inspection-detail` / `corrective-action-detail`),
- the outer `aria-label` value (`"Violation detail"` / `"Inspection detail"` / `"Corrective Action detail"`).

The inner `aria-label="Empty entity rail"` on the `<div class="card entity-rail__empty">` is **the same string** across all three empty variants — this satisfies UX-DR's instruction that the empty state itself names what is empty (UX-DR §EntityRail at line 875: "empty (\"Select an entity to see its detail here.\")") **and** the per-rail outer `aria-label` names *which* rail it is. The empty-state aria-label is **not** parameterized by entity-type — there is one empty rail; the outer landmark identifies its scope.

**And** the `<aside>` is the **only** outermost element — no wrapping `<div>`, no `<section>` parent. This is a top-level page landmark per [ux-design-specification.md:1216](../planning-artifacts/ux-design-specification.md) (UX §"Document landmark structure" — "optional `<aside>` (EntityRail)"). The wrapper renders the `<aside>` directly; the consumer page slots it into the layout grid via the responsive-collapse CSS (AC5) — no per-stack template adds a wrapping container.

**And** the empty-state `<p>` text is literal: `Select an entity to see its detail here.` (period included, en-quote-less, no `&apos;` since there are no apostrophes — the byte sequence is plain ASCII for this string).

### AC3 — EntityRail loaded-shell markup contract (slot wiring, dismiss-× button)

**Given** the EntityRail contract at [ux-design-specification.md:871–877](../planning-artifacts/ux-design-specification.md)
**When** I inspect `fieldmark_shared/components/entity_rail/canonical.html` for the `loaded-shell-violation` variant
**Then** the rendered HTML is **exactly** this shape:

```html
<aside
  aria-label="Violation detail"
  aria-live="polite"
  class="entity-rail entity-rail--loaded"
  id="violation-detail"
  role="region"
  tabindex="-1"
>
  <header class="entity-rail__header">
    <span class="entity-rail__entity-type">Violation</span>
    <button
      aria-label="Close Violation detail"
      class="entity-rail__dismiss"
      type="button"
    >×</button>
  </header>
  <div class="entity-rail__body">__BODY__</div>
  <div class="entity-rail__footer">__FOOTER__</div>
</aside>
```

**And** the three loaded-shell variants differ only in:
- the `id` attribute value,
- the outer `aria-label` value (`"<EntityType> detail"`),
- the inner `<span class="entity-rail__entity-type">` text (the entity-type label, no `text-transform: uppercase` — UX-DR specifies the header strip carries the *entity-type label*, which is the literal label string the caller passes; do not call `.upper()` / `.ToUpper()` and do not add a CSS uppercase rule this story — the header is plain text per UX line 874),
- the dismiss button's `aria-label="Close <EntityType> detail"`.

**Dismiss-button shape.** The dismiss × is a `<button type="button">` carrying:
- `class="entity-rail__dismiss"`,
- `aria-label="Close <EntityType> detail"`,
- **no** `hx-get`, `hx-post`, `hx-target`, `hx-swap`, `hx-trigger`, `onclick`, or any handler attribute.

The button is a **markup placeholder** — wiring it to actually clear the rail is the consumer story's responsibility (Story 2.11+ will add `hx-get` to a "render-empty-rail" endpoint and `hx-target="#<rail-id>" hx-swap="outerHTML"`). The wrapper does not own the dismiss behavior; it owns the dismiss *affordance*. A per-stack negative-attribute test asserts the rendered output for any loaded-shell variant does not contain any of `hx-get`, `hx-post`, `hx-target`, `hx-swap`, `hx-trigger`, `onclick=`.

**And** the visible × character is the **U+00D7 MULTIPLICATION SIGN** (UTF-8 bytes `C3 97`), not the ASCII letter `x`, not the `×` entity `&times;`, and not U+2715 / U+2716 / U+10005 (multiplication X variants). The byte sequence is part of the contract — per-stack tests assert the raw byte.

**Slot pass-through.** The `body_slot` and `footer_slot` props are rendered **verbatim** without escaping, transformation, or wrapping. The slots are placed inside `<div class="entity-rail__body">` and `<div class="entity-rail__footer">` containers; the slot content is the *children* of those `<div>`s. Per-stack rendering primitives:

- **.NET (Razor):** `@Html.Raw(Model.BodySlot ?? string.Empty)` and `@Html.Raw(Model.FooterSlot ?? string.Empty)` inside the wrapper. **This is the explicit exception to the "no `Html.Raw`" rule from Stories 2.4 / 2.5.** The trust boundary moves to the consumer page — the wrapper *must* emit caller-supplied HTML verbatim so the consumer can place a violation-detail partial inside without it being entity-encoded. The wrapper's `README.md` records this exception (AC1 §contract item 7 — "the caller is responsible for the safety of HTML they place in the slots"). The grep guard from Stories 2.4 / 2.5 that forbids `Html.Raw` in component wrappers MUST be updated to **exempt this single wrapper file** (`_EntityRail.cshtml`) and **only** the two slot-rendering positions; any other use of `Html.Raw` in this file is still forbidden. A per-stack lint asserts this scope: the file may contain exactly two `Html.Raw` occurrences, both inside `<div class="entity-rail__body">` / `<div class="entity-rail__footer">` contexts.
- **Django:** `{{ body_slot|safe }}` and `{{ footer_slot|safe }}`. Same exception: the wrapper is the slot-pass-through point, the consumer is the trust boundary. The grep guard exempts `_entity_rail.html` for these two `|safe` filters and nothing else.
- **Go:** `{{.BodySlot | safeHTML}}` and `{{.FooterSlot | safeHTML}}` where `safeHTML` is a template function that returns `template.HTML(s)`. If `safeHTML` does not yet exist as a registered template function, this story registers it in the existing template-function map (search `fieldmark-go/internal/web/templates/` for the function-map registration — likely in `templates.go` or wherever Story 1.5 / 1.12 wired the function map). The function is generic — name it `safeHTML`, document it in a one-line code comment as "Render caller-supplied HTML verbatim; intended for component-slot pass-through. Caller is the trust boundary.", and reuse it for future slotted components (TabStrip's panel slot in Story 2.7, etc.). The grep guard for `template.HTML(` is left in place for **all other files**; this single helper file is the documented exception.

The slot mechanism is the **only place** in any wrapper this story produces where caller-supplied HTML is rendered raw. The `entity_type_label` prop is still framework-escaped (it is plain text, not HTML); the `id` prop is still framework-escaped (defense in depth — `id` is expected to be a known canonical string from the UX vocabulary but the escape is cheap insurance).

**Per-stack wrapper paths** (idiomatic per stack — no shared template, no symlinked partial):

- **.NET (Razor partial):** `FieldMark/FieldMark.Web/Pages/Shared/Components/_EntityRail.cshtml` (NEW). Invoked as `<partial name="Shared/Components/_EntityRail" model="@(new EntityRailViewModel(id, entityTypeLabel, entityLoaded, bodySlot, footerSlot))" />`. View model lives in-file (record at bottom), same precedent as Stories 1.12 / 2.4 / 2.5.
- **Django (template include):** `fieldmark_py/templates/components/_entity_rail.html` (NEW). Invoked as `{% include "components/_entity_rail.html" with id="violation-detail" entity_type_label="Violation" entity_loaded=False body_slot=None footer_slot=None %}`. When the loaded variant is rendered, `entity_loaded=True` and the caller supplies pre-rendered HTML strings for the slots (typically via `{% with body_slot=violation_partial_html footer_slot=violation_actions_html %}` after rendering those partials to string).
- **Go (`html/template` `{{define}}` block):** `fieldmark-go/internal/web/templates/components/entity_rail.html` (NEW). Defines `{{define "entity_rail"}}…{{end}}`. Args struct in a sibling `entity_rail.go` file (NEW; follows Story 2.5 sibling-`.go`-file precedent): `type EntityRailArgs struct { ID, EntityTypeLabel string; EntityLoaded bool; BodySlot, FooterSlot template.HTML }` — note `template.HTML` on the slot fields is the **idiomatic Go opt-in** to slot pass-through; the caller constructs the value as `template.HTML(htmlString)` at the slot boundary. The args-struct definition is the slot-trust-boundary declaration in Go.

### AC4 — Pure empty-vs-loaded selector (no embedded conditional logic in templates beyond the variant choice)

**Given** the empty-vs-loaded conditional in the wrapper
**When** I inspect each stack's wrapper template
**Then** the **only** conditional inside the template body is a single binary branch on `entity_loaded`. There is no nested `if`/`else` inside either branch; the empty branch renders the empty-state card, the loaded branch renders the header strip + body slot + footer slot.

**And** the **footer slot is conditional** within the loaded branch — if `footer_slot` is null/empty, the `<div class="entity-rail__footer">` is **omitted entirely** (not rendered as an empty `<div>`). This is the only nested conditional permitted. A per-stack snapshot variant is added if the test author judges the "loaded-shell with no footer" surface needs coverage (this would be a seventh variant `loaded-shell-no-footer-violation`; treat the six-variant floor in AC1 as the minimum, not the ceiling — add the seventh if it clarifies the footer-omission contract for the test).

**And** a per-stack unit test exercises four cases:
1. `entity_loaded=false, body_slot=null, footer_slot=null` → empty-state shell.
2. `entity_loaded=true, body_slot="<p>body</p>", footer_slot="<button>Save</button>"` → loaded shell with both slots.
3. `entity_loaded=true, body_slot="<p>body</p>", footer_slot=null` → loaded shell with body only; **no** `<div class="entity-rail__footer">` element in the output (assert by absence).
4. `entity_loaded=true, body_slot=null, footer_slot=null` → **edge case** — loaded shell with neither slot; renders the header strip, an empty `<div class="entity-rail__body"></div>`, and omits the footer. A test asserts this is the rendered shape and does not crash. (Rationale: the consumer might call the wrapper at the "shell" stage and populate the slots later — though this is unusual, the wrapper must not error.)

### AC5 — Responsive collapse rule in `_layout.css` (UX §"Layout collapse rules" at [ux-design-specification.md:1162–1169](../planning-artifacts/ux-design-specification.md))

**Given** the responsive-collapse rule at UX §"Layout collapse rules" line 1169 ("Sticky right rail" at desktop ≥1280; "Un-fixes; stacks beneath list" at tablet 768–1279; "Stacks beneath list" at mobile <768) and the Tailwind breakpoint discipline at [ux-design-specification.md:1180–1191](../planning-artifacts/ux-design-specification.md)
**When** I inspect `fieldmark_shared/src/_layout.css`
**Then** a new block is appended at the bottom of the file (preserving the existing `app-container` / `header` / `main` / `footer` / `body` rules untouched) with this exact shape:

```css
/* ─── EntityRail responsive collapse ─────────────────────────────────────────
   Per UX-DR Layout-collapse rules (ux-design-specification.md §Responsive
   Breakpoints). Owned by Story 2.6.

   Desktop ≥ 1280px (Tailwind xl:): sticky right rail at the right one-third
   of the content area; rail is `position: sticky` with `top` matching the
   header strip height.

   Below 1280px (tablet + mobile): rail un-fixes, stacks beneath the list
   (no transform — CSS grid template change in the consuming page; the rail
   itself simply releases `position: sticky`).
   ─────────────────────────────────────────────────────────────────────── */

.entity-rail {
  width: 100%;
  /* Empty rail and loaded rail share base layout; variant modifiers
     (entity-rail--empty / entity-rail--loaded) carry the cosmetic differences
     (the empty state uses a Basecoat .card surface; the loaded state uses
     no surface so its inner header/body/footer can paint freely). */
}

@media (min-width: 1280px) {
  .entity-rail {
    position: sticky;
    /* top offset = header strip height (3.5rem from _layout.css `header nav`)
       + main padding-block (1.5rem from _layout.css `main`) = 5rem. */
    top: 5rem;
    /* Rail max-height fills the remaining viewport so its internal
       body slot can scroll independently of the page. */
    max-height: calc(100vh - 5rem - 1.5rem);
    overflow-y: auto;
  }
}
```

**And** no other CSS file in `fieldmark_shared/src/` is touched this story. **Do not** add a `.entity-rail` rule to `_components.css`; the rail's geometry is a *layout* concern, not a component-surface concern, and `_layout.css` is the only place layout rules live.

**And** the `top: 5rem` offset is derived from the existing `header nav { height: 3.5rem }` + `main { padding-block: 1.5rem }` constants in `_layout.css`. If a future story changes either constant, the EntityRail top offset must be updated in the same change — record this in the CSS comment block above (the comment already names the source of the 5rem value). Do not use a CSS custom property for the offset this story (overengineering for one consumer; a future header-resize story can introduce `--header-height` and propagate the variable).

**And** after editing `_layout.css`, run `cd fieldmark_shared && pnpm run build` and commit the regenerated `dist/fieldmark.css`. The new rule must appear in the compiled bundle; if `pnpm run build` reports any LightningCSS or Tailwind warning, fix it before commit (per the `fieldmark_shared/CLAUDE.md` §"Build-Script Defensive Defaults" — fatal warnings exit non-zero).

### AC6 — Playwright responsive-viewport test (the **only** new e2e test this story lands)

**Given** UX §"Layout collapse rules" line 1169 and the responsive-collapse rule from AC5
**When** I inspect `e2e/tests/shared/entity-rail-responsive.spec.ts` (or the existing e2e suite location — verify before creating; the path may be `e2e/tests/shared/` or `tests/e2e/shared/` depending on the Story 1.14 layout)
**Then** a new spec file exists with **three** test cases — one per viewport width — that load a minimal test fixture page (a page that renders the empty EntityRail wrapper alone, served by **any one** of the three stacks; the e2e suite already configures cross-stack base URLs in the Story 1.14 Playwright project setup — match that setup). The test cases:

1. **Desktop ≥ 1280px** — viewport `{ width: 1280, height: 800 }`. Assert that `aside.entity-rail` has computed `position: sticky` and `top: 80px` (5rem at 16px root font-size).
2. **Tablet 768–1279px** — viewport `{ width: 1024, height: 768 }`. Assert that `aside.entity-rail` has computed `position: static` (no sticky) and that `top` is `auto` or the computed default. The rail's bounding rect must be **below** the simulated list element in the page's normal flow.
3. **Mobile <768px** — viewport `{ width: 375, height: 667 }`. Same assertions as the tablet case — `position: static`, rail stacked below list.

**And** the test fixture page is created in **one** stack (.NET is the recommended stack since its `/_test/render-partial/…` endpoint scaffold from Story 2.4 already provides a partial-rendering surface; if that scaffold is not yet committed, fall back to a one-off Django dev-only view at `/__test__/entity-rail-fixture/` gated behind `DEBUG=True`). The page renders a stub list `<ul style="height:1200px;background:lightblue"><li>list</li></ul>` next to the empty EntityRail, so the sticky behavior at desktop has visible vertical scroll room to be observable. **Do not** introduce a production route; the fixture page must be gated behind a debug / `#if DEBUG` flag so it does not appear in `make parity`.

**And** the three test cases pass at the documented breakpoint boundaries — the assertion at `width=1280` must trigger the `(min-width: 1280px)` media query (it does — `min-width` is inclusive); the assertion at `width=1279` (if added as a boundary case) must NOT trigger it. The test author may add a fourth boundary case (`width=1279`) if they judge the boundary worth nailing down. Three is the floor.

**And** the test file's top-of-file comment links to UX §"Layout collapse rules" so a future reader knows the source of the breakpoint values. Do not hardcode breakpoint numbers in test assertions without naming the source.

### AC7 — Tab-swap independence (UX-DR24 — rail survives tab-content swaps)

**Given** UX-DR §"Independent of tab content" at [ux-design-specification.md:1041](../planning-artifacts/ux-design-specification.md) ("The right rail is independent of tab swap; survives tab switches")
**When** the Project Detail screen (Story 2.11) swaps `#project-detail-tab-content` via an HTMX request
**Then** the rail (`#violation-detail` / `#inspection-detail` / `#corrective-action-detail`) MUST NOT be cleared by the tab swap. This story does **not** itself land Project Detail, but it does land a **structural assertion** that the rail is a *sibling* of the tab-content target in the canonical layout — not a child — by recording the canonical Project Detail grid template in `README.md` §contract item 8 ("Layout placement"):

```
Canonical Project Detail layout (informative — implemented in Story 2.11):

  <main>
    <section id="project-header-strip">…</section>
    <nav role="tablist">…TabStrip…</nav>
    <div class="project-detail-grid">
      <div id="project-detail-tab-content" role="tabpanel">
        …tab content (Summary / Inspections / Violations / Audit)…
      </div>
      <aside id="violation-detail" …>…EntityRail…</aside>
    </div>
  </main>

The rail is a sibling of #project-detail-tab-content inside the grid
container, not a descendant. A tab swap targets #project-detail-tab-content
and replaces only that node; the sibling <aside> is untouched.
```

The Project Detail page (Story 2.11) will reference this informative section when composing the layout. This story does not introduce the grid container CSS — Story 2.11 owns that (since the grid is a Project-Detail concern, not an EntityRail concern). The README section is **descriptive guidance**, not a CSS commitment.

### AC8 — Component edge-case checklist coverage (per [component-edge-case-checklist.md](../../docs/reference/component-edge-case-checklist.md))

Walked the nine categories — only the **applicable** ones get an AC:

**Given** category 6 (text overflow & special characters in user-visible strings)
**When** the wrapper receives an `entity_type_label` prop with XSS-prone characters (`<`, `>`, `&`, `"`, `'`)
**Then** each framework's default auto-escaping converts them to entities — verified by a per-stack XSS-payload round-trip test: `entity_type_label="<script>alert(1)</script>"` produces both `aria-label="&lt;script&gt;alert(1)&lt;/script&gt; detail"` **and** `<span class="entity-rail__entity-type">&lt;script&gt;alert(1)&lt;/script&gt;</span>` in the rendered output. The `body_slot` / `footer_slot` props are **NOT** subject to this test — they are the documented slot-pass-through points (AC3) where the wrapper opts out of escaping with deliberate per-stack mechanisms; the trust boundary is the consumer page.

**Given** category 6 also applies to long entity-type labels (e.g., a 200-char label)
**When** the loaded shell renders with a long label in the header strip
**Then** CSS controls the overflow via the existing Story 1.14 `.truncate` utility on `.entity-rail__entity-type`. Add the `.truncate` class to the canonical `<span class="entity-rail__entity-type">` — if the existing `.truncate` utility is not yet declared in `_components.css` (Story 1.14 should have landed it; verify), defer to Story 1.14's resolution and do not redeclare. The test for long-label rendering is a Playwright visual-regression case **deferred to the Epic 7 visual-regression suite** — do not add it this story (overengineering for the markup-only contract).

**Given** category 8 (forced-colors / high-contrast mode)
**When** I render the EntityRail under `forced-colors: active`
**Then** the empty-state Basecoat `.card` already has the Story 1.14 `_a11y.css` defenses (`forced-color-adjust: auto; border: 1px solid ButtonText`) on Basecoat surface elements. Verify the rule covers `.card`; if it does not, the **existing rule** in `_a11y.css` is extended (do not create a new file). The loaded shell's header / body / footer divs are unstyled by surface — they inherit from `<aside>` which has no background — so forced-colors mode has nothing to lose against. The `<aside>` landmark itself is identified by `role="region"` + `aria-label`, which screen readers and the OS color scheme both honor.

**Given** category 9 (empty / whitespace text input to derived values)
**When** the wrapper receives an empty or whitespace-only `entity_type_label`
**Then** the wrapper renders the label slot with the literal empty / whitespace content (does not crash). The `aria-label` value becomes `" detail"` (leading space) for whitespace input — this is acceptable per the empty/whitespace-input policy (the wrapper does not derive — it just renders). A per-stack unit test asserts `entity_type_label=""` and `entity_type_label="   "` do not crash the wrapper and produce a valid `<aside>` element. The fallback name "??" / "unnamed" from [component-edge-case-checklist.md §9](../../docs/reference/component-edge-case-checklist.md) does **not** apply here — that fallback is for *derived* tokens (initials, slugs), not literal display labels. The wrapper does not derive.

**Given** categories 1 (unknown enum), 2 (font load), 3 (JS init), 4 (AG Grid overlays), 5 (stacking), 7 (reduced motion)
**When** I evaluate against this story's deliverables
**Then** they are **N/A** with this rationale:
- **1 (unknown enum):** EntityRail has no vocabulary token props. `id`, `entity_type_label`, slots are all free-form strings. No unknown-token resolution needed.
- **2 (font load):** No custom font; inherits from page chrome.
- **3 (JS init):** Zero-JS component. No initialization marker, no progressive enhancement. The rail renders fully functional with `javaScriptEnabled: false`.
- **4 (AG Grid overlays):** Not an AG Grid story.
- **5 (stacking):** The rail does not queue items. A future "rail shows a stack of detail history" story (no such story planned) would invoke this category; not now.
- **7 (reduced motion):** No transitions / animations introduced. The responsive collapse is a media-query class change, not an animation.

### AC9 — Security-defaults checklist coverage (per [security-defaults.md](../../docs/reference/security-defaults.md))

Walked the seven categories — only the **applicable** one gets an AC:

**Given** category 3 (allowlist validation on writes) — **adapted here** for output-escaping on read since this story has no writes
**When** the wrapper renders `entity_type_label` or `id` (the non-slot string props)
**Then** the framework's default auto-escaping is in force — verified by the per-stack XSS-payload round-trip test in AC8 §category-6. The `body_slot` / `footer_slot` props are the **explicit documented exception** per AC3 — they are slot-pass-through points where the consumer is the trust boundary. The grep guard from Stories 2.4 / 2.5 forbidding `Html.Raw` / `|safe` / `template.HTML(` in component wrappers is **scoped-exempted** for the three EntityRail wrapper files, and **only** for the two slot positions per file. Any other use of those tokens in any of the three wrapper files is forbidden, and the per-stack grep guard MUST be a positive test (assert exactly N occurrences in the file, where N is the documented slot count — two for .NET, two for Django, two for Go via `safeHTML`).

**Given** categories 1 (open-redirect), 2 (cookie attributes), 4 (dynamic RegExp), 5 (filesystem writes), 6 (CSRF posture), 7 (stub-auth warnings)
**When** I evaluate against this story's deliverables
**Then** they are **N/A** — no redirects, no cookies, no regex on user input, no filesystem writes (the `pnpm run build` step is the existing build process from Story 1.14, not new tooling), no routes introduced in the production stack (the AC6 fixture page is debug-gated), no auth changes.

### AC10 — Cross-stack architecture principle three-deliverable check (root [CLAUDE.md](../../CLAUDE.md))

This story introduces **one cross-stack contract surface** — the `<aside id="<rail-id>" role="region" tabindex="-1" aria-label="<EntityType> detail" aria-live="polite">` landmark shape that is the canonical HTMX target for `#violation-detail`, `#inspection-detail`, `#corrective-action-detail` — and produces all three deliverables:

1. **Documentation contract:** `fieldmark_shared/components/entity_rail/README.md` (AC1) + appended row in `docs/reference/component-canonical-examples.md` (AC1) + the responsive-collapse comment block in `_layout.css` (AC5) + the canonical Project Detail layout snippet in README §contract item 8 (AC7).
2. **Native implementation per stack:** the three wrapper templates in idiomatic locations (AC2 / AC3) + the per-stack slot-pass-through mechanism (`Html.Raw` / `|safe` / `safeHTML`) each implemented natively, no shared codec, no generated stubs, no symlinked partial.
3. **Per-stack conformance test:** snapshot tests over six variants (AC2 / AC3) + slot pass-through and footer-omission unit tests (AC4) + the cross-stack Playwright responsive-viewport test (AC6) + the per-stack XSS-payload escaping test for non-slot props (AC8) + the per-stack scoped grep guard counting exactly two slot-pass-through usages (AC9).

**And** there is **no new file** in `fieldmark_shared/` that lists the rail target ids (`#violation-detail` etc.) as a shared symbol manifest. The canonical id vocabulary is documented in [ux-design-specification.md:1114](../planning-artifacts/ux-design-specification.md) and the README's §contract item 2 (Props — `id: string`). The wrapper accepts any caller-supplied string; the consumer story is responsible for using the canonical id. The cross-stack invariant is enforced by the snapshot test (canonical.html names the three canonical ids in the three loaded-shell variants) plus the consumer story's own AC list, not by a shared symbol file.

### AC11 — `make parity` clean, no new production routes introduced

**Given** all wrappers + tests + the CSS rule land
**When** I run `make parity` from the repo root
**Then** the route-parity script reports the **same** drift baseline as Story 2.5 — no new `GET /…` or `POST /…` production endpoints. The AC6 Playwright fixture page (if added as a `/__test__/entity-rail-fixture/` route on Django or `/_test/render-partial/entity-rail` on .NET) MUST be gated behind a debug flag and excluded from the parity diff. If the existing Story 2.4 `/_test/render-partial/…` scaffold is the host, the route is already debug-gated and no new route entry is needed. `pg_indexes` diff: zero (no DB changes).

### AC12 — Build, type, lint, and test gates green on every stack

- **.NET:** `cd FieldMark && dotnet csharpier check . && dotnet build && dotnet test && dotnet test FieldMark.Tests.Integration/FieldMark.Tests.Integration.csproj` — clean. The snapshot test class passes with one `[Theory]` row per variant; the slot-pass-through unit tests pass.
- **Django:** `cd fieldmark_py && uv run ruff check . && uv run mypy . && uv run pytest && uv run pytest -m integration` — clean. The snapshot test passes; the slot-pass-through and footer-omission tests pass.
- **Go:** `cd fieldmark-go && make check && go test ./... && go test -tags=integration ./...` — clean. The template snapshot test passes; the `safeHTML` template-function registration is exercised by at least one test.
- **`fieldmark_shared`:** `cd fieldmark_shared && pnpm install && pnpm run build` — clean; `dist/fieldmark.css` regenerated and committed (the `.entity-rail` block from AC5 propagates into the compiled bundle).
- **E2E:** the new Playwright spec (AC6) passes against the chosen host stack with three viewport cases. The other two stacks are not required to run this spec — the layout rule lives in shared CSS, so testing one host stack covers the cross-stack invariant.
- From repo root: `make parity` exits 0 (AC11) and `make test-all` (canonical pre-merge gate) exits 0.

## Tasks / Subtasks

- [ ] **Task 1: Author canonical example + README in `fieldmark_shared/`** (AC: #1, #2, #3, #10)
  - [ ] 1.1 Create `fieldmark_shared/components/entity_rail/canonical.html` with six variant blocks (`empty-violation`, `empty-inspection`, `empty-corrective-action`, `loaded-shell-violation`, `loaded-shell-inspection`, `loaded-shell-corrective-action`). Include the U+00D7 multiplication sign as the dismiss-× character.
  - [ ] 1.2 Create `fieldmark_shared/components/entity_rail/README.md` per AC1 §contract-fixed-order — nine sections (Purpose, Props, Variants, ARIA, Slot contract, Allowed classes, Snapshot equality, Responsive collapse, Focus invariant) plus the canonical Project Detail layout snippet in §contract item 8 (AC7).
  - [ ] 1.3 Append one row to `docs/reference/component-canonical-examples.md` Component Index — `EntityRail`, fixture path, README path, three wrapper paths, three test paths.

- [ ] **Task 2: Author the responsive-collapse CSS rule in `_layout.css`** (AC: #5, #10, #12)
  - [ ] 2.1 Append the EntityRail block at the bottom of `fieldmark_shared/src/_layout.css` per AC5 §exact-shape. Include the comment block naming the UX-DR source and the 5rem-offset derivation.
  - [ ] 2.2 Run `cd fieldmark_shared && pnpm run build`. Verify no LightningCSS / Tailwind warnings. Commit the regenerated `dist/fieldmark.css`.

- [ ] **Task 3: .NET wrapper + tests** (AC: #2, #3, #4, #8, #9, #12)
  - [ ] 3.1 Create `FieldMark/FieldMark.Web/Pages/Shared/Components/_EntityRail.cshtml` with in-file `EntityRailViewModel` record. Implement the slot pass-through via exactly two `@Html.Raw(Model.BodySlot ?? string.Empty)` / `@Html.Raw(Model.FooterSlot ?? string.Empty)` calls, inside `<div class="entity-rail__body">` / `<div class="entity-rail__footer">` containers. Footer omission: wrap the `<div class="entity-rail__footer">…</div>` in `@if (Model.FooterSlot is not null) { … }`.
  - [ ] 3.2 Top-of-file comment references `docs/reference/component-canonical-examples.md`.
  - [ ] 3.3 Create `FieldMark/FieldMark.Tests.Web/Components/EntityRailSnapshotTests.cs` (or `FieldMark.Tests.Integration/Components/…` per the Story 2.4 host decision). One `[Theory]` row per variant; reuse the Story 2.4 partial-render scaffold.
  - [ ] 3.4 Slot-pass-through + footer-omission unit tests per AC4 (four cases).
  - [ ] 3.5 XSS round-trip test for `entity_type_label` (AC8 §category-6 — non-slot props are framework-escaped).
  - [ ] 3.6 Negative HTMX-producer-attribute test on the dismiss × (AC3).
  - [ ] 3.7 Scoped grep guard (AC9) — CI lane assertion that `Html.Raw` appears exactly **twice** in `_EntityRail.cshtml` and zero times in any other component wrapper file. Update the existing Story 2.4 / 2.5 grep guard configuration to add this file-scoped exemption.
  - [ ] 3.8 Run `dotnet csharpier check . && dotnet build && dotnet test && dotnet test FieldMark.Tests.Integration/` — clean.

- [ ] **Task 4: Django wrapper + tests** (AC: #2, #3, #4, #8, #9, #12)
  - [ ] 4.1 Create `fieldmark_py/templates/components/_entity_rail.html` (NEW). Implement the slot pass-through via exactly two `{{ body_slot|safe }}` / `{{ footer_slot|safe }}` filters. Footer omission: wrap the `<div class="entity-rail__footer">…</div>` in `{% if footer_slot %}…{% endif %}`.
  - [ ] 4.2 Top-of-file comment references `docs/reference/component-canonical-examples.md`.
  - [ ] 4.3 Create `fieldmark_py/components/tests/test_entity_rail_snapshot.py` — `@pytest.mark.parametrize` over six variants; load variant block from `canonical.html` via the Story 2.4 path-walker; normalize; byte-equal assert.
  - [ ] 4.4 Slot-pass-through + footer-omission tests (AC4).
  - [ ] 4.5 XSS round-trip test for `entity_type_label`.
  - [ ] 4.6 Negative HTMX-producer-attribute test on the dismiss ×.
  - [ ] 4.7 Scoped grep guard — CI lane assertion that `|safe` appears exactly **twice** in `_entity_rail.html` and zero times in other component wrappers.
  - [ ] 4.8 Run `uv run ruff check . && uv run mypy . && uv run pytest && uv run pytest -m integration` — clean.

- [ ] **Task 5: Go wrapper + tests** (AC: #2, #3, #4, #8, #9, #12)
  - [ ] 5.1 Locate the existing Go template-function-map registration (search `fieldmark-go/internal/web/templates/` — likely `templates.go` from Story 1.5 / 1.12). If a `safeHTML` function does not yet exist, register it: `"safeHTML": func(s string) template.HTML { return template.HTML(s) }`. Document the function with a one-line code comment per AC3. If the function map registration lives in multiple files (per stack-internal scoping), add the registration to the one Story 2.4 / 2.5 wrappers reuse.
  - [ ] 5.2 Create `fieldmark-go/internal/web/templates/components/entity_rail.go` (NEW) — `type EntityRailArgs struct { ID, EntityTypeLabel string; EntityLoaded bool; BodySlot, FooterSlot template.HTML }`.
  - [ ] 5.3 Create `fieldmark-go/internal/web/templates/components/entity_rail.html` (NEW) — `{{define "entity_rail"}}…{{end}}`. Footer omission: `{{if .FooterSlot}}<div class="entity-rail__footer">{{.FooterSlot}}</div>{{end}}`. Body slot: `<div class="entity-rail__body">{{.BodySlot}}</div>` (the `template.HTML`-typed field renders verbatim without the `safeHTML` function call needed — Go's `html/template` honors `template.HTML` directly). The `safeHTML` template function is registered for use by *callers* who construct `EntityRailArgs` from string inputs; the wrapper itself relies on the `template.HTML` field type.
  - [ ] 5.4 Top-of-file comments in both files reference `docs/reference/component-canonical-examples.md`.
  - [ ] 5.5 Create `fieldmark-go/internal/web/templates/components/entity_rail_test.go` — table-driven sub-tests per variant. Plus four-case slot-pass-through / footer-omission test.
  - [ ] 5.6 XSS round-trip test for `EntityTypeLabel`; negative HTMX-producer-attribute test on dismiss ×.
  - [ ] 5.7 Scoped grep guard — CI lane assertion that `template.HTML(` appears in `entity_rail.go` (the args struct uses the type) but **does not** appear in `entity_rail.html` (the template body relies on the typed field, not a cast). The exact occurrence count: one in `entity_rail.go` (the field type declaration is a type identifier, not a cast — the grep should target `template.HTML(` with the paren); zero in `entity_rail.html`. Adjust the grep target precisely to match this. Other component template files remain forbidden.
  - [ ] 5.8 Run `make check && go test ./... && go test -tags=integration ./...` — clean.

- [ ] **Task 6: Playwright responsive-viewport test** (AC: #6, #12)
  - [ ] 6.1 Locate the existing e2e suite directory (verify `e2e/tests/shared/` vs `tests/e2e/shared/` — check Story 1.14's Playwright project setup).
  - [ ] 6.2 Author the fixture page (recommend the existing Story 2.4 `/_test/render-partial/…` .NET endpoint; fall back to a Django debug-gated view if 2.4's scaffold is not yet committed).
  - [ ] 6.3 Create `e2e/tests/shared/entity-rail-responsive.spec.ts` — three viewport test cases per AC6 (1280, 1024, 375). Top-of-file comment links to UX §"Layout collapse rules".
  - [ ] 6.4 Run the spec against the host stack. All three viewport cases pass.

- [ ] **Task 7: Cross-stack verification + parity** (AC: #10, #11, #12)
  - [ ] 7.1 Run `make parity` — route diff equals the Story 2.5 baseline; no new production routes. `pg_indexes` zero diff.
  - [ ] 7.2 Run `make test-all` — green.
  - [ ] 7.3 Confirm each new wrapper file's top-of-file comment references `docs/reference/component-canonical-examples.md`.
  - [ ] 7.4 Verify the Component Index row for EntityRail is correctly populated.
  - [ ] 7.5 Verify the `_layout.css` change rendered into `dist/fieldmark.css` (the compiled file contains the `.entity-rail` block).

- [ ] **Task 8: Story sign-off** (AC: all)
  - [ ] 8.1 Populate the Sign-off block below; flip sprint-status to `review`.

## Dev Notes

### Critical context (read before writing code)

- **Markup-only — zero JS, no production routes, no DB.** EntityRail is a structural container plus one CSS block. If you find yourself adding a `<script>`, a focus-handler JS file, a `hx-trigger`, or any production HTTP endpoint, you are out of scope. Stop and re-read AC11. The focus-after-swap behavior is *consumer-owned* — the wrapper just emits `tabindex="-1"` so a future consumer's `autofocus` / `HX-Trigger` script has a focus target.
- **The slot-pass-through is the one place this wrapper is "unsafe" by design.** `Html.Raw` / `|safe` / `template.HTML` would be a bug in StatusBadge or DashboardTile — there, the props are plain text. In EntityRail, the slot props are *intentional* HTML pass-through; the consumer page renders a `#violation-detail` partial to a string and hands it to the wrapper for landmark wrapping. This is the **opposite** of a security gap; forcing the wrapper to escape the slot would prevent the wrapper from ever wrapping HTML content. The trust boundary moves to the consumer, where it always lived for partials. The grep guard scoping (file + count exactly two per file) is how this story keeps the exception from being abused.
- **The responsive-collapse rule is the ONLY CSS this story adds.** Resist the urge to add `.entity-rail__header` / `.entity-rail__body` / `.entity-rail__footer` styling beyond what the layout-collapse rule requires. Those are cosmetic concerns — they can be added in a future polish story or alongside the first consumer (Story 2.11) when there's actual content to style around. This story's job is the *contract surface*, not the *visual polish*.
- **Sequencing with Stories 2.4 / 2.5.** This story uses the per-component directory convention, the variant-delimiter parser, the snapshot-test harness, the path-walker, the `Pages/Shared/Components/` Razor sub-directory, and the partial-render scaffold endpoint that Stories 2.4 and 2.5 introduce. **Block on them** — do not pre-implement those primitives here. If during implementation you find a primitive missing, flag it as a Story 2.4 / 2.5 review-round patch rather than reimplementing it.
- **The `aria-live="polite"` is on the `<aside>` itself.** UX-DR §"Live region politeness" line 1223 lists EntityRail among the polite live regions. This means screen readers announce changes when the rail's content is swapped (a partial inserted via HTMX). Combined with `aria-atomic` being **absent** (defaults to `false`), only the *changed* content is announced, not the whole rail re-spoken — which is what we want when only the body or footer slot updates.
- **No `aria-atomic` attribute.** Do not add `aria-atomic="true"` (would make the whole rail re-announce on every change — too noisy) or `aria-atomic="false"` (the default — adding it is redundant noise in the markup). The attribute is absent intentionally.
- **The dismiss × is a placeholder.** Story 2.11 (and downstream consumer stories) will wire the × to clear the rail back to its empty state. This story emits the button shape only. The button must be focusable and keyboard-activatable by default (`<button type="button">` is both); the wrapper does not give it `disabled` because the empty-the-rail action is always valid when the rail is loaded.
- **No header strip on the empty-state variant.** UX-DR specifies the header strip exists on the *loaded* shell (entity-type label + dismiss). The empty state is just the card with the instruction. Do not render the header strip on empty — that would imply there is something to dismiss when there is not.
- **The `<aside>` is a landmark — there is one per page.** UX §"Document landmark structure" at [ux-design-specification.md:1216](../planning-artifacts/ux-design-specification.md) allows "optional `<aside>` (EntityRail)" per page. If a consumer page ever needs two rails simultaneously, this design assumption breaks — but no MVP story requires it. The wrapper does not enforce uniqueness (it can't — it doesn't see the page). Document the constraint in the README and rely on per-page audit.
- **The 5rem sticky-top offset is brittle but acceptable.** It encodes the header-strip height (3.5rem) + the main `padding-block` (1.5rem). If either of those `_layout.css` constants changes, the EntityRail offset must move in lock-step. A CSS custom property like `--header-height` would decouple them — but introducing the variable across the existing layout is overengineering for one consumer. A future header-resize story can introduce the variable and propagate; until then, the comment block in AC5 names the source of the 5rem so a future editor knows what to change.

### Component-specific notes

- **The dismiss × character matters byte-for-byte.** U+00D7 (×, multiplication sign) is the conventional dismiss glyph in design-system literature. U+2715 / U+2716 are decorative variants used in some icon fonts. ASCII `x` is the wrong character (different letterforms and no convention). HTML entity `&times;` and `&#x00D7;` render the same glyph but are different byte sequences — the byte-equality contract specifies the literal Unicode character. Each stack's template engine should emit it from source as the literal character (not the entity); verify by hex-dumping `canonical.html` to confirm `C3 97` is present at the dismiss-button position.
- **`role="region"` on `<aside>` is intentional, not redundant.** ARIA 1.2 specifies that an `<aside>` already has an implicit `role="complementary"`; explicitly setting `role="region"` *overrides* the implicit complementary role, which is appropriate here because the rail is more accurately a generic named region (the entity-detail container) than a "complementary" tangentially-related section. UX-DR §"EntityRail" line 876 explicitly says `role="region"`. Per WAI-ARIA Authoring Practices, regions require an accessible name — `aria-label` provides it. This is correct usage; the per-stack accessibility-lint should not flag the explicit role.
- **`tabindex="-1"` is essential for programmatic focus.** Without it, `.focus()` calls on a non-interactive element silently fail in some browsers, and screen reader virtual-buffer mode can't land focus correctly. `-1` (as opposed to `0`) keeps the rail out of the natural tab order — only programmatic focus lands here, which is what the post-swap focus convention requires.
- **The slot pass-through grep-guard count is `exactly two`, not `at most two`.** A wrapper that "forgets" one of the two slot positions would silently drop slot content; the test must catch absence as well as excess. Each stack's grep should target the specific token in the specific file and assert `count == 2` (or `count == 1` for Go's `safeHTML` function call — but Go's mechanism is the `template.HTML` field type, not a function call in the template, so the grep target for Go is the type declaration in `entity_rail.go`, not a call site in `entity_rail.html`).

### Edge cases (per [component-edge-case-checklist.md](../../docs/reference/component-edge-case-checklist.md))

Walked the nine categories — see AC8 for the per-category ACs. Summary: categories **6**, **8**, **9** apply and have ACs; categories **1**, **2**, **3**, **4**, **5**, **7** are N/A with rationale recorded in AC8.

### Security defaults (per [security-defaults.md](../../docs/reference/security-defaults.md))

Walked the seven categories — see AC9 for the per-category ACs. Summary: category **3** applies (adapted as output-escaping on read for non-slot props, with documented slot-pass-through exceptions for the two slot positions); categories **1**, **2**, **4**, **5**, **6**, **7** are N/A.

### Cross-stack contract three-deliverable check

This story introduces one cross-stack contract surface — the `<aside id role tabindex aria-label aria-live>` landmark shape and the responsive-collapse rule — and produces all three deliverables: (1) the per-component `README.md` + appended row in `docs/reference/component-canonical-examples.md` + the layout-collapse comment block in `_layout.css` + the Project Detail layout snippet in the README, (2) per-stack wrapper templates with idiomatic slot pass-through mechanisms (`Html.Raw` / `|safe` / `template.HTML` field type), (3) per-stack snapshot tests + slot/footer-omission tests + Playwright responsive-viewport test + XSS escaping tests + scoped grep guards. See AC10.

### Files this story modifies vs creates

| File | New / Modified | Purpose |
|---|---|---|
| `fieldmark_shared/components/entity_rail/canonical.html` | NEW | six variant blocks |
| `fieldmark_shared/components/entity_rail/README.md` | NEW | contract + canonical Project Detail layout snippet |
| `fieldmark_shared/src/_layout.css` | MODIFY | append `.entity-rail` block per AC5 |
| `fieldmark_shared/dist/fieldmark.css` | MODIFY (regenerated) | commit after `pnpm run build` |
| `docs/reference/component-canonical-examples.md` | MODIFY | append one row to Component Index |
| `FieldMark/FieldMark.Web/Pages/Shared/Components/_EntityRail.cshtml` | NEW | wrapper with in-file view model |
| `FieldMark/FieldMark.Tests.{Web,Integration}/Components/EntityRailSnapshotTests.cs` | NEW | snapshot + slot tests |
| `fieldmark_py/templates/components/_entity_rail.html` | NEW | wrapper |
| `fieldmark_py/components/tests/test_entity_rail_snapshot.py` | NEW | snapshot + slot tests |
| `fieldmark-go/internal/web/templates/components/entity_rail.go` | NEW | args struct |
| `fieldmark-go/internal/web/templates/components/entity_rail.html` | NEW | `{{define "entity_rail"}}` wrapper |
| `fieldmark-go/internal/web/templates/components/entity_rail_test.go` | NEW | snapshot + slot tests |
| `fieldmark-go/internal/web/templates/templates.go` (or wherever Go function-map lives) | MODIFY (if `safeHTML` not yet registered) | register `safeHTML` template function for caller use |
| `e2e/tests/shared/entity-rail-responsive.spec.ts` | NEW | three-viewport collapse test |
| (`docs/reference/component-canonical-examples.md` already listed above) | | |

Anything outside this list — Project Detail page, the actual dismiss-button wiring, the focus-on-swap script, any consumer page that populates the rail, AG Grid row-click wiring, route registration, any DB change — is out of scope. Resist the urge.

### Files to read fully before editing

- [_bmad-output/planning-artifacts/ux-design-specification.md:871–877](../planning-artifacts/ux-design-specification.md) — EntityRail UX-DR spec (Custom Components).
- [_bmad-output/planning-artifacts/ux-design-specification.md:1041](../planning-artifacts/ux-design-specification.md) — rail survives tab swaps (UX-DR24).
- [_bmad-output/planning-artifacts/ux-design-specification.md:1049–1076](../planning-artifacts/ux-design-specification.md) — Pattern 6 (List+Detail Co-Presence) + Pattern 7 (Empty State With Next Action).
- [_bmad-output/planning-artifacts/ux-design-specification.md:1162–1189](../planning-artifacts/ux-design-specification.md) — Layout collapse rules + Breakpoint strategy (binding for AC5).
- [_bmad-output/planning-artifacts/ux-design-specification.md:1114](../planning-artifacts/ux-design-specification.md) — canonical HTMX target IDs (rail ids listed).
- [_bmad-output/planning-artifacts/ux-design-specification.md:1213–1230](../planning-artifacts/ux-design-specification.md) — Document landmark structure + focus management + live-region politeness (binding for AC2 / AC3 / AC8).
- [_bmad-output/planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md](../planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md) §Story 2.6 — epic AC source.
- [docs/reference/component-edge-case-checklist.md](../../docs/reference/component-edge-case-checklist.md) — nine-category walkthrough; binding for AC8.
- [docs/reference/security-defaults.md](../../docs/reference/security-defaults.md) — seven-category walkthrough; binding for AC9.
- [fieldmark_shared/CLAUDE.md](../../fieldmark_shared/CLAUDE.md) §"Snapshot-test pipeline" — normalization pipeline this story's tests inherit; §"Build-Script Defensive Defaults" — fatal-warning policy for `pnpm run build`.
- [fieldmark_shared/src/_layout.css](../../fieldmark_shared/src/_layout.css) — existing layout file this story appends to. Read fully; do not duplicate or override any existing rule.
- [_bmad-output/implementation-artifacts/2-4-implement-phase-2-markup-only-components-statusbadge-inlinealert-auditrow-dashboardtile.md](2-4-implement-phase-2-markup-only-components-statusbadge-inlinealert-auditrow-dashboardtile.md) — per-component-directory convention, snapshot harness, path-walker, host decision for tests, grep-guard configuration.
- [_bmad-output/implementation-artifacts/2-5-implement-compliancetile-component-and-compliance-tile-oob-target.md](2-5-implement-compliancetile-component-and-compliance-tile-oob-target.md) — Go sibling-`.go`-file precedent (this story extends it).
- [FieldMark/FieldMark.Web/Pages/Shared/_ActionButton.cshtml](../../FieldMark/FieldMark.Web/Pages/Shared/_ActionButton.cshtml) — .NET wrapper precedent (in-file view model).
- [fieldmark_py/templates/components/_action_button.html](../../fieldmark_py/templates/components/_action_button.html) — Django `{% include … with … %}` precedent.
- [fieldmark-go/internal/web/templates/components/action_button.html](../../fieldmark-go/internal/web/templates/components/action_button.html) and [_test.go sibling](../../fieldmark-go/internal/web/templates/components/action_button_test.go) — Go `{{define}}` + table-driven test precedent.
- Stack rules: [FieldMark/CLAUDE.md](../../FieldMark/CLAUDE.md), [fieldmark_py/CLAUDE.md](../../fieldmark_py/CLAUDE.md), [fieldmark-go/CLAUDE.md](../../fieldmark-go/CLAUDE.md).
- Root cross-stack invariants: [CLAUDE.md](../../CLAUDE.md) §"Cross-Stack Architecture Principle" — binding for AC10.

### Project Structure Notes

- The Razor `Pages/Shared/Components/` sub-directory is created by Story 2.4. This story drops `_EntityRail.cshtml` alongside the four Story 2.4 wrappers + the Story 2.5 ComplianceTile wrapper. If Story 2.4 has not yet created the sub-directory at implementation time, this story must.
- The Django `templates/components/` and `components/tests/` directories are touched by Stories 2.4 / 2.5 — match their resolution.
- The Go `internal/web/templates/components/` directory follows the same pattern (sibling-`.go`-file from Story 2.5 is reused here).
- The e2e Playwright suite location may vary — Story 1.14 introduced cross-stack visual-regression Playwright tests. Verify the directory layout (`e2e/tests/shared/` vs `tests/e2e/shared/`) before creating `entity-rail-responsive.spec.ts`.
- The Go template-function map registration may live in different files per stack-internal conventions. Search for `template.FuncMap` or `Funcs(` in `fieldmark-go/internal/web/` to find the established location; add `safeHTML` there.

### References

- AC source: [_bmad-output/planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md](../planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md) §Story 2.6
- UX-DR EntityRail spec: [ux-design-specification.md:871–877](../planning-artifacts/ux-design-specification.md)
- UX Pattern 6 (List+Detail Co-Presence): [ux-design-specification.md:1049–1064](../planning-artifacts/ux-design-specification.md)
- UX Pattern 7 (Empty State With Next Action): [ux-design-specification.md:1066–1076](../planning-artifacts/ux-design-specification.md)
- Layout collapse rules: [ux-design-specification.md:1162–1169](../planning-artifacts/ux-design-specification.md)
- Breakpoint strategy: [ux-design-specification.md:1180–1191](../planning-artifacts/ux-design-specification.md)
- Canonical HTMX target IDs: [ux-design-specification.md:1114](../planning-artifacts/ux-design-specification.md)
- Document landmark structure: [ux-design-specification.md:1216](../planning-artifacts/ux-design-specification.md)
- Focus management on HTMX swaps: [ux-design-specification.md:1218–1221](../planning-artifacts/ux-design-specification.md)
- Live region politeness: [ux-design-specification.md:1222–1226](../planning-artifacts/ux-design-specification.md)
- Per-component-directory convention precedent: [Story 2.4](2-4-implement-phase-2-markup-only-components-statusbadge-inlinealert-auditrow-dashboardtile.md)
- Go sibling-`.go`-file precedent: [Story 2.5](2-5-implement-compliancetile-component-and-compliance-tile-oob-target.md)
- Component edge-case checklist: [docs/reference/component-edge-case-checklist.md](../../docs/reference/component-edge-case-checklist.md)
- Security defaults checklist: [docs/reference/security-defaults.md](../../docs/reference/security-defaults.md)
- Snapshot-test pipeline: [fieldmark_shared/CLAUDE.md](../../fieldmark_shared/CLAUDE.md) §"Snapshot-test pipeline"
- Cross-Stack Architecture Principle: root [CLAUDE.md](../../CLAUDE.md) §Cross-Stack Architecture Principle
- Stack rules: [FieldMark/CLAUDE.md](../../FieldMark/CLAUDE.md), [fieldmark_py/CLAUDE.md](../../fieldmark_py/CLAUDE.md), [fieldmark-go/CLAUDE.md](../../fieldmark-go/CLAUDE.md)

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
| Deferred-work entries | _none new — this story has no deferred items. The Story 2.4-followup unknown-token logger entry is unrelated; the EntityRail slot-pass-through is a documented design exception, not a deferred concern._ |
| Dev-notes divergences from epic AC | The epic AC says "either `HX-Trigger`-driven focus script or autofocus" — this story explicitly **defers** the focus-invocation script to consumer stories (2.11+) and lands only `tabindex="-1"` as the focus surface. Rationale: the wrapper is consumer-agnostic; the consumer (e.g., Story 2.11 Project Detail) chooses between `autofocus` on the inserted partial root vs an `HX-Trigger`-driven base-layout listener. Forcing the choice here would either (a) introduce JS the wrapper doesn't need, or (b) lock downstream consumers to one mechanism. The AC9 §AC6 Playwright test verifies the *responsive-collapse* behavior, not the focus invocation; focus invocation will be covered by Story 2.11's e2e tests when the first real consumer lands. |

### Review Findings

_to be populated by code-review_
