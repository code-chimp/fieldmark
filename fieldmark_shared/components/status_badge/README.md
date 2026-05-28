# StatusBadge

## Purpose

Renders an entity state or severity as a Basecoat badge with deterministic semantic color.

## Required Props

| Prop | Type |
|---|---|
| `entity` | string: `Project`, `Inspection`, `Violation`, `CorrectiveAction`, `Severity` |
| `value` | string: canonical state or severity value |
| `severity` | optional string used when `entity=Violation` and `value=Open` |

## Variant List

`project-active`, `project-on-hold`, `project-closed`, `inspection-scheduled`, `inspection-in-progress`, `inspection-completed-pass`, `inspection-completed-conditional`, `inspection-completed-fail`, `inspection-cancelled`, `violation-open-critical-high`, `violation-open-medium-low`, `violation-in-progress`, `violation-resolved`, `violation-voided`, `corrective-action-submitted`, `corrective-action-under-review`, `corrective-action-approved`, `corrective-action-rejected`, `severity-critical`, `severity-high`, `severity-medium`, `severity-low`, `unknown`.

## ARIA Invariants

No ARIA role is applied by the badge itself; the visible text always names the state so color is never the sole information carrier.

## Allowed Basecoat / Utility Class Vocabulary

`badge`, `badge-project-active`, `badge-project-onhold`, `badge-project-closed`, `badge-inspection-scheduled`, `badge-inspection-inprogress`, `badge-inspection-pass`, `badge-inspection-conditional`, `badge-inspection-fail`, `badge-inspection-cancelled`, `badge-violation-open-high`, `badge-violation-open-low`, `badge-violation-inprogress`, `badge-violation-resolved`, `badge-violation-voided`, `badge-ca-submitted`, `badge-ca-underreview`, `badge-ca-approved`, `badge-ca-rejected`, `badge-severity-critical`, `badge-severity-high`, `badge-severity-medium`, `badge-severity-low`, `badge-bump`, `badge-unknown`.

## Snapshot Equality Requirement

Per-stack wrappers MUST render output byte-equal to the matching variant block in `canonical.html` after the standard normalization defined in `fieldmark_shared/CLAUDE.md` §'Snapshot-test pipeline'.

## Unknown Vocabulary Handling

Unknown values render the `--unknown` fallback variant per Dev Notes §'Decision — unknown-token handling'.
