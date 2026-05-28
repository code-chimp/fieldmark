# InlineAlert

## Purpose

Renders an in-flow domain message with severity, icon, title, body, and optional metadata.

## Required Props

| Prop | Type |
|---|---|
| `severity` | string: `danger`, `warning`, `info`, `success` |
| `title` | string |
| `message` | string |
| `meta` | optional string |

## Variant List

`danger`, `warning`, `info`, `success`, `unknown`.

## ARIA Invariants

`danger` and `warning` render `role="alert"`; `info`, `success`, and `unknown` render `role="status"`. The icon is `aria-hidden="true"` and is always paired with visible title text.

## Allowed Basecoat / Utility Class Vocabulary

`alert`, `alert-danger`, `alert-warning`, `alert-info`, `alert-success`, `alert-unknown`, `alert-icon`, `alert-title`, `alert-message`, `alert-meta`.

## Snapshot Equality Requirement

Per-stack wrappers MUST render output byte-equal to the matching variant block in `canonical.html` after the standard normalization defined in `fieldmark_shared/CLAUDE.md` §'Snapshot-test pipeline'.

## Unknown Vocabulary Handling

Unknown values render the `--unknown` fallback variant per Dev Notes §'Decision — unknown-token handling'.
