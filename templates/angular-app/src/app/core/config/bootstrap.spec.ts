import { provideHttpClient, withInterceptors } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { fakeAppConfig } from '../../testing/fakes';
import { AuthService } from '../auth/auth.service';
import { authInterceptor } from '../interceptors/auth.interceptor';
import { correlationIdInterceptor } from '../interceptors/correlation-id.interceptor';
import { errorInterceptor } from '../interceptors/error.interceptor';
import { ConfigService } from './config.service';

/**
 * O bootstrap carrega o config.json por um HttpClient que já tem os interceptors. Eles rodam ANTES de a
 * configuração existir e não podem depender dela (regressão: o authInterceptor chamava config()).
 */
describe('carga do config.json com os interceptors reais', () => {
  beforeEach(() => {
    sessionStorage.clear();
    TestBed.configureTestingModule({
      providers: [
        provideHttpClient(
          withInterceptors([correlationIdInterceptor, authInterceptor, errorInterceptor]),
        ),
        provideHttpClientTesting(),
      ],
    });
  });

  it('carrega sem sessão', async () => {
    const config = TestBed.inject(ConfigService);
    const loading = config.load();
    TestBed.inject(HttpTestingController).expectOne('config.json').flush(fakeAppConfig());
    await loading;

    expect(config.config().apiBaseUrl).toBe('https://api.example.com');
  });

  it('carrega para quem volta com um token na sessão', async () => {
    TestBed.inject(AuthService).setToken('t0k3n', Date.now() + 60_000);
    const config = TestBed.inject(ConfigService);

    const loading = config.load();
    const req = TestBed.inject(HttpTestingController).expectOne('config.json');
    expect(req.request.headers.has('Authorization')).toBe(false); // o config.json é público
    req.flush(fakeAppConfig());
    await loading;

    expect(config.current()?.auth.clientId).toBe('client-123');
  });

  it('current() é null até a configuração carregar', () => {
    expect(TestBed.inject(ConfigService).current()).toBeNull();
  });
});
