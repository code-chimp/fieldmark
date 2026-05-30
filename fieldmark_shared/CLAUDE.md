# CLAUDE.md — Shared Front-End Assets (fieldmark_shared)

This file provides guidance to Claude Code (claude.ai/code) when working in the `fieldmark_shared/` project. Read alongside the root `CLAUDE.md`.

## Purpose

`fieldmark_shared` is the single source of truth for all shared front-end assets in the FieldMark monorepo:

- **CSS** — Tailwind v4 + Basecoat source compiled to `dist/fieldmark.css`
- **Vendor JS** — `htmx.min.js`, `ag-grid-enterprise.min.js`, and `theme-toggle/theme-toggle.js` live in `vendor/` and are symlinked into each stack's static directory. AG Grid **Enterprise** is used to demonstrate the true Server-Side Row Model; the demo runs **without a license key** and the "unlicensed" watermark is an accepted, deliberate tradeoff.
- **Vendor fonts** — Inter and JetBrains Mono woff2 files live in `vendor/fonts/` and are referenced by `_fonts.css`

No stack has its own copy of these files. Adding a library here and symlinking it into the stacks is the only permitted way to introduce a shared JS dependency.

## Directory Layout

```
fieldmark_shared/
├── src/
│   ├── fieldmark.css       Main entry point — imports Tailwind, Basecoat, and all partials
│   ├── _fonts.css          @font-face declarations for Inter and JetBrains Mono
│   ├── _tokens.css         Semantic color tokens, status-badge vocabulary, score-band mapping
│   ├── _components.css     Shared UI component styles (ThemeToggle pill, brand mark, etc.)
│   ├── _layout.css         Container, gutter collapse, body text-size defaults
│   └── _ag-grid.css        AG Grid Quartz theme overrides
├── dist/
│   └── fieldmark.css       Compiled output — commit this file
├── vendor/
│   ├── ag-grid/
│   │   └── 35.3.0/
│   │       └── ag-grid-enterprise.min.js     Enterprise UMD bundle (includes Community); no license key — watermark accepted
│   ├── htmx/
│   │   └── htmx.min.js
│   ├── theme-toggle/
│   │   └── theme-toggle.js                    Client-side theme listener; syncs aria-pressed on 3-button pill
│   ├── tabstrip/
│   │   └── tabstrip.js                        Arrow-key navigation for <nav role="tablist" data-tabstrip>; ES5 IIFE; 23 LOC
│   ├── fonts/
│   │   ├── inter/
│   │   │   └── InterVariable.woff2            Inter v4.1 variable font
│   │   └── jetbrains-mono/
│   │       └── JetBrainsMono[wght].woff2      JetBrains Mono v2.304 variable font
│   └── img/                                   Brand images — favicon suite and OG image
│       ├── favicon.svg
│       ├── favicon-16.png, favicon-32.png
│       ├── apple-touch-icon.png
│       ├── icon-192.png, icon-512.png
│       └── fieldmark-og.png
├── package.json
└── node_modules/            gitignored
```

## CSS File Organization

| File | Purpose |
|---|---|
| `fieldmark.css` | Main entry point — imports Tailwind, Basecoat, and all partials |
| `_fonts.css` | `@font-face` declarations for Inter and JetBrains Mono |
| `_tokens.css` | Semantic color tokens, status-badge vocabulary, score-band mapping, `.tnum` utility |
| `_components.css` | Shared UI component styles (ThemeToggle icon visibility, future shared components) |
| `_layout.css` | Container (max-w-screen-2xl), gutter collapse (px-6→px-4 at 640px), body text-sm default |
| `_ag-grid.css` | AG Grid Quartz theme overrides aligned with Tailwind neutral palette |

The underscore-prefix convention signals partials imported by `fieldmark.css`.

## Pinned Dependencies

