import { TestBed } from '@angular/core/testing';
import { Router } from '@angular/router';
import { fakeConfigService, fakeIdp } from '../../testing/fakes';
import { ConfigService } from '../config/config.service';
import { AuthCallback } from './auth-callback';
import { AuthService } from './auth.service';

async function start(auth: AuthService, returnUrl: string): Promise<string> {
  vi.stubGlobal('fetch', fakeIdp());
  await auth.login(returnUrl);
  return (JSON.parse(sessionStorage.getItem('oidc_pending')!) as { state: string }).state;
}

describe('AuthCallback', () => {
  const assign = vi.fn();

  beforeEach(() => {
    sessionStorage.clear();
    TestBed.configureTestingModule({
      providers: [{ provide: ConfigService, useValue: fakeConfigService() }],
    });
    vi.stubGlobal('location', { assign, origin: 'https://app.example.com', search: '' });
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    assign.mockReset();
  });

  it('termina o login, guarda o token e volta para a página de origem', async () => {
    const auth = TestBed.inject(AuthService);
    const state = await start(auth, '/example');
    vi.stubGlobal('fetch', fakeIdp());
    vi.stubGlobal('location', {
      assign,
      origin: 'https://app.example.com',
      search: `?code=abc&state=${state}`,
    });
    const navigate = vi.spyOn(TestBed.inject(Router), 'navigateByUrl').mockResolvedValue(true);

    const fixture = TestBed.createComponent(AuthCallback);
    fixture.detectChanges();
    await vi.waitFor(() => expect(navigate).toHaveBeenCalledWith('/example', { replaceUrl: true }));

    expect(auth.token()).toBe('AT');
  });

  it('mostra o erro e oferece tentar de novo quando o provedor recusa', async () => {
    const auth = TestBed.inject(AuthService);
    const state = await start(auth, '/example');
    vi.stubGlobal('fetch', fakeIdp(400));
    vi.stubGlobal('location', {
      assign,
      origin: 'https://app.example.com',
      search: `?code=abc&state=${state}`,
    });

    const fixture = TestBed.createComponent(AuthCallback);
    fixture.detectChanges();
    await vi.waitFor(() => {
      fixture.detectChanges();
      expect(fixture.nativeElement.querySelector('[role="alert"]')?.textContent).toContain(
        'HTTP 400',
      );
    });

    expect(fixture.nativeElement.querySelector('button')?.textContent).toContain('Tentar de novo');
    expect(auth.token()).toBeNull();
  });
});
