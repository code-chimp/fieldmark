// Contract: docs/reference/component-canonical-examples.md
//
// Sibling args file for tab_strip.html. Defines TabStripArgs, TabSpec, and the
// tabTabindex helper registered in the template function map (main.go engine.AddFunc).
package components

import (
	"fmt"
	"strings"
)

// TabSpec describes a single tab in the tablist.
// BadgeCount is a pointer so nil signals "no badge" while 0 renders a badge with text "0".
type TabSpec struct {
	ID         string
	Label      string
	HxGet      string
	HxTarget   string
	BadgeCount *int
}

// HasBadge reports whether a badge should render. Use this in templates instead of
// {{if .BadgeCount}} because a pointer-to-zero is truthy as a pointer but would render
// the zero incorrectly with a plain if-check that dereferences to 0.
func (t TabSpec) HasBadge() bool {
	return t.BadgeCount != nil
}

// TabStripArgs is the data context for the tab_strip component template.
type TabStripArgs struct {
	ID          string // optional; OOB swap requires a non-empty ID
	AriaLabel   string // required — panics on execute if empty
	Tabs        []TabSpec
	ActiveIndex int
}

// TabTabindex returns "0" for the active tab and "-1" for all others.
// Registered as the "tabTabindex" template function.
func TabTabindex(activeIndex, i int) string {
	if i == activeIndex {
		return "0"
	}
	return "-1"
}

// TabAriaControls strips a leading "#" from the hxTarget CSS selector.
// aria-controls must be an element id without the "#" prefix.
func TabAriaControls(hxTarget string) string {
	if len(hxTarget) > 0 && hxTarget[0] == '#' {
		return hxTarget[1:]
	}
	return hxTarget
}

// TabRequiredString returns the value if non-empty/non-whitespace, or an error.
// Registered as the "tabRequired" template function. Used to enforce required string
// props (e.g. AriaLabel) at template-execution time rather than silently rendering empty.
func TabRequiredString(propName, value string) (string, error) {
	if strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("TabStrip: %s is required and must not be empty or whitespace", propName)
	}
	return value, nil
}
