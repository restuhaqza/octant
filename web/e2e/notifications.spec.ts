// Copyright (c) 2026 the Octant contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
//

import { expect, test } from '@playwright/test';
import { openApp } from './helpers';

test.describe('notifications', () => {
  // Requires the notification center (toast stack + bell badge) added in the
  // "stack toasts and add a notification center" change. Un-skip
  // (test.fixme -> test) once that lands on master.
  test.fixme(
    'pushing an alert shows a notifier toast and increments the bell badge',
    async ({ page }) => {
      await openApp(page, 'alert');

      // Toast body, rendered from event.octant.dev/alert.
      const toastText = page.locator('.notifier-toasts .alert-text');
      await expect(toastText).toHaveCount(1);
      await expect(toastText).toHaveText(
        'e2e: this is a pushed alert notification.'
      );

      // The history badge counts non-loading signals.
      await expect(page.locator('.notifier-toggle__badge')).toHaveText('1');
    }
  );
});
