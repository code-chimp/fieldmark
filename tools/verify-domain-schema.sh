#!/usr/bin/env bash
# Verifies the FieldMark canonical domain.* schema after `make reset` or `make up`.
# Exits 0 with a success banner; exits non-zero with a precise diff on failure.
# Side-effect-free: read-only queries only.
set -euo pipefail

DSN="postgresql://fieldmark:fieldmark@localhost:5432/fieldmark"

PASS=0
FAIL=0
MESSAGES=()

psql_query() {
    psql -tAq "$DSN" -c "$1" 2>&1
}

check() {
    local label="$1"
    local result="$2"
    local expected="$3"

    if [ "$result" = "$expected" ]; then
        PASS=$((PASS + 1))
    else
        FAIL=$((FAIL + 1))
        MESSAGES+=("FAIL [$label]")
        MESSAGES+=("  expected: $expected")
        MESSAGES+=("  got:      $result")
    fi
}

# ---------------------------------------------------------------------------
# 1. Schema presence — exactly 5 user schemas
# ---------------------------------------------------------------------------
SCHEMA_LIST=$(psql_query "
    SELECT schema_name
    FROM information_schema.schemata
    WHERE schema_name NOT IN ('public', 'information_schema', 'pg_catalog', 'pg_toast')
      AND schema_name NOT LIKE 'pg_%'
    ORDER BY schema_name
" | tr '\n' ',')
SCHEMA_LIST="${SCHEMA_LIST%,}"  # strip trailing comma

EXPECTED_SCHEMAS="django_auth,domain,dotnet_auth,fiber_auth,infra"
check "schemas present" "$SCHEMA_LIST" "$EXPECTED_SCHEMAS"

# ---------------------------------------------------------------------------
# 2. Table presence — exactly 12 tables in domain.*
# ---------------------------------------------------------------------------
TABLE_LIST=$(psql_query "
    SELECT table_name
    FROM information_schema.tables
    WHERE table_schema = 'domain'
    ORDER BY table_name
" | tr '\n' ',')
TABLE_LIST="${TABLE_LIST%,}"

EXPECTED_TABLES="audit_entry,compliance_rule,corrective_action,finding,inspection,job_site,project,project_inspector,project_trade_scope,trade_type,violation,violation_category"
check "domain tables" "$TABLE_LIST" "$EXPECTED_TABLES"

# Table count
TABLE_COUNT=$(psql_query "
    SELECT count(*)
    FROM information_schema.tables
    WHERE table_schema = 'domain'
")
check "domain table count" "$TABLE_COUNT" "12"

# ---------------------------------------------------------------------------
# 3. Reference seed row counts
# ---------------------------------------------------------------------------
TRADE_COUNT=$(psql_query "SELECT count(*) FROM domain.trade_type")
if [ "$TRADE_COUNT" -ge 4 ] 2>/dev/null; then
    PASS=$((PASS + 1))
else
    FAIL=$((FAIL + 1))
    MESSAGES+=("FAIL [trade_type count >= 4]: got $TRADE_COUNT")
fi

CATEGORY_COUNT=$(psql_query "SELECT count(*) FROM domain.violation_category")
if [ "$CATEGORY_COUNT" -ge 4 ] 2>/dev/null; then
    PASS=$((PASS + 1))
else
    FAIL=$((FAIL + 1))
    MESSAGES+=("FAIL [violation_category count >= 4]: got $CATEGORY_COUNT")
fi

RULE_COUNT=$(psql_query "SELECT count(*) FROM domain.compliance_rule")
check "compliance_rule count = 4" "$RULE_COUNT" "4"

# ---------------------------------------------------------------------------
# 4. Spot-check seed content — ELEC and PLUMB codes must be active
# ---------------------------------------------------------------------------
SPOT_COUNT=$(psql_query "
    SELECT count(*)
    FROM domain.trade_type
    WHERE code IN ('ELEC', 'PLUMB')
      AND active = true
")
check "trade_type spot-check (ELEC + PLUMB active)" "$SPOT_COUNT" "2"

# ---------------------------------------------------------------------------
# Result
# ---------------------------------------------------------------------------
if [ "$FAIL" -eq 0 ]; then
    echo "OK domain schema verified (5 schemas, 12 tables, $TRADE_COUNT trade types, $CATEGORY_COUNT violation categories, $RULE_COUNT compliance rules)"
    exit 0
else
    echo "FAIL domain schema verification failed ($FAIL check(s) failed, $PASS passed)"
    for msg in "${MESSAGES[@]}"; do
        echo "  $msg"
    done
    exit 1
fi
