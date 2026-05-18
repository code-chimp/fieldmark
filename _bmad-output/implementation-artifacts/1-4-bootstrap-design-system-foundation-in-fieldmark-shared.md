# Story 1.4: Bootstrap Design System Foundation in `fieldmark_shared/`

Status: done

## Story

As a developer styling any FieldMark screen on any stack,
I want one compiled CSS bundle with the Basecoat component vocabulary, semantic tokens, status-badge vocabulary, typography, and vendored JS,
So that I can render byte-identical markup across the three stacks without authoring per-stack CSS.

## Acceptance Criteria

1. **Given** the repo at HEAD **When** I inspect `fieldmark_shared/package.json` **Then** `tailwindcss@4.x` is pinned to an exact patch and `basecoat-css` is pinned to an exact pre-1.0 patch (e.g., `0.3.11`) — no `^` or `~` ranges (UX-DR1) **And** the version pins are documented in `_bmad-output/planning-artifacts/architecture.md` alongside HTMX and AG Grid.

2. **Given** `fieldmark_shared/src/fieldmark.css` **When** I read it **Then** it imports Basecoat's CSS, the AG Grid Quartz theme, and declares the five semantic color tokens `--color-success`, `--color-warning`, `--color-danger`, `--color-info`, `--color-neutral` (UX-DR2) with both light and dark variants **And** each token meets ≥ 4.5:1 contrast against `neutral-50/100` and `neutral-900/950`, with a one-line comment recording the contrast ratio at design time.

3. **Given** the same file **When** I read it **Then** the status-badge color vocabulary (UX-DR3) for Project, Inspection, Violation (with severity overlay), CorrectiveAction, and Severity is encoded as deterministic class-to-token mappings **And** the compliance-score threshold mapping (UX-DR4) is encoded as a single CSS rule keyed on `data-score-band` (`healthy`, `watch`, `concern`, `critical`).

4. **Given** `fieldmark_shared/src/` **When** I read its CSS **Then** Inter and JetBrains Mono are referenced as `@font-face` declarations pointing to self-hosted woff2 files under `fieldmark_shared/vendor/fonts/` (UX-DR6) **And** body default is `text-sm` (14px), `font-feature-settings: "tnum"` is applied via a `.tnum` utility to compliance score, timestamps, counts, and any DOM element with numeric updating values.

5. **Given** the spacing scale (UX-DR8) **When** I read `fieldmark_shared/src/_layout.css` **Then** it uses only Tailwind defaults — no custom breakpoints — and documents `max-w-screen-2xl` container + `px-6 → px-4` gutter collapse with a single comment per rule naming the collapse point it implements.

6. **Given** the vendored JS strategy (Architecture D15) **When** I inspect `fieldmark_shared/vendor/` **Then** `htmx/htmx.min.js` and `ag-grid/35.2.1/ag-grid-community.min.js` are committed; each stack's `vendor/` static dir has directory symlinks pointing here **And** each stack's static directory symlinks `dist/fieldmark.css` and the vendor directory.

7. **Given** the design system is built **When** I run `cd fieldmark_shared && pnpm run build` (alias `make css`) **Then** `fieldmark_shared/dist/fieldmark.css` is produced **And** the compiled file is committed (no build step required after clone).

## Tasks / Subtasks

