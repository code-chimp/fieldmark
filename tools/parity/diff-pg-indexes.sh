#!/usr/bin/env bash
# Compares the live domain.* pg_indexes against the committed canonical snapshot.
# Exits 0 if they match; exits non-zero and prints the diff otherwise.
# To refresh the canonical: tools/parity/dump-pg-indexes.sh > tools/parity/canonical-pg-indexes.txt
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CANONICAL="$SCRIPT_DIR/canonical-pg-indexes.txt"

if [ ! -f "$CANONICAL" ]; then
    echo "ERROR: canonical-pg-indexes.txt not found at $CANONICAL" >&2
    echo "  Run: tools/parity/dump-pg-indexes.sh > tools/parity/canonical-pg-indexes.txt" >&2
    exit 1
fi

TMPFILE="$(mktemp)" || { echo "ERROR: mktemp failed" >&2; exit 1; }
trap 'rm -f "$TMPFILE"' EXIT

echo "  dumping live domain pg_indexes..."
"$SCRIPT_DIR/dump-pg-indexes.sh" > "$TMPFILE"

if diff -u "$CANONICAL" "$TMPFILE" > /dev/null 2>&1; then
    INDEX_COUNT=$(wc -l < "$CANONICAL" | tr -d ' ')
    echo "OK pg_indexes parity verified ($INDEX_COUNT indexes)"
    exit 0
else
    echo "DRIFT [domain pg_indexes vs canonical]"
    diff -u "$CANONICAL" "$TMPFILE" || true
    exit 1
fi
