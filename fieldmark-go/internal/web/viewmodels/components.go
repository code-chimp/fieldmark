package viewmodels

import "strings"

// StatusBadgeVM is the data context for the status_badge component template.
type StatusBadgeVM struct {
	Entity    string
	Value     string
	Severity  string
	ClassName string
	Label     string
}

// ResolveStatusBadge maps entity/value to the shared status badge presentation.
func ResolveStatusBadge(entity, value string) StatusBadgeVM {
	vm := StatusBadgeVM{
		Entity:    entity,
		Value:     value,
		ClassName: "badge-unknown",
		Label:     strings.TrimSpace(value),
	}
	switch strings.ToLower(strings.TrimSpace(entity)) {
	case "project":
		switch strings.ToLower(strings.TrimSpace(value)) {
		case "active":
			vm.ClassName = "badge-project-active"
			vm.Label = "Active"
		case "onhold", "on hold":
			vm.ClassName = "badge-project-onhold"
			vm.Label = "On Hold"
		case "closed":
			vm.ClassName = "badge-project-closed"
			vm.Label = "Closed"
		}
	}
	return vm
}

// InlineAlertVM is the data context for the inline_alert component template.
type InlineAlertVM struct {
	Severity   string
	AlertClass string
	Role       string
	Icon       string
	Title      string
	Message    string
	Meta       string
}

// AuditRowVM is the data context for the audit_row component template.
type AuditRowVM struct {
	Action          string
	ActionClass     string
	ActorName       string
	OccurredAt      string
	Absolute        string
	Relative        string
	BeforeAfterJSON string
	Expanded        bool
}

// ActorDisplay applies the cross-stack AuditRow empty actor fallback.
func (vm AuditRowVM) ActorDisplay() string {
	if strings.TrimSpace(vm.ActorName) == "" {
		return "unnamed"
	}
	return vm.ActorName
}

// ShowInitialsFallback reports whether AuditRow should render the deterministic
// empty actor initials fallback.
func (vm AuditRowVM) ShowInitialsFallback() bool {
	return strings.TrimSpace(vm.ActorName) == ""
}

// DashboardTileVM is the data context for the dashboard_tile component template.
type DashboardTileVM struct {
	TileID       string
	Label        string
	DisplayValue string
	ValueClass   string
	Secondary    string
	RoleStatus   bool
}

// DisplayValueText applies the cross-stack DashboardTile empty value fallback.
func (vm DashboardTileVM) DisplayValueText() string {
	if strings.TrimSpace(vm.DisplayValue) == "" {
		return "—"
	}
	return vm.DisplayValue
}
