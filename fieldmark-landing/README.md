# FieldMark Landing Page

Static marketing and project-orientation page for FieldMark, the server-authoritative HTMX reference implementation. The landing page explains the thesis, compares the .NET, Django, and Go/Fiber stacks, and links visitors to the live implementations and source repository.

This is not one of the three application stacks. It does not connect to PostgreSQL, does not participate in parity checks, and does not own any FieldMark domain behavior.

## What This Page Covers

- The FieldMark thesis: server-owned application state with targeted browser interactivity
- The shared architecture: three backend stacks, one infrastructure-owned PostgreSQL `domain` schema
- The stack comparison: ASP.NET Core Razor Pages, Django Templates, and Go/Fiber
- The domain shape: projects, inspections, violations, corrective actions, scoring, and audit trails
- Links to the public repository and live stack deployments

For the implementation details behind those summaries, use the stack READMEs:

- [.NET README](../FieldMark/README.md) — Razor Pages + EF Core
- [Django README](../fieldmark_py/README.md) — Django Templates + Django ORM
- [Go README](../fieldmark-go/README.md) — Fiber + pgx
- [Shared assets README](../fieldmark_shared/README.md) — shared CSS and vendored JS used by the app stacks

## Project Structure

```
fieldmark-landing/
├── index.html                Single static page with embedded CSS and minimal JS
├── favicon.ico               Legacy favicon fallback
├── site.webmanifest          Web app manifest
└── static/
    ├── fonts/                Local Inter and JetBrains Mono webfonts
    └── img/                  Icons, Open Graph image, and stack artwork
```

The page is intentionally self-contained. CSS is embedded in `index.html`, fonts and images are local, and the only JavaScript handles theme selection and smooth anchor navigation.

## Local Preview

Serve the directory from `fieldmark-landing/` so root-relative paths such as `/static/img/favicon.svg` and `/site.webmanifest` resolve correctly:

```bash
cd fieldmark-landing
python3 -m http.server 8080
```

Open `http://localhost:8080`.

Opening `index.html` directly from the filesystem is not the preferred preview path because several asset references are root-relative.

## Deployment

The landing page can be deployed by any static host that serves `fieldmark-landing/` as the web root.

Expected production routes:

| Path | File |
|---|---|
| `/` | `index.html` |
| `/site.webmanifest` | `site.webmanifest` |
| `/favicon.ico` | `favicon.ico` |
| `/static/*` | `static/*` |

The page currently links to:

| Target | URL |
|---|---|
| Repository | `https://github.com/code-chimp/fieldmark` |
| .NET stack | `https://dotnet.fieldmark.dev` |
| Django stack | `https://django.fieldmark.dev` |
| Go/Fiber stack | `https://fiber.fieldmark.dev` |

Update `index.html` if deployment domains change.

## Editing Guidelines

- Keep the landing page static unless there is a concrete need for a build step.
- Do not add client-side state management or frontend routing.
- Keep implementation claims aligned with the root README and the three stack READMEs.
- Treat stack versions, schema ownership, parity claims, and architecture rules as factual documentation. If one changes in the app READMEs, update this page.
- Prefer local assets over CDN dependencies so the page remains portable and durable.
- Keep accessibility basics intact: semantic headings, useful link labels, image alt text, and keyboard-visible controls.

## Asset Notes

The page includes generated/project artwork plus third-party mascot references. Current footer attribution in `index.html` covers:

- Go Gopher artwork by Renee French, CC BY 4.0
- .NET dotnet-bot illustrations by the .NET Foundation, CC0 1.0
- HTMX logo, 0-Clause BSD
- Django trademark notice and community mascot inspiration

When adding or replacing artwork, update both this README and the footer attribution if the licensing context changes.

## Relationship To FieldMark

FieldMark itself is the multi-stack reference application documented at the repository root. The landing page is presentation material for that project: it should describe the architecture accurately, but it should not introduce separate architecture, runtime dependencies, or source-of-truth documentation that conflicts with the app stacks.
