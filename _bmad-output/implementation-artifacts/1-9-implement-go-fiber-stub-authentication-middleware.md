# Story 1.9: Implement Go/Fiber stub authentication middleware

Status: done

## Story

As a developer running the Go stack at MVP,
I want a stub authentication mechanism that hydrates a configurable user identity onto every request,
So that the Go stack can render role-aware pages and exercise the cross-stack parity contract while real auth remains deferred per ADR-012.

## Acceptance Criteria

1. **`fiber_auth.users` and `fiber_auth.user_roles` tables exist after running a Go-owned migration step.** The DDL is hand-authored SQL embedded in the Go binary and applied by a one-shot command `go run ./cmd/migrate-fiber-auth` (run manually after `make reset`). After applying:
   - `\dt fiber_auth.*` lists exactly: `users`, `user_roles`.
   - `fiber_auth.users` has columns `id uuid PRIMARY KEY`, `username varchar(64) NOT NULL UNIQUE`, `display_name varchar(128) NOT NULL`, `created_at timestamptz NOT NULL DEFAULT now()`.
   - `fiber_auth.user_roles` has columns `user_id uuid NOT NULL REFERENCES fiber_auth.users(id) ON DELETE CASCADE`, `role varchar(64) NOT NULL` with a `CHECK` constraint restricting `role` to the five canonical values (`ADMIN`, `COMPLIANCE_OFFICER`, `INSPECTOR`, `SITE_SUPERVISOR`, `EXECUTIVE`), `PRIMARY KEY (user_id, role)`.
   - Re-running the migrate command against an already-migrated database succeeds with no errors and no row mutations (idempotent via `CREATE TABLE IF NOT EXISTS` + `CREATE INDEX IF NOT EXISTS`).
   - No DDL targets `domain.*` (verified by grep over the embedded SQL).

