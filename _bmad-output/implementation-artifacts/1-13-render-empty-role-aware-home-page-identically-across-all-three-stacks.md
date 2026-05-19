# Story 1.13: Render empty role-aware Home page identically across all three stacks

Status: ready-for-dev

## Story

As any authenticated user on any of the three FieldMark stacks,
I want to land on a clean Home page that reflects who I am and offers the theme toggle,
So that I can confirm I am logged in on the right stack with the right identity before the product features land — and so Epic 1 closes with a byte-identical chrome surface across .NET, Django, and Go (FR58).

## Acceptance Criteria

1. **Authenticated `/` renders a single canonical Home page on every stack.** The page is composed of (in DOM order, inside `<main id="main-content">`):
   - The existing `<div id="flash-region" role="status" aria-live="polite" aria-atomic="false">` (Story 1.5 — unchanged; do not re-emit, do not nest).
   - A single `<h1>FieldMark</h1>` (UX-DR33 — exactly one `<h1>` per page; never `<h1>Home</h1>` or `<h1>Welcome</h1>`).
   - The role badge described in AC #4 (no other element between the `<h1>` and the badge).
   - A single `<p class="text-muted">Your projects will appear here.</p>` placeholder. The string is verbatim — case, period, and quoting are part of the cross-stack contract verified by AC #7.
   - No `<form>`, no `<button>`, no AG Grid panel, no HTMX-loaded fragment, no inline `<script>`. This is intentionally empty; Epic 2 (Story 2.10) fills it.

2. **The page chrome (`<header>` + `<nav aria-label="Main">`) renders the four canonical chrome controls in this exact left-to-right order on every stack:**
   1. The FieldMark wordmark — `<a href="/" class="fm-wordmark" aria-label="FieldMark home">FieldMark</a>` — first focusable element inside `<nav aria-label="Main">`.
   2. A nav spacer (`<div class="ml-auto flex items-center gap-3">`) pushing the right-side cluster.
   3. The existing `<button class="theme-toggle" data-theme-toggle …>` from Story 1.6 (unmoved, unmodified except for being inside the new cluster).
   4. The Avatar + LogoutMenu composite from AC #3.
   The skip-link from Story 1.5 (`<a href="#main-content" class="skip-link">`) remains the first focusable element of `<body>` — it is **outside** `<header>` and unchanged.

3. **The Avatar component is rendered as a server-decided `<button class="avatar-menu" type="button" aria-haspopup="true" aria-expanded="false" aria-controls="avatar-menu-dropdown">` containing a `<span class="avatar avatar-initials" aria-hidden="true">{initials}</span>` and a visually-hidden `<span class="sr-only">User menu for {full_name} ({role})</span>` for screen readers.** Behavior at this story:
   - `{initials}` is derived deterministically from `full_name` by taking the first character of the first whitespace-separated token plus the first character of the last whitespace-separated token, uppercased; for a single-token name take the first two characters uppercased; if `full_name` is empty fall back to the first two characters of `username`. Helper functions per stack: `.NET — FieldMark/FieldMark.Web/Helpers/AvatarInitials.cs:Initials(string fullName, string username)`; `Django — fieldmark_py/fieldmark/avatar.py:initials(full_name, username)`; `Go — fieldmark-go/internal/web/viewmodels/avatar.go:Initials(fullName, username string) string`. Each helper has unit tests covering: empty full name → username fallback, single-token name → first two characters, two+ token name → first+last initials, Unicode characters preserved as-is (no transliteration), all uppercase output.
   - The dropdown panel is a sibling `<ul id="avatar-menu-dropdown" class="menu hidden" role="menu">` containing exactly one item: `<li role="none"><a role="menuitem" class="menu-item" href="/logout">Sign out</a></li>` (the `/logout` route was registered by Story 1.11). The dropdown is `hidden` by default; **no client-side JS is added by this story to open it** — it is keyboard-discoverable via Tab (the `<a href="/logout">` is the next focusable element after the avatar `<button>`) and visible-on-`:focus-within` via CSS (see Task 2). Epic 2+ may later wire a click toggle; this story is markup-only.
   - The whole avatar+dropdown sits inside a `<div class="avatar-menu-wrapper relative">` so the CSS `:focus-within` rule has a positioning context.

