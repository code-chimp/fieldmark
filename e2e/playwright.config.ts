import 'dotenv/config';
import { defineConfig, devices } from '@playwright/test';

/**
 * FieldMark shared E2E harness — see docs/FieldMark_Project_Level_Playwright_E2E_Harness.md
 * and _bmad-output/planning-artifacts/research/playwright-e2e-philosophy.md
 *
 * Mode B (default): Playwright projects dotnet / django / fiber — same shared specs, different baseURL.
 * Mode A: run one backend only, e.g. `pnpm exec playwright test --project=fiber`.
 */

const dotnetURL = process.env.DOTNET_URL ?? 'http://localhost:4000';
const djangoURL = process.env.DJANGO_URL ?? 'http://localhost:8000';
const fiberURL = process.env.FIBER_URL ?? 'http://localhost:3000';

export default defineConfig({
  testDir: './tests',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: 'html',
  use: {
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
  },

  projects: [
    {
      name: 'dotnet',
      testMatch: ['tests/shared/**/*.spec.ts', 'tests/dotnet/**/*.spec.ts'],
      use: {
        ...devices['Desktop Chrome'],
        baseURL: dotnetURL,
      },
    },
    {
      name: 'django',
      testMatch: ['tests/shared/**/*.spec.ts', 'tests/django/**/*.spec.ts'],
      use: {
        ...devices['Desktop Chrome'],
        baseURL: djangoURL,
      },
    },
    {
      name: 'fiber',
      testMatch: ['tests/shared/**/*.spec.ts', 'tests/fiber/**/*.spec.ts'],
      use: {
        ...devices['Desktop Chrome'],
        baseURL: fiberURL,
      },
    },
  ],
});
