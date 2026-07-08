import { Component } from '@angular/core';
import { RouterOutlet } from '@angular/router';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [RouterOutlet],
  template: `
    <h1>Flatty Budget — Angular Shell</h1>
    <router-outlet />
  `,
})
export class AppComponent {}
