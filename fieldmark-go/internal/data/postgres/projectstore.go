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

// ProjectStore is the narrow per-aggregate read interface for domain.project.
// Write methods (Create, Save) land in Story 2.8 / 2.12 when consuming
// handlers know the shape they need.
type ProjectStore interface {
	Load(ctx context.Context, id uuid.UUID) (*entities.Project, error)
	LoadWithRelations(
		ctx context.Context,
		id uuid.UUID,
	) (*entities.Project, []entities.JobSite, []entities.ProjectTradeScope, []entities.ProjectInspector, error)
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
	row := r.QueryRow(
		ctx,
		`SELECT `+projectColumns+` FROM domain.project WHERE id = $1`,
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
