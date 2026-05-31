## Deferred from: code review rerun of 2-10-compliance-dashboard-with-portfolio-tiles.md (2026-05-31)

- **AG Grid `detail` mode silently drops row-click when `data-grid-target` is absent** — the `if (target)` guard in `ag-grid-panel.js` silently no-ops when `data-grid-target` is missing, with no `console.warn`. This is pre-existing behavior from Story 2.9. Add a console warning in the `else` branch when a future story modifies this file.
- **Go nil-pool pattern prevents authorized-200 integration test for `GET /dashboard`** — the `Pool: nil` test stub used in Go handler tests cannot reach `readStats`; an authorized-role 200 response is covered by the template test only. Address when the Go test harness gains a real Postgres pool.

## Deferred from: code review of 2-10-compliance-dashboard-with-portfolio-tiles.md (2026-05-31)

- **Go home chrome tests exercise dead test fixture** — `buildHomeApp` in `home_test.go` wires `pages/home` rendering rather than the redirect; tests pass but do not exercise the production `/` route. Refactoring the Go home test suite to use the real router is a larger task; address when the home-page test architecture is revisited.
- **Go nil-pool `/dashboard` branch returns empty HTTP 200** — `main.go` stub wires a no-op handler for `GET /dashboard` when `pool == nil`, returning 200 with no body. Consistent with the project's existing no-pool stub pattern for other routes; address if the stub-mode UX becomes a concern.
- **`make parity` route-dump check is a no-op** — the parity tooling was not scaffolded (Story 1.3 gap); `GET /dashboard` cannot be verified in all three stack route dumps until the tool lands.

## Deferred from: Story 2.7 — TabStrip component (2026-05-30)

- **Story 2.7-followup — TabStrip badge semantic monoculture**: The badge `aria-label` is hard-coded to `"<count> unread"`. If a future consumer needs a different semantic (e.g., a "high priority count" badge), the wrapper needs an additional `badge_aria_template` prop per stack. Currently deferred; add when a non-unread-count consumer lands. Track: add `badge_aria_template: string?` prop to all three stack wrappers and the canonical fixture.

## Deferred from: code review of 2-4-implement-phase-2-markup-only-components (2026-05-28)

- `StatusBadgeVM.Severity` field (`fieldmark-go/internal/web/viewmodels/components.go`) is dead exported state — no template or resolver reads it; acceptable for the markup-only story scope since resolution logic (`(Entity, Value, Severity)` → `(ClassName, Label)`) will live in the first handler that constructs a `StatusBadgeVM` from domain values (Story 2.10 / 2.11). Address when that handler is written.

## Resolved by Story 1.14 (2026-05-21)

All 2026-05-17 entries are accounted for below.

| Deferred entry | Resolution |
|---|---|
| Reduced-motion users see abrupt transitions | RESOLVED — AC1.1 / Task 2: global `@media (prefers-reduced-motion: reduce)` rule in `_a11y.css` |
| Unknown `badge-*` or `data-score-band` values render neutral only | RESOLVED — AC2.4 / Task 4: `.badge-unknown` CSS fallback; per-stack warning log + `"unknown"` token |
| Font files 404 / blocked → no visual regression test | RESOLVED — AC1.3 / Task 3: `font-display: swap` already present; Playwright font-fallback + CLS test added |
| Sidebar hidden or jumps when `[data-sidebar-initialized]` never set | RESOLVED — AC2.5 / Task 5: CSS default to visible + static; PE override in `_components.css`; Playwright no-JS test added |
| AG Grid empty/loading state not styled distinctly | RESOLVED — AC2.6 / Task 6: `.ag-overlay-loading-center` and `.ag-overlay-no-rows-center` rules in `_ag-grid.css` |
| Toaster accumulates unlimited toasts without height/scroll limit | RESOLVED — AC2.7 / Task 7: `.toaster { max-height: calc(5 × ...); overflow-y: auto }` in `_components.css`; Playwright visual regression test added |
| Tooltip `data-tooltip` with HTML entities or >container-xs clips silently | RESOLVED — AC2.8 / Task 8: `[data-tooltip]::before { max-width; white-space: normal; word-break; text-overflow }` in `_components.css`; per-stack escaping tests added; tooltip escaping rule documented in per-stack CLAUDE.md files |
| `@source` globs silently fail if stack directory renamed | RESOLVED — AC3.9 / Task 9: `scripts/check-sources.mjs` wired as `prebuild` |
| Basecoat 0.3.11 pre-1.0 class names may shift on minor upgrade | RESOLVED — AC4.13 / Task 12: `scripts/check-basecoat-classes.mjs` wired in `prebuild`; `docs/basecoat-upgrade-checklist.md` created |
| Forced-colors mode loses badge/score meaning (color-only) | RESOLVED — AC1.2 / Task 2: `@media (forced-colors: active)` block in `_components.css` with `forced-color-adjust: auto; border: 1px solid ButtonText` |
| optimize-css.mjs crashes on LightningCSS version bump / pnpm store change | RESOLVED — AC3.10 / Task 10: fatal warning types (`error`, `unsupported`) now exit non-zero |
| Script has zero error handling for missing input / permission issues | RESOLVED — AC3.10 / Task 10: missing input, directory input, empty input all exit non-zero with clear messages |
| In-place mutation without backup or dry-run mode | RESOLVED — already implemented (atomic `.tmp` → rename); verified in Task 10 tests |
| No logging or propagation of LightningCSS recovered errors | RESOLVED — AC3.10 / Task 10: all LightningCSS warnings logged to stderr; fatal types cause non-zero exit |
| Script assumes pnpm + ESM; breaks under npm/yarn or CommonJS | RESOLVED — AC3.11 / Task 11: `"packageManager": "pnpm@11.0.8"` + `preinstall` guard in `package.json` |
| Optimization step not mentioned in build docs or CLAUDE.md | RESOLVED — AC4.12 / Task 12: `docs/getting-started.md` CSS pipeline section added; root `CLAUDE.md` pointer added |
| Future Basecoat minor release may re-introduce unmergeable duplicates | RESOLVED — AC4.13 / Task 12: class smoke test + upgrade checklist; `build:raw` bypass documented |

