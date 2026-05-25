from django.contrib import admin
from django.contrib.auth.decorators import login_not_required
from django.urls import path
from django.views.generic import TemplateView

from . import views

urlpatterns = [
    path("robots.txt", login_not_required(TemplateView.as_view(template_name="robots.txt", content_type="text/plain"))),
    path(".well-known/security.txt", login_not_required(TemplateView.as_view(template_name="security.txt", content_type="text/plain"))),
    path("admin/", admin.site.urls),
    path("login", views.login_view, name="login"),
    path("logout", views.logout_view, name="logout"),
    path("", views.home, name="home"),
    path("privacy/", views.privacy, name="privacy"),
    path("fragments/compliance-tile/", views.compliance_tile, name="compliance_tile"),
    path("preferences/theme", views.set_theme, name="set_theme"),
]
