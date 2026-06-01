package handlers

import (
	"context"
	"log/slog"

	"github.com/gofiber/fiber/v3"

	"github.com/code-chimp/fieldmark-go/internal/app"
	"github.com/code-chimp/fieldmark-go/internal/domain"
	"github.com/code-chimp/fieldmark-go/internal/domain/entities"
	"github.com/code-chimp/fieldmark-go/internal/web/auth"
	"github.com/code-chimp/fieldmark-go/internal/web/viewmodels"
)

type ReferenceStore interface {
	ListTradeTypes(ctx context.Context) ([]entities.TradeType, error)
	ListViolationCategories(ctx context.Context) ([]entities.ViolationCategory, error)
	ListComplianceRules(ctx context.Context) ([]entities.ComplianceRule, error)
}

type AdminReferenceHandlers struct {
	Reference ReferenceStore
}

type complianceRuleRow struct {
	Code           string
	Name           string
	Description    string
	RuleKind       string
	ParametersJSON string
	Active         bool
}

type tradeTypeRow struct {
	Code        string
	Name        string
	Description string
	Active      bool
}

type violationCategoryRow struct {
	Code            string
	Name            string
	TradeTypeID     string
	DefaultSeverity string
	Description     string
	Active          bool
}

func (h *AdminReferenceHandlers) AdminReferenceIndex(c fiber.Ctx) error {
	actor := auth.ActorFromCtx(c)
	if actor == nil || actor.Role != string(domain.RoleAdmin) {
		c.Status(fiber.StatusForbidden)
		return c.SendString("You do not have permission to access this page.")
	}

	tradeTypes, err := h.Reference.ListTradeTypes(c.Context())
	if err != nil {
		return err
	}
	categories, err := h.Reference.ListViolationCategories(c.Context())
	if err != nil {
		return err
	}
	rules, err := h.Reference.ListComplianceRules(c.Context())
	if err != nil {
		return err
	}

	m := referenceBaseMap(c)
	m["Title"] = "Reference Data"
	m["TradeTypes"] = tradeTypeRows(tradeTypes)
	m["ViolationCategories"] = violationCategoryRows(categories)
	ruleRows := make([]complianceRuleRow, 0, len(rules))
	for _, rule := range rules {
		ruleRows = append(ruleRows, complianceRuleRow{
			Code:           rule.Code,
			Name:           rule.Name,
			Description:    rule.Description,
			RuleKind:       rule.RuleKind,
			ParametersJSON: string(rule.Parameters),
			Active:         rule.Active,
		})
	}
	m["ComplianceRules"] = ruleRows

	return c.Render("pages/admin_reference", m)
}

func (h *AdminReferenceHandlers) TradeTypesIndex(c fiber.Ctx) error {
	actor := auth.ActorFromCtx(c)
	if actor == nil || actor.Role != string(domain.RoleAdmin) {
		c.Status(fiber.StatusForbidden)
		return c.SendString("You do not have permission to access this page.")
	}

	tradeTypes, err := h.Reference.ListTradeTypes(c.Context())
	if err != nil {
		return err
	}

	m := referenceBaseMap(c)
	m["Title"] = "Trade Types"
	m["TradeTypes"] = tradeTypeRows(tradeTypes)
	return c.Render("pages/admin_reference_trade_types", m)
}

func (h *AdminReferenceHandlers) ViolationCategoriesIndex(c fiber.Ctx) error {
	actor := auth.ActorFromCtx(c)
	if actor == nil || actor.Role != string(domain.RoleAdmin) {
		c.Status(fiber.StatusForbidden)
		return c.SendString("You do not have permission to access this page.")
	}

	categories, err := h.Reference.ListViolationCategories(c.Context())
	if err != nil {
		return err
	}

	m := referenceBaseMap(c)
	m["Title"] = "Violation Categories"
	m["ViolationCategories"] = violationCategoryRows(categories)
	return c.Render("pages/admin_reference_violation_categories", m)
}

func (h *AdminReferenceHandlers) ComplianceRulesIndex(c fiber.Ctx) error {
	actor := auth.ActorFromCtx(c)
	if actor == nil || actor.Role != string(domain.RoleAdmin) {
		c.Status(fiber.StatusForbidden)
		return c.SendString("You do not have permission to access this page.")
	}

	rules, err := h.Reference.ListComplianceRules(c.Context())
	if err != nil {
		return err
	}

	m := referenceBaseMap(c)
	m["Title"] = "Compliance Rules"
	ruleRows := make([]complianceRuleRow, 0, len(rules))
	for _, rule := range rules {
		ruleRows = append(ruleRows, complianceRuleRow{
			Code:           rule.Code,
			Name:           rule.Name,
			Description:    rule.Description,
			RuleKind:       rule.RuleKind,
			ParametersJSON: string(rule.Parameters),
			Active:         rule.Active,
		})
	}
	m["ComplianceRules"] = ruleRows
	return c.Render("pages/admin_reference_compliance_rules", m)
}

func tradeTypeRows(tradeTypes []entities.TradeType) []tradeTypeRow {
	rows := make([]tradeTypeRow, 0, len(tradeTypes))
	for _, tradeType := range tradeTypes {
		rows = append(rows, tradeTypeRow{
			Code:        tradeType.Code,
			Name:        tradeType.Name,
			Description: optionalString(tradeType.Description),
			Active:      tradeType.Active,
		})
	}
	return rows
}

func violationCategoryRows(categories []entities.ViolationCategory) []violationCategoryRow {
	rows := make([]violationCategoryRow, 0, len(categories))
	for _, category := range categories {
		tradeTypeID := ""
		if category.TradeTypeID != nil {
			tradeTypeID = category.TradeTypeID.String()
		}
		rows = append(rows, violationCategoryRow{
			Code:            category.Code,
			Name:            category.Name,
			TradeTypeID:     tradeTypeID,
			DefaultSeverity: category.DefaultSeverity,
			Description:     optionalString(category.Description),
			Active:          category.Active,
		})
	}
	return rows
}

func optionalString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func referenceBaseMap(c fiber.Ctx) fiber.Map {
	theme, next := themeEntries(c)
	actor := auth.ActorFromCtx(c)
	if actor == nil {
		actor = app.Anonymous()
	}

	role := domain.Role(actor.Role)
	badgeToken := role.BadgeToken()
	if badgeToken == "unknown" && actor.Role != "" {
		slog.Warn("unknown role badge token", "role", actor.Role)
	}
	return fiber.Map{
		"FmTheme":         theme,
		"FmThemeNext":     next,
		"FmThemeResolved": theme,
		"Actor":           actor,
		"RoleLabel":       role.Label(),
		"RoleBadgeToken":  badgeToken,
		"FullName":        actor.DisplayName,
		"Initials":        viewmodels.Initials(actor.DisplayName, actor.Username),
	}
}
