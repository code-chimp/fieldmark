package auth

import (
	"testing"

	"github.com/google/uuid"

	"github.com/code-chimp/fieldmark-go/internal/app"
	"github.com/code-chimp/fieldmark-go/internal/domain"
)

func TestCan_AnonymousActor_ReturnsFalse(t *testing.T) {
	resetForTests()
	RegisterAction("test.allow_admin", domain.RoleAdmin)

	if Can(nil, "test.allow_admin", uuid.Nil) {
		t.Fatal("Can: expected false for nil actor")
	}
	anon := app.Anonymous()
	if Can(anon, "test.allow_admin", uuid.Nil) {
		t.Fatal("Can: expected false for anonymous actor")
	}
}

func TestCan_AdminActor_ReturnsTrueForAdminScopedAction(t *testing.T) {
	resetForTests()
	RegisterAction("test.allow_admin", domain.RoleAdmin)

	actor := &app.Actor{
		ID:       uuid.New(),
		Username: "aisha",
		Role:     string(domain.RoleAdmin),
	}
	if !Can(actor, "test.allow_admin", uuid.Nil) {
		t.Fatal("Can: expected true for ADMIN actor on admin-scoped action")
	}
}

func TestCan_NonAdminActor_ReturnsFalseForAdminScopedAction(t *testing.T) {
	resetForTests()
	RegisterAction("test.allow_admin", domain.RoleAdmin)

	actor := &app.Actor{
		ID:       uuid.New(),
		Username: "pat",
		Role:     string(domain.RoleSiteSupervisor),
	}
	if Can(actor, "test.allow_admin", uuid.Nil) {
		t.Fatal("Can: expected false for SITE_SUPERVISOR actor on admin-scoped action")
	}
}

func TestCan_UnknownAction_ReturnsFalse(t *testing.T) {
	resetForTests()

	actor := &app.Actor{
		ID:       uuid.New(),
		Username: "aisha",
		Role:     string(domain.RoleAdmin),
	}
	if Can(actor, "test.unmapped", uuid.Nil) {
		t.Fatal("Can: expected false for action not in the map")
	}
}

func TestProjectActions_AdminOnly(t *testing.T) {
	resetForTests()
	for _, action := range []string{"project.place_on_hold", "project.resume", "project.close"} {
		RegisterAction(action, domain.RoleAdmin)
	}

	admin := &app.Actor{ID: uuid.New(), Username: "aisha", Role: string(domain.RoleAdmin)}
	executive := &app.Actor{ID: uuid.New(), Username: "eli", Role: string(domain.RoleExecutive)}
	inspector := &app.Actor{ID: uuid.New(), Username: "ravi", Role: string(domain.RoleInspector)}

	for _, action := range []string{"project.place_on_hold", "project.resume", "project.close"} {
		if !Can(admin, action, uuid.Nil) {
			t.Fatalf("expected admin to have %s", action)
		}
		if Can(executive, action, uuid.Nil) {
			t.Fatalf("expected executive to lack %s", action)
		}
		if Can(inspector, action, uuid.Nil) {
			t.Fatalf("expected inspector to lack %s", action)
		}
	}
}
