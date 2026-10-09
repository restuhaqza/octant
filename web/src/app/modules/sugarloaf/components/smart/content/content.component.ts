/*
 * Copyright (c) 2020 the Octant contributors. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 *
 */
import {
  ChangeDetectionStrategy,
  ChangeDetectorRef,
  Component,
  NgZone,
  OnDestroy,
  OnInit,
  ViewChild,
} from '@angular/core';
import { Params, Router, UrlSegment } from '@angular/router';
import {
  ContentResponse,
  ExtensionView,
  View,
} from 'src/app/modules/shared/models/content';
import { IconService } from '../../../../shared/services/icon/icon.service';
import { ContentService } from '../../../../shared/services/content/content.service';
import { isEqual } from 'lodash';
import { Subscription } from 'rxjs';
import { LoadingService } from 'src/app/modules/shared/services/loading/loading.service';
import { OverlayScrollbarsComponent } from 'overlayscrollbars-ngx';
import { EventListeners, PartialOptions } from 'overlayscrollbars';

@Component({
  standalone: false,
  selector: 'app-overview',
  templateUrl: './content.component.html',
  styleUrls: ['./content.component.scss'],
  changeDetection: ChangeDetectionStrategy.Default,
})
export class ContentComponent implements OnInit, OnDestroy {
  @ViewChild('contentScrollbar', { read: OverlayScrollbarsComponent })
  contentScrollbar: OverlayScrollbarsComponent;

  hasTabs = false;
  hasReceivedContent = false;
  notFound = false;
  title: View[] = null;
  views: View[] = null;
  titleComponents: View[] = null;
  extView: ExtensionView = null;
  singleView: View = null;
  private contentSubscription: Subscription;
  private previousUrl = '';
  private defaultPath: string;
  private previousParams: Params;
  private loadingSubscription: Subscription;
  public showSpinner = false;
  currentPath = '';
  // https://github.com/KingSora/OverlayScrollbars/issues/257
  options: PartialOptions = {};

  events: EventListeners = {
    scroll: instance => {
      this.contentService.setScrollPos(
        instance.elements().scrollOffsetElement.scrollTop
      );
    },
  };

  constructor(
    private router: Router,
    private iconService: IconService,
    private contentService: ContentService,
    private loadingService: LoadingService,
    private ngZone: NgZone,
    private cdr: ChangeDetectorRef
  ) {}

  ngOnInit() {
    this.updatePath(this.router.routerState.snapshot.url);

    this.contentSubscription = this.contentService.current.subscribe(
      contentResponse => {
        // zone.js 0.16 dropped the default WebSocket patch, so websocket
        // emissions land outside the Angular zone; re-enter the zone and mark
        // the view so the update is actually rendered.
        this.ngZone.run(() => {
          this.setContent(contentResponse);
          this.cdr.markForCheck();
        });
      }
    );

    this.loadingService
      .withDelay(this.loadingService.requestComplete, 650, 1000)
      .subscribe(v => {
        this.showSpinner = v;
      });
  }

  ngOnDestroy() {
    this.resetView();
    if (this.contentSubscription) {
      this.contentSubscription.unsubscribe();
    }
    if (this.loadingSubscription) {
      this.loadingSubscription.unsubscribe();
    }
  }

  updatePath(url: string) {
    const tree = this.router.parseUrl(url);

    const primary = tree.root.children.primary;
    let segments = [];
    if (primary) {
      segments = primary.segments;
    }

    this.handlePathChange(segments, tree.queryParams, false);
  }

  private handlePathChange(
    segments: UrlSegment[],
    queryParams: Params,
    force: boolean
  ) {
    const urlPath = segments.map(u => u.path).join('/');
    this.currentPath = urlPath;
    const currentPath = urlPath || this.defaultPath;
    if (
      force ||
      (currentPath && currentPath !== this.previousUrl) ||
      !isEqual(queryParams, this.previousParams)
    ) {
      if (this.previousUrl === currentPath) {
        return;
      }

      this.previousParams = queryParams;
      this.resetView();
      this.contentService.setContentPath(currentPath, queryParams);
    }
  }

  private resetView() {
    this.title = null;
    this.views = null;
    this.titleComponents = null;
    this.notFound = false;
  }

  private setContent = (contentResponse: ContentResponse) => {
    if (
      this.currentPath.length > 0 &&
      contentResponse.currentPath !== this.currentPath
    ) {
      return; // ignore premature updates
    }

    const views = contentResponse.content.viewComponents;
    if (!views || views.length === 0) {
      // A response that targets the current path but carries no views has
      // nothing to render. Only treat it as not-found when the path is
      // explicit, so the initial empty seed (currentPath '') during startup
      // keeps showing the loading state.
      this.hasReceivedContent = false;
      this.notFound =
        !!contentResponse.currentPath &&
        contentResponse.currentPath === this.currentPath;
      return;
    }

    this.notFound = false;
    this.extView = contentResponse.content.extensionComponent;
    this.views = views;
    this.title = contentResponse.content.title;
    this.titleComponents = contentResponse.content.titleComponents;

    this.hasReceivedContent = true;
  };

  onScroll(event) {
    this.contentService.setScrollPos(event.target.scrollTop);
  }
}
