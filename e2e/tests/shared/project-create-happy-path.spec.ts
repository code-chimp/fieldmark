/**
 * Project-create happy path — AC #10, Story 2.8.
 *
 * Runs against all three stacks (shared/ means it is included in dotnet,
 * django, and fiber Playwright projects). Each stack must:
 *   1. Accept login as ADMIN (aisha / FieldMark!2026).
 *   2. Navigate to GET /projects/new and render the form.
 *   3. Fill the canonical fields and submit.
 *   4. Redirect to /projects/<uuid>.
 *   5. Render the project name in the destination heading.
 *
 * See docs/reference/project-create-form-contract.md for the field contract.
 *
 * Trade-type UUID: a1b2c3d4-0001-0001-0001-000000000001 (Electrical — from 020_domain_seed.sql)
 */

import { expect, test } from '../../fixtures/base';

const ADMIN_USERNAME = 'aisha';
const ADMIN_PASSWORD = 'FieldMark!2026';
const ELEC_TRADE_ID = 'a1b2c3d4-0001-0001-0001-000000000001';

const consoleErrors: string[] = [];

test.describe('Project create — happy path', () => {
  test.beforeEach(({ page }) => {
    consoleErrors.length = 0;
    page.on('pageerror', (err) => consoleErrors.push(err.message));
    page.on('console', (msg) => {
      if (msg.type() === 'error') consoleErrors.push(msg.text());
    });
  });

  test('ADMIN can create a project and land on the detail page', async ({ page, baseURL }) => {
    // ── Login ──────────────────────────────────────────────────────────────
    const loginResp = await page.goto('/login');
    expect(loginResp?.status(), 'login page loads').toBeLessThan(400);

    // Django uses a real login form; .NET uses identity; Go uses a stub switcher.
    // All three have username + password fields.  Try both form shapes.
    const stackIsGo =
      (baseURL ?? '').includes('3000') ||
      (await page.locator('button[name="username"]').count()) > 0;

    if (stackIsGo) {
      // Go stub: click the button for the user by name.
      const goLoginButton = page.locator(`button:has-text("Aisha"), button[value="${ADMIN_USERNAME}"], form input[value="${ADMIN_USERNAME}"] ~ button`).first();
      if (await goLoginButton.count() > 0) {
        await goLoginButton.click();
      } else {
        // Fall back: look for a form that submits username directly.
        await page.fill('input[name="username"]', ADMIN_USERNAME);
        await page.getByRole('button', { name: /sign in|login/i }).click();
      }
    } else {
      await page.fill('input[name="username"]', ADMIN_USERNAME);
      await page.fill('input[name="password"]', ADMIN_PASSWORD);
      await page.getByRole('button', { name: /sign in|login/i }).click();
    }

    // After login, should land on /
    await page.waitForURL('**/');

    // ── Navigate to /projects/new ──────────────────────────────────────────
    const formResp = await page.goto('/projects/new');
    expect(formResp?.status(), 'form page loads').toBe(200);

    await expect(page.getByRole('heading', { name: /Create Project/i })).toBeVisible();
    await expect(page.locator('input[name="code"]')).toBeVisible();

    // ── Fill the form ──────────────────────────────────────────────────────
    const uniqueSuffix = Date.now().toString(36).toUpperCase().slice(-4);
    const projectCode = `E2E${uniqueSuffix}`;
    const projectName = `E2E Project ${uniqueSuffix}`;

    await page.fill('input[name="code"]', projectCode);
    await page.fill('input[name="name"]', projectName);
    await page.fill('input[name="start_date"]', '2026-06-01');

    // Select the Electrical trade scope option.
    // The select may be a <select multiple> — select by value.
    const tradeSelect = page.locator('select[name="trade_scope_ids"]');
    if (await tradeSelect.count() > 0) {
      await tradeSelect.selectOption(ELEC_TRADE_ID);
    }

    // ── Submit ────────────────────────────────────────────────────────────
    const [response] = await Promise.all([
      page.waitForNavigation({ waitUntil: 'load' }),
      page.getByRole('button', { name: /Create Project/i }).click(),
    ]);

    // ── Assert redirect to /projects/<uuid> ────────────────────────────────
    const finalUrl = page.url();
    expect(finalUrl, 'redirected to project detail').toMatch(
      /\/projects\/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/
    );

    // ── Assert project name in heading ─────────────────────────────────────
    await expect(page.getByRole('heading', { name: projectName })).toBeVisible();

    // ── No console errors ──────────────────────────────────────────────────
    expect(consoleErrors, 'no JS console errors').toHaveLength(0);
  });
});
