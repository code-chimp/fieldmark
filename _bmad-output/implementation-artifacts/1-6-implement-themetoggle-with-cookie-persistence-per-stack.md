# Story 1.6: Implement ThemeToggle with cookie persistence per stack

Status: ready-for-dev

## Story

As any user on any stack,
I want a single header-strip control that cycles System → Light → Dark with my preference remembered across sessions,
So that the application matches my environment without flashing the wrong theme on first paint.

## Acceptance Criteria

1. **First-paint theme resolution (no cookie):** When no `fm_theme` cookie is present, the server emits `<html data-theme="system" ...>`. A 5-line inline `<script>` placed in `<head>` (before any external CSS/JS) reads `window.matchMedia('(prefers-color-scheme: dark)')` and replaces `data-theme` on `<html>` with `"light"` or `"dark"` before first paint (UX-DR5). This is the **only** inline JavaScript in the application; its presence is documented in `_bmad-output/planning-artifacts/architecture.md` under Frontend Architecture (D15 or new sub-decision).

2. **First-paint with persisted preference:** When `fm_theme` cookie is `"light"` or `"dark"`, the server emits `<html data-theme="<value>">` directly; the inline script becomes a no-op for non-`system` values. When `fm_theme="system"`, the same inline script resolves and replaces it. No flash of wrong theme is visible in any case.

3. **ThemeToggle component markup (UX-DR15):** Renders in the header strip beside the user avatar slot as a 36×36 icon button. `aria-label="Theme: <current>; activate to cycle"` where `<current>` is the resolved theme name (`system`, `light`, or `dark`). Displays a Lucide icon — Monitor (system), Sun (light), Moon (dark) — reflecting the *resolved* theme. Renders as `<button type="button" hx-post="/preferences/theme" hx-vals='{"value":"<next>"}' hx-swap="none">…</button>`.

4. **POST `/preferences/theme` endpoint (canonical, all three stacks):** Accepts form-encoded body with field `value` ∈ {`system`, `light`, `dark`}. Cycle order is `system → light → dark → system`. Server validates the value (rejects unknown with HTTP 400, no cookie change). On valid value, server sets `Set-Cookie: fm_theme=<value>; Path=/; SameSite=Lax; Max-Age=31536000` (no `HttpOnly` — the client listener may read it; no `Secure` because local-only MVP). Returns HTTP **204** with response header `HX-Trigger: theme-changed`. No response body. GET on this path is not registered (no fallback).

5. **Client-side listener (vendored, ≤ 20 LOC):** `fieldmark_shared/vendor/theme-toggle/theme-toggle.js` listens for the `theme-changed` event on `document.body` (HTMX dispatches custom events from `HX-Trigger` on the body). On fire, it (a) reads the cycled value from `document.cookie` (`fm_theme`), (b) resolves `system` via `matchMedia` if needed, (c) writes the resolved value to `<html>` `data-theme`, (d) updates the toggle button's `aria-label` and icon to reflect the new current + next state. Script is loaded once via `<script src="…/vendor/theme-toggle/theme-toggle.js"></script>` after HTMX in every stack's base layout. Total file ≤ 20 lines of executable code (excluding comments and the IIFE wrapper).

6. **Cross-stack byte-parity:** For identical inputs (no cookie, then `fm_theme=light`, then `fm_theme=dark`), the rendered `<html>` opening tag, the inline `<script>` block, and the ThemeToggle button HTML are **byte-identical** across the three stacks when captured via `curl /` and normalized through the same anti-flicker pipeline used in story 1-5 (whitespace-collapsed, attribute-sorted).

7. **`/preferences/theme` is registered as POST in all three route dumps:** `make parity` exits clean with `post /preferences/theme` appearing once in each stack's route inventory. This requires updating the .NET and Django route dumpers to emit the actual HTTP method (they currently hardcode `get`); the Fiber dumper already emits proper methods. See **Parity Tooling Updates** below — these are part of this story.

8. **Keyboard activation:** Tabbing to the toggle and pressing Space or Enter cycles the theme (native `<button>` behavior). After activation, a screen reader announces the updated `aria-label` text containing the new current state and the next-cycle hint.

## Tasks / Subtasks

