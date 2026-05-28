# Story 2.4: Implement Phase-2 markup-only components — StatusBadge, InlineAlert, AuditRow, DashboardTile

Status: done

Epic: 2 — Project Lifecycle & Compliance Dashboard
Source AC: [_bmad-output/planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md](../planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md) §Story 2.4
Canonical examples directory: [fieldmark_shared/components/](../../fieldmark_shared/components/) — this story introduces the per-component sub-directory convention (`<component>/canonical.html` + `<component>/README.md`) alongside the existing flat `action_button.example.html` (Story 1.12) and `login-form.example.html` / `login-error-region.example.html` (Story 1.11).

## Story

As a developer rendering Project Detail (Story 2.11), the Compliance Dashboard (Story 2.10), and every state-change response that emits an audit row (Story 2.12 onward),
I want four small Basecoat-compliant wrapper templates per stack — **StatusBadge**, **InlineAlert**, **AuditRow**, **DashboardTile** — with byte-identical output across all three stacks against a checked-in canonical example,
So that subsequent stories can compose these screens without inventing markup per page, the state-machine-as-UI invariant (UX §"Critical Success Moments") holds across stacks, and the Phase-2 component gate for anchor screens is unblocked.

**Scope boundary:** this story produces (a) one canonical example directory per component under `fieldmark_shared/components/<component>/` with `canonical.html` (variant-delimited fixture, same `<!-- variant: name -->` convention as `action_button.example.html`) plus `README.md` (contract: required props, ARIA, allowed class vocabulary), (b) one markup-only wrapper template per component per stack in that stack's idiomatic component-template location, (c) per-stack snapshot tests that parse `canonical.html` by variant delimiter and assert byte-equality after the existing whitespace-and-attribute-order normalization used by the Story 1.11 / 1.12 snapshot harness, and (d) an `docs/reference/component-canonical-examples.md` index that lists every component, its `canonical.html` path, its per-stack wrapper paths, and its per-stack snapshot-test paths. **Out of scope:** ComplianceTile (Story 2.5 — needs the `#compliance-tile` OOB target wiring), EntityRail (Story 2.6 — responsive collapse), TabStrip (Story 2.7 — arrow-key JS), AGGridPanel (Story 2.9 — AG Grid bundle), any consumer page that *uses* these components (Stories 2.8 / 2.10 / 2.11+), any JavaScript (these are all zero-JS components), any audit-row OOB plumbing (markup only — the `aria-live="polite"` parent is rendered by AuditRow's *consumer page* in Story 2.13, not by the row itself), and the unknown-vocabulary-fallback runtime warning logger (deferred per "Decision — unknown-token handling" in Dev Notes).

## Acceptance Criteria

### AC1 — Canonical-example directories under `fieldmark_shared/components/`

**Given** the canonical-example contract from Story 1.12 (flat `action_button.example.html`)
**When** I inspect `fieldmark_shared/components/`
**Then** four new sub-directories exist with this exact layout:

```
fieldmark_shared/components/
├── status_badge/
│   ├── canonical.html
│   └── README.md
├── inline_alert/
│   ├── canonical.html
│   └── README.md
├── audit_row/
│   ├── canonical.html
│   └── README.md
├── dashboard_tile/
│   ├── canonical.html
│   └── README.md
├── action_button.example.html          # untouched (Story 1.12 — flat form remains)
├── login-form.example.html             # untouched (Story 1.11)
├── login-error-region.example.html     # untouched (Story 1.11)
└── README.md                           # MODIFY — add a "Per-component directories" section documenting the new convention alongside the legacy flat form
```

**And** each `canonical.html` follows the same variant-delimited format as `action_button.example.html` — a leading HTML comment block documenting `fixture:` inputs, then `<!-- variant: <name> (inputs: ...) -->` delimiters separating each variant's expected output.

**And** each `README.md` documents the component's contract in this fixed-order: (1) Purpose (one sentence), (2) Required props with types, (3) Variant list (the names that appear as `<!-- variant: -->` delimiters in `canonical.html`), (4) ARIA invariants (`role`, `aria-live`, `aria-expanded`, etc.), (5) Allowed Basecoat / utility class vocabulary, (6) Snapshot-equality requirement (one-line: "Per-stack wrappers MUST render output byte-equal to the matching variant block in `canonical.html` after the standard normalization defined in `fieldmark_shared/CLAUDE.md` §'Snapshot-test pipeline'"), (7) Unknown-vocabulary handling (one-line: "Unknown values render the `--unknown` fallback variant per Dev Notes §'Decision — unknown-token handling'").

**And** `fieldmark_shared/components/README.md` is MODIFIED with a new "Per-component directories" section that records the introduction of the directory convention in this story, naming the four new directories, and notes that the flat form is acceptable for existing components and that future component stories may use either form at the story author's discretion. The section explicitly does not deprecate the flat form (changing existing examples is out of scope).

### AC2 — StatusBadge component (UX-DR §"StatusBadge", UX-DR §"Status badge vocabulary")

**Given** the canonical entity-state vocabulary in [ux-design-specification.md:452–476](../planning-artifacts/ux-design-specification.md) and the audit-actions table in [docs/reference/audit-actions.md](../../docs/reference/audit-actions.md)
**When** I inspect `fieldmark_shared/components/status_badge/canonical.html`
**Then** the file contains one variant block per row in the UX status-badge vocabulary table — `project-active`, `project-on-hold`, `project-closed`, `inspection-scheduled`, `inspection-in-progress`, `inspection-completed-pass`, `inspection-completed-conditional`, `inspection-completed-fail`, `inspection-cancelled`, `violation-open-critical-high`, `violation-open-medium-low`, `violation-in-progress`, `violation-resolved`, `violation-voided`, `corrective-action-submitted`, `corrective-action-under-review`, `corrective-action-approved`, `corrective-action-rejected`, `severity-critical`, `severity-high`, `severity-medium`, `severity-low` — plus one `unknown` variant for the unknown-vocabulary fallback. Each variant block renders the Basecoat `<span class="badge ...">` markup with the semantic-color CSS variable token classes that the existing `_tokens.css` already declares (do not invent new tokens this story).

**And** the visible text label inside each badge **always** names the state in human-readable form (e.g., `Active`, `On Hold`, `Critical`) — color is never the sole information carrier (WCAG 1.4.1, UX §"Information Architecture"). The label string is part of the byte-equality contract.

**And** the `severity-critical` and `severity-high` variants include the `badge-bump` class per UX §"Status badges" (severity bump). The other variants do not.

**Given** each stack's wrapper
**When** I render it with `entity=Project, value=Active`
**Then** the output equals the `project-active` variant block byte-for-byte after the Story 1.11 normalization (whitespace collapse + alphabetical attribute order + trimmed lines).

**Per-stack wrapper paths** (idiomatic per stack — no shared template, no symlinked partial):

- **.NET (Razor partial):** `FieldMark/FieldMark.Web/Pages/Shared/Components/_StatusBadge.cshtml` (NEW). Invoked as `<partial name="Shared/Components/_StatusBadge" model="@(new StatusBadgeViewModel(entity, value, severity))" />`. The view model lives in the **same file** at the bottom (`@functions { public record StatusBadgeViewModel(...) }`) — do not create a sibling `.cs` file; the Story 1.12 `_ActionButton.cshtml` precedent is "view model in-file".
- **Django (template include):** `fieldmark_py/templates/components/_status_badge.html` (NEW). Invoked as `{% include "components/_status_badge.html" with entity="Project" value="Active" severity=None %}`. The Story 1.12 `_action_button.html` precedent uses the same `{% include … with … %}` pattern.
- **Go (html/template `{{define}}` block):** `fieldmark-go/internal/web/templates/components/status_badge.html` (NEW). Defines `{{define "status_badge"}}…{{end}}` (snake_case template name matching the Story 1.12 `action_button` precedent in [components/action_button.html](../../fieldmark-go/internal/web/templates/components/action_button.html)). Invoked as `{{template "status_badge" $statusBadgeArgs}}` where `$statusBadgeArgs` is a context struct value the caller constructs.

**And** for an `unknown` token (e.g., `entity=Violation, value=Foobar`), all three stacks render the `unknown` variant block — a `<span class="badge badge-unknown">Foobar</span>` with a `1px dashed` outline (declare `.badge-unknown` in `fieldmark_shared/src/_tokens.css` alongside the other badge variants; this is the only `fieldmark_shared/src/` edit this story makes). **No runtime warning is logged** this story — see Dev Notes §"Decision — unknown-token handling" for the rationale. A unit test asserts the fallback class is emitted.

### AC3 — InlineAlert component (UX-DR §"InlineAlert", UX §Pattern 3 — Errors Render In Place)

**Given** the InlineAlert contract in [ux-design-specification.md:887–893](../planning-artifacts/ux-design-specification.md)
**When** I inspect `fieldmark_shared/components/inline_alert/canonical.html`
**Then** the file contains four variants — `danger`, `warning`, `info`, `success` — plus an `unknown` variant. Each variant renders a Basecoat `<div class="alert alert-<severity>">` with:
- `role="alert"` for `danger` and `warning` (assertive, per WAI-ARIA — fires the SR announcement on insertion).
- `role="status"` for `info` and `success` (polite).
- A Lucide-sourced inline `<svg aria-hidden="true">` icon paired with the title text (icon-text pairing per UX §"Information Architecture" — icon is never the sole carrier of meaning). Use these Lucide icons by name: `triangle-alert` (danger), `alert-triangle` is the legacy name — verify in the existing `_ThemeToggle` / `_FlashRegion` partials which Lucide naming convention this codebase has settled on; do not introduce a new Lucide build pipeline. Inline the SVG path directly into `canonical.html` so the byte-equality contract is closed — do not reference an external sprite this story.
- A `<strong class="alert-title">` title element followed by a `<p class="alert-message">` body. Optional `<p class="alert-meta">` metadata line per the UX prop list.
- The `unknown` variant collapses to `role="status"` (the safe default) with `class="alert alert-unknown"`.

**And** each stack's wrapper renders the matching variant byte-equal to `canonical.html` for the same input set.

**Per-stack wrapper paths:**

