import { cleanup, render, screen } from '@testing-library/react';
import { act } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { fakeJwt, useConfig } from '../../test/helpers';
import { setConfig } from '../config/config';
import { RequireAdmin, RequireAnyRole, RequireAuth, RequireRole } from './guards';
import { tokenStore } from './token-store';

const assign = vi.fn();

beforeEach(() => {
  vi.stubGlobal('location', { assign });
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

  it('sem token leva ao login e não mostra o conteúdo', () => {
    useConfig();
    render(<RequireAuth>{secret}</RequireAuth>);
    expect(screen.queryByText('conteúdo restrito')).not.toBeInTheDocument();
    expect(assign).toHaveBeenCalledWith('/login');
  });

  it('com token mostra o conteúdo e reage ao logout', () => {
    useConfig();
    tokenStore.set(fakeJwt({ sub: '1' }));
    render(<RequireAuth>{secret}</RequireAuth>);
    expect(screen.getByText('conteúdo restrito')).toBeInTheDocument();

    act(() => tokenStore.clear());
    expect(screen.queryByText('conteúdo restrito')).not.toBeInTheDocument();
  });
});

describe('guards por role (espelham @HasRole, @HasAnyRole e @IsAdmin)', () => {
  it('RequireRole libera a role certa e bloqueia as outras', () => {
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
