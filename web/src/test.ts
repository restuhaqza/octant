// Copyright (c) 2019 the Octant contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
//

// This file is required by karma.conf.js and loads recursively all the .spec and framework files
import 'zone.js/testing';
import { getTestBed, TestBed } from '@angular/core/testing';
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
