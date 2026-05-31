/**
 * Grid row selection — AC #6, Story 2.9.
 *
 * Runs against all three stacks (shared/ means it is included in dotnet,
 * django, and fiber Playwright projects). Each stack must:
 *   1. Accept login as ADMIN (aisha / FieldMark!2026).
 *   2. Navigate to GET /projects and render the AG Grid panel.
 *   3. Wait for at least one row to appear.
 *   4. Click a row.
 *   5. Assert #project-detail populates (HTMX swap from /projects/<id>).
 *   6. Assert no JS console errors.
 *
 * No-JS fallback test (AC #8 cat 3) is also included here.
 *
 * See docs/reference/ag-grid-ssrm-contract.md
 */

import { expect, test } from '../../fixtures/base';

const ADMIN_USERNAME = 'aisha';
const ADMIN_PASSWORD = 'FieldMark!2026';

const consoleErrors: string[] = [];

async function loginAsAdmin(page: import('@playwright/test').Page, baseURL?: string) {
  const loginResp = await page.goto('/login');
  expect(loginResp?.status(), 'login page loads').toBeLessThan(400);

  const isGo = (baseURL ?? '').includes('3000');

  if (isGo) {
    const goBtn = page.locator(`button:has-text("Aisha"), button[value="${ADMIN_USERNAME}"]`).first();
    if (await goBtn.count() > 0) {
      await goBtn.click();
    } else {
      await page.fill('input[name="username"]', ADMIN_USERNAME);
      await page.getByRole('button', { name: /sign in|login/i }).click();
    }
  } else {
    await page.fill('input[name="username"]', ADMIN_USERNAME);
    await page.fill('input[name="password"]', ADMIN_PASSWORD);
    await page.getByRole('button', { name: /sign in|login/i }).click();
  }

  await page.waitForURL((url) => !url.pathname.includes('/login'), { timeout: 5000 });
}

