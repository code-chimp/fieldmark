# Story 2.12: Place-On-Hold and Resume transitions with three-region OOB orchestration

Status: ready-for-dev

Epic: 2 — Project Lifecycle & Compliance Dashboard
Source AC: [_bmad-output/planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md](../planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md) §Story 2.12
Canonical DDL: [docker/postgres/init/010_domain_tables.sql](../../docker/postgres/init/010_domain_tables.sql) — `domain.project` (58–73, `status` CHECK `{Active,OnHold,Closed}`), `domain.audit_entry`
Pattern doc owned by this story: [docs/how-to/three-region-oob-orchestration.md](../../docs/how-to/three-region-oob-orchestration.md) (currently a **skeleton**; this story populates it — AC9)

Depends on (all **done** unless noted):
- **Story 2.11** — Project Detail anchor screen (`GET /projects/<id>` dual-mode, header strip with `#compliance-tile`, TabStrip `#project-detail-tabstrip`, `#project-detail-tab-content`, empty EntityRail `#violation-detail`, the **ActionButton row** with `place-on-hold-btn` / `resume-btn` / `close-btn`, the three `project.*` action registrations, and the read-only `Can*` predicates). **⚠️ 2.11 status is `in-progress` (not `done`) at the time this story was authored** — the detail screen, registrations (`project.place_on_hold`/`resume`/`close` → `ADMIN`), and `CanPlaceOnHold/CanResume/CanClose` predicates are already implemented per stack ([Go main.go:133-135](../../fieldmark-go/cmd/web/main.go), [Django views.py:33-34](../../fieldmark_py/projects/views.py), [.NET Project.cs:27-38](../../FieldMark/FieldMark.Domain/Entities/Project.cs)). **This story does not start until 2.11 reaches `review`/`done`.** Decision 1 below corrects two integration defects in 2.11's render that this story's mutating flow depends on.
- **Story 2.2** — `domain.audit_entry` mapping + per-stack `append_audit_entry()` helper called **inside the caller's open transaction** (.NET [`IAuditAppender.Append(...)`](../../FieldMark/FieldMark.Data/Auditing/IAuditAppender.cs), Django [`append_audit_entry(*, actor_id, action, …)`](../../fieldmark_py/audit/append.py), Go [`AuditEntryStore.Append(ctx, tx, *AuditEntry)`](../../fieldmark-go/internal/data/postgres/auditentrystore.go)). The action strings **`ProjectPlacedOnHold`** and **`ProjectResumed`** already exist in the canonical list + JSON fixture and in all three per-stack enums ([audit-actions.md](../../docs/reference/audit-actions.md), [audit-actions.json](../../docs/reference/audit-actions.json)) — **do not add them; emit them.**
- **Story 2.8** — the canonical **write-handler shape** (authorize → `transaction.atomic` / `IDbContextTransaction` / `pgx.Tx` → load → entity method → `append_audit_entry` → commit) and the **422/InlineAlert re-render** pattern; mirror it for the 409 path. `Project.create(...)` is the precedent entity method ([Project.cs:47](../../FieldMark/FieldMark.Domain/Entities/Project.cs), [models.py:72](../../fieldmark_py/projects/models.py)) — `place_on_hold`/`resume` are the next behavior methods.
- **Story 2.5** — **ComplianceTile** wrapper + the `#compliance-tile` OOB-capable target (`role="status"`, `aria-live="polite"`, `aria-atomic="true"`). This story re-renders it OOB on success (score unchanged — the OOB pattern is what is exercised).
- **Story 2.4** — **InlineAlert** wrapper (`_InlineAlert.cshtml` / `_inline_alert.html` / `inline_alert.html`; props `severity, title, message, meta?`) for the 409 in-place explanation.
- **Story 2.13** (**backlog, not done**) — builds the real Audit tab and the live `#audit-log` target. **Until 2.13 lands, `#audit-log` is not present in the DOM** (2.11 renders the Audit tab as a placeholder). See Decision 4: this story still **emits** the OOB `#audit-log` fragment per the three-region contract; it lands live once 2.13 renders the target. Audit-row correctness is proven this story at the **data layer**, not by DOM landing.
- **Story 1.12** — `can(actor, action)` primitive. **Story 1.11** — unauthenticated → `/login` redirect, canonical 403 body. **Story 2.9** — the `/projects` list page whose row-click swaps `GET /projects/<id>` into `<aside id="project-detail">` ([projects_index.html:19](../../fieldmark-go/internal/web/templates/pages/projects_index.html)).

## Story

As a Project Manager (fulfilled by the **ADMIN** role in the seeded role set — Decision 3),
I want to place an Active Project on hold and resume it back to Active, each with a recorded reason and an audit row written in the same transaction,
So that the canonical **three-region OOB orchestration** (primary `#project-detail` partial + OOB `#compliance-tile` + OOB `#audit-log`) and its **negative cases** (403 → zero OOB; 409 → originating partial + InlineAlert, zero OOB) are proven on the Project aggregate before the anchor demo lands in Epic 5 (FR12, FR13, UX-DR20/22/23).

