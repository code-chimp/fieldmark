// Package web — ssrm.go
//
// SSRM request parser and SQL translator for AG Grid server-side row model
// endpoints. Validates all client-supplied colIds, operators, and values
// against strict allowlists before producing parameterized SQL clauses.
//
// Contract: docs/reference/ag-grid-ssrm-contract.md
package web

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// SsrmRequest mirrors AG Grid's IServerSideGetRowsRequest (camelCase vendor vocabulary).
type SsrmRequest struct {
	StartRow    int                        `json:"startRow"`
	EndRow      int                        `json:"endRow"`
	SortModel   []SsrmSortItem             `json:"sortModel"`
	FilterModel map[string]json.RawMessage `json:"filterModel"`
}

// SsrmSortItem is one entry in sortModel.
type SsrmSortItem struct {
	ColID string `json:"colId"`
	Sort  string `json:"sort"`
}

// SsrmParsed is the validated, translated result ready for SQL construction.
type SsrmParsed struct {
	Limit  int
	Offset int
	// OrderClauses are safe SQL tokens (never raw client strings).
	OrderClauses []string
	// WhereFragments are parameterised fragments ("code ILIKE $N").
	WhereFragments []string
	// Args holds bind values corresponding to $1..$N placeholders.
	Args []any
}

// projectColAllowlist maps allowlisted colId strings to their safe SQL column names.
var projectColAllowlist = map[string]string{
	"code":                   "code",
	"name":                   "name",
	"status":                 "status",
	"compliance_score":       "compliance_score",
	"start_date":             "start_date",
	"target_completion_date": "target_completion_date",
}

// colFilterType is the single allowed filterType per column.
// A request carrying the wrong filterType for a column → 400.
var colFilterType = map[string]string{
	"code":                   "text",
	"name":                   "text",
	"status":                 "set",
	"compliance_score":       "number",
	"start_date":             "date",
	"target_completion_date": "date",
}

var textOperators = map[string]bool{
	"equals": true, "notEqual": true, "contains": true, "notContains": true,
	"startsWith": true, "endsWith": true, "blank": true, "notBlank": true,
}
var numberOperators = map[string]bool{
	"equals": true, "notEqual": true, "greaterThan": true, "greaterThanOrEqual": true,
	"lessThan": true, "lessThanOrEqual": true, "inRange": true, "blank": true, "notBlank": true,
}
var dateOperators = map[string]bool{
	"equals": true, "notEqual": true, "greaterThan": true, "lessThan": true,
	"inRange": true, "blank": true, "notBlank": true,
}
var statusAllowlist = map[string]bool{"Active": true, "OnHold": true, "Closed": true}

