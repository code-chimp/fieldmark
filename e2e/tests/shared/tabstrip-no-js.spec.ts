/**
 * TabStrip progressive-enhancement test with JavaScript disabled (AC7 §category-3 / Story 2.7).
 *
 * Verifies that when tabstrip.js fails to load or JavaScript is disabled, the TabStrip
 * degrades gracefully: all tabs visible, layout stable, no hidden content.
 *
 * This mirrors the sidebar-no-js.spec.ts pattern from Story 1.14.
 * - component-edge-case-checklist.md §3 (JavaScript fails to initialize)
 * - fieldmark_shared/CLAUDE.md §"Sidebar progressive enhancement" (reference pattern)
 */

import { expect, test } from '@playwright/test';

const FIXTURE_URL = '/__test__/tab-strip-fixture/';

test.use({ javaScriptEnabled: false });

test.describe('TabStrip progressive enhancement (JS disabled)', () => {
  test('all tabs are visible when JS is disabled', async ({ page }) => {
    const resp = await page.goto(FIXTURE_URL);
    if (!resp || resp.status() === 404) {
      test.skip(true, 'Fixture page not available — only runs against Django with DEBUG=True');
      return;
    }
    // All four tabs must be visible
    await expect(page.locator('#tab-summary')).toBeVisible();
    await expect(page.locator('#tab-inspections')).toBeVisible();
    await expect(page.locator('#tab-violations')).toBeVisible();
    await expect(page.locator('#tab-audit')).toBeVisible();
  });

  test('no layout shift: tabs nav is present in DOM without JS', async ({ page }) => {
    const resp = await page.goto(FIXTURE_URL);
    if (!resp || resp.status() === 404) {
      test.skip(true, 'Fixture page not available');
      return;
    }
    const nav = page.locator('nav[data-tabstrip]');
    await expect(nav).toBeVisible();
    // The nav should have the tab-strip class
    await expect(nav).toHaveClass(/tab-strip/);
  });

  test('no hidden content: tab panel visible without JS', async ({ page }) => {
    const resp = await page.goto(FIXTURE_URL);
    if (!resp || resp.status() === 404) {
      test.skip(true, 'Fixture page not available');
      return;
    }
    const panel = page.locator('#project-detail-tab-content');
    await expect(panel).toBeVisible();
  });
});
