import { getConfig } from '../config/config';
import { buildLoginUrl, buildLogoutUrl, completeLogin, type OidcSettings } from './oidc';
import { tokenStore } from './token-store';

export const CALLBACK_PATH = '/auth/callback';

function settings(): OidcSettings {
  const { auth } = getConfig();
  return {
    issuer: auth.issuer,
    clientId: auth.clientId,
    scope: auth.scope,
    redirectUri: `${window.location.origin}${CALLBACK_PATH}`,
  };
}

function currentUrl(): string {
  return window.location.pathname + window.location.search;
}

/** Leva o navegador ao provedor de identidade; ao voltar, a rota /auth/callback termina o login. */
export async function login(returnUrl: string = currentUrl()): Promise<void> {
  window.location.assign(await buildLoginUrl(settings(), returnUrl, sessionStorage));
}

/** Chamada pela rota /auth/callback: troca o code pelos tokens e devolve para onde voltar. */
export async function finishLogin(search: string = window.location.search): Promise<string> {
  const { tokens, returnUrl } = await completeLogin(settings(), search, sessionStorage);
  tokenStore.set(tokens.accessToken, tokens.expiresAt, tokens.idToken);
  return returnUrl;
}

/** Limpa a sessão local e, se o provedor tiver logout, encerra a sessão lá também. */
export async function logout(): Promise<void> {
  const idToken = tokenStore.getIdToken() ?? undefined;
  tokenStore.clear();
  const url = await buildLogoutUrl(settings(), idToken, window.location.origin).catch(() => null);
  window.location.assign(url ?? '/');
}
