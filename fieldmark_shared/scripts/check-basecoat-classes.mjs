/**
 * Basecoat class-name smoke test: assert the CSS classes the design system
 * relies on are still present in the installed Basecoat distribution.
 *
 * Why: Basecoat is pre-1.0 and may rename classes on minor version bumps.
 * This guard catches the breakage at build time rather than in the browser.
 * See docs/basecoat-upgrade-checklist.md for the upgrade procedure.
 *
 * Usage: node scripts/check-basecoat-classes.mjs
 *
 * Wired into the "prebuild" script in package.json (runs after check-sources).
 * Also useful as a standalone gate after `pnpm update basecoat-css`.
 */

import { readFileSync } from 'node:fs';
import { resolve, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';

const __dir = dirname(fileURLToPath(import.meta.url));
const root = resolve(__dir, '..');

// The set of Basecoat class names the FieldMark design system requires.
// These are classes that come FROM Basecoat (not FieldMark custom classes).
// Update this list when adding new Basecoat classes to fieldmark CSS.
// Run `node scripts/check-basecoat-classes.mjs` after updating the list.
//
// AC4.13 reconciliation note:
//   The original AC text listed `.menu` and `.menu-item` as classes to verify.
//   Those are FieldMark-custom classes defined in `src/_components.css` — they do
//   NOT come from Basecoat's distribution and are intentionally absent here.
//   Adding them would make this check permanently fail. The class list below
//   reflects what FieldMark actually depends on FROM Basecoat, including `.table`
//   (used by Epic 2 data-grid stories) in place of the incorrectly-cited menu classes.
const REQUIRED_CLASSES = [
  '.btn',
  '.badge',
  '.alert',
  '.field',
  '.toaster',
  '.toast',
  '.sidebar',
  '.table',
];

const basecoatPath = resolve(root, 'node_modules/basecoat-css/dist/basecoat.css');
let css;
try {
  css = readFileSync(basecoatPath, 'utf8');
} catch (err) {
  process.stderr.write(
    `check-basecoat-classes: cannot read Basecoat distribution at ${basecoatPath}\n` +
    `  Run \`pnpm install\` first.\n  ${err.message}\n`
  );
  process.exit(1);
}

// Selector-context regex: the class must be preceded by start-of-line, whitespace,
// comma, or a brace — i.e., only positions where a CSS selector token can appear.
// This prevents false-passes when the class name occurs in a non-selector position
// (e.g., inside a CSS comment or a string value). The suffix guard prevents substring
// matches: `.btn` inside `.btn-primary` is excluded because `-` follows.
function classIsPresent(cls) {
  const escaped = cls.replace(/\./g, '\\.');
  return new RegExp('(?:^|[,{}\\s])' + escaped + '(?![-\\w])', 'm').test(css);
}

const missing = REQUIRED_CLASSES.filter(cls => !classIsPresent(cls));

if (missing.length > 0) {
  process.stderr.write(
    `check-basecoat-classes: the following required Basecoat classes are missing from ${basecoatPath}:\n` +
    missing.map(c => `  ${c}`).join('\n') + '\n\n' +
    `  A Basecoat upgrade may have renamed or removed these classes.\n` +
    `  See docs/basecoat-upgrade-checklist.md for the upgrade procedure.\n`
  );
  process.exit(1);
}

console.log(`check-basecoat-classes: all ${REQUIRED_CLASSES.length} required Basecoat class(es) present — OK`);
