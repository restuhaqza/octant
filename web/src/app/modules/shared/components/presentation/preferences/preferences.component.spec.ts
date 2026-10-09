import { ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';
import { SimpleChange } from '@angular/core';

import { PreferencesComponent } from './preferences.component';
import { BrowserModule } from '@angular/platform-browser';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import { Preferences } from '../../../models/preference';

describe('PreferencesComponent', () => {
  let component: PreferencesComponent;
  let fixture: ComponentFixture<PreferencesComponent>;

  beforeEach(
    waitForAsync(() => {
      TestBed.configureTestingModule({
        declarations: [PreferencesComponent],
        imports: [BrowserModule, ReactiveFormsModule, FormsModule],
      }).compileComponents();
    })
  );

  beforeEach(() => {
    fixture = TestBed.createComponent(PreferencesComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('replaces the valueChanges subscription on change and releases it on destroy', () => {
    const preferences: Preferences = {
      updateName: 'update',
      panels: [
        {
          name: 'general',
          sections: [
            {
              name: 'section',
              elements: [
                {
                  type: 'input',
                  name: 'name',
                  value: 'octant',
                  config: { label: 'Name', placeholder: '' },
                },
              ],
            },
          ],
        },
      ],
    };

    component.preferences = preferences;
    component.ngOnChanges({
      preferences: new SimpleChange(null, preferences, true),
    });
    const firstSub = component['formSub'];
    expect(firstSub.closed).toBe(false);

    // Any subsequent change rebuilds the form; the old stream must be dropped.
    component.ngOnChanges({ isOpen: new SimpleChange(false, true, false) });
    const secondSub = component['formSub'];
    expect(firstSub.closed).toBe(true);
    expect(secondSub).not.toBe(firstSub);

    component.ngOnDestroy();
    expect(secondSub.closed).toBe(true);
  });
});
