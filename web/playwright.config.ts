// Copyright (c) 2026 the Octant contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
//

import { defineConfig, devices } from '@playwright/test';

/**
 * End-to-end harness for the production web build.
 *
 * `webServer` boots the mock Octant backend (e2e/mock-backend.mjs), which
 * serves web/dist/octant and pushes deterministic fixtures over
 * /api/v1/stream. No real cluster or network access is required.
 *
 * Run `npm run build` first so dist/octant exists, then:
 *   npx playwright test
 *
 * Set PLAYWRIGHT_CHROMIUM_EXECUTABLE to use a system Chrome instead of the
 * Playwright-downloaded browser (needed in sandboxes where the download is
 * blocked), e.g. PLAYWRIGHT_CHROMIUM_EXECUTABLE=/home/agent/chrome.sh.
 */

const PORT = Number(process.env.OCTANT_E2E_PORT || 4321);
const BASE_URL = `http://127.0.0.1:${PORT}`;
const executablePath = process.env.PLAYWRIGHT_CHROMIUM_EXECUTABLE;

export default defineConfig({
  testDir: './e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  timeout: 30_000,
  expect: { timeout: 10_000 },
  reporter: process.env.CI
    ? [['github'], ['list']]
    : [['list']],
  use: {
    baseURL: BASE_URL,
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
    ...(executablePath
      ? { launchOptions: { executablePath, args: ['--no-sandbox', '--disable-dev-shm-usage'] } }
      : {}),
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
  webServer: {
    command: `node e2e/mock-backend.mjs --port ${PORT}`,
    url: `${BASE_URL}/healthz`,
    reuseExistingServer: !process.env.CI,
    timeout: 120_000,
    stdout: 'pipe',
    stderr: 'pipe',
  },
});
