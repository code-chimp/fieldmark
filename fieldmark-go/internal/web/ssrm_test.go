package web_test

import (
	"encoding/json"
	"testing"

	web "github.com/code-chimp/fieldmark-go/internal/web"
)

func req(startRow, endRow int, sort []web.SsrmSortItem, filter map[string]json.RawMessage) web.SsrmRequest {
	if sort == nil {
		sort = []web.SsrmSortItem{}
	}
	if filter == nil {
		filter = map[string]json.RawMessage{}
	}
	return web.SsrmRequest{StartRow: startRow, EndRow: endRow, SortModel: sort, FilterModel: filter}
}

func rawJSON(s string) json.RawMessage { return json.RawMessage(s) }

// ─── Pagination bounds ────────────────────────────────────────────────────────

func TestParseSsrm_StartRowNegative_Returns400(t *testing.T) {
	_, err := web.ParseSsrmRequest(req(-1, 10, nil, nil))
	if err == nil {
		t.Fatal("expected error for negative startRow")
	}
}

func TestParseSsrm_EndRowNotGreaterThanStart_Returns400(t *testing.T) {
	_, err := web.ParseSsrmRequest(req(5, 5, nil, nil))
	if err == nil {
		t.Fatal("expected error for endRow == startRow")
	}
}

func TestParseSsrm_PageSizeExceedsMax_Returns400(t *testing.T) {
	_, err := web.ParseSsrmRequest(req(0, 1001, nil, nil))
	if err == nil {
		t.Fatal("expected error for page size > 1000")
	}
}

