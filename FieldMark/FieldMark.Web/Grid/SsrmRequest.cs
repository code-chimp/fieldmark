// SSRM request types and parser — docs/reference/ag-grid-ssrm-contract.md
// Validates all client-supplied colIds, operators, and values against strict
// allowlists; produces an error message for any violation.
using System.Text.Json;
using System.Text.Json.Serialization;

namespace FieldMark.Web.Grid;

/// <summary>
/// AG Grid IServerSideGetRowsRequest (camelCase vendor vocabulary; not subject to
/// the project's snake_case-on-the-wire rule).
/// </summary>
public sealed class SsrmRequest
{
    [JsonPropertyName("startRow")]
    public int StartRow { get; init; }

    [JsonPropertyName("endRow")]
    public int EndRow { get; init; }

    [JsonPropertyName("sortModel")]
    public List<SsrmSortItem> SortModel { get; init; } = [];

    [JsonPropertyName("filterModel")]
    public Dictionary<string, SsrmFilterEntry> FilterModel { get; init; } = [];
}

public sealed class SsrmSortItem
{
    [JsonPropertyName("colId")]
    public string ColId { get; init; } = "";

    [JsonPropertyName("sort")]
    public string Sort { get; init; } = "";
}

public sealed class SsrmFilterEntry
{
    [JsonPropertyName("filterType")]
    public string FilterType { get; init; } = "";

    [JsonPropertyName("type")]
    public string? Type { get; init; }

    // AG Grid sends "filter" as either a string (text filters) or a number (number filters).
    // JsonElement? accepts both without type-mismatch errors; callers read it via
    // FilterAsString / FilterAsDouble helpers below.
    [JsonPropertyName("filter")]
    public JsonElement? Filter { get; init; }

    [JsonPropertyName("filterTo")]
    public JsonElement? FilterTo { get; init; }

    [JsonPropertyName("dateFrom")]
    public string? DateFrom { get; init; }

    [JsonPropertyName("dateTo")]
    public string? DateTo { get; init; }

    [JsonPropertyName("values")]
    public List<string>? Values { get; init; }

    // Convenience accessors used by the handler.
    public string? FilterAsString =>
        Filter?.ValueKind == JsonValueKind.String ? Filter.Value.GetString() : null;

    public double? FilterAsDouble
    {
        get
        {
            if (Filter?.ValueKind is not JsonValueKind.Number) return null;
            // TryGetDouble avoids OverflowException on out-of-range JSON numbers.
            return Filter.Value.TryGetDouble(out var d) ? d : null;
        }
    }

    public double? FilterToAsDouble
    {
        get
        {
            if (FilterTo?.ValueKind is not JsonValueKind.Number) return null;
            return FilterTo.Value.TryGetDouble(out var d) ? d : null;
        }
    }
}

public sealed class SsrmValidationException(string message) : Exception(message) { }

/// <summary>
/// Validated, parsed pagination + sort + filter ready for query construction.
/// MatchNothing is true when a Set Filter has values=[]; callers short-circuit
/// to {rows:[], lastRow:0} without hitting the DB.
/// </summary>
public sealed class SsrmParsed
{
    public bool MatchNothing { get; init; }
    public int Limit { get; init; }
    public int Offset { get; init; }
    // Validated sort entries; first is primary. Tiebreaker "id ASC" is always appended.
    public IReadOnlyList<(string ColId, bool Descending)> SortEntries { get; init; } = [];
    // Validated filter entries for the handler to apply via IQueryable.
    public IReadOnlyList<SsrmFilterEntry> ValidatedFilters { get; init; } = [];
    // Map from colId to its validated filter entry (parallel to ValidatedFilters).
    public IReadOnlyDictionary<string, SsrmFilterEntry> FiltersByColId { get; init; }
        = new Dictionary<string, SsrmFilterEntry>();
}

public static class SsrmParser
{
    private static readonly HashSet<string> ColAllowlist =
    [
        "code", "name", "status", "compliance_score", "start_date", "target_completion_date",
    ];

    // Each column accepts exactly one filterType. Sending the wrong type → 400.
    private static readonly Dictionary<string, string> ColFilterType = new()
    {
        ["code"]                   = "text",
        ["name"]                   = "text",
        ["status"]                 = "set",
        ["compliance_score"]       = "number",
        ["start_date"]             = "date",
        ["target_completion_date"] = "date",
    };

    private static readonly HashSet<string> StatusAllowlist = ["Active", "OnHold", "Closed"];

    private static readonly HashSet<string> TextOps =
    [
        "equals", "notEqual", "contains", "notContains", "startsWith", "endsWith", "blank", "notBlank",
    ];
    private static readonly HashSet<string> NumberOps =
    [
        "equals", "notEqual", "greaterThan", "greaterThanOrEqual", "lessThan", "lessThanOrEqual",
        "inRange", "blank", "notBlank",
    ];
    private static readonly HashSet<string> DateOps =
    [
        "equals", "notEqual", "greaterThan", "lessThan", "inRange", "blank", "notBlank",
    ];

