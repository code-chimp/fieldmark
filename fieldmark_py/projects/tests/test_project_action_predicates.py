from projects.models import Project, ProjectStatus


def _project_with_status(status: str) -> Project:
    project = Project()
    project.status = status
    return project


def test_action_predicates_active():
    project = _project_with_status(ProjectStatus.ACTIVE)
    assert project.can_place_on_hold() is True
    assert project.can_resume() is False
    assert project.can_close() is True


def test_action_predicates_on_hold():
    project = _project_with_status(ProjectStatus.ON_HOLD)
    assert project.can_place_on_hold() is False
    assert project.can_resume() is True
    assert project.can_close() is False


def test_action_predicates_closed():
    project = _project_with_status(ProjectStatus.CLOSED)
    assert project.can_place_on_hold() is False
    assert project.can_resume() is False
    assert project.can_close() is False
