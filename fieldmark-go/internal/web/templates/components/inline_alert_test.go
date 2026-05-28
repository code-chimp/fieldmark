package components_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/code-chimp/fieldmark-go/internal/web/viewmodels"
)

var inlineAlertFixture = viewmodels.InlineAlertVM{
	Title:   "Action blocked",
	Message: "Resolve open violations before closing.",
	Meta:    "Project PM-104",
}

func TestInlineAlertVariantsMatchCanonical(t *testing.T) {
	cases := map[string]viewmodels.InlineAlertVM{
		"danger":  {AlertClass: "alert-danger", Role: "alert", Icon: "warning"},
		"warning": {AlertClass: "alert-warning", Role: "alert", Icon: "warning"},
		"info":    {AlertClass: "alert-info", Role: "status", Icon: "info"},
		"success": {AlertClass: "alert-success", Role: "status", Icon: "success"},
		"unknown": {AlertClass: "alert-unknown", Role: "status", Icon: "info"},
	}
	for variant, vm := range cases {
		vm.Title = inlineAlertFixture.Title
		vm.Message = inlineAlertFixture.Message
		vm.Meta = inlineAlertFixture.Meta
		t.Run(variant, func(t *testing.T) {
			assertComponentSnapshot(t, "inline_alert", "inline_alert", variant, vm)
		})
	}
}

func TestInlineAlertEscapesUserStrings(t *testing.T) {
	html := renderComponent(t, "inline_alert", viewmodels.InlineAlertVM{
		AlertClass: "alert-danger",
		Role:       "alert",
		Icon:       "warning",
		Title:      "<script>alert(1)</script>",
		Message:    "<script>alert(1)</script>",
		Meta:       "<script>alert(1)</script>",
	})
	if !strings.Contains(html, "&lt;script&gt;alert(1)&lt;/script&gt;") {
		t.Fatalf("expected escaped script payload, got %q", html)
	}
	if strings.Contains(html, "<script>alert(1)</script>") {
		t.Fatalf("raw script payload must not appear in rendered inline alert: %q", html)
	}
}

func TestInlineAlertUnknownFallbackClass(t *testing.T) {
	html := renderComponent(t, "inline_alert", viewmodels.InlineAlertVM{
		AlertClass: "alert-unknown",
		Role:       "status",
		Icon:       "info",
		Title:      inlineAlertFixture.Title,
		Message:    inlineAlertFixture.Message,
		Meta:       inlineAlertFixture.Meta,
	})
	if !strings.Contains(html, "alert-unknown") {
		t.Fatalf("unknown inline alert must include fallback class, got %q", html)
	}
}

func TestInlineAlertTemplateDoesNotUseTemplateHTML(t *testing.T) {
	_, thisFile, _, _ := runtime.Caller(0)
	b, err := os.ReadFile(filepath.Join(filepath.Dir(thisFile), "inline_alert.html"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "template.HTML(") {
		t.Fatal("inline alert template must not use template.HTML")
	}
}
