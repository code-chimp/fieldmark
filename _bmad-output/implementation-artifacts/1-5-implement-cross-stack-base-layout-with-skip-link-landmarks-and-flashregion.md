# Story 1.5: Implement cross-stack base layout with skip-link, landmarks, and FlashRegion

Status: ready-for-dev

## Story

As a screen-reader user landing on any FieldMark page on any stack,
I want a consistent landmark structure, a working skip-link, and a polite live region for system announcements,
So that I can navigate the application predictably regardless of which stack served the page.

## Acceptance Criteria

1. **Skip-link:** Each stack's base layout (`_Layout.cshtml`, `base.html`, `layouts/base.html`) has a "Skip to main content" link as the first focusable element targeting `#main-content`. The link is visually hidden until focused.

2. **Landmark structure:** Exactly one `<header>`, one `<nav aria-label="Main">`, one `<main id="main-content">`, an optional `<aside>` slot for EntityRail, and an optional `<footer>`. No nested landmarks of same role.

3. **Heading hierarchy:** Exactly one `<h1>` per page; heading levels never skip (no `<h3>` without a prior `<h2>`).

4. **FlashRegion:** `<div id="flash-region" role="status" aria-live="polite" aria-atomic="false">` is present in page chrome, empty by default, renders messages from a per-stack `flash_messages()` template helper.

5. **Focus styling:** `:focus-visible` ring is 2px wide at 2px offset in body text color. Touch targets render at >= 44x44px under `(pointer: coarse)` media query.

6. **Cross-stack parity:** The rendered HTML chrome (header skeleton, nav skeleton, skip-link, FlashRegion, main slot, footer skeleton) of `/` is byte-identical across all three stacks (modulo per-stack server-rendered values — none expected at this story).

## Tasks / Subtasks

