---
name: feedback_story_2_4_rule_changes
description: Rules and reference doc updates made after Story 2.4's 5-round review to prevent recurring error patterns
metadata:
  type: feedback
---

After Story 2.4 required 5 review rounds and 24 patch items, rules were added to 7 files to prevent the recurring patterns. Summary of what was added and where:

**Why:** The same classes of mistake (grep guard missing on the 4th component, .NET model property without `S()`, Django `{% if value %}` treating `"0"` as falsy, XSS tests missing the negative assertion) appeared independently in multiple rounds. These are now codified rules, not just review observations.

**How to apply:** When implementing any component story, cross-check each rule before submitting.

---

## Changes made

### `FieldMark/CLAUDE.md` — new "Razor Component Rules" section
- All `.cshtml` model property reads MUST use `S()` helper (not `@Model.Foo` directly)
- No `Html.Raw` in component templates (XSS vector)
- Every component test file MUST contain an `Html.Raw` grep guard `[Fact]`

### `fieldmark_py/CLAUDE.md` — Django Component Template Rules + Coding Standards addition
- `{% if value %}` is falsy for `"0"` — use `{% if value is not None and value != "" %}` for numeric-like props
- No `|safe` filter in component templates
- Every component test file MUST contain a `|safe` grep guard test

### `fieldmark-go/CLAUDE.md` — new "Go Component Template Rules" section
- No `template.HTML(...)` casts in templates or view models
- Every component test file MUST contain a `template.HTML(` grep guard sub-test
- Table-driven tests: Go 1.22+ gives range variables per-iteration scope, so `tc := tc` is no longer needed and should not be used (it was removed from `inline_alert_test.go` in round 7). The real concern is mutating a shared VM struct inside the loop body before `t.Run`; construct a fresh VM per sub-test instead

### `docs/reference/component-edge-case-checklist.md`
- Category 9 updated: tests MUST cover empty string, whitespace-only, AND zero-value as distinct cases
- NEW Category 10: Cross-component test parity — for every test type in {grep-guard, unknown-fallback, xss-positive, xss-negative, edge-cases}, assert coverage across all sibling components, not N−1

### `docs/reference/security-defaults.md`
- NEW section 3a: XSS round-trip test completeness — requires (1) bare `<script>alert(1)</script>` payload, (2) `Contains(escaped)`, (3) `NotContains(raw)`, and (4) all user-visible props tested including conditional props (must pass non-empty payload to trigger the branch)

### `fieldmark_shared/CLAUDE.md` — two new rules
- Accessibility media queries (`forced-colors`, `prefers-reduced-motion`) belong exclusively in `_a11y.css` — never in `_tokens.css` or other partials
- `canonical.html` and the component `README.md` Variant List MUST be updated in the same commit when a variant is added

### `docs/reference/hard-rules.md`
- Added: Sibling component test parity rule (links to checklist §10)
