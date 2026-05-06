# CLAUDE.md — Shared CSS (fieldmark_style)

This file provides guidance to Claude Code (claude.ai/code) when working in the `fieldmark_style/` CSS build project. Read alongside the root `CLAUDE.md`.

## Commands

```bash
npm install            # first-time setup
npm run build          # compile once (development)
npm run build:prod     # compile and minify (production)
npm run watch          # watch mode — run alongside app dev servers
```

## How It Works

`src/fieldmark.css` is the sole input file. It uses Tailwind v4 (`@import "tailwindcss"`) and `@source` directives to scan both app template directories:

```
@source "../../FieldMark/FieldMark.Web/Pages/**/*.cshtml"
@source "../../fieldmark_py/templates/**/*.html"
```

The compiled output at `dist/fieldmark.css` is the file both apps consume. It is symlinked — not copied — into each app's static directory:

- `FieldMark/FieldMark.Web/wwwroot/css/fieldmark.css`
- `fieldmark_py/static/vendor/fieldmark.css`

Commit `dist/fieldmark.css`. The symlinks depend on it existing; without it, fresh checkouts have broken static file references before the build runs.

## Watch Mode with App Dev Servers

```bash
# Terminal 1 — CSS watcher
cd fieldmark_style && npm run watch

# Terminal 2 — .NET
cd FieldMark && dotnet watch run --project FieldMark.Web

# Terminal 3 — Django
cd fieldmark_py && uv run python manage.py runserver
```

## Rules

- `src/fieldmark.css` is the only file that imports Tailwind. Do not add Tailwind imports or directives anywhere inside `FieldMark/` or `fieldmark_py/`.
- Neither app has a dependency on Node or Tailwind. They consume `dist/fieldmark.css` as a plain static file.
- Do not add per-app stylesheets that reintroduce a CSS framework. All shared styles belong here.
- `node_modules/` is gitignored. `dist/fieldmark.css` is not — commit it.
