package handlers

import (
	"context"
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

var fmThemeCycle = map[string]string{"system": "light", "light": "dark", "dark": "system"}

func themeEntries(c fiber.Ctx) (theme, next string) {
	v := c.Cookies("fm_theme", "system")
	if v != "system" && v != "light" && v != "dark" {
		v = "system"
	}
	return v, fmThemeCycle[v]
}

// LoginHandlers holds the DB pool for auth-related route handlers.
type LoginHandlers struct {
	Pool *pgxpool.Pool
}

func (h *LoginHandlers) GetLogin(c fiber.Ctx) error {
	if !auth.ActorFromCtx(c).IsAnonymous() {
		return c.Redirect().Status(fiber.StatusFound).To("/")
	}
	users, err := h.listSeededUsers(c.Context())
	theme, next := themeEntries(c)
	if err != nil {
		log.Printf("login: list users: %v", err)
		return c.Render("pages/login", fiber.Map{
			"Title":           "Sign in",
			"Users":           nil,
			"Error":           "Unable to list users — check the database connection.",
			"FmTheme":         theme,
			"FmThemeNext":     next,
			"FmThemeResolved": theme,
		})
	}
	return c.Render("pages/login", fiber.Map{
		"Title":           "Sign in",
		"Users":           users,
		"FmTheme":         theme,
		"FmThemeNext":     next,
		"FmThemeResolved": theme,
	})
}

func (h *LoginHandlers) PostLogin(c fiber.Ctx) error {
	username := strings.TrimSpace(c.FormValue("username"))
	if username == "" {
		return h.renderInvalid(c, "Username is required.")
	}
	found, err := h.lookupUser(c.Context(), username)
	if err != nil {
		log.Printf("login: lookup %q: %v", username, err)
		return h.renderInvalid(c, "Internal error — check server logs.")
	}
	if !found {
		return h.renderInvalid(c, "Unknown user — pick from the list.")
	}
	c.Cookie(&fiber.Cookie{
		Name:     auth.CookieName(),
		Value:    username,
		Path:     "/",
		MaxAge:   31536000,
		SameSite: "Lax",
		HTTPOnly: false, // dev stub — not a credential
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
	theme, next := themeEntries(c)
	c.Status(fiber.StatusUnprocessableEntity)
	return c.Render("pages/login", fiber.Map{
		"Title":           "Sign in",
		"Users":           users,
		"Error":           message,
		"FmTheme":         theme,
		"FmThemeNext":     next,
		"FmThemeResolved": theme,
	})
}

func (h *LoginHandlers) listSeededUsers(ctx context.Context) ([]seededUser, error) {
	if h.Pool == nil {
		return nil, nil
	}
	const q = `
	  SELECT u.username, u.display_name, COALESCE(MIN(r.role), '') AS role
	    FROM fiber_auth.users u
	    LEFT JOIN fiber_auth.user_roles r ON r.user_id = u.id
	GROUP BY u.username, u.display_name
	ORDER BY u.username
	`
	rows, err := h.Pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []seededUser
	for rows.Next() {
		var u seededUser
		if err := rows.Scan(&u.Username, &u.DisplayName, &u.Role); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func (h *LoginHandlers) lookupUser(ctx context.Context, username string) (bool, error) {
	const q = `SELECT 1 FROM fiber_auth.users WHERE username = $1`
	var dummy int
	err := h.Pool.QueryRow(ctx, q, username).Scan(&dummy)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
