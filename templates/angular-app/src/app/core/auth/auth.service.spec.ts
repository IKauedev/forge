import { TestBed } from '@angular/core/testing';
import { DISCOVERY, fakeConfigService, fakeIdp } from '../../testing/fakes';
import { ConfigService } from '../config/config.service';
import { AuthService } from './auth.service';

describe('AuthService', () => {
  const assign = vi.fn();
  let auth: AuthService;

  beforeEach(() => {
    sessionStorage.clear();
    vi.stubGlobal('location', {
      assign,
      origin: 'https://app.example.com',
      pathname: '/example',
      search: '?a=1',
    });
    vi.stubGlobal('fetch', fakeIdp());
    TestBed.configureTestingModule({
      providers: [{ provide: ConfigService, useValue: fakeConfigService() }],
    });
    auth = TestBed.inject(AuthService);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    assign.mockReset();
  });

  it('guarda o token e o limpa', () => {
    auth.setToken('abc', Date.now() + 60_000);
    expect(auth.isAuthenticated()).toBe(true);
    expect(sessionStorage.getItem('access_token')).toBe('abc');

    auth.clear();
    expect(auth.isAuthenticated()).toBe(false);
    expect(sessionStorage.getItem('access_token')).toBeNull();
  });

  it('um token vencido conta como ausente', () => {
    auth.setToken('abc', Date.now() - 1);
    expect(auth.token()).toBeNull();
    expect(auth.isAuthenticated()).toBe(false);
  });

  it('retoma a sessão guardada no sessionStorage', () => {
    auth.setToken('abc', Date.now() + 60_000);
    TestBed.resetTestingModule();
    TestBed.configureTestingModule({
      providers: [{ provide: ConfigService, useValue: fakeConfigService() }],
    });

    expect(TestBed.inject(AuthService).token()).toBe('abc');
  });

  it('login() leva ao provedor com a página atual como retorno por padrão', async () => {
    await auth.login();

    const url = new URL(assign.mock.calls[0][0] as string);
    expect(`${url.origin}${url.pathname}`).toBe(DISCOVERY.authorization_endpoint);
    expect(url.searchParams.get('redirect_uri')).toBe('https://app.example.com/auth/callback');
    expect(JSON.parse(sessionStorage.getItem('oidc_pending')!).returnUrl).toBe('/example?a=1');
  });

  it('logout() limpa a sessão e usa o logout do provedor', async () => {
    auth.setToken('abc', Date.now() + 60_000, 'IT');

    await auth.logout();

    expect(auth.isAuthenticated()).toBe(false);
    const url = new URL(assign.mock.calls[0][0] as string);
    expect(`${url.origin}${url.pathname}`).toBe(DISCOVERY.end_session_endpoint);
    expect(url.searchParams.get('id_token_hint')).toBe('IT');
  });
});
