import { TestBed, waitForAsync } from '@angular/core/testing';
import { OverlayScrollbarsComponent } from 'overlayscrollbars-ngx';
import { ApplyYAMLComponent } from '../../components/smart/apply-yaml/apply-yaml.component';
import { FilterTextPipe } from './filtertext.pipe';

describe('FilterTextPipe', () => {
  beforeEach(
    waitForAsync(() => {
      TestBed.configureTestingModule({
        declarations: [ApplyYAMLComponent],
        imports: [OverlayScrollbarsComponent],
      });
    })
  );
  it('create an instance', () => {
    const pipe = new FilterTextPipe();
    expect(pipe).toBeTruthy();
  });
});
