# Epic 1: Walking Skeleton — Auth, Design System & Parity Foundation

Lay the cross-stack foundation so every later epic implements only its own domain delta. After Epic 1, `make up && make run-{net,django,go}` produces three stacks rendering byte-identical chrome at a role-aware empty home page, with the light/dark theme toggle, FlashRegion, and the affordance-trichotomy primitive in place. `make parity` runs clean.

## Story 1.1: Confirm three native scaffolds, root Makefile, and Docker Compose harness

As a developer joining FieldMark for the first time,
I want a single documented set of commands that bring all three stacks and the database up locally,
So that I can run the application on every stack from a clean clone in minutes.

**Acceptance Criteria:**

**Given** a clean clone of the repository
**When** I run `make up` from the repo root
**Then** Postgres 17 starts via `docker compose up -d` and is reachable on `localhost:5432` with `fieldmark/fieldmark/fieldmark`
**And** the init scripts under `docker/postgres/init/` run automatically on first volume creation.

**Given** Postgres is up
**When** I run `make run-net`, `make run-django`, and `make run-go` (each in its own shell)
**Then** the three stacks bind to their native ports (.NET :5000, Django :8000, Fiber :3000)
**And** each stack reads `FIELDMARK_DATABASE_URL` (defaulting to the local Postgres URL) and connects without error.

**Given** the repo at HEAD
**When** I inspect the top-level `Makefile`
**Then** it exposes targets `up`, `down`, `reset`, `run-net`, `run-django`, `run-go`, `test-net`, `test-django`, `test-go`, `e2e`, `parity`, `css` per Architecture D20
**And** each target succeeds (or no-ops cleanly) on a fresh clone.

**Given** the repo at HEAD
**When** I inspect the three stack directories `FieldMark/`, `fieldmark_py/`, `fieldmark-go/`
**Then** each matches the Architecture §Initialization Commands layout (`.NET`: Web/Domain/Data class libs + xUnit projects; Django: `projects`, `inspections`, `violations`, `compliance`, `audit`, `reference`, `grid` apps with `uv` deps pinned; Go: `cmd/web` + `internal/{app,data,domain,web}`)
**And** each stack's README documents how to run it.

---

## Story 1.2: Verify Postgres init scripts produce the canonical `domain.*` schema on a fresh volume

As a developer working across three stacks,
I want a single command that destroys and re-creates the database in a known canonical state,
So that any drift between framework mapping code and the infrastructure-owned schema surfaces immediately.

**Acceptance Criteria:**

**Given** a running database with arbitrary local state
**When** I run `make reset` (`docker compose down -v && docker compose up -d`)
**Then** the volume is destroyed and recreated
**And** `001_schemas.sql`, `010_domain_tables.sql`, and `020_domain_seed.sql` execute in order with no errors visible in `docker logs`.

**Given** the database has been initialized
**When** I connect with `psql` and run `\dn`
**Then** the schemas `domain`, `dotnet_auth`, `django_auth`, `fiber_auth`, `infra` are all present.

**Given** the database has been initialized
**When** I run `SELECT table_name FROM information_schema.tables WHERE table_schema='domain' ORDER BY table_name`
**Then** exactly 12 tables are returned: `audit_entry`, `compliance_rule`, `corrective_action`, `finding`, `inspection`, `job_site`, `project`, `project_inspector`, `project_trade_scope`, `trade_type`, `violation`, `violation_category`.

**Given** the database has been initialized
**When** I inspect `domain.trade_type`, `domain.violation_category`, and `domain.compliance_rule`
**Then** the reference rows from `020_domain_seed.sql` are present and identical to the file's `INSERT` statements (verified by row count + a `SELECT` sample).

**Given** the canonical DDL is owned by infrastructure (ADR-014)
**When** I grep each stack for tooling that could mutate the `domain` schema (`dotnet ef migrations add` against a DbContext whose `HasDefaultSchema` is `"domain"`, Django `makemigrations` against a `domain.*` model with `Meta.managed = True`, Go migration tools targeting `domain.*`)
**Then** zero matches are found
**And** each stack's README explicitly states that `domain.*` is infrastructure-owned and that framework migrations only apply to its `*_auth` schema.

---

## Story 1.3: Establish `tools/parity/` and `make parity` with per-stack `--dump-routes`