**Scope boundary.** This story produces, per stack:
- (a) Two **domain transition methods** on the `Project` entity — `place_on_hold(reason)` (`Active → OnHold`) and `resume(reason?)` (`OnHold → Active`) — each mutating `Status` and raising a **typed domain exception** when called from the wrong state (Decision 6). These are the first *mutating* behavior methods on `Project` after `create`.
- (b) Four routes per stack: `GET /projects/<id>/place-on-hold` + `GET /projects/<id>/resume` (return the **inline reason-form fragment** — Decision 2) and `POST /projects/<id>/place-on-hold` + `POST /projects/<id>/resume` (run the canonical mutating flow).
- (c) The **three-region success response** composition: re-rendered `#project-detail` body partial (innerHTML — Decision 1) **plus** OOB `#compliance-tile` **plus** OOB `#audit-log` row fragment.
- (d) The **403** (zero OOB) and **409** (originating partial re-rendered with *current* state + InlineAlert, zero OOB) responses.
- (e) The **inline reason form** revealed by the present-state Place-on-Hold / Resume buttons (Decision 2), with server-side reason validation (length cap + character-class — security-defaults cat 3).
- (f) The **Decision-1 integration corrections** to 2.11's render so the mutating target resolves in both the standalone and list-embedded contexts (stable `id="project-detail"` wrapper present in both; action buttons swap `innerHTML`, not `outerHTML`).
- (g) The populated cross-stack how-to **`docs/how-to/three-region-oob-orchestration.md`** (currently a skeleton) + a per-stack **three-region conformance test** (Decision 5 / AC9).
- (h) Per-stack tests + E2E (single round trip + timing) + `make parity`.

**Out of scope:**
- **Project close** (`project.close`, `ProjectClosed`, the closure gate) — **Epic 6**. The `close-btn` rendered by 2.11 is untouched by this story; its `hx-post` continues to 404 until Epic 6.
- **The real Audit tab and live `#audit-log` rendering** — **Story 2.13**. This story emits the OOB `#audit-log` fragment per the contract but does not build the audit list (Decision 4).
- **Compliance-score recompute** — hold/resume do **not** change the score. The OOB `#compliance-tile` re-renders the *unchanged* score (the OOB mechanism, not a score mutation, is the deliverable). Score recompute lands with violation void / CA approve (Epics 4/5).
- **EntityRail row-selection / population** — the rail stays empty (Epic 3/4).
- **Broadening the grant beyond ADMIN** — Decision 3 keeps the 2.11 `ADMIN`-only registrations. If a distinct `ProjectManager` role is ever seeded it follows the Change Procedure.
- Any `domain.*` schema change (`pg_indexes` zero-diff).

---

## ⚠️ Decisions baked into this story (read first)

Each is implemented as written and listed in the Sign-off block for reviewer ratification.

