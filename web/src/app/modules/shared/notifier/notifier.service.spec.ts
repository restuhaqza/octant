// Copyright (c) 2019 the Octant contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
//

import { TestBed } from '@angular/core/testing';
import {
  notifierHistoryLimit,
  NotifierService,
  NotifierSignalType,
} from './notifier.service';

describe('NotifierService', () => {
  beforeEach(() => TestBed.configureTestingModule({}));

  it('should be created', () => {
    const service: NotifierService = TestBed.inject(NotifierService);
    expect(service).toBeTruthy();
    expect(service.globalSignalsStream.getValue()).toEqual([]);
  });

  it('should be able to add signals', () => {
    const service = new NotifierService();
    const loadingSignalID = service.pushSignal(
      NotifierSignalType.LOADING,
      true
    );
    const errorSignalID = service.pushSignal(
      NotifierSignalType.ERROR,
      'You are doing it wrong'
    );
    const observedSignals = service.globalSignalsStream.getValue();
    const expectedSignals = [
      {
        id: loadingSignalID,
        sessionID: 'baseSignal',
        type: NotifierSignalType.LOADING,
        data: true,
        timestamp: jasmine.any(Number),
      },
      {
        id: errorSignalID,
        sessionID: 'baseSignal',
        type: NotifierSignalType.ERROR,
        data: 'You are doing it wrong',
        timestamp: jasmine.any(Number),
      },
    ];
    expect(observedSignals).toEqual(expectedSignals);
  });

  it('should be able to remove signals', () => {
    const service = new NotifierService();

    const warningSignalID = service.pushSignal(
      NotifierSignalType.WARNING,
      'nope nope nope'
    );
    const loadingSignalID = service.pushSignal(
      NotifierSignalType.LOADING,
      false
    );
    const errorSignalID = service.pushSignal(
      NotifierSignalType.ERROR,
      'No more worky'
    );

    const initialObservedSignals = service.globalSignalsStream.getValue();
    const initialExpectedSignals = [
      {
        id: warningSignalID,
        sessionID: 'baseSignal',
        type: NotifierSignalType.WARNING,
        data: 'nope nope nope',
        timestamp: jasmine.any(Number),
      },
      {
        id: loadingSignalID,
        sessionID: 'baseSignal',
        type: NotifierSignalType.LOADING,
        data: false,
        timestamp: jasmine.any(Number),
      },
      {
        id: errorSignalID,
        sessionID: 'baseSignal',
        type: NotifierSignalType.ERROR,
        data: 'No more worky',
        timestamp: jasmine.any(Number),
      },
    ];
    expect(initialObservedSignals).toEqual(initialExpectedSignals);

    service.removeSignals([loadingSignalID, '', null]);

    const currentObservedSignals = service.globalSignalsStream.getValue();
    const currentExpectedSignals = [
      {
        id: warningSignalID,
        sessionID: 'baseSignal',
        type: NotifierSignalType.WARNING,
        data: 'nope nope nope',
        timestamp: jasmine.any(Number),
      },
      {
        id: errorSignalID,
        sessionID: 'baseSignal',
        type: NotifierSignalType.ERROR,
        data: 'No more worky',
        timestamp: jasmine.any(Number),
      },
    ];
    expect(currentObservedSignals).toEqual(currentExpectedSignals);
  });

  it('should be able to create a signal session', () => {
    const service = new NotifierService();

    const warningSignalID = service.pushSignal(
      NotifierSignalType.WARNING,
      'nope nope nope'
    );

    const session = service.createSession();

    const loadingSignalID = session.pushSignal(
      NotifierSignalType.LOADING,
      false
    );
    const errorSignalID = session.pushSignal(
      NotifierSignalType.ERROR,
      'No more worky'
    );

    const initialObservedSignals = service.globalSignalsStream.getValue();
    const initialExpectedSignals = [
      {
        id: warningSignalID,
        sessionID: 'baseSignal',
        type: NotifierSignalType.WARNING,
        data: 'nope nope nope',
        timestamp: jasmine.any(Number),
      },
      {
        id: loadingSignalID,
        sessionID: session.id,
        type: NotifierSignalType.LOADING,
        data: false,
        timestamp: jasmine.any(Number),
      },
      {
        id: errorSignalID,
        sessionID: session.id,
        type: NotifierSignalType.ERROR,
        data: 'No more worky',
        timestamp: jasmine.any(Number),
      },
    ];
    expect(initialObservedSignals).toEqual(initialExpectedSignals);

    session.removeAllSignals();

    const currentObservedSignals = service.globalSignalsStream.getValue();
    const currentExpectedSignals = [
      {
        id: warningSignalID,
        sessionID: 'baseSignal',
        type: NotifierSignalType.WARNING,
        data: 'nope nope nope',
        timestamp: jasmine.any(Number),
      },
    ];
    expect(currentObservedSignals).toEqual(currentExpectedSignals);
  });

  it('keeps non-loading signals in history, newest first and without loading', () => {
    const service = new NotifierService();

    service.pushSignal(NotifierSignalType.LOADING, true);
    service.pushSignal(NotifierSignalType.WARNING, 'first');
    service.pushSignal(NotifierSignalType.ERROR, 'second');

    const history = service.history.getValue();

    expect(history.map(signal => signal.data)).toEqual(['second', 'first']);
  });

  it('bounds the history to the configured limit', () => {
    const service = new NotifierService();

    for (let i = 0; i < notifierHistoryLimit + 5; i++) {
      service.pushSignal(NotifierSignalType.INFO, `signal ${i}`);
    }

    expect(service.history.getValue().length).toBe(notifierHistoryLimit);
  });

  it('removes a signal from any session by id', () => {
    const service = new NotifierService();
    const session = service.createSession();
    const id = session.pushSignal(NotifierSignalType.ERROR, 'session signal');

    expect(service.globalSignalsStream.getValue().length).toBe(1);
    expect(service.removeSignalGlobally(id)).toBe(true);
    expect(service.globalSignalsStream.getValue().length).toBe(0);
    expect(service.removeSignalGlobally(id)).toBe(false);
  });

  it('clears the history', () => {
    const service = new NotifierService();
    service.pushSignal(NotifierSignalType.INFO, 'something');

    expect(service.history.getValue().length).toBe(1);
    service.clearHistory();
    expect(service.history.getValue().length).toBe(0);
  });
});
