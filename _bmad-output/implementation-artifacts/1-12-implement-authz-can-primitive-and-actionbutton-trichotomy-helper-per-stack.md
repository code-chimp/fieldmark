# Story 1.12: Implement `authz.Can` primitive and ActionButton trichotomy helper per stack

Status: ready-for-dev

## Story

As a developer rendering an action affordance on any FieldMark screen,
I want a single template helper per stack that decides `absent | disabled-with-tooltip | present` per the affordance trichotomy (UX-DR10, UX-DR21, Pattern 2),
So that future epics introduce action buttons by calling one helper with `permission` and `state_allows` booleans — never re-deciding the rendering rule per screen, and never letting a stack drift from the cross-stack contract (FR5, FR6, FR58, FR59).

## Acceptance Criteria

1. **Each stack exposes a uniform authorization decision primitive `Can(user, action, entity?) → bool` (FR5).** Canonical per-stack call signatures and locations:
   - **.NET** — `FieldMark.Web.Authorization.DomainPolicies.Can(ClaimsPrincipal user, string action, Guid? entityId = null) : bool` in `FieldMark/FieldMark.Web/Authorization/DomainPolicies.cs`.
   - **Django** — `fieldmark.authz.can(user, action: str, entity_id: uuid.UUID | None = None) -> bool` in `fieldmark_py/fieldmark/authz.py`.
   - **Go** — `authz.Can(actor *app.Actor, action string, entityID uuid.UUID) bool` in `fieldmark-go/internal/web/auth/authz.go` (a sibling of `stub.go` from Story 1.9). `uuid.Nil` denotes "no entity scope".
   - At this story there are no domain entities, so the entity-scope parameter is accepted by the signature but not yet consulted by any internal rule — see AC #2 for the role-only Epic-1 behavior. The signature shape is the cross-stack contract; the rules behind it grow with Epics 2+.

