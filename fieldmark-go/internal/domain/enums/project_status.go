// Package enums holds string-typed enums whose storage values are pinned by
// the canonical DDL. Behavior methods live with the consuming entity (none
// for ProjectStatus this story; lands in Story 2.8 / 2.12).
package enums

// ProjectStatus values match the DDL CHECK constraint on domain.project.status
// exactly — PascalCase per docker/postgres/init/010_domain_tables.sql line 71.
// The DDL is binding; the epic AC's SCREAMING_SNAKE_CASE note is superseded.
type ProjectStatus string

const (
	ProjectStatusActive ProjectStatus = "Active"
	ProjectStatusOnHold ProjectStatus = "OnHold"
	ProjectStatusClosed ProjectStatus = "Closed"
)
