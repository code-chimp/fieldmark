import type { Page } from '@playwright/test';

/** Navigate using relative paths so Playwright applies project baseURL. */
export async function openDashboard(page: Page): Promise<void> {
  await page.goto('/');
}
