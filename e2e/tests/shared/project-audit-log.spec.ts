import { execFileSync } from 'node:child_process';
import AxeBuilder from '@axe-core/playwright';
import { expect, test } from '../../fixtures/base';
import { projectSlotForBaseUrl } from './helpers';

const ADMIN_USERNAME = 'aisha';
const ADMIN_PASSWORD = 'FieldMark!2026';
const DB_URL = 'postgres://fieldmark:fieldmark@localhost:5432/fieldmark';

type HtmxWindow = Window & {
  htmx: {
    ajax: (
      method: string,
      url: string,
      options: { target: string; swap: string; values: Record<string, string> },
    ) => Promise<void> | void;
  };
};

function runPsql(sql: string) {
  return execFileSync('psql', [DB_URL, '-c', sql], {
    encoding: 'utf8',
    stdio: 'pipe',
  });
}

function queryPsql(sql: string) {
  return execFileSync('psql', [DB_URL, '-tA', '-c', sql], {
    encoding: 'utf8',
    stdio: 'pipe',
  }).trim();
}

function resetSharedProjectsToActive() {
  runPsql(
    "update domain.project set status = 'Active' where code in ('E2EWLFO2S','P_32c6cd55','E2EWLFMTB');",
  );
}

function seedLoadMoreRows(projectId: string) {
  runPsql(`
delete from domain.audit_entry
where project_id = '${projectId}'::uuid
  and metadata ->> 'reason' like 'E2E load more %';

insert into domain.audit_entry
  (id, occurred_at, actor_id, action, entity_type, entity_id, project_id, before_state, after_state, metadata)
select
  (
    substr(md5(format('%s-audit-%s', '${projectId}', g)), 1, 8) || '-' ||
    substr(md5(format('%s-audit-%s', '${projectId}', g)), 9, 4) || '-' ||
    substr(md5(format('%s-audit-%s', '${projectId}', g)), 13, 4) || '-' ||
    substr(md5(format('%s-audit-%s', '${projectId}', g)), 17, 4) || '-' ||
    substr(md5(format('%s-audit-%s', '${projectId}', g)), 21, 12)
  )::uuid,
  timestamp with time zone '2025-01-01 00:00:00+00' + make_interval(mins => g),
  (
    substr(md5(format('%s-actor-%s', '${projectId}', g)), 1, 8) || '-' ||
    substr(md5(format('%s-actor-%s', '${projectId}', g)), 9, 4) || '-' ||
    substr(md5(format('%s-actor-%s', '${projectId}', g)), 13, 4) || '-' ||
    substr(md5(format('%s-actor-%s', '${projectId}', g)), 17, 4) || '-' ||
    substr(md5(format('%s-actor-%s', '${projectId}', g)), 21, 12)
  )::uuid,
  case when g % 2 = 0 then 'ProjectPlacedOnHold' else 'ProjectResumed' end,
  'Project',
  '${projectId}'::uuid,
  '${projectId}'::uuid,
  '{"status":"Active"}'::jsonb,
  '{"status":"OnHold"}'::jsonb,
  jsonb_build_object('reason', format('E2E load more %s', g))
from generate_series(1, 105) as g;
`);
}

function projectAuditCount(projectId: string) {
  return Number.parseInt(
    queryPsql(
      `select count(*) from domain.audit_entry where project_id = '${projectId}'::uuid;`,
    ),
    10,
  );
}

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

