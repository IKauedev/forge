import { AppConfig } from '../core/config/app-config';

export const ISSUER = 'https://idp.example.com/pool';

/** Documento de descoberta OIDC do provedor falso dos testes. */
export const DISCOVERY = {
  issuer: ISSUER,
  authorization_endpoint: 'https://idp.example.com/oauth2/authorize',
  token_endpoint: 'https://idp.example.com/oauth2/token',
  end_session_endpoint: 'https://idp.example.com/logout',
};

export function fakeAppConfig(auth: Partial<AppConfig['auth']> = {}): AppConfig {
  return {
    apiBaseUrl: 'https://api.example.com',
    auth: {
      enabled: true,
      issuer: ISSUER,
      clientId: 'client-123',
      scope: 'openid profile email',
      ...auth,
    },
  };
}

/** ConfigService já carregado, para os testes que não precisam do HTTP. */
export function fakeConfigService(auth: Partial<AppConfig['auth']> = {}) {
  const config = fakeAppConfig(auth);
  return { config: () => config, current: () => config };
}

/** Provedor OIDC falso: descoberta e endpoint de token. */
export function fakeIdp(tokenStatus = 200) {
  return vi.fn(async (input: RequestInfo | URL) => {
    const url = String(input);
    if (url.endsWith('/.well-known/openid-configuration')) return Response.json(DISCOVERY);
    if (url === DISCOVERY.token_endpoint) {
      return tokenStatus === 200
        ? Response.json({ access_token: 'AT', id_token: 'IT', expires_in: 600 })
        : new Response('', { status: tokenStatus });
    }
    return new Response('', { status: 404 });
  });
}
