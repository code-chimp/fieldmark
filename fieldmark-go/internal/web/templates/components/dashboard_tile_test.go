package components_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/code-chimp/fieldmark-go/internal/web/viewmodels"
)

var dashboardTileFixture = viewmodels.DashboardTileVM{
	TileID:       "open-violations-tile",
	Label:        "Open Violations",
	DisplayValue: "12",
}

func TestDashboardTileTemplateDoesNotUseTemplateHTML(t *testing.T) {
	_, thisFile, _, _ := runtime.Caller(0)
	b, err := os.ReadFile(filepath.Join(filepath.Dir(thisFile), "dashboard_tile.html"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "template.HTML(") {
		t.Fatal("dashboard tile template must not use template.HTML")
	}
}

func TestDashboardTileVariantsMatchCanonical(t *testing.T) {
	cases := map[string]viewmodels.DashboardTileVM{
		"populated": dashboardTileFixture,
		"zero-value": func() viewmodels.DashboardTileVM {
			vm := dashboardTileFixture
			vm.DisplayValue = "0"
			return vm
		}(),
		"populated-with-secondary": func() viewmodels.DashboardTileVM {
			vm := dashboardTileFixture
			vm.Secondary = "3 critical"
			return vm
		}(),
		"populated-with-color": func() viewmodels.DashboardTileVM {
			vm := dashboardTileFixture
			vm.ValueClass = " text-danger"
			return vm
		}(),
		"empty": func() viewmodels.DashboardTileVM {
			vm := dashboardTileFixture
			vm.DisplayValue = "—"
			return vm
		}(),
		"status-region": func() viewmodels.DashboardTileVM {
			vm := dashboardTileFixture
			vm.RoleStatus = true
			return vm
		}(),
	}
	for variant, vm := range cases {
		t.Run(variant, func(t *testing.T) {
			assertComponentSnapshot(t, "dashboard_tile", "dashboard_tile", variant, vm)
		})
	}
}

func TestDashboardTileWhitespaceOnlyValueMatchesEmptyCanonical(t *testing.T) {
	vm := dashboardTileFixture
	vm.DisplayValue = "   "
	assertComponentSnapshot(t, "dashboard_tile", "dashboard_tile", "empty", vm)
}