As an agent or developer modifying any of the three stacks,
I want a single local command that detects cross-stack drift on routes and database indexes,
So that I catch divergence before it reaches code review — without depending on CI.

**Acceptance Criteria:**

**Given** the repo at HEAD
**When** I inspect `tools/parity/`
**Then** the directory contains executable scripts `dump-pg-indexes.sh`, `dump-routes-net.sh`, `dump-routes-django.sh`, `dump-routes-fiber.sh`, `diff-routes.sh`, `diff-pg-indexes.sh` (per Architecture D19).

**Given** each stack
**When** I invoke its route-dump subcommand
**Then** .NET responds to `dotnet run --project FieldMark/FieldMark.Web -- --dump-routes`, Django responds to `manage.py show_urls` (or equivalent custom command), and Go responds to `go run ./cmd/web -dump-routes`
**And** each command writes a normalized line-per-route list (METHOD + path) to stdout, sorted, with language casing normalized to lowercase.

**Given** the database has been initialized and all three stacks are buildable
**When** I run `make parity` from the repo root
**Then** the script invokes `diff-routes.sh` (comparing all three route dumps) and `diff-pg-indexes.sh` (snapshotting `pg_indexes WHERE schemaname='domain'` against the canonical file)
**And** both diffs exit `0` (clean).

**Given** I intentionally add a route to one stack and not the others
**When** I run `make parity`
**Then** the command exits non-zero and prints the diff identifying the divergent route.

**Given** `tools/git-hooks/pre-commit.sample` is committed
**When** I read it
**Then** it shows how to opt in to running `make parity` on commits touching any of `FieldMark/`, `fieldmark_py/`, `fieldmark-go/`, or `docker/postgres/init/`.

---

## Story 1.4: Bootstrap design system foundation in `fieldmark_shared/`

As a developer styling any FieldMark screen on any stack,
I want one compiled CSS bundle with the Basecoat component vocabulary, semantic tokens, status-badge vocabulary, typography, and vendored JS,
So that I can render byte-identical markup across the three stacks without authoring per-stack CSS.

**Acceptance Criteria:**

**Given** the repo at HEAD
**When** I inspect `fieldmark_shared/package.json`
**Then** `tailwindcss@4.x` is pinned to an exact patch and `basecoat-css` is pinned to an exact pre-1.0 patch (e.g., `0.3.11`) — no `^` or `~` ranges (UX-DR1)
**And** the version pins are documented in `_bmad-output/planning-artifacts/architecture.md` alongside HTMX and AG Grid.

**Given** `fieldmark_shared/src/fieldmark.css`
**When** I read it
**Then** it imports Basecoat's CSS, the AG Grid Quartz theme, and declares the five semantic color tokens `--color-success`, `--color-warning`, `--color-danger`, `--color-info`, `--color-neutral` (UX-DR2) with both light and dark variants
**And** each token meets ≥ 4.5:1 contrast against `neutral-50/100` and `neutral-900/950`, with a one-line comment recording the contrast ratio at design time.

**Given** the same file
**When** I read it
**Then** the status-badge color vocabulary (UX-DR3) for Project, Inspection, Violation (with severity overlay), CorrectiveAction, and Severity is encoded as deterministic class-to-token mappings
**And** the compliance-score threshold mapping (UX-DR4) is encoded as a single CSS rule keyed on `data-score-band` (`healthy`, `watch`, `concern`, `critical`).

**Given** `fieldmark_shared/src/`
**When** I read its CSS
**Then** Inter and JetBrains Mono are referenced as `@font-face` declarations pointing to self-hosted woff2 files under `fieldmark_shared/vendor/fonts/` (UX-DR6)
**And** body default is `text-sm` (14px), `font-feature-settings: "tnum"` is applied via a `.tnum` utility to compliance score, timestamps, counts, and any DOM element with numeric updating values.

**Given** the spacing scale (UX-DR8)
**When** I read `fieldmark_shared/src/_layout.css`
**Then** it uses only Tailwind defaults — no custom breakpoints — and documents `max-w-screen-2xl` container + `px-6 → px-4` gutter collapse with a single comment per rule naming the collapse point it implements.

