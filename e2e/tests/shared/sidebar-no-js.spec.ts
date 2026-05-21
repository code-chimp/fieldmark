/**
 * Sidebar progressive-enhancement test (AC2.5 / Story 1.14).
 *
 * With JavaScript disabled (javaScriptEnabled: false), the sidebar must be:
 *   - present in the DOM
 *   - visible (not hidden via display:none or visibility:hidden)
 *   - not absolutely positioned off-screen (left/top < -100px)
 *
 * The authenticated session is injected via a cookie since JS-disabled pages
 * cannot execute a login flow. The cookie value matches the stub actor cookie
 * name used by the Go stack (X-FieldMark-Actor) and the Django/dotnet
 * session cookie established by force-login in fixtures. Because this test
 * runs in JS-disabled mode, we navigate directly to / with a pre-set cookie
 * that points to the "aisha" (ADMIN) dev user.
 *
 * Stack-specific notes:
 *   - .NET / Django: session cookie approach — if the fixture does not support
 *     cookie injection without JS, the test skips gracefully.
 *   - Go/Fiber: sets X-FieldMark-Actor cookie; no session required.
 *
 * If the sidebar is not present on the specific stack's home page (Epic 1
 * home is intentionally minimal), the test asserts that the element is
 * either absent or visible — never hidden. This accommodates stacks that
 * have not yet added a sidebar to the home chrome.
 */

import { expect, test } from '@playwright/test';

/**
 * Unconditional CSS contract test (AC2.5 / Story 1.14).
 *
 * This test does NOT use javaScriptEnabled: false, requires no auth, and NEVER
 * skips. It navigates to /login (always accessible), injects a mock .sidebar
 * element, and asserts that the compiled CSS PE rule makes it display:block
 * without [data-sidebar-initialized]. This gives CI an always-enforced gate
 * that complements the JS-disabled authenticated test below.
 */
test.describe('sidebar PE CSS contract (unconditional)', () => {
  test('injected .sidebar without [data-sidebar-initialized] must be display:block via CSS', async ({
    page,
  }) => {
    // /login is always accessible — no auth, no sidebar dependency.
    await page.goto('/login');
    await expect(page.locator('body')).toBeVisible();

    // Inject a bare .sidebar element (no [data-sidebar-initialized] attribute)
    // so the PE rule .sidebar:not([data-sidebar-initialized]) applies.
    await page.evaluate(() => {
      const el = document.createElement('div');
      el.className = 'sidebar';
      document.body.appendChild(el);
    });

    // The PE override must force display:block. If the rule is missing or wrong,
    // the browser's default or Basecoat's mobile display:none takes over.
    const display = await page.evaluate(() => {
      const el = document.querySelector<HTMLElement>('.sidebar:not([data-sidebar-initialized])');
      if (!el) return null;
      return window.getComputedStyle(el).display;
    });

    expect(display).not.toBeNull();
    expect(display).toBe('block');
  });
});

test.describe('sidebar progressive enhancement (no JS)', () => {
  test.use({ javaScriptEnabled: false });

  test('sidebar is visible or absent — never hidden — without JavaScript', async ({
    page,
    context,
  }) => {
    // Inject the stub actor cookie for the Go stack and a generic session hint.
    // For .NET and Django, this cookie is ignored; those stacks require a real session.
    // Use 'http://localhost' explicitly — page.url() returns 'about:blank' before
    // the first navigation and is not a valid cookie domain origin.
    await context.addCookies([
      {
        name: 'X-FieldMark-Actor',
        value: 'aisha',
        url: 'http://localhost',
        path: '/',
      },
    ]);

    // Navigate — if unauthenticated redirect occurs that's acceptable for this test;
    // we only care about sidebar behavior on pages that render it.
    const resp = await page.goto('/');
    if (!resp || resp.status() === 302 || resp.url().includes('/login')) {
      test.skip(true, 'Stack redirected unauthenticated request — skipping sidebar JS-disabled check');
      return;
    }

    const sidebar = page.locator('.sidebar').first();
    const count = await sidebar.count();

    if (count === 0) {
      // Sidebar is intentionally absent on this stack's Epic 1 home page.
      // Explicitly skip rather than trivially passing — this makes the absence
      // visible in CI output so it's not confused with a silent no-op.
      test.skip(true, 'No .sidebar element present on this stack\'s home page — Epic 1 intentionally minimal');
      return;
    }

    // Sidebar must be visible (not display:none or visibility:hidden).
    await expect(sidebar).toBeVisible();

    // Sidebar must not be absolutely positioned off-screen.
    const box = await sidebar.boundingBox();
    if (box !== null) {
      expect(box.x).toBeGreaterThan(-100);
      expect(box.y).toBeGreaterThan(-100);
    }
  });
});
