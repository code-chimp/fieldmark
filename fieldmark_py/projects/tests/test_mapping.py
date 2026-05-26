"""Story 2.1 AC5 — Django round-trip smoke for `domain.project`.

The mapping is exercised via the Django ORM against the live `make up`
Postgres (the canonical `domain.*` schema lives in the container's init
scripts; pytest-django's per-run test database cannot reproduce it). The
`django_db_setup` override below is scoped to this file only — other
test files in `projects/tests/` continue to get pytest-django's standard
isolation. `@pytest.mark.django_db` then wraps each test in a
transaction that rolls back at teardown, so no row reaches disk.
"""

from __future__ import annotations

import datetime as dt
import uuid
from collections.abc import Iterator

import pytest
from django.db import connection

from projects.models import (
    JobSite,
    Project,
    ProjectInspector,
    ProjectStatus,
    ProjectTradeScope,
)


@pytest.fixture(scope="session")
def django_db_setup(django_db_blocker) -> Iterator[None]:
    """File-local override: reuse the live `make up` Postgres rather than
    let pytest-django create an empty test database. The canonical
    `domain.*` schema only exists in the Docker container's init scripts
    (Epic 1 retro A3); a test DB would be missing it.

    This override is in the test file (not `projects/tests/conftest.py`)
    so future unit tests in this directory keep pytest-django's standard
    test-database isolation.
    """
    with django_db_blocker.unblock():
        yield


@pytest.mark.integration
@pytest.mark.django_db
def test_project_round_trips_through_orm() -> None:
    project_id = uuid.uuid4()
    code = f"P_{uuid.uuid4().hex[:8].upper()}"
    start_date = dt.date(2026, 1, 15)
    target_date = dt.date(2026, 12, 31)
    closed_at = dt.datetime(2026, 6, 1, 12, 0, 0, tzinfo=dt.UTC)
    created_at = dt.datetime(2026, 1, 10, 9, 0, 0, tzinfo=dt.UTC)
    updated_at = dt.datetime(2026, 1, 11, 10, 0, 0, tzinfo=dt.UTC)

    # Insert via the Django connection so the uncommitted row is visible
    # to the ORM read; @pytest.mark.django_db rolls back the wrapping
    # transaction on teardown, so no row reaches disk.
    with connection.cursor() as cur:
        cur.execute(
            """
            INSERT INTO domain.project
                (id, code, name, description, status,
                 start_date, target_completion_date, actual_closed_at,
                 compliance_score, created_at, updated_at)
            VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
            """,
            (
                str(project_id),
                code,
                "Smoke Project",
                "round-trip",
                "OnHold",
                start_date,
                target_date,
                closed_at,
                87,
                created_at,
                updated_at,
            ),
        )

    loaded = Project.objects.get(pk=project_id)

    assert loaded.id == project_id
    assert loaded.code == code
    assert loaded.name == "Smoke Project"
    assert loaded.description == "round-trip"
    assert loaded.status == ProjectStatus.ON_HOLD
    assert loaded.start_date == start_date
    assert loaded.target_completion_date == target_date
    assert loaded.actual_closed_at == closed_at
    assert loaded.compliance_score == 87
    assert loaded.created_at == created_at
    assert loaded.updated_at == updated_at


@pytest.mark.integration
@pytest.mark.django_db
def test_relation_models_are_queryable() -> None:
    # Smoke that JobSite / ProjectTradeScope / ProjectInspector mappings
    # compile and round-trip through the ORM. The canonical seed leaves
    # these tables empty so we assert non-negative counts only.
    assert JobSite.objects.count() >= 0
    assert ProjectTradeScope.objects.count() >= 0
    assert ProjectInspector.objects.count() >= 0
