package handlers_test

import (
	"context"
	htmlpkg "html"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/code-chimp/fieldmark-go/internal/app"
	"github.com/code-chimp/fieldmark-go/internal/domain"
	"github.com/code-chimp/fieldmark-go/internal/domain/entities"
	"github.com/code-chimp/fieldmark-go/internal/web/auth"
	"github.com/code-chimp/fieldmark-go/internal/web/handlers"
)

type referenceStoreStub struct {
	tradeTypes []entities.TradeType
	categories []entities.ViolationCategory
	rules      []entities.ComplianceRule
	called     bool
}

func (s *referenceStoreStub) ListTradeTypes(context.Context) ([]entities.TradeType, error) {
	s.called = true
	return s.tradeTypes, nil
}

func (s *referenceStoreStub) ListViolationCategories(context.Context) ([]entities.ViolationCategory, error) {
	s.called = true
	return s.categories, nil
}

func (s *referenceStoreStub) ListComplianceRules(context.Context) ([]entities.ComplianceRule, error) {
	s.called = true
	return s.rules, nil
}

func buildAdminReferenceApp(actor *app.Actor, store *referenceStoreStub) *fiber.App {
	a := newTestApp()
	if actor != nil {
		a.Use(injectActor(actor))
	}
	h := &handlers.AdminReferenceHandlers{Reference: store}
	a.Get("/admin/reference", auth.RequireAuth(), h.AdminReferenceIndex)
	return a
}

func TestAdminReferenceAdminRendersReferenceSections(t *testing.T) {
	description := "Electrical systems"
	tradeTypeID := uuid.New()
	store := &referenceStoreStub{
		tradeTypes: []entities.TradeType{
			{Code: "ELEC", Name: "Electrical", Description: &description, Active: true},
		},
		categories: []entities.ViolationCategory{
			{
				Code:            "ELEC_NO_GFCI",
				Name:            "Missing GFCI Protection",
				TradeTypeID:     &tradeTypeID,
				DefaultSeverity: "High",
				Description:     &description,
				Active:          true,
			},
		},
		rules: []entities.ComplianceRule{
			{
				Code:        "OPEN_VIOLATION_GATE",
				Name:        "Open Violation Closure Gate",
				Description: "Blocks closure with open violations",
				RuleKind:    "ClosureGate",
				Parameters:  []byte(`{"blocking_statuses":["Open","InProgress"]}`),
				Active:      true,
			},
		},
	}
	actor := &app.Actor{ID: uuid.New(), Username: "aisha", DisplayName: "Aisha Patel", Role: string(domain.RoleAdmin)}
	a := buildAdminReferenceApp(actor, store)

	req, _ := http.NewRequest(http.MethodGet, "/admin/reference", nil)
	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	html := string(body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", resp.StatusCode, html)
	}
	if strings.Index(html, "<h2>Trade Types</h2>") > strings.Index(html, "<h2>Violation Categories</h2>") {
		t.Fatal("Trade Types section must precede Violation Categories")
	}
	if strings.Index(html, "<h2>Violation Categories</h2>") > strings.Index(html, "<h2>Compliance Rules</h2>") {
		t.Fatal("Violation Categories section must precede Compliance Rules")
	}
	if strings.Count(html, "<tbody>") != 3 {
		t.Fatalf("expected three tables, got body:\n%s", html)
	}
	if !strings.Contains(htmlpkg.UnescapeString(html), `{"blocking_statuses":["Open","InProgress"]}`) {
		t.Fatal("expected compact JSON parameters disclosure")
	}
}

func TestAdminReferenceNonAdminReturns403WithoutReferenceState(t *testing.T) {
	for _, role := range []domain.Role{
		domain.RoleComplianceOfficer,
		domain.RoleInspector,
		domain.RoleSiteSupervisor,
		domain.RoleExecutive,
	} {
		t.Run(string(role), func(t *testing.T) {
			store := &referenceStoreStub{}
			actor := &app.Actor{ID: uuid.New(), Username: "user", DisplayName: "Role User", Role: string(role)}
			a := buildAdminReferenceApp(actor, store)

			req, _ := http.NewRequest(http.MethodGet, "/admin/reference", nil)
			resp, err := a.Test(req)
			if err != nil {
				t.Fatalf("test request failed: %v", err)
			}
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			html := string(body)

			if resp.StatusCode != http.StatusForbidden {
				t.Fatalf("expected 403, got %d; body: %s", resp.StatusCode, html)
			}
			if store.called {
				t.Fatal("reference store must not be called for non-admin users")
			}
			for _, protected := range []string{"ELEC", "ELEC_NO_GFCI", "OPEN_VIOLATION_GATE", "rule_kind", "parameters"} {
				if strings.Contains(html, protected) {
					t.Fatalf("403 leaked protected string %q in body: %s", protected, html)
				}
			}
		})
	}
}
