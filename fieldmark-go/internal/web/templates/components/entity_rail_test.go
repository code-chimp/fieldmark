package components_test

import (
	"html/template"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	components "github.com/code-chimp/fieldmark-go/internal/web/templates/components"
)

func emptyRailArgs(id, entityTypeLabel string) components.EntityRailArgs {
	return components.EntityRailArgs{
		ID:              id,
		EntityTypeLabel: entityTypeLabel,
		EntityLoaded:    false,
	}
}

func loadedRailArgs(id, entityTypeLabel string, body, footer template.HTML) components.EntityRailArgs {
	return components.EntityRailArgs{
		ID:              id,
		EntityTypeLabel: entityTypeLabel,
		EntityLoaded:    true,
		BodySlot:        body,
		FooterSlot:      footer,
	}
}

func railHTML(t *testing.T, args components.EntityRailArgs) string {
	t.Helper()
	return renderComponent(t, "entity_rail", args)
}

// TestEntityRailTemplateDoesNotUseTemplateHTMLCast asserts that entity_rail.html
// does not contain template.HTML( — the template relies on the typed field, not casts.
func TestEntityRailTemplateDoesNotUseTemplateHTMLCast(t *testing.T) {
	_, thisFile, _, _ := runtime.Caller(0)
	b, err := os.ReadFile(filepath.Join(filepath.Dir(thisFile), "entity_rail.html"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "template.HTML(") {
		t.Fatal("entity_rail.html must not use template.HTML( — use the typed EntityRailArgs field instead")
	}
}

// TestEntityRailOtherTemplatesDoNotUseTemplateHTMLCast extends the guard to all other component html files.
func TestEntityRailOtherTemplatesDoNotUseTemplateHTMLCast(t *testing.T) {
	_, thisFile, _, _ := runtime.Caller(0)
	dir := filepath.Dir(thisFile)
	for _, filename := range []string{
		"action_button.html",
		"audit_row.html",
		"compliance_tile.html",
		"dashboard_tile.html",
		"inline_alert.html",
		"status_badge.html",
	} {
		b, err := os.ReadFile(filepath.Join(dir, filename))
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(b), "template.HTML(") {
			t.Fatalf("%s must not use template.HTML(", filename)
		}
	}
}

// TestEntityRailVariantsMatchCanonical asserts byte-equality against the six canonical variants.
func TestEntityRailVariantsMatchCanonical(t *testing.T) {
	cases := map[string]components.EntityRailArgs{
		"empty-violation":                emptyRailArgs("violation-detail", "Violation"),
		"empty-inspection":               emptyRailArgs("inspection-detail", "Inspection"),
		"empty-corrective-action":        emptyRailArgs("corrective-action-detail", "Corrective Action"),
		"loaded-shell-violation":         loadedRailArgs("violation-detail", "Violation", "__BODY__", "__FOOTER__"),
		"loaded-shell-inspection":        loadedRailArgs("inspection-detail", "Inspection", "__BODY__", "__FOOTER__"),
		"loaded-shell-corrective-action": loadedRailArgs("corrective-action-detail", "Corrective Action", "__BODY__", "__FOOTER__"),
	}
	for variant, args := range cases {
		t.Run(variant, func(t *testing.T) {
			assertComponentSnapshot(t, "entity_rail", "entity_rail", variant, args)
		})
	}
}

// AC4 — four-case slot / footer-omission coverage

func TestEntityRailEmptyStateRendersEmptyCard(t *testing.T) {
	html := railHTML(t, emptyRailArgs("violation-detail", "Violation"))
	if !strings.Contains(html, "entity-rail--empty") {
		t.Errorf("expected entity-rail--empty in output:\n%s", html)
	}
	if !strings.Contains(html, "Empty entity rail") {
		t.Errorf("expected empty-state aria-label in output:\n%s", html)
	}
	if !strings.Contains(html, "Select an entity to see its detail here.") {
		t.Errorf("expected empty-state text in output:\n%s", html)
	}
	if strings.Contains(html, "entity-rail__body") {
		t.Errorf("empty state must not render entity-rail__body:\n%s", html)
	}
}

