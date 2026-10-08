/* Copyright (c) 2020 the Octant contributors. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';
import { CodeComponent } from './code.component';

describe('CodeComponent', () => {
  let component: CodeComponent;
  let fixture: ComponentFixture<CodeComponent>;

  beforeEach(
    waitForAsync(() => {
      TestBed.configureTestingModule({
        declarations: [CodeComponent],
      }).compileComponents();
    })
  );

  beforeEach(() => {
    fixture = TestBed.createComponent(CodeComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('copy button copies text', async () => {
    spyOn(component, 'copyToClipboard');

    const button = fixture.debugElement.nativeElement.querySelector('button');
    button.click();

    await fixture.whenStable();

    expect(component.copyToClipboard).toHaveBeenCalled();
  });
});