**Given** the vendored JS strategy (Architecture D15)
**When** I inspect `fieldmark_shared/vendor/`
**Then** `htmx/htmx.min.js and ag-grid/35.2.1/ag-grid-community.min.js are committed; each stack's vendor/ static dir has directory symlinks pointing here
**And** each stack's static directory symlinks `dist/fieldmark.css` and the vendor directory.

**Given** the design system is built
**When** I run `cd fieldmark_shared && npm run build` (alias `make css`)
**Then** `fieldmark_shared/dist/fieldmark.css` is produced
**And** the compiled file is committed (no build step required after clone).

---

## Story 1.5: Implement cross-stack base layout with skip-link, landmarks, and FlashRegion

As a screen-reader user landing on any FieldMark page on any stack,
I want a consistent landmark structure, a working skip-link, and a polite live region for system announcements,
So that I can navigate the application predictably regardless of which stack served the page.

**Acceptance Criteria:**

**Given** each stack
**When** I open the rendered base layout (Razor `_Layout.cshtml`, Django `base.html`, Go `layouts/base.tmpl`)
**Then** the document body's first focusable element is a "Skip to main content" link that targets `#main-content` (UX-DR33)
**And** the link is visually hidden until focused.

**Given** the same base layout
**When** I inspect the document structure
**Then** there is exactly one `<header>`, one `<nav aria-label="Main">`, one `<main id="main-content">`, an optional `<aside>` slot for EntityRail, and an optional `<footer>`
**And** there are no nested landmarks of the same role.

**Given** every page rendered by any stack
**When** I count `<h1>` elements
**Then** exactly one is present (the page title), and heading levels never skip (no `<h3>` without a prior `<h2>` in the same section) (UX-DR33).

**Given** the base layout
**When** I inspect it
**Then** `#flash-region` is present as a `<div id="flash-region" role="status" aria-live="polite" aria-atomic="false">` in page chrome (UX-DR14, UX-DR32)
**And** it is empty by default and renders any messages from a per-stack `flash_messages()` template helper.

**Given** focus styling (UX-DR35)
**When** I tab through any rendered page
**Then** the `:focus-visible` ring is 2px wide at 2px offset, in body text color
**And** touch targets render at ≥ 44×44px under `(pointer: coarse)` media query.

**Given** the three stacks
**When** I capture the rendered HTML of `/` on each stack
**Then** the chrome (header skeleton, nav skeleton, skip-link, FlashRegion, main slot, footer skeleton) is byte-identical modulo any per-stack server-rendered values (none expected at this story).

---

## Story 1.6: Implement ThemeToggle with cookie persistence per stack

As any user on any stack,
I want a single header-strip control that cycles System → Light → Dark with my preference remembered across sessions,
So that the application matches my environment without flashing the wrong theme on first paint.

**Acceptance Criteria:**

**Given** I land on any page with no prior preference
**When** the page renders
**Then** the server emits `<html data-theme="system">` and a 5-line inline `<script>` resolves `prefers-color-scheme` and sets `data-theme="light"` or `data-theme="dark"` before first paint (UX-DR5)
**And** that inline script is the only inline JavaScript in the application; its presence is documented in the architecture doc.

**Given** the ThemeToggle component renders in the header strip beside the user avatar slot
**When** I inspect it (UX-DR15)
**Then** it is a 36×36 icon button with `aria-label="Theme: <current>; activate to cycle"`
**And** the displayed Lucide icon (Sun / Moon / Monitor) reflects the *currently resolved* theme.

**Given** I click the ThemeToggle
**When** the click fires
**Then** an HTMX `hx-post` is sent to `/preferences/theme` with the cycled value (`system` → `light` → `dark` → `system`)
**And** the server sets `Set-Cookie: fm_theme=<value>; Path=/; SameSite=Lax; Max-Age=31536000` and returns HTTP `204` with `HX-Trigger: theme-changed`
**And** a small client-side listener (≤ 20 LOC, vendored as `theme-toggle.js`) updates `data-theme` on `<html>` immediately.

**Given** I refresh the page after setting a preference
**When** the page renders
**Then** the server reads the `fm_theme` cookie and emits the correct `data-theme` attribute before first paint
**And** no flash of wrong theme is visible.

**Given** the three stacks
**When** I capture the rendered Theme Toggle markup on each
**Then** the HTML is byte-identical for identical inputs
**And** the `/preferences/theme` endpoint exists at the same path on all three (verified by `make parity`).

**Given** the user activates the toggle by keyboard (Space or Enter)
**When** I observe in a screen reader
**Then** the cycle works and the `aria-label` value updates to describe the new current + next state.