// ParseSsrmRequest validates and translates an SsrmRequest into SsrmParsed.
// Returns a non-nil error (with a human-readable message) on any validation failure.
func ParseSsrmRequest(r SsrmRequest) (*SsrmParsed, error) {
	if r.StartRow < 0 {
		return nil, fmt.Errorf("startRow must be >= 0")
	}
	if r.EndRow <= r.StartRow {
		return nil, fmt.Errorf("endRow must be greater than startRow")
	}
	if r.EndRow-r.StartRow > 1000 {
		return nil, fmt.Errorf("page size exceeds maximum of 1000")
	}

	p := &SsrmParsed{
		Limit:  r.EndRow - r.StartRow,
		Offset: r.StartRow,
	}

	// --- Sort ---
	for _, s := range r.SortModel {
		col, ok := projectColAllowlist[s.ColID]
		if !ok {
			return nil, fmt.Errorf("unknown column: %s", s.ColID)
		}
		if s.Sort != "asc" && s.Sort != "desc" {
			return nil, fmt.Errorf("invalid sort direction: %s", s.Sort)
		}
		p.OrderClauses = append(p.OrderClauses, col+" "+s.Sort)
	}
	// Always append tiebreaker (stable pagination per contract doc).
	p.OrderClauses = append(p.OrderClauses, "id ASC")

	// --- Filter ---
	for colID, rawFilter := range r.FilterModel {
		col, ok := projectColAllowlist[colID]
		if !ok {
			return nil, fmt.Errorf("unknown column: %s", colID)
		}

		// Peek at filterType
		var peek struct {
			FilterType string `json:"filterType"`
		}
		if err := json.Unmarshal(rawFilter, &peek); err != nil {
			return nil, fmt.Errorf("invalid filter for column %s", colID)
		}

		// Reject filterType that doesn't match the column's declared type.
		if expected, known := colFilterType[colID]; known && peek.FilterType != expected {
			return nil, fmt.Errorf("column '%s' only accepts filterType '%s', got '%s'", colID, expected, peek.FilterType)
		}

		switch peek.FilterType {
		case "set":
			if colID != "status" {
				return nil, fmt.Errorf("set filter only supported on status column")
			}
			var f struct {
				FilterType string   `json:"filterType"`
				Values     []string `json:"values"`
			}
			if err := json.Unmarshal(rawFilter, &f); err != nil {
				return nil, fmt.Errorf("invalid set filter for status")
			}
			for _, v := range f.Values {
				if !statusAllowlist[v] {
					return nil, fmt.Errorf("invalid status value: %s", v)
				}
			}
			if len(f.Values) == 0 {
				p.WhereFragments = append(p.WhereFragments, "FALSE")
			} else {
				idx := len(p.Args) + 1
				p.Args = append(p.Args, f.Values)
				p.WhereFragments = append(p.WhereFragments, fmt.Sprintf("%s = ANY($%d)", col, idx))
			}

		case "text":
			frag, err := parseTextFilter(col, rawFilter, &p.Args)
			if err != nil {
				return nil, err
			}
			if frag != "" {
				p.WhereFragments = append(p.WhereFragments, frag)
			}

		case "number":
			frag, err := parseNumberFilter(col, rawFilter, &p.Args)
			if err != nil {
				return nil, err
			}
			if frag != "" {
				p.WhereFragments = append(p.WhereFragments, frag)
			}

		case "date":
			frag, err := parseDateFilter(col, rawFilter, &p.Args)
			if err != nil {
				return nil, err
			}
			if frag != "" {
				p.WhereFragments = append(p.WhereFragments, frag)
			}

		default:
			return nil, fmt.Errorf("unknown filterType for column %s", colID)
		}
	}

	return p, nil
}

// BuildWhereClause returns "WHERE ..." or "" if no filters.
func (p *SsrmParsed) BuildWhereClause() string {
	if len(p.WhereFragments) == 0 {
		return ""
	}
	return "WHERE " + strings.Join(p.WhereFragments, " AND ")
}

// BuildOrderByClause returns "ORDER BY col1 asc, col2 desc, id ASC".
func (p *SsrmParsed) BuildOrderByClause() string {
	return "ORDER BY " + strings.Join(p.OrderClauses, ", ")
}

// --- text filter ---

func parseTextFilter(col string, raw json.RawMessage, args *[]any) (string, error) {
	var f struct {
		FilterType string `json:"filterType"`
		Type       string `json:"type"`
		Filter     string `json:"filter"`
	}
	if err := json.Unmarshal(raw, &f); err != nil {
		return "", fmt.Errorf("invalid text filter for %s", col)
	}
	if !textOperators[f.Type] {
		return "", fmt.Errorf("invalid operator '%s' for column '%s'", f.Type, col)
	}
	idx := len(*args) + 1
	switch f.Type {
	case "blank":
		return fmt.Sprintf("(%s IS NULL OR %s = '')", col, col), nil
	case "notBlank":
		return fmt.Sprintf("(%s IS NOT NULL AND %s != '')", col, col), nil
	case "equals":
		*args = append(*args, f.Filter)
		return fmt.Sprintf("%s = $%d", col, idx), nil
	case "notEqual":
		*args = append(*args, f.Filter)
		return fmt.Sprintf("%s != $%d", col, idx), nil
	case "contains":
		*args = append(*args, "%"+f.Filter+"%")
		return fmt.Sprintf("%s ILIKE $%d", col, idx), nil
	case "notContains":
		*args = append(*args, "%"+f.Filter+"%")
		return fmt.Sprintf("%s NOT ILIKE $%d", col, idx), nil
	case "startsWith":
		*args = append(*args, f.Filter+"%")
		return fmt.Sprintf("%s ILIKE $%d", col, idx), nil
	case "endsWith":
		*args = append(*args, "%"+f.Filter)
		return fmt.Sprintf("%s ILIKE $%d", col, idx), nil
	}
	return "", nil
}

// --- number filter ---

