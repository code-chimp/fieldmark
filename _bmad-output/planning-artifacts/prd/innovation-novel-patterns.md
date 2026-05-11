# Innovation & Novel Patterns

## Normalizing the Pragmatic-Islands Pattern

FieldMark stakes a position that the official HTMX project supports but does not heavily teach: **rich third-party JavaScript controls (AG Grid, Kendo Grid, charting libraries, rich text editors, etc.) belong inside an HTMX architecture as accepted, vanilla patterns — not as exceptions, escape hatches, or apologies.**

This is not a claim of originality. The official HTMX documentation ships first-class mechanisms for the pattern: the [`json-enc`](https://v1.htmx.org/extensions/json-enc/) extension for JSON encoding, the [`htmx.process()`](https://htmx.org/docs/) hook for initializing dynamically-added DOM (including content placed by third-party JavaScript), and dedicated examples for [Web Components](https://htmx.org/examples/web-components/) and AlpineJS. The [Two Approaches To Decoupling](https://htmx.org/essays/two-approaches-to-decoupling/) essay further endorses splitting an application API (hypermedia) from a data API (JSON), which implicitly permits both to coexist in the same application. The architectural permission is unambiguous.

What the official material _doesn't_ heavily emphasize is the **rich-island integration pattern at scale** — using a mature data grid as a load-bearing UI component while keeping the architecture server-authoritative. That gap is reasonable: introductory teaching focuses on hypermedia mechanics, where HTMX is most differentiated. But it leaves enterprise teams evaluating HTMX without a worked example of the integration they care about most, which can lead to the wrong conclusion that their dependency on tools like AG Grid disqualifies the architecture.

The contribution FieldMark makes is not novelty; it's making this pattern boring enough to copy. The architectural argument should compete on delivered business value, not on the absence of well-supported third-party tooling.

### What this position actually says

- A JavaScript island for a data grid, chart, calendar, or editor is **a valid, prevalent, real-world pattern** that does not invalidate a server-driven architecture.
- The enforceable architectural rule isn't "no JavaScript" — it's "**no business logic, no source-of-truth state, no workflow orchestration on the client**." A grid that loads server-side rows, fires HTTP requests on selection, and never owns business rules respects the rule. A client component that re-implements domain logic in JavaScript violates it, regardless of which library it uses.
- JSON endpoints supporting a grid are a different contract for a different consumer. Hypermedia is the wire format for human-facing partials; structured JSON is the wire format for machine-facing consumers (grids, charts, third-party integrations). Both can live in the same application without contradiction — and the HTMX project's own essays endorse this split.
- The line between "acceptable island" and "creeping SPA" is enforceable and worth enforcing: declarative grid configuration, server-fed data, no client-side row computation, no client-owned filter or sort logic the server doesn't also know about, row interactions delegate back to HTMX for detail rendering.

### Why FieldMark takes the position explicitly

The teams most likely to benefit from server-driven architecture are enterprise teams already invested in heavy backend stacks (.NET, Java, Python, Rails) and dependent on mature data-grid tooling. For those teams, "you must give up AG Grid to adopt HTMX" is a non-starter conclusion — and one they may reach on their own without seeing the alternative demonstrated end-to-end. FieldMark's contribution is to put a worked example on record.

## Validation Approach

The position is validated within the artifact in two ways:

1. **The integration itself.** AG Grid is integrated identically across all three stacks using the server-side row model. Row selection drives HTMX detail-panel loads. The grid never owns business state. The architectural rules do not bend to accommodate it.
2. **The audience-objection test.** AG Grid (or its equivalents — Kendo, TanStack Table, ag-Charts) is predicted to come up in nearly every Q&A because it represents the most common real-world objection. If the answer — here's the JSON contract, here's how it stays an island, here's why this is fine — closes the objection rather than triggering more, the position holds.

## Risk Mitigation

The principal risk is **island creep**: grid configuration grows over time to include client-side filters, computed columns, custom row logic, or selection-driven UI updates that the server doesn't render. Mitigations:

- AG Grid configuration is reviewed under the same architectural rules as backend code, not waved through as "frontend stuff."
- Any client-side logic beyond declarative grid configuration and HTMX-trigger glue is a defect requiring justification.
- E2E tests assert the grid drives detail panels via HTMX (not client-side state), so regressions are caught at the cross-stack parity boundary.
- The PRD-binding architectural constraints are explicit on this: JavaScript outside AG Grid wiring and minimal UX glue is a defect across all three stacks.