async function loadActiveProject(
  page: import('@playwright/test').Page,
  baseURL?: string,
): Promise<{ holdPath: string; projectId: string }> {
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
    if (!rowText.includes('Active')) {
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

    const holdPath = await page
      .locator('#place-on-hold-btn')
      .getAttribute('hx-get');
    if (!holdPath?.endsWith('/place-on-hold')) {
      continue;
    }
    const projectId = /\/projects\/([^/]+)\/place-on-hold$/.exec(holdPath)?.[1];
    if (!projectId) {
      continue;
    }

    if (fallbackIndex === null) {
      fallbackIndex = i;
    }
    if (candidateCount === targetSlot) {
      return { holdPath, projectId };
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
    const holdPath = await page
      .locator('#place-on-hold-btn')
      .getAttribute('hx-get');
    if (!holdPath?.endsWith('/place-on-hold')) {
      throw new Error('fallback project does not expose place-on-hold action');
    }
    const projectId = /\/projects\/([^/]+)\/place-on-hold$/.exec(holdPath)?.[1];
    if (!projectId) {
      throw new Error('could not derive project id from fallback hold path');
    }
    return { holdPath, projectId };
  }

  throw new Error('no Active project with place-on-hold action was found');
}

test.describe('Project audit log', () => {
  test.describe.configure({ mode: 'serial' });

  test('audit tab load more appends unique rows and expanded disclosure shows JSON', async ({
    page,
    baseURL,
  }) => {
    resetSharedProjectsToActive();
    const { projectId } = await loadActiveProject(page, baseURL);
    seedLoadMoreRows(projectId);
    const expectedTotal = projectAuditCount(projectId);
    expect(expectedTotal).toBeGreaterThan(100);

    await page.locator('#tab-audit').click();
    await expect(page.locator('#project-detail-tab-content')).toHaveAttribute(
      'aria-labelledby',
      'tab-audit',
    );
    await expect(page.locator('#audit-log')).toBeVisible();
    await expect(page.locator('#audit-log-load-more')).toBeVisible();

    const beforeCount = await page.locator('#audit-log > li.audit-row').count();
    expect(beforeCount).toBe(100);

    await page
      .locator('#audit-log > li.audit-row details summary')
      .first()
      .click();
    await expect(
      page
        .locator('#audit-log > li.audit-row details[open] .font-mono')
        .first(),
    ).toBeVisible();

    await page.locator('#audit-log-load-more button').click();

    await expect
      .poll(async () => page.locator('#audit-log > li.audit-row').count())
      .toBe(expectedTotal);

    await expect(page.locator('#audit-log-load-more')).toHaveCount(0);
  });

  test('audit tab has no new axe violations with one disclosure expanded', async ({
    page,
    baseURL,
  }) => {
    resetSharedProjectsToActive();
    const { projectId } = await loadActiveProject(page, baseURL);
    seedLoadMoreRows(projectId);

    await page.locator('#tab-audit').click();
    await expect(page.locator('#audit-log')).toBeVisible();
    await page
      .locator('#audit-log > li.audit-row details summary')
      .first()
      .click();
    await expect(
      page
        .locator('#audit-log > li.audit-row details[open] .font-mono')
        .first(),
    ).toBeVisible();

    const results = await new AxeBuilder({ page })
      .include('#project-detail-tab-content')
      .analyze();

    expect(
      results.violations,
      JSON.stringify(results.violations, null, 2),
    ).toEqual([]);
  });

  test('audit tab stays active and receives the prepended row on transition', async ({
    page,
    baseURL,
  }) => {
    resetSharedProjectsToActive();
    const { holdPath } = await loadActiveProject(page, baseURL);

    const auditTab = page.locator('#tab-audit');
    const auditPath = await auditTab.getAttribute('hx-get');
    expect(auditPath).toMatch(/\/projects\/.+\/tabs\/audit$/);

    await auditTab.click();
    await expect(page.locator('#project-detail-tab-content')).toHaveAttribute(
      'aria-labelledby',
      'tab-audit',
    );
    await expect(page.locator('#audit-log')).toBeVisible();

    const beforeRows = await page.locator('#audit-log > li.audit-row').count();
    const beforeHtml = await page.locator('#project-detail').innerHTML();

    await page.evaluate(
      async ({ holdPath }) => {
        await (window as HtmxWindow).htmx.ajax('GET', holdPath, {
          target: '#project-detail-tab-content',
          swap: 'innerHTML',
          values: {
            current_tab: 'audit',
          },
        });
      },
      { holdPath },
    );
    await expect(
      page.locator('#project-detail-tab-content textarea[name="reason"]'),
    ).toBeVisible();
    await page
      .locator('#project-detail-tab-content textarea[name="reason"]')
      .fill('Weather delay');
    await page
      .locator('#project-detail-tab-content button[type="submit"]')
      .click();

    await page.waitForFunction(
      (previous) => {
        const el = document.getElementById('project-detail');
        return !!el && el.innerHTML !== previous;
      },
      beforeHtml,
      { timeout: 10000 },
    );

    await expect(page.locator('#tab-audit')).toHaveAttribute(
      'aria-selected',
      'true',
    );
    await expect(page.locator('#audit-log')).toBeVisible();
    await expect(
      page.locator('#audit-log > li.audit-row').first(),
    ).toContainText('ProjectPlacedOnHold');
    const afterRows = await page.locator('#audit-log > li.audit-row').count();
    expect(afterRows).toBe(beforeRows + 1);
  });
});
