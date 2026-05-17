#!/usr/bin/env bash
# Dumps the normalized .NET route inventory to stdout.
# Output format: one line per route, "METHOD /path", sorted, all lowercase.
# dotnet run emits "Using launch settings..." and "Building..." to stdout;
# we filter to only lines matching the canonical route format.
# Real dotnet errors surface on stderr (not suppressed here).
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

dotnet run --project "$REPO_ROOT/FieldMark/FieldMark.Web" -- --dump-routes \
    | grep -E '^(get|post|put|patch|delete|head|options) /' || true
