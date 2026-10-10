import { beforeEach, describe, expect, it, vi } from 'vitest';
import { DISCOVERY, ISSUER } from '../../test/helpers';
import {
  OidcError,
  buildLoginUrl,
  buildLogoutUrl,
  completeLogin,
  discover,
  pkceChallenge,
  randomString,
  safeReturnUrl,
  type OidcSettings,
} from './oidc';

const settings: OidcSettings = {
  issuer: ISSUER,
  clientId: 'client-123',
  scope: 'openid profile email',
  redirectUri: 'https://app.example.com/auth/callback',
};

/** Servidor OIDC falso: descoberta + endpoint de token. */
function idp(
  tokenResponse: Response = Response.json({ access_token: 'AT', id_token: 'IT', expires_in: 600 }),
) {
  return vi.fn(async (input: RequestInfo | URL) => {
    const url = String(input);
    if (url.endsWith('/.well-known/openid-configuration')) return Response.json(DISCOVERY);
    if (url === DISCOVERY.token_endpoint) return tokenResponse.clone();
    return new Response('', { status: 404 });
  }) as unknown as typeof fetch;
}

beforeEach(() => sessionStorage.clear());

describe('PKCE', () => {
  it('calcula o code_challenge S256 do exemplo do RFC 7636', async () => {
    expect(await pkceChallenge('dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk')).toBe(
      'E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM',
    );
  });

  it('gera strings aleatórias base64url diferentes', () => {
    const a = randomString();
    expect(a).toMatch(/^[A-Za-z0-9_-]{43}$/);
    expect(randomString()).not.toBe(a);
  });
});

describe('safeReturnUrl', () => {
  it('aceita só caminhos da própria aplicação', () => {
    expect(safeReturnUrl('/example?x=1')).toBe('/example?x=1');
    for (const bad of [
      'https://evil.com',
      '//evil.com',
      '/\\evil.com',
      'example',
      '',
      null,
      undefined,
    ]) {
      expect(safeReturnUrl(bad)).toBe('/');
    }
  });
});

describe('discover', () => {
  it('lê o documento de descoberta do emissor', async () => {
    const e = await discover(`${ISSUER}/`, idp());
    expect(e.authorization_endpoint).toBe(DISCOVERY.authorization_endpoint);
    expect(e.token_endpoint).toBe(DISCOVERY.token_endpoint);
  });

  it('recusa um emissor diferente do informado', async () => {
    const other = vi.fn(async () =>
      Response.json({ ...DISCOVERY, issuer: 'https://outro.com' }),
    ) as unknown as typeof fetch;
    await expect(discover(ISSUER, other)).rejects.toThrow('não confere');
  });

  it('falha com HTTP de erro ou documento incompleto', async () => {
    const down = vi.fn(async () => new Response('', { status: 500 })) as unknown as typeof fetch;
    await expect(discover(ISSUER, down)).rejects.toBeInstanceOf(OidcError);
    const partial = vi.fn(async () => Response.json({ issuer: ISSUER })) as unknown as typeof fetch;
    await expect(discover(ISSUER, partial)).rejects.toThrow('authorization_endpoint');
  });
});

describe('buildLoginUrl', () => {
  it('monta a URL de autorização com PKCE e guarda state e verifier', async () => {
    const url = new URL(await buildLoginUrl(settings, '/example', sessionStorage, idp()));
    const q = url.searchParams;

    expect(`${url.origin}${url.pathname}`).toBe(DISCOVERY.authorization_endpoint);
    expect(q.get('response_type')).toBe('code');
    expect(q.get('client_id')).toBe('client-123');
    expect(q.get('redirect_uri')).toBe(settings.redirectUri);
    expect(q.get('scope')).toBe('openid profile email');
    expect(q.get('code_challenge_method')).toBe('S256');

    const pending = JSON.parse(sessionStorage.getItem('oidc_pending')!) as {
      state: string;
      verifier: string;
      returnUrl: string;
    };
    expect(q.get('state')).toBe(pending.state);
    expect(q.get('code_challenge')).toBe(await pkceChallenge(pending.verifier));
    expect(pending.returnUrl).toBe('/example');
    expect(JSON.stringify(url)).not.toContain(pending.verifier); // o verifier nunca vai na URL
  });

  it('troca um returnUrl externo por /', async () => {
    await buildLoginUrl(settings, 'https://evil.com', sessionStorage, idp());
    expect(JSON.parse(sessionStorage.getItem('oidc_pending')!).returnUrl).toBe('/');
  });
});

