# Story 1.4: Bootstrap Design System Foundation in `fieldmark_shared/`

Status: ready-for-dev

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

- [ ] Task 1: Add Basecoat dependency and pin versions (AC: #1)
  - [ ] 1.1: Install `basecoat-css` (npm package name) at version `0.3.11` with exact version pin in `fieldmark_shared/package.json`
  - [ ] 1.2: Pin `tailwindcss` and `@tailwindcss/cli` to exact versions (remove any `^` or `~`)
  - [ ] 1.3: Run `pnpm install` to update lockfile
  - [ ] 1.4: Document Basecoat version pin in `architecture.md` alongside HTMX 4.0.0-beta2 and AG Grid 35.2.1

- [ ] Task 2: Self-host Inter and JetBrains Mono fonts (AC: #4)
  - [ ] 2.1: Download Inter variable woff2 into `fieldmark_shared/vendor/fonts/inter/`
  - [ ] 2.2: Download JetBrains Mono variable woff2 into `fieldmark_shared/vendor/fonts/jetbrains-mono/`
  - [ ] 2.3: Create `fieldmark_shared/src/_fonts.css` with `@font-face` declarations

- [ ] Task 3: Restructure `fieldmark.css` with Basecoat + semantic tokens (AC: #2, #3)
  - [ ] 3.1: Import Basecoat CSS into `fieldmark.css`
  - [ ] 3.2: Import `_fonts.css`
  - [ ] 3.3: Define the five semantic color tokens with light and dark variants
  - [ ] 3.4: Add contrast ratio comments for each token
  - [ ] 3.5: Define status-badge class-to-token mappings for all entity states
  - [ ] 3.6: Define compliance-score threshold mapping keyed on `data-score-band`

- [ ] Task 4: Create AG Grid theme overrides (AC: #2)
  - [ ] 4.1: Create `fieldmark_shared/src/_ag-grid.css` with Quartz theme overrides aligning AG Grid with Basecoat palette

- [ ] Task 5: Create layout foundation (AC: #5)
  - [ ] 5.1: Create `fieldmark_shared/src/_layout.css` with container, gutter collapse, and spacing documentation

- [ ] Task 6: Add `.tnum` utility class (AC: #4)
  - [ ] 6.1: Define `.tnum` with `font-feature-settings: "tnum"` in the CSS source

- [ ] Task 7: Compile and verify (AC: #7)
  - [ ] 7.1: Run `pnpm run build` — verify `dist/fieldmark.css` is produced
  - [ ] 7.2: Verify all symlinks still resolve correctly in all three stacks
  - [ ] 7.3: Commit `dist/fieldmark.css`

- [ ] Task 8: Verify existing vendor JS and symlinks (AC: #6)
  - [ ] 8.1: Confirm `vendor/htmx/htmx.min.js` and `vendor/ag-grid/35.2.1/ag-grid-community.min.js` exist
  - [ ] 8.2: Confirm all three stacks' symlinks resolve

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

claude-opus-4-6

### Completion Notes

- Ultimate context engine analysis completed — comprehensive developer guide created.
- `fieldmark_shared/` has a working Tailwind v4 pipeline but NO Basecoat yet — this story adds it.
- Existing symlinks in all three stacks are working and must not be broken.
- Package manager is `pnpm` (not npm or yarn) — use `pnpm install` and `pnpm run build`.
- Five new source files, two font files, one updated package.json, one updated architecture.md.
- No template changes in any stack — CSS infrastructure only.
- `dist/fieldmark.css` and vendor font files must be committed.

### File List

- `fieldmark_shared/package.json` — update (add basecoat, pin versions)
- `fieldmark_shared/pnpm-lock.yaml` — update (regenerated)
- `fieldmark_shared/src/fieldmark.css` — update (add imports, @theme)
- `fieldmark_shared/src/_fonts.css` — new
- `fieldmark_shared/src/_tokens.css` — new
- `fieldmark_shared/src/_layout.css` — new
- `fieldmark_shared/src/_ag-grid.css` — new
- `fieldmark_shared/vendor/fonts/inter/InterVariable.woff2` — new
- `fieldmark_shared/vendor/fonts/jetbrains-mono/JetBrainsMono[wght].woff2` — new
- `fieldmark_shared/dist/fieldmark.css` — update (recompiled)
- `_bmad-output/planning-artifacts/architecture.md` — update (Basecoat version pin)
- `fieldmark_shared/CLAUDE.md` — update (new CSS partials, fonts)
