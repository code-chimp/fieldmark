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
