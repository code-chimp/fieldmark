package components_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/code-chimp/fieldmark-go/internal/web/testutil"
	"github.com/code-chimp/fieldmark-go/internal/web/viewmodels"
)

func TestStatusBadgeVariantsMatchCanonical(t *testing.T) {
	cases := map[string]viewmodels.StatusBadgeVM{
		"project-active":                   {ClassName: "badge-project-active", Label: "Active"},
		"project-on-hold":                  {ClassName: "badge-project-onhold", Label: "On Hold"},
		"project-closed":                   {ClassName: "badge-project-closed", Label: "Closed"},
		"inspection-scheduled":             {ClassName: "badge-inspection-scheduled", Label: "Scheduled"},
		"inspection-in-progress":           {ClassName: "badge-inspection-inprogress", Label: "In Progress"},
		"inspection-completed-pass":        {ClassName: "badge-inspection-pass", Label: "Pass"},
		"inspection-completed-conditional": {ClassName: "badge-inspection-conditional", Label: "Conditional"},
		"inspection-completed-fail":        {ClassName: "badge-inspection-fail", Label: "Fail"},
		"inspection-cancelled":             {ClassName: "badge-inspection-cancelled", Label: "Cancelled"},
		"violation-open-critical-high":     {ClassName: "badge-violation-open-high", Label: "Open"},
		"violation-open-medium-low":        {ClassName: "badge-violation-open-low", Label: "Open"},
		"violation-in-progress":            {ClassName: "badge-violation-inprogress", Label: "In Progress"},
		"violation-resolved":               {ClassName: "badge-violation-resolved", Label: "Resolved"},
		"violation-voided":                 {ClassName: "badge-violation-voided", Label: "Voided"},
		"corrective-action-submitted":      {ClassName: "badge-ca-submitted", Label: "Submitted"},
		"corrective-action-under-review":   {ClassName: "badge-ca-underreview", Label: "Under Review"},
		"corrective-action-approved":       {ClassName: "badge-ca-approved", Label: "Approved"},
		"corrective-action-rejected":       {ClassName: "badge-ca-rejected", Label: "Rejected"},
		"severity-critical":                {ClassName: "badge-severity-critical badge-bump", Label: "Critical"},
		"severity-high":                    {ClassName: "badge-severity-high badge-bump", Label: "High"},
		"severity-medium":                  {ClassName: "badge-severity-medium", Label: "Medium"},
		"severity-low":                     {ClassName: "badge-severity-low", Label: "Low"},
		"unknown":                          {ClassName: "badge-unknown", Label: "Foobar"},
	}
	for variant, vm := range cases {
		t.Run(variant, func(t *testing.T) {
			assertComponentSnapshot(t, "status_badge", "status_badge", variant, vm)
		})
	}
}

func TestStatusBadgeUnknownFallbackClass(t *testing.T) {
	html := testutil.NormaliseComponent(renderComponent(t, "status_badge", viewmodels.StatusBadgeVM{ClassName: "badge-unknown", Label: "Foobar"}))
	if !strings.Contains(html, "badge-unknown") {
		t.Fatalf("unknown status badge must include fallback class, got %q", html)
	}
}

func TestStatusBadgeOpenLowClassMatchesMediumLowVariant(t *testing.T) {
	vm := viewmodels.StatusBadgeVM{ClassName: "badge-violation-open-low", Label: "Open"}
	assertComponentSnapshot(t, "status_badge", "status_badge", "violation-open-medium-low", vm)
}

func TestStatusBadgeTemplateDoesNotUseTemplateHTML(t *testing.T) {
	_, thisFile, _, _ := runtime.Caller(0)
	b, err := os.ReadFile(filepath.Join(filepath.Dir(thisFile), "status_badge.html"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "template.HTML(") {
		t.Fatal("status badge template must not use template.HTML")
	}
}