- [x] Task 1: Add Basecoat dependency and pin versions (AC: #1)
  - [x] 1.1: Install `basecoat-css` (npm package name) at version `0.3.11` with exact version pin in `fieldmark_shared/package.json`
  - [x] 1.2: Pin `tailwindcss` and `@tailwindcss/cli` to exact versions (remove any `^` or `~`)
  - [x] 1.3: Run `pnpm install` to update lockfile
  - [x] 1.4: Document Basecoat version pin in `architecture.md` alongside HTMX 4.0.0-beta2 and AG Grid 35.2.1

- [x] Task 2: Self-host Inter and JetBrains Mono fonts (AC: #4)
  - [x] 2.1: Download Inter variable woff2 into `fieldmark_shared/vendor/fonts/inter/`
  - [x] 2.2: Download JetBrains Mono variable woff2 into `fieldmark_shared/vendor/fonts/jetbrains-mono/`
  - [x] 2.3: Create `fieldmark_shared/src/_fonts.css` with `@font-face` declarations

- [x] Task 3: Restructure `fieldmark.css` with Basecoat + semantic tokens (AC: #2, #3)
  - [x] 3.1: Import Basecoat CSS into `fieldmark.css`
  - [x] 3.2: Import `_fonts.css`
  - [x] 3.3: Define the five semantic color tokens with light and dark variants
  - [x] 3.4: Add contrast ratio comments for each token
  - [x] 3.5: Define status-badge class-to-token mappings for all entity states
  - [x] 3.6: Define compliance-score threshold mapping keyed on `data-score-band`

- [x] Task 4: Create AG Grid theme overrides (AC: #2)
  - [x] 4.1: Create `fieldmark_shared/src/_ag-grid.css` with Quartz theme overrides aligning AG Grid with Basecoat palette

- [x] Task 5: Create layout foundation (AC: #5)
  - [x] 5.1: Create `fieldmark_shared/src/_layout.css` with container, gutter collapse, and spacing documentation

- [x] Task 6: Add `.tnum` utility class (AC: #4)
  - [x] 6.1: Define `.tnum` with `font-feature-settings: "tnum"` in the CSS source

- [x] Task 7: Compile and verify (AC: #7)
  - [x] 7.1: Run `pnpm run build` — verify `dist/fieldmark.css` is produced
  - [x] 7.2: Verify all symlinks still resolve correctly in all three stacks
  - [x] 7.3: Commit `dist/fieldmark.css`

- [x] Task 8: Verify existing vendor JS and symlinks (AC: #6)
  - [x] 8.1: Confirm `vendor/htmx/htmx.min.js` and `vendor/ag-grid/35.2.1/ag-grid-community.min.js` exist
  - [x] 8.2: Confirm all three stacks' symlinks resolve

## Dev Notes

### Brownfield Posture

`fieldmark_shared/` already exists with a working Tailwind v4 build pipeline. The current state:

- `package.json` has `@tailwindcss/cli: "4.2.4"` (devDependency). **No Basecoat yet.**
- `src/fieldmark.css` has only `@import "tailwindcss"` and three `@source` directives pointing to the three stacks' template directories.
- `dist/fieldmark.css` exists and is committed — it's a vanilla Tailwind output with reset + utilities.
- `vendor/htmx/htmx.min.js` (v4.0.0-beta2) and `vendor/ag-grid/35.2.1/ag-grid-community.min.js` are committed.
- All three stacks have working symlinks in their `vendor/` static dirs pointing to `fieldmark_shared/dist/fieldmark.css`, `vendor/ag-grid`, and `vendor/htmx`.
- `pnpm` is the package manager (pnpm-lock.yaml present, not npm or yarn).

**Do not break existing symlinks.** The `dist/fieldmark.css` symlink path and the `vendor/` directory symlinks are already wired and working.

### Architectural Constraints

- **Architecture D14** — AG Grid Quartz theme compiled into `fieldmark_shared/dist/fieldmark.css` as part of the Tailwind compile pass. Overrides in `_ag-grid.css`.
- **Architecture D15** — Vendor locally, no CDN. HTMX and AG Grid are already vendored.
- **Architecture D16** — Manual Tailwind compile via `pnpm run build`. Compiled `dist/` is committed. No watcher needed in CI.
- **UX-DR1** — Basecoat pinned to exact pre-1.0 patch version (no `^` or `~` ranges).
- **UX-DR2** — Five semantic color tokens with ≥ 4.5:1 contrast in both light and dark.
- **UX-DR3** — Status-badge color vocabulary is deterministic (entity-state → token).
- **UX-DR4** — Compliance score thresholds: ≥90 healthy, 70–89 watch, 50–69 concern, <50 critical.
- **UX-DR6** — Self-hosted Inter and JetBrains Mono; no Google Fonts.
- **UX-DR8** — Tailwind defaults only. No custom breakpoints. Standard spacing scale.

### Basecoat Integration

**Package name:** `basecoat-css` on npm ([npmjs.com/package/basecoat-css](https://npmjs.com/package/basecoat-css)). The project is at [github.com/hunvreus/basecoat](https://github.com/hunvreus/basecoat). **Not** the `basecoat` package (that's an unrelated 0.1.0 package).

**Latest version:** `0.3.11` (latest as of 2026-05-17). Available versions range from 0.1.0 to 0.3.11.

**Tailwind v4 integration:** Basecoat is Tailwind v4–native. Import Basecoat's CSS in `fieldmark.css` using:

```css
@import "basecoat-css";
```

This imports Basecoat's component styles into the Tailwind v4 pipeline. The `@import` must come after `@import "tailwindcss"` so Basecoat's component classes are available.

**Version pinning:** Pin with exact version in `package.json`:

```json
"basecoat-css": "0.3.11"
```

**Important:** Basecoat is pre-1.0. Breaking changes possible between minor versions. Pin exact and treat upgrades as a coordinated three-stack story.

### Semantic Color Token Definitions

Exact token values from the UX Design Specification:

| Token | Light mode | Dark mode | Usage |
|---|---|---|---|
| `--color-success` | `emerald-600` | `emerald-400` | Resolved, Approved, Pass, Healthy score |
| `--color-warning` | `amber-600` | `amber-400` | InProgress, UnderReview, Conditional, mid-band score |
| `--color-danger` | `rose-600` | `rose-400` | Critical/High severity, Failed, low score, 409 errors |
| `--color-info` | `sky-600` | `sky-400` | Submitted, neutral informational |
| `--color-neutral` | `slate-600` | `slate-400` | Open, Scheduled, default pre-action state |

Use Tailwind v4's `oklch()` color values for these. The Tailwind default palette values:

- `emerald-600`: `oklch(0.596 0.145 163.225)` / `emerald-400`: `oklch(0.765 0.177 163.223)`
- `amber-600`: `oklch(0.666 0.179 58.318)` / `amber-400`: `oklch(0.828 0.189 84.429)`
- `rose-600`: `oklch(0.586 0.209 16.833)` / `rose-400`: `oklch(0.712 0.194 13.428)`
- `sky-600`: `oklch(0.588 0.158 231.824)` / `sky-400`: `oklch(0.746 0.16 232.661)`
- `slate-600`: `oklch(0.446 0.043 257.281)` / `slate-400`: `oklch(0.704 0.04 256.788)`

Each must meet ≥ 4.5:1 against `neutral-50/100` (light surfaces) and `neutral-900/950` (dark surfaces). Record the measured ratio in a comment.

### Status Badge Vocabulary

Deterministic entity-state → token mappings. Use CSS classes that encode the entity and state:

| Entity | State | Token | CSS Class |
|---|---|---|---|
| Project | Active | `--color-success` | `.badge-project-active` |
| Project | OnHold | `--color-warning` | `.badge-project-onhold` |
| Project | Closed | `--color-neutral` | `.badge-project-closed` |
| Inspection | Scheduled | `--color-info` | `.badge-inspection-scheduled` |
| Inspection | InProgress | `--color-warning` | `.badge-inspection-inprogress` |
| Inspection | Completed (Pass) | `--color-success` | `.badge-inspection-pass` |
| Inspection | Completed (Conditional) | `--color-warning` | `.badge-inspection-conditional` |
| Inspection | Completed (Fail) | `--color-danger` | `.badge-inspection-fail` |
| Inspection | Cancelled | `--color-neutral` | `.badge-inspection-cancelled` |
| Violation | Open (Critical/High) | `--color-danger` | `.badge-violation-open-high` |
| Violation | Open (Medium/Low) | `--color-warning` | `.badge-violation-open-low` |
| Violation | InProgress | `--color-warning` | `.badge-violation-inprogress` |
| Violation | Resolved | `--color-success` | `.badge-violation-resolved` |
| Violation | Voided | `--color-neutral` | `.badge-violation-voided` |
| CorrectiveAction | Submitted | `--color-info` | `.badge-ca-submitted` |
| CorrectiveAction | UnderReview | `--color-warning` | `.badge-ca-underreview` |
| CorrectiveAction | Approved | `--color-success` | `.badge-ca-approved` |
| CorrectiveAction | Rejected | `--color-danger` | `.badge-ca-rejected` |
| Severity | Critical | `--color-danger` (filled) | `.badge-severity-critical` |
| Severity | High | `--color-danger` (outline) | `.badge-severity-high` |
| Severity | Medium | `--color-warning` | `.badge-severity-medium` |
| Severity | Low | `--color-neutral` | `.badge-severity-low` |

### Compliance Score Threshold Mapping

CSS rule keyed on `data-score-band` attribute:

```css
[data-score-band="healthy"] { color: var(--color-success); }
[data-score-band="watch"]   { color: var(--color-warning); }
[data-score-band="concern"] { color: var(--color-warning); }  /* darker variant */
[data-score-band="critical"]{ color: var(--color-danger); }
```

Score → band: ≥ 90 = healthy, 70–89 = watch, 50–69 = concern, < 50 = critical. Band assignment is server-side (emitted as `data-score-band` attribute). CSS only styles it.

### Typography Setup

**`@font-face` declarations** in `_fonts.css`:

- Inter variable (woff2) — regular axis covers all weights 100–900.
- JetBrains Mono variable (woff2) — regular axis covers standard weights.

Both files go under `vendor/fonts/`. Download the variable font `.woff2` files from the official releases:
- Inter: https://github.com/rsms/inter/releases (get `InterVariable.woff2`)
- JetBrains Mono: https://github.com/JetBrains/JetBrainsMono/releases (get `JetBrainsMono[wght].woff2`)

Set the Tailwind font families in `fieldmark.css`:

```css
@theme {
  --font-sans: "Inter", ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
  --font-mono: "JetBrains Mono", ui-monospace, SFMono-Regular, Menlo, monospace;
}
```

Default body text is `text-sm` (14px) — this is set in the layout CSS or base styles.

### AG Grid Theme Overrides

`_ag-grid.css` aligns the AG Grid Quartz theme's background, border, and text colors with the Basecoat/Tailwind neutral palette. Import AG Grid's Quartz CSS and override CSS custom properties. The grid should not look foreign within the FieldMark UI.

### CSS File Organization

After this story, `fieldmark_shared/src/` contains:

| File | Purpose |
|---|---|
| `fieldmark.css` | Main entry point — imports Tailwind, Basecoat, and all partials |
| `_fonts.css` | `@font-face` declarations for Inter and JetBrains Mono |
| `_tokens.css` | Semantic color tokens, status-badge vocabulary, score-band mapping |
| `_layout.css` | Container, gutter collapse, spacing documentation |
| `_ag-grid.css` | AG Grid Quartz theme overrides |

The underscore-prefix convention signals partials imported by `fieldmark.css`.

### Source Tree — Files to Create/Modify

| File | Type | Description |
|---|---|---|
| `fieldmark_shared/package.json` | UPDATE | Add `basecoat` dependency; pin all versions exactly |
| `fieldmark_shared/pnpm-lock.yaml` | UPDATE | Regenerated by `pnpm install` |
| `fieldmark_shared/src/fieldmark.css` | UPDATE | Add Basecoat import, partial imports, `@theme` overrides |
| `fieldmark_shared/src/_fonts.css` | NEW | `@font-face` declarations for Inter and JetBrains Mono |
| `fieldmark_shared/src/_tokens.css` | NEW | Semantic tokens, status-badge classes, score-band mapping |
| `fieldmark_shared/src/_layout.css` | NEW | Container, gutter, spacing foundation |
| `fieldmark_shared/src/_ag-grid.css` | NEW | AG Grid Quartz theme overrides |
| `fieldmark_shared/vendor/fonts/inter/InterVariable.woff2` | NEW | Self-hosted Inter font |
| `fieldmark_shared/vendor/fonts/jetbrains-mono/JetBrainsMono[wght].woff2` | NEW | Self-hosted JetBrains Mono font |
| `fieldmark_shared/dist/fieldmark.css` | UPDATE | Recompiled output |
| `_bmad-output/planning-artifacts/architecture.md` | UPDATE | Document Basecoat version pin |
| `fieldmark_shared/CLAUDE.md` | UPDATE | Document new CSS partials and font vendoring |

### Files to READ Before Implementation

- `fieldmark_shared/package.json` — current dependencies; exact version format
- `fieldmark_shared/src/fieldmark.css` — current CSS entry point; preserve `@source` directives
- `fieldmark_shared/CLAUDE.md` — current documentation; update with new structure
- `fieldmark_shared/dist/fieldmark.css` — verify current build output structure
- `_bmad-output/planning-artifacts/architecture.md` (D14, D15, D16 sections) — where to add Basecoat version pin

### Hard Rules

- **Do not break existing symlinks.** All three stacks' `vendor/` symlinks must continue to resolve after changes.
- **Do not modify template files in any stack.** This story is CSS infrastructure only.
- **Do not introduce per-stack CSS.** All CSS lives in `fieldmark_shared/src/`.
- **Pin all npm dependencies to exact versions.** No `^` or `~` ranges.
- **Use `pnpm` as the package manager.** Do not switch to `npm` or `yarn`.
- **Commit `dist/fieldmark.css` and all `vendor/fonts/` files.** These are not gitignored.
- **Do not modify `.gitignore` to exclude font files or dist.**
- **Preserve the existing `@source` directives** in `fieldmark.css` — they scan all three stacks for Tailwind class references.
- **Self-host all fonts.** No external requests to Google Fonts or any CDN.

### Testing Strategy

1. Run `cd fieldmark_shared && pnpm install && pnpm run build` — must succeed.
2. Verify `dist/fieldmark.css` contains Basecoat component styles (search for known Basecoat class selectors).
3. Verify semantic color tokens are defined in both light and dark variants.
4. Verify `.tnum` utility class is present in the compiled output.
5. Verify status-badge classes are present in the compiled output.
6. Verify `[data-score-band]` rules are present.
7. Verify font files exist at `vendor/fonts/inter/` and `vendor/fonts/jetbrains-mono/`.
8. Verify all three stacks' symlinks still resolve: `ls -la FieldMark/FieldMark.Web/wwwroot/vendor/fieldmark.css`, `ls -la fieldmark_py/static/vendor/fieldmark.css`, `ls -la fieldmark-go/internal/web/static/vendor/fieldmark.css`.
9. Start each stack's dev server and verify the page loads without CSS errors.

### Previous Story Intelligence (Story 1.3)

- Story 1.3 established the `tools/parity/` shell scripts and per-stack route dump commands. No CSS changes.
- The `fieldmark_shared/` directory was not modified by stories 1.1–1.3.
- The shell script convention (`set -euo pipefail`, `#!/usr/bin/env bash`) established by Story 1.2 does not apply here (this story is CSS/npm only).
- Stories 1.1 and 1.2 confirmed all three stacks build and run. After this story, verify they still do.

### Git Intelligence

- Latest commits (1.1, 1.2) confirmed scaffolds and SQL init scripts.
- `fieldmark_shared/` was scaffolded during the pre-planning phase (commit `a4fcc76` and earlier).
- The package manager is `pnpm` — lockfile is `pnpm-lock.yaml`.
- Tailwind CLI `4.2.4` is already installed as `@tailwindcss/cli`.

### Project Structure Notes

- `fieldmark_shared/src/` currently has only `fieldmark.css`. This story adds four new partial files (`_fonts.css`, `_tokens.css`, `_layout.css`, `_ag-grid.css`).
- `fieldmark_shared/vendor/` currently has `htmx/` and `ag-grid/`. This story adds `fonts/` with two subdirectories.
- The `dist/` directory continues to contain only `fieldmark.css` — compiled output.
- No new symlinks are created by this story; existing symlinks are preserved.

### Implementation Notes

- **Inter v4.1** `InterVariable.woff2` was sourced from the official GitHub release zip (`Inter-4.1.zip`). The variable font was in `web/InterVariable.woff2` within the archive.
- **JetBrains Mono v2.304** release only included TTF variable fonts (no woff2 variable). The `JetBrainsMono[wght].ttf` from `fonts/variable/` was converted to woff2 using `fonttools` (`font.flavor = 'woff2'`). Output verified at 113KB.
- **`theme()` function** in `_layout.css` — Tailwind v4 does not support `theme(--spacing-6)` syntax. Raw CSS values used instead (`1.5rem`, `1rem`, `640px`, `1536px`) — these are the equivalent Tailwind default scale values.
- **Basecoat dark mode** — Basecoat registers `@custom-variant dark (&:is(.dark *))` class-based dark mode. Our `_tokens.css` dark overrides use `.dark { ... }` to match this pattern.

### References

- [Source: architecture.md §D14] — AG Grid theming via `_ag-grid.css`
- [Source: architecture.md §D15] — Vendor locally, no CDN
- [Source: architecture.md §D16] — Manual Tailwind compilation, committed dist
- [Source: ux-design-specification.md §Design System Foundation] — Basecoat adoption rationale and carve-outs
- [Source: ux-design-specification.md §Visual Design Foundation] — Color system, semantic tokens, typography
- [Source: epics.md §UX Design Requirements] — UX-DR1 through UX-DR8
- [Source: epics.md §Story 1.4] — Acceptance criteria
- [Source: fieldmark_shared/CLAUDE.md] — Current directory layout, build commands, symlink paths

## Dev Agent Record

### Agent Model Used

claude-sonnet-4-6

### Debug Log

- `theme(--spacing-6)` syntax unsupported in Tailwind v4 `theme()` function — replaced with raw values (`1.5rem`, `1rem`) in `_layout.css`.
- JetBrains Mono v2.304 release has no woff2 variable font — TTF converted to woff2 via `fonttools` with `font.flavor = 'woff2'`.

### Completion Notes

- `basecoat-css@0.3.11` installed as exact-pinned dependency; documented in architecture.md D15.
- Four new CSS partials created: `_fonts.css`, `_tokens.css`, `_layout.css`, `_ag-grid.css`.
- `fieldmark.css` updated with Basecoat import, all partial imports, and `@theme` font family overrides.
- Inter v4.1 (`InterVariable.woff2`) and JetBrains Mono v2.304 (`JetBrainsMono[wght].woff2`) self-hosted in `vendor/fonts/`.
- `pnpm run build` succeeds; compiled CSS is 134,544 bytes with Basecoat components, semantic tokens, status badges, score bands, font-face declarations, and AG Grid overrides all verified in output.
- All three stacks' symlinks still resolve to the updated `dist/fieldmark.css`.
- `fieldmark_shared/CLAUDE.md` updated with new directory layout, CSS partial documentation, and pinned dependency table.

### File List

- `fieldmark_shared/package.json` — updated (added basecoat-css@0.3.11; two-step build script)
- `fieldmark_shared/pnpm-lock.yaml` — updated (regenerated by pnpm install)
- `fieldmark_shared/src/fieldmark.css` — updated (Basecoat import, partial imports, @theme)
- `fieldmark_shared/src/_fonts.css` — new
- `fieldmark_shared/src/_tokens.css` — new
- `fieldmark_shared/src/_layout.css` — new
- `fieldmark_shared/src/_ag-grid.css` — new
- `fieldmark_shared/scripts/optimize-css.mjs` — new (LightningCSS post-build deduplication; --minify flag; dynamic pnpm store resolution; error handling)
- `fieldmark_shared/vendor/fonts/inter/InterVariable.woff2` — new (Inter v4.1, 352KB)
- `fieldmark_shared/vendor/fonts/jetbrains-mono/JetBrainsMono[wght].woff2` — new (JBM v2.304, 113KB)
- `fieldmark_shared/dist/fieldmark.css` — updated (recompiled + optimized, 131KB)
- `_bmad-output/planning-artifacts/architecture.md` — updated (Basecoat version pin in D15)
- `fieldmark_shared/CLAUDE.md` — updated (new CSS partials, font vendoring, dependency table, build scripts)

### Change Log

- 2026-05-17: Bootstrapped design system foundation — installed Basecoat CSS, self-hosted Inter and JetBrains Mono fonts, created four CSS partials (tokens, fonts, layout, ag-grid), updated fieldmark.css entry point, recompiled dist/fieldmark.css. All three stacks' symlinks verified intact.
- 2026-05-18: Addressed code review patch findings — added LightningCSS post-build optimization step (`scripts/optimize-css.mjs`) to merge duplicate selectors emitted by Tailwind v4/Basecoat compilation; `&:disabled` duplicates reduced 31→16, bundle bytes reduced 134,588→130,991. Updated `package.json` `build` script to two-step pipeline (tailwindcss + optimize-css.mjs).
- 2026-05-18: Addressed rerun review patch findings — (1) `scripts/optimize-css.mjs` confirmed present and wired; (2) added consecutive `content: var(--tw-content)` deduplication pass to `optimize-css.mjs`, reducing occurrences 43→34; (3) `!important` in sidebar icon collapse is intentional Basecoat 0.3.11 behavior (`padding:0 !important` for `data-collapsible="icon"`) — cannot change without forking; documented; (4) combined `.badge-violation-open-low, .badge-violation-inprogress` into selector list in `_tokens.css`. Bundle rebuilt: 130,991→130,669 B.
- 2026-05-18: Addressed chunked review (3rd round) findings — Decisions: (1) font files in git mandated by story hard rules + UX-DR6 self-hosted requirement; (2) CSS bundle size expected — Tailwind v4 does purge via @source, Basecoat full component CSS is baseline, ~131KB/~30KB gzipped; (3) relative paths are intentional monorepo design per Architecture D15. Patches: (1) replaced hardcoded `lightningcss@1.32.0` pnpm store path with dynamic `readdirSync` scan that finds any installed version; (2) added try/catch around readFileSync, transform, writeFileSync with process.exit(1) on error; (3) added `--minify` flag support to `optimize-css.mjs` and wired `build:prod` to use it — prod now runs LightningCSS selector merge on minified output (149→147KB); (4) broadened regex to `content:[ \t]*var\(--tw-content\);[ \t]*\r?\n` to tolerate whitespace variations and CRLF. Updated `fieldmark_shared/CLAUDE.md` to document two-step pipeline for both build targets.
- 2026-05-18: Addressed 4th review round (Rerun - Commit 4b65026) — (1) Replaced `readdirSync` pnpm store scan with resolution via `@tailwindcss/cli/package.json` dependency chain — version-independent, no file-system traversal; (2) added `result.warnings` surfacing to stderr so LightningCSS parse warnings are visible in CI; (3) remaining `&:focus-visible`/`&[aria-invalid]` "duplicate" selector blocks are NOT mergeable — they contain `@supports (color: color-mix(...))` conditionals that LightningCSS cannot merge without minification — documented in script comment; (4) regex already handles `--minify` skip + CRLF from prior round; (5) added `mkdirSync` call to ensure output directory exists before write; (6) cross-platform: `readdirSync` removed, CRLF handled, build/watch race is file-system-inherent and cannot be locked at this layer.
- 2026-05-18: Addressed 5th review round (Rerun - Working Tree) — Root-cause fix: added `lightningcss@1.32.0` as explicit devDependency so direct `req('lightningcss')` always resolves; entire fallback chain eliminated. (1) No more pnpm store resolution fragility; (2) normalized line endings to LF before regex (`/\r\n/g → \n`) — handles mixed CRLF/LF input; simplified regex to `\n` only after normalization; (3) atomic write: write to `output.tmp` then `renameSync` to final path — POSIX-atomic, prevents concurrent-reader corruption; (4) all `process.exit(1)` calls already have distinct `stderr` messages describing the failure mode — added `\n` suffix to all stderr writes for CI log readability; (5) added output path containment guard: `normalize(output).startsWith(normalize(root))` — rejects traversal attempts. `fieldmark_shared/package.json` updated; `pnpm install` run.
- 2026-05-18: Addressed 6th review round (Rerun - Latest Commit/Working Tree) — (1) Added `.tmp` cleanup via `unlinkSync(tmpOutput)` in catch block — temp file never lingers on failure; (2) Regex minified/CRLF finding: already handled — `!shouldMinify` guard + LF normalization; marked stale; (3) Added `realpathSync` check on output directory after `mkdirSync` to detect symlinks resolving outside root; (4) Added Windows EPERM/EACCES fallback: on rename failure with those codes, remove `.tmp` and fall back to direct `writeFileSync`; (5) Added zero-byte input guard — exits with clear message if Tailwind produced an empty file; (6) Warnings-silently-ignored finding: stale — lines 100–104 already write all LightningCSS warnings to stderr; marked resolved.
- 2026-05-18: Addressed 7th review round — (1) Symlink TOCTOU: unfixable at Node.js userspace (requires kernel `O_NOFOLLOW` not exposed in fs API); documented in comment; normalize guard still active as backstop; (2) Added `statSync` + `isFile()` check before `readFileSync` for cleaner error on directory/device input; (3) Changed silent `catch {}` on `realpathSync` to `catch (err) { stderr.write(...) }` — failure reason now visible in CI logs; (4) Wrapped Windows fallback `writeFileSync(output, css)` in its own try/catch with `process.exit(1)` — partial-write failures on EACCES now surface properly.

### Review Findings

**Code review complete.** 0 `decision-needed`, 3 `patch`, 10 `defer`, 0 dismissed as noise.
Findings written to the review findings section in 1-4-bootstrap-design-system-foundation-in-fieldmark-shared.md.

#### decision-needed
(none)

#### patch
- [x] [Review][Patch] Duplicate disabled state rules bloat output [fieldmark_shared/dist/fieldmark.css:1400-1500]
- [x] [Review][Patch] Massive CSS bundle growth from Basecoat integration [fieldmark_shared/dist/fieldmark.css +3496 lines]
- [x] [Review][Patch] Repeated SVG and hover rule duplication [fieldmark_shared/dist/fieldmark.css]

#### defer
- [x] [Review][Defer] Reduced-motion users see abrupt sidebar/toast/tooltip transitions — deferred, pre-existing
- [x] [Review][Defer] Unknown badge-* or data-score-band values render with default/neutral color only — deferred, pre-existing
- [x] [Review][Defer] Font files 404 or blocked → only system fallback, no visual regression test — deferred, pre-existing
- [x] [Review][Defer] Sidebar remains hidden or jumps if `[data-sidebar-initialized]` never set — deferred, pre-existing
- [x] [Review][Defer] AG Grid empty/loading state not styled distinctly from Basecoat table — deferred, pre-existing
- [x] [Review][Defer] Toaster accumulates unlimited toasts without height/scroll limit — deferred, pre-existing
- [x] [Review][Defer] Tooltip `data-tooltip` with HTML entities or >container-xs text clips silently — deferred, pre-existing
- [x] [Review][Defer] @source globs silently fail if any stack directory is renamed/moved — deferred, pre-existing
- [x] [Review][Defer] Basecoat 0.3.11 (pre-1.0) class names may shift on minor upgrade — deferred, pre-existing
- [x] [Review][Defer] High-contrast / forced-colors mode loses badge/score meaning (color-only) — deferred, pre-existing

### Review Findings (Rerun - 2026-05-17)

**Code review complete (rerun).** 0 `decision-needed`, 4 `patch`, 7 `defer`, 0 dismissed.
Findings appended for latest changes (optimize-css.mjs addition + CSS refinements).

#### decision-needed
(none)

#### patch
- [x] [Review][Patch] Build script references missing optimize-css.mjs [fieldmark_shared/package.json:7]
- [x] [Review][Patch] Repeated `content: var(--tw-content)` declarations (dozens) [fieldmark_shared/dist/fieldmark.css]
- [x] [Review][Patch] `!important` in sidebar icon collapse rule [fieldmark_shared/dist/fieldmark.css]
- [x] [Review][Patch] Duplicate `.badge-violation-open-low, .badge-violation-inprogress` combined rule [fieldmark_shared/dist/fieldmark.css]

#### defer
- [x] [Review][Defer] optimize-css.mjs crashes on LightningCSS version bump or pnpm store change — deferred, post-processing
- [x] [Review][Defer] Script has zero error handling for missing input file or permission issues — deferred, post-processing
- [x] [Review][Defer] In-place mutation without backup or dry-run mode — deferred, post-processing
- [x] [Review][Defer] No logging or propagation of LightningCSS recovered errors — deferred, post-processing
- [x] [Review][Defer] Script assumes pnpm + ESM environment; breaks under npm/yarn or CommonJS — deferred, post-processing
- [x] [Review][Defer] Optimization step not mentioned in build docs or CLAUDE.md — deferred, post-processing
- [x] [Review][Defer] Future Basecoat minor release may re-introduce unmergeable duplicates — deferred, post-processing

### Review Findings (Rerun - Commit 4b65026, 2026-05-18)

**Code review complete.** 0 `decision-needed`, 6 `patch`, 4 `defer`, 3 dismissed.
Blind Hunter + Edge Case Hunter on latest fixes; Acceptance Auditor: clean (all prior action items implemented correctly).

#### decision-needed
(none)

#### patch
- [x] [Review][Patch] optimize-css.mjs dynamic `readdirSync` + `createRequire` on pnpm store path (supply-chain risk) [fieldmark_shared/scripts/optimize-css.mjs:8-42]
- [x] [Review][Patch] `errorRecovery: true` silently swallows LightningCSS errors; broken CSS can ship [fieldmark_shared/scripts/optimize-css.mjs:94]
- [x] [Review][Patch] Remaining duplicate selector rules after dedup pass (focus-visible, aria-invalid, :is(.dark *)) [fieldmark_shared/dist/fieldmark.css]
- [x] [Review][Patch] Regex dedup fragile on minified output or future LightningCSS whitespace changes [fieldmark_shared/scripts/optimize-css.mjs:58]
- [x] [Review][Patch] No pre/post validation that vendor fonts exist or dist/ is writable [fieldmark_shared/scripts/optimize-css.mjs + package.json]
- [x] [Review][Patch] Cross-platform issues: pnpm store scan, line endings, concurrent build/watch race on dist/ [optimize-css.mjs + package.json]

#### defer
- [x] [Review][Defer] Unpinned major `basecoat-css@0.3.11` (pre-1.0) transitive risk — deferred, dependency policy
- [x] [Review][Defer] `build:raw` script undocumented / dead code — deferred, tooling hygiene
- [x] [Review][Defer] Duplicate token definitions in _tokens.css / _ag-grid.css vs Basecoat — deferred, long-term maintenance
- [x] [Review][Defer] @source globs silently drop classes if new template dir added — deferred, build robustness

#### dismissed
- Inline oklch tokens and @property formatting inconsistencies — acceptable generated output variance.
- build:raw added but unused — minor documentation gap, not a defect in this story.

### Review Findings (Chunked fieldmark_shared - 2026-05-18)

**Code review complete.** 3 `decision-needed`, 4 `patch`, 2 `defer`, 2 dismissed as noise.
Findings from Blind Hunter + Edge Case Hunter layers on fieldmark_shared/ design-system bootstrap changes only (out-of-scope items excluded).

#### decision-needed
- [x] [Review][Decision] Commit 350KB+ binary variable fonts (Inter/JetBrainsMono woff2) directly into git — bloats clones, history, and every PR diff touching CSS.
- [x] [Review][Decision] `dist/fieldmark.css` ~4k+ lines with no tree-shaking/purging — risks massive payload across all three stacks.
- [x] [Review][Decision] Relative paths in `_fonts.css` (`../vendor/fonts/...`) and `fieldmark.css` `@source` globs assume fixed monorepo layout and symlink depths.

#### patch
- [x] [Review][Patch] Hardcoded pnpm store path `node_modules/.pnpm/lightningcss@1.32.0/...` in optimize-css.mjs:37 — breaks on pnpm update, clean install, or different OS.
- [x] [Review][Patch] No try/catch or validation around LightningCSS transform + write in optimize-css.mjs — silent corruption possible.
- [x] [Review][Patch] Optimize step (`&& node scripts/optimize-css.mjs`) present only in default `build`; missing from `watch`, `build:raw`, `build:prod` — un-deduplicated CSS ships in those flows.

### Review Findings (Rerun - Working Tree, 2026-05-18)

**Code review complete.** 0 `decision-needed`, 5 `patch`, 3 `defer`, 3 dismissed.
Blind + Edge layers on latest optimize-css.mjs + related fixes; Acceptance Auditor: clean.

#### decision-needed
(none)

#### patch
- [x] [Review][Patch] `require.resolve` + `createRequire` chain for lightningcss is fragile across pnpm layouts [fieldmark_shared/scripts/optimize-css.mjs:40-50]
- [x] [Review][Patch] Regex dedup sensitive to mixed LF/CRLF line endings and whitespace [fieldmark_shared/scripts/optimize-css.mjs:95]
- [x] [Review][Patch] No file locking / atomic write — concurrent builds can corrupt dist/fieldmark.css [optimize-css.mjs]
- [x] [Review][Patch] `process.exit(1)` on all errors loses failure-mode detail for CI/Makefile [optimize-css.mjs]
- [x] [Review][Patch] No guard against absolute/traversal output paths or malicious args [optimize-css.mjs]

#### defer
- [x] [Review][Defer] `@import "basecoat-css"` with only package.json pin — deferred, dependency policy
- [x] [Review][Defer] No unit tests, input sanitization, or dry-run for optimize script — deferred, test coverage

### Review Findings (Rerun - Latest Commit / Working Tree, 2026-05-18)

**Code review complete.** 0 `decision-needed`, 6 `patch`, 4 `defer`, 3 dismissed.
Blind Hunter + Edge Case Hunter on current state; Acceptance Auditor: clean.

#### decision-needed
(none)

#### patch
- [x] [Review][Patch] Atomic renameSync lacks try/finally cleanup of .tmp file on failure [optimize-css.mjs:137]
- [x] [Review][Patch] Regex fails on minified single-line output or CRLF remnants [optimize-css.mjs:64-68, 119]
- [x] [Review][Patch] Path guard (normalize + startsWith) defeated by symlinks [optimize-css.mjs:40-42]
- [x] [Review][Patch] Windows renameSync race with antivirus/indexers (EPERM) [optimize-css.mjs:126-129]
- [x] [Review][Patch] No handling for empty/zero-byte input CSS [optimize-css.mjs:79-84]
- [x] [Review][Patch] Warnings + errorRecovery:true silently ignored; malformed CSS can ship [optimize-css.mjs:100-104]

#### defer
- [x] [Review][Defer] basecoat-css@0.3.11 unpinned beyond exact in package.json — deferred, dependency policy
- [x] [Review][Defer] @source globs silently break on template rename or new stack dir — deferred, build robustness
- [x] [Review][Defer] Duplicate oklch literals across _tokens/_layout/_ag-grid/_fonts.css — deferred, token centralization
- [x] [Review][Defer] Binary fonts without LFS or .gitattributes — deferred, repo hygiene

#### dismissed
- Build latency (~50-100 ms) and shebang-less .mjs — acceptable for current scope.
- Assumes content: var(--tw-content) only in pseudo-elements — low risk in current Tailwind + Basecoat usage.
- [x] [Review][Defer] Custom properties + badge classes lack @supports / contrast verification — deferred, accessibility polish

#### dismissed

### Review Findings (Rerun - 2026-05-18)

**Code review complete.** 0 `decision-needed`, 4 `patch`, 2 `defer`, 5 dismissed.
Blind Hunter + Edge Case Hunter; Acceptance Auditor: clean (all ACs satisfied, zero defects).

#### decision-needed
(none)

#### patch
- [x] [Review][Patch] Symlink TOCTOU between mkdir and realpath [optimize-css.mjs]
- [x] [Review][Patch] No fs.stat / isFile check before readFileSync [optimize-css.mjs]
- [x] [Review][Patch] Silent suppression of realpathSync errors [optimize-css.mjs]
- [x] [Review][Patch] Atomic write fallback on Windows leaves partial file on EACCES [optimize-css.mjs]

#### defer
- [x] [Review][Defer] Regex dedup assumes exact whitespace + LF after normalization [optimize-css.mjs]
- [x] [Review][Defer] No encoding/BOM handling on CSS input [optimize-css.mjs]

#### dismissed
- Documentation-only changes and snapshot build output claims — low risk.
- Pinned dep table without Renovate policy — tooling hygiene, not code defect.
- Hardcoded 'fieldmark.css' filename in transform — minor debug impact.
- LightningCSS warnings not promoted to failure — acceptable for current scope.
- Build time overhead (100-200 ms) and lack of SRI on fonts — acceptable for foundation story.
- Potential visual regression from Tailwind syntax changes — covered by existing e2e/visual testing plans.
- [x] [Review][Patch] Regex dedup `/([ \t]*content: var\(--tw-content\);\n)\1+/g` assumes exact post-LightningCSS whitespace — fragile to future Tailwind/LightningCSS changes.

#### defer
- [x] [Review][Defer] Variable font `font-weight: 100 900` syntax unsupported in Safari <16 / some Android WebViews — deferred, pre-existing browser quirk.
- [x] [Review][Defer] No CI gate enforcing `dist/fieldmark.css` matches source after build — deferred, infrastructure.

#### dismissed
- Underscore-prefixed partial naming (`_tokens.css` etc.) — project convention choice, not a defect.
- Minor generated CSS formatting inconsistencies (spacing, quotes) — acceptable output variance from two-stage build.
