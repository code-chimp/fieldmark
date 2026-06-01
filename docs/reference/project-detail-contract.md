# Project Detail Contract

Story: 2.11  
Canonical screen: `GET /projects/:id`

## Routes

- `GET /projects/:id`
- `GET /projects/:id/tabs/:tab` where `:tab ∈ {summary, inspections, violations, audit}`

## Dual-Mode Rule

- When `HX-Request: true` on `GET /projects/:id`, return the body fragment only (no page chrome).
- Without `HX-Request`, return full page chrome around the same body markup.

## Canonical IDs

- `project-detail` (stable detail-shell wrapper; main re-render target)
- `project-header-strip`
- `compliance-tile`
- `project-detail-tabstrip`
- `project-detail-tab-content`
- `violation-detail`

The `compliance-tile` is nested inside `#project-detail`. Any future OOB compliance-tile re-render must remain byte-equivalent to the in-shell render to avoid cross-region drift.

## Tab Contract

- Tab order is fixed: `Summary`, `Inspections`, `Violations`, `Audit`.
- Tab click targets `#project-detail-tab-content` via `hx-swap="innerHTML"`.
- Non-HTMX direct navigation to `GET /projects/:id/tabs/:tab` redirects to `GET /projects/:id`.
- Tab response returns:
  - panel HTML for `#project-detail-tab-content`
  - OOB tabstrip nav with `id="project-detail-tabstrip"` and `hx-swap-oob="outerHTML"`
- Active tab uses `aria-selected="true"` and `tabindex="0"`, others use `false`/`-1`.

## Panel/Focus Contract

- Panel root always has:
  - `id="project-detail-tab-content"`
  - `role="tabpanel"`
  - `aria-labelledby="<active-tab-id>"`
  - `tabindex="-1"`
- Focus move after swap is implemented by `autofocus` on inserted panel root.

## Status/Errors

- Unauthorized (`!project.read`) → HTTP `403`.
- Unknown/invalid `id` or unknown `tab` → HTTP `404`.
- Unauthenticated access follows stack login redirect behavior.

## Story 2.12 Hand-off

- Story 2.12 whole-panel update targets `#project-detail` with `hx-swap="innerHTML"`.
