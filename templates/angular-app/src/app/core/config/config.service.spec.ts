import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { ConfigService } from './config.service';

describe('ConfigService', () => {
  let service: ConfigService;
  let http: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting()],
    });
    service = TestBed.inject(ConfigService);
    http = TestBed.inject(HttpTestingController);
  });

  it('falha ao ler a configuração antes de carregar', () => {
    expect(() => service.config()).toThrow();
  });

  it('carrega o config.json e expõe o valor', async () => {
    const loading = service.load();
    http.expectOne('config.json').flush({
      apiBaseUrl: 'https://api.example.com',
      auth: { enabled: true, issuer: 'https://idp.example.com', clientId: 'c', scope: 'openid' },
    });
    await loading;

    expect(service.config().apiBaseUrl).toBe('https://api.example.com');
    expect(service.config().auth.enabled).toBe(true);
  });
});