    public static SsrmParsed Parse(SsrmRequest req)
    {
        if (req.StartRow < 0)
            throw new SsrmValidationException("startRow must be >= 0");
        if (req.EndRow <= req.StartRow)
            throw new SsrmValidationException("endRow must be greater than startRow");
        if (req.EndRow - req.StartRow > 1000)
            throw new SsrmValidationException("page size exceeds maximum of 1000");

        // --- Sort validation ---
        // req.SortModel/FilterModel default to [] / {} but explicit JSON null overrides the default.
        if (req.SortModel is null)
            throw new SsrmValidationException("sortModel must be an array");
        if (req.FilterModel is null)
            throw new SsrmValidationException("filterModel must be an object");

        var sortEntries = new List<(string, bool)>();
        foreach (var s in req.SortModel)
        {
            if (s is null)
                throw new SsrmValidationException("sortModel entries must not be null");
            if (!ColAllowlist.Contains(s.ColId))
                throw new SsrmValidationException($"unknown column: {s.ColId}");
            if (s.Sort != "asc" && s.Sort != "desc")
                throw new SsrmValidationException($"invalid sort direction: {s.Sort}");
            sortEntries.Add((s.ColId, s.Sort == "desc"));
        }
        sortEntries.Add(("id", false)); // stable tiebreaker

        // --- Filter validation ---
        bool matchNothing = false;
        var validatedFilters = new Dictionary<string, SsrmFilterEntry>();

        foreach (var (colId, f) in req.FilterModel)
        {
            if (f is null)
                throw new SsrmValidationException($"filterModel entry for '{colId}' must not be null");
            if (!ColAllowlist.Contains(colId))
                throw new SsrmValidationException($"unknown column: {colId}");

            // Reject filterType that doesn't match the column's declared type.
            if (ColFilterType.TryGetValue(colId, out var expectedType)
                && f.FilterType != expectedType)
                throw new SsrmValidationException(
                    $"column '{colId}' only accepts filterType '{expectedType}', got '{f.FilterType}'"
                );

            switch (f.FilterType)
            {
                case "set":
                    if (colId != "status")
                        throw new SsrmValidationException("set filter only supported on status column");
                    var values = f.Values ?? [];
                    foreach (var v in values)
                        if (!StatusAllowlist.Contains(v))
                            throw new SsrmValidationException($"invalid status value: {v}");
                    if (values.Count == 0)
                        matchNothing = true;
                    break;

                case "text":
                    if (!TextOps.Contains(f.Type ?? ""))
                        throw new SsrmValidationException(
                            $"invalid operator '{f.Type}' for column '{colId}'"
                        );
                    // For operators that use the filter value, it must be a JSON string.
                    // A non-string (e.g. numeric 42) would silently coerce to "" via FilterAsString.
                    if (f.Type is not ("blank" or "notBlank")
                        && f.Filter.HasValue
                        && f.Filter.Value.ValueKind != JsonValueKind.String)
                        throw new SsrmValidationException(
                            $"text filter 'filter' for column '{colId}' must be a string"
                        );
                    break;

                case "number":
                    if (!NumberOps.Contains(f.Type ?? ""))
                        throw new SsrmValidationException(
                            $"invalid operator '{f.Type}' for column '{colId}'"
                        );
                    // Validate that operands required by the operator are present as numbers.
                    if (f.Type is "equals" or "notEqual" or "greaterThan" or "greaterThanOrEqual"
                                or "lessThan" or "lessThanOrEqual" && f.FilterAsDouble is null)
                        throw new SsrmValidationException(
                            $"operator '{f.Type}' for column '{colId}' requires a numeric 'filter' value"
                        );
                    if (f.Type == "inRange" && (f.FilterAsDouble is null || f.FilterToAsDouble is null))
                        throw new SsrmValidationException(
                            $"inRange for column '{colId}' requires numeric 'filter' and 'filterTo' values"
                        );
                    break;

                case "date":
                    if (!DateOps.Contains(f.Type ?? ""))
                        throw new SsrmValidationException(
                            $"invalid operator '{f.Type}' for column '{colId}'"
                        );
                    // Validate dateFrom for operators that need it; reject malformed date strings.
                    if (f.Type is "equals" or "notEqual" or "greaterThan" or "lessThan")
                    {
                        if (string.IsNullOrEmpty(f.DateFrom))
                            throw new SsrmValidationException(
                                $"operator '{f.Type}' for column '{colId}' requires 'dateFrom'"
                            );
                        if (!DateOnly.TryParseExact(f.DateFrom, "yyyy-MM-dd", out _))
                            throw new SsrmValidationException(
                                $"invalid dateFrom value '{f.DateFrom}' for column '{colId}' — expected YYYY-MM-DD"
                            );
                    }
                    if (f.Type == "inRange")
                    {
                        if (string.IsNullOrEmpty(f.DateFrom) || string.IsNullOrEmpty(f.DateTo))
                            throw new SsrmValidationException(
                                $"inRange for column '{colId}' requires 'dateFrom' and 'dateTo'"
                            );
                        if (!DateOnly.TryParseExact(f.DateFrom, "yyyy-MM-dd", out _))
                            throw new SsrmValidationException(
                                $"invalid dateFrom value '{f.DateFrom}' for column '{colId}' — expected YYYY-MM-DD"
                            );
                        if (!DateOnly.TryParseExact(f.DateTo, "yyyy-MM-dd", out _))
                            throw new SsrmValidationException(
                                $"invalid dateTo value '{f.DateTo}' for column '{colId}' — expected YYYY-MM-DD"
                            );
                    }
                    break;

                default:
                    throw new SsrmValidationException($"unknown filterType for column {colId}");
            }

            validatedFilters[colId] = f;
        }

        return new SsrmParsed
        {
            MatchNothing = matchNothing,
            Limit = req.EndRow - req.StartRow,
            Offset = req.StartRow,
            SortEntries = sortEntries,
            FiltersByColId = validatedFilters,
        };
    }
}
