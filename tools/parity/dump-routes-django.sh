#!/usr/bin/env bash
# Dumps the normalized Django route inventory to stdout.
# Output format: one line per route, "METHOD /path", sorted, all lowercase.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

cd "$REPO_ROOT/fieldmark_py"
exec uv run python manage.py dump_routes
