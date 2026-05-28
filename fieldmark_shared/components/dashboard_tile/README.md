# DashboardTile

## Purpose

Renders a compact dashboard summary tile with label, value, optional secondary text, and optional status-region semantics.

## Required Props

| Prop | Type |
|---|---|
| `tile_id` | string |
| `label` | string |
| `value` | optional string or number |
| `secondary` | optional string |
| `value_color` | optional string: `success`, `warning`, `danger`, `info`, `neutral` |
| `role_status` | bool |

## Variant List

`populated`, `zero-value`, `populated-with-secondary`, `populated-with-color`, `empty`, `status-region`.

## ARIA Invariants

Only OOB-updated consumer variants set `role="status"` on the outer section.

## Allowed Basecoat / Utility Class Vocabulary

`dashboard-tile`, `dashboard-tile__label`, `dashboard-tile__value`, `dashboard-tile__secondary`, `text-3xl`, `font-bold`, `tnum`, `text-success`, `text-warning`, `text-danger`, `text-info`, `text-neutral`.

## Snapshot Equality Requirement

Per-stack wrappers MUST render output byte-equal to the matching variant block in `canonical.html` after the standard normalization defined in `fieldmark_shared/CLAUDE.md` §'Snapshot-test pipeline'.

## Unknown Vocabulary Handling

Unknown `value_color` values render no semantic color utility class. DashboardTile has no `--unknown` fallback variant because its primary value remains visible as text without an additional status vocabulary signal.