- **.NET:** `FieldMark/FieldMark.Web/Pages/Shared/Components/_InlineAlert.cshtml`.
- **Django:** `fieldmark_py/templates/components/_inline_alert.html`.
- **Go:** `fieldmark-go/internal/web/templates/components/inline_alert.html` defining `{{define "inline_alert"}}…{{end}}`.

**Given** any user-supplied prop string is rendered into the alert (`title`, `message`, `meta`)
**When** the prop contains characters in the XSS-prone set (`<`, `>`, `&`, `"`, `'`)
**Then** the rendered output contains the HTML-entity-escaped form, never the raw character. Each framework's default auto-escaping covers this — Razor `@Model.Message`, Django `{{ message }}` (no `|safe`), Go `html/template` `{{.Message}}` (no `template.HTML`). A per-stack test asserts a payload containing `<script>alert(1)</script>` round-trips as `&lt;script&gt;alert(1)&lt;/script&gt;` in the rendered output. **No `|safe`, no `Html.Raw`, no `template.HTML`** anywhere in this story.

### AC4 — AuditRow component (UX-DR §"AuditRow", UX §Pattern 4 — Audit Row As Receipt)

**Given** the AuditRow contract in [ux-design-specification.md:851–857](../planning-artifacts/ux-design-specification.md) and the canonical audit-action vocabulary in [docs/reference/audit-actions.md](../../docs/reference/audit-actions.md)
**When** I inspect `fieldmark_shared/components/audit_row/canonical.html`
**Then** the file contains these variants: `default` (action + actor + relative timestamp, no before/after disclosure), `with-disclosure-collapsed` (disclosure present, `aria-expanded="false"`, before/after JSON not visible), `with-disclosure-expanded` (`aria-expanded="true"`, JSON shown inside `<pre><code class="font-mono">` in JetBrains Mono per UX-DR §"AuditRow"), `unknown-action` (action token not in the canonical list — renders StatusBadge `unknown` variant inline for the action; rest of the row degrades gracefully), and `empty-actor` (actor name is empty or whitespace — renders the fallback initials `??` and display name `unnamed` per [component-edge-case-checklist.md §9](../../docs/reference/component-edge-case-checklist.md)).

**And** each row markup is:

```html
<li class="audit-row" data-audit-action="<ActionString>">
  <span class="audit-row__action">{{ StatusBadge entity=Audit value=<action> }}</span>
  <span class="audit-row__actor">{{ actor_name }}</span>
  <time class="audit-row__timestamp tnum"
        datetime="<ISO-8601-UTC>"
        title="<absolute-formatted>"
        >{{ relative }}</time>
  <details class="audit-row__disclosure">
    <summary aria-expanded="false">Show change</summary>
    <pre><code class="font-mono">{{ before_after_json_compact }}</code></pre>
  </details>
</li>
```

**And** the `<li>` is intentionally bare — the `aria-live="polite"` parent (`<ul id="audit-log" role="log" aria-live="polite">`) is owned by the **consumer page** (the Project Audit Log tab, Story 2.13), not by AuditRow itself. The AuditRow `README.md` documents this as the "live region is parent-owned" invariant. A consumer-side conformance assertion will land in Story 2.13.

**And** the `<time datetime="">` value is **always the ISO-8601 UTC string** for the audit-entry's `occurred_at`. The visible content is the relative form (`3 minutes ago`); the `title` attribute is the absolute form (`2026-05-28 14:23:01 UTC`). Cross-stack relative-time formatting MUST agree to the second — use each stack's built-in (`TimeSince` in .NET, `humanize.naturaltime` in Django, a per-stack helper in Go) and lock the output via a deterministic test fixture (mock-now to a fixed instant; assert the rendered string).

**Per-stack wrapper paths:**

- **.NET:** `FieldMark/FieldMark.Web/Pages/Shared/Components/_AuditRow.cshtml`.
- **Django:** `fieldmark_py/templates/components/_audit_row.html`.
- **Go:** `fieldmark-go/internal/web/templates/components/audit_row.html` (`{{define "audit_row"}}`).

**Given** before/after JSON containing `<script>` or other escape-sensitive payload
**When** the row renders with disclosure
**Then** the JSON content is HTML-entity-escaped inside `<pre><code>` (it is text content, not markup — auto-escape applies). A per-stack test asserts a `{"value": "<script>"}` payload round-trips as `{"value": "&lt;script&gt;"}` in the rendered output.

### AC5 — DashboardTile component (UX-DR §"DashboardTile")

**Given** the DashboardTile contract in [ux-design-specification.md:879–885](../planning-artifacts/ux-design-specification.md)
**When** I inspect `fieldmark_shared/components/dashboard_tile/canonical.html`
**Then** the file contains these variants: `populated` (label + numeric value), `populated-with-secondary` (label + value + secondary text or sub-badges), `populated-with-color` (label + value with a semantic `--color-<token>` class applied to the value text), `empty` (label + `—` em-dash per UX), `status-region` (`role="status"` on the outer container — the variant the consumer uses when the tile will receive OOB updates from a state-change response). Five variants total.

**And** each variant renders:

```html
<section class="dashboard-tile" id="<tile-id>"  <!-- role="status" added for the status-region variant -->>
  <p class="dashboard-tile__label">{{ LABEL_UPPERCASE }}</p>
  <p class="dashboard-tile__value text-3xl font-bold tnum">{{ value }}</p>
  <p class="dashboard-tile__secondary">{{ secondary }}</p>  <!-- only when secondary is provided -->
</section>
```

**And** the `LABEL_UPPERCASE` content is **CSS-uppercased via `text-transform: uppercase`** on the `.dashboard-tile__label` class — the wrapper's input prop is the human-cased label (`"Open Violations"`), and the visual uppercase comes from CSS. Do not call `.upper()` / `.ToUpper()` / `strings.ToUpper(...)` in the wrapper; the byte-equality contract is on the rendered HTML, which carries the original-case text. (This decision keeps locale-correct uppercase deferrable; the CSS rule is locale-neutral and matches the UX-DR specification of the visual outcome.)

**And** the `populated-with-color` variant adds a single `text-<semantic-token>` class to `.dashboard-tile__value` — the token is one of the existing `_tokens.css` semantic colors (`text-success` / `text-warning` / `text-danger` / `text-info` / `text-neutral`). Unknown color tokens fall through to no class (no `.badge-unknown` equivalent here — DashboardTile color is decorative, not state-bearing).

**Per-stack wrapper paths:**

- **.NET:** `FieldMark/FieldMark.Web/Pages/Shared/Components/_DashboardTile.cshtml`.
- **Django:** `fieldmark_py/templates/components/_dashboard_tile.html`.
- **Go:** `fieldmark-go/internal/web/templates/components/dashboard_tile.html` (`{{define "dashboard_tile"}}`).

### AC6 — Cross-stack snapshot-test conformance (UX-DR §"Component identity is markup, not class")

**Given** each stack's existing snapshot harness from Story 1.11 (login-form snapshot) and Story 1.12 (action-button snapshot)
**When** I run the per-stack snapshot tests for the four new components
**Then** for **every** variant block in each `canonical.html`, the corresponding stack wrapper renders output byte-equal to that block after the standard normalization (extract the named region, strip per-stack antiforgery noise where applicable, collapse whitespace runs, trim lines, sort attributes alphabetically — same pipeline as `fieldmark_shared/CLAUDE.md` §"Snapshot-test pipeline").

**And** the test parses `canonical.html` by `<!-- variant: <name> -->` delimiter at test time — the variant name (the regex-captured identifier between `variant:` and `(inputs:` or `-->`) is the test-case name. The fixture inputs documented in the `inputs:` clause are *human-readable hints only* for the test author; the test code maps variant-name → wrapper-input-args explicitly per stack. Do not parse the `inputs:` clause programmatically — it is documentation, not data.

**And** the test reads `canonical.html` from disk by computing a path relative to the test-run working directory, walking up to the repo root, then descending into `fieldmark_shared/components/<component>/canonical.html`. This is the same path-resolution pattern the Story 2.2 `AuditActionConformanceTests` uses for `docs/reference/audit-actions.json` — do not reinvent the path-walker.

**Per-stack test paths:**

- **.NET:** `FieldMark/FieldMark.Tests.Web/Components/StatusBadgeSnapshotTests.cs`, `InlineAlertSnapshotTests.cs`, `AuditRowSnapshotTests.cs`, `DashboardTileSnapshotTests.cs`. The harness uses `WebApplicationFactory<Program>` to render the Razor partial via a dedicated `/_test/render-partial/<name>` endpoint, the same scaffold pattern Story 1.12 introduced — confirm the existing Story 1.12 endpoint shape before reinventing the host. If the project `FieldMark.Tests.Web` does not yet exist, create it as a sibling to `FieldMark.Tests.Domain` and `FieldMark.Tests.Integration`; if `FieldMark.Tests.Integration` already hosts wrapper-rendering helpers, place the four files there and skip the new project.
- **Django:** `fieldmark_py/components/tests/test_status_badge_snapshot.py`, `test_inline_alert_snapshot.py`, `test_audit_row_snapshot.py`, `test_dashboard_tile_snapshot.py`. Use `django.test.RequestFactory` + `django.template.loader.render_to_string` directly (no live server needed); the Story 1.12 `_action_button.html` snapshot test is the precedent. Place the tests in a new `fieldmark_py/components/tests/` package if `components/` is a fresh app, or in `fieldmark_py/tests/components/` if there's an existing site-wide `tests/` location — verify before writing.
- **Go:** `fieldmark-go/internal/web/templates/components/status_badge_test.go`, `inline_alert_test.go`, `audit_row_test.go`, `dashboard_tile_test.go`. Use the existing harness from `action_button_test.go` ([components/action_button_test.go](../../fieldmark-go/internal/web/templates/components/action_button_test.go)) — execute the `{{define}}` block against a context struct, capture the writer output, normalize, compare.

**And** the snapshot test for `StatusBadge` exercises **all** 23 variant blocks (the 22 vocabulary variants plus `unknown`), not a subset. The other three components exercise their full variant lists (4 + 1 for InlineAlert; 5 for AuditRow; 5 for DashboardTile). Total: 23 + 5 + 5 + 5 = **38 snapshot assertions per stack** = 114 across the three stacks. The test harness MUST emit one logical assertion per variant so a failure names the variant (do not `assertAll` — name the variant in the failure).

