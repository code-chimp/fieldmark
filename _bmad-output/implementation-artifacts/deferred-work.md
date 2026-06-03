# Status corrections (2026-06-03, John/PM — verified during Epic 2 hardening triage)

> Corrections to earlier deferred entries that have since been resolved or were inaccurate. Kept as a record so the same misunderstanding does not recur. Originals left in place below for history.

- **CORRECTION — D-2.10 "`make parity` route-dump is a no-op / tooling not scaffolded" is STALE/INACCURATE.** `tools/parity/` is fully scaffolded and functional: `diff-routes.sh` (dumps all three stacks, normalizes, pairwise `.NET vs Django` + `.NET vs Fiber` diff with non-zero exit on drift) and `diff-pg-indexes.sh` (live `domain.*` `pg_indexes` vs committed `canonical-pg-indexes.txt`). The routes tool landed in Story 1.3 (`e1s3 establish tools parity`) and was extended in Story 2.3 (`e2s3 expose read api`); `dump-routes-net.sh` was wired the same day (2026-05-31) the D-2.10 note was written, which likely explains the reviewer's observation. `make parity` is the gate for every Epic 3+ `.c` story (cross-stack story-split policy, 2026-06-03) and **does** enforce. Verified by live run on 2026-06-03 (see story 2.15 / sprint notes). **No parity-tooling story needed.**
- **CORRECTION — D-2.1 "routes-parity gate is red because .NET lacks `/robots.txt` + `/.well-known/security.txt`" is RESOLVED via exemption.** `diff-routes.sh` contains `filter_ignored_non_business_routes()` which strips `^(get|options) /(robots\.txt|\.well-known/security\.txt)$` from all three stacks before diffing (added in the Story 2.3 commit `cc6ba23`). The D-2.1 note's "either land the endpoints on .NET or formally exempt them" choice was taken as **exempt**, so the gate is **not red** on this account. Follow-up (tracked in Story 2.15): land the two routes on .NET to match Django+Go and **remove the exemption filter** so parity verifies them too — converting a permanent carve-out into a stronger gate.

---

## Deferred from: code review of 2-12 Group 4 tests (2026-06-03)

- **D-2.12-G4-1 — `GET /resume` 403 and .NET GET 403 tests absent**: Auth path is shared with GET place-on-hold which is tested; risk low. AC4 technically covers GET too but the auth middleware is shared.
- **D-2.12-G4-2 — Go missing resume success integration test**: The place-on-hold integration test exercises the full handler code path; resume is symmetric. P5 coverage gap.
- **D-2.12-G4-3 — AC4/AC9: 403 tests never assert `CountOobRegions == 0` explicitly**: Plain-text 403 body (`"You do not have permission to access this page."`) makes OOB structurally impossible; invariant is guaranteed by the response type. Explicit conformance assertion would be mechanical.

## Deferred from: code review of 2-12 Group 3 rerun (2026-06-03)

- **D-2.12-G3R-1 — Go `{{ if .Alert }}` truthy for zero-value struct**: `html/template` treats any non-nil struct as truthy. If a future caller injects a zero-value `InlineAlertVM{}` (rather than omitting the key), the template renders an empty alert box. No current caller does this; no observable defect. Fix: change check to `{{ if .Alert.Title }}`.
- **D-2.12-G3R-2 — .NET `Html.Raw` in `_ProjectTransitionForm.cshtml`**: Injects a fixed developer-controlled attribute string (no user data, no XSS risk). File is a page partial, not a `Pages/Shared/Components/` component, so the "no `Html.Raw`" rule's scope is ambiguous. Idiomatic fix: replace with two nullable Razor attribute expressions (`aria-invalid="@(condition ? "true" : null)"` and `aria-describedby="@(condition ? "reason-error" : null)"`).

## Deferred from: code review of 2-12 Group 3 templates (2026-06-03)

