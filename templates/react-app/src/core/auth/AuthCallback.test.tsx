import { cleanup, render, screen } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { DISCOVERY, useConfig } from '../../test/helpers';
import { setConfig } from '../config/config';
import AuthCallback from './AuthCallback';
import { buildLoginUrl } from './oidc';
import { tokenStore } from './token-store';

const settings = {
  issuer: DISCOVERY.issuer,
  clientId: 'client-123',
  scope: 'openid',
  redirectUri: 'https://app.example.com/auth/callback',
};

function idp(tokenStatus = 200) {
  return vi.fn(async (input: RequestInfo | URL) => {
    const url = String(input);
    if (url.endsWith('/.well-known/openid-configuration')) return Response.json(DISCOVERY);
    if (url === DISCOVERY.token_endpoint) {
      return tokenStatus === 200
        ? Response.json({ access_token: 'AT', id_token: 'IT', expires_in: 600 })
        : new Response('', { status: tokenStatus });
    }
    return new Response('', { status: 404 });
  });
}

function renderCallback() {
  return render(
    <MemoryRouter initialEntries={['/auth/callback']}>
      <Routes>
        <Route path="/auth/callback" element={<AuthCallback />} />
        <Route path="/example" element={<p>página de destino</p>} />
      </Routes>
    </MemoryRouter>,
  );
}

async function startLogin(returnUrl: string) {
  await buildLoginUrl(settings, returnUrl, sessionStorage, idp() as unknown as typeof fetch);
  return (JSON.parse(sessionStorage.getItem('oidc_pending')!) as { state: string }).state;
}

beforeEach(() => {
  sessionStorage.clear();
  useConfig({ clientId: 'client-123', scope: 'openid' });
});

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
  tokenStore.clear();
  setConfig(null);
});

describe('AuthCallback', () => {
  it('termina o login, guarda o token e volta para a página de origem', async () => {
    const state = await startLogin('/example');
    vi.stubGlobal('fetch', idp());
    vi.stubGlobal('location', {
      origin: 'https://app.example.com',
      search: `?code=abc&state=${state}`,
      assign: vi.fn(),
    });

    renderCallback();

    expect(await screen.findByText('página de destino')).toBeInTheDocument();
    expect(tokenStore.get()).toBe('AT');
    expect(tokenStore.getIdToken()).toBe('IT');
  });

  it('mostra o erro e oferece tentar de novo quando o provedor recusa', async () => {
    const state = await startLogin('/example');
    vi.stubGlobal('fetch', idp(400));
    vi.stubGlobal('location', {
      origin: 'https://app.example.com',
      search: `?code=abc&state=${state}`,
      assign: vi.fn(),
    });

    renderCallback();

    expect(await screen.findByRole('alert')).toHaveTextContent('HTTP 400');
    expect(screen.getByRole('button', { name: 'Tentar de novo' })).toBeInTheDocument();
    expect(tokenStore.get()).toBeNull();
  });

  it('recusa um retorno com state que não é o da tentativa', async () => {
    await startLogin('/example');
    vi.stubGlobal('fetch', idp());
    vi.stubGlobal('location', {
      origin: 'https://app.example.com',
      search: '?code=abc&state=forjado',
      assign: vi.fn(),
    });

    renderCallback();

    expect(await screen.findByRole('alert')).toHaveTextContent('Estado');
    expect(tokenStore.get()).toBeNull();
  });
});
