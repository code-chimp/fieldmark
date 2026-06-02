package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/code-chimp/fieldmark-go/internal/domain/entities"
)

// ProjectStore is the narrow per-aggregate interface for domain.project.
// Write methods introduced by Story 2.8.
type ProjectStore interface {
	Load(ctx context.Context, id uuid.UUID) (*entities.Project, error)
	LoadWithRelations(
		ctx context.Context,
		id uuid.UUID,
	) (*entities.Project, []entities.JobSite, []entities.ProjectTradeScope, []entities.ProjectInspector, error)
	// CreateInTx persists the project + join rows within the caller's open transaction.
	// Callers own the transaction lifecycle (begin / commit / rollback).
	CreateInTx(ctx context.Context, tx pgx.Tx, created *entities.CreatedProject) error
}

type projectStorePg struct {
	pool *pgxpool.Pool
}

// NewProjectStore returns a ProjectStore backed by the provided pgx pool.
func NewProjectStore(pool *pgxpool.Pool) ProjectStore {
	return &projectStorePg{pool: pool}
}

const projectColumns = `id, code, name, description, status,
	start_date, target_completion_date, actual_closed_at,
	compliance_score, created_at, updated_at`

// Querier is the narrow read-only interface satisfied by both
// *pgxpool.Pool and pgx.Tx. Exported so tests (and any future callers
// that already hold a transaction) can drive the same scan code paths
// the production store uses, instead of maintaining a parallel
// shadow-scan implementation in tests.
type Querier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

// LoadProjectFrom reads a single domain.project row via the provided
// Querier. Returns ErrProjectNotFound if no row matches. Useful for
// callers (typically tests) that hold an open pgx.Tx and want to
// exercise the production scan logic without committing.
func LoadProjectFrom(ctx context.Context, q Querier, id uuid.UUID) (*entities.Project, error) {
	return loadProject(ctx, q, id)
}

// LoadProjectForUpdateFrom reads a single domain.project row and acquires a
// row-level lock within the caller's open transaction.
func LoadProjectForUpdateFrom(ctx context.Context, q Querier, id uuid.UUID) (*entities.Project, error) {
	return loadProjectForUpdate(ctx, q, id)
}

// LoadTradeScopesFrom reads the project's trade-scope join rows using the
// caller's provided Querier, typically an open transaction.
func LoadTradeScopesFrom(ctx context.Context, q Querier, id uuid.UUID) ([]entities.ProjectTradeScope, error) {
	return loadTradeScopes(ctx, q, id)
}

// LoadInspectorsFrom reads the project's inspector join rows using the caller's
// provided Querier, typically an open transaction.
func LoadInspectorsFrom(ctx context.Context, q Querier, id uuid.UUID) ([]entities.ProjectInspector, error) {
	return loadInspectors(ctx, q, id)
}

func scanProject(row pgx.Row, p *entities.Project) error {
	return row.Scan(
		&p.ID,
		&p.Code,
		&p.Name,
		&p.Description,
		&p.Status,
		&p.StartDate,
		&p.TargetCompletionDate,
		&p.ActualClosedAt,
		&p.ComplianceScore,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
}

func (s *projectStorePg) Load(ctx context.Context, id uuid.UUID) (*entities.Project, error) {
	return loadProject(ctx, s.pool, id)
}

func loadProject(ctx context.Context, r Querier, id uuid.UUID) (*entities.Project, error) {
	return loadProjectBySQL(ctx, r, `SELECT `+projectColumns+` FROM domain.project WHERE id = $1`, id)
}

func loadProjectForUpdate(ctx context.Context, r Querier, id uuid.UUID) (*entities.Project, error) {
	return loadProjectBySQL(ctx, r, `SELECT `+projectColumns+` FROM domain.project WHERE id = $1 FOR UPDATE`, id)
}

func loadProjectBySQL(ctx context.Context, r Querier, sql string, id uuid.UUID) (*entities.Project, error) {
	row := r.QueryRow(
		ctx,
		sql,
		id,
	)
	var p entities.Project
	if err := scanProject(row, &p); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("projectstore: load: %w", err)
	}
	return &p, nil
}

