import { Pipe, PipeTransform } from '@angular/core';

@Pipe({
  standalone: false,
  name: 'join',
})
export class JoinPipe implements PipeTransform {
  transform(values: string[], delimiter: string): string {
    return values?.join(delimiter) || '';
  }
}
