package domain_test

import (
	"testing"

	"github.com/code-chimp/fieldmark-go/internal/domain"
)

func TestRoleLabel(t *testing.T) {
	cases := []struct {
		role domain.Role
		want string
	}{
		{domain.RoleAdmin, "Admin"},
		{domain.RoleComplianceOfficer, "Compliance Officer"},
		{domain.RoleInspector, "Inspector"},
		{domain.RoleSiteSupervisor, "Site Supervisor"},
		{domain.RoleExecutive, "Executive"},
	}
	for _, tc := range cases {
		if got := tc.role.Label(); got != tc.want {
			t.Errorf("Role(%q).Label() = %q, want %q", tc.role, got, tc.want)
		}
	}
}

func TestRoleLabelUnknown(t *testing.T) {
	if got := domain.Role("UNKNOWN").Label(); got != "" {
		t.Errorf("Role(UNKNOWN).Label() = %q, want %q", got, "")
	}
}

func TestRoleBadgeToken(t *testing.T) {
	cases := []struct {
		role domain.Role
		want string
	}{
		{domain.RoleAdmin, "danger"},
		{domain.RoleComplianceOfficer, "info"},
		{domain.RoleInspector, "warning"},
		{domain.RoleSiteSupervisor, "neutral"},
		{domain.RoleExecutive, "success"},
	}
	for _, tc := range cases {
		if got := tc.role.BadgeToken(); got != tc.want {
			t.Errorf("Role(%q).BadgeToken() = %q, want %q", tc.role, got, tc.want)
		}
	}
}