2. **The Epic-1 implementation consults the user's conceptual role(s) only (entity-scope rules deferred).** Behavior on every stack:
   - If the user is anonymous (not authenticated / no resolved role): `Can` returns `false` for every action string.
   - If the user has at least one of the five canonical roles (`ADMIN`, `COMPLIANCE_OFFICER`, `INSPECTOR`, `SITE_SUPERVISOR`, `EXECUTIVE`): `Can` consults an internal `ActionRoleMap` (a `Dictionary<string, ISet<string>>` / `dict[str, frozenset[str]]` / `map[string]map[string]struct{}`) keyed by `action` string and returning the set of roles permitted to perform that action. Membership test: `actor.Roles ∩ map[action] ≠ ∅` ⇒ `true`; else `false`.
   - The `ActionRoleMap` ships **empty** in this story (no actions registered — Epic 1 has no live use sites; subsequent stories register their own actions). For every action string passed in by callers — including the test-only action `"test.allow_admin"` registered in unit tests (see AC #7) — `Can` returns `false` if the action is not in the map; otherwise the role-intersection rule applies.
   - Entity-scope rules (e.g., "Site Supervisor can act on a Violation only if assigned to it") are **not** evaluated in this story. The `entityId`/`entity_id`/`entityID` parameter is captured in the signature and forwarded to a single internal extension point (`evaluateEntityScope(action, entityId)` returning `true` by default) so Epic 2+ stories can wire entity rules without changing the call-site contract. Document the deferral in code comments.

3. **A canonical ActionButton example exists at `fieldmark_shared/components/action_button.example.html` and is the cross-stack snapshot target (UX-DR10, FR58).** The file contains exactly three `<!-- variant: ... -->`-delimited blocks representing the three states; each variant is rendered from a fixed input set documented inline. The three variants:
   - **`absent`** — `permission=false`. The block is the literal empty string (zero bytes between the variant delimiters). Verified by the per-stack snapshot test asserting the helper returns an empty string for this input.
   - **`disabled`** — `permission=true && state_allows=false`. A single Basecoat-styled `<button>` carrying: `type="button"`, `disabled`, `aria-disabled="true"`, `class="btn btn-secondary"`, `data-tooltip="<disabled_reason>"`, `aria-describedby="<id>-reason"`, an accessible label as visible text, and a sibling visually-hidden `<span id="<id>-reason" class="sr-only"><disabled_reason></span>` so screen readers receive the reason text. (Basecoat's `[data-tooltip]` is a CSS-only pseudo-element invisible to assistive tech; the `sr-only` span carrying the same text closes the WCAG gap — see Dev Notes for the rationale.)
   - **`present`** — `permission=true && state_allows=true`. A Basecoat-styled `<button>` carrying: `type="button"`, `class="btn btn-primary"`, `hx-post="<hx_post>"`, `hx-target="<hx_target>"`, `hx-swap="outerHTML"`, `hx-disabled-elt="this"` (UX-DR27, FR64), and the accessible label as visible text. No `data-tooltip`, no `aria-describedby`.
   - The exact canonical input set used to generate every variant is captured in the file header as YAML-in-HTML-comment so the three stack wrappers can read identical inputs from a single source of truth. Suggested fixture id: `id="ab-fixture-1"`, label `"Approve Resolution"`, `hx_post="/violations/00000000-0000-0000-0000-000000000001/corrective-actions/00000000-0000-0000-0000-000000000002/approve"`, `hx_target="#violation-detail"`, `disabled_reason="Awaiting review"`.

4. **Each stack ships an ActionButton template wrapper that, given those canonical inputs, renders byte-identical output to the corresponding variant block in `action_button.example.html` (FR58).** Wrapper locations and shapes:
   - **.NET** — `FieldMark/FieldMark.Web/Pages/Shared/_ActionButton.cshtml` as a Razor partial with a strongly-typed `ActionButtonVm` model declared at `FieldMark/FieldMark.Web/ViewModels/Components/ActionButtonVm.cs`. Properties: `Id : string`, `Permission : bool`, `StateAllows : bool`, `Label : string`, `HxPost : string`, `HxTarget : string`, `DisabledReason : string?`. Rendered via `<partial name="_ActionButton" model="@vm" />`. Domain-pure projection: the VM is constructed in handler/page-model code (or in the `_ActionButton` partial caller); no DI inside the partial.
   - **Django** — `fieldmark_py/templates/components/_action_button.html` as a regular template. Included via `{% include "components/_action_button.html" with id=... permission=... state_allows=... label=... hx_post=... hx_target=... disabled_reason=... %}`. The template is logic-light: three `{% if/elif/else %}` branches. No new template tags, no `inclusion_tag` decorator — the architecture forbids speculative abstractions.
   - **Go** — `fieldmark-go/internal/web/templates/components/action_button.tmpl` as a Go `html/template` definition named `action_button` (i.e., `{{ define "action_button" }}...{{ end }}`). Invoked via `{{ template "action_button" .ActionButton }}` where the data context is a small `ActionButtonVM` struct (`internal/web/viewmodels/action_button.go`) carrying the same seven fields. The template package's `engine.AddFunc` registration in `cmd/web/main.go` is **not** modified — no custom func is needed (the template is data-driven, not function-driven).
   - "Byte-identical" is normalized: tabs/spaces between tags collapsed to a single space, leading/trailing whitespace per line trimmed, `<!-- ... -->` HTML comments stripped, attribute order canonicalized alphabetically. The normalization helper is **already shared** with Stories 1.5 and 1.6 (`fieldmark_py/tests/normalize_html.py`, `FieldMark.Tests.Integration/Helpers/NormaliseHtml.cs`, `fieldmark-go/internal/web/testutil/normalizehtml.go`) — reuse them; do not author a fourth normalizer.

5. **The disabled state implements full accessibility per UX-DR21 / UX-DR10 / WCAG 2.1 AA.** Specifically:
   - The disabled button retains focus (it is a real `<button>` with `disabled` plus `aria-disabled="true"`, NOT a CSS-hidden element; `disabled` removes it from default browser tab order but `aria-disabled="true"` keeps it discoverable to assistive tech and we explicitly add `tabindex="0"` to bring it back into keyboard tab order so screen-reader users can navigate to it and hear the reason — this is the documented Basecoat-disabled-affordance pattern).
   - The `data-tooltip` attribute renders the Basecoat CSS-only tooltip on hover/focus (UX-DR10 surface).
   - The same `disabled_reason` text is duplicated into a visually-hidden `<span id="<id>-reason" class="sr-only">` and referenced via the button's `aria-describedby="<id>-reason"`. This ensures keyboard / screen-reader users receive the reason without depending on Basecoat's `::before` pseudo-element (which is invisible to assistive tech).
   - Verified by `axe-core` and by an explicit keyboard-navigation unit test (see Task 5.4 — focuses the disabled button, asserts `:focus-visible` outline applies, and that the linked `sr-only` reason node exists in the DOM with the same text content).

6. **Per-stack `Role` value object is introduced (deferred from Stories 1.7 / 1.8 / 1.9).** Each stack centralizes the five canonical role names so the helper, `Can`, and any future call site share one source of truth. Locations and shapes:
   - **.NET** — `FieldMark/FieldMark.Domain/ValueObjects/Role.cs` declaring `public sealed record Role` with five `public static readonly Role` instances (`Admin`, `ComplianceOfficer`, `Inspector`, `SiteSupervisor`, `Executive`), each carrying a single `Name` string property holding the canonical screaming-snake value (`"ADMIN"`, etc.). Add a static `Parse(string)` method and a static `All` collection. `FieldMark.Domain` still has zero outbound references (this is a pure type — no EF Core, no framework).
   - **Django** — `fieldmark_py/fieldmark/roles.py` declaring a `Role` Enum (subclass of `str, Enum` so values stringify cleanly) with the same five members. `seed_groups` (Story 1.8) is **refactored** to import from `Role` rather than carrying its own `CANONICAL_GROUPS` tuple — same behavior, one source of truth. Tests for `seed_groups` continue to pass with the import change.
   - **Go** — `fieldmark-go/internal/domain/role.go` declaring `type Role string` with five `const` values. `internal/web/auth/lookup.go` (Story 1.9) and `internal/web/auth/stub.go` are **refactored** to import the canonical names from `domain.Role` rather than referencing bare string literals. (The `CHECK` constraint in `fiber_auth.user_roles` from Story 1.9 stays as-is — DB constraint is authoritative; the Go const just consumes it.) `internal/domain` continues to hold no outbound non-stdlib imports.
   - The `RoleSeeder` in .NET (Story 1.7) is **refactored** to read from `FieldMark.Domain.ValueObjects.Role.All` rather than carrying its own private `CanonicalRoles` array — same behavior, one source of truth. The existing role-seeding unit/integration coverage continues to pass.

7. **Unit tests prove the helper and `Can` invariants on every stack (FR66, FR58).** Each stack carries the following test set; each set is independently runnable and uses no DB:
   - **`Can_AnonymousActor_ReturnsFalse`** — supply an anonymous actor + the test-only action `"test.allow_admin"` (registered for the test only via a fixture / test-fixture-only `ActionRoleMap` mutation that wraps the production helper with one entry); assert `false`.
   - **`Can_AdminActor_ReturnsTrueForAdminScopedAction`** — register `"test.allow_admin" → {ADMIN}` in the test fixture; supply an authenticated actor with role `ADMIN`; assert `true`.
   - **`Can_NonAdminActor_ReturnsFalseForAdminScopedAction`** — supply an authenticated actor with role `SITE_SUPERVISOR`; assert `false`.
   - **`Can_UnknownAction_ReturnsFalse`** — supply an authenticated `ADMIN` actor and an action string not in the map (`"test.unmapped"`); assert `false`.
   - **`ActionButton_PermissionFalse_RendersEmpty`** — call the helper with `permission=false`; assert the normalized output is the empty string.
   - **`ActionButton_DisabledVariant_MatchesCanonicalSnapshot`** — call the helper with the canonical disabled-variant inputs (AC #3); normalize; compare to the `<!-- variant: disabled -->` block of `fieldmark_shared/components/action_button.example.html`; assert byte-identical.
   - **`ActionButton_PresentVariant_MatchesCanonicalSnapshot`** — same shape for the present variant.
   - **`ActionButton_DisabledVariant_HasScreenReaderReason`** — parse the disabled-variant output with the stack's idiomatic HTML parser (`HtmlAgilityPack` on .NET, `BeautifulSoup` on Django, `golang.org/x/net/html` on Go) and assert: a `<button>` exists with `disabled`, `aria-disabled="true"`, `tabindex="0"`, `data-tooltip` matching the reason text, and `aria-describedby` whose value matches the `id` of a sibling element bearing `class="sr-only"` and text content equal to the reason text.
   - Test files: `FieldMark.Tests.Domain/Authorization/CanTests.cs`, `FieldMark.Tests.Integration/Components/ActionButtonRenderingTests.cs` (.NET); `fieldmark_py/fieldmark/tests/test_authz.py`, `fieldmark_py/fieldmark/tests/test_action_button_template.py` (Django); `fieldmark-go/internal/web/auth/authz_test.go`, `fieldmark-go/internal/web/templates/components/action_button_test.go` (Go).

8. **No live ActionButton call sites exist in Epic 1 (FR58 parity invariant).** Epic 1 has zero rendered action affordances — the helper exists for Epic 2+. Verified by:
   - **.NET** — `grep -rn '_ActionButton' FieldMark/FieldMark.Web/Pages/` (excluding `FieldMark/FieldMark.Web/Pages/Shared/_ActionButton.cshtml` itself) returns zero matches.
   - **Django** — `grep -rn 'components/_action_button.html' fieldmark_py/` (excluding `fieldmark_py/templates/components/_action_button.html` itself and the test files) returns zero matches.
   - **Go** — `grep -rn 'template "action_button"' fieldmark-go/internal/web/templates/` (excluding the component file itself and the test file) returns zero matches.

9. **`make parity` exits 0; route inventory and `pg_indexes` are unchanged.** Story 1.12 adds zero routes and zero DDL. Verified by:
   - From repo root: `make parity` — exits 0.
   - From each stack root, the `--dump-routes` output is byte-identical to its HEAD-before-this-story output.
   - `tools/parity/canonical-pg-indexes.txt` is **not** edited — auth schemas are out of scope; `domain.*` is unchanged.

10. **Each stack's `CLAUDE.md` `Authorization` section (or new `## Authorization` section after `## Authentication`) is added or updated.** Each stack's CLAUDE.md documents: (a) where `Can` lives, (b) its signature, (c) the `ActionRoleMap` extension point and where future stories should register new actions, (d) where the `Role` value object lives and that role-name string literals elsewhere are a defect, (e) the ActionButton wrapper path and that `permission` is decided by `Can` (not by the template). Cross-references the canonical `fieldmark_shared/components/action_button.example.html`.

11. **Build, type, lint, and test gates stay green on every stack.**
    - **.NET:** `cd FieldMark && dotnet build && dotnet test` — clean. `dotnet csharpier format .` reports zero diffs. `TreatWarningsAsErrors=true` honoured.
    - **Django:** `cd fieldmark_py && uv run ruff check . && uv run mypy . && uv run pytest` — clean. New tests from AC #7 pass.
    - **Go:** `cd fieldmark-go && make fmt-check && make vet && make staticcheck && make test` — clean.

## Tasks / Subtasks

- [ ] Task 1: Read all upstream story artifacts and confirm dependency posture (AC: all)
  - [ ] 1.1 Read [Story 1.7](_bmad-output/implementation-artifacts/1-7-wire-asp-net-core-identity-to-dotnet-auth-schema-with-conceptual-roles.md) — note: `RoleSeeder.cs` carries `private static readonly string[] CanonicalRoles` that Story 1.12 must consolidate into `Role.cs` (Story 1.7 Dev Notes "Role names — strings now, enum later" calls out this refactor explicitly).
  - [ ] 1.2 Read [Story 1.8](_bmad-output/implementation-artifacts/1-8-wire-django-built-in-auth-to-django-auth-schema-with-conceptual-role-groups.md) — note: `seed_groups.py` carries `CANONICAL_GROUPS = (...)` that Story 1.12 must consolidate. The Story 1.8 Dev Notes "Do not introduce a top-level constant or enum elsewhere (e.g., fieldmark/roles.py) for the role names. Story 1.12 (authz.Can primitive) is the right place" — this story is that place.
  - [ ] 1.3 Read [Story 1.9](_bmad-output/implementation-artifacts/1-9-implement-go-fiber-stub-authentication-middleware.md) — note: the `fiber_auth.user_roles.role` CHECK constraint carries the five canonical names; `internal/web/auth/lookup.go` reads them. Story 1.9 Dev Notes "Do not introduce a domain.Role enum at this story. The five role names live as a CHECK constraint... The typed Role value object lands with Story 1.12" — this story is that place.
  - [ ] 1.4 Read the architecture's [Authorization expression pattern](_bmad-output/planning-artifacts/architecture.md#process-patterns) and the [`can_*` boolean rendering pattern](_bmad-output/planning-artifacts/architecture.md#process-patterns) — note: `Can` is the single call site every handler uses; the wrapper takes the **already-computed** `permission` bool, not a `(user, action)` pair. Templates do not call `Can` directly.
  - [ ] 1.5 Read [`fieldmark_shared/CLAUDE.md`](fieldmark_shared/CLAUDE.md) (if present) and confirm the `components/` subdirectory (sibling of `src/`) does not exist yet — it is created by this story, as noted in [UX spec line 931](_bmad-output/planning-artifacts/ux-design-specification.md). Story 1.4 (`review`) established `src/` and `dist/`; this story adds `components/`.
  - [ ] 1.6 If Story 1.4 has not merged when this story begins, surface the dependency — the canonical Basecoat classes (`.btn`, `.btn-primary`, `.btn-secondary`, `.sr-only`, `[data-tooltip]` styling) come from `fieldmark_shared/dist/fieldmark.css` which Story 1.4 produces. The wrapper templates produce markup; the styling depends on Story 1.4 having shipped.

- [ ] Task 2: Author the canonical `fieldmark_shared/components/action_button.example.html` (AC: #3)
  - [ ] 2.1 Create the directory `fieldmark_shared/components/` (sibling of `src/`). Add a `README.md` inside with one paragraph stating that this directory holds canonical static HTML examples per the UX spec "Canonical examples in `fieldmark_shared/components/`" rule, and that every per-stack template wrapper is snapshot-tested against the relevant file here.
  - [ ] 2.2 Create `fieldmark_shared/components/action_button.example.html`. Use exactly this structure (replace `__REASON__` etc. with the canonical fixture values; do **not** add per-stack class names — Basecoat classes only):

    ```html
    <!--
    ActionButton component — canonical example.
    Each variant is delimited by a `<!-- variant: <name> -->` line. The block
    between two delimiters is the helper's expected output for the inputs
    documented in `inputs:` for that variant. The per-stack snapshot tests
    parse this file by delimiter, normalize whitespace + comments + attribute
    order, and compare against the helper output.

    fixture:
      id: ab-fixture-1
      label: Approve Resolution
      hx_post: /violations/00000000-0000-0000-0000-000000000001/corrective-actions/00000000-0000-0000-0000-000000000002/approve
      hx_target: #violation-detail
      disabled_reason: Awaiting review
    -->

    <!-- variant: absent (inputs: permission=false, state_allows=*) -->
    <!-- variant: disabled (inputs: permission=true, state_allows=false) -->
    <button
      type="button"
      id="ab-fixture-1"
      class="btn btn-secondary"
      disabled
      aria-disabled="true"
      tabindex="0"
      data-tooltip="Awaiting review"
      aria-describedby="ab-fixture-1-reason"
    >Approve Resolution</button>
    <span id="ab-fixture-1-reason" class="sr-only">Awaiting review</span>

    <!-- variant: present (inputs: permission=true, state_allows=true) -->
    <button
      type="button"
      id="ab-fixture-1"
      class="btn btn-primary"
      hx-post="/violations/00000000-0000-0000-0000-000000000001/corrective-actions/00000000-0000-0000-0000-000000000002/approve"
      hx-target="#violation-detail"
      hx-swap="outerHTML"
      hx-disabled-elt="this"
    >Approve Resolution</button>
    ```

    Note: the `absent` variant's block is intentionally empty (a single newline between its delimiter and the next variant's delimiter). The snapshot helper strips this to an empty string after normalization.

  - [ ] 2.3 Confirm no other consumer relies on `fieldmark_shared/components/` existing — `fieldmark_shared/package.json`'s `build` script writes only to `dist/fieldmark.css`, so the new `components/` directory is purely a source-of-truth folder and does not need pipeline changes.

- [ ] Task 3: .NET implementation — `Role`, `Can`, ActionButton partial, tests (AC: #1, #2, #4, #5, #6, #7, #10)
  - [ ] 3.1 Create `FieldMark/FieldMark.Domain/ValueObjects/Role.cs`:

    ```csharp
    namespace FieldMark.Domain.ValueObjects;

    /// <summary>
    /// Conceptual role of an authenticated FieldMark user. The five canonical
    /// names are persisted in dotnet_auth, django_auth, and fiber_auth — this
    /// type is the single .NET-side source of truth for them.
    /// </summary>
    public sealed record Role
    {
        public string Name { get; }

        private Role(string name) => Name = name;

        public static readonly Role Admin             = new("ADMIN");
        public static readonly Role ComplianceOfficer = new("COMPLIANCE_OFFICER");
        public static readonly Role Inspector         = new("INSPECTOR");
        public static readonly Role SiteSupervisor    = new("SITE_SUPERVISOR");
        public static readonly Role Executive         = new("EXECUTIVE");

        public static IReadOnlyList<Role> All { get; } = new[]
        {
            Admin, ComplianceOfficer, Inspector, SiteSupervisor, Executive,
        };

        public static Role Parse(string name) =>
            All.FirstOrDefault(r => r.Name == name)
            ?? throw new ArgumentException($"Unknown role name: {name}", nameof(name));

        public override string ToString() => Name;
    }
    ```

  - [ ] 3.2 Refactor `FieldMark/FieldMark.Web/SeedData/RoleSeeder.cs` (Story 1.7) to read from `Role.All`:

    ```csharp
    // Replace:
    private static readonly string[] CanonicalRoles = { "ADMIN", "COMPLIANCE_OFFICER", ... };

    // With:
    private static readonly IReadOnlyList<string> CanonicalRoles =
        FieldMark.Domain.ValueObjects.Role.All.Select(r => r.Name).ToList();
    ```

    Add `using FieldMark.Domain.ValueObjects;` at the top. The existing Story 1.7 unit/integration coverage continues to pass without test edits — the public contract of `RoleSeeder.SeedAsync` is unchanged.

  - [ ] 3.3 Create `FieldMark/FieldMark.Web/Authorization/DomainPolicies.cs`:

    ```csharp
    using System.Security.Claims;
    using FieldMark.Domain.ValueObjects;

    namespace FieldMark.Web.Authorization;

    /// <summary>
    /// The single .NET-side authorization decision call site (FR5).
    /// Epic 1: role-only checks (entity-scope rules deferred to Epic 2+).
    /// </summary>
    public static class DomainPolicies
    {
        // Action → roles permitted. Stories from Epic 2+ register their actions
        // by appending to this map at composition time (see RegisterAction).
        // Story 1.12 ships the map empty — there are no live actions in Epic 1.
        private static readonly Dictionary<string, HashSet<string>> ActionRoleMap = new();

        /// <summary>
        /// Register an action → permitted-roles mapping. Call from Program.cs
        /// or from a per-aggregate registrar (e.g., `ProjectPolicies.Register()`).
        /// </summary>
        public static void RegisterAction(string action, params Role[] roles)
        {
            if (!ActionRoleMap.TryGetValue(action, out var set))
            {
                set = new HashSet<string>();
                ActionRoleMap[action] = set;
            }
            foreach (var role in roles) set.Add(role.Name);
        }

        /// <summary>
        /// Return true if the user is authenticated and permitted to perform
        /// `action` (optionally scoped to `entityId`).
        /// </summary>
        public static bool Can(ClaimsPrincipal user, string action, Guid? entityId = null)
        {
            if (user.Identity is not { IsAuthenticated: true }) return false;
            if (!ActionRoleMap.TryGetValue(action, out var permittedRoles)) return false;

            foreach (var permittedRole in permittedRoles)
            {
                if (user.IsInRole(permittedRole))
                    return EvaluateEntityScope(action, entityId);
            }
            return false;
        }

        // Single extension point for Epic 2+ entity-scope rules (e.g.,
        // "Site Supervisor can act on a Violation only if assigned to it").
        // Today every action is role-coarse; flip individual entries to do
        // entity-scope work when that story arrives.
        private static bool EvaluateEntityScope(string action, Guid? entityId) => true;

        // Test-only escape hatch. Production callers must use RegisterAction.
        internal static void ResetForTests() => ActionRoleMap.Clear();
    }
    ```

  - [ ] 3.4 Create `FieldMark/FieldMark.Web/ViewModels/Components/ActionButtonVm.cs`:

    ```csharp
    namespace FieldMark.Web.ViewModels.Components;

    public sealed record ActionButtonVm(
        string Id,
        bool   Permission,
        bool   StateAllows,
        string Label,
        string HxPost,
        string HxTarget,
        string? DisabledReason = null
    );
    ```

  - [ ] 3.5 Create `FieldMark/FieldMark.Web/Pages/Shared/_ActionButton.cshtml`:

    ```razor
    @model FieldMark.Web.ViewModels.Components.ActionButtonVm
    @{
        // Trichotomy decision (UX-DR21, FR6).
        if (!Model.Permission) { return; }
    }
    @if (!Model.StateAllows)
    {
        <button
            type="button"
            id="@Model.Id"
            class="btn btn-secondary"
            disabled
            aria-disabled="true"
            tabindex="0"
            data-tooltip="@Model.DisabledReason"
            aria-describedby="@(Model.Id)-reason"
        >@Model.Label</button>
        <span id="@(Model.Id)-reason" class="sr-only">@Model.DisabledReason</span>
    }
    else
    {
        <button
            type="button"
            id="@Model.Id"
            class="btn btn-primary"
            hx-post="@Model.HxPost"
            hx-target="@Model.HxTarget"
            hx-swap="outerHTML"
            hx-disabled-elt="this"
        >@Model.Label</button>
    }
    ```

  - [ ] 3.6 Create `FieldMark.Tests.Domain/Authorization/CanTests.cs` covering the four `Can_*` cases from AC #7. Use a hand-built `ClaimsPrincipal` (no DI, no Identity scaffolding). Inside each test, call `DomainPolicies.ResetForTests()` before registering the test action, to prevent inter-test state bleed.
  - [ ] 3.7 Create `FieldMark.Tests.Integration/Components/ActionButtonRenderingTests.cs` covering the four ActionButton snapshot/parse cases from AC #7. Render the partial via `RazorEngineCore` or by spinning up the test host (`WebApplicationFactory<Program>`) and resolving `IRazorViewEngine` — pick whichever the existing integration suite already uses; if neither, prefer `WebApplicationFactory` since it lives in `FieldMark.Tests.Integration` already. Read the canonical example file via `File.ReadAllText` from a path computed off the test assembly location (`../../../../../../fieldmark_shared/components/action_button.example.html`) — this same pattern was established in Story 1.5/1.6 integration tests, follow the same conventions.

- [ ] Task 4: Django implementation — `Role`, `authz.can`, ActionButton include, tests (AC: #1, #2, #4, #5, #6, #7, #10)
  - [ ] 4.1 Create `fieldmark_py/fieldmark/roles.py`:

    ```python
    """Canonical conceptual-role names — single Django-side source of truth.

    Mirrored in dotnet_auth, django_auth, and fiber_auth. The string values
    are persisted as-is (Group.name on Django; ASPNetRole.NormalizedName on .NET;
    fiber_auth.user_roles.role on Go), so they must match across stacks.
    """

    from enum import Enum


    class Role(str, Enum):
        ADMIN              = "ADMIN"
        COMPLIANCE_OFFICER = "COMPLIANCE_OFFICER"
        INSPECTOR          = "INSPECTOR"
        SITE_SUPERVISOR    = "SITE_SUPERVISOR"
        EXECUTIVE          = "EXECUTIVE"

        def __str__(self) -> str:  # so f"{Role.ADMIN}" → "ADMIN", not "Role.ADMIN"
            return self.value
    ```

  - [ ] 4.2 Refactor `fieldmark_py/tools/management/commands/seed_groups.py` (Story 1.8) to import from `Role`:

    ```python
    from fieldmark.roles import Role

    CANONICAL_GROUPS = tuple(r.value for r in Role)
    ```

    Existing `tools/tests/test_seed_groups.py` continues to pass unchanged — `CANONICAL_GROUPS` still emits the same five strings.

  - [ ] 4.3 Create `fieldmark_py/fieldmark/authz.py`:

    ```python
    """Authorization decision primitive (FR5).

    `can(user, action, entity_id=None) -> bool` is the single Django-side call
    site every view uses to decide whether an action is permitted. Epic 1:
    role-only checks (entity-scope rules deferred to Epic 2+).
    """

    from __future__ import annotations

    import uuid
    from typing import Any

    from fieldmark.roles import Role

    # Action → frozenset of permitted role names. Epic 1 ships empty;
    # subsequent stories register their actions via register_action().
    _ACTION_ROLE_MAP: dict[str, frozenset[str]] = {}


    def register_action(action: str, *roles: Role) -> None:
        """Register an action → permitted-roles mapping. Call at app-load
        time (typically from an aggregate app's `apps.py:ready()` — except
        Django signals are banned, so call it from a module-level statement
        in the same package as the action's handler instead).
        """
        existing = _ACTION_ROLE_MAP.get(action, frozenset())
        _ACTION_ROLE_MAP[action] = existing | frozenset(r.value for r in roles)


    def can(user: Any, action: str, entity_id: uuid.UUID | None = None) -> bool:
        if not getattr(user, "is_authenticated", False):
            return False
        permitted = _ACTION_ROLE_MAP.get(action)
        if permitted is None:
            return False
        user_roles = set(user.groups.values_list("name", flat=True))
        if not (user_roles & permitted):
            return False
        return _evaluate_entity_scope(action, entity_id)


    def _evaluate_entity_scope(action: str, entity_id: uuid.UUID | None) -> bool:
        # Single extension point for Epic 2+ entity-scope rules. Today every
        # action is role-coarse; future stories swap this for per-action
        # entity-scope evaluators.
        return True


    # Test-only escape hatch. Production callers must use register_action().
    def _reset_for_tests() -> None:
        _ACTION_ROLE_MAP.clear()
    ```

  - [ ] 4.4 Create `fieldmark_py/templates/components/_action_button.html`:

    ```django
    {% comment %}
      ActionButton — affordance trichotomy (UX-DR10, UX-DR21).

      Inputs:
        - id: stable element id (string)
        - permission: bool (server-computed via authz.can)
        - state_allows: bool (server-computed from entity state)
        - label: string
        - hx_post: string (HTMX POST target)
        - hx_target: string (CSS selector)
        - disabled_reason: string (rendered in tooltip + sr-only span when disabled)
    {% endcomment %}
    {% if not permission %}{% else %}{% if not state_allows %}<button
      type="button"
      id="{{ id }}"
      class="btn btn-secondary"
      disabled
      aria-disabled="true"
      tabindex="0"
      data-tooltip="{{ disabled_reason }}"
      aria-describedby="{{ id }}-reason"
    >{{ label }}</button>
    <span id="{{ id }}-reason" class="sr-only">{{ disabled_reason }}</span>{% else %}<button
      type="button"
      id="{{ id }}"
      class="btn btn-primary"
      hx-post="{{ hx_post }}"
      hx-target="{{ hx_target }}"
      hx-swap="outerHTML"
      hx-disabled-elt="this"
    >{{ label }}</button>{% endif %}{% endif %}
    ```

    The collapsed whitespace shape (no leading newlines inside the `{% if/else %}` branches) matters for snapshot byte-identity — Django's default template trimming would otherwise inject newlines. Verify with a manual render before running the snapshot test.

  - [ ] 4.5 Create `fieldmark_py/fieldmark/tests/__init__.py` (empty) and `fieldmark_py/fieldmark/tests/test_authz.py` covering the four `can_*` cases from AC #7. Use `pytest`'s `monkeypatch` fixture to reset `_ACTION_ROLE_MAP` between tests via the module's `_reset_for_tests()` helper.
  - [ ] 4.6 Create `fieldmark_py/fieldmark/tests/test_action_button_template.py` covering the four ActionButton snapshot/parse cases. Render the template via Django's `render_to_string("components/_action_button.html", context)` and read the canonical example via `Path(__file__).resolve().parents[3] / "fieldmark_shared/components/action_button.example.html"`. Use `tests/normalize_html.py` (already established in Story 1.5/1.6) — do not author a fresh normalizer. Parse the disabled variant with `BeautifulSoup` (`bs4` is already in `pyproject.toml` dev deps as part of pytest-django's transitive tree; if not, add it as a test-only dependency).
  - [ ] 4.7 Add `fieldmark` to `pytest.ini`'s `testpaths` list (it currently lists the aggregate apps + `tools` from Story 1.8; the `fieldmark/` package itself was not previously test-pathed because it held no tests).

- [ ] Task 5: Go implementation — `Role`, `authz.Can`, ActionButton template, tests (AC: #1, #2, #4, #5, #6, #7, #10)
  - [ ] 5.1 Create `fieldmark-go/internal/domain/role.go`:

    ```go
    // Package domain holds entities and behaviour with zero outbound non-
    // stdlib imports. Role is the single Go-side source of truth for the
    // five canonical conceptual-role names; mirrored in dotnet_auth,
    // django_auth, and fiber_auth.user_roles.role (CHECK constraint).
    package domain

    type Role string

    const (
        RoleAdmin             Role = "ADMIN"
        RoleComplianceOfficer Role = "COMPLIANCE_OFFICER"
        RoleInspector         Role = "INSPECTOR"
        RoleSiteSupervisor    Role = "SITE_SUPERVISOR"
        RoleExecutive         Role = "EXECUTIVE"
    )

    // AllRoles enumerates the canonical names in deterministic order. Used
    // by seeders, tests, and any future fiber_auth.user_roles validator.
    var AllRoles = []Role{
        RoleAdmin, RoleComplianceOfficer, RoleInspector, RoleSiteSupervisor, RoleExecutive,
    }
    ```

  - [ ] 5.2 Refactor `fieldmark-go/internal/web/auth/lookup.go` and `fieldmark-go/internal/web/auth/stub.go` (Story 1.9) to reference `domain.Role*` constants where string literals appear (e.g., the alphabetical-min role result is still a `string` on the wire but the const can be typed when stored in `app.Actor.Role`). If `app.Actor.Role` was typed as `string` in Story 1.9 (it was), keep it `string` here to avoid a refactor cascade — but document in `internal/app/actor.go` that valid values are constrained to `domain.AllRoles` (CHECK enforced at DB layer; lookup returns one of those strings). Adding the typed `Role` is the architectural step; rewriting every consumer to use it is Epic 2's concern as actors flow into authz checks.
  - [ ] 5.3 Create `fieldmark-go/internal/web/auth/authz.go`:

    ```go
    package auth

    import (
        "sync"

        "github.com/google/uuid"

        "github.com/code-chimp/fieldmark-go/internal/app"
        "github.com/code-chimp/fieldmark-go/internal/domain"
    )

    // actionRoleMap holds the action → permitted-roles mapping. Stories from
    // Epic 2+ register their actions at composition time. Story 1.12 ships
    // empty — Epic 1 has no live actions.
    var (
        actionRoleMapMu sync.RWMutex
        actionRoleMap   = map[string]map[domain.Role]struct{}{}
    )

    // RegisterAction registers an action → permitted-roles mapping.
    func RegisterAction(action string, roles ...domain.Role) {
        actionRoleMapMu.Lock()
        defer actionRoleMapMu.Unlock()
        set, ok := actionRoleMap[action]
        if !ok {
            set = map[domain.Role]struct{}{}
            actionRoleMap[action] = set
        }
        for _, r := range roles {
            set[r] = struct{}{}
        }
    }

    // Can returns true if the actor is authenticated and permitted to perform
    // action (optionally scoped to entityID; uuid.Nil means "no entity scope").
    // Epic 1: role-only checks (entity-scope rules deferred).
    func Can(actor *app.Actor, action string, entityID uuid.UUID) bool {
        if actor == nil || actor.IsAnonymous() {
            return false
        }
        actionRoleMapMu.RLock()
        permitted, ok := actionRoleMap[action]
        actionRoleMapMu.RUnlock()
        if !ok {
            return false
        }
        if _, hit := permitted[domain.Role(actor.Role)]; !hit {
            return false
        }
        return evaluateEntityScope(action, entityID)
    }

    // Single extension point for Epic 2+ entity-scope rules.
    func evaluateEntityScope(action string, entityID uuid.UUID) bool { return true }

    // resetForTests clears the map. Test-only — internal package access only.
    func resetForTests() {
        actionRoleMapMu.Lock()
        defer actionRoleMapMu.Unlock()
        actionRoleMap = map[string]map[domain.Role]struct{}{}
    }
    ```

  - [ ] 5.4 Create `fieldmark-go/internal/web/viewmodels/action_button.go`:

    ```go
    package viewmodels

    // ActionButtonVM is the data context for the action_button template
    // (internal/web/templates/components/action_button.tmpl). Build it in
    // handlers (or in a per-aggregate to_vm helper) using auth.Can to
    // populate Permission and an entity-method-derived bool for StateAllows.
    type ActionButtonVM struct {
        ID             string
        Permission     bool
        StateAllows    bool
        Label          string
        HxPost         string
        HxTarget       string
        DisabledReason string
    }
    ```

  - [ ] 5.5 Create `fieldmark-go/internal/web/templates/components/action_button.tmpl`:

    ```gotmpl
    {{ define "action_button" -}}
    {{- if not .Permission -}}{{- else -}}{{- if not .StateAllows -}}<button
      type="button"
      id="{{ .ID }}"
      class="btn btn-secondary"
      disabled
      aria-disabled="true"
      tabindex="0"
      data-tooltip="{{ .DisabledReason }}"
      aria-describedby="{{ .ID }}-reason"
    >{{ .Label }}</button>
    <span id="{{ .ID }}-reason" class="sr-only">{{ .DisabledReason }}</span>{{- else -}}<button
      type="button"
      id="{{ .ID }}"
      class="btn btn-primary"
      hx-post="{{ .HxPost }}"
      hx-target="{{ .HxTarget }}"
      hx-swap="outerHTML"
      hx-disabled-elt="this"
    >{{ .Label }}</button>{{- end -}}{{- end -}}
    {{- end }}
    ```

    The `{{- ... -}}` whitespace-trim markers are required for snapshot byte-identity — Go templates inject newlines and whitespace around `{{ }}` actions by default.

  - [ ] 5.6 Register the components subdirectory with the template engine. In `fieldmark-go/cmd/web/main.go`, the existing `html.New("./internal/web/templates", ".html")` only loads `.html` files. Since `action_button.tmpl` uses `.tmpl`, **either** (a) name it `action_button.html` to match the engine's extension (preferred — keep one extension across the codebase), or (b) extend the engine to load both extensions. Go with (a): rename to `action_button.html`. Update the AC-#4 path and Task 5.5's snippet filename accordingly. (Note: `gofiber/template/html/v2` walks recursively, so the `components/` subdirectory is loaded automatically with no main.go changes.)
  - [ ] 5.7 Create `fieldmark-go/internal/web/auth/authz_test.go` covering the four `Can_*` cases (note: file is colocated with `authz.go` so the unexported `resetForTests` is accessible — this is the standard library posture; no testify, no mocks, no DB). Inputs use `app.Anonymous()` and hand-constructed `&app.Actor{ID: ..., Username: "...", Role: "ADMIN"}` actors.
  - [ ] 5.8 Create `fieldmark-go/internal/web/templates/components/action_button_test.go` (package `components_test` or `components`, both acceptable — match the file's location). Render the template via a standalone `html/template` parse + `Execute` against a freshly-constructed `template.New("action_button").Parse(...)` (cleaner than spinning up Fiber for this unit test). For accessibility parsing of the disabled variant, use `golang.org/x/net/html` — it is **not** currently a Go dependency; add it via `go get golang.org/x/net/html` and confirm `go mod tidy` is clean. Read the canonical example via `os.ReadFile` with a path computed relative to the test file (`../../../../../fieldmark_shared/components/action_button.example.html`).

- [ ] Task 6: Cross-stack snapshot byte-identity verification (AC: #4)
  - [ ] 6.1 Each per-stack snapshot test reads the *same* `fieldmark_shared/components/action_button.example.html` and uses its *own* stack's `normalize_html` helper. The expected property: after normalization, each stack's helper output for the canonical fixture equals the corresponding variant block from the example file. This is asserted independently three times (once per stack); there is no central "cross-stack diff" runner at this story — the cross-stack guarantee comes from all three stacks comparing against the same canonical file with the same normalization rules.
  - [ ] 6.2 Manually verify cross-stack identity once during dev (this is **not** an automated step in Story 1.12; it is a hand-check to catch silent normalizer divergence). From each stack, write the rendered ActionButton output (disabled and present variants) for the canonical fixture to `/tmp/`, then `diff` them. They must match exactly post-normalization (the diff tool may show trivial whitespace differences pre-normalization; that is acceptable as long as each stack's snapshot test passes). Roll back the temp files before commit.
  - [ ] 6.3 If the three stack outputs diverge in a way that cannot be reconciled by adjusting `_action_button.html` / `_ActionButton.cshtml` / `action_button.html` template syntax (i.e., the divergence is a Basecoat class-name dispute or an attribute-order preference), file it as a defect rather than papering over with normalizer special-casing. The normalizer is the same across stacks; divergent inputs are the bug.

- [ ] Task 7: Verify no live call sites and parity (AC: #8, #9)
  - [ ] 7.1 Run the three `grep` commands from AC #8. Each must return zero matches (excluding the component files themselves and the test files that reference them).
  - [ ] 7.2 From repo root: `make parity` — exits 0.
  - [ ] 7.3 From each stack root, capture the `--dump-routes` output and `diff` against HEAD-before-this-story. Zero diff on all three.
  - [ ] 7.4 Confirm `tools/parity/canonical-pg-indexes.txt` is unchanged in the diff.

- [ ] Task 8: Update each stack's `CLAUDE.md` (AC: #10)
  - [ ] 8.1 **`FieldMark/CLAUDE.md`** — add a `## Authorization` section after `## Authentication`. Cover:
    - The single .NET-side `Can` call site lives at `FieldMark.Web/Authorization/DomainPolicies.cs` with signature `Can(ClaimsPrincipal user, string action, Guid? entityId = null) : bool`. Handlers and Razor view-models call it; templates do not.
    - Role names live in `FieldMark.Domain/ValueObjects/Role.cs`. Hard-coded role-name string literals anywhere else are a defect.
    - Actions are registered via `DomainPolicies.RegisterAction(action, roles...)`. Epic 2+ stories register their actions at composition time (in `Program.cs` or in a per-aggregate `<Aggregate>Policies.Register()` helper). Story 1.12 ships the map empty.
    - The ActionButton partial lives at `Pages/Shared/_ActionButton.cshtml` with VM at `ViewModels/Components/ActionButtonVm.cs`. The trichotomy decision is made by the partial; the caller supplies pre-computed `permission` (from `DomainPolicies.Can`) and `state_allows` (from the entity's `can_*` predicate). Templates never call `Can` directly.
    - Cross-references `fieldmark_shared/components/action_button.example.html` as the canonical snapshot target.

  - [ ] 8.2 **`fieldmark_py/CLAUDE.md`** — add a `## Authorization` section after `## Authentication`. Same shape, Python-flavored: `fieldmark/authz.py:can(user, action, entity_id=None)`; `Role` in `fieldmark/roles.py`; `register_action(action, *roles)` at module-load time (no signals); ActionButton at `templates/components/_action_button.html` consumes pre-computed booleans.
  - [ ] 8.3 **`fieldmark-go/CLAUDE.md`** — add a `## Authorization` section after `## Authentication`. Same shape, Go-flavored: `internal/web/auth/authz.go:Can(actor, action, entityID)`; `domain.Role` const set in `internal/domain/role.go`; `RegisterAction(action, roles...)` at composition (typically a per-aggregate `init()` in `internal/web/handlers/` or an explicit call from `cmd/web/main.go`); ActionButton at `internal/web/templates/components/action_button.html`.

- [ ] Task 9: Verify all gates green (AC: #11)
  - [ ] 9.1 **.NET:** `cd FieldMark && dotnet csharpier format . && dotnet build && dotnet test` — all green; `dotnet csharpier check .` reports zero diffs.
  - [ ] 9.2 **Django:** `cd fieldmark_py && uv run ruff check . && uv run mypy . && uv run pytest` — all green.
  - [ ] 9.3 **Go:** `cd fieldmark-go && make check` — all green (`fmt-check` + `vet` + `staticcheck` + `test`).
  - [ ] 9.4 From repo root: `make parity` — exits 0.

## Dev Notes

### Brownfield posture — what exists today (read before writing anything)

Cross-stack state at HEAD of branch `feature/1.4_bootstrap-design-system`:

- **Stories landed:** 1.1, 1.2, 1.3 are `done`. Story 1.4 (design system) is `review`. Stories 1.5–1.11 are `ready-for-dev` — their implementation may or may not have started in parallel; this story should be authored against what's in the planning artifacts plus what the merged stories produced, not against an assumption that 1.5–1.11 have already shipped. If any of 1.7 / 1.8 / 1.9 has merged when Story 1.12 begins, Tasks 3.2 / 4.2 / 5.2 refactor what they shipped. If they have not yet merged, those tasks still apply — they introduce the canonical Role type and back-port the refactor as the merge proceeds.
- **`fieldmark_shared/components/` does not exist.** Story 1.4 (review) populates `src/` and `dist/`. This story creates `components/` as a sibling for canonical example HTML — first usage is `action_button.example.html`. Stories 1.5–1.11's UX-DR-named components (StatusBadge, ComplianceTile, etc.) will each follow the same pattern in their own stories.
- **No aggregate domain entities exist yet.** `FieldMark/FieldMark.Domain/Entities/` has only `EntityClass.cs` (scaffold placeholder); `FieldMark/FieldMark.Domain/ValueObjects/` has only `ValueObjClass.cs`. The five domain entities (Project, Inspection, Violation, CorrectiveAction, AuditEntry) land in Epics 2–5. Story 1.12 introduces the first real `ValueObjects/Role.cs` file. (You may delete `ValueObjClass.cs` and `EntityClass.cs` once Role lands and compiles — they are dead scaffolds. Treat that deletion as in-scope cleanup for this story.)
- **Django app `fieldmark/` exists** (the project package) with `settings.py`, `urls.py`, and now `authz.py` and `roles.py` are added by Story 1.12. No `fieldmark/tests/` directory exists yet — this story creates it. The aggregate apps (`projects/`, `inspections/`, …) have empty `models.py` / `views.py` / `migrations/` (Epic 2+).
- **Go `internal/domain/` is empty** (per `fieldmark-go/internal/domain/` — only the directory exists). `internal/web/templates/` has `layouts/`, `pages/`, `fragments/`, `partials/` but no `components/` subdirectory yet. Story 1.12 creates `internal/web/templates/components/`.
- **The `Can` primitive is the canonical handler call site per the architecture's Canonical Request Flow** (Step 1, Authorize). Every code stub in the architecture's "Approve a corrective action" example uses `Can`: `_authz.Can(User, "violation.approve_resolution", id)` (.NET), `authz.can(request.user, "violation.approve_resolution", violation_id)` (Django), `h.authz.Can(actor, "violation.approve_resolution", violationID)` (Go). Story 1.12 is the first story that makes those call sites compilable — until now they were illustrative only.

### Why the helper takes `permission` (a bool), not `(user, action)`

The architecture's "Action-button rendering pattern" makes this explicit: view models carry computed `can_*` booleans. Templates conditionally render based on those booleans. Computed at render time, in the same handler that returns the partial. **Never on the client.**

This separates two concerns:

1. **`Can(user, action, entityId?)`** — the authorization decision. Called *once* in the handler, *outside* the template. Returns `bool`.
2. **`ActionButton(permission, state_allows, ...)`** — the rendering helper. Takes the already-decided permission bool plus a state-allows bool, and renders one of three things. Has no authorization knowledge.

The trichotomy lives in (2). The permission decision lives in (1). The mistake the trichotomy prevents — "user clicks a button they shouldn't have seen" — depends on (1) being honest; the mistake (2) prevents is "two stacks render the same trichotomy differently."

Templates calling `Can` directly is forbidden because (a) it puts authorization logic in the view layer where reviewers don't look for it, (b) it makes the call shape per-stack-idiomatic (`@authz.Can(...)`-style Razor function vs. `{% if authz.can(...) %}`-style Django tag vs. Go `{{ if authz.Can ... }}` function map) which directly violates FR58 cross-stack identity. Keep `Can` in the handler.

### Why a static `DomainPolicies` (not a registered `IAuthorizationService`)

ASP.NET Core's native `IAuthorizationService.AuthorizeAsync(user, resource, policyName)` is the framework idiom for entity-scoped authorization. The AC at line 771 of epics.md prescribes a different shape: `DomainPolicies.Can(...)`. Two reasons we honour the AC's static-call shape:

1. **Cross-stack call site identity (FR58).** A static call `DomainPolicies.Can(...)` mirrors Django `authz.can(...)` and Go `authz.Can(...)` more directly than `await _authzService.AuthorizeAsync(...)` does. The architecture's canonical request flow has the same `Can(...)` shape across all three stacks; using `IAuthorizationService` on .NET only would break that visual identity at every handler call site.
2. **No async needed.** Role-only checks are synchronous in-memory dictionary lookups. Wrapping in async machinery to please `IAuthorizationService`'s shape adds friction without value.

When entity-scope rules arrive (Epic 2+), the `EvaluateEntityScope(action, entityId)` extension point is where the typically-async DB-touching work lands. At that point the .NET implementation may grow to internally call a registered `IAuthorizationHandler` — but the *public* call shape (`DomainPolicies.Can(user, action, entityId)`) stays static and cross-stack-identical.

If a future story needs the ASP.NET Core authorization pipeline (e.g., `[Authorize(Policy = "...")]` attribute on a Razor Page), `DomainPolicies.Register(IServiceCollection)` can be added to wire policies whose handlers internally delegate to `Can`. That wiring is not in scope for Story 1.12 — Epic 1 has no protected handlers.

### Why the `disabled` button needs both `data-tooltip` AND a sibling `sr-only` span

Basecoat 0.3.11's `[data-tooltip]` implementation is a CSS-only `::before`/`::after` pseudo-element pattern. The tooltip text is in `attr(data-tooltip)`; the visual rendering is via pseudo-elements. Pseudo-element content is **inconsistently exposed** to assistive technology — Safari + VoiceOver does fine, but many screen reader / browser combinations (NVDA + Firefox in particular, common in WCAG audits) ignore `::before`-generated content.

The UX-DR21 / WCAG 2.1 AA invariant — "every disabled control carries an explanation" — must be honoured for screen-reader users too, not just hover-state visual users. The fix is to duplicate the reason text into a real DOM node (`<span class="sr-only">`) that is:
- visually hidden by the `.sr-only` Basecoat utility (clip-path / clip / absolute positioning),
- linked from the `<button>` via `aria-describedby`,
- therefore announced reliably by every screen reader as part of the button's accessible description.

The `data-tooltip` attribute remains because it is the visual UX (hover affordance) and because Basecoat upgrades may begin exposing tooltip text to AT in the future — keeping the attribute means we automatically benefit. The `sr-only` span is the durable correctness path; the tooltip is the visual UX layer.

**Do not** rely on `title="<reason>"` instead — `title` is keyboard-inaccessible and inconsistent across browsers. We avoid it project-wide.

**Do not** add `tabindex="-1"` to the disabled button. UX-DR21 explicitly says the disabled button retains its place in the tab order so a keyboard user can navigate to it and hear the reason. `tabindex="0"` on a `disabled` button is the documented Basecoat-disabled-affordance pattern — `disabled` removes the default focusability; `tabindex="0"` restores it.

### Why `Role.All` / `Role.AllRoles` / `Role` enum is the right consolidation point

Stories 1.7, 1.8, and 1.9 each carry their own copy of the canonical role-name array:
- .NET: `private static readonly string[] CanonicalRoles` in `RoleSeeder.cs`.
- Django: `CANONICAL_GROUPS = (...)` tuple in `seed_groups.py`.
- Go: `CONSTRAINT user_roles_role_check CHECK (role IN ('ADMIN', ...))` in `001_initial.sql` (DB-layer) + implicit string-literal contract in `lookup.go`.

Three copies, three idioms — predicted churn the moment a role name changes (or a sixth role is added). Story 1.12 consolidates each stack to one location:
- .NET: `Role.All` (the seeder imports it).
- Django: `Role` enum (the seeder iterates it).
- Go: `domain.AllRoles` (no current seeder reads it directly because Story 1.10's seeder is per-stack and the Go seeder reads from the manifest; but `domain.Role` consts make all in-Go references consistent).

The DB-layer CHECK constraint in `fiber_auth.user_roles.role` (Story 1.9) stays as-is and is the authoritative gate at write time — `domain.Role` is the Go-side type that matches what passes the CHECK.

### How `register_action` + `ActionRoleMap` evolves in Epic 2+

Story 1.12 ships the map empty. Epic 2's first story (Project domain methods) will add actions like:

```csharp
// .NET — somewhere in Program.cs or a ProjectPolicies.Register() helper
DomainPolicies.RegisterAction("project.close",
    Role.Admin, Role.ComplianceOfficer, Role.SiteSupervisor);
DomainPolicies.RegisterAction("project.place_on_hold",
    Role.Admin, Role.ComplianceOfficer);
```

```python
# Django — top-level in fieldmark/policies.py or per-app policies.py
from fieldmark.authz import register_action
from fieldmark.roles import Role

register_action("project.close", Role.ADMIN, Role.COMPLIANCE_OFFICER, Role.SITE_SUPERVISOR)
```

```go
// Go — in handler-package init() or main.go composition
auth.RegisterAction("project.close",
    domain.RoleAdmin, domain.RoleComplianceOfficer, domain.RoleSiteSupervisor)
```

Each action string is registered exactly once per stack (a defect to register the same action twice with different role sets). The action strings themselves are part of the cross-stack canonical inventory (FR58) — they must be byte-identical across stacks. Future stories add their action strings to `_bmad-output/planning-artifacts/architecture.md`'s "Canonical inventory" section (alongside audit action strings) — that's an Epic 2 concern, not Story 1.12's.

### Anti-patterns that must NOT slip in

- ❌ Calling `Can` from a template (Razor `@DomainPolicies.Can(...)`, Django `{% if authz.can ... %}`, Go `{{ if authz.Can ... }}`). The decision is made in the handler; the template renders booleans. Cross-stack FR58 identity depends on this.
- ❌ Hard-coded role-name string literals anywhere except `Role.cs` / `roles.py` / `role.go` and the canonical `fiber_auth.user_roles.role` CHECK constraint. After Tasks 3.2 / 4.2 / 5.2, `RoleSeeder.cs` / `seed_groups.py` reference `Role.All` / `Role`. Adding a new literal `"ADMIN"` elsewhere is a defect — search for it in code review.
- ❌ Rendering the disabled variant without `aria-describedby` + a real DOM `sr-only` reason span. Per UX-DR21 + WCAG 2.1 AA, pseudo-element-only tooltip text is not an accessible reason. The dev review pass on Story 1.4 explicitly deferred the related "tooltip clips silently" issue — this story does NOT regress the accessibility surface.
- ❌ Adding `title="<reason>"` to the disabled button as a fallback. Keyboard-inaccessible; inconsistent AT support.
- ❌ Using `tabindex="-1"` on the disabled button. Removes it from the keyboard tab order; defeats UX-DR21.
- ❌ Adding `[Authorize(Roles = "...")]` attributes or `@login_required`/`@user_passes_test` decorators to Razor Pages / Django views in this story. Story 1.11 wires `app.UseAuthentication()` / `LoginRequiredMiddleware` / `auth.RequireAuth()`; Story 1.12 ships the `Can` primitive that handlers will *call* once those handlers exist. No existing handler should grow an `[Authorize]` attribute as a side effect of this story.
- ❌ Wiring the ActionButton helper into any existing template (Home page, Privacy page, dashboard fragment). AC #8 forbids live call sites in Epic 1; the primitive exists for Epic 2 onward.
- ❌ Modifying any audit-action-string list, route inventory, or `pg_indexes` snapshot. Story 1.12 changes none of those.
- ❌ Adding a new dependency to `fieldmark_shared/package.json`. The `components/` directory is static example HTML; no build step is added.
- ❌ Replacing Basecoat's `[data-tooltip]` with a JS tooltip library (Floating UI, Tippy, etc.). Project rule: Basecoat is the component vocabulary; the visible-tooltip surface is Basecoat's CSS. Accessibility is layered on via `sr-only`, not by swapping libraries.
- ❌ Using `inclusion_tag` on the Django side, or a `partial` view-component on .NET, or `template.FuncMap` on Go, to "abstract over the trichotomy". The trichotomy IS the helper. The template body does the if-elif-else. No additional indirection layer.
- ❌ Async `Can` (`Task<bool> Can(...)`). Synchronous in-memory dictionary lookup. Async machinery does not earn its weight here; entity-scope work lands inside `EvaluateEntityScope` which can introduce its own narrow async path if needed.
- ❌ Returning `false` for `Can(authenticated_user, action, entity)` when the action is in the map but the role check fails *for a specific user* — and silently swapping that to "absent" in the template — when the contract is "the user *would not be permitted at all*, so the button is absent." Conversely, when the user *would be permitted but the entity state blocks*, the contract is "disabled-with-tooltip." If you find yourself making `Can` aware of state-allows, stop — those are two booleans, not one.
- ❌ Calling `Can` inside a `Models/View` initializer that has no access to a `ClaimsPrincipal` / `HttpRequest` / `fiber.Ctx`. The decision belongs in the handler (where the user is known) — the view model carries the *result*. If you cannot construct the view model without the user, the answer is to pass the user to the view-model-construction call, not to push `Can` deeper.

### Project Structure Notes

Files this story adds:

- **Shared:** `fieldmark_shared/components/README.md`, `fieldmark_shared/components/action_button.example.html`
- **.NET (new):** `FieldMark/FieldMark.Domain/ValueObjects/Role.cs`, `FieldMark/FieldMark.Web/Authorization/DomainPolicies.cs`, `FieldMark/FieldMark.Web/ViewModels/Components/ActionButtonVm.cs`, `FieldMark/FieldMark.Web/Pages/Shared/_ActionButton.cshtml`, `FieldMark.Tests.Domain/Authorization/CanTests.cs`, `FieldMark.Tests.Integration/Components/ActionButtonRenderingTests.cs`
- **Django (new):** `fieldmark_py/fieldmark/roles.py`, `fieldmark_py/fieldmark/authz.py`, `fieldmark_py/templates/components/_action_button.html`, `fieldmark_py/fieldmark/tests/__init__.py`, `fieldmark_py/fieldmark/tests/test_authz.py`, `fieldmark_py/fieldmark/tests/test_action_button_template.py`
- **Go (new):** `fieldmark-go/internal/domain/role.go`, `fieldmark-go/internal/web/auth/authz.go`, `fieldmark-go/internal/web/viewmodels/action_button.go`, `fieldmark-go/internal/web/templates/components/action_button.html`, `fieldmark-go/internal/web/auth/authz_test.go`, `fieldmark-go/internal/web/templates/components/action_button_test.go`

Files this story updates:

- **.NET (update):** `FieldMark/FieldMark.Web/SeedData/RoleSeeder.cs` (read role names from `Role.All`), `FieldMark/CLAUDE.md` (add `## Authorization` section). Delete the scaffold stubs `FieldMark.Domain/Entities/EntityClass.cs` and `FieldMark.Domain/ValueObjects/ValueObjClass.cs` (the latter is now replaced by `Role.cs`; both were placeholders).
- **Django (update):** `fieldmark_py/tools/management/commands/seed_groups.py` (import `Role`), `fieldmark_py/pytest.ini` (add `fieldmark` to `testpaths`), `fieldmark_py/CLAUDE.md` (add `## Authorization` section).
- **Go (update):** `fieldmark-go/internal/web/auth/lookup.go` and `fieldmark-go/internal/web/auth/stub.go` (light: reference `domain.Role*` consts where the canonical names appear), `fieldmark-go/go.mod` / `go.sum` (`go get golang.org/x/net/html` for the HTML-parser test dep), `fieldmark-go/CLAUDE.md` (add `## Authorization` section).

All file locations align with [Architecture Repository Directory Structure](_bmad-output/planning-artifacts/architecture.md#complete-repository-directory-structure):

- `FieldMark.Domain/ValueObjects/Role.cs` ← architecture line 1027.
- `FieldMark.Web/Authorization/DomainPolicies.cs` ← architecture line 1050.
- `fieldmark/authz.py` ← architecture line 1123.
- `internal/web/auth/` (the Go authz module is a sibling of `internal/web/auth/stub.go` from Story 1.9; the architecture's reference to authz inside `internal/app/` (line 1266) refers to the `Deps` wiring point, not the `Can` function — Story 1.12's `Can` lives in `internal/web/auth/` because it operates on `*app.Actor` which the web layer already produces).
- `internal/web/templates/components/` ← architecture mentions `internal/web/templates/` (line 1218); the `components/` subdirectory is new in this story and parallels the cross-stack canonical pattern.

### Testing Standards

Per [Architecture Testing](_bmad-output/planning-artifacts/architecture.md) and each stack's `CLAUDE.md`:

- **.NET:** Unit tests in `FieldMark.Tests.Domain/` for `Can` (pure-logic; hand-built `ClaimsPrincipal`; no DB). Integration tests in `FieldMark.Tests.Integration/` for the Razor partial rendering (uses `WebApplicationFactory` or a comparable in-process render harness — match what Story 1.5 / 1.6 already established). No SQLite.
- **Django:** `@pytest.mark.django_db` for the `authz.can` tests (they touch `user.groups.values_list(...)` which is ORM — real Postgres via `pytest-django`). Template tests use `django.test.utils.override_settings` + `render_to_string` and do not need `@pytest.mark.django_db`.
- **Go:** Standard library `testing` only — no testify, no gomock. Tests in `authz_test.go` are pure-logic (in-memory map). Template tests parse the `html/template` directly and use `golang.org/x/net/html` for accessible-attributes parsing.
- **All stacks:** Reuse the existing `normalize_html` helper established in Stories 1.5 / 1.6 — do **not** author a fourth normalizer. If a helper does not yet exist on a given stack at this point in the branch, Story 1.5's PR established the canonical algorithm; copy the implementation, do not invent a new one.
- **No Playwright in this story.** The canonical component example gallery and per-stack snapshot tests cover the cross-stack identity surface; Playwright comes online in Epic 7 (or earlier as anchor screens land).

### Previous Story Intelligence

**Story 1.7 (.NET — likely `review` or `done` when 1.12 starts).** Lessons that transfer directly:
- Idempotent identifier list pattern (Story 1.7: `RoleManager.RoleExistsAsync` → create if missing). Story 1.12 inherits the *role names* from 1.7 — consolidating them to `Role.All` is the explicit Story 1.7 / 1.8 / 1.9 deferral coming due.
- `AddIdentityCore` (not `AddDefaultIdentity`) is what keeps `/identity/*` routes off the parity inventory. Story 1.12 adds zero routes and must not regress this.
- `TreatWarningsAsErrors=true` is enforced via `Directory.Build.props`. Watch for `CS86xx` warnings from new auth code (the `ClaimsPrincipal` parameter on `Can` is non-null; mark it `[NotNull]` if the analyzer wants it — but if a vanilla `ClaimsPrincipal user` compiles clean, leave it).
- CSharpier is the formatter; run `dotnet csharpier format .` before commit.

**Story 1.8 (Django — likely `review` or `done`).** Lessons:
- Schema-isolation mechanism is `search_path` on `OPTIONS`; not a router. This story does not touch DB plumbing.
- `Group.objects.get_or_create(name=...)` is the idempotent path; Story 1.12 does not duplicate that — the role-Group seeding from 1.8 stays as-is; only the source of the role-name list changes.
- "Real Postgres in tests" — the `authz.can` test uses `user.groups`, which is ORM; the `@pytest.mark.django_db` decorator is required and `pytest-django` handles the connection.
- `dump_routes` is unchanged; verify the route inventory post-1.12.

**Story 1.9 (Go — likely `review` or `done`).** Lessons:
- `pgxpool.Pool` everywhere (Story 1.9 migrated from `pgx.Conn`). This story does not touch `db.go`.
- `*app.Actor` is the canonical principal type; `auth.ActorFromCtx(c)` is the canonical extractor. `Can` takes `*app.Actor`, not a raw username string.
- `internal/app/` stays Fiber-free (`fiber.Ctx must not escape internal/web`). `Can` lives in `internal/web/auth/` for the same reason.
- Standard library `testing` only.

**Story 1.10 (dev-user manifest seeding).** When merged, it populates `dotnet_auth.users` / `django_auth.auth_user` (+ dev-uuid side table) / `fiber_auth.users` + `fiber_auth.user_roles` from `docker/postgres/init/seed-uuids/dev-users.json`. The role names in that manifest must match `Role.All` — Story 1.12 is the consolidation point that makes "do the manifest names match the canonical Role list?" a one-place check.

**Story 1.11 (login / logout).** When merged, `app.UseAuthentication()` / `LoginRequiredMiddleware` / `auth.RequireAuth()` are wired. After 1.11 + 1.12, the next Epic-1 story is 1.13 (role-aware home page) — which is the *first* story that actually calls `Can`. Story 1.12's helpers exist to be called by 1.13 and Epic 2+.

### Git Intelligence

Recent commits (most relevant to this story):
- `d03f0fe feat: e1s3 establish tools parity` — established the `--dump-routes` invariants and `make parity`. Story 1.12 preserves both (zero new routes).
- `cbf47e9 feat: e1s2 verfied sql init scripts` — confirmed `fiber_auth.user_roles.role` CHECK constraint. The Go `domain.Role` constants must mirror this CHECK literally.
- `a6fac88 feat: e1s1 confirm scaffolds` — base scaffolds across all three stacks. Story 1.12 introduces the first real Domain/ValueObjects/Role code on .NET and Go, and the first real `fieldmark/authz.py` on Django.
- `a4fcc76 task: complete BMAD planning phase` — planning artifacts pinned at this commit. AC and Dev Notes cite from these.

No prior commit has introduced a typed `Role` value object or an `authz.Can` primitive on any stack. Story 1.12 is the first.

### Latest Technical Information

- **.NET 10 / EF Core 10.0.7** in use. `ClaimsPrincipal` and `IsInRole(string)` are stable since .NET Framework 4.5; no version concerns. `HashSet<T>` membership is O(1).
- **Django 6.0.4 / Python 3.14+** in use. `Enum` (`str, Enum` mixin) is stable since 3.11; `frozenset` is stdlib. No new deps. `bs4` (BeautifulSoup) is **already** in `pyproject.toml` dev deps (`beautifulsoup4` — confirm with `uv tree | grep beautifulsoup`; if absent, add it as a test-only dep via `uv add --dev beautifulsoup4`).
- **Go 1.26.2 / Fiber v3.2.0 / pgx v5.9.2** in use. `sync.RWMutex` on the `actionRoleMap` is needed because Fiber goroutines may read concurrently (writes happen at startup or in tests, so the RW split is correct). `golang.org/x/net/html` is **not** currently in `go.mod` — Task 5.8 adds it. After `go get golang.org/x/net/html`, run `go mod tidy` and verify no other indirect deps shift.
- **Basecoat 0.3.11** is the pinned design-system version. `[data-tooltip]` and `.sr-only` and `.btn`/`.btn-primary`/`.btn-secondary` are all present in the `dist/fieldmark.css` Story 1.4 produces. No new CSS is required by Story 1.12.
- **HTMX 2.x** (the vendor version pinned in `fieldmark_shared/vendor/htmx/`). `hx-post`, `hx-target`, `hx-swap`, `hx-disabled-elt` are all HTMX 2 attributes with stable semantics. `hx-swap="outerHTML"` replaces the targeted element with the response body — the canonical pattern for a state-changing action that swaps the originating partial (matching the architecture's canonical request flow Step 8 + UX-DR21 + Pattern 1 three-region round trip).

### References

- [Architecture — Authorization expression pattern](_bmad-output/planning-artifacts/architecture.md#process-patterns) — single `Can` call site; templates render computed booleans, never call `Can`.
- [Architecture — Action-button rendering pattern](_bmad-output/planning-artifacts/architecture.md#process-patterns) — `can_*` boolean view-model fields; trichotomy (absent / disabled / enabled).
- [Architecture — Repository Directory Structure](_bmad-output/planning-artifacts/architecture.md#complete-repository-directory-structure) — `Role.cs`, `DomainPolicies.cs`, `fieldmark/authz.py`, `internal/web/auth/` locations.
- [Architecture — Architectural Boundaries → Authentication / authorization](_bmad-output/planning-artifacts/architecture.md#architectural-boundaries) — opaque UUID refs; per-stack-idiomatic auth implementation; single `Can` call site.
- [Architecture — Canonical Request Flow code stubs](_bmad-output/planning-artifacts/architecture.md#process-patterns) — the three `_authz.Can` / `authz.can` / `h.authz.Can` shapes Story 1.12 makes compilable.
- [PRD FR5, FR6](_bmad-output/planning-artifacts/prd/functional-requirements.md) — Authorization decision primitive; server-decided action-button absence.
- [PRD FR58, FR59](_bmad-output/planning-artifacts/prd/functional-requirements.md) — Cross-stack identical routes, method names; identical observable behavior across stacks.
- [PRD FR64](_bmad-output/planning-artifacts/prd/functional-requirements.md) — Buttons that trigger HTMX requests indicate disabled state during request (`hx-disabled-elt`).
- [PRD FR60–FR63](_bmad-output/planning-artifacts/prd/functional-requirements.md) — Accessibility (keyboard, ARIA, focus, aria-live) — applies to the disabled-button surface this story introduces.
- [UX spec — ActionButton component](_bmad-output/planning-artifacts/ux-design-specification.md) — Props, anatomy, states, accessibility, cross-stack invariant.
- [UX spec — Pattern 2: Affordance Trichotomy](_bmad-output/planning-artifacts/ux-design-specification.md) — Rule, prohibition, used-by, verification.
- [UX spec — Canonical examples in `fieldmark_shared/components/`](_bmad-output/planning-artifacts/ux-design-specification.md) — Storybook-style static HTML example per component.
- [Epic 1 Story 1.12](_bmad-output/planning-artifacts/epics.md) — AC source (epics.md is canonical for stories per the workflow).
- [docs/hard-rules.md](docs/hard-rules.md) — Backend authority, no service layers, real PostgreSQL in tests, stack symmetry.
- [FieldMark/CLAUDE.md](FieldMark/CLAUDE.md) — `FieldMark.Domain` has zero outbound references (Role must compile in a Domain with zero deps).
- [fieldmark_py/CLAUDE.md](fieldmark_py/CLAUDE.md) — No signals; no business logic in views/forms/managers/middleware.
- [fieldmark-go/CLAUDE.md](fieldmark-go/CLAUDE.md) — `fiber.Ctx` stays in `internal/web/`; `internal/domain/` has zero non-stdlib imports.
- [Story 1.7 implementation artifact](_bmad-output/implementation-artifacts/1-7-wire-asp-net-core-identity-to-dotnet-auth-schema-with-conceptual-roles.md) — Role-name strings live in seeder "until Story 1.12"; this story consolidates them.
- [Story 1.8 implementation artifact](_bmad-output/implementation-artifacts/1-8-wire-django-built-in-auth-to-django-auth-schema-with-conceptual-role-groups.md) — Same deferral on Django; this story consolidates.
- [Story 1.9 implementation artifact](_bmad-output/implementation-artifacts/1-9-implement-go-fiber-stub-authentication-middleware.md) — Same deferral on Go; this story consolidates; `*app.Actor` shape is the input to `Can`.
- [Story 1.4 implementation artifact](_bmad-output/implementation-artifacts/1-4-bootstrap-design-system-foundation-in-fieldmark-shared.md) — Basecoat 0.3.11 is the design-system version; `.btn`, `.btn-primary`, `.btn-secondary`, `.sr-only`, `[data-tooltip]` styling all come from this story.

## Dev Agent Record

### Agent Model Used

_(populated by dev agent)_

### Debug Log References

### Completion Notes List

### File List
