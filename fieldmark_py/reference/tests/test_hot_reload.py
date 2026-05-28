from __future__ import annotations

from collections.abc import Iterator

import pytest
from django.db import connection

from reference import queries


@pytest.fixture(scope="session")
def django_db_setup(django_db_blocker) -> Iterator[None]:
    """Use the live `make up` database so domain.* comes from init scripts."""
    with django_db_blocker.unblock():
        yield


@pytest.fixture()
def live_reference_schema(request, django_db_blocker) -> Iterator[None]:
    markexpr = (getattr(request.config.option, "markexpr", "") or "").strip()
    is_integration_lane = "integration" in markexpr and "not integration" not in markexpr
    with django_db_blocker.unblock():
        with connection.cursor() as cur:
            cur.execute("SELECT to_regclass('domain.compliance_rule')")
            row = cur.fetchone()
        if row is None or row[0] is None:
            msg = (
                "domain.compliance_rule not present on Django default connection — "
                "integration test requires the live fieldmark DB."
            )
            if is_integration_lane:
                pytest.fail(msg)
            pytest.skip(msg)
        yield


@pytest.mark.integration
@pytest.mark.django_db
def test_queries_see_compliance_rule_update_without_recreating_reader(live_reference_schema) -> None:
    code = "OPEN_VIOLATION_GATE"
    updated_name = "Open Violation Closure Gate (UPDATED)"

    original = next(rule for rule in queries.list_compliance_rules() if rule.code == code)
    assert original.name != updated_name

    try:
        with connection.cursor() as cur:
            cur.execute(
                "UPDATE domain.compliance_rule SET name = %s WHERE code = %s",
                [updated_name, code],
            )

        updated = next(rule for rule in queries.list_compliance_rules() if rule.code == code)
        assert updated.name == updated_name
    finally:
        with connection.cursor() as cur:
            cur.execute(
                "UPDATE domain.compliance_rule SET name = %s WHERE code = %s",
                [original.name, code],
            )
