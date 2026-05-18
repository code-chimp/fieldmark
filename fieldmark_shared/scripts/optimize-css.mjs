/**
 * Post-build CSS optimization using LightningCSS.
 *
 * Pass 1 — selector merging (LightningCSS):
 *   Tailwind v4 compiles each utility class into a separate rule block.
 *   When Basecoat uses multiple utilities targeting the same pseudo-selector
 *   (e.g., &:disabled, &>svg), the output has consecutive duplicate selectors.
 *   LightningCSS merges them; pass --minify to keep the output minified (prod).
 *
 *   NOTE: Some same-selector blocks (e.g., &:focus-visible, &[aria-invalid]) are
 *   intentionally kept separate because one block contains an @supports conditional
 *   that LightningCSS cannot safely merge without minification.
 *
 * Pass 2 — content: var(--tw-content) deduplication (regex, dev builds only):
 *   Tailwind v4 emits this declaration in every pseudo-element utility block.
 *   After LightningCSS merges the blocks, consecutive copies accumulate in the
 *   same rule. The last declaration is the effective one — earlier consecutive
 *   duplicates are removed. Skipped in --minify mode (single-line output).
 *   Input is normalized to LF before the regex, so mixed LF/CRLF input is safe.
 *
 * Usage: node scripts/optimize-css.mjs [--minify] [input] [output]
 *   --minify   Minify the output (for build:prod). Default: false (dev/readable).
 *   input      Path to input CSS. Default: dist/fieldmark.css
 *   output     Path for output CSS. Default: same as input (in-place)
 *
 * lightningcss is an explicit devDependency (see package.json) so it is always
 * available via a direct require — no pnpm store scanning needed.
 */

import { readFileSync, writeFileSync, mkdirSync, renameSync, unlinkSync, realpathSync, statSync } from 'node:fs';
import { createRequire } from 'node:module';
import { resolve, dirname, normalize } from 'node:path';
import { fileURLToPath } from 'node:url';

const __dir = dirname(fileURLToPath(import.meta.url));
const root = resolve(__dir, '..');

// Parse flags and positional args separately so --minify can appear anywhere.
const flags = new Set(process.argv.slice(2).filter(a => a.startsWith('--')));
const positionals = process.argv.slice(2).filter(a => !a.startsWith('--'));

const shouldMinify = flags.has('--minify');
const input  = resolve(positionals[0] ?? resolve(root, 'dist/fieldmark.css'));
const output = resolve(positionals[1] ?? input);

// Guard: output must stay within the project root (catches obvious ../traversal in args).
if (!normalize(output).startsWith(normalize(root) + '/') && normalize(output) !== normalize(root)) {
  process.stderr.write(`optimize-css: output path "${output}" must be inside the project root.\n`);
  process.exit(1);
}

// ---------------------------------------------------------------------------
// Resolve LightningCSS.
// lightningcss is listed as a devDependency in package.json, so this is a
// direct require — no fallback or store scanning needed.
// ---------------------------------------------------------------------------
const req = createRequire(import.meta.url);
let transform;
try {
  ({ transform } = req('lightningcss'));
} catch (err) {
  process.stderr.write(`optimize-css: cannot load lightningcss — run \`pnpm install\` first.\n  ${err.message}\n`);
  process.exit(1);
}

// ---------------------------------------------------------------------------
// Ensure output directory exists, then verify its real path (resolves symlinks)
// is still inside the project root.
// ---------------------------------------------------------------------------
try {
  mkdirSync(dirname(output), { recursive: true });
} catch (err) {
  process.stderr.write(`optimize-css: cannot create output directory "${dirname(output)}": ${err.message}\n`);
  process.exit(1);
}

// NOTE: There is an inherent TOCTOU window between mkdirSync and realpathSync —
// a symlink swap could occur in between. Eliminating this requires kernel-level
// O_NOFOLLOW semantics not exposed by Node's fs API. In a trusted build
// environment this risk is acceptable; the normalize guard above stops obvious
// argument-injection attacks before we even reach the filesystem.
try {
  const realOutDir = realpathSync(dirname(output));
  const realRoot   = realpathSync(root);
  if (!realOutDir.startsWith(realRoot + '/') && realOutDir !== realRoot) {
    process.stderr.write(`optimize-css: output directory "${dirname(output)}" resolves via symlink to "${realOutDir}", outside project root.\n`);
    process.exit(1);
  }
} catch (err) {
  // Log the reason realpathSync failed so it's visible in CI; the normalize
  // check above is still active as a backstop.
  process.stderr.write(`optimize-css: note — cannot verify real path of output directory: ${err.message}\n`);
}

