// Contract: docs/reference/component-canonical-examples.md
//
// Sibling helper file for compliance_tile.html — hosts the pure band-resolver because
// Go's html/template cannot express the four-arm decision concisely. Other component
// templates that ship without a logic surface remain template-only.
package components

import "strconv"

// ComplianceTileArgs is the data context for the compliance_tile component template.
// Score uses *int so nil is the no-data signal; 0 is a legitimate score (Critical band).
type ComplianceTileArgs struct {
	Score *int
	Label string
	ID    string
	Band  complianceBand
}

// NewComplianceTileArgs constructs ComplianceTileArgs with the band pre-resolved.
func NewComplianceTileArgs(score *int, label, id string) ComplianceTileArgs {
	return ComplianceTileArgs{
		Score: score,
		Label: label,
		ID:    id,
		Band:  resolveComplianceBand(score),
	}
}

// DisplayValue returns the score as a string, or an em-dash for no-data / out-of-range.
func (a ComplianceTileArgs) DisplayValue() string {
	if a.Score == nil || *a.Score < 0 || *a.Score > 100 {
		return "—"
	}
	return strconv.Itoa(*a.Score)
}

type complianceBand struct {
	ValueClass     string
	ThresholdWord  string
	ThresholdClass string
	RenderP        bool
}

var (
	bandHealthy  = complianceBand{"text-success", "Healthy", "text-success", true}
	bandWatch    = complianceBand{"text-warning", "Watch", "text-warning", true}
	bandConcern  = complianceBand{"text-warning-strong", "Concern", "text-warning-strong", true}
	bandCritical = complianceBand{"text-danger", "Critical", "text-danger", true}
	bandNoData   = complianceBand{"text-neutral", "", "", false}
)

func resolveComplianceBand(score *int) complianceBand {
	if score == nil || *score < 0 || *score > 100 {
		return bandNoData
	}
	switch {
	case *score >= 90:
		return bandHealthy
	case *score >= 70:
		return bandWatch
	case *score >= 50:
		return bandConcern
	default:
		return bandCritical
	}
}
