# Web App Specific Requirements

## Project-Type Overview

FieldMark is a server-rendered multi-page web application with HTMX-driven partial updates and AG Grid as a scoped JavaScript island. It is not a single-page application, not a progressive web application, and explicitly rejects SPA architecture. Server-side rendering produces the canonical HTML for every screen; HTMX `hx-get` / `hx-post` requests return HTML partials for in-page updates; AG Grid loads server-side rows via JSON endpoints. There is no client-side routing, no client-side state management, no service worker, and no offline mode.

## Browser Support Matrix

| Browser           | Supported versions                                                |
| ----------------- | ----------------------------------------------------------------- |
| Chrome            | Last 2 stable versions                                            |
| Firefox           | Last 2 stable versions                                            |
| Safari            | Last 2 stable versions                                            |
| Edge              | Last 2 stable versions (Chromium)                                 |
| Internet Explorer | Not supported (deprecated)                                        |
| Mobile browsers   | Best-effort; not the primary target — see Responsive Design below |

The browser matrix applies identically across all three stacks. A feature that works in Chromium but not Safari is a defect, not an acceptable tradeoff.

## Responsive Design

The application is **desktop-first responsive**. Tailwind CSS v4 is the styling layer (`fieldmark_shared/` is the sole CSS source, compiled to `dist/fieldmark.css` and symlinked into all three apps). Tailwind's default breakpoint utilities are used directly — no custom breakpoint scheme.

- **Primary target:** desktop browsers, ≥ 1280px width. All workflows are designed and tested at this width.
- **Secondary target:** tablet (≥ 768px). Layouts collapse via Tailwind's standard `md:` / `lg:` breakpoints; AG Grid remains usable; HTMX partials render correctly. Site supervisors using a tablet on-site are a credible use case.
- **Tertiary target:** mobile (< 768px). The site does not break, but UX is not optimized. AG Grid in particular is acknowledged to be a poor mobile experience by design — server-side row models in narrow viewports are a known constraint; this is consistent with the brief's out-of-scope on mobile-native.

Responsive layout is delivered through Tailwind utilities applied during template authoring, not through media-query authoring or per-stack CSS. The compiled CSS is identical across all three stacks.

## Performance Targets

Already defined in §Success Criteria — Measurable Outcomes. Summary:

- HTMX partial-swap perceived latency: ≤ 200 ms p95 (local dev).
- AG Grid row selection → detail panel: ≤ 300 ms p95.
- Compliance score tile updates in the same round trip as the triggering action.
- No full-page reload on any state-changing action.

These targets apply identically across all three stacks. Cross-stack latency divergence > 50ms p95 on the same scenario is a defect requiring investigation.

## SEO Strategy

**Not applicable.** FieldMark is an internal enterprise application behind authentication. There is no public surface, no content discoverability, no sitemap, and no SEO concern. Pages do not need to be indexable, render server-side metadata for crawlers, or support OpenGraph / Twitter Card markup. Authentication is required for every route except a single public landing page (which carries no business content).

## Accessibility (WCAG 2.1 AA Target)

FieldMark targets **WCAG 2.1 Level AA conformance** across all three stacks. This is treated as a first-class technical and discipline requirement, not an afterthought. The rationale is twofold: (1) enterprise customers — the kind of teams the architecture is being demonstrated to — frequently have accessibility procurement requirements, and an HTMX architecture that fails accessibility undercuts the adoption argument; (2) the demo audience includes people who do this for a living and will notice. More fundamentally, accessibility is a baseline obligation: the people who depend on adaptive technology are real users, and technical teams do well to adopt that mindset by default rather than retrofit it.

### Enforcement strategy

The challenge: server-side rendering stacks (Razor, Django templates, Go `html/template`) do not have a native equivalent of React's `eslint-plugin-jsx-a11y`. Static template-level a11y linting is uneven across these ecosystems. The lever that works _cross-stack_ and _cheaply_ is runtime accessibility scanning against rendered HTML.

**Primary enforcement — axe-core via Playwright.**

- Every Playwright E2E scenario runs an `@axe-core/playwright` accessibility scan as an assertion against the rendered page or partial.
- Scans run identically across all three stacks. Same scenario, same axe ruleset, same expected outcome.
- WCAG 2.1 AA rules are enabled; specific known-impossible rules (if any emerge) are explicitly disabled with a documented rationale, never silently ignored.
- A new accessibility violation introduced by any stack is a build-blocking defect at the cross-stack parity boundary.

