"""Story 2.2 AC5 — Django transactional integrity for ``append_audit_entry``.

Two phases.

**Rollback phase**: open a single connection, insert a project row + audit
entry inside an explicit ``with transaction.atomic():`` block, raise to force
rollback, then verify on a fresh connection that neither row persists.

**Commit phase**: same setup but allow the atomic block to commit, verify
both rows are visible from a fresh connection, then clean up the audit row
first (the FK from ``domain.audit_entry.project_id`` to ``domain.project`` is
``REFERENCES`` without ``ON DELETE CASCADE``) followed by the project row.
The cleanup is test-only — production code never issues UPDATE/DELETE
against ``domain.audit_entry``.
"""

from __future__ import annotations

import datetime
import os
import uuid
from collections.abc import Iterator
from contextlib import contextmanager

import psycopg
import pytest
from django.db import connection, transaction

from audit.actions import AuditAction
from audit.append import append_audit_entry


def _dsn() -> str:
    return (os.environ.get("FIELDMARK_DATABASE_URL") or "").strip() or (
        "postgres://fieldmark:fieldmark@localhost:5432/fieldmark"
    )


def _count(cur: psycopg.Cursor, sql: str, *params: object) -> int:
    cur.execute(sql, params)
    row = cur.fetchone()
    return int(row[0]) if row else 0


class _Rollback(Exception):
    """Raised inside the atomic block to force Django to roll the transaction back."""


@contextmanager
def _open_psycopg() -> Iterator[psycopg.Connection]:
    try:
        conn = psycopg.connect(_dsn(), autocommit=False)
    except psycopg.OperationalError as exc:
        pytest.skip(f"Postgres not reachable for integration test: {exc}")
    try:
        yield conn
    finally:
        conn.close()


@pytest.fixture
def django_db_unblock(request, django_db_blocker):
    """Bypass pytest-django's DB-access guard for raw-DB integration tests.

    pytest-django's auto-test-DB machinery is the wrong fit here: ``domain.*``
    is created by init scripts in the live container, not by Django migrations
    against an ephemeral ``test_*`` database. Unblock direct access and let
    DATABASES["default"] hit FIELDMARK_DATABASE_URL.

    Skip-vs-fail policy (hard-rules.md §"A skipped test is not a verified test"):
    the ``make test-django-integration`` lane is the guaranteed-precondition
    lane for these tests. When the ``integration`` marker is active, a missing
    ``domain.*`` schema is a defect (the lane should always see the live DB),
    so we ``pytest.fail``. When the marker is absent (the broader
    ``uv run pytest`` lane), pytest-django typically interposes a test DB
    without the schema; skipping there is acceptable. Unreachable Postgres is
    always treated as an environmental skip — matches the existing
    ``test_db_rollback.py`` convention.
    """
    # Detect the *lane*, not the marker on the test itself. The marker is
    # always present on these tests; what differentiates lanes is whether the
    # user passed ``-m integration`` on the command line. pytest stores that
    # in ``request.config.option.markexpr``. In the integration lane, missing
    # preconditions are defects (fail loudly); in the default lane they are
    # expected (skip).
    markexpr = (getattr(request.config.option, "markexpr", "") or "").strip()
    is_integration_lane = (
        "integration" in markexpr and "not integration" not in markexpr
    )
    with django_db_blocker.unblock():
        try:
            with connection.cursor() as cur:
                cur.execute("SELECT to_regclass('domain.project')")
                row = cur.fetchone()
                if row is None or row[0] is None:
                    msg = (
                        "domain.project not present on Django default connection — "
                        "integration test requires the live fieldmark DB."
                    )
                    if is_integration_lane:
                        pytest.fail(msg)
                    pytest.skip(msg)
        except psycopg.OperationalError as exc:
            if is_integration_lane:
                pytest.fail(f"Postgres not reachable in integration lane: {exc}")
            pytest.skip(f"Postgres not reachable for integration test: {exc}")
        yield