---

## Story 1.7: Wire ASP.NET Core Identity to `dotnet_auth` schema with conceptual roles

As an administrator using the .NET stack,
I want framework-native authentication backed by the `dotnet_auth` schema with the canonical password policy,
So that user identity is owned by .NET and never leaks into the `domain.*` schema.

**Acceptance Criteria:**

**Given** the .NET solution
**When** I inspect `FieldMark.Data` and `FieldMark.Web`
**Then** an `AuthDbContext` is configured with `modelBuilder.HasDefaultSchema("dotnet_auth")` and `UseSnakeCaseNamingConvention()` (Architecture D6)
**And** all seven Identity tables (`users`, `roles`, `user_roles`, `role_claims`, `user_claims`, `user_logins`, `user_tokens`) are mapped into `dotnet_auth`.

**Given** the password policy
**When** I read the Identity options registration
**Then** `RequireDigit = true`, `RequireLowercase = true`, `RequireUppercase = true`, `RequireNonAlphanumeric = false`, `RequiredLength = 10` are set.

**Given** Identity migrations
**When** I list `FieldMark.Data/Migrations/Auth/`
**Then** initial migration files exist that create the seven `dotnet_auth.*` tables only
**And** no migration touches `domain.*` (verified by grep).

**Given** Identity is wired
**When** the application starts for the first time after `make reset`
**Then** the five canonical role records are seeded into `dotnet_auth.roles` with names `ADMIN`, `COMPLIANCE_OFFICER`, `INSPECTOR`, `SITE_SUPERVISOR`, `EXECUTIVE`
**And** seeding is idempotent (running the seeder twice produces the same state).

**Given** parity tooling
**When** I run `make parity`
**Then** the route inventory diff remains clean (no .NET-only auth routes break parity — Django and Go have equivalent endpoints).

---

## Story 1.8: Wire Django built-in `auth` to `django_auth` schema with conceptual-role Groups

As an administrator using the Django stack,
I want framework-native authentication backed by the `django_auth` schema with role assignment via Groups,
So that Django's identity layer mirrors the .NET stack's isolation and never touches `domain.*`.

**Acceptance Criteria:**

**Given** `fieldmark_py/fieldmark/settings.py`
**When** I read database routing
**Then** a `DatabaseRouter` is configured (or `db_table` overrides applied to Django auth models) so that `auth_user`, `auth_group`, `auth_permission`, `auth_user_groups`, `auth_user_user_permissions`, `auth_group_permissions`, `django_session`, `django_admin_log` all resolve into the `django_auth` schema (Architecture D7).

**Given** Django migrations
**When** I run `uv run python manage.py migrate`
**Then** auth tables are created in `django_auth` and no auth migration touches `domain.*` (verified by inspecting `django_migrations` table).

**Given** the auth schema is migrated
**When** the application starts (or a one-shot management command runs)
**Then** five Django Groups are present: `ADMIN`, `COMPLIANCE_OFFICER`, `INSPECTOR`, `SITE_SUPERVISOR`, `EXECUTIVE`
**And** the seeding management command is idempotent.

**Given** `make parity`
**When** I run it after Django auth is wired
**Then** route inventory diff stays clean and `pg_indexes` for `domain.*` shows zero changes from the canonical inventory.

---

## Story 1.9: Implement Go/Fiber stub authentication middleware

As a developer running the Go stack at MVP,
I want a stub authentication mechanism that injects a configurable user identity into the request context,
So that the Go stack can render role-aware pages and exercise the cross-stack parity contract while real auth remains deferred per ADR-012.

**Acceptance Criteria:**

**Given** `fieldmark-go/internal/web/auth/`
**When** I inspect it
**Then** a `StubAuthMiddleware` exists that reads a user identifier from (in order) the `X-FieldMark-Actor` header, the `FIELDMARK_STUB_ACTOR` env var, or falls back to an "anonymous" sentinel.

**Given** the middleware resolves a user id
**When** the request context is hydrated
**Then** the user's UUID, username, and resolved conceptual role are bound to `c.Locals("user", ...)` and accessible from any handler
**And** the middleware looks up the user from a small `fiber_auth.users` + `fiber_auth.user_roles` pair of tables it owns (seeded in Story 1.10).

**Given** a request arrives with no identity
**When** the handler is `[`Authorize required`]`
**Then** the middleware returns HTTP `302` to `/login` (which renders a user-switcher stub list, not a real form).

