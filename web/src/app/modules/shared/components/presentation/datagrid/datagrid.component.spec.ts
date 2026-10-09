// Copyright (c) 2019 the Octant contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
//

import { ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';
import { DatagridComponent } from './datagrid.component';
import { TableView } from 'src/app/modules/shared/models/content';
import { SharedModule } from '../../../shared.module';
import { windowProvider, WindowToken } from '../../../../../window';

describe('DatagridComponent', () => {
  let component: DatagridComponent;
  let fixture: ComponentFixture<DatagridComponent>;

  const tableView = (emptyContent: string): TableView =>
    ({
      metadata: { type: 'table', title: [] },
      config: {
        columns: [{ name: 'Name', accessor: 'Name' }],
        rows: [],
        emptyContent,
        loading: false,
        filters: {},
      },
    } as unknown as TableView);

  beforeEach(
    waitForAsync(() => {
      TestBed.configureTestingModule({
        imports: [SharedModule],
        providers: [{ provide: WindowToken, useFactory: windowProvider }],
      }).compileComponents();
    })
  );

  beforeEach(() => {
    fixture = TestBed.createComponent(DatagridComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('shows the server-provided empty content when there are no rows', async () => {
    component.view = tableView("We couldn't find any things!");

    await new Promise(resolve => setTimeout(resolve));
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).toContain(
      "We couldn't find any things!"
    );
  });

  it('falls back to a default placeholder when empty content is blank', async () => {
    component.view = tableView('');

    await new Promise(resolve => setTimeout(resolve));
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).toContain('No items to display.');
  });
});
