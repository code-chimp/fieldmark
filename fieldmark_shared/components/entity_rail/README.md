# EntityRail Component

## 1. Purpose

Right-rail container for the currently-selected entity's detail; the canonical `#violation-detail` / `#inspection-detail` / `#corrective-action-detail` HTMX target shape with stable focus surface and responsive collapse.

## 2. Required Props

| Prop | Type | Description |
|---|---|---|
| `id` | `string` | The stable HTMX target id (e.g. `violation-detail`). Framework-escaped. |
| `entity_type_label` | `string` | Used in `aria-label` and the loaded-shell header. Framework-escaped. |
| `entity_loaded` | `bool` | Selects empty (`false`) vs loaded (`true`) variant. |
| `body_slot` | `string?` | Caller-supplied inner HTML for the body slot (loaded variant only). Rendered verbatim — see Slot Contract. |
| `footer_slot` | `string?` | Caller-supplied action-footer markup (loaded variant only). Omitted entirely when null/empty. Rendered verbatim. |

## 3. Variant List

1. `empty-violation` — empty state, id=`violation-detail`
2. `empty-inspection` — empty state, id=`inspection-detail`
3. `empty-corrective-action` — empty state, id=`corrective-action-detail`
4. `loaded-shell-violation` — loaded shell with sentinel slots, id=`violation-detail`
5. `loaded-shell-inspection` — loaded shell with sentinel slots, id=`inspection-detail`
6. `loaded-shell-corrective-action` — loaded shell with sentinel slots, id=`corrective-action-detail`

## 4. ARIA Invariants

- `role="region"` on `<aside>` — overrides the implicit `complementary` role so the rail is identified as a named region, per UX-DR §"EntityRail" line 876.
- `aria-label="<entity_type_label> detail"` on `<aside>` for loaded variant; outer `aria-label="<EntityType> detail"` for empty variant; inner `aria-label="Empty entity rail"` on the empty-state card `<div>`.
- `tabindex="-1"` always — keeps the rail out of the natural tab order but makes it a valid programmatic focus target.
- `aria-live="polite"` per UX-DR §"Live region politeness" line 1223 — screen readers announce changes when HTMX swaps content into the rail; `aria-atomic` is absent (defaults `false`) so only changed content is announced.

## 5. Slot Contract

- **Header strip** — wrapper-owned in both states. Empty variant renders the empty-state card; loaded variant renders the entity-type label + dismiss-× button.
- **Body slot** — caller-owned via `body_slot`. Placed inside `<div class="entity-rail__body">`. Rendered verbatim without escaping.
- **Footer slot** — caller-owned via `footer_slot`. Placed inside `<div class="entity-rail__footer">`. Omitted entirely (no empty `<div>`) when null or empty. Rendered verbatim without escaping.

**Security note:** `body_slot` and `footer_slot` are rendered raw — the caller is the trust boundary, not the wrapper. The grep guard for each stack asserts exactly **two** raw-render occurrences per wrapper file (one for body, one for footer) so the exception cannot silently expand.

## 6. Allowed Class Vocabulary

| Class | Element |
|---|---|
| `entity-rail` | `<aside>` base class |
| `entity-rail--empty` | `<aside>` modifier — empty state |
| `entity-rail--loaded` | `<aside>` modifier — loaded state |
| `entity-rail__header` | `<header>` in loaded state |
| `entity-rail__entity-type` | `<span>` entity-type label; also carries Tailwind `truncate` utility |
| `entity-rail__dismiss` | dismiss `<button>` in loaded state |
| `entity-rail__body` | `<div>` body slot container |
| `entity-rail__footer` | `<div>` footer slot container |
| `entity-rail__empty` | `<div>` empty-state card container |
| `card` | Basecoat surface class on empty-state `<div class="card entity-rail__empty">` |

## 7. Snapshot-Equality Requirement

Per-stack wrappers MUST render output byte-equal to the matching variant block in `canonical.html` after the standard normalization defined in `fieldmark_shared/CLAUDE.md` §"Snapshot-test pipeline". Caller-supplied `body_slot` and `footer_slot` strings pass through verbatim — the wrapper does not escape, transform, or wrap them. The caller is responsible for the safety of HTML they place in the slots; the consumer page is the trust boundary, not the wrapper.

## 8. Responsive-Collapse Contract

Sticky right rail at ≥ 1280px; un-fixes and stacks beneath the list at < 1280px. Rule lives in `fieldmark_shared/src/_layout.css` under `/* EntityRail responsive collapse */`.

**Layout placement (informative — implemented in Story 2.11):**

```
Canonical Project Detail layout:

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

The rail survives tab-content swaps because it is a sibling of `#project-detail-tab-content`, not a descendant. Story 2.11 owns the grid container CSS.

## 9. Focus-on-Swap Invariant

Wrapper emits `tabindex="-1"` so consumer-supplied focus-after-swap (`autofocus` on inserted partial root, or `HX-Trigger` event handler) lands focus on the rail root. The wrapper itself does not invoke `.focus()` — that is a consumer concern (Story 2.11+).
