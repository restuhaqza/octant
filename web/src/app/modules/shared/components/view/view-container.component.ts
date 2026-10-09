/*
 * Copyright (c) 2020 the Octant contributors. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
  AfterViewInit,
  ChangeDetectionStrategy,
  Component,
  ComponentRef,
  EventEmitter,
  Inject,
  Input,
  OnDestroy,
  OnInit,
  Output,
  Type,
  ViewChild,
} from '@angular/core';
import { Subscription } from 'rxjs';
import { View } from '../../models/content';
import { ViewHostDirective } from '../../directives/view-host/view-host.directive';
import {
  ComponentMapping,
  DYNAMIC_COMPONENTS_MAPPING,
} from '../../dynamic-components';
import { MissingComponentComponent } from '../missing-component/missing-component.component';

interface Viewer {
  view: View;
  viewInit: EventEmitter<void>;
  ping: () => void;
}

@Component({
  standalone: false,
  selector: 'app-view-container',
  template: `<ng-container appView></ng-container>`,
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class ViewContainerComponent
  implements OnInit, AfterViewInit, OnDestroy
{
  @ViewChild(ViewHostDirective, { static: true }) appView: ViewHostDirective;
  private viewValue: View;
  @Input() set view(v: View) {
    if (v && v.metadata) {
      const cur = JSON.stringify(v);
      if (this.previous !== cur) {
        this.previous = cur;
        this.loadView(v);
      }
    }
    this.viewValue = v;
  }
  get view(): View {
    return this.viewValue;
  }
  @Input() enableDebug = false;
  @Output() viewInit: EventEmitter<void> = new EventEmitter<void>();

  private start: number;
  public componentRef: ComponentRef<Viewer>;
  private previous: string;
  private viewInitSub: Subscription;

  constructor(
    @Inject(DYNAMIC_COMPONENTS_MAPPING)
    private componentMappings: ComponentMapping
  ) {}

  ngOnInit(): void {
    if (this.enableDebug) {
      this.start = new Date().getTime();
    }
  }

  ngAfterViewInit() {
    if (!this.enableDebug || !this.view) {
      return;
    }

    console.log(
      `${this.view?.metadata?.type}: ${new Date().getTime() - this.start}`
    );
  }

  ngOnDestroy(): void {
    this.viewInitSub?.unsubscribe();
  }

  loadView(view: View) {
    const componentChanged =
      this.componentRef &&
      view.metadata.type !== this.componentRef.instance.view.metadata.type;

    if (!this.componentRef || componentChanged) {
      const viewType = view.metadata.type;
      let component: Type<any> = this.componentMappings[viewType];
      if (!component) {
        component = MissingComponentComponent;
      }

      // Drop the previous view's hook before it is destroyed (and before a new
      // one is created) so a type change does not leave a dangling subscriber
      // and an in-place update does not stack a fresh one.
      this.viewInitSub?.unsubscribe();
      this.viewInitSub = undefined;

      const viewContainerRef = this.appView.viewContainerRef;
      viewContainerRef.clear();

      this.componentRef = viewContainerRef.createComponent<Viewer>(component);
    }

    if (this.componentRef.componentType === MissingComponentComponent) {
      // The fallback is not a Viewer: it declares no `view` input and no
      // `viewInit` output. Keep the direct assignment (used by the type-change
      // check above) and surface the unknown type in its template.
      this.componentRef.instance.view = view;
      this.componentRef.setInput('name', view.metadata.type);
      return;
    }

    // Subscribe exactly once per created component: loadView() runs on every
    // in-place update, so subscribing unconditionally here would leak a
    // subscriber per update and make `viewInit` fire once per update.
    if (!this.viewInitSub) {
      this.viewInitSub = this.componentRef.instance.viewInit.subscribe(_ =>
        this.viewInit.emit()
      );
    }

    // setInput() marks the dynamically created (possibly OnPush) view dirty.
    // A direct `instance.view = view` assignment does not, because the
    // component is created imperatively instead of bound through a template,
    // so in-place refreshes for the same contentPath left the leaf view stale.
    this.componentRef.setInput('view', view);
  }
}
