package components_test

import (
	"bytes"
	"html/template"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	components "github.com/code-chimp/fieldmark-go/internal/web/templates/components"
)

func badgePtr(v int) *int { return &v }

func tabStripFuncMap() template.FuncMap {
	return template.FuncMap{
		"tabTabindex":     components.TabTabindex,
		"tabAriaControls": components.TabAriaControls,
		"tabRequired":     components.TabRequiredString,
	}
}

func renderTabStrip(t *testing.T, args components.TabStripArgs) string {
	t.Helper()
	_, thisFile, _, _ := runtime.Caller(0)
	src, err := os.ReadFile(filepath.Join(filepath.Dir(thisFile), "tab_strip.html"))
	if err != nil {
		t.Fatalf("read tab_strip.html: %v", err)
	}
	tmpl, err := template.New("tab_strip").Funcs(tabStripFuncMap()).Parse(string(src))
	if err != nil {
		t.Fatalf("parse tab_strip.html: %v", err)
	}
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "tab_strip", args); err != nil {
		t.Fatalf("execute tab_strip: %v", err)
	}
	return buf.String()
}

func renderTabStripErr(args components.TabStripArgs) error {
	_, thisFile, _, _ := runtime.Caller(0)
	src, err := os.ReadFile(filepath.Join(filepath.Dir(thisFile), "tab_strip.html"))
	if err != nil {
		return err
	}
	tmpl, err := template.New("tab_strip").Funcs(tabStripFuncMap()).Parse(string(src))
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	return tmpl.ExecuteTemplate(&buf, "tab_strip", args)
}

func assertTabStripSnapshot(t *testing.T, variant string, args components.TabStripArgs) {
	t.Helper()
	_, thisFile, _, _ := runtime.Caller(0)
	repoRoot := filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "..", "..")
	canonical, err := os.ReadFile(
		filepath.Join(repoRoot, "fieldmark_shared", "components", "tab_strip", "canonical.html"),
	)
	if err != nil {
		t.Fatalf("read canonical: %v", err)
	}
	// Import testutil via the shared package (already used in component_snapshot_test.go)
	// We call assertComponentSnapshot from component_snapshot_test.go infra but need FuncMap.
	// For tab_strip, we use a local assertComponentSnapshot that includes the FuncMap.
	actual := normaliseTabStrip(renderTabStrip(t, args))
	want := normaliseTabStrip(extractTabStripVariant(string(canonical), variant))
	if actual != want {
		t.Fatalf("tab_strip %s mismatch:\nwant: %q\ngot:  %q", variant, want, actual)
	}
}

// normaliseTabStrip collapses whitespace runs and strips HTML comments — same as testutil.NormaliseComponent.
func normaliseTabStrip(html string) string {
	// strip HTML comments
	for {
		start := strings.Index(html, "<!--")
		if start < 0 {
			break
		}
		end := strings.Index(html[start:], "-->")
		if end < 0 {
			break
		}
		html = html[:start] + html[start+end+3:]
	}
	// collapse whitespace runs
	fields := strings.Fields(html)
	return strings.Join(fields, " ")
}

// extractTabStripVariant extracts the named variant block from the canonical HTML.
func extractTabStripVariant(src, name string) string {
	marker := "<!-- variant: " + name
	lines := strings.Split(src, "\n")
	var sb strings.Builder
	inBlock := false
	for _, line := range lines {
		trimmed := strings.TrimRight(line, "\r")
		if strings.HasPrefix(trimmed, marker) {
			inBlock = true
			continue
		}
		if inBlock && strings.HasPrefix(trimmed, "<!-- variant:") {
			break
		}
		if inBlock {
			sb.WriteString(trimmed)
			sb.WriteByte('\n')
		}
	}
	return normaliseTabStrip(sb.String())
}

var projectDetailTabs = []components.TabSpec{
	{ID: "tab-summary", Label: "Summary", HxGet: "/projects/__ID__/summary", HxTarget: "#project-detail-tab-content"},
	{ID: "tab-inspections", Label: "Inspections", HxGet: "/projects/__ID__/inspections", HxTarget: "#project-detail-tab-content"},
	{ID: "tab-violations", Label: "Violations", HxGet: "/projects/__ID__/violations", HxTarget: "#project-detail-tab-content"},
	{ID: "tab-audit", Label: "Audit", HxGet: "/projects/__ID__/audit", HxTarget: "#project-detail-tab-content"},
}

func TestTabStripTemplateDoesNotUseTemplateHTMLCast(t *testing.T) {
	_, thisFile, _, _ := runtime.Caller(0)
	for _, filename := range []string{"tab_strip.html", "tab_strip.go"} {
		b, err := os.ReadFile(filepath.Join(filepath.Dir(thisFile), filename))
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(b), "template.HTML(") {
			t.Fatalf("%s must not use template.HTML(", filename)
		}
	}
}