- **D-2.12-G3-1 — Django `_detail_transition_response.html` inlines compliance tile markup**: The `_compliance_tile.html` component emits the `<section id="compliance-tile">` wrapper itself; placing it inside another `<section hx-swap-oob>` would double-nest. The inline approach with `{% compliance_band %}` template tag is the correct workaround given this constraint. Diverges from .NET (partial) and Go (template call) but no runtime defect.
- **D-2.12-G3-2 — .NET `_DetailBody.cshtml` hardcodes `ActiveIndex = 0` in TabStrip call**: `LoadDetailAsync(id, null, ct)` always sets `ActiveTabIndex = 0` on the success path so no current observable defect. Latent if `_DetailBody` is reused in a non-summary-tab context. Fix: pass `Model.ActiveTabIndex` instead of literal `0`.
- **D-2.12-G3-3 — .NET success response does not emit tab-strip OOB refresh**: After a transition, `#project-detail-tabstrip` in the DOM is not refreshed. No badge counts wired, no current defect. Would need a fourth OOB region if tab badges are ever added.

## Deferred from: code review of 2-12 Group 2 rerun 3 (2026-06-03)

- **D-2.12-G2R3-1 — Go `buildVM` tab-switch `fiber.ErrNotFound` not caught by post-commit guard**: `buildVMWithLoadedProjectData` returns `fiber.ErrNotFound` from its `default:` arm on unknown `:tab` parameter values. The post-commit guard in `postTransition` only checks `postgres.ErrProjectNotFound`. This code path is unreachable today because `postTransition` calls `buildVM` with no `:tab` segment. Fiber would send 404 via its `*fiber.Error` handler regardless. Theoretical defensive gap only.
- **D-2.12-G2R3-2 — Successful commit + concurrent deletion returns empty 404 (cross-stack design gap)**: After `tx.Commit` succeeds, if `buildVM` cannot find the project (concurrent deletion), all three stacks return an empty 404 with no user feedback. The HTMX client silently no-ops. Resolution requires a cross-stack story to return `HX-Redirect` to the projects list or an `HX-Trigger` toast notification on this path.

## Deferred from: code review of 2-12 Group 2 rerun 2 (2026-06-02)

- **D-2.12-G2R2-1 — Django internal asymmetry: `place_on_hold` passes raw reason, `resume` passes trimmed**: `project_place_on_hold` calls `project.place_on_hold(reason)` with the raw form value; `project_resume` calls `project.resume(reason.strip() or None)`. Domain methods currently ignore the parameter so no observable defect. Same root class as D-2.12-G2R-4 (Go untrimmed). Fix: change to `project.place_on_hold(reason.strip())`.

## Deferred from: code review of 2-12 Group 2 rerun (2026-06-02)

- **D-2.12-G2R-1 — All three stacks: `updated_at` not set on status transition**: Go's `UPDATE` sets only `status`; Django's `save(update_fields=["status"])` skips `updated_at`; .NET's `SaveChangesAsync` relies on `ValueGeneratedOnAdd` (write-once). Story 2.8 precedent has the same behavior and passed review; no AC requires it. Schema-layer fix (PostgreSQL `BEFORE UPDATE` trigger on `domain.project`) would resolve without cross-stack code changes.
- **D-2.12-G2R-2 — Django conflict path N+1 queries**: Inside `transaction.atomic()`, `Project.objects.select_for_update().get(pk=id)` has no `prefetch_related`. If `InvalidProjectTransition` is raised, `_render_transition_conflict` → `_build_project_detail_context` fires lazy SELECTs for trade scopes and inspectors outside the rolled-back transaction (different MVCC snapshot). Data is still correct; inefficiency only. Fix: add `.prefetch_related("trade_scopes", "inspector_assignments")` to the `select_for_update` call.
- **D-2.12-G2R-3 — .NET `BeforeAfterJson.after.status` serialization asymmetry**: In the success path, `before` is deserialized from `beforeState.RootElement.GetRawText()` but `after` is constructed from the `ProjectStatus` page-model property populated by `LoadDetailAsync`. Values are identical today but the two sides use different serialization paths. Fix: read `after` symmetrically from `afterState.RootElement.GetRawText()`.
- **D-2.12-G2R-4 — Go passes untrimmed reason to domain methods**: `project.PlaceOnHold(reason)` and `project.Resume(reason)` receive the raw `c.FormValue("reason")` string; .NET passes the already-trimmed `rawReason`. Domain methods currently discard the parameter (`_ = reason`), so no current defect. Latent contract drift if methods ever use the parameter. Fix: pass `strings.TrimSpace(reason)` to both calls.
- **D-2.12-G2R-5 — Go `m["ComplianceTileOOB"]` uses pre-commit in-memory score**: `components.NewComplianceTileArgs(&project.ComplianceScore, ...)` references the pre-commit `project` variable, not the freshly-reloaded project from `buildVM`. Django and .NET reload the project before building the OOB tile. hold/resume do not change `compliance_score`, so no current observable defect. Fix: use the score from the `buildVM` result map's loaded project.

