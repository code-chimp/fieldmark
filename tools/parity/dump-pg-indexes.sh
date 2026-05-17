#!/usr/bin/env bash
# Dumps the live domain.* index inventory from PostgreSQL to stdout.
# Format: "indexname|indexdef" — one row per index, ordered by indexname.
# Pipe output to canonical-pg-indexes.txt to refresh the snapshot.
set -euo pipefail

DSN="${FIELDMARK_DATABASE_URL:-postgresql://fieldmark:fieldmark@localhost:5432/fieldmark}"

psql -tAq "$DSN" -c "
    SELECT indexname, indexdef
    FROM pg_indexes
    WHERE schemaname = 'domain'
    ORDER BY indexname
"
