import { lazy, Suspense } from 'react';
import { Navigate, type RouteObject } from 'react-router';
import { RequireAuth } from './core/auth/guards';
import Shell from './layout/Shell';
import NotFound from './shared/NotFound';

// Rotas lazy: cada feature vira um chunk separado.
const AddressLookup = lazy(() => import('./features/example/AddressLookup'));
const AuthCallback = lazy(() => import('./core/auth/AuthCallback'));

export const routes: RouteObject[] = [
  {
    element: <Shell />,
    children: [
      { index: true, element: <Navigate to="/example" replace /> },
      {
        path: 'example',
        element: (
          <RequireAuth>
            <Suspense fallback={<p>Carregando…</p>}>
              <AddressLookup />
            </Suspense>
          </RequireAuth>
        ),
      },
      {
        // pública: é por aqui que o provedor devolve o usuário depois do login
        path: 'auth/callback',
        element: (
          <Suspense fallback={<p>Entrando…</p>}>
            <AuthCallback />
          </Suspense>
        ),
      },
      { path: '*', element: <NotFound /> },
    ],
  },
];
