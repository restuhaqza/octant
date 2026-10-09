// Copyright (c) 2026 the Octant contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
//

import { expect, test } from '@playwright/test';
import { openApp } from './helpers';

test.describe('content datagrid', () => {
  test('renders a datagrid with rows from a pushed content message', async ({
    page,
  }) => {
    await openApp(page, 'table');

    const grid = page.locator('clr-datagrid');
    await expect(grid).toBeVisible();

    // Columns come from the pushed table view.
    await expect(grid.locator('clr-dg-column')).toHaveCount(6);
    await expect(grid).toContainText('Name');
    await expect(grid).toContainText('Ready');

    // Rows come from the pushed table view (fixture trims the real capture to 3).
    const rows = page.locator('clr-dg-row');
    await expect(rows).toHaveCount(3);
    await expect(grid).toContainText('cilium-2bjhh');
    await expect(grid).toContainText('coredns-5448979df7-d9kfk');

    // The page title is derived from the content title.
    await expect(page).toHaveTitle(/Pods/);
  });
});