4. **The role badge uses the StatusBadge color vocabulary and the canonical role→color mapping defined below.** Rendered as `<span class="badge badge-{token}" role="status">{role_label}</span>` where:
   - `{role_label}` is the resolved conceptual role title-cased with a space (`"Admin"`, `"Compliance Officer"`, `"Inspector"`, `"Site Supervisor"`, `"Executive"`). When the user has multiple roles, the **first role alphabetically** (matching the Go stub's tie-break from Story 1.9) is rendered; the multi-role surface is out of scope for Epic 1.
   - `{token}` is from the canonical mapping (locked here for Story 1.13; future role-color decisions amend this table):

     | Role | `{token}` | `--color-*` source | Rationale |
     |---|---|---|---|
     | `ADMIN` | `danger` | `--color-danger` | Highest-privilege role; visually distinct |
     | `COMPLIANCE_OFFICER` | `info` | `--color-info` | Read/audit posture |
     | `INSPECTOR` | `warning` | `--color-warning` | Field action posture |
     | `SITE_SUPERVISOR` | `neutral` | `--color-neutral` | Default actor posture |
     | `EXECUTIVE` | `success` | `--color-success` | Portfolio-level read posture |

   - This mapping lives in **one place per stack** alongside the existing role-name source of truth (Story 1.12 deferred this; Story 1.13 lands it). Locations: `.NET — FieldMark/FieldMark.Domain/ValueObjects/Role.cs` adds a `BadgeToken` property on each `Role` instance; `Django — fieldmark_py/fieldmark/roles.py` adds a `BADGE_TOKENS: dict[Role, str]` constant; `Go — fieldmark-go/internal/domain/role.go` adds a `func (r Role) BadgeToken() string` method. Across stacks the mapping table is byte-identical (verified by a small cross-stack snapshot test — see Task 7.4). The badge element itself uses the Basecoat `.badge` class plus a stack-shared modifier (`.badge-danger`, etc.) defined in `fieldmark_shared/src/_components.css`.
   - The badge text — never color alone — is the accessibility surface (UX-DR for "color paired with text"; UX spec line 848).

5. **Unauthenticated `/` redirects to `/login` (FR4 / Story 1.11 reassertion).** Story 1.11 already wires this on each stack (.NET fallback authorization policy; Django `LoginRequiredMiddleware`; Go `auth.RequireAuth()` on the app's protected route group). Story 1.13's only obligation here is to ensure the Home page handlers do **not** bypass that machinery on any stack — specifically:
   - **.NET:** the new `IndexModel` (or replacement page model) is not decorated with `[AllowAnonymous]`. It inherits the fallback policy.
   - **Django:** the `home(request)` view does not carry `@login_not_required` (Django 5+ decorator) and is not added to `LoginRequiredMiddleware`'s allowlist.
   - **Go:** the `/` route remains registered inside the `app.Group("/", auth.RequireAuth())` block from Story 1.11 — Story 1.13 does not move it to the unauthenticated `app.Get(...)` block.
   - Verified by an integration test per stack that issues a GET to `/` with no auth cookie/session and asserts HTTP 302 with `Location: /login`.

6. **Zero WCAG 2.1 AA violations under axe-core.** Each stack carries an integration test that renders `/` with an authenticated dev user, captures the HTML, and runs `axe-core` (Node CLI invoked from the test, or `@axe-core/playwright` if a Playwright harness is already wired by Story 1.5/1.11) with the **default WCAG 2.1 AA ruleset**. Zero violations is the gate. (UX-DR39 — applies to every rendered page; locked in here as the first instance.) Specific surfaces under audit by this story:
   - Single `<h1>`; heading levels never skip.
   - All interactive controls have accessible names (wordmark `aria-label`, theme-toggle `aria-label` from Story 1.6, avatar `<button>` accessible name via the `sr-only` description span, logout `<a>` visible text).
   - Color is never the only carrier of information (the role badge's text content is the role name).
   - Focus order matches DOM order (verified by AC #7).
   - Touch targets ≥ 44×44px under `(pointer: coarse)` — already established by Story 1.5; the new avatar button must inherit those styles.

7. **Tab order through the rendered page is exactly: Skip-Link → Wordmark → ThemeToggle → Avatar menu button → Logout link → page body.** Verified by an integration test that programmatically tabs through the rendered document (using each stack's idiomatic harness — see Testing Standards) and asserts the focused element sequence by `id`/`aria-label`. The visible focus ring (`:focus-visible` from Story 1.5) appears at every step (UX-DR35); regression-tested by an axe-core rule for focus visibility + a unit assertion that the `outline` computed style is not `none` for each focused element. (Story 1.11 may have placed Logout inside an avatar dropdown — Story 1.13 keeps Logout as a sibling `<a>` of the avatar button so it sits next in DOM order; if Story 1.11 used a different layout, refactor to this contract — see Dev Notes.)

8. **Cross-stack byte-parity of the chrome and the role badge is verified.** When the same dev user (same UUID per Story 1.10) renders `/` on each of the three stacks:
   - The `<header>` block (open tag through close tag) is byte-identical after the chrome-normalization pipeline established in Stories 1.5 and 1.6 (whitespace collapsed, attribute order canonicalized, HTML comments stripped). The only per-stack divergence permitted is the `asp-append-version` query-string suffix on `.NET` static asset URLs (already exempted by 1.5/1.6 diff allowlist).
   - The role-badge `<span>` is byte-identical (no per-stack class names — Basecoat + the shared `badge-{token}` modifier only).
   - The `<h1>FieldMark</h1>` and the placeholder `<p>` are byte-identical.
   - The verification is integrated into the existing chrome-diff script from Story 1.5 (`tools/parity/diff-chrome.sh` if present, otherwise per-stack `curl` + `diff` in the test harness). The diff is captured as test output, not just an exit-code check, so review can see what changed.

9. **`make parity` exits 0 and Epic 1 is complete.** Story 1.13 adds zero new routes (every stack already has `/` from 1.1; `/login`, `/logout` from 1.11; `/preferences/theme` from 1.6; `/fragments/compliance-tile` from 1.1; auth/identity routes wired by 1.7/1.8/1.9). Verified by:
   - From repo root: `make parity` — exits 0; route inventory diff is empty across the three stacks; `tools/parity/canonical-pg-indexes.txt` is unchanged.
   - The Epic-1 closure assertion: every story key `1-1-*` through `1-13-*` in `_bmad-output/implementation-artifacts/sprint-status.yaml` is at status `done` or `review` (Story 1.13 itself moves to `review` after dev; the retrospective entry remains `optional`). The closing change to `epic-1` status to `done` happens in retrospective (out of scope for this story; this story only enables it).

10. **Each stack's `CLAUDE.md` Home/Identity section is added or updated.** Each stack's `CLAUDE.md` gains a short `## Home page` section (after `## Authorization` from Story 1.12) documenting:
    - Where the Home page lives (file path).
    - That the page is intentionally empty in Epic 1 and is replaced by the dashboard in Story 2.10.
    - The chrome composition order (AC #2) and that any new chrome control must be added in all three stacks in the same commit (FR58).
    - The role→badge-token mapping (AC #4) and that the badge surface is the first cross-stack visual proof of identity.

11. **Build, type, lint, and test gates stay green on every stack.**
    - **.NET:** `cd FieldMark && dotnet csharpier format . && dotnet build && dotnet test` — clean. `TreatWarningsAsErrors=true` honoured. `dotnet csharpier check .` reports zero diffs.
    - **Django:** `cd fieldmark_py && uv run ruff check . && uv run mypy . && uv run pytest` — clean.
    - **Go:** `cd fieldmark-go && make check` — clean (`fmt-check` + `vet` + `staticcheck` + `test`).
    - From repo root: `make parity` — exits 0 (per AC #9).

## Tasks / Subtasks

- [ ] Task 1: Read upstream story artifacts and confirm dependency posture (AC: all)
  - [ ] 1.1 Read [Story 1.5](_bmad-output/implementation-artifacts/1-5-implement-cross-stack-base-layout-with-skip-link-landmarks-and-flashregion.md) — note: layouts already carry skip-link → `<header><nav aria-label="Main">…</nav></header>` → `<main id="main-content">` → `<footer>`. The Home page extends the `<main>` slot only; the chrome edits in AC #2 go inside the existing `<header>`/`<nav>`.
  - [ ] 1.2 Read [Story 1.6](_bmad-output/implementation-artifacts/1-6-implement-themetoggle-with-cookie-persistence-per-stack.md) — note: the ThemeToggle button is already rendered in the header by each layout. Story 1.13 wraps it in the right-cluster `<div>` per AC #2 but does **not** modify the button itself.
  - [ ] 1.3 Read [Story 1.10](_bmad-output/implementation-artifacts/1-10-author-shared-uuid-dev-user-manifest-and-per-stack-idempotent-seed-runners.md) — extract: which dev usernames are seeded (`alice@admin`, `bob@inspector`, etc. or whatever 1.10 chooses); these are the test fixtures for AC #5 / #6 / #7 / #8.
  - [ ] 1.4 Read [Story 1.11](_bmad-output/implementation-artifacts/1-11-login-logout-and-unauthenticated-redirect-across-all-three-stacks.md) — note: `/login` and `/logout` routes exist on every stack. **Check whether Story 1.11 placed a "Sign out" link in the header chrome and where.** If yes, Story 1.13's Task X.3 moves that link into the avatar-menu wrapper described in AC #3 (do not duplicate). If no, this story adds it.
  - [ ] 1.5 Read [Story 1.12](_bmad-output/implementation-artifacts/1-12-implement-authz-can-primitive-and-actionbutton-trichotomy-helper-per-stack.md) — note: `Role.All` (.NET), `Role` enum (Django), `domain.Role` consts (Go) are the canonical role-name sources. Story 1.13's AC #4 mapping extends those types — do **not** add a parallel enum elsewhere.
  - [ ] 1.6 Read each stack's existing Home/landing page to confirm what to delete:
    - `.NET — FieldMark/FieldMark.Web/Pages/Index.cshtml` + `Index.cshtml.cs` (placeholder "Welcome" markup from scaffold).
    - `Django — fieldmark_py/templates/pages/home.html` (and confirm `views.home` is the right handler).
    - `Go — fieldmark-go/internal/web/templates/pages/dashboard.html` (placeholder with compliance-tile hx-get; the hx-get fragment hook must be **removed** at this story per AC #1's "no HTMX-loaded fragment" rule — the compliance tile lands in Story 2.10).
  - [ ] 1.7 Confirm `fieldmark_shared/src/_components.css` exists (Stories 1.4 / 1.12 introduced it). If absent on the current branch, create it and wire it from `src/fieldmark.css`.

- [ ] Task 2: Add Home-page CSS + badge-token + avatar utilities to the shared design system (AC: #3, #4, #7)
  - [ ] 2.1 In `fieldmark_shared/src/_components.css` add:

    ```css
    /* Role badge tokens — pairs with .badge from Basecoat */
    .badge-danger   { background: color-mix(in srgb, var(--color-danger)   15%, transparent); color: var(--color-danger); }
    .badge-info     { background: color-mix(in srgb, var(--color-info)     15%, transparent); color: var(--color-info); }
    .badge-warning  { background: color-mix(in srgb, var(--color-warning)  15%, transparent); color: var(--color-warning); }
    .badge-neutral  { background: color-mix(in srgb, var(--color-neutral)  15%, transparent); color: var(--color-neutral); }
    .badge-success  { background: color-mix(in srgb, var(--color-success)  15%, transparent); color: var(--color-success); }

    /* Avatar (header chrome) */
    .avatar-menu-wrapper { position: relative; display: inline-flex; align-items: center; gap: 0.5rem; }
    .avatar-menu { width: 36px; height: 36px; padding: 0; border: 0; background: transparent; cursor: pointer; }
    .avatar { display: inline-flex; align-items: center; justify-content: center;
              width: 36px; height: 36px; border-radius: 9999px;
              background: var(--color-neutral); color: var(--color-background, #fff);
              font-weight: 600; font-size: 0.75rem; letter-spacing: 0.02em; }

    /* Avatar dropdown: keyboard-revealable, no JS at this story */
    .avatar-menu-wrapper .menu.hidden { display: none; }
    .avatar-menu-wrapper:focus-within .menu.hidden { display: block; position: absolute; right: 0; top: calc(100% + 4px);
                                                     background: var(--color-surface, #fff);
                                                     border: 1px solid var(--color-border, currentColor);
                                                     border-radius: 6px; min-width: 12rem; padding: 0.25rem 0; }
    .menu-item { display: block; padding: 0.5rem 0.75rem; color: inherit; text-decoration: none; }
    .menu-item:hover, .menu-item:focus-visible { background: color-mix(in srgb, currentColor 8%, transparent); }

    /* Wordmark */
    .fm-wordmark { font-weight: 700; letter-spacing: 0.02em; text-decoration: none; color: inherit; }

    /* Placeholder paragraph */
    .text-muted { color: var(--color-muted-foreground, color-mix(in srgb, currentColor 60%, transparent)); }
    ```

  - [ ] 2.2 Rebuild + commit: `cd fieldmark_shared && pnpm run build` then commit the regenerated `dist/fieldmark.css`. **Do not edit `dist/fieldmark.css` by hand** (D16 — manual CSS compile rule; the file is generated).
  - [ ] 2.3 Smoke-check the CSS: open `dist/fieldmark.css` and `grep -E 'badge-(danger|info|warning|neutral|success)' dist/fieldmark.css` — all five must be present.

- [ ] Task 3: .NET implementation — Home page, role badge, avatar partial, tests (AC: #1, #2, #3, #4, #5, #6, #7, #10, #11)
  - [ ] 3.1 Delete the scaffold-era welcome markup in `FieldMark/FieldMark.Web/Pages/Index.cshtml` and replace with the canonical Home markup:

    ```cshtml
    @page
    @model FieldMark.Web.Pages.IndexModel
    @{
        ViewData["Title"] = "FieldMark";
    }
    <h1>FieldMark</h1>
    <span class="badge badge-@Model.RoleBadgeToken" role="status">@Model.RoleLabel</span>
    <p class="text-muted">Your projects will appear here.</p>
    ```

  - [ ] 3.2 Rewrite `Pages/Index.cshtml.cs` to expose `RoleLabel`, `RoleBadgeToken`, `FullName`, `Initials`:

    ```csharp
    using FieldMark.Domain.ValueObjects;
    using FieldMark.Web.Helpers;
    using Microsoft.AspNetCore.Mvc.RazorPages;

    namespace FieldMark.Web.Pages;

    public class IndexModel : PageModel
    {
        public string RoleLabel { get; private set; } = string.Empty;
        public string RoleBadgeToken { get; private set; } = "neutral";
        public string FullName { get; private set; } = string.Empty;
        public string Initials { get; private set; } = "??";

        public void OnGet()
        {
            // User is guaranteed authenticated here (fallback policy from Story 1.11).
            // Resolve the first role alphabetically; multi-role is post-MVP.
            var roleName = User.Claims
                .Where(c => c.Type == System.Security.Claims.ClaimTypes.Role)
                .Select(c => c.Value)
                .OrderBy(s => s, StringComparer.Ordinal)
                .FirstOrDefault() ?? string.Empty;

            var role = Role.All.FirstOrDefault(r => r.Name == roleName);
            if (role is not null)
            {
                RoleLabel = role.Label;          // see Task 3.3
                RoleBadgeToken = role.BadgeToken; // see Task 3.3
            }

            FullName = User.Identity?.Name ?? string.Empty;
            Initials = AvatarInitials.From(FullName, User.Identity?.Name);
        }
    }
    ```

  - [ ] 3.3 Extend `FieldMark/FieldMark.Domain/ValueObjects/Role.cs` (introduced by Story 1.12) with two computed properties — `Label` (Title-Case-With-Space form per AC #4) and `BadgeToken` (token per the AC #4 mapping). Domain has zero outbound references — both are pure string constants on the value object. Update Story 1.12's `Role.Parse` and `Role.All` tests to cover the new properties (no new test files; extend `CanTests.cs` or split out a `RoleTests.cs`).
  - [ ] 3.4 Create `FieldMark/FieldMark.Web/Helpers/AvatarInitials.cs` with `public static string From(string? fullName, string? usernameFallback)` implementing the algorithm in AC #3. Unit tests in `FieldMark.Tests.Domain/Helpers/AvatarInitialsTests.cs` — but note: helper lives in `FieldMark.Web`, so the test must live in an integration test project or a new `FieldMark.Tests.Web` xUnit project. **Preferred path:** put the helper in `FieldMark.Domain/Services/AvatarInitials.cs` (pure string transformation, no DI, zero outbound references — stays Domain-pure) and the tests in `FieldMark.Tests.Domain/Services/AvatarInitialsTests.cs`. Update the `using` in `IndexModel` to point at the Domain namespace.
  - [ ] 3.5 Update `FieldMark/FieldMark.Web/Pages/Shared/_Layout.cshtml` — restructure the `<header><nav aria-label="Main">…</nav></header>` block to the canonical chrome order per AC #2:

    ```cshtml
    <header>
      <nav aria-label="Main">
        <a href="/" class="fm-wordmark" aria-label="FieldMark home">FieldMark</a>
        <div class="ml-auto flex items-center gap-3">
          <partial name="_ThemeToggle" model="@theme" />
          @if (User.Identity?.IsAuthenticated == true)
          {
              <partial name="_AvatarMenu" model="@(new FieldMark.Web.ViewModels.AvatarMenuVm(User))" />
          }
        </div>
      </nav>
    </header>
    ```

    The pre-existing skip-link before `<header>` is unchanged. If Story 1.11 already added a Logout link in the header outside an avatar dropdown, **remove it** from the layout — it now lives inside the avatar menu.

  - [ ] 3.6 Create `FieldMark/FieldMark.Web/Pages/Shared/_AvatarMenu.cshtml` and its strongly-typed VM `FieldMark/FieldMark.Web/ViewModels/AvatarMenuVm.cs`. The VM exposes `FullName`, `Initials`, `RoleLabel` derived from the `ClaimsPrincipal` constructor argument (call `AvatarInitials.From(...)` once; do not recompute per render). Markup is verbatim per AC #3 — `<div class="avatar-menu-wrapper relative">` wrapping the `<button>` and the `<ul id="avatar-menu-dropdown" class="menu hidden" role="menu">`.
  - [ ] 3.7 Add unit tests:
    - `FieldMark.Tests.Domain/ValueObjects/RoleTests.cs` — covers `Role.All[i].Label` and `Role.All[i].BadgeToken` for all five canonical roles (table-driven xUnit `[Theory]`).
    - `FieldMark.Tests.Domain/Services/AvatarInitialsTests.cs` — covers the seven cases listed in AC #3.
  - [ ] 3.8 Add integration tests in `FieldMark.Tests.Integration/Pages/HomePageTests.cs`:
    - `Home_Unauthenticated_RedirectsToLogin` — GET `/` with no auth cookie returns 302 with `Location: /login` (AC #5).
    - `Home_AuthenticatedAdmin_RendersRoleBadgeAndPlaceholder` — sign in via the test helper (Story 1.11 provides one), GET `/`, assert `<h1>FieldMark</h1>`, `<span class="badge badge-danger" role="status">Admin</span>`, and the placeholder paragraph all present in order.
    - `Home_AuthenticatedAnyRole_PassesAxe` — render `/`, capture HTML, run axe-core via the `Deque.AxeCore.Playwright` integration if present in solution; otherwise invoke `npx @axe-core/cli` against a temp file. Assert zero violations (AC #6).
    - `Home_TabOrder_MatchesContract` — using a headless browser harness from Story 1.5 or 1.11 (`Playwright` is preferred), `Tab` repeatedly and assert focused element sequence (AC #7).
    - `Home_ChromeMatchesParitySnapshot` — render with the canonical dev user, normalize, compare to `_bmad-output/implementation-artifacts/_parity-snapshots/home-chrome.normalized.html` (the snapshot file is authored by Task 7 and committed).
  - [ ] 3.9 Update `FieldMark/CLAUDE.md` — add `## Home page` section per AC #10.

- [ ] Task 4: Django implementation — Home page, role badge, avatar partial, tests (AC: #1, #2, #3, #4, #5, #6, #7, #10, #11)
  - [ ] 4.1 Rewrite `fieldmark_py/templates/pages/home.html` to the canonical Home markup:

    ```django
    {% extends "base.html" %}
    {% block title %}FieldMark{% endblock %}
    {% block content %}
      <h1>FieldMark</h1>
      <span class="badge badge-{{ role_badge_token }}" role="status">{{ role_label }}</span>
      <p class="text-muted">Your projects will appear here.</p>
    {% endblock %}
    ```

  - [ ] 4.2 Update `fieldmark_py/fieldmark/views.py` — replace the existing `home(request)` with a version that resolves role and renders the canonical context. Authentication enforcement is the `LoginRequiredMiddleware` from Story 1.11; do **not** add `@login_required` (would double-protect with no benefit and break parity with the .NET fallback policy idiom):

    ```python
    from django.shortcuts import render
    from fieldmark.roles import Role, BADGE_TOKENS, LABELS

    def home(request):
        # request.user is guaranteed authenticated here (LoginRequiredMiddleware from 1.11)
        group_names = sorted(request.user.groups.values_list("name", flat=True))
        role_name = group_names[0] if group_names else ""
        try:
            role = Role(role_name)
        except ValueError:
            role = None
        return render(request, "pages/home.html", {
            "role_label": LABELS.get(role, ""),
            "role_badge_token": BADGE_TOKENS.get(role, "neutral"),
        })
    ```

  - [ ] 4.3 Extend `fieldmark_py/fieldmark/roles.py` (Story 1.12) with two module-level dicts: `LABELS: dict[Role, str]` (Title-Case-With-Space per AC #4) and `BADGE_TOKENS: dict[Role, str]` (per AC #4 mapping). **Both keyed by the `Role` enum** — not by the string value — so a typo on the call-site fails at type-check time under `mypy`.
  - [ ] 4.4 Create `fieldmark_py/fieldmark/avatar.py:initials(full_name: str | None, username_fallback: str | None) -> str` implementing the AC #3 algorithm. Unit tests in `fieldmark_py/fieldmark/tests/test_avatar.py` covering the seven cases.
  - [ ] 4.5 Update `fieldmark_py/templates/base.html` — restructure the existing `<header><nav aria-label="Main">…</nav></header>` block to the canonical chrome order per AC #2, mirroring the .NET layout:

    ```django
    <header>
      <nav aria-label="Main">
        <a href="/" class="fm-wordmark" aria-label="FieldMark home">FieldMark</a>
        <div class="ml-auto flex items-center gap-3">
          {% include "_theme_toggle.html" with current=fm_theme %}
          {% if request.user.is_authenticated %}
            {% include "_avatar_menu.html" with user=request.user %}
          {% endif %}
        </div>
      </nav>
    </header>
    ```

  - [ ] 4.6 Create `fieldmark_py/templates/_avatar_menu.html` rendering the AC #3 markup. It computes initials inline via a small template tag (`fieldmark/templatetags/avatar.py:initials_for(user)`) — keep template logic minimal; the actual implementation is the `avatar.py` helper. Register the template tag library in the template via `{% load avatar %}` at the top.
  - [ ] 4.7 Add integration tests in `fieldmark_py/fieldmark/tests/test_home_page.py`:
    - `test_home_unauthenticated_redirects_to_login` — client.get("/") with no session asserts 302 and `Location: /login` (AC #5). Uses `@pytest.mark.django_db`.
    - `test_home_authenticated_admin_renders_role_badge_and_placeholder` — log in as the seeded admin user from Story 1.10; GET `/`; assert markup per AC #1 and #4.
    - `test_home_authenticated_any_role_passes_axe` — render, capture HTML, run `axe-core` via `subprocess.run(["npx", "@axe-core/cli", html_file])` and assert zero violations.
    - `test_home_tab_order_matches_contract` — uses `pytest-playwright` if present; otherwise document a manual recipe and skip the test (annotate with `pytest.mark.manual_skip`).
    - `test_home_chrome_matches_parity_snapshot` — normalize, compare to the same snapshot file from Task 3.8.
  - [ ] 4.8 Update `fieldmark_py/CLAUDE.md` per AC #10.

- [ ] Task 5: Go/Fiber implementation — Home page, role badge, avatar partial, tests (AC: #1, #2, #3, #4, #5, #6, #7, #10, #11)
  - [ ] 5.1 Rename `fieldmark-go/internal/web/templates/pages/dashboard.html` → `pages/home.html`. Strip the compliance-tile placeholder and the "Go (Fiber) stack — standup milestone placeholder" line. New content:

    ```go-template
    {{template "base" .}}

    {{define "title"}}FieldMark{{end}}

    {{define "content"}}
    <h1>FieldMark</h1>
    <span class="badge badge-{{.RoleBadgeToken}}" role="status">{{.RoleLabel}}</span>
    <p class="text-muted">Your projects will appear here.</p>
    {{end}}
    ```

  - [ ] 5.2 Extend `fieldmark-go/internal/domain/role.go` (Story 1.12) with methods `func (r Role) Label() string` and `func (r Role) BadgeToken() string` per AC #4. Pure functions; zero outbound non-stdlib imports.
  - [ ] 5.3 Create `fieldmark-go/internal/web/viewmodels/avatar.go:Initials(fullName, usernameFallback string) string` per AC #3. Unit tests in `fieldmark-go/internal/web/viewmodels/avatar_test.go`.
  - [ ] 5.4 In `fieldmark-go/cmd/web/main.go` (or wherever the `/` handler lives after Story 1.11):
    - Update the `/` handler (now under the authenticated `app.Group("/", auth.RequireAuth())` block from Story 1.11) to render `pages/home` with a `fiber.Map` containing `Title`, `FmTheme` (already injected by 1.6), `Actor` (already in `c.Locals("user")` from 1.9), `RoleLabel`, `RoleBadgeToken`, `Initials`, `FullName`.
    - Helper: `func renderHomeContext(c fiber.Ctx) fiber.Map { ... }` — keeps the handler thin.
    - **Do not** preserve the old `/fragments/compliance-tile` hx-trigger="load" — Story 2.10 introduces the real dashboard; the Home page is empty until then.
  - [ ] 5.5 Update `fieldmark-go/internal/web/templates/partials/header.html` to the canonical chrome order per AC #2:

    ```go-template
    {{define "header"}}
    <header>
      <nav aria-label="Main">
        <a href="/" class="fm-wordmark" aria-label="FieldMark home">FieldMark</a>
        <div class="ml-auto flex items-center gap-3">
          {{template "theme_toggle" .}}
          {{if .Actor.IsAuthenticated}}{{template "avatar_menu" .}}{{end}}
        </div>
      </nav>
    </header>
    {{end}}
    ```

    If Story 1.11 placed a logout link directly in the header, move it inside the avatar_menu partial.

  - [ ] 5.6 Create `fieldmark-go/internal/web/templates/partials/avatar_menu.html` per AC #3 markup, referencing `.Actor.FullName`, `.Initials`, `.RoleLabel` from the view-model context.
  - [ ] 5.7 Add integration tests:
    - `fieldmark-go/internal/web/handlers/home_test.go` (or `cmd/web/main_test.go` if the handler lives there) — `TestHomeUnauthenticatedRedirectsToLogin`, `TestHomeAuthenticatedRendersRoleBadgeAndPlaceholder`, `TestHomeChromeMatchesParitySnapshot`. Use Fiber's `app.Test(req)` harness already established by Stories 1.9 / 1.11.
    - axe-core run: launch the binary against a test port, `curl /` with a test auth cookie, pipe through `npx @axe-core/cli`. If the Go test environment cannot invoke Node, skip with a documented manual recipe.
    - Tab-order test: use `chromedp` if a dep is already in `go.mod`; otherwise document the manual recipe and skip.
  - [ ] 5.8 Update `fieldmark-go/CLAUDE.md` per AC #10.

- [ ] Task 6: Author the parity-snapshot fixture (AC: #8)
  - [ ] 6.1 Create `_bmad-output/implementation-artifacts/_parity-snapshots/` directory.
  - [ ] 6.2 Render `/` on each stack with the **canonical dev user** from Story 1.10 (the user whose role is `ADMIN`, so the badge is `badge-danger Admin` — most distinctive for snapshot diffs).
  - [ ] 6.3 Normalize each captured HTML through the chrome-normalization pipeline (Story 1.5's `tools/parity/normalize-html.sh` or the per-stack normalizer helpers from Story 1.12).
  - [ ] 6.4 Confirm the three normalized outputs are byte-identical. If they diverge: **fix the markup, not the snapshot** — divergence is a defect (FR58).
  - [ ] 6.5 Commit the single byte-identical normalized HTML as `_bmad-output/implementation-artifacts/_parity-snapshots/home-chrome.normalized.html`. This is the test fixture for AC #8's per-stack integration tests.

- [ ] Task 7: Cross-stack verification, parity, and Epic-1 closure preparation (AC: #8, #9)
  - [ ] 7.1 Start all three stacks (`make up` then `make run-net`, `make run-django`, `make run-go` in separate shells).
  - [ ] 7.2 `curl -s --cookie "<dev-admin-session-cookie>" http://localhost:5000/ http://localhost:8000/ http://localhost:3000/` — capture each output; pass through the chrome normalizer; diff. Must be byte-identical (modulo the documented `asp-append-version` exemption).
  - [ ] 7.3 `make parity` — must exit 0; route inventory diff empty; `pg_indexes` diff empty.
  - [ ] 7.4 Add a tiny cross-stack mapping snapshot test: `tools/parity/role-badge-tokens.sh` — invokes one read endpoint per stack (or one CLI subcommand) that prints `<role-name>\t<token>\t<label>` for each of the five roles. Asserts the three outputs are byte-identical. Cheapest path: a `--dump-role-badges` flag added to each stack's existing CLI (already-existing pattern from Story 1.6's `--dump-routes`). If that surface area is too much, drop this to a documented manual recipe and capture the three outputs as an asciidoc in `tools/parity/role-badge-tokens.expected.txt` — but the **automated** path is preferred and a non-skip for Epic-1 closure rigor.
  - [ ] 7.5 Walk every story key in Epic 1 in `sprint-status.yaml` — confirm all are `done` or `review` (Story 1.13 itself moves to `review` after dev). Note any still in `in-progress` or `backlog` and surface in the completion notes — Epic 1 retrospective will run after they close.

- [ ] Task 8: Verify all gates green (AC: #11)
  - [ ] 8.1 **.NET:** `cd FieldMark && dotnet csharpier format . && dotnet build && dotnet test` — all green; `dotnet csharpier check .` reports zero diffs.
  - [ ] 8.2 **Django:** `cd fieldmark_py && uv run ruff check . && uv run mypy . && uv run pytest` — all green.
  - [ ] 8.3 **Go:** `cd fieldmark-go && make check` — all green.
  - [ ] 8.4 From repo root: `make parity` — exits 0.

## Dev Notes

### Brownfield posture — what exists today (read before writing anything)

Cross-stack state at HEAD of branch `feature/1.6_theme-toggle`:

- **Stories landed:** 1.1, 1.2, 1.3, 1.4, 1.5 are `done`. Story 1.6 is `in-progress` (this branch). Stories 1.7–1.12 are `ready-for-dev`. **Stories 1.7 through 1.12 are prerequisites for this story** — Story 1.13 consumes `Role.All` / `Role` / `domain.Role` (1.12), `Can` (1.12, transitively via the fallback authorization), `/login` and `/logout` (1.11), the dev-user manifest with deterministic UUIDs (1.10), and the framework-native authentication wiring (1.7/1.8/1.9). When this story begins, all of 1.7–1.12 must be `review` or `done` — flag any that are not at story-kickoff time. If any are still `ready-for-dev`, coordinate sequencing rather than block-implement around them; the chrome edits depend on the avatar/role surface being real, not stubbed.
- **The current Home pages are scaffold placeholders.** All three stacks render a "Welcome" / "Dashboard placeholder" page at `/` today (the .NET `Index.cshtml` carries Microsoft's template welcome paragraph; the Go `dashboard.html` carries a compliance-tile hx-get that is **not** real — it's a stand-up demo placeholder). Story 1.13 replaces all three with the canonical Home markup and removes those placeholders entirely. The Go dashboard's `hx-get="/fragments/compliance-tile"` is **not** preserved — the real Compliance Dashboard lands in Story 2.10 (Epic 2). Leaving it in would render an empty broken tile on the Home page; remove it and let 2.10 add the dashboard back as a separate page or as the new contents of Home.
- **`<header>` chrome currently carries the wordmark and the ThemeToggle.** Story 1.5 introduced the landmark structure; Story 1.6 added the ThemeToggle. Story 1.11 adds a "Sign out" surface (link or form-post button) — its exact placement is whatever 1.11 chose; Story 1.13 normalizes that placement into the avatar dropdown.
- **`fieldmark_shared/src/_components.css` is the right home for the new CSS** — Stories 1.4 / 1.12 established it. If it doesn't exist on the branch when this story starts, create it and `@import` it from `src/fieldmark.css` (the import order must keep Basecoat first, then tokens, then `_a11y.css` from 1.5, then `_components.css`).
- **No JS is added by this story.** The avatar dropdown is keyboard-revealable via `:focus-within`. A click-toggle is post-MVP; do not introduce one here even though "it would be nicer" — Epic 1's JS budget is intentionally near-zero (5 lines of inline first-paint resolver from 1.6 + 20 lines for theme-toggle.js + AG Grid + HTMX). Adding 10–20 LOC for a dropdown toggle is in scope for Epic 2's first JS-bearing component, not here. (See architecture's JS budget line in UX spec §932–939.)

### Why role label & badge-token live on the `Role` value object (and not on the view model)

Story 1.12 makes `Role` the single source of truth for role names across each stack. Story 1.13's mapping (role → label, role → badge-token) is the second axis of that table. Two ways to add it:

1. **Carry it on `Role`** — `Role.Admin.Label`, `Role.Admin.BadgeToken`. One source of truth; view models read from it; cross-stack diff is structural (the type itself).
2. **Carry it on the view model** — `IndexModel.RoleLabel`, `IndexModel.RoleBadgeToken`. Convenient locally, but every future screen that wants a role badge (audit row, user picker, EntityRail header, …) re-implements the mapping. Two implementations diverge.

The architecture's stack-symmetry hard rule (root CLAUDE.md "Stack symmetry … Divergence = defect") favors (1). The Domain layer's "zero outbound references" rule is preserved because both properties are pure strings.

A near-future alternative is to push the mapping into a separate `RoleAffordance` value object (`Role + Label + BadgeToken + Icon + …`) — that's right when the table grows past two columns. At two columns, putting them on `Role` is the simpler shape.

### Why "first role alphabetically" for multi-role users

Story 1.9's Go stub middleware already chose this tie-break (`auth/lookup.go` orders `user_roles.role` ASC and takes the first). The .NET and Django stacks at MVP rarely have multi-role users (Story 1.10's dev manifest seeds one role per user), but if/when they do, the badge must be deterministic. "First alphabetically" is:

- Deterministic — same input, same output, across stacks.
- Cheap — no role hierarchy or precedence table to author and maintain.
- Honest — when multi-role becomes a real product surface (post-MVP), this default will be replaced by a UI affordance ("you have 3 roles — switch context") and a per-session active-role cookie. The badge color/label at that point reflects the *active* role, not a tie-break. Story 1.13's tie-break is explicitly a placeholder for that future affordance.

The alternative "highest-privilege wins" (ADMIN > COMPLIANCE_OFFICER > …) imports a role hierarchy concept the system does not have anywhere else (`Can` decides by membership, not by ordering). Don't introduce it for a badge.

### Why the avatar dropdown opens on `:focus-within` instead of click

Three reasons:

1. **Zero JS.** The trichotomy is markup; the visibility rule is CSS. Adding a click toggle requires (a) an event listener (b) ARIA state management (`aria-expanded` true/false) (c) a click-outside dismisser (d) Escape-key dismisser. All of that is fine to write, but it's ~30 LOC of JS for a control that Epic 1 has no users for. Defer to Epic 2's first JS-bearing component story.
2. **Discoverability via keyboard.** The Tab order is wordmark → theme-toggle → avatar-button → logout-link. When the avatar button is focused (or any descendant), `:focus-within` reveals the dropdown; tabbing forward focuses the logout link (which is inside the now-visible dropdown); tabbing again leaves the `:focus-within` zone and the dropdown hides. Native keyboard semantics; no JS.
3. **Mouse users can't see the menu yet.** This is the conscious tradeoff. A mouse user landing on Home sees an avatar circle with their initials and the role badge but cannot pop the menu open with a click. They can sign out by typing `/logout` or by Tab-Tab-Tab. **This is acceptable for Epic 1 because the only Epic-1 user action other than the theme toggle is "sign out", and the test users for Epic 1 are developers verifying the parity surface — not real end users.** Story 2.x will add the click toggle as part of the first user-facing flow.

If the click-toggle gap feels too sharp during code review, the acceptable compromise is to add a single CSS rule that also reveals on `:hover` (`.avatar-menu-wrapper:hover .menu.hidden { display: block; … }`). That gives mouse users an open path while keeping JS at zero. Decide during review; both shapes are inside the cross-stack contract.

### Why removing the Go compliance-tile placeholder is in scope

The Go `pages/dashboard.html` carries:

```html
<div id="compliance-tile" hx-get="/fragments/compliance-tile" hx-trigger="load" …>
```

This was Story 1.1's stand-up milestone proof that HTMX + template engine + handler wiring all worked. Story 2.10 introduces the real Compliance Dashboard with the real `#compliance-tile` (and friends). Leaving the placeholder on the Home page now means:

- On page load, Home fires an `hx-get` to a fragment endpoint that returns a placeholder tile. The user sees a flicker. AC #1 ("no HTMX-loaded fragment") forbids this.
- AC #6 (axe-core zero violations) is at risk — the placeholder's contrast may or may not pass under Basecoat's dark theme; the cheapest defense is to delete the placeholder.
- Cross-stack parity fails: .NET and Django do not have this fragment surface on their Home pages; only Go does. Either add it to all three or remove it from Go. Removing is correct because Epic 1 has no Compliance Dashboard requirement — that's Epic 2.

Story 2.10 reintroduces the dashboard as the **new contents of Home** (the empty Home becomes the populated dashboard for authenticated users). At that point Home's `<h1>` may change from `FieldMark` to the dashboard's title; the role badge may move to the avatar dropdown; the placeholder paragraph is deleted. That refactor is Story 2.10's, not 1.13's — 1.13 just makes Home cleanly empty.

### Why the parity snapshot is one file, not three

A cross-stack byte-parity contract has exactly one expected output. Three "expected" files (one per stack) would silently allow drift — a maintainer fixing a Django-side change would update only the Django expected file, .NET and Go would continue to match their own stale fixtures, and the cross-stack diff would still pass (each test against its own file) even though the three runtime outputs had diverged.

One file forces the three integration tests to diff against the same source of truth. If any stack's output diverges, that stack's test fails with a readable diff against the canonical snapshot. The fix is in that stack's markup, not in the snapshot file.

The snapshot's directory `_bmad-output/implementation-artifacts/_parity-snapshots/` is new — the underscore prefix keeps it visually separate from the per-story `.md` files. Future stories that have a cross-stack byte-parity requirement (e.g., 2.4 Phase-2 components) drop their normalized fixtures here too.

### Why the helper `AvatarInitials.From(...)` lives in Domain (.NET) and `avatar.py` (Django) and `viewmodels/avatar.go` (Go)

The placement is **per-stack-idiomatic**, not byte-identical. The architectural rule each stack honors:

- **.NET:** `FieldMark.Domain` has zero outbound references. A pure `string → string` transformation belongs there if it has no DI, no I/O. Putting it in `FieldMark.Web/Helpers/` would split the helper away from the `Role` types it conceptually neighbors. Putting it in Domain alongside `Role.cs` is the right shape. (`FieldMark.Domain/Services/` is fine; or just inline as a static method on a `User` value object when one exists. At Story 1.13 there is no `User` value object — Domain has no users; that's Epic 5+ territory. Stand-alone static is correct for now.)
- **Django:** Django doesn't have a layered architecture concept analogous to "Domain" — Story 1.12 placed `roles.py` and `authz.py` directly under the `fieldmark/` project package, and `avatar.py` belongs there too. Tests in `fieldmark/tests/test_avatar.py`.
- **Go:** Go's `internal/domain/` is reserved for entities and domain methods; the avatar helper is presentation-flavored (it produces display text from a name string), so it belongs in `internal/web/viewmodels/` alongside the view models that consume it. This mirrors Story 1.12's placement of `viewmodels/action_button.go`.

The behavioral contract — the algorithm in AC #3 — is byte-identical across stacks. The placement and idiom may differ. Cross-stack tests in Task 7.4 (if implemented as the `--dump-role-badges`-shaped surface) prove the **output** matches; the **placement** is per-stack-idiomatic.

### Anti-patterns that must NOT slip in

- ❌ Adding any inline `<script>` on the Home page. The 5-line first-paint resolver from Story 1.6 is the **only** inline JS permitted in the application.
- ❌ Adding an `hx-get` to the Home page body to load any fragment. Home is empty; no AJAX.
- ❌ Adding `@AllowAnonymous` (.NET) / `@login_not_required` (Django) / removing the `/` route from the `auth.RequireAuth()` group (Go) to "make the page render in the parity test without auth." The parity tests must use a real authenticated session via Story 1.11's test helper.
- ❌ Hard-coding role labels (`"Admin"`, `"Compliance Officer"`) anywhere except the `Role` value object's `Label` property / `LABELS` dict / `Label()` method.
- ❌ Hard-coding badge tokens (`"danger"`, `"info"`, …) anywhere except `Role.BadgeToken` / `BADGE_TOKENS` / `Role.BadgeToken()`. A future contributor adding a role badge to an AuditRow in Epic 2 must read from the same source.
- ❌ Replacing `<h1>FieldMark</h1>` with `<h1>{{ user.name }}</h1>` or `<h1>Welcome, {{ user.name }}</h1>` or any other variant. AC #1 is explicit. The user's name lives in the avatar's accessible description; it is not the page heading.
- ❌ Adding a second `<h1>` to the page (e.g., for the role badge). One `<h1>` per page (UX-DR33 from Story 1.5).
- ❌ Putting the role badge **before** the `<h1>` in DOM order. AC #1 fixes the order: `<h1>` then `<span class="badge">` then `<p>`.
- ❌ Using a `<div>` for the role badge instead of `<span role="status">`. The element is inline; the role is `status` so screen readers treat it as a live region update target (consistent with Story 1.5's FlashRegion pattern).
- ❌ Using Tailwind classes like `bg-red-500 text-white` directly on the badge. The whole point of `--color-*` semantic tokens (Story 1.4) is that theme switches re-color the badge without markup changes. Use `.badge-danger` etc.
- ❌ Adding a sixth role (or worse — adding a "Guest" / "Anonymous" role badge to handle the unauthenticated case). Unauthenticated users do not see Home (AC #5 redirect). The role-resolution code paths that fall back to "no role" are defensive; they should not render a badge at all (omit the `<span>`) rather than render a placeholder. The integration test for AC #5 verifies the redirect; the "no role" path is not a rendered surface.
- ❌ Wiring a JS dropdown toggle in this story. (Acceptable: add a `:hover` reveal rule to the CSS as a mouse-user accommodation — see Dev Notes "Why … `:focus-within`".)
- ❌ Adding a `data-testid` attribute "to make the integration test easier." The integration tests use stable ARIA attributes (`role="status"`, `aria-label`), stable IDs (`#main-content`, `#avatar-menu-dropdown`), and stable class names (`.fm-wordmark`, `.theme-toggle`, `.avatar-menu`, `.badge-danger`, …). `data-testid` would pollute the parity snapshot.
- ❌ Editing the Story 1.5 normalize-html pipeline or the Story 1.12 attribute-canonicalization helper. They are shared — edits must be cross-stack consistent and are out of scope for this story. If you find a normalization gap, document it in `deferred-work.md` and proceed.

### Project Structure Notes

Files this story adds:

- **Shared:** updates to `fieldmark_shared/src/_components.css`; regenerated `fieldmark_shared/dist/fieldmark.css`.
- **.NET (new):** `FieldMark/FieldMark.Domain/Services/AvatarInitials.cs`, `FieldMark/FieldMark.Web/Pages/Shared/_AvatarMenu.cshtml`, `FieldMark/FieldMark.Web/ViewModels/AvatarMenuVm.cs`, `FieldMark.Tests.Domain/ValueObjects/RoleTests.cs`, `FieldMark.Tests.Domain/Services/AvatarInitialsTests.cs`, `FieldMark.Tests.Integration/Pages/HomePageTests.cs`.
- **Django (new):** `fieldmark_py/fieldmark/avatar.py`, `fieldmark_py/fieldmark/templatetags/__init__.py`, `fieldmark_py/fieldmark/templatetags/avatar.py`, `fieldmark_py/templates/_avatar_menu.html`, `fieldmark_py/fieldmark/tests/test_avatar.py`, `fieldmark_py/fieldmark/tests/test_home_page.py`.
- **Go (new):** `fieldmark-go/internal/web/viewmodels/avatar.go`, `fieldmark-go/internal/web/viewmodels/avatar_test.go`, `fieldmark-go/internal/web/templates/partials/avatar_menu.html`, `fieldmark-go/internal/web/handlers/home_test.go` (or `cmd/web/main_test.go`).
- **Shared snapshot:** `_bmad-output/implementation-artifacts/_parity-snapshots/home-chrome.normalized.html` (one file; the cross-stack contract).
- **Optional cross-stack tool:** `tools/parity/role-badge-tokens.sh` + per-stack `--dump-role-badges` flag plumbing (Task 7.4; promote to required if review wants the closure rigor).

Files this story updates:

- **.NET (update):** `FieldMark/FieldMark.Domain/ValueObjects/Role.cs` (add `Label` + `BadgeToken`), `FieldMark/FieldMark.Web/Pages/Index.cshtml` + `Index.cshtml.cs` (rewrite), `FieldMark/FieldMark.Web/Pages/Shared/_Layout.cshtml` (chrome order), `FieldMark/CLAUDE.md` (Home page section).
- **Django (update):** `fieldmark_py/fieldmark/roles.py` (add `LABELS`, `BADGE_TOKENS`), `fieldmark_py/fieldmark/views.py` (rewrite `home`), `fieldmark_py/templates/pages/home.html` (rewrite), `fieldmark_py/templates/base.html` (chrome order), `fieldmark_py/CLAUDE.md` (Home page section).
- **Go (update):** `fieldmark-go/internal/domain/role.go` (add `Label()` + `BadgeToken()`), `fieldmark-go/internal/web/templates/pages/dashboard.html` → `home.html` (rename + rewrite), `fieldmark-go/internal/web/templates/partials/header.html` (chrome order), `fieldmark-go/cmd/web/main.go` (handler context), `fieldmark-go/CLAUDE.md` (Home page section).

All file locations align with [Architecture Repository Directory Structure](_bmad-output/planning-artifacts/architecture.md#complete-repository-directory-structure):

- `FieldMark.Domain/Services/` — line 1025 (Domain-pure helpers).
- `FieldMark.Web/Pages/Shared/_AvatarMenu.cshtml` — sibling of `_Layout.cshtml`, `_ThemeToggle.cshtml`, `_ActionButton.cshtml` per the established Razor partial pattern.
- `fieldmark/templatetags/` — line 1131 (template-tag-library home).
- `internal/web/viewmodels/` — line 1224 (view-model home; sibling of `action_button.go` from 1.12).
- `internal/web/templates/partials/` — line 1218 (partial home; sibling of `theme_toggle.html` from 1.6).

### Testing Standards

Per [Architecture Testing](_bmad-output/planning-artifacts/architecture.md) and each stack's `CLAUDE.md`:

- **.NET:** Unit tests in `FieldMark.Tests.Domain/` for `Role.Label`, `Role.BadgeToken`, `AvatarInitials.From` (pure-logic; no DB). Integration tests in `FieldMark.Tests.Integration/` for the Home page (uses `WebApplicationFactory<>` with the test auth scheme from Story 1.11). Real Postgres via Testcontainers — never SQLite.
- **Django:** Unit tests in `fieldmark_py/fieldmark/tests/test_avatar.py` (pure-logic; no DB). Integration tests in `test_home_page.py` use `@pytest.mark.django_db` plus `django.test.Client.force_login(...)` with a seeded dev user from Story 1.10. Real Postgres via `pytest-django`.
- **Go:** Standard library `testing` only — no testify. Unit tests for `Role.Label()` / `Role.BadgeToken()` / `Initials(...)` are pure-logic. Integration tests for the Home page use Fiber's `app.Test(httptest.NewRequest(...))` harness from Stories 1.9 / 1.11. For axe-core, shell out to `npx @axe-core/cli` via `os/exec`; skip with a `t.Skip("npm not available")` if the binary is missing — surface the skip count in CI summary.
- **All stacks:** Reuse the normalize-html helper from Stories 1.5 / 1.6 / 1.12 — **do not** author a fourth normalizer. The parity-snapshot comparison runs the same normalizer over the captured runtime HTML and the committed snapshot.
- **axe-core specifics:** Use the WCAG 2.1 AA ruleset (axe's default tag set: `wcag2a`, `wcag2aa`, `wcag21a`, `wcag21aa`). Do **not** disable rules — if a rule fires, fix the markup. The expected count is zero on every stack. Document the axe version in each stack's `package.json` / `requirements-dev.txt` / `tools/axe-version.txt` (or rely on the `@axe-core/cli` version pinned in CI).
- **Tab-order specifics:** Each stack's integration test programmatically focuses each element via the harness's `keyboard.press("Tab")` or equivalent, captures `document.activeElement` after each press, and asserts the sequence via a stable selector list. Five Tab presses cover the contract (Skip → Wordmark → ThemeToggle → Avatar → Logout); a sixth Tab lands on the page body's first focusable element or `<body>` itself.

### Previous Story Intelligence

**Story 1.5 (`done`).** Lessons:
- The base layout file in each stack already carries `<header><nav aria-label="Main">`, `<main id="main-content">`, `<footer>` plus the skip-link. Story 1.13 edits the *inside* of `<nav>` (Task X.5 per stack) without touching the landmark structure. The skip-link is unchanged and remains the first focusable element.
- The `_FlashRegion.cshtml` / `_flash_region.html` / `partials/flash_region.html` partials remain stubbed-empty in this story — Home does not flash any messages on first render. (Story 1.11 may add a one-shot "Signed in as …" flash; that's 1.11's concern.)
- The chrome-normalization pipeline that Story 1.5 established in its dev-agent verification (curl + normalize + diff) is the same pipeline this story's parity snapshot uses.

**Story 1.6 (`in-progress` on this branch).** Lessons:
- The ThemeToggle button is rendered inside each stack's `<header>` block. Story 1.13's chrome edits wrap it in a right-side cluster `<div>` — the button itself is unmodified.
- The first-paint inline `<script>` and the `theme-toggle.js` listener are unchanged.
- The `data-theme` attribute on `<html>` is unchanged.

**Story 1.7 (.NET — likely `review` or `done` when 1.13 starts).** Lessons:
- `ClaimsPrincipal.User` carries role claims after 1.7's Identity wiring; `User.Claims.Where(c => c.Type == ClaimTypes.Role)` is the canonical extraction shape. The Razor `User.IsInRole("ADMIN")` short-cut also works but does not surface the role name string we need for the badge.
- `User.Identity?.Name` returns the username after Identity sign-in. The full-name field is custom (added in 1.7 or 1.10 — confirm); if absent at this story, fall back to `User.Identity.Name`.

**Story 1.8 (Django).** Lessons:
- `request.user.is_authenticated` is the guard; `LoginRequiredMiddleware` from 1.11 enforces it before the view runs.
- `request.user.groups.values_list("name", flat=True)` returns the role names (Story 1.8's role-Group seeding). `sorted(...)` gives the alphabetical tie-break.

**Story 1.9 (Go).** Lessons:
- `c.Locals("user").(*app.Actor)` is the canonical extractor (Story 1.9 sets it in `StubAuthMiddleware`).
- `Actor.IsAuthenticated()` is a method (or `actor.Username != "anonymous"` is the literal check). Use the method shape — it's cleaner in templates.
- `Actor.Roles` is a slice; the first element is the alphabetical-first role per 1.9's stub.

**Story 1.10 (dev-user manifest seeding).** Lessons:
- The same UUID per username is seeded across all three stacks; the dev admin user is the recommended fixture for the parity snapshot (consistent role across stacks = consistent badge).
- The full-name field per user is set; use it for the avatar initials. The username field is a fallback if full-name is empty.

**Story 1.11 (login / logout).** Lessons:
- `/login` and `/logout` routes are registered on every stack. The Home page's avatar dropdown points at `/logout` directly.
- The unauthenticated-redirect contract for `/` is wired (`.NET` fallback policy; `Django` `LoginRequiredMiddleware`; `Go` `auth.RequireAuth()` on the route group). Story 1.13's AC #5 reasserts; no new wiring is needed.
- The test helper for "sign in as user X" lives in each stack's integration test project — Story 1.13's integration tests should reuse it, not author a parallel sign-in path.
- **If Story 1.11 placed a "Sign out" surface in the header outside an avatar dropdown,** Story 1.13's chrome edits relocate it into the avatar menu per AC #3. Surface this as a coordination point in dev-review; the alternative (leaving Logout outside the avatar) breaks the AC #7 tab-order contract.

**Story 1.12 (`authz.Can` + ActionButton).** Lessons:
- `Role.All` (.NET) / `Role` enum (Django) / `domain.Role` const set (Go) is the canonical role-name source. Story 1.13 extends each with the label and badge-token mapping per AC #4 — **do not** add a parallel mapping table elsewhere.
- The `_components.css` partial in `fieldmark_shared/src/` is the right home for new component CSS — Story 1.12 established it.
- The HTML normalizer and snapshot-test harness on each stack are shared infrastructure; reuse them.

### Git Intelligence

Recent commits (most relevant to this story):
- `ac376d7 feat: e1s5 cross stack layout` — established the chrome landmark structure this story extends.
- `d0b9577 feat: e1s4 bootstrap design system` — established the semantic color tokens (`--color-success`, `--color-danger`, etc.) that the badge tokens consume.
- `d03f0fe feat: e1s3 establish tools parity` — `make parity` is the AC #9 gate.

Branch state: `feature/1.6_theme-toggle` has uncommitted work for Story 1.6 (the in-progress story). Story 1.13 belongs on its own branch (`feature/1.13_role-aware-home-page` or per the project's branch convention) — do not start it on top of 1.6's branch. Start it from `main` after 1.6 and the intermediate stories 1.7–1.12 have merged.

Commit convention from git history: `feat: :sparkles: e1s{N} {description}` — use `feat: :sparkles: e1s13 role-aware home page across three stacks`.

### Latest Technical Information

- **.NET 10 / EF Core 10.0.7** in use. `ClaimsPrincipal` and the `User` accessor on `PageModel` are stable framework features; no version concerns. Razor partials with strongly-typed VMs (`<partial name="..." model="@vm" />`) are the canonical render shape per `.NET CLAUDE.md`.
- **Django 6.0.4 / Python 3.14+** in use. `pytest-django` provides `client.force_login(user)` for the integration tests. The template-tag library pattern (`templatetags/avatar.py` + `{% load avatar %}`) is the canonical way to put a small Python helper behind a `{{ value|filter }}` or `{% tag arg %}` template invocation.
- **Go 1.26.2 / Fiber v3.2.0 / pgx v5.9.2** in use. `c.Locals("user").(*app.Actor)` is the established context-access pattern from Story 1.9.
- **Basecoat 0.3.11** is the pinned design-system version. `.badge` is a Basecoat class; the `badge-danger` / `-info` / `-warning` / `-neutral` / `-success` modifiers are FieldMark extensions added in this story.
- **axe-core** — most-recent stable on npm at story-author time is `axe-core@4.10.x` (under `@axe-core/cli@4.10.x`). Use the default WCAG 2.1 AA tags. If a future Basecoat upgrade introduces a `<dialog>` or other pattern with new ARIA expectations, axe will catch it — keep the rule set at default.
- **Avatar/initials Unicode handling:** the helper preserves input characters as-is (no transliteration to ASCII). A user named "Ää Öö" gets initials "ÄÖ"; a user named "李 明" gets "李明" (or "李李" if single-token in Chinese — the algorithm operates on Unicode whitespace tokens). This is per AC #3 — document it in the helper's tests.

### References

- [Epic 1 Story 1.13](_bmad-output/planning-artifacts/epics.md) — AC source (epics.md is canonical per the workflow).
- [Architecture — Repository Directory Structure](_bmad-output/planning-artifacts/architecture.md#complete-repository-directory-structure) — file location alignment.
- [Architecture — Frontend Architecture](_bmad-output/planning-artifacts/architecture.md#frontend-architecture) — JS budget, no-client-state rule, HTMX target IDs.
- [PRD FR4](_bmad-output/planning-artifacts/prd/functional-requirements.md) — unauthenticated redirect to framework-local login.
- [PRD FR58, FR59](_bmad-output/planning-artifacts/prd/functional-requirements.md) — cross-stack identical routes, methods, observable behavior.
- [PRD FR60–FR63](_bmad-output/planning-artifacts/prd/functional-requirements.md) — accessibility (keyboard, ARIA, focus, aria-live).
- [UX spec — StatusBadge](_bmad-output/planning-artifacts/ux-design-specification.md) — color tokens, "text + color always paired", role badge surface.
- [UX spec — Avatar](_bmad-output/planning-artifacts/ux-design-specification.md) — header avatar, initials fallback.
- [UX spec — ThemeToggle](_bmad-output/planning-artifacts/ux-design-specification.md) — header placement beside avatar (Story 1.6 implemented; 1.13 wraps it in the right-cluster).
- [UX spec — Skip-link / focus / heading hierarchy (Step 8)](_bmad-output/planning-artifacts/ux-design-specification.md) — UX-DR33, UX-DR35, UX-DR39 source surfaces (Story 1.5 implemented; 1.13 first content page exercising them).
- [docs/hard-rules.md](docs/hard-rules.md) — stack symmetry, backend authority, no client state.
- [Story 1.5 implementation artifact](_bmad-output/implementation-artifacts/1-5-implement-cross-stack-base-layout-with-skip-link-landmarks-and-flashregion.md) — chrome structure this story extends.
- [Story 1.6 implementation artifact](_bmad-output/implementation-artifacts/1-6-implement-themetoggle-with-cookie-persistence-per-stack.md) — ThemeToggle markup wrapped by this story's chrome edits.
- [Story 1.10 implementation artifact](_bmad-output/implementation-artifacts/1-10-author-shared-uuid-dev-user-manifest-and-per-stack-idempotent-seed-runners.md) — dev users for parity snapshot fixture.
- [Story 1.11 implementation artifact](_bmad-output/implementation-artifacts/1-11-login-logout-and-unauthenticated-redirect-across-all-three-stacks.md) — `/login`, `/logout`, redirect contract, test sign-in helpers.
- [Story 1.12 implementation artifact](_bmad-output/implementation-artifacts/1-12-implement-authz-can-primitive-and-actionbutton-trichotomy-helper-per-stack.md) — `Role` value object extended by this story; `_components.css` precedent.
- [FieldMark/CLAUDE.md](FieldMark/CLAUDE.md), [fieldmark_py/CLAUDE.md](fieldmark_py/CLAUDE.md), [fieldmark-go/CLAUDE.md](fieldmark-go/CLAUDE.md), [fieldmark_shared/CLAUDE.md](fieldmark_shared/CLAUDE.md) — per-stack architectural rules.

## Dev Agent Record

### Agent Model Used

_(populated by dev agent)_

### Debug Log References

### Completion Notes List

- Ultimate context engine analysis completed — comprehensive developer guide created.

### File List
