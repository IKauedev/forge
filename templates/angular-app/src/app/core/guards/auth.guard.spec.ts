import { TestBed } from '@angular/core/testing';
import { ActivatedRouteSnapshot, RouterStateSnapshot } from '@angular/router';
import { DISCOVERY, fakeConfigService } from '../../testing/fakes';
import { AuthService } from '../auth/auth.service';
import { ConfigService } from '../config/config.service';
import { authGuard } from './auth.guard';

const assign = vi.fn();

function run(enabled: boolean, token: string | null): boolean {
  TestBed.configureTestingModule({
    providers: [{ provide: ConfigService, useValue: fakeConfigService({ enabled }) }],
  });
  if (token) {
    TestBed.inject(AuthService).setToken(token);
  }
  return TestBed.runInInjectionContext(
    () =>
      authGuard(
        {} as ActivatedRouteSnapshot,
        { url: '/example?a=1' } as RouterStateSnapshot,
      ) as boolean,
  );
}

describe('authGuard', () => {
  beforeEach(() => {
    sessionStorage.clear();
    vi.stubGlobal('location', {
      assign,
      origin: 'https://app.example.com',
      pathname: '/',
      search: '',
    });
    vi.stubGlobal(
      'fetch',
      vi.fn(async () => Response.json(DISCOVERY)),
    );
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    assign.mockReset();
  });

  it('libera quando a autenticação está desligada', () => {
    expect(run(false, null)).toBe(true);
    expect(assign).not.toHaveBeenCalled();
  });

  it('libera com token quando a autenticação está ligada', () => {
    expect(run(true, 'abc')).toBe(true);
  });

  it('sem token bloqueia e leva ao provedor, guardando a página de origem', async () => {
    expect(run(true, null)).toBe(false);

    await vi.waitFor(() => expect(assign).toHaveBeenCalledTimes(1));
    const url = new URL(assign.mock.calls[0][0] as string);
    expect(`${url.origin}${url.pathname}`).toBe(DISCOVERY.authorization_endpoint);
    expect(JSON.parse(sessionStorage.getItem('oidc_pending')!).returnUrl).toBe('/example?a=1');
  });

  it('um token vencido conta como não autenticado', async () => {
    TestBed.configureTestingModule({
      providers: [{ provide: ConfigService, useValue: fakeConfigService() }],
    });
    TestBed.inject(AuthService).setToken('velho', Date.now() - 1000);
    const allowed = TestBed.runInInjectionContext(
      () =>
        authGuard(
          {} as ActivatedRouteSnapshot,
          { url: '/example' } as RouterStateSnapshot,
        ) as boolean,
    );

    expect(allowed).toBe(false);
    await vi.waitFor(() => expect(assign).toHaveBeenCalled());
  });
});
