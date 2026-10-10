import { tokenStore } from '../auth/token-store';
import { getConfig } from '../config/config';

export const CORRELATION_HEADER = 'X-Correlation-Id';

export class ApiError extends Error {
  readonly status: number;
  readonly body: unknown;

  constructor(status: number, body: unknown) {
    super(`A API respondeu HTTP ${status}`);
    this.name = 'ApiError';
    this.status = status;
    this.body = body;
  }
}

/**
 * Chamada à API configurada: acrescenta o id de correlação e o Bearer token, limpa o token e leva
 * ao login em 401 (se a autenticação estiver ligada) e transforma erros HTTP em ApiError.
 */
export async function apiFetch<T>(path: string, init: RequestInit = {}): Promise<T> {
  const { apiBaseUrl, auth } = getConfig();

  const headers = new Headers(init.headers);
  if (!headers.has(CORRELATION_HEADER)) headers.set(CORRELATION_HEADER, crypto.randomUUID());
  const token = tokenStore.get();
  if (token) headers.set('Authorization', `Bearer ${token}`);

  const response = await fetch(`${apiBaseUrl}${path}`, { ...init, headers });
  if (response.ok) {
    return (response.status === 204 ? undefined : await response.json()) as T;
  }

  if (response.status === 401) {
    tokenStore.clear();
    if (auth.enabled) window.location.assign(auth.loginUrl);
  }
  throw new ApiError(response.status, await response.json().catch(() => null));
}
