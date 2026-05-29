# ComplianceTile Component

## Purpose

Display a compliance score with threshold-derived semantic color; serves as the canonical `#compliance-tile` OOB target.

## Required Props

| Prop | Type | Description |
|---|---|---|
| `score` | `int? (0–100, null for no-data)` | The compliance score to display. `null` or out-of-range renders the no-data variant (em-dash). |
| `label` | `string` | Human-cased display label (e.g. `"Compliance"`, `"Portfolio Compliance"`). CSS applies uppercase via `.compliance-tile__label`. |
| `id` | `string` | Caller-supplied element id — use `"compliance-tile"` for project context, `"compliance-tile-portfolio"` for portfolio context. |

`context` is documentation-only — it appears in `canonical.html` variant labels but is not a template prop. The wrapper receives `id` directly so the same wrapper serves both contexts.

## Variant List

1. `healthy-project` — score=95, label=Compliance, id=compliance-tile
2. `watch-project` — score=82, label=Compliance, id=compliance-tile
3. `concern-project` — score=58, label=Compliance, id=compliance-tile
4. `critical-project` — score=37, label=Compliance, id=compliance-tile
5. `healthy-portfolio` — score=91, label=Portfolio Compliance, id=compliance-tile-portfolio
6. `critical-portfolio` — score=42, label=Portfolio Compliance, id=compliance-tile-portfolio
7. `no-data-project` — score=null, label=Compliance, id=compliance-tile
8. `boundary-90` — score=90, label=Compliance, id=compliance-tile (verifies ≥90 inclusive)
9. `boundary-70` — score=70, label=Compliance, id=compliance-tile (verifies ≥70 inclusive)
10. `boundary-50` — score=50, label=Compliance, id=compliance-tile (verifies ≥50 inclusive)
11. `boundary-49` — score=49, label=Compliance, id=compliance-tile (verifies <50 → Critical)

## ARIA Invariants

All three of the following attributes are mandatory on the outer `<section>` and must never be weakened:

- `role="status"` — identifies the region as a live status region
- `aria-live="polite"` — announces score changes after the current speech completes
- `aria-atomic="true"` — announces the entire tile as a unit on OOB swap (e.g. "Compliance, 87, Watch")

## Threshold Table

| Band (score) | Semantic color class | Threshold word |
|---|---|---|
| ≥ 90 (Healthy) | `text-success` | Healthy |
| 70 – 89 (Watch) | `text-warning` | Watch |
| 50 – 69 (Concern) | `text-warning-strong` | Concern |
| < 50 (Critical) | `text-danger` | Critical |
| `null` or out-of-range | `text-neutral` | _(omitted — no threshold `<p>` rendered)_ |

Source: UX-DR §"Compliance score thresholds" (ux-design-specification.md:478–489).

## Allowed Class Vocabulary

Value element classes: `text-success`, `text-warning`, `text-warning-strong`, `text-danger`, `text-neutral` (no-data), `text-3xl`, `font-bold`, `tnum`.

- `tnum` (tabular numbers) is mandatory on every populated variant — prevents column jitter on OOB updates.
- `text-warning-strong` maps to `--color-warning-strong` in `_tokens.css` (amber-700 light / amber-300 dark).
- Class `.compliance-tile__label` has `text-transform: uppercase` declared in `_tokens.css`.

## Snapshot-Equality Requirement

Per-stack wrappers MUST render output byte-equal to the matching variant block in `canonical.html` after the standard normalization defined in `fieldmark_shared/CLAUDE.md` §"Snapshot-test pipeline".

## Unknown-Vocabulary Handling

Score values outside 0–100 (i.e. `< 0` or `> 100`) render as the `no-data` variant (em-dash, `text-neutral`, threshold `<p>` omitted). A per-stack unit test asserts this. No runtime warning logger this story — see `_bmad-output/implementation-artifacts/deferred-work.md` §"Story 2.4-followup — unknown-token runtime warning logger".

## OOB-Target Invariant

This component is a pure OOB **target** — it never emits HTMX producer attributes (`hx-get`, `hx-post`, `hx-target`, `hx-swap`, `hx-trigger`) or `<script>` tags. Downstream stories (2.10, 2.11, 2.12, 5.5) emit OOB swaps that replace the entire `<section>`; the `aria-atomic="true"` causes the screen reader to announce the whole tile as a unit.
