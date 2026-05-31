// Package handlers — grid_projects_handler.go
//
// POST /grid/projects — AG Grid SSRM data endpoint.
// Returns { "rows": [...], "lastRow": N } per the contract at
// docs/reference/ag-grid-ssrm-contract.md
package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	web "github.com/code-chimp/fieldmark-go/internal/web"
	"github.com/code-chimp/fieldmark-go/internal/web/auth"
)

// GridProjectsHandlers groups dependencies for the grid endpoint.
type GridProjectsHandlers struct {
	Pool *pgxpool.Pool
}

// projectGridRow is the manually-projected wire type (NFR6 — no AutoMapper).
// Keys match the contract doc exactly: snake_case, seven columns.
type projectGridRow struct {
	ID                   string  `json:"id"`
	Code                 string  `json:"code"`
	Name                 string  `json:"name"`
	Status               string  `json:"status"`
	ComplianceScore      int     `json:"compliance_score"`
	StartDate            string  `json:"start_date"`
	TargetCompletionDate *string `json:"target_completion_date"`
}

// gridResponse is the SSRM envelope (camelCase keys per AG Grid vendor vocabulary).
type gridResponse struct {
	Rows    []projectGridRow `json:"rows"`
	LastRow int              `json:"lastRow"`
}

// gridError is the 400 error envelope.
type gridError struct {
	Error string `json:"error"`
}

// PostGridProjects handles POST /grid/projects.
// Read-only query — no transaction, no audit entry, no state change.
// CSRF: exempt (read; AG Grid datasource does not send tokens).
// See docs/reference/ag-grid-ssrm-contract.md
func (h *GridProjectsHandlers) PostGridProjects(c fiber.Ctx) error {
	actor := auth.ActorFromCtx(c)
	if !auth.Can(actor, "project.read", uuid.Nil) {
		return c.Status(fiber.StatusForbidden).JSON(gridError{Error: "forbidden"})
	}

	// Pre-check: detect explicit null for sortModel/filterModel before full unmarshal.
	// Go's json.Unmarshal treats null slices/maps as nil (indistinguishable from absent),
	// so explicit null would silently be accepted as empty. The contract requires arrays/objects.
	var rawFields map[string]json.RawMessage
	if err := json.Unmarshal(c.Body(), &rawFields); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(gridError{Error: "invalid request body"})
	}
	if sm, ok := rawFields["sortModel"]; ok && string(sm) == "null" {
		return c.Status(fiber.StatusBadRequest).JSON(gridError{Error: "sortModel must be an array"})
	}
	if fm, ok := rawFields["filterModel"]; ok && string(fm) == "null" {
		return c.Status(fiber.StatusBadRequest).JSON(gridError{Error: "filterModel must be an object"})
	}

	var req web.SsrmRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(gridError{Error: "invalid request body"})
	}

	parsed, err := web.ParseSsrmRequest(req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(gridError{Error: err.Error()})
	}

	where := parsed.BuildWhereClause()
	orderBy := parsed.BuildOrderByClause()

	// --- Count query ---
	countSQL := fmt.Sprintf(`SELECT COUNT(*) FROM domain.project %s`, where)
	var total int
	if err := h.Pool.QueryRow(c.Context(), countSQL, parsed.Args...).Scan(&total); err != nil {
		slog.Error("grid/projects: count query", "err", err)
		return err
	}

	// --- Data query (manual projection — NFR6) ---
	dataSQL := fmt.Sprintf(`
		SELECT id, code, name, status, compliance_score, start_date, target_completion_date
		  FROM domain.project
		  %s
		  %s
		 LIMIT $%d OFFSET $%d`,
		where,
		orderBy,
		len(parsed.Args)+1,
		len(parsed.Args)+2,
	)
	dataArgs := append(parsed.Args, parsed.Limit, parsed.Offset) //nolint:gocritic

	rows, err := h.Pool.Query(c.Context(), dataSQL, dataArgs...)
	if err != nil {
		slog.Error("grid/projects: data query", "err", err)
		return err
	}
	defer rows.Close()

	result := make([]projectGridRow, 0, parsed.Limit)
	for rows.Next() {
		var (
			id         uuid.UUID
			code       string
			name       string
			status     string
			score      int
			startDate  time.Time
			targetDate *time.Time
		)
		if err := rows.Scan(&id, &code, &name, &status, &score, &startDate, &targetDate); err != nil {
			slog.Error("grid/projects: scan row", "err", err)
			return err
		}
		row := projectGridRow{
			ID:              id.String(),
			Code:            code,
			Name:            name,
			Status:          status,
			ComplianceScore: score,
			StartDate:       startDate.Format("2006-01-02"),
		}
		if targetDate != nil {
			s := targetDate.Format("2006-01-02")
			row.TargetCompletionDate = &s
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		slog.Error("grid/projects: rows err", "err", err)
		return err
	}

	return c.JSON(gridResponse{Rows: result, LastRow: total})
}
