# fieldmark_shared

**Shared front-end assets for all three FieldMark stacks** — the single source of truth for compiled CSS and vendored JavaScript. No stack has its own copy of these files.

The three consuming stacks are [`FieldMark/`](../FieldMark/README.md) (.NET Razor Pages), [`fieldmark_py/`](../fieldmark_py/README.md) (Django Templates), and [`fieldmark-go/`](../fieldmark-go/README.md) (Go/Fiber). All three consume `dist/fieldmark.css` and the contents of `vendor/` via symlinks — never by copying files.

---

## What Lives Here

| Path | Purpose |
|---|---|
| `src/fieldmark.css` | Tailwind v4 source — the only file that imports Tailwind |
| `dist/fieldmark.css` | Compiled output — **committed to the repo** |
| `vendor/htmx/` | Vendored HTMX (pinned version) |
| `vendor/ag-grid/` | Vendored AG Grid Enterprise (pinned version; no license key — watermark accepted) |

---

## Directory Layout

```
fieldmark_shared/
├── src/
│   └── fieldmark.css        Tailwind v4 source (sole CSS input)
├── dist/
│   └── fieldmark.css        Compiled output — commit this file
├── vendor/
│   ├── ag-grid/
│   │   └── 35.3.0/
│   │       └── ag-grid-enterprise.min.js
│   └── htmx/
│       └── htmx.min.js
├── package.json
└── node_modules/            gitignored
```

---

## Tech Stack

| Tool | Purpose | Version |
|---|---|---|
| Tailwind CSS | Utility-first CSS framework | 4.x (pinned) |
| Basecoat | Component vocabulary layered on Tailwind | pre-1.0 patch (pinned) |
| HTMX | Hypermedia interactivity | 4.x (pinned) |
| AG Grid Enterprise | Server-side row model data grids (true SSRM; no license key — watermark accepted) | 35.3.0 (pinned) |
| pnpm | Package manager | workspace root |

Version pins are exact (no `^` or `~` ranges). Any upgrade is a coordinated three-stack story — bump the pin here, re-baseline visual regression snapshots, verify all three stacks.

---

## Prerequisites

- [Node.js 20+](https://nodejs.org/) (for the Tailwind build only — Tailwind's Oxide engine requires Node ≥ 20; no stack has a runtime Node dependency)
- [pnpm](https://pnpm.io/) (`npm install -g pnpm`)

No stack requires Node to run. They consume `dist/fieldmark.css` and the vendor files as plain static assets.

---

## Getting Started

**Install dependencies (first-time only):**

```bash
cd fieldmark_shared
pnpm install
```

**Compile CSS once:**

```bash
pnpm run build
```

**Watch mode (run alongside app dev servers):**

```bash
pnpm run watch
```

`dist/fieldmark.css` is committed to the repo so a fresh clone does not require a build step before any dev server starts.

---

## How the CSS Build Works

`src/fieldmark.css` uses `@source` directives to scan all three app template directories for Tailwind class usage:

```css
@source "../../FieldMark/FieldMark.Web/Pages/**/*.cshtml"
@source "../../fieldmark_py/templates/**/*.html"
@source "../../fieldmark-go/internal/web/templates/**/*.html"
```

The compiled `dist/fieldmark.css` is symlinked — never copied — into each stack's static tree:

| Stack | Symlink |
|---|---|
| .NET | `FieldMark/FieldMark.Web/wwwroot/vendor/fieldmark.css` |
| Django | `fieldmark_py/static/vendor/fieldmark.css` |
| Go/Fiber | `fieldmark-go/internal/web/static/vendor/fieldmark.css` |

---

## How the Vendor JS Symlinks Work

Each stack symlinks the `vendor/ag-grid/` and `vendor/htmx/` **directories** (not individual files) into its own static tree:

| Stack | ag-grid symlink | htmx symlink |
|---|---|---|
| .NET | `wwwroot/vendor/ag-grid` | `wwwroot/vendor/htmx` |
| Django | `static/vendor/ag-grid` | `static/vendor/htmx` |
| Go/Fiber | `internal/web/static/vendor/ag-grid` | `internal/web/static/vendor/htmx` |

All paths are relative so the repo works regardless of where it is cloned. No CDN dependencies — vendor assets are local.

---

## Development Workflow

Run the CSS watcher alongside whichever app dev servers you need:

```bash
# Terminal 1 — CSS watcher
cd fieldmark_shared && pnpm run watch

# Terminal 2 — .NET
cd FieldMark && dotnet watch run --project FieldMark.Web

# Terminal 3 — Django
cd fieldmark_py && uv run python manage.py runserver

# Terminal 4 — Go/Fiber
cd fieldmark-go && go run ./cmd/web
```

---

## Rules

- `src/fieldmark.css` is the **only** file that imports Tailwind. Do not add Tailwind imports anywhere inside `FieldMark/`, `fieldmark_py/`, or `fieldmark-go/`.
- No stack has a Node or npm dependency. They consume `dist/fieldmark.css` and vendor JS as plain static files via symlinks.
- Do not add per-stack stylesheets that reintroduce a CSS framework. All shared styles belong here.
- `dist/fieldmark.css` and everything in `vendor/` are **committed** — do not gitignore them.
- To add a new shared JS library: add it to `vendor/`, create directory symlinks in all three stacks, and update this README and the root `CLAUDE.md`.

---

## Related Documentation

- [Root README](../README.md) — project overview and monorepo structure
- [.NET README](../FieldMark/README.md) — .NET Razor Pages stack
- [Django README](../fieldmark_py/README.md) — Django Templates stack
- [Go README](../fieldmark-go/README.md) — Go/Fiber stack
- [Architecture](../_bmad-output/planning-artifacts/architecture.md) — cross-stack architectural decisions