// LoadWithRelations reads the project and its three relation tables from a
// single REPEATABLE READ read-only transaction. The four queries then see
// the same database snapshot — a concurrent writer cannot make the
// project's job_sites / trade_scopes / inspectors disagree with each other
// or with the parent project row.
func (s *projectStorePg) LoadWithRelations(
	ctx context.Context,
	id uuid.UUID,
) (*entities.Project, []entities.JobSite, []entities.ProjectTradeScope, []entities.ProjectInspector, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.RepeatableRead,
		AccessMode: pgx.ReadOnly,
	})
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("projectstore: begin read tx: %w", err)
	}
	// Read-only transaction; commit vs rollback is equivalent for visibility,
	// but rolling back is the cheaper signal that we never intended to write.
	defer func() { _ = tx.Rollback(ctx) }()

	project, err := loadProject(ctx, tx, id)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	sites, err := loadJobSites(ctx, tx, id)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	scopes, err := loadTradeScopes(ctx, tx, id)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	inspectors, err := loadInspectors(ctx, tx, id)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	return project, sites, scopes, inspectors, nil
}

// CreateInTx persists the project + join rows in FK order:
//  1. domain.project
//  2. domain.project_trade_scope (one row per scope)
//  3. domain.project_inspector   (one row per inspector; zero rows if empty)
//
// All writes share the caller's transaction so the entire create is atomic.
func (s *projectStorePg) CreateInTx(
	ctx context.Context,
	tx pgx.Tx,
	created *entities.CreatedProject,
) error {
	p := created.Project
	_, err := tx.Exec(
		ctx,
		`INSERT INTO domain.project
			(id, code, name, description, status, start_date, target_completion_date,
			 compliance_score, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, DEFAULT, DEFAULT)`,
		p.ID, p.Code, p.Name, p.Description, string(p.Status),
		p.StartDate, p.TargetCompletionDate, p.ComplianceScore,
	)
	if err != nil {
		return fmt.Errorf("projectstore: insert project: %w", err)
	}

	for _, sc := range created.Scopes {
		if _, err = tx.Exec(
			ctx,
			`INSERT INTO domain.project_trade_scope (project_id, trade_type_id) VALUES ($1, $2)`,
			sc.ProjectID, sc.TradeTypeID,
		); err != nil {
			return fmt.Errorf("projectstore: insert trade scope: %w", err)
		}
	}

	for _, insp := range created.Inspectors {
		if _, err = tx.Exec(
			ctx,
			`INSERT INTO domain.project_inspector (project_id, user_id) VALUES ($1, $2)`,
			insp.ProjectID, insp.UserID,
		); err != nil {
			return fmt.Errorf("projectstore: insert inspector: %w", err)
		}
	}
	return nil
}

func loadJobSites(ctx context.Context, r Querier, projectID uuid.UUID) ([]entities.JobSite, error) {
	rows, err := r.Query(
		ctx,
		`SELECT id, project_id, label, address FROM domain.job_site WHERE project_id = $1`,
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("projectstore: load job_sites: %w", err)
	}
	defer rows.Close()

	var sites []entities.JobSite
	for rows.Next() {
		var js entities.JobSite
		if err := rows.Scan(&js.ID, &js.ProjectID, &js.Label, &js.Address); err != nil {
			return nil, fmt.Errorf("projectstore: scan job_site: %w", err)
		}
		sites = append(sites, js)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("projectstore: iterate job_sites: %w", err)
	}
	return sites, nil
}

func loadTradeScopes(ctx context.Context, r Querier, projectID uuid.UUID) ([]entities.ProjectTradeScope, error) {
	rows, err := r.Query(
		ctx,
		`SELECT project_id, trade_type_id FROM domain.project_trade_scope WHERE project_id = $1`,
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("projectstore: load trade_scopes: %w", err)
	}
	defer rows.Close()

	var scopes []entities.ProjectTradeScope
	for rows.Next() {
		var sc entities.ProjectTradeScope
		if err := rows.Scan(&sc.ProjectID, &sc.TradeTypeID); err != nil {
			return nil, fmt.Errorf("projectstore: scan trade_scope: %w", err)
		}
		scopes = append(scopes, sc)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("projectstore: iterate trade_scopes: %w", err)
	}
	return scopes, nil
}

func loadInspectors(ctx context.Context, r Querier, projectID uuid.UUID) ([]entities.ProjectInspector, error) {
	rows, err := r.Query(
		ctx,
		`SELECT project_id, user_id FROM domain.project_inspector WHERE project_id = $1`,
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("projectstore: load inspectors: %w", err)
	}
	defer rows.Close()

	var inspectors []entities.ProjectInspector
	for rows.Next() {
		var pi entities.ProjectInspector
		if err := rows.Scan(&pi.ProjectID, &pi.UserID); err != nil {
			return nil, fmt.Errorf("projectstore: scan inspector: %w", err)
		}
		inspectors = append(inspectors, pi)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("projectstore: iterate inspectors: %w", err)
	}
	return inspectors, nil
}