| Library | Version | Location |
|---|---|---|
| Tailwind CSS CLI | 4.2.4 (exact) | `devDependencies` |
| LightningCSS | 1.32.0 (exact) | `devDependencies` |
| Basecoat CSS | 0.3.11 (exact, pre-1.0) | `dependencies` |
| HTMX | 4.0.0-beta2 | `vendor/htmx/htmx.min.js` |
| AG Grid Enterprise | 35.3.0 | `vendor/ag-grid/35.3.0/` (Enterprise UMD bundle; no license key — "unlicensed" watermark is an accepted demo tradeoff for showing true SSRM) |
| Inter font | 4.1 | `vendor/fonts/inter/InterVariable.woff2` |
| JetBrains Mono | 2.304 | `vendor/fonts/jetbrains-mono/JetBrainsMono[wght].woff2` |

All npm dependencies use exact version pins — no `^` or `~` ranges.

## Commands

```bash
pnpm install           # first-time setup
pnpm run build         # compile + optimize once (development) — Tailwind then LightningCSS dedup pass
pnpm run build:raw     # Tailwind compile only (skips optimization; for debugging)
pnpm run build:prod    # compile, minify, and optimize (production)
pnpm run watch         # watch mode — run alongside app dev servers (no optimization pass)
```

## How the CSS Build Works

The default `pnpm run build` is a two-step pipeline:
1. **Tailwind** compiles `src/fieldmark.css` → `dist/fieldmark.css` (4,133 lines raw)
2. **`scripts/optimize-css.mjs`** runs LightningCSS over the output to merge duplicate selectors that Tailwind v4 emits when Basecoat uses multiple utilities on the same pseudo-selector (e.g. `&:disabled`, `&>svg`), then removes consecutive `content: var(--tw-content)` duplicates within merged pseudo-element blocks. Result: ~4,606 lines, ~131KB. The line count is slightly higher than the Tailwind output because LightningCSS normalizes some selector syntax, but bytes are lower due to deduplication.

`build:prod` runs the same two-step pipeline with `--minify` in both stages — Tailwind minifies, then LightningCSS merges duplicate selectors on the minified output (~147KB). The optimization step is intentionally absent from `watch` (latency) and `build:raw` (debug use).

`src/fieldmark.css` is the sole Tailwind input. It uses `@source` directives to scan all three app template directories:

```css
@source "../../FieldMark/FieldMark.Web/Pages/**/*.cshtml"
@source "../../fieldmark_py/templates/**/*.html"
@source "../../fieldmark-go/internal/web/templates/**/*.html"
```

The compiled output `dist/fieldmark.css` is symlinked — never copied — into each stack:

| Stack | Symlink location |
|---|---|
| .NET | `FieldMark/FieldMark.Web/wwwroot/vendor/fieldmark.css` |
| Django | `fieldmark_py/static/vendor/fieldmark.css` |
| Go/Fiber | `fieldmark-go/internal/web/static/vendor/fieldmark.css` |

Commit `dist/fieldmark.css`. Fresh checkouts need the compiled file to exist before any dev server starts.

## How the Vendor JS Symlinks Work

Each stack symlinks vendor directories (not individual files) into its own static tree:

| Stack | ag-grid symlink | htmx symlink | theme-toggle symlink | tabstrip symlink | img symlink |
|---|---|---|---|---|---|
| .NET | `wwwroot/vendor/ag-grid` → `../../../../fieldmark_shared/vendor/ag-grid` | `wwwroot/vendor/htmx` → `../../../../fieldmark_shared/vendor/htmx` | `wwwroot/vendor/theme-toggle` → `../../../../fieldmark_shared/vendor/theme-toggle` | `wwwroot/vendor/tabstrip` → `../../../../fieldmark_shared/vendor/tabstrip` | `wwwroot/vendor/img` → `../../../../fieldmark_shared/vendor/img` |
| Django | `static/vendor/ag-grid` → `../../../fieldmark_shared/vendor/ag-grid` | `static/vendor/htmx` → `../../../fieldmark_shared/vendor/htmx` | `static/vendor/theme-toggle` → `../../../fieldmark_shared/vendor/theme-toggle` | `static/vendor/tabstrip` → `../../../fieldmark_shared/vendor/tabstrip` | `static/vendor/img` → `../../../fieldmark_shared/vendor/img` |
| Go/Fiber | `internal/web/static/vendor/ag-grid` → `../../../../../fieldmark_shared/vendor/ag-grid` | `internal/web/static/vendor/htmx` → `../../../../../fieldmark_shared/vendor/htmx` | `internal/web/static/vendor/theme-toggle` → `../../../../../fieldmark_shared/vendor/theme-toggle` | `internal/web/static/vendor/tabstrip` → `../../../../../fieldmark_shared/vendor/tabstrip` | `internal/web/static/vendor/img` → `../../../../../fieldmark_shared/vendor/img` |

