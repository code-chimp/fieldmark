# Story 2.5: Implement ComplianceTile component and `#compliance-tile` OOB target

Status: ready-for-dev

Epic: 2 — Project Lifecycle & Compliance Dashboard
Source AC: [_bmad-output/planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md](../planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md) §Story 2.5
Depends on: Story 2.4 (introduces the per-component directory convention under `fieldmark_shared/components/<name>/`, the `canonical.html` + `README.md` fixture format, and the per-stack snapshot harness for the four markup-only wrappers). This story extends that convention with a fifth component — ComplianceTile — that adds the threshold-color rule and serves as the canonical `#compliance-tile` OOB target.

## Story

As a developer rendering the Project Detail anchor screen (Story 2.11), the Compliance Dashboard (Story 2.10), and every state-change response that updates the compliance score in the same round trip (Stories 2.12 Place-on-Hold, 5.5 Approve Corrective Action — the talk's "three regions, one round trip" hinge),
I want a markup-only **ComplianceTile** wrapper per stack — byte-identical HTML across .NET, Django, and Go — that renders a 0–100 compliance score with threshold-band semantic color, a textual threshold word ("Healthy" / "Watch" / "Concern" / "Critical"), a no-data em-dash variant, and a stable `aria-live="polite"` `<section>` with id `#compliance-tile` (project context) or `#compliance-tile-portfolio` (portfolio context),
So that downstream stories (2.10, 2.11, 2.12, 5.5) can emit OOB swaps targeting `#compliance-tile` without inventing markup, the `aria-atomic="true"` announcement on score change is consistent across stacks, and color is never the sole information carrier (UX §"Information Architecture" / WCAG 1.4.1).

**Scope boundary:** this story produces (a) `fieldmark_shared/components/compliance_tile/canonical.html` (variant-delimited fixture, same convention Story 2.4 introduces) + `README.md` (contract), (b) one markup-only wrapper template per stack in that stack's idiomatic component-template location, (c) per-stack snapshot tests asserting byte-equality against `canonical.html` after the Story 1.11 / 1.12 / 2.4 normalization, (d) a new row in the `docs/reference/component-canonical-examples.md` Component Index (the doc Story 2.4 creates). **Out of scope:** the actual OOB-swap producers (Story 2.12 Place-on-Hold orchestration, Story 5.5 Approve CA orchestration — these merely *target* `#compliance-tile`; they are not implemented here), any consumer page that *uses* the tile (Story 2.10 Compliance Dashboard, Story 2.11 Project Detail), any compliance-score computation logic (already covered by Story 3.3 rule engine — this story renders whatever score the caller passes in), any threshold-crossing color transition or animation (UX §"Phase 3 — Polish optimizations" defers this; UX-DR explicitly flags "may violate calm register and be cut"), any JavaScript (zero-JS component), and the unknown-vocabulary-fallback runtime warning logger (deferred per the same "Story 2.4-followup" entry in `deferred-work.md`).

## Acceptance Criteria

### AC1 — Canonical-example directory under `fieldmark_shared/components/`

**Given** the per-component-directory convention introduced by Story 2.4
**When** I inspect `fieldmark_shared/components/`
**Then** a new sub-directory exists with this exact layout:

```
fieldmark_shared/components/compliance_tile/
├── canonical.html
└── README.md
```

**And** `canonical.html` follows the same variant-delimited format as `action_button.example.html` (Story 1.12) and the four `canonical.html` files from Story 2.4 — a leading HTML comment block documenting `fixture:` inputs, then `<!-- variant: <name> (inputs: ...) -->` delimiters separating each variant's expected output.

**And** `canonical.html` contains **exactly nine variant blocks**:
1. `healthy-project` — `score=95, label="Compliance", context="project", id="compliance-tile"` → success color, threshold word "Healthy", id `compliance-tile`.
2. `watch-project` — `score=82, label="Compliance", context="project", id="compliance-tile"` → warning-lighter, "Watch".
3. `concern-project` — `score=58, label="Compliance", context="project", id="compliance-tile"` → warning-darker, "Concern".
4. `critical-project` — `score=37, label="Compliance", context="project", id="compliance-tile"` → danger, "Critical".
5. `healthy-portfolio` — `score=91, label="Portfolio Compliance", context="portfolio", id="compliance-tile-portfolio"` → success, "Healthy", id `compliance-tile-portfolio`.
6. `critical-portfolio` — `score=42, label="Portfolio Compliance", context="portfolio", id="compliance-tile-portfolio"` → danger, "Critical", id `compliance-tile-portfolio`.
7. `no-data-project` — `score=null, label="Compliance", context="project", id="compliance-tile"` → renders em-dash `—` in the value slot, threshold word omitted (no text), neutral color (no semantic-color class on the value).
8. `boundary-90` — `score=90, label="Compliance", context="project", id="compliance-tile"` → success, "Healthy" (verifies the `≥ 90` inclusive boundary).
9. `boundary-70` — `score=70, label="Compliance", context="project", id="compliance-tile"` → warning-lighter, "Watch" (verifies the `≥ 70` inclusive boundary; mirror boundary at 50 is covered by `concern-project` at 58 and `critical-project` at 37 — but **also** add a `boundary-50` and a `boundary-49` block if the test author judges the boundary surface needs more coverage; nine blocks is the floor, not the ceiling — see Dev Notes §"Boundary coverage rationale").

**And** `README.md` documents the contract in this fixed order (matching Story 2.4's per-component README pattern): (1) Purpose (one sentence — "Display a compliance score with threshold-derived semantic color; serves as the canonical `#compliance-tile` OOB target"), (2) Required props with types (`score: int? (0–100, null for no-data)`, `label: string`, `context: "project" | "portfolio"`, `id: string` — caller-supplied so the same wrapper can render `#compliance-tile` or `#compliance-tile-portfolio`), (3) Variant list (the nine names that appear as `<!-- variant: -->` delimiters), (4) ARIA invariants (`role="status"`, `aria-live="polite"`, `aria-atomic="true"` — all three are mandatory and on the outer `<section>`), (5) Threshold table (band → semantic color → threshold word, exactly mirroring UX §"Compliance score thresholds" at [ux-design-specification.md:478–489](../planning-artifacts/ux-design-specification.md)), (6) Allowed class vocabulary (`text-success`, `text-warning-strong`, `text-warning`, `text-danger`, `text-3xl`, `font-bold`, `tnum`, `text-neutral` for no-data — names match the tokens already declared in `_tokens.css`; verify the warning-lighter / warning-darker distinction is already represented in `_tokens.css` as Story 2.4 expected — if `text-warning-strong` does not exist, see AC2 §token-derivation), (7) Snapshot-equality requirement (one-line: "Per-stack wrappers MUST render output byte-equal to the matching variant block in `canonical.html` after the standard normalization defined in `fieldmark_shared/CLAUDE.md` §'Snapshot-test pipeline'"), (8) Unknown-vocabulary handling (one-line: "Score values outside 0–100 render as the `no-data` variant; a per-stack unit test asserts this. No runtime warning logger this story — same rationale as Story 2.4 Dev Notes §'Decision — unknown-token handling'.").

**And** `docs/reference/component-canonical-examples.md` (Story 2.4) is MODIFIED — one new row appended to the Component Index table with columns: `ComplianceTile`, `fieldmark_shared/components/compliance_tile/canonical.html`, `…/README.md`, then the three per-stack wrapper paths from AC2 and the three per-stack test paths from AC6.

### AC2 — ComplianceTile component markup contract (UX-DR §"ComplianceTile", UX §"Compliance score thresholds")

**Given** the ComplianceTile contract at [ux-design-specification.md:835–841](../planning-artifacts/ux-design-specification.md) and the threshold table at [ux-design-specification.md:478–489](../planning-artifacts/ux-design-specification.md)
**When** I inspect `fieldmark_shared/components/compliance_tile/canonical.html` for the `healthy-project` variant
**Then** the rendered HTML is **exactly** this shape (whitespace-normalized; attribute order alphabetized for the byte-equality comparison):

```html
<section
  aria-atomic="true"
  aria-live="polite"
  class="compliance-tile"
  id="compliance-tile"
  role="status"
>
  <p class="compliance-tile__label">Compliance</p>
  <p class="compliance-tile__value text-3xl font-bold tnum text-success">95</p>
  <p class="compliance-tile__threshold text-success">Healthy</p>
</section>
```

**And** the four populated variants follow the same shape, with these per-band substitutions on `.compliance-tile__value` and `.compliance-tile__threshold`:

| Band (score) | Value class | Threshold word | Threshold class |
|---|---|---|---|
| ≥ 90 (Healthy) | `text-success` | `Healthy` | `text-success` |
| 70 – 89 (Watch) | `text-warning` | `Watch` | `text-warning` |
| 50 – 69 (Concern) | `text-warning-strong` | `Concern` | `text-warning-strong` |
| < 50 (Critical) | `text-danger` | `Critical` | `text-danger` |
| `null` (No data) | `text-neutral` | _omitted — no `<p class="compliance-tile__threshold">` element rendered_ | _n/a_ |

**Token derivation.** UX §"Compliance score thresholds" specifies "warning (lighter)" for the 70–89 band and "warning (darker)" for the 50–69 band — the canonical token names are `text-warning` (lighter / amber-400 dark / amber-600 light) for 70–89 and `text-warning-strong` (darker — derived once via `_tokens.css`) for 50–69. **Verify** the `_tokens.css` rules from Story 2.4's edits already declare both tokens. If only `text-warning` exists, this story extends `_tokens.css` by adding a `--color-warning-strong` HSL value and a `text-warning-strong { color: var(--color-warning-strong); }` rule, paired with a dark-mode override. Use the existing Tailwind amber palette per the UX-DR semantic-color table: light = amber-700; dark = amber-300 (one step darker than the existing `text-warning` to provide the visible mid-band distinction the four-band threshold rule requires). Add the rule **in `_tokens.css`** alongside the other text-color utilities — this is the only `fieldmark_shared/src/` edit this story makes.

**And** the `no-data` variant collapses to:

```html
<section
  aria-atomic="true"
  aria-live="polite"
  class="compliance-tile"
  id="compliance-tile"
  role="status"
>
  <p class="compliance-tile__label">Compliance</p>
  <p class="compliance-tile__value text-3xl font-bold tnum text-neutral">—</p>
</section>
```

The `—` is the U+2014 EM DASH character (UTF-8 bytes `E2 80 94`), **not** the ASCII hyphen-minus and **not** the HTML entity `&mdash;` — the byte-equality contract is on the raw output, and the three frameworks' default auto-escaping does not encode `—` (it's not in the XSS-prone set). Per-stack tests assert the byte sequence directly.

**And** the wrapper exposes the **caller-supplied `id`** prop verbatim — when the caller passes `id="compliance-tile-portfolio"`, the outer `<section>` carries that id, not `#compliance-tile`. The wrapper does not derive the id from the `context` prop; `context` is **documentation-only** (the `README.md` and `canonical.html` use it to label the variant set, but the wrapper template itself receives only `id`, `label`, `score`). This keeps the wrapper API minimal — Story 2.10 and Story 2.11 pass `id` explicitly. A per-stack unit test asserts that passing `id="compliance-tile-portfolio"` produces the `healthy-portfolio` variant byte-equal output for `score=91, label="Portfolio Compliance"`.

**And** the `LABEL_UPPERCASE` content follows the same Story 2.4 DashboardTile rule — **CSS-uppercased via `text-transform: uppercase`** on the `.compliance-tile__label` class. The wrapper input is the human-cased label (`"Compliance"` / `"Portfolio Compliance"`); the visual uppercase comes from CSS. Do not call `.upper()` / `.ToUpper()` / `strings.ToUpper(...)` in the wrapper. Confirm `_tokens.css` (post-Story-2.4) declares this rule on `.dashboard-tile__label`; this story adds the identical `text-transform: uppercase` rule on `.compliance-tile__label` (one new selector, same declaration). If Story 2.4 elected to group `text-transform: uppercase` under a shared utility selector instead, follow that precedent — do not invent a new pattern.

**Per-stack wrapper paths** (idiomatic per stack — no shared template, no symlinked partial):

- **.NET (Razor partial):** `FieldMark/FieldMark.Web/Pages/Shared/Components/_ComplianceTile.cshtml` (NEW). Invoked as `<partial name="Shared/Components/_ComplianceTile" model="@(new ComplianceTileViewModel(score, label, id))" />`. The view model lives in the **same file** at the bottom (`@functions { public record ComplianceTileViewModel(int? Score, string Label, string Id) }`) — matching Story 1.12 / 2.4 precedent. Place the file alongside the four Story 2.4 wrappers in `Pages/Shared/Components/` (the sub-directory Story 2.4 creates).
- **Django (template include):** `fieldmark_py/templates/components/_compliance_tile.html` (NEW). Invoked as `{% include "components/_compliance_tile.html" with score=95 label="Compliance" id="compliance-tile" %}`.
- **Go (`html/template` `{{define}}` block):** `fieldmark-go/internal/web/templates/components/compliance_tile.html` (NEW). Defines `{{define "compliance_tile"}}…{{end}}` (snake_case template name matching the Story 1.12 / 2.4 precedent). Invoked as `{{template "compliance_tile" $complianceTileArgs}}`. The context struct is `type ComplianceTileArgs struct { Score *int; Label string; ID string }` (note `*int` so `nil` is the no-data signal — Go's `0` is a legitimate score, distinct from no-data).

**No-data signaling per stack.** All three stacks model "no data" as the **absence of a score**, not as `0`:
- .NET: `int? Score` — `null` is no-data.
- Django: `score=None` — `{% if score is None %}…{% endif %}`.
- Go: `*int` — `nil` is no-data.
A per-stack test asserts that `score=0` renders the `critical-project`-shape variant (band: < 50 → danger, "Critical"), **not** the `no-data` variant. This is the same test fixture, asserted on each stack.

### AC3 — Band determination is pure and table-driven (zero JS, zero conditional sprawl)

**Given** the band-to-color-and-word mapping in AC2
**When** I inspect each stack's wrapper implementation
**Then** the band derivation is a **pure server-side function** that takes `int? score` and returns the four-tuple `(value_class, threshold_word, threshold_class, render_threshold_p)`. The function is expressed as a small lookup — not a chain of `if/else if/else` past three levels deep — and is **trivially auditable** for the four band boundaries (90, 70, 50) + the no-data branch. Suggested form per stack:

- **.NET:** a `private static (string ValueClass, string ThresholdWord, string ThresholdClass, bool RenderP) ResolveBand(int? score)` static method on the in-file view model `ComplianceTileViewModel`. Use a `switch` expression — five arms: `null`, `>= 90`, `>= 70`, `>= 50`, `_`. The view model exposes the resolved tuple as four properties so the `.cshtml` template emits them directly with no `@if` logic in the markup.
- **Django:** a template tag is unnecessary — instead, the `_compliance_tile.html` template receives a context dictionary pre-populated by a **template filter or context helper** named `compliance_band` registered in `fieldmark_py/components/templatetags/compliance_tile_tags.py` (NEW). The filter takes `score` and returns a dict `{"value_class": ..., "threshold_word": ..., "threshold_class": ..., "render_p": ...}`. The template body then renders the resolved values without `{% if %}` cascades — at most one `{% if band.render_p %}` for the optional `<p>`. The filter is in a single file with the five-arm decision.
- **Go:** a package-level function `func resolveComplianceBand(score *int) complianceBand { … }` in `fieldmark-go/internal/web/templates/components/compliance_tile.go` (NEW — siblings the `.html` template). The function returns a value of type `complianceBand struct { ValueClass, ThresholdWord, ThresholdClass string; RenderP bool }`. The struct is funneled into the template context. Five-arm `switch` (one for `nil`, then `>= 90`, `>= 70`, `>= 50`, default).

**And** each stack has a **unit test on the pure function** (separate from the snapshot test): one row per band boundary, asserting the returned tuple matches the expected `(value_class, threshold_word, threshold_class, render_p)` for the canonical input. Boundary inputs to test: `null`, `100`, `90`, `89`, `70`, `69`, `50`, `49`, `0`. Nine inputs. The test names the boundary it exercises so a failure surfaces the exact band edge that drifted. This unit test runs faster than the snapshot test and catches band-logic regressions without re-rendering HTML.

**And** the band function rejects **out-of-range scores** (`score < 0` or `score > 100`) by returning the same tuple as the `null` case — the wrapper renders the no-data variant. Do not throw, do not log this story (the warning-log deferral applies — see Dev Notes). A per-stack unit test asserts `score=-1` and `score=101` both produce the no-data tuple.

### AC4 — `#compliance-tile` is the canonical OOB target for downstream stories

**Given** the canonical HTMX target IDs declared in UX §"State Machine As UI" → "Canonical HTMX target IDs" at [ux-design-specification.md:1114](../planning-artifacts/ux-design-specification.md) and the three-region OOB rule at [ux-design-specification.md:979–981](../planning-artifacts/ux-design-specification.md)
**When** Story 2.10 (Compliance Dashboard), Story 2.11 (Project Detail), Story 2.12 (Place-on-Hold three-region orchestration), and Story 5.5 (Approve Corrective Action anchor demo) compose their pages
**Then** they MUST emit the ComplianceTile wrapper at the position the layout requires (Project Detail header strip uses `id="compliance-tile"`; Compliance Dashboard portfolio tile uses `id="compliance-tile-portfolio"`) without inventing markup. This story does **not** itself land any consumer page, but it does land a **conformance assertion** that the wrapper's output is the canonical target shape:

**And** a per-stack **target-shape conformance test** asserts that the rendered tile contains, in the outer `<section>`: `id="<caller-supplied>"`, `role="status"`, `aria-live="polite"`, `aria-atomic="true"`, and `class="compliance-tile"`. These five attributes are the contract surface that downstream HTMX OOB swaps depend on — if any drifts, OOB updates will silently fail to announce or fail to target. The test is **distinct from the snapshot test**: it asserts presence-of-attribute regardless of byte-equality, so a future cosmetic CSS class addition doesn't break the OOB contract.

**And** a **negative-test pair** asserts that the wrapper does **not** emit any of: `hx-get`, `hx-post`, `hx-target`, `hx-swap`, `hx-trigger`, `<script>` tag, `onload=`, or any `data-*` attribute named `data-htmx-*`. The tile is a pure OOB *target* — it never emits HTMX *producer* attributes. A grep-style assertion in the snapshot-test file verifies none of these tokens appear in the rendered output.

**And** the ComplianceTile `README.md` records the OOB-target invariant explicitly (Story 2.4's per-component README precedent has the "live region is parent-owned" invariant for AuditRow — this story adds the symmetric "this region is the OOB target — wrapper never produces, only receives" invariant).

### AC5 — Cross-stack snapshot-test conformance

**Given** the Story 2.4 snapshot harness (per-stack: `WebApplicationFactory<Program>` partial-render endpoint on .NET; `django.template.loader.render_to_string` on Django; the `bytes.Buffer` template-execute helper on Go) and the Story 1.11 normalization pipeline (whitespace collapse, alphabetical attribute order, trimmed lines)
**When** I run the per-stack snapshot tests for ComplianceTile
**Then** for **every** variant block in `canonical.html` — minimum nine, more if boundary coverage extended per AC1 §boundary-rationale — the corresponding stack wrapper renders output byte-equal to that block after normalization.

**And** the test parses `canonical.html` by `<!-- variant: <name> -->` delimiter at test time using the same parser Story 2.4 introduces. Do not reinvent the parser; if Story 2.4's parser is not yet committed when this story implements, **wait for it** rather than building a parallel one (this is a sequencing constraint, not a scope expansion — see Dev Notes §"Sequencing with Story 2.4").

**Per-stack test paths:**

- **.NET:** `FieldMark/FieldMark.Tests.Web/Components/ComplianceTileSnapshotTests.cs` (or `FieldMark.Tests.Integration/Components/ComplianceTileSnapshotTests.cs` if Story 2.4 elects to land the snapshot tests there — match Story 2.4's host decision; do not split). One `[Theory]` row per variant. Plus `ComplianceTileBandTests.cs` for the AC3 pure-function tests (boundary inputs).
- **Django:** `fieldmark_py/components/tests/test_compliance_tile_snapshot.py` (location matches Story 2.4's resolution of `components/tests/` vs `tests/components/`). One `@pytest.mark.parametrize` case per variant. Plus `test_compliance_band.py` for the AC3 pure-function tests.
- **Go:** `fieldmark-go/internal/web/templates/components/compliance_tile_test.go`. Table-driven sub-tests `t.Run(variantName, …)`. Plus a sibling table-driven test for `resolveComplianceBand` boundary inputs (place in the same file or a sibling `compliance_tile_band_test.go` — Go testing convention is either; match Story 2.4 / 1.12 precedent).

**And** the snapshot test for each stack exercises **all** variants (nine minimum), not a subset. The test harness MUST emit one logical assertion per variant so a failure names the variant (do not `assertAll` — name the variant in the failure).

### AC6 — Component edge-case checklist coverage (per [component-edge-case-checklist.md](../../docs/reference/component-edge-case-checklist.md))

Walked the nine categories — only the **applicable** ones get an AC:

**Given** category 1 (unknown enum / vocabulary values)
**When** ComplianceTile receives an out-of-range score (`< 0` or `> 100`) or `null`
**Then** the wrapper renders the documented `no-data` variant (em-dash, `text-neutral`, threshold `<p>` omitted). A per-stack unit test asserts the `no-data` output is emitted for `score=-1`, `score=101`, and `score=null` (AC3). **No runtime warning logger this story** — the existing "Story 2.4-followup" entry in `deferred-work.md` is extended to cover ComplianceTile out-of-range scores (one-line append to that same entry; do not create a new deferred entry).

**Given** category 6 (text overflow & special characters in user-visible strings)
**When** the wrapper renders the `label` prop (user-controlled in the sense that Story 2.10 / 2.11 / consumer screens supply it from layout config; not directly from a `<form>` POST in this story)
**Then** each framework's default auto-escaping converts `<`, `>`, `&`, `"`, `'` to entities — verified by a per-stack XSS-payload round-trip test: passing `label="<script>alert(1)</script>"` produces `&lt;script&gt;alert(1)&lt;/script&gt;` in the rendered output. **No `|safe`, no `Html.Raw`, no `template.HTML`** in the wrapper this story.

**Given** category 8 (forced-colors / high-contrast mode)
**When** I render ComplianceTile under `forced-colors: active`
**Then** the existing `_a11y.css` rule from Story 1.14 must already cover any styling defenses needed — but ComplianceTile's color is on **text** (`text-success`, `text-warning`, `text-warning-strong`, `text-danger`, `text-neutral`), not on a background or border. In forced-colors mode, these `color:` declarations are overridden by the OS color scheme — which is the correct behavior, **provided** the threshold *word* ("Healthy" / "Watch" / "Concern" / "Critical") is always rendered alongside the color so the information remains carried in text. AC2 mandates the threshold-word `<p>` for all populated variants, satisfying this category by construction.

**Given** category 9 (empty / whitespace text input to derived values)
**When** the wrapper receives an empty or whitespace-only `label` prop
**Then** the wrapper renders the label slot with the empty / whitespace content as-is. There is no "derived initials" or "derived slug" concern here — the label is the literal display text. **However**, a per-stack unit test asserts that a whitespace-only label does not crash the wrapper (no `NullReferenceException`, no template rendering error). The test passes `label=""` and `label="   "` and asserts the rendered output is a valid `<section>` with the label `<p>` present but empty / whitespace.

**Given** categories 2 (font load), 3 (JS init), 4 (AG Grid overlays), 5 (stacking), 7 (reduced motion)
**When** I evaluate against this story's deliverables
**Then** they are **N/A** with this rationale:
- **2 (font load):** ComplianceTile uses no custom fonts beyond what the page already declares (Inter for body, JetBrains Mono for code). The `font-display: swap` rule in `_fonts.css` (Story 1.14) is the existing canonical resolution. No CLS test added this story.
- **3 (JS init):** Zero-JS component. No initialization marker, no progressive enhancement. The tile renders fully functional with `javaScriptEnabled: false`.
- **4 (AG Grid overlays):** Not an AG Grid story.
- **5 (stacking):** No queueing.
- **7 (reduced motion):** UX-DR explicitly defers any threshold-crossing color transition to Phase 3 ("may violate calm register and be cut" per [ux-design-specification.md:952](../planning-artifacts/ux-design-specification.md)). No animations introduced this story; the global `@media (prefers-reduced-motion: reduce)` rule in `_a11y.css` covers any inherited Basecoat animation if one ever lands on `.compliance-tile`.

### AC7 — Security-defaults checklist coverage (per [security-defaults.md](../../docs/reference/security-defaults.md))

Walked the seven categories — only the **applicable** one gets an AC:

**Given** category 3 (allowlist validation on writes) — **adapted here** for output-escaping on read since this story has no writes
**When** the wrapper renders the `label` prop or any other string input
**Then** the framework's default auto-escaping is in force — verified by the XSS-payload round-trip test in AC6 §category-6. **No `|safe` (Django), no `Html.Raw` / `@Html.Raw` (Razor), no `template.HTML` cast (Go) anywhere in the wrapper.** A grep-style conformance check in the per-stack lint lane confirms the absence of these tokens within the new wrapper file (mirror the Story 2.4 grep guard).

**Given** categories 1 (open-redirect), 2 (cookie attributes), 4 (dynamic RegExp), 5 (filesystem writes), 6 (CSRF posture), 7 (stub-auth warnings)
**When** I evaluate against this story's deliverables
**Then** they are **N/A** — no redirects, no cookies, no regex on user input, no filesystem writes, no routes (the wrapper is markup-only — invoked from server-rendered pages but introduces no new HTTP endpoint), no auth changes.

### AC8 — Cross-stack architecture principle three-deliverable check (root [CLAUDE.md](../../CLAUDE.md))

This story introduces **one cross-stack contract surface** — the `#compliance-tile` / `#compliance-tile-portfolio` OOB target — and produces all three required deliverables per the Cross-Stack Architecture Principle:

1. **Documentation contract:** `fieldmark_shared/components/compliance_tile/README.md` (AC1) + the appended row in `docs/reference/component-canonical-examples.md` (AC1) + the OOB-target invariant recorded in the README (AC4).
2. **Native implementation per stack:** the three wrapper templates in idiomatic locations (AC2) + the three pure-function band resolvers (AC3) — each fully native to its stack with no shared codec, no generated stubs, no symlinked partial.
3. **Per-stack conformance test:** the snapshot tests (AC5) + the pure-function boundary tests (AC3) + the target-shape attribute test (AC4) + the negative HTMX-producer-attribute grep (AC4).

**And** there is **no new file** in `fieldmark_shared/src/` that declares the threshold band words, color tokens, or score boundaries as a shared manifest. The four threshold words ("Healthy" / "Watch" / "Concern" / "Critical") and the three boundary values (90, 70, 50) appear in each stack's band-resolver code **and** in `canonical.html` as the source-of-truth fixture. The cross-stack invariant is enforced by the snapshot test, not by a shared symbol file.

**And** the `text-warning-strong` token added to `_tokens.css` (AC2 §token-derivation) is a **CSS rule**, not a symbol manifest — it joins the existing semantic-color token set and is consumed by class-name reference, the same way the Story 1.14 / 2.4 tokens are consumed. This does not violate the principle (the principle prohibits shared *vocabulary lists in code* — it explicitly allows the shared design-system bundle, which is the one symlinked artifact across stacks).

### AC9 — `make parity` clean, no new routes introduced

**Given** all wrappers + tests land
**When** I run `make parity` from the repo root
**Then** the route-parity script reports the **same** drift baseline as Story 2.4 — no new `GET /…` or `POST /…` endpoints are introduced this story. (If the Razor partial-render test endpoint from Story 2.4 (`/_test/render-partial/...`) is reused for ComplianceTile snapshot tests, no new route is added; if a new test endpoint variant is needed, it MUST be gated behind `#if DEBUG` or the equivalent build-time exclusion so it does not appear in the production route table or the parity diff.) `pg_indexes` diff: zero (no DB changes).

### AC10 — Build, type, lint, and test gates green on every stack

- **.NET:** `cd FieldMark && dotnet csharpier check . && dotnet build && dotnet test && dotnet test FieldMark.Tests.Integration/FieldMark.Tests.Integration.csproj` — clean. The snapshot test class passes with one `[Theory]` row per variant; the band test class passes with one row per boundary input.
- **Django:** `cd fieldmark_py && uv run ruff check . && uv run mypy . && uv run pytest && uv run pytest -m integration` — clean. The snapshot test passes with parametrized cases; the band-filter test passes.
- **Go:** `cd fieldmark-go && make check && go test ./... && go test -tags=integration ./...` — clean. Both the template snapshot test and the `resolveComplianceBand` table-driven test pass.
- **`fieldmark_shared`:** `cd fieldmark_shared && pnpm install && pnpm run build` — clean; `dist/fieldmark.css` regenerated and committed (the `text-warning-strong` rule from AC2 propagates into the compiled bundle).
- From repo root: `make parity` exits 0 (AC9) and `make test-all` (canonical pre-merge gate) exits 0.

## Tasks / Subtasks

- [ ] **Task 1: Author canonical example + README in `fieldmark_shared/`** (AC: #1, #2, #8)
  - [ ] 1.1 Create `fieldmark_shared/components/compliance_tile/canonical.html` with nine variant blocks (`healthy-project`, `watch-project`, `concern-project`, `critical-project`, `healthy-portfolio`, `critical-portfolio`, `no-data-project`, `boundary-90`, `boundary-70`).
  - [ ] 1.2 Create `fieldmark_shared/components/compliance_tile/README.md` per AC1 §contract-fixed-order — eight sections (Purpose, Props, Variants, ARIA, Threshold table, Allowed classes, Snapshot equality, Unknown-vocab handling).
  - [ ] 1.3 Append one row to `docs/reference/component-canonical-examples.md` Component Index — `ComplianceTile`, fixture path, README path, three wrapper paths, three test paths.
  - [ ] 1.4 Verify `_tokens.css` declares both `text-warning` and `text-warning-strong`. If `text-warning-strong` is absent, add the CSS variable (`--color-warning-strong`) and the utility selector (`text-warning-strong { color: var(--color-warning-strong); }`) with a dark-mode override using amber-300. Run `pnpm run build`. Commit the regenerated `dist/fieldmark.css`.
  - [ ] 1.5 Verify `_tokens.css` (post-Story-2.4) declares `text-transform: uppercase` on `.dashboard-tile__label` and that the equivalent rule for `.compliance-tile__label` is added (one new selector). If Story 2.4 grouped the uppercase rule under a shared utility, reuse that grouping — do not introduce a new pattern.

- [ ] **Task 2: .NET wrapper + tests** (AC: #2, #3, #4, #5, #6, #7, #10)
  - [ ] 2.1 Create `FieldMark/FieldMark.Web/Pages/Shared/Components/_ComplianceTile.cshtml` — partial + in-file `ComplianceTileViewModel` record with `Score: int?`, `Label: string`, `Id: string`, plus the `ResolveBand` static method returning the four-tuple.
  - [ ] 2.2 Top-of-file comment references `docs/reference/component-canonical-examples.md` (matching Story 2.4 convention).
  - [ ] 2.3 Create `FieldMark/FieldMark.Tests.Web/Components/ComplianceTileSnapshotTests.cs` (or land in `FieldMark.Tests.Integration/Components/` per the Story 2.4 host decision). One `[Theory]` row per variant; reuse the Story 2.4 partial-render scaffold.
  - [ ] 2.4 Create `FieldMark.Tests.{Web,Integration}/Components/ComplianceTileBandTests.cs` — nine `[InlineData]` rows for the boundary inputs (`null, 100, 90, 89, 70, 69, 50, 49, 0`) plus out-of-range (`-1, 101`). Each row asserts the returned tuple matches the expected band.
  - [ ] 2.5 XSS round-trip test in the snapshot test class (one method asserting `label="<script>"` round-trips as `&lt;script&gt;`).
  - [ ] 2.6 Target-shape attribute conformance test (AC4) — one method asserting the five required outer-`<section>` attributes are present for a default render.
  - [ ] 2.7 Negative test (AC4) — assertion that the rendered output does not contain any of `hx-get`, `hx-post`, `hx-target`, `hx-swap`, `hx-trigger`, `<script`, `onload=`, `data-htmx-`.
  - [ ] 2.8 Grep guard (AC7) — CI lane assertion that `Html.Raw` does not appear in `_ComplianceTile.cshtml`.
  - [ ] 2.9 Run `dotnet csharpier check . && dotnet build && dotnet test && dotnet test FieldMark.Tests.Integration/` — clean.

- [ ] **Task 3: Django wrapper + tests** (AC: #2, #3, #4, #5, #6, #7, #10)
  - [ ] 3.1 Create `fieldmark_py/components/templatetags/compliance_tile_tags.py` (NEW) — register a `compliance_band` filter (or simple_tag) that takes `score: Optional[int]` and returns the band dict.
  - [ ] 3.2 Create `fieldmark_py/templates/components/_compliance_tile.html` (NEW) — template body uses the filter result; no `{% if %}` cascades except the single optional `<p class="compliance-tile__threshold">`.
  - [ ] 3.3 Top-of-file comment references `docs/reference/component-canonical-examples.md`.
  - [ ] 3.4 Create `fieldmark_py/components/tests/test_compliance_tile_snapshot.py` — `@pytest.mark.parametrize` over variant names; loads variant block from `canonical.html` via the Story 2.4 path-walker; normalizes; asserts byte-equal.
  - [ ] 3.5 Create `fieldmark_py/components/tests/test_compliance_band.py` — table-driven tests over `compliance_band` filter for the nine boundary inputs + out-of-range.
  - [ ] 3.6 XSS round-trip test; target-shape conformance test; negative HTMX-producer-attribute test; whitespace-label test.
  - [ ] 3.7 Grep guard — CI lane assertion that `|safe` does not appear in `_compliance_tile.html`.
  - [ ] 3.8 Run `uv run ruff check . && uv run mypy . && uv run pytest && uv run pytest -m integration` — clean.

- [ ] **Task 4: Go wrapper + tests** (AC: #2, #3, #4, #5, #6, #7, #10)
  - [ ] 4.1 Create `fieldmark-go/internal/web/templates/components/compliance_tile.go` (NEW) — declare `type ComplianceTileArgs struct { Score *int; Label string; ID string }`, `type complianceBand struct { ValueClass, ThresholdWord, ThresholdClass string; RenderP bool }`, and `func resolveComplianceBand(score *int) complianceBand`.
  - [ ] 4.2 Create `fieldmark-go/internal/web/templates/components/compliance_tile.html` (NEW) — `{{define "compliance_tile"}}…{{end}}`. The template receives a struct that bundles `ComplianceTileArgs` with the resolved `complianceBand` so the template has no logic beyond the single optional-`<p>` conditional.
  - [ ] 4.3 Top-of-file comments in both files reference `docs/reference/component-canonical-examples.md`.
  - [ ] 4.4 Create `fieldmark-go/internal/web/templates/components/compliance_tile_test.go` — mirrors `action_button_test.go` harness (parse template, execute against context, write to `bytes.Buffer`, normalize, compare). Table-driven `t.Run(variantName, …)` per variant.
  - [ ] 4.5 Add table-driven sub-tests (same file or `compliance_tile_band_test.go`) for `resolveComplianceBand` boundary inputs.
  - [ ] 4.6 XSS round-trip test; target-shape conformance test; negative HTMX-producer-attribute test; whitespace-label test.
  - [ ] 4.7 Grep guard — CI lane assertion that `template.HTML(` does not appear in `compliance_tile.go` or `compliance_tile.html`.
  - [ ] 4.8 Run `make check && go test ./... && go test -tags=integration ./...` — clean.

- [ ] **Task 5: Cross-stack verification + parity** (AC: #5, #8, #9, #10)
  - [ ] 5.1 Run `make parity` — route diff equals the Story 2.4 baseline; no new routes. `pg_indexes` zero diff.
  - [ ] 5.2 Run `make test-all` — green.
  - [ ] 5.3 Confirm each new wrapper file's top-of-file comment references `docs/reference/component-canonical-examples.md`.
  - [ ] 5.4 Verify the Component Index in `docs/reference/component-canonical-examples.md` lists ComplianceTile correctly with all paths.
  - [ ] 5.5 Append the one-line ComplianceTile out-of-range extension to the existing "Story 2.4-followup — unknown-token runtime warning logger" entry in `_bmad-output/implementation-artifacts/deferred-work.md`. Do **not** create a new deferred entry — this story's warning-log gap collapses into the existing one because the resolution (per-stack request-scoped logger lookup) is identical.

- [ ] **Task 6: Story sign-off** (AC: all)
  - [ ] 6.1 Populate the Sign-off block below; flip sprint-status to `review`.

## Dev Notes

### Critical context (read before writing code)

- **Markup-only — zero JS, zero handlers, zero routes, zero animations.** ComplianceTile is a pure rendering wrapper. The threshold-crossing color transition is **explicitly deferred** to Phase 3 (UX §"Phase 3 — Polish optimizations" flags it as "may violate calm register and be cut"). If you find yourself adding a `<script>`, a CSS transition on `.compliance-tile__value`, a `hx-trigger`, or any per-stack HTTP endpoint, you are out of scope. Stop and re-read AC4.
- **The cross-stack contract is byte-equality of rendered HTML**, same as Story 2.4. Each stack's template syntax differs; the snapshot tests assert the *output* matches `canonical.html`. The Razor in-file view model, the Django template tag for band resolution, and the Go sibling `.go` file are three different mechanisms for the same logical transformation — that's idiomatic per stack, which is what the Cross-Stack Architecture Principle requires.
- **Sequencing with Story 2.4.** This story uses the per-component directory convention, the variant-delimiter parser, the snapshot-test harness, and the path-walker that Story 2.4 introduces. If Story 2.4 is not yet **done** when this story enters implementation, **block on it** — do not pre-implement those primitives here. Specifically: do not create the Story 2.4 parser; do not create the Razor partial-render endpoint; do not modify `docs/reference/component-canonical-examples.md` before Story 2.4 lands the file. The dependency is one-way (2.5 depends on 2.4, not vice versa) and the artifacts must arrive in order. If during implementation you find a Story 2.4 primitive missing, flag it as a Story 2.4 review-round patch item rather than reimplementing it here.
- **The `id` prop is caller-supplied, not derived from `context`.** Story 2.10 (Compliance Dashboard) passes `id="compliance-tile-portfolio"`; Story 2.11 (Project Detail) passes `id="compliance-tile"`. The wrapper does not look at `context` to pick the id — `context` exists in `canonical.html` and `README.md` only as variant-label documentation. This decision keeps the wrapper API minimal (three props: `score`, `label`, `id`) and makes the cross-stack template syntax trivial. If a future consumer needs a third id (e.g., for an embedded sub-region), the wrapper accommodates it without code change.
- **No-data is `null`, not `0`.** A project with `compliance_score = 0` is in the Critical band — it has scored zero, which is a real datum. A project that has **not yet been scored** (no inspections completed) is no-data — em-dash render. The three stacks model this as `int?` (.NET), `Optional[int]` (Django — though templates receive `score=None`), `*int` (Go). The per-stack tests for `score=0` vs `score=null` enforce the distinction.
- **The four threshold words are part of the byte-equality contract.** "Healthy" / "Watch" / "Concern" / "Critical" are English strings; they appear in `canonical.html` as literal text. Internationalization is out of scope (PRD does not list i18n as MVP). If a future i18n story lands, the threshold words become resource-keys and `canonical.html` carries the English defaults — but the cross-stack invariant is the *rendered* English text today.
- **Boundary coverage rationale.** UX §"Compliance score thresholds" specifies the bands as `≥ 90`, `70–89`, `50–69`, `< 50`. The boundaries (89/90, 69/70, 49/50) are where off-by-one bugs hide. The nine variant blocks include explicit `boundary-90` and `boundary-70` cases; the band test (AC3) exercises `100, 90, 89, 70, 69, 50, 49, 0` plus `null` and `-1` / `101`. If the test author judges the surface needs more (e.g., `91`, `71`, `51`) add them — the nine variants in `canonical.html` are the floor, the band test is the precision instrument.
- **Decision — unknown-token / out-of-range handling.** Same disposition as Story 2.4's "Decision — unknown-token handling": render the documented fallback (no-data variant for out-of-range scores) **without** a runtime warning log this story. The Story 2.4-followup entry in `deferred-work.md` is extended (one-line append, not a new entry) to cover ComplianceTile out-of-range scores. Rationale: identical to Story 2.4's — the per-stack request-scoped logger lookup is the friction; the user-visible signal (the em-dash em-render) is in place.
- **`text-warning-strong` is one new token, not a new pattern.** If Story 2.4's `_tokens.css` edits already added it (Story 2.4 references `text-warning` and a "darker" warning in the DashboardTile `populated-with-color` variant — verify before adding), reuse it. If not, this story adds it. Do not invent a competing token name (`text-warning-2`, `text-amber-darker`, etc.) — match the semantic-token naming convention already established (`text-success`, `text-warning`, `text-danger`, `text-info`, `text-neutral`).

### Component-specific notes

- **`role="status"` + `aria-live="polite"` + `aria-atomic="true"` is the announcement triad.** The OOB swap that updates `#compliance-tile` (Story 2.12, 5.5) replaces the entire `<section>` content — `aria-atomic="true"` causes the screen reader to announce the whole tile as a unit (e.g., "Portfolio Compliance, 87, Watch") rather than incrementally. This is by design and is the user-visible payoff of the OOB pattern. Do not weaken any of the three ARIA attributes.
- **`tnum` on the value is non-negotiable.** UX §"Tabular numerals" pins `font-feature-settings: "tnum"` on all updating numbers, specifically calling out compliance score values. When the OOB swap replaces `95` with `89`, the `tnum` utility prevents the column from jittering. The `.tnum` class is declared in `_tokens.css` per `fieldmark_shared/CLAUDE.md`. Do not parameterize this — every populated variant carries `tnum`.
- **The `compliance-tile__label`, `compliance-tile__value`, `compliance-tile__threshold` class names are BEM-style and intentional.** They match the Story 2.4 DashboardTile pattern (`dashboard-tile__label`, `dashboard-tile__value`, `dashboard-tile__secondary`). Future styling adjustments target these classes; the semantic-color token classes (`text-success` etc.) are layered on top. Do not collapse the BEM classes into a single `compliance-tile` selector — the structure carries the styling hooks.

### Edge cases (per [component-edge-case-checklist.md](../../docs/reference/component-edge-case-checklist.md))

Walked the nine categories — see AC6 for the per-category ACs. Summary: categories **1**, **6**, **8**, **9** apply and have ACs; categories **2**, **3**, **4**, **5**, **7** are N/A with rationale recorded in AC6.

### Security defaults (per [security-defaults.md](../../docs/reference/security-defaults.md))

Walked the seven categories — see AC7 for the per-category ACs. Summary: category **3** applies (adapted as output-escaping on read since this story has no writes); categories **1**, **2**, **4**, **5**, **6**, **7** are N/A with rationale recorded in AC7.

### Cross-stack contract three-deliverable check

This story introduces one cross-stack contract surface — the `#compliance-tile` / `#compliance-tile-portfolio` OOB target shape and the threshold-band rule — and produces all three deliverables: (1) the per-component `README.md` plus the appended row in `docs/reference/component-canonical-examples.md`, (2) per-stack wrapper templates + pure band resolvers in idiomatic locations, (3) per-stack snapshot tests, band-resolver boundary tests, target-shape attribute tests, and HTMX-producer-attribute negative grep. See AC8.

### Files this story modifies vs creates

| File | New / Modified | Purpose |
|---|---|---|
| `fieldmark_shared/components/compliance_tile/canonical.html` | NEW | nine variant blocks |
| `fieldmark_shared/components/compliance_tile/README.md` | NEW | contract |
| `fieldmark_shared/src/_tokens.css` | MODIFY (only if needed) | add `text-warning-strong` token + `.compliance-tile__label { text-transform: uppercase }` rule (or reuse Story 2.4 grouping) |
| `fieldmark_shared/dist/fieldmark.css` | MODIFY (regenerated) | commit after `pnpm run build` |
| `docs/reference/component-canonical-examples.md` | MODIFY | append one row to the Component Index |
| `FieldMark/FieldMark.Web/Pages/Shared/Components/_ComplianceTile.cshtml` | NEW | wrapper with in-file view model |
| `FieldMark/FieldMark.Tests.{Web,Integration}/Components/ComplianceTileSnapshotTests.cs` | NEW | snapshot tests |
| `FieldMark/FieldMark.Tests.{Web,Integration}/Components/ComplianceTileBandTests.cs` | NEW | band-resolver boundary tests |
| `fieldmark_py/components/templatetags/compliance_tile_tags.py` | NEW | `compliance_band` filter |
| `fieldmark_py/templates/components/_compliance_tile.html` | NEW | wrapper |
| `fieldmark_py/components/tests/test_compliance_tile_snapshot.py` | NEW | snapshot tests |
| `fieldmark_py/components/tests/test_compliance_band.py` | NEW | band-filter boundary tests |
| `fieldmark-go/internal/web/templates/components/compliance_tile.go` | NEW | args struct + `resolveComplianceBand` |
| `fieldmark-go/internal/web/templates/components/compliance_tile.html` | NEW | `{{define "compliance_tile"}}` wrapper |
| `fieldmark-go/internal/web/templates/components/compliance_tile_test.go` | NEW | snapshot + band tests (or split into a sibling `_band_test.go`) |
| `_bmad-output/implementation-artifacts/deferred-work.md` | MODIFY | extend the existing "Story 2.4-followup — unknown-token runtime warning logger" entry to cover ComplianceTile out-of-range scores (one-line append) |

Anything outside this list — Compliance Dashboard page, Project Detail page, Place-on-Hold orchestration, Approve CA orchestration, compliance-score computation logic, threshold-crossing animation, route registration, any DB change — is out of scope. Resist the urge.

### Files to read fully before editing

- [_bmad-output/planning-artifacts/ux-design-specification.md:478–489](../planning-artifacts/ux-design-specification.md) — semantic color tokens and compliance-score threshold table. Binding for AC2.
- [_bmad-output/planning-artifacts/ux-design-specification.md:820–956](../planning-artifacts/ux-design-specification.md) — Custom Components section, specifically the ComplianceTile contract at lines 835–841 and the Phase-3 deferral of threshold transitions at line 952.
- [_bmad-output/planning-artifacts/ux-design-specification.md:1114](../planning-artifacts/ux-design-specification.md) — canonical HTMX target IDs (`#compliance-tile` is listed).
- [_bmad-output/planning-artifacts/ux-design-specification.md:979–981](../planning-artifacts/ux-design-specification.md) — three-region OOB orchestration rule (the downstream consumers of `#compliance-tile`).
- [_bmad-output/planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md](../planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md) §Story 2.5 — epic AC source.
- [docs/reference/component-edge-case-checklist.md](../../docs/reference/component-edge-case-checklist.md) — nine-category walkthrough; binding for AC6.
- [docs/reference/security-defaults.md](../../docs/reference/security-defaults.md) — seven-category walkthrough; binding for AC7.
- [fieldmark_shared/CLAUDE.md](../../fieldmark_shared/CLAUDE.md) §"Snapshot-test pipeline" — normalization pipeline this story's tests inherit.
- [_bmad-output/implementation-artifacts/2-4-implement-phase-2-markup-only-components-statusbadge-inlinealert-auditrow-dashboardtile.md](2-4-implement-phase-2-markup-only-components-statusbadge-inlinealert-auditrow-dashboardtile.md) — per-component-directory convention, snapshot harness, path-walker, host decision for tests, `_tokens.css` uppercase-label rule precedent. This story extends Story 2.4 mechanically.
- [fieldmark_shared/components/action_button.example.html](../../fieldmark_shared/components/action_button.example.html) — variant-delimiter format precedent.
- [FieldMark/FieldMark.Web/Pages/Shared/_ActionButton.cshtml](../../FieldMark/FieldMark.Web/Pages/Shared/_ActionButton.cshtml) — .NET wrapper precedent (in-file view model + static helper method).
- [fieldmark_py/templates/components/_action_button.html](../../fieldmark_py/templates/components/_action_button.html) — Django `{% include … with … %}` precedent.
- [fieldmark-go/internal/web/templates/components/action_button.html](../../fieldmark-go/internal/web/templates/components/action_button.html) and [_test.go sibling](../../fieldmark-go/internal/web/templates/components/action_button_test.go) — Go `{{define}}` + sibling-`.go`-helper + table-driven test precedent.
- Stack rules: [FieldMark/CLAUDE.md](../../FieldMark/CLAUDE.md), [fieldmark_py/CLAUDE.md](../../fieldmark_py/CLAUDE.md), [fieldmark-go/CLAUDE.md](../../fieldmark-go/CLAUDE.md).
- Root cross-stack invariants: [CLAUDE.md](../../CLAUDE.md) §"Cross-Stack Architecture Principle" — binding for AC8.

### Project Structure Notes

- The `Pages/Shared/Components/` Razor sub-directory is created by Story 2.4. This story drops `_ComplianceTile.cshtml` alongside the four Story 2.4 wrappers; if Story 2.4 has not yet created the sub-directory at implementation time, create it here (block on the host decision Story 2.4 makes — `FieldMark.Tests.Web` vs `FieldMark.Tests.Integration` for the snapshot test project).
- The Django `templates/components/` and `components/tests/` directories are touched by Story 2.4 — verify the chosen test-location resolution and match it.
- The Go `internal/web/templates/components/` directory is the established location. The sibling-`.go`-file pattern (`compliance_tile.go` next to `compliance_tile.html` + `compliance_tile_test.go`) is consistent with Go convention; no precedent yet exists for a sibling helper file alongside a template (Story 2.4's wrappers are template-only), so this story introduces it. Document the rationale in the file's top-of-file comment ("Sibling helper file for `compliance_tile.html` — hosts pure band-resolver because Go's `html/template` cannot express the four-arm decision concisely. Other component templates that ship without a logic surface remain template-only.").

### References

- AC source: [_bmad-output/planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md](../planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md) §Story 2.5
- UX-DR ComplianceTile spec: [ux-design-specification.md:835–841](../planning-artifacts/ux-design-specification.md)
- UX-DR threshold table: [ux-design-specification.md:478–489](../planning-artifacts/ux-design-specification.md)
- Canonical HTMX target IDs: [ux-design-specification.md:1114](../planning-artifacts/ux-design-specification.md)
- Three-region OOB orchestration: [ux-design-specification.md:979–981](../planning-artifacts/ux-design-specification.md)
- Phase-3 deferral of threshold transitions: [ux-design-specification.md:952](../planning-artifacts/ux-design-specification.md)
- Component edge-case checklist: [docs/reference/component-edge-case-checklist.md](../../docs/reference/component-edge-case-checklist.md)
- Security defaults checklist: [docs/reference/security-defaults.md](../../docs/reference/security-defaults.md)
- Snapshot-test pipeline: [fieldmark_shared/CLAUDE.md](../../fieldmark_shared/CLAUDE.md) §"Snapshot-test pipeline"
- Per-component-directory convention precedent: [Story 2.4](2-4-implement-phase-2-markup-only-components-statusbadge-inlinealert-auditrow-dashboardtile.md)
- Wrapper precedent (Story 1.12): [.NET _ActionButton.cshtml](../../FieldMark/FieldMark.Web/Pages/Shared/_ActionButton.cshtml), [Django _action_button.html](../../fieldmark_py/templates/components/_action_button.html), [Go action_button.html](../../fieldmark-go/internal/web/templates/components/action_button.html)
- Cross-Stack Architecture Principle: root [CLAUDE.md](../../CLAUDE.md) §Cross-Stack Architecture Principle
- Stack rules: [FieldMark/CLAUDE.md](../../FieldMark/CLAUDE.md), [fieldmark_py/CLAUDE.md](../../fieldmark_py/CLAUDE.md), [fieldmark-go/CLAUDE.md](../../fieldmark-go/CLAUDE.md)

## Dev Agent Record

### Agent Model Used

_to be populated by dev-story_

### Debug Log References

### Completion Notes List

### File List

## Sign-off

| Field | Value |
|---|---|
| Final review date | _pending_ |
| Total review rounds | 0 |
| Final reviewer verdict | _pending — story created, status `ready-for-dev`_ |
| Deferred-work entries | _none new — Story 2.4-followup entry in `deferred-work.md` is extended (one-line append) at implementation time to cover ComplianceTile out-of-range scores per Dev Notes §"Decision — unknown-token / out-of-range handling"_ |
| Dev-notes divergences from epic AC | The epic AC enumerates four threshold bands and mandates `aria-live="polite" aria-atomic="true"` on `#compliance-tile`. This story adds (a) the explicit `text-warning-strong` token for the 50–69 band's "warning (darker)" distinction, justified by UX §lines 478–489; (b) the caller-supplied-`id` design rather than deriving id from `context`, justified for wrapper-API minimality; (c) a separate target-shape attribute conformance test + negative HTMX-producer-attribute test, justified by AC4 OOB-target-contract surface concerns. None of these contradict the epic AC; all are additive precision. |

### Review Findings

_to be populated by code-review_
