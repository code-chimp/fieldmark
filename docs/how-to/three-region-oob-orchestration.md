# How-To: Three-Region OOB Orchestration

> **Status:** skeleton (scaffolded by Epic 1 retrospective action item A4, 2026-05-25).
> Content is populated by **Story 2.12** (Place-on-Hold / Resume transitions).

This recipe describes the canonical pattern for HTMX responses that update **three regions in a single round trip**: the primary entity partial, the header `#compliance-tile` (OOB), and the `#audit-log` (OOB). Each stack implements the recipe natively; per-stack conformance tests assert response shape.

See the root [CLAUDE.md](../../CLAUDE.md) **Cross-Stack Architecture Principle** for why this lives as documentation rather than as a shared template fragment.

---

## When to Use This Pattern

TODO (Story 2.12): document the trigger condition — domain mutations on a Project (or future aggregate) that affect:

1. The entity's own detail view, AND
2. The portfolio-level compliance tile, AND
3. The audit log.

The canonical first use is `POST /projects/<id>/place-on-hold` and `POST /projects/<id>/resume`.

---

## Successful Response Composition

TODO (Story 2.12): document the response structure with a worked example.

```html
<!-- Primary partial: re-renders the entity -->
<div id="project-detail">
  TODO: project detail markup after mutation
</div>

<!-- OOB swap: compliance tile (header strip) -->
<section id="compliance-tile" hx-swap-oob="true" role="status" aria-live="polite" aria-atomic="true">
  TODO: re-rendered tile (score may be unchanged; structural re-render still required)
</section>

<!-- OOB swap: audit log prepend -->
<tr id="audit-log" hx-swap-oob="afterbegin">
  TODO: new audit row prepended
</tr>
```

- **HTMX target:** the request specifies `hx-target="#project-detail"` and `hx-swap="innerHTML"` (or equivalent).
- **OOB regions:** declared with `hx-swap-oob` per HTMX convention.
- **Single round trip:** all three regions ship in one response body; no follow-up requests.

---

## Negative Cases (Critical — UX-DR22)

TODO (Story 2.12): document the rule that error responses **MUST NOT** emit OOB swaps.

| Outcome | HTTP | Response body | OOB regions |
|---|---|---|---|
| Success | 200 | main partial + OOB `#compliance-tile` + OOB `#audit-log` | **3** |
| Unauthorized | 403 | TODO (no entity state leakage per FR7) | **0** |
| Rule violation (e.g., already on hold) | 409 | main `#project-detail` re-rendered with *current* state + InlineAlert | **0** |
| Validation failure | 422 | form partial with `aria-invalid` | **0** |

---

## Per-Stack Native Implementations

- **.NET** — TODO: Razor partial composition, `Response.Headers["HX-Trigger"]` usage if needed.
- **Django** — TODO: template `{% include %}`/`{% extends %}` composition.
- **Go** — TODO: `html/template` block composition.

No shared template fragment, no symlinked partial.

---

## Conformance Test Contract

Each stack ships a conformance test that exercises three flows:

1. **Success** — asserts the response contains exactly the three documented regions (main `#project-detail` + OOB `#compliance-tile` + OOB `#audit-log`).
2. **403** — asserts the response contains zero OOB regions.
3. **409** — asserts the response re-renders the main partial with current state plus InlineAlert, with zero OOB regions.

TODO (Story 2.12): define the per-stack test locations and any shared fixture (a parser that counts OOB regions from a response body, perhaps).

---

## Timing Contract (NFR1)

TODO (Story 2.12 + retro action item A5): the orchestration must complete with local-dev p95 ≤ 200 ms per stack and cross-stack divergence ≤ 50 ms p95. The timing harness (A5) measures this.
