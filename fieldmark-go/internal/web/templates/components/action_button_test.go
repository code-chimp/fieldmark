package components_test

import (
	"bytes"
	"html/template"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"golang.org/x/net/html"

	"github.com/code-chimp/fieldmark-go/internal/web/testutil"
	"github.com/code-chimp/fieldmark-go/internal/web/viewmodels"
)

// canonicalExamplePath resolves the canonical action_button.example.html from
// the fieldmark_shared/components/ directory relative to this test file.
func canonicalExamplePath(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	// thisFile: .../fieldmark-go/internal/web/templates/components/action_button_test.go
	// 5 levels up: components→templates→web→internal→fieldmark-go→repo-root
	repoRoot := filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "..", "..")
	return filepath.Join(repoRoot, "fieldmark_shared", "components", "action_button.example.html")
}

func readCanonical(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile(canonicalExamplePath(t))
	if err != nil {
		t.Fatalf("readCanonical: %v", err)
	}
	return string(b)
}

// renderActionButton renders the action_button template with the given VM.
func renderActionButton(t *testing.T, vm viewmodels.ActionButtonVM) string {
	t.Helper()
	_, thisFile, _, _ := runtime.Caller(0)
	tmplPath := filepath.Join(filepath.Dir(thisFile), "action_button.html")

	src, err := os.ReadFile(tmplPath)
	if err != nil {
		t.Fatalf("read template: %v", err)
	}

	tmpl, err := template.New("action_button").Parse(string(src))
	if err != nil {
		t.Fatalf("parse template: %v", err)
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "action_button", vm); err != nil {
		t.Fatalf("execute template: %v", err)
	}
	return buf.String()
}

// canonical fixture matching action_button.example.html
var fixture = viewmodels.ActionButtonVM{
	ID:             "ab-fixture-1",
	Permission:     true,
	StateAllows:    false,
	Label:          "Approve Resolution",
	HxPost:         "/violations/00000000-0000-0000-0000-000000000001/corrective-actions/00000000-0000-0000-0000-000000000002/approve",
	HxTarget:       "#violation-detail",
	DisabledReason: "Awaiting review",
}

func TestActionButton_PermissionFalse_RendersEmpty(t *testing.T) {
	vm := fixture
	vm.Permission = false
	html := testutil.NormaliseComponent(renderActionButton(t, vm))
	if html != "" {
		t.Fatalf("absent variant: expected empty output, got %q", html)
	}
}

func TestActionButton_DisabledVariant_MatchesCanonicalSnapshot(t *testing.T) {
	vm := fixture // permission=true, state_allows=false
	actual := testutil.NormaliseComponent(renderActionButton(t, vm))
	canonical := testutil.ExtractVariant(readCanonical(t), "disabled")
	if actual != canonical {
		t.Fatalf("disabled variant mismatch:\nwant: %q\ngot:  %q", canonical, actual)
	}
}

func TestActionButton_PresentVariant_MatchesCanonicalSnapshot(t *testing.T) {
	vm := fixture
	vm.StateAllows = true
	actual := testutil.NormaliseComponent(renderActionButton(t, vm))
	canonical := testutil.ExtractVariant(readCanonical(t), "present")
	if actual != canonical {
		t.Fatalf("present variant mismatch:\nwant: %q\ngot:  %q", canonical, actual)
	}
}

func TestActionButton_DisabledVariant_HasScreenReaderReason(t *testing.T) {
	vm := fixture // permission=true, state_allows=false
	rendered := renderActionButton(t, vm)

	doc, err := html.Parse(strings.NewReader(rendered))
	if err != nil {
		t.Fatalf("html.Parse: %v", err)
	}

	var button, srSpan *html.Node
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "button":
				button = n
			case "span":
				if attrVal(n, "id") == "ab-fixture-1-reason" {
					srSpan = n
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)

	if button == nil {
		t.Fatal("disabled variant: no <button> element found")
	}
	if attrVal(button, "aria-disabled") != "true" {
		t.Errorf("button aria-disabled: want %q, got %q", "true", attrVal(button, "aria-disabled"))
	}
	if attrVal(button, "tabindex") != "0" {
		t.Errorf("button tabindex: want %q, got %q", "0", attrVal(button, "tabindex"))
	}
	if attrVal(button, "data-tooltip") != "Awaiting review" {
		t.Errorf("button data-tooltip: want %q, got %q", "Awaiting review", attrVal(button, "data-tooltip"))
	}
	describedBy := attrVal(button, "aria-describedby")
	if describedBy != "ab-fixture-1-reason" {
		t.Errorf("button aria-describedby: want %q, got %q", "ab-fixture-1-reason", describedBy)
	}

	if srSpan == nil {
		t.Fatal("disabled variant: sr-only reason span not found")
	}
	if !strings.Contains(attrVal(srSpan, "class"), "sr-only") {
		t.Errorf("sr-only span class: want to contain %q, got %q", "sr-only", attrVal(srSpan, "class"))
	}
	if text := nodeText(srSpan); text != "Awaiting review" {
		t.Errorf("sr-only span text: want %q, got %q", "Awaiting review", text)
	}
}

func attrVal(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func nodeText(n *html.Node) string {
	var sb strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			sb.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return strings.TrimSpace(sb.String())
}
