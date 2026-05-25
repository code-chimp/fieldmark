/**
 * Smoke tests for scripts/check-sources.mjs (AC3.9 / Story 1.14).
 *
 * Tests:
 *   (a) Happy path: real fieldmark.css → exits 0, OK message
 *   (b) Broken glob: @source pointing at a nonexistent directory → exits 1,
 *       clear stderr message — this is the failure-path coverage required by AC3.9.
 *
 * Run: node --test tests/check-sources.test.mjs
 *      (from fieldmark_shared/ directory)
 */

import { test } from 'node:test';
import assert from 'node:assert/strict';
import { spawnSync } from 'node:child_process';
import { mkdtempSync, writeFileSync, rmSync } from 'node:fs';
import { resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { tmpdir } from 'node:os';

const __dir = fileURLToPath(new URL('.', import.meta.url));
const scriptPath = resolve(__dir, '..', 'scripts', 'check-sources.mjs');

test('check-sources exits 0 when run against real fieldmark.css', () => {
  const result = spawnSync('node', [scriptPath], {
    encoding: 'utf8',
    cwd: resolve(__dir, '..'),
    env: { ...process.env },
  });
  assert.equal(result.status, 0, `unexpected failure: ${result.stderr}`);
  assert.ok(result.stdout.includes('OK'), `expected OK message, got: ${result.stdout}`);
});

test('check-sources exits non-zero when a @source glob matches zero files', () => {
  // Create a temp CSS file with a deliberately broken @source glob.
  const tmpDir = mkdtempSync(resolve(tmpdir(), 'check-sources-test-'));
  const fakeCss = resolve(tmpDir, 'fake.css');
  try {
    writeFileSync(fakeCss, '@source "this-directory-does-not-exist/**/*.html";\n');
    const result = spawnSync('node', [scriptPath, `--css-path=${fakeCss}`], {
      encoding: 'utf8',
      cwd: resolve(__dir, '..'),
      env: { ...process.env },
    });
    assert.notEqual(result.status, 0, 'must exit non-zero for broken glob');
    assert.ok(
      result.stderr.includes('0 files') || result.stderr.includes('resolved to 0'),
      `expected zero-files error message, got: ${result.stderr}`
    );
  } finally {
    rmSync(tmpDir, { recursive: true, force: true });
  }
});
