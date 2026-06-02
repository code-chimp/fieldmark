"""URL configuration for the projects app (Story 2.8 + 2.9).

See docs/reference/project-create-form-contract.md for the create form contract.
See docs/reference/ag-grid-ssrm-contract.md for the list page contract.

  GET  /projects          → project_list      (Story 2.9)
  GET  /projects/new      → project_create_get
  POST /projects/         → project_create_post  (GET → 405 via @require_POST)
  GET  /projects/<uuid>   → project_detail
  GET  /projects/<uuid>/tabs/{summary,inspections,violations,audit} → tab content + OOB tabstrip
"""

from django.urls import path

from . import views

urlpatterns = [
    path("projects", views.project_list, name="project_list"),
    path("projects/new", views.project_create_get, name="project_create"),
    # @require_POST on project_create_post returns 405 with Allow: POST on GET.
    path("projects/", views.project_create_post, name="project_collection"),
    path("projects/<uuid:id>", views.project_detail, name="project_detail"),
    path("projects/<uuid:id>/place-on-hold", views.project_place_on_hold, name="project_place_on_hold"),
    path("projects/<uuid:id>/resume", views.project_resume, name="project_resume"),
    path("projects/<uuid:id>/tabs/<str:tab>", views.project_tab, name="project_tab"),
]