---

## Deferred from: code review of 1-4-bootstrap-design-system-foundation-in-fieldmark-shared.md (2026-05-17)

- Reduced-motion users see abrupt sidebar/toast/tooltip transitions (prefers-reduced-motion)
- Unknown badge-* or data-score-band values render with default/neutral color only
- Font files 404 or blocked → only system fallback, no visual regression test
- Sidebar remains hidden or jumps if `[data-sidebar-initialized]` never set
- AG Grid empty/loading state not styled distinctly from Basecoat table
- Toaster accumulates unlimited toasts without height/scroll limit
- Tooltip `data-tooltip` with HTML entities or >container-xs text clips silently
- @source globs silently fail if any stack directory is renamed/moved
- Basecoat 0.3.11 (pre-1.0) class names may shift on minor upgrade
- High-contrast / forced-colors mode loses badge/score meaning (color-only)

## Deferred from: code review (rerun) of 1-4-bootstrap-design-system-foundation-in-fieldmark-shared.md (2026-05-17)

- optimize-css.mjs crashes on LightningCSS version bump or pnpm store change
- Script has zero error handling for missing input file or permission issues
- In-place mutation without backup or dry-run mode
- No logging or propagation of LightningCSS recovered errors
- Script assumes pnpm + ESM environment; breaks under npm/yarn or CommonJS
- Optimization step not mentioned in build docs or CLAUDE.md
- Future Basecoat minor release may re-introduce unmergeable duplicates
## Deferred from: code review of 1-14-harden-design-system-foundation-and-build-tooling-against-known-edge-cases.md (2026-05-22)

- Conflicting source-of-truth guidance points to stale planning artifacts (`CLAUDE.md` vs `docs/README.md` vs stale statements in planning artifacts). Deferred as pre-existing documentation governance debt.
- AGENTS pre-kickoff note is stale versus current parity/e2e scaffolding and can suppress verification behavior. Deferred as pre-existing docs debt.
- Stack CLAUDE files repeat stale planning-artifacts authority language, reintroducing drift risk. Deferred as pre-existing docs debt.

## Deferred from: Story 2.1 (2026-05-26)

- `make parity` routes diff is failing pre-existing Story 2.1 — Django and Fiber expose `/robots.txt` and `/.well-known/security.txt` but the .NET stack does not. Verified pre-existing by stashing 2.1 changes and re-running. Story 2.1 introduces zero new routes (AC4) so this is out of scope, but the routes-parity gate is currently red and needs a dedicated story to either land the two endpoints on .NET or formally exempt them from the diff.

## Deferred from: Story 2.4 (2026-05-28)

- Story 2.4-followup — unknown-token runtime warning logger per [component-edge-case-checklist.md §1](../../docs/reference/component-edge-case-checklist.md) canonical resolution; deferred from Story 2.4 per Dev Notes §"Decision — unknown-token handling". Story 2.4 shipped the user-visible fallback class and per-stack fallback assertions; request-scoped operator logging remains follow-up work. Extended by Story 2.5 to also cover ComplianceTile out-of-range scores (< 0 or > 100) — same per-stack resolution pattern applies (no-data variant rendered, no log this story).