- [ ] Task 1: Update `fieldmark_shared/src/fieldmark.css` with skip-link, focus, and touch-target styles (AC: #1, #5)
  - [ ] 1.1 Add `.skip-link` class (sr-only by default, visible on `:focus`)
  - [ ] 1.2 Add `:focus-visible` ring: 2px solid, 2px offset, body text color
  - [ ] 1.3 Add `(pointer: coarse)` media query for 44x44px minimum touch targets on interactive elements
  - [ ] 1.4 Recompile: `cd fieldmark_shared && pnpm run build`
  - [ ] 1.5 Commit updated `dist/fieldmark.css`

- [ ] Task 2: Rewrite .NET `_Layout.cshtml` (AC: #1, #2, #3, #4, #6)
  - [ ] 2.1 Add skip-link as first child of `<body>`
  - [ ] 2.2 Restructure to semantic landmarks: `<header>` containing `<nav aria-label="Main">`, `<main id="main-content">`, `<footer>`
  - [ ] 2.3 Add `<div id="flash-region" role="status" aria-live="polite" aria-atomic="false"></div>` inside `<main>` before `@RenderBody()`
  - [ ] 2.4 Create `_FlashRegion.cshtml` partial for message rendering
  - [ ] 2.5 Ensure exactly one `<h1>` via page title convention

- [ ] Task 3: Rewrite Django `base.html` (AC: #1, #2, #3, #4, #6)
  - [ ] 3.1 Add skip-link as first child of `<body>`
  - [ ] 3.2 Restructure to semantic landmarks: `<header>` containing `<nav aria-label="Main">`, `<main id="main-content">`, `<footer>`
  - [ ] 3.3 Add `<div id="flash-region" role="status" aria-live="polite" aria-atomic="false"></div>` inside `<main>` before `{% block content %}`
  - [ ] 3.4 Create `_flash_region.html` partial with `{% for message in messages %}` loop using Django messages framework
  - [ ] 3.5 Ensure exactly one `<h1>` via block convention

- [ ] Task 4: Rewrite Go `layouts/base.html`, `partials/header.html`, `partials/footer.html` (AC: #1, #2, #3, #4, #6)
  - [ ] 4.1 Add skip-link as first child of `<body>` in base layout
  - [ ] 4.2 Restructure header partial: `<header>` containing `<nav aria-label="Main">`
  - [ ] 4.3 Restructure footer partial: semantic `<footer>`
  - [ ] 4.4 Add `<div id="flash-region" role="status" aria-live="polite" aria-atomic="false"></div>` inside `<main>` before content block
  - [ ] 4.5 Create flash region template helper (Fiber context-based)
  - [ ] 4.6 Ensure exactly one `<h1>` via block convention

- [ ] Task 5: Cross-stack byte-parity verification (AC: #6)
  - [ ] 5.1 Start all three stacks and curl `/` from each
  - [ ] 5.2 Normalize and diff the rendered chrome HTML
  - [ ] 5.3 Fix any divergences until byte-identical

## Dev Notes

### Current State of Layout Files (READ BEFORE CHANGING)

All three stacks already have working base layouts from story 1.1. The task is to **rewrite them in place** to add accessibility landmarks, skip-link, and FlashRegion while preserving existing functionality.

**.NET — `FieldMark/FieldMark.Web/Pages/Shared/_Layout.cshtml` (50 lines):**
- Has `<header>`, `<nav>` (no `aria-label`), `<main role="main">`, `<footer>`
- Has mobile hamburger toggle with inline `onclick` JS
- Uses `asp-append-version="true"` on CSS link and `site.js` script
- Scripts: AG Grid in `<head>`, HTMX + site.js before `</body>`
- Has `@RenderSectionAsync("Scripts", required: false)` for page-specific scripts
- **Missing:** skip-link, `aria-label="Main"` on nav, `id="main-content"` on main, FlashRegion, focus styles

**Django — `fieldmark_py/templates/base.html` (34 lines):**
- Minimal: `<header>` with `<nav>`, `<main>`, `<footer>`
- Has `hx-headers='{"X-CSRFToken": "{{ csrf_token }}"}'` on `<body>` — MUST PRESERVE
- Has `{% block extra_css %}`, `{% block content %}`, `{% block extra_scripts %}`
- Scripts: AG Grid then HTMX before `</body>`
- **Missing:** skip-link, `aria-label="Main"` on nav, `id="main-content"` on main, FlashRegion

**Go — `fieldmark-go/internal/web/templates/layouts/base.html` (30 lines):**
- Uses `{{template "header" .}}` and `{{template "footer" .}}` (separate partials)
- Header partial (`partials/header.html`): `<header>` wrapping `<nav>` (no `aria-label`)
- Footer partial (`partials/footer.html`): `<footer>` with copyright
- Base has `<main class="container mx-auto px-4 py-6">`, blocks for `title`, `content`, `extra_css`, `scripts`
- Scripts: AG Grid then HTMX before `</body>`
- **Missing:** skip-link, `aria-label="Main"` on nav, `id="main-content"` on main, FlashRegion

### Target HTML Structure (all three stacks must produce this)

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>{Page Title} — FieldMark</title>
    <link rel="stylesheet" href="{vendor/fieldmark.css}">
    {extra_css_block}
</head>
<body{django: hx-headers for CSRF}>
    <a href="#main-content" class="skip-link">Skip to main content</a>

    <header>
        <nav aria-label="Main" class="...">
            <a href="/">FieldMark</a>
            {nav links}
        </nav>
    </header>

    <main id="main-content">
        <div id="flash-region" role="status" aria-live="polite" aria-atomic="false"></div>
        {page content}
    </main>

    <footer>
        <p>&copy; 2026 FieldMark</p>
    </footer>

    <script src="{vendor/ag-grid/35.2.1/ag-grid-community.min.js}"></script>
    <script src="{vendor/htmx/htmx.min.js}"></script>
    {extra_scripts_block}
</body>
</html>
```

### CSS Requirements for `fieldmark_shared/src/fieldmark.css`

Add these rules (Tailwind v4 `@layer` or `@utility` as appropriate):

**Skip-link:** Visually hidden by default, visible on focus. Pattern:
```css
.skip-link {
  position: absolute;
  left: -9999px;
  top: auto;
  width: 1px;
  height: 1px;
  overflow: hidden;
  z-index: 9999;
}
.skip-link:focus {
  position: fixed;
  top: 0;
  left: 0;
  width: auto;
  height: auto;
  padding: 0.75rem 1.5rem;
  background: var(--color-background, #fff);
  color: var(--color-foreground, #000);
  font-weight: 600;
  z-index: 9999;
  outline: 2px solid currentColor;
  outline-offset: 2px;
}
```

**Focus ring (global):**
```css
:focus-visible {
  outline: 2px solid currentColor;
  outline-offset: 2px;
}
```

**Touch targets:**
```css
@media (pointer: coarse) {
  button, a, [role="button"], input, select, textarea {
    min-height: 44px;
    min-width: 44px;
  }
}
```

After adding, rebuild: `cd fieldmark_shared && pnpm run build` and commit `dist/fieldmark.css`.

### FlashRegion Implementation Per Stack

**The FlashRegion is empty by default.** At this story, no messages will be rendered — the `<div>` must simply be present in the DOM. Future stories (1.11+) will wire flash messages through each stack's native mechanism.

Prepare the plumbing now:

- **.NET:** Create `Pages/Shared/_FlashRegion.cshtml`. Use `TempData["FlashMessages"]` or a similar mechanism. For now, render nothing inside the div.
- **Django:** Create `templates/_flash_region.html`. Use the Django messages framework (`{% for message in messages %}`). For now, render nothing.
- **Go:** Create `templates/partials/flash_region.html` as a named template `{{define "flash_region"}}`. Use a `FlashMessages` field on the base view model. For now, render nothing.

### Architecture Compliance

- **D11 — HTMX target ID inventory:** `#flash-region` is canonical. Do not invent new IDs.
- **D12 — Partial naming:** `.NET: _FlashRegion.cshtml`, Django: `_flash_region.html`, Go: `flash_region.html` (in `partials/`)
- **UX-DR14, UX-DR32:** FlashRegion is `aria-live="polite"`, `aria-atomic="false"`, `role="status"`
- **UX-DR33:** Skip-link targets `#main-content`, visually hidden until focused
- **UX-DR35:** Focus ring 2px at 2px offset; touch targets >= 44x44px on coarse pointers
- **FR60:** All interactive controls keyboard-operable; tab order matches visual order; focus visible
- **FR62:** HTMX swaps shift focus to swapped region (plumbing only — no HTMX swaps at this story)

### Anti-Patterns to Avoid

- Do NOT add `role="main"` on `<main>` — the element itself is the landmark. The .NET layout currently has `role="main"` which is redundant; remove it.
- Do NOT nest landmarks of the same role (e.g., `<nav>` inside `<nav>`).
- Do NOT use `aria-live="assertive"` on FlashRegion — polite is correct for non-blocking announcements. `assertive` is reserved for InlineAlert (danger/warning) per UX spec.
- Do NOT use `aria-atomic="true"` on FlashRegion — `false` is specified so only new messages are announced, not the entire region.
- Do NOT add business logic, auth checks, or role-based rendering at this story.
- Do NOT modify AG Grid script loading order — AG Grid must load before HTMX (documented constraint).
- Do NOT remove the Django CSRF token `hx-headers` attribute on `<body>`.
- Do NOT remove the .NET `@RenderSectionAsync("Scripts")` call.

### Stack-Specific Gotchas

- **.NET:** Remove the inline `onclick` JS on the hamburger button — it will be replaced with proper accessible navigation in a later story. For now, keep navigation simple (no mobile toggle). Remove `role="main"` from `<main>`.
- **Django:** Keep `hx-headers` CSRF on `<body>`. The `{% load static %}` tag must remain at top of file.
- **Go:** The header and footer are separate template files (`partials/header.html`, `partials/footer.html`). Update them in place. The skip-link goes in `layouts/base.html` before `{{template "header" .}}`.

### Testing Approach

No automated tests at this story (Playwright E2E comes in Epic 7). Verification is manual:
1. Start each stack and load `/` in a browser
2. Tab — first focus should land on the skip-link (visible when focused)
3. Press Enter on skip-link — focus moves to `<main id="main-content">`
4. Inspect DOM: verify landmark structure, FlashRegion attributes, heading hierarchy
5. Diff rendered HTML across stacks for byte-parity of chrome

### Project Structure Notes

Files created or modified:

| Action | Stack | File |
|--------|-------|------|
| UPDATE | shared | `fieldmark_shared/src/fieldmark.css` |
| UPDATE | shared | `fieldmark_shared/dist/fieldmark.css` (recompile) |
| UPDATE | .NET | `FieldMark/FieldMark.Web/Pages/Shared/_Layout.cshtml` |
| NEW | .NET | `FieldMark/FieldMark.Web/Pages/Shared/_FlashRegion.cshtml` |
| UPDATE | Django | `fieldmark_py/templates/base.html` |
| NEW | Django | `fieldmark_py/templates/_flash_region.html` |
| UPDATE | Go | `fieldmark-go/internal/web/templates/layouts/base.html` |
| UPDATE | Go | `fieldmark-go/internal/web/templates/partials/header.html` |
| UPDATE | Go | `fieldmark-go/internal/web/templates/partials/footer.html` |
| NEW | Go | `fieldmark-go/internal/web/templates/partials/flash_region.html` |

### Previous Story Intelligence

Stories 1.1 and 1.2 are done (scaffolds confirmed, SQL init scripts verified). Stories 1.3 (parity tooling) and 1.4 (design system CSS) are `ready-for-dev` but not yet implemented.

**Dependency note:** Story 1.4 adds Basecoat, semantic color tokens, fonts, and status-badge CSS. Story 1.5 depends on 1.4 for the compiled CSS foundation. If 1.4 is not yet complete when 1.5 starts, the focus/skip-link CSS additions should be compatible with whatever CSS state exists — they are additive rules that don't conflict with Basecoat integration.

**Commit convention from git history:** `feat: :sparkles: e1s{N} {description}` — use `feat: :sparkles: e1s5 cross-stack base layout with skip-link landmarks and flash-region`.

### References

- [Source: _bmad-output/planning-artifacts/architecture.md — D11 HTMX target ID inventory]
- [Source: _bmad-output/planning-artifacts/architecture.md — D12 Partial-naming convention]
- [Source: _bmad-output/planning-artifacts/epics.md — Story 1.5]
- [Source: _bmad-output/planning-artifacts/prd/web-app-specific-requirements.md — Accessibility WCAG 2.1 AA]
- [Source: _bmad-output/planning-artifacts/ux-design-specification.md — FlashRegion, focus management, skip-link]
- [Source: Root CLAUDE.md — Canonical HTMX Target IDs]
- [Source: FieldMark/CLAUDE.md — .NET stack rules]
- [Source: fieldmark_py/CLAUDE.md — Django stack rules]
- [Source: fieldmark-go/CLAUDE.md — Go stack rules]

## Dev Agent Record

### Agent Model Used

### Debug Log References

### Completion Notes List

### File List
