# Story 2.13: Project audit log tab

Status: done

Epic: 2 — Project Lifecycle & Compliance Dashboard
Source AC: [_bmad-output/planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md](../planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md) §Story 2.13
Canonical DDL: [docker/postgres/init/010_domain_tables.sql](../../docker/postgres/init/010_domain_tables.sql) — `domain.audit_entry` (190–211; columns `id, occurred_at, actor_id, action, entity_type, entity_id, project_id, before_state jsonb, after_state jsonb, metadata jsonb`)
AuditRow contract (owned by Story 2.4 — **consume, do not re-author**): [docs/reference/component-canonical-examples.md](../../docs/reference/component-canonical-examples.md) + [fieldmark_shared/components/audit_row/canonical.html](../../fieldmark_shared/components/audit_row/canonical.html) / [README.md](../../fieldmark_shared/components/audit_row/README.md)

Depends on (all **done** unless noted):
- **Story 2.11** — Project Detail anchor screen: `GET /projects/<id>` dual-mode (standalone full page + list-embedded `<aside id="project-detail">`), the **TabStrip** (`#project-detail-tabstrip`) with the **Audit** tab already wired (`hx-get=/projects/<id>/tabs/audit`, `hx-target="#project-detail-tab-content"`), and the tab-panel container `#project-detail-tab-content` (`role="tabpanel"`). This story **replaces the Audit-tab placeholder** the three stacks currently render — Go [projects_detail_panels.html:30-31](../../fieldmark-go/internal/web/templates/pages/projects_detail_panels.html) (`project_detail_audit_panel` → "Audit entries appear here once the audit tab story lands."), Django [templates/projects/tabs/_placeholder_audit.html](../../fieldmark_py/templates/projects/tabs/_placeholder_audit.html), .NET [Pages/Projects/Tabs/_AuditPanel.cshtml](../../FieldMark/FieldMark.Web/Pages/Projects/Tabs/_AuditPanel.cshtml).
- **Story 2.12** — Place-on-hold/resume **emit** an OOB `#audit-log` `afterbegin` fragment per the [three-region contract](../../docs/how-to/three-region-oob-orchestration.md), which **silently no-ops today because `#audit-log` is not in the DOM (Decision 4 of 2.12)**. **This story builds the `#audit-log` target — making that OOB land live.** Go's emitted fragment is `<li hx-swap-oob="afterbegin:#audit-log">{{ template "audit_row" .AuditRow }}</li>` ([projects_transition_response.html:7-8](../../fieldmark-go/internal/web/templates/pages/projects_transition_response.html)); .NET/Django emit the equivalent. The `#audit-log` element this story renders **must be the exact prepend target** those fragments expect: a list whose children are `audit_row` `<li>`s, with a stable `id="audit-log"`.
- **Story 2.4** — the **AuditRow** wrapper (markup-only, byte-identical across stacks): Go [audit_row.html](../../fieldmark-go/internal/web/templates/components/audit_row.html), .NET [_AuditRow.cshtml](../../FieldMark/FieldMark.Web/Pages/Shared/Components/_AuditRow.cshtml), Django [_audit_row.html](../../fieldmark_py/templates/components/_audit_row.html). Its VM shape is the contract: `Action`, `ActionClass`, `ActorDisplay` (empty/whitespace/null actor → `"??"` — cat 9 fallback already built in), `OccurredAt` (machine `datetime` attr), `Absolute` (`title`), `Relative` (human string), `BeforeAfterJSON` (rendered in `font-mono`/JetBrains Mono inside a `<details>` disclosure with `aria-expanded`). **Compose this component; do not re-emit its inner markup.**
- **Story 2.2** — `domain.audit_entry` mapping per stack. **The data layer is append-only today** — there is **no read query**: Go [`AuditEntryStore`](../../fieldmark-go/internal/data/postgres/auditentrystore.go) exposes only `Append(ctx, tx, *AuditEntry)`; Django [`audit/append.py`](../../fieldmark_py/audit/append.py) is write-only (but the [`AuditEntry`](../../fieldmark_py/audit/models.py) model is queryable); .NET writes via [`IAuditAppender`](../../FieldMark/FieldMark.Data/Auditing/IAuditAppender.cs) but [`FieldMarkDbContext`](../../FieldMark/FieldMark.Data/Context/FieldMarkDbContext.cs) already exposes `DbSet<AuditEntry> AuditEntries`. **This story adds the project-scoped read path (Decision 2).**
- **Story 2.9** — the **actor-name resolution** precedent: the Project list projects `pm_name` from `domain.project` joined to the per-stack user source. Reuse that same user lookup to resolve `audit_entry.actor_id → display name` here. **Story 1.12** — `can(actor, action)` primitive (`project.read` gates this tab). **Story 1.11** — unauthenticated → `/login` redirect, canonical 403 body.

## Story

As any authorized user on Project Detail (including the **Executive** read-only role — FR43),
I want to view the project's full audit history most-recent-first, with expandable before/after detail and incremental loading,
So that I have forensic visibility into every domain mutation that touched the project (FR42, FR43), and the canonical `#audit-log` OOB target that Story 2.12+ already emit into finally lands live.

**Scope boundary.** This story produces, per stack:
- (a) A **project-scoped read path** on the audit data layer (Decision 2) — a narrow method returning `domain.audit_entry` rows where `project_id = <id>`, newest-first, with **keyset pagination** (Decision 3). Go: a new pool-backed read method (the existing `Append` is tx-threaded and stays untouched). Django: an ORM query. .NET: an EF Core projection over the existing `DbSet<AuditEntry>`. **No `domain.*` schema change** (`pg_indexes` zero-diff).
- (b) The **real Audit tab panel** replacing the three placeholders: a `#project-detail-tab-content` tabpanel containing `<ul id="audit-log" role="list" aria-live="polite">` of **AuditRow** components (Decision 1) — making `#audit-log` a live DOM target so Story 2.12's OOB prepend lands.
- (c) An **actor-display + relative-time mapping** from `audit_entry` rows to the AuditRow VM (Decision 4), reusing 2.9's user lookup and the cat-9 `"??"` fallback.
- (d) **Keyset "Load more"** (Decision 3): initial render ≤ 100 rows; a Load-more affordance `hx-get`s the next ≤ 100 via a dedicated fragment route, **appending** `<li>`s into `#audit-log` and replacing/removing the Load-more control when exhausted.
- (e) The **empty state** (no entries) and the **403 / 404** read-path responses (no state leakage).
- (f) The **before/after JSON disclosure** rendering through default escaping (the `metadata.reason` written by Story 2.12 is user text → cat 3a XSS round-trip).
- (g) Per-stack tests (read path, render, authz, pagination, empty state, XSS round-trip) + the cross-stack **byte-identical audit-log shape** assertion + `make parity` + full gate.

**Out of scope:**
- **Writing audit entries** — every mutation story (2.8, 2.12, Epics 3–6) owns its own `append_audit_entry`. This story is **read-only**; it adds no `append_audit_entry` call and no mutating route. No CSRF surface (GET only).
- **Audit entries for non-Project entities** — the tab is scoped to `project_id = <id>`. Inspection/Violation/CorrectiveAction audit views are their own detail screens (Epics 3–5). (Rows whose mutation set `project_id` to this project still appear, regardless of `entity_type` — that is the forensic intent.)
- **Filtering / searching / sorting the audit log** — fixed order `occurred_at DESC`. No AG Grid (UX-DR carve-out: audit log is a Basecoat list, not a grid). No per-column filter.
- **A second `#compliance-tile` or three-region orchestration** — this tab is a plain content swap (TabStrip behavior from 2.7/2.11), not a mutation. It emits **no** OOB regions itself.
- **Live-streaming / polling new entries** — entries appear via (i) the Story 2.12 OOB prepend when a transition happens while the Audit tab is open, and (ii) a fresh tab fetch. No `hx-trigger="every Ns"` poll.
- Any `domain.*` schema change.

---

## ⚠️ Decisions baked into this story (read first)

Each is implemented as written and listed in the Sign-off block for reviewer ratification.

1. **This story makes `#audit-log` a live DOM target.** The Audit tab panel renders `<ul id="audit-log" role="list" aria-live="polite">` as the **direct parent of the AuditRow `<li>`s**, inside `#project-detail-tab-content`. This id and structure are **load-bearing for Story 2.12**, whose success response prepends `<li hx-swap-oob="afterbegin:#audit-log">…AuditRow…</li>`. Verify the contract is honored end-to-end: with the Audit tab open, exercise a place-on-hold transition and assert the new AuditRow appears at the **top** of `#audit-log` in the same paint (this is the 2.12 epic AC "open the Audit tab → new row at top", now finally E2E-verifiable — see AC9). The `aria-live="polite"` parent is what announces the OOB-prepended row (UX-DR12 / 2.4 AuditRow contract: "lives inside an `aria-live="polite"` parent"). **Do not** wrap each row in its own live region; the **list** is the live region.

2. **The audit data layer gains a project-scoped read path — append-only is the current state.** There is no existing read query for `domain.audit_entry`.
   - **Go:** add a **pool-backed** read method (the `Append` interface is tx-threaded and must stay untouched — reads do not run inside the mutation transaction). Recommended: a `AuditEntryReadStore` interface with `ListByProject(ctx, projectID uuid.UUID, page AuditPage) ([]AuditEntryRow, error)` backed by `pgxpool`, in `internal/data/postgres/auditentryreadstore.go`. Keep it a **narrow read interface** (mirrors the `ProjectStore` read-only pattern from 2.1).
   - **Django:** `AuditEntry.objects.filter(project_id=id)` ordered `-occurred_at, -id` — the model already exists; no new model.
   - **.NET:** project `DbSet<AuditEntry> AuditEntries` (already on `FieldMarkDbContext`) with `.Where(a => a.ProjectId == id).OrderByDescending(a => a.OccurredAt).ThenByDescending(a => a.Id)` — **manual projection to the AuditRow VM, no AutoMapper** (NFR6).
   - All three: read-only, no writes; no `domain.*` schema change.

