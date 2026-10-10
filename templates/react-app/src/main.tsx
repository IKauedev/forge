import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { createBrowserRouter, RouterProvider } from 'react-router';
import { loadConfig } from './core/config/config';
import { routes } from './routes';

// A configuração de runtime precisa estar pronta antes de qualquer componente ler getConfig().
loadConfig()
  .then(() => {
    createRoot(document.getElementById('root')!).render(
      <StrictMode>
        <RouterProvider router={createBrowserRouter(routes)} />
      </StrictMode>,
    );
  })
  .catch((err: unknown) => {
    console.error(err);
    document.getElementById('root')!.textContent = 'Não foi possível carregar a configuração.';
  });