@pytest.mark.integration
def test_rollback_leaves_no_audit_row(django_db_unblock) -> None:
    project_id = uuid.uuid4()
    actor_id = uuid.uuid4()
    code = f"AUD_{uuid.uuid4().hex[:10].upper()}"

    # Both the raw project insert and the ORM audit insert must share the
    # same Django connection so they roll back together. Phase the project
    # insert through Django's connection rather than a side psycopg.
    try:
        with pytest.raises(_Rollback):
            with transaction.atomic():
                with connection.cursor() as cur:
                    cur.execute(
                        """
                        INSERT INTO domain.project
                            (id, code, name, status, start_date, compliance_score,
                             created_at, updated_at)
                        VALUES
                            (%s, %s, %s, %s, %s, %s, now(), now())
                        """,
                        (
                            str(project_id),
                            code,
                            "Audit Smoke Project",
                            "Active",
                            datetime.date(2026, 1, 1),
                            100,
                        ),
                    )
                append_audit_entry(
                    actor_id=actor_id,
                    action=AuditAction.PROJECT_CREATED,
                    entity_type="Project",
                    entity_id=project_id,
                    project_id=project_id,
                )
                raise _Rollback
    except psycopg.OperationalError as exc:
        pytest.skip(f"Postgres not reachable for integration test: {exc}")

    # Verify on a fresh connection that nothing persisted.
    with _open_psycopg() as conn:
        with conn.cursor() as cur:
            assert (
                _count(
                    cur,
                    "SELECT count(*) FROM domain.project WHERE id = %s",
                    str(project_id),
                )
                == 0
            ), "project rollback must not persist"
            assert (
                _count(
                    cur,
                    "SELECT count(*) FROM domain.audit_entry WHERE entity_id = %s",
                    str(project_id),
                )
                == 0
            ), "audit rollback must not leave an orphan row"


@pytest.mark.integration
def test_commit_persists_both_rows_then_cleanup_succeeds(django_db_unblock) -> None:
    project_id = uuid.uuid4()
    actor_id = uuid.uuid4()
    code = f"AUD_{uuid.uuid4().hex[:10].upper()}"

    try:
        with transaction.atomic():
            with connection.cursor() as cur:
                cur.execute(
                    """
                    INSERT INTO domain.project
                        (id, code, name, status, start_date, compliance_score,
                         created_at, updated_at)
                    VALUES
                        (%s, %s, %s, %s, %s, %s, now(), now())
                    """,
                    (
                        str(project_id),
                        code,
                        "Audit Smoke Project",
                        "Active",
                        datetime.date(2026, 1, 1),
                        100,
                    ),
                )
            append_audit_entry(
                actor_id=actor_id,
                action=AuditAction.PROJECT_PLACED_ON_HOLD,
                entity_type="Project",
                entity_id=project_id,
                project_id=project_id,
            )
    except psycopg.OperationalError as exc:
        pytest.skip(f"Postgres not reachable for integration test: {exc}")

    # Fresh connection — both rows present, action string verbatim.
    with _open_psycopg() as conn:
        try:
            with conn.cursor() as cur:
                assert (
                    _count(
                        cur,
                        "SELECT count(*) FROM domain.project WHERE id = %s",
                        str(project_id),
                    )
                    == 1
                )
                assert (
                    _count(
                        cur,
                        "SELECT count(*) FROM domain.audit_entry WHERE entity_id = %s",
                        str(project_id),
                    )
                    == 1
                )
                cur.execute(
                    "SELECT action FROM domain.audit_entry WHERE entity_id = %s",
                    (str(project_id),),
                )
                row = cur.fetchone()
                assert row is not None and row[0] == "ProjectPlacedOnHold"
        finally:
            # Cleanup runs unconditionally so a mid-assertion failure does not
            # leak test rows. Audit first (project_id FK has no CASCADE).
            with conn.cursor() as cur:
                cur.execute(
                    "DELETE FROM domain.audit_entry WHERE entity_id = %s",
                    (str(project_id),),
                )
                cur.execute(
                    "DELETE FROM domain.project WHERE id = %s",
                    (str(project_id),),
                )
            conn.commit()