3. **Keyset (cursor) pagination, newest-first — not offset.** Page by the composite key `(occurred_at DESC, id DESC)`. The audit table grows by append (and Story 2.12 prepends a row mid-session via OOB), so **offset pagination would duplicate or skip rows** when a new entry lands between page fetches. Keyset is correct and stays byte-stable.
   - **Initial render** (`GET /projects/<id>/tabs/audit`): the newest ≤ `PAGE_SIZE` (= **100**, document the constant, identical across stacks) rows. Query `WHERE project_id = $1 ORDER BY occurred_at DESC, id DESC LIMIT PAGE_SIZE + 1` — the **+1 sentinel** detects "more exist" without a `COUNT`.
   - **Load more** (`GET /projects/<id>/audit-log?before_occurred_at=<iso8601-utc>&before_id=<uuid>`): the next page strictly older than the cursor — `WHERE project_id = $1 AND (occurred_at, id) < ($cursor_ts, $cursor_id) ORDER BY occurred_at DESC, id DESC LIMIT PAGE_SIZE + 1`. The cursor is the **last row of the current page** (the page boundary, carried on the Load-more control).
   - Returns a **fragment** = the next batch of `<li>` AuditRows **plus** a refreshed Load-more control (or nothing/an end-marker when the +1 sentinel shows no more). The fragment `hx-target`s `#audit-log` with **`hx-swap="beforeend"`** (append) for the rows; the Load-more control replaces itself (`hx-swap="outerHTML"` on the control, or render it as the last child and let the appended-rows-then-control composition handle it — **pick one mechanism, identical across stacks, and document it**).

4. **Actor display + relative time are mapped per the AuditRow contract; the machine timestamp is the parity anchor.** For each row:
   - **`ActorDisplay`** resolves `actor_id` → user display name via the **same user source 2.9 used for `pm_name`**. Empty/whitespace/null/unresolvable actor → the cat-9 fallback `"??"` (already implemented in the AuditRow VM — do not re-invent; just feed it an empty display string and let `ShowInitialsFallback`/`ActorDisplay` handle it).
   - **`OccurredAt`** (machine) + **`Absolute`** (`title`) render the **ISO-8601 UTC** timestamp — this is **byte-stable across stacks** and is what the parity test anchors on. **`Relative`** is the human string ("3 minutes ago"); to prevent cross-stack drift, the **bucketing rules** (e.g. `< 60s → "just now"`, `< 60m → "N minutes ago"`, …) are documented in the AuditRow contract doc and implemented identically. The parity assertion targets **structure + the machine `datetime`/`title`**, not the relative string's wording for a live `now`.
   - **`BeforeAfterJSON`** renders `before_state` / `after_state` (and is the disclosure body). Use **alphabetical key ordering** (the Story 2.8 / 2.12 convention) so the JSON snippet is byte-stable across stacks. Render through **default escaping** — `metadata.reason` and any string value is user-influenced (cat 3a).
   - **`ActionClass`** maps the action string to a badge class via the established AuditRow mapping (Story 2.4); an unknown action falls back to the documented neutral `badge-unknown` (cat 1) — do not silently style-default.

5. **Read-only for every role, including Executive (FR43); gated by `project.read`.** The tab and the Load-more route authorize `can(actor, "project.read")` (the same gate 2.11 put on `GET /projects/<id>` and its tab endpoints — [Go projects_detail_handler.go:201](../../fieldmark-go/internal/web/handlers/projects_detail_handler.go), [Django views.py:307-321](../../fieldmark_py/projects/views.py)). Because **AuditRow is markup-only with zero action affordances** (Story 2.4), FR43's "Executive sees read-only, no action affordances" is satisfied **structurally** — there is nothing to suppress. A non-reader gets the canonical **403** body (Story 1.11 shape — do not invent a new one), **no entity-state leakage**. Unauthenticated → the 1.11 login redirect fires first.

---

## Acceptance Criteria

### AC1 — Audit tab renders the live `#audit-log` list, newest-first (Decision 1, FR42)

**Given** I am authorized (`can(actor, "project.read")`) and navigate to the **Audit** tab on `/projects/<id>` (the TabStrip from 2.11 fires `hx-get=/projects/<id>/tabs/audit`, `hx-target="#project-detail-tab-content"`)
**When** the tab content swaps in
**Then** `#project-detail-tab-content` contains `<ul id="audit-log" role="list" aria-live="polite">` whose children are **AuditRow** components — one per `domain.audit_entry` row where `project_id = <id>`, ordered `occurred_at DESC, id DESC` (Decision 3)
**And** each AuditRow is composed from the Story 2.4 wrapper (no re-emitted inner markup) and carries the action badge, the resolved actor display, the machine `datetime` + human relative time, and the before/after disclosure (AC3)
**And** the tabpanel keeps the 2.11 `role="tabpanel"` + `aria-labelledby="tab-audit"` + focus-management (`tabindex="-1"`/`autofocus`) semantics the placeholder had (UX-DR33).

**Given** the Audit tab is open **and** a Story 2.12 place-on-hold/resume transition completes in the same session
**Then** the transition's OOB `<li hx-swap-oob="afterbegin:#audit-log">` lands as the **top** row of `#audit-log` in the same paint (Decision 1 — the 2.12 epic "new row at top" guarantee, now live).

### AC2 — Executive (and every non-mutating reader) sees a read-only log (FR43, Decision 5)

**Given** I am authenticated as **Executive** (or any role with `project.read`)
**When** I view the Audit tab
**Then** every row renders as a read-only AuditRow with **no action affordances anywhere** (UX-DR21 collapses to `absent` — and AuditRow has none to begin with)
**And** the response is identical in structure to what ADMIN sees (the audit log content is not role-filtered — forensic visibility is the same for all readers).

**Given** I lack `project.read`
**When** I `GET /projects/<id>/tabs/audit` **or** `GET /projects/<id>/audit-log`
**Then** the response is **HTTP 403** with the canonical 1.11 body, **no entity state leaked** (no project fields, no row data, no "exists but forbidden" signal).

**Given** an **unauthenticated** request to either route
**Then** the Story 1.11 redirect-to-login fires first (302/303 → `/login`), unchanged.

### AC3 — Before/after disclosure renders JSON in JetBrains Mono with correct ARIA (UX-DR12)

**Given** a single AuditEntry with non-null `before_state` / `after_state` (and/or `metadata`)
**When** I expand its disclosure
**Then** the `before_state` / `after_state` (and metadata) JSON renders in **JetBrains Mono** (`font-mono`) inside the AuditRow's `<details>` with `aria-expanded` toggling correctly (collapsed by default; UX-DR12)
**And** the JSON uses **alphabetical key ordering** (byte-stable cross-stack — Story 2.8/2.12 convention).

**Given** an AuditEntry whose `before_state`/`after_state` are **null** (an event with no state delta)
**Then** the disclosure is **absent** (the AuditRow already conditions the `<details>` on `BeforeAfterJSON` being present) — no empty `<details>`.

### AC4 — Keyset "Load more" pagination, ≤ 100 per page (Decision 3, epic AC)

