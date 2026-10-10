import { HttpClient, provideHttpClient, withInterceptors } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { AuthService } from '../auth/auth.service';
import { ConfigService } from '../config/config.service';
import { authInterceptor } from './auth.interceptor';
import { CORRELATION_HEADER, correlationIdInterceptor } from './correlation-id.interceptor';
import { errorInterceptor } from './error.interceptor';

const API = 'https://api.example.com';

describe('interceptors', () => {
  let http: HttpClient;
  let controller: HttpTestingController;
  let auth: AuthService;

  beforeEach(() => {
    sessionStorage.clear();
    TestBed.configureTestingModule({
      providers: [
        provideHttpClient(
          withInterceptors([correlationIdInterceptor, authInterceptor, errorInterceptor]),
        ),
        provideHttpClientTesting(),
        {
          provide: ConfigService,
          useValue: {
            config: () => ({ apiBaseUrl: API, auth: { enabled: false, loginUrl: '/login' } }),
          },
        },
      ],
    });
    http = TestBed.inject(HttpClient);
    controller = TestBed.inject(HttpTestingController);
    auth = TestBed.inject(AuthService);
  });

  afterEach(() => controller.verify());

  it('adiciona X-Correlation-Id em toda requisição', () => {
    http.get(`${API}/x`).subscribe();
    const req = controller.expectOne(`${API}/x`);
    expect(req.request.headers.get(CORRELATION_HEADER)).toMatch(/^[0-9a-f-]{36}$/);
    req.flush({});
  });

  it('mantém o X-Correlation-Id já informado', () => {
    http.get(`${API}/x`, { headers: { [CORRELATION_HEADER]: 'abc' } }).subscribe();
    const req = controller.expectOne(`${API}/x`);
    expect(req.request.headers.get(CORRELATION_HEADER)).toBe('abc');
    req.flush({});
  });

  it('envia o Bearer só para a API configurada', () => {
    auth.setToken('t0k3n');
    http.get(`${API}/x`).subscribe();
    http.get('https://other.example.com/y').subscribe();

    expect(controller.expectOne(`${API}/x`).request.headers.get('Authorization')).toBe(
      'Bearer t0k3n',
    );
    const other = controller.expectOne('https://other.example.com/y');
    expect(other.request.headers.has('Authorization')).toBe(false);
    controller.match(() => true).forEach((r) => r.flush({}));
  });

  it('limpa o token ao receber 401', () => {
    auth.setToken('t0k3n');
    http.get(`${API}/x`).subscribe({ error: () => undefined });
    controller.expectOne(`${API}/x`).flush('no', { status: 401, statusText: 'Unauthorized' });

    expect(auth.isAuthenticated()).toBe(false);
  });
});
