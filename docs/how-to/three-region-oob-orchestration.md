# How-To: Three-Region OOB Orchestration

> **Status:** live — populated by Story 2.12, 2026-05-31.

This recipe is the canonical pattern for HTMX mutations that update three regions in one response: the primary entity detail region, the header `#compliance-tile`, and the `#audit-log` prepend target. Each stack composes the response natively; there is no shared template fragment.

The first implementation is `POST /projects/<id>/place-on-hold` and `POST /projects/<id>/resume`.

---

## When to Use This Pattern

Use this pattern when a server-side domain mutation changes the entity detail view and must also refresh adjacent status surfaces in the same user-visible paint.

For Project transitions, the response updates:

1. `#project-detail` — the main detail body after the status transition.
2. `#compliance-tile` — re-rendered out-of-band, even when the score is unchanged.
3. `#audit-log` — an out-of-band prepend containing the new audit row.

Do not use this pattern for read-only tab swaps, validation-only form errors, or unauthorized responses.

---

## Successful Response Composition

The initiating form posts with:

```html
<form
  hx-post="/projects/<id>/place-on-hold"
  hx-target="#project-detail"
  hx-swap="innerHTML">
</form>
```

The successful response body contains the main detail inner-content partial plus exactly two OOB regions:

```html
<section id="project-header-strip">
  <!-- StatusBadge reflects the new project status. -->
  <section id="compliance-tile" role="status" aria-live="polite" aria-atomic="true">
    <!-- In-band tile copy inside the refreshed detail body. -->
  </section>
</section>

<section class="project-detail-grid">
  <!-- Summary panel, flipped action-button trichotomy, tabs, empty rail. -->
</section>

<section id="compliance-tile" hx-swap-oob="true" role="status" aria-live="polite" aria-atomic="true">
  <!-- Same compliance score, structural refresh in the header target. -->
</section>

<li hx-swap-oob="afterbegin:#audit-log">
  <!-- AuditRow for ProjectPlacedOnHold or ProjectResumed. -->
</li>
```

The main `#project-detail` wrapper is not part of the response body and is never OOB. The request target owns that swap and preserves the outer wrapper in both contexts:

- Standalone detail page: the page template owns `<div id="project-detail">...</div>`.
- List-embedded detail rail: the list page owns `<aside id="project-detail">...</aside>`.

The compliance tile and audit row are the only OOB regions, so `hx-swap-oob` count is `2`.

`#audit-log` is emitted even though Story 2.13 owns the live Audit tab target. Until that target exists in the DOM, HTMX drops the OOB fragment. This is intentional; Story 2.12 proves audit correctness at the data layer and response-shape level.

---

## Negative Cases

Error responses must not emit OOB swaps. They either return the canonical forbidden body, the reason form, or the current detail body with an InlineAlert.

| Outcome | HTTP | Response body | OOB regions |
|---|---:|---|---:|
| Success | 200 | main `#project-detail` partial + OOB `#compliance-tile` + OOB `#audit-log` prepend | 2 |
| Unauthorized | 403 | canonical forbidden body with no entity details | 0 |
| Invalid transition | 409 | current `#project-detail` body + InlineAlert, status unchanged | 0 |
| Validation failure | 422 | reason form fragment with `aria-invalid` / `aria-describedby` | 0 |

The 409 path catches only the typed project-transition exception. Generic exceptions must surface as failures, not as user-facing 409 responses.

---

## Per-Stack Native Implementations

- **.NET** — `Pages/Projects/PlaceOnHold.cshtml`, `Resume.cshtml`, `Detail.cshtml.cs`, `_ProjectTransitionForm.cshtml`, and `_DetailTransitionResponse.cshtml`.
- **Django** — `projects/views.py`, `projects/urls.py`, `_project_transition_form.html`, and `_detail_transition_response.html`.
- **Go** — `internal/web/handlers/projects_transition_handler.go`, `cmd/web/main.go`, `projects_transition_form.html`, and `projects_transition_response.html`.

All three stacks use the same logical order: authorize, validate, open one transaction, load aggregate, call domain method, append audit row in the same transaction, persist status, commit, render the three-region response.

---

## Conformance Test Contract

Each stack should cover these response-shape checks:

1. Success: response contains main `#project-detail`, OOB `#compliance-tile`, OOB `#audit-log`, and exactly two `hx-swap-oob` attributes.
2. 403: response contains zero `hx-swap-oob` attributes and no project fields.
3. 409: response contains current `#project-detail`, an InlineAlert `role="alert"`, and zero `hx-swap-oob` attributes.
4. 422: response contains the reason form field with `aria-invalid="true"` and zero `hx-swap-oob` attributes.

Preferred helper: a small per-stack parser/helper that counts `hx-swap-oob` in the rendered body. Keep it local to each stack's test suite; do not add a shared parser package.

---

## Timing Contract

The target NFR is local-dev p95 <= 200 ms per stack with cross-stack divergence <= 50 ms p95. If the timing harness from the Epic 1 retrospective action A5 is unavailable, record measured local timings in the story sign-off and note the harness dependency instead of blocking implementation.
