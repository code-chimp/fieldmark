# AuditRow

## Purpose

Renders one audit log entry as a receipt row with action, actor, timestamp, and optional before/after disclosure.

## Required Props

| Prop | Type |
|---|---|
| `action` | string: canonical audit action |
| `actor_name` | string |
| `occurred_at` | ISO-8601 UTC string |
| `relative` | string |
| `absolute` | string |
| `before_after_json` | optional compact JSON string |
| `expanded` | bool |

## Variant List

`default`, `with-disclosure-collapsed`, `with-disclosure-expanded`, `unknown-action`, `empty-actor`.

## ARIA Invariants

The `<li>` is intentionally bare; the live region is parent-owned by the consumer page. Disclosure state is carried by the native `<details open>` attribute; `<summary>` does not add `aria-expanded`.

## Allowed Basecoat / Utility Class Vocabulary

`audit-row`, `audit-row__action`, `audit-row__actor`, `audit-row__initials`, `audit-row__timestamp`, `audit-row__disclosure`, `badge`, `badge-audit-action`, `badge-unknown`, `tnum`, `font-mono`.

## Snapshot Equality Requirement

Per-stack wrappers MUST render output byte-equal to the matching variant block in `canonical.html` after the standard normalization defined in `fieldmark_shared/CLAUDE.md` §'Snapshot-test pipeline'.

## Unknown Vocabulary Handling

AuditRow has no `--unknown` fallback variant for the row itself. An unknown `action` value causes the embedded StatusBadge badge to render with `badge-unknown`; the row structure is otherwise identical to the `default` variant. The AuditRow variant list does not include an `--unknown` entry — only the five named variants above.
