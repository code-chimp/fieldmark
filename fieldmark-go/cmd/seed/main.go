// Command seed reads docker/postgres/init/seed-uuids/dev-users.json and
// writes six users into fiber_auth.users + fiber_auth.user_roles using the
// canonical UUIDs from the manifest. Idempotent via ON CONFLICT.
//
// No password storage: the Go stack uses stub auth (ADR-012, Story 1.9) —
// identity is asserted via the X-FieldMark-Actor cookie/header carrying a
// username, not a password. The manifest's "password" field is read and
// discarded by this seeder; .NET and Django persist it via their own
// framework's password hasher.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type manifestEntry struct {
	ID          string  `json:"id"`
	Username    string  `json:"username"`
	DisplayName string  `json:"display_name"`
	Password    string  `json:"password"` // intentionally unused (see file header)
	Role        *string `json:"role"`     // null for the no-role test user
}

type manifest struct {
	Users []manifestEntry `json:"users"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "seed:", err)
		os.Exit(1)
	}
}

func run() error {
	path, err := findManifest()
	if err != nil {
		return fmt.Errorf("find manifest: %w", err)
	}

	m, err := parseManifest(path)
	if err != nil {
		return fmt.Errorf("parse manifest %s: %w", path, err)
	}

	dsn := strings.TrimSpace(os.Getenv("FIELDMARK_DATABASE_URL"))
	if dsn == "" {
		dsn = "postgres://fieldmark:fieldmark@localhost:5432/fieldmark"
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("pgxpool: %w", err)
	}
	defer pool.Close()

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	for _, e := range m.Users {
		if _, err := tx.Exec(ctx, `
			INSERT INTO fiber_auth.users (id, username, display_name)
			    VALUES ($1, $2, $3)
			    ON CONFLICT (id) DO UPDATE
			       SET username     = EXCLUDED.username,
			           display_name = EXCLUDED.display_name
		`, e.ID, e.Username, e.DisplayName); err != nil {
			return fmt.Errorf("upsert user %s: %w", e.Username, err)
		}

		if e.Role == nil {
			// No-role entry: skip both user_roles statements (AC #5).
			// The manifest is the only writer for fiber_auth.user_roles, so a
			// no-role user will never have a stale row to clean up in normal
			// operation. Manual overrides are out of scope.
			continue
		}

		if _, err := tx.Exec(ctx, `
			INSERT INTO fiber_auth.user_roles (user_id, role)
			    VALUES ($1, $2)
			    ON CONFLICT (user_id, role) DO NOTHING
		`, e.ID, *e.Role); err != nil {
			return fmt.Errorf("upsert role for %s: %w", e.Username, err)
		}

		if _, err := tx.Exec(ctx,
			`DELETE FROM fiber_auth.user_roles WHERE user_id = $1 AND role <> $2`,
			e.ID, *e.Role,
		); err != nil {
			return fmt.Errorf("prune roles for %s: %w", e.Username, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	log.Printf("seed_dev_users: %d users, 0 errors", len(m.Users))
	return nil
}

func parseManifest(path string) (*manifest, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var m manifest
	if err := json.NewDecoder(f).Decode(&m); err != nil {
		return nil, err
	}
	if len(m.Users) != 6 {
		return nil, fmt.Errorf("manifest must contain exactly 6 users, got %d", len(m.Users))
	}
	validRoles := map[string]bool{
		"ADMIN": true, "COMPLIANCE_OFFICER": true,
		"INSPECTOR": true, "SITE_SUPERVISOR": true, "EXECUTIVE": true,
	}
	for _, u := range m.Users {
		if u.ID == "" || u.Username == "" || u.DisplayName == "" || u.Password == "" {
			return nil, fmt.Errorf("manifest entry %q has empty required field(s)", u.Username)
		}
		if u.Role != nil && !validRoles[*u.Role] {
			return nil, fmt.Errorf("manifest entry %q has invalid role %q", u.Username, *u.Role)
		}
	}
	return &m, nil
}

// findManifest walks up from the current working directory looking for
// docker/postgres/init/seed-uuids/dev-users.json. Returns the absolute path
// or an error if not found within 5 ancestor levels.
func findManifest() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	cur := cwd
	for range 5 {
		candidate := filepath.Join(cur, "docker", "postgres", "init", "seed-uuids", "dev-users.json")
		if _, err := os.Stat(candidate); err == nil {
			return filepath.Abs(candidate)
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			break
		}
		cur = parent
	}
	return "", fmt.Errorf("dev-users.json not found within 5 ancestors of %s", cwd)
}
