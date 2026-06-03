# Story 2.15: Harden Epic 2 — consolidated deferred-work pass

Status: backlog (SHELL — populated 2026-06-03; manual-AC-test findings to be appended after Story 2.13)

Epic: 2 — Project Lifecycle & Compliance Dashboard
Pattern precedent: [Story 1.14](1-14-harden-design-system-foundation-and-build-tooling-against-known-edge-cases.md) — Epic 1's consolidated hardening pass (17 deferred items dispatched in one disciplined story). This story is the Epic 2 equivalent.
Source: [deferred-work.md](deferred-work.md) — the Epic 2 entries (2.1, 2.4, 2.7, 2.10, 2.11, 2.12, 2.14) plus the parity correction block at the top.

## Purpose

One disciplined hardening pass over the ~50 deferred items that accumulated during Epic 2 (mostly from the 2.11/2.12 review marathons), **plus** the findings from Tim's manual AC testing of the application after Story 2.13 lands. Mirrors 1.14: resolve known debt in one consolidated story rather than dribbling it across Epic 3.

**Sequencing:** authored after 2.13; **manual-AC-test findings (Section H) are added before dev starts**, so the manual pass and the deferred backlog are hardened together. Lands before the Epic 2 retrospective.

> **This is a shell.** Sections A–G enumerate the known deferred items grouped by theme. Section H is the placeholder for manual-AC-test findings. The cross-stack symmetry rule applies: a fix in one stack must be mirrored in the other two unless the item is explicitly stack-specific.

---

## In scope — grouped by theme (each box = one deferred item resolved)

### Group A — Transaction / concurrency correctness (project transitions)
_Highest value. Mostly "no current observable defect" but latent audit/data-integrity risk._
- [ ] **D-2.12-G2R-1** — `updated_at` not set on status transition (all 3 stacks). **Recommended fix: PostgreSQL `BEFORE UPDATE` trigger on `domain.project`** (schema-layer, resolves all three without cross-stack code changes — but note this is a `domain` schema change via `docker/postgres/init/`, requires `make reset`, and must update `canonical-pg-indexes.txt` if indexes change). Confirm approach before dev.
- [ ] **D-2.12-G1-1** — Django stale `before_state` snapshot (capture from the locked instance inside the txn).
- [ ] **D-2.12-G1-2** — Django `_render_transition_conflict` reloads project without lock.
- [ ] **D-2.12-G2R-3** — .NET `before/after` serialization asymmetry (read `after` symmetrically from `afterState.RootElement.GetRawText()`).
- [ ] **D-2.12-G2R-5** — Go `ComplianceTileOOB` uses pre-commit in-memory score (use the `buildVM`-reloaded score).
- [ ] **D-2.12-G2-1** — .NET `AuditRow.OccurredAt` sampled after `CommitAsync` (capture before commit, pass through).

### Group B — Reason-handling consistency
- [ ] **D-2.12-G2R2-1** — Django `place_on_hold` passes raw reason, `resume` passes trimmed (normalize to trimmed).
- [ ] **D-2.12-G2R-4** — Go passes untrimmed reason to domain methods (`strings.TrimSpace`).
- [ ] **D-2.12-G2-2** — control-char regex `[\x00-\x1F\x7F]` rejects `\t`/`\n`/`\r` (all 3). **Product decision required:** keep single-line-only (document the constraint in form UI) **or** allow multi-line reasons (relax the regex). Decide, then make all three stacks identical.

