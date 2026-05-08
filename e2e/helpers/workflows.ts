/**
 * Multi-step user workflows (login → action → assert visible state).
 * Expand as shared product flows stabilize.
 */

import type { Page } from '@playwright/test';

/** Placeholder for future violation-resolution workflow. */
export async function noopPlaceholder(_page: Page): Promise<void> {
  await Promise.resolve();
}
