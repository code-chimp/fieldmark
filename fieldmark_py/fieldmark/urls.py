from django.contrib import admin
from django.urls import path

from . import views

urlpatterns = [
    path("admin/", admin.site.urls),
    path("login", views.login_view, name="login"),
    path("logout", views.logout_view, name="logout"),
    path("", views.home, name="home"),
    path("privacy/", views.privacy, name="privacy"),
    path("fragments/compliance-tile/", views.compliance_tile, name="compliance_tile"),
    path("preferences/theme", views.set_theme, name="set_theme"),
]