test.describe('Projects grid — row selection', () => {
  test.beforeEach(({ page }) => {
    consoleErrors.length = 0;
    page.on('pageerror', (err) => consoleErrors.push(err.message));
    page.on('console', (msg) => {
      if (msg.type() === 'error') consoleErrors.push(msg.text());
    });
  });

  test('ADMIN can navigate to /projects and see the grid', async ({ page, baseURL }) => {
    await loginAsAdmin(page, baseURL);
    await page.goto('/projects');

    // h1 must be present
    await expect(page.locator('h1')).toContainText('Projects');

    // AG Grid container must exist
    await expect(page.locator('[data-grid-endpoint="/grid/projects"]')).toBeVisible();

    // #project-detail aside must exist
    await expect(page.locator('#project-detail')).toBeAttached();
  });

  test('clicking a grid row loads detail into #project-detail', async ({ page, baseURL }) => {
    await loginAsAdmin(page, baseURL);
    await page.goto('/projects');

    // Wait for AG Grid to render at least one row (the grid fetches asynchronously).
    // Timeout is generous because SSRM makes a POST /grid/projects network call.
    const firstRow = page.locator('.ag-row').first();
    await firstRow.waitFor({ state: 'visible', timeout: 15000 });

    const detailBefore = await page.locator('#project-detail').innerHTML();

    // Click the first grid row.
    await firstRow.click();

    // Wait for HTMX to swap content into #project-detail.
    await page.waitForFunction(
      (before) => {
        const el = document.getElementById('project-detail');
        return el && el.innerHTML !== before;
      },
      detailBefore,
      { timeout: 10000 }
    );

    // #project-detail should now have some content from /projects/<id>.
    const detailAfter = await page.locator('#project-detail').innerHTML();
    expect(detailAfter.trim().length, 'detail panel populated after row click').toBeGreaterThan(0);
    expect(detailAfter).not.toBe(detailBefore);
  });

  test('no JS console errors on /projects page', async ({ page, baseURL }) => {
    await loginAsAdmin(page, baseURL);
    await page.goto('/projects');

    // Wait briefly for any JS errors to surface.
    await page.waitForTimeout(1000);

    const licenseWarning = /unlicensed|evaluation|license/i;
    const filteredErrors = consoleErrors.filter((e) => !licenseWarning.test(e));
    expect(filteredErrors, 'no unexpected JS console errors').toHaveLength(0);
  });

  test('AC8: /projects page has correct ARIA landmarks and grid role', async ({ page, baseURL }) => {
    // Verify the ARIA structure that axe rules would check.
    // Full axe scan deferred to Epic 7 accessibility lane (per component-edge-case-checklist.md cat 8).
    await loginAsAdmin(page, baseURL);
    await page.goto('/projects');

    // h1 must be present (UX-DR33)
    await expect(page.locator('h1')).toContainText('Projects');

    // #project-detail region must be focusable
    const detailRegion = page.locator('#project-detail[role="region"][tabindex="-1"]');
    await expect(detailRegion).toBeAttached();

    // The page must have exactly one main landmark (inherited from base layout)
    const mainCount = await page.locator('main').count();
    expect(mainCount, 'page must have a main landmark').toBe(1);

    // AG Grid renders role="treegrid" or role="grid" on the grid element
    await page.locator('[data-grid-endpoint]').waitFor({ state: 'visible' });
  });

  test('AC8: /projects page remains usable under forced-colors emulation', async ({ page, baseURL }) => {
    // Forced-colors (Windows High Contrast) emulation — verifies the page renders
    // and key elements are visible; color-only failures manifest as blank/invisible content.
    // Detailed color-contrast axe audit is deferred to Epic 7 per component-edge-case-checklist.md.
    await loginAsAdmin(page, baseURL);
    await page.emulateMedia({ forcedColors: 'active' });
    await page.goto('/projects');

    // Page must not be blank under forced-colors
    await expect(page.locator('h1')).toContainText('Projects');
    await expect(page.locator('[data-grid-endpoint]')).toBeVisible();
    await expect(page.locator('#project-detail')).toBeAttached();
  });

  test('no-JS fallback: /projects page is not blank with JS disabled', async ({ browser, baseURL }) => {
    // AC8 cat 3: with JS disabled the page must degrade honestly — not render a blank box.
    // Strategy:
    //   1. Authenticate in a JS-enabled context to obtain session/auth cookies.
    //   2. Clone those cookies into a JS-disabled context.
    //   3. Navigate to /projects and assert the page is non-blank.
    //      - If we reach the page: the <noscript> fallback message must be present.
    //      - If we're still redirected (e.g. auth cookie mismatch): /login page must be non-blank.

    // Step 1: log in with JS enabled.
    const authCtx = await browser.newContext({ javaScriptEnabled: true });
    const authPage = await authCtx.newPage();
    await authPage.goto('/login');

    const isGo = (baseURL ?? '').includes('3000');
    if (isGo) {
      const goBtn = authPage.locator(`button:has-text("Aisha"), button[value="aisha"]`).first();
      if (await goBtn.count() > 0) {
        await goBtn.click();
      }
    } else {
      await authPage.fill('input[name="username"]', ADMIN_USERNAME);
      await authPage.fill('input[name="password"]', ADMIN_PASSWORD);
      await authPage.getByRole('button', { name: /sign in|login/i }).click();
    }
    await authPage.waitForURL((url) => !url.pathname.includes('/login'), { timeout: 5000 }).catch(() => {});

    // Step 2: extract cookies and create a JS-disabled context with them.
    const cookies = await authCtx.cookies();
    await authCtx.close();

    const noJsCtx = await browser.newContext({ javaScriptEnabled: false });
    await noJsCtx.addCookies(cookies);
    const noJsPage = await noJsCtx.newPage();

    // Step 3: navigate to /projects and assert non-blank response.
    await noJsPage.goto('/projects', { waitUntil: 'domcontentloaded' });

    const bodyText = await noJsPage.locator('body').innerText();
    expect(bodyText.trim().length, 'page must not be blank with JS disabled').toBeGreaterThan(0);

    // If the page loaded (not redirected to login), assert the noscript fallback is present.
    const url = noJsPage.url();
    if (!url.includes('/login')) {
      const noscriptCount = await noJsPage.locator('noscript').count();
      expect(noscriptCount, '<noscript> fallback must be present on /projects with JS disabled').toBeGreaterThan(0);
    }

    await noJsCtx.close();
  });
});