## Deferred from: code review of 2-12 Group 2 handlers + routes (2026-06-02)

- **D-2.12-G2-1 — .NET `AuditRow.OccurredAt` sampled after `CommitAsync`**: The `OccurredAt` and `Absolute` fields in the success view-model are set to `DateTimeOffset.UtcNow` after the DB commit completes. The audit row in the database uses whatever timestamp `IAuditAppender.Append` assigns internally (before `SaveChangesAsync`). The two timestamps may diverge by the commit round-trip time (~1–10 ms). Cosmetic display skew only; the DB row is authoritative. Fix: capture the timestamp before the commit and pass it through to the view model.
- **D-2.12-G2-2 — Control-char regex `[\x00-\x1F\x7F]` rejects embedded newlines/tabs (all three stacks)**: The reason validation regex rejects `\t` (0x09), `\n` (0x0A), and `\r` (0x0D), which are control characters per the spec. Multi-line reasons (e.g., pasted from a document) will fail with "invalid control characters." Spec says "reject control characters" — this is spec-compliant behavior. Document as a known single-line-only constraint in the form UI. Revisit if product requirements change to allow multi-line reasons.
- **D-2.12-G2-3 — Django double project load on POST**: `project_place_on_hold` and `project_resume` POST paths call `_load_project_or_404(id)` before the `transaction.atomic()` block (non-locking, for early 404) and then re-fetch with `select_for_update()` inside the block. The outer load result is discarded. Redundant DB round-trip. The P5 fix (initialize `before_state = {}` before the `with` block and rely on the locked inner load for 404) can eliminate the outer load.
- **D-2.12-G2-4 — Go InlineAlert position in transition form**: `projects_transition_form.html` renders the InlineAlert after the `<textarea>`, while .NET `_ProjectTransitionForm.cshtml` and Django `_project_transition_form.html` render it before the label+textarea. Cross-stack visual parity difference on 422 re-renders. Deferred to Group 3 template review.

## Deferred from: code review of domain entity diff (2026-06-02)

- **D-2.12-G1-DomainMethods-1 — Django `place_on_hold` / `resume` lack "no save" docstring**: Both model methods mutate `self.status` in memory without calling `self.save()`. Persistence is intentionally the handler's responsibility (`views.py` calls `project.save(update_fields=["status"])` after each call). Without a docstring note, a future developer could call the method directly and believe the change was persisted. Add a one-line docstring to each method noting "Does not call save(); caller is responsible for persistence." Deferred — documentation gap only, no code defect; consistent with project pattern for other domain model methods.

## Deferred from: code review of 2-12-place-on-hold-and-resume-transitions-with-three-region-oob-orchestration.md — Group 1 domain layer (2026-06-02)

- **D-2.12-G1-1 — Django stale `before_state` snapshot in place-on-hold and resume views**: `before_state = {"status": project.status}` is captured from the pre-lock, pre-transaction fetch in both `project_place_on_hold` and `project_resume` (views.py). Inside the `transaction.atomic()` the row is re-fetched with `select_for_update()`. A concurrent transition between the two fetches could produce a stale `before_state` in the audit entry. Fix: capture `before_state` from the locked instance inside the transaction. Deferred to handler review (Group 2).
- **D-2.12-G1-2 — Django `_render_transition_conflict` reloads project without lock**: The 409 conflict render calls `_load_project_or_404(id)` (a plain non-locking read) outside any transaction after the `InvalidProjectTransition` exception is caught. A concurrent transition between the exception and this re-load could produce a 409 body showing a third state. Low-probability but confusing. Deferred to handler review (Group 2).

