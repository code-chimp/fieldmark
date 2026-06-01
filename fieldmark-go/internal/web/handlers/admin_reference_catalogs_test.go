package handlers_test

import (
	"context"
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

type referenceCatalogStoreStub struct {
	tradeTypes []entities.TradeType
	categories []entities.ViolationCategory
	rules      []entities.ComplianceRule

	tradeCalls    int
	categoryCalls int
	ruleCalls     int
}

func (s *referenceCatalogStoreStub) ListTradeTypes(context.Context) ([]entities.TradeType, error) {
	s.tradeCalls++
	return s.tradeTypes, nil
}

func (s *referenceCatalogStoreStub) ListViolationCategories(context.Context) ([]entities.ViolationCategory, error) {
	s.categoryCalls++
	return s.categories, nil
}

func (s *referenceCatalogStoreStub) ListComplianceRules(context.Context) ([]entities.ComplianceRule, error) {
	s.ruleCalls++
	return s.rules, nil
}

func buildAdminReferenceCatalogApp(actor *app.Actor, store *referenceCatalogStoreStub) *fiber.App {
	a := newTestApp()
	if actor != nil {
		a.Use(injectActor(actor))
	}
	h := &handlers.AdminReferenceHandlers{Reference: store}
	a.Get("/admin/reference/trade-types", auth.RequireAuth(), h.TradeTypesIndex)
	a.Get("/admin/reference/violation-categories", auth.RequireAuth(), h.ViolationCategoriesIndex)
	a.Get("/admin/reference/compliance-rules", auth.RequireAuth(), h.ComplianceRulesIndex)
	return a
}

func TestAdminReferenceCatalogsAdminRendersPages(t *testing.T) {
	description := "Electrical systems"
	tradeTypeID := uuid.New()
	store := &referenceCatalogStoreStub{
		tradeTypes: []entities.TradeType{{Code: "ELEC", Name: "Electrical", Description: &description, Active: true}},
		categories: []entities.ViolationCategory{{
			Code:            "ELEC_NO_GFCI",
			Name:            "Missing GFCI Protection",
			TradeTypeID:     &tradeTypeID,
			DefaultSeverity: "High",
			Description:     &description,
			Active:          true,
		}},
		rules: []entities.ComplianceRule{{
			Code:        "OPEN_VIOLATION_GATE",
			Name:        "Open Violation Closure Gate",
			Description: "Blocks closure with open violations",
			RuleKind:    "ClosureGate",
			Parameters:  []byte(`{"blocking_statuses":["Open","InProgress"]}`),
			Active:      true,
		}},
	}
	actor := &app.Actor{ID: uuid.New(), Username: "aisha", DisplayName: "Aisha Patel", Role: string(domain.RoleAdmin)}
	a := buildAdminReferenceCatalogApp(actor, store)

	cases := []struct {
		path     string
		heading  string
		expected string
		absentHref string
	}{
		{path: "/admin/reference/trade-types", heading: "<h1>Trade Types</h1>", expected: "ELEC", absentHref: `href="/admin/reference/trade-types"`},
		{path: "/admin/reference/violation-categories", heading: "<h1>Violation Categories</h1>", expected: "ELEC_NO_GFCI", absentHref: `href="/admin/reference/violation-categories"`},
		{path: "/admin/reference/compliance-rules", heading: "<h1>Compliance Rules</h1>", expected: "OPEN_VIOLATION_GATE", absentHref: `href="/admin/reference/compliance-rules"`},
	}

	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, tc.path, nil)
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
			if !strings.Contains(html, tc.heading) || !strings.Contains(html, tc.expected) {
				t.Fatalf("expected heading/content missing for %s; body: %s", tc.path, html)
			}
			if !strings.Contains(html, `aria-label="Reference catalogs"`) {
				t.Fatalf("expected reference sub-nav in %s", tc.path)
			}
			if strings.Contains(html, tc.absentHref) {
				t.Fatalf("unexpected self-link %q in %s", tc.absentHref, tc.path)
			}
		})
	}

	if store.tradeCalls != 1 || store.categoryCalls != 1 || store.ruleCalls != 1 {
		t.Fatalf("expected exactly one call per catalog read method, got trade=%d category=%d rule=%d", store.tradeCalls, store.categoryCalls, store.ruleCalls)
	}
}

func TestAdminReferenceCatalogsNonAdminReturns403WithoutReferenceState(t *testing.T) {
	for _, role := range []domain.Role{domain.RoleComplianceOfficer, domain.RoleInspector, domain.RoleSiteSupervisor, domain.RoleExecutive} {
		for _, path := range []string{"/admin/reference/trade-types", "/admin/reference/violation-categories", "/admin/reference/compliance-rules"} {
			t.Run(string(role)+"_"+path, func(t *testing.T) {
				store := &referenceCatalogStoreStub{}
				actor := &app.Actor{ID: uuid.New(), Username: "user", DisplayName: "Role User", Role: string(role)}
				a := buildAdminReferenceCatalogApp(actor, store)

				req, _ := http.NewRequest(http.MethodGet, path, nil)
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
				if store.tradeCalls != 0 || store.categoryCalls != 0 || store.ruleCalls != 0 {
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
}

func TestAdminReferenceCatalogsEmptyState(t *testing.T) {
	store := &referenceCatalogStoreStub{}
	actor := &app.Actor{ID: uuid.New(), Username: "aisha", DisplayName: "Aisha Patel", Role: string(domain.RoleAdmin)}
	a := buildAdminReferenceCatalogApp(actor, store)

	cases := []struct {
		path    string
		emptyMsg string
	}{
		{path: "/admin/reference/trade-types", emptyMsg: "No trade types defined."},
		{path: "/admin/reference/violation-categories", emptyMsg: "No violation categories defined."},
		{path: "/admin/reference/compliance-rules", emptyMsg: "No compliance rules defined."},
	}

	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, tc.path, nil)
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
			if !strings.Contains(html, tc.emptyMsg) {
				t.Fatalf("expected empty-state message %q; body: %s", tc.emptyMsg, html)
			}
		})
	}
}
