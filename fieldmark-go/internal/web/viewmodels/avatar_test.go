package viewmodels

import "testing"

func TestInitialsEmptyFullNameFallsBackToUsername(t *testing.T) {
	if got := Initials("", "alice"); got != "AL" {
		t.Errorf("got %q, want %q", got, "AL")
	}
}

func TestInitialsNilEquivalentFullNameFallsBackToUsername(t *testing.T) {
	if got := Initials("   ", "bob"); got != "BO" {
		t.Errorf("got %q, want %q", got, "BO")
	}
}

func TestInitialsSingleTokenFirstTwoCharsUppercased(t *testing.T) {
	if got := Initials("Alice", "alice"); got != "AL" {
		t.Errorf("got %q, want %q", got, "AL")
	}
}

func TestInitialsTwoTokenFirstAndLastInitials(t *testing.T) {
	if got := Initials("Alice Admin", "alice"); got != "AA" {
		t.Errorf("got %q, want %q", got, "AA")
	}
}

func TestInitialsThreePlusTokenFirstAndLastInitials(t *testing.T) {
	if got := Initials("Alice Marie Admin", "alice"); got != "AA" {
		t.Errorf("got %q, want %q", got, "AA")
	}
}

func TestInitialsUnicodePreservedAsIs(t *testing.T) {
	if got := Initials("Ää Öö", "aao"); got != "ÄÖ" {
		t.Errorf("got %q, want %q", got, "ÄÖ")
	}
}

func TestInitialsSingleTokenUnicodeFirstTwoChars(t *testing.T) {
	if got := Initials("李明", "user"); got != "李明" {
		t.Errorf("got %q, want %q", got, "李明")
	}
}

func TestInitialsBothEmptyReturnsFallbackToken(t *testing.T) {
	if got := Initials("", ""); got != "??" {
		t.Errorf("got %q, want %q", got, "??")
	}
}

func TestInitialsWhitespaceOnlyBothReturnsFallbackToken(t *testing.T) {
	if got := Initials("   ", "   "); got != "??" {
		t.Errorf("got %q, want %q", got, "??")
	}
}
