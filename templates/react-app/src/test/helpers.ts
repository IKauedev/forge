import { setConfig, type AppConfig } from '../core/config/config';

export function useConfig(overrides: Partial<AppConfig['auth']> = {}): AppConfig {
  const config: AppConfig = {
    apiBaseUrl: 'https://api.example.com',
    auth: { enabled: true, loginUrl: '/login', rolesClaim: 'roles', ...overrides },
  };
  setConfig(config);
  return config;
}

function base64url(value: unknown): string {
  const bytes = new TextEncoder().encode(JSON.stringify(value));
  return btoa(String.fromCharCode(...bytes))
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=+$/, '');
}

/** JWT sem assinatura válida: o frontend só lê as claims. */
export function fakeJwt(claims: Record<string, unknown>): string {
  return `${base64url({ alg: 'none' })}.${base64url(claims)}.sig`;
}