2. **`internal/web/auth/` package exists and exposes `StubAuthMiddleware`.** The middleware resolves an actor identifier from, in order: (a) the `X-FieldMark-Actor` cookie, (b) the `X-FieldMark-Actor` HTTP header, (c) the `FIELDMARK_STUB_ACTOR` environment variable, (d) the literal sentinel `"anonymous"`. The identifier is treated as a `username` string (not a UUID — debug-friendly, matches Story 1.11's user-switcher cookie semantics).

3. **The middleware hydrates the request context with an `*app.Actor` via `c.Locals("user", actor)`.** Resolution rules:
   - If the resolved identifier is `"anonymous"` (or empty after trimming): bind a sentinel `app.Anonymous()` actor (zero `uuid.Nil` ID, username `"anonymous"`, role `""`).
   - Otherwise: query `fiber_auth.users` joined to `fiber_auth.user_roles` by username; on hit, bind an `*app.Actor` with the row's `id`, `username`, and **first** matched `role` (alphabetically sorted; multi-role support is post-MVP).
   - On query error: log the error and bind the anonymous sentinel. Do **not** return HTTP 500 — the application must remain navigable on auth-store failure, since this is stub posture.
   - On lookup miss (username not in `fiber_auth.users`): bind the anonymous sentinel. Do **not** auto-create users; that is Story 1.10's seeder's job.

4. **`internal/web/auth.RequireAuth` middleware factory exists but is not yet applied to any current route.** When applied to a route group in a future story (1-11), an unauthenticated request (Actor is `Anonymous()`) returns HTTP `302` with `Location: /login`. Story 1-9 ships the factory and a `auth_test.go` unit test proving the redirect contract; **it does not register `RequireAuth` on `/`, `/privacy`, or `/fragments/compliance-tile`** — those remain public until 1-11 reshapes the route inventory across all three stacks together.

5. **No new routes are registered.** The route inventory from `go run ./cmd/web -dump-routes` is byte-identical to its HEAD-before-this-story output. No `/login`, no `/logout`, no `/auth/*`. The cross-stack `/login` and `/logout` paths land in Story 1.11 across all three stacks simultaneously, preserving the AC-#5 invariant in Story 1.7 and Story 1.8.

6. **`StubAuthMiddleware` is wired into `cmd/web/main.go` before route registration.** It runs on every request and hydrates `c.Locals("user", ...)` even on currently-public routes — so future handlers added in any story can read the actor without further wiring. Wiring point: between `app.Use(logger.New())` and the static middleware (so even static asset handlers see hydrated context if they need it — none do today, but the placement is the canonical "site-wide middleware" slot).

7. **Database connectivity is upgraded from `pgx.Conn` to `pgxpool.Pool`.** The middleware needs concurrent-safe DB access; `pgx.Conn` is single-connection and unsafe across goroutines. `internal/data/postgres/db.go`'s `Connect(dsn)` returns `*pgxpool.Pool` (renamed signature: it still validates with a `Ping`). Pool size is the pgxpool default (no explicit override at this story — the 4×CPU sizing from Architecture D3 lands when load shape becomes observable). `cmd/web/main.go` is updated to defer `pool.Close()` (note: `pool.Close()` is synchronous and takes no arguments, unlike `conn.Close(ctx)`).

8. **`cmd/migrate-fiber-auth/main.go` exists and is documented in the README.** Invocation: `go run ./cmd/migrate-fiber-auth` (no flags). Reads `FIELDMARK_DATABASE_URL` (defaulting to the local `postgres://fieldmark:fieldmark@localhost:5432/fieldmark`), opens a temporary pool, applies the embedded SQL inside a single transaction, logs `"fiber_auth: schema applied (idempotent)"`, exits 0. On SQL error, rolls back and exits non-zero with the wrapped error printed to stderr.

9. **`fieldmark-go/CLAUDE.md` `## Authentication` section is rewritten.** New content covers:
   - Stub middleware lives at `internal/web/auth/`. ADR-012 explicitly defers real Go auth; this story closes the placeholder gap.
   - Identifier-resolution order (cookie → header → env → anonymous).
   - The `fiber_auth.users` + `fiber_auth.user_roles` tables are framework-local (ADR-012), Go-owned. Schema is created via `go run ./cmd/migrate-fiber-auth` after `make reset`. **Never** colocate with `domain.*` DDL.
   - Login/logout HTML and the unauthenticated-redirect contract land in Story 1.11.
   - Replacement of the stub with real auth is epic-sized work outside MVP scope (sessions, password hashing, CSRF tokens for cookie-auth, user management UI). Do **not** grow the stub into real auth incrementally.

10. **`fieldmark-go/README.md` "Getting Started" gains a fiber-auth migrate step.** After step 3 (Run the application), insert a new step **between Install dependencies and Run the application**: `go run ./cmd/migrate-fiber-auth` with a one-line explanation that this creates the framework-local `fiber_auth` tables (idempotent — safe to re-run). The "Current State" paragraph is updated: stub authentication middleware is wired; the application reads an actor identity from the `X-FieldMark-Actor` cookie/header or the `FIELDMARK_STUB_ACTOR` env var.

11. **`make parity` exits 0.** Route inventory diff stays clean (Story 1.9 adds no routes — see AC #5). `pg_indexes WHERE schemaname='domain'` is unchanged from the canonical snapshot — Story 1.9 touches only `fiber_auth`, never `domain`.

12. **Build, vet, staticcheck, and tests stay green.** From `fieldmark-go/`:
    - `make fmt-check` — zero diffs.
    - `make vet` — zero issues.
    - `make staticcheck` — zero issues.
    - `make test` — all tests pass, including new unit tests for `StubAuthMiddleware` resolution order and `RequireAuth` redirect contract (Task 6).

## Tasks / Subtasks

- [x] Task 1: Author the framework-local `fiber_auth` DDL (AC: #1)
  - [x] 1.1 Create `fieldmark-go/internal/data/postgres/migrations/fiber_auth/001_initial.sql` with the following content (snake_case column names per the canonical naming convention; `SCREAMING_SNAKE_CASE` enum values per Architecture §Naming Conventions; `CREATE TABLE IF NOT EXISTS` for idempotence):

    ```sql
    -- fiber_auth bootstrap. Framework-local per ADR-012. Owned by the Go stack.
    -- Re-runnable: every statement is IF NOT EXISTS.
    -- DO NOT issue any DDL against domain.* here.

    CREATE TABLE IF NOT EXISTS fiber_auth.users (
        id           uuid          PRIMARY KEY,
        username     varchar(64)   NOT NULL UNIQUE,
        display_name varchar(128)  NOT NULL,
        created_at   timestamptz   NOT NULL DEFAULT now()
    );

    CREATE TABLE IF NOT EXISTS fiber_auth.user_roles (
        user_id uuid        NOT NULL REFERENCES fiber_auth.users(id) ON DELETE CASCADE,
        role    varchar(64) NOT NULL,
        PRIMARY KEY (user_id, role),
        CONSTRAINT user_roles_role_check CHECK (role IN (
            'ADMIN', 'COMPLIANCE_OFFICER', 'INSPECTOR', 'SITE_SUPERVISOR', 'EXECUTIVE'
        ))
    );

    CREATE INDEX IF NOT EXISTS idx_fiber_auth_user_roles_user_id
        ON fiber_auth.user_roles(user_id);
    ```

  - [x] 1.2 Create `fieldmark-go/internal/data/postgres/migrations/fiber_auth/embed.go`:

    ```go
    // Package fiberauthmigrations embeds the framework-local DDL for the
    // fiber_auth schema and exposes it as a single SQL script applied by
    // cmd/migrate-fiber-auth.
    //
    // ADR-012: fiber_auth tables are framework-local. domain.* DDL lives in
    // docker/postgres/init/ and is owned by infrastructure (ADR-014); nothing
    // in this package may target domain.*.
    package fiberauthmigrations

    import _ "embed"

    //go:embed 001_initial.sql
    var InitialSQL string
    ```

  - [x] 1.3 Verify with `grep -n 'domain' fieldmark-go/internal/data/postgres/migrations/fiber_auth/*.sql` — must return zero matches.

- [x] Task 2: Upgrade `internal/data/postgres/db.go` from `pgx.Conn` to `pgxpool.Pool` (AC: #7)
  - [x] 2.1 Rewrite `fieldmark-go/internal/data/postgres/db.go`:

    ```go
    // Package postgres provides database connectivity and persistence adapters
    // for the FieldMark Go stack. All SQL targets the infrastructure-owned
    // domain schema (domain.*) plus the framework-local fiber_auth schema.
    // This package owns nothing in domain — it only reads and writes to
    // tables created by docker/postgres/init scripts; fiber_auth DDL is
    // applied by cmd/migrate-fiber-auth.
    package postgres

    import (
        "context"
        "fmt"

        "github.com/jackc/pgx/v5/pgxpool"
    )

    // Connect opens a pgxpool against the FieldMark PostgreSQL database and
    // validates it with a Ping. The caller is responsible for closing the
    // pool via pool.Close() at shutdown.
    //
    // dsn must be a valid libpq-style connection string or URL, e.g.:
    //
    //   postgres://fieldmark:fieldmark@localhost:5432/fieldmark
    func Connect(dsn string) (*pgxpool.Pool, error) {
        ctx := context.Background()

        pool, err := pgxpool.New(ctx, dsn)
        if err != nil {
            return nil, fmt.Errorf("postgres: pool open: %w", err)
        }

        if err := pool.Ping(ctx); err != nil {
            pool.Close()
            return nil, fmt.Errorf("postgres: ping: %w", err)
        }

        return pool, nil
    }
    ```

  - [x] 2.2 Update `fieldmark-go/cmd/web/main.go` to:
    - Receive `*pgxpool.Pool` from `postgres.Connect(dsn)`.
    - Replace `defer func() { _ = conn.Close(context.Background()) }()` with `defer pool.Close()`.
    - Pass the pool into the auth middleware factory (Task 4).
  - [x] 2.3 Confirm `go.mod` does **not** need a new direct dependency — `pgxpool` is a subpackage of `github.com/jackc/pgx/v5` already in the require block. After Task 2.1 compiles, run `go mod tidy`; the only expected change is `github.com/google/uuid` moving from `// indirect` to a direct require (because Task 3 imports it directly).

- [x] Task 3: Author the `internal/app/actor.go` Actor type (AC: #2, #3)
  - [x] 3.1 Create `fieldmark-go/internal/app/actor.go`:

    ```go
    // Package app is the THIN coordinator — wiring only (Deps struct, env
    // config, Actor type). It must not import fiber/v3, and no business
    // rules live here. See fieldmark-go/CLAUDE.md.
    package app

    import "github.com/google/uuid"

    // Actor is the resolved request principal. Constructed by the auth
    // middleware (internal/web/auth) and read by handlers via the web
    // layer's ActorFromCtx helper. Lives in app/ so future packages
    // (e.g., domain audit-entry helpers) can take an Actor parameter
    // without depending on web/.
    type Actor struct {
        ID       uuid.UUID
        Username string
        Role     string // "" for anonymous; one of the five canonical roles otherwise
    }

    // Anonymous returns the sentinel actor representing an unauthenticated
    // request. ID is uuid.Nil. Role is the empty string.
    func Anonymous() *Actor {
        return &Actor{Username: "anonymous"}
    }

    // IsAnonymous is true when the Actor has no resolved identity.
    func (a *Actor) IsAnonymous() bool {
        return a == nil || a.ID == uuid.Nil
    }
    ```

  - [x] 3.2 Do **not** introduce a `domain.Role` enum at this story. The five role names live as a CHECK constraint in `001_initial.sql` (AC #1) and as a string-matching helper inside `internal/web/auth/lookup.go`. The typed `Role` value object lands with Story 1.12 (`authz.Can` primitive) across all three stacks — mirroring the Story 1.7 and Story 1.8 decisions to keep role names in the seeder until 1.12.

- [x] Task 4: Author the stub middleware + lookup (AC: #2, #3, #6)
  - [x] 4.1 Create `fieldmark-go/internal/web/auth/lookup.go`:

    ```go
    // Package auth holds the framework-local stub authentication middleware
    // for the Go/Fiber stack (ADR-012 deferral). It owns reads against the
    // fiber_auth schema; writes are reserved for Story 1.10's seeder.
    package auth

    import (
        "context"
        "errors"
        "fmt"

        "github.com/jackc/pgx/v5"
        "github.com/jackc/pgx/v5/pgxpool"

        "github.com/code-chimp/fieldmark-go/internal/app"
    )

    // lookupByUsername returns the resolved Actor for the given username,
    // joining fiber_auth.users to fiber_auth.user_roles. On lookup miss
    // returns (nil, nil) — callers should treat that as anonymous. Multi-
    // role users return the alphabetically-first role; multi-role support
    // is post-MVP (Story 1.12 introduces the typed Role value object and
    // can revisit then).
    func lookupByUsername(ctx context.Context, pool *pgxpool.Pool, username string) (*app.Actor, error) {
        const q = `
            SELECT u.id, u.username, COALESCE(MIN(r.role), '') AS role
              FROM fiber_auth.users u
              LEFT JOIN fiber_auth.user_roles r ON r.user_id = u.id
             WHERE u.username = $1
          GROUP BY u.id, u.username
        `
        var a app.Actor
        err := pool.QueryRow(ctx, q, username).Scan(&a.ID, &a.Username, &a.Role)
        if err != nil {
            if errors.Is(err, pgx.ErrNoRows) {
                return nil, nil
            }
            return nil, fmt.Errorf("auth: lookupByUsername: %w", err)
        }
        return &a, nil
    }
    ```

  - [x] 4.2 Create `fieldmark-go/internal/web/auth/stub.go`:

    ```go
    package auth

    import (
        "log"
        "os"
        "strings"

        "github.com/gofiber/fiber/v3"
        "github.com/jackc/pgx/v5/pgxpool"

        "github.com/code-chimp/fieldmark-go/internal/app"
    )

    // localsKey is the c.Locals() key under which the request's *app.Actor
    // is stored. Use ActorFromCtx to read it; never type-assert directly.
    const localsKey = "user"

    // cookieName / headerName are the two request-borne identifier carriers.
    // The cookie is set by Story 1.11's /login user-switcher; the header is
    // for scripted tests and ad-hoc curl flows. envVar is the deployment-
    // fixed fallback (set in docker-compose or .env for parity scenarios).
    const (
        cookieName = "X-FieldMark-Actor"
        headerName = "X-FieldMark-Actor"
        envVar     = "FIELDMARK_STUB_ACTOR"
    )

    // StubAuthMiddleware returns a Fiber middleware that hydrates an *app.Actor
    // onto c.Locals(localsKey) for every request. Resolution order: cookie,
    // header, env var, anonymous. Lookup failure (DB error or miss) falls
    // back to anonymous; the application remains navigable so the developer
    // can see logs and fix the auth store. ADR-012 stub posture: this is
    // intentional and not a production-grade pattern.
    func StubAuthMiddleware(pool *pgxpool.Pool) fiber.Handler {
        return func(c fiber.Ctx) error {
            username := resolveUsername(c)
            if username == "" || username == "anonymous" {
                c.Locals(localsKey, app.Anonymous())
                return c.Next()
            }
            actor, err := lookupByUsername(c.Context(), pool, username)
            if err != nil {
                log.Printf("auth: lookup error for %q: %v (binding anonymous)", username, err)
                c.Locals(localsKey, app.Anonymous())
                return c.Next()
            }
            if actor == nil {
                c.Locals(localsKey, app.Anonymous())
                return c.Next()
            }
            c.Locals(localsKey, actor)
            return c.Next()
        }
    }

    // RequireAuth returns a middleware that 302-redirects unauthenticated
    // requests to /login. NOT applied to any route in Story 1.9; Story 1.11
    // mounts it on business routes once /login exists.
    func RequireAuth() fiber.Handler {
        return func(c fiber.Ctx) error {
            actor := ActorFromCtx(c)
            if actor.IsAnonymous() {
                return c.Redirect().Status(fiber.StatusFound).To("/login")
            }
            return c.Next()
        }
    }

    // ActorFromCtx reads the hydrated *app.Actor from c.Locals. Returns
    // app.Anonymous() if the middleware did not run or stored an unexpected
    // type (defensive — never panic on a missing or wrong-typed locals).
    func ActorFromCtx(c fiber.Ctx) *app.Actor {
        v := c.Locals(localsKey)
        if a, ok := v.(*app.Actor); ok && a != nil {
            return a
        }
        return app.Anonymous()
    }

    func resolveUsername(c fiber.Ctx) string {
        if v := strings.TrimSpace(c.Cookies(cookieName)); v != "" {
            return v
        }
        if v := strings.TrimSpace(c.Get(headerName)); v != "" {
            return v
        }
        return strings.TrimSpace(os.Getenv(envVar))
    }
    ```

  - [x] 4.3 The `fiber.Ctx`-typed `ActorFromCtx` lives in `internal/web/auth/` (not `internal/app/`) so the `internal/app` package stays Fiber-free per `fieldmark-go/CLAUDE.md` "fiber.Ctx must not escape internal/web". The architecture diagram (line 1203) lists `internal/app/actor.go` — that is the `Actor` struct only (Task 3.1). The context-reader function lives in `web/auth/` because it requires a Fiber type.

- [x] Task 5: Wire the middleware into `cmd/web/main.go` (AC: #6, #7)
  - [x] 5.1 In `fieldmark-go/cmd/web/main.go`, after `app.Use(logger.New())` and **before** `app.Use("/static", static.New(...))`, add:

    ```go
    app.Use(auth.StubAuthMiddleware(pool))
    ```

    Required imports: `"github.com/code-chimp/fieldmark-go/internal/web/auth"`.

  - [x] 5.2 Move the database connection from after the `-dump-routes` short-circuit to **before** route registration — the middleware factory needs the pool at route-wiring time. The `-dump-routes` early-return must still avoid touching the database. Restructure so the order is:
    1. Parse `-dump-routes` flag (existing).
    2. If flag set: build a minimal Fiber app, register routes, dump, exit. **Skip pool open entirely** — preserves Story 1.3's invariant that route dump never needs a live DB.
    3. Else: open pool, build Fiber app with middleware, register routes, listen.

    Suggested refactor (replace lines 20–105 of `main.go`):

    ```go
    func main() {
        dumpRoutes := flag.Bool("dump-routes", false, "print normalized route inventory and exit")
        flag.Parse()

        if *dumpRoutes {
            runDumpRoutes()
            return
        }

        runServer()
    }

    func runDumpRoutes() {
        app := buildApp(nil) // nil pool — middleware factory must accept nil and no-op (see Task 5.3)
        registerRoutes(app)
        var lines []string
        for _, r := range app.GetRoutes(true) {
            method := strings.ToLower(r.Method)
            path := strings.ToLower(r.Path)
            if strings.HasPrefix(path, "/static") || method == "head" {
                continue
            }
            lines = append(lines, fmt.Sprintf("%s %s", method, path))
        }
        sort.Strings(lines)
        for _, l := range lines { fmt.Println(l) }
    }

    func runServer() {
        dsn := strings.TrimSpace(os.Getenv("FIELDMARK_DATABASE_URL"))
        if dsn == "" {
            dsn = "postgres://fieldmark:fieldmark@localhost:5432/fieldmark"
        }
        pool, err := postgres.Connect(dsn)
        if err != nil {
            log.Fatalf("database connection failed: %v", err)
        }
        defer pool.Close()
        log.Println("database connection validated")

        app := buildApp(pool)
        registerRoutes(app)
        log.Fatal(app.Listen(":3000"))
    }
    ```

    `buildApp(pool)` constructs the Fiber app with engine + middleware (logger, then `auth.StubAuthMiddleware(pool)` — unless pool is nil, in which case the auth middleware is omitted for the route-dump path). `registerRoutes(app)` is the existing route registrations (lines 45–61 of current main.go).

  - [x] 5.3 Alternative if Task 5.2's refactor feels too invasive: leave the existing structure and gate the middleware registration inline:

    ```go
    if !*dumpRoutes {
        pool, err := postgres.Connect(dsn)
        // ... (existing pool setup)
        app.Use(auth.StubAuthMiddleware(pool))
    }
    ```

    Either pattern is acceptable. The refactor (5.2) is preferred for readability; the inline gate (5.3) is acceptable if the dev prefers minimum diff. **Do not** register `StubAuthMiddleware` in the `-dump-routes` path — it would force a live DB on parity invocations and break Story 1.3.

  - [x] 5.4 Confirm `make` target alignment: from repo root, `make run-go` continues to work without any additional flag. From repo root, the parity script `tools/parity/dump-routes-fiber.sh` (which `cd`s into `fieldmark-go` and runs `go run ./cmd/web -dump-routes`) continues to exit 0 with the same output. **Do not edit `tools/parity/dump-routes-fiber.sh`** — it is correct as written.

- [x] Task 6: Write unit tests for the middleware (AC: #4, #12)
  - [x] 6.1 Create `fieldmark-go/internal/web/auth/stub_test.go`. Use the standard library `testing` package (the project rule per `fieldmark-go/CLAUDE.md` and `fiber-reference.md` §Unit testing — no third-party framework, no testify, no mocks of pgx). Tests focus on the *resolution-order* behavior of `resolveUsername` and the redirect contract of `RequireAuth` — neither of those touches the database, so no `pgxpool` fixture is required. Tests that exercise `lookupByUsername` belong in a future integration test file (`//go:build integration`); do **not** add a DB-integration test in this story.

    ```go
    package auth

    import (
        "net/http/httptest"
        "testing"

        "github.com/gofiber/fiber/v3"

        "github.com/code-chimp/fieldmark-go/internal/app"
    )

    func TestResolveUsername_CookieWinsOverHeader(t *testing.T) {
        a := fiber.New()
        a.Get("/probe", func(c fiber.Ctx) error {
            return c.SendString(resolveUsername(c))
        })

        req := httptest.NewRequest("GET", "/probe", nil)
        req.AddCookie(&http.Cookie{Name: cookieName, Value: "marisol"})
        req.Header.Set(headerName, "pat")
        resp, _ := a.Test(req)
        body := readBody(t, resp)
        if body != "marisol" {
            t.Fatalf("want marisol, got %q", body)
        }
    }

    func TestResolveUsername_HeaderWhenNoCookie(t *testing.T) { /* analogous */ }
    func TestResolveUsername_EnvWhenNoCookieAndNoHeader(t *testing.T) {
        t.Setenv(envVar, "kenji")
        // ...
    }
    func TestResolveUsername_EmptyWhenNoneProvided(t *testing.T) { /* ... */ }

    func TestRequireAuth_RedirectsAnonymousToLogin(t *testing.T) {
        a := fiber.New()
        a.Use(func(c fiber.Ctx) error {
            c.Locals(localsKey, app.Anonymous())
            return c.Next()
        })
        a.Get("/secure", RequireAuth(), func(c fiber.Ctx) error {
            return c.SendString("ok")
        })

        resp, _ := a.Test(httptest.NewRequest("GET", "/secure", nil))
        if resp.StatusCode != fiber.StatusFound {
            t.Fatalf("want 302, got %d", resp.StatusCode)
        }
        if loc := resp.Header.Get("Location"); loc != "/login" {
            t.Fatalf("want /login, got %q", loc)
        }
    }

    func TestRequireAuth_PassesAuthenticatedThrough(t *testing.T) {
        a := fiber.New()
        a.Use(func(c fiber.Ctx) error {
            c.Locals(localsKey, &app.Actor{Username: "marisol", Role: "ADMIN"})
            return c.Next()
        })
        a.Get("/secure", RequireAuth(), func(c fiber.Ctx) error { return c.SendString("ok") })
        // ... assert 200 + body == "ok"
    }
    ```

    (Sketches above — fill in the omitted helpers (`readBody`, `http.Cookie` import, etc.) and the four resolution-order test bodies. The pattern is `fiber.New() + a.Test(req)` with no real network — see `fiber-reference.md` §Unit testing for the canonical posture.)

  - [x] 6.2 Do **not** unit-test `lookupByUsername` here — it requires a real DB connection, and the project rule is "real PostgreSQL only" (root `CLAUDE.md` → `docs/hard-rules.md`). An integration-tagged test (`//go:build integration`) is the right home; defer it to Story 1.10 or 1.11 when end-to-end coverage of the lookup is meaningful.
  - [x] 6.3 Run `go test ./internal/web/auth/...` — all tests pass.

- [x] Task 7: Author `cmd/migrate-fiber-auth/main.go` (AC: #1, #8)
  - [x] 7.1 Create `fieldmark-go/cmd/migrate-fiber-auth/main.go`:

    ```go
    // Command migrate-fiber-auth applies the framework-local fiber_auth DDL
    // (idempotent via CREATE TABLE IF NOT EXISTS). Invoke after `make reset`
    // to bring up the Go-stack auth tables. ADR-012 stub posture: this is
    // not a general-purpose migration runner; real auth migration tooling
    // lands when the deferred Go-auth epic begins.
    package main

    import (
        "context"
        "log"
        "os"
        "strings"

        "github.com/jackc/pgx/v5/pgxpool"

        fam "github.com/code-chimp/fieldmark-go/internal/data/postgres/migrations/fiber_auth"
    )

    func main() {
        dsn := strings.TrimSpace(os.Getenv("FIELDMARK_DATABASE_URL"))
        if dsn == "" {
            dsn = "postgres://fieldmark:fieldmark@localhost:5432/fieldmark"
        }

        ctx := context.Background()
        pool, err := pgxpool.New(ctx, dsn)
        if err != nil {
            log.Fatalf("migrate-fiber-auth: pool: %v", err)
        }
        defer pool.Close()

        tx, err := pool.Begin(ctx)
        if err != nil {
            log.Fatalf("migrate-fiber-auth: begin: %v", err)
        }
        defer func() { _ = tx.Rollback(ctx) }() // no-op after commit

        if _, err := tx.Exec(ctx, fam.InitialSQL); err != nil {
            log.Fatalf("migrate-fiber-auth: exec: %v", err)
        }
        if err := tx.Commit(ctx); err != nil {
            log.Fatalf("migrate-fiber-auth: commit: %v", err)
        }

        log.Println("fiber_auth: schema applied (idempotent)")
    }
    ```

  - [x] 7.2 Verify the build: `go build ./cmd/migrate-fiber-auth` — no errors. Then run it twice end-to-end:
    - `make reset` (from repo root) — destroys volume, re-runs `001_schemas.sql` (creates `fiber_auth` schema as empty).
    - `cd fieldmark-go && go run ./cmd/migrate-fiber-auth` — applies DDL; expect `"fiber_auth: schema applied (idempotent)"`.
    - Re-run `go run ./cmd/migrate-fiber-auth` — same output, exit 0, no DDL errors (the `IF NOT EXISTS` clauses make every statement a no-op on the second pass).
  - [x] 7.3 Verify via `psql -h localhost -U fieldmark -d fieldmark` (password `fieldmark`):
    - `\dt fiber_auth.*` — exactly two tables: `users`, `user_roles`.
    - `\d+ fiber_auth.users` — columns and types match AC #1.
    - `\d+ fiber_auth.user_roles` — `CHECK (role IN ('ADMIN', 'COMPLIANCE_OFFICER', 'INSPECTOR', 'SITE_SUPERVISOR', 'EXECUTIVE'))` constraint is present.
    - `\di fiber_auth.*` — `idx_fiber_auth_user_roles_user_id` is present.

- [x] Task 8: Update `fieldmark-go/CLAUDE.md` (AC: #9)
  - [x] 8.1 Rewrite the `## Authentication` section (currently lines 106–110). New content:

    ```markdown
    ## Authentication

    The Go stack uses a **stub authentication middleware** (ADR-012 explicit deferral). Real auth — sessions, password hashing, CSRF, login forms, user management UI — is an epic-sized follow-on, not MVP scope.

    **Where it lives:** `internal/web/auth/` (middleware + lookup). The hydrated principal type is `app.Actor` in `internal/app/actor.go`.

    **How identity is resolved (in order):**
    1. `X-FieldMark-Actor` cookie (set by Story 1.11's /login user-switcher).
    2. `X-FieldMark-Actor` HTTP header (for scripted tests and ad-hoc curl).
    3. `FIELDMARK_STUB_ACTOR` environment variable (deployment-fixed fallback).
    4. Otherwise: anonymous (sentinel `app.Anonymous()`).

    The resolved username is looked up against `fiber_auth.users` joined to `fiber_auth.user_roles`. On miss or DB error, the request binds the anonymous actor and continues — the application stays navigable while the developer investigates.

    **Schema ownership:** `fiber_auth.users` and `fiber_auth.user_roles` are framework-local (ADR-012), Go-owned, defined in `internal/data/postgres/migrations/fiber_auth/001_initial.sql`. Apply with `go run ./cmd/migrate-fiber-auth` after `make reset`. **Never** colocate this DDL with `domain.*` init scripts — `domain` is infrastructure-owned (ADR-014).

    **What lands later:**
    - Story 1.10 seeds the six dev users (`marisol`, `pat`, `aisha`, `ravi`, `kenji`, plus a no-role test user) into `fiber_auth` from the shared UUID manifest.
    - Story 1.11 introduces `/login` (user-switcher stub list rendered as Basecoat buttons) and `/logout` on all three stacks simultaneously, plus mounts `auth.RequireAuth()` on business routes.

    **Replacing the stub with real auth is out of MVP scope.** Do not grow the stub incrementally — when real auth lands, it lands as a coherent epic (session tables, password hashing via `golang.org/x/crypto/bcrypt`, CSRF middleware, real login form, registration/management UI, password reset flow). Until then: this is the stub posture.
    ```

  - [x] 8.2 The existing `## Database` section is correct and unaffected. The new "Authentication" section supersedes its closing paragraph ("This stack reads and writes to the shared `domain` schema and will eventually own `fiber_auth` for authentication") only insofar as `fiber_auth` is now actively owned, not "eventually"; lightly amend that sentence to read: "This stack reads and writes to the shared `domain` schema and owns the framework-local `fiber_auth` schema for stub authentication (see §Authentication)."

- [x] Task 9: Update `fieldmark-go/README.md` (AC: #10)
  - [x] 9.1 In the "Getting Started" section (currently lines 141–164), insert a new step between "Install dependencies" (step 2) and "Run the application" (step 3):

    ```markdown
    **3. Apply framework-local auth schema:**

    ```bash
    go run ./cmd/migrate-fiber-auth
    ```

    Creates the `fiber_auth.users` and `fiber_auth.user_roles` tables (idempotent — safe to re-run). Required once after `make reset`.
    ```

    Renumber the existing "Run the application" step to **4**.

  - [x] 9.2 Update the "Current State" paragraph (currently lines 167–169). New text:

    > Standup is complete. The Fiber server starts, validates the Postgres connection, serves static assets, renders a full-page dashboard route and one HTMX fragment route (`/fragments/compliance-tile`), and hydrates a stub authentication actor onto every request from the `X-FieldMark-Actor` cookie/header or the `FIELDMARK_STUB_ACTOR` env var. Folder layout matches the structure above. Domain implementation begins with the first feature story.

- [x] Task 10: Verify QA gates and parity (AC: #11, #12)
  - [x] 10.1 From `fieldmark-go/`: `make fmt-check && make vet && make staticcheck && make test` — all green.
  - [x] 10.2 From `fieldmark-go/`: `go run ./cmd/web -dump-routes` — diff against `git show HEAD:fieldmark-go/cmd/web/main.go` baseline. Output must be byte-identical (same three routes: `get /`, `get /fragments/compliance-tile`, `get /privacy`). Capture both: `git stash && go run ./cmd/web -dump-routes > /tmp/before.txt && git stash pop && go run ./cmd/web -dump-routes > /tmp/after.txt && diff /tmp/before.txt /tmp/after.txt` — zero diff lines.
  - [x] 10.3 From repo root: `make parity` — exits 0. The three stacks' route dumps are identical; `tools/parity/canonical-pg-indexes.txt` snapshot for `domain.*` is unchanged.
  - [x] 10.4 Manual smoke test (optional but useful):
    - `make reset && cd fieldmark-go && go run ./cmd/migrate-fiber-auth`.
    - Insert one row manually for the smoke test (Story 1.10 will automate this):
      ```sql
      INSERT INTO fiber_auth.users (id, username, display_name)
          VALUES ('00000000-0000-0000-0000-000000000001', 'marisol', 'Marisol (smoke test)');
      INSERT INTO fiber_auth.user_roles (user_id, role)
          VALUES ('00000000-0000-0000-0000-000000000001', 'COMPLIANCE_OFFICER');
      ```
    - In another terminal, `make run-go`. Then:
      - `curl -i http://localhost:3000/` — expect 200, anonymous request.
      - `curl -i -H 'X-FieldMark-Actor: marisol' http://localhost:3000/` — expect 200; the log line shows the resolved actor (add a temporary `log.Printf` in a handler if you want to observe; remove before commit).
      - `curl -i -H 'X-FieldMark-Actor: nobody' http://localhost:3000/` — expect 200, anonymous (lookup miss → anonymous).
    - Roll back the manual insert before completing the story (`DELETE FROM fiber_auth.user_roles; DELETE FROM fiber_auth.users;`) — Story 1.10's seeder owns the persistent dev-user rows.

## Dev Notes

### Brownfield posture — what exists today (read before writing anything)

State of the Go stack at HEAD of this branch:

- `fieldmark-go/cmd/web/main.go` — single-file entry. Reads `FIELDMARK_DATABASE_URL`, opens a single `pgx.Conn` (not a pool), registers three routes, listens on `:3000`. Handles the `-dump-routes` flag with an explicit early-return before the DB connect (Story 1.3 invariant — parity must not require a live DB). **This story rewrites the DB connect to use `pgxpool.Pool` (Task 2) and adds middleware wiring (Task 5).** The `-dump-routes` early-return invariant must be preserved.
- `fieldmark-go/internal/data/postgres/db.go` — owns the `Connect(dsn) (*pgx.Conn, error)` function. **This story rewrites it to return `*pgxpool.Pool`** (Task 2.1). Tradeoff explained in "Why pgxpool, not pgx.Conn" below.
- `fieldmark-go/internal/app/` — currently three empty `.gitkeep` directories (`dto/`, `ports/`, `services/`) plus `doc.go`. Story 1.9 creates the first real Go file there: `actor.go`. The `Deps` struct described in the architecture (line 1201) and `fieldmark-go/CLAUDE.md` lines 73–74 is **not** introduced at this story — no aggregate stores exist yet (Epic 2+), so a `Deps` container has nothing to hold beyond the pool, and adding it now would be speculative wiring. Story 1.9's middleware factory takes the pool directly as a parameter; Epic 2's first story can introduce `app.Deps{Pool, Projects, ...}` when it has stores to hold.
- `fieldmark-go/internal/web/auth/` — does not exist yet. `internal/web/middleware/.gitkeep` exists (the architecture diagram puts a generic `middleware/` directory there too). **Place the stub auth code at `internal/web/auth/`** per the architecture line 1214 (`auth/` is a sibling of `middleware/`, not a child) — auth is a distinct concern from generic request middleware (request-id, logging, error recovery) and warrants its own package.
- `fieldmark-go/cmd/migrate-fiber-auth/` — does not exist. Story 1.9 creates it.
- `fieldmark-go/cmd/seed/` — does not exist. Story 1.10 creates it.
- `fieldmark-go/internal/data/postgres/migrations/` — does not exist. Story 1.9 creates the `fiber_auth/` subfolder and its embedded SQL.
- `fieldmark-go/internal/data/postgres/stores/.gitkeep` — present but empty. No aggregate stores exist yet — that work starts at Epic 2.
- `fieldmark-go/CLAUDE.md` (line 106–110) currently says "`fiber_auth` schema exists in the database but Fiber authentication is **deferred by design**." Story 1.9 **is** the story that closes the placeholder gap with the stub; Task 8 rewrites this section.
- `tools/parity/dump-routes-fiber.sh` invokes `go run ./cmd/web -dump-routes`. Story 1.3 verified this works without a live DB. Story 1.9's main.go refactor (Task 5.2 or 5.3) must preserve that behavior — the `-dump-routes` path must not call `postgres.Connect`.
- `tools/parity/canonical-pg-indexes.txt` snapshots `domain.*` indexes only — `fiber_auth` indexes are out of scope. Don't touch this file.
- `docker/postgres/init/001_schemas.sql` line 21 already creates the `fiber_auth` schema empty. The schema is in place; Story 1.9 populates its first two tables.
- `domain.*` tables (12 of them) already exist (Story 1.2). Domain `created_by_user_id` / `actor_id` columns are opaque `uuid` columns with no FK to any auth schema — that contract is held by Story 1.2's DDL and remains untouched here.

### Why `pgxpool`, not `pgx.Conn`

The current `Connect` returns `*pgx.Conn` (single connection). Fiber serves concurrent requests on goroutines; `pgx.Conn` is **not** safe for concurrent use. Today there are only three handlers and none of them touch the database, so the unsafety is latent. Once Story 1.9's middleware runs `lookupByUsername` on every request, concurrent goroutines will race the connection.

`pgxpool.Pool` is the standard pgx concurrent-safe wrapper. Switching now (one-file change) is cheaper than retrofitting it under load. Architecture D3 (line 320) already specifies pgxpool for the Go stack with a default sizing of 4×CPU — we adopt pgxpool here, but **do not** override the default pool size (pgxpool's default is fine for the dev/MVP footprint, and an explicit size belongs with the first observable load shape, not this story).

`pool.Close()` is **synchronous and takes no arguments**, unlike `pgx.Conn.Close(ctx)`. The corresponding `defer` in `main.go` changes from `defer func() { _ = conn.Close(context.Background()) }()` to `defer pool.Close()`. This is the most common subtle defect in pgx.Conn → pgxpool migrations.

### Why the middleware lives at `internal/web/auth/`, not `internal/web/middleware/`

The architecture diagram (line 1214) places `auth/` as a sibling of `middleware/` under `internal/web/`:

```
fieldmark-go/internal/web/
├── handlers/
├── auth/           ← stub middleware here
├── middleware/     ← request-id, logger, recovery, etc. (when they land)
├── viewmodels/
└── ...
```

Reasons:
- Auth has its own DDL ownership (`fiber_auth.*`), its own schema lookups, its own redirect contract — none of which apply to a generic request-id or logger middleware.
- Keeping `internal/web/auth/` as a dedicated package makes the import explicit: `auth.StubAuthMiddleware(pool)` and `auth.RequireAuth()`. If it lived under `middleware/`, the call site would read `middleware.StubAuth(...)`, which is less descriptive.
- Mirrors the architecture's intentional separation. This is **not** a deviation from the diagram.

### Why no `app.Deps` wiring in this story

The architecture (line 1201) describes `internal/app/deps.go` as a `Deps` struct holding "DB pool, all Stores, Authz." At Story 1.9 there are no aggregate stores (Epic 2 introduces them) and the authz primitive lands in Story 1.12. A `Deps{Pool: pool}` struct that holds only the pool is degenerate — it adds an indirection layer without coordinating anything. The middleware factory takes `*pgxpool.Pool` directly (Task 4.2); the first Epic-2 story (which adds `ProjectStore`) is the right place to introduce `Deps` and refactor the middleware to take `Deps` instead.

Don't add `app.Deps` here. Don't speculate the shape.

### Why a manual `cmd/migrate-fiber-auth` instead of startup auto-migration

The hard rule "no auto-migration tooling against the `domain` schema" applies to `domain.*` only — `fiber_auth.*` is framework-local. The question is what shape "framework-local migration" takes in Go, where there's no idiomatic native migration tool baked into Fiber the way EF Core / Django migrations are baked in.

Three options considered and rejected:

1. **`CREATE TABLE IF NOT EXISTS` in `cmd/web/main.go` startup.** Tempting because it's one less command for the developer to remember. Rejected: muddies the boundary between "app startup" and "schema bootstrap." A future on-call engineer reading `main.go` shouldn't have to wonder whether the server creates tables. Migrations belong in their own command.
2. **A third-party migration library (`migrate`, `goose`, `tern`).** Rejected for stub posture: we have **one** DDL file and no concept of rollback or version history at the stub phase. Adding 200 KB of dependency for a 30-line SQL file is overhead, and choosing one of three popular tools is a decision that belongs in the deferred real-auth epic, not here.
3. **A SQL file under `docker/postgres/init/`.** Rejected: violates ADR-012 ("`*_auth` schemas are framework-local"). The infra init scripts own `domain.*` only (ADR-014); adding `fiber_auth` DDL there would mean Postgres init runs auth DDL before the Go stack exists, breaking the framework-local boundary.

The chosen option (one-shot `cmd/migrate-fiber-auth` command applying embedded SQL inside a transaction) is the minimum viable framework-local migration step. When the deferred real-auth epic begins, it can swap in a proper migration tool — but at that point there are multiple versioned DDLs to manage and the tool earns its weight.

### Why `X-FieldMark-Actor` carries `username` (not UUID)

The stub identifier travels three paths: cookie (set by Story 1.11's user-switcher), header (for scripted tests), env var (for dev/test deployment defaults). All three are human-readable strings — UUIDs are not. `username` is what a developer types in a curl flag (`-H 'X-FieldMark-Actor: marisol'`) and what shows up in server logs.

The internal Actor still carries the UUID (`Actor.ID`) — that's what `domain.audit_entry.actor_id` will reference (Story 1.10 onwards). The username is the *handle*; the UUID is the *identity*. The middleware's job is to translate the handle to the identity by querying `fiber_auth.users`.

Story 1.11's user-switcher will set the cookie to `username`, not the UUID, for the same reason.

### Why the middleware is permissive on failure (anonymous, not 500)

Stub posture: this is dev/MVP code. If the DB is unreachable, returning HTTP 500 on every request hides the failure mode behind an opaque error page; logging the failure and degrading to anonymous lets the developer see logs, navigate public routes, and still observe the broken state.

In real auth (deferred epic), failure should be hard — auth-store unreachability is a deployment defect, not a graceful-degradation case. The "permissive on failure" rule is **stub-only** and must be reversed when the deferred epic lands. Document this in CLAUDE.md so the future engineer doesn't carry the leniency forward.

### Multi-role users — chosen handling, deferred refinement

`fiber_auth.user_roles` has a composite PK `(user_id, role)`, so a user can hold multiple roles. Story 1.9's `lookupByUsername` returns the **alphabetically-first** role via `MIN(r.role)`. Rationale:

- `Actor.Role` is a single string today; that matches the .NET Story 1.7 and Django Story 1.8 idioms (single conceptual role per user at MVP).
- Story 1.10's manifest assigns exactly one role per dev user.
- Story 1.12 introduces the typed `Role` value object across all three stacks. *Multi-role* is its decision to make, not this story's. If the answer is "users can hold multiple roles and authz unions them," 1.12 changes `Actor.Role` to `[]string` (or to a `RoleSet` type) and updates the three lookups in lockstep.

Don't model multi-role today. Don't add a `[]string` field to `Actor` speculatively.

### Anti-patterns that must NOT slip in

- ❌ Adding `/login`, `/logout`, or `/auth/*` routes to `cmd/web/main.go`. Story 1.11's scope. Adding them in 1.9 breaks `make parity` (.NET and Django don't have these routes yet).
- ❌ Calling `auth.RequireAuth()` on `/`, `/privacy`, or `/fragments/compliance-tile`. Those are public until Story 1.11 re-shapes the route inventory cross-stack.
- ❌ Putting `fiber_auth` DDL in `docker/postgres/init/`. Violates ADR-012. The DDL lives in Go-stack-local SQL embedded in the Go binary.
- ❌ Running `CREATE TABLE` in `cmd/web/main.go` startup. Schema bootstrap and app startup are distinct concerns — keep them separate (one-shot command vs. server loop).
- ❌ Adding a `Role` enum / constants module in `internal/domain/` or `internal/app/`. Defer to Story 1.12. The five canonical names live as a SQL `CHECK` constraint (Task 1.1) and an implicit contract in the lookup; that's enough at this story.
- ❌ Adding `app.Deps` with only a `Pool` field. Degenerate wiring. Wait for Epic 2 when the first store lands.
- ❌ Passing `fiber.Ctx` into `internal/app/` or `internal/domain/`. Hard rule (`fieldmark-go/CLAUDE.md`): `fiber.Ctx` stays in `internal/web/`. `ActorFromCtx` (the only function that takes `fiber.Ctx`) lives in `internal/web/auth/`, not in `internal/app/`.
- ❌ Pinning `pgxpool` pool size in `db.go`. Architecture D3's "4×CPU" sizing is a directional note; the explicit size belongs with the first observable load shape, not this story. pgxpool's default is fine for now.
- ❌ Using `testify` / `gomock` / `mockery` in the new tests. Standard library `testing` only — see `fiber-reference.md` §Unit testing and `fieldmark-go/CLAUDE.md`.
- ❌ Mocking pgx in unit tests. The unit tests in Task 6 cover the pure-Go resolution-order and redirect logic that does not touch the DB. DB-touching tests are integration-tagged (`//go:build integration`) and live in a separate file — defer to Story 1.10/1.11.
- ❌ Returning HTTP 500 on `fiber_auth.users` lookup failure. Stub-posture rule: bind anonymous and continue. Document the inversion that real-auth must apply.
- ❌ Storing the resolved actor anywhere other than `c.Locals("user", actor)`. Don't introduce a global, don't stash in `pgxpool` context, don't piggyback on a request-id middleware.
- ❌ Editing `tools/parity/canonical-pg-indexes.txt` or `tools/parity/dump-routes-fiber.sh`. Both are correct as written; Story 1.9 changes neither route inventory nor `domain.*` indexes.

### Project Structure Notes

Files this story adds or modifies:

- **New:** `fieldmark-go/internal/app/actor.go`
- **New:** `fieldmark-go/internal/web/auth/stub.go`
- **New:** `fieldmark-go/internal/web/auth/lookup.go`
- **New:** `fieldmark-go/internal/web/auth/stub_test.go`
- **New:** `fieldmark-go/internal/data/postgres/migrations/fiber_auth/001_initial.sql`
- **New:** `fieldmark-go/internal/data/postgres/migrations/fiber_auth/embed.go`
- **New:** `fieldmark-go/cmd/migrate-fiber-auth/main.go`
- **Update:** `fieldmark-go/internal/data/postgres/db.go` — `pgx.Conn` → `pgxpool.Pool`.
- **Update:** `fieldmark-go/cmd/web/main.go` — wire middleware; `pool.Close()` instead of `conn.Close(ctx)`; preserve `-dump-routes` early-return.
- **Update:** `fieldmark-go/go.mod` — `github.com/google/uuid` moves from indirect to direct require (no version bump). `go mod tidy` handles it.
- **Update:** `fieldmark-go/CLAUDE.md` — rewrite `## Authentication` section (Task 8.1) and lightly amend the `## Database` closing paragraph (Task 8.2).
- **Update:** `fieldmark-go/README.md` — insert `migrate-fiber-auth` step in Getting Started (Task 9.1); update Current State paragraph (Task 9.2).

No file under `FieldMark/`, `fieldmark_py/`, `docker/`, `fieldmark_shared/`, `tools/`, or root-level config is modified. All file locations align with the architecture's `fieldmark-go/` tree (lines 1167–1224).

### Testing Standards

Per `fiber-reference.md` §Unit testing and `fieldmark-go/CLAUDE.md`:

- **Standard library `testing` package only** — no testify, no gomock, no mockery.
- **Real PostgreSQL for any DB-touching test.** The unit tests in Task 6 are explicitly DB-free (`resolveUsername` and `RequireAuth` redirect logic only) and use `fiber.New() + a.Test(req)` without a real server bind. DB-touching coverage is integration-tagged (`//go:build integration`) and deferred to Story 1.10 or 1.11.
- **Test files colocated with source:** `internal/web/auth/stub_test.go` lives next to `stub.go`.
- **No test of route registration, template rendering, or middleware chain composition.** Those are Fiber concerns, not domain behavior — Playwright covers cross-stack end-to-end (Epic 7).
- **No new test dependencies in `go.mod`.** If you find yourself adding `testify` to "save a few lines," stop — the project rule is hard.

### Previous Story Intelligence

**Story 1.7 (.NET — currently in `review`)** is the structural analog for the .NET stack:

- Idempotent seeding via "exists check then insert" (.NET: `RoleManager.RoleExistsAsync(name)`; Go equivalent for Story 1.10 will be `INSERT ... ON CONFLICT DO NOTHING`). Story 1.9 itself doesn't seed users — it sets up the schema and middleware that 1.10's seeder will write into.
- No login/logout UI in this story. Story 1.7 used `AddIdentityCore` (not `AddDefaultIdentity`) to avoid scaffolding `/Identity/Account/*` Razor pages. Story 1.9's equivalent is "don't add `/login`, `/logout`, or `/auth/*` routes to `cmd/web/main.go`." The parity invariant is binding.
- Role names live in the schema/seeder, not in the domain. Story 1.7 explicitly did not create a `Role` enum in `FieldMark.Domain`. Story 1.9 makes the same call — the five names live as a SQL `CHECK` constraint and as implicit lookup contract; no `internal/domain/role.go` is introduced. Story 1.12 designs the typed `Role` value object across all three stacks at once.
- Migration ownership is bifurcated. .NET has `FieldMark.Data/Migrations/Auth/` (auth context only). Go has `internal/data/postgres/migrations/fiber_auth/` (auth schema only). Neither stack writes any DDL targeting `domain.*` (ADR-014).
- `--dump-routes` early-return is sacred. Story 1.3 fixed both .NET and Go to short-circuit before DB connect; Story 1.7 preserved it; Story 1.9 must preserve it through the main.go restructure (Task 5.2).

**Story 1.8 (Django — currently `ready-for-dev`)** is the Django counterpart:

- Schema isolation via the simplest available mechanism (Django: `search_path` on the DB connection; .NET: separate `DbContext`; Go: separate schema target in hand-authored SQL). Three stacks, three idioms, same outcome.
- Idempotence is required, not optional. Story 1.8 uses `Group.objects.get_or_create()`; Story 1.10 (Go seeder, a future story) will use `INSERT ... ON CONFLICT DO NOTHING`. Story 1.9's migration uses `CREATE TABLE IF NOT EXISTS` + `CREATE INDEX IF NOT EXISTS` — the DDL-layer equivalent.
- No `/login`, `/logout`, or admin scaffolding added. Pure schema + bootstrap. Same posture for Story 1.9.
- Role names stay in the seeder until 1.12. Same for Go.

**Story 1.3 (parity — done)** established the contract that `make parity` exits 0 across `routes` and `pg_indexes`. Story 1.9 is the first story since 1.3 to risk *new schema visible to parity* — but `pg_indexes` snapshot filters `WHERE schemaname='domain'`, so `fiber_auth` indexes are not part of the parity surface. Routes inventory remains clean because Story 1.9 adds zero routes.

**Story 1.4 (design system — currently `review`)** is unrelated to this story's surface. No direct dependency.

### Git Intelligence

Recent commits and their relevance to Story 1.9:

- `d03f0fe feat: e1s3 establish tools parity` — established the `tools/parity/dump-routes-fiber.sh` invocation and the `-dump-routes` early-return in `cmd/web/main.go`. Story 1.9's main.go refactor (Task 5.2) must preserve both. Read the current `cmd/web/main.go` carefully before refactoring.
- `cbf47e9 feat: e1s2 verified sql init scripts` — confirmed `docker/postgres/init/001_schemas.sql` creates the `fiber_auth` schema (empty). Story 1.9 populates it via `cmd/migrate-fiber-auth`.
- `a6fac88 feat: e1s1 confirm scaffolds` — the Go scaffold this story extends. `cmd/web/main.go`, `internal/data/postgres/db.go`, the `internal/{app,data,domain,web}/` skeleton with `.gitkeep` placeholders all ship from this commit. Story 1.9 is the first story to add real Go files under `internal/app/` and `internal/web/auth/`.
- `a4fcc76 task: complete BMAD planning phase` — the planning artifacts (`architecture.md`, `epics.md`, PRD shards) are pinned at this commit. Story 1.9's AC and Dev Notes cite these as written.

No prior commit has added any auth-related code to the Go stack. Story 1.9 is the first.

### Latest Technical Information

- **Go 1.26.2** is in use (`go.mod` line 3). The `//go:embed` directive (Task 1.2) is stable since Go 1.16; no version concerns.
- **Fiber v3.2.0** is the active framework (`go.mod` line 6). API stability is high since v3.0 GA. Key v3 changes relative to v2 that affect this story:
  - Middleware signature is `fiber.Handler` (alias for `func(fiber.Ctx) error`), and Fiber v3 uses **`fiber.Ctx` as a value type** in handler signatures — not `*fiber.Ctx`. Task 4.2's middleware snippets reflect this. Do **not** write `*fiber.Ctx`.
  - Redirect API: `c.Redirect().Status(fiber.StatusFound).To("/login")` is the v3 idiom (Task 4.2). The v2 form `c.Redirect("/login", fiber.StatusFound)` is gone.
  - `c.Cookies(name)` reads cookies; `c.Get(name)` reads headers. Both are unchanged from v2.
  - `c.Locals(key, value)` to write; `c.Locals(key)` to read. The Locals API is unchanged.
- **pgx v5.9.2** with `pgxpool` subpackage (`go.mod` line 8). `pgxpool.New(ctx, dsn)` returns `*pgxpool.Pool`. `pool.Close()` is synchronous (no args). `pool.Begin(ctx)` returns `pgx.Tx`. `pool.QueryRow(ctx, sql, args...).Scan(...)` is the canonical row-read pattern (Task 4.1).
- **`github.com/google/uuid` v1.6.0** is already an indirect dependency. Task 3 imports it directly; `go mod tidy` after Task 3 promotes it to a direct require (no version change).
- **`golang.org/x/tools` / `staticcheck`** — pinned in `go.mod` via the `tool` block (lines 40–43). Run via `go tool ...` per `fieldmark-go/CLAUDE.md` — no global install needed. Story 1.9 introduces no new tools.
- **PostgreSQL 17** is the target. `CHECK` constraints, `varchar(N)`, `timestamptz`, `uuid`, `ON DELETE CASCADE`, and `IF NOT EXISTS` on `CREATE TABLE`/`CREATE INDEX` are all stable.
- **No new packages need to be added to `go.mod`.** `pgx/v5/pgxpool` is a subpackage of the existing pgx dep; `google/uuid` and `embed` are stdlib/indirect-already-present.

### References

- [Architecture: Authentication & Security → D8 (Go/Fiber auth deferred)](_bmad-output/planning-artifacts/architecture.md#authentication--security) — ADR-012 explicit deferral; stub middleware injects a configurable `actor_id`.
- [Architecture: D3 Connection pooling — pgxpool default sizing for Go](_bmad-output/planning-artifacts/architecture.md#data-architecture) — adopt pgxpool; 4×CPU sizing is directional, not pinned here.
- [Architecture: D5 Connection string standardization — `FIELDMARK_DATABASE_URL`](_bmad-output/planning-artifacts/architecture.md#data-architecture) — env var name and local default URL.
- [Architecture: Repository Directory Structure → `fieldmark-go/`](_bmad-output/planning-artifacts/architecture.md#complete-repository-directory-structure) — file locations for `cmd/`, `internal/app/`, `internal/web/auth/`, etc. (lines 1167–1224).
- [Architecture: Naming Conventions](_bmad-output/planning-artifacts/architecture.md#naming-conventions) — `snake_case` schema/table/column; `SCREAMING_SNAKE_CASE` enums in DB and on the wire; canonical route casing rules.
- [Architecture: Architectural Boundaries → Authentication / authorization](_bmad-output/planning-artifacts/architecture.md#architectural-boundaries) — opaque UUID refs in `domain.*`; auth implementation is per-stack-idiomatic; single `authz.Can` call site (Story 1.12, not now).
- [PRD FR1–FR8 — Authentication & Authorization](_bmad-output/planning-artifacts/prd/functional-requirements.md) — framework-local authentication; conceptual roles.
- [PRD architectural-constraints-prd-binding.md — Authentication & Authorization (ADR-012)](_bmad-output/planning-artifacts/prd/architectural-constraints-prd-binding.md) — schema isolation contract; per-stack auth ownership; opaque UUIDs.
- [PRD web-app-specific-requirements.md](_bmad-output/planning-artifacts/prd/web-app-specific-requirements.md) — "Authentication is required on every route except a single public landing page" (the unauthenticated-redirect contract that Story 1.11 will turn on).
- [docs/hard-rules.md](docs/hard-rules.md) — backend authority, infrastructure-owned domain schema, real PostgreSQL in tests, no service layers, no auto-migration against `domain`.
- [fieldmark-go/CLAUDE.md](fieldmark-go/CLAUDE.md) — Go-specific rules; `fiber.Ctx` must not escape `internal/web/`; no business rules in middleware or `app`; no generic repository abstractions.
- [_bmad-output/planning-artifacts/research/fiber-reference.md §Authentication & Authorization Policy](_bmad-output/planning-artifacts/research/fiber-reference.md) — stub posture per ADR-012; `fiber_auth` tables framework-local; no FKs from `domain.*` to `fiber_auth.*`.
- [_bmad-output/planning-artifacts/research/fiber-reference.md §Unit testing](_bmad-output/planning-artifacts/research/fiber-reference.md) — standard library `testing` only; real PostgreSQL for DB-touching tests.
- [_bmad-output/planning-artifacts/research/authentication-authorization-primer.md](_bmad-output/planning-artifacts/research/authentication-authorization-primer.md) — ADR-012 narrative; opaque UUID references; trade-offs accepted.
- [Story 1.7 implementation artifact](_bmad-output/implementation-artifacts/1-7-wire-asp-net-core-identity-to-dotnet-auth-schema-with-conceptual-roles.md) — .NET counterpart; idempotent role pattern; no UI in this story.
- [Story 1.8 implementation artifact](_bmad-output/implementation-artifacts/1-8-wire-django-built-in-auth-to-django-auth-schema-with-conceptual-role-groups.md) — Django counterpart; schema-isolation via `search_path`; same "no UI yet" posture.
- [Story 1.3 implementation artifact](_bmad-output/implementation-artifacts/1-3-establish-tools-parity-and-make-parity-with-per-stack-dump-routes.md) — `-dump-routes` early-return invariant must be preserved; `make parity` contract.
- [docker/postgres/init/001_schemas.sql](docker/postgres/init/001_schemas.sql) — `fiber_auth` schema already created (empty).

## Dev Agent Record

### Agent Model Used

claude-sonnet-4-6

### Debug Log References

- Fixed `TestRequireAuth_PassesAuthenticatedThrough`: test stub actor needed a non-nil UUID (zero UUID → `IsAnonymous()` true → 302 redirect). Added `uuid.MustParse("00000000-0000-0000-0000-000000000001")` to the test actor. Story spec sketch omitted the ID field — the correct invariant is that `IsAnonymous()` checks `a.ID == uuid.Nil`, so any real user must have a non-nil ID.

### Completion Notes List

- Implemented all 10 tasks against all 12 ACs.
- Used Task 5.2 (full `buildApp`/`registerRoutes`/`runDumpRoutes`/`runServer` refactor) — cleaner than the inline-gate alternative (5.3); `-dump-routes` path confirmed to not call `postgres.Connect`.
- `go mod tidy` promoted `github.com/jackc/puddle/v2` (pgxpool's internal pool library) and `github.com/google/uuid` from indirect to direct as expected.
- `make fmt-check`, `make vet`, `make staticcheck`, `make test` all green.
- Route dump byte-identical (4 routes unchanged); `make parity` exits 0.
- `go run ./cmd/migrate-fiber-auth` applied idempotently twice against live DB; `\dt fiber_auth.*` shows exactly `users` and `user_roles`.

### File List

- `fieldmark-go/internal/data/postgres/migrations/fiber_auth/001_initial.sql` (new)
- `fieldmark-go/internal/data/postgres/migrations/fiber_auth/embed.go` (new)
- `fieldmark-go/internal/app/actor.go` (new)
- `fieldmark-go/internal/web/auth/lookup.go` (new)
- `fieldmark-go/internal/web/auth/stub.go` (new)
- `fieldmark-go/internal/web/auth/stub_test.go` (new)
- `fieldmark-go/cmd/migrate-fiber-auth/main.go` (new)
- `fieldmark-go/internal/data/postgres/db.go` (updated — pgx.Conn → pgxpool.Pool)
- `fieldmark-go/cmd/web/main.go` (updated — refactored to buildApp/registerRoutes/runServer/runDumpRoutes; pool.Close(); auth middleware wired)
- `fieldmark-go/go.mod` (updated — uuid and puddle/v2 promoted to direct)
- `fieldmark-go/go.sum` (updated — pgxpool dependency entries added)
- `fieldmark-go/CLAUDE.md` (updated — Authentication section rewritten; Database paragraph amended)
- `fieldmark-go/README.md` (updated — migrate-fiber-auth step added; Current State updated)
- `_bmad-output/implementation-artifacts/sprint-status.yaml` (updated — 1-9 set to review)

## Change Log

- 2026-05-19: Story 1.9 implemented — stub auth middleware wired, fiber_auth DDL created, pgx.Conn upgraded to pgxpool.Pool, all 10 tasks complete, 6 unit tests green, parity clean.
- 2026-05-19: Code review follow-ups addressed — 2 decision items dismissed (ADR-012 stub posture), 3 patch items resolved (DDL cross-reference comment, FIELDMARK_STUB_ACTOR startup warning, pgxpool acquisition lifecycle comment).

### Review Findings

**decision-needed** (dismissed — known stub limitations per ADR-012):
- [x] [Review][Decision] Stub authentication trivially spoofable via `X-FieldMark-Actor` header or cookie — accepts untrusted client-controlled identity without validation/signature. Attacker can impersonate any seeded user. **Dismissed: ADR-012 stub posture by design; real auth is a deferred epic.**
- [x] [Review][Decision] Anonymous fallback on DB lookup failure or missing user silently degrades security/audit while keeping app navigable — masks misconfigs or attacks. **Dismissed: explicit stub-posture choice per Dev Notes "Why the middleware is permissive on failure"; must be reversed when real auth lands.**

**patch** (resolved):
- [x] [Review][Patch] Role names as bare strings with CHECK constraint only — no Go constants/enum; risk of typos or drift between DDL and lookup.go. **Resolved: No Go code compares against role name literals — the DB CHECK constraint is the sole enforcer. Added cross-reference comment in `lookupByUsername` pointing to the DDL as the authority. Go role constants deferred to Story 1.12 (typed Role value object across all stacks) per story spec.**
- [x] [Review][Patch] Environment variable fallback (`FIELDMARK_STUB_ACTOR`) allows fixed identity in deployments; if mis-set, all users share one principal. No validation/logging shown. **Resolved: `StubAuthMiddleware` factory now logs a warning at startup when `FIELDMARK_STUB_ACTOR` is set, explicitly naming the env var and the resolved identity.**
- [x] [Review][Patch] pgxpool upgrade + connection management — pool.Close() vs old conn semantics; potential leaks if any path retains old Conn usage (main.go refactor shown but no explicit error handling around acquisition). **Resolved: grep of all .go files confirms zero remaining `pgx.Conn` / `pgx.Connect` usage. `pool.Close()` (synchronous, no args) is correctly deferred in both `runServer` and `cmd/migrate-fiber-auth`. Pool acquisition errors surface through `pool.QueryRow().Scan()` and are handled in `StubAuthMiddleware` (log + bind anonymous). Added clarifying comment in `lookupByUsername` explaining the pgxpool acquire/release lifecycle.**