// ---------------------------------------------------------------------------
// Read input. Check it is a regular file first for a clearer error when a
// directory or device is passed, then verify it isn't empty.
// ---------------------------------------------------------------------------
try {
  const stat = statSync(input);
  if (!stat.isFile()) {
    process.stderr.write(`optimize-css: input "${input}" is not a regular file.\n`);
    process.exit(1);
  }
} catch (err) {
  process.stderr.write(`optimize-css: cannot stat input "${input}": ${err.message}\n`);
  process.exit(1);
}

let code;
try {
  code = readFileSync(input);
} catch (err) {
  process.stderr.write(`optimize-css: cannot read input "${input}": ${err.message}\n`);
  process.exit(1);
}

if (code.length === 0) {
  process.stderr.write(`optimize-css: input "${input}" is empty (0 bytes) — Tailwind build may have failed.\n`);
  process.exit(1);
}

// ---------------------------------------------------------------------------
// Run LightningCSS selector merge (pass 1).
// errorRecovery keeps parsing past recoverable errors (e.g. Tailwind-specific
// syntax). All warnings are written to stderr so CI logs surface them.
// ---------------------------------------------------------------------------
let css;
try {
  const result = transform({
    filename: 'fieldmark.css',
    code,
    minify: shouldMinify,
    errorRecovery: true,
  });

  if (result.warnings?.length) {
    for (const w of result.warnings) {
      process.stderr.write(`optimize-css warning [${w.type}]: ${w.message}\n`);
    }
  }

  // Normalize to LF before string operations to handle mixed or CRLF line endings.
  css = Buffer.from(result.code).toString('utf8').replace(/\r\n/g, '\n');
} catch (err) {
  process.stderr.write(`optimize-css: LightningCSS transform failed: ${err.message}\n`);
  process.exit(1);
}

// ---------------------------------------------------------------------------
// Collapse consecutive content: var(--tw-content) duplicates (pass 2, dev only).
// Minified output is single-line; the regex would never match and is skipped.
// Line endings are already normalized to LF above.
// ---------------------------------------------------------------------------
if (!shouldMinify) {
  css = css.replace(/([ \t]*content:[ \t]*var\(--tw-content\);[ \t]*\n)\1+/g, '$1');
}

// ---------------------------------------------------------------------------
// Write output atomically: write to .tmp then rename to final path.
// On POSIX this is atomic (readers never see a partial write).
// On Windows, antivirus/indexers can EPERM the rename; fall back to direct
// write in that case — less atomic but functional on a single-writer build.
// The .tmp file is cleaned up on any failure path.
// ---------------------------------------------------------------------------
const tmpOutput = output + '.tmp';
try {
  writeFileSync(tmpOutput, css);
  try {
    renameSync(tmpOutput, output);
  } catch (renameErr) {
    if (renameErr.code === 'EPERM' || renameErr.code === 'EACCES') {
      // Windows fallback: remove the locked .tmp, write directly.
      try { unlinkSync(tmpOutput); } catch { /* best effort */ }
      try {
        writeFileSync(output, css);
      } catch (writeErr) {
        process.stderr.write(`optimize-css: fallback direct write failed: ${writeErr.message}\n`);
        process.exit(1);
      }
    } else {
      throw renameErr;
    }
  }
} catch (err) {
  // Ensure .tmp never lingers on unexpected failures.
  try { unlinkSync(tmpOutput); } catch { /* best effort */ }
  process.stderr.write(`optimize-css: cannot write output "${output}": ${err.message}\n`);
  process.exit(1);
}

const inLines  = code.toString().split('\n').length;
const outLines = css.split('\n').length;
console.log(`optimize-css: ${inLines} lines → ${outLines} lines (${code.length} B → ${Buffer.byteLength(css)} B)`);
