import { HttpClient } from '@angular/common/http';
import { Injectable, inject } from '@angular/core';
import { Observable } from 'rxjs';
import { ConfigService } from '../../core/config/config.service';

export interface Address {
  cep: string;
  street: string;
  district: string;
  city: string;
  state: string;
}

/** Consome GET /addresses/{cep} da API gerada por spring-api. */
@Injectable({ providedIn: 'root' })
export class AddressService {
  private readonly http = inject(HttpClient);
  private readonly config = inject(ConfigService);

  lookup(cep: string): Observable<Address> {
    const digits = cep.replace(/\D/g, '');
    return this.http.get<Address>(`${this.config.config().apiBaseUrl}/addresses/${digits}`);
  }
}
