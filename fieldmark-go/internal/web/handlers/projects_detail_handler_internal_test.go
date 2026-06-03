package handlers

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/code-chimp/fieldmark-go/internal/data/postgres"
)

func TestRenderAuditJSON_RecursivelySortsNestedObjects(t *testing.T) {
	row := postgres.AuditEntryRow{
		ID:          uuid.MustParse("a2000000-0000-0000-0000-000000000004"),
		OccurredAt:  time.Date(2026, 6, 3, 15, 0, 0, 0, time.UTC),
		BeforeState: []byte(`{"zebra":{"delta":2,"alpha":1}}`),
		AfterState:  []byte(`{"items":[{"zulu":2,"bravo":1}],"alpha":1}`),
		Metadata:    []byte(`{"reason":"<script>alert(1)</script>"}`),
	}

	got := renderAuditJSON(row)
	want := `{"after":{"alpha":1,"items":[{"bravo":1,"zulu":2}]},"before":{"zebra":{"alpha":1,"delta":2}},"metadata":{"reason":"\u003cscript\u003ealert(1)\u003c/script\u003e"}}`
	if got != want {
		t.Fatalf("renderAuditJSON() mismatch:\nwant: %s\ngot:  %s", want, got)
	}
}