- [ ] Task 1: Vendor the client-side listener and document the inline script convention (AC: #1, #5)
  - [ ] 1.1 Create `fieldmark_shared/vendor/theme-toggle/theme-toggle.js` (≤ 20 LOC; spec below in Dev Notes)
  - [ ] 1.2 Create symlinks in each stack's vendor/ static dir following the existing `vendor/htmx` symlink pattern:
    - `.NET: FieldMark/FieldMark.Web/wwwroot/vendor/theme-toggle → ../../../../fieldmark_shared/vendor/theme-toggle`
    - `Django: fieldmark_py/static/vendor/theme-toggle → ../../../fieldmark_shared/vendor/theme-toggle`
    - `Go: fieldmark-go/internal/web/static/vendor/theme-toggle → ../../../../../fieldmark_shared/vendor/theme-toggle`
  - [ ] 1.3 Update `fieldmark_shared/CLAUDE.md` directory layout and rules sections to mention the new `vendor/theme-toggle/` subdir and the symlink table
  - [ ] 1.4 Add a sub-section to `_bmad-output/planning-artifacts/architecture.md` Frontend Architecture documenting the single inline `<script>` exception (its 5 lines, why it must be inline + blocking, its location in `<head>`)

- [ ] Task 2: Add ThemeToggle CSS to shared design system (AC: #3, #8)
  - [ ] 2.1 Add `.theme-toggle` button styles to `fieldmark_shared/src/_tokens.css` (or a new `_components.css` if cleaner) — 36×36, centered icon, hover/focus per Basecoat icon-button conventions
  - [ ] 2.2 Add three icon variants (`.theme-toggle__icon--system`, `…--light`, `…--dark`), each visible only when the parent `<html data-theme>` matches; or use a single icon element that swaps via `currentColor` SVG href. **Choose the all-CSS approach (visibility toggle) — no JS to switch icons during first paint.** See Dev Notes.
  - [ ] 2.3 Rebuild: `cd fieldmark_shared && pnpm run build` and commit `dist/fieldmark.css`

- [ ] Task 3: .NET implementation — server-side cookie read, partial, endpoint, dumper update (AC: #1, #2, #3, #4, #7)
  - [ ] 3.1 Create `FieldMark/FieldMark.Web/Pages/Shared/_ThemeToggle.cshtml` (markup partial — see Dev Notes for canonical HTML)
  - [ ] 3.2 Update `Pages/Shared/_Layout.cshtml`:
    - Read `Context.Request.Cookies["fm_theme"]` in the layout (or via a `@functions` block); default to `"system"` when absent or invalid
    - Emit `<html lang="en" data-theme="@theme">`
    - Emit the 5-line inline `<script>` immediately after `<head>` opens (before stylesheet link) — the script body is given in Dev Notes
    - Include `<script src="~/vendor/theme-toggle/theme-toggle.js"></script>` after the HTMX script tag
    - Render the partial in the header strip: `<partial name="_ThemeToggle" model="@theme" />` placed beside the (future) avatar slot per UX-DR15
  - [ ] 3.3 Create `Pages/Preferences/Theme.cshtml` + `Theme.cshtml.cs` with `@page "/preferences/theme"` directive, `OnPostAsync` handler:
    - Read `Request.Form["value"]`
    - Validate against allowed set `{"system","light","dark"}`; on invalid return `BadRequest()`
    - Append cookie via `Response.Cookies.Append("fm_theme", value, new CookieOptions { Path = "/", SameSite = SameSiteMode.Lax, MaxAge = TimeSpan.FromSeconds(31536000) })`
    - Set `Response.Headers["HX-Trigger"] = "theme-changed"`
    - Return `new StatusCodeResult(204)`
    - Add `[IgnoreAntiforgeryToken]` attribute on the handler — see Dev Notes "Antiforgery handling" for the rationale and explicit policy
  - [ ] 3.4 Update `FieldMark/FieldMark.Web/Tools/DumpRoutes.cs` to emit the actual HTTP method per endpoint by reading `HttpMethodMetadata` from each `RouteEndpoint`:
    - Replace the hardcoded `"get "` prefix with a method derived from `ep.Metadata.GetMetadata<HttpMethodMetadata>()?.HttpMethods` (lowercased, one line per (method,path) pair)
    - If a single Razor Page exposes both GET and POST (it will, by default, for any page with both `OnGet` and `OnPost`), emit a line per method
    - Keep all existing filtering (Error, Admin area, Index alias suppression)
  - [ ] 3.5 Verify routes: `dotnet run --project FieldMark/FieldMark.Web -- --dump-routes` includes both `get /` (if Index still applies) and `post /preferences/theme`

- [ ] Task 4: Django implementation — middleware/template context, partial, endpoint, dumper update (AC: #1, #2, #3, #4, #7)
  - [ ] 4.1 Create `fieldmark_py/templates/_theme_toggle.html` (partial — see Dev Notes for canonical HTML)
  - [ ] 4.2 Create a small context processor at `fieldmark_py/fieldmark/context_processors.py` exposing `fm_theme` from `request.COOKIES.get("fm_theme", "system")`, validated against the allowed set (fall back to `"system"` on unknown):
    - Register it in `settings.TEMPLATES[0]["OPTIONS"]["context_processors"]`
  - [ ] 4.3 Update `fieldmark_py/templates/base.html`:
    - Change `<html lang="en">` to `<html lang="en" data-theme="{{ fm_theme }}">`
    - Insert the 5-line inline `<script>` at the top of `<head>` (before stylesheet)
    - Include `<script src="{% static 'vendor/theme-toggle/theme-toggle.js' %}"></script>` after HTMX
    - Render the toggle in the header: `{% include "_theme_toggle.html" with current=fm_theme %}` next to the (future) avatar slot
  - [ ] 4.4 Add view `set_theme(request)` in `fieldmark_py/fieldmark/views.py`:
    - Decorate with `@require_POST`
    - Decorate with `@csrf_exempt` (or supply CSRF token; see Dev Notes — preferred path is to keep CSRF because the base layout already injects `X-CSRFToken` via `hx-headers`)
    - Validate `request.POST.get("value")`; on invalid return `HttpResponseBadRequest()`
    - Build `response = HttpResponse(status=204)`
    - `response.set_cookie("fm_theme", value, max_age=31536000, path="/", samesite="Lax")`
    - `response["HX-Trigger"] = "theme-changed"`
    - Return response
  - [ ] 4.5 Add URL pattern: `path("preferences/theme", views.set_theme, name="set_theme")` in `fieldmark_py/fieldmark/urls.py` — **no trailing slash** to match the canonical path exactly
  - [ ] 4.6 If keeping CSRF: confirm `base.html` retains the existing `hx-headers='{"X-CSRFToken": "{{ csrf_token }}"}'` attribute (this was preserved in story 1-5). Do NOT use `@csrf_exempt`.
  - [ ] 4.7 Update `fieldmark_py/tools/management/commands/dump_routes.py` to emit the actual HTTP method per route. Strategy:
    - For each `URLPattern`, inspect `pattern.callback`. If the callback is wrapped (Django wraps with `csrf_protect`, `require_POST`, etc.), unwrap via `inspect.unwrap` or walk `__wrapped__`.
    - Read `pattern.callback.view_class.http_method_names` for class-based views, OR detect `require_POST` / `require_http_methods` decorators via the function's `__qualname__` or by inspecting the closure cells of `django.views.decorators.http.require_http_methods` (returns a list on the wrapper).
    - Pragmatic fallback: introspect the function's source or rely on an explicit registry — see Dev Notes "Django dump-routes strategy" for the chosen approach (registry-based is simplest and traceable).
    - Emit `post /preferences/theme` (and continue emitting `get` lines for all other routes).
  - [ ] 4.8 Verify routes: `cd fieldmark_py && uv run python manage.py dump_routes` includes `post /preferences/theme`

- [ ] Task 5: Go/Fiber implementation — middleware-free cookie read, partial, endpoint (AC: #1, #2, #3, #4, #7)
  - [ ] 5.1 Create `fieldmark-go/internal/web/templates/partials/theme_toggle.html` (partial — see Dev Notes for canonical HTML; uses `{{define "theme_toggle"}}`)
  - [ ] 5.2 Update `fieldmark-go/internal/web/templates/layouts/base.html`:
    - Change `<html lang="en">` to `<html lang="en" data-theme="{{.FmTheme}}">`
    - Insert the 5-line inline `<script>` at the top of `<head>`
    - Add `<script src="/static/vendor/theme-toggle/theme-toggle.js"></script>` after HTMX
  - [ ] 5.3 Update `fieldmark-go/internal/web/templates/partials/header.html` to include `{{template "theme_toggle" .}}` beside the (future) avatar slot
  - [ ] 5.4 In `fieldmark-go/cmd/web/main.go`:
    - Add a small helper `resolveFmTheme(c fiber.Ctx) string` that reads `c.Cookies("fm_theme", "system")` and validates against `{"system","light","dark"}` (fallback `"system"` on unknown)
    - In each existing page handler (`/` and `/privacy`), inject `"FmTheme": resolveFmTheme(c)` into the `fiber.Map`
    - Add new route: `app.Post("/preferences/theme", func(c fiber.Ctx) error { … })` with:
      - `value := c.FormValue("value")`
      - Validate; on invalid return `c.SendStatus(400)`
      - `c.Cookie(&fiber.Cookie{Name: "fm_theme", Value: value, Path: "/", MaxAge: 31536000, SameSite: "Lax"})`
      - `c.Set("HX-Trigger", "theme-changed")`
      - `return c.SendStatus(204)`
    - Use the third-arg `""` (no layout) form is **not** applicable here because there is no Render call — we return 204 directly
  - [ ] 5.5 Verify routes: `cd fieldmark-go && go run ./cmd/web -dump-routes` includes `post /preferences/theme` (Fiber dumper already method-aware)

- [ ] Task 6: Parity verification and cross-stack byte-diff (AC: #6, #7)
  - [ ] 6.1 Start all three stacks (`make up`, then `make run-net`, `make run-django`, `make run-go` in separate shells)
  - [ ] 6.2 `curl -s -i http://localhost:5000/ http://localhost:8000/ http://localhost:3000/` — confirm `<html data-theme="system">` in all three
  - [ ] 6.3 `curl -i -X POST -d "value=light" http://localhost:5000/preferences/theme` (and equivalent for Django and Go) — confirm `204`, `Set-Cookie: fm_theme=light; …`, and `HX-Trigger: theme-changed`
  - [ ] 6.4 `curl -s --cookie "fm_theme=dark" http://localhost:5000/` (etc.) — confirm `<html data-theme="dark">`
  - [ ] 6.5 Diff the captured chrome HTML across stacks (header + inline script + toggle button) — must be byte-identical modulo per-stack server-rendered values (none expected)
  - [ ] 6.6 `make parity` — must exit 0 with the new POST route present in all three dumps

- [ ] Task 7: Manual keyboard + screen-reader smoke test (AC: #8)
  - [ ] 7.1 In each stack, tab to the ThemeToggle, press Space — confirm theme changes immediately (no full page reload)
  - [ ] 7.2 Confirm `aria-label` text updates (use browser devtools to inspect after activation)
  - [ ] 7.3 Optional but recommended: drive VoiceOver/NVDA over the button to confirm the announcement reflects the new state

## Dev Notes

### Current State (READ BEFORE CHANGING)

**`fieldmark_shared/src/fieldmark.css` (10 lines):** Imports Tailwind v4; declares three `@source` directives for the three stacks' template directories. **Story 1-4 (`ready-for-dev`) will add Basecoat, semantic tokens, fonts, and partial CSS files.** Story 1-6 must add the ThemeToggle component CSS into whatever structure story 1-4 establishes (likely `_components.css` or extending `_tokens.css`). If story 1-4 has not been implemented when this story starts, coordinate: 1-4 is a hard prerequisite for parts of 1-6's CSS work.

**Base layouts (modified by story 1-5):**
- **`.NET — FieldMark/FieldMark.Web/Pages/Shared/_Layout.cshtml`:** Story 1-5 introduces skip-link, `<nav aria-label="Main">`, `<main id="main-content">`, and `#flash-region`. **1-6 builds on the 1-5 result**, not the pre-1-5 layout. If 1-5 has not yet been implemented when 1-6 starts, treat the 1-5 target HTML (documented in story 1-5 §Target HTML Structure) as the contract this story extends.
- **`Django — fieldmark_py/templates/base.html`:** Same as above. Must preserve `hx-headers='{"X-CSRFToken": "{{ csrf_token }}"}'` on `<body>` (established by 1-1, preserved by 1-5). This attribute is what makes CSRF work for the ThemeToggle's `hx-post`.
- **`Go — fieldmark-go/internal/web/templates/layouts/base.html` + `partials/header.html`:** 1-5 introduces skip-link in base, restructures header/footer to semantic landmarks. The Go view model must gain a `FmTheme` field; existing handlers pass anonymous `fiber.Map`, so just add the key.

**`fieldmark-go/cmd/web/main.go`:** Currently uses `flag.Bool("dump-routes", …)`. The Fiber dumper already prints `lowercase(method) lowercase(path)` per route, so adding `app.Post("/preferences/theme", …)` is automatically picked up.

**Parity scripts:** `tools/parity/diff-routes.sh` compares .NET ↔ Django and .NET ↔ Fiber (transitivity covers the third pair). The `dump-routes-net.sh` script greps for lines starting with the seven HTTP methods, so when DumpRoutes.cs emits `post /preferences/theme`, it will pass through. Same for `dump-routes-django.sh` (no grep filter — pipes the management command's full output).

**Existing endpoint conventions (1-1 baseline):** All three stacks register `/` (dashboard), `/privacy`, and `/fragments/compliance-tile` as GET. The compliance-tile fragment is the only non-page route today. Story 1-6 adds the first POST route.

### Canonical Inline `<script>` (5 lines; identical across all three stacks)

Place inside `<head>`, after `<meta name="viewport">`, **before** the stylesheet `<link>` (so it runs before paint):

```html
<script>
(function(){var d=document.documentElement,t=d.getAttribute('data-theme');
if(t!=='system')return;
d.setAttribute('data-theme',window.matchMedia('(prefers-color-scheme: dark)').matches?'dark':'light');})();
</script>
```

This is 5 lines including the `<script>` tags. The IIFE prevents global leakage. Reads the server-rendered `data-theme`; if `"system"`, replaces it with the resolved value. No-op for `"light"` or `"dark"` (which the server emitted directly). This is the **only** inline JavaScript in the application — document it in `architecture.md` Frontend Architecture.

### Canonical Vendored Listener `theme-toggle.js` (≤ 20 LOC executable)

```javascript
(function () {
  'use strict';
  var ORDER = ['system', 'light', 'dark'];
  var ICONS = { system: 'Monitor', light: 'Sun', dark: 'Moon' };

  function readCookie(name) {
    var m = document.cookie.match(new RegExp('(^| )' + name + '=([^;]+)'));
    return m ? m[2] : 'system';
  }

  function resolve(pref) {
    if (pref !== 'system') return pref;
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
  }

  document.body.addEventListener('theme-changed', function () {
    var pref = readCookie('fm_theme');
    var resolved = resolve(pref);
    var next = ORDER[(ORDER.indexOf(pref) + 1) % ORDER.length];
    document.documentElement.setAttribute('data-theme', pref);
    var btn = document.querySelector('[data-theme-toggle]');
    if (!btn) return;
    btn.setAttribute('aria-label', 'Theme: ' + pref + '; activate to cycle (next: ' + next + ')');
    btn.dataset.themeResolved = resolved;
    // Update hx-vals so next click sends the next value in the cycle:
    btn.setAttribute('hx-vals', JSON.stringify({ value: next }));
  });
})();
```

Notes on this listener:
- Writes the cycled preference (`pref`) to `<html data-theme>`. CSS handles `[data-theme="system"]` via `prefers-color-scheme` media queries (story 1-4's token CSS must include the `@media (prefers-color-scheme: dark)` branch keyed inside `[data-theme="system"]`).
- Updates the button's `hx-vals` so the *next* click sends the *next* cycle value. The server-rendered initial markup must already include `hx-vals` with the correct next value.
- Uses `[data-theme-toggle]` as a stable selector — the ThemeToggle button must carry this attribute.
- Total: 20 lines of executable code (excluding blank lines and comments). If you can shave it shorter without harming readability, do.

### Canonical ThemeToggle Button HTML (must be byte-identical across stacks)

```html
<button type="button"
        class="theme-toggle"
        data-theme-toggle
        data-theme-resolved="{resolved}"
        aria-label="Theme: {current}; activate to cycle (next: {next})"
        hx-post="/preferences/theme"
        hx-vals='{"value":"{next}"}'
        hx-swap="none">
  <svg class="theme-toggle__icon theme-toggle__icon--system" aria-hidden="true" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="3" width="20" height="14" rx="2"/><line x1="8" y1="21" x2="16" y2="21"/><line x1="12" y1="17" x2="12" y2="21"/></svg>
  <svg class="theme-toggle__icon theme-toggle__icon--light" aria-hidden="true" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="5"/><line x1="12" y1="1" x2="12" y2="3"/><line x1="12" y1="21" x2="12" y2="23"/><line x1="4.22" y1="4.22" x2="5.64" y2="5.64"/><line x1="18.36" y1="18.36" x2="19.78" y2="19.78"/><line x1="1" y1="12" x2="3" y2="12"/><line x1="21" y1="12" x2="23" y2="12"/><line x1="4.22" y1="19.78" x2="5.64" y2="18.36"/><line x1="18.36" y1="5.64" x2="19.78" y2="4.22"/></svg>
  <svg class="theme-toggle__icon theme-toggle__icon--dark" aria-hidden="true" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/></svg>
</button>
```

`{current}` and `{next}` are server-rendered. `{resolved}` is the resolved theme name (used by CSS to show the correct icon — only the icon matching the resolved theme is visible at any time). All three SVG icons are present; CSS hides two via `[data-theme-resolved="…"] .theme-toggle__icon--…` rules.

**Why all three SVGs server-rendered:** so the icon swap on click is instant (no JS-driven DOM construction), and the byte-identical chrome diff across stacks remains true. CSS does the visibility based on `data-theme-resolved`.

### CSS for icon visibility (add to story 1-4's `_components.css` or `_tokens.css`)

```css
.theme-toggle { width: 36px; height: 36px; display: inline-flex; align-items: center; justify-content: center; position: relative; }
.theme-toggle__icon { width: 20px; height: 20px; display: none; }
.theme-toggle[data-theme-resolved="system"] .theme-toggle__icon--system,
.theme-toggle[data-theme-resolved="light"]  .theme-toggle__icon--light,
.theme-toggle[data-theme-resolved="dark"]   .theme-toggle__icon--dark { display: block; }
```

**Note on `[data-theme="system"]` color resolution:** Story 1-4 defines the semantic color tokens. When `<html data-theme="system">` is rendered server-side and the inline script runs, the script *replaces* it with `"light"` or `"dark"`. So your CSS only needs to handle `[data-theme="light"]` and `[data-theme="dark"]` — it never sees `"system"` after first paint, because the inline script has already swapped it. If 1-4's CSS structure assumes a `[data-theme="system"]` branch with nested media queries, simplify it — the system value is resolved before any CSS reads it.

### Antiforgery handling per stack

- **.NET:** Razor Pages by default require an antiforgery token on POST. The ThemeToggle's `hx-post` does not currently send one. Apply `[IgnoreAntiforgeryToken]` on the `Theme.cshtml.cs` PageModel. Document this in the file as a justified exception: "Theme preference is non-security-sensitive UI state; no CSRF protection required." Do NOT disable global antiforgery — keep the default for every other handler in the app.
- **Django:** The existing `hx-headers='{"X-CSRFToken": "{{ csrf_token }}"}'` on `<body>` automatically attaches the CSRF token to all HTMX POSTs. **Do not use `@csrf_exempt`** — keep CSRF on for parity with the rest of the app. Verify with browser devtools that the header is sent.
- **Go/Fiber:** Fiber has no CSRF middleware mounted today (and won't until auth lands). No special handling needed.

### Cookie semantics (must match exactly across stacks)

| Attribute | Value | Notes |
|---|---|---|
| Name | `fm_theme` | Lowercase, underscore separator |
| Value | `system` ∣ `light` ∣ `dark` | Server validates; rejects any other value |
| `Path` | `/` | Site-wide |
| `Max-Age` | `31536000` | 1 year in seconds |
| `SameSite` | `Lax` | Allow first-party navigation |
| `HttpOnly` | **omitted** | Client JS reads `document.cookie` in the listener |
| `Secure` | **omitted** | MVP is local-only (http://localhost) |
| `Domain` | **omitted** | Defaults to current host |

### Endpoint contract (must match exactly across stacks)

```
POST /preferences/theme
Content-Type: application/x-www-form-urlencoded
Body: value=<system|light|dark>

200 cases never occur — only:
204 No Content
  Set-Cookie: fm_theme=<value>; Path=/; SameSite=Lax; Max-Age=31536000
  HX-Trigger: theme-changed
  (no body)

400 Bad Request
  (when value is missing or not in the allowed set; no Set-Cookie)
```

GET on this path returns 404/405 (depending on framework default) — do not register a GET handler.

### Files to create/modify

| Action | Stack | File |
|---|---|---|
| NEW | shared | `fieldmark_shared/vendor/theme-toggle/theme-toggle.js` |
| UPDATE | shared | `fieldmark_shared/src/_tokens.css` or `_components.css` (per 1-4 structure) — add `.theme-toggle` + icon visibility rules |
| UPDATE | shared | `fieldmark_shared/dist/fieldmark.css` (recompile) |
| UPDATE | shared | `fieldmark_shared/CLAUDE.md` (vendor table + new subdir) |
| UPDATE | docs | `_bmad-output/planning-artifacts/architecture.md` (document the single inline script exception) |
| NEW | .NET | `FieldMark/FieldMark.Web/Pages/Shared/_ThemeToggle.cshtml` |
| NEW | .NET | `FieldMark/FieldMark.Web/Pages/Preferences/Theme.cshtml` |
| NEW | .NET | `FieldMark/FieldMark.Web/Pages/Preferences/Theme.cshtml.cs` |
| UPDATE | .NET | `FieldMark/FieldMark.Web/Pages/Shared/_Layout.cshtml` (data-theme, inline script, partial render, theme-toggle.js script tag) |
| UPDATE | .NET | `FieldMark/FieldMark.Web/Tools/DumpRoutes.cs` (emit actual HTTP method per endpoint) |
| NEW | .NET symlink | `FieldMark/FieldMark.Web/wwwroot/vendor/theme-toggle` → `../../../../fieldmark_shared/vendor/theme-toggle` |
| NEW | Django | `fieldmark_py/templates/_theme_toggle.html` |
| NEW | Django | `fieldmark_py/fieldmark/context_processors.py` |
| UPDATE | Django | `fieldmark_py/fieldmark/settings.py` (register context processor) |
| UPDATE | Django | `fieldmark_py/fieldmark/views.py` (add `set_theme` view) |
| UPDATE | Django | `fieldmark_py/fieldmark/urls.py` (add `preferences/theme` URL) |
| UPDATE | Django | `fieldmark_py/templates/base.html` (data-theme, inline script, include partial, theme-toggle.js) |
| UPDATE | Django | `fieldmark_py/tools/management/commands/dump_routes.py` (method-aware emit) |
| NEW | Django symlink | `fieldmark_py/static/vendor/theme-toggle` → `../../../fieldmark_shared/vendor/theme-toggle` |
| NEW | Go | `fieldmark-go/internal/web/templates/partials/theme_toggle.html` |
| UPDATE | Go | `fieldmark-go/internal/web/templates/layouts/base.html` (data-theme, inline script, theme-toggle.js) |
| UPDATE | Go | `fieldmark-go/internal/web/templates/partials/header.html` (include theme_toggle partial) |
| UPDATE | Go | `fieldmark-go/cmd/web/main.go` (resolveFmTheme helper, FmTheme in render maps, POST handler) |
| NEW | Go symlink | `fieldmark-go/internal/web/static/vendor/theme-toggle` → `../../../../../fieldmark_shared/vendor/theme-toggle` |

### Django dump-routes strategy

The current dump_routes.py emits `get <path>` for every route. To emit the correct HTTP method for `/preferences/theme` (POST), choose the simplest workable approach:

**Recommended: introspect `require_POST` / `require_http_methods` decorators.** Django's `require_http_methods` sets an attribute on the wrapped view function. Specifically, the wrapper has the original function accessible via `__wrapped__` (when decorators use `functools.wraps`), and `require_http_methods` itself stores the methods list on the wrapper as a closure variable. The robust approach:

```python
def _methods_for(callback):
    # Class-based views expose http_method_names; the URLPattern.callback
    # is the result of as_view(), which sets view_class on the callback.
    view_class = getattr(callback, "view_class", None)
    if view_class is not None:
        # Filter to declared methods (default = all). Better: detect overridden methods.
        return [m for m in view_class.http_method_names if hasattr(view_class, m)]
    # Function-based view decorated with require_http_methods stores the list
    # on the wrapper as `request_method_list` (this is an implementation
    # detail of django.views.decorators.http but stable since Django 1.x).
    methods = getattr(callback, "request_method_list", None)
    if methods is not None:
        return [m.lower() for m in methods]
    # Undecorated function-based view: assume GET (the safe default for FieldMark
    # given every existing route is a page render).
    return ["get"]
```

Update the `_collect` recursion to emit one line per (method, path) pair. Test by running `dump_routes` after adding the `set_theme` view — `post /preferences/theme` must appear exactly once.

### .NET dump-routes strategy

Modify `DumpRoutes.cs` to read `HttpMethodMetadata` from each endpoint. For Razor Pages, the framework registers separate endpoints per HTTP method when handlers (`OnGet`, `OnPost`) are declared, each with `HttpMethodMetadata` containing the single method. Replace the hardcoded `"get "` prefix:

```csharp
var methods = ep.Metadata.GetMetadata<HttpMethodMetadata>()?.HttpMethods;
if (methods is null || methods.Count == 0)
    return Array.Empty<string>();
return methods.Select(m => $"{m.ToLowerInvariant()} {path}");
```

For a Razor Page with both `OnGet` and `OnPost`, two endpoints are registered and the dumper will emit both lines.

### Architecture Compliance

- **D11 — HTMX target ID inventory:** ThemeToggle uses `hx-swap="none"` and returns 204; no DOM swap occurs, no target ID is needed. Do not add a new ID to the canonical inventory.
- **D12 — Partial-naming convention:** .NET `_ThemeToggle.cshtml`, Django `_theme_toggle.html`, Go `theme_toggle.html` (in `partials/`).
- **D15 — Vendor locally, no CDN:** the listener is vendored under `fieldmark_shared/vendor/theme-toggle/`; symlinked into each stack.
- **D16 — Manual CSS compile:** rebuild + commit `dist/fieldmark.css`.
- **UX-DR5:** First-paint resolution exactly as specified (5-line inline script, `fm_theme` cookie, `HX-Trigger: theme-changed`, 204 response).
- **UX-DR15:** 36×36 button beside avatar; `aria-label` describes current + next; keyboard activatable.
- **FR54 (POST-only mutations):** `/preferences/theme` is POST only — never GET.
- **FR60 (keyboard operable):** native `<button>` Space/Enter activation.
- **FR62 (focus after swap):** N/A — `hx-swap="none"` means no swap; focus stays on the button.
- **Cross-stack symmetry (root CLAUDE.md hard rule):** the endpoint path, HTTP method, cookie attributes, response status, and header are canonical.

### Anti-Patterns to Avoid

- Do NOT add a second inline `<script>` anywhere. The 5-line first-paint resolver is the only inline JS permitted.
- Do NOT use `localStorage` for the preference. Cookie only — the server must read it.
- Do NOT set `HttpOnly` on the cookie (the listener reads it via `document.cookie`).
- Do NOT register a GET handler at `/preferences/theme` (POST-only per FR54).
- Do NOT use `@csrf_exempt` in Django; the existing `hx-headers` body attribute provides the token.
- Do NOT issue a redirect from the POST handler — return 204 with `HX-Trigger`.
- Do NOT swap any DOM element from the response — `hx-swap="none"` is intentional; the listener does the visual update.
- Do NOT add a watcher to `prefers-color-scheme` changes that auto-updates when `fm_theme="system"`. The MVP behavior is: theme resolves once per page load (first paint). System-pref changes mid-session do not retroactively flip the UI. (This is a deliberate simplicity choice; can revisit later.)
- Do NOT name the cookie anything other than `fm_theme`. Path/SameSite/Max-Age must match exactly.
- Do NOT introduce a per-stack `theme.js` — the listener is shared and vendored.
- Do NOT skip updating the dumpers. AC #7 is part of this story; the parity script must show `post /preferences/theme` in all three dumps.

### Testing Approach

No automated tests at this story (Playwright comes in Epic 7). Verification is manual + parity-tooling driven:

1. Build all three stacks; start them (or rely on `make parity` which dumps routes without DB).
2. Run `make parity` — must exit 0 with `post /preferences/theme` in all three dumps.
3. For each stack, in a browser:
   a. Open `/` in private/incognito (no cookie). Expect no flash. `data-theme` ends up resolved to OS preference.
   b. Click the ThemeToggle. Confirm the icon and theme change instantly.
   c. Reload. Confirm the new theme is applied at first paint (no flash).
   d. Cycle through all three states (system → light → dark → system). Confirm each persists.
   e. Tab to the toggle and press Space — confirm keyboard activation works.
4. Cross-stack diff: `curl -s --cookie "fm_theme=light" http://localhost:5000/ http://localhost:8000/ http://localhost:3000/`; compare the `<head>` and header chrome — should be byte-identical (after whitespace normalization).

### Previous Story Intelligence

- **Stories 1-1, 1-2, 1-3:** done. Scaffolds, SQL init, parity tooling confirmed.
- **Story 1-4 (ready-for-dev):** Bootstraps design system — Basecoat, semantic tokens, fonts, status badge vocab, `dist/fieldmark.css`. **Prerequisite** for 1-6's ThemeToggle CSS. The `.theme-toggle` rules must land in whatever partial structure 1-4 creates (`_components.css` is the natural home if 1-4 creates one; otherwise extend `_tokens.css`).
- **Story 1-5 (ready-for-dev):** Cross-stack base layout — skip-link, landmarks, FlashRegion. **Prerequisite** for 1-6 because 1-6 modifies the same layout files (`_Layout.cshtml`, `base.html`, `layouts/base.html` + `partials/header.html`). The recommended sequence is 1-4 → 1-5 → 1-6. If 1-5 is in progress when 1-6 starts, coordinate the layout edits to avoid merge conflicts.
- **Story 1-5's preserved invariants** that 1-6 must continue to preserve:
  - Django: `hx-headers='{"X-CSRFToken": "{{ csrf_token }}"}'` on `<body>` — keep it; the ThemeToggle's POST relies on it.
  - .NET: `@RenderSectionAsync("Scripts", required: false)` call — preserve.
  - All: AG Grid script loads before HTMX — preserve. The new `theme-toggle.js` loads **after** HTMX so it can use the HTMX-dispatched `theme-changed` event.

### Git Intelligence

- Commit convention from history: `feat: :sparkles: e1s{N} {description}`. Use `feat: :sparkles: e1s6 themetoggle with cookie persistence per stack`.
- Recent commits (`d03f0fe`, `cbf47e9`, `a6fac88`) closed 1-3, 1-2, 1-1 respectively. Story 1-4's git work is still pending (the `fieldmark_shared/package.json`, `pnpm-lock.yaml`, and `vendor/fonts/` files appear modified/untracked in current `git status`).
- This story is a 4-file-touched story per stack plus shared and docs — keep the commit focused.

### Project Structure Notes

- No new top-level directories. New subdirs are: `fieldmark_shared/vendor/theme-toggle/`, `FieldMark/FieldMark.Web/Pages/Preferences/`, and the three new symlink targets in each stack's vendor dir.
- The `Pages/Preferences/` directory is new in the .NET stack but matches the `Pages/Fragments/` pattern already established (one folder per loose grouping under the Razor Pages root).
- Django adds a `context_processors.py` to the `fieldmark/` project package (alongside `urls.py`, `views.py`) — no new app, no new directory.
- Go does not need a new directory; the handler is registered inline in `main.go` per the existing pattern (no `internal/web/handlers/` is in use yet — that directory has only `.gitkeep`).

### References

- [Source: _bmad-output/planning-artifacts/epics.md — Story 1.6, UX-DR5, UX-DR15]
- [Source: _bmad-output/planning-artifacts/ux-design-specification.md §Step 8 — theme switch convention, lines 489–501]
- [Source: _bmad-output/planning-artifacts/ux-design-specification.md §ThemeToggle component spec, lines 894–900]
- [Source: _bmad-output/planning-artifacts/architecture.md §D11 HTMX target IDs, §D15 vendoring, §D16 Tailwind compile]
- [Source: _bmad-output/planning-artifacts/prd/web-app-specific-requirements.md — FR54 POST-only, FR60 keyboard, FR62 focus]
- [Source: _bmad-output/implementation-artifacts/1-5-implement-cross-stack-base-layout-…md — target HTML structure, layout edits this story extends]
- [Source: _bmad-output/implementation-artifacts/1-4-bootstrap-design-system-foundation-in-fieldmark-shared.md — CSS partial structure, Basecoat pin]
- [Source: docs/hard-rules.md — stack symmetry on routes; canonical wire format]
- [Source: docs/architecture.md — canonical HTMX target IDs, HTMX patterns]
- [Source: CLAUDE.md (root) — three-stack constraint, story-not-done-until-three-stack-pass]
- [Source: FieldMark/CLAUDE.md, fieldmark_py/CLAUDE.md, fieldmark-go/CLAUDE.md, fieldmark_shared/CLAUDE.md — per-stack rules]

## Dev Agent Record

### Agent Model Used

claude-opus-4-7

### Debug Log References

### Completion Notes List

- Ultimate context engine analysis completed — comprehensive developer guide created.
- Story extends the layouts established by 1-5 and the CSS foundation established by 1-4 (both `ready-for-dev` at the time of story creation).
- Adds the first POST route in the codebase and the first need for method-aware route dumping in .NET and Django.
- The single inline `<script>` is documented as a deliberate exception to "no inline JS" and must remain the only one in the application.
- Five canonical artifacts must remain byte-identical across stacks: the `<html data-theme>` attribute, the 5-line inline `<script>`, the ThemeToggle button HTML, the `Set-Cookie` header, and the `HX-Trigger: theme-changed` response header.

### File List
