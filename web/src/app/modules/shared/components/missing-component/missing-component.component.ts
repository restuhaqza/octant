import { Component, Input } from '@angular/core';

@Component({
  standalone: false,
  selector: 'app-missing-component',
  templateUrl: './missing-component.component.html',
  styleUrls: ['./missing-component.component.sass'],
})
export class MissingComponentComponent {
  @Input() name: string;

  constructor() {}
}
