/**
 * Project transition flow — Story 2.12.
 *
 * Shared across dotnet / django / fiber Playwright projects.
 *
 * Coverage:
 *   1. Log in as ADMIN.
 *   2. Load an existing Active project from the shared /projects grid.
 *   3. Exercise place-on-hold from the detail rail.
 *   4. Assert the detail region re-renders in place and the action trichotomy flips.
 *   5. Simulate a stale-state POST /resume from an Active project and assert the
 *      409 path swaps an inline alert into #project-detail without destroying the wrapper.
 */

import { expect, test } from '../../fixtures/base';
import { projectSlotForBaseUrl } from './helpers';

const ADMIN_USERNAME = 'aisha';
const ADMIN_PASSWORD = 'FieldMark!2026';

type HtmxWindow = Window & {
  htmx: {
    ajax: (
      method: string,
      url: string,
      options: { target: string; swap: string; values: Record<string, string> },
    ) => Promise<void> | void;
  };
};

async function loginAsAdmin(
  page: import('@playwright/test').Page,
  baseURL?: string,
) {
  const loginResp = await page.goto('/login');
  expect(loginResp?.status(), 'login page loads').toBeLessThan(400);

  const isGo = (baseURL ?? '').includes('3000');
  if (isGo) {
    const goBtn = page
      .locator(`button:has-text("Aisha"), button[value="${ADMIN_USERNAME}"]`)
      .first();
    if ((await goBtn.count()) > 0) {
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

  await page.waitForURL((url) => !url.pathname.includes('/login'), {
    timeout: 5000,
  });
}

async function loadProjectFromGrid(
  page: import('@playwright/test').Page,
  status: 'Active' | 'OnHold',
  actionButtonId: '#place-on-hold-btn' | '#resume-btn',
  actionSuffix: '/place-on-hold' | '/resume',
  baseURL?: string,
) {
  await loginAsAdmin(page, baseURL);

  const listResp = await page.goto('/projects');
  expect(listResp?.status(), 'projects list loads').toBe(200);

  const rows = page.locator('.ag-center-cols-container .ag-row');
  await expect(rows.first()).toBeVisible({ timeout: 10000 });

  const detailRegion = page.locator('#project-detail');
  const rowCount = await rows.count();
  const targetSlot = projectSlotForBaseUrl(baseURL);
  let candidateCount = 0;
  let fallbackIndex: number | null = null;
  for (let i = 0; i < Math.min(rowCount, 20); i += 1) {
    const row = rows.nth(i);
    const rowText = (await row.textContent()) ?? '';
    if (!rowText.includes(status)) {
      continue;
    }

    const before = await detailRegion.innerHTML();

    await row.click();
    await page.waitForFunction(
      (previous) => {
        const el = document.getElementById('project-detail');
        return !!el && el.innerHTML !== previous;
      },
      before,
      { timeout: 10000 },
    );

    const holdPath = await page.locator(actionButtonId).getAttribute('hx-get');
    if (!holdPath?.endsWith(actionSuffix)) {
      continue;
    }

    if (fallbackIndex === null) {
      fallbackIndex = i;
    }
    if (candidateCount === targetSlot) {
      return;
    }
    candidateCount += 1;
  }

  if (fallbackIndex !== null) {
    const before = await detailRegion.innerHTML();

    await rows.nth(fallbackIndex).click();
    await page.waitForFunction(
      (previous) => {
        const el = document.getElementById('project-detail');
        return !!el && el.innerHTML !== previous;
      },
      before,
      { timeout: 10000 },
    );
    return;
  }

  throw new Error(`no ${status} project with ${actionSuffix} action was found`);
}

function captureHiddenFormValues(page: import('@playwright/test').Page) {
  return page
    .locator('#project-action-form input[type="hidden"]')
    .evaluateAll((inputs) =>
      inputs.reduce<Record<string, string>>((values, input) => {
        const name = input.getAttribute('name');
        if (!name) {
          return values;
        }

        values[name] = (input as HTMLInputElement).value;
        return values;
      }, {}),
    );
}

test.describe('Project transitions', () => {
  test.describe.configure({ mode: 'serial' });

  test('place-on-hold re-renders detail in place and flips actions', async ({
    page,
    baseURL,
  }) => {
    await loadProjectFromGrid(
      page,
      'Active',
      '#place-on-hold-btn',
      '/place-on-hold',
      baseURL,
    );

    const detailRegion = page.locator('#project-detail');
    await expect(detailRegion).toBeAttached();
    await expect(page.locator('#place-on-hold-btn')).toHaveAttribute(
      'hx-get',
      /\/projects\/.+\/place-on-hold$/,
    );

    await page.locator('#place-on-hold-btn').click();
    await expect(
      page.locator('#project-action-form textarea[name="reason"]'),
    ).toBeVisible();
    await page
      .locator('#project-action-form textarea[name="reason"]')
      .fill('Weather delay');

    const beforeSubmit = await detailRegion.innerHTML();
    await page.locator('#project-action-form button[type="submit"]').click();
    await page.waitForFunction(
      (before) => {
        const el = document.getElementById('project-detail');
        return !!el && el.innerHTML !== before;
      },
      beforeSubmit,
      { timeout: 10000 },
    );

    await expect(detailRegion).toBeAttached();
    await expect(page.locator('#resume-btn')).toHaveAttribute(
      'hx-get',
      /\/projects\/.+\/resume$/,
    );
    await expect(page.locator('#place-on-hold-btn')).toHaveAttribute(
      'disabled',
      '',
    );
    await expect(page.locator('#project-action-form')).toBeAttached();
  });

  test('stale resume POST shows inline alert without destroying the wrapper', async ({
    page,
    baseURL,
  }) => {
    await loadProjectFromGrid(
      page,
      'OnHold',
      '#resume-btn',
      '/resume',
      baseURL,
    );

    const detailRegion = page.locator('#project-detail');
    await expect(detailRegion).toBeAttached();
    await page.locator('#resume-btn').click();
    await expect(
      page.locator('#project-action-form textarea[name="reason"]'),
    ).toBeVisible();

    const actionPath =
      (await page
        .locator('#project-action-form form')
        .getAttribute('action')) ?? '';
    expect(actionPath).toMatch(/\/projects\/.+\/resume$/);
    const hiddenValues = await captureHiddenFormValues(page);

    await page
      .locator('#project-action-form textarea[name="reason"]')
      .fill('Resuming from hold');

    const before = await detailRegion.innerHTML();
    await page.locator('#project-action-form button[type="submit"]').click();

    await page.waitForFunction(
      (previous) => {
        const el = document.getElementById('project-detail');
        return !!el && el.innerHTML !== previous;
      },
      before,
      { timeout: 10000 },
    );

    await expect(page.locator('#place-on-hold-btn')).toHaveAttribute(
      'hx-get',
      /\/projects\/.+\/place-on-hold$/,
    );

    const beforeConflict = await detailRegion.innerHTML();
    await page.evaluate(
      async ({ actionPath, hiddenValues }) => {
        await (window as HtmxWindow).htmx.ajax('POST', actionPath, {
          target: '#project-detail',
          swap: 'innerHTML',
          values: {
            ...hiddenValues,
            reason: 'stale request',
          },
        });
      },
      { actionPath, hiddenValues },
    );

    await page.waitForFunction(
      (previous) => {
        const el = document.getElementById('project-detail');
        return !!el && el.innerHTML !== previous;
      },
      beforeConflict,
      { timeout: 10000 },
    );

    await expect(detailRegion).toBeAttached();
    await expect(page.getByRole('alert')).toContainText(
      "Couldn't resume project",
    );
    await expect(page.getByRole('alert')).toContainText(
      'Project is not on hold',
    );
  });
});