All paths are relative so the repo works regardless of where it is cloned.

## Watch Mode with App Dev Servers

```bash
# Terminal 1 — CSS watcher
cd fieldmark_shared && pnpm run watch

# Terminal 2 — .NET
cd FieldMark && dotnet watch run --project FieldMark.Web

# Terminal 3 — Django
cd fieldmark_py && uv run python manage.py runserver

# Terminal 4 — Go/Fiber
cd fieldmark-go && go run ./cmd/server
```

## Rules

- `src/fieldmark.css` is the only file that imports Tailwind. Do not add Tailwind imports anywhere inside `FieldMark/`, `fieldmark_py/`, or `fieldmark-go/`.
- No stack has a Node or npm dependency. They consume `dist/fieldmark.css` and the vendor JS as plain static files via symlinks.
- Do not add per-app stylesheets that reintroduce a CSS framework. All shared styles belong here.
- To add a new vendor JS library: drop it in `vendor/`, create symlinks in all three stacks, update this doc and the root `CLAUDE.md`.
- **HTMX 4.0 event names use colon-separated format, not camelCase.** Event listeners in vendor JS must use the new names: `htmx:after:swap` (not `htmx:afterSwap`), `htmx:before:swap` (not `htmx:beforeSwap`), etc. Events are dispatched on `document`, not `document.body` — use `document.addEventListener(...)`. This was discovered during Story 2.7's tabstrip.js OOB-swap re-attachment; the original code used the HTMX 1.x `htmx:afterSwap` on `document.body` which silently never fired under HTMX 4.0-beta2.
- `node_modules/` is gitignored. `dist/fieldmark.css` and everything in `vendor/` are **not** — commit them.
- All npm dependencies must use exact version pins — no `^` or `~` ranges.
- CSS partials use underscore prefix (`_fonts.css`, `_tokens.css`, `_layout.css`, `_ag-grid.css`). They are imported into `fieldmark.css` only.
- When editing CSS, run `pnpm run build` and commit both the source changes and the updated `dist/fieldmark.css`.
- **Accessibility media queries belong exclusively in `_a11y.css`.** Never add `@media (forced-colors: active)` or `@media (prefers-reduced-motion: reduce)` blocks to `_tokens.css`, `_components.css`, or any other partial. Splitting these blocks across files produces duplicate `@media` rules in `dist/fieldmark.css` and makes it impossible to audit accessibility coverage from a single file. (Story 2.4 round 3: a `forced-colors` block was erroneously placed in `_tokens.css`.)

## Component Canonical Examples — Maintenance Rules

When creating or modifying a component under `fieldmark_shared/components/<name>/`:

- **`canonical.html` and `README.md` travel together.** Whenever a new `<!-- variant: name -->` block is added to `canonical.html`, the corresponding entry in the README's "Variant List" section must be updated in the same commit. A `canonical.html` whose variant count differs from the README's Variant List is a documentation defect. (Story 2.4 round 4: `zero-value` variant was added to `dashboard_tile/canonical.html` but not to `README.md`.)
- When per-stack snapshot tests are updated to cover a new variant, the README must also reflect that variant before the story is marked complete.

## Build-Script Defensive Defaults

Lessons absorbed during Story 1.4's six review rounds on `scripts/optimize-css.mjs`. Any new script in `fieldmark_shared/scripts/` must follow these defaults — they apply to every build-tool addition, not just CSS.

