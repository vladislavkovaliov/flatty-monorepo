import { Routes } from '@angular/router';

export const routes: Routes = [
  {
    path: '',
    pathMatch: 'full',
    redirectTo: 'external',
  },
  {
    path: 'external',
    loadComponent: () =>
      import('./components/external/external.component').then(
        (m) => m.ExternalComponent
      ),
  },
];
