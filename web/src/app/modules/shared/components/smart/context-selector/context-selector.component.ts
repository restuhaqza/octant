// Copyright (c) 2019 the Octant contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
//

import { ChangeDetectorRef, Component, OnDestroy, OnInit } from '@angular/core';
import { ClarityIcons, clusterIcon } from '@clr/angular/icon';
import {
  ContextDescription,
  KubeContextService,
} from '../../../services/kube-context/kube-context.service';
import { Subscription } from 'rxjs';

@Component({
  standalone: false,
  selector: 'app-context-selector',
  templateUrl: './context-selector.component.html',
  styleUrls: ['./context-selector.component.scss'],
})
export class ContextSelectorComponent implements OnInit, OnDestroy {
  contexts: ContextDescription[] = [];
  selected = '';

  private subscriptions = new Subscription();

  constructor(
    private kubeContext: KubeContextService,
    private cdr: ChangeDetectorRef
  ) {
    ClarityIcons.addIcons(clusterIcon);
  }

  ngOnInit() {
    // Contexts and the selection are pushed from the websocket outside a normal
    // change-detection cycle, so the view is not repainted on its own. Mark the
    // view for check after each emission, mirroring ContentComponent.
    this.subscriptions.add(
      this.kubeContext.contexts().subscribe(contexts => {
        this.contexts = contexts;
        this.cdr.markForCheck();
      })
    );
    this.subscriptions.add(
      this.kubeContext.selected().subscribe(selected => {
        this.selected = selected;
        this.cdr.markForCheck();
      })
    );
  }

  ngOnDestroy(): void {
    this.subscriptions.unsubscribe();
  }

  contextClass(context: ContextDescription) {
    const active = this.selected === context.name ? ['active'] : [];
    return ['context-button', ...active];
  }

  selectContext(context: ContextDescription) {
    this.kubeContext.select(context);
  }

  trackByFn(index, item) {
    return index;
  }
}
