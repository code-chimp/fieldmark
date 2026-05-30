/**
 * EntityRail responsive-collapse test (AC6 / Story 2.6).
 *
 * Asserts the CSS breakpoint behaviour from UX-DR §"Layout collapse rules"
 * (ux-design-specification.md §Responsive Breakpoints):
 *
 *   Desktop ≥ 1280px  → aside.entity-rail: position: sticky, top: 80px (5rem at 16px base)
 *   Tablet  1024px    → aside.entity-rail: position: static
 *   Mobile   375px    → aside.entity-rail: position: static
 *
 * The fixture page is served by the Django debug-gated view at
 * /__test__/entity-rail-fixture/ (gated behind DEBUG=True, excluded from
 * make parity via the __test__ prefix rule in dump_routes.py).
 *
 * These tests run in shared/ so they execute against the project=django run.
 * The layout rule lives in fieldmark_shared/dist/fieldmark.css (symlinked into
 * each stack) so testing one host stack covers the cross-stack invariant.
 */

import { expect, test } from '@playwright/test';

const FIXTURE_URL = '/__test__/entity-rail-fixture/';

test.describe('EntityRail responsive collapse', () => {
  test('desktop ≥1280px: aside.entity-rail is position:sticky with top:80px', async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1280, height: 800 });
    const resp = await page.goto(FIXTURE_URL);
    if (!resp || resp.status() === 404) {
      test.skip(true, 'Fixture page not available on this stack — skipping (only runs against django)');
      return;
    }

    const [position, top] = await page.evaluate(() => {
      const el = document.querySelector<HTMLElement>('aside.entity-rail');
      if (!el) return [null, null];
      const styles = window.getComputedStyle(el);
      return [styles.position, styles.top];
    });

    expect(position, 'at 1280px entity-rail must be position:sticky').toBe('sticky');
    // 5rem at 16px root font-size = 80px
    expect(top, 'at 1280px entity-rail top must be 80px (5rem)').toBe('80px');
  });

  test('tablet 1024px: aside.entity-rail is position:static (un-fixes)', async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1024, height: 768 });
    const resp = await page.goto(FIXTURE_URL);
    if (!resp || resp.status() === 404) {
      test.skip(true, 'Fixture page not available on this stack — skipping (only runs against django)');
      return;
    }

    const computed = await page.evaluate(() => {
      const el = document.querySelector<HTMLElement>('aside.entity-rail');
      const list = document.querySelector<HTMLElement>('.fixture-list');
      if (!el || !list) return null;
      const styles = window.getComputedStyle(el);
      const scrollY = window.scrollY;
      return {
        position: styles.position,
        top: styles.top,
        railPageTop: el.getBoundingClientRect().top + scrollY,
        listPageBottom: list.getBoundingClientRect().bottom + scrollY,
      };
    });

    expect(computed, 'computed styles should be present').not.toBeNull();
    expect(computed!.position, 'at 1024px entity-rail must be position:static').toBe('static');
    // AC6: top is auto/default when position:static
    expect(computed!.top, 'at 1024px entity-rail top must be auto (position:static)').toBe('auto');
    // AC6: rail stacks below the list in normal flow
    expect(
      computed!.railPageTop,
      'at 1024px rail must be positioned below the list'
    ).toBeGreaterThanOrEqual(computed!.listPageBottom);
  });

  test('mobile 375px: aside.entity-rail is position:static (stacks beneath list)', async ({
    page,
  }) => {
    await page.setViewportSize({ width: 375, height: 667 });
    const resp = await page.goto(FIXTURE_URL);
    if (!resp || resp.status() === 404) {
      test.skip(true, 'Fixture page not available on this stack — skipping (only runs against django)');
      return;
    }

    const computed = await page.evaluate(() => {
      const el = document.querySelector<HTMLElement>('aside.entity-rail');
      const list = document.querySelector<HTMLElement>('.fixture-list');
      if (!el || !list) return null;
      const styles = window.getComputedStyle(el);
      const scrollY = window.scrollY;
      return {
        position: styles.position,
        top: styles.top,
        railPageTop: el.getBoundingClientRect().top + scrollY,
        listPageBottom: list.getBoundingClientRect().bottom + scrollY,
      };
    });

    expect(computed, 'computed styles should be present').not.toBeNull();
    expect(computed!.position, 'at 375px entity-rail must be position:static').toBe('static');
    // AC6: top is auto/default when position:static
    expect(computed!.top, 'at 375px entity-rail top must be auto (position:static)').toBe('auto');
    // AC6: rail stacks below the list in normal flow
    expect(
      computed!.railPageTop,
      'at 375px rail must be positioned below the list'
    ).toBeGreaterThanOrEqual(computed!.listPageBottom);
  });
});