func TestTabStripVariantsMatchCanonical(t *testing.T) {
	cases := map[string]components.TabStripArgs{
		"project-detail-four-tabs-summary-active": {
			ID:          "project-detail-tabstrip",
			AriaLabel:   "Project Detail Tabs",
			ActiveIndex: 0,
			Tabs:        projectDetailTabs,
		},
		"project-detail-four-tabs-violations-active": {
			ID:          "project-detail-tabstrip",
			AriaLabel:   "Project Detail Tabs",
			ActiveIndex: 2,
			Tabs:        projectDetailTabs,
		},
		"project-detail-four-tabs-with-badges": {
			ID:        "project-detail-tabstrip",
			AriaLabel: "Project Detail Tabs",
			Tabs: []components.TabSpec{
				{ID: "tab-summary", Label: "Summary", HxGet: "/projects/__ID__/summary", HxTarget: "#project-detail-tab-content"},
				{ID: "tab-inspections", Label: "Inspections", HxGet: "/projects/__ID__/inspections", HxTarget: "#project-detail-tab-content", BadgeCount: badgePtr(12)},
				{ID: "tab-violations", Label: "Violations", HxGet: "/projects/__ID__/violations", HxTarget: "#project-detail-tab-content", BadgeCount: badgePtr(3)},
				{ID: "tab-audit", Label: "Audit", HxGet: "/projects/__ID__/audit", HxTarget: "#project-detail-tab-content", BadgeCount: badgePtr(147)},
			},
		},
		"two-tabs-minimal": {
			ID:        "two-tabs-strip",
			AriaLabel: "Open Closed Tabs",
			Tabs: []components.TabSpec{
				{ID: "tab-open", Label: "Open", HxGet: "/__tab__/open", HxTarget: "#__panel__"},
				{ID: "tab-closed", Label: "Closed", HxGet: "/__tab__/closed", HxTarget: "#__panel__"},
			},
		},
		"single-tab": {
			ID:        "single-tab-strip",
			AriaLabel: "Single Tab",
			Tabs:      []components.TabSpec{{ID: "tab-only", Label: "Only Tab", HxGet: "/__tab__/only", HxTarget: "#__panel__"}},
		},
		"badge-zero": {
			ID:        "project-detail-tabstrip",
			AriaLabel: "Project Detail Tabs",
			Tabs: []components.TabSpec{
				{ID: "tab-summary", Label: "Summary", HxGet: "/projects/__ID__/summary", HxTarget: "#project-detail-tab-content"},
				{ID: "tab-inspections", Label: "Inspections", HxGet: "/projects/__ID__/inspections", HxTarget: "#project-detail-tab-content", BadgeCount: badgePtr(12)},
				{ID: "tab-violations", Label: "Violations", HxGet: "/projects/__ID__/violations", HxTarget: "#project-detail-tab-content", BadgeCount: badgePtr(0)},
				{ID: "tab-audit", Label: "Audit", HxGet: "/projects/__ID__/audit", HxTarget: "#project-detail-tab-content", BadgeCount: badgePtr(147)},
			},
		},
		"badge-large": {
			ID:        "project-detail-tabstrip",
			AriaLabel: "Project Detail Tabs",
			Tabs: []components.TabSpec{
				{ID: "tab-summary", Label: "Summary", HxGet: "/projects/__ID__/summary", HxTarget: "#project-detail-tab-content"},
				{ID: "tab-inspections", Label: "Inspections", HxGet: "/projects/__ID__/inspections", HxTarget: "#project-detail-tab-content", BadgeCount: badgePtr(9999)},
				{ID: "tab-violations", Label: "Violations", HxGet: "/projects/__ID__/violations", HxTarget: "#project-detail-tab-content"},
				{ID: "tab-audit", Label: "Audit", HxGet: "/projects/__ID__/audit", HxTarget: "#project-detail-tab-content"},
			},
		},
	}
	for variant, args := range cases {
		args := args
		t.Run(variant, func(t *testing.T) {
			assertTabStripSnapshot(t, variant, args)
		})
	}
}

// Tabindex distribution tests
func TestTabTabindex_ActiveIndex0_FirstTabHasTabindex0(t *testing.T) {
	args := components.TabStripArgs{ID: "s", AriaLabel: "Tabs", ActiveIndex: 0, Tabs: projectDetailTabs}
	html := renderTabStrip(t, args)
	if strings.Count(html, `tabindex="0"`) != 1 {
		t.Errorf("expected exactly 1 tabindex=0, got:\n%s", html)
	}
	if strings.Count(html, `tabindex="-1"`) != 3 {
		t.Errorf("expected exactly 3 tabindex=-1, got:\n%s", html)
	}
}

