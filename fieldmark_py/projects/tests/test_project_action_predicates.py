import pytest

from projects.models import Project, ProjectStatus
from projects.errors import InvalidProjectTransition


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


def test_place_on_hold_active_to_on_hold():
    project = _project_with_status(ProjectStatus.ACTIVE)
    project.place_on_hold("maintenance window")
    assert project.status == ProjectStatus.ON_HOLD


@pytest.mark.parametrize("status", [ProjectStatus.ON_HOLD, ProjectStatus.CLOSED])
def test_place_on_hold_invalid_states_raise(status: str):
    project = _project_with_status(status)
    with pytest.raises(InvalidProjectTransition, match="Project is already on hold"):
        project.place_on_hold("maintenance window")


def test_resume_on_hold_to_active():
    project = _project_with_status(ProjectStatus.ON_HOLD)
    project.resume("back online")
    assert project.status == ProjectStatus.ACTIVE


@pytest.mark.parametrize("status", [ProjectStatus.ACTIVE, ProjectStatus.CLOSED])
def test_resume_invalid_states_raise(status: str):
    project = _project_with_status(status)
    with pytest.raises(InvalidProjectTransition, match="Project is not on hold"):
        project.resume("back online")
