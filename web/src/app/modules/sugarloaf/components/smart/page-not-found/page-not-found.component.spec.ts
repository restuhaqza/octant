// Copyright (c) 2019 the Octant contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
//

import { ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';
import { Location } from '@angular/common';
import { OverlayscrollbarsModule } from 'overlayscrollbars-ngx';
import { PageNotFoundComponent } from './page-not-found.component';

describe('PageNotFoundComponent', () => {
  let component: PageNotFoundComponent;
  let fixture: ComponentFixture<PageNotFoundComponent>;
  let location: jasmine.SpyObj<Location>;

  beforeEach(
    waitForAsync(() => {
      location = jasmine.createSpyObj('Location', ['back']);
      TestBed.configureTestingModule({
        declarations: [PageNotFoundComponent],
        imports: [OverlayscrollbarsModule],
        providers: [{ provide: Location, useValue: location }],
      }).compileComponents();
    })
  );

  beforeEach(() => {
    fixture = TestBed.createComponent(PageNotFoundComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('should explain that the page was not found', () => {
    expect(fixture.nativeElement.textContent).toContain('Page not found');
  });

  it('should go back in history when the action is used', () => {
    const button = fixture.nativeElement.querySelector('button');

    button.click();

    expect(location.back).toHaveBeenCalled();
  });
});
