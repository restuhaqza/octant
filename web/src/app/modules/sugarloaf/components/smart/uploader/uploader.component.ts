// Copyright (c) 2020 the Octant contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
//

import {
  ChangeDetectionStrategy,
  Component,
  Inject,
  OnDestroy,
  OnInit,
} from '@angular/core';
import { WebsocketService } from '../../../../../data/services/websocket/websocket.service';
import { Subscription } from 'rxjs';
import { WindowToken } from '../../../../../window';

@Component({
  standalone: false,
  selector: 'app-uploader',
  templateUrl: './uploader.component.html',
  styleUrls: ['./uploader.component.scss'],
  // The upload/loading overlay is toggled by a websocket push that has no
  // DOM event or @Input to piggyback on. This is a simple root shell component
  // with no RxJS subscription lifecycle, so opt into eager checking instead of
  // threading a ChangeDetectorRef through each handler.
  changeDetection: ChangeDetectionStrategy.Eager,
})
export class UploaderComponent implements OnInit, OnDestroy {
  inputValue: string;
  showModal: boolean;

  private contentSubscription: Subscription;

  constructor(
    private websocketService: WebsocketService,
    @Inject(WindowToken) private window: Window
  ) {}

  ngOnInit(): void {
    this.websocketService.registerHandler('event.octant.dev/loading', () => {
      this.showModal = true;
    });
    this.websocketService.registerHandler('event.octant.dev/refresh', () => {
      setTimeout(this.window.location.reload.bind(this.window.location), 1000);
    });

    this.websocketService.sendMessage('action.octant.dev/loading', {
      loading: true,
    });
  }

  ngOnDestroy(): void {
    if (this.contentSubscription) {
      this.contentSubscription.unsubscribe();
    }
  }

  upload() {
    this.websocketService.sendMessage('action.octant.dev/uploadKubeConfig', {
      kubeConfig: window.btoa(this.inputValue),
    });
  }

  updateInput(event: HTMLInputElement) {
    this.inputValue = String(event);
  }

  hasInput(): boolean {
    return !this.inputValue || this.inputValue.length === 0;
  }
}
