// Copyright (c) 2019 the Octant contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
//

import { ChangeDetectorRef, Component, OnDestroy, OnInit } from '@angular/core';
import { Subscription } from 'rxjs';
import { Alert } from 'src/app/modules/shared/models/content';
import {
  NotifierService,
  NotifierSignal,
  NotifierSignalType,
} from 'src/app/modules/shared/notifier/notifier.service';

export interface NotifierToast {
  id: string;
  alert: Alert;
  time: string;
}

const alertStatus = (type: NotifierSignalType): string => {
  switch (type) {
    case NotifierSignalType.ERROR:
      return 'danger';
    case NotifierSignalType.WARNING:
      return 'warning';
    case NotifierSignalType.SUCCESS:
      return 'success';
    default:
      return 'info';
  }
};

@Component({
  standalone: false,
  selector: 'app-notifier',
  templateUrl: './notifier.component.html',
  styleUrls: ['./notifier.component.scss'],
})
export class NotifierComponent implements OnInit, OnDestroy {
  loading = false;
  toasts: NotifierToast[] = [];
  history: NotifierToast[] = [];
  historyOpen = false;

  private signalsSubscription: Subscription;
  private historySubscription: Subscription;

  constructor(
    private notifierService: NotifierService,
    private cdr: ChangeDetectorRef
  ) {}

  ngOnInit() {
    this.signalsSubscription =
      this.notifierService.globalSignalsStream.subscribe(signals => {
        this.loading = signals.some(
          signal => signal.type === NotifierSignalType.LOADING
        );
        this.toasts = signals
          .filter(signal => signal.type !== NotifierSignalType.LOADING)
          .map(signal => this.toToast(signal));
        this.cdr.markForCheck();
      });

    this.historySubscription = this.notifierService.history.subscribe(
      signals => {
        this.history = signals.map(signal => this.toToast(signal));
        this.cdr.markForCheck();
      }
    );
  }

  ngOnDestroy() {
    this.signalsSubscription?.unsubscribe();
    this.historySubscription?.unsubscribe();
  }

  dismiss(toast: NotifierToast): void {
    this.notifierService.removeSignalGlobally(toast.id);
  }

  dismissAll(): void {
    this.toasts.forEach(toast =>
      this.notifierService.removeSignalGlobally(toast.id)
    );
  }

  toggleHistory(): void {
    this.historyOpen = !this.historyOpen;
  }

  clearHistory(): void {
    this.notifierService.clearHistory();
  }

  identifyToast = (_index: number, toast: NotifierToast): string => toast.id;

  private toToast(signal: NotifierSignal): NotifierToast {
    return {
      id: signal.id,
      time: new Date(signal.timestamp).toLocaleTimeString(),
      alert: {
        status: alertStatus(signal.type),
        type: 'light',
        message: typeof signal.data === 'string' ? signal.data : '',
        closable: true,
      },
    };
  }
}
