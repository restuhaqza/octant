// Copyright (c) 2026 the Octant contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
//

import { expect, Page } from '@playwright/test';

/** Cookie the mock backend reads to pick the pushed fixture scenario. */
export const SCENARIO_COOKIE = 'octant-e2e-scenario';

/**
 * Seed the scenario cookie for this browser context and open the app.
 *
 * The mock backend serves the built app at `/` and pushes a deterministic
 * fixture sequence over the websocket on connect. The scenario is carried by a
 * cookie so specs can run in parallel without shared server state.
 */
export async function openApp(page: Page, scenario: string): Promise<void> {
  await page.context().addCookies([
    {
      name: SCENARIO_COOKIE,
      value: scenario,
      domain: '127.0.0.1',
      path: '/',
    },
  ]);

  await page.goto('/');

  // Shell is up once Angular has rendered the container header.
  await expect(page.locator('header.header')).toBeVisible();
}
