import { ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';

import { FormViewContainerComponent } from './form-view-container.component';
import { CUSTOM_ELEMENTS_SCHEMA, Component } from '@angular/core';
import {
  FormArray,
  FormBuilder,
  FormGroup,
  FormsModule,
  ReactiveFormsModule,
} from '@angular/forms';
import { ClrInputModule } from '@clr/angular/forms/input';
import { ClrTextareaModule } from '@clr/angular/forms/textarea';
import { ClrSelectModule } from '@clr/angular/forms/select';
import { ClrRadioModule } from '@clr/angular/forms/radio';
import { ClrCheckboxModule } from '@clr/angular/forms/checkbox';
import { ActionForm } from '../../models/content';
import { FormHelper } from '../../models/form-helper';

@Component({
  standalone: false,
  template:
    '<app-form-view-container [form]="form" [formGroupContainer]="formGroup"></app-form-view-container>',
})
class TestWrapperComponent {
  form: ActionForm;
  formGroup: FormGroup;
}

describe('FormViewContainerComponent', () => {
  let component: TestWrapperComponent;
  let fixture: ComponentFixture<TestWrapperComponent>;
  let element: HTMLDivElement;
  let formHelper;

  const formBuilder: FormBuilder = new FormBuilder();

  beforeEach(
    waitForAsync(() => {
      TestBed.configureTestingModule({
        declarations: [TestWrapperComponent, FormViewContainerComponent],
        imports: [
          ReactiveFormsModule,
          FormsModule,
          ClrInputModule,
          ClrTextareaModule,
          ClrSelectModule,
          ClrRadioModule,
          ClrCheckboxModule,
        ],
        schemas: [CUSTOM_ELEMENTS_SCHEMA],
        providers: [{ provide: FormBuilder, useValue: formBuilder }],
      }).compileComponents();
    })
  );

  beforeEach(() => {
    fixture = TestBed.createComponent(TestWrapperComponent);
    component = fixture.componentInstance;
    element = fixture.nativeElement;
    formHelper = new FormHelper();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  describe('form group', () => {
    it('creates radio', () => {
      component.form = {
        fields: [
          {
            config: {
              configuration: {
                choices: [
                  { label: 'a', value: 'a', checked: true },
                  { label: 'b', value: 'b', checked: false },
                  { label: 'c', value: 'c', checked: false },
                ],
              },
              label: 'label',
              name: 'name',
              type: 'radio',
              value: null,
              placeholder: '',
              error: null,
              validators: null,
            },
            metadata: { type: 'formField' },
          },
        ],
      };
      component.formGroup = formHelper.createFromGroup(
        component.form,
        formBuilder
      );
      fixture.detectChanges();
      expect(element.querySelector('clr-radio-container')).not.toBeNull();
      expect(element.querySelector('clr-radio-wrapper')).not.toBeNull();
    });

    it('should create a select and verify is selected', () => {
      const name = 'name';
      component.form = {
        fields: [
          {
            config: {
              configuration: {
                choices: [
                  { label: 'a', value: 'a', checked: true },
                  { label: 'b', value: 'b', checked: false },
                  { label: 'c', value: 'c', checked: false },
                ],
              },
              label: 'label',
              name,
              type: 'select',
              value: null,
              placeholder: '',
              error: null,
              validators: null,
            },
            metadata: { type: 'formField' },
          },
        ],
      };
      component.formGroup = formHelper.createFromGroup(
        component.form,
        formBuilder
      );
      fixture.detectChanges();

      const selected = (
        component.formGroup.get(name) as FormArray
      ).getRawValue();
      expect(selected[0]).toEqual('a');
      expect(element.querySelector('clr-select-container')).not.toBeNull();
    });

    it('should create a select and verify is NOT selected', () => {
      const name = 'name';
      component.form = {
        fields: [
          {
            config: {
              configuration: {
                choices: [
                  { label: 'd', value: 'd', checked: false },
                  { label: 'b', value: 'b', checked: false },
                  { label: 'c', value: 'c', checked: false },
                ],
              },
              label: 'label',
              name,
              type: 'select',
              value: null,
              placeholder: '',
              error: null,
              validators: null,
            },
            metadata: { type: 'formField' },
          },
        ],
      };
      component.formGroup = formHelper.createFromGroup(
        component.form,
        formBuilder
      );
      fixture.detectChanges();

      const selected = (
        component.formGroup.get(name) as FormArray
      ).getRawValue();
      expect(selected[0]).toEqual(undefined);
      expect(element.querySelector('clr-select-container')).not.toBeNull();
    });
  });
});
