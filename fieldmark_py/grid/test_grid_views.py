"""View-level tests for grid/views.py.

Tests that only need request/response machinery (auth, 400 validation)
run without a database by using Django's test client with a mocked user.
"""

from __future__ import annotations

import json
from unittest.mock import MagicMock, patch

import pytest
from django.test import Client, RequestFactory

from grid import views as grid_views

VALID_BODY = json.dumps({"startRow": 0, "endRow": 10, "sortModel": [], "filterModel": {}}).encode()


# ─── Auth ─────────────────────────────────────────────────────────────────────

@pytest.mark.django_db
def test_grid_projects_unauthenticated_returns_403(client: Client):
    resp = client.post("/grid/projects", data=VALID_BODY, content_type="application/json")
    # Unauthenticated: LoginRequiredMiddleware redirects to /login, or grid view returns 403.
    assert resp.status_code in (302, 403)


# ─── 400 validation (no DB needed) ───────────────────────────────────────────

def _authed_post(body: bytes | str) -> int:
    """POST /grid/projects with a mocked authenticated user that has project.read."""
    factory = RequestFactory()
    if isinstance(body, str):
        body = body.encode()
    req = factory.post("/grid/projects", data=body, content_type="application/json")

    mock_user = MagicMock()
    mock_user.is_authenticated = True

    with patch("grid.views.can", return_value=True):
        req.user = mock_user
        resp = grid_views.grid_projects(req)
    return resp.status_code


def test_grid_projects_invalid_json_returns_400():
    assert _authed_post(b"NOT JSON") == 400


def test_grid_projects_unknown_col_id_returns_400():
    body = json.dumps({"startRow": 0, "endRow": 10, "sortModel": [{"colId": "UNKNOWN", "sort": "asc"}], "filterModel": {}})
    assert _authed_post(body) == 400


def test_grid_projects_injection_col_id_returns_400():
    body = json.dumps({"startRow": 0, "endRow": 10, "sortModel": [{"colId": "code; DROP TABLE domain.project --", "sort": "asc"}], "filterModel": {}})
    assert _authed_post(body) == 400


def test_grid_projects_invalid_sort_direction_returns_400():
    body = json.dumps({"startRow": 0, "endRow": 10, "sortModel": [{"colId": "code", "sort": "INVALID"}], "filterModel": {}})
    assert _authed_post(body) == 400


def test_grid_projects_negative_start_row_returns_400():
    body = json.dumps({"startRow": -1, "endRow": 10, "sortModel": [], "filterModel": {}})
    assert _authed_post(body) == 400


def test_grid_projects_page_size_exceeds_max_returns_400():
    body = json.dumps({"startRow": 0, "endRow": 1001, "sortModel": [], "filterModel": {}})
    assert _authed_post(body) == 400


def test_grid_projects_invalid_status_value_returns_400():
    body = json.dumps({"startRow": 0, "endRow": 10, "sortModel": [], "filterModel": {"status": {"filterType": "set", "values": ["INVALID"]}}})
    assert _authed_post(body) == 400


def test_grid_projects_injection_status_value_returns_400():
    body = json.dumps({"startRow": 0, "endRow": 10, "sortModel": [], "filterModel": {"status": {"filterType": "set", "values": ["Active' OR '1'='1"]}}})
    assert _authed_post(body) == 400
