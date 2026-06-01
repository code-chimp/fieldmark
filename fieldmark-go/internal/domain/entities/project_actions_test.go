package entities_test

import (
	"testing"

	"github.com/code-chimp/fieldmark-go/internal/domain/entities"
	"github.com/code-chimp/fieldmark-go/internal/domain/enums"
)

func TestProjectActionPredicates_Active(t *testing.T) {
	project := entities.Project{Status: enums.ProjectStatusActive}
	if !project.CanPlaceOnHold() {
		t.Fatal("CanPlaceOnHold should be true for Active")
	}
	if project.CanResume() {
		t.Fatal("CanResume should be false for Active")
	}
	if !project.CanClose() {
		t.Fatal("CanClose should be true for Active")
	}
}

func TestProjectActionPredicates_OnHold(t *testing.T) {
	project := entities.Project{Status: enums.ProjectStatusOnHold}
	if project.CanPlaceOnHold() {
		t.Fatal("CanPlaceOnHold should be false for OnHold")
	}
	if !project.CanResume() {
		t.Fatal("CanResume should be true for OnHold")
	}
	if project.CanClose() {
		t.Fatal("CanClose should be false for OnHold")
	}
}

func TestProjectActionPredicates_Closed(t *testing.T) {
	project := entities.Project{Status: enums.ProjectStatusClosed}
	if project.CanPlaceOnHold() {
		t.Fatal("CanPlaceOnHold should be false for Closed")
	}
	if project.CanResume() {
		t.Fatal("CanResume should be false for Closed")
	}
	if project.CanClose() {
		t.Fatal("CanClose should be false for Closed")
	}
}
