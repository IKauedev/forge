import { cleanup, render, screen, waitFor } from '@testing-library/react';
import { act } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { DISCOVERY, fakeJwt, useConfig } from '../../test/helpers';
import { setConfig } from '../config/config';
import { RequireAdmin, RequireAnyRole, RequireAuth, RequireRole } from './guards';
import { tokenStore } from './token-store';

const assign = vi.fn();

beforeEach(() => {
  vi.stubGlobal('location', {
    assign,
    origin: 'https://app.example.com',
    pathname: '/example',
    search: '?a=1',
  });
  // provedor falso: só a descoberta é necessária para montar a URL de login
  vi.stubGlobal(
    'fetch',
    vi.fn(async () => Response.json(DISCOVERY)),
  );
});

afterEach(() => {
  cleanup(); // antes de limpar token/config: os componentes montados assinam o tokenStore
  vi.unstubAllGlobals();
  assign.mockReset();
  tokenStore.clear();
  setConfig(null);
});

const secret = <p>conteúdo restrito</p>;

describe('RequireAuth', () => {
  it('com a autenticação desligada mostra o conteúdo', () => {
    useConfig({ enabled: false });
    render(<RequireAuth>{secret}</RequireAuth>);
    expect(screen.getByText('conteúdo restrito')).toBeInTheDocument();
  });

  it('sem token leva ao provedor, guardando a página de origem, e não mostra o conteúdo', async () => {
    useConfig();
    render(<RequireAuth>{secret}</RequireAuth>);

    await waitFor(() => expect(assign).toHaveBeenCalledTimes(1));
    const url = new URL(assign.mock.calls[0][0] as string);
    expect(`${url.origin}${url.pathname}`).toBe(DISCOVERY.authorization_endpoint);
    expect(url.searchParams.get('redirect_uri')).toBe('https://app.example.com/auth/callback');
    expect(JSON.parse(sessionStorage.getItem('oidc_pending')!).returnUrl).toBe('/example?a=1');
    expect(screen.queryByText('conteúdo restrito')).not.toBeInTheDocument();
  });

  it('mostra o erro quando o provedor está indisponível', async () => {
    useConfig();
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => new Response('', { status: 503 })),
    );
    render(<RequireAuth>{secret}</RequireAuth>);

    expect(await screen.findByRole('alert')).toHaveTextContent('HTTP 503');
    expect(assign).not.toHaveBeenCalled();
  });

  it('com token mostra o conteúdo e reage ao logout', () => {
    useConfig();
    tokenStore.set(fakeJwt({ sub: '1' }));
    render(<RequireAuth>{secret}</RequireAuth>);
    expect(screen.getByText('conteúdo restrito')).toBeInTheDocument();
    expect(assign).not.toHaveBeenCalled();

    act(() => tokenStore.clear());
    expect(screen.queryByText('conteúdo restrito')).not.toBeInTheDocument();
  });

  it('um token vencido conta como não autenticado', async () => {
    useConfig();
    tokenStore.set(fakeJwt({ sub: '1' }), Date.now() - 1000);
    render(<RequireAuth>{secret}</RequireAuth>);

    await waitFor(() => expect(assign).toHaveBeenCalled());
    expect(screen.queryByText('conteúdo restrito')).not.toBeInTheDocument();
  });
});

describe('guards por role (espelham @HasRole, @HasAnyRole e @IsAdmin)', () => {
  it('RequireRole libera a role certa', () => {
    useConfig();
    tokenStore.set(fakeJwt({ roles: ['user'] }));

    render(<RequireRole role="USER">{secret}</RequireRole>);
    expect(screen.getByText('conteúdo restrito')).toBeInTheDocument();
  });

  it('RequireAdmin bloqueia quem não é ADMIN', () => {
    useConfig();
    tokenStore.set(fakeJwt({ roles: ['USER'] }));

    render(<RequireAdmin>{secret}</RequireAdmin>);

    expect(screen.queryByText('conteúdo restrito')).not.toBeInTheDocument();
    expect(screen.getByRole('alert')).toHaveTextContent('permissão');
  });

  it('RequireAnyRole libera com qualquer uma das roles', () => {
    useConfig();
    tokenStore.set(fakeJwt({ roles: ['ROLE_ADMIN'] }));

    render(<RequireAnyRole roles={['USER', 'ADMIN']}>{secret}</RequireAnyRole>);

    expect(screen.getByText('conteúdo restrito')).toBeInTheDocument();
  });

  it('com a autenticação desligada não exige role', () => {
    useConfig({ enabled: false });
    render(<RequireAdmin>{secret}</RequireAdmin>);
    expect(screen.getByText('conteúdo restrito')).toBeInTheDocument();
  });
});
