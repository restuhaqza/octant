// Copyright (c) 2019 the Octant contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
//
import { CommonModule, Location } from '@angular/common';
import { HttpClientModule } from '@angular/common/http';
import {
  Injectable,
  NgModule,
  provideAppInitializer,
  provideZoneChangeDetection,
} from '@angular/core';
import { RouteReuseStrategy, RouterModule } from '@angular/router';
import { HomeComponent } from './components/smart/home/home.component';
import { AppRoutingModule } from './app-routing.module';
import { BrowserModule } from '@angular/platform-browser';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { highlightProvider } from './modules/shared/highlight';
import { MonacoEditorModule } from '@materia-ui/ngx-monaco-editor';
import { ComponentReuseStrategy } from './modules/shared/component-reuse.strategy';
import { windowProvider, WindowToken } from './window';
import {
  loadChartIconSet,
  loadCommerceIconSet,
  loadCoreIconSet,
  loadEssentialIconSet,
  loadMediaIconSet,
  loadMiniIconSet,
  loadSocialIconSet,
  loadTechnologyIconSet,
  loadTextEditIconSet,
  loadTravelIconSet,
} from '@clr/angular/icon';

@Injectable()
export class UnstripTrailingSlashLocation extends Location {
  public static stripTrailingSlash(url: string): string {
    return url;
  }
}

@NgModule({
  declarations: [HomeComponent],
  imports: [
    CommonModule,
    BrowserModule,
    BrowserAnimationsModule,
    HttpClientModule,
    RouterModule,
    MonacoEditorModule,
    // routing loads last
    AppRoutingModule,
  ],
  providers: [
    // Angular 20+ defaults to zoneless change detection. Octant still relies on
    // zone.js (see polyfills.ts) and its components mutate plain (non-signal)
    // state, so async updates (websocket content, timers) would otherwise not
    // trigger change detection and the UI would stay stuck on "Loading".
    provideZoneChangeDetection(),
    // Clarity 18 no longer ships global icon collections. Octant receives icon
    // shape names from the server and renders them with Clarity's ClrIcon, so
    // register every built-in collection once at startup.
    provideAppInitializer(() => {
      loadCoreIconSet();
      loadEssentialIconSet();
      loadTechnologyIconSet();
      loadMediaIconSet();
      loadChartIconSet();
      loadCommerceIconSet();
      loadMiniIconSet();
      loadSocialIconSet();
      loadTextEditIconSet();
      loadTravelIconSet();
    }),
    {
      provide: Location,
      useClass: UnstripTrailingSlashLocation,
    },
    highlightProvider(),
    { provide: RouteReuseStrategy, useClass: ComponentReuseStrategy },
    { provide: WindowToken, useFactory: windowProvider },
  ],
  bootstrap: [HomeComponent],
})
export class AppModule {}