## Deferred from: code review round 3 of 2-11-project-detail-anchor-screen-with-header-strip-tabstrip-and-entityrail.md (2026-06-01)

- **D-2.11-R3-1 — Go status-mutation tests lack `t.Cleanup` on `stubProjectStatus`**: `TestGetProjectsDetail_AdminClosedAllButtonsDisabled` and related tests mutate `stubProjectStatus` after `makeProjectsDetailApp` without a `t.Cleanup`. Currently safe (sequential execution; `makeProjectsDetailApp` resets on entry). Address if `-shuffle` or `t.Parallel` is introduced.
- **D-2.11-R3-2 — `.NET` Summary VM built before redirect check**: `OnGetAsync` loads the full project + trades + inspectors + auth users before the `if (!isHtmx && tab is not null) return Redirect(...)` guard fires. No correctness defect; wasted DB work on non-HTMX tab redirect. Refactor the load to happen after the redirect check if handler performance becomes a concern.
- **D-2.11-R3-3 — 403 body is custom text, not canonical Story 1.11 403 page**: All three stacks return plain-text `"You do not have permission to access this page."` rather than the Story 1.11 canonical 403 HTML page. AC6 specifies HTTP 403 status code only; body shape is not AC-required. Consistent with prior story patterns. Align if a canonical 403 template is ever introduced.
- **D-2.11-R3-4 — Django `{% include panel_template %}` variable path**: `_tab_response.html` uses `{% include panel_template %}` where `panel_template` is a view-set string literal. Safe today (never sourced from request input), but fragile if refactored. Add a code comment documenting "must always be a hardcoded literal" when next touching this template.
- **D-2.11-R3-5 — Django `prefetch_related` bypassed by `.values_list()` in `_build_project_detail_context`**: `project.trade_scopes.values_list(...)` and `project.inspector_assignments.values_list(...)` create new querysets that bypass the `prefetch_related` cache — two extra DB hits per request. Fix: iterate `.all()` to use the cache, then extract values in Python. Address if request latency becomes a concern.
- **D-2.11-R3-6 — No test for no-role user + non-HTMX GET on tab URL**: The R2 authz-before-redirect fix is correct but the combined unauthorized + non-HTMX path is untested. Add `no_role_user` hitting `/projects/{id}/tabs/violations` without HX-Request and asserting 403 in all three stacks.
- **D-2.11-R3-7 — Go `buildVM` nil project not guarded after `LoadWithRelations`**: If `LoadWithRelations` returns `(nil, nil, nil, nil, nil)` (nil project, nil error — a store contract violation), `project.Code` panics. Add `if project == nil { return nil, postgres.ErrProjectNotFound }` guard when next touching `buildVM`.
- **D-2.11-R3-8 — `.NET` `_DetailBody.cshtml` hardcodes `ActiveIndex = 0`**: The TabStrip is always rendered with Summary as active on full-page load, per Decision 6. Correct for current scope; potential footgun if a future story needs to render a non-Summary tab as active on initial load (e.g., a tab deep-link URL scheme).

## Deferred from: code review round 2 of 2-11-project-detail-anchor-screen-with-header-strip-tabstrip-and-entityrail.md (2026-06-01)

- **D-2.11-R2-1 — Inspector silent drop for deleted auth user (cross-stack)**: When a `domain.project_inspector.user_id` has no matching row in `dotnet_auth.users` / `django_auth.auth_user` / `fiber_auth.users` (deleted user), the inspector is silently omitted from the Summary panel list with no fallback display or operator log. Address when user-lifecycle management / delete handling is in scope.
- **D-2.11-R2-2 — `_SummaryPanel` no `#project-action-form` slot**: Story 2.12 Decision 1 designates adding this slot as Task 0 in that story. Epic-sanctioned; no action needed on Story 2.11.
- **D-2.11-R2-3 — `HtmxMode` test ARIA coverage weak**: The `.NET` `HtmxMode` conformance test asserts `id="violation-detail"` is present but does not verify `role="region"` or `aria-live` on the rail, nor that `hx-swap-oob` is absent from the main response. Address when next touching the conformance tests.
- **D-2.11-R2-4 — `autofocus` on OOB-swapped panel reliability**: `autofocus` is the chosen focus-move mechanism per UX-DR31. Spec-compliant and working in HTMX 4.0-beta2 + Chromium. If cross-browser issues emerge, add `HX-Trigger: {"focusPanel": true}` + JS listener fallback.

