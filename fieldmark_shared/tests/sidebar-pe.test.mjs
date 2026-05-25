/**
 * CSS content assertion for sidebar progressive enhancement (AC2.5 / Story 1.14).
 *
 * The Playwright sidebar-no-JS test skips in Epic 1 because no stack's home page
 * has a sidebar yet. This test provides non-skippable coverage of the underlying
 * CSS contract by asserting the PE override rule is present in the compiled
 * dist/fieldmark.css — no browser, no auth, no server required.
 *
 * What it verifies:
 *   .sidebar:not([data-sidebar-initialized]) must have display:block and
 *   position:static overrides in the compiled output. If these rules are removed
 *   or regressed, the sidebar will be hidden/off-screen when JS fails to load.
 *
 * Run: node --test tests/sidebar-pe.test.mjs
 *      (from fieldmark_shared/ directory, after pnpm run build)
 */

import { test } from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const __dir = fileURLToPath(new URL('.', import.meta.url));
const distCss = resolve(__dir, '..', 'dist', 'fieldmark.css');

let css;
try {
  css = readFileSync(distCss, 'utf8');
} catch {
  throw new Error(
    `dist/fieldmark.css not found — run \`pnpm run build\` first.\n` +
    `  Path checked: ${distCss}`
  );
}

test('compiled CSS contains sidebar PE override: display:block when not initialized', () => {
  // The selector .sidebar:not([data-sidebar-initialized]) must be present.
  assert.ok(
    css.includes('.sidebar:not([data-sidebar-initialized])'),
    'dist/fieldmark.css must contain .sidebar:not([data-sidebar-initialized]) — ' +
    'the progressive-enhancement override was removed or not compiled'
  );
});

test('compiled CSS sidebar PE override forces display:block', () => {
  // Find the PE selector block and assert display:block is in it.
  // LightningCSS may reorder properties but must preserve display:block.
  const peIdx = css.indexOf('.sidebar:not([data-sidebar-initialized])');
  assert.ok(peIdx !== -1, 'PE selector not found in dist/fieldmark.css');

  // Grab a window of text after the selector to check the rule body.
  const ruleWindow = css.slice(peIdx, peIdx + 300);
  assert.ok(
    ruleWindow.includes('display:block') || ruleWindow.includes('display: block'),
    `PE rule block must set display:block. Got:\n  ${ruleWindow.slice(0, 200)}`
  );
});

test('compiled CSS sidebar PE override forces position:static', () => {
  const peIdx = css.indexOf('.sidebar:not([data-sidebar-initialized])');
  assert.ok(peIdx !== -1, 'PE selector not found in dist/fieldmark.css');
  const ruleWindow = css.slice(peIdx, peIdx + 300);
  assert.ok(
    ruleWindow.includes('position:static') || ruleWindow.includes('position: static'),
    `PE rule block must set position:static. Got:\n  ${ruleWindow.slice(0, 200)}`
  );
});