func TestParseSsrm_ValidPagination_LimitOffset(t *testing.T) {
	p, err := web.ParseSsrmRequest(req(10, 20, nil, nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Limit != 10 {
		t.Errorf("expected Limit=10, got %d", p.Limit)
	}
	if p.Offset != 10 {
		t.Errorf("expected Offset=10, got %d", p.Offset)
	}
}

// ─── Sort allowlist ───────────────────────────────────────────────────────────

func TestParseSsrm_UnknownSortColId_Returns400(t *testing.T) {
	_, err := web.ParseSsrmRequest(req(0, 10, []web.SsrmSortItem{{ColID: "description", Sort: "asc"}}, nil))
	if err == nil {
		t.Fatal("expected error for unknown colId")
	}
}

func TestParseSsrm_InvalidSortDirection_Returns400(t *testing.T) {
	_, err := web.ParseSsrmRequest(req(0, 10, []web.SsrmSortItem{{ColID: "code", Sort: "DESC"}}, nil))
	if err == nil {
		t.Fatal("expected error for invalid sort direction")
	}
}

func TestParseSsrm_InjectionInSortColId_Returns400(t *testing.T) {
	_, err := web.ParseSsrmRequest(req(0, 10, []web.SsrmSortItem{{ColID: "code; DROP TABLE domain.project --", Sort: "asc"}}, nil))
	if err == nil {
		t.Fatal("expected error for injection-style colId")
	}
}

func TestParseSsrm_ValidSort_AlwaysHasTiebreaker(t *testing.T) {
	p, err := web.ParseSsrmRequest(req(0, 10, []web.SsrmSortItem{{ColID: "compliance_score", Sort: "asc"}}, nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	last := p.OrderClauses[len(p.OrderClauses)-1]
	if last != "id ASC" {
		t.Errorf("expected tiebreaker 'id ASC', got %q", last)
	}
}

func TestParseSsrm_EmptySort_DefaultOrderIncludesTiebreaker(t *testing.T) {
	p, err := web.ParseSsrmRequest(req(0, 10, nil, nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, c := range p.OrderClauses {
		if c == "id ASC" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'id ASC' tiebreaker in order clauses, got %v", p.OrderClauses)
	}
}

// ─── Filter allowlist ─────────────────────────────────────────────────────────

func TestParseSsrm_UnknownFilterColId_Returns400(t *testing.T) {
	_, err := web.ParseSsrmRequest(req(0, 10, nil, map[string]json.RawMessage{
		"description": rawJSON(`{"filterType":"text","type":"contains","filter":"test"}`),
	}))
	if err == nil {
		t.Fatal("expected error for unknown filter colId")
	}
}

func TestParseSsrm_InjectionInFilterColId_Returns400(t *testing.T) {
	_, err := web.ParseSsrmRequest(req(0, 10, nil, map[string]json.RawMessage{
		"code; DROP TABLE domain.project --": rawJSON(`{"filterType":"text","type":"contains","filter":"x"}`),
	}))
	if err == nil {
		t.Fatal("expected error for injection-style filter colId")
	}
}

func TestParseSsrm_InvalidTextOperator_Returns400(t *testing.T) {
	_, err := web.ParseSsrmRequest(req(0, 10, nil, map[string]json.RawMessage{
		"code": rawJSON(`{"filterType":"text","type":"INVALID","filter":"x"}`),
	}))
	if err == nil {
		t.Fatal("expected error for invalid text operator")
	}
}

func TestParseSsrm_InvalidNumberOperator_Returns400(t *testing.T) {
	_, err := web.ParseSsrmRequest(req(0, 10, nil, map[string]json.RawMessage{
		"compliance_score": rawJSON(`{"filterType":"number","type":"INVALID","filter":50}`),
	}))
	if err == nil {
		t.Fatal("expected error for invalid number operator")
	}
}

func TestParseSsrm_InvalidDateOperator_Returns400(t *testing.T) {
	_, err := web.ParseSsrmRequest(req(0, 10, nil, map[string]json.RawMessage{
		"start_date": rawJSON(`{"filterType":"date","type":"INVALID","dateFrom":"2026-01-01"}`),
	}))
	if err == nil {
		t.Fatal("expected error for invalid date operator")
	}
}

func TestParseSsrm_InvalidStatusValue_Returns400(t *testing.T) {
	_, err := web.ParseSsrmRequest(req(0, 10, nil, map[string]json.RawMessage{
		"status": rawJSON(`{"filterType":"set","values":["Active","INVALID"]}`),
	}))
	if err == nil {
		t.Fatal("expected error for invalid status value")
	}
}

func TestParseSsrm_InjectionInStatusValue_Returns400(t *testing.T) {
	_, err := web.ParseSsrmRequest(req(0, 10, nil, map[string]json.RawMessage{
		"status": rawJSON(`{"filterType":"set","values":["Active' OR '1'='1"]}`),
	}))
	if err == nil {
		t.Fatal("expected error for injection-style status value")
	}
}

func TestParseSsrm_EmptyStatusValues_ReturnsFalseFragment(t *testing.T) {
	p, err := web.ParseSsrmRequest(req(0, 10, nil, map[string]json.RawMessage{
		"status": rawJSON(`{"filterType":"set","values":[]}`),
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(p.WhereFragments) == 0 || p.WhereFragments[0] != "FALSE" {
		t.Errorf("expected FALSE fragment for empty status values, got %v", p.WhereFragments)
	}
}

// ─── SQL generation ───────────────────────────────────────────────────────────

func TestParseSsrm_NumberFilterMissingOperand_Returns400(t *testing.T) {
	// "equals" without a "filter" value must → 400, not nil-deref panic.
	_, err := web.ParseSsrmRequest(req(0, 10, nil, map[string]json.RawMessage{
		"compliance_score": rawJSON(`{"filterType":"number","type":"equals"}`),
	}))
	if err == nil {
		t.Fatal("expected error for number filter missing 'filter' value")
	}
}

func TestParseSsrm_WrongFilterTypeForColumn_Returns400(t *testing.T) {
	// compliance_score only accepts "number"; sending "text" must → 400.
	_, err := web.ParseSsrmRequest(req(0, 10, nil, map[string]json.RawMessage{
		"compliance_score": rawJSON(`{"filterType":"text","type":"contains","filter":"50"}`),
	}))
	if err == nil {
		t.Fatal("expected error for wrong filterType on compliance_score")
	}
}

func TestParseSsrm_TextFilterTypeOnDateColumn_Returns400(t *testing.T) {
	_, err := web.ParseSsrmRequest(req(0, 10, nil, map[string]json.RawMessage{
		"start_date": rawJSON(`{"filterType":"text","type":"contains","filter":"2026"}`),
	}))
	if err == nil {
		t.Fatal("expected error for text filterType on start_date")
	}
}

func TestParseSsrm_DateFilterMissingDateFrom_Returns400(t *testing.T) {
	_, err := web.ParseSsrmRequest(req(0, 10, nil, map[string]json.RawMessage{
		"start_date": rawJSON(`{"filterType":"date","type":"equals"}`),
	}))
	if err == nil {
		t.Fatal("expected error for date filter missing dateFrom")
	}
}

func TestParseSsrm_DateFilterMalformedDate_Returns400(t *testing.T) {
	_, err := web.ParseSsrmRequest(req(0, 10, nil, map[string]json.RawMessage{
		"start_date": rawJSON(`{"filterType":"date","type":"equals","dateFrom":"not-a-date"}`),
	}))
	if err == nil {
		t.Fatal("expected error for malformed dateFrom value")
	}
}

func TestParseSsrm_DateFilterInRange_MissingDateTo_Returns400(t *testing.T) {
	_, err := web.ParseSsrmRequest(req(0, 10, nil, map[string]json.RawMessage{
		"start_date": rawJSON(`{"filterType":"date","type":"inRange","dateFrom":"2026-01-01"}`),
	}))
	if err == nil {
		t.Fatal("expected error for inRange missing dateTo")
	}
}

func TestParseSsrm_TextContainsFilter_ParameterizedILIKE(t *testing.T) {
	p, err := web.ParseSsrmRequest(req(0, 10, nil, map[string]json.RawMessage{
		"code": rawJSON(`{"filterType":"text","type":"contains","filter":"BLDG"}`),
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	where := p.BuildWhereClause()
	if where == "" {
		t.Fatal("expected non-empty WHERE clause")
	}
	// Filter value must be a bind parameter, not interpolated.
	if len(p.Args) == 0 {
		t.Fatal("expected at least one bind argument")
	}
	if p.Args[0] != "%BLDG%" {
		t.Errorf("expected arg[0]='%%BLDG%%', got %v", p.Args[0])
	}
}
