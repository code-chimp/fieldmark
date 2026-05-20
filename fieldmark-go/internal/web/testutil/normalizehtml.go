// Package testutil provides shared helpers for web-layer snapshot tests.
package testutil

import (
	"regexp"
	"strings"
)

var (
	commentRe    = regexp.MustCompile(`(?s)<!--.*?-->`)
	whitespaceRe = regexp.MustCompile(`\s+`)
)

// NormaliseComponent strips HTML comments, collapses whitespace runs to a
// single space, and trims the result. Used for cross-stack snapshot comparison.
func NormaliseComponent(html string) string {
	html = commentRe.ReplaceAllString(html, "")
	html = whitespaceRe.ReplaceAllString(html, " ")
	return strings.TrimSpace(html)
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
		if strings.HasPrefix(trimmed, startMarker) {
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
