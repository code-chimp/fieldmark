package entities_test

import (
	"errors"
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

func TestProjectPlaceOnHold_ActiveToOnHold(t *testing.T) {
	project := entities.Project{Status: enums.ProjectStatusActive}
	if err := project.PlaceOnHold("maintenance window"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if project.Status != enums.ProjectStatusOnHold {
		t.Fatalf("status=%q want %q", project.Status, enums.ProjectStatusOnHold)
	}
}

func TestProjectPlaceOnHold_InvalidStates(t *testing.T) {
	for _, status := range []enums.ProjectStatus{enums.ProjectStatusOnHold, enums.ProjectStatusClosed} {
		t.Run(string(status), func(t *testing.T) {
			project := entities.Project{Status: status}
			err := project.PlaceOnHold("maintenance window")
			if err == nil {
				t.Fatalf("status=%q expected error", status)
			}
			if !errors.Is(err, entities.ErrInvalidProjectTransition) {
				t.Errorf("status=%q expected ErrInvalidProjectTransition got %v", status, err)
			}
			if err.Error() != "Project is already on hold" {
				t.Errorf("status=%q unexpected message: %v", status, err)
			}
		})
	}
}

func TestProjectResume_OnHoldToActive(t *testing.T) {
	project := entities.Project{Status: enums.ProjectStatusOnHold}
	if err := project.Resume("back online"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if project.Status != enums.ProjectStatusActive {
		t.Fatalf("status=%q want %q", project.Status, enums.ProjectStatusActive)
	}
}

func TestProjectResume_InvalidStates(t *testing.T) {
	for _, status := range []enums.ProjectStatus{enums.ProjectStatusActive, enums.ProjectStatusClosed} {
		t.Run(string(status), func(t *testing.T) {
			project := entities.Project{Status: status}
			err := project.Resume("back online")
			if err == nil {
				t.Fatalf("status=%q expected error", status)
			}
			if !errors.Is(err, entities.ErrInvalidProjectTransition) {
				t.Errorf("status=%q expected ErrInvalidProjectTransition got %v", status, err)
			}
			if err.Error() != "Project is not on hold" {
				t.Errorf("status=%q unexpected message: %v", status, err)
			}
		})
	}
}
