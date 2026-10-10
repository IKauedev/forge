import { Routes } from '@angular/router';
import { authGuard } from './core/guards/auth.guard';

export const routes: Routes = [
  { path: '', pathMatch: 'full', redirectTo: 'example' },
  {
    path: 'example',
    canActivate: [authGuard],
    loadChildren: () => import('./features/example/example.routes').then((m) => m.EXAMPLE_ROUTES),
  },
  {
    path: '**',
    loadComponent: () => import('./shared/not-found').then((m) => m.NotFound),
  },
];