**Secondary enforcement — discipline and review.**

- Templates use semantic HTML matched to the _kind of interaction_, not by reflexive defaults. State-changing actions ("I am pressing a control") use `<button>`. Navigation requests use `<a>`. Form submissions use `<form>`. Row-or-card selectors that trigger detail loads — when the underlying element isn't naturally interactive — carry `role="button"`, `tabindex="0"`, an accessible name, and Enter/Space keyboard handlers (AG Grid provides this; hand-rolled grids must do it explicitly). HTMX-driven elements with no user-initiated trigger (`hx-trigger="revealed"`, `hx-trigger="every Ns"`) have no a11y operability requirement and use whatever element is semantically appropriate to the content. The principle: HTMX chooses the _server-communication mechanism_; WCAG governs the _element semantics_. The two decisions are independent and both must be made deliberately.
- All interactive elements are keyboard-operable. Tab order matches visual order. Focus is visible.
- Color contrast meets AA at 4.5:1 for normal text, 3:1 for large text. Tailwind's default palette is reviewed at component-design time; deviations are documented.
- Form errors are announced via `aria-invalid` and `aria-describedby` linking to error messages, not just color.
- Images have `alt` attributes. Decorative images use `alt=""` deliberately.
- PR review treats a11y violations as code defects, not nice-to-haves.

### HTMX-specific accessibility considerations

HTMX's swap behavior introduces accessibility concerns that don't exist in traditional MPAs and are easy to overlook. The list below is the floor, not the ceiling — additional patterns may surface during implementation and should be added here:

- **Focus management on swaps.** When HTMX replaces a region, focus does not automatically move to it. For screen reader users, an in-place swap can pass unnoticed. FieldMark's convention: after an HTMX swap that represents a meaningful state change (e.g., violation resolved, corrective action submitted), the swapped partial includes an element with `tabindex="-1"` and `autofocus` — or the server returns an `HX-Trigger` event that runs a small documented focus-shift script. Each pattern in use is documented in the stack reference docs.
- **Live regions for out-of-band swaps.** The compliance score tile and any other `hx-swap-oob` target updates without user action on that region. These regions carry `aria-live="polite"` (or `aria-live="assertive"` where appropriate) so screen readers announce the change. Every OOB swap site is documented per the architectural constraint already established.
- **Loading and disabled states.** Buttons that trigger HTMX requests use `hx-disabled-elt="this"` (or equivalent) so the disabled state is announced to assistive technology during the request. A spinner that is purely visual is insufficient.
- **HTTP 409 error rendering.** When a domain rule rejects an action and the server re-renders the partial with an error, the error message is associated with the relevant control via `aria-describedby` and is announced via a live region on first render.

### What we deliberately do not do

- We do not aim for WCAG 2.1 AAA — the cost is high and the audience for a teaching artifact does not require it.
- We do not aim for full Section 508 procurement compliance documentation; the architecture supports it, but the artifact does not produce VPATs.
- We do not run manual screen-reader testing in CI. Manual testing happens at major milestones; runtime axe-core scanning catches the common defects.

## Implementation Considerations

- **HTMX and AG Grid versions are pinned identically across all three stacks.** Version mismatch is a build-blocking defect (per root `README.md`). Upgrades are a coordinated three-stack story, not a per-stack maintenance task.
- **Tailwind compiled CSS is committed to the repository.** No build step is required after cloning to run any stack. `fieldmark_shared/dist/fieldmark.css` is the artifact; CSS authoring happens against `fieldmark_shared/src/fieldmark.css`.
- **No client-side build pipeline beyond Tailwind.** No webpack, no Vite, no Turbopack, no esbuild for application code. AG Grid is loaded from its distributed bundle. HTMX is loaded from its distributed file. The only "build" anyone runs locally is Tailwind, and only when CSS source changes.
- **Authentication is required on every route** except a single public landing page (no business content). Per ADR-012, each stack implements auth in its own framework-local schema; landing redirects to the framework's login page.
- **No service worker, no offline mode, no PWA install prompt.** These are explicitly out of scope per the brief.
