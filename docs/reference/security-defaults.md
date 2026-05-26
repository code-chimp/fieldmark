# Security Defaults

Per-stack defensive defaults that must be in place whenever code handles user input, cookies, redirects, or filesystem writes. Each item has a canonical resolution proven in Epic 1's review history — implement that resolution unless there's a documented reason to deviate.

This list exists because Epic 1 stories shipped functionally-correct code that omitted standard framework safeguards, which the reviewer then caught one-by-one. Front-loading them as a checklist removes an entire class of review findings.

---

## 1. Open-redirect on return-target parameters

**Failure mode:** Login/auth-required redirect accepts a `ReturnUrl` / `next` / `return_url` parameter without verifying it's a local path. Attacker can craft `…/login?ReturnUrl=https://evil.example/phish` and harvest credentials after the login.

**Canonical resolution per stack:**
- **.NET** — `Url.IsLocalUrl(returnUrl)` guard on every read; fall back to `/` for non-local targets. Apply on both GET (initial render of the form's hidden input) and POST (the redirect after sign-in succeeds).
- **Django** — `django.utils.http.url_has_allowed_host_and_scheme(url, allowed_hosts={request.get_host()}, require_https=request.is_secure())` guard. Use `django.contrib.auth`'s built-in `LoginView` whenever possible — it does this for you.
- **Go** — explicit check that the parsed URL has empty `Scheme` and `Host`, or matches the current request host. Reject absolute URLs from query params.

**Reference:** Story 1.11 re-review patch on .NET `ReturnUrl` round-trip.

---

## 2. Cookie attribute defaults

**Failure mode:** Cookies set without explicit `SameSite`, `Secure`, `HttpOnly`, and `Path` attributes drift across stacks; cross-site request risk; client JS reads cookies that should be HTTP-only (or can't read cookies that need to be client-readable).

**Canonical resolution:** every `Set-Cookie` in any handler must declare all five attributes intentionally. Document the choice when `HttpOnly` or `Secure` is intentionally omitted (e.g., theme cookie deliberately readable by client JS — see Story 1.6).

| Cookie purpose | SameSite | Secure | HttpOnly | Path | Max-Age |
|---|---|---|---|---|---|
| Session / auth | `Lax` | yes (prod) | **yes** | `/` | per session lifetime |
| Client-readable preference (theme, locale) | `Lax` | yes (prod) | **no** (intentional) | `/` | 1 year |
| CSRF | `Lax` | yes (prod) | depends on stack | `/` | per session |

**Reference:** Story 1.6 cookie semantics table in dev notes; round 2 patches on Go `Secure`/`HttpOnly` documentation.

---

## 3. Strict allowlist validation on user-controlled writes

**Failure mode:** Handler writes a user-supplied value to a cookie, database, log, or filesystem without verifying it's in the allowed set. Stored values may be malformed, oversized, or contain control characters.

**Canonical resolution:**
- For enum-like values (theme `{system, light, dark}`, severity `{info, warning, danger}`): explicit set membership check, reject with HTTP 400 / 422 on mismatch.
- For free-form text (project name, reason): length cap *and* character-class validation appropriate to the field. Persist the validated value, not the raw input.
- No "validate by trying to use it" — validate before any write.

**Reference:** Story 1.6 round 2 patches on theme-value sanitization; Story 1.10 Go seeder strict-validation patches.

---

## 4. No dynamic `RegExp` on untrusted input

**Failure mode:** Client- or server-side code constructs a regex from user-controlled input (`new RegExp(`(^|; )${name}=...`)`). Special regex characters in the input change the pattern's meaning — at best, a false-negative match; at worst, ReDoS or unintended capture.

**Canonical resolution:**
- Prefer literal regex (`/(^| )fm_theme=([^;]+)/`) when the pattern is known.
- If the pattern truly must be dynamic, escape regex special characters in the interpolated value *and* document why dynamic construction is necessary.

**Reference:** Story 1.6 round 2 cookie-regex injection patch.

---

## 5. Filesystem-write defaults in tooling

**Failure mode:** Build scripts and tooling accept output paths without normalization; absolute paths, `..` traversal, symlink-following, or in-place mutation can corrupt the working tree.

**Canonical resolution:**
- Normalize output paths and assert they're inside the expected project directory.
- Reject absolute paths and paths containing `..` segments.
- Atomic write: write to `<target>.tmp`, then rename. Wrap in `try/finally` for `.tmp` cleanup on failure.
- Path guards must defeat symlinks — resolve real path before the prefix check.

See also: [fieldmark_shared/CLAUDE.md](../../fieldmark_shared/CLAUDE.md) §"Build-Script Defensive Defaults".

**Reference:** Story 1.4 rounds 4–5 patches on `optimize-css.mjs` path handling.

---

## 6. CSRF posture per stack

**Failure mode:** Cross-stack drift on which mutations require CSRF tokens; one stack accepts unauthenticated state-changing POSTs while the others reject them.

**Canonical resolution per stack:**
- **.NET** — `[ValidateAntiForgeryToken]` (or `services.AddAntiforgery()` global filter) on every state-changing endpoint. HTMX requests include the token via `hx-headers` or a per-form hidden input.
- **Django** — middleware on by default; HTMX requests include the token via `hx-headers` set on `<body>` (`{"X-CSRFToken": "{{ csrf_token }}"}`).
- **Go** — Fiber has no built-in CSRF middleware in scope yet; document the exemption in stack `CLAUDE.md` and revisit when real auth lands (currently deferred per ADR-012).

The cross-stack divergence is acceptable as long as it's *documented* and the missing protection in Go is tracked.

**Reference:** Story 1.6 decision-needed item on Go CSRF; Story 1.11 antiforgery dev-notes section.

---

## 7. Identity / session attributes

**Failure mode:** Stub/dev auth accepts client-supplied identity headers without warning the operator that they're running in stub mode.

**Canonical resolution:**
- Stub auth middleware logs a warning at startup naming the env var and resolved identity (Go `StubAuthMiddleware`).
- Anonymous-fallback behavior on lookup failure is documented and visible in logs — never silent.
- Pre-production deploy checklist verifies stub flags are off.

**Reference:** Story 1.9 dev notes "Why the middleware is permissive on failure"; review patches on `FIELDMARK_STUB_ACTOR` warning log.

---

## How to Use This Checklist

**When authoring a story (`bmad-create-story`):** for any story that handles forms, cookies, redirects, user input, or filesystem writes, add Given/When/Then AC blocks for the applicable categories. The story 1.11 / 1.6 / 1.4 review-round counts are evidence of what happens otherwise.

**When implementing a story:** if a category is in the AC list, implement the canonical resolution for the stack you're touching. If you deviate (e.g., theme cookie deliberately omits `HttpOnly`), document the deviation in dev notes with rationale.

**When reviewing a story (`bmad-code-review`):** for any code that touches the trigger surfaces above, walk the seven categories and flag any defaults that are missing or silently absent.

---

*Ratified by Epic 1 retrospective 2026-05-25.*