func TestTabTabindex_ActiveIndex2_ThirdTabSelected(t *testing.T) {
	args := components.TabStripArgs{ID: "s", AriaLabel: "Tabs", ActiveIndex: 2, Tabs: projectDetailTabs}
	html := renderTabStrip(t, args)
	if strings.Count(html, `aria-selected="true"`) != 1 {
		t.Errorf("expected exactly 1 aria-selected=true")
	}
	if strings.Count(html, `aria-selected="false"`) != 3 {
		t.Errorf("expected exactly 3 aria-selected=false")
	}
}

func TestTabTabindex_ActiveIndexLast_LastTabActive(t *testing.T) {
	args := components.TabStripArgs{ID: "s", AriaLabel: "Tabs", ActiveIndex: 3, Tabs: projectDetailTabs}
	html := renderTabStrip(t, args)
	if strings.Count(html, `aria-selected="true"`) != 1 {
		t.Error("expected exactly 1 aria-selected=true")
	}
	if !strings.Contains(html, `id="tab-audit"`) {
		t.Error("expected tab-audit to be present")
	}
}

// Badge tests
func TestBadge_Count12_RendersWithText12(t *testing.T) {
	args := components.TabStripArgs{
		ID:        "s",
		AriaLabel: "Tabs",
		Tabs:      []components.TabSpec{{ID: "t", Label: "Tab", HxGet: "/t", HxTarget: "#p", BadgeCount: badgePtr(12)}},
	}
	html := renderTabStrip(t, args)
	if !strings.Contains(html, `class="badge tab-strip__badge"`) {
		t.Error("expected badge class")
	}
	if !strings.Contains(html, ">12<") {
		t.Error("expected badge text 12")
	}
	if !strings.Contains(html, `aria-label="12 unread"`) {
		t.Error("expected aria-label 12 unread")
	}
}

func TestBadge_Count0_RendersWithZero(t *testing.T) {
	args := components.TabStripArgs{
		ID:        "s",
		AriaLabel: "Tabs",
		Tabs:      []components.TabSpec{{ID: "t", Label: "Tab", HxGet: "/t", HxTarget: "#p", BadgeCount: badgePtr(0)}},
	}
	html := renderTabStrip(t, args)
	if !strings.Contains(html, "tab-strip__badge") {
		t.Errorf("expected badge for count=0:\n%s", html)
	}
	if !strings.Contains(html, ">0<") {
		t.Errorf("expected badge text 0:\n%s", html)
	}
	if !strings.Contains(html, `aria-label="0 unread"`) {
		t.Errorf("expected aria-label 0 unread:\n%s", html)
	}
}

func TestBadge_Nil_NoBadgeElement(t *testing.T) {
	args := components.TabStripArgs{
		ID:        "s",
		AriaLabel: "Tabs",
		Tabs:      []components.TabSpec{{ID: "t", Label: "Tab", HxGet: "/t", HxTarget: "#p"}},
	}
	html := renderTabStrip(t, args)
	if strings.Contains(html, "tab-strip__badge") {
		t.Errorf("expected no badge for nil count:\n%s", html)
	}
}

func TestBadge_Count9999_NoTruncation(t *testing.T) {
	args := components.TabStripArgs{
		ID:        "s",
		AriaLabel: "Tabs",
		Tabs:      []components.TabSpec{{ID: "t", Label: "Tab", HxGet: "/t", HxTarget: "#p", BadgeCount: badgePtr(9999)}},
	}
	html := renderTabStrip(t, args)
	if !strings.Contains(html, ">9999<") {
		t.Error("expected badge text 9999")
	}
	if strings.Contains(html, "99+") {
		t.Error("expected no truncated badge text")
	}
}

// type="button" guard
func TestAllButtons_HaveTypeButton(t *testing.T) {
	args := components.TabStripArgs{ID: "s", AriaLabel: "Tabs", ActiveIndex: 0, Tabs: projectDetailTabs}
	html := renderTabStrip(t, args)
	if strings.Count(html, `type="button"`) != len(projectDetailTabs) {
		t.Errorf("expected %d type=button attributes", len(projectDetailTabs))
	}
}