describe('completeLogin', () => {
  async function started() {
    await buildLoginUrl(settings, '/example', sessionStorage, idp());
    return JSON.parse(sessionStorage.getItem('oidc_pending')!) as {
      state: string;
      verifier: string;
    };
  }

  it('troca o code pelos tokens enviando o code_verifier', async () => {
    const pending = await started();
    const fetchMock = idp();

    const { tokens, returnUrl } = await completeLogin(
      settings,
      `?code=abc&state=${pending.state}`,
      sessionStorage,
      () => 1_000,
      fetchMock,
    );

    expect(tokens).toEqual({ accessToken: 'AT', idToken: 'IT', expiresAt: 1_000 + 600_000 });
    expect(returnUrl).toBe('/example');
    const tokenCall = (fetchMock as unknown as ReturnType<typeof vi.fn>).mock.calls.find(
      (c) => String(c[0]) === DISCOVERY.token_endpoint,
    )!;
    const body = new URLSearchParams(String((tokenCall[1] as RequestInit).body));
    expect(Object.fromEntries(body)).toEqual({
      grant_type: 'authorization_code',
      code: 'abc',
      redirect_uri: settings.redirectUri,
      client_id: 'client-123',
      code_verifier: pending.verifier,
    });
    expect(sessionStorage.getItem('oidc_pending')).toBeNull();
  });

  it('recusa um state diferente (proteção contra CSRF)', async () => {
    await started();
    await expect(
      completeLogin(settings, '?code=abc&state=forjado', sessionStorage, Date.now, idp()),
    ).rejects.toThrow('Estado');
  });

  it('só aceita o retorno uma vez', async () => {
    const pending = await started();
    await completeLogin(
      settings,
      `?code=abc&state=${pending.state}`,
      sessionStorage,
      Date.now,
      idp(),
    );
    await expect(
      completeLogin(settings, `?code=abc&state=${pending.state}`, sessionStorage, Date.now, idp()),
    ).rejects.toThrow('Estado');
  });

  it('mostra o erro devolvido pelo provedor', async () => {
    await started();
    await expect(
      completeLogin(
        settings,
        '?error=access_denied&error_description=Usu%C3%A1rio+cancelou',
        sessionStorage,
        Date.now,
        idp(),
      ),
    ).rejects.toThrow('Usuário cancelou');
  });

  it('falha sem code e quando o provedor recusa o code', async () => {
    const pending = await started();
    await expect(
      completeLogin(settings, `?state=${pending.state}`, sessionStorage, Date.now, idp()),
    ).rejects.toThrow('código');
    const p2 = await started();
    await expect(
      completeLogin(
        settings,
        `?code=x&state=${p2.state}`,
        sessionStorage,
        Date.now,
        idp(new Response('', { status: 400 })),
      ),
    ).rejects.toThrow('HTTP 400');
  });
});

describe('buildLogoutUrl', () => {
  it('usa o end_session_endpoint com o id_token_hint', async () => {
    const url = new URL((await buildLogoutUrl(settings, 'IT', 'https://app.example.com', idp()))!);
    expect(url.searchParams.get('id_token_hint')).toBe('IT');
    expect(url.searchParams.get('post_logout_redirect_uri')).toBe('https://app.example.com');
  });

  it('devolve null quando o provedor não publica logout', async () => {
    const noLogout = vi.fn(async () =>
      Response.json({ ...DISCOVERY, end_session_endpoint: undefined }),
    ) as unknown as typeof fetch;
    expect(await buildLogoutUrl(settings, 'IT', 'https://app.example.com', noLogout)).toBeNull();
  });
});