## Deferred from: code review of 2-11-project-detail-anchor-screen-with-header-strip-tabstrip-and-entityrail.md (2026-06-01)

- **D-2.11-1 — `IsTabResponse` flag set before tab validation (.NET)**: `Detail.cshtml.cs` sets `IsTabResponse = tab is not null` before the `NotFound()` guard fires on invalid tab values. No current defect (guard fires correctly), but the ordering is fragile to future refactoring that bypasses the guard. Address when next touching `Detail.cshtml.cs`.
- **D-2.11-2 — Go nil-pool `loadInspectorNames` silent swallow**: `if len(inspectorIDs) == 0 || h.Pool == nil { return out, nil }` silently returns empty inspector names with no error or log when `Pool` is nil. Pre-existing nil-pool pattern throughout Go test suite. Address when Go test harness gains a real Postgres pool.
- **D-2.11-3 — Go `projects_detail_tab_response.html` no fallback for unknown `PanelTemplate`**: Four `{{if eq .PanelTemplate "..."}}` blocks with no `{{else}}` branch — silently emits empty panel + OOB tabstrip if `PanelTemplate` drifts. Handler switch guards it today; add a defensive `{{else}}` when next touching this template.
- **D-2.11-4 — .NET `ProjectActionPredicateTests` uses reflection to set `Status`**: `typeof(Project).GetProperty("Status")!.SetValue(project, status)` will break silently if `Status` becomes `init`-only. Use a factory constructor overload when next touching the entity.
- **D-2.11-5 — `compliance-tile` present-once not asserted**: AC1 states `#compliance-tile` must appear exactly once. Tests verify presence but not uniqueness. Low risk on this read page; add a `ContainSingle` assertion when next touching the conformance tests.
- **D-2.11-6 — `javaScriptEnabled:false` E2E test absent (Task 7.3)**: Playwright Chromium environment constraint prevented running the no-JS test. Must land on a CI lane that can launch Chromium. Add as part of Epic 7 E2E build-out.
- **D-2.11-7 — E2E scenario + `make parity` unverified (Task 8)**: Same Playwright environment constraint. Five new routes need parity-tool verification; add to Epic 7 E2E sprint.

## Deferred from: code review rerun of 2-14-reference-data-read-pages-for-administrator.md (2026-05-31)

- **D-R1 — Dead `_subnav.html` partial**: `fieldmark_py/reference/templates/reference/_subnav.html` contains the 4-link self-referential nav and is no longer referenced by any template (inline sub-navs replaced it). Accidentally `{% include %}`ing it in a future template would silently reintroduce the AC5 self-link bug. Delete the file.
- **D-R2 — .NET sub-nav `NotContain` assertion scope**: `AdminReferenceCatalogPagesTests.cs` line 51 uses `html.Should().NotContain("/admin/reference/trade-types")` against the full body. A tighter assertion scoped to `href="..."` would be more precise and resilient to future layout changes that might include the URL in a meta tag or breadcrumb.
- **D-R3 — Django `index.html` (Story 2.3) `category.trade_type` unguarded**: The Story 2.3 overview page `violation_categories` column renders `{{ category.trade_type }}` without `|default_if_none:""`. Explicitly out of scope for Story 2.14 per spec Decisions note 1; track as a pre-existing gap in Story 2.3.

## Deferred from: code review of 2-14-reference-data-read-pages-for-administrator.md (2026-05-31)