### AC7 — `docs/reference/component-canonical-examples.md` index (cross-stack contract doc)

**Given** the Cross-Stack Architecture Principle (root [CLAUDE.md](../../CLAUDE.md) §Cross-Stack Architecture Principle) — cross-stack invariants live as **documentation contracts**, not as shared code
**When** I inspect `docs/reference/component-canonical-examples.md`
**Then** the document is **NEW** and contains, in this order:
1. A short Status block ("populated by Story 2.4, 2026-05-28") mirroring [docs/reference/audit-actions.md](../../docs/reference/audit-actions.md).
2. A "Why" pointer to the Cross-Stack Architecture Principle.
3. A **Component Index** table with columns: `Component`, `Canonical example`, `README`, `.NET wrapper`, `Django wrapper`, `Go wrapper`, `.NET test`, `Django test`, `Go test` — one row per component (initially four rows: StatusBadge, InlineAlert, AuditRow, DashboardTile). Future component stories add rows.
4. A "Snapshot-test pipeline" section that points to `fieldmark_shared/CLAUDE.md` §"Snapshot-test pipeline" (the canonical description) and adds component-specific notes only where the per-component normalization deviates — Story 2.4 has no deviations, so this section is one paragraph that says so.
5. A "Change Procedure" section: adding a component is a four-step change — (a) author `fieldmark_shared/components/<name>/canonical.html` + `README.md`, (b) implement the per-stack wrappers, (c) implement the per-stack snapshot tests, (d) append a row to the Component Index above. Mirrors the `audit-actions.md` change-procedure pattern.

**And** the top-of-file comment in each per-stack wrapper template references this document URL (the same convention `AuditAction.cs` / `actions.py` / `audit_action.go` follow).

### AC8 — Component edge-case checklist coverage (per [component-edge-case-checklist.md](../../docs/reference/component-edge-case-checklist.md))

Walked the nine categories — only the **applicable** ones get an AC:

