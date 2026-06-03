# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Repository Is

FieldMark is a construction compliance and inspection management system implemented **across three parallel stacks** — .NET (Razor Pages + HTMX), Django (Templates + HTMX), and Go (Fiber + HTMX) — against a shared PostgreSQL 17 database. It is a teaching artifact demonstrating server-authoritative architecture as an alternative to SPAs. A story is never done until all three stacks pass it.

See [docs/tutorials/getting-started.md](docs/tutorials/getting-started.md) for infrastructure, quick start, and setup. See [docs/explanation/architecture.md](docs/explanation/architecture.md) for full architecture, patterns, and domain details. CSS pipeline (two-step build, pre-build checks, pnpm guard, Basecoat upgrade procedure): [fieldmark_shared/CLAUDE.md](fieldmark_shared/CLAUDE.md) and [docs/tutorials/getting-started.md#2-build-shared-css-optional](docs/tutorials/getting-started.md).

The `docs/` folder is organized by the Diátaxis framework: `tutorials/` (learning-oriented), `how-to/` (problem-oriented), `reference/` (information-oriented), `explanation/` (understanding-oriented). Pick the quadrant that matches the reader's need.

## Hard Rules (all stacks)

See [docs/reference/hard-rules.md](docs/reference/hard-rules.md). Stack-specific rules live in each project's own `CLAUDE.md`.

## Playwright E2E Rule

When an agent runs Playwright from [`e2e/`](e2e/), treat browser launch as a likely sandbox boundary on macOS-hosted agent runners.

- If a Playwright command fails during browser startup with a permission error such as `Permission denied (1100)`, `MachPortRendezvousServer`, or another browser bootstrap denial, do **not** debug selectors first.
- Re-run the same Playwright command with escalation / outside the sandbox so Chromium can launch normally.
- Prefer **single-project reruns** (`--project=dotnet|django|fiber`) when tests mutate shared DB state; reset or reseed known fixture rows between project runs if the scenario is stateful.
- Do not assume switching from `devices["Desktop Chrome"]` to explicit `browserName: "chromium"` will remove sandbox failures. In this repo the practical issue is browser-launch permission, not the Playwright browser preset.

See [e2e/README.md](e2e/README.md) for the concrete runbook.

## Cross-Stack Architecture Principle

Each stack (.NET, Django, Go) is a self-contained, idiomatic application. A native developer opening one stack must see every enum, DTO, DAO, handler, and test in its expected location with no surprises.

**Shared only via symlink:** the compiled design-system bundle (`fieldmark_shared/dist/fieldmark.css`) and vendored static assets (`htmx.min.js`, `ag-grid-enterprise.min.js`, fonts). That is the full list. (AG Grid **Enterprise** is used to demonstrate the true Server-Side Row Model; the demo runs without a license key and the "unlicensed" watermark is an accepted tradeoff.)

**Cross-stack invariants live as documentation contracts**, not as shared code:
- A spec page under `docs/reference/` (data contracts: audit actions, AG Grid SSRM wire format, role names, canonical HTMX target IDs, form field names for cross-stack forms) or `docs/how-to/` (orchestration patterns: three-region OOB, canonical request flow).
- A native implementation in each stack — idiomatic to that stack's framework.
- A per-stack conformance test asserting the native implementation matches the documented contract.

**Form-contract corollary:** when a form appears in ≥2 stacks (login, project-create, place-on-hold, corrective-action submit), the canonical field names, hidden-input names, and return-target conventions must appear in the story AC list or in a contract doc. Per-stack drift between template field names and handler bindings is a recurring Epic 1 bug class.

**Anti-patterns:**
- Extracting cross-stack constants into a shared package, generated stubs, or a symlinked manifest file.
- A shared template engine, partial, or component fragment.
- Any artifact that requires a developer working in one stack to read files in another stack to understand their own code.

This principle was ratified in the Epic 1 retrospective (2026-05-25). It overrides any earlier guidance that suggested otherwise.

**Split-story corollary (ratified 2026-06-03):** Large behavioral/UI-integration stories may be split into a three-part group — **`.a` reference (.NET)**, **`.b` port (Django + Go)**, **`.c` parity & definition-of-done**. Within such a group, `.a` and `.b` are legitimately "done" in a deliberately stack-divergent state; the *"a story is never done until all three stacks pass it"* invariant and `make parity` are asserted on the **`.c` story**, which carries the cross-stack conformance tests, byte-identical snapshots, E2E, and NFR-timing gates. The dependency chain `.a → .b → .c` is hard: `.b` does not start until `.a` is reviewed clean. Small single-transition stories and pure data-layer mapping stories stay unified. See [docs/how-to/cross-stack-story-splitting.md](docs/how-to/cross-stack-story-splitting.md).

## Key Reference Documents

- [docs/explanation/architecture.md](docs/explanation/architecture.md) and [docs/tutorials/getting-started.md](docs/tutorials/getting-started.md) — canonical docs (progressive disclosure from this file).
- `_bmad-output/planning-artifacts/architecture.md` — architectural source of truth (decisions, patterns, structure, validation)
- `_bmad-output/planning-artifacts/prd/` — product requirements source of truth (sharded; index at `prd/index.md`)

The `_bmad-output/planning-artifacts/research/` folder contains pre-kickoff priming material. It is not maintained and is not authoritative.
