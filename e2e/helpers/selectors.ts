/**
 * Locator strategy for FieldMark E2E (semantic-first; avoid CSS-class coupling).
 *
 * Priority:
 * 1. Playwright accessibility locators — getByRole, getByLabel, getByPlaceholder, getByText — when
 *    product copy and roles are parity-locked across .NET / Django / Fiber.
 * 2. Stable HTMX / layout IDs from root CLAUDE.md — #project-detail, #compliance-tile,
 *    #violation-detail, #audit-log (must match across stacks).
 * 3. data-testid — only when markup or accessible names diverge by stack but behavior must align.
 *
 * Centralize locators here so fallback policy stays consistent.
 */

import type { Locator, Page } from '@playwright/test';

/** Example: primary action using accessible name (prefer exact copy from UX guide). */
export function primaryButton(page: Page, name: string | RegExp): Locator {
  return page.getByRole('button', { name });
}

/** HTMX fragment roots — contract IDs (see root CLAUDE.md). */
export function complianceTile(page: Page): Locator {
  return page.locator('#compliance-tile');
}

export function projectDetail(page: Page): Locator {
  return page.locator('#project-detail');
}

export function violationDetail(page: Page): Locator {
  return page.locator('#violation-detail');
}

export function auditLog(page: Page): Locator {
  return page.locator('#audit-log');
}

/** Prefer explicit test id only when semantic locators are not portable across stacks. */
export function byTestId(page: Page, id: string): Locator {
  return page.getByTestId(id);
}
