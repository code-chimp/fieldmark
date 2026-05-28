// Story 2.2 AC6 — Go conformance gate. AllAuditActions must match the
// canonical fixture at docs/reference/audit-actions.json exactly. Pure unit
// test, no DB, no build tag — runs under `go test ./...`.
package enums

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"testing"
)

func TestAuditActionMatchesCanonicalFixture(t *testing.T) {
	fixturePath := locateFixture(t)

	raw, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("read fixture %s: %v", fixturePath, err)
	}

	var doc struct {
		Actions []string `json:"actions"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("parse fixture: %v", err)
	}

	// Cardinality first — set equality alone masks duplicate fixture entries.
	if dupes := duplicates(doc.Actions); len(dupes) > 0 {
		t.Errorf("audit-actions.json contains duplicate entries: %v", dupes)
	}

	canonical := toSet(doc.Actions)
	native := make(map[string]struct{}, len(AllAuditActions))
	for _, a := range AllAuditActions {
		native[string(a)] = struct{}{}
	}

	missingFromNative := diff(canonical, native)
	extrasInNative := diff(native, canonical)

	if len(missingFromNative) > 0 {
		t.Errorf("canonical actions missing from Go AllAuditActions: %v", missingFromNative)
	}
	if len(extrasInNative) > 0 {
		t.Errorf("Go AllAuditActions has extras not in fixture: %v", extrasInNative)
	}
}

// TestAllAuditActions_includesEveryDeclaredConst guards the bypass path the
// fixture-comparison test cannot see: a new `AuditAction...` const declared in
// audit_action.go but accidentally omitted from AllAuditActions would let the
// fixture check pass while the slice silently lags. We parse the source file
// with go/parser and assert every declared const of type AuditAction appears
// in AllAuditActions.
func TestAllAuditActions_includesEveryDeclaredConst(t *testing.T) {
	declared := declaredAuditActionConsts(t)
	allSet := make(map[string]struct{}, len(AllAuditActions))
	for _, a := range AllAuditActions {
		allSet[string(a)] = struct{}{}
	}

	missing := make([]string, 0)
	for name, value := range declared {
		if _, ok := allSet[value]; !ok {
			missing = append(missing, name+" = "+value)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Errorf("AuditAction constants declared but not added to AllAuditActions: %v", missing)
	}
}

// declaredAuditActionConsts parses audit_action.go (sibling file) and returns
// a map of declared const symbol name → string literal value, for every const
// whose declared type is AuditAction.
func declaredAuditActionConsts(t *testing.T) map[string]string {
	t.Helper()
	_, thisFile, _, _ := runtime.Caller(0)
	sourcePath := filepath.Join(filepath.Dir(thisFile), "audit_action.go")

	fset := token.NewFileSet()
	parsed, err := parser.ParseFile(fset, sourcePath, nil, parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("parse %s: %v", sourcePath, err)
	}

	out := make(map[string]string)
	for _, decl := range parsed.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}
		// Track the most recent explicit type within the const block so
		// implicit-type ValueSpecs inside `const (...)` inherit correctly.
		var currentType string
		for _, spec := range gen.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			if vs.Type != nil {
				if ident, ok := vs.Type.(*ast.Ident); ok {
					currentType = ident.Name
				}
			}
			if currentType != "AuditAction" {
				continue
			}
			for i, name := range vs.Names {
				if i >= len(vs.Values) {
					continue
				}
				lit, ok := vs.Values[i].(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					continue
				}
				// Unquote the string literal.
				if len(lit.Value) >= 2 {
					out[name.Name] = lit.Value[1 : len(lit.Value)-1]
				}
			}
		}
	}
	if len(out) == 0 {
		t.Fatalf("found no AuditAction constants in %s — parser may be broken", sourcePath)
	}
	return out
}

func duplicates(items []string) []string {
	seen := make(map[string]int, len(items))
	for _, it := range items {
		seen[it]++
	}
	out := make([]string, 0)
	for k, n := range seen {
		if n > 1 {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}

func toSet(items []string) map[string]struct{} {
	out := make(map[string]struct{}, len(items))
	for _, it := range items {
		out[it] = struct{}{}
	}
	return out
}

func diff(a, b map[string]struct{}) []string {
	out := make([]string, 0)
	for k := range a {
		if _, ok := b[k]; !ok {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}

// locateFixture walks up from this test file's directory until it finds
// docs/reference/audit-actions.json at the repo root. runtime.Caller is the
// stable anchor — `go test`'s working directory is the package dir, but the
// walk is the same either way.
func locateFixture(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}
	dir := filepath.Dir(thisFile)
	for {
		candidate := filepath.Join(dir, "docs", "reference", "audit-actions.json")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not locate docs/reference/audit-actions.json walking up from %s",
				filepath.Dir(thisFile))
		}
		dir = parent
	}
}
