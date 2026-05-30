# Project Create Form — Cross-Stack Contract

<!-- See: docs/reference/project-create-form-contract.md -->
<!-- Populated by Story 2.8, 2026-05-30 -->

## Status

Populated by Story 2.8, 2026-05-30. Follows the same conventions as
`docs/reference/audit-actions.md`.

## Why This Document Exists

The root `CLAUDE.md` §"Form-contract corollary" (ratified in the Epic 1
retrospective, 2026-05-25) requires that any form appearing in ≥ 2 stacks must
have its field names, hidden-input names, and return-target conventions recorded
in a contract doc. The login open-redirect bug and the cookie-regex injection
(both discovered during Epic 1 code review) both traced to uncoordinated
per-stack interpretations of a form's field names and redirect parameters. This
document prevents that class of bug for the project-create form.

## Routes

| Route | Method | Purpose |
|---|---|---|
| `/projects/new` | `GET` | Render the empty create form. |
| `/projects/` | `POST` | Submit the form; creates the project and redirects. |

The **trailing slash on POST** is canonical and deliberate — it mirrors the
Django idiom and the epic AC text. Per-stack handlers must accept the
trailing-slash form at the minimum. The no-slash form (`POST /projects`) is
acceptable as an alias if the framework normalises it without an extra
round-trip redirect; but the canonical advertised URL in the form `action`
attribute is always `/projects/`. Do not emit a 308 permanent redirect to
normalise slashes — that adds a spurious round-trip for HTMX partial forms.

## HTTP Method Matrix

| Request | Expected response |
|---|---|
| `GET /projects/new` — authenticated ADMIN | 200 + form HTML |
| `GET /projects/new` — unauthenticated | 302/303 → `/login?return_url=/projects/new` |
| `GET /projects/new` — authenticated non-ADMIN | 403 |
| `POST /projects/` — valid input, ADMIN | 200 + `HX-Redirect: /projects/<id>` header (or 303 `Location: /projects/<id>` for non-HTMX) |
| `POST /projects/` — invalid input | 422 + re-rendered form partial |
| `POST /projects/` — unauthenticated | 302/303 → `/login` (no return_url for POST) |
| `POST /projects/` — non-ADMIN | 403 |
| `POST /projects/new` | 405 Method Not Allowed |
| `GET /projects/` | 405 Method Not Allowed with `Allow: POST` header |
| `GET /projects` (no slash) | 404 (Story 2.9 will register this route) |

### Framework-specific 405 behaviour

- **.NET Razor Pages**: the `Index.cshtml.cs` PageModel at `/projects/` exposes
  only `OnGetAsync` that returns `StatusCode(405)` with
  `Response.Headers.Allow = "POST"`. There is no POST handler at this URL in
  the page model; the form `POST /projects/` is handled by a separate minimal
  API endpoint or a second page at a different path — see Task 4.3 in the
  story.
- **Django**: `@require_http_methods(["POST"])` on the `POST /projects/`
  endpoint returns 405 natively with the `Allow` header.
- **Go / Fiber**: `app.Post("/projects/", ...)` returns 405 on GET natively if
  `app.Get` is not registered for the same path.

## Form Field Name Contract

DOM order below is the canonical order. Per-stack handlers **must** bind to
these exact `name` attributes.

| Field name | HTML element | Required | Constraints |
|---|---|---|---|
| `code` | `<input type="text">` | yes | trimmed; non-empty; max 32 chars; pattern `^[A-Z0-9][A-Z0-9-]*$`; unique across all projects (case-sensitive UNIQUE constraint) |
| `name` | `<input type="text">` | yes | trimmed; non-empty; max 200 chars |
| `description` | `<textarea>` | no | trimmed; max 10 000 chars (application-level cap; DDL is unbounded `TEXT`) |
| `start_date` | `<input type="date">` | yes | must parse as ISO date (YYYY-MM-DD) |
| `target_completion_date` | `<input type="date">` | no | if provided, must be ≥ `start_date` |
| `trade_scope_ids` | `<select multiple>` or repeated `<input type="checkbox">` | yes | at least one; each UUID must exist in `domain.trade_type` with `active=true` |
| `inspector_ids` | `<select multiple>` or repeated `<input type="checkbox">` | no | each UUID must exist as a user with the `INSPECTOR` conceptual role and active |

### Form-shape latitude for multi-selects

Snapshot tests assert that the canonical `name` attribute appears on an element
of the appropriate type and that all reference-data rows render as options or
checkboxes. Tests do **not** byte-compare the multi-select element type
(`<select multiple>` vs `<input type="checkbox">`) — that is per-stack
rendering latitude.

## Hidden-Input List

| Stack | Hidden input | Purpose |
|---|---|---|
| .NET | `__RequestVerificationToken` | ASP.NET Core antiforgery (required) |
| Django | `csrfmiddlewaretoken` | Django CSRF middleware (required) |
| Go | _(none)_ | ADR-012 exempts Go from CSRF middleware in MVP; the form contains no CSRF hidden input |

The Go exemption is a documented design decision, not an oversight. The negative
assertion is tested: a per-stack test asserts that the Go form does not contain
a CSRF hidden input.

## Return-Target Convention

On successful `POST /projects/`:

