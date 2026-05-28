package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/code-chimp/fieldmark-go/internal/domain/entities"
)

type ReferenceStore interface {
	ListTradeTypes(ctx context.Context) ([]entities.TradeType, error)
	ListViolationCategories(ctx context.Context) ([]entities.ViolationCategory, error)
	ListComplianceRules(ctx context.Context) ([]entities.ComplianceRule, error)
}

type referenceStorePg struct {
	pool *pgxpool.Pool
}

func NewReferenceStore(pool *pgxpool.Pool) ReferenceStore {
	return &referenceStorePg{pool: pool}
}

func (s *referenceStorePg) ListTradeTypes(ctx context.Context) ([]entities.TradeType, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, code, name, description, active
		   FROM domain.trade_type
		  ORDER BY code ASC`)
	if err != nil {
		return nil, fmt.Errorf("referencestore: list trade types: %w", err)
	}
	defer rows.Close()

	var out []entities.TradeType
	for rows.Next() {
		var t entities.TradeType
		if err := rows.Scan(&t.ID, &t.Code, &t.Name, &t.Description, &t.Active); err != nil {
			return nil, fmt.Errorf("referencestore: scan trade type: %w", err)
		}
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("referencestore: iterate trade types: %w", err)
	}
	return out, nil
}

func (s *referenceStorePg) ListViolationCategories(ctx context.Context) ([]entities.ViolationCategory, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, code, name, trade_type_id, default_severity, description, active
		   FROM domain.violation_category
		  ORDER BY code ASC`)
	if err != nil {
		return nil, fmt.Errorf("referencestore: list violation categories: %w", err)
	}
	defer rows.Close()

	var out []entities.ViolationCategory
	for rows.Next() {
		var v entities.ViolationCategory
		if err := rows.Scan(
			&v.ID,
			&v.Code,
			&v.Name,
			&v.TradeTypeID,
			&v.DefaultSeverity,
			&v.Description,
			&v.Active,
		); err != nil {
			return nil, fmt.Errorf("referencestore: scan violation category: %w", err)
		}
		out = append(out, v)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("referencestore: iterate violation categories: %w", err)
	}
	return out, nil
}

func (s *referenceStorePg) ListComplianceRules(ctx context.Context) ([]entities.ComplianceRule, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, code, name, description, rule_kind, parameters, active
		   FROM domain.compliance_rule
		  ORDER BY code ASC`)
	if err != nil {
		return nil, fmt.Errorf("referencestore: list compliance rules: %w", err)
	}
	defer rows.Close()

	var out []entities.ComplianceRule
	for rows.Next() {
		var r entities.ComplianceRule
		if err := rows.Scan(
			&r.ID,
			&r.Code,
			&r.Name,
			&r.Description,
			&r.RuleKind,
			&r.Parameters,
			&r.Active,
		); err != nil {
			return nil, fmt.Errorf("referencestore: scan compliance rule: %w", err)
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("referencestore: iterate compliance rules: %w", err)
	}
	return out, nil
}
