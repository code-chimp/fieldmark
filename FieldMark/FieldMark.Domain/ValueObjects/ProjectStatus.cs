namespace FieldMark.Domain.ValueObjects;

// Storage form is the PascalCase string per the DDL CHECK constraint at
// docker/postgres/init/010_domain_tables.sql line 71 — 'Active', 'OnHold',
// 'Closed'. The DDL is binding; the epic AC's SCREAMING_SNAKE_CASE note
// (sourced from non-authoritative research/) is superseded.
public enum ProjectStatus
{
    Active,
    OnHold,
    Closed,
}
