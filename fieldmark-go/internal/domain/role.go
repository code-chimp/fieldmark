package domain

// Role is the single Go-side source of truth for the five canonical
// conceptual-role names; mirrored in dotnet_auth, django_auth, and
// fiber_auth.user_roles.role (CHECK constraint).
type Role string

const (
	RoleAdmin             Role = "ADMIN"
	RoleComplianceOfficer Role = "COMPLIANCE_OFFICER"
	RoleInspector         Role = "INSPECTOR"
	RoleSiteSupervisor    Role = "SITE_SUPERVISOR"
	RoleExecutive         Role = "EXECUTIVE"
)

// AllRoles enumerates the canonical names in deterministic order.
var AllRoles = []Role{
	RoleAdmin, RoleComplianceOfficer, RoleInspector, RoleSiteSupervisor, RoleExecutive,
}

// Label returns the title-cased display label for the role (AC #4, Story 1.13).
func (r Role) Label() string {
	switch r {
	case RoleAdmin:
		return "Admin"
	case RoleComplianceOfficer:
		return "Compliance Officer"
	case RoleInspector:
		return "Inspector"
	case RoleSiteSupervisor:
		return "Site Supervisor"
	case RoleExecutive:
		return "Executive"
	default:
		return ""
	}
}

// BadgeToken returns the CSS badge modifier token for the role (AC #4, Story 1.13).
func (r Role) BadgeToken() string {
	switch r {
	case RoleAdmin:
		return "danger"
	case RoleComplianceOfficer:
		return "info"
	case RoleInspector:
		return "warning"
	case RoleSiteSupervisor:
		return "neutral"
	case RoleExecutive:
		return "success"
	default:
		return "neutral"
	}
}
