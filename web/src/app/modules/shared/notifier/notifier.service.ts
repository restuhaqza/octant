// Copyright (c) 2019 the Octant contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
//

import { Injectable } from '@angular/core';
import findIndex from 'lodash/findIndex';
import forEach from 'lodash/forEach';
import pullAt from 'lodash/pullAt';
import remove from 'lodash/remove';
import uniqueId from 'lodash/uniqueId';
import { BehaviorSubject } from 'rxjs';

export enum NotifierSignalType {
  LOADING = 'LOADING',
  ERROR = 'ERROR',
  WARNING = 'WARNING',
  INFO = 'INFO',
  SUCCESS = 'SUCCESS',
}

export interface NotifierSignal {
  id: string;
  sessionID: string;
  type: NotifierSignalType;
  data: boolean | string;
  timestamp: number;
}

/**
 * How many past notifications the notification center keeps.
 */
export const notifierHistoryLimit = 50;

export class NotifierSession {
  id: string;

  constructor(
    private globalSignalsStream: BehaviorSubject<NotifierSignal[]>,
    private historyStream: BehaviorSubject<NotifierSignal[]>,
    private uniqueIDPrefix: string
  ) {
    this.id = uniqueIDPrefix;
  }

  pushSignal(type: NotifierSignalType, data: boolean | string): string {
    const currentSignals = this.globalSignalsStream.getValue();
    const newSignalID = uniqueId(this.uniqueIDPrefix);
    const newSignal: NotifierSignal = {
      id: newSignalID,
      sessionID: this.uniqueIDPrefix,
      type,
      data,
      timestamp: Date.now(),
    };
    this.globalSignalsStream.next([...currentSignals, newSignal]);

    // Loading is a transient, non-actionable state; only keep the notifications
    // a user may want to review later.
    if (type !== NotifierSignalType.LOADING) {
      const history = this.historyStream.getValue();
      this.historyStream.next(
        [newSignal, ...history].slice(0, notifierHistoryLimit)
      );
    }

    return newSignalID;
  }

  removeSignal(id: string): boolean {
    const currentSignals = this.globalSignalsStream.getValue();
    const foundSignalIndex = findIndex(currentSignals, {
      id,
      sessionID: this.uniqueIDPrefix,
    });
    if (foundSignalIndex < 0) {
      return false;
    }

    const newSignalList = [...currentSignals];
    pullAt(newSignalList, foundSignalIndex);
    this.globalSignalsStream.next(newSignalList);
    return true;
  }

  removeSignals(ids: string[]): void {
    forEach(ids, (id: string) => {
      if (id) {
        this.removeSignal(id);
      }
    });
  }

  removeAllSignals(): void {
    const currentSignals = this.globalSignalsStream.getValue();
    const newSignalList = [...currentSignals];
    remove(newSignalList, { sessionID: this.uniqueIDPrefix });
    this.globalSignalsStream.next(newSignalList);
  }
}

@Injectable({
  providedIn: 'root',
})
export class NotifierService {
  baseSignalSession: NotifierSession;
  globalSignalsStream: BehaviorSubject<NotifierSignal[]> = new BehaviorSubject(
    []
  );
  history: BehaviorSubject<NotifierSignal[]> = new BehaviorSubject([]);

  constructor() {
    this.baseSignalSession = new NotifierSession(
      this.globalSignalsStream,
      this.history,
      'baseSignal'
    );
  }

  pushSignal(type: NotifierSignalType, data: boolean | string): string {
    return this.baseSignalSession.pushSignal(type, data);
  }

  removeSignal(id: string): boolean {
    return this.baseSignalSession.removeSignal(id);
  }

  removeSignals(ids: string[]): void {
    return this.baseSignalSession.removeSignals(ids);
  }

  /**
   * Removes a signal regardless of which session created it. Used by the
   * notification UI to dismiss a toast whose origin session (for example the
   * websocket connection) is not otherwise reachable.
   */
  removeSignalGlobally(id: string): boolean {
    const currentSignals = this.globalSignalsStream.getValue();
    const foundSignalIndex = findIndex(currentSignals, { id });
    if (foundSignalIndex < 0) {
      return false;
    }

    const newSignalList = [...currentSignals];
    pullAt(newSignalList, foundSignalIndex);
    this.globalSignalsStream.next(newSignalList);
    return true;
  }

  clearHistory(): void {
    this.history.next([]);
  }

  createSession(): NotifierSession {
    return new NotifierSession(
      this.globalSignalsStream,
      this.history,
      uniqueId('signalSession')
    );
  }
}
