import { Routes } from '@angular/router';

export const EXAMPLE_ROUTES: Routes = [
  {
    path: '',
    loadComponent: () => import('./address-lookup').then((m) => m.AddressLookup),
  },
];
