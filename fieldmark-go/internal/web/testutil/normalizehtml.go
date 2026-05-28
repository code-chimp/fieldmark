// Package testutil provides shared helpers for web-layer snapshot tests.
package testutil

import (
	"regexp"
	"strings"
)

var (
	commentRe    = regexp.MustCompile(`(?s)<!--.*?-->`)
	whitespaceRe = regexp.MustCompile(`\s+`)
	// hiddenInputRe strips <input type="hidden" ...> elements added by CSRF frameworks
	// so parity snapshot comparisons are not broken by per-request tokens.
	hiddenInputRe = regexp.MustCompile(`(?i)<input[^>]+type="hidden"[^>]*>`)
)

// NormaliseComponent strips HTML comments, collapses whitespace runs to a
// single space, and trims the result. Used for cross-stack snapshot comparison.
func NormaliseComponent(html string) string {
	html = commentRe.ReplaceAllString(html, "")
	html = strings.ReplaceAll(html, "&#34;", "&quot;")
	html = whitespaceRe.ReplaceAllString(html, " ")
	return strings.TrimSpace(html)
}

// NormaliseForParity calls NormaliseComponent and additionally strips
// <input type="hidden"> elements so CSRF tokens do not break cross-stack diffs.
func NormaliseForParity(html string) string {
	html = hiddenInputRe.ReplaceAllString(html, "")
	return NormaliseComponent(html)
}

// ExtractVariant extracts the content of a named variant block from a
// canonical component example file and returns it normalised.
// Blocks are delimited by <!-- variant: <name> ... --> comment lines.
func ExtractVariant(exampleContent, variantName string) string {
	startMarker := "<!-- variant: " + variantName
	lines := strings.Split(exampleContent, "\n")
	inBlock := false
	var collected []string

	for _, line := range lines {
		trimmed := strings.TrimRight(line, " \t\r")
		if trimmed == startMarker+" -->" || strings.HasPrefix(trimmed, startMarker+" ") {
			inBlock = true
			continue
		}
		if inBlock && strings.HasPrefix(trimmed, "<!-- variant:") {
			break
		}
		if inBlock {
			collected = append(collected, trimmed)
		}
	}

	return NormaliseComponent(strings.Join(collected, "\n"))
}
