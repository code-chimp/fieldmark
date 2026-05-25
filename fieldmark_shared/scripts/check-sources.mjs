/**
 * Pre-build sanity check: verify every @source glob in src/fieldmark.css resolves
 * to at least one file. Exits non-zero with an actionable error if any glob is empty.
 *
 * Usage: node scripts/check-sources.mjs
 *
 * Wired as the first step in the "prebuild" script in package.json so it runs
 * automatically before every `pnpm run build` or `pnpm run build:prod`.
 *
 * Failure demo (deliberately-broken glob):
 *   1. Rename any matched directory (e.g. mv FieldMark FieldMark_bak).
 *   2. Run `pnpm run build` — exits 1 with:
 *      check-sources: glob "@source \"../../FieldMark/FieldMark.Web/Pages/**\/*.cshtml\""
 *        → resolved to 0 files. Glob may be stale after a directory rename or move.
 *   3. Rename back — build proceeds normally.
 */

import { readFileSync } from 'node:fs';
import { glob } from 'node:fs/promises';
import { resolve, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';

// Hard runtime guard — `glob` from node:fs/promises requires Node >=22.
// The `engines` field in package.json is advisory; this check is enforced.
const nodeMajor = parseInt(process.versions.node.split('.')[0], 10);
if (nodeMajor < 22) {
  process.stderr.write(
    `check-sources: requires Node >=22 (found ${process.version}).\n` +
    `  Upgrade Node or run \`pnpm run build:raw\` to skip the source check.\n`
  );
  process.exit(1);
}

const __dir = dirname(fileURLToPath(import.meta.url));
const root = resolve(__dir, '..');

// Allow --css-path=<path> override for automated testing of the failure path.
const cssFlagArg = process.argv.find(a => a.startsWith('--css-path='));
const cssPath = cssFlagArg
  ? resolve(cssFlagArg.slice('--css-path='.length))
  : resolve(root, 'src/fieldmark.css');

let source;
try {
  source = readFileSync(cssPath, 'utf8');
} catch (err) {
  process.stderr.write(`check-sources: cannot read ${cssPath}: ${err.message}\n`);
  process.exit(1);
}

// Extract every @source "..." pattern from the CSS file.
const sourceRe = /@source\s+"([^"]+)"/g;
const globs = [];
let match;
while ((match = sourceRe.exec(source)) !== null) {
  globs.push(match[1]);
}

if (globs.length === 0) {
  process.stderr.write('check-sources: no @source directives found in src/fieldmark.css — expected at least one.\n');
  process.exit(1);
}

const cssDir = dirname(cssPath);
let failed = false;

for (const pattern of globs) {
  const absPattern = resolve(cssDir, pattern);
  const files = [];
  try {
    for await (const f of glob(absPattern)) {
      files.push(f);
      break; // one match is enough — we only need to know it's non-empty
    }
  } catch (err) {
    process.stderr.write(
      `check-sources: glob "${pattern}" threw an error: ${err.message}\n` +
      `  Glob may reference a path that does not exist.\n`
    );
    failed = true;
    continue;
  }
  if (files.length === 0) {
    process.stderr.write(
      `check-sources: glob "${pattern}" resolved to 0 files.\n` +
      `  Glob may be stale after a directory rename or move.\n` +
      `  Update src/fieldmark.css @source directives to match the current layout.\n`
    );
    failed = true;
  }
}

if (failed) {
  process.exit(1);
}

console.log(`check-sources: all ${globs.length} @source glob(s) resolve to files — OK`);
