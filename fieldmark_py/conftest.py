"""Top-level pytest configuration for the Django stack.

The ``domain_db`` fixture (see audit/tests/test_db_rollback.py for usage) opens
a raw psycopg connection to the running ``make up`` Postgres, begins a
transaction, yields a cursor, and rolls back on teardown. This sidesteps the
auto-created pytest-django test database — the canonical ``domain.*`` schema
lives in init scripts the DB container runs once at first boot, so we exercise
the real schema directly per the project convention (real PostgreSQL only, no
SQLite, no Django-managed DDL on domain tables).

Tests that need the fixture wear the ``integration`` marker so ``make
test-django`` can stay fast and the integration lane (``make
test-django-integration``) is opted into explicitly.
"""

from __future__ import annotations

import os
from collections.abc import Iterator

import psycopg
import pytest


def pytest_configure(config: pytest.Config) -> None:
    config.addinivalue_line(
        "markers",
        "integration: requires a running Postgres with the canonical domain.* schema (make up).",
    )


@pytest.fixture
def domain_db() -> Iterator[psycopg.Cursor]:
    """Yield a cursor inside a transaction that rolls back on teardown.

    Connects to ``FIELDMARK_DATABASE_URL`` (falling back to the local compose
    DSN). Autocommit is left off so the explicit rollback in teardown
    guarantees no row reaches disk — Story 2.2 will rely on this to verify
    append_audit_entry's transactional semantics.
    """

    dsn = (os.environ.get("FIELDMARK_DATABASE_URL") or "").strip() or (
        "postgres://fieldmark:fieldmark@localhost:5432/fieldmark"
    )

    try:
        conn = psycopg.connect(dsn, autocommit=False)
    except psycopg.OperationalError as exc:
        pytest.skip(f"Postgres not reachable for integration test: {exc}")

    try:
        with conn.cursor() as cur:
            yield cur
    finally:
        conn.rollback()
        conn.close()
