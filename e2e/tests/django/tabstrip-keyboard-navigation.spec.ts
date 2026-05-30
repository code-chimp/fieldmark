/**
 * TabStrip keyboard-navigation test (AC6 / Story 2.7).
 *
 * Lives in tests/django/ — this spec requires the Django debug-gated fixture at
 * /__test__/tab-strip-fixture/ (gated behind DEBUG=True). Running this spec
 * against dotnet or fiber would 404; the django Playwright project guarantees
 * the fixture is reachable. See playwright.config.ts §projects for the lane config.
 *
 * UX §"No client-side tests for accessibility patterns" exception — TabStrip arrow-key
 * navigation is one of three documented exceptions.
 * - UX §"No client-side tests" exception list: ux-design-specification.md:1252
 * - UX-DR §"TabStrip" component spec: ux-design-specification.md:910–916
 * - WAI-ARIA Authoring Practices "Tabs" pattern: https://www.w3.org/WAI/ARIA/apg/patterns/tabs/
 *   (documentation reference only — not a code dependency)
 */

import { expect, test } from '@playwright/test';

const FIXTURE_URL = '/__test__/tab-strip-fixture/';
const SINGLE_TAB_URL = '/__test__/tab-strip-fixture/?variant=single-tab';

test.describe('TabStrip keyboard navigation', () => {
  let consoleErrors: string[] = [];

  test.beforeEach(async ({ page }) => {
    consoleErrors = [];
    page.on('pageerror', (err) => consoleErrors.push(err.message));
    page.on('console', (msg) => {
      if (msg.type() === 'error') consoleErrors.push(msg.text());
    });
  });

  test.afterEach(() => {
    expect(consoleErrors, 'no JS console errors').toHaveLength(0);
  });

  test('fixture page loads and first tab is focusable', async ({ page }) => {
    const resp = await page.goto(FIXTURE_URL);
    expect(
      resp?.status(),
      'TabStrip fixture must return 200 — ensure Django is running with DEBUG=True',
    ).toBe(200);
    const firstTab = page.locator('#tab-summary');
    await expect(firstTab).toBeVisible();
    await firstTab.focus();
    await expect(firstTab).toBeFocused();
  });

  test('ArrowRight moves focus forward', async ({ page }) => {
    await page.goto(FIXTURE_URL);
    await page.locator('#tab-summary').focus();
    await page.keyboard.press('ArrowRight');
    await expect(page.locator('#tab-inspections')).toBeFocused();
    await expect(page.locator('#tab-inspections')).toHaveAttribute('tabindex', '0');
    await expect(page.locator('#tab-summary')).toHaveAttribute('tabindex', '-1');
  });

  test('ArrowRight wraps from last tab to first', async ({ page }) => {
    await page.goto(FIXTURE_URL);
    await page.locator('#tab-audit').focus();
    await page.keyboard.press('ArrowRight');
    await expect(page.locator('#tab-summary')).toBeFocused();
  });

  test('ArrowLeft wraps from first tab to last', async ({ page }) => {
    await page.goto(FIXTURE_URL);
    await page.locator('#tab-summary').focus();
    await page.keyboard.press('ArrowLeft');
    await expect(page.locator('#tab-audit')).toBeFocused();
  });

  test('Home moves focus to first tab, End moves to last', async ({ page }) => {
    await page.goto(FIXTURE_URL);
    await page.locator('#tab-violations').focus();
    await page.keyboard.press('Home');
    await expect(page.locator('#tab-summary')).toBeFocused();
    await page.keyboard.press('End');
    await expect(page.locator('#tab-audit')).toBeFocused();
  });

  test('Enter activates tab and fires HTMX request', async ({ page }) => {
    await page.goto(FIXTURE_URL);
    let requestFired = false;
    await page.route('**/projects/__ID__/inspections', (route) => {
      requestFired = true;
      route.fulfill({
        status: 200,
        contentType: 'text/html',
        body: '<p id="inspections-content">Inspections content</p>',
      });
    });
    await page.locator('#tab-inspections').focus();
    await page.keyboard.press('Enter');
    await page.waitForTimeout(200);
    expect(requestFired).toBe(true);
    const panel = page.locator('#project-detail-tab-content');
    await expect(panel).toContainText('Inspections content');
  });

  test('Space activates tab and fires HTMX request', async ({ page }) => {
    await page.goto(FIXTURE_URL);
    let requestFired = false;
    await page.route('**/projects/__ID__/violations', (route) => {
      requestFired = true;
      route.fulfill({ status: 200, contentType: 'text/html', body: '<p>Violations content</p>' });
    });
    await page.locator('#tab-violations').focus();
    await page.keyboard.press('Space');
    await page.waitForTimeout(200);
    expect(requestFired).toBe(true);
  });

  test('single-tab: arrow keys are no-op without JS error', async ({ page }) => {
    await page.goto(SINGLE_TAB_URL);
    const tab = page.locator('#tab-only');
    await expect(tab).toBeVisible();
    await tab.focus();
    await page.keyboard.press('ArrowRight');
    await expect(tab).toBeFocused();
    await page.keyboard.press('ArrowLeft');
    await expect(tab).toBeFocused();
  });

  test('OOB-swap re-attachment: arrow keys work after tab-content swap', async ({ page }) => {
    await page.goto(FIXTURE_URL);
    await page.route('**/projects/__ID__/inspections', (route) => {
      route.fulfill({
        status: 200,
        contentType: 'text/html',
        body: `
          <p>Inspections panel</p>
          <nav
            aria-label="Project Detail Tabs"
            class="tab-strip"
            data-tabstrip
            hx-swap-oob="true"
            id="project-detail-tabstrip"
            role="tablist"
          >
            <button aria-controls="project-detail-tab-content" aria-selected="false" class="tab-strip__tab" hx-get="/projects/__ID__/summary" hx-swap="innerHTML" hx-target="#project-detail-tab-content" id="tab-summary" role="tab" tabindex="-1" type="button"><span class="tab-strip__label">Summary</span></button>
            <button aria-controls="project-detail-tab-content" aria-selected="true" class="tab-strip__tab" hx-get="/projects/__ID__/inspections" hx-swap="innerHTML" hx-target="#project-detail-tab-content" id="tab-inspections" role="tab" tabindex="0" type="button"><span class="tab-strip__label">Inspections</span></button>
            <button aria-controls="project-detail-tab-content" aria-selected="false" class="tab-strip__tab" hx-get="/projects/__ID__/violations" hx-swap="innerHTML" hx-target="#project-detail-tab-content" id="tab-violations" role="tab" tabindex="-1" type="button"><span class="tab-strip__label">Violations</span></button>
            <button aria-controls="project-detail-tab-content" aria-selected="false" class="tab-strip__tab" hx-get="/projects/__ID__/audit" hx-swap="innerHTML" hx-target="#project-detail-tab-content" id="tab-audit" role="tab" tabindex="-1" type="button"><span class="tab-strip__label">Audit</span></button>
          </nav>
        `,
      });
    });
    await page.locator('#tab-inspections').click();
    await page.waitForTimeout(500);
    await page.locator('#tab-inspections').focus();
    await page.keyboard.press('ArrowRight');
    await expect(page.locator('#tab-violations')).toBeFocused();
  });
});