func parseNumberFilter(col string, raw json.RawMessage, args *[]any) (string, error) {
	var f struct {
		FilterType string   `json:"filterType"`
		Type       string   `json:"type"`
		Filter     *float64 `json:"filter"`
		FilterTo   *float64 `json:"filterTo"`
	}
	if err := json.Unmarshal(raw, &f); err != nil {
		return "", fmt.Errorf("invalid number filter for %s", col)
	}
	if !numberOperators[f.Type] {
		return "", fmt.Errorf("invalid operator '%s' for column '%s'", f.Type, col)
	}
	// Guard against nil-deref: operators that require a value must have filter present.
	needsFilter := f.Type != "blank" && f.Type != "notBlank"
	if needsFilter && f.Filter == nil {
		return "", fmt.Errorf("operator '%s' for column '%s' requires a numeric 'filter' value", f.Type, col)
	}
	if f.Type == "inRange" && f.FilterTo == nil {
		return "", fmt.Errorf("inRange for column '%s' requires numeric 'filterTo' value", col)
	}
	idx := len(*args) + 1
	switch f.Type {
	case "blank":
		return fmt.Sprintf("%s IS NULL", col), nil
	case "notBlank":
		return fmt.Sprintf("%s IS NOT NULL", col), nil
	case "equals":
		*args = append(*args, *f.Filter)
		return fmt.Sprintf("%s = $%d", col, idx), nil
	case "notEqual":
		*args = append(*args, *f.Filter)
		return fmt.Sprintf("%s != $%d", col, idx), nil
	case "greaterThan":
		*args = append(*args, *f.Filter)
		return fmt.Sprintf("%s > $%d", col, idx), nil
	case "greaterThanOrEqual":
		*args = append(*args, *f.Filter)
		return fmt.Sprintf("%s >= $%d", col, idx), nil
	case "lessThan":
		*args = append(*args, *f.Filter)
		return fmt.Sprintf("%s < $%d", col, idx), nil
	case "lessThanOrEqual":
		*args = append(*args, *f.Filter)
		return fmt.Sprintf("%s <= $%d", col, idx), nil
	case "inRange":
		*args = append(*args, *f.Filter, *f.FilterTo)
		return fmt.Sprintf("%s >= $%d AND %s <= $%d", col, idx, col, idx+1), nil
	}
	return "", nil
}

// --- date filter ---

const dateLayout = "2006-01-02"

func validateDate(val, col, field string) error {
	if val == "" {
		return fmt.Errorf("operator for column '%s' requires a non-empty '%s'", col, field)
	}
	if _, err := time.Parse(dateLayout, val); err != nil {
		return fmt.Errorf("invalid %s value '%s' for column '%s' — expected YYYY-MM-DD", field, val, col)
	}
	return nil
}

func parseDateFilter(col string, raw json.RawMessage, args *[]any) (string, error) {
	var f struct {
		FilterType string `json:"filterType"`
		Type       string `json:"type"`
		DateFrom   string `json:"dateFrom"`
		DateTo     string `json:"dateTo"`
	}
	if err := json.Unmarshal(raw, &f); err != nil {
		return "", fmt.Errorf("invalid date filter for %s", col)
	}
	if !dateOperators[f.Type] {
		return "", fmt.Errorf("invalid operator '%s' for column '%s'", f.Type, col)
	}

	// Validate date strings before building SQL — a malformed date would produce a DB error (500)
	// instead of the required 400 contract response.
	needsFrom := f.Type != "blank" && f.Type != "notBlank"
	if needsFrom {
		if err := validateDate(f.DateFrom, col, "dateFrom"); err != nil {
			return "", err
		}
	}
	if f.Type == "inRange" {
		if err := validateDate(f.DateTo, col, "dateTo"); err != nil {
			return "", err
		}
	}

	idx := len(*args) + 1
	switch f.Type {
	case "blank":
		return fmt.Sprintf("%s IS NULL", col), nil
	case "notBlank":
		return fmt.Sprintf("%s IS NOT NULL", col), nil
	case "equals":
		*args = append(*args, f.DateFrom)
		return fmt.Sprintf("%s = $%d", col, idx), nil
	case "notEqual":
		*args = append(*args, f.DateFrom)
		return fmt.Sprintf("%s != $%d", col, idx), nil
	case "greaterThan":
		*args = append(*args, f.DateFrom)
		return fmt.Sprintf("%s > $%d", col, idx), nil
	case "lessThan":
		*args = append(*args, f.DateFrom)
		return fmt.Sprintf("%s < $%d", col, idx), nil
	case "inRange":
		*args = append(*args, f.DateFrom, f.DateTo)
		return fmt.Sprintf("%s >= $%d AND %s <= $%d", col, idx, col, idx+1), nil
	}
	return "", nil
}