**Given** ADR-012 explicitly defers real Go auth
**When** I read `fieldmark-go/CLAUDE.md`
**Then** the stub strategy is documented along with the explicit deferral and what landing real auth would look like (epic-sized work, not MVP).

**Given** `make parity`
**When** I run it
**Then** the Go stack's route inventory matches .NET and Django modulo language casing — including the `/login` and `/logout` paths.

---

## Story 1.10: Author shared UUID dev-user manifest and per-stack idempotent seed runners

As a developer running cross-stack scenarios,
I want every stack's dev users to share identical UUIDs,
So that audit comparison and cross-stack E2E parity tests can assert on actor identity without translation tables.

**Acceptance Criteria:**

**Given** `docker/postgres/init/seed-uuids/dev-users.json`
**When** I read it
**Then** it contains exactly six users: Marisol (`COMPLIANCE_OFFICER`), Pat (`SITE_SUPERVISOR`), Aisha (`ADMIN`), an inspector "Ravi" (`INSPECTOR`), Kenji (`EXECUTIVE`), and a no-role test user
**And** each entry has a canonical UUID (UUIDv7 preferred), a username, a display name, an initial password, and a role.

**Given** the .NET seeder `FieldMark.Web/SeedData/DevUsers.cs`
**When** `make run-net` starts the application with an empty database
**Then** the seeder reads the JSON manifest and writes the six users to `dotnet_auth.users` with the manifest's UUIDs as primary keys, hashed via ASP.NET Core Identity's `IPasswordHasher`
**And** running the seeder twice produces no duplicates and no errors (idempotent).

**Given** the Django seeder `fieldmark_py/<app>/management/commands/seed_dev_users.py`
**When** I run `uv run python manage.py seed_dev_users`
**Then** the six users are written to `django_auth.auth_user` with the manifest UUIDs (stored in a `uuid` column or as `username=<uuid>` if Django auth's PK contract is incompatible — chosen approach documented in the command's docstring)
**And** users are assigned to their conceptual-role Group.

**Given** the Go seeder `fieldmark-go/cmd/seed/main.go`
**When** I run `go run ./cmd/seed`
**Then** the six users are written to `fiber_auth.users` + `fiber_auth.user_roles` with the manifest UUIDs.

**Given** all three seeders have run
**When** I query each stack's auth tables for `id`/`uuid` of `marisol`
**Then** the returned UUID is identical across all three stacks (verified by SQL spot-check).

**Given** `020_domain_seed.sql` already seeds reference data
**When** I inspect the per-stack seed runners
**Then** none of them write into `domain.*` (reference data ownership stays with infrastructure SQL).

---

## Story 1.11: Login, logout, and unauthenticated-redirect across all three stacks

As any FieldMark user,
I want to log in with my username and password on .NET and Django and to pick my actor on Go,
So that the application identifies me on every request and rejects access to business routes until I authenticate.

**Acceptance Criteria:**

**Given** I am unauthenticated on any stack
**When** I request any business route (e.g., `/`, `/projects`, `/dashboard`)
**Then** I am redirected to `/login` (FR4)
**And** the response is HTTP `302` (or framework-equivalent).

**Given** the .NET login page
**When** it renders
**Then** the form is built from Basecoat input components with `<label>`-associated inputs
**And** on validation failure each invalid field renders `aria-invalid="true"` + `aria-describedby` linking to its error message, and the form partial is re-rendered with HTTP `422` containing a top InlineAlert with `role="alert"` and a link to the first invalid field (UX-DR34, FR61).

**Given** the Django login page
**When** it renders
**Then** the same form contract holds (Basecoat markup, label association, 422 + `aria-invalid`/`aria-describedby` on failure) — byte-identical markup verified by snapshot.

**Given** the Go login page
**When** it renders
**Then** it presents a list of seeded users from `fiber_auth.users` styled as Basecoat buttons; clicking a user sets the `X-FieldMark-Actor` cookie and redirects to `/`
**And** the page is clearly labeled as a development stub per ADR-012.

**Given** I am authenticated on any stack
**When** I click Log Out
**Then** the session is terminated (FR3) and I am redirected to `/login`
**And** subsequent requests to business routes redirect to `/login` again.

