// Copyright (c) 2019 the Octant contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
//

// This file is required by karma.conf.js and loads recursively all the .spec and framework files
import 'zone.js/testing';
import { ComponentFixture, getTestBed, TestBed } from '@angular/core/testing';
import { ChangeDetectorRef } from '@angular/core';
import {
  BrowserDynamicTestingModule,
  platformBrowserDynamicTesting,
} from '@angular/platform-browser-dynamic/testing';
import { provideZoneChangeDetection } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ClarityModule } from '@clr/angular';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { RouterTestingModule } from '@angular/router/testing';

// First, initialize the Angular testing environment.
getTestBed().initTestEnvironment(
  [
    BrowserDynamicTestingModule,
    RouterTestingModule,
    CommonModule,
    FormsModule,
    ClarityModule,
    HttpClientTestingModule,
  ],
  platformBrowserDynamicTesting()
);

// Angular 21 defaults TestBed to zoneless change detection. This application
// still uses zone.js, and the existing specs drive change detection with
// fixture.detectChanges() while mutating plain (non-signal) state. Without an
// explicit zone change detection provider those mutations are not reflected and
// dev-mode reports NG0100. Provide zone change detection for every spec.
beforeEach(() => {
  TestBed.configureTestingModule({
    providers: [provideZoneChangeDetection()],
  });
});

// Angular 22 regression workaround: ComponentFixture.changeDetectorRef still
// points at the fixture root view (as in Angular 21), but unlike 21, refreshing
// it no longer descends into the component under test, so fixture.detectChanges()
// stops re-evaluating the component template after the first render. Mark the
// component under test dirty before running the original detectChanges(), which
// restores the Angular <=21 behaviour (the root view refresh descends into the
// now-dirty component view) while keeping NgZone, effect flush and checkNoChanges
// intact.
const fixturePrototype = ComponentFixture.prototype as any;
const originalDetectChanges = fixturePrototype.detectChanges;
fixturePrototype.detectChanges = function detectChanges(
  checkNoChanges = true
): void {
  try {
    this.debugElement?.injector?.get(ChangeDetectorRef)?.markForCheck();
  } catch {
    // No component ChangeDetectorRef available; use the default path only.
  }
  originalDetectChanges.call(this, checkNoChanges);
};
