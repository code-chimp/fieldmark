package viewmodels

import "strings"

// Initials derives display initials from a full name or username fallback (AC #3, Story 1.13).
//
// Algorithm:
//   - Two+ whitespace-separated tokens → first char of first + first char of last, uppercased.
//   - Single-token full name → first two characters, uppercased.
//   - Empty full name → first two characters of usernameFallback, uppercased.
//   - Unicode characters are preserved as-is (no transliteration).
func Initials(fullName, usernameFallback string) string {
	name := strings.TrimSpace(fullName)
	if name != "" {
		tokens := strings.Fields(name)
		if len(tokens) >= 2 {
			return strings.ToUpper(string([]rune(tokens[0])[:1]) + string([]rune(tokens[len(tokens)-1])[:1]))
		}
		runes := []rune(tokens[0])
		if len(runes) >= 2 {
			return strings.ToUpper(string(runes[:2]))
		}
		return strings.ToUpper(string(runes))
	}

	fallback := strings.TrimSpace(usernameFallback)
	runes := []rune(fallback)
	if len(runes) >= 2 {
		return strings.ToUpper(string(runes[:2]))
	}
	if len(runes) > 0 {
		return strings.ToUpper(string(runes))
	}
	return "??"
}