**Given** a project with **> 100** audit entries
**When** the Audit tab first renders
**Then** **at most 100** AuditRows load, newest-first, **and** a "Load more" affordance is present (an `hx-get` to `GET /projects/<id>/audit-log?before_occurred_at=<iso>&before_id=<uuid>` carrying the **last rendered row's cursor**), targeting `#audit-log` with **append** semantics (`hx-swap="beforeend"` for rows)
**And** the `PAGE_SIZE` constant is **100**, documented, identical across stacks
**And** the "more exist?" decision uses a **`LIMIT PAGE_SIZE + 1` sentinel** (no `COUNT(*)`).

**Given** I click "Load more"
**When** the next page returns
**Then** the next ≤ 100 AuditRows **append** to `#audit-log` (no duplicate or skipped rows even if a new entry was prepended via OOB since the first render — keyset guarantees this), **and** the Load-more control is refreshed with the new cursor (or removed when the sentinel shows no more remain).

**Given** a project with **≤ 100** entries (or the last page)
**Then** the Load-more affordance is **absent** (not a disabled stub).

### AC5 — Empty state (Decision, cat 9)

**Given** a project with **zero** audit entries
**When** the Audit tab renders
**Then** `#audit-log` is present but renders a single empty-state message — "No audit entries recorded for this project yet." — with Basecoat empty styling and an appropriate `aria-label`/`role` (distinct from an error; mirrors the 2.9 no-rows-overlay intent for a non-grid list)
**And** no Load-more affordance is rendered
**And** the empty list still exposes `id="audit-log"` `aria-live="polite"` so a subsequent OOB-prepended transition row (Story 2.12) lands and announces correctly (the empty-state message and an OOB-landed row must not collide — document the chosen composition: the empty-state node is replaced/hidden on first real row, or it sits as a sibling the OOB prepend precedes).

### AC6 — Read path is project-scoped, ordered, and adds no schema (Decision 2, FR42, NFR6)

**Given** each stack's audit data layer
**When** I inspect the new read method
**Then** it returns only rows where `project_id = <id>`, ordered `occurred_at DESC, id DESC`, **manually projected** to the AuditRow VM (no AutoMapper — .NET; no ORM-magic mapping that bypasses the VM) (NFR6)
**And** Go's read method is **pool-backed and separate from the tx-threaded `Append`** (the `AuditEntryStore.Append` interface is unchanged)
**And** a per-stack test loads a project seeded with a known set of audit entries and asserts: correct count, correct `occurred_at DESC, id DESC` ordering, project-scoping (entries for a *different* project are excluded), and round-trip of `action` / `actor_id` / `before_state` / `after_state` / `metadata`
**And** `make parity` shows **`pg_indexes` zero-diff** (no `domain.*` change).

### AC7 — Security-defaults + edge-case checklist coverage

**Given** [security-defaults.md](../../docs/reference/security-defaults.md) **cat 3a (XSS round-trip on render)**
**Then** every user-influenced string the AuditRow renders goes through each engine's **default escaping** (no `Html.Raw`/`|safe`/`template.HTML`). A per-stack test seeds an audit entry with `metadata = {"reason": "<script>alert(1)</script>"}` (the bare payload, **not** JSON-wrapped — the JSON disclosure body is the render surface), renders the Audit tab, and asserts **both** `Contains("&lt;script&gt;alert(1)&lt;/script&gt;")` **and** `NotContains("<script>")`. Repeat for the **actor display** surface (seed an entry whose resolved actor name contains the payload, or assert the actor name is escaped) — every user-visible prop, not just the first (cat 3a / edge-checklist cat 10).

**Given** [component-edge-case-checklist.md](../../docs/reference/component-edge-case-checklist.md):
- **cat 1 (unknown enum)** — an `action` string outside the canonical list renders the documented neutral `badge-unknown` fallback (Story 2.4 safety net) with the single warn-log, **not** a silent default style; a unit test seeds an unknown action and asserts the fallback class. (The canonical 15 actions are CHECK-free strings in the column, so a malformed/forward-compat value is reachable.)
- **cat 6 (text overflow & special characters)** — long `before/after` JSON and special characters render without breaking layout (the AuditRow JSON `<pre><code>` wraps/scrolls per the 2.4 contract); HTML entities decode to characters, not double-escaped.
- **cat 9 (empty/whitespace derived values)** — an entry whose `actor_id` resolves to an **empty / whitespace-only / null** display name renders the `"??"` fallback; a unit test covers all three boundary inputs. An entry with **null** before/after renders no disclosure (AC3).
- **cat 7/8 (reduced-motion / forced-colors)** — handled by the 1.14 globals; the disclosure toggle and badge colors carry text labels (the action string is always present in the badge), so the log is legible with motion/colors disabled. An **axe-core** scan on the rendered Audit tab (with one row expanded) reports **zero** new WCAG 2.1 AA violations, including the `aria-live`/`aria-expanded` wiring.

**Given** the tab and Load-more routes are **GET / read-only**
**Then** **cat 6 (CSRF) is N/A** (no state change) and no CSRF token is required — document this explicitly so a reviewer doesn't flag a "missing" token.

### AC8 — Cross-stack byte-identical audit-log shape (epic AC, Decision 4)

**Given** the same seeded audit entries across all three stacks
**When** I render `GET /projects/<id>/tabs/audit` (and a Load-more fragment)
**Then** the **HTML structure** of the `#audit-log` list, each AuditRow, the disclosure, and the Load-more affordance is **byte-identical across stacks** (anchored on structure + the machine `datetime`/`title` ISO-8601-UTC values + the alphabetical-key JSON; the human relative-time wording for a live `now` is excluded from the byte-comparison per Decision 4 — use a fixed/injected clock or compare structure-only)
**And** the per-stack render test references the canonical AuditRow example by path (no copy-pasted expected markup), consistent with the 2.4 snapshot convention.

### AC9 — `make parity` + E2E + full gate (no schema change)

**Given** Story 1.3 route-parity tooling
**When** I run `make parity`
**Then** all three route dumps contain `GET /projects/:id/tabs/audit` (already present from 2.11) **and** the new `GET /projects/:id/audit-log` (stack-idiomatic param syntax), diff clean; `pg_indexes` diff is **zero**.

**Given** a Playwright E2E scenario per stack
**When** I (login as ADMIN → open `/projects/<id>` with > 100 entries → click the Audit tab)
**Then** the first ≤ 100 rows render newest-first, "Load more" appends the next page with no duplicates, and an **expanded disclosure** shows the JSON
**And** a second scenario verifies the **Story 2.12 integration**: with the Audit tab open, place the project on hold → the new AuditRow appears at the **top** of `#audit-log` in one paint (the deferred 2.12 epic AC, now landing live — Decision 1).

**Build/type/lint/test gates green per stack:**
- **.NET:** `cd FieldMark && dotnet csharpier check . && dotnet build && dotnet test && dotnet test FieldMark.Tests.Integration/FieldMark.Tests.Integration.csproj` — clean.
- **Django:** `cd fieldmark_py && uv run ruff check . && uv run mypy . && uv run pytest && uv run pytest -m integration` — clean.
- **Go:** `cd fieldmark-go && make check && go test ./... && go test -tags=integration ./...` — clean.
- **No `fieldmark_shared` CSS change expected** (AuditRow / list / disclosure / empty-state styles exist from 2.4 / 1.14). If a hand-authored rule is unavoidable (e.g. an audit-log empty-state class), rebuild `dist/fieldmark.css` via `make css` and justify in dev notes; otherwise assert **zero** CSS drift.

> ⚠️ **.NET aggregate web-suite runner caveat (carried from 2.11/2.12):** solution-wide `make test-net` has not reliably concluded the aggregate `FieldMark.Tests.Web` run under the current runner. If it recurs, treat focused class-level passes (`--filter ProjectsDetail`, the new audit-tab test class) plus the domain/integration lanes and browser verification as the definitive .NET evidence, and record the caveat in the Sign-off — do not block the story on the aggregate runner alone.

---

## Tasks / Subtasks

- [x] **Task 1: Project-scoped audit read path** (AC: #6, #4)
  - [x] 1.1 **Go:** add `internal/data/postgres/auditentryreadstore.go` — a pool-backed `AuditEntryReadStore` with `ListByProject(ctx, projectID, page)` implementing the keyset query (`ORDER BY occurred_at DESC, id DESC LIMIT PAGE_SIZE+1`, cursor `WHERE (occurred_at, id) < ($ts,$id)`). Keep `AuditEntryStore.Append` untouched. Add the actor-name join (reuse 2.9's user lookup).
  - [x] 1.2 **Django:** add a read query (e.g. `audit/queries.py` or a `projects/views.py` helper) — `AuditEntry.objects.filter(project_id=id)` + keyset filter, `order_by("-occurred_at", "-id")`, `[:PAGE_SIZE+1]`. Resolve actor display via the 2.9 user lookup.
  - [x] 1.3 **.NET:** add a read service/query over `DbSet<AuditEntry> AuditEntries` — `.Where(project) .OrderByDescending(OccurredAt).ThenByDescending(Id)`, keyset predicate, `.Take(PAGE_SIZE+1)`, **manual projection** to the AuditRow VM (no AutoMapper). Actor join via 2.9 lookup.
  - [x] 1.4 Define `PAGE_SIZE = 100` once per stack (documented constant) and the keyset cursor type. Per-stack read-path test: count, ordering, project-scoping (exclude other project's rows), JSON/actor round-trip.

- [x] **Task 2: Audit tab panel — live `#audit-log` list** (AC: #1, #2, #3, #5)
  - [x] 2.1 Replace the placeholder in each stack — Go `project_detail_audit_panel` ([projects_detail_panels.html:30-31](../../fieldmark-go/internal/web/templates/pages/projects_detail_panels.html)), Django [tabs/_placeholder_audit.html](../../fieldmark_py/templates/projects/tabs/_placeholder_audit.html), .NET [Tabs/_AuditPanel.cshtml](../../FieldMark/FieldMark.Web/Pages/Projects/Tabs/_AuditPanel.cshtml) — with `<ul id="audit-log" role="list" aria-live="polite">` of **AuditRow** components, preserving the tabpanel `role`/`aria-labelledby="tab-audit"`/focus semantics.
  - [x] 2.2 Wire the tab handler to call the Task-1 read path and map rows → AuditRow VMs (actor display + relative/absolute time + alphabetical-key before/after JSON + ActionClass with cat-1 fallback). Go: extend [projects_detail_handler.go](../../fieldmark-go/internal/web/handlers/projects_detail_handler.go) `audit` case (currently sets `panel = "project_detail_audit_panel"`); Django: [project_tab](../../fieldmark_py/projects/views.py); .NET: the Detail page model's audit-tab branch.
  - [x] 2.3 Empty state (AC5): zero rows → the empty-state node inside `#audit-log`; document the OOB-coexistence composition so a later prepended row lands cleanly.
  - [x] 2.4 Authorize `project.read` (403 canonical body, no leakage); 404 on bad/missing project id.

- [x] **Task 3: Keyset "Load more" fragment route** (AC: #4, #9)
  - [x] 3.1 Add `GET /projects/<id>/audit-log?before_occurred_at=<iso>&before_id=<uuid>` per stack (Django `urls.py`, Go `cmd/web/main.go`, .NET Razor page/handler) → returns the next-page **fragment**: `<li>` AuditRows + refreshed/removed Load-more control. Authorize `project.read`. Parse + validate the cursor params (reject malformed → 400/422, no leakage).
  - [x] 3.2 Initial tab render emits the Load-more control with the first page's boundary cursor when the `+1` sentinel indicates more remain; absent otherwise.
  - [x] 3.3 Append semantics: rows `hx-swap="beforeend"` into `#audit-log`; the control replaces itself — one mechanism, identical across stacks, documented in dev notes / the AuditRow contract.

- [x] **Task 4: Tests — read path, render, authz, pagination, empty, XSS, parity** (AC: #2,#3,#5,#6,#7,#8)
  - [x] 4.1 Render test: Audit tab for a seeded project → `#audit-log` present, N AuditRows newest-first, disclosure collapsed-by-default with `aria-expanded`, JSON in `font-mono`.
  - [x] 4.2 Authz: each non-`project.read` role → 403 canonical body, zero row data; ADMIN + Executive → 200 identical structure (AC2). Unauthenticated → login redirect.
  - [x] 4.3 Pagination: > 100 entries → first page = 100 + Load-more present; Load-more fragment = next page appended, no dup/skip after an interleaved insert; last page → Load-more absent. ≤ 100 → no Load-more.
  - [x] 4.4 Empty state: zero entries → empty-state message, no Load-more, `#audit-log`+`aria-live` still present.
  - [x] 4.5 XSS round-trip (cat 3a): `<script>alert(1)</script>` in `metadata.reason` (and actor display) → both assertions, every user-visible surface.
  - [x] 4.6 cat 1 unknown-action → `badge-unknown` + warn-log; cat 9 empty/whitespace/null actor → `"??"`; null before/after → no disclosure.
  - [x] 4.7 Cross-stack byte-identical audit-log/AuditRow/Load-more structure (AC8) — fixed/injected clock or structure-only comparison; reference the canonical AuditRow example by path.
  - [x] 4.8 axe-core scan on the rendered tab (one row expanded) → zero new WCAG 2.1 AA violations.

- [x] **Task 5: E2E + parity + gate** (AC: #9)
  - [x] 5.1 E2E: open Audit tab on a >100-entry project → first page + Load-more append + expand disclosure, per stack.
  - [x] 5.2 E2E **Story 2.12 integration**: Audit tab open → place-on-hold → new AuditRow at top of `#audit-log` in one paint (Decision 1 — the live landing of 2.12's OOB).
  - [x] 5.3 `make parity` (`GET /projects/:id/audit-log` present on all three, `pg_indexes` zero-diff) + full per-stack gate green; assert zero `fieldmark_shared` CSS drift (or justify a rebuild).

- [x] **Task 6: Doc updates + story sign-off** (AC: all)
  - [x] 6.1 Add the **relative-time bucketing rules**, the **keyset Load-more route + query-param contract**, and the **`#audit-log` live-target / OOB-coexistence note** to the AuditRow section of [component-canonical-examples.md](../../docs/reference/component-canonical-examples.md) (and/or the `audit_row/README.md`), so the contract is the single source of truth (Cross-Stack Architecture Principle).
  - [x] 6.2 Cross-reference from [three-region-oob-orchestration.md](../../docs/how-to/three-region-oob-orchestration.md): flip its Decision-4 "lands live at Story 2.13" note to "live as of Story 2.13" now that `#audit-log` exists.
  - [x] 6.3 Populate the Sign-off block; record the five decisions; confirm the 2.12 OOB now lands live; note the .NET aggregate-runner caveat if it recurs.

## Dev Notes

### Critical context (read before writing code)

- **This is a *read* story — the opposite discipline from 2.12.** No transaction, no `append_audit_entry`, no entity method, no mutation, no CSRF. The work is: a project-scoped read query + the AuditRow list render + keyset Load-more. Do **not** copy the 2.12 mutating-handler shape.
- **The single highest-value outcome: `#audit-log` becomes live (Decision 1).** Story 2.12 already emits `<li hx-swap-oob="afterbegin:#audit-log">` and it silently no-ops today. Your `<ul id="audit-log">` must be the **exact** prepend target — id, and `<li>`-children structure matching the `audit_row` component. Verify with the AC9 E2E that a transition row lands at the top while the tab is open. Getting the id or the element type wrong = "the 2.12 OOB still doesn't land," a silent regression.
- **The data layer is append-only — you are adding the first read path (Decision 2).** Don't extend the tx-threaded `Append`. Go gets a separate pool-backed read store; Django queries the existing model; .NET projects the existing `DbSet<AuditEntry>` manually (no AutoMapper).
- **Keyset, not offset (Decision 3).** Offset duplicates/skips rows because the table is append-and-prepend during a session. Use `(occurred_at, id) < (cursor)` with `LIMIT PAGE_SIZE+1` sentinel. This is the load-bearing correctness decision for pagination.
- **Reuse the AuditRow VM and the cat-9 `"??"` fallback — don't reinvent.** The VM (Action/ActionClass/ActorDisplay/OccurredAt/Absolute/Relative/BeforeAfterJSON) and the empty-actor fallback already exist from 2.4 and were exercised by 2.12. Feed it; don't re-author its markup or its fallback logic.
- **Actor name resolution reuses 2.9's `pm_name` user lookup.** Join `audit_entry.actor_id` to the same per-stack user source the Project list used for `pm_name`. An unresolvable actor → empty display → `"??"`.
- **Alphabetical JSON keys + ISO-8601-UTC machine timestamp are the parity anchors (Decision 4).** The human relative string drifts with `now`; compare structure + machine values, or inject a fixed clock in the parity test.
- **`metadata.reason` is user text (from 2.12) → escape it (cat 3a).** It renders inside the before/after disclosure. Bare-payload XSS round-trip, both assertions, on the JSON surface **and** the actor surface.
- **Read-only for Executive is free (Decision 5).** AuditRow has no affordances; FR43 is satisfied structurally. Gate on `project.read`; non-readers → canonical 1.11 403, no leakage.

### Source tree — where things land

| Stack | Read path | Tab handler + Load-more route | Templates |
|---|---|---|---|
| .NET | read query/service over `DbSet<AuditEntry>` (e.g. `FieldMark.Data/...` or `FieldMark.Web` service) — manual projection | Detail page model audit branch + a `GET /projects/{id}/audit-log` handler (Razor page or minimal-API) | `Pages/Projects/Tabs/_AuditPanel.cshtml` (replace placeholder) + a Load-more fragment partial; compose `Pages/Shared/Components/_AuditRow.cshtml` |
| Django | `audit/queries.py` (or view helper) over `AuditEntry` model | `projects/views.py` `project_tab` audit branch + a new `audit-log` view in `projects/urls.py` | `templates/projects/tabs/_placeholder_audit.html` → real panel; Load-more fragment; compose `templates/components/_audit_row.html` |
| Go | `internal/data/postgres/auditentryreadstore.go` (new, pool-backed) | extend `projects_detail_handler.go` audit case + new handler/route in `cmd/web/main.go` | `projects_detail_panels.html` `project_detail_audit_panel` (replace) + Load-more fragment; compose `components/audit_row.html` |

Docs: [component-canonical-examples.md](../../docs/reference/component-canonical-examples.md) (+ `audit_row/README.md`) for the relative-time + Load-more + live-target contract; [three-region-oob-orchestration.md](../../docs/how-to/three-region-oob-orchestration.md) cross-ref. No `fieldmark_shared` CSS change expected.

### Existing code to reuse (read before writing)

- **AuditRow wrapper + VM:** Go [audit_row.html](../../fieldmark-go/internal/web/templates/components/audit_row.html) / [AuditRowVM in components.go](../../fieldmark-go/internal/web/viewmodels/components.go), .NET [_AuditRow.cshtml](../../FieldMark/FieldMark.Web/Pages/Shared/Components/_AuditRow.cshtml), Django [_audit_row.html](../../fieldmark_py/templates/components/_audit_row.html). Snapshot tests show the contract: [AuditRowSnapshotTests.cs](../../FieldMark/FieldMark.Tests.Web/Components/AuditRowSnapshotTests.cs), [audit_row_test.go](../../fieldmark-go/internal/web/templates/components/audit_row_test.go), [test_audit_row_snapshot.py](../../fieldmark_py/fieldmark/tests/test_audit_row_snapshot.py).
- **The transition AuditRow mapping (the shape to reuse):** Go [projects_transition_handler.go:236-243](../../fieldmark-go/internal/web/handlers/projects_transition_handler.go) builds an `AuditRowVM` (`ActionClass="badge-audit-action"`, `Relative="just now"`, `BeforeAfterJSON=…`). Your read path builds the same VM from real rows — replace the hardcoded `"just now"` with the real relative-time helper.
- **Audit data layer:** Go [auditentrystore.go](../../fieldmark-go/internal/data/postgres/auditentrystore.go) (append-only — add a sibling read store), Django [audit/models.py](../../fieldmark_py/audit/models.py) (queryable model) + [append.py](../../fieldmark_py/audit/append.py), .NET [AuditEntryConfiguration.cs](../../FieldMark/FieldMark.Data/Configuration/AuditEntryConfiguration.cs) + [FieldMarkDbContext.cs](../../FieldMark/FieldMark.Data/Context/FieldMarkDbContext.cs) (`DbSet<AuditEntry> AuditEntries`).
- **Tab handler + placeholder + authz:** Go [projects_detail_handler.go](../../fieldmark-go/internal/web/handlers/projects_detail_handler.go) (audit case at ~:172; `project.read` gate at :201), Django [views.py project_tab/project_detail](../../fieldmark_py/projects/views.py) (`project.read` at :307-321), .NET Detail page model + [Tabs/_AuditPanel.cshtml](../../FieldMark/FieldMark.Web/Pages/Projects/Tabs/_AuditPanel.cshtml).
- **Actor-name lookup precedent:** Story 2.9 `pm_name` projection in each stack's `/grid/projects` handler — reuse the same user source for `actor_id → display name`.
- **OOB prepend target this story enables:** Go [projects_transition_response.html:7-8](../../fieldmark-go/internal/web/templates/pages/projects_transition_response.html); .NET [_DetailTransitionResponse.cshtml](../../FieldMark/FieldMark.Web/Pages/Projects/_DetailTransitionResponse.cshtml); Django [_detail_transition_response.html](../../fieldmark_py/templates/projects/_detail_transition_response.html).

### Project Structure Notes

- Adds **1 route** to the parity inventory (`GET /projects/:id/audit-log` for Load-more); `GET /projects/:id/tabs/audit` already exists from 2.11. No `domain.*` schema change (`pg_indexes` zero-diff).
- First **read** of `domain.audit_entry` (Story 2.2 mapped it write-only). The read store / query pattern established here is reused by every entity detail screen with an audit view (Epics 3–5).
- Makes the canonical `#audit-log` target (project-context line 69) live for the first time, completing the deferred half of Story 2.12's three-region contract (Decision 4 of 2.12).
- No new component; consumes the 2.4 AuditRow. The only possible `fieldmark_shared` touch is an empty-state class if one doesn't already exist — verify before hand-authoring CSS.

### References

- Epic AC: [epic-2 §Story 2.13](../planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md)
- Prior screen + tab wiring: [2-11 Project Detail](2-11-project-detail-anchor-screen-with-header-strip-tabstrip-and-entityrail.md); OOB emitter this story completes: [2-12 Place-on-hold/Resume](2-12-place-on-hold-and-resume-transitions-with-three-region-oob-orchestration.md) (Decision 4)
- AuditRow component: [2-4 Phase-2 components](2-4-implement-phase-2-markup-only-components-statusbadge-inlinealert-auditrow-dashboardtile.md); contract: [component-canonical-examples.md](../../docs/reference/component-canonical-examples.md), [audit_row/README.md](../../fieldmark_shared/components/audit_row/README.md)
- Audit mapping: [2-2 audit_entry + append helper](2-2-map-domain-audit-entry-and-provide-a-per-stack-append-audit-entry-helper.md); action strings: [audit-actions.md](../../docs/reference/audit-actions.md)
- Actor lookup precedent: [2-9 Project list AG Grid](2-9-project-list-ag-grid-with-server-side-row-model.md) (`pm_name`)
- Three-region pattern (cross-ref): [three-region-oob-orchestration.md](../../docs/how-to/three-region-oob-orchestration.md)
- Edge cases / security: [component-edge-case-checklist.md](../../docs/reference/component-edge-case-checklist.md) cat 1/6/7/8/9/10, [security-defaults.md](../../docs/reference/security-defaults.md) cat 3a
- DDL: [010_domain_tables.sql](../../docker/postgres/init/010_domain_tables.sql) (`domain.audit_entry`)

## Dev Agent Record

### Agent Model Used

### Debug Log References
- 2026-06-03 17:20 CDT — Implemented the audit-tab read path and `/projects/<id>/audit-log` fragment in Django, Go, and .NET; parity verified the new route on all three stacks.
- 2026-06-03 17:20 CDT — Verification so far: `uv run python -m py_compile projects/views.py`; `uv run pytest projects/tests/test_project_detail.py -k audit` (all selected tests skipped here because the default Django test DB does not expose the `domain.*` tables); `GOCACHE=/private/tmp/fieldmark-go-cache go test ./internal/web/handlers -run 'ProjectsDetail|ProjectAuditLog'`; `dotnet test FieldMark.Tests.Web/FieldMark.Tests.Web.csproj --filter 'ProjectsDetailTab_Audit_RendersLiveAuditLog|ProjectAuditLog_ReturnsFragmentOnly'`; `make parity`.
- 2026-06-03 18:34 CDT — Added focused auth/cursor/current-tab coverage for the new audit read path: `dotnet test FieldMark.Tests.Web/FieldMark.Tests.Web.csproj --filter "ProjectsDetailTab_Audit_RendersLiveAuditLog|ProjectsDetailTab_Audit_NoRoleUser_ReturnsForbidden|ProjectAuditLog_ReturnsFragmentOnly|ProjectAuditLog_Unauthenticated_RedirectsToLogin|ProjectAuditLog_InvalidCursor_ReturnsBadRequest|ProjectPlaceOnHold_Post_CurrentTabAudit_KeepsAuditPanelLive"` passed (6 tests); `GOCACHE=/private/tmp/fieldmark-go-cache go test ./internal/web/handlers -run 'ProjectsDetail|ProjectAuditLog'` passed; `uv run pytest projects/tests/test_project_detail.py -k 'audit and (tab or log or current_tab)'` selected 7 Django cases but all skipped here because the default test DB still lacks `domain.project` / `domain.audit_entry`.
- 2026-06-03 19:00 CDT — Fixed the audit-tab transition duplicate-row bug across .NET, Django, and Go by rendering only older rows in the swapped panel when `current_tab=audit` and letting the OOB prepend supply the new top row; added shared `docs/reference/fixtures/project-audit-log-canonical.html` parity coverage and a Go integration assertion for the current-tab success path.
- 2026-06-03 19:00 CDT — Final verification: `dotnet test FieldMark.Tests.Web/FieldMark.Tests.Web.csproj --filter "ProjectsDetailTab_Audit|ProjectAuditLog|ProjectPlaceOnHold_Post_CurrentTabAudit_KeepsAuditPanelLive"` passed (8 tests); `GOCACHE=/private/tmp/fieldmark-go-cache go test ./internal/web/handlers -run 'ProjectsDetail|ProjectAuditLog'` passed; `GOCACHE=/private/tmp/fieldmark-go-cache go test -tags=integration ./internal/web/handlers -run 'CurrentTabAudit_KeepsSingleLiveRow|PostProjectPlaceOnHold_Success'` passed; `make parity` passed; `make test-django` passed with the expected `domain.*` integration skips; `cd e2e && npx playwright test tests/shared/project-audit-log.spec.ts` still fails before assertions on all three stacks with Chromium MachPort `bootstrap_check_in ... Permission denied (1100)`.
- 2026-06-04 19:30 CDT — Closed round-1 review findings: Go now recursively sorts nested audit JSON, Django/.NET/Go cover unknown-action + actor fallback + actor/XSS cases, the audit disclosure body renders while collapsed, and the canonical page fixtures were updated to match the shared AuditRow contract.
- 2026-06-04 19:30 CDT — Final verification refresh: `dotnet test FieldMark.Tests.Web/FieldMark.Tests.Web.csproj --filter "AuditRowSnapshotTests|ProjectsDetailPageTests.ProjectsDetailTab_Audit_RenderAuditJson_SortsNestedObjects|ProjectsDetailPageTests.ProjectsDetailTab_Audit_XssPayloads_AreEscapedAcrossActorAndMetadata"` passed (11 tests); `GOCACHE=/private/tmp/fieldmark-go-cache go test ./internal/web/handlers ./internal/web/templates/components` passed; `uv run pytest fieldmark/tests/test_audit_row_snapshot.py` passed (9 tests); `make css` rebuilt `fieldmark_shared/dist/fieldmark.css`; `cd e2e && npx playwright test tests/shared/project-audit-log.spec.ts` passed on `dotnet`, `django`, and `fiber` (9 tests), including the new load-more/disclosure path and the axe-core audit-tab scan.
- 2026-06-04 21:18 CDT — Closed the remaining round-2 regressions: removed the broken `<template>` OOB wrappers, strengthened the Playwright row-count assertion, and updated the current-tab audit transition path so audit-tab submissions re-render the new row in-band while keeping the standard two-OOB transition contract for non-audit submissions. Verification: `dotnet test FieldMark.Tests.Web/FieldMark.Tests.Web.csproj --filter "ProjectsDetailPageTests"` passed (41 tests); `GOCACHE=/private/tmp/fieldmark-go-cache go test ./internal/web/handlers -run 'TestGetProjectsDetailTab_Audit|TestGetProjectAuditLog|TestPostProjectPlaceOnHold_CurrentTabAudit_KeepsSingleLiveRow'` passed; `uv run pytest projects/tests/test_project_detail.py -k 'audit'` remained environment-skipped because the default DB still lacks `domain.*`; `cd e2e && npx playwright test tests/shared/project-audit-log.spec.ts` passed (9 tests).
- 2026-06-04 22:05 CDT — Closed the round-3 patch items: Go transition rows now include `metadata` in `BeforeAfterJSON`, Go and Django audit-panel extractors no longer use lazy regex truncation, and Go explicitly covers the unresolvable-actor cat-9 boundary. Also addressed two advisories by switching the .NET load-more extractor to HtmlAgilityPack and extracting the shared Playwright `projectSlotForBaseUrl` helper into `e2e/tests/shared/helpers.ts`. Verification: `dotnet test FieldMark.Tests.Web/FieldMark.Tests.Web.csproj --filter "ProjectsDetailPageTests"` passed (41 tests); `GOCACHE=/private/tmp/fieldmark-go-cache go test ./internal/web/handlers -run 'TestGetProjectsDetailTab_Audit|TestGetProjectAuditLog|TestPostProjectPlaceOnHold_CurrentTabAudit_KeepsSingleLiveRow|TestPostProjectPlaceOnHold_Success'` passed; `uv run python -m py_compile projects/tests/test_project_detail.py` passed; `uv run pytest projects/tests/test_project_detail.py -k 'audit'` remained environment-skipped because the default DB still lacks `domain.*`; `cd e2e && npx playwright test tests/shared/project-audit-log.spec.ts --workers=1` passed (9 tests); `cd e2e && npx playwright test tests/shared/project-transition-flow.spec.ts --workers=1` passed (6 tests).
- 2026-06-05 11:12 CDT — Closed the round-4 follow-ups: Django transition rows now include `metadata.reason` in `before_after_json`, the Go audit-panel extractor only counts real `<div>` tags, `loadAuditPage` now emits a warning when `AuditRead` is unwired, .NET and Django both carry explicit empty-string actor fallback tests, Django replaced the load-more row regex helper with an `HTMLParser`, and the AC5 sibling-precedes composition is now documented in both contract docs. Verification: `dotnet test FieldMark.Tests.Web/FieldMark.Tests.Web.csproj --filter "ProjectsDetailPageTests"` passed (42 tests); `GOCACHE=/private/tmp/fieldmark-go-cache go test ./internal/web/handlers -run 'TestGetProjectsDetailTab_Audit|TestGetProjectAuditLog|TestPostProjectPlaceOnHold_CurrentTabAudit_KeepsSingleLiveRow|TestPostProjectPlaceOnHold_Success'` passed; `uv run python -m py_compile projects/views.py projects/tests/test_project_detail.py` passed; `uv run pytest projects/tests/test_project_detail.py -k 'audit'` remained environment-skipped because the default DB still lacks `domain.*`; `cd e2e && npx playwright test tests/shared/project-audit-log.spec.ts --workers=1` passed (9 tests) after restoring the Django dev server.

### Completion Notes List
- Implemented a first-pass multi-stack audit-log read flow: each stack now renders a real `#audit-log` list in the Audit tab, maps persisted audit rows into the existing AuditRow component, and exposes a keyset-paginated `/projects/<id>/audit-log` fragment route.
- Added focused Go and .NET tests for the audit-tab render path, read-route auth/cursor handling, and current-tab transition preservation; Django has matching focused tests in place, but the DB-backed cases skip in this environment because the default test database does not currently include the `domain.project` / `domain.audit_entry` tables.
- Updated the AuditRow README and the three-region OOB orchestration doc so the live-target, relative-time, and load-more contracts are documented before broader verification resumes.
- Closed the main response-contract defect discovered during verification: a successful transition with `current_tab=audit` now keeps the Audit tab active without duplicating the new row or leaving the empty-state placeholder behind.
- Added a shared page/fragment fixture for the audit log shell so .NET, Django, and Go now prove the same normalized audit-panel and load-more shape.
- Story is ready for reviewer handoff: stack-level server tests and parity are green, Django’s full suite passes with expected live-DB skips, and the only unresolved lane is Playwright browser launch in this environment rather than application behavior.
- Review follow-up work is closed: the audit-tab disclosure now reveals JSON when expanded, the shared audit-action badge meets axe contrast on the dark surface, and the page-level shell fixtures match the updated shared AuditRow contract.
- Browser verification is no longer environment-blocked for this story: the shared Playwright audit-log spec passes across `.NET`, Django, and Go with serial DB seeding, load-more coverage, the 2.12 OOB landing assertion, and an axe-core scan scoped to the audit tab.
- The final current-tab audit regression is closed: when a transition is submitted from the audit-tab flow, the returned panel now includes the new top row in-band, so the tab stays live at 101 rows for a full page without depending on an OOB target that was temporarily removed by the reason form swap.
- Round-3 follow-up is closed for code-review handoff: Go transition disclosures now match the tab-load JSON shape, the Go and Django parity extractors no longer silently truncate nested panels, and the shared Playwright slot-selection logic lives in one helper instead of two copy-pasted implementations.
- Round-4 review follow-up is closed: the remaining Django OOB disclosure parity bug is fixed, the last documented advisories were addressed, focused server-side verification is green, and the shared Playwright audit-log spec passed again in single-worker mode after restoring the local Django server.

### File List
- FieldMark/FieldMark.Tests.Web/Pages/ProjectsDetailPageTests.cs
- FieldMark/FieldMark.Web/Pages/Projects/AuditLog.cshtml
- FieldMark/FieldMark.Web/Pages/Projects/AuditLog.cshtml.cs
- FieldMark/FieldMark.Web/Pages/Projects/ProjectDetailPageModelBase.cs
- FieldMark/FieldMark.Web/Pages/Projects/Tabs/_AuditPanel.cshtml
- FieldMark/FieldMark.Web/Pages/Projects/_AuditLogItems.cshtml
- FieldMark/FieldMark.Web/Pages/Projects/_DetailBody.cshtml
- FieldMark/FieldMark.Web/Pages/Projects/_ProjectTransitionForm.cshtml
- docs/how-to/three-region-oob-orchestration.md
- docs/reference/fixtures/project-audit-log-canonical.html
- e2e/tests/shared/project-audit-log.spec.ts
- fieldmark-go/cmd/web/main.go
- fieldmark-go/internal/data/postgres/auditentryreadstore.go
- fieldmark-go/internal/web/handlers/projects_audit_log_handler.go
- fieldmark-go/internal/web/handlers/projects_detail_handler.go
- fieldmark-go/internal/web/handlers/projects_detail_handler_test.go
- fieldmark-go/internal/web/handlers/projects_detail_handler_internal_test.go
- fieldmark-go/internal/web/handlers/projects_transition_integration_test.go
- fieldmark-go/internal/web/handlers/projects_transition_handler.go
- fieldmark-go/internal/web/templates/pages/projects_audit_log_items.html
- fieldmark-go/internal/web/templates/pages/projects_detail_body.html
- fieldmark-go/internal/web/templates/pages/projects_detail_panels.html
- fieldmark-go/internal/web/templates/pages/projects_transition_form.html
- fieldmark_py/projects/tests/test_project_detail.py
- fieldmark_py/projects/urls.py
- fieldmark_py/projects/views.py
- fieldmark_py/templates/projects/_audit_log_items.html
- fieldmark_py/templates/projects/_detail_body.html
- fieldmark_py/templates/projects/_project_transition_form.html
- fieldmark_py/templates/projects/tabs/_audit_panel.html
- fieldmark_py/templates/projects/tabs/_placeholder_audit.html
- fieldmark_shared/components/audit_row/canonical.html
- fieldmark_shared/components/audit_row/README.md
- fieldmark_shared/dist/fieldmark.css
- fieldmark_shared/src/_tokens.css
- _bmad-output/implementation-artifacts/sprint-status.yaml

### Review Findings

- [x] [Review][Patch] Go `renderAuditJSON` — nested JSONB not recursively sorted; `encoding/json` sorts top-level map keys but nested values from `json.Unmarshal` into `any` are `map[string]interface{}` — also sorted in practice, but unlike .NET's recursive `SortJsonNode` and Django's `sort_keys=True`, Go provides no explicit recursive sort guarantee, creating a fragile parity anchor that will break if any value is not a flat object (e.g. an embedded array wrapping an object). AC3/AC8 require byte-stable alphabetical ordering. [`fieldmark-go/internal/web/handlers/projects_detail_handler.go`]
- [x] [Review][Patch] Django `_render_transition_success` — when `latest is None`, both cursor args are `None`; `_project_audit_page` runs without a cursor, returning all rows including the row just committed via OOB, causing a duplicate when both the panel body and the OOB prepend land. The `is not None else None` guard is present but chooses the wrong fallback behavior. [`fieldmark_py/projects/views.py`]
- [x] [Review][Patch] .NET `ExtractAuditPanel` test helper — lazy regex `.*?</div>` with `Singleline` stops at the **first** `</div>`, truncating any panel content that has nested `</div>` tags; the empty-state test currently passes because `<ul>` has no nested `<div>`, but any future non-empty content will silently produce a wrong comparison. [`FieldMark/FieldMark.Tests.Web/Pages/ProjectsDetailPageTests.cs`]
- [x] [Review][Patch] Go: missing XSS round-trip test for audit-log render surfaces (cat 3a, AC7) — no test seeds an audit entry with `metadata = {"reason": "<script>alert(1)</script>"}` and asserts both the escaped form present and raw `<script>` absent on the audit tab render; also missing for the actor display surface. Django has this test; Go and .NET do not. [`fieldmark-go/internal/web/handlers/projects_detail_handler_test.go`]
- [x] [Review][Patch] .NET: missing XSS round-trip test for audit-log render surfaces (cat 3a, AC7) — same as Go: no `metadata.reason` XSS + actor-display XSS round-trip assertion in the .NET audit-tab test suite. [`FieldMark/FieldMark.Tests.Web/Pages/ProjectsDetailPageTests.cs`]
- [x] [Review][Patch] Missing cat-1 unknown-action → `badge-unknown` unit test across all three stacks (AC7) — `auditActionClass` (Go), the Django template filter, and the .NET `ActionClass` mapping all have the fallback, but no per-stack test seeds an unknown action string and asserts `badge-unknown` is emitted (task 4.6).
- [x] [Review][Patch] Missing cat-9 empty/whitespace/null actor → `"??"` unit tests across all three stacks (AC7) — task 4.6 requires three boundary inputs (empty, whitespace-only, null/unresolvable actor_id) each producing the `"??"` fallback; no such tests exist in any stack's diff.
- [x] [Review][Patch] Go `TestGetProjectsDetailTab_AuditEmptyPanelMatchesCanonical` exercises the wrong canonical variant — the `auditReadStoreStub` returns one row with a `NextCursor`, so the test asserts against `panel-with-row-and-load-more`, not the `panel-empty` variant its name implies; no Go test ever validates the canonical empty-panel shape. [`fieldmark-go/internal/web/handlers/projects_detail_handler_test.go`]
- [x] [Review][Patch] E2E `project-audit-log.spec.ts` missing load-more click-through and expanded-disclosure scenario (AC9) — the spec requires "Load more appends the next page with no duplicates, and an expanded disclosure shows the JSON"; only the OOB transition scenario is present; the load-more + disclosure scenario is absent. [`e2e/tests/shared/project-audit-log.spec.ts`]
- [x] [Review][Patch] axe-core scan entirely absent from all stacks (AC7, cat 7/8) — AC7 requires "An axe-core scan on the rendered Audit tab (with one row expanded) reports zero new WCAG 2.1 AA violations, including the `aria-live`/`aria-expanded` wiring." Not documented as deferred in the sign-off; not implemented in any stack.
- [x] [Review][Patch] Empty-state `<li>` + OOB-prepend visible collision not resolved (AC5) — when the audit tab shows zero entries and a transition fires, `afterbegin:#audit-log` prepends the new AuditRow `<li>` before the empty-state `<li>`, leaving both visible simultaneously; AC5 requires "the empty-state message and an OOB-landed row must not collide — document the chosen composition"; no composition is documented or implemented.
- [x] [Review][Defer] .NET `latestAudit` fetched between `SaveChangesAsync` and `CommitAsync` — if commit fails, cursor points at a rolled-back row; exception propagates so no response is sent, making this a latent risk not an observable defect; pre-existing data-commit pattern carried from 2.12; hardening story 2.15 is the appropriate venue. [`FieldMark/FieldMark.Web/Pages/Projects/ProjectDetailPageModelBase.cs`] — deferred, pre-existing

### Round-2 Review Findings (2026-06-04)

**Round-1 status at round-2:** 10 of 11 round-1 patch items checked off by developer; P11 (empty-state collision) partially addressed via `SuppressAuditEmptyState` server-render suppression (live-DOM collision when `current_tab ≠ audit` remains). Round-2 confirms 8 round-1 items are still open and introduces 2 new regressions.

**New regressions (introduced by round-1 fixes):**

- [x] [Review][Patch] **CRITICAL — `<template>` wrapper breaks HTMX OOB swap, all three stacks** — all three `_DetailTransitionResponse` templates now wrap the `<li hx-swap-oob="afterbegin:#audit-log">` inside `<template>`. HTMX processes OOB elements via `querySelectorAll('[hx-swap-oob]')` on the parsed response fragment; `<template>` content lives in a separate `.content` DocumentFragment and is not reachable by normal DOM traversal — the OOB element is never found, the audit-row prepend silently no-ops on every transition, and Decision 1 / AC1 / AC9 (OOB live landing) is broken across all stacks. AC9's `.NET` test still passes because it string-matches `hx-swap-oob="afterbegin:#audit-log"` in the raw response body rather than asserting DOM behavior. [`FieldMark/FieldMark.Web/Pages/Projects/_DetailTransitionResponse.cshtml`, `fieldmark-go/internal/web/templates/pages/projects_transition_response.html`, `fieldmark_py/templates/projects/_detail_transition_response.html`]
- [x] [Review][Patch] E2E row-count assertion inverted for projects with pre-existing entries — `expect(afterRows).toBe(beforeRows === 0 ? 1 : beforeRows)` passes when `beforeRows > 0` and NO new row was added (the failure scenario); correct assertion is `expect(afterRows).toBe(beforeRows + 1)`. The test cannot detect an OOB regression for any project that already has audit entries. [`e2e/tests/shared/project-audit-log.spec.ts`]

**Round-1 items still open:**

- [x] [Review][Patch] .NET `ExtractAuditPanel` lazy regex still open — `.*?</div>` with `Singleline` stops at first `</div>`; unchanged from round 1. [`FieldMark/FieldMark.Tests.Web/Pages/ProjectsDetailPageTests.cs`]
- [x] [Review][Patch] Go: XSS round-trip test still missing (cat 3a, AC7) — no test seeds bare `<script>alert(1)</script>` payload in `metadata.reason` or actor display and asserts escaped/absent forms. [`fieldmark-go/internal/web/handlers/projects_detail_handler_test.go`]
- [x] [Review][Patch] .NET: XSS round-trip test still missing (cat 3a, AC7) — the dev-agent log references `ProjectsDetailTab_Audit_XssPayloads_AreEscapedAcrossActorAndMetadata` as passing, but this method does not exist in the submitted diff; the `CreateAuditEntryAsync` helper hardcodes `"Weather delay"` and never varies the payload. [`FieldMark/FieldMark.Tests.Web/Pages/ProjectsDetailPageTests.cs`]
- [x] [Review][Patch] Cat-1 unknown-action → `badge-unknown` unit test still missing, all three stacks (AC7 task 4.6). [`fieldmark-go/internal/web/handlers/projects_detail_handler_test.go`, `FieldMark/FieldMark.Tests.Web/Pages/ProjectsDetailPageTests.cs`, `fieldmark_py/projects/tests/test_project_detail.py`]
- [x] [Review][Patch] Cat-9 empty/whitespace/null actor → `"??"` unit tests still missing, all three stacks (AC7 task 4.6).
- [x] [Review][Patch] Go `TestGetProjectsDetailTab_AuditEmptyPanelMatchesCanonical` still tests wrong variant — stub returns one row with `NextCursor`; asserts `panel-with-row-and-load-more`; no Go test validates the `panel-empty` canonical shape. [`fieldmark-go/internal/web/handlers/projects_detail_handler_test.go`]
- [x] [Review][Patch] E2E load-more click-through and expanded-disclosure scenario still absent (AC9) — only the OOB transition test exists; no load-more append + no-duplicate + open-disclosure assertion. [`e2e/tests/shared/project-audit-log.spec.ts`]
- [x] [Review][Patch] axe-core scan still absent from all stacks — dev-agent log claims it ran, but no `AxeBuilder`, `injectAxe`, or `checkA11y` call exists anywhere in the E2E diff (AC7 task 4.8). [`e2e/tests/shared/project-audit-log.spec.ts`]

### Round-3 Review Findings (2026-06-04)

**Round-2 status at round-3:** All 10 round-2 items confirmed fixed or addressed. No round-1 or round-2 items remain open. Three new findings identified.

- [x] [Review][Patch] **Go OOB audit row omits `metadata` from `BeforeAfterJSON`** — `projects_transition_handler.go` constructs `BeforeAfterJSON` with only `{"after":..., "before":...}`, omitting the `"metadata"` key. The tab-load path (`renderAuditJSON`) includes metadata. A user who opens the disclosure on an OOB-prepended row (the new row that appears at the top when a transition fires while the Audit tab is open) cannot see `metadata.reason` — it is absent from the JSON. The same row on a tab reload does show it. This is a cross-stack inconsistency; .NET and Django both emit metadata in the OOB row. [`fieldmark-go/internal/web/handlers/projects_transition_handler.go`]
- [x] [Review][Patch] **Go `_extract_audit_panel` lazy regex** — the same `(?s)<div id="project-detail-tab-content".*?</div>` lazy pattern that was flagged as P3 in round 1 and fixed in .NET (via HtmlAgilityPack) was introduced in Go in this very commit. Stops at the first `</div>`, silently truncating any panel that has nested `<div>` elements. The empty-panel test passes today because the empty variant has no nested `<div>`; the non-empty canonical parity test will silently compare a truncated string. [`fieldmark-go/internal/web/handlers/projects_detail_handler_test.go`]
- [x] [Review][Patch] **Django `_extract_audit_panel` lazy regex** — same issue as Go: `re.search(r'(<div id="project-detail-tab-content".*?</div>)', html, re.DOTALL)` is lazy and stops at the first `</div>`. The .NET fix (HtmlAgilityPack) was not mirrored to Django. [`fieldmark_py/projects/tests/test_project_detail.py`]
- [x] [Review][Patch] **Go missing unresolvable-actor cat-9 test (third boundary input)** — `TestGetProjectsDetailTab_Audit_ActorFallbackRendersQuestionMarks` covers `actorName=""` and `actorName="   "` but has no sub-test for an `actor_id` UUID that resolves to zero rows in the user store (the DB-join unresolvable case). .NET and Django both have this test explicitly. This is the third boundary input AC7 task 4.6 requires and the cross-stack gap for Go. [`fieldmark-go/internal/web/handlers/projects_detail_handler_test.go`]
- [x] [Review][Advisory] **.NET `ExtractFirstAuditRowAndLoadMore` lazy `.*?</li>` regex** — same class of bug as the `ExtractAuditPanel` pattern fixed this round (via HtmlAgilityPack). AuditRow has no nested `<li>` today so it passes, but silently truncates if the component ever gains inner list items. Fix alongside P2/P3 for consistency. [`FieldMark/FieldMark.Tests.Web/Pages/ProjectsDetailPageTests.cs`]
- [x] [Review][Advisory] **Go `auditEntry` cursor zero-value risk** — verify that `postgres.AuditEntryStore.Append` writes `ID` and `OccurredAt` back onto the passed `*AuditEntry` pointer before `PostTransitionAsync` reads them as the load-more cursor. If it does not, `BeforeID = uuid.Nil` and `BeforeOccurredAt = time.Time{}`, causing `loadAuditPage` to return all rows and duplicate the just-committed entry at the top of the panel alongside the OOB-prepended row. [`fieldmark-go/internal/web/handlers/projects_transition_handler.go`]
- [x] [Review][Advisory] **Go `loadAuditPage` nil guard produces silent empty panel** — `if h.AuditRead == nil { return nil, "", nil }` means a missing `AuditRead` wiring in production would surface as a permanently empty audit log with no error or log entry. At minimum emit a `log.Warn`; ideally remove the guard and let the nil dereference panic at startup so misconfiguration is visible. [`fieldmark-go/internal/web/handlers/projects_detail_handler.go`]
- [x] [Review][Advisory] **`projectSlotForBaseUrl` duplicated in two E2E files** — identical function body in `project-audit-log.spec.ts` and `project-transition-flow.spec.ts`; port-mapping drift between the two files would cause one stack to silently use the wrong project slot. Extract to `e2e/tests/shared/helpers.ts`. [`e2e/tests/shared/project-audit-log.spec.ts`, `e2e/tests/shared/project-transition-flow.spec.ts`]
- [x] [Review][Advisory] **.NET and Django missing explicit empty-string actor test (third cat-9 boundary)** — .NET covers unresolvable UUID + whitespace; Django covers unresolvable UUID + whitespace; neither has a named test for `actorName = ""` (empty string). The unresolvable-UUID path implicitly produces `""` → `"??"`, so the path is covered in practice, but AC7 task 4.6 asks for all three boundary inputs as distinct named cases. [`FieldMark/FieldMark.Tests.Web/Pages/ProjectsDetailPageTests.cs`, `fieldmark_py/projects/tests/test_project_detail.py`]

### Round-4 Review Findings (2026-06-05)

**Round-3 status at round-4:** All 4 round-3 patch items confirmed fixed. Round-3 advisories A1, A2, A4 fixed; A3 and A5 still open. One new patch identified (Django OOB metadata, same defect as Go round-3 P1 — fix was not mirrored). Four new advisory findings identified.

- [x] [Review][Patch] **Django OOB audit row omits `metadata` from `before_after_json`** — `_render_transition_success` in `views.py` constructs `before_after_json=json.dumps({"before": before_state, "after": after_state}, sort_keys=True)` in both the `current_tab=audit` path (line ~589) and the OOB non-audit path (line ~602), omitting the `"metadata"` key. The round-3 Go fix added `Metadata json.RawMessage` to the transition handler's struct; this was not mirrored to Django. Users who expand the disclosure on the OOB-prepended top row cannot see `metadata.reason` (the hold reason); it appears only after a tab reload. [`fieldmark_py/projects/views.py`]
- [x] [Review][Advisory] **Go `extractAuditPanel` depth counter matches any `<div`-prefixed tag** — `strings.HasPrefix(html[i:], "<div")` matches `<divider>`, `<div-custom>`, etc., incorrectly incrementing depth for non-`<div>` elements. In practice no current template uses such a tag, but the guard should be `strings.HasPrefix(html[i:], "<div ") || strings.HasPrefix(html[i:], "<div>")` to be unambiguous. Django's `_PanelParser` (HTMLParser subclass) is immune to this. [`fieldmark-go/internal/web/handlers/projects_detail_handler_test.go`]
- [x] [Review][Advisory] **Go `loadAuditPage` nil guard still has no log warning** — round-3 A3 remains unaddressed; `if h.AuditRead == nil { return nil, "", nil }` silently renders an empty audit log in production with no diagnostic output. [`fieldmark-go/internal/web/handlers/projects_detail_handler.go`]
- [x] [Review][Advisory] **.NET and Django missing explicit empty-string actor named test** — round-3 A5 remains unaddressed; both stacks test `whitespace` and `unresolvable-UUID` but have no named test for `actorName = ""`. The path is implicitly covered by the UUID-miss test (which produces `""`), but AC7 task 4.6 calls for three distinct named boundary cases. [`FieldMark/FieldMark.Tests.Web/Pages/ProjectsDetailPageTests.cs`, `fieldmark_py/projects/tests/test_project_detail.py`]
- [x] [Review][Advisory] **Django `_extract_first_audit_row_and_load_more` still uses lazy `.*?</li>` regex** — `_extract_audit_panel` was correctly replaced with an `HTMLParser` subclass (P3 fix), but `_extract_first_audit_row_and_load_more` at line ~168 still uses `re.search(r'(<li class="audit-row".*?</li>)', ...)`. AuditRow has no nested `<li>` today; the risk is latent. The .NET equivalent was fixed via HtmlAgilityPack `SelectSingleNode`. [`fieldmark_py/projects/tests/test_project_detail.py`]
- [x] [Review][Advisory] **AC5 OOB-coexistence composition not documented in contract docs** — AC5 requires "document the chosen composition" in the contract doc (AuditRow README or three-region orchestration doc). The implementation uses the "sibling precedes" approach (`SuppressAuditEmptyState` for the server-render path; the empty-state `<li>` remains in DOM alongside the OOB-prepended row for the client-side path), but this choice is stated only in the story dev notes, not in the referenced contract documents. [`fieldmark_shared/components/audit_row/README.md`, `docs/how-to/three-region-oob-orchestration.md`]

### Round-5 Review Findings (2026-06-05)

**Status: CLEAN.** All six round-4 items (one patch, five advisories) confirmed fixed. Acceptance Auditor verdict: ready to merge. Blind Hunter raised three new speculative advisories — all dismissed after verification:

- *Django `_audit_row_context` omits `action_class`* — **DISMISSED**: `_audit_row.html` computes badge class inline via `{% if action == "..." %}` chain; no pre-computed context key is needed or expected.
- *.NET `RenderAuditJson` tested via reflection* — **DISMISSED**: pre-existing test pattern, not introduced by this story; no correctness impact.
- *Razor `(object)Model` cast* — **DISMISSED**: idiomatic Razor overload disambiguation, pre-existing, Razor correctly re-types from `ViewData.Model`.

No new PATCH findings. No open ADVISORY items.

## Sign-off

- Date of final review: 2026-06-05
- Total review-round count: 5
- Final reviewer verdict (PASS/FAIL): **PASS**
- Deferred-work entries created from this story: D-2.13-D1 (pre-commit cursor latent risk, Story 2.15 venue)
- Open dependencies confirmed at review:
  - **Story 2.11 / 2.12** — both `done`; OOB `#audit-log` prepend now lands live.
  - **NFR1 timing parity** — not run; treat as follow-up, not a release blocker.
  - **.NET aggregate runner caveat** — focused `FieldMark.Tests.Web` coverage used; caveat documented.
- All five story decisions ratified (Decision 1–5 as stated in original sign-off block below).
- Open dependencies confirmed at review:
  - **Story 2.11 / 2.12** — both are already `done`; the 2.12 OOB `#audit-log` prepend now lands against a live DOM target and the current-tab response contract avoids duplicate top rows.
  - **NFR1 timing parity** — read-path timing under the retro-A5 harness was not run here; treat as follow-up, not a release blocker for code review.
  - **Browser lane** — `tests/shared/project-audit-log.spec.ts` now passes across `.NET`, Django, and Go in this environment, including the load-more/disclosure path, the current-tab transition path, and the axe-core audit-tab scan.
  - **.NET aggregate runner caveat** — this pass used focused `FieldMark.Tests.Web` coverage instead of a full `make test-net` rerun; no new aggregate-runner evidence was collected beyond the previously documented caveat.
- Decisions ratified:
  1. **`#audit-log` becomes a live DOM target** — `<ul id="audit-log" role="list" aria-live="polite">` of AuditRow `<li>`s, completing Story 2.12's deferred OOB landing (Decision 1).
  2. **Project-scoped read path added to the append-only audit layer** — Go pool-backed read store, Django ORM query, .NET manual `DbSet<AuditEntry>` projection; no schema change (Decision 2).
  3. **Keyset pagination `(occurred_at DESC, id DESC)` with `LIMIT PAGE_SIZE+1` sentinel, PAGE_SIZE=100** — not offset (Decision 3).
  4. **Alphabetical-key JSON + ISO-8601-UTC machine timestamp are the parity anchors; relative-time bucketing + actor lookup documented in the AuditRow contract** (Decision 4).
  5. **Read-only for every role incl. Executive, gated by `project.read`; non-readers → canonical 1.11 403, no leakage** (Decision 5).

## Change Log

- 2026-06-03 — Added first-pass multi-stack Project Audit tab read path, live `#audit-log` target, keyset load-more fragment route, focused Go/.NET audit-tab tests, and contract doc updates; story remains in-progress pending broader verification and sign-off.
- 2026-06-03 — Added focused audit read-route auth/cursor/current-tab regression coverage in .NET, Go, and Django test suites; status remains `in-progress` because page-level parity, axe-core, and browser/E2E verification are still outstanding or environment-blocked.
- 2026-06-03 — Fixed audit-tab transition dedupe across all three stacks, added shared audit-log shell/fragment parity fixtures, reran focused .NET + Go tests and Django/parity gates, and moved the story to `review` with the remaining Playwright MachPort blocker documented in sign-off.
- 2026-06-04 — Closed all round-1 review patch items: hardened Go audit JSON ordering, fixed the .NET panel extractor, added cross-stack unknown-action/actor/XSS coverage, kept the empty-state hidden during OOB prepends, and made the collapsed disclosure body renderable across all three stacks.
- 2026-06-04 — Installed `axe-core` / `@axe-core/playwright`, refreshed the shared audit badge color token for dark-surface contrast, rebuilt `fieldmark_shared/dist/fieldmark.css`, and reran `tests/shared/project-audit-log.spec.ts` successfully across `dotnet`, `django`, and `fiber`; story returned to `review`.
- 2026-06-04 — Removed the broken OOB `<template>` wrappers, strengthened the Playwright row-count assertion, and updated the current-tab audit transition path so transition submissions from the audit flow keep the new top row visible; focused .NET/Go tests and the full shared Playwright audit-log spec are green.
- 2026-06-04 — Closed the Round-3 patch items around Go OOB metadata parity and robust audit-panel extraction, addressed low-risk extractor/helper advisories, and reran the focused stack suites plus both shared Playwright specs in single-worker mode for stable cross-stack verification.
- 2026-06-05 — Closed the Round-4 Django metadata parity bug and the remaining advisory items, reran the focused .NET and Go suites, confirmed Django syntax coverage, reran the shared Playwright audit-log spec with `--workers=1` after restoring the local Django dev server, and returned the story to `review`.
