# fieldmark_shared/components

This directory holds canonical static HTML examples for FieldMark UI components, per the UX spec "Canonical examples in `fieldmark_shared/components/`" rule (UX-DR10).

Each `.example.html` file contains variant blocks delimited by `<!-- variant: <name> -->` comments. Every per-stack template wrapper is snapshot-tested against the corresponding variant block here — a failing snapshot means the stack has drifted from the cross-stack contract.

The build script in `fieldmark_shared/package.json` writes only to `dist/fieldmark.css`; this directory requires no build step.

## Per-component directories

Story 2.4 introduces a directory convention for components that need more than one artifact:

- `status_badge/canonical.html` and `status_badge/README.md`
- `inline_alert/canonical.html` and `inline_alert/README.md`
- `audit_row/canonical.html` and `audit_row/README.md`
- `dashboard_tile/canonical.html` and `dashboard_tile/README.md`

The existing flat `*.example.html` form remains valid for previously shipped examples such as `action_button.example.html`, `login-form.example.html`, and `login-error-region.example.html`. Future component stories may use either form at the story author's discretion; the flat form is not deprecated.
