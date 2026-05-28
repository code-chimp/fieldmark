package components_test

import (
	"bytes"
	"html/template"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/code-chimp/fieldmark-go/internal/web/testutil"
)

func componentCanonical(t *testing.T, component string) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	repoRoot := filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "..", "..")
	return filepath.Join(repoRoot, "fieldmark_shared", "components", component, "canonical.html")
}

func readComponentCanonical(t *testing.T, component string) string {
	t.Helper()
	b, err := os.ReadFile(componentCanonical(t, component))
	if err != nil {
		t.Fatalf("read canonical: %v", err)
	}
	return string(b)
}

func renderComponent(t *testing.T, name string, data any) string {
	t.Helper()
	_, thisFile, _, _ := runtime.Caller(0)
	src, err := os.ReadFile(filepath.Join(filepath.Dir(thisFile), name+".html"))
	if err != nil {
		t.Fatalf("read template: %v", err)
	}
	tmpl, err := template.New(name).Parse(string(src))
	if err != nil {
		t.Fatalf("parse template: %v", err)
	}
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		t.Fatalf("execute template: %v", err)
	}
	return buf.String()
}

func assertComponentSnapshot(t *testing.T, component, templateName, variant string, data any) {
	t.Helper()
	actual := testutil.NormaliseComponent(renderComponent(t, templateName, data))
	canonical := testutil.ExtractVariant(readComponentCanonical(t, component), variant)
	if actual != canonical {
		t.Fatalf("%s %s mismatch:\nwant: %q\ngot:  %q", component, variant, canonical, actual)
	}
}
