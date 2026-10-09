// Copyright (c) 2019 the Octant contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
//

import { ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';
import { SimpleChange } from '@angular/core';
import { TabsComponent } from './tabs.component';
import { SharedModule } from '../../../shared.module';
import { windowProvider, WindowToken } from '../../../../../window';
import { OctantTooltipComponent } from '../octant-tooltip/octant-tooltip';
import { SliderService } from 'src/app/modules/shared/slider/slider.service';
import { TextView } from 'src/app/modules/shared/models/content';

describe('TabsComponent', () => {
  let component: TabsComponent;
  let fixture: ComponentFixture<TabsComponent>;

  beforeEach(
    waitForAsync(() => {
      TestBed.configureTestingModule({
        declarations: [OctantTooltipComponent],
        imports: [SharedModule],
        providers: [{ provide: WindowToken, useFactory: windowProvider }],
      }).compileComponents();
    })
  );

  beforeEach(() => {
    fixture = TestBed.createComponent(TabsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('keeps a single activeTab subscription across input changes and releases it on destroy', () => {
    const slider = TestBed.inject(SliderService);
    const subscribeSpy = spyOn(slider.activeTab, 'subscribe').and.callThrough();

    component.extView = true;
    const views: TextView[] = [
      {
        config: { value: 'one' },
        metadata: { type: 'text', title: [], accessor: 'one' },
      },
      {
        config: { value: 'two' },
        metadata: { type: 'text', title: [], accessor: 'two' },
      },
    ];

    component.ngOnChanges({ views: new SimpleChange(null, views, true) });
    component.ngOnChanges({ views: new SimpleChange(views, views, false) });

    expect(subscribeSpy).toHaveBeenCalledTimes(1);

    const sub = component['activeTabSub'];
    expect(sub.closed).toBe(false);

    component.ngOnDestroy();
    expect(sub.closed).toBe(true);
  });
});
