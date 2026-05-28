package components_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/code-chimp/fieldmark-go/internal/web/viewmodels"
)

var auditRowFixture = viewmodels.AuditRowVM{
	Action:      "ProjectCreated",
	ActionClass: "badge-audit-action",
	ActorName:   "Aisha Stone",
	OccurredAt:  "2026-05-28T14:20:01Z",
	Absolute:    "2026-05-28 14:20:01 UTC",
	Relative:    "3 minutes ago",
}

func TestAuditRowVariantsMatchCanonical(t *testing.T) {
	cases := map[string]viewmodels.AuditRowVM{
		"default": auditRowFixture,
		"with-disclosure-collapsed": func() viewmodels.AuditRowVM {
			vm := auditRowFixture
			vm.BeforeAfterJSON = `{"after":{"status":"ACTIVE"},"before":{"status":"DRAFT"}}`
			return vm
		}(),
		"with-disclosure-expanded": func() viewmodels.AuditRowVM {
			vm := auditRowFixture
			vm.BeforeAfterJSON = `{"after":{"status":"ACTIVE"},"before":{"status":"DRAFT"}}`
			vm.Expanded = true
			return vm
		}(),
		"unknown-action": func() viewmodels.AuditRowVM {
			vm := auditRowFixture
			vm.Action = "UnknownAction"
			vm.ActionClass = "badge-unknown"
			return vm
		}(),
		"empty-actor": func() viewmodels.AuditRowVM {
			vm := auditRowFixture
			vm.ActorName = ""
			return vm
		}(),
	}
	for variant, vm := range cases {
		t.Run(variant, func(t *testing.T) {
			assertComponentSnapshot(t, "audit_row", "audit_row", variant, vm)
		})
	}
}

func TestAuditRowEscapesJSONText(t *testing.T) {
	vm := auditRowFixture
	vm.BeforeAfterJSON = `<script>alert(1)</script>`
	vm.Expanded = true
	html := renderComponent(t, "audit_row", vm)
	if !strings.Contains(html, `&lt;script&gt;alert(1)&lt;/script&gt;`) {
		t.Fatalf("expected escaped JSON payload, got %q", html)
	}
	if strings.Contains(html, `<script>alert(1)</script>`) {
		t.Fatalf("raw script payload must not appear in rendered audit row: %q", html)
	}
}

func TestAuditRowWhitespaceOnlyActorMatchesEmptyActorCanonical(t *testing.T) {
	vm := auditRowFixture
	vm.ActorName = "   "
	assertComponentSnapshot(t, "audit_row", "audit_row", "empty-actor", vm)
}

func TestAuditRowUnknownActionFallbackClass(t *testing.T) {
	vm := auditRowFixture
	vm.Action = "UnknownAction"
	vm.ActionClass = "badge-unknown"
	html := renderComponent(t, "audit_row", vm)
	if !strings.Contains(html, "badge-unknown") {
		t.Fatalf("unknown audit action must include fallback class, got %q", html)
	}
}

func TestAuditRowEmptyActorFallbackIsTemplateOwned(t *testing.T) {
	vm := auditRowFixture
	vm.ActorName = ""
	html := renderComponent(t, "audit_row", vm)
	if !strings.Contains(html, `<span class="audit-row__actor">unnamed</span>`) {
		t.Fatalf("empty actor must render unnamed fallback, got %q", html)
	}
	if !strings.Contains(html, `<span class="audit-row__initials">??</span>`) {
		t.Fatalf("empty actor must render initials fallback, got %q", html)
	}
}

func TestAuditRowTemplateDoesNotUseTemplateHTML(t *testing.T) {
	_, thisFile, _, _ := runtime.Caller(0)
	b, err := os.ReadFile(filepath.Join(filepath.Dir(thisFile), "audit_row.html"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "template.HTML(") {
		t.Fatal("audit row template must not use template.HTML")
	}
}
