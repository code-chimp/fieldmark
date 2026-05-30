"""URL configuration for the projects app (Story 2.8).

See docs/reference/project-create-form-contract.md for the route contract.

  GET  /projects/new      → project_create_get
  POST /projects/         → project_create_post  (GET → 405 via @require_POST)
  GET  /projects/<uuid>   → project_detail_stub  (stub; Story 2.11 replaces)
"""

from django.urls import path

from . import views

urlpatterns = [
    path("projects/new", views.project_create_get, name="project_create"),
    # @require_POST on project_create_post returns 405 with Allow: POST on GET.
    path("projects/", views.project_create_post, name="project_collection"),
    path("projects/<uuid:id>", views.project_detail_stub, name="project_detail"),
]
