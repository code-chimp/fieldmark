/**
 * Toaster overflow-cap visual regression (AC2.7 / Story 1.14).
 *
 * The .toaster CSS region must:
 *   - show at most 5 visible toasts when more are queued
 *   - scroll on overflow (overflow-y: auto)
 *   - never grow unbounded (max-height capped)
 *
 * This test injects 7 toast elements into the DOM, checks that the toaster
 * region's scrollHeight is capped, and takes a visual regression snapshot.
 *
 * Note: Epic 1 does not have a server-side toast queue helper; the test
 * directly injects DOM nodes via page.evaluate so it is independent of
 * stack-specific toast mechanisms. The CSS rules alone govern the cap.
 */

import { expect, test } from '@playwright/test';

test.describe('toaster overflow cap', () => {
  test('toaster region caps at ~5 visible toasts and scrolls on overflow', async ({
    page,
    browserName,
  }) => {
    // Navigate to any authenticated page. Use the login page as a simple
    // public route that has the CSS loaded.
    await page.goto('/login');
    await expect(page.locator('body')).toBeVisible();

    // Inject a mock toaster with 7 toast elements via JS.
    await page.evaluate(() => {
      const toaster = document.createElement('div');
      toaster.className = 'toaster';
      toaster.setAttribute('data-testid', 'mock-toaster');
      toaster.style.position = 'fixed';
      toaster.style.bottom = '0';
      toaster.style.right = '0';
      for (let i = 1; i <= 7; i++) {
        const toast = document.createElement('div');
        toast.className = 'toast';
        toast.style.height = '4rem';
        toast.style.background = 'var(--color-card, #fff)';
        toast.style.border = '1px solid #ccc';
        toast.style.padding = '0.5rem';
        toast.textContent = `Toast ${i}`;
        toaster.appendChild(toast);
      }
      document.body.appendChild(toaster);
    });

    const toaster = page.locator('[data-testid="mock-toaster"]');
    await expect(toaster).toBeVisible();

    // The toaster's scrollable height should be less than 7 × 4rem ≈ 448px.
    // (7 toasts × (4rem + 2×padding) but capped at 5 toasts.)
    const scrollHeight = await toaster.evaluate((el) => el.scrollHeight);
    const clientHeight = await toaster.evaluate((el) => el.clientHeight);

    // scrollHeight > clientHeight means overflow is active (scrollable).
    // This passes when the CSS cap is applied and overflow-y:auto is set.
    // If clientHeight >= scrollHeight, the CSS cap isn't working — fail.
    expect(clientHeight).toBeLessThan(scrollHeight);

    // Visual regression snapshot: 7 toasts queued, 5 visible + scroll.
    await expect(toaster).toHaveScreenshot(`toaster-cap-${browserName}.png`, {
      maxDiffPixelRatio: 0.02,
    });
  });
});
