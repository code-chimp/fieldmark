# Story 1.11: Login, logout, and unauthenticated-redirect across all three stacks

Status: ready-for-dev

## Story

As any FieldMark user,
I want to log in with my username and password on .NET and Django, pick my actor on Go, and log out cleanly on every stack,
So that the application identifies me on every request, redirects me to `/login` when I am not authenticated, and the cross-stack parity contract holds for the new `/login` and `/logout` routes.

## Acceptance Criteria

1. **`/login` (GET) renders an identically-structured login surface on every stack.**
   - **.NET (`/login`)** renders a username + password form built from Basecoat input components inside the existing `_Layout.cshtml` chrome (skip-link, landmarks, FlashRegion, ThemeToggle — Stories 1.5 / 1.6).
   - **Django (`/login`)** renders the same form with byte-identical markup (Basecoat classes, label associations, error-region wiring). Snapshot parity asserted by a unit test (see AC #10).
   - **Go (`/login`)** renders a user-switcher: the list of seeded dev users from `fiber_auth.users` styled as Basecoat buttons, plus a labelled banner explaining this is a development stub per ADR-012. Each button submits via `POST /login` and carries the user's `username` as the form value. The page intentionally diverges from the .NET / Django form (a credential form would be a lie — Go has no password storage; ADR-012 / Story 1.10).
   - All three pages place exactly one `<h1>` ("Sign in to FieldMark"), use `<main id="main-content">`, render the FlashRegion in chrome, render the ThemeToggle in chrome, and pass `@axe-core/playwright` with zero violations.

2. **Unauthenticated requests to any business route redirect to `/login` with HTTP 302 (FR4).**
   - On .NET, the cookie-authentication middleware's `LoginPath` is set to `/login`. Any `[Authorize]`-protected page (or page protected by the application's fallback auth policy) responds with `302 Location: /login?ReturnUrl=<encoded-original>` when the request is unauthenticated.
   - On Django, every business view is protected. The chosen mechanism is `settings.LOGIN_URL = "/login"` + an application-wide `LoginRequiredMiddleware` (Django 5.2+ shipped class; also acceptable: a project-wide custom middleware that wraps `login_required` around every view) that excludes the `/login` and `/logout` paths and the static prefix. Unauthenticated requests respond `302 Location: /login?next=<encoded-original>`.
   - On Go, `auth.RequireAuth()` (already authored in Story 1.9) is mounted on every existing business route. Currently public routes (`/`, `/privacy`, `/fragments/compliance-tile`) become authenticated routes; the only un-protected paths in this story are `/login` (GET, POST), `/logout` (POST), `/preferences/theme` (POST — Story 1.6), and `/static/*`.