- **Atomic writes.** Write to `<target>.tmp`, then `renameSync` to the final path. Wrap in `try/finally` so the `.tmp` is cleaned up on failure. Never mutate the target file in place.
- **No hardcoded package-manager paths.** Resolve dependencies via `require.resolve` / `createRequire` against `import.meta.url`, not by scanning `node_modules/.pnpm/...`. Hardcoded store paths break on pnpm upgrade, clean install, and other package managers.
- **Fatal warnings exit non-zero.** Tool wrappers (LightningCSS, esbuild, etc.) often have an `errorRecovery` mode that silently continues on bad input. Either disable it, or classify warning severities and `process.exit(1)` on the fatal types. Silent success on broken input is the worst failure mode.
- **Validate inputs before writing.** Missing input file, directory passed as file, empty input, non-writable output dir, absolute or `..` traversal paths — all exit non-zero with a clear message before touching the filesystem.
- **Engine guards.** If the script uses Node ≥ N features, both `engines.node` in `package.json` *and* a runtime check (`if (parseInt(process.versions.node) < N) process.exit(1)`) — `engines` is advisory only.
- **Package-manager guard.** `preinstall` script that enforces pnpm (or whatever the project uses); without it, npm/yarn users hit cryptic mid-pipeline failures.
- **Surface tool warnings to stderr.** Don't swallow LightningCSS / esbuild / Tailwind warnings — log them. The reviewer should see them; CI should see them.
- **No script silently outputs zero.** `@source` glob misses, empty input directories, missing dependencies → fail loud with an actionable error, not a zero-byte file.

## Component Examples

The `components/` directory holds canonical reference HTML fragments that serve as cross-stack snapshot-test targets. These are **not** live HTML pages — they contain only the component markup, without surrounding layout chrome.

| File | Purpose |
|---|---|
| `login-form.example.html` | Canonical login form (`<form>…</form>` block) that .NET and Django snapshot tests assert byte-identical output against |
| `login-error-region.example.html` | Canonical inline-alert block rendered above the form on HTTP 422 |

### Snapshot-test pipeline

Each stack normalises the rendered `GET /login` response body before comparing:
1. Extract the `<form id="login-form">…</form>` block (or the error-region block).
2. Strip per-stack antiforgery noise: `<input name="__RequestVerificationToken">` (.NET) and `<input name="csrfmiddlewaretoken">` (Django).
3. Normalise whitespace (collapse runs, trim lines) and sort attributes alphabetically.
4. Assert byte-equal against the corresponding example file.

The Go stack is **excluded** from the login-form snapshot assertion — its `/login` renders a user-switcher, not a credential form (see ADR-012).

### Cross-stack change rule

Any change to `login-form.example.html` or `login-error-region.example.html` **must** be applied simultaneously to the .NET Razor template (`Pages/Account/Login.cshtml`) and the Django template (`templates/_login.html`). The snapshot tests will fail if the stacks drift from the canonical reference. The Go user-switcher template is exempt.

## Sidebar progressive enhancement

The sidebar uses `[data-sidebar-initialized]` as a JS activation gate. **The default CSS must always render the sidebar visible and non-collapsible.** Only when `[data-sidebar-initialized]` is present on the element should collapse/slide behavior engage.

This means: if the sidebar JS fails (404, CSP block, or JS disabled), the sidebar degrades gracefully to a visible, static nav — not hidden, not jumping.

CSS rule (in `_components.css`):

```css
.sidebar:not([data-sidebar-initialized]) {
  display: block !important;
  position: static !important;
  transform: none !important;
}
.sidebar:not([data-sidebar-initialized]) nav {
  display: block !important;
}
```

The `!important` is intentional: it overrides Basecoat's component-level mobile `display: none` that fires before `[data-sidebar-initialized]` is set. Without it, the sidebar is invisible on first paint in mobile viewports.

**Testing:** `e2e/tests/shared/sidebar-no-js.spec.ts` verifies this with `javaScriptEnabled: false`.

## AG Grid empty / loading states

AG Grid Quartz theme overlay styles are in `src/_ag-grid.css`. They target the overlay center elements that AG Grid renders when the grid is loading or has no rows:

- `.ag-overlay-loading-center` — shows a spinner-adjacent tinted background and italic "Loading…" text, visually distinct from a Basecoat empty `<table>` row.
- `.ag-overlay-no-rows-center` — shows a muted neutral background with "No records found" text.

Both states have dark-mode overrides (`.dark .ag-theme-quartz ...`).

These rules are keyed on AG Grid's own class names — do not rename them. Epic 2's AG Grid feature stories will exercise these states in integration tests; the CSS is wired now so the visual contract is consistent from the first feature that uses a grid.