**Given** category 1 (unknown enum / vocabulary values)
**When** any of the three vocab-bearing components (StatusBadge entity-state value, InlineAlert severity, AuditRow action) receives an unknown token
**Then** the wrapper renders the documented `unknown` variant (`.badge-unknown` dashed outline for StatusBadge; `alert-unknown` neutral surface for InlineAlert; StatusBadge `unknown` inline for AuditRow's action slot). A per-stack unit test asserts the fallback class is emitted for an unknown input. **No runtime warning logger this story** — see Dev Notes §"Decision — unknown-token handling".

**Given** category 6 (text overflow & special characters in user-visible strings)
**When** AuditRow, InlineAlert, or DashboardTile render user-supplied strings (actor name, alert title/message, tile secondary text, audit before/after JSON)
**Then** each framework's default auto-escaping converts `<`, `>`, `&`, `"`, `'` to entities — verified by the XSS-payload round-trip tests in AC3 and AC4. Long content (descriptions over ~120 chars, JSON over ~200 chars) clips via `text-overflow: ellipsis` on the relevant text-row class **or** is wrapped in a `<details>` disclosure when the component already uses one (AuditRow). Add no new truncation CSS this story — the Story 1.14 `_a11y.css` / `_components.css` `.truncate` utility is the existing precedent; use it on `.audit-row__actor`, `.alert-message`, `.dashboard-tile__secondary` as appropriate.

**Given** category 8 (forced-colors / high-contrast mode)
**When** I render any StatusBadge or InlineAlert in `forced-colors: active`
**Then** the existing `_a11y.css` rule landed in Story 1.14 (`@media (forced-colors: active) { .badge, .alert { border: 1px solid ButtonText; forced-color-adjust: auto; } }`) provides a visible border so the badge/alert remains distinguishable. Verify the rule already covers `.badge` and `.alert`; if it does not, **extend the existing rule in [fieldmark_shared/src/_a11y.css](../../fieldmark_shared/src/_a11y.css) — do not create a new file**. Color is never the sole information carrier (the text label per AC2 / AC3 is the primary carrier; forced-colors is the secondary defense).

**Given** category 9 (empty / whitespace text input to derived values)
**When** AuditRow receives an empty or whitespace-only `actor_name`
**Then** the row renders the deterministic fallback `??` for initials (if the wrapper derives initials this story — it does not; initials are an Avatar concern out of scope here) and `unnamed` for the display name. A per-stack unit test for both empty-string and whitespace-only inputs.

**Given** categories 2 (font load), 3 (JS init), 4 (AG Grid overlays), 5 (stacking), 7 (reduced motion)
**When** I evaluate these against this story's deliverables
**Then** they are **N/A** with this rationale:
- **2 (font load):** AuditRow uses JetBrains Mono inside `<pre><code class="font-mono">`. The Story 1.14 `font-display: swap` rule in `_fonts.css` is the existing canonical resolution and requires no extension here. CLS guard tests already live in Story 1.14's Playwright suite — adding a new lane this story is out of scope.
- **3 (JS init):** All four components are zero-JS. No initialization marker, no progressive enhancement needed.
- **4 (AG Grid overlays):** Not an AG Grid story.
- **5 (stacking):** No queueing components introduced this story.
- **7 (reduced motion):** No transitions or animations are introduced; the global `@media (prefers-reduced-motion: reduce)` rule in `_a11y.css` (Story 1.14) covers any inherited Basecoat animation.

### AC9 — Security-defaults checklist coverage (per [security-defaults.md](../../docs/reference/security-defaults.md))

Walked the seven categories — only the **applicable** one gets an AC:

**Given** category 3 (allowlist validation on writes) — **adapted here** for output-escaping on read since this story has no writes
**When** any wrapper renders a user-supplied string (alert title / message / meta; audit actor name; audit before/after JSON; tile secondary text)
**Then** the framework's default auto-escaping is in force — verified by the per-stack XSS-payload round-trip tests in AC3 and AC4. **No `|safe` (Django), no `Html.Raw` / `@Html.Raw` (Razor), no `template.HTML` cast (Go) anywhere in any wrapper this story.** A grep-style conformance check in the per-stack lint lane confirms the absence of these tokens within the four new wrapper files.

**Given** categories 1 (open-redirect), 2 (cookie attributes), 4 (dynamic RegExp), 5 (filesystem writes), 6 (CSRF posture), 7 (stub-auth warnings)
**When** I evaluate against this story's deliverables
**Then** they are **N/A** — no redirects, no cookies, no regex on user input, no filesystem writes, no routes (the wrappers are markup-only — they are invoked from server-rendered pages but introduce no new HTTP endpoint), no auth changes.

### AC10 — Cross-stack architecture principle three-deliverable check (root [CLAUDE.md](../../CLAUDE.md))

This story introduces **one cross-stack contract** — the `<component>/canonical.html` byte-equality contract — and produces all three required deliverables:

1. **Documentation contract:** `docs/reference/component-canonical-examples.md` (AC7) + the per-component `README.md` files (AC1).
2. **Native implementation per stack:** the four wrapper templates per stack in idiomatic locations (AC2 / AC3 / AC4 / AC5).
3. **Per-stack conformance test:** the 38 snapshot assertions per stack (AC6).

**And** there is **no new file** in `fieldmark_shared/` that lists vocabulary tokens, action strings, or severity names — the `canonical.html` files are visual fixtures, not symbol manifests. The status-badge vocabulary remains DDL-owned (severity strings, audit-action strings already published in `docs/reference/audit-actions.md`).

**And** there is **no shared template engine, no symlinked partial** — each stack's wrapper lives natively under its own template tree.

### AC11 — `make parity` clean, no new routes introduced

**Given** all wrappers + tests land
**When** I run `make parity` from the repo root
**Then** the route-parity script reports the **same** drift baseline as Story 2.3 — no new `GET /…` or `POST /…` endpoints are introduced this story. (The Razor `/_test/render-partial/...` endpoint, if added, MUST be gated behind `#if DEBUG` or the equivalent build-time exclusion so it does not appear in the production route table or the parity diff. The Story 1.12 precedent applies.) `pg_indexes` diff: zero (no DB changes).

### AC12 — Build, type, lint, and test gates green on every stack

- **.NET:** `cd FieldMark && dotnet csharpier check . && dotnet build && dotnet test && dotnet test FieldMark.Tests.Integration/FieldMark.Tests.Integration.csproj` — clean. The four snapshot test classes pass with one `[Theory]` row per variant.
- **Django:** `cd fieldmark_py && uv run ruff check . && uv run mypy . && uv run pytest && uv run pytest -m integration` — clean. The four snapshot tests pass with one parametrized case per variant.
- **Go:** `cd fieldmark-go && make check && go test ./... && go test -tags=integration ./...` — clean. The four `*_test.go` files exercise each variant via table-driven sub-tests.
- **`fieldmark_shared`:** `cd fieldmark_shared && pnpm install && pnpm run build` — clean; `dist/fieldmark.css` regenerated and committed (the `.badge-unknown` rule and any other `_tokens.css` edits propagate into the compiled bundle).
- From repo root: `make parity` exits 0 (AC11) and `make test-all` (the canonical pre-merge gate from Story 2.2 round 3) exits 0.


- [x] **Task 1: Author canonical examples + per-component READMEs in `fieldmark_shared/`** (AC: #1, #2, #3, #4, #5, #10)
  - [x] 1.1 Create `fieldmark_shared/components/status_badge/canonical.html` with 23 variant blocks (22 vocabulary + `unknown`).
  - [x] 1.2 Create `fieldmark_shared/components/status_badge/README.md` per AC1 §contract-fixed-order.
  - [x] 1.3 Create `fieldmark_shared/components/inline_alert/canonical.html` (5 variants) + `README.md`.
  - [x] 1.4 Create `fieldmark_shared/components/audit_row/canonical.html` (5 variants) + `README.md`.
  - [x] 1.5 Create `fieldmark_shared/components/dashboard_tile/canonical.html` (5 variants) + `README.md`.
  - [x] 1.6 Modify `fieldmark_shared/components/README.md` — add "Per-component directories" section documenting the new convention alongside the legacy flat form (do not deprecate the flat form).
  - [x] 1.7 Add `.badge-unknown` and `.alert-unknown` rules to `fieldmark_shared/src/_tokens.css` (dashed border, neutral surface). Run `pnpm run build`. Commit the regenerated `dist/fieldmark.css`.
  - [x] 1.8 Verify the existing `@media (forced-colors: active)` rule in `_a11y.css` covers `.badge` and `.alert`. If absent, extend the existing rule (do not create a new file).

- [x] **Task 2: Author the cross-stack contract doc** (AC: #7, #10)
  - [x] 2.1 Create `docs/reference/component-canonical-examples.md` per AC7 §five-section-order.
  - [x] 2.2 Each per-stack wrapper file (Tasks 3 / 4 / 5) gets a top-of-file comment referencing this doc URL.

- [x] **Task 3: .NET wrappers + snapshot tests** (AC: #2, #3, #4, #5, #6, #9, #12)
  - [x] 3.1 `FieldMark/FieldMark.Web/Pages/Shared/Components/_StatusBadge.cshtml` — partial + in-file `StatusBadgeViewModel`.
  - [x] 3.2 `_InlineAlert.cshtml` — same pattern, in-file view model.
  - [x] 3.3 `_AuditRow.cshtml` — same pattern; embedded StatusBadge invocation for the action slot.
  - [x] 3.4 `_DashboardTile.cshtml` — same pattern.
  - [x] 3.5 Confirm or create `FieldMark.Tests.Web` project (or land tests in `FieldMark.Tests.Integration` per AC6 §host-decision); add `StatusBadgeSnapshotTests.cs`, `InlineAlertSnapshotTests.cs`, `AuditRowSnapshotTests.cs`, `DashboardTileSnapshotTests.cs`.
  - [x] 3.6 Each test class: one `[Theory]` with one `[InlineData]` per variant; test method calls the existing partial-render scaffold from Story 1.12; normalizes; asserts byte-equal.
  - [x] 3.7 If the partial-render scaffold from Story 1.12 does not yet exist in test infrastructure, confirm the harness shape before reinventing it. The Story 1.12 `_ActionButton.cshtml` snapshot test is binding precedent.
  - [x] 3.8 XSS round-trip tests for InlineAlert (AC3) and AuditRow (AC4).
  - [x] 3.9 Unknown-token tests for StatusBadge, InlineAlert, AuditRow (AC8 §category-1) and empty-actor test for AuditRow (AC8 §category-9).
  - [x] 3.10 Grep guard test (or CI lane step) asserting `Html.Raw` does not appear in any of the four new `.cshtml` files (AC9).
  - [x] 3.11 Run `dotnet csharpier check . && dotnet build && dotnet test && dotnet test FieldMark.Tests.Integration/` — all green.

- [x] **Task 4: Django wrappers + snapshot tests** (AC: #2, #3, #4, #5, #6, #9, #12)
  - [x] 4.1 `fieldmark_py/templates/components/_status_badge.html`.
  - [x] 4.2 `_inline_alert.html`.
  - [x] 4.3 `_audit_row.html` — embeds `{% include "components/_status_badge.html" with entity="Audit" value=action %}` for the action slot.
  - [x] 4.4 `_dashboard_tile.html`.
  - [x] 4.5 Locate or create `fieldmark_py/components/tests/` (or the site-wide tests location — verify before writing); add `test_status_badge_snapshot.py`, `test_inline_alert_snapshot.py`, `test_audit_row_snapshot.py`, `test_dashboard_tile_snapshot.py`. Use `django.template.loader.render_to_string` + `RequestFactory`.
  - [x] 4.6 Parametrize per variant (`@pytest.mark.parametrize("variant", ["project-active", "project-on-hold", ...])`); the test loads the variant block from `canonical.html`, normalizes, asserts equal.
  - [x] 4.7 XSS round-trip tests (AC3, AC4); unknown-token + empty-actor tests (AC8); grep guard asserting `|safe` does not appear in any of the four new templates (AC9).
  - [x] 4.8 Run `uv run ruff check . && uv run mypy . && uv run pytest && uv run pytest -m integration` — all green.

- [x] **Task 5: Go wrappers + snapshot tests** (AC: #2, #3, #4, #5, #6, #9, #12)
  - [x] 5.1 `fieldmark-go/internal/web/templates/components/status_badge.html` (`{{define "status_badge"}}…{{end}}`).
  - [x] 5.2 `inline_alert.html` (`{{define "inline_alert"}}`).
  - [x] 5.3 `audit_row.html` (`{{define "audit_row"}}`) — embeds `{{template "status_badge" $args}}` for the action slot.
  - [x] 5.4 `dashboard_tile.html` (`{{define "dashboard_tile"}}`).
  - [x] 5.5 `status_badge_test.go`, `inline_alert_test.go`, `audit_row_test.go`, `dashboard_tile_test.go` — mirror the Story 1.12 `action_button_test.go` harness (parse template via existing render helper, write to a `bytes.Buffer`, normalize, compare against the variant block).
  - [x] 5.6 Table-driven sub-tests (`t.Run(variantName, …)`) so a failure names the variant.
  - [x] 5.7 XSS round-trip tests; unknown-token + empty-actor tests; grep guard asserting `template.HTML(` does not appear in the four wrapper files (AC9).
  - [x] 5.8 Run `make check && go test ./... && go test -tags=integration ./...` — all green.

- [x] **Task 6: Cross-stack verification + parity** (AC: #6, #10, #11, #12)
  - [x] 6.1 Run `make parity` — route diff equals the Story 2.3 baseline; no new routes. `pg_indexes` zero diff.
  - [x] 6.2 Run `make test-all` — green.
  - [x] 6.3 Confirm `grep -rn "TT_CONCRETE\|ProjectPlacedOnHold\|Critical\|High\|Medium\|Low" fieldmark_shared/components/<new dirs>/canonical.html` is the **only** place those literals appear in shared code paths (they are expected in canonical.html — that's the fixture; they MUST NOT appear in a new shared symbol manifest).
  - [x] 6.4 Verify each new wrapper file's top-of-file comment references `docs/reference/component-canonical-examples.md`.
  - [x] 6.5 Verify `docs/reference/component-canonical-examples.md` Component Index lists every wrapper + every test path correctly (a one-time hand-check; future stories add rows).

- [x] **Task 7: Story sign-off** (AC: all)
  - [x] 7.1 Populate the Sign-off block below; flip sprint-status to `review`.


### Critical context (read before writing code)

- **Markup-only — zero JS, zero handlers, zero routes.** These wrappers exist so consumer pages (Story 2.8 form, Story 2.10 dashboard, Story 2.11 detail screen, Story 2.13 audit log tab) can compose the screen without inventing markup. If you find yourself adding `hx-post`, a JS init script, a Django view function, a `[HttpGet]` Razor handler, or a Fiber route, you are out of scope. Stop and re-read AC11.
- **The cross-stack contract is byte-equality of rendered HTML**, not template-syntax parity. Each stack's wrapper template syntax may differ (Razor vs Django template tags vs Go `{{define}}`); the snapshot tests assert the *output* matches `canonical.html` after the standard normalization. Razor's auto-encoding, Django's auto-escape, and Go's `html/template` context-aware escape all produce different intermediate forms but converge on the same final HTML — that's the point of the contract.
- **The canonical-examples directory is the contract, not a stack.** `fieldmark_shared/components/<name>/canonical.html` is a *fixture* — no build step, no Tailwind input, no live HTML page. It exists to be `read`-ed by the per-stack snapshot tests. Do not symlink it into any stack's static directory. Do not include it from any stack's template.
- **Decision — unknown-token handling.** The [component-edge-case-checklist.md §1](../../docs/reference/component-edge-case-checklist.md) canonical resolution prescribes both a fallback class **and** a single server-side warning log on first encounter per request. **This story implements only the fallback class** (`.badge-unknown`, `alert-unknown`, AuditRow embeds the StatusBadge `unknown` variant). The runtime warning log is deferred to a Story 2.4-follow-up entry in `deferred-work.md` (write the entry as part of this story's sign-off). Rationale: the warning logger requires a per-stack request-scoped logger lookup (DI service in .NET, `request.META` adapter in Django, `c.Locals(...)` in Fiber), and the wiring varies enough per stack that it would dominate this story's review surface. The fallback class is the user-visible signal; the log is the operator-visible signal — both have value, but they can ship independently and the user-facing signal ships first. If a reviewer disagrees with this split, the recourse is to scope-creep a single shared abstraction here (which the Cross-Stack Architecture Principle disallows) or land the per-stack loggers in a follow-up; not to silently drop the fallback class.
- **`canonical.html` rendering is a fixture, not a runtime artifact.** It does **not** need to round-trip through any template engine. Author it by hand to match the expected output exactly. The fixtures already in the directory (`action_button.example.html`, `login-form.example.html`) are the precedent — neither was generated; both were hand-authored. A misalignment between a stack wrapper's output and `canonical.html` is fixed by editing the wrapper, **not** by regenerating the fixture from one stack's output (which would silently lock in that stack's idiosyncrasies as the contract).
- **Attribute-order normalization matters.** Razor, Django, and Go html/template emit attributes in declaration order — which can differ if the wrapper declares the attributes in different orders. The Story 1.11 snapshot pipeline sorts attributes alphabetically before comparing. Verify the existing normalizer per stack handles the four new components' attribute sets — if any wrapper emits an attribute (e.g., `data-audit-action="..."` on AuditRow) that the normalizer doesn't yet support, extend the normalizer; do not work around it by re-ordering attributes in `canonical.html`.
- **Relative-time formatting is a parity hazard.** "3 minutes ago" in .NET, Django, and Go can drift by one minute depending on rounding and the surrounding code path. The AuditRow snapshot test MUST mock the current time to a deterministic fixed instant (`2026-05-28T14:23:01Z`) and assert each stack renders the same `<time>` content. If the three stacks' built-in humanizers disagree on rounding (e.g., 30s vs 60s as "a minute ago"), document the divergence in the AuditRow `README.md` and pick the "round down to the floor minute" rule across all three stacks — implement a per-stack helper if needed. **Do not commit a test that asserts each stack against itself** — the test must assert each stack against the canonical fixture.
- **JSON before/after disclosure is JSON-as-text.** `before_state` and `after_state` are stored as JSONB columns (Story 2.2) but render as escape-aware *text* inside `<pre><code>` — they are not parsed, not pretty-printed beyond a single compact serialization. Use `JsonSerializer.Serialize(state, new JsonSerializerOptions { WriteIndented = false })` (.NET), `json.dumps(state, separators=(',', ':'))` (Django), `json.Marshal(state)` then string-cast (Go). The exact serialization output may differ in key ordering between stacks — JSON objects don't guarantee key order, and each stack's serializer has its own default. **This is acceptable** as long as the snapshot fixture uses canonical key ordering that matches *all three stacks' serialization* — author the `with-disclosure-expanded` variant's JSON content by hand to use alphabetical key order, then verify each stack's serializer emits the same order for the same input. If a stack's serializer does not preserve alphabetical order, pass it through a per-stack helper that sorts keys before serializing.

### Component-specific notes

- **StatusBadge — vocabulary derivation.** The 22 vocabulary variants enumerated in AC2 come directly from [ux-design-specification.md:454–476](../planning-artifacts/ux-design-specification.md). Do not invent variants. If a state in the UX spec is missing from `canonical.html`, fix `canonical.html` to add it (and add the wrapper render + test case). Severity badges (`severity-critical`, `severity-high`) take the `badge-bump` class; everything else does not.
- **InlineAlert — Lucide icon naming.** Lucide has renamed icons over its lifecycle (`triangle-alert` ↔ `alert-triangle`). Pick the name **already in use** in this codebase (search `fieldmark_shared/src/` and existing component templates) and stay consistent. If no Lucide icon SVG is yet inlined anywhere, pick the names from the current Lucide v0.x release the project pins (check `package.json` if Lucide is an npm dep; otherwise the icons are hand-pasted from lucide.dev — record which version in the InlineAlert `README.md`).
- **AuditRow — live region is parent-owned.** The `<li class="audit-row">` does **not** carry `aria-live` or `role="log"`. The parent `<ul id="audit-log">` (rendered by the consumer page in Story 2.13) does. This is by design — the OOB swap pattern requires one stable live region for the whole list, not a per-row one. The AuditRow `README.md` records this invariant; Story 2.13's AC will pick up the parent-region conformance check.
- **AuditRow — `<time datetime="">` ISO format.** Always `YYYY-MM-DDTHH:MM:SSZ` in UTC (no offset, no fractional seconds). The visible content is the relative form; the `title` attribute is the absolute form (`YYYY-MM-DD HH:MM:SS UTC`). Use the consumer's `occurred_at` field (Story 2.2 ships this on `domain.audit_entry`) directly — do not re-parse.
- **DashboardTile — `text-3xl font-bold tnum` is fixed.** UX-DR pins the numeric value to `text-3xl font-bold` plus the `tnum` utility (the `.tnum` class is declared in `_tokens.css` per `fieldmark_shared/CLAUDE.md`). Do not parameterize the size class; the prop is the *value*, not the visual treatment.

### Edge cases (per [component-edge-case-checklist.md](../../docs/reference/component-edge-case-checklist.md))

Walked the nine categories — see AC8 for the per-category ACs. Summary: categories **1**, **6**, **8**, **9** apply and have ACs; categories **2**, **3**, **4**, **5**, **7** are N/A with rationale recorded in AC8.

### Security defaults (per [security-defaults.md](../../docs/reference/security-defaults.md))

Walked the seven categories — see AC9 for the per-category ACs. Summary: category **3** applies (adapted as output-escaping on read since this story has no writes); categories **1**, **2**, **4**, **5**, **6**, **7** are N/A with rationale recorded in AC9.

### Cross-stack contract three-deliverable check

This story introduces one cross-stack contract — the `<component>/canonical.html` byte-equality contract — and produces all three deliverables: (1) the `docs/reference/component-canonical-examples.md` doc plus per-component `README.md` files, (2) per-stack wrapper templates in idiomatic locations, (3) per-stack snapshot tests asserting equality against the fixture. See AC10.

### Files this story modifies vs creates

| File | New / Modified | Purpose |
|---|---|---|
| `fieldmark_shared/components/status_badge/canonical.html` | NEW | 23 variant blocks |
| `fieldmark_shared/components/status_badge/README.md` | NEW | contract |
| `fieldmark_shared/components/inline_alert/canonical.html` | NEW | 5 variants |
| `fieldmark_shared/components/inline_alert/README.md` | NEW | contract |
| `fieldmark_shared/components/audit_row/canonical.html` | NEW | 5 variants |
| `fieldmark_shared/components/audit_row/README.md` | NEW | contract |
| `fieldmark_shared/components/dashboard_tile/canonical.html` | NEW | 5 variants |
| `fieldmark_shared/components/dashboard_tile/README.md` | NEW | contract |
| `fieldmark_shared/components/README.md` | MODIFY | "Per-component directories" section |
| `fieldmark_shared/src/_tokens.css` | MODIFY | add `.badge-unknown`, `.alert-unknown` |
| `fieldmark_shared/src/_a11y.css` | MODIFY (only if needed) | extend `@media (forced-colors: active)` rule to cover `.badge` / `.alert` if not already |
| `fieldmark_shared/dist/fieldmark.css` | MODIFY (regenerated) | commit after `pnpm run build` |
| `docs/reference/component-canonical-examples.md` | NEW | cross-stack contract index |
| `FieldMark/FieldMark.Web/Pages/Shared/Components/_StatusBadge.cshtml` | NEW | wrapper |
| `FieldMark/FieldMark.Web/Pages/Shared/Components/_InlineAlert.cshtml` | NEW | wrapper |
| `FieldMark/FieldMark.Web/Pages/Shared/Components/_AuditRow.cshtml` | NEW | wrapper |
| `FieldMark/FieldMark.Web/Pages/Shared/Components/_DashboardTile.cshtml` | NEW | wrapper |
| `FieldMark/FieldMark.Tests.Web/Components/*SnapshotTests.cs` × 4 | NEW (or under `FieldMark.Tests.Integration/Components/` if no `Tests.Web` project) | snapshot tests |
| `fieldmark_py/templates/components/_status_badge.html` | NEW | wrapper |
| `fieldmark_py/templates/components/_inline_alert.html` | NEW | wrapper |
| `fieldmark_py/templates/components/_audit_row.html` | NEW | wrapper |
| `fieldmark_py/templates/components/_dashboard_tile.html` | NEW | wrapper |
| `fieldmark_py/components/tests/test_*_snapshot.py` × 4 | NEW | snapshot tests (location confirmed per AC6) |
| `fieldmark-go/internal/web/templates/components/status_badge.html` | NEW | wrapper |
| `fieldmark-go/internal/web/templates/components/inline_alert.html` | NEW | wrapper |
| `fieldmark-go/internal/web/templates/components/audit_row.html` | NEW | wrapper |
| `fieldmark-go/internal/web/templates/components/dashboard_tile.html` | NEW | wrapper |
| `fieldmark-go/internal/web/templates/components/*_test.go` × 4 | NEW | snapshot tests |
| `_bmad-output/implementation-artifacts/deferred-work.md` | MODIFY | new entry: "Story 2.4-followup — unknown-token runtime warning logger per [component-edge-case-checklist.md §1](../../docs/reference/component-edge-case-checklist.md) canonical resolution; deferred from Story 2.4 per Dev Notes §'Decision — unknown-token handling'" |

Anything outside this list — ComplianceTile, EntityRail, TabStrip, AGGridPanel, any consumer page (Project Detail, Compliance Dashboard, Audit Log tab), AG Grid endpoint scaffolding, the unknown-token runtime warning logger, route registration, any DB change — is out of scope. Resist the urge.

### Files to read fully before editing

- [_bmad-output/planning-artifacts/ux-design-specification.md:820–956](../planning-artifacts/ux-design-specification.md) — Custom Components section (UX-DR for StatusBadge, AuditRow, DashboardTile, InlineAlert) and Implementation Roadmap (Phase-1 / Phase-2 gate conditions).
- [_bmad-output/planning-artifacts/ux-design-specification.md:440–489](../planning-artifacts/ux-design-specification.md) — semantic color tokens, status-badge vocabulary table, compliance-score thresholds. Binding for AC2.
- [docs/reference/audit-actions.md](../../docs/reference/audit-actions.md) — canonical audit-action vocabulary; AuditRow's action prop must accept any of the 15 canonical strings plus the `unknown` fallback.
- [docs/reference/component-edge-case-checklist.md](../../docs/reference/component-edge-case-checklist.md) — nine-category walkthrough; binding for AC8.
- [docs/reference/security-defaults.md](../../docs/reference/security-defaults.md) — seven-category walkthrough; binding for AC9.
- [fieldmark_shared/CLAUDE.md](../../fieldmark_shared/CLAUDE.md) §"Snapshot-test pipeline" — the four-step normalization pipeline this story's tests inherit.
- [fieldmark_shared/components/action_button.example.html](../../fieldmark_shared/components/action_button.example.html) — variant-delimiter format precedent.
- [fieldmark_shared/components/README.md](../../fieldmark_shared/components/README.md) — current convention; this story modifies it to document the per-component-directory variant.
- [FieldMark/FieldMark.Web/Pages/Shared/_ActionButton.cshtml](../../FieldMark/FieldMark.Web/Pages/Shared/_ActionButton.cshtml) — .NET wrapper precedent (in-file view model).
- [fieldmark_py/templates/components/_action_button.html](../../fieldmark_py/templates/components/_action_button.html) — Django `{% include … with … %}` precedent.
- [fieldmark-go/internal/web/templates/components/action_button.html](../../fieldmark-go/internal/web/templates/components/action_button.html) and its [_test.go sibling](../../fieldmark-go/internal/web/templates/components/action_button_test.go) — Go `{{define}}` + table-driven test precedent.
- [_bmad-output/implementation-artifacts/2-2-map-domain-audit-entry-and-provide-a-per-stack-append-audit-entry-helper.md](2-2-map-domain-audit-entry-and-provide-a-per-stack-append-audit-entry-helper.md) — fixture-from-doc conformance-test path-walking precedent (`AuditActionConformanceTests` walks up to repo root to find `docs/reference/audit-actions.json`). Same pattern applies to `canonical.html`.
- Stack rules: [FieldMark/CLAUDE.md](../../FieldMark/CLAUDE.md), [fieldmark_py/CLAUDE.md](../../fieldmark_py/CLAUDE.md), [fieldmark-go/CLAUDE.md](../../fieldmark-go/CLAUDE.md).
- Root cross-stack invariants: [CLAUDE.md](../../CLAUDE.md) §"Cross-Stack Architecture Principle" — binding for AC10.

### Project Structure Notes

- The `fieldmark_shared/components/` directory currently holds three flat `*.example.html` files. This story introduces the per-component-directory convention (`<name>/canonical.html` + `<name>/README.md`) **alongside** the flat form. The flat form is not deprecated; future stories may use either form. The directory README is updated to document both forms.
- The Razor `Pages/Shared/Components/` sub-directory does **not** currently exist — the existing `_ActionButton.cshtml`, `_AvatarMenu.cshtml`, `_FlashRegion.cshtml`, `_ThemeToggle.cshtml` live flat under `Pages/Shared/`. This story creates the `Components/` sub-directory to group the new wrappers. Existing flat partials are not moved (out of scope; the directory hierarchy is additive).
- The Django `templates/components/` directory exists with `_action_button.html` (Story 1.12). New wrappers go alongside.
- The Go `internal/web/templates/components/` directory exists with `action_button.html` + `action_button_test.go` (Story 1.12). New wrappers + tests go alongside, flat at this level.

### References

- AC source: [_bmad-output/planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md](../planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md) §Story 2.4
- UX-DR component specs: [ux-design-specification.md:820–956](../planning-artifacts/ux-design-specification.md)
- UX-DR vocab tokens: [ux-design-specification.md:440–489](../planning-artifacts/ux-design-specification.md)
- Audit-actions canonical list: [docs/reference/audit-actions.md](../../docs/reference/audit-actions.md)
- Component edge-case checklist: [docs/reference/component-edge-case-checklist.md](../../docs/reference/component-edge-case-checklist.md)
- Security defaults checklist: [docs/reference/security-defaults.md](../../docs/reference/security-defaults.md)
- Snapshot-test pipeline: [fieldmark_shared/CLAUDE.md](../../fieldmark_shared/CLAUDE.md) §"Snapshot-test pipeline"
- Variant-delimiter precedent: [fieldmark_shared/components/action_button.example.html](../../fieldmark_shared/components/action_button.example.html)
- Wrapper precedents (Story 1.12): [.NET _ActionButton.cshtml](../../FieldMark/FieldMark.Web/Pages/Shared/_ActionButton.cshtml), [Django _action_button.html](../../fieldmark_py/templates/components/_action_button.html), [Go action_button.html](../../fieldmark-go/internal/web/templates/components/action_button.html)
- Conformance-test path-walking precedent: [_bmad-output/implementation-artifacts/2-2-...md](2-2-map-domain-audit-entry-and-provide-a-per-stack-append-audit-entry-helper.md) §Per-Stack Native Implementations
- Cross-Stack Architecture Principle: root [CLAUDE.md](../../CLAUDE.md) §Cross-Stack Architecture Principle
- Stack rules: [FieldMark/CLAUDE.md](../../FieldMark/CLAUDE.md), [fieldmark_py/CLAUDE.md](../../fieldmark_py/CLAUDE.md), [fieldmark-go/CLAUDE.md](../../fieldmark-go/CLAUDE.md)

## Dev Agent Record

### Agent Model Used

GPT-5 Codex

### Debug Log References

- `make help`
- `./tools/verify-domain-schema.sh`
- `make css`
- `dotnet csharpier check .`
- `dotnet build`
- `dotnet test FieldMark.Tests.Web/FieldMark.Tests.Web.csproj --filter FullyQualifiedName~DashboardTileSnapshotTests`
- `dotnet test FieldMark.Tests.Web/FieldMark.Tests.Web.csproj --filter "FullyQualifiedName~AuditRowSnapshotTests|FullyQualifiedName~InlineAlertSnapshotTests|FullyQualifiedName~StatusBadgeSnapshotTests"`
- `dotnet test FieldMark.Tests.Web/FieldMark.Tests.Web.csproj --filter "FullyQualifiedName~DashboardTileSnapshotTests|FullyQualifiedName~AuditRowSnapshotTests"`
- `dotnet test FieldMark.Tests.Web/FieldMark.Tests.Web.csproj --filter FullyQualifiedName~DashboardTileSnapshotTests`
- `dotnet test FieldMark.Tests.Web/FieldMark.Tests.Web.csproj --filter FullyQualifiedName~Components`
- `uv run ruff check .`
- `uv run mypy .`
- `uv run pytest fieldmark/tests/test_dashboard_tile_snapshot.py`
- `uv run pytest fieldmark/tests/test_audit_row_snapshot.py fieldmark/tests/test_inline_alert_snapshot.py fieldmark/tests/test_dashboard_tile_snapshot.py`
- `uv run pytest fieldmark/tests/test_dashboard_tile_snapshot.py`
- `uv run pytest fieldmark/tests/test_status_badge_snapshot.py fieldmark/tests/test_inline_alert_snapshot.py fieldmark/tests/test_audit_row_snapshot.py fieldmark/tests/test_dashboard_tile_snapshot.py`
- `GOCACHE=/private/tmp/fieldmark-go-cache go test ./internal/web/templates/components`
- `GOCACHE=/private/tmp/fieldmark-go-cache go test ./internal/web/templates/components ./internal/web/viewmodels`
- `GOCACHE=/private/tmp/fieldmark-go-cache STATICCHECK_CACHE=/private/tmp/fieldmark-staticcheck-cache make check`
- `make parity`
- `GOCACHE=/private/tmp/fieldmark-go-cache STATICCHECK_CACHE=/private/tmp/fieldmark-staticcheck-cache make test-all`

### Completion Notes List

- Implemented the four canonical component directories under `fieldmark_shared/components/` with variant-delimited fixtures and per-component README contracts.
- Added native .NET Razor, Django template, and Go `html/template` wrappers for StatusBadge, InlineAlert, AuditRow, and DashboardTile. The wrappers are markup-only and introduce no runtime routes, handlers, JavaScript, or database changes.
- Added per-stack snapshot tests for all 38 component variants, plus XSS escaping, unknown-token fallback, empty actor fallback, and unsafe-rendering grep guards.
- Regenerated `fieldmark_shared/dist/fieldmark.css` after adding unknown fallback styles, text semantic utilities, and forced-colors badge/alert coverage.
- Added the Story 2.4 deferred-work entry for unknown-token request-scoped runtime warning logging.
- Verification passed: schema check, CSS build, .NET format/build/component tests, Django ruff/mypy/component tests, Go `make check`, `make parity`, and `make test-all`.
- Resolved all 10 round-1 review patch items: removed AuditRow `summary[aria-expanded]`, hardened XSS negative assertions, moved Go AuditRow empty-actor fallback into the template VM contract, added missing Go coverage guards, improved repo-root diagnostics, and corrected the component contract document section order.
- Resolved all 3 round-2 review patch items: preserved zero values in Django DashboardTile rendering, added the missing .NET DashboardTile `Html.Raw` guard, and renamed the Go StatusBadge low-class fixture test to match its actual assertion.
- Resolved all 3 round-3 review patch items: added the DashboardTile zero-value canonical snapshot across all stacks, consolidated forced-colors overrides into `_a11y.css` and rebuilt shared CSS, and applied .NET AuditRow null-safe timestamp rendering.
- Resolved all 2 round-4 review patch items: documented the DashboardTile `zero-value` variant and routed .NET DashboardTile string props through the null-safe helper.
- Resolved all 6 round-5 review patch items: added whitespace-only AuditRow actor coverage in all stacks, tightened .NET InlineAlert null-safe reads, added missing unsafe-rendering grep guards, exercised InlineAlert meta escaping, and added direct InlineAlert/AuditRow unknown fallback assertions.
- Resolved all 4 round-6 review patch items: corrected DashboardTile unknown-vocabulary documentation, updated Go table-test guidance for Go 1.22+, clarified XSS-test applicability for server-computed tile labels, and made whitespace-only DashboardTile values fall back consistently across all stacks.

### File List

- `FieldMark/FieldMark.Tests.Web/Components/AuditRowSnapshotTests.cs`
- `FieldMark/FieldMark.Tests.Web/Components/ComponentRenderFixture.cs`
- `FieldMark/FieldMark.Tests.Web/Components/DashboardTileSnapshotTests.cs`
- `FieldMark/FieldMark.Tests.Web/Components/InlineAlertSnapshotTests.cs`
- `FieldMark/FieldMark.Tests.Web/Components/StatusBadgeSnapshotTests.cs`
- `FieldMark/FieldMark.Tests.Web/Helpers/NormaliseHtml.cs`
- `FieldMark/FieldMark.Web/Pages/Shared/Components/_AuditRow.cshtml`
- `FieldMark/FieldMark.Web/Pages/Shared/Components/_DashboardTile.cshtml`
- `FieldMark/FieldMark.Web/Pages/Shared/Components/_InlineAlert.cshtml`
- `FieldMark/FieldMark.Web/Pages/Shared/Components/_StatusBadge.cshtml`
- `_bmad-output/implementation-artifacts/2-4-implement-phase-2-markup-only-components-statusbadge-inlinealert-auditrow-dashboardtile.md`
- `_bmad-output/implementation-artifacts/deferred-work.md`
- `_bmad-output/implementation-artifacts/sprint-status.yaml`
- `docs/reference/component-canonical-examples.md`
- `docs/reference/security-defaults.md`
- `fieldmark-go/CLAUDE.md`
- `fieldmark-go/internal/web/templates/components/audit_row.html`
- `fieldmark-go/internal/web/templates/components/audit_row_test.go`
- `fieldmark-go/internal/web/templates/components/component_snapshot_test.go`
- `fieldmark-go/internal/web/templates/components/dashboard_tile.html`
- `fieldmark-go/internal/web/templates/components/dashboard_tile_test.go`
- `fieldmark-go/internal/web/templates/components/inline_alert.html`
- `fieldmark-go/internal/web/templates/components/inline_alert_test.go`
- `fieldmark-go/internal/web/templates/components/status_badge.html`
- `fieldmark-go/internal/web/templates/components/status_badge_test.go`
- `fieldmark-go/internal/web/testutil/normalizehtml.go`
- `fieldmark-go/internal/web/viewmodels/components.go`
- `fieldmark_py/fieldmark/tests/component_fixtures.py`
- `fieldmark_py/fieldmark/tests/normalize_html.py`
- `fieldmark_py/fieldmark/tests/test_audit_row_snapshot.py`
- `fieldmark_py/fieldmark/tests/test_dashboard_tile_snapshot.py`
- `fieldmark_py/fieldmark/tests/test_inline_alert_snapshot.py`
- `fieldmark_py/fieldmark/tests/test_status_badge_snapshot.py`
- `fieldmark_py/templates/components/_audit_row.html`
- `fieldmark_py/templates/components/_dashboard_tile.html`
- `fieldmark_py/templates/components/_inline_alert.html`
- `fieldmark_py/templates/components/_status_badge.html`
- `fieldmark_shared/components/README.md`
- `fieldmark_shared/components/audit_row/README.md`
- `fieldmark_shared/components/audit_row/canonical.html`
- `fieldmark_shared/components/dashboard_tile/README.md`
- `fieldmark_shared/components/dashboard_tile/canonical.html`
- `fieldmark_shared/components/inline_alert/README.md`
- `fieldmark_shared/components/inline_alert/canonical.html`
- `fieldmark_shared/components/status_badge/README.md`
- `fieldmark_shared/components/status_badge/canonical.html`
- `fieldmark_shared/dist/fieldmark.css`
- `fieldmark_shared/src/_a11y.css`
- `fieldmark_shared/src/_components.css`
- `fieldmark_shared/src/_tokens.css`

### Change Log

- 2026-05-28 — Implemented Story 2.4 markup-only components, canonical fixtures, per-stack wrappers/tests, shared CSS fallback styles, component contract documentation, and deferred-work follow-up entry.
- 2026-05-28 — Addressed round-1 code review findings; 10 patch items resolved and full validation gates passed.
- 2026-05-28 — Addressed round-2 code review findings; 3 patch items resolved and full validation gates passed.
- 2026-05-28 — Addressed round-3 code review findings; 3 patch items resolved, 1 item deferred, and full validation gates passed.
- 2026-05-28 — Addressed round-4 code review findings; 2 patch items resolved and full validation gates passed.
- 2026-05-28 — Addressed round-5 code review findings; 6 patch items resolved, 2 items deferred, and full validation gates passed.
- 2026-05-28 — Addressed round-6 code review findings; 4 patch items resolved and full validation gates passed.

## Sign-off

| Field | Value |
|---|---|
| Final review date | 2026-05-28 |
| Total review rounds | 7 |
| Final reviewer verdict | All ACs satisfied; 33 patch items resolved across 7 rounds; 4 Round 7 patches applied directly by reviewer; story `done` |
| Deferred-work entries | (1) Story 2.4-followup — unknown-token runtime warning logger; (2) `StatusBadgeVM.Severity` dead field; (3) Go `AuditRowVM.ActionClass` constructor enforcement; (4) Go component snapshot harness nested-template parsing — deferred to first handler story / future harness story |
| Dev-notes divergences from epic AC | The epic AC says "Phase-2 markup-only components" but per [ux-design-specification.md:943–948](../planning-artifacts/ux-design-specification.md) StatusBadge / AuditRow / InlineAlert are **Phase 1** and DashboardTile is **Phase 2**. The bundling into one story is deliberate — these four are the markup-only-zero-JS subset; ComplianceTile / EntityRail / TabStrip / AGGridPanel land in dedicated downstream stories (2.5 / 2.6 / 2.7 / 2.9) because each ships behavior (OOB target / responsive collapse / arrow-key JS / AG Grid bundle). Recording the rationale here rather than amending the epic. |

### Review Findings

- [x] [Review][Patch] `aria-expanded` on `<summary>` is semantically dubious — `<details>`/`<summary>` manages open state via the `open` attribute natively; `aria-expanded` is redundant and can conflict with AT signals if the two diverge — implementing agent should resolve: either remove from canonical + all wrappers + update snapshots (correct per spec) or document as an explicit cross-browser AT compatibility deviation in each component README [`fieldmark_shared/components/audit_row/canonical.html`, `FieldMark/FieldMark.Web/Pages/Shared/Components/_AuditRow.cshtml`, `fieldmark_py/templates/components/_audit_row.html`, `fieldmark-go/internal/web/templates/components/audit_row.html`]

- [x] [Review][Patch] Go InlineAlert test loop mutates `vm` after range-assign with no `t.Parallel()` guard — latent data-corruption trap if subtests are ever parallelized [`fieldmark-go/internal/web/templates/components/inline_alert_test.go`]
- [x] [Review][Patch] .NET `_AuditRow.cshtml` derives `emptyActor` from `actor == "unnamed"` (post-substituted value) instead of raw `string.IsNullOrWhiteSpace(Model.ActorName)` — a real actor named "unnamed" would incorrectly receive the initials fallback [`FieldMark/FieldMark.Web/Pages/Shared/Components/_AuditRow.cshtml`]
- [x] [Review][Patch] Go `StatusBadgeSnapshotTests`: `violation-low` severity variant not exercised — `violation-open-medium-low` tests only `Severity = "Medium"`; the `"Low"` arm in the switch is dead from a regression-protection standpoint [`fieldmark-go/internal/web/templates/components/status_badge_test.go`]
- [x] [Review][Patch] `ComponentRenderFixture.cs` throws `"Repo root not found"` with no starting-path diagnostic — add the walked path to the error message so CI failures are triageable [`FieldMark/FieldMark.Tests.Web/Components/ComponentRenderFixture.cs`]
- [x] [Review][Patch] `docs/reference/component-canonical-examples.md` missing "Why pointer" section (spec requires 5 `##` sections in order; file has 3); Status block is an informal blockquote instead of a named section — add the missing §2 and promote the status line to a proper `## Status` section [`docs/reference/component-canonical-examples.md`]
- [x] [Review][Patch] Go `dashboard_tile_test.go` missing `TestDashboardTileTemplateDoesNotUseTemplateHTML` grep guard — the other three Go component test files all have this guard; DashboardTile is unprotected [`fieldmark-go/internal/web/templates/components/dashboard_tile_test.go`]
- [x] [Review][Patch] Go AuditRow: empty `ActorName` → `"unnamed"` transform is caller-owned with no unit test that passes an empty string and asserts "unnamed" output; the Go template renders `{{ .ActorName }}` verbatim and would emit an empty span if a handler forgets to pre-compute the value (AC8 §category-9) [`fieldmark-go/internal/web/templates/components/audit_row.html`, `fieldmark-go/internal/web/viewmodels/components.go`]
- [x] [Review][Patch] All three stacks' InlineAlert XSS tests assert `Contains(escaped)` but omit `NotContains(raw "<script>")` — a rendering regression that emitted both escaped and raw forms would still pass the current assertion [`FieldMark/FieldMark.Tests.Web/Components/InlineAlertSnapshotTests.cs`, `fieldmark_py/fieldmark/tests/test_inline_alert_snapshot.py`, `fieldmark-go/internal/web/templates/components/inline_alert_test.go`]
- [x] [Review][Patch] AuditRow XSS test payload is `{"value":"<script>"}` (JSON-wrapped) not the spec-prescribed bare `<script>alert(1)</script>` string (AC4 §XSS test) [`FieldMark/FieldMark.Tests.Web/Components/AuditRowSnapshotTests.cs`, `fieldmark_py/fieldmark/tests/test_audit_row_snapshot.py`, `fieldmark-go/internal/web/templates/components/audit_row_test.go`]

#### Round 2 findings (2026-05-28)

- [x] [Review][Patch] Django `_dashboard_tile.html` `{% if value %}` treats `"0"` as falsy — a zero-count tile (e.g., 0 open violations) renders `—` instead of `0`, while .NET uses `string.IsNullOrWhiteSpace` (correctly keeps `"0"`); fix with `{% if value is not None and value != "" %}` or `{% if value|stringformat:"s" %}` [`fieldmark_py/templates/components/_dashboard_tile.html`]
- [x] [Review][Patch] .NET `DashboardTileSnapshotTests.cs` missing `Html.Raw` grep guard — all three other .NET component test files (`StatusBadgeSnapshotTests`, `InlineAlertSnapshotTests`, `AuditRowSnapshotTests`) include a `[Fact]` asserting `Html.Raw` is absent from the wrapper file; DashboardTile is unprotected [`FieldMark/FieldMark.Tests.Web/Components/DashboardTileSnapshotTests.cs`]
- [x] [Review][Patch] Go `TestStatusBadgeViolationOpenLowSeverityMatchesMediumLowVariant` is misleadingly named — `StatusBadgeVM.Severity` is never read by the template (template renders `{{ .ClassName }}` directly); the test only validates the pre-baked `ClassName` value, not any resolution logic; rename the test to reflect what it actually asserts, or add a separate `NewStatusBadgeVM` helper test that validates `(Violation, Open, Low)` → `badge-violation-open-low` [`fieldmark-go/internal/web/templates/components/status_badge_test.go`]

#### Round 3 findings (2026-05-28)

- [x] [Review][Patch] .NET `DashboardTileSnapshotTests` missing `value="0"` snapshot variant — Django gained an explicit zero-value test (`test_dashboard_tile_zero_value_renders_zero`) via the R2-P1 fix; .NET has no equivalent, leaving the zero-value path untested there; add a `{"zero-value", TileModel(value: "0")}` entry to `Variants` (and a corresponding `<!-- variant: zero-value -->` block in `dashboard_tile/canonical.html` if one does not already exist) [`FieldMark/FieldMark.Tests.Web/Components/DashboardTileSnapshotTests.cs`]
- [x] [Review][Patch] `fieldmark_shared/src/_tokens.css` contains a `@media (forced-colors: active)` block — forced-colors overrides belong exclusively in `fieldmark_shared/src/_a11y.css` per AC1 §1.8; the errant block in `_tokens.css` produces a duplicate `@media (forced-colors: active) { .badge, .alert { ... } }` rule in `dist/fieldmark.css`; remove the forced-colors block from `_tokens.css` and rebuild dist [`fieldmark_shared/src/_tokens.css`, `fieldmark_shared/dist/fieldmark.css`]
- [x] [Review][Patch] `.NET _AuditRow.cshtml` uses `@Model.OccurredAt`, `@Model.Absolute`, and `@Model.Relative` directly without the `S()` null-safety helper — a null value produces invalid `datetime=""` HTML, whereas all other model properties flow through `S()` which safely returns `""`; wrap these three in `S()` for consistency [`FieldMark/FieldMark.Web/Pages/Shared/Components/_AuditRow.cshtml`]
- [x] [Review][Defer] `StatusBadgeVM.Severity` field is dead exported state — no template or resolver reads it; structural residue from the rename-only R2-P3 resolution; acceptable for this markup-only story since resolution logic lives in future handler stories — deferred to when the first handler constructs a `StatusBadgeVM` from domain values [`fieldmark-go/internal/web/viewmodels/components.go`]

#### Round 4 findings (2026-05-28)

- [x] [Review][Patch] `fieldmark_shared/components/dashboard_tile/README.md` Variant List omits `zero-value` — the R3-P1 patch added a `zero-value` variant to `canonical.html` and all three stacks' snapshot tests but did not update the README contract doc; Variant List still enumerates only 5 variants [`fieldmark_shared/components/dashboard_tile/README.md`]
- [x] [Review][Patch] `.NET _DashboardTile.cshtml` — `@Model.TileId` and `@Model.Label` accessed directly without the `S()` null-safety helper, inconsistent with all other model reads in all four wrapper templates; null `TileId` produces invalid `id=""` HTML [`FieldMark/FieldMark.Web/Pages/Shared/Components/_DashboardTile.cshtml`]

#### Round 5 findings (2026-05-28)

- [x] [Review][Patch] Whitespace-only actor (`"   "`) not tested in any stack's AuditRow tests — `.NET` uses `IsNullOrWhiteSpace`, Go uses `TrimSpace`, Django uses `.strip`, all three correctly handle whitespace, but all snapshot tests pass only `""` (empty string); a regression on the whitespace-only path would go undetected; add a whitespace-only actor test case in all three stacks matched against the `empty-actor` canonical variant [all three `audit_row` test files]
- [x] [Review][Patch] `.NET _InlineAlert.cshtml` renders `@Model.Title` and `@Model.Message` directly without the `S()` null-safety helper — every other model read in all four wrapper templates routes through local variables extracted via `S()`; an ExpandoObject model missing either key throws `RuntimeBinderException` rather than rendering safely as `""`; apply the same `var title = S(Model.Title); var message = S(Model.Message);` pattern [`FieldMark/FieldMark.Web/Pages/Shared/Components/_InlineAlert.cshtml`]
- [x] [Review][Patch] `.NET StatusBadgeSnapshotTests.cs` missing `[Fact]` `Html.Raw` grep guard — the other three `.NET` component test files (`InlineAlertSnapshotTests`, `AuditRowSnapshotTests`, `DashboardTileSnapshotTests`) all include an `XxxTemplateDoesNotUseHtmlRaw` fact; `StatusBadgeSnapshotTests` is the only one without this guard, leaving the `_StatusBadge.cshtml` file unprotected against future `Html.Raw` regressions [`FieldMark/FieldMark.Tests.Web/Components/StatusBadgeSnapshotTests.cs`]
- [x] [Review][Patch] Django `test_dashboard_tile_snapshot.py` missing `|safe` grep guard test — the other three Django component test files all include a `test_xxx_template_does_not_use_safe_filter` function; `test_dashboard_tile_snapshot.py` is the only one without this guard [`fieldmark_py/fieldmark/tests/test_dashboard_tile_snapshot.py`]
- [x] [Review][Patch] InlineAlert XSS test passes `meta=""` so the `alert-meta` block never renders — `meta` is a user-visible field; its auto-escaping path is never exercised by the XSS test; pass an XSS payload as `meta` and add a `NotContains(raw)` assertion alongside the existing `Contains(escaped)` assertion in all three stacks [`.NET InlineAlertSnapshotTests.cs`, `fieldmark_py/fieldmark/tests/test_inline_alert_snapshot.py`, `fieldmark-go/internal/web/templates/components/inline_alert_test.go`]
- [x] [Review][Patch] No dedicated unknown-fallback-class assertion for `InlineAlert` or `AuditRow` — `StatusBadge` has a targeted `TestStatusBadgeUnknownFallbackClass` / `test_status_badge_unknown_fallback_class` / `StatusBadgeUnknownFallbackClass` fact in all three stacks; `InlineAlert` and `AuditRow` only have implicit coverage via the `"unknown"` snapshot variant; add dedicated `TestInlineAlertUnknownFallbackClass` and `TestAuditRowUnknownActionFallbackClass` (Go/Django equivalents) that directly assert the `alert-unknown` / `badge-unknown` class is present [`fieldmark-go/internal/web/templates/components/inline_alert_test.go`, `fieldmark-go/internal/web/templates/components/audit_row_test.go`, `fieldmark_py/fieldmark/tests/test_inline_alert_snapshot.py`, `fieldmark_py/fieldmark/tests/test_audit_row_snapshot.py`, `FieldMark/FieldMark.Tests.Web/Components/AuditRowSnapshotTests.cs`]
- [x] [Review][Defer] Go `AuditRowVM.ActionClass` is caller-computed with no constructor enforcement — a handler that forgets to set `ActionClass` when constructing a VM renders `<span class="badge ">Action</span>` (empty class); deferred to the first handler story (2.10/2.11) that constructs `AuditRowVM` from domain values, which should introduce a `NewAuditRowVM` constructor [`fieldmark-go/internal/web/viewmodels/components.go`]
- [x] [Review][Defer] Go `component_snapshot_test.go` `renderComponent` only parses the single component file — any nested `{{template "status_badge" ...}}` call within a component would cause `ExecuteTemplate` to fail; the current implementation inlines badge logic via VM fields so tests pass, but the harness is fragile for any future nested template use; deferred as pre-existing design choice for the markup-only scope [`fieldmark-go/internal/web/templates/components/component_snapshot_test.go`]

#### Round 6 findings (2026-05-28)

- [x] [Review][Patch] `fieldmark_shared/components/dashboard_tile/README.md` "Unknown Vocabulary Handling" section is inaccurate — DashboardTile has no `--unknown` fallback variant; unknown `value_color` silently applies no class (the `ColorClass()` helper returns `""` for unrecognized values); the boilerplate wording implies an `--unknown` variant exists and will mislead a future developer into adding a spurious variant [`fieldmark_shared/components/dashboard_tile/README.md`]
- [x] [Review][Patch] `fieldmark-go/CLAUDE.md` `tc := tc` loop-capture rule is inaccurate for Go 1.22+ — Go 1.22 changed loop variable scoping to per-iteration by default, making `tc := tc` a no-op; the rule should describe the actual concern ("avoid mutating a VM struct after the range variable is captured, especially before passing to `t.Run`") rather than prescribing a now-redundant idiom; additionally the rule is inconsistently applied — only `inline_alert_test.go` uses it, the other three component tests do not [`fieldmark-go/CLAUDE.md`]
- [x] [Review][Patch] `docs/reference/security-defaults.md` §3a is self-inconsistent — §3a explicitly lists "tile label" as a user-supplied string requiring XSS round-trip tests, but no DashboardTile XSS test exists in any stack; either revise §3a to distinguish server-computed display values from user-entered strings (with DashboardTile labels noted as server-computed and therefore exempt) or add the missing XSS tests across all three stacks [`docs/reference/security-defaults.md`]
- [x] [Review][Patch] Cross-stack whitespace-only `value` divergence in DashboardTile — Django `{% if value == 0 or value %}` treats `"   "` (whitespace-only) as truthy and renders raw spaces; `.NET` `string.IsNullOrWhiteSpace` correctly substitutes the em-dash fallback; Go pre-computes `DisplayValue` so behavior depends on the caller; add a whitespace-only value test that asserts the em-dash fallback across all three stacks, matching the pattern from the R5-P1 AuditRow fix [all three dashboard_tile test files]

#### Round 7 findings (2026-05-28) — applied directly

- [x] [Review][Patch] `fieldmark_shared/components/audit_row/README.md` "Unknown Vocabulary Handling" has the same inaccuracy R6-P1 fixed for DashboardTile — says "Unknown values render the `--unknown` fallback variant" but AuditRow has no `--unknown` row variant; an unknown `action` causes the embedded StatusBadge to render `badge-unknown`; the row structure is the same as `default` [`fieldmark_shared/components/audit_row/README.md`] — **fixed directly**
- [x] [Review][Patch] `fieldmark-go/internal/web/templates/components/inline_alert_test.go` still used `vm := vm` pattern after Go CLAUDE.md (R6-P2) explicitly identified it as "no longer the safety mechanism" in Go 1.22+; only the inline_alert test used it; removed to match the other three component tests and the updated rule [`fieldmark-go/internal/web/templates/components/inline_alert_test.go`] — **fixed directly**
- [x] [Review][Patch] `fieldmark_py/CLAUDE.md` prescribed `{% if value is not None and value != "" %}` but the actual `_dashboard_tile.html` fix uses `{% if value == 0 or value is not None and value|stringformat:"s"|slugify %}`; a developer following the doc would write a condition that passes for whitespace-only strings; updated to match the actual implementation with explanation [`fieldmark_py/CLAUDE.md`] — **fixed directly**
- [x] [Review][Patch] `memory/feedback_story_2_4_rule_changes.md` still advised "capture `tc := tc`" contradicting Go CLAUDE.md R6-P2 update; memory file updated to remove the deprecated idiom and describe the actual concern [`_bmad-output/memory/feedback_story_2_4_rule_changes.md`] — **fixed directly**
