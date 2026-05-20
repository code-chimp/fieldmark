package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// sixUserJSON builds a minimal valid six-user manifest, replacing the first
// entry with the provided overrides so individual field tests stay terse.
func sixUserJSON(overrideFirst string) string {
	base := `{"users":[
	  ` + overrideFirst + `,
	  {"id":"01923456-7890-7abc-def0-000000000002","username":"pat","display_name":"Pat Smith","password":"x","role":"SITE_SUPERVISOR"},
	  {"id":"01923456-7890-7abc-def0-000000000003","username":"aisha","display_name":"Aisha Patel","password":"x","role":"ADMIN"},
	  {"id":"01923456-7890-7abc-def0-000000000004","username":"ravi","display_name":"Ravi Kumar","password":"x","role":"INSPECTOR"},
	  {"id":"01923456-7890-7abc-def0-000000000005","username":"kenji","display_name":"Kenji Tanaka","password":"x","role":"EXECUTIVE"},
	  {"id":"01923456-7890-7abc-def0-000000000006","username":"testuser","display_name":"Test User","password":"x","role":null}
	]}`
	return base
}

func writeManifest(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "dev-users.json")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestParseManifest_RoundTrip(t *testing.T) {
	first := `{"id":"01923456-7890-7abc-def0-123456789abc","username":"marisol","display_name":"Marisol Vega","password":"x","role":"COMPLIANCE_OFFICER"}`
	path := writeManifest(t, sixUserJSON(first))
	m, err := parseManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Users) != 6 {
		t.Fatalf("want 6 users, got %d", len(m.Users))
	}
	if m.Users[5].Role != nil {
		t.Fatalf("testuser role should be nil, got %v", *m.Users[5].Role)
	}
}

func TestParseManifest_RejectsWrongCount(t *testing.T) {
	path := writeManifest(t, `{"users":[]}`)
	if _, err := parseManifest(path); err == nil {
		t.Fatal("want error for zero users, got nil")
	}
	// 5-user manifest also rejected
	short := `{"users":[
	  {"id":"01923456-7890-7abc-def0-000000000001","username":"a","display_name":"A","password":"x","role":null},
	  {"id":"01923456-7890-7abc-def0-000000000002","username":"b","display_name":"B","password":"x","role":null},
	  {"id":"01923456-7890-7abc-def0-000000000003","username":"c","display_name":"C","password":"x","role":null},
	  {"id":"01923456-7890-7abc-def0-000000000004","username":"d","display_name":"D","password":"x","role":null},
	  {"id":"01923456-7890-7abc-def0-000000000005","username":"e","display_name":"E","password":"x","role":null}
	]}`
	if _, err := parseManifest(writeManifest(t, short)); err == nil {
		t.Fatal("want error for 5 users, got nil")
	}
}

func TestParseManifest_RejectsEmptyFields(t *testing.T) {
	cases := []struct {
		name  string
		entry string
	}{
		{"empty display_name", `{"id":"01923456-7890-7abc-def0-123456789abc","username":"marisol","display_name":"","password":"x","role":"COMPLIANCE_OFFICER"}`},
		{"empty username", `{"id":"01923456-7890-7abc-def0-123456789abc","username":"","display_name":"Marisol Vega","password":"x","role":"COMPLIANCE_OFFICER"}`},
		{"empty password", `{"id":"01923456-7890-7abc-def0-123456789abc","username":"marisol","display_name":"Marisol Vega","password":"","role":"COMPLIANCE_OFFICER"}`},
	}
	for _, c := range cases {
		if _, err := parseManifest(writeManifest(t, sixUserJSON(c.entry))); !strings.Contains(err.Error(), "empty required") {
			t.Fatalf("%s: want 'empty required' error, got %v", c.name, err)
		}
	}
}

func TestParseManifest_RejectsInvalidRole(t *testing.T) {
	bad := `{"id":"01923456-7890-7abc-def0-123456789abc","username":"marisol","display_name":"Marisol Vega","password":"x","role":"HACKER"}`
	if _, err := parseManifest(writeManifest(t, sixUserJSON(bad))); !strings.Contains(err.Error(), "invalid role") {
		t.Fatalf("want 'invalid role' error, got %v", err)
	}
	// null role should still be accepted
	nullRole := `{"id":"01923456-7890-7abc-def0-123456789abc","username":"marisol","display_name":"Marisol Vega","password":"x","role":null}`
	if _, err := parseManifest(writeManifest(t, sixUserJSON(nullRole))); err != nil {
		t.Fatalf("null role should be accepted, got %v", err)
	}
}
