package components_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	components "github.com/code-chimp/fieldmark-go/internal/web/templates/components"
)

func intPtr(v int) *int { return &v }

func newTileArgs(score *int, label, id string) components.ComplianceTileArgs {
	return components.NewComplianceTileArgs(score, label, id)
}

func tileHTML(t *testing.T, args components.ComplianceTileArgs) string {
	t.Helper()
	return renderComponent(t, "compliance_tile", args)
}

func TestComplianceTileTemplateDoesNotUseTemplateHTML(t *testing.T) {
	_, thisFile, _, _ := runtime.Caller(0)
	for _, filename := range []string{"compliance_tile.html", "compliance_tile.go"} {
		b, err := os.ReadFile(filepath.Join(filepath.Dir(thisFile), filename))
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(b), "template.HTML(") {
			t.Fatalf("%s must not use template.HTML", filename)
		}
	}
}

func TestComplianceTileVariantsMatchCanonical(t *testing.T) {
	cases := map[string]components.ComplianceTileArgs{
		"healthy-project":    newTileArgs(intPtr(95), "Compliance", "compliance-tile"),
		"watch-project":      newTileArgs(intPtr(82), "Compliance", "compliance-tile"),
		"concern-project":    newTileArgs(intPtr(58), "Compliance", "compliance-tile"),
		"critical-project":   newTileArgs(intPtr(37), "Compliance", "compliance-tile"),
		"healthy-portfolio":  newTileArgs(intPtr(91), "Portfolio Compliance", "compliance-tile-portfolio"),
		"critical-portfolio": newTileArgs(intPtr(42), "Portfolio Compliance", "compliance-tile-portfolio"),
		"no-data-project":    newTileArgs(nil, "Compliance", "compliance-tile"),
		"boundary-90":        newTileArgs(intPtr(90), "Compliance", "compliance-tile"),
		"boundary-70":        newTileArgs(intPtr(70), "Compliance", "compliance-tile"),
		"boundary-50":        newTileArgs(intPtr(50), "Compliance", "compliance-tile"),
		"boundary-49":        newTileArgs(intPtr(49), "Compliance", "compliance-tile"),
	}
	for variant, args := range cases {
		t.Run(variant, func(t *testing.T) {
			assertComponentSnapshot(t, "compliance_tile", "compliance_tile", variant, args)
		})
	}
}

// TestComplianceTileBandBoundaries asserts resolveComplianceBand tuple output directly per AC3.
// Tests the pure band function via ComplianceTileArgs.Band — no HTML rendering.
func TestComplianceTileBandBoundaries(t *testing.T) {
	cases := []struct {
		name            string
		score           *int
		wantValueClass  string
		wantWord        string
		wantThreshClass string
		wantRenderP     bool
	}{
		{"null-no-data", nil, "text-neutral", "", "", false},
		{"100-healthy", intPtr(100), "text-success", "Healthy", "text-success", true},
		{"90-healthy-inclusive", intPtr(90), "text-success", "Healthy", "text-success", true},
		{"89-watch", intPtr(89), "text-warning", "Watch", "text-warning", true},
		{"70-watch-inclusive", intPtr(70), "text-warning", "Watch", "text-warning", true},
		{"69-concern", intPtr(69), "text-warning-strong", "Concern", "text-warning-strong", true},
		{"50-concern-inclusive", intPtr(50), "text-warning-strong", "Concern", "text-warning-strong", true},
		{"49-critical", intPtr(49), "text-danger", "Critical", "text-danger", true},
		{"0-critical", intPtr(0), "text-danger", "Critical", "text-danger", true},
		{"minus1-out-of-range", intPtr(-1), "text-neutral", "", "", false},
		{"101-out-of-range", intPtr(101), "text-neutral", "", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			band := newTileArgs(tc.score, "Compliance", "compliance-tile").Band
			if band.ValueClass != tc.wantValueClass {
				t.Errorf("ValueClass: want %q got %q", tc.wantValueClass, band.ValueClass)
			}
			if band.ThresholdWord != tc.wantWord {
				t.Errorf("ThresholdWord: want %q got %q", tc.wantWord, band.ThresholdWord)
			}
			if band.ThresholdClass != tc.wantThreshClass {
				t.Errorf("ThresholdClass: want %q got %q", tc.wantThreshClass, band.ThresholdClass)
			}
			if band.RenderP != tc.wantRenderP {
				t.Errorf("RenderP: want %v got %v", tc.wantRenderP, band.RenderP)
			}
		})
	}
}

func TestComplianceTileScore0RendersAsCriticalNotNoData(t *testing.T) {
	html := tileHTML(t, newTileArgs(intPtr(0), "Compliance", "compliance-tile"))
	if !strings.Contains(html, "text-danger") {
		t.Errorf("score=0 must render as Critical (text-danger), got:\n%s", html)
	}
	if strings.Contains(html, "—") {
		t.Errorf("score=0 must not render em-dash (not no-data), got:\n%s", html)
	}
}

func TestComplianceTileXSSPayloadEscaped(t *testing.T) {
	html := tileHTML(t, newTileArgs(intPtr(95), "<script>alert(1)</script>", "compliance-tile"))
	if strings.Contains(html, "<script>alert(1)</script>") {
		t.Error("XSS payload must be escaped")
	}
	if strings.Contains(html, "<script>") {
		t.Errorf("generic raw <script> tag must not appear in output:\n%s", html) // security-defaults 3a
	}
	if !strings.Contains(html, "&lt;script&gt;") {
		t.Errorf("expected escaped payload in output:\n%s", html)
	}
}

func TestComplianceTileWhitespaceOnlyLabelDoesNotCrash(t *testing.T) {
	for _, label := range []string{"", "   "} {
		html := tileHTML(t, newTileArgs(intPtr(95), label, "compliance-tile"))
		if !strings.Contains(html, "<section") {
			t.Errorf("label=%q: expected valid <section> in output:\n%s", label, html)
		}
	}
}

func TestComplianceTileTargetShapeAttributes(t *testing.T) {
	html := tileHTML(t, newTileArgs(intPtr(95), "Compliance", "compliance-tile"))
	for _, want := range []string{
		`id="compliance-tile"`,
		`role="status"`,
		`aria-live="polite"`,
		`aria-atomic="true"`,
		`class="compliance-tile"`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("expected %q in output:\n%s", want, html)
		}
	}
}

func TestComplianceTileNoHtmxProducerAttributes(t *testing.T) {
	html := tileHTML(t, newTileArgs(intPtr(95), "Compliance", "compliance-tile"))
	for _, forbidden := range []string{"hx-get", "hx-post", "hx-target", "hx-swap", "hx-trigger", "<script", "onload=", "data-htmx-"} {
		if strings.Contains(html, forbidden) {
			t.Errorf("found forbidden token %q in output:\n%s", forbidden, html)
		}
	}
}
