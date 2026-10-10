import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { ConfigService } from '../../core/config/config.service';
import { AddressService } from './address.service';

describe('AddressService', () => {
  it('consulta a API usando só os dígitos do CEP', () => {
    TestBed.configureTestingModule({
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        {
          provide: ConfigService,
          useValue: {
            config: () => ({ apiBaseUrl: 'https://api.example.com', auth: { enabled: false } }),
          },
        },
      ],
    });
    const http = TestBed.inject(HttpTestingController);
    let result: unknown;

    TestBed.inject(AddressService)
      .lookup('01310-100')
      .subscribe((a) => (result = a));
    http.expectOne('https://api.example.com/addresses/01310100').flush({ city: 'São Paulo' });

    expect(result).toEqual({ city: 'São Paulo' });
    http.verify();
  });
});