1. **The mutating main-region target is a stable `#project-detail` wrapper, swapped `innerHTML` — and 2.11's render must be corrected to make this work.** This is the load-bearing decision; the three-region flow is broken without it.

   **The defect (verified in 2.11's current in-progress code):** 2.11's `place-on-hold-btn` renders `hx-target="#project-detail"` with **`hx-swap="outerHTML"`** ([action_button.html:10](../../fieldmark-go/internal/web/templates/components/action_button.html)), but the detail **body partial** (`projects_detail_body`) starts with `<section id="project-header-strip">` and carries **no `id="project-detail"` wrapper** ([projects_detail_body.html](../../fieldmark-go/internal/web/templates/pages/projects_detail_body.html)). Consequences:
   - **Standalone `/projects/<id>`:** `#project-detail` does not exist in the page (only `<main>` wraps the body). The POST response has **no swap target → silently dropped → nothing updates.**
   - **List-embedded (`/projects` rail):** `#project-detail` *is* `<aside id="project-detail" role="region" aria-label="Project detail" tabindex="-1">`. An `outerHTML` swap **destroys that `<aside>`** (id, role, aria-label, tabindex) and replaces it with the body partial — breaking the rail for every subsequent interaction.

   **The fix (owned by this story):**
   - The detail screen exposes a **stable wrapper element carrying `id="project-detail"` in *both* render modes.** Standalone full page: the `<main>` (or a `<div>` directly inside it) carries `id="project-detail"`. List-embedded: the existing `<aside id="project-detail">` already is that wrapper. The **body partial is always the *inner* content** of `#project-detail` (header strip + `.project-detail-grid` + rail), never the wrapper itself.
   - The action buttons (and the reason-form submit) use **`hx-target="#project-detail"` `hx-swap="innerHTML"`** (correcting 2.11's `outerHTML`). The POST success/409 response is the **inner body partial** (which re-renders header strip + tabs + Summary panel + rail) — re-rendering inside the persistent wrapper in both contexts, never destroying it.
   - This matches the [three-region how-to](../../docs/how-to/three-region-oob-orchestration.md) (`hx-target="#project-detail"`, `hx-swap="innerHTML"`) and the UX spec ("re-render `#project-detail`").

   **If 2.11 already shipped these (wrapper id + `innerHTML`) by the time this story runs, Task 0 is a no-op verification.** Otherwise this story corrects them. Either way, the integration is verified by exercising the `/projects` list row-click **and** the standalone page — not just one.

2. **The transition is a two-step reveal: present button `hx-get`s the inline reason form; the form `hx-post`s the transition.** The epic AC requires "an inline form expands requesting a reason; submission fires `POST …`". Mechanism (mirrors 2.8's `GET /projects/new` form → `POST /projects/`):
   - The **present-state** Place-on-Hold / Resume button uses `hx-get="/projects/<id>/place-on-hold"` (resp. `/resume`), `hx-target` a small inline slot inside the Summary panel's action area (e.g. `#project-action-form`), `hx-swap="innerHTML"`. (This supersedes 2.11's direct `hx-post` on these two buttons — the **`close-btn` is unchanged**, it remains a direct affordance Epic 6 will wire.)
   - The **`GET`** endpoint returns the **reason-form fragment**: a `<form hx-post="/projects/<id>/place-on-hold" hx-target="#project-detail" hx-swap="innerHTML" hx-disabled-elt="find button[type=submit]">` with a `<label>`+`<textarea name="reason">`, the per-stack CSRF token, a submit button, and a Cancel control that clears the slot. The form **also** works without JS (`method="post" action="/projects/<id>/place-on-hold"`).
   - The **`POST`** endpoint runs the canonical mutating flow (AC3).
   - Disabled-state and absent-state buttons are unchanged (trichotomy from 2.11). A disabled button never reveals a form.

3. **"Project Manager" = the `ADMIN` role; the grants are unchanged from 2.11.** The seeded role set is `{ADMIN, EXECUTIVE, INSPECTOR, SITE_SUPERVISOR, COMPLIANCE_OFFICER}` — there is **no distinct `ProjectManager` role** (consistent with Story 2.8 "PM/Admin" → `ADMIN`, and 2.11's 5-role trichotomy test). 2.11 already registered `project.place_on_hold` / `project.resume` → `ADMIN` only ([main.go:133-134](../../fieldmark-go/cmd/web/main.go), [views.py:33-34](../../fieldmark_py/projects/views.py)). **This story does not re-register or broaden them.** The persona "Project Manager" in the epic maps to `ADMIN`. The authz tests assert `ADMIN` permitted, the other four roles 403.

4. **The OOB `#audit-log` fragment is emitted per the contract, but does not land live until Story 2.13.** 2.11 renders the Audit tab as a placeholder; the live `#audit-log` target is built by 2.13 (backlog). HTMX **silently drops** an `hx-swap-oob` whose target id is absent from the current DOM — so the emitted `#audit-log` fragment no-ops in the live app this story. This is correct, sanctioned, and documented:
   - The three-region **response body** always contains all three regions (the conformance test asserts this at the string level — AC9 — independent of DOM presence).
   - **Audit-row correctness is proven at the data layer** this story: a per-stack test asserts exactly one `domain.audit_entry` row was committed with the right `action` / `before_state` / `after_state` / `metadata.reason` (AC6).
   - The epic AC "open the Audit tab → new row at top" becomes fully E2E-verifiable **once 2.13 lands**; this story notes the dependency and does not block on it.
   - **Do not invent a temporary `#audit-log` target** in the Summary panel to make the OOB land — that would be thrown away by 2.13 and would violate the canonical layout.

5. **Three-region response shape is a documentation contract + per-stack native composition + per-stack conformance test** (Cross-Stack Architecture Principle). This story populates the skeleton [docs/how-to/three-region-oob-orchestration.md](../../docs/how-to/three-region-oob-orchestration.md) and ships a conformance test per stack (AC9). No shared template fragment; each stack composes natively (Razor partials, Django `{% include %}`, Go `html/template` blocks).

6. **The wrong-state transition raises a typed domain exception → HTTP 409.** `place_on_hold` raises when `Status != Active`; `resume` raises when `Status != OnHold`. The exception is a **named domain type** per stack (not a bare `ValueError`/`ArgumentException`), so the handler can catch *only* it and map to 409 (a generic catch would mask bugs). Names: .NET `InvalidProjectTransitionException`, Django `InvalidProjectTransition(DomainError)`, Go `ErrInvalidProjectTransition` (sentinel, wrapped with context). The exception **message is user-visible** in the InlineAlert and is part of the cross-stack parity contract: `"Project is already on hold"` (place-on-hold from non-Active) / `"Project is not on hold"` (resume from non-OnHold).

---

## Acceptance Criteria

### AC1 — Reason-form reveal (`GET`) on the present-state buttons (Decision 2)

**Given** I am authorized (`can(actor, "project.place_on_hold")`) and the project is **Active**
**When** I click the present-state **Place on Hold** button
**Then** HTMX fires `hx-get="/projects/<id>/place-on-hold"` into the inline action slot (`#project-action-form`, `hx-swap="innerHTML"`)
**And** the response is a `role="form"`-bearing fragment containing a `<label for="reason">` + `<textarea id="reason" name="reason" required maxlength="…">`, the per-stack CSRF/antiforgery token (Go: none per ADR-012), a `<button type="submit">` (e.g. "Place on hold"), and a Cancel control that empties the slot (`hx-get` a blank fragment, or `hx-on` clear — pick one mechanism, identical across stacks)
**And** the `<form>` carries `hx-post="/projects/<id>/place-on-hold"`, `hx-target="#project-detail"`, `hx-swap="innerHTML"`, `hx-disabled-elt="find button[type=submit]"`, **and** the no-JS fallback `method="post" action="/projects/<id>/place-on-hold"`.

**Given** the same for **Resume** when the project is **OnHold** (`/projects/<id>/resume`), with the reason `<textarea>` present (epic: resume reason is optional metadata, but the form still collects it — `required` may be omitted on resume; document the choice and keep it identical across stacks).

**Given** I lack the permission, or the project is not in the action's source state
**Then** the button is **absent** or **disabled** (2.11 trichotomy) and reveals nothing; a direct `GET` to the reason-form endpoint by an unauthorized user returns **403** (AC4 shape).

### AC2 — `place_on_hold` / `resume` domain transition methods (Decision 6)

**Given** the `Project` entity per stack
**When** `place_on_hold(reason)` is called on an **Active** project
**Then** `Status` becomes `OnHold` (pure in-memory mutation; the handler persists) and the method returns enough to snapshot `before`/`after` (or the handler snapshots around the call — pick the per-stack idiom)
**And** calling it on a non-Active project raises the typed `InvalidProjectTransition*` exception with message `"Project is already on hold"`.

**When** `resume(reason?)` is called on an **OnHold** project
**Then** `Status` becomes `Active`; calling it on a non-OnHold project raises with message `"Project is not on hold"`.

**And** unit tests per stack cover: `Active --place_on_hold--> OnHold`; `OnHold --resume--> Active`; and the raise on every other `{Active, OnHold, Closed}` source state for each method (cat-9 boundary discipline — all states, not just the happy one).
**And** these are **pure domain methods** — no DB, no audit, no transaction inside the entity (same discipline as `Project.create`).

### AC3 — `POST` happy path: canonical flow + three-region response (UX-DR20/23, FR40/57)

**Given** I am authorized and the project is in the valid source state, and I submit a valid reason
**When** `POST /projects/<id>/place-on-hold` (or `/resume`) runs
**Then** the handler executes the canonical write flow **in this order**: authorize (`can`) → **open exactly one transaction** → load the `Project` aggregate → call `project.place_on_hold(reason)` (resp. `.resume(reason)`) → `append_audit_entry(action=ProjectPlacedOnHold|ProjectResumed, actor, entity_type="Project", entity_id=<id>, project_id=<id>, before_state, after_state, metadata={"reason": <validated reason>})` **inside the same transaction** → persist the project row → **commit**
**And** the **response body contains exactly three regions**:
  1. **Main** — the re-rendered `#project-detail` **inner body partial** (header strip with the **new StatusBadge**, the `.project-detail-grid`, Summary panel with the **flipped trichotomy** — e.g. after place-on-hold: Place-on-Hold now disabled, Resume now present — and the empty rail), targeted by the request's `hx-target="#project-detail"` `hx-swap="innerHTML"`.
  2. **OOB `#compliance-tile`** — `<section id="compliance-tile" hx-swap-oob="true" role="status" aria-live="polite" aria-atomic="true">` re-rendered with the **unchanged** score (structural re-render only — Decision: score is not mutated by hold/resume).
  3. **OOB `#audit-log`** — the new AuditRow fragment with `hx-swap-oob="afterbegin"` (prepend) targeting `#audit-log` (lands live once 2.13 renders the target — Decision 4).
**And** the HTTP status is **200** and there is **exactly one** HTTP round trip (no follow-up request).
**And** focus is managed per the existing swap convention (the re-rendered `#project-detail` / its focusable surface) so the change is announced (WCAG 2.1 AA across swaps).

### AC4 — `403` unauthorized: zero OOB, no state leakage (UX-DR22, FR7/56)

**Given** the requester lacks the action permission (any of the four non-ADMIN roles, or the no-role `testuser`)
**When** they `GET` or `POST` either transition endpoint
**Then** the response is **HTTP 403** with the canonical 403 body (Story 1.11 shape — do not invent a new one), **no entity state leaked** (no project fields, no "exists but forbidden" signal)
**And** the response contains **zero** OOB regions (no `#compliance-tile`, no `#audit-log`) — assert `hx-swap-oob` absent from the body.

**Given** an **unauthenticated** request to any of the four endpoints
**Then** the Story 1.11 redirect-to-login fires first (302/303 → `/login`), unchanged.

### AC5 — `409` rule violation: originating partial + InlineAlert, zero OOB (UX-DR22, FR55)

**Given** I am authorized but the project is in the wrong source state (e.g. `POST /projects/<id>/place-on-hold` on an already-`OnHold` project — a stale page or a concurrent transition)
**When** the entity method raises `InvalidProjectTransition*` and the exception bubbles to the handler
**Then** the handler **catches only that typed exception**, rolls back the transaction (no audit row written), and returns **HTTP 409**
**And** the body is the **same `#project-detail` inner partial re-rendered showing the project's *current* (unchanged) state**, with an **InlineAlert** (`severity="danger"`, `role="alert"`, `title` e.g. "Couldn't place the project on hold", `message` = the exception's user-visible message) rendered at the top of the Summary panel
**And** the response contains **zero** OOB regions (no `#compliance-tile`, no `#audit-log`) — assert `hx-swap-oob` absent.

**Given** a missing/invalid project id
**Then** the response is **HTTP 404** with no state leakage (no OOB).

**Given** a missing/blank/over-length/invalid-character `reason` on place-on-hold (where required)
**Then** the response is **HTTP 422** with the reason-form fragment re-rendered showing the InlineAlert + `aria-invalid="true"` + `aria-describedby` on the `<textarea>` (mirror 2.8's 422 shape), **zero OOB** (a validation failure is not a transition).

### AC6 — Audit-row correctness (data layer; cross-stack-identical strings)

**Given** a successful place-on-hold (resp. resume)
**When** I inspect `domain.audit_entry`
**Then** exactly **one** new row exists with: `action = "ProjectPlacedOnHold"` (resp. `"ProjectResumed"`) — verbatim PascalCase per [audit-actions.md](../../docs/reference/audit-actions.md); `actor_id` = the request user; `entity_type = "Project"`; `entity_id = project_id = <id>`; `before_state` = JSON snapshot with `status` = the pre-transition value; `after_state` = JSON snapshot with `status` = the post-transition value; `metadata = {"reason": "<validated reason>"}`
**And** `before_state` / `after_state` use **alphabetical key ordering** for byte-stable cross-stack snapshots (same convention as Story 2.8 `ProjectCreated`); document the exact key set in the contract doc.
**And** a per-stack test asserts the row's fields (this is the AC4-epic "audit row at top of Audit tab" guarantee, proven at the data layer this story per Decision 4; the rendered-tab assertion is deferred to Story 2.13's E2E).
**And** on **403** and **409**, **no** `domain.audit_entry` row is written (assert count unchanged).

### AC7 — `docs/how-to/three-region-oob-orchestration.md` populated (Decision 5, AC-epic)

**Given** the skeleton at [docs/how-to/three-region-oob-orchestration.md](../../docs/how-to/three-region-oob-orchestration.md)
**When** I read it after this story
**Then** every `TODO` is replaced and it specifies: **when to use** the pattern (a domain mutation affecting the entity partial + the header tile + the audit log; canonical first use = place-on-hold/resume); the **success composition** (worked example with the real `#project-detail` inner partial + OOB `#compliance-tile` + OOB `#audit-log` afterbegin, `hx-target="#project-detail"` `hx-swap="innerHTML"`); the **negative-case table** (200→3 regions, 403→0, 409→0 originating partial + InlineAlert, 422→0 form partial); the **`#audit-log` deferred-landing note** (Decision 4 — emitted always, lands live at 2.13); the **per-stack native composition** notes (Razor / Django includes / Go blocks, no shared fragment); the **conformance-test contract** (AC9); and the **timing contract** (NFR1).
**And** the top-of-file Status line is updated from "skeleton" to "live — populated by Story 2.12, 2026-05-31".

### AC8 — Security-defaults + edge-case checklist coverage

**Given** [security-defaults.md](../../docs/reference/security-defaults.md) **cat 3 (strict allowlist on user-controlled writes)**
**Then** `reason` is **free-form text** → validated before any write with a **length cap** (e.g. ≤ 2000 chars; pick a value, document it, identical across stacks) **and** a character-class guard (reject control characters); the **validated** value is persisted to `metadata.reason`, never the raw input. Over-length/invalid → 422 (AC5).

**Given** **cat 6 (CSRF posture)**
**Then** the two `POST` endpoints are state-changing → each carries its stack's existing CSRF posture: .NET `[ValidateAntiForgeryToken]` (token via the form's hidden input), Django CSRF middleware (token via the form / `hx-headers`), Go documented ADR-012 exemption. The HTMX form includes the token exactly as the 2.8 create form does.

**Given** **security-defaults cat 3a (XSS round-trip on render)**
**Then** the `reason` echoed into the InlineAlert/audit-influenced render goes through each engine's **default escaping** (no `Html.Raw`/`|safe`/`template.HTML`); a per-stack test passes the bare payload `<script>alert(1)</script>` as `reason`, drives the flow, and asserts **both** `Contains("&lt;script&gt;alert(1)&lt;/script&gt;")` **and** `NotContains("<script>")` wherever the reason renders.

**Given** [component-edge-case-checklist.md](../../docs/reference/component-edge-case-checklist.md): **cat 1** (StatusBadge always hits a known `{Active,OnHold,Closed}` value — DB-CHECK-constrained, 2.4 `badge-unknown` is the safety net); **cat 3** (no new JS — the reveal/post is HTMX; with JS off the form posts via `method/action` and the page navigates, still functional); **cat 7/8** (reduced-motion / forced-colors handled by 1.14 globals; the StatusBadge/tile pair colour with text). An **axe-core** scan on the post-transition `#project-detail` render reports **zero** new WCAG 2.1 AA violations, including the InlineAlert `role="alert"` on the 409 path.

### AC9 — Per-stack three-region conformance test (Decision 5)

**Given** each stack's test suite
**When** I run the Project three-region conformance test
**Then** it exercises three flows and asserts response **shape** (string/parse level, DB-independent where possible, integration-tagged where a real transition is needed):
  1. **Success (200)** — body contains **exactly** the three documented regions: main `#project-detail` (no `hx-swap-oob` on it), OOB `#compliance-tile` (`hx-swap-oob` present), OOB `#audit-log` (`hx-swap-oob="afterbegin"`). Count of OOB regions = **2**.
  2. **403** — body contains **zero** `hx-swap-oob` regions.
  3. **409** — body re-renders the `#project-detail` partial with current state + an InlineAlert (`role="alert"`), **zero** `hx-swap-oob` regions.
**And** a shared helper (per stack) that counts OOB regions in a response body is acceptable (and recommended) — document its location in the contract doc.

### AC10 — E2E single round trip + timing parity (NFR1)

**Given** a Playwright E2E scenario per stack
**When** the place-on-hold action is exercised end-to-end (login as ADMIN → open `/projects/<id>` Active → click Place on Hold → fill reason → submit)
**Then** the network panel shows **exactly one** POST (no follow-up request), the `#project-detail` partial re-renders with status `OnHold` and the flipped trichotomy, and the OOB `#compliance-tile` re-renders **in the same paint**
**And** a 409 scenario (resume an Active project via a stale/forged request) renders the InlineAlert in place with status unchanged and **no** OOB swap
**And** local-dev **p95 ≤ 200 ms** per stack with cross-stack divergence **≤ 50 ms p95** (NFR1) — capture via the timing harness if present; if the harness (retro action A5) is not yet available, record measured p95 in the sign-off and note the harness dependency rather than blocking.

### AC11 — `make parity` + full gate (no schema change)

**Given** Story 1.3 route-parity tooling
**When** I run `make parity`
**Then** all three route dumps contain `GET /projects/:id/place-on-hold`, `POST /projects/:id/place-on-hold`, `GET /projects/:id/resume`, `POST /projects/:id/resume` (stack-idiomatic param syntax), diff clean; `pg_indexes` diff is **zero** (no `domain.*` change).

**Build/type/lint/test gates green per stack:**
- **.NET:** `cd FieldMark && dotnet csharpier check . && dotnet build && dotnet test && dotnet test FieldMark.Tests.Integration/FieldMark.Tests.Integration.csproj` — clean.
- **Django:** `cd fieldmark_py && uv run ruff check . && uv run mypy . && uv run pytest && uv run pytest -m integration` — clean.
- **Go:** `cd fieldmark-go && make check && go test ./... && go test -tags=integration ./...` — clean.
- **No `fieldmark_shared` CSS change expected** (the InlineAlert/StatusBadge/tile styles already exist). If a hand-authored rule is unavoidable, rebuild `dist/fieldmark.css` and justify in dev notes; otherwise assert **zero** CSS drift.

---

## Tasks / Subtasks

- [ ] **Task 0: Decision-1 integration correction to the 2.11 render** (AC: #3, #5; prereq for everything)
  - [ ] 0.1 Ensure the standalone `GET /projects/<id>` full page wraps the body partial in an element carrying **`id="project-detail"`** (give `<main>` the id, or a `<div id="project-detail">` directly inside `<main>`); the list-embedded `<aside id="project-detail">` already is that wrapper. The **body partial stays the inner content** (header strip + `.project-detail-grid` + rail) in both modes. Per stack: Go [projects_detail_body.html](../../fieldmark-go/internal/web/templates/pages/projects_detail_body.html) + the page that wraps it; Django `templates/projects/detail.html`; .NET `Pages/Projects/Detail.cshtml`.
  - [ ] 0.2 Change the `place-on-hold-btn` / `resume-btn` present-state `hx-swap` from `outerHTML` → **`innerHTML`** (and target `#project-detail`). Note: 2.11 renders these via the shared ActionButton component with a fixed `hx-swap="outerHTML"` ([action_button.html:10](../../fieldmark-go/internal/web/templates/components/action_button.html)) — make `hx-swap` a **per-button prop** (default `outerHTML` to preserve other callers) and pass `innerHTML` for these two, **or** switch these buttons to the Decision-2 `hx-get` reveal (which makes the button's own swap irrelevant — the *form* sets `innerHTML`). Prefer the Decision-2 reveal; if so, 0.2 reduces to "the reason-form's `hx-swap` is `innerHTML`" and the button only reveals the form.
  - [ ] 0.3 Verify by exercising **both** the `/projects` list row-click → place-on-hold **and** the standalone `/projects/<id>` → place-on-hold; the `<aside id="project-detail">` (list) and `<main id="project-detail">` (standalone) must survive and re-render their inner content. Add a regression assertion that the wrapper element is not destroyed.

- [ ] **Task 1: Domain transition methods + typed exception** (AC: #2)
  - [ ] 1.1 .NET `FieldMark.Domain/Entities/Project.cs`: `public void PlaceOnHold(string reason)` (`Active`→`OnHold`, else `throw new InvalidProjectTransitionException("Project is already on hold")`); `public void Resume(string? reason)` (`OnHold`→`Active`, else `"Project is not on hold"`). Add `InvalidProjectTransitionException` (domain exception type). Pure; no persistence.
  - [ ] 1.2 Django `projects/models.py` `Project`: `def place_on_hold(self, reason: str) -> None` / `def resume(self, reason: str | None = None) -> None` raising `InvalidProjectTransition(DomainError)`. Pure mutation of `self.status`; handler calls `.save()`.
  - [ ] 1.3 Go `internal/domain/entities/project.go`: `func (p *Project) PlaceOnHold(reason string) error` / `func (p *Project) Resume(reason string) error` returning a wrapped `ErrInvalidProjectTransition` sentinel on wrong state (Go-idiomatic — error return, not panic). Mutates `p.Status`.
  - [ ] 1.4 Domain unit tests per stack: happy transition + raise on every other `{Active,OnHold,Closed}` source state, for both methods. Assert the user-visible message strings are exactly `"Project is already on hold"` / `"Project is not on hold"`.

- [ ] **Task 2: `GET` reason-form endpoints + inline slot** (AC: #1)
  - [ ] 2.1 Add `#project-action-form` inline slot in the Summary panel's action area (2.11 template) and wire the present-state Place-on-Hold / Resume buttons to `hx-get` their reason-form endpoint into it (Decision 2). Disabled/absent buttons unchanged.
  - [ ] 2.2 `GET /projects/<id>/place-on-hold` + `/resume` per stack → reason-form fragment (label + `<textarea name="reason">` + CSRF + submit + Cancel; `hx-post`/`hx-target="#project-detail"`/`hx-swap="innerHTML"`/`hx-disabled-elt`; no-JS `method`/`action`). Authorize `project.place_on_hold`/`resume` (403) and 404 on bad id. Reason `required` on hold; document the resume choice.

- [ ] **Task 3: `POST` transition handlers — canonical flow + three-region composition** (AC: #3, #5, #6, #8)
  - [ ] 3.1 `POST /projects/<id>/place-on-hold` + `/resume` per stack: authorize → validate `reason` (cat 3: length cap + char-class; 422 on fail) → open transaction → load aggregate (404 if missing) → call entity method (catch only `InvalidProjectTransition*` → rollback → 409) → `append_audit_entry(...)` with alphabetical-key `before`/`after` + `metadata.reason` → persist → commit.
  - [ ] 3.2 Success (200) response: re-render `#project-detail` inner partial (new StatusBadge + flipped trichotomy) + OOB `#compliance-tile` (unchanged score) + OOB `#audit-log` afterbegin AuditRow. Compose natively per stack (Razor partials / Django includes / Go blocks) — no shared fragment.
  - [ ] 3.3 409 response: re-render `#project-detail` inner partial with **current** state + InlineAlert (`severity=danger`, `role=alert`), **zero OOB**. 422: re-render reason-form fragment + InlineAlert + `aria-invalid`/`aria-describedby`, zero OOB. 403: canonical 1.11 body, zero OOB.
  - [ ] 3.4 Routes registered in each stack's router (Django `urls.py`, Go `cmd/web/main.go`, .NET Razor page/handler or minimal-API). Keep parity-route names.

- [ ] **Task 4: Populate the three-region how-to doc** (AC: #7)
  - [ ] 4.1 Replace every `TODO` in [three-region-oob-orchestration.md](../../docs/how-to/three-region-oob-orchestration.md): when-to-use, worked success composition (real markup), negative-case table, `#audit-log` deferred-landing note (Decision 4), per-stack native composition, conformance-test contract, timing contract. Flip the Status line to "live — Story 2.12".

- [ ] **Task 5: Per-stack three-region conformance test + handler/flow tests** (AC: #3, #4, #5, #6, #8, #9)
  - [ ] 5.1 Conformance test (AC9): success = 2 OOB regions + main; 403 = 0 OOB; 409 = originating partial + InlineAlert + 0 OOB. Add the OOB-counting helper; document its path in the contract doc.
  - [ ] 5.2 Happy-path integration test: ADMIN place-on-hold → status `OnHold`, exactly one `audit_entry` row with correct fields (AC6), one round trip. Resume mirror.
  - [ ] 5.3 403 (each non-ADMIN role + no-role) → 403, zero OOB, zero audit rows. Unauthenticated → login redirect.
  - [ ] 5.4 409 (place-on-hold on OnHold; resume on Active) → 409, originating partial + InlineAlert, zero OOB, zero audit rows. 404 (bad id). 422 (blank/over-length/control-char reason), zero OOB/audit.
  - [ ] 5.5 XSS round-trip (cat 3a): `<script>` in `reason`, both assertions, wherever reason renders.

- [ ] **Task 6: E2E + parity + gate** (AC: #10, #11)
  - [ ] 6.1 E2E place-on-hold (one POST, `#project-detail` + OOB tile in one paint, flipped trichotomy) + 409 (InlineAlert in place, status unchanged, no OOB), per stack. Capture p95 timing (NFR1) or note the harness dependency.
  - [ ] 6.2 `make parity` (four new routes present, `pg_indexes` zero-diff) + full per-stack gate green; assert zero `fieldmark_shared` CSS drift.

- [ ] **Task 7: Story sign-off** (AC: all)
  - [ ] 7.1 Populate the Sign-off block; record the six decisions; note the Story 2.13 dependency for live `#audit-log` landing and the NFR1 timing-harness status; flip sprint-status to `review`.

## Dev Notes

### Critical context (read before writing code)

- **This is a *mutating* story — the opposite discipline from 2.10/2.11 reads.** Authorize → **one transaction** → load → entity method → **`append_audit_entry` in the same transaction** → commit. Mirror Story 2.8's create handler exactly; the only differences are the entity method (`place_on_hold`/`resume` vs `create`), the action string, and the **three-region response** instead of an `HX-Redirect`.
- **Decision 1 is the highest-risk item — verify it first (Task 0).** The 2.11 render as it currently stands makes the mutating target unreachable on the standalone page and destructive on the list page. Do not write the POST handler before the `#project-detail` wrapper + `innerHTML` swap are corrected and verified in **both** contexts. This is the single most likely source of a "works in the test, broken in the app" failure.
- **`#audit-log` does not exist in the live DOM until Story 2.13 (Decision 4).** Emit the OOB fragment anyway (the contract + conformance test require it). Prove the audit row at the data layer (AC6). Do **not** fabricate a temporary `#audit-log` to make the OOB land — 2.13 owns that target.
- **Score is unchanged by hold/resume.** The OOB `#compliance-tile` re-renders the *same* score. The deliverable is the OOB *mechanism* (one round trip touches the tile region), not a score change. Do not call any score-recompute path.
- **Catch only the typed transition exception for 409 (Decision 6).** A broad `except Exception` / `catch` would map genuine bugs (a null-ref, a DB error) to 409 and hide them. Catch `InvalidProjectTransition*` specifically; let everything else 500.
- **`reason` is user input → validate before write (security-defaults cat 3).** Length cap + control-char reject; persist the validated value into `metadata.reason`. The reason renders back through default escaping (cat 3a XSS test).
- **"Project Manager" = ADMIN; do not register or broaden grants (Decision 3).** 2.11 already registered `project.place_on_hold`/`resume` → `ADMIN`. This story *uses* them. Adding a `ProjectManager` role or broadening here is out of scope.
- **Action strings already exist — emit, don't add.** `ProjectPlacedOnHold` / `ProjectResumed` are in the canonical list + JSON fixture + all three enums (Story 2.2). The audit conformance test (Story 2.2) already asserts set equality; adding a duplicate would break it.
- **`before`/`after` JSON uses alphabetical keys** (Story 2.8 convention) for byte-stable cross-stack snapshots. `metadata = {"reason": "<validated>"}`.
- **Two-step reveal (Decision 2), not a direct POST.** The present button `hx-get`s the reason form; the form `hx-post`s the transition. The `close-btn` is untouched (Epic 6).

### Source tree — where things land

| Stack | Transition methods + exception | GET form + POST handlers + routes | Templates (form fragment, 3-region response, 409 partial) |
|---|---|---|---|
| .NET | `FieldMark.Domain/Entities/Project.cs` + `InvalidProjectTransitionException` | `Pages/Projects/` handlers (or minimal-API `MapPost`) + route registration | Razor partials under `Pages/Projects/` / `Pages/Shared/` (reason form, `#project-detail` inner partial, OOB tile, OOB audit row, InlineAlert) |
| Django | `projects/models.py` `Project` + `InvalidProjectTransition` (in `projects/` domain-errors module) | `projects/views.py` + `projects/urls.py` | `templates/projects/` partials (`{% include %}` composition) |
| Go | `internal/domain/entities/project.go` + `ErrInvalidProjectTransition` | `internal/web/handlers/projects_transition_handler.go` (or extend the detail handler) + routes in `cmd/web/main.go` | `internal/web/templates/pages/` blocks (reason form fragment, three-region response, 409 partial) |

Doc: [docs/how-to/three-region-oob-orchestration.md](../../docs/how-to/three-region-oob-orchestration.md) (populate). No `fieldmark_shared` change expected.

### Existing code to reuse (read before writing)

- **Write-handler precedent:** Story 2.8 create handler (tx + `append_audit_entry` + 422/InlineAlert re-render) — the closest template. Entity-method precedent: `Project.create` ([Project.cs:47](../../FieldMark/FieldMark.Domain/Entities/Project.cs), [models.py:72](../../fieldmark_py/projects/models.py)).
- **Audit helper:** .NET [`IAuditAppender.Append(...)`](../../FieldMark/FieldMark.Data/Auditing/IAuditAppender.cs), Django [`append_audit_entry(*, …)`](../../fieldmark_py/audit/append.py), Go [`AuditEntryStore.Append(ctx, tx, *AuditEntry)`](../../fieldmark-go/internal/data/postgres/auditentrystore.go). All called inside the caller's open transaction.
- **Components:** ComplianceTile (2.5, `#compliance-tile` OOB target), StatusBadge (2.4), InlineAlert (2.4 — [_InlineAlert.cshtml](../../FieldMark/FieldMark.Web/Pages/Shared/Components/_InlineAlert.cshtml) / [_inline_alert.html](../../fieldmark_py/templates/components/_inline_alert.html) / [inline_alert.html](../../fieldmark-go/internal/web/templates/components/inline_alert.html)), AuditRow (2.4), ActionButton (1.12). Compose; do not re-emit inner markup.
- **Detail render (2.11):** [projects_detail_body.html](../../fieldmark-go/internal/web/templates/pages/projects_detail_body.html), [projects_detail_handler.go](../../fieldmark-go/internal/web/handlers/projects_detail_handler.go), [views.py project_detail](../../fieldmark_py/projects/views.py), `Pages/Projects/Detail.cshtml`. The Decision-1 corrections live here.
- **`can()` / registrations:** Story 1.12 + 2.11 ([main.go:133-135](../../fieldmark-go/cmd/web/main.go), [views.py:33-34](../../fieldmark_py/projects/views.py)).
- **Enums:** `ProjectStatus {Active,OnHold,Closed}` ([project_status.go:12](../../fieldmark-go/internal/domain/enums/project_status.go), [ProjectStatus.cs:7](../../FieldMark/FieldMark.Domain/ValueObjects/ProjectStatus.cs), [models.py:24](../../fieldmark_py/projects/models.py)); `AuditAction.ProjectPlacedOnHold`/`ProjectResumed`.

### Project Structure Notes

- Adds **4 routes** to the parity inventory (`GET`+`POST` × hold/resume); no `domain.*` schema change (`pg_indexes` zero-diff).
- First **mutating** transition methods on `Project` after `create`; first use of the **three-region OOB** pattern (populates the how-to doc the rest of Epics 4–6 reuse for Approve/Resolve/Void/Close).
- Decision 1 touches the 2.11 detail render (wrapper id + swap mode) — a small, contained correction, but verify it does not regress the 2.11 dual-mode tests or the 2.9 list row-click.
- `#audit-log` OOB lands live only after Story 2.13; tracked in sign-off.

### References

- Epic AC: [epic-2 §Story 2.12](../planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md)
- Pattern doc (owned): [three-region-oob-orchestration.md](../../docs/how-to/three-region-oob-orchestration.md); UX journeys + patterns: [ux-design-specification.md](../planning-artifacts/ux-design-specification.md) (Journey 3, Journey Patterns §1–2, lines 798–803)
- Audit contract: [audit-actions.md](../../docs/reference/audit-actions.md), [audit-actions.json](../../docs/reference/audit-actions.json)
- Write precedent: [2-8 create form](2-8-project-create-form-pm-admin.md); prior screen: [2-11 Project Detail](2-11-project-detail-anchor-screen-with-header-strip-tabstrip-and-entityrail.md)
- Edge cases / security: [component-edge-case-checklist.md](../../docs/reference/component-edge-case-checklist.md) cat 1/3/7/8, [security-defaults.md](../../docs/reference/security-defaults.md) cat 3/3a/6
- DDL: [010_domain_tables.sql](../../docker/postgres/init/010_domain_tables.sql) (`domain.project.status` CHECK, `domain.audit_entry`)

## Dev Agent Record

### Agent Model Used

### Debug Log References

### Completion Notes List

### File List

## Sign-off

- Date of final review: _pending_
- Total review-round count: _pending_
- Final reviewer verdict (PASS/FAIL): _pending_
- Deferred-work entries created from this story: _pending_
- Open dependencies to confirm at review:
  - **Story 2.11 must be `review`/`done` before this story starts** (detail screen, registrations, predicates). _pending_
  - **Live `#audit-log` OOB landing depends on Story 2.13** (backlog) — this story emits the fragment + proves the row at the data layer (Decision 4). _pending_
  - **NFR1 timing parity** depends on the retro action-A5 timing harness if not yet present — record measured p95 if the harness is unavailable. _pending_
- Decisions requiring ratification (confirm or overturn at review):
  1. **Mutating target = stable `#project-detail` wrapper, `hx-swap="innerHTML"`** — corrects 2.11's `outerHTML` + missing standalone wrapper id (Decision 1). _pending_
  2. **Two-step reveal**: present button `hx-get`s the inline reason form; form `hx-post`s the transition (Decision 2). _pending_
  3. **"Project Manager" = ADMIN**; grants unchanged from 2.11 (Decision 3). _pending_
  4. **OOB `#audit-log` emitted per contract but lands live only at Story 2.13**; audit row proven at the data layer this story (Decision 4). _pending_
  5. **Three-region shape = doc contract + native composition + per-stack conformance test** (Decision 5). _pending_
  6. **Wrong-state transition raises a typed domain exception → 409**, message part of the parity contract (Decision 6). _pending_
