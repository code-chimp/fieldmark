## Deferred from: code review of 1-4-bootstrap-design-system-foundation-in-fieldmark-shared.md (2026-05-17)

- Reduced-motion users see abrupt sidebar/toast/tooltip transitions (prefers-reduced-motion)
- Unknown badge-* or data-score-band values render with default/neutral color only
- Font files 404 or blocked → only system fallback, no visual regression test
- Sidebar remains hidden or jumps if `[data-sidebar-initialized]` never set
- AG Grid empty/loading state not styled distinctly from Basecoat table
- Toaster accumulates unlimited toasts without height/scroll limit
- Tooltip `data-tooltip` with HTML entities or >container-xs text clips silently
- @source globs silently fail if any stack directory is renamed/moved
- Basecoat 0.3.11 (pre-1.0) class names may shift on minor upgrade
- High-contrast / forced-colors mode loses badge/score meaning (color-only)

## Deferred from: code review (rerun) of 1-4-bootstrap-design-system-foundation-in-fieldmark-shared.md (2026-05-17)

- optimize-css.mjs crashes on LightningCSS version bump or pnpm store change
- Script has zero error handling for missing input file or permission issues
- In-place mutation without backup or dry-run mode
- No logging or propagation of LightningCSS recovered errors
- Script assumes pnpm + ESM environment; breaks under npm/yarn or CommonJS
- Optimization step not mentioned in build docs or CLAUDE.md
- Future Basecoat minor release may re-introduce unmergeable duplicates