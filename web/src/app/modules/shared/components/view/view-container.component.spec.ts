/*
 * Copyright (c) 2020 the Octant contributors. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';

import { ViewContainerComponent } from './view-container.component';
import { DYNAMIC_COMPONENTS_MAPPING } from '../../dynamic-components';
import { TextComponent } from '../presentation/text/text.component';
import { PodStatusView, TextView } from '../../models/content';
import { SharedModule } from '../../shared.module';
import { PodStatusComponent } from '../presentation/pod-status/pod-status.component';
import { windowProvider, WindowToken } from '../../../../window';

describe('ViewContainerComponent', () => {
  let component: ViewContainerComponent;
  let fixture: ComponentFixture<ViewContainerComponent>;

  beforeEach(
    waitForAsync(() => {
      TestBed.configureTestingModule({
        declarations: [ViewContainerComponent],
        imports: [SharedModule],
        providers: [
          { provide: WindowToken, useFactory: windowProvider },
          {
            provide: DYNAMIC_COMPONENTS_MAPPING,
            useValue: {
              text: TextComponent,
              podStatus: PodStatusComponent,
            },
          },
        ],
      }).compileComponents();
    })
  );

  beforeEach(() => {
    fixture = TestBed.createComponent(ViewContainerComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('should mark the created view for check on an in-place update', () => {
    const firstView: TextView = {
      config: { value: 'first value' },
      metadata: { type: 'text', title: [], accessor: 'accessor' },
    };

    const secondView: TextView = {
      config: { value: 'second value' },
      metadata: { type: 'text', title: [], accessor: 'accessor' },
    };

    component.view = firstView;
    fixture.detectChanges();
    expect(fixture.nativeElement.textContent).toContain('first value');

    // Same contentPath (same view type) -> the dynamically created component
    // is updated in place rather than recreated. Its view must still be marked
    // for check, otherwise the OnPush leaf goes stale.
    component.view = secondView;
    expect(
      (component.componentRef.instance as unknown as TextComponent).value
    ).toBe('second value');
    fixture.detectChanges();
    expect(fixture.nativeElement.textContent).toContain('second value');
    expect(fixture.nativeElement.textContent).not.toContain('first value');
  });

  it('should subscribe to the child viewInit hook only once across in-place updates', () => {
    const makeText = (value: string): TextView => ({
      config: { value },
      metadata: { type: 'text', title: [], accessor: 'accessor' },
    });

    component.view = makeText('first');
    fixture.detectChanges();

    const child = component.componentRef.instance as unknown as TextComponent;
    const emitSpy = spyOn(component.viewInit, 'emit');

    // Several in-place refreshes for the same contentPath.
    component.view = makeText('second');
    component.view = makeText('third');
    fixture.detectChanges();

    child.viewInit.emit();

    expect(emitSpy).toHaveBeenCalledTimes(1);
  });

  it('should resubscribe after a component type change without leaking the old hook', () => {
    const textView: TextView = {
      config: { value: 'some text' },
      metadata: { type: 'text', title: [], accessor: 'accessor' },
    };

    component.view = textView;
    fixture.detectChanges();
    const oldChild = component.componentRef
      .instance as unknown as TextComponent;

    const podStatusView: PodStatusView = {
      metadata: { type: 'podStatus' },
      config: {
        pods: {
          pod1: {
            details: [textView],
            status: 'ok',
          },
        },
      },
    };

    component.view = podStatusView;
    fixture.detectChanges();
    const newChild = component.componentRef
      .instance as unknown as PodStatusComponent;

    const emitSpy = spyOn(component.viewInit, 'emit');

    newChild.viewInit.emit();
    // The previous, now-destroyed child must no longer feed the parent.
    oldChild.viewInit.emit();

    expect(emitSpy).toHaveBeenCalledTimes(1);
  });

  it('should use the missing-component fallback for unknown view types', () => {
    const unknownView = {
      config: {},
      metadata: { type: 'notARealViewType', title: [], accessor: 'accessor' },
    } as unknown as TextView;

    expect(() => {
      component.view = unknownView;
      fixture.detectChanges();
    }).not.toThrow();

    expect(fixture.nativeElement.textContent).toContain(
      'notARealViewType not implemented'
    );
  });

  it('should recreate component when different type', () => {
    const textView: TextView = {
      config: { value: 'some text' },
      metadata: { type: 'text', title: [], accessor: 'accessor' },
    };

    const podStatusView: PodStatusView = {
      metadata: { type: 'podStatus' },
      config: {
        pods: {
          pod1: {
            details: [textView],
            status: 'ok',
          },
        },
      },
    };

    component.view = textView;
    fixture.detectChanges();
    expect(component.componentRef.instance.view.metadata.type).toEqual('text');
    expect(component.componentRef.componentType.name).toEqual('TextComponent');

    component.view = podStatusView;
    fixture.detectChanges();
    expect(component.componentRef.instance.view.metadata.type).toEqual(
      'podStatus'
    );
    expect(component.componentRef.componentType.name).toEqual(
      'PodStatusComponent'
    );
  });
});
