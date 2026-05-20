"""Per-request actor helpers — read-only view of the authenticated principal."""

from dataclasses import dataclass
from uuid import UUID

from django.http import HttpRequest


@dataclass(frozen=True)
class CurrentActor:
    id: UUID
    username: str
    roles: tuple[str, ...]

    @property
    def is_anonymous(self) -> bool:
        return self.username == "anonymous"


ANONYMOUS = CurrentActor(
    id=UUID("00000000-0000-0000-0000-000000000000"),
    username="anonymous",
    roles=(),
)


def current_actor(request: HttpRequest) -> CurrentActor:
    user = request.user
    if not user.is_authenticated:
        return ANONYMOUS
    # dev_uuid is the related_name on DevUserUuid (Story 1.10, tools app).
    try:
        uuid_value = user.dev_uuid.uuid
    except AttributeError as exc:
        raise RuntimeError(
            f"Authenticated user {user.username!r} has no DevUserUuid row. "
            "Run `uv run python manage.py seed_dev_users` to populate the manifest."
        ) from exc
    roles = tuple(user.groups.values_list("name", flat=True))
    return CurrentActor(id=uuid_value, username=user.username, roles=roles)
