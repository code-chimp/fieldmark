from django.conf import settings
from django.contrib import admin
from django.contrib.auth.decorators import login_not_required
from django.urls import include, path
from django.views.generic import TemplateView

from reference.views import reference_index

from . import views

urlpatterns = [
    path("robots.txt", login_not_required(TemplateView.as_view(template_name="robots.txt", content_type="text/plain"))),
    path(".well-known/security.txt", login_not_required(TemplateView.as_view(template_name="security.txt", content_type="text/plain"))),
    # Must precede Django Admin's admin/ mount; URL patterns resolve in declaration order.
    path("admin/reference", reference_index, name="reference_index"),
    path("admin/", admin.site.urls),
    path("login", views.login_view, name="login"),
    path("logout", views.logout_view, name="logout"),
    path("", views.home, name="home"),
    path("privacy/", views.privacy, name="privacy"),
    path("fragments/compliance-tile/", views.compliance_tile, name="compliance_tile"),
    path("preferences/theme", views.set_theme, name="set_theme"),
    # Projects app — Story 2.8.
    path("", include("projects.urls")),
    # Grid endpoints — Story 2.9.
    path("", include("grid.urls")),
]

if settings.DEBUG:
    urlpatterns += [
        path("__test__/entity-rail-fixture/", views.entity_rail_fixture, name="entity_rail_fixture"),
        path("__test__/tab-strip-fixture/", views.tab_strip_fixture, name="tab_strip_fixture"),
    ]