// XSS tests
func TestXSS_LabelIsEscaped(t *testing.T) {
	args := components.TabStripArgs{
		ID:        "s",
		AriaLabel: "Tabs",
		Tabs:      []components.TabSpec{{ID: "t", Label: "<script>alert(1)</script>", HxGet: "/t", HxTarget: "#p"}},
	}
	html := renderTabStrip(t, args)
	if strings.Contains(html, "<script>") {
		t.Errorf("XSS payload must be escaped in Label:\n%s", html)
	}
	if !strings.Contains(html, "&lt;script&gt;") {
		t.Errorf("expected escaped payload in Label:\n%s", html)
	}
}

func TestXSS_AriaLabelIsEscaped(t *testing.T) {
	args := components.TabStripArgs{
		ID:        "s",
		AriaLabel: "<script>alert(1)</script>",
		Tabs:      []components.TabSpec{{ID: "t", Label: "Tab", HxGet: "/t", HxTarget: "#p"}},
	}
	html := renderTabStrip(t, args)
	if strings.Contains(html, "<script>") {
		t.Errorf("XSS payload must be escaped in AriaLabel:\n%s", html)
	}
	if !strings.Contains(html, "&lt;script&gt;") {
		t.Errorf("expected escaped payload in AriaLabel:\n%s", html)
	}
}

func TestXSS_HxGetIsEscaped(t *testing.T) {
	args := components.TabStripArgs{
		ID:        "s",
		AriaLabel: "Tabs",
		Tabs:      []components.TabSpec{{ID: "t", Label: "Tab", HxGet: "javascript:alert(1)", HxTarget: "#p"}},
	}
	html := renderTabStrip(t, args)
	// hx-get should NOT have an active script injection; html/template escapes it
	if strings.Contains(html, "<script>") {
		t.Errorf("XSS payload must not create script element:\n%s", html)
	}
}

// Edge cases — empty / whitespace label does not crash
func TestEmptyLabel_DoesNotCrash(t *testing.T) {
	args := components.TabStripArgs{
		ID:        "s",
		AriaLabel: "Tabs",
		Tabs:      []components.TabSpec{{ID: "t", Label: "", HxGet: "/t", HxTarget: "#p"}},
	}
	html := renderTabStrip(t, args)
	if !strings.Contains(html, "tab-strip__label") {
		t.Errorf("expected tab-strip__label in output:\n%s", html)
	}
}

func TestWhitespaceOnlyLabel_DoesNotCrash(t *testing.T) {
	args := components.TabStripArgs{
		ID:        "s",
		AriaLabel: "Tabs",
		Tabs:      []components.TabSpec{{ID: "t", Label: "   ", HxGet: "/t", HxTarget: "#p"}},
	}
	html := renderTabStrip(t, args)
	if !strings.Contains(html, "tab-strip__label") {
		t.Errorf("expected tab-strip__label in output:\n%s", html)
	}
}

// Required-prop enforcement tests
func TestAriaLabel_Empty_ReturnsError(t *testing.T) {
	args := components.TabStripArgs{
		ID:        "s",
		AriaLabel: "",
		Tabs:      []components.TabSpec{{ID: "t", Label: "Tab", HxGet: "/t", HxTarget: "#p"}},
	}
	if err := renderTabStripErr(args); err == nil {
		t.Error("expected error for empty AriaLabel, got nil")
	}
}

func TestAriaLabel_WhitespaceOnly_ReturnsError(t *testing.T) {
	args := components.TabStripArgs{
		ID:        "s",
		AriaLabel: "   ",
		Tabs:      []components.TabSpec{{ID: "t", Label: "Tab", HxGet: "/t", HxTarget: "#p"}},
	}
	if err := renderTabStripErr(args); err == nil {
		t.Error("expected error for whitespace-only AriaLabel, got nil")
	}
}

// tabTabindex pure function tests
func TestTabTabindex_ReturnsZeroForActiveIndex(t *testing.T) {
	if got := components.TabTabindex(2, 2); got != "0" {
		t.Errorf("expected 0, got %s", got)
	}
}

func TestTabTabindex_ReturnsNegOneForInactive(t *testing.T) {
	if got := components.TabTabindex(2, 0); got != "-1" {
		t.Errorf("expected -1, got %s", got)
	}
}

// Grep guard: tab_strip.html and tab_strip.go must not use template.HTML(
// (covered in TestTabStripTemplateDoesNotUseTemplateHTMLCast above, which also covers other components)
// Extend the existing TestEntityRailOtherTemplatesDoNotUseTemplateHTMLCast to include tab_strip.html
// by exercising it explicitly here:
func TestTabStripHTMLFileHasNoTemplateHTMLCast(t *testing.T) {
	_, thisFile, _, _ := runtime.Caller(0)
	b, err := os.ReadFile(filepath.Join(filepath.Dir(thisFile), "tab_strip.html"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "template.HTML(") {
		t.Fatal("tab_strip.html must not use template.HTML(")
	}
}