3. **`POST /login` authenticates and redirects.**
   - **.NET (`POST /login`)**: handler binds `username` + `password` + optional `returnUrl`. Calls `SignInManager.PasswordSignInAsync(username, password, isPersistent: true, lockoutOnFailure: false)`. On success: redirect 302 to `returnUrl` if it is a local URL, else `/`. On failure: re-render `_Login.cshtml` with HTTP **422**, `aria-invalid="true"` + `aria-describedby` on the offending fields, and a top `InlineAlert` with `role="alert"` containing the error count and a link (`<a href="#field-username">`) to the first invalid field (FR55a, FR61, UX-DR34). No state mutated on 422 (no session cookie set).
   - **Django (`POST /login`)**: handler binds `username` + `password` + optional `next`. Uses `django.contrib.auth.authenticate(request, username=..., password=...)`, then `django.contrib.auth.login(request, user)` on success — never bypassing the framework's password hashing. Same 422 + `aria-invalid` + `aria-describedby` + top `InlineAlert` shape on failure. Redirect on success: `next` if it is a safe URL (`django.utils.http.url_has_allowed_host_and_scheme(next, allowed_hosts={request.get_host()})`), else `/`. CSRF token required on form submission (Django's default; the form template carries `{% csrf_token %}`).
   - **Go (`POST /login`)**: handler binds form field `username`. Looks up the username in `fiber_auth.users`; if absent, re-renders the user-switcher with an `InlineAlert` (`role="alert"`, "Unknown user — pick from the list") at HTTP **422**. If present, sets `Set-Cookie: X-FieldMark-Actor=<username>; Path=/; SameSite=Lax; Max-Age=31536000` (no `HttpOnly` — the cookie is dev-only, debug-friendly, and is not a credential). Redirects 302 to `/`. No password field is read (and none is rendered on the form per AC #1).

4. **`POST /logout` terminates the session and redirects to `/login` (FR3, FR54).**
   - **.NET (`POST /logout`)**: handler calls `SignInManager.SignOutAsync()` and returns `LocalRedirect("/login")`. GET is **not** registered for this path. `[ValidateAntiForgeryToken]` (or the framework default) is honoured.
   - **Django (`POST /logout`)**: handler is `@require_POST`-decorated; calls `django.contrib.auth.logout(request)` and returns `redirect("/login")`. GET is **not** registered. CSRF token required (no `@csrf_exempt`).
   - **Go (`POST /logout`)**: handler clears the actor cookie (`c.Cookie(&fiber.Cookie{Name: "X-FieldMark-Actor", Value: "", Path: "/", MaxAge: -1, SameSite: "lax"})`) and returns 302 to `/login`. GET is not registered.
   - In all three stacks: a subsequent request to any business route after logout 302-redirects to `/login` (verified by the integration test in Task 9 on .NET and Task 10 on Django; on Go, by an integration-tagged Go test or a manual smoke step in Task 11).

5. **An authenticated request can resolve the actor's UUID and conceptual role on every stack (FR2).**
   - **.NET**: a `ClaimsPrincipal` extension method `User.GetActorId() : Guid` returns the user's UUID by parsing `User.FindFirstValue(ClaimTypes.NameIdentifier)`. A second extension `User.GetConceptualRoles() : IReadOnlyList<string>` returns the role claims (`ClaimTypes.Role`) — already populated by Identity's default cookie-claims pipeline because Story 1.7 wired `AddRoles<IdentityRole<Guid>>()` and Story 1.10's `DevUsersSeeder` calls `UserManager.AddToRoleAsync`. Both helpers live in `FieldMark.Web/Authentication/ClaimsPrincipalExtensions.cs` and have unit-test coverage with a hand-built `ClaimsPrincipal` (Task 8).
   - **Django**: `request.user.is_authenticated` and `request.user.groups.values_list("name", flat=True)` return the conceptual roles. A new helper `fieldmark.authn.current_actor(request) -> CurrentActor` (a small `dataclasses.dataclass` carrying `id: UUID`, `username: str`, `roles: list[str]`) reads the side-table UUID (`request.user.dev_uuid.uuid` via the `related_name="dev_uuid"` introduced in Story 1.10) and the group names. Lives in `fieldmark_py/fieldmark/authn.py`.
   - **Go**: `auth.ActorFromCtx(c)` (already authored in Story 1.9) returns `*app.Actor` with `ID uuid.UUID`, `Username string`, `Role string`. **No new helper is added in this story** — the contract from 1.9 is the contract.

6. **Unauthorized direct POST to a privileged route returns HTTP 403 without leaking entity state (FR7, FR56).**
   - Because Epic 1 has zero business state-change handlers, this AC is satisfied by **a single guarded test endpoint per stack** registered for the duration of an integration test only — **not** by adding a real, persisted route. Specifically: each stack's integration test suite registers a temporary route `POST /__authz_probe` that requires the conceptual role `ADMIN` via the stack's idiomatic mechanism (.NET `[Authorize(Roles = "ADMIN")]`, Django `@user_passes_test(lambda u: u.groups.filter(name="ADMIN").exists())`, Go `authz.RequireRole("ADMIN")` factory added to `internal/web/auth/`). The test logs in as `diego` (`SITE_SUPERVISOR`), POSTs to the probe, and asserts response status is exactly 403 and response body does **not** contain the canonical state strings `"Active"`, `"OnHold"`, `"Closed"`, `"InProgress"`, `"Open"`, `"Resolved"`, `"Voided"` (a regex assertion). The probe route is **not** registered in production code paths and is gated behind a test-only configuration flag — confirmed by `make parity` (AC #11) seeing zero `/__authz_probe` entries in the dumped route inventories.

7. **The dump-routes output for each stack lists exactly these new HTTP method/path pairs (in addition to the prior inventory plus Story 1.6's `post /preferences/theme`):**
   - `get  /login`
   - `post /login`
   - `post /logout`

   All three stacks emit the same three lines (modulo per-stack route dumper formatting that was unified in Story 1.6 — every dumper now emits actual HTTP methods). No `/logout` GET line; no `/auth/*` routes; no `/identity/*` routes from .NET (`AddIdentityCore` — not `AddDefaultIdentity` — was deliberately chosen in Story 1.7 to keep that off the inventory).

8. **`make parity` exits 0.** After this story lands:
   - Route inventory diff across all three stacks: zero. The two new lines above appear identically on every stack.
   - `pg_indexes` snapshot for the `domain` schema (`tools/parity/canonical-pg-indexes.txt`): unchanged. Story 1.11 touches **zero** DDL — it consumes the auth schemas seeded by Stories 1.7 / 1.8 / 1.9 / 1.10 and writes session/cookie state only.

9. **Build / type / lint / test gates stay green on each stack.**
   - **.NET:** `cd FieldMark && dotnet build && dotnet test` — clean. `dotnet csharpier format .` reports zero diffs. `TreatWarningsAsErrors=true` honoured.
   - **Django:** `cd fieldmark_py && uv run ruff check . && uv run mypy . && uv run pytest` — clean.
   - **Go:** `cd fieldmark-go && make fmt-check && make vet && make staticcheck && make test` — clean.

10. **A cross-stack snapshot test asserts byte-identical login-form markup on .NET and Django.** A new test fixture in `e2e/` (or, if `e2e/` does not yet contain Playwright setup at this branch's HEAD, a per-stack unit-level snapshot test colocated with each stack's template tests) captures the rendered HTML of `GET /login` on .NET and Django, normalises whitespace and attribute order using the same pipeline established in Story 1.5 / Story 1.6 (`tests/normalize_html.py` on Django, `FieldMark.Tests.Integration/Helpers/NormaliseHtml.cs` on .NET, both ports of the same algorithm), and asserts byte-identical output for: (a) the form's `<form>...</form>` block, (b) the inline error region block. The Go stack is **excluded** from this assertion because its `/login` is intentionally a different surface (user-switcher list, no password input).

11. **Each stack's `CLAUDE.md` Authentication section is updated to reflect login wiring.** The previously-deferred Story 1.11 step is now done:
    - `FieldMark/CLAUDE.md` documents `app.UseAuthentication()` + `app.UseAuthorization()` are wired; the `/login` Razor Page lives at `Pages/Account/Login.cshtml(.cs)`; the cookie authentication scheme's `LoginPath = "/login"` is configured; the fallback authorization policy requires authenticated users on every page except `Pages/Account/Login.cshtml` and `Pages/Account/Logout.cshtml`.
    - `fieldmark_py/CLAUDE.md` documents the `LoginRequiredMiddleware` placement, the excluded paths, and the `fieldmark.authn.current_actor()` helper.
    - `fieldmark-go/CLAUDE.md` documents that `auth.RequireAuth()` is now mounted on every business route (so existing handlers like the root index now read a guaranteed-non-nil authenticated actor via `auth.ActorFromCtx`), and that `/login` renders a user-switcher per ADR-012.

12. **Root `README.md` Getting Started gains a `make seed` reminder before first login.** The login flow on .NET and Django depends on the dev-user manifest having been seeded into each stack's auth tables (Story 1.10). The README's existing "after `make up`" sequence gains one line: "Run `make seed` once to populate dev users — the login page on .NET / Django will only accept these credentials, and the Go user-switcher only lists seeded users."

## Tasks / Subtasks

- [ ] Task 1: Read all upstream story artifacts and confirm dependency posture (AC: all)
  - [ ] 1.1 Read `_bmad-output/implementation-artifacts/1-5-implement-cross-stack-base-layout-with-skip-link-landmarks-and-flashregion.md` — note the FlashRegion shape (`<div id="flash-region" role="status" aria-live="polite" aria-atomic="false">`), the `flash_messages()` helper signature per stack, and the layout file paths (`_Layout.cshtml`, `templates/base.html`, `internal/web/templates/layouts/base.html`).
  - [ ] 1.2 Read `_bmad-output/implementation-artifacts/1-6-implement-themetoggle-with-cookie-persistence-per-stack.md` — note the `dump_routes` change (Tasks 3.4 / 4.7) that makes each dumper emit actual HTTP methods. This story relies on that change; if 1.6 is not merged yet when 1.11 runs, surface a clear dependency stop.
  - [ ] 1.3 Read `_bmad-output/implementation-artifacts/1-7-wire-asp-net-core-identity-to-dotnet-auth-schema-with-conceptual-roles.md` — note: `AddIdentityCore` (not `AddDefaultIdentity`) was used; `AddSignInManager` is registered; `app.UseAuthentication()` was deliberately **not** added (left for this story); `Pages/Account/*` was deliberately **not** scaffolded (left for this story).
  - [ ] 1.4 Read `_bmad-output/implementation-artifacts/1-8-wire-django-built-in-auth-to-django-auth-schema-with-conceptual-role-groups.md` — note: `django.contrib.auth` is installed; `MIDDLEWARE` includes the default auth/session middleware; no `LoginRequiredMiddleware` yet; no `urls.py` entries for `/login` or `/logout`; CSRF middleware is active.
  - [ ] 1.5 Read `_bmad-output/implementation-artifacts/1-9-implement-go-fiber-stub-authentication-middleware.md` — note: `StubAuthMiddleware` hydrates an `*app.Actor` on every request; `RequireAuth` exists but is not yet mounted; `ActorFromCtx` reads the actor; the cookie name is `X-FieldMark-Actor`; `fiber_auth.users` / `fiber_auth.user_roles` are the lookup tables.
  - [ ] 1.6 Read `_bmad-output/implementation-artifacts/1-10-author-shared-uuid-dev-user-manifest-and-per-stack-idempotent-seed-runners.md` — note: six seeded users (`marisol`, `diego`, `aisha`, `ravi`, `kenji`, `testuser`); shared password is dev-only (e.g., `FieldMark!2026`); on Django the canonical UUID lives in the `django_auth.dev_user_uuid` side table accessible via `request.user.dev_uuid.uuid`; the Go seeder writes `fiber_auth.users` rows but **does not** persist passwords.
  - [ ] 1.7 If **any** of Stories 1.5, 1.6, 1.7, 1.8, 1.9, 1.10 has not merged into the working branch when this story begins, surface the dependency and pause. Story 1.11 cannot land standalone — login forms need the base layout (1.5) and ThemeToggle (1.6) in chrome, Identity wiring (1.7) on .NET, search-path + Groups (1.8) on Django, stub middleware (1.9) on Go, and seeded dev users (1.10) on every stack to be useful.

- [ ] Task 2: Author the canonical login-form markup contract in `fieldmark_shared/components/login-form.example.html` (AC: #1, #10)
  - [ ] 2.1 Create `fieldmark_shared/components/login-form.example.html` (this is the canonical reference that .NET and Django snapshot tests assert against, modeled on Architecture line 784 — "fieldmark_shared/components/action_button.example.html" — and the "canonical component example gallery" pattern named in Epic 7). It contains the full `<form>...</form>` block (the form itself only — outer chrome is the per-stack layout):

    ```html
    <form method="post" action="/login" class="space-y-4" id="login-form" novalidate>
      <input type="hidden" name="return_url" value="" />
      <div class="form-field">
        <label for="field-username" class="label">Username</label>
        <input
          type="text"
          id="field-username"
          name="username"
          class="input"
          autocomplete="username"
          autocapitalize="off"
          autocorrect="off"
          spellcheck="false"
          required
        />
        <p id="field-username-error" class="form-error" hidden></p>
      </div>
      <div class="form-field">
        <label for="field-password" class="label">Password</label>
        <input
          type="password"
          id="field-password"
          name="password"
          class="input"
          autocomplete="current-password"
          required
        />
        <p id="field-password-error" class="form-error" hidden></p>
      </div>
      <button type="submit" class="btn btn-primary w-full">Sign in</button>
    </form>
    ```

  - [ ] 2.2 Create `fieldmark_shared/components/login-error-region.example.html` (the InlineAlert block, rendered above the form on 422):

    ```html
    <div role="alert" class="alert alert-danger" id="login-errors">
      <p class="alert-title">Sign in failed — please correct the highlighted fields.</p>
      <p>
        <a href="#field-username" class="alert-link">Go to first invalid field</a>
      </p>
    </div>
    ```

  - [ ] 2.3 Document both files in `fieldmark_shared/CLAUDE.md` under a new `## Component Examples` section: their purpose (canonical reference for cross-stack form snapshots), the snapshot-test pipeline (whitespace normalisation, attribute sorting), and the rule that any change must be applied to .NET and Django **simultaneously** (and the Go user-switcher is exempt — see AC #1).
  - [ ] 2.4 Do **not** add CSS in this story to support these classes if Story 1.4's design system already provides `.form-field`, `.label`, `.input`, `.form-error`, `.btn`, `.btn-primary`, `.alert`, `.alert-danger`, `.alert-title`, `.alert-link`. If any class is missing, add the minimum necessary rules to `fieldmark_shared/src/_components.css` (or the file Story 1.4 established) and rebuild `dist/fieldmark.css`. Verify by grepping the dist file before adding.

- [ ] Task 3: Implement the .NET login page (AC: #1, #3, #11)
  - [ ] 3.1 Create `FieldMark/FieldMark.Web/Pages/Account/Login.cshtml`. Page directive: `@page "/login"` (route override — keeps the URL `/login`, not `/Account/Login`). Content: extends `_Layout.cshtml`, places one `<h1>Sign in to FieldMark</h1>` inside `<main id="main-content">`, then renders the canonical form markup from `fieldmark_shared/components/login-form.example.html`. Hidden `return_url` value bound to `Model.ReturnUrl`. Top `InlineAlert` block (from `login-error-region.example.html`) is conditionally rendered when `Model.HasErrors` is true.

    Important: **do not** use ASP.NET Core's `<form asp-page="...">` tag helper for this form. The tag helper inserts an antiforgery token but it also injects framework-controlled attributes (`asp-route-*`) that would diverge from the canonical markup. Instead, write a plain `<form method="post" action="/login">` and render the antiforgery token manually via `@Html.AntiForgeryToken()` placed **before** the `<input type="hidden" name="return_url" ...>` line. Verify the rendered HTML matches the canonical example bit-for-bit save for the appended `<input name="__RequestVerificationToken" ...>` tag, which is excluded from the snapshot test (Task 12.4 lists the exclusion).
  - [ ] 3.2 Create `FieldMark/FieldMark.Web/Pages/Account/Login.cshtml.cs`. Class `LoginModel : PageModel`. Inject `SignInManager<IdentityUser<Guid>> signInManager` and `UserManager<IdentityUser<Guid>> userManager` via constructor.
    - `[BindProperty] public string Username { get; set; } = "";`
    - `[BindProperty] public string Password { get; set; } = "";`
    - `[BindProperty(SupportsGet = true)] public string? ReturnUrl { get; set; }`
    - `public bool HasErrors => !ModelState.IsValid;`
    - `public IDictionary<string, string?> FieldErrors { get; } = new Dictionary<string, string?>();` — populated on validation failure (key = field id `"field-username"` / `"field-password"`, value = error message). The Razor page reads this dict to attach `aria-invalid` and to fill the `#field-<name>-error` `<p>` elements.
  - [ ] 3.3 Implement `OnGet()`: `if (User.Identity?.IsAuthenticated == true) return LocalRedirect(ReturnUrl ?? "/"); return Page();`. The early-redirect path means an already-authenticated user hitting `/login` does not see the form.
  - [ ] 3.4 Implement `OnPostAsync()`:
    - If `string.IsNullOrWhiteSpace(Username)` add field error to `FieldErrors["field-username"]` ("Username is required.") and to `ModelState`. Same for `Password`. If any field error: return `Page()` with `Response.StatusCode = 422` set before return.
    - Otherwise call `var result = await _signInManager.PasswordSignInAsync(Username, Password, isPersistent: true, lockoutOnFailure: false);`
    - If `result.Succeeded`: `if (Url.IsLocalUrl(ReturnUrl)) return LocalRedirect(ReturnUrl); return LocalRedirect("/");` (LocalUrl check is the canonical defence against open-redirect — never trust the form's `return_url`).
    - If `!result.Succeeded`: add a generic non-field-bound error (`ModelState.AddModelError(string.Empty, "Invalid username or password.")`) and **also** mark `field-username` and `field-password` as invalid for the inline-alert link target (`FieldErrors["field-username"] = ""`). Set `Response.StatusCode = 422` and return `Page()`. Critical: do **not** disclose which of (username, password) was wrong — the generic message is the only safe disclosure (defence-in-depth even for a dev artifact; aligns with the project's NFR2 server-authoritative posture).
  - [ ] 3.5 Update `FieldMark/FieldMark.Web/Pages/Shared/_Layout.cshtml` (or wherever the body of the layout renders chrome) only if needed: ensure the `<a href="/login">Sign in</a>` link is conditional on `!User.Identity?.IsAuthenticated == true`, and conditionally renders an inline `<form method="post" action="/logout">` containing the antiforgery token and a `<button type="submit" class="btn btn-link">Sign out</button>` when authenticated. This is the only safe way to expose logout in a header without violating FR54 (POST-only state changes). Visually placed beside the ThemeToggle (Story 1.6 placement convention).
  - [ ] 3.6 Create `FieldMark/FieldMark.Web/Pages/Account/Logout.cshtml.cs` (no `.cshtml` view file needed; this is a handler-only page). Page directive on a one-line file (`Logout.cshtml`): `@page "/logout"`. The `.cs` file defines `LogoutModel : PageModel` with **only** `OnPostAsync()` — no `OnGet`. The handler: `await _signInManager.SignOutAsync(); return LocalRedirect("/login");`. Inject `SignInManager<IdentityUser<Guid>>` via constructor. Apply `[ValidateAntiForgeryToken]` (or rely on the framework default; in .NET 10's Razor Pages, antiforgery validation is automatic for POST — confirm by checking the Story 1.7 / 1.10 `Program.cs` doesn't disable it).

- [ ] Task 4: Wire .NET authentication into the request pipeline and configure cookie scheme (AC: #2, #11)
  - [ ] 4.1 In `FieldMark/FieldMark.Web/Program.cs`, **after** the `AddIdentityCore<IdentityUser<Guid>>(...)...AddSignInManager().AddDefaultTokenProviders()` chain (from Story 1.7) and **before** `var app = builder.Build();`, add:

    ```csharp
    builder.Services
        .AddAuthentication(IdentityConstants.ApplicationScheme)
        .AddCookie(IdentityConstants.ApplicationScheme, options =>
        {
            options.LoginPath = "/login";
            options.LogoutPath = "/logout";
            options.AccessDeniedPath = "/login";
            options.ExpireTimeSpan = TimeSpan.FromDays(14);
            options.SlidingExpiration = true;
            options.Cookie.SameSite = SameSiteMode.Lax;
            options.Cookie.SecurePolicy = CookieSecurePolicy.SameAsRequest;
            options.Cookie.HttpOnly = true;
        });

    builder.Services.AddAuthorization(options =>
    {
        options.FallbackPolicy = new AuthorizationPolicyBuilder()
            .RequireAuthenticatedUser()
            .Build();
    });

    builder.Services.AddRazorPages(options =>
    {
        options.Conventions.AllowAnonymousToPage("/Account/Login");
        options.Conventions.AllowAnonymousToPage("/Account/Logout");
        // /preferences/theme is callable by anonymous users so the theme works on /login
        options.Conventions.AllowAnonymousToFolder("/Preferences");
    });
    ```

    Notes: `IdentityConstants.ApplicationScheme` is the canonical scheme name Identity's `SignInManager` writes claims to; using it (rather than a hand-named `"FieldMarkCookie"` scheme) means `SignInManager.PasswordSignInAsync` works out of the box. The fallback policy + `AllowAnonymousToPage` exemptions implement AC #2 cleanly without per-page `[Authorize]` attributes — Microsoft's documented "secure by default" pattern.
  - [ ] 4.2 Insert middleware in the request pipeline at exactly this order, **between** `app.UseStaticFiles()` and `app.UseRouting()` is **wrong** — the correct order is:

    ```csharp
    app.UseHttpsRedirection(); // already present
    app.UseStaticFiles();      // already present
    app.UseRouting();          // already present
    app.UseAuthentication();   // NEW — must come after UseRouting, before UseAuthorization
    app.UseAuthorization();    // NEW — must come after UseAuthentication
    app.MapRazorPages();       // already present
    ```

    If the existing `Program.cs` does not call `app.UseRouting()` explicitly (Razor Pages with `MapRazorPages` adds it implicitly), insert `app.UseRouting()` explicitly before `app.UseAuthentication()`. The .NET documentation is emphatic that `UseAuthentication` must precede `UseAuthorization` and both must follow `UseRouting`.
  - [ ] 4.3 Confirm the `--dump-routes` early-return path (Story 1.3, preserved through 1.7 and 1.10) still runs **before** `app.Run()` and **does not** require a live DB. Auth middleware is registered in `Services`; it does not connect to the DB at startup. Run `dotnet run --project FieldMark.Web -- --dump-routes` and confirm it exits 0 without `make up`.

- [ ] Task 5: Implement the Django login view, logout view, URLs, and login-required middleware (AC: #1, #3, #4, #11)
  - [ ] 5.1 Create `fieldmark_py/fieldmark/views.py` if it does not exist (Story 1.6 may already have created it for the theme endpoint). Add two new view functions:

    ```python
    from dataclasses import dataclass
    from typing import Any
    from django.contrib.auth import authenticate, login, logout
    from django.http import HttpRequest, HttpResponse
    from django.shortcuts import redirect, render
    from django.utils.http import url_has_allowed_host_and_scheme
    from django.views.decorators.http import require_http_methods, require_POST


    @dataclass
    class LoginFieldError:
        username: str | None = None
        password: str | None = None
        general: str | None = None

        def has_any(self) -> bool:
            return bool(self.username or self.password or self.general)


    @require_http_methods(["GET", "POST"])
    def login_view(request: HttpRequest) -> HttpResponse:
        if request.user.is_authenticated:
            next_url = request.GET.get("next", "/")
            if url_has_allowed_host_and_scheme(next_url, allowed_hosts={request.get_host()}):
                return redirect(next_url)
            return redirect("/")

        errors = LoginFieldError()
        username = ""
        next_url = request.POST.get("next") or request.GET.get("next") or ""

        if request.method == "POST":
            username = (request.POST.get("username") or "").strip()
            password = request.POST.get("password") or ""
            if not username:
                errors.username = "Username is required."
            if not password:
                errors.password = "Password is required."
            if not errors.has_any():
                user = authenticate(request, username=username, password=password)
                if user is None:
                    errors.general = "Invalid username or password."
                    errors.username = ""  # marks the field as invalid for link target
                    errors.password = ""
                else:
                    login(request, user)
                    if url_has_allowed_host_and_scheme(next_url, allowed_hosts={request.get_host()}):
                        return redirect(next_url)
                    return redirect("/")

        status = 422 if (request.method == "POST" and errors.has_any()) else 200
        return render(
            request,
            "_login.html",
            {"errors": errors, "username": username, "next": next_url},
            status=status,
        )


    @require_POST
    def logout_view(request: HttpRequest) -> HttpResponse:
        logout(request)
        return redirect("/login")
    ```

  - [ ] 5.2 Create `fieldmark_py/templates/_login.html`. Extend `base.html`. Place one `<h1>Sign in to FieldMark</h1>` inside `{% block main %}`. Render the canonical form markup from `fieldmark_shared/components/login-form.example.html`. The `{% csrf_token %}` tag goes immediately inside `<form>` **after** the `<input type="hidden" name="return_url" ...>` — Django's tag emits `<input type="hidden" name="csrfmiddlewaretoken" ...>`, which is excluded from the snapshot test (Task 12.4). The `return_url` hidden input's `value="{{ next }}"` Django expression must produce identical markup byte-for-byte to the .NET version on identical inputs (empty `return_url=""` is the empty-case both stacks render).

    Conditionally render the InlineAlert block when `errors.has_any()` is true, using the canonical `fieldmark_shared/components/login-error-region.example.html` shape. Each field's `aria-invalid="true"` and `aria-describedby="field-<name>-error"` and the inline `<p id="field-<name>-error">{{ errors.<name> }}</p>` text (with `hidden` removed) render only when that field's error is non-None.
  - [ ] 5.3 Add URL patterns in `fieldmark_py/fieldmark/urls.py`:

    ```python
    from fieldmark import views
    # ... existing patterns ...
    path("login", views.login_view, name="login"),
    path("logout", views.logout_view, name="logout"),
    ```

    **No trailing slashes** — matches Story 1.6's `path("preferences/theme", ...)` convention. The canonical inventory carries `/login` and `/logout`, not `/login/` and `/logout/`.
  - [ ] 5.4 Add to `fieldmark_py/fieldmark/settings.py`:
    - `LOGIN_URL = "/login"`
    - `LOGIN_REDIRECT_URL = "/"`
    - `LOGOUT_REDIRECT_URL = "/login"`
    - Add `"django.contrib.auth.middleware.LoginRequiredMiddleware"` to `MIDDLEWARE`, **after** `"django.contrib.auth.middleware.AuthenticationMiddleware"`. (Django 5.2+ ships this class. If `pyproject.toml` pins an older Django, write a small functional-equivalent middleware in `fieldmark_py/fieldmark/middleware.py` that wraps every view with `login_required` unless the view has the `login_not_required` decorator from `django.contrib.auth.decorators`. Confirm Django version via `uv run python -c "import django; print(django.__version__)"` — Story 1.8 noted "Django 6.0.4"; `LoginRequiredMiddleware` is available.)
    - Decorate `login_view` and `logout_view` with `@login_not_required` from `django.contrib.auth.decorators` so the middleware does not loop. The theme endpoint must also be reachable when unauthenticated (so the toggle works on `/login`); decorate Story 1.6's `set_theme` view with `@login_not_required` in this story as a one-line follow-up — note this is a tiny modification to an upstream story's file and is the correct place to make it.

- [ ] Task 6: Implement the Go user-switcher login, the logout cookie-clear, and mount `RequireAuth` on business routes (AC: #1, #3, #4, #11)
  - [ ] 6.1 Create `fieldmark-go/internal/web/handlers/auth.go`:

    ```go
    package handlers

    import (
        "errors"
        "log"
        "strings"

        "github.com/gofiber/fiber/v3"
        "github.com/jackc/pgx/v5"
        "github.com/jackc/pgx/v5/pgxpool"

        "github.com/code-chimp/fieldmark-go/internal/web/auth"
    )

    type seededUser struct {
        Username    string
        DisplayName string
        Role        string
    }

    type LoginHandlers struct {
        Pool *pgxpool.Pool
    }

    func (h *LoginHandlers) GetLogin(c fiber.Ctx) error {
        if !auth.ActorFromCtx(c).IsAnonymous() {
            return c.Redirect().Status(fiber.StatusFound).To("/")
        }
        users, err := h.listSeededUsers(c.Context())
        if err != nil {
            log.Printf("login: list users: %v", err)
            // Render the page anyway with an empty list and an inline error;
            // do not 500 — keeps the dev artifact navigable.
            return c.Render("pages/login", fiber.Map{
                "Users":   nil,
                "Error":   "Unable to list users — check the database connection.",
                "Status":  fiber.StatusOK,
                "FmTheme": c.Cookies("fm_theme", "system"),
            })
        }
        return c.Render("pages/login", fiber.Map{
            "Users":   users,
            "Status":  fiber.StatusOK,
            "FmTheme": c.Cookies("fm_theme", "system"),
        })
    }

    func (h *LoginHandlers) PostLogin(c fiber.Ctx) error {
        username := strings.TrimSpace(c.FormValue("username"))
        if username == "" {
            return h.renderInvalid(c, "Username is required.")
        }
        actor, err := h.lookupUser(c.Context(), username)
        if err != nil {
            log.Printf("login: lookup %q: %v", username, err)
            return h.renderInvalid(c, "Internal error — check server logs.")
        }
        if actor == nil {
            return h.renderInvalid(c, "Unknown user — pick from the list.")
        }
        c.Cookie(&fiber.Cookie{
            Name:     auth.CookieName(), // see Task 6.4 — expose const from auth pkg
            Value:    username,
            Path:     "/",
            MaxAge:   31536000,
            SameSite: "Lax",
            HTTPOnly: false, // dev stub; not a credential
        })
        return c.Redirect().Status(fiber.StatusFound).To("/")
    }

    func (h *LoginHandlers) PostLogout(c fiber.Ctx) error {
        c.Cookie(&fiber.Cookie{
            Name:     auth.CookieName(),
            Value:    "",
            Path:     "/",
            MaxAge:   -1,
            SameSite: "Lax",
        })
        return c.Redirect().Status(fiber.StatusFound).To("/login")
    }

    func (h *LoginHandlers) renderInvalid(c fiber.Ctx, message string) error {
        users, _ := h.listSeededUsers(c.Context()) // best-effort
        c.Status(fiber.StatusUnprocessableEntity)
        return c.Render("pages/login", fiber.Map{
            "Users":   users,
            "Error":   message,
            "FmTheme": c.Cookies("fm_theme", "system"),
        })
    }

    func (h *LoginHandlers) listSeededUsers(ctx fiber.Ctx) ([]seededUser, error) {
        const q = `
          SELECT u.username, u.display_name, COALESCE(MIN(r.role), '') AS role
            FROM fiber_auth.users u
            LEFT JOIN fiber_auth.user_roles r ON r.user_id = u.id
        GROUP BY u.username, u.display_name
        ORDER BY u.username
        `
        rows, err := h.Pool.Query(ctx, q)
        if err != nil { return nil, err }
        defer rows.Close()
        var out []seededUser
        for rows.Next() {
            var u seededUser
            if err := rows.Scan(&u.Username, &u.DisplayName, &u.Role); err != nil { return nil, err }
            out = append(out, u)
        }
        return out, rows.Err()
    }

    func (h *LoginHandlers) lookupUser(ctx fiber.Ctx, username string) (*struct{ ID string }, error) {
        const q = `SELECT id::text FROM fiber_auth.users WHERE username = $1`
        var id string
        err := h.Pool.QueryRow(ctx, q, username).Scan(&id)
        if err != nil {
            if errors.Is(err, pgx.ErrNoRows) { return nil, nil }
            return nil, err
        }
        return &struct{ ID string }{ID: id}, nil
    }
    ```

    Notes: this handler reads `fiber_auth.users` directly. That is acceptable per `fieldmark-go/CLAUDE.md`'s "no repository abstractions" rule — Stories 1.9 and 1.10 set the precedent for the auth package owning its own SQL. The `fiber.Ctx`-typed list/lookup helpers use `c.Context()` to derive a `context.Context`; do **not** type-assert the return of `c.Context()` and do not pass `fiber.Ctx` into `internal/data` or `internal/app`.
  - [ ] 6.2 Expose `auth.CookieName() string` in `fieldmark-go/internal/web/auth/stub.go` as a tiny accessor returning the package-private `cookieName` constant — the handler package needs it for cookie writes. Add right below the existing `cookieName`/`headerName`/`envVar` const block:

    ```go
    // CookieName returns the cookie name carrying the resolved actor username.
    // Exposed for the login/logout handlers; do not use elsewhere.
    func CookieName() string { return cookieName }
    ```

  - [ ] 6.3 Create `fieldmark-go/internal/web/templates/pages/login.html`. Extends the base layout via `{{template "base" .}}` (or however Story 1.5 wires the layout — match it exactly). Inside the `main` block, render one `<h1>Sign in to FieldMark</h1>`, then a labelled banner:

    ```html
    <div class="alert alert-info" role="status">
      <p class="alert-title">Development stub</p>
      <p>Real Go authentication is intentionally deferred (ADR-012). Pick a user to sign in as.</p>
    </div>
    ```

    Then conditionally an InlineAlert if `.Error` is set:

    ```html
    {{if .Error}}
    <div role="alert" class="alert alert-danger" id="login-errors">
      <p class="alert-title">{{.Error}}</p>
    </div>
    {{end}}
    ```

    Then the user list, rendered as one `<form method="post" action="/login">` per user — each containing a hidden `<input name="username" value="{{.Username}}">` and a `<button type="submit" class="btn btn-secondary">{{.DisplayName}} <span class="badge">{{.Role}}</span></button>`. Using one form per user avoids per-button JavaScript and keeps the page server-rendered. If `.Users` is empty (e.g., DB unreachable), render a single helpful message: "No seeded users found — run `make seed` first."
  - [ ] 6.4 Update `fieldmark-go/cmd/web/main.go` (the Task 5.2 refactor from Story 1.9 should be in place):
    - Inside `registerRoutes(app, deps)`, register:

      ```go
      h := &handlers.LoginHandlers{Pool: deps.Pool}
      app.Get("/login", h.GetLogin)
      app.Post("/login", h.PostLogin)
      app.Post("/logout", h.PostLogout)
      ```

    - Mount `auth.RequireAuth()` on the existing business route group. The simplest way given Story 1.9's posture (which left the routes globally registered) is to **selectively** apply `RequireAuth()` to each non-public handler:

      ```go
      app.Get("/", auth.RequireAuth(), existingIndexHandler)
      app.Get("/privacy", auth.RequireAuth(), existingPrivacyHandler)
      app.Get("/fragments/compliance-tile", auth.RequireAuth(), existingFragmentHandler)
      ```

      Do **not** mount `RequireAuth()` on `/login`, `/logout`, `/preferences/theme`, or `/static/*`. If the existing index handler signature does not accept a leading middleware, refactor it to be a `fiber.Handler` (it already is; this is a one-line per-route change).
    - The `StubAuthMiddleware` registered at the application level (Story 1.9) **stays** — it hydrates the actor on every request, including on `/login` (so `GetLogin` can check `auth.ActorFromCtx(c).IsAnonymous()` and redirect already-signed-in visitors away). `RequireAuth` is the gate; `StubAuthMiddleware` is the hydrator.
  - [ ] 6.5 Update `fieldmark-go/internal/web/templates/partials/header.html` (or the partial that Story 1.5/1.6 created for the header chrome): if the actor is anonymous, render `<a href="/login" class="btn btn-link">Sign in</a>`; if authenticated, render an inline `<form method="post" action="/logout"><button type="submit" class="btn btn-link">Sign out ({{.Actor.Username}})</button></form>`. The actor is passed into every render via the existing view-model pipeline; if not, add a small `viewModelForLayout(c) Layout` helper in `internal/web/viewmodels/layout.go` that bundles `Actor`, `FmTheme`, and any other chrome-level values into one struct, and is called from every handler before rendering.

- [ ] Task 7: Update the route dumpers to emit the three new routes correctly (AC: #7, #8)
  - [ ] 7.1 .NET — `FieldMark/FieldMark.Web/Tools/DumpRoutes.cs` already emits actual HTTP methods after Story 1.6 (AC #7 of that story). Run `dotnet run --project FieldMark.Web -- --dump-routes` and confirm the new lines (`get /login`, `post /login`, `post /logout`) appear. If they do not, the issue is that Razor Pages auto-discovers handlers (`OnGet` → GET, `OnPost` → POST); confirm `Pages/Account/Login.cshtml` has both `@page "/login"` and `OnGet`/`OnPostAsync` handlers, and `Pages/Account/Logout.cshtml` has `@page "/logout"` and an `OnPostAsync` handler only.
  - [ ] 7.2 Django — `fieldmark_py/tools/management/commands/dump_routes.py` already emits actual HTTP methods after Story 1.6 (per its AC #7). The `@require_http_methods(["GET", "POST"])` decorator on `login_view` and `@require_POST` on `logout_view` should be detected by whatever introspection strategy Story 1.6 chose. Run `uv run python manage.py dump_routes` and confirm. If `login_view` only emits one line (e.g., only `get /login` or only `post /login`), inspect the introspection logic — `require_http_methods` stores the allowed methods on the wrapper's closure cells; Story 1.6's pragmatic-fallback registry might need an entry for `login_view`. If a registry-based fallback is in use, add `("login_view", ["GET", "POST"])` and `("logout_view", ["POST"])` entries.
  - [ ] 7.3 Go — `fieldmark-go/cmd/web/main.go` (and the dump-routes path inside it, if Story 1.9's Task 5.2 refactor is in place) walks `app.GetRoutes(true)` and emits each route's method. After adding the three new routes in Task 6.4, run `go run ./cmd/web -dump-routes` and confirm `get /login`, `post /login`, `post /logout` are emitted.
  - [ ] 7.4 Run `make parity` from repo root. The expected diff is **zero**. If any pairwise diff appears, the most likely root cause is one stack registering a method/path that another stack does not — e.g., .NET emitting `get /logout` because Razor Pages auto-registered an `OnGet` handler on `Pages/Account/Logout.cshtml` (the `.cshtml` file should be empty/header-only, and there must be no `OnGet` method in `Logout.cshtml.cs`); or Django emitting `/login/` with a trailing slash (Task 5.3 forbids this — `APPEND_SLASH=True` is Django's default but the `path("login", ...)` pattern itself decides the canonical form). Diagnose and fix root cause; do **not** edit `tools/parity/diff-routes.sh`.

- [ ] Task 8: Add the `ClaimsPrincipalExtensions` helpers on .NET and the `current_actor` helper on Django (AC: #5)
  - [ ] 8.1 .NET — create `FieldMark/FieldMark.Web/Authentication/ClaimsPrincipalExtensions.cs`:

    ```csharp
    using System.Security.Claims;

    namespace FieldMark.Web.Authentication;

    public static class ClaimsPrincipalExtensions
    {
        public static Guid GetActorId(this ClaimsPrincipal user)
        {
            var raw = user.FindFirstValue(ClaimTypes.NameIdentifier);
            if (string.IsNullOrWhiteSpace(raw) || !Guid.TryParse(raw, out var id))
            {
                throw new InvalidOperationException(
                    "GetActorId called on an unauthenticated or claim-less principal. " +
                    "Guard with User.Identity.IsAuthenticated or use the [Authorize] attribute.");
            }
            return id;
        }

        public static IReadOnlyList<string> GetConceptualRoles(this ClaimsPrincipal user) =>
            user.FindAll(ClaimTypes.Role).Select(c => c.Value).ToList();
    }
    ```

  - [ ] 8.2 .NET — add `FieldMark/FieldMark.Tests.Domain/ClaimsPrincipalExtensionsTests.cs` (note: this is in the Domain test project for now, even though the extension lives in Web — the test is a pure assertion and doesn't pull a host; if the project prefers a Web test project, create `FieldMark.Tests.Web/` and put it there). Tests:
    - `GetActorId_ReturnsGuid_FromNameIdentifierClaim` — build a `ClaimsPrincipal` with `new Claim(ClaimTypes.NameIdentifier, "01923456-7890-7abc-def0-123456789abc")`; assert returned `Guid` matches.
    - `GetActorId_ThrowsWhenClaimMissing` — empty principal; assert `InvalidOperationException`.
    - `GetActorId_ThrowsWhenClaimNotGuid` — `Claim(ClaimTypes.NameIdentifier, "not-a-guid")`; assert exception.
    - `GetConceptualRoles_ReturnsAllRoleClaims` — principal with two `ClaimTypes.Role` claims; assert returned list has both, in claim-insertion order.
  - [ ] 8.3 Django — create `fieldmark_py/fieldmark/authn.py`:

    ```python
    """Per-request actor helpers — read-only view of the authenticated principal."""

    from dataclasses import dataclass
    from uuid import UUID

    from django.http import HttpRequest


    @dataclass(frozen=True)
    class CurrentActor:
        id: UUID
        username: str
        roles: tuple[str, ...]

        @property
        def is_anonymous(self) -> bool:
            return self.username == "anonymous"


    ANONYMOUS = CurrentActor(
        id=UUID("00000000-0000-0000-0000-000000000000"),
        username="anonymous",
        roles=(),
    )


    def current_actor(request: HttpRequest) -> CurrentActor:
        user = request.user
        if not user.is_authenticated:
            return ANONYMOUS
        # dev_uuid is the related_name on DevUserUuid (Story 1.10, tools app).
        try:
            uuid_value = user.dev_uuid.uuid
        except AttributeError as exc:
            raise RuntimeError(
                f"Authenticated user {user.username!r} has no DevUserUuid row. "
                "Run `uv run python manage.py seed_dev_users` to populate the manifest."
            ) from exc
        roles = tuple(user.groups.values_list("name", flat=True))
        return CurrentActor(id=uuid_value, username=user.username, roles=roles)
    ```

  - [ ] 8.4 Django — add tests at `fieldmark_py/fieldmark/tests/test_authn.py` (create the package with `__init__.py` if not present):

    ```python
    """Tests for fieldmark.authn.current_actor."""

    import uuid

    import pytest
    from django.contrib.auth.models import Group, User
    from django.test import RequestFactory

    from fieldmark.authn import ANONYMOUS, current_actor
    from tools.models import DevUserUuid


    @pytest.mark.django_db
    def test_anonymous_request_returns_anonymous():
        request = RequestFactory().get("/")
        # Default RequestFactory().user is AnonymousUser (is_authenticated == False).
        from django.contrib.auth.models import AnonymousUser
        request.user = AnonymousUser()
        assert current_actor(request) == ANONYMOUS


    @pytest.mark.django_db
    def test_authenticated_user_returns_uuid_username_and_roles():
        Group.objects.create(name="ADMIN")
        u = User.objects.create(username="aisha")
        u.groups.set([Group.objects.get(name="ADMIN")])
        canonical = uuid.uuid4()
        DevUserUuid.objects.create(user_id=u.pk, uuid=canonical)

        request = RequestFactory().get("/")
        request.user = u

        actor = current_actor(request)
        assert actor.id == canonical
        assert actor.username == "aisha"
        assert actor.roles == ("ADMIN",)
    ```

  - [ ] 8.5 Go — **no new helper added.** `auth.ActorFromCtx(c)` from Story 1.9 already exposes ID/username/role. Document this as the canonical accessor in `fieldmark-go/CLAUDE.md` (Task 13).

- [ ] Task 9: Add the .NET integration test for unauthenticated redirect, login success, and logout (AC: #2, #3, #4, #6, #9)
  - [ ] 9.1 If `FieldMark.Tests.Integration` does not yet have a Testcontainers fixture set up (Story 1.7 deferred this; the project ships an empty stub), add the minimum: `Fixtures/PostgresFixture.cs` spinning up `postgres:17` with `docker/postgres/init` mounted, plus the auth migration applied. If a fixture exists at HEAD, reuse it.
  - [ ] 9.2 Create `FieldMark.Tests.Integration/AuthFlowTests.cs`. Test methods (use `WebApplicationFactory<Program>` against the Postgres fixture and a seeded dev-user database — call `RoleSeeder.SeedAsync` + `DevUsersSeeder.SeedAsync` in the fixture's `InitializeAsync`):
    - `Get_BusinessRoute_WhileUnauthenticated_Redirects302ToLogin` — `var resp = await client.GetAsync("/");` with `client.AllowAutoRedirect = false`; assert `resp.StatusCode == HttpStatusCode.Found` and `resp.Headers.Location.PathAndQuery.StartsWith("/login")`.
    - `Post_Login_WithValidCredentials_RedirectsToHome` — POST `/login` with form `username=marisol&password=FieldMark!2026`; assert 302 to `/`.
    - `Post_Login_WithInvalidPassword_Returns422AndDoesNotSetCookie` — POST `/login` with `username=marisol&password=wrong`; assert 422; assert response Set-Cookie does **not** contain `IdentityConstants.ApplicationScheme`'s cookie; assert response body contains `id="login-errors"` and `role="alert"`.
    - `Post_Logout_TerminatesSessionAndRedirectsToLogin` — sign in as `marisol`, then POST `/logout`; assert 302 to `/login`; assert subsequent GET `/` 302-redirects to `/login`.
    - `Post_AuthzProbe_AsSiteSupervisor_Returns403WithoutLeakingState` — register a temporary in-test route at `/__authz_probe` requiring `[Authorize(Roles = "ADMIN")]`; sign in as `diego` (SITE_SUPERVISOR); POST `/__authz_probe`; assert 403; assert response body does not contain any of the canonical state-leak strings (use a `Assert.DoesNotContain` over `["Active", "OnHold", "Closed", "InProgress", "Open", "Resolved", "Voided"]`).
  - [ ] 9.3 The probe route registration must happen **only** inside the test fixture (`WebApplicationFactory.ConfigureWebHost`) so it does not appear in production routes. Verify by running `make parity` after the test passes — the probe must not appear.

- [ ] Task 10: Add the Django integration tests (AC: #2, #3, #4, #6, #9)
  - [ ] 10.1 Create `fieldmark_py/fieldmark/tests/test_auth_flow.py`. Use `pytest.mark.django_db` and `django.test.Client`:
    - `test_unauthenticated_request_to_root_redirects_to_login(db, client)` — `resp = client.get("/", follow=False)`; `assert resp.status_code == 302 and resp.url.startswith("/login")`.
    - `test_post_login_with_valid_credentials_redirects_to_home(db, client, seeded_users)` — `seeded_users` is a fixture that runs `call_command("seed_groups")` + `call_command("seed_dev_users")`; `resp = client.post("/login", {"username": "marisol", "password": "FieldMark!2026"})`; `assert resp.status_code == 302 and resp.url == "/"`; `assert "_auth_user_id" in client.session`.
    - `test_post_login_with_invalid_password_returns_422_and_no_session(db, client, seeded_users)` — assert `resp.status_code == 422`; `assert "_auth_user_id" not in client.session`; `assert b'role="alert"' in resp.content`; `assert b'aria-invalid="true"' in resp.content`.
    - `test_post_logout_clears_session_and_redirects_to_login(db, client, seeded_users)` — log in, then `client.post("/logout")`; `assert resp.status_code == 302 and resp.url == "/login"`; subsequent `client.get("/", follow=False)` returns 302 to `/login`.
    - `test_post_authz_probe_as_site_supervisor_returns_403_without_state_leak(db, client, seeded_users)` — uses Django's `urls.urlpatterns` override mechanism (`@override_settings(ROOT_URLCONF=...)` or a test-only URL include) to register `/__authz_probe`; log in as `diego`; POST; assert 403 and the canonical-state-strings absence.
  - [ ] 10.2 Add a conftest fixture at `fieldmark_py/conftest.py` (or extend an existing one) for the `seeded_users` fixture that seeds groups + dev users idempotently.
  - [ ] 10.3 The `csrf_token` on Django's test client is handled automatically when using `Client(enforce_csrf_checks=False)` (the default). Do not pass `enforce_csrf_checks=True` for these tests — the goal is to assert the auth flow, not CSRF specifically (Story 1.8 already verified CSRF middleware is wired).

- [ ] Task 11: Add Go integration-tagged tests for login, logout, and require-auth (AC: #2, #3, #4, #6, #9)
  - [ ] 11.1 Create `fieldmark-go/internal/web/handlers/auth_test.go` (unit, no DB):
    - `TestPostLogin_AnonymousActorOnEmptyUsername_RendersInvalid` — use `fiber.New() + a.Test(req)`; assert 422 in response status.
    - `TestPostLogout_ClearsCookie` — POST `/logout`; assert 302 to `/login`; assert response Set-Cookie has `X-FieldMark-Actor=; Max-Age=0` (or equivalent: Fiber emits `MaxAge=-1` as deletion).
  - [ ] 11.2 Create `fieldmark-go/internal/web/handlers/auth_integration_test.go` (integration-tagged: `//go:build integration`):
    - Spins up a Postgres container via `testcontainers-go`, applies `docker/postgres/init/001_schemas.sql` + `010_domain_tables.sql`, runs `cmd/migrate-fiber-auth` against the test DB, runs `cmd/seed` against the test DB.
    - `TestRequireAuth_UnauthenticatedRequest_Returns302ToLogin` — `GET /` with no cookie; assert 302 to `/login`.
    - `TestPostLogin_WithSeededUsername_SetsCookieAndRedirectsHome` — POST `/login` with `username=marisol`; assert 302 to `/`; assert Set-Cookie contains `X-FieldMark-Actor=marisol`.
    - `TestPostLogin_WithUnknownUsername_Returns422` — POST `/login` with `username=nobody`; assert 422.
    - `TestRequireAuth_AfterLogout_Returns302ToLogin` — log in, POST logout, GET `/`; assert 302.
    - `TestRoleCheck_AsSiteSupervisor_Returns403WithoutStateLeak` — register a probe route requiring ADMIN role (via a new `authz.RequireRole("ADMIN")` middleware factory added in this story); log in as `diego`; POST; assert 403; assert response body has none of the state-leak strings.
  - [ ] 11.3 Add `fieldmark-go/internal/web/auth/authz.go` with a tiny `RequireRole(role string) fiber.Handler` factory:

    ```go
    package auth

    import "github.com/gofiber/fiber/v3"

    func RequireRole(role string) fiber.Handler {
        return func(c fiber.Ctx) error {
            actor := ActorFromCtx(c)
            if actor.IsAnonymous() {
                return c.Redirect().Status(fiber.StatusFound).To("/login")
            }
            if actor.Role != role {
                // 403 without entity-state leakage: empty body, generic status text.
                c.Status(fiber.StatusForbidden)
                return c.SendString("Forbidden.")
            }
            return c.Next()
        }
    }
    ```

    This is a minimal first step. The full `authz.Can` primitive (which understands entity-scope rules) lands in Story 1.12 — `RequireRole` here is the limited form needed only for AC #6 in this story. Do **not** generalise this into `authz.Can` here; the typed `Role` value object and the entity-scope logic belong in 1.12.
  - [ ] 11.4 Confirm the integration build runs (`make test-integration` if defined, else `go test -tags=integration ./internal/web/handlers/...`) and passes.

- [ ] Task 12: Cross-stack snapshot test for login form markup (AC: #10)
  - [ ] 12.1 Pick the snapshot strategy. Two acceptable paths:
    - **(a) Per-stack unit-level snapshot tests** — each stack normalises its rendered `/login` body using a shared `normalize_html` algorithm (the one introduced in Story 1.5 / 1.6) and asserts equality against `fieldmark_shared/components/login-form.example.html`. .NET and Django both run their own snapshot; Go is excluded. **Recommended** because no `e2e/` Playwright harness exists yet at this branch's HEAD (Epic 7).
    - **(b) Playwright snapshot test** — defer until Epic 7. Note here that the canonical reference file exists and is consumed when E2E lands.
  - [ ] 12.2 Choose path (a). On .NET: add `FieldMark.Tests.Integration/LoginFormSnapshotTests.cs`. It GETs `/login` against the test host, parses the response body for the `<form id="login-form">...</form>` block, normalises whitespace/attributes, strips the antiforgery token input, and `Assert.Equal`s against the contents of `fieldmark_shared/components/login-form.example.html` (loaded via `File.ReadAllText` with path resolved through `env.ContentRootPath`).
  - [ ] 12.3 On Django: add `fieldmark_py/fieldmark/tests/test_login_snapshot.py`. Uses `Client().get("/login")`, parses the same form block, strips the `csrfmiddlewaretoken` hidden input, and asserts equality against the file at `BASE_DIR.parent / "fieldmark_shared" / "components" / "login-form.example.html"`.
  - [ ] 12.4 The normalisation must strip exactly these elements before comparing: `<input name="__RequestVerificationToken">` (the .NET antiforgery) and `<input name="csrfmiddlewaretoken">` (the Django CSRF token). Document this exception list in the test source as a comment so a future engineer knows what is in the canonical file vs. what is per-stack noise.

- [ ] Task 13: Update each stack's `CLAUDE.md` (AC: #11)
  - [ ] 13.1 `.NET — FieldMark/CLAUDE.md` — rewrite the `## Authentication` section (left by Story 1.7 with a "login pages added in 1.11" placeholder). New content:
    - `app.UseAuthentication()` and `app.UseAuthorization()` are wired in `Program.cs` between `UseRouting` and `MapRazorPages`. The cookie scheme uses `IdentityConstants.ApplicationScheme`; `LoginPath = "/login"`; default 14-day sliding cookie.
    - Fallback authorization policy: `RequireAuthenticatedUser()`. Anonymous access granted to `/Account/Login`, `/Account/Logout`, and `/Preferences/Theme` only.
    - `/login` is `Pages/Account/Login.cshtml(.cs)`. Uses `SignInManager.PasswordSignInAsync` with `isPersistent: true`. On failure, returns 422 with `aria-invalid` + `aria-describedby` per field and a top `role="alert"` summary linking to the first invalid field (FR55a, UX-DR34).
    - `/logout` is `Pages/Account/Logout.cshtml(.cs)` (POST only). Calls `SignInManager.SignOutAsync()` then `LocalRedirect("/login")`. GET is not registered.
    - Reading the actor: `User.GetActorId()` (UUID from `NameIdentifier` claim) and `User.GetConceptualRoles()` (list of `ClaimTypes.Role` values). Both in `FieldMark.Web.Authentication.ClaimsPrincipalExtensions`.
    - Story 1.12 will introduce the typed `Role` value object and the `authz.Can` primitive — current role checks use string comparison on the claim values.
  - [ ] 13.2 `Django — fieldmark_py/CLAUDE.md` — add to the `## Authentication` section (created by Story 1.8):
    - `LoginRequiredMiddleware` is the third entry in `MIDDLEWARE`. Views that should be reachable while unauthenticated must be decorated with `@login_not_required`. Currently exempt: `login_view`, `logout_view`, `set_theme` (Story 1.6).
    - `/login` is `fieldmark.views.login_view`. Uses `django.contrib.auth.authenticate` + `login`. On failure returns 422 with the same field-error semantics as .NET (snapshot test enforces parity).
    - `/logout` is `fieldmark.views.logout_view`, `@require_POST`. Calls `django.contrib.auth.logout` then redirects to `/login`.
    - Reading the actor: `fieldmark.authn.current_actor(request) -> CurrentActor`. Returns `ANONYMOUS` for unauthenticated requests. UUID is read from `request.user.dev_uuid.uuid` (the `DevUserUuid` side table — Story 1.10).
    - Roles are Django Groups: `user.groups.values_list("name", flat=True)`. The five canonical groups are seeded by `seed_groups` (Story 1.8).
  - [ ] 13.3 `Go — fieldmark-go/CLAUDE.md` — extend the `## Authentication` section (created by Story 1.9):
    - `auth.RequireAuth()` is now mounted on every business route. Currently public-only routes are `/login`, `/logout`, `/preferences/theme`, and `/static/*`.
    - `/login` (GET, POST) renders a user-switcher backed by `fiber_auth.users`. The `POST /login` handler sets the `X-FieldMark-Actor` cookie to the submitted username; `POST /logout` clears it. No password is ever read or stored — ADR-012 stub posture.
    - Reading the actor: `auth.ActorFromCtx(c) -> *app.Actor`. `RequireAuth()` guarantees a non-anonymous actor inside its protected handlers.
    - `auth.RequireRole(role string)` is the minimal role-gate helper (newly added in 1.11). The full `authz.Can(user, action, entity)` primitive lands in Story 1.12.
  - [ ] 13.4 Root `CLAUDE.md` — no changes required if it currently delegates auth specifics to per-stack `CLAUDE.md` files; spot-check that it does.

- [ ] Task 14: Update root README "Getting Started" (AC: #12)
  - [ ] 14.1 In the root `README.md`, find the existing Getting Started block (likely after `make up` and before `make run-net`). Insert one line that the dev should run `make seed` once before first login, with a one-line explanation: ".NET and Django login will only accept seeded users; the Go user-switcher only lists seeded users."
  - [ ] 14.2 Add a small "Login credentials (dev only)" subsection:
    - Username: any of `marisol`, `diego`, `aisha`, `ravi`, `kenji`, `testuser`.
    - Password (`.NET` and `Django`): whatever is committed in `docker/postgres/init/seed-uuids/dev-users.json` (default per Story 1.10 example: `FieldMark!2026`). Note that the password is dev-only.
    - On Go, the user is picked from a list — no password entered.

- [ ] Task 15: Run the full verification suite (AC: #8, #9)
  - [ ] 15.1 From repo root: `make reset && make seed` — clean DB, seed all three stacks.
  - [ ] 15.2 Start each stack in turn (`make run-net`, then Ctrl-C; `make run-django`, then Ctrl-C; `make run-go`, then Ctrl-C) and manually verify:
    - GET `/` redirects 302 to `/login`.
    - GET `/login` renders the form (or user-switcher on Go).
    - POST `/login` with the seeded credentials redirects to `/`.
    - POST `/logout` redirects to `/login`; subsequent GET `/` redirects to `/login`.
  - [ ] 15.3 `make parity` — exits 0. Three new lines (`get /login`, `post /login`, `post /logout`) appear identically on every stack.
  - [ ] 15.4 Run each stack's test suite per AC #9.

## Dev Notes

### Brownfield posture — what exists today (read before writing anything)

State of the three stacks at HEAD of this branch (informed by 1.7 / 1.8 / 1.9 / 1.10 having been written but most still `ready-for-dev`):

- **.NET** — Story 1.7's `AuthDbContext`, `AddIdentityCore<IdentityUser<Guid>>` + roles + sign-in-manager, and `RoleSeeder` are in place. Story 1.10's `DevUsersSeeder` populates `dotnet_auth.users` from the shared manifest. **Story 1.7 deliberately did not add `app.UseAuthentication()` or scaffold `/Identity/Account/*`** — both are this story's job. The `Program.cs` `--dump-routes` early-return (Story 1.3) is sacred; preserve it.
- **Django** — Story 1.8 enabled `django.contrib.auth`, `search_path=django_auth,public`, and seeded the five canonical Groups. Story 1.10 added the `DevUserUuid` side-table model and the `seed_dev_users` command. `LoginRequiredMiddleware` and `LOGIN_URL` are not yet configured. CSRF middleware is active by default.
- **Go** — Story 1.9 wired the `StubAuthMiddleware` (cookie → header → env → anonymous resolution) on every request, exposed `auth.RequireAuth()` (but did not mount it), and added `cmd/migrate-fiber-auth` for `fiber_auth.*` schema. Story 1.10 seeded `fiber_auth.users` and `fiber_auth.user_roles` from the shared manifest. The Go stack has **no** password storage and **never** will at MVP — Story 1.11's `/login` here is a stub user-switcher per ADR-012.
- **Shared** — Story 1.5's base layout / FlashRegion / skip-link landmarks are in every stack's chrome. Story 1.6's ThemeToggle is wired with the `fm_theme` cookie endpoint at `POST /preferences/theme`. Story 1.6's `dump_routes` updates made every dumper method-aware; this story relies on that.
- **Stories 1.5–1.10 are `ready-for-dev`** — they have not all merged at the time this story was written. Story 1.11 cannot land standalone. Task 1 makes the dependency check explicit; in practice the dev should land 1.5 → 1.6 → 1.7 → 1.8 → 1.9 → 1.10 → 1.11 in order (with 1.4 being upstream of 1.5 for the design system).

### Why `app.UseAuthentication()` lives in 1.11, not 1.7

Story 1.7 wired Identity services into DI and applied the EF migration. It deliberately did **not** call `app.UseAuthentication()` because:

1. Adding the middleware before any `/login` page exists changes request-pipeline behavior without an observable consumer — premature wiring.
2. The cookie scheme's `LoginPath`, `AccessDeniedPath`, etc. are this story's design decisions (where unauthenticated redirects go, how long the cookie lives). Cramming them into 1.7 would have leaked 1.11 concerns into the schema-wiring story.
3. `app.UseAuthorization()` with a fallback `RequireAuthenticatedUser()` policy actually breaks the current `/` and `/privacy` pages on .NET — those are anonymous-allowed today. Doing it in 1.7 would have either left them unprotected (defeating the policy) or required `[AllowAnonymous]` everywhere (premature). 1.11 is where the protection comes online cleanly.

### Why `LoginRequiredMiddleware` (Django) instead of per-view `@login_required`

Django offers four shapes for "every view requires auth":

| Approach | Pros | Cons |
|---|---|---|
| `@login_required` decorator on every view | Explicit, per-view control | Easy to forget on a new view; ergonomically poor |
| `LoginRequiredMixin` on class-based views | Explicit | Useless if any view is function-based; FieldMark uses function-based views |
| `LoginRequiredMiddleware` (Django 5.2+) | Implicit, secure-by-default | Requires `@login_not_required` on exemptions |
| Custom middleware wrapping `login_required` | Same effect | Reinvents what Django 5.2 ships |

The middleware approach is chosen because: (a) it is "secure by default" — adding a new view doesn't accidentally expose it; (b) the exempted-paths list is small and explicit (`login`, `logout`, `set_theme`); (c) Django 6.0.4 (already pinned per Story 1.8) ships the class.

### Why .NET uses `IdentityConstants.ApplicationScheme` and not a custom scheme name

`SignInManager.PasswordSignInAsync` writes the principal under `IdentityConstants.ApplicationScheme`. If a different scheme is configured, the sign-in succeeds (Identity wrote the cookie under its own scheme) but the auth middleware reads from the configured scheme — and sees no claims. The result: login appears to succeed (302 to `/`) but the subsequent request to `/` is unauthenticated and 302s right back to `/login`. This is a classic .NET Identity foot-gun. Use `IdentityConstants.ApplicationScheme` as both the registration name in `AddAuthentication(...)` and the scheme name in `AddCookie(...)` — the value is `"Identity.Application"` but using the constant prevents typos.

### Why `isPersistent: true` on sign-in

Local-dev artifact. Survives browser restarts (14-day cookie). For a production system this would be opt-in via a "Remember me" checkbox; here the dev iterates faster without re-logging-in on every server restart. Documented in `FieldMark/CLAUDE.md` (Task 13.1).

### Why 422 on form-validation failure (and 409 vs 422 distinction)

The PRD distinguishes (FR55 vs FR55a):

- **HTTP 409** — domain rule rejected the action (e.g., "cannot close a project with open violations"). The originating partial re-renders with an inline error; entity state is unchanged.
- **HTTP 422** — server-side input validation failed (malformed input, missing required field, type error). Same rendering contract; same partial; same `aria-invalid` + `aria-describedby` + top `InlineAlert`.

Login validation failure is **input validation**, not a domain rule violation (there is no `User.SignIn()` entity method). Use 422. The UX-DR34 acceptance from the epic file says "the form partial is re-rendered with HTTP 422" — that is the contract.

### Why the .NET form does **not** use `<form asp-page="...">` tag helpers

Two reasons:

1. **Markup parity.** The Razor tag helpers inject framework-specific attributes (e.g., `asp-action`, `asp-route-*`, sometimes additional anti-forgery wiring). Django and Go would have to mimic these exactly or the snapshot test (AC #10) fails. Writing a plain `<form>` makes the markup deterministic and snapshot-comparable.
2. **Antiforgery is preserved.** `@Html.AntiForgeryToken()` inserts the hidden input that the framework's antiforgery filter validates on POST. The token is excluded from snapshot diffing (Task 12.4) — Django's `{% csrf_token %}` is the per-stack equivalent and is also excluded.

### Why the Go login is a user-switcher (and not a password form)

ADR-012: real Go authentication is deferred. The Go stack has no password storage (Story 1.10 explicitly does not persist the manifest's `password` field for Go). A password form would be a UX lie — pretending the password matters when the stub middleware actually trusts the `X-FieldMark-Actor` cookie regardless of password. The user-switcher is the **honest** dev surface: pick a user, set the cookie, sign in.

The user-switcher diverges from the .NET / Django form (AC #1 names the divergence explicitly and the snapshot test (AC #10) excludes Go). When the deferred real-auth Go epic lands, the user-switcher will be replaced with a credential form; until then, the stub posture is documented in `fieldmark-go/CLAUDE.md`.

### Why a fallback authorization policy on .NET, not `[Authorize]` everywhere

Two equivalent shapes:

```csharp
// Shape A — fallback policy:
options.FallbackPolicy = new AuthorizationPolicyBuilder().RequireAuthenticatedUser().Build();
options.Conventions.AllowAnonymousToPage("/Account/Login");
// + AllowAnonymousToPage for each public path

// Shape B — per-page [Authorize]:
[Authorize] public class IndexModel : PageModel { ... }
[Authorize] public class PrivacyModel : PageModel { ... }
// + every other page
```

Shape A is "secure by default" — adding a new page doesn't accidentally expose it. Shape B is "open by default — opt in to security." For a teaching artifact whose theme is "server-authoritative everything," secure-by-default is the right teach. The exemptions are explicit and reviewable in `Program.cs`.

### How to test 403-without-state-leak for AC #6

The epic AC says "an unauthorized direct request returns HTTP 403 without leaking the entity state." Epic 1 has zero business state-change handlers — the AC is essentially "wire the 403 path correctly." The probe-route pattern in Task 9.3 / 10.1 / 11.3 satisfies this:

- Register a `/__authz_probe` endpoint **only inside test code** (test fixture / overridden URLconf / integration build tag).
- Require an ADMIN role on it.
- Log in as a non-ADMIN user (`diego` is SITE_SUPERVISOR — the seed manifest's most ergonomic foil).
- POST to the probe; assert 403; assert the response body does not contain any of the canonical entity-state strings.

This proves the 403 path is wired without introducing a probe in production. `make parity` (AC #11) verifies the probe is absent from the dumped routes.

### Open-redirect defence

The `return_url` / `next` parameter is the canonical open-redirect vector. All three stacks must validate it before honouring it:

- **.NET:** `Url.IsLocalUrl(ReturnUrl)` — built-in, returns true only for relative URLs and same-host absolute URLs.
- **Django:** `django.utils.http.url_has_allowed_host_and_scheme(next_url, allowed_hosts={request.get_host()})` — Django's canonical check; pass `require_https=False` (local dev).
- **Go:** the Go login redirect target is hard-coded to `/` in this story (the user-switcher doesn't carry a `next` param). If a `next` is ever added, write a small `isSafeRedirect(target string) bool` that allows only paths starting with `/` and not starting with `//` (which Go's HTTP libraries treat as protocol-relative).

### Antiforgery / CSRF posture

- **.NET:** Razor Pages enables antiforgery by default for `OnPostAsync` handlers. The `@Html.AntiForgeryToken()` call (Task 3.1) emits the matching hidden input. Do **not** add `[IgnoreAntiforgeryToken]`. The Story 1.6 `[IgnoreAntiforgeryToken]` on `/preferences/theme` is an unrelated exception for HTMX cookie writes; login forms post real credentials and must be CSRF-protected.
- **Django:** `CsrfViewMiddleware` is active by default (it is in the framework defaults Story 1.8 left untouched). The `{% csrf_token %}` tag in `_login.html` emits the matching token. Story 1.6's `set_theme` is exempted because HTMX requests carry the token via `hx-headers` (per Story 1.6 Task 4.6); the login form uses the standard `csrfmiddlewaretoken` POST field instead.
- **Go:** Stub auth, no CSRF middleware. The `X-FieldMark-Actor` cookie is dev-only and not a real credential. When the deferred real-Go-auth epic lands, CSRF middleware lands with it (documented in `fieldmark-go/CLAUDE.md` per Task 13.3).

### Cookie attributes — the SecurePolicy nuance

`SecurePolicy = CookieSecurePolicy.SameAsRequest` (Task 4.1) is **not** the same as `Always`. The latter forces `Secure` on the cookie regardless of request scheme — which breaks local-dev over `http://localhost:5000`. `SameAsRequest` matches the request scheme: HTTPS in production, HTTP locally. For an MVP that ships only as a local-dev artifact this is correct. If/when deployment becomes a goal, this changes — and that change is its own ADR (per Architecture §Deployment).

### Anti-patterns that must NOT slip in

- ❌ Adding `[Authorize]` attributes to every Razor Page individually. Use the fallback authorization policy + explicit `AllowAnonymousToPage` exemptions.
- ❌ Adding `app.UseAuthentication()` after `app.MapRazorPages()`. The middleware order matters; `UseAuthentication()` must precede `MapRazorPages()`.
- ❌ Putting `app.UseAuthentication()` between `UseHttpsRedirection` and `UseStaticFiles`. Static files must be served before auth so unauthenticated users see the favicon and stylesheets — the canonical order is documented in Task 4.2.
- ❌ Catching the form's password and re-rendering it on validation failure. The 422 response must clear the `password` field (do not pass it back to the view model). Render the username back (UX convenience) but never the password.
- ❌ Logging the password anywhere — request body, log line, exception detail. Identity's own logging is sufficient; do not add `logger.LogInformation($"Login attempt: {username}/{password}")`. This is true at MVP and forever.
- ❌ Returning a different status code on bad username vs bad password. Both return 422 with the generic "Invalid username or password." message. Disclosing which one was wrong is a documented user-enumeration vector.
- ❌ Setting `HttpOnly = false` on the .NET / Django session cookies. Identity defaults to `HttpOnly = true`; preserve it. Only the Go stub's `X-FieldMark-Actor` cookie has `HttpOnly = false`, and that is documented as dev-only.
- ❌ Adding GET handlers for `/logout`. POST-only, per FR54 (state-changing actions never use GET). The link in the header chrome is a `<form method="post">` wrapping a `<button>` — accessible, semantic, no JavaScript.
- ❌ Implementing the `authz.Can` primitive in this story. That is Story 1.12's scope. Story 1.11 ships `RequireRole(role)` (Go) and uses `[Authorize(Roles = ...)]` (.NET) / `@user_passes_test` (Django) for the AC #6 probe test only — the typed `Role` value object and entity-scope rules belong to 1.12.
- ❌ Mounting `RequireAuth()` on `/static/*` (Go). Static assets must be served to anonymous requests; the layout's stylesheets and the ThemeToggle script need to load on `/login` itself.
- ❌ Mounting `LoginRequiredMiddleware` (Django) before `AuthenticationMiddleware`. The order is: `AuthenticationMiddleware` first (sets `request.user`), then `LoginRequiredMiddleware` (reads `request.user.is_authenticated`).
- ❌ Editing `tools/parity/canonical-pg-indexes.txt`. Story 1.11 touches zero DDL.
- ❌ Editing `tools/parity/diff-routes.sh`. The new routes must just appear identically in all three dumps; the diff script is correct as written.
- ❌ Adding a "Sign up" or "Register" link to the login page. The dev users are seed-managed (Story 1.10); user registration is an explicit non-goal at MVP (no FR covers it). Do not add a registration page anywhere on any stack.
- ❌ Adding password reset, change-password, or email-confirmation flows. All deferred (no FRs cover them at MVP). Identity's `AddDefaultTokenProviders()` registration from Story 1.7 means the underlying machinery is available, but no UI is added.
- ❌ Mocking ASP.NET Core Identity in tests (e.g., `Mock<UserManager<...>>`). The integration tests in Task 9 hit a real Postgres via Testcontainers and use the real `SignInManager`. Per the project hard rule: no SQLite, real Postgres in tests.
- ❌ Adding `AddDefaultIdentity<IdentityUser<Guid>>` to fix a "Login page doesn't render" symptom. Story 1.7 deliberately uses `AddIdentityCore`. The login page in 1.11 is hand-authored; `AddDefaultIdentity` would add `/Identity/Account/*` Razor scaffolding and break the parity invariant.

### Project Structure Notes

Files this story adds or modifies:

**Shared:**
- **New:** `fieldmark_shared/components/login-form.example.html`
- **New:** `fieldmark_shared/components/login-error-region.example.html`
- **Update:** `fieldmark_shared/CLAUDE.md` — add `## Component Examples` section.
- **Update (conditional):** `fieldmark_shared/src/_components.css` — only if any of the form classes is missing.
- **Update (conditional):** `fieldmark_shared/dist/fieldmark.css` — rebuild if `_components.css` changed.

**.NET:**
- **New:** `FieldMark/FieldMark.Web/Pages/Account/Login.cshtml` (+ `.cs`)
- **New:** `FieldMark/FieldMark.Web/Pages/Account/Logout.cshtml` (+ `.cs`)
- **New:** `FieldMark/FieldMark.Web/Authentication/ClaimsPrincipalExtensions.cs`
- **New:** `FieldMark.Tests.Domain/ClaimsPrincipalExtensionsTests.cs` (or move to a new `FieldMark.Tests.Web/`)
- **New:** `FieldMark.Tests.Integration/AuthFlowTests.cs`
- **New:** `FieldMark.Tests.Integration/LoginFormSnapshotTests.cs`
- **New (conditional):** `FieldMark.Tests.Integration/Fixtures/PostgresFixture.cs` — if not already present.
- **New (conditional):** `FieldMark.Tests.Integration/Helpers/NormaliseHtml.cs` — if not already present from Story 1.5/1.6.
- **Update:** `FieldMark/FieldMark.Web/Program.cs` — add `AddAuthentication`/`AddCookie`/`AddAuthorization`/`AddRazorPages(options =>...)` blocks; insert `UseAuthentication`/`UseAuthorization` middleware.
- **Update:** `FieldMark/FieldMark.Web/Pages/Shared/_Layout.cshtml` — header chrome login/logout links.
- **Update:** `FieldMark/CLAUDE.md` — rewrite `## Authentication`.

**Django:**
- **New:** `fieldmark_py/templates/_login.html`
- **New:** `fieldmark_py/fieldmark/authn.py`
- **New:** `fieldmark_py/fieldmark/tests/__init__.py` (if not present)
- **New:** `fieldmark_py/fieldmark/tests/test_authn.py`
- **New:** `fieldmark_py/fieldmark/tests/test_auth_flow.py`
- **New:** `fieldmark_py/fieldmark/tests/test_login_snapshot.py`
- **Update:** `fieldmark_py/fieldmark/views.py` — add `login_view`, `logout_view`; decorate `set_theme` with `@login_not_required`.
- **Update:** `fieldmark_py/fieldmark/urls.py` — add `/login`, `/logout` routes.
- **Update:** `fieldmark_py/fieldmark/settings.py` — `LOGIN_URL`, `LoginRequiredMiddleware`.
- **Update:** `fieldmark_py/conftest.py` (or create) — `seeded_users` fixture.
- **Update:** `fieldmark_py/pytest.ini` — add `fieldmark` to `testpaths` if missing.
- **Update:** `fieldmark_py/CLAUDE.md` — extend `## Authentication`.

**Go:**
- **New:** `fieldmark-go/internal/web/handlers/auth.go`
- **New:** `fieldmark-go/internal/web/handlers/auth_test.go`
- **New:** `fieldmark-go/internal/web/handlers/auth_integration_test.go`
- **New:** `fieldmark-go/internal/web/auth/authz.go`
- **New:** `fieldmark-go/internal/web/templates/pages/login.html`
- **New (conditional):** `fieldmark-go/internal/web/viewmodels/layout.go` — if Story 1.5/1.6 did not already create a chrome-level view model.
- **Update:** `fieldmark-go/internal/web/auth/stub.go` — add `CookieName()` accessor.
- **Update:** `fieldmark-go/cmd/web/main.go` — register `/login`/`/logout`, mount `RequireAuth()` on business routes.
- **Update:** `fieldmark-go/internal/web/templates/partials/header.html` — Sign in / Sign out link.
- **Update:** `fieldmark-go/CLAUDE.md` — extend `## Authentication`.

**Root:**
- **Update:** `README.md` — `make seed` hint + dev credentials note.

No file under `_bmad-output/planning-artifacts/`, `docker/`, `docs/`, `e2e/`, `tools/parity/`, or `tools/git-hooks/` is modified by this story.

### Testing Standards

Per root `CLAUDE.md` → `docs/hard-rules.md` and each stack's CLAUDE.md:

- **No SQLite** — real Postgres for every DB-touching test. .NET uses Testcontainers; Django uses `@pytest.mark.django_db`; Go uses `//go:build integration` against a local or Testcontainers Postgres.
- **No mocks of framework auth primitives** — no `Mock<UserManager<...>>`, no `mock.patch("django.contrib.auth.authenticate")`. Test against the real machinery hitting a real DB.
- **Snapshot tests strip per-stack noise** — antiforgery tokens (.NET / Django) and any framework-injected attributes are documented in the test source.
- **Probe route stays inside test code** — the AC #6 `/__authz_probe` route is registered only in test fixtures, never in production code paths. `make parity` enforces.
- **Cross-stack parity is the gating mechanism** — if a test passes on one stack and fails on another, the story is not done.

### Previous Story Intelligence

**Story 1.5 (base layout — `ready-for-dev`):**
- Skip-link, landmarks, FlashRegion are in chrome on every stack. The login page renders inside the same layout chrome — its `<h1>` is the page-level heading; the FlashRegion announces post-redirect flash messages (e.g., "You have been signed out.").
- Per-stack template paths: `_Layout.cshtml`, `templates/base.html`, `internal/web/templates/layouts/base.html`.

**Story 1.6 (ThemeToggle — `ready-for-dev`):**
- `/preferences/theme` (POST) is registered on every stack. It must be reachable while unauthenticated so the theme works on the login page itself — Tasks 4.1 (`AllowAnonymousToFolder("/Preferences")`), 5.4 (`@login_not_required` on `set_theme`), and 6.4 (don't mount `RequireAuth` on `/preferences/theme`) all enforce this.
- The `dump_routes` upgrades from 1.6 are the reason `make parity` can now distinguish `post /login` from `get /login`.
- The ThemeToggle script (`fieldmark_shared/vendor/theme-toggle/theme-toggle.js`) is the only inline-adjacent JavaScript in the application; the login page picks up this contract unchanged.

**Story 1.7 (.NET Identity — `review`):**
- `AddIdentityCore<IdentityUser<Guid>>` + roles + sign-in-manager is in place. Five canonical roles seeded in `dotnet_auth.roles`. Identity tables under `dotnet_auth.*`.
- `app.UseAuthentication()` is **not** wired — Task 4 wires it.
- No `Pages/Account/*` scaffolded — Tasks 3 and 3.6 author them.

**Story 1.8 (Django auth — `ready-for-dev`):**
- `django.contrib.auth` is enabled; `search_path=django_auth,public` is configured; five conceptual-role Groups seeded.
- No `LoginRequiredMiddleware`, no `LOGIN_URL`, no `/login` / `/logout` urls — Tasks 5.3 / 5.4 add them.
- CSRF middleware is active by default; the login form uses `{% csrf_token %}`.

**Story 1.9 (Go stub auth — `ready-for-dev`):**
- `StubAuthMiddleware` hydrates `*app.Actor` from cookie/header/env. `RequireAuth()` factory exists but is not mounted on any route.
- `fiber_auth.users` + `fiber_auth.user_roles` schema is in place via `cmd/migrate-fiber-auth`.
- Cookie name `X-FieldMark-Actor` carries `username` (not UUID); the middleware does the username → UUID translation.

**Story 1.10 (dev-user manifest — `ready-for-dev`):**
- Six users seeded across all three stacks from `docker/postgres/init/seed-uuids/dev-users.json`.
- Shared password (e.g., `FieldMark!2026`) usable on .NET and Django. Go ignores password.
- Django stores the canonical UUID in `django_auth.dev_user_uuid` (side table); access via `user.dev_uuid.uuid`.
- `make seed` runs all three seeders.

**Story 1.3 (parity tooling — `done`):**
- `make parity` is the cross-stack guard. AC #8 of this story is bound by it.
- `--dump-routes` early-return is sacred; preserve it in any `Program.cs` / `main.go` / `manage.py` edits.

### Git Intelligence

Recent commits (most relevant to this story):

- `d03f0fe feat: e1s3 establish tools parity` — `make parity` baseline. Story 1.11 adds three new routes per stack; expect those three lines to appear identically in all dumps after this story.
- `cbf47e9 feat: e1s2 verified sql init scripts` — `*_auth` schemas already created (empty pre-1.7/1.8/1.9). Story 1.11 consumes the populated schemas from upstream seed work.
- `a6fac88 feat: e1s1 confirm scaffolds` — original `cmd/web/main.go`, `Program.cs`, `urls.py` baselines. By the time 1.11 runs, those files have been touched by 1.5–1.10; verify against the actual HEAD before editing.

No prior commit has added a login flow on any stack. Story 1.11 is the first.

### Latest Technical Information

- **.NET 10 / EF Core 10.0.7** — `SignInManager<TUser>` is the canonical sign-in machinery. `AddAuthentication(scheme).AddCookie(scheme, opts)` configures the cookie scheme; `IdentityConstants.ApplicationScheme` is the constant. Razor Pages' `LocalRedirect` is open-redirect-safe.
- **ASP.NET Core 10 fallback authorization policy** — `options.FallbackPolicy = new AuthorizationPolicyBuilder().RequireAuthenticatedUser().Build();` is the secure-by-default pattern. `options.Conventions.AllowAnonymousToPage(path)` is the exemption mechanism.
- **Django 6.0.4** — `LoginRequiredMiddleware` (added in Django 5.2) is available. `@login_not_required` is the exemption decorator (Django 5.2+). `django.utils.http.url_has_allowed_host_and_scheme` is the open-redirect safe-check.
- **Fiber v3.2.0** — `c.Cookie(*fiber.Cookie)` to write; `c.Cookies(name)` to read. `c.Redirect().Status(StatusFound).To(path)` is the redirect idiom. `c.Render(template, data)` is the html/template renderer (set up by Story 1.5).
- **pgx v5.9.2** — `pool.QueryRow(ctx, sql, args...).Scan(...)` for single-row reads (used by `lookupUser`); `pool.Query(ctx, sql, args...)` + `rows.Next()` + `rows.Scan(...)` for list reads (used by `listSeededUsers`). `pgx.ErrNoRows` is the canonical no-rows sentinel.
- **Postgres 17** — `CHECK` constraints, `UNIQUE` indexes, foreign keys with `ON DELETE CASCADE` are all stable. The auth schemas are populated; the story performs only DML (cookie state) at runtime and no DDL.
- **No new package dependencies on any stack.** Identity (.NET), `django.contrib.auth` (Django), and Fiber + pgx (Go) are all already in scope from upstream stories.

### References

- [Architecture: Authentication & Security → D6, D7, D8, D9](_bmad-output/planning-artifacts/architecture.md#authentication--security) — locked decisions and resolved opens.
- [Architecture: Architectural Boundaries → Authentication / authorization](_bmad-output/planning-artifacts/architecture.md#architectural-boundaries) — opaque UUID refs; per-stack-idiomatic auth implementation.
- [Architecture: API & Communication Patterns → D13 Error rendering pattern](_bmad-output/planning-artifacts/architecture.md#api--communication-patterns) — 409 / 422 / 403 handling shapes.
- [Architecture: Repository Directory Structure](_bmad-output/planning-artifacts/architecture.md#complete-repository-directory-structure) — file locations for `Pages/Account/*`, `templates/_login.html`, `internal/web/handlers/auth.go`, `internal/web/templates/pages/login.html`.
- [PRD FR1–FR8 — Authentication & Authorization](_bmad-output/planning-artifacts/prd/functional-requirements.md) — framework-local authentication; conceptual roles; FR3 logout, FR4 redirect, FR7 authorization rejection.
- [PRD FR54, FR55, FR55a, FR56, FR60–FR64](_bmad-output/planning-artifacts/prd/functional-requirements.md) — POST-only state changes; 409/422/403 contracts; form validation a11y.
- [PRD architectural-constraints-prd-binding.md — Authentication & Authorization (ADR-012)](_bmad-output/planning-artifacts/prd/architectural-constraints-prd-binding.md) — schema isolation; per-stack auth ownership; Go stub posture.
- [UX-DR14, UX-DR32](_bmad-output/planning-artifacts/ux-design-specification.md) — FlashRegion / live-region politeness.
- [UX-DR15](_bmad-output/planning-artifacts/ux-design-specification.md) — ThemeToggle placement in header chrome.
- [UX-DR33](_bmad-output/planning-artifacts/ux-design-specification.md) — skip-link, landmarks, heading hierarchy.
- [UX-DR34](_bmad-output/planning-artifacts/ux-design-specification.md) — form validation announcement (422 / `aria-invalid` / `aria-describedby` / `role="alert"`).
- [UX-DR35](_bmad-output/planning-artifacts/ux-design-specification.md) — focus styling and touch-target sizes.
- [docs/hard-rules.md](docs/hard-rules.md) — backend authority; real Postgres in tests; stack symmetry on routes; POST-only state changes.
- [FieldMark/CLAUDE.md](FieldMark/CLAUDE.md) — .NET-specific rules.
- [fieldmark_py/CLAUDE.md](fieldmark_py/CLAUDE.md) — Django-specific rules; no signals.
- [fieldmark-go/CLAUDE.md](fieldmark-go/CLAUDE.md) — Go-specific rules; `fiber.Ctx` stays in `internal/web/`.
- [Story 1.3 implementation artifact](_bmad-output/implementation-artifacts/1-3-establish-tools-parity-and-make-parity-with-per-stack-dump-routes.md) — `--dump-routes` invariant; `make parity` contract.
- [Story 1.5 implementation artifact](_bmad-output/implementation-artifacts/1-5-implement-cross-stack-base-layout-with-skip-link-landmarks-and-flashregion.md) — base layout shape; FlashRegion contract.
- [Story 1.6 implementation artifact](_bmad-output/implementation-artifacts/1-6-implement-themetoggle-with-cookie-persistence-per-stack.md) — `/preferences/theme` route; dump-routes method awareness.
- [Story 1.7 implementation artifact](_bmad-output/implementation-artifacts/1-7-wire-asp-net-core-identity-to-dotnet-auth-schema-with-conceptual-roles.md) — .NET Identity registration; deliberate `app.UseAuthentication()` deferral to this story.
- [Story 1.8 implementation artifact](_bmad-output/implementation-artifacts/1-8-wire-django-built-in-auth-to-django-auth-schema-with-conceptual-role-groups.md) — Django auth setup; CSRF posture; deliberate `LoginRequiredMiddleware` deferral.
- [Story 1.9 implementation artifact](_bmad-output/implementation-artifacts/1-9-implement-go-fiber-stub-authentication-middleware.md) — `StubAuthMiddleware`, `RequireAuth()` factory; the cookie / header / env resolution order.
- [Story 1.10 implementation artifact](_bmad-output/implementation-artifacts/1-10-author-shared-uuid-dev-user-manifest-and-per-stack-idempotent-seed-runners.md) — dev-user manifest; six seeded users; password expectation; `DevUserUuid` side table on Django.

## Dev Agent Record

### Agent Model Used

_(populated by dev agent)_

### Debug Log References

### Completion Notes List

### File List
