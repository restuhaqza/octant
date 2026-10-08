// Copyright (c) 2019 the Octant contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
//
import { Injectable } from '@angular/core';
import { ClarityIcons } from '@clr/angular';

export interface IconAble {
  iconName?: string;
  iconSource?: string;
}

@Injectable({
  providedIn: 'root',
})
export class IconService {
  constructor() {}

  load(item: IconAble): string {
    if (!item.iconName || item.iconName === '') {
      return '';
    }

    // Clarity 18 dropped the global `window.ClarityIcons` helpers (`has`/`add`)
    // in favour of the `ClarityIcons` class exported from `@clr/angular`
    // (`getIconShape`/`addIcons`). Using the old globals threw and left the
    // shape unregistered, so `clr-icon` fell back to the "unknown" icon.
    if (!ClarityIcons.getIconShape(item.iconName) && item.iconSource) {
      ClarityIcons.addIcons([item.iconName, item.iconSource]);
    }

    return item.iconName;
  }
}
