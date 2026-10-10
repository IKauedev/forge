/**
 * OpenID Connect com Authorization Code + PKCE, sem biblioteca. Funciona com qualquer provedor que
 * publique /.well-known/openid-configuration (Cognito, Keycloak, Auth0, Entra ID...).
 *
 * É código puro (recebe storage e fetch por parâmetro) para ser testado sem navegador.
 */

export interface OidcSettings {
  issuer: string;
  clientId: string;
  scope: string;
  /** URL absoluta da página /auth/callback desta aplicação (cadastre-a no provedor). */
  redirectUri: string;
}

export interface Endpoints {
  authorization_endpoint: string;
  token_endpoint: string;
  end_session_endpoint?: string;
}

export interface Tokens {
  accessToken: string;
  idToken?: string;
  /** Epoch em milissegundos em que o access token expira. */
  expiresAt: number;
}

export class OidcError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'OidcError';
  }
}

const PENDING_KEY = 'oidc_pending';
const DEFAULT_LIFETIME_S = 3600;

interface Pending {
  state: string;
  verifier: string;
  returnUrl: string;
}

export function base64Url(bytes: Uint8Array): string {
  let binary = '';
  bytes.forEach((b) => (binary += String.fromCharCode(b)));
  return btoa(binary).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
}

/** String aleatória segura (para state e code_verifier). */
export function randomString(byteLength = 32): string {
  return base64Url(crypto.getRandomValues(new Uint8Array(byteLength)));
}

/** code_challenge = BASE64URL(SHA-256(code_verifier)) (RFC 7636, método S256). */
export async function pkceChallenge(verifier: string): Promise<string> {
  const digest = await crypto.subtle.digest('SHA-256', new TextEncoder().encode(verifier));
  return base64Url(new Uint8Array(digest));
}

/** Só caminhos da própria aplicação: evita que um returnUrl malicioso vire um redirecionamento aberto. */
export function safeReturnUrl(url: string | null | undefined): string {
  if (!url || !url.startsWith('/') || url.startsWith('//') || url.includes('\\')) return '/';
  return url;
}

export async function discover(issuer: string, fetchFn: typeof fetch = fetch): Promise<Endpoints> {
  const base = issuer.replace(/\/+$/, '');
  const response = await fetchFn(`${base}/.well-known/openid-configuration`);
  if (!response.ok) {
    throw new OidcError(
      `Não foi possível ler a configuração OIDC de ${base} (HTTP ${response.status})`,
    );
  }
  const doc = (await response.json()) as Partial<Endpoints> & { issuer?: string };
  if (doc.issuer && doc.issuer.replace(/\/+$/, '') !== base) {
    throw new OidcError('O emissor informado não confere com o publicado pelo provedor');
  }
  if (!doc.authorization_endpoint || !doc.token_endpoint) {
    throw new OidcError('A configuração OIDC não traz authorization_endpoint e token_endpoint');
  }
  return {
    authorization_endpoint: doc.authorization_endpoint,
    token_endpoint: doc.token_endpoint,
    end_session_endpoint: doc.end_session_endpoint,
  };
}

/** Guarda state e verifier e devolve a URL do provedor para onde o navegador deve ir. */
export async function buildLoginUrl(
  settings: OidcSettings,
  returnUrl: string,
  storage: Storage,
  fetchFn: typeof fetch = fetch,
): Promise<string> {
  const endpoints = await discover(settings.issuer, fetchFn);
  const pending: Pending = {
    state: randomString(),
    verifier: randomString(48),
    returnUrl: safeReturnUrl(returnUrl),
  };
  storage.setItem(PENDING_KEY, JSON.stringify(pending));

  const url = new URL(endpoints.authorization_endpoint);
  url.searchParams.set('response_type', 'code');
  url.searchParams.set('client_id', settings.clientId);
  url.searchParams.set('redirect_uri', settings.redirectUri);
  url.searchParams.set('scope', settings.scope);
  url.searchParams.set('state', pending.state);
  url.searchParams.set('code_challenge', await pkceChallenge(pending.verifier));
  url.searchParams.set('code_challenge_method', 'S256');
  return url.toString();
}

function takePending(storage: Storage): Pending | null {
  const raw = storage.getItem(PENDING_KEY);
  storage.removeItem(PENDING_KEY); // uso único: um code só pode ser trocado uma vez
  if (!raw) return null;
  try {
    return JSON.parse(raw) as Pending;
  } catch {
    return null;
  }
}

/** Valida o retorno do provedor (?code=&state=) e troca o code pelos tokens. */
export async function completeLogin(
  settings: OidcSettings,
  search: string,
  storage: Storage,
  now: () => number = Date.now,
  fetchFn: typeof fetch = fetch,
): Promise<{ tokens: Tokens; returnUrl: string }> {
  const params = new URLSearchParams(search);
  const pending = takePending(storage);

  const error = params.get('error');
  if (error) {
    throw new OidcError(params.get('error_description') ?? error);
  }
  const code = params.get('code');
  if (!code) throw new OidcError('O provedor não devolveu o código de autorização');
  if (!pending || params.get('state') !== pending.state) {
    throw new OidcError('Estado da autenticação inválido ou expirado; tente entrar de novo');
  }

  const endpoints = await discover(settings.issuer, fetchFn);
  const response = await fetchFn(endpoints.token_endpoint, {
    method: 'POST',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: new URLSearchParams({
      grant_type: 'authorization_code',
      code,
      redirect_uri: settings.redirectUri,
      client_id: settings.clientId,
      code_verifier: pending.verifier,
    }),
  });
  if (!response.ok) {
    throw new OidcError(`O provedor recusou o código (HTTP ${response.status})`);
  }
  const body = (await response.json()) as {
    access_token?: string;
    id_token?: string;
    expires_in?: number;
  };
  if (!body.access_token) throw new OidcError('A resposta do provedor não traz access_token');

  return {
    tokens: {
      accessToken: body.access_token,
      idToken: body.id_token,
      expiresAt: now() + (body.expires_in ?? DEFAULT_LIFETIME_S) * 1000,
    },
    returnUrl: pending.returnUrl,
  };
}

/** URL de logout do provedor, se ele publicar end_session_endpoint (o Cognito usa /logout próprio). */
export async function buildLogoutUrl(
  settings: OidcSettings,
  idToken: string | undefined,
  postLogoutRedirectUri: string,
  fetchFn: typeof fetch = fetch,
): Promise<string | null> {
  const { end_session_endpoint } = await discover(settings.issuer, fetchFn);
  if (!end_session_endpoint) return null;
  const url = new URL(end_session_endpoint);
  if (idToken) url.searchParams.set('id_token_hint', idToken);
  url.searchParams.set('post_logout_redirect_uri', postLogoutRedirectUri);
  url.searchParams.set('client_id', settings.clientId);
  return url.toString();
}
