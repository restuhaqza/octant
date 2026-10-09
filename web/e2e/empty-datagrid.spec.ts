// Copyright (c) 2026 the Octant contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
//

import { expect, test } from '@playwright/test';
import { openApp } from './helpers';

test.describe('empty datagrid', () => {
  test('renders the empty placeholder (and no rows) when there is no emptyContent', async ({
    page,
  }) => {
    await openApp(page, 'empty');

    const grid = page.locator('clr-datagrid');
    await expect(grid).toBeVisible();

    // Columns still render, so the grid itself is healthy.
    await expect(grid.locator('clr-dg-column')).toHaveCount(3);

    // No data rows.
    await expect(page.locator('clr-dg-row')).toHaveCount(0);

    // Clarity marks the placeholder region as the empty state.
    await expect(
      page.locator('.datagrid-placeholder.datagrid-empty')
    ).toBeVisible();
  });

  // Known regression: the "No items to display." fallback no longer renders.
  // The fork replaced the `#emptyPlaceholder` template in
  // web/src/app/modules/shared/components/presentation/datagrid/datagrid.component.html
  // with an HTML comment, so `emptyContent: ''` yields an empty placeholder.
  // Un-skip (test.fixme -> test) once the fallback text is restored.
  test.fixme(
    'shows the "No items to display." fallback when emptyContent is blank',
    async ({ page }) => {
      await openApp(page, 'empty');

      await expect(
        page.locator('.datagrid-placeholder-content')
      ).toHaveText('No items to display.');
    }
  );
});