### Group C — Cross-stack divergences (parity drift the route/index tool can't catch)
- [ ] **D-2.11-R3-3** — 403 body is custom plain text, not the canonical Story 1.11 403 page (all 3). Align if a canonical 403 template is adopted — **decide whether to introduce one** (touches several stories' 403 paths).
- [ ] **D-2.12-G2-4** — Go `InlineAlert` rendered after the `<textarea>`; .NET/Django render it before. Align position (422 re-render visual parity).
- [ ] **D-2.12-G3-1** — Django transition response inlines compliance-tile markup vs .NET partial / Go template call (no runtime defect; document or unify).

### Group D — Defensive guards / latent footguns (cheap, mechanical)
- [ ] **D-2.12-G3R-1** — Go `{{ if .Alert }}` → `{{ if .Alert.Title }}` (zero-value struct truthiness).
- [ ] **D-2.12-G3R-2** — .NET `Html.Raw` in `_ProjectTransitionForm.cshtml` → two nullable Razor attribute expressions.
- [ ] **D-2.12-G3-2 / D-2.11-R3-8** (dup) — .NET `_DetailBody.cshtml` hardcodes `ActiveIndex = 0` → pass `Model.ActiveTabIndex`.
- [ ] **D-2.11-R3-7** — Go `buildVM` nil-project guard after `LoadWithRelations`.
- [ ] **D-2.11-3** — Go `projects_detail_tab_response.html` add defensive `{{else}}` for unknown `PanelTemplate`.
- [ ] **D-2.12-G2R3-1** — Go `buildVM` `fiber.ErrNotFound` not caught by post-commit guard (defensive; unreachable today).
- [ ] **D-2.11-1** — .NET `IsTabResponse` flag set before tab validation (reorder).
- [ ] **D-2.11-R3-4** — Django `{% include panel_template %}` — add "must be a hardcoded literal" comment.
- [ ] **D-2.12-G1-DomainMethods-1** — Django `place_on_hold`/`resume` add "does not call save()" docstring.
- [ ] **D-2.11-4** — .NET `ProjectActionPredicateTests` reflection-sets `Status` → factory constructor overload.
- [ ] **D-2.10 (AG Grid)** — `ag-grid-panel.js` silently drops row-click when `data-grid-target` absent → add `console.warn` in the `else`. _(Directly relevant to Story 3.4a, which reuses this panel.)_

### Group E — Test hardening (coverage gaps)
- [ ] **D-2.12-G4-3** — assert `CountOobRegions == 0` explicitly on 403/409 paths.
- [ ] **D-2.12-G4-1** — add `GET /resume` 403 + .NET GET 403 tests.
- [ ] **D-2.12-G4-2** — Go resume success integration test.
- [ ] **D-2.11-R3-6** — no-role user + non-HTMX GET on tab URL → assert 403 (all 3).
- [ ] **D-2.11-R2-3** — strengthen `.NET HtmxMode` conformance ARIA coverage (`role="region"`, `aria-live`, OOB-absent).
- [ ] **D-2.11-5** — assert `#compliance-tile` present **exactly once** (`ContainSingle`).
- [ ] **D-2.11-R3-1** — Go status-mutation tests add `t.Cleanup` for `stubProjectStatus`.
- [ ] **D-R2** — .NET sub-nav `NotContain` scoped to `href="..."`.
- [ ] **D1** — Go reference-page call-counter per-handler isolation (or document the cumulative-sequential assumption).

### Group F — Dead code & efficiency
- [ ] **D-R1** — delete dead `fieldmark_py/reference/templates/reference/_subnav.html`.
- [ ] **D2** — remove unreachable `actor == nil` guard in Go `admin_reference.go`.
- [ ] **D-2.11-R3-2** — .NET Summary VM built before the redirect check (move load after guard).
- [ ] **D-2.11-R3-5** — Django `prefetch_related` bypassed by `.values_list()` (iterate `.all()`).
- [ ] **D-2.12-G2-3** — Django double project load on POST (eliminate the outer non-locking load).
- [ ] **D-2.12-G2R-2** — Django conflict-path N+1 (`prefetch_related` on the `select_for_update`). _(Also concurrency — see Group A.)_

### Group G — Parity hardening: robots.txt / security.txt symmetry
- [ ] **D-2.1 (resolve properly)** — land `GET /robots.txt` and `GET /.well-known/security.txt` on the **.NET** stack to match Django + Go, **then delete `filter_ignored_non_business_routes()` from `tools/parity/diff-routes.sh`** so `make parity` verifies all three serve them. Converts the permanent exemption into a stronger gate. _(Verified 2026-06-03: parity is currently green at 24 routes / 21 indexes **with** the exemption; after this change it should stay green **without** it, now covering the two routes.)_

### Section H — Manual AC-test findings (TO BE POPULATED after Story 2.13)
> _Placeholder. After 2.13 lands, Tim runs a manual AC pass across the application; findings are enumerated here as `MT-1, MT-2, …` with the same fix-in-all-three-stacks discipline. This section is expected to be the largest single source of items (mirrors how 1.14 absorbed late findings)._
- [ ] _MT-1 …_

---

## Explicitly OUT of scope (do NOT pull these in)

- **D-2.12-G2R3-2 — concurrent-deletion-after-commit returns a silent empty 404.** Needs *new* cross-stack behavior (`HX-Redirect` to the list or an `HX-Trigger` toast). **→ its own dedicated story**, not a patch.
- **All E2E / `javaScriptEnabled:false` / Playwright items** (D-2.11-6, D-2.11-7, D3, the 2.10 e2e gaps). **→ Epic 7** E2E build-out.
- **Go nil-pool test-harness items** (D-2.11-2, the 2.10 nil-pool gaps, Go home dead-fixture). Blocked on the Go integration harness gaining a real Postgres pool — **→ verify Epic 1 retro action A3 status; its own enabling story if A3 didn't land for Go.**
- **Docs-governance trio** (deferred-work lines for 1.14: stale source-of-truth language in `CLAUDE.md`/`AGENTS.md`/stack CLAUDEs). **→ tech-writer (Paige); this is retro action A2 follow-up**, not code hardening.
- **D-2.11-R2-1 — inspector silently dropped for deleted auth user.** Same cross-stack-identity problem as the `inspector_name` decision in **Story 3.4a** — **resolve there (Epic 3)**, not here.
- **D-2.11-R2-4** (autofocus fallback — only if cross-browser issue emerges), **D-2.12-G3-3** (tab-strip OOB refresh — only if tab badges are added), **2.7-followup** (TabStrip badge prop — only when a non-unread consumer lands). Defer until their trigger condition exists.
- **D-R3** — Django Story 2.3 `category.trade_type` unguarded — pre-existing 2.3 gap; fold in only if trivial, else track on 2.3.

---

## Acceptance criteria (shell)

- [ ] Every checked item in Groups A–G is resolved, with the fix mirrored across all three stacks unless explicitly stack-specific.
- [ ] All Section H manual-test findings resolved.
- [ ] `make test-net`, `make test-django`, `make test-go` all green.
- [ ] `make parity` green **with the robots/security exemption removed** (Group G) — routes now include `/robots.txt` + `/.well-known/security.txt` on all three.
- [ ] Group A schema change (if the `updated_at` trigger is chosen): `canonical-pg-indexes.txt` refreshed if needed; `make reset` documented in the story sign-off.
- [ ] Each resolved item's `deferred-work.md` entry annotated "Resolved by Story 2.15" (or moved to a Resolved section), mirroring the 1.14 resolution table.
- [ ] Product decisions recorded: Group B multi-line-reason call; Group C canonical-403 call.

## Open product/architecture decisions to resolve before/at dev start
1. **Group A** — `updated_at`: schema trigger (recommended) vs per-stack code. Schema change implies `make reset`.
2. **Group B** — single-line-only reasons (document) vs allow multi-line (relax regex).
3. **Group C** — introduce a canonical 403 template (touches multiple stories) vs leave plain-text 403 and just document the convention.
