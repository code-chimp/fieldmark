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
//
// The role string returned from the DB is constrained by the Postgres CHECK
// on fiber_auth.user_roles.role to the canonical five values enumerated in
// domain.AllRoles. Actor.Role is kept as string to avoid a cascade change to
// callers; the domain.Role* constants are the authoritative Go-side names.
func lookupByUsername(ctx context.Context, pool *pgxpool.Pool, username string) (*app.Actor, error) {
	const q = `
        SELECT u.id, u.username, u.display_name, COALESCE(MIN(r.role), '') AS role
          FROM fiber_auth.users u
          LEFT JOIN fiber_auth.user_roles r ON r.user_id = u.id
         WHERE u.username = $1
      GROUP BY u.id, u.username, u.display_name
    `
	var a app.Actor
	// pgxpool.QueryRow acquires a connection from the pool, executes the
	// query, and releases the connection on Scan — no manual connection
	// management required. Pool exhaustion or network errors surface as
	// non-nil err here and are handled by the caller (StubAuthMiddleware
	// logs and binds anonymous rather than returning HTTP 500).
	err := pool.QueryRow(ctx, q, username).Scan(&a.ID, &a.Username, &a.DisplayName, &a.Role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("auth: lookupByUsername: %w", err)
	}
	return &a, nil
}
