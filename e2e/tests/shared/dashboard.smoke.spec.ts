import { expect, test } from '../../fixtures/base';
import { expectOkNavigation } from '../../helpers/assertions';

test.describe('shared dashboard smoke', () => {
  test('root route responds without client error', async ({ page }) => {
    await expectOkNavigation(page, '/');
    await expect(page.locator('body')).toBeVisible();
  });
});
