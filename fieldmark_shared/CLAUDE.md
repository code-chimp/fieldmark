# CLAUDE.md — Shared Front-End Assets (fieldmark_shared)

This file provides guidance to Claude Code (claude.ai/code) when working in the `fieldmark_shared/` project. Read alongside the root `CLAUDE.md`.

## Purpose

`fieldmark_shared` is the single source of truth for all shared front-end assets in the FieldMark monorepo:

- **CSS** — Tailwind v4 + Basecoat source compiled to `dist/fieldmark.css`
- **Vendor JS** — `htmx.min.js`, `ag-grid-community.min.js`, and `theme-toggle/theme-toggle.js` live in `vendor/` and are symlinked into each stack's static directory
- **Vendor fonts** — Inter and JetBrains Mono woff2 files live in `vendor/fonts/` and are referenced by `_fonts.css`

No stack has its own copy of these files. Adding a library here and symlinking it into the stacks is the only permitted way to introduce a shared JS dependency.

## Directory Layout

```
fieldmark_shared/
├── src/
│   ├── fieldmark.css       Main entry point — imports Tailwind, Basecoat, and all partials
│   ├── _fonts.css          @font-face declarations for Inter and JetBrains Mono
│   ├── _tokens.css         Semantic color tokens, status-badge vocabulary, score-band mapping
│   ├── _components.css     Shared UI component styles (ThemeToggle, etc.)
│   ├── _layout.css         Container, gutter collapse, body text-size defaults
│   └── _ag-grid.css        AG Grid Quartz theme overrides
├── dist/
│   └── fieldmark.css       Compiled output — commit this file
├── vendor/
│   ├── ag-grid/
│   │   └── 35.2.1/
│   │       └── ag-grid-community.min.js
│   ├── htmx/
│   │   └── htmx.min.js
│   ├── theme-toggle/
│   │   └── theme-toggle.js                    Client-side theme listener (≤20 LOC; story 1-6)
│   └── fonts/
│       ├── inter/
│       │   └── InterVariable.woff2            Inter v4.1 variable font
│       └── jetbrains-mono/
│           └── JetBrainsMono[wght].woff2      JetBrains Mono v2.304 variable font
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
| AG Grid Community | 35.2.1 | `vendor/ag-grid/35.2.1/` |
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

| Stack | ag-grid symlink | htmx symlink | theme-toggle symlink |
|---|---|---|---|
| .NET | `wwwroot/vendor/ag-grid` → `../../../../fieldmark_shared/vendor/ag-grid` | `wwwroot/vendor/htmx` → `../../../../fieldmark_shared/vendor/htmx` | `wwwroot/vendor/theme-toggle` → `../../../../fieldmark_shared/vendor/theme-toggle` |
| Django | `static/vendor/ag-grid` → `../../../fieldmark_shared/vendor/ag-grid` | `static/vendor/htmx` → `../../../fieldmark_shared/vendor/htmx` | `static/vendor/theme-toggle` → `../../../fieldmark_shared/vendor/theme-toggle` |
| Go/Fiber | `internal/web/static/vendor/ag-grid` → `../../../../../fieldmark_shared/vendor/ag-grid` | `internal/web/static/vendor/htmx` → `../../../../../fieldmark_shared/vendor/htmx` | `internal/web/static/vendor/theme-toggle` → `../../../../../fieldmark_shared/vendor/theme-toggle` |

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
- `node_modules/` is gitignored. `dist/fieldmark.css` and everything in `vendor/` are **not** — commit them.
- All npm dependencies must use exact version pins — no `^` or `~` ranges.
- CSS partials use underscore prefix (`_fonts.css`, `_tokens.css`, `_layout.css`, `_ag-grid.css`). They are imported into `fieldmark.css` only.
- When editing CSS, run `pnpm run build` and commit both the source changes and the updated `dist/fieldmark.css`.