- **HTMX request** (`HX-Request: true` header present): respond HTTP 200 with
  `HX-Redirect: /projects/<new-project-id>` header and an empty body (or a
  one-line comment). HTMX triggers `window.location = '/projects/<id>'`.
- **Non-HTMX request** (no `HX-Request` header — e.g. JS disabled): respond
  HTTP 303 See Other with `Location: /projects/<new-project-id>`. 303 follows
  the POST → redirect → GET pattern and preserves the GET method on follow-up.

The redirect target is always server-decided (`/projects/<server-generated
UUID>`). There is no `return_url` parameter on this form — this is a create
flow, not a login flow. The user cannot influence the redirect target.

## 422 Response Body Shape

The response is the **re-rendered form partial** containing:

1. An `InlineAlert` block at the top (`role="alert"`, `severity="danger"`,
   title `"Couldn't create the project"`, message `"<n> error(s) must be
   resolved before this project can be created."`).
2. Each invalid field: `aria-invalid="true"` + `aria-describedby="<field-id>-error"`.
3. Each invalid field followed by `<p id="<field-id>-error" class="form-error"
   role="alert">{{message}}</p>`.
4. Failed input values are echoed back (user does not re-type `name`,
   `description`, etc.). Exception: unknown/inactive UUIDs in `trade_scope_ids`
   and `inspector_ids` are dropped from the re-selection set (see AC4 in the
   story file).
5. The response body contains **no** `hx-swap-oob` attributes (UX Pattern 3 —
   no OOB swaps on failed actions).

HTTP status is 422 Unprocessable Entity — not 400, not 200.

## 403 Response Body Shape

When the requester lacks `project.create`, both `GET /projects/new` and
`POST /projects/` return HTTP 403. The 403 body is byte-equal to the canonical
403 body from Story 1.11 (per AC5 in the story file). No state leakage — the
403 body must not enumerate field names, UUIDs, or any signal that this URL is
special.

## Audit Emission Contract (`ProjectCreated`)

Every successful `POST /projects/` writes one `domain.audit_entry` row:

| Column | Value |
|---|---|
| `action` | `"ProjectCreated"` |
| `actor_id` | requesting user's canonical UUID |
| `entity_type` | `"Project"` |
| `entity_id` | new project's UUID |
| `project_id` | new project's UUID (denormalization) |
| `before_state` | `NULL` |
| `after_state` | see `after_state` JSON schema below |
| `metadata` | `NULL` |

### `after_state` JSON schema for `ProjectCreated`

Keys are alphabetical; `null` fields are present (not omitted); UUID lists are
sorted lexically (so two stacks creating identical projects produce byte-identical
`after_state`). `compliance_score: 100` is the DDL default.

```json
{
  "code": "<code>",
  "compliance_score": 100,
  "description": null,
  "inspector_ids": ["<uuid>", "..."],
  "name": "<name>",
  "start_date": "<YYYY-MM-DD>",
  "status": "Active",
  "target_completion_date": null,
  "trade_scope_ids": ["<uuid>", "..."]
}
```

## Why `HX-Redirect` (not partial swap)

A brand-new entity has no existing in-page partial container — the destination
is a separate screen (Project Detail, Story 2.11). Partial swap into the current
page would produce a disorienting in-place replacement of the form with a detail
view; full navigation is the honest signal. The non-HTMX fallback (303 +
`Location`) ensures the no-JS browser still works correctly.

## Why a plain `<button type="submit">` (not ActionButton)

The ActionButton trichotomy (absent / disabled / present) applies to *action
affordances* — buttons that fire `hx-post` for a single state-changing action
on an existing entity. The form-submit button is a different concept: it commits
the form values the user just typed. By the time the user is viewing the create
form they have already passed the `project.create` permission check (403 is
returned before the form renders). The submit button is always present and is a
plain `<button type="submit">Create Project</button>`.

## Transaction Shape (canonical reference for Epic 2–6)

Every state-changing handler from Story 2.8 onward follows this five-step shape:

1. **Validate** input at the request boundary (field types, lengths, allowlists,
   CSRF).
2. **Open transaction** against the `domain` schema.
3. **Load reference rows** inside the transaction (trade types, inspectors) for
   validation. For soft-deleted rows (deactivated trades / inspectors), prefer
   the FK as the final guarantee rather than re-checking active flags on every
   write — the race window is microseconds and a recoverable state.
4. **Call entity method** (`Project.create`, etc.) to produce new entity +
   relation collections.
5. **Persist** domain writes in FK order: aggregate row → join tables →
   `AuditEntry` (via `append_audit_entry`).
6. **Commit**. On any error inside the transaction the entire transaction rolls
   back — no orphan rows, no orphan audit entry.
7. **Respond**: HTMX → 200 + `HX-Redirect`; non-HTMX → 303 + `Location`.

## Change Procedure

Adding a field or changing a field name follows the same procedure as
`docs/reference/audit-actions.md`:

1. ADR amendment recorded in the epic file and in this document.
2. Update all three per-stack handler files and their snapshot/form-binding
   tests.
3. Run `make parity` and confirm the route inventory diff is clean.
4. Run `make test-all` and confirm all three stacks pass.

Per-stack handler and template files that implement this contract must include
a top-of-file comment referencing this document URL (`docs/reference/project-create-form-contract.md`).
