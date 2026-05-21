/**
 * Font-fallback visual regression + CLS test (AC1.3 / Story 1.14).
 *
 * Blocks all .woff2 font requests to simulate a CDN/network failure or
 * local-dev 404. Asserts:
 *   1. A screenshot baseline passes within the configured tolerance (font
 *      fallback should be visually acceptable, not pixel-perfect).
 *   2. Cumulative Layout Shift (CLS) stays ≤ 0.1 — fonts are loaded with
 *      `font-display: swap`, so swap should not cause unbounded layout shift.
 *
 * Baselines are stored under __screenshots__/ (Playwright default) and are
 * committed to git. Re-baseline only when the change is intentional:
 *   pnpm exec playwright test --project=<stack> --update-snapshots font-fallback
 *
 * The baseline intentionally captures system-font rendering so the test
 * fails only on *unintended* changes, not on expected font-fallback behavior.
 */

import { expect, test } from '@playwright/test';

test.describe('font fallback resilience', () => {
  test('login page renders acceptably with fonts blocked (visual regression)', async ({
    page,
    browserName,
  }) => {
    // Block all font requests to simulate network failure.
    await page.route('**/*.woff2', (route) => route.abort());
    await page.route('**/*.woff', (route) => route.abort());

    await page.goto('/login');
    await expect(page.locator('body')).toBeVisible();

    // Visual regression: system-font fallback should be acceptable.
    // Tolerance: 0.5% pixel difference allowed for anti-aliasing variance.
    await expect(page).toHaveScreenshot(`login-font-fallback-${browserName}.png`, {
      maxDiffPixelRatio: 0.005,
    });
  });

  test('login page CLS ≤ 0.1 with fonts blocked', async ({ page, browserName }) => {
    // Inject CLS observer before navigation so we capture all shifts.
    // Note: this test blocks fonts entirely — it does NOT test font-swap CLS.
    // It verifies that a font-load failure (404/network block) does not itself
    // cause unbounded layout shift (the page must be stable without any fonts).
    await page.addInitScript(() => {
      (window as unknown as { __cls: number }).__cls = 0;
      // Browser-side feature guard: do NOT call observe() unconditionally.
      // If layout-shift is unsupported, observe() throws and corrupts the page load.
      // The test-side guard after navigation will see __cls === 0 and skip cleanly.
      const po = typeof PerformanceObserver !== 'undefined' ? PerformanceObserver : null;
      if (!po?.supportedEntryTypes?.includes('layout-shift')) return;
      const observer = new po((list) => {
        for (const entry of list.getEntries()) {
          // @ts-expect-error — hadRecentInput is on LayoutShift
          if (!entry.hadRecentInput) {
            // @ts-expect-error — value is on LayoutShift
            (window as unknown as { __cls: number }).__cls += entry.value;
          }
        }
      });
      observer.observe({ type: 'layout-shift', buffered: true });
    });

    await page.route('**/*.woff2', (route) => route.abort());
    await page.route('**/*.woff', (route) => route.abort());

    await page.goto('/login');
    await expect(page.locator('body')).toBeVisible();

    // Guard: layout-shift PerformanceObserver is Chromium-only in Playwright's
    // bundled browsers. Skip gracefully rather than asserting a 0 from a missing API.
    const hasLayoutShift = await page.evaluate(() => {
      try {
        return (PerformanceObserver as { supportedEntryTypes?: string[] }).supportedEntryTypes?.includes('layout-shift') ?? false;
      } catch {
        return false;
      }
    });
    if (!hasLayoutShift) {
      // Chromium must support layout-shift — if it doesn't, that is a CI
      // environment failure, not a browser-capability gap. Throw to surface it.
      if (browserName === 'chromium') {
        throw new Error(
          'layout-shift PerformanceObserver must be supported in Chromium — ' +
          'CLS gate enforcement requires at least one always-capable browser lane.'
        );
      }
      test.skip(true, `layout-shift PerformanceObserver not supported in ${browserName}`);
      return;
    }

    // Wait for network to go idle (all font requests, even aborted ones, settled)
    // rather than using a fixed timeout, which is fragile under CI load variance.
    await page.waitForLoadState('networkidle');

    const cls = await page.evaluate(() => (window as unknown as { __cls: number }).__cls);
    expect(cls).toBeLessThanOrEqual(0.1);
  });
});
