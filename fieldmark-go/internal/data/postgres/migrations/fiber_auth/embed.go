// Package fiberauthmigrations embeds the framework-local DDL for the
// fiber_auth schema and exposes it as a single SQL script applied by
// cmd/migrate-fiber-auth.
//
// ADR-012: fiber_auth tables are framework-local. domain.* DDL lives in
// docker/postgres/init/ and is owned by infrastructure (ADR-014); nothing
// in this package may target domain.*.
package fiberauthmigrations

import _ "embed"

//go:embed 001_initial.sql
var InitialSQL string
