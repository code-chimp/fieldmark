//go:build integration

package postgres_test

import (
	"context"
	"testing"

	"github.com/code-chimp/fieldmark-go/internal/data/postgres"
)

func TestReferenceStoreSeesComplianceRuleUpdateWithoutRecreation(t *testing.T) {
	pool := openPool(t)
	defer pool.Close()

	ctx := context.Background()
	store := postgres.NewReferenceStore(pool)
	const code = "OPEN_VIOLATION_GATE"
	const updatedName = "Open Violation Closure Gate (UPDATED)"

	before, err := store.ListComplianceRules(ctx)
	if err != nil {
		t.Fatalf("initial list: %v", err)
	}
	originalName := ""
	for _, rule := range before {
		if rule.Code == code {
			originalName = rule.Name
			break
		}
	}
	if originalName == "" {
		t.Fatalf("seeded rule %s not found", code)
	}
	if originalName == updatedName {
		t.Fatalf("seeded rule already has test update name %q", updatedName)
	}

	if _, err := pool.Exec(ctx,
		`UPDATE domain.compliance_rule SET name = $1 WHERE code = $2`,
		updatedName, code,
	); err != nil {
		t.Fatalf("update: %v", err)
	}
	defer func() {
		_, _ = pool.Exec(ctx,
			`UPDATE domain.compliance_rule SET name = $1 WHERE code = $2`,
			originalName, code,
		)
	}()

	after, err := store.ListComplianceRules(ctx)
	if err != nil {
		t.Fatalf("second list: %v", err)
	}
	for _, rule := range after {
		if rule.Code == code {
			if rule.Name != updatedName {
				t.Fatalf("updated rule name = %q, want %q", rule.Name, updatedName)
			}
			return
		}
	}
	t.Fatalf("seeded rule %s not found after update", code)
}
