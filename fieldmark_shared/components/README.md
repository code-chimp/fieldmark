# fieldmark_shared/components

This directory holds canonical static HTML examples for FieldMark UI components, per the UX spec "Canonical examples in `fieldmark_shared/components/`" rule (UX-DR10).

Each `.example.html` file contains variant blocks delimited by `<!-- variant: <name> -->` comments. Every per-stack template wrapper is snapshot-tested against the corresponding variant block here — a failing snapshot means the stack has drifted from the cross-stack contract.

The build script in `fieldmark_shared/package.json` writes only to `dist/fieldmark.css`; this directory requires no build step.
