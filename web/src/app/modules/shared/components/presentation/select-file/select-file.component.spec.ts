// Copyright (c) 2021 the Octant contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
//

import { ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';
import { SelectFileComponent } from './select-file.component';
import { SharedModule } from '../../../shared.module';
import { windowProvider, WindowToken } from '../../../../../window';

describe('SelectFileComponent', () => {
  let component: SelectFileComponent;
  let fixture: ComponentFixture<SelectFileComponent>;

  beforeEach(
    waitForAsync(() => {
      TestBed.configureTestingModule({
        imports: [SharedModule],
        providers: [{ provide: WindowToken, useFactory: windowProvider }],
      }).compileComponents();
    })
  );

  beforeEach(() => {
    fixture = TestBed.createComponent(SelectFileComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('shows the error status message with a clr-error ancestor', () => {
    component.status = 'error';
    component.statusMessage = 'bad file';
    fixture.detectChanges();

    const error = fixture.nativeElement.querySelector(
      '.clr-error clr-control-error'
    );
    expect(error).not.toBeNull();
    expect(error.textContent).toContain('bad file');
  });

  it('shows the success status message with a clr-success ancestor', () => {
    component.status = 'success';
    component.statusMessage = 'good file';
    fixture.detectChanges();

    const success = fixture.nativeElement.querySelector(
      '.clr-success clr-control-success'
    );
    expect(success).not.toBeNull();
    expect(success.textContent).toContain('good file');
  });
});
