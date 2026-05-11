# CLAUDE.md — Shared Front-End Assets (fieldmark_shared)

This file provides guidance to Claude Code (claude.ai/code) when working in the `fieldmark_shared/` project. Read alongside the root `CLAUDE.md`.

## Purpose

`fieldmark_shared` is the single source of truth for all shared front-end assets in the FieldMark monorepo:

- **CSS** — Tailwind v4 source compiled to `dist/fieldmark.css`
- **Vendor JS** — `htmx.min.js` and `ag-grid-community.min.js` live in `vendor/` and are symlinked into each stack's static directory

No stack has its own copy of these files. Adding a library here and symlinking it into the stacks is the only permitted way to introduce a shared JS dependency.

## Directory Layout

```
fieldmark_shared/
├── src/
│   └── fieldmark.css        Tailwind v4 source (sole CSS input)
├── dist/
│   └── fieldmark.css        Compiled output — commit this file
├── vendor/
│   ├── ag-grid/
│   │   └── 35.2.1/
│   │       └── ag-grid-community.min.js
│   └── htmx/
│       └── htmx.min.js
├── package.json
└── node_modules/            gitignored
```

## Commands

```bash
npm install            # first-time setup
npm run build          # compile once (development)
npm run build:prod     # compile and minify (production)
npm run watch          # watch mode — run alongside app dev servers
```

## How the CSS Build Works

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

Each stack symlinks the `vendor/ag-grid` and `vendor/htmx` directories (not individual files) into its own static tree:

| Stack | ag-grid symlink | htmx symlink |
|---|---|---|
| .NET | `wwwroot/vendor/ag-grid` → `../../../../fieldmark_shared/vendor/ag-grid` | `wwwroot/vendor/htmx` → `../../../../fieldmark_shared/vendor/htmx` |
| Django | `static/vendor/ag-grid` → `../../../fieldmark_shared/vendor/ag-grid` | `static/vendor/htmx` → `../../../fieldmark_shared/vendor/htmx` |
| Go/Fiber | `internal/web/static/vendor/ag-grid` → `../../../../../fieldmark_shared/vendor/ag-grid` | `internal/web/static/vendor/htmx` → `../../../../../fieldmark_shared/vendor/htmx` |

All paths are relative so the repo works regardless of where it is cloned.

## Watch Mode with App Dev Servers

```bash
# Terminal 1 — CSS watcher
cd fieldmark_shared && npm run watch

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
