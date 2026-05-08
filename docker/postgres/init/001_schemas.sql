-- FieldMark Postgres bootstrap: schema creation.
--
-- This script is mounted into the postgres container at
-- /docker-entrypoint-initdb.d/001_schemas.sql and runs automatically on first
-- container startup (when the data volume is empty).
--
-- Schema ownership (see _bmad-output/planning-artifacts/research/architecture-decisions.md):
--   - domain        infrastructure-owned (ADR-014); shared business tables
--   - django_auth   Django stack auth/admin/sessions (ADR-012)
--   - dotnet_auth   ASP.NET Core Identity tables (ADR-012)
--   - fiber_auth    Go/Fiber auth tables when implemented (ADR-012)
--   - infra         reserved for cross-stack metadata (e.g. migration ledgers)
--
-- Per ADR-013, schemas are infrastructure, not framework data. No framework
-- migration tooling is permitted to CREATE SCHEMA. If you need a new schema,
-- add it here, destroy the volume, and recreate it.

CREATE SCHEMA IF NOT EXISTS domain;
CREATE SCHEMA IF NOT EXISTS django_auth;
CREATE SCHEMA IF NOT EXISTS dotnet_auth;
CREATE SCHEMA IF NOT EXISTS fiber_auth;
CREATE SCHEMA IF NOT EXISTS infra;
