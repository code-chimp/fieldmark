from django.shortcuts import render


def home(request):
    return render(request, "pages/home.html")


def privacy(request):
    return render(request, "pages/privacy.html")


def compliance_tile(request):
    return render(request, "fragments/compliance_tile.html")