- **D1 — Go call-counter assertion is cumulative across 3 sub-tests**: `store.tradeCalls/categoryCalls/ruleCalls` are asserted collectively after all three route tests run. Correct for sequential execution and handlers are obviously single-purpose, but does not strictly prove per-handler isolation and would race if sub-tests were ever parallelized. Address if sub-test parallelism is ever introduced or if a per-handler isolation test is required by a future AC.
- **D2 — `actor == nil` dead code in Go handlers**: `auth.ActorFromCtx` always returns non-nil (falls back to `app.Anonymous()`); the nil guard in `TradeTypesIndex`, `ViolationCategoriesIndex`, and `ComplianceRulesIndex` is unreachable. Remove when next touching `admin_reference.go` to avoid misleading future maintainers.
- **D3 — AC7 cat 3 Playwright `javaScriptEnabled:false` test absent**: Pages are purely server-rendered and degrade correctly with JS off. The test coverage assertion is missing but no production defect exists. Add when the e2e suite covers admin reference pages.

## Deferred from: code review rerun of 2-10-compliance-dashboard-with-portfolio-tiles.md (2026-05-31)

- **AG Grid `detail` mode silently drops row-click when `data-grid-target` is absent** — the `if (target)` guard in `ag-grid-panel.js` silently no-ops when `data-grid-target` is missing, with no `console.warn`. This is pre-existing behavior from Story 2.9. Add a console warning in the `else` branch when a future story modifies this file.
- **Go nil-pool pattern prevents authorized-200 integration test for `GET /dashboard`** — the `Pool: nil` test stub used in Go handler tests cannot reach `readStats`; an authorized-role 200 response is covered by the template test only. Address when the Go test harness gains a real Postgres pool.

## Deferred from: code review of 2-10-compliance-dashboard-with-portfolio-tiles.md (2026-05-31)

- **Go home chrome tests exercise dead test fixture** — `buildHomeApp` in `home_test.go` wires `pages/home` rendering rather than the redirect; tests pass but do not exercise the production `/` route. Refactoring the Go home test suite to use the real router is a larger task; address when the home-page test architecture is revisited.
- **Go nil-pool `/dashboard` branch returns empty HTTP 200** — `main.go` stub wires a no-op handler for `GET /dashboard` when `pool == nil`, returning 200 with no body. Consistent with the project's existing no-pool stub pattern for other routes; address if the stub-mode UX becomes a concern.
- **`make parity` route-dump check is a no-op** — the parity tooling was not scaffolded (Story 1.3 gap); `GET /dashboard` cannot be verified in all three stack route dumps until the tool lands.

## Deferred from: Story 2.7 — TabStrip component (2026-05-30)

- **Story 2.7-followup — TabStrip badge semantic monoculture**: The badge `aria-label` is hard-coded to `"<count> unread"`. If a future consumer needs a different semantic (e.g., a "high priority count" badge), the wrapper needs an additional `badge_aria_template` prop per stack. Currently deferred; add when a non-unread-count consumer lands. Track: add `badge_aria_template: string?` prop to all three stack wrappers and the canonical fixture.

## Resolved by Story 2.11 (2026-06-01)

- `StatusBadgeVM.Severity` deferred item from Story 2.4 is resolved by introducing `ResolveStatusBadge(entity, value string)` and using it in the Go Project Detail handler (`fieldmark-go/internal/web/viewmodels/components.go`, `fieldmark-go/internal/web/handlers/projects_detail_handler.go`).

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

## Deferred from: code review of 2-13-project-audit-log-tab (2026-06-04)

- D-2.13-D1 — **.NET `latestAudit` fetched between `SaveChangesAsync` and `CommitAsync`**: `ProjectDetailPageModelBase.cs` reads the just-written audit entry inside the open transaction (after `SaveChangesAsync`, before `CommitAsync`) and uses its `OccurredAt`/`Id` to build the load-more cursor. If `CommitAsync` subsequently fails, the cursor points at a rolled-back row; because the exception propagates and no response is sent, there is no observable defect today. Pre-existing data-commit pattern carried from Story 2.12; hardening story 2.15 is the appropriate venue to harden the transaction boundary.
