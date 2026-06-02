class DomainError(Exception):
    """Base typed domain error for project aggregate behavior."""


class InvalidProjectTransition(DomainError):
    """Raised when a project state transition is invalid for current status."""
