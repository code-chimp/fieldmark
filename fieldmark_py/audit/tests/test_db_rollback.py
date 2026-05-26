"""Smoke test for action item A3 (Epic 1 retro): real-DB integration harness.

Opens a transaction against ``domain.trade_type``, asserts the row is visible
inside the transaction, rolls back via the ``domain_db`` fixture's teardown,
then opens a fresh connection to confirm nothing persisted. Story 2.2 will
build on this pattern to exercise ``append_audit_entry``.
"""

from __future__ import annotations

import os
import uuid

import psycopg
import pytest


@pytest.mark.integration
def test_insert_visible_in_transaction(domain_db: psycopg.Cursor) -> None:
    code = f"TEST_{uuid.uuid4().hex[:8].upper()}"
    domain_db.execute(
        "INSERT INTO domain.trade_type (id, code, name) VALUES (%s, %s, %s)",
        (str(uuid.uuid4()), code, "Rollback smoke"),
    )
    domain_db.execute(
        "SELECT count(*) FROM domain.trade_type WHERE code = %s",
        (code,),
    )
    row = domain_db.fetchone()
    assert row is not None and row[0] == 1


@pytest.mark.integration
def test_rollback_leaves_no_trace() -> None:
    code = f"TEST_{uuid.uuid4().hex[:8].upper()}"

    dsn = (os.environ.get("FIELDMARK_DATABASE_URL") or "").strip() or (
        "postgres://fieldmark:fieldmark@localhost:5432/fieldmark"
    )

    # Phase 1: open a tx, insert, roll back explicitly.
    try:
        conn = psycopg.connect(dsn, autocommit=False)
    except psycopg.OperationalError as exc:
        pytest.skip(f"Postgres not reachable for integration test: {exc}")

    try:
        with conn.cursor() as cur:
            cur.execute(
                "INSERT INTO domain.trade_type (id, code, name) VALUES (%s, %s, %s)",
                (str(uuid.uuid4()), code, "Rollback smoke"),
            )
        conn.rollback()
    finally:
        conn.close()

    # Phase 2: fresh connection — must not see the row.
    with psycopg.connect(dsn, autocommit=True) as conn2:
        with conn2.cursor() as cur:
            cur.execute(
                "SELECT count(*) FROM domain.trade_type WHERE code = %s",
                (code,),
            )
            row = cur.fetchone()
            assert row is not None and row[0] == 0, "rollback must not persist the row"


@pytest.mark.integration
def test_reference_seed_present(domain_db: psycopg.Cursor) -> None:
    domain_db.execute("SELECT count(*) FROM domain.trade_type WHERE active")
    row = domain_db.fetchone()
    assert row is not None and row[0] > 0, "init scripts should have populated reference data"
