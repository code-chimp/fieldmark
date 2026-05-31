from fieldmark.views import dashboard_context_from_raw


def test_dashboard_context_empty_sets_map_to_nulls():
    ctx = dashboard_context_from_raw(
        portfolio_score=None,
        project_count=0,
        active_count=0,
        violation_count=0,
        overdue_total=0,
        overdue_breakdown="",
        inspection_count=0,
        week_count=0,
    )
    assert ctx["portfolio_score"] is None
    assert ctx["active_projects"] is None
    assert ctx["overdue_violations"] is None
    assert ctx["inspections_week"] is None


def test_dashboard_context_existing_data_keeps_zero_counts():
    ctx = dashboard_context_from_raw(
        portfolio_score=88,
        project_count=5,
        active_count=0,
        violation_count=2,
        overdue_total=0,
        overdue_breakdown="",
        inspection_count=4,
        week_count=0,
    )
    assert ctx["portfolio_score"] == 88
    assert ctx["active_projects"] == 0
    assert ctx["overdue_violations"] == 0
    assert ctx["inspections_week"] == 0
