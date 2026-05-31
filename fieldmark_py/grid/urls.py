"""URL configuration for the grid app (Story 2.9).

  POST /grid/projects → grid_projects  (SSRM endpoint)
"""

from django.urls import path

from . import views

urlpatterns = [
    path("grid/projects", views.grid_projects, name="grid_projects"),
]
