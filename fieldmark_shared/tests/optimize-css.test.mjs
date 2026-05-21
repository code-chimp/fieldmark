/**
 * Smoke tests for scripts/optimize-css.mjs failure paths (AC3.10 / Story 1.14).
 *
 * Run: node --test tests/optimize-css.test.mjs
 *      (from fieldmark_shared/ directory)
 *
 * Tests:
 *   (a) Missing input file → exit 1, clear stderr message
 *   (b) Empty input file → exit 1, documented message
 *   (c) Directory passed as input → exit 1, clear stderr message
 *   (d) No lingering .tmp file after failure
 *
 * NOTE: The script guards that input/output paths are within the project root,
 * so all test paths are placed within fieldmark_shared/dist/ (the normal output dir).
 */

import { test } from 'node:test';
import assert from 'node:assert/strict';
import { spawnSync } from 'node:child_process';
import { writeFileSync, mkdirSync, existsSync, unlinkSync, rmSync } from 'node:fs';
import { resolve, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const __dir = fileURLToPath(new URL('.', import.meta.url));
const root = resolve(__dir, '..');
const scriptPath = resolve(root, 'scripts', 'optimize-css.mjs');
const distDir = resolve(root, 'dist');

function run(args) {
  return spawnSync('node', [scriptPath, ...args], {
    encoding: 'utf8',
    cwd: root,
    env: { ...process.env },
  });
}

// Ensure dist/ exists so we can place test files there.
mkdirSync(distDir, { recursive: true });

test('missing input file exits non-zero with clear message', () => {
  const missingInput = join(distDir, '_test_missing_input.css');
  // Ensure the file genuinely does not exist.
  if (existsSync(missingInput)) unlinkSync(missingInput);

  const result = run([missingInput]);
  assert.notEqual(result.status, 0, 'must exit non-zero for missing input');
  assert.ok(
    result.stderr.includes('cannot stat input') || result.stderr.includes('cannot read input'),
    `expected error about missing input, got: ${result.stderr}`
  );
});

test('directory passed as input exits non-zero with clear message', () => {
  // distDir itself is a directory — pass it as the input path.
  const result = run([distDir]);
  assert.notEqual(result.status, 0, 'must exit non-zero when input is a directory');
  assert.ok(
    result.stderr.includes('not a regular file'),
    `expected "not a regular file" message, got: ${result.stderr}`
  );
});

test('empty input file exits non-zero with documented message', () => {
  const emptyFile = join(distDir, '_test_empty_input.css');
  try {
    writeFileSync(emptyFile, '');
    const result = run([emptyFile]);
    assert.notEqual(result.status, 0, 'must exit non-zero for empty input');
    assert.ok(
      result.stderr.includes('empty') || result.stderr.includes('0 bytes'),
      `expected empty-file error, got: ${result.stderr}`
    );
  } finally {
    if (existsSync(emptyFile)) unlinkSync(emptyFile);
  }
});

test('no lingering .tmp file after failure on empty input', () => {
  const inputFile = join(distDir, '_test_empty_for_tmp.css');
  const outputFile = join(distDir, '_test_output_for_tmp.css');
  const tmpOutput = outputFile + '.tmp';
  try {
    writeFileSync(inputFile, '');
    run([inputFile, outputFile]);
    assert.ok(
      !existsSync(tmpOutput),
      `.tmp file must not linger after failure: ${tmpOutput} should not exist`
    );
  } finally {
    if (existsSync(inputFile)) unlinkSync(inputFile);
    if (existsSync(outputFile)) unlinkSync(outputFile);
    if (existsSync(tmpOutput)) unlinkSync(tmpOutput);
  }
});