**Given** I am authenticated
**When** any request is handled
**Then** the handler can resolve my UUID and conceptual role(s) via the per-stack equivalent of `currentUser` (FR2)
**And** an unauthorized direct request (e.g., POSTing to `/projects/:id/close` without role) returns HTTP `403` without leaking the entity state (FR7, FR56).

**Given** `make parity`
**When** I run it with all auth wired
**Then** routes `/login`, `/logout`, `/preferences/theme` exist on all three stacks and the diff is clean.

---

## Story 1.12: Implement `authz.Can` primitive and ActionButton trichotomy helper per stack

As a developer rendering an action affordance on any FieldMark screen,
I want a single template helper that decides `absent | disabled-with-tooltip | present` per the affordance trichotomy,
So that future epics can introduce action buttons without re-deciding the rendering rule per screen.

**Acceptance Criteria:**

**Given** each stack
**When** I inspect its authorization module
**Then** it exposes a function with the signature `Can(user, action: string, entity?) -> bool` (.NET: `DomainPolicies.Can(...)`, Django: `fieldmark.authz.can(...)`, Go: `authz.Can(...)`) (FR5)
**And** the function consults the user's conceptual role(s) and any entity-scope rules (e.g., assignment, ownership) — initially trivial since no entities exist yet.

**Given** the ActionButton template helper (UX-DR10, UX-DR21)
**When** I inspect each stack's wrapper (`Pages/Shared/_ActionButton.cshtml`, `templates/components/_action_button.html`, `internal/web/templates/components/action_button.tmpl`)
**Then** it accepts `permission: bool`, `state_allows: bool`, `label: string`, `hx_post: string`, `hx_target: string`, and optional `disabled_reason: string`
**And** it implements the trichotomy:
- `permission=false` → renders nothing
- `permission=true && state_allows=false` → renders a Basecoat `<button disabled aria-disabled="true">` with a tooltip carrying `disabled_reason` and `aria-describedby` linking to the tooltip
- `permission=true && state_allows=true` → renders a Basecoat `<button hx-post=... hx-target=... hx-swap=... hx-disabled-elt="this">` (UX-DR27, FR64).

**Given** identical inputs on all three stacks
**When** I render the same ActionButton invocation
**Then** the produced HTML is byte-identical (verified by a unit test per stack snapshotting against a canonical example in `fieldmark_shared/components/action_button.example.html`).

**Given** the ActionButton renders a disabled button
**When** I navigate by keyboard
**Then** the disabled button retains its place in the tab order, the tooltip is keyboard-reachable, and the `aria-describedby` association is announced by screen readers (UX-DR21).

**Given** Epic 1 has no live use sites for ActionButton
**When** I grep each stack's templates for usages
**Then** zero rendered call sites are found (the primitive exists for Epic 2 onward) — but the unit-test snapshots prove the helper renders correctly.

---

## Story 1.13: Render empty role-aware Home page identically across all three stacks

As any authenticated user on any stack,
I want to land on a clean Home page that reflects who I am and offers the theme toggle,
So that I can confirm I am logged in on the right stack with the right identity before the product features land.

**Acceptance Criteria:**

**Given** I am authenticated on any stack
**When** I navigate to `/`
**Then** I see a page with:
- a `<header>` containing the FieldMark wordmark (left), the ThemeToggle (right of avatar), and an avatar showing my initials (Story 1.6)
- a single `<h1>FieldMark</h1>` (UX-DR33)
- a role badge using the StatusBadge color vocabulary showing my resolved conceptual role
- an empty content slot with a placeholder string ("Your projects will appear here.")
- the FlashRegion (`#flash-region`) in chrome (Story 1.5).

**Given** I render `/` on each of the three stacks while logged in as the same user (same UUID, courtesy of Story 1.10)
**When** I capture the rendered HTML
**Then** the chrome and the role badge are byte-identical across stacks (Basecoat-classed markup; no per-stack class names).

**Given** I am unauthenticated
**When** I navigate to `/`
**Then** I am redirected to `/login` (FR4 — already covered in Story 1.11; reasserted here for the empty Home).

**Given** the page renders
**When** I run an axe-core scan
**Then** zero WCAG 2.1 AA violations are reported (UX-DR39 — applies to every rendered page; locked in here as the first instance).

**Given** I tab through the page
**When** I observe focus order
**Then** Skip-Link → ThemeToggle → Avatar Menu → Logout → page body, in that order, with the visible focus ring at every step (UX-DR35).

