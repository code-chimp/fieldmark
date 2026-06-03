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

## Audit Log Consumer Contract

Story 2.13 makes AuditRow live inside the Project Detail Audit tab. Consumers must use this exact parent structure:

- `<ul id="audit-log" role="list" aria-live="polite">`
- direct child `<li class="audit-row">` items rendered by the native AuditRow wrapper
- optional trailing `<li id="audit-log-load-more">` containing the load-more button

The Story 2.12 transition response prepends a new row with `hx-swap-oob="afterbegin:#audit-log"`, so the `#audit-log` id and direct-`<li>` child structure are load-bearing.

### Empty-State / OOB Coexistence

The empty-state message is rendered as a sibling `<li>` inside `#audit-log`.

- When the server re-renders the Audit tab after a transition (`current_tab=audit`), it suppresses the empty-state row and returns the new audit row in-band at the top of the list.
- When a transition response lands out-of-band while the existing client-side empty-state row is still present, HTMX prepends the new AuditRow before that sibling. Consumers should treat this as the chosen composition for Story 2.13 rather than replacing the client-side empty-state node in-place.

### Relative-Time Buckets

Use identical bucket wording across stacks:

- `< 60s` → `just now`
- `< 60m` → `N minute(s) ago`
- `< 24h` → `N hour(s) ago`
- `< 30d` → `N day(s) ago`
- `< 12 months` → `N month(s) ago`
- otherwise → `N year(s) ago`

### Keyset Load-More Contract

The load-more route is `GET /projects/<id>/audit-log?before_occurred_at=<iso8601-utc>&before_id=<uuid>`.

- Page size is `100`
- Ordering is `occurred_at DESC, id DESC`
- Cursor semantics are strict keyset: fetch rows where `(occurred_at, id) < (cursor_occurred_at, cursor_id)`
- Consumers render the next-page control as:
  `<li id="audit-log-load-more"><button hx-get="..." hx-target="closest li" hx-swap="outerHTML">Load more</button></li>`
- The fragment response replaces that trailing `<li>` with the next batch of `<li class="audit-row">` rows plus a refreshed trailing load-more `<li>` when more rows remain
