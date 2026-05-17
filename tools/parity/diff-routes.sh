#!/usr/bin/env bash
# Compares the route inventory across all three stacks.
# Exits 0 if all three match; exits non-zero and prints the diff otherwise.
#
# Strategy: diff .NET vs Django and .NET vs Fiber. If A=B and A=C then B=C,
# so the third pairwise comparison is redundant and omitted.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

TMPDIR_PARITY=""

cleanup() {
    [ -n "$TMPDIR_PARITY" ] && rm -rf "$TMPDIR_PARITY"
}
trap cleanup EXIT

TMPDIR_PARITY="$(mktemp -d)" || { echo "ERROR: mktemp -d failed" >&2; exit 1; }
NET_FILE="$TMPDIR_PARITY/net.txt"
DJANGO_FILE="$TMPDIR_PARITY/django.txt"
FIBER_FILE="$TMPDIR_PARITY/fiber.txt"

echo "  dumping .NET routes..."
"$SCRIPT_DIR/dump-routes-net.sh" > "$NET_FILE"

echo "  dumping Django routes..."
"$SCRIPT_DIR/dump-routes-django.sh" > "$DJANGO_FILE"

echo "  dumping Fiber routes..."
"$SCRIPT_DIR/dump-routes-fiber.sh" > "$FIBER_FILE"

FAIL=0

diff_pair() {
    local label="$1"
    local file_a="$2"
    local file_b="$3"
    if ! diff -u "$file_a" "$file_b" > /dev/null 2>&1; then
        echo "DRIFT [$label]"
        diff -u "$file_a" "$file_b" || true
        FAIL=1
    fi
}

# Two comparisons suffice: .NET is the reference. If .NET=Django and .NET=Fiber,
# transitivity guarantees Django=Fiber.
diff_pair ".NET vs Django" "$NET_FILE" "$DJANGO_FILE"
diff_pair ".NET vs Fiber"  "$NET_FILE" "$FIBER_FILE"

if [ "$FAIL" -eq 0 ]; then
    echo "OK routes parity verified ($(wc -l < "$NET_FILE" | tr -d ' ') routes)"
    exit 0
else
    exit 1
fi