**Given** `make parity`
**When** I run it after Story 1.13 lands
**Then** route inventory and `pg_indexes` for `domain.*` are clean across all three stacks.

## Story 1.14: Harden design-system foundation and build tooling against known edge cases

As a developer about to begin Epic 2 feature work,
I want the Epic 1 foundation hardened against the edge cases surfaced during code review of Stories 1.4–1.13,
So that user-facing features ride on a base that degrades gracefully and doesn't regress under upgrades or hostile inputs.

This story consolidates the items captured in `_bmad-output/implementation-artifacts/deferred-work.md` (2026-05-17 entries). It is the final story of Epic 1; `make parity` (asserted in Story 1.13) must remain clean.

**Acceptance Criteria:**

**AC1 — Accessibility & motion preferences**

**Given** a user with `prefers-reduced-motion: reduce`
**When** the sidebar opens/closes, a toast appears, or a tooltip shows
**Then** transitions are instant (no animation)
**And** an axe-core scan on the affected pages still reports zero WCAG 2.1 AA violations.

**Given** a user in `forced-colors: active` (Windows High Contrast) mode
**When** any StatusBadge or `data-score-band` element renders
**Then** meaning is conveyed by text or icon, not color alone.

**Given** the brand fonts return 404 or are blocked by the network
**When** the page renders with system-font fallback
**Then** a checked-in Playwright visual regression snapshot for the login page passes within the configured tolerance
**And** Cumulative Layout Shift remains ≤ 0.1.

**AC2 — Component robustness**

**Given** a `badge-*` or `data-score-band` value not in the documented vocabulary
**When** the component renders
**Then** it falls back to a documented "unknown" style and emits a single server-side warning log (not a console error).

**Given** the `[data-sidebar-initialized]` attribute is never set because the sidebar JS fails to load
**When** the page renders
**Then** the sidebar renders in a documented degraded state (visible, non-collapsible) — not hidden, not jumping.

**Given** AG Grid has zero rows or is in a loading state
**When** the grid renders
**Then** empty and loading states use design-system styling that is visually distinct from the Basecoat table empty state and is documented in `fieldmark_shared/`.

**Given** more than 5 toasts are queued
**When** the toaster renders
**Then** only 5 toasts are visible at once, the toast region scrolls on overflow, and the region's height never grows unbounded.

**Given** a `data-tooltip` containing HTML entities or text exceeding the `container-xs` width
**When** the tooltip displays
**Then** entities render as their characters (no raw `&amp;`) and overflow text wraps or truncates with ellipsis — never silently clips.

**AC3 — Build tooling hardening**

**Given** a stack directory is renamed or any `@source` glob no longer matches files
**When** the CSS build runs
**Then** the build fails loudly with an actionable error pointing at the bad glob (not silent zero-output).

**Given** `optimize-css.mjs` is run against a missing input file, against a read-only target, or with LightningCSS recovered errors present
**When** the script executes
**Then** it exits non-zero with a clear error message
**And** writes to a `.tmp` file before atomic rename (no in-place mutation without backup)
**And** propagates any LightningCSS warnings to stderr.

**Given** the project is checked out fresh on a machine using npm or yarn instead of pnpm
**When** a developer follows the build docs
**Then** the build either works or fails immediately with a clear "pnpm + ESM required" message (no silent breakage mid-pipeline).

**AC4 — Documentation & upgrade-resilience**

**Given** a developer reads `docs/getting-started.md` and root `CLAUDE.md`
**When** they look for the CSS pipeline
**Then** the `optimize-css.mjs` step is documented, including its inputs, outputs, failure modes, and how to bypass it locally during debugging.

**Given** Basecoat publishes a minor version with renamed classes or reintroduced unmergeable duplicates
**When** `pnpm update` runs and the build executes
**Then** a pinned class-name smoke test (or version-range assertion) in the build fails fast with a pointer to a documented upgrade checklist.

**AC5 — Epic 1 exit**

**Given** Story 1.14 lands
**When** `make parity` runs
**Then** route inventory and `pg_indexes` for `domain.*` remain clean across all three stacks
**And** the deferred-work entries dated 2026-05-17 in `_bmad-output/implementation-artifacts/deferred-work.md` are either resolved or explicitly re-deferred with a written rationale
**And** Epic 1 is closed.
