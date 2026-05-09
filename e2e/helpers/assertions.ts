/**
 * Shared assertion helpers for visible domain-driven state (not framework internals).
 */

import { expect, type Page } from '@playwright/test';

export async function expectOkNavigation(
  page: Page,
  path: string,
): Promise<void> {
  const response = await page.goto(path);
  expect(response, `navigation to ${path}`).toBeTruthy();
  expect(response?.status(), `HTTP status for ${path}`).toBeLessThan(400);
}
