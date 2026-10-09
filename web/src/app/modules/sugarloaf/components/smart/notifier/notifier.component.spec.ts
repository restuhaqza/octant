// Copyright (c) 2019 the Octant contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
//

import { ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';
import { ClarityModule } from '@clr/angular';
import { SharedModule } from 'src/app/modules/shared/shared.module';
import {
  NotifierService,
  NotifierSignalType,
} from 'src/app/modules/shared/notifier/notifier.service';

import { NotifierComponent } from './notifier.component';

describe('NotifierComponent', () => {
  let component: NotifierComponent;
  let fixture: ComponentFixture<NotifierComponent>;
  let notifier: NotifierService;

  beforeEach(
    waitForAsync(() => {
      TestBed.configureTestingModule({
        declarations: [NotifierComponent],
        imports: [ClarityModule, SharedModule],
      }).compileComponents();
    })
  );

  beforeEach(() => {
    fixture = TestBed.createComponent(NotifierComponent);
    component = fixture.componentInstance;
    notifier = TestBed.inject(NotifierService);
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('renders a toast per notification instead of only the last one', () => {
    notifier.pushSignal(NotifierSignalType.ERROR, 'boom');
    notifier.pushSignal(NotifierSignalType.WARNING, 'careful');
    fixture.detectChanges();

    expect(component.toasts.length).toBe(2);
    expect(fixture.nativeElement.textContent).toContain('boom');
    expect(fixture.nativeElement.textContent).toContain('careful');
  });

  it('dismisses a single toast', () => {
    notifier.pushSignal(NotifierSignalType.ERROR, 'boom');
    notifier.pushSignal(NotifierSignalType.WARNING, 'careful');
    fixture.detectChanges();

    component.dismiss(component.toasts[0]);
    fixture.detectChanges();

    expect(component.toasts.length).toBe(1);
    expect(fixture.nativeElement.textContent).not.toContain('boom');
    expect(fixture.nativeElement.textContent).toContain('careful');
  });

  it('dismisses all toasts', () => {
    notifier.pushSignal(NotifierSignalType.ERROR, 'boom');
    notifier.pushSignal(NotifierSignalType.WARNING, 'careful');
    fixture.detectChanges();

    component.dismissAll();
    fixture.detectChanges();

    expect(component.toasts.length).toBe(0);
  });

  it('keeps a notification history that can be toggled and cleared', () => {
    notifier.pushSignal(NotifierSignalType.ERROR, 'boom');
    fixture.detectChanges();

    expect(component.history.length).toBe(1);

    component.toggleHistory();
    fixture.detectChanges();
    expect(component.historyOpen).toBe(true);

    component.clearHistory();
    fixture.detectChanges();
    expect(component.history.length).toBe(0);
  });
});
