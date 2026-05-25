# Basecoat Upgrade Checklist

Basecoat CSS is pinned at an exact version in `fieldmark_shared/package.json` (currently `0.3.11`). Because Basecoat is pre-1.0, minor versions may rename classes, change compiled selector shapes, or reintroduce unmergeable duplicates that the `optimize-css.mjs` dedup pass has to handle.

## Why the pinned class smoke test exists

`fieldmark_shared/scripts/check-basecoat-classes.mjs` greps `node_modules/basecoat-css/dist/basecoat.css` for a small list of classes the FieldMark design system relies on. It runs automatically as part of `prebuild`. If Basecoat renames `.badge` to `.badge-item` (for example), the smoke test fails immediately at build time rather than silently producing a broken UI.

## Pinned classes (as of 0.3.11)

| Class | Purpose |
|---|---|
| `.btn` | Button component |
| `.badge` | Status badge base |
| `.alert` | Alert / inline-notification |
| `.field` | Form field wrapper |
| `.toaster` | Toast region container |
| `.toast` | Individual toast element |
| `.sidebar` | Sidebar navigation container |
| `.table` | Table / data-grid styling base (used by Epic 2 AG Grid feature stories) |

**Note:** `.menu` and `.menu-item` are FieldMark-custom classes defined in `fieldmark_shared/src/_components.css`, **not** Basecoat classes. Do not add them to this list.

## How to run the smoke test manually

```bash
cd fieldmark_shared
node scripts/check-basecoat-classes.mjs
```

Exits 0 with an `OK` message if all pinned classes are present. Exits non-zero listing the missing classes if any are absent.

## Upgrade procedure

Follow these steps in order when Basecoat publishes a new version:

1. **Read the CHANGELOG.** Check `node_modules/basecoat-css/CHANGELOG.md` (or the GitHub releases page) for any renamed classes, removed utilities, or structural changes to the compiled output.

2. **Diff the compiled CSS.** Compare the new `node_modules/basecoat-css/dist/basecoat.css` against the current pinned version using `git diff` or a text differ. Look specifically for:
   - Renamed or removed class selectors in the pinned list above.
   - New duplicate-selector patterns that `optimize-css.mjs` may not handle cleanly.
   - Changes to pseudo-selector shapes (`:disabled`, `::before`, `::after`) that interact with Tailwind v4's Basecoat integration.

3. **Update the version pin.** In `fieldmark_shared/package.json`, change the exact version string for `basecoat-css`. Use an exact pin — no `^` or `~`.

4. **Run the smoke test.** `node scripts/check-basecoat-classes.mjs` must exit 0. If it fails, update `fieldmark_shared/src/_components.css` to match the new class names and re-run.

5. **Run the full build.** `pnpm run build` must exit 0 (this includes the `prebuild` source check and the class smoke test).

6. **Run all three stack test suites.**
   ```bash
   cd FieldMark && dotnet build && dotnet test
   cd fieldmark_py && uv run ruff check . && uv run mypy . && uv run pytest
   cd fieldmark-go && make check
   ```

7. **Run `make parity` from the repo root.** Route inventory and `pg_indexes` must remain clean.

8. **Visual inspection.** Start each dev server and visually check: login page, home page (role badge, avatar menu), any pages that use `.btn`, `.badge`, `.alert`, or `.field`. Also check dark mode.

9. **Update `fieldmark_shared/CLAUDE.md`.** Update the version in the Pinned Dependencies table.

10. **Commit all changes together** — `package.json`, `dist/fieldmark.css`, and any `_components.css` adjustments — in a single commit referencing the Basecoat version bump.

## If the `optimize-css.mjs` output grows unexpectedly

After a Basecoat upgrade, if `dist/fieldmark.css` is significantly larger than before, it may mean new unmergeable duplicates were introduced. To investigate:

1. Run `pnpm run build:raw` to get the raw Tailwind output.
2. Diff `dist/fieldmark.css` (raw) against the optimized output.
3. If LightningCSS is not merging a new duplicate pattern, add the relevant selector to `scripts/optimize-css.mjs`'s dedup logic and document why in that file.
