// Copyright (c) 2026 the Octant contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
//

import { expect, test } from '@playwright/test';
import { openApp } from './helpers';

test.describe('app shell', () => {
  test('renders the header and vertical navigation from pushed websocket messages', async ({
    page,
  }) => {
    await openApp(page, 'table');

    // Header, from the static shell + event.octant.dev/buildInfo.
    const header = page.locator('header.header');
    await expect(header).toBeVisible();
    await expect(header.locator('.title')).toHaveText('Octant');

    // Vertical nav, driven by event.octant.dev/navigation.
    const nav = page.locator('clr-vertical-nav');
    await expect(nav).toBeVisible();

    // Module tabs come from the navigation sections.
    await expect(nav).toContainText('Namespace Overview');
    await expect(nav).toContainText('Cluster Overview');

    // The selected module's children render as nav links.
    await expect(nav).toContainText('Workloads');
    await expect(nav).toContainText('Deployments');
  });
});