func TestEntityRailLoadedWithBothSlotsRendersBodyAndFooter(t *testing.T) {
	args := loadedRailArgs("violation-detail", "Violation", "<p>body</p>", "<button>Save</button>")
	html := railHTML(t, args)
	if !strings.Contains(html, "entity-rail--loaded") {
		t.Errorf("expected entity-rail--loaded:\n%s", html)
	}
	if !strings.Contains(html, "<p>body</p>") {
		t.Errorf("expected body slot content:\n%s", html)
	}
	if !strings.Contains(html, "<button>Save</button>") {
		t.Errorf("expected footer slot content:\n%s", html)
	}
	if !strings.Contains(html, "entity-rail__footer") {
		t.Errorf("expected entity-rail__footer:\n%s", html)
	}
}

func TestEntityRailLoadedBodyOnlyOmitsFooterDiv(t *testing.T) {
	args := loadedRailArgs("violation-detail", "Violation", "<p>body</p>", "")
	html := railHTML(t, args)
	if !strings.Contains(html, "entity-rail__body") {
		t.Errorf("expected entity-rail__body:\n%s", html)
	}
	if strings.Contains(html, "entity-rail__footer") {
		t.Errorf("footer div must be omitted when FooterSlot is empty:\n%s", html)
	}
}

func TestEntityRailLoadedNoSlotsRendersHeaderAndEmptyBody(t *testing.T) {
	args := loadedRailArgs("violation-detail", "Violation", "", "")
	html := railHTML(t, args)
	if !strings.Contains(html, "entity-rail--loaded") {
		t.Errorf("expected entity-rail--loaded:\n%s", html)
	}
	if !strings.Contains(html, "entity-rail__header") {
		t.Errorf("expected entity-rail__header:\n%s", html)
	}
	if !strings.Contains(html, "entity-rail__body") {
		t.Errorf("expected entity-rail__body:\n%s", html)
	}
	if strings.Contains(html, "entity-rail__footer") {
		t.Errorf("footer div must be omitted when FooterSlot is empty:\n%s", html)
	}
}

// AC8 — XSS round-trip: EntityTypeLabel is framework-escaped (non-slot prop)

func TestEntityRailXSSPayloadInEntityTypeLabelIsEscaped(t *testing.T) {
	html := railHTML(t, emptyRailArgs("violation-detail", "<script>alert(1)</script>"))
	if strings.Contains(html, "<script>alert(1)</script>") {
		t.Error("XSS payload must be escaped in EntityTypeLabel")
	}
	if strings.Contains(html, "<script>") {
		t.Errorf("raw <script> tag must not appear in output:\n%s", html)
	}
	if !strings.Contains(html, "&lt;script&gt;") {
		t.Errorf("expected escaped payload in output:\n%s", html)
	}
}

func TestEntityRailXSSPayloadInLoadedLabelEscapedInSpanAndAriaLabel(t *testing.T) {
	args := loadedRailArgs("violation-detail", "<script>alert(1)</script>", "", "")
	html := railHTML(t, args)
	if strings.Contains(html, "<script>") {
		t.Errorf("raw <script> tag must not appear in loaded output:\n%s", html)
	}
	if !strings.Contains(html, "&lt;script&gt;") {
		t.Errorf("expected escaped payload in loaded output:\n%s", html)
	}
}

// AC8 §category 9 — empty/whitespace EntityTypeLabel does not crash

func TestEntityRailWhitespaceOrEmptyLabelDoesNotCrash(t *testing.T) {
	for _, label := range []string{"", "   "} {
		html := railHTML(t, emptyRailArgs("violation-detail", label))
		if !strings.Contains(html, "<aside") {
			t.Errorf("label=%q: expected valid <aside> in output:\n%s", label, html)
		}
		if !strings.Contains(html, "entity-rail") {
			t.Errorf("label=%q: expected entity-rail class in output:\n%s", label, html)
		}
	}
}

// AC3 — no HTMX producer attributes on dismiss button

func TestEntityRailLoadedShellDismissButtonHasNoHtmxAttributes(t *testing.T) {
	html := railHTML(t, loadedRailArgs("violation-detail", "Violation", "__BODY__", "__FOOTER__"))
	for _, forbidden := range []string{"hx-get", "hx-post", "hx-target", "hx-swap", "hx-trigger", "onclick="} {
		if strings.Contains(html, forbidden) {
			t.Errorf("found forbidden token %q in dismiss button output:\n%s", forbidden, html)
		}
	}
}
