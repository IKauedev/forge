import { apiFetch } from '../../core/http/http';

export interface Address {
  cep: string;
  street: string;
  district: string;
  city: string;
  state: string;
}

/** Consome GET /addresses/{cep} da API gerada por spring-api. */
export function lookupAddress(cep: string): Promise<Address> {
  const digits = cep.replace(/\D/g, '');
  return apiFetch<Address>(`/addresses/${digits}`);
}
