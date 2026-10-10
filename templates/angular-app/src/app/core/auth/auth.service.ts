import { Injectable, computed, inject, signal } from '@angular/core';
import { ConfigService } from '../config/config.service';
import { buildLoginUrl, buildLogoutUrl, completeLogin, type OidcSettings } from './oidc';

export const CALLBACK_PATH = '/auth/callback';

const KEYS = { access: 'access_token', idToken: 'id_token', expiresAt: 'access_token_expires_at' };

/**
 * Sessão do usuário: guarda o access token (com validade) e conduz o login OIDC com PKCE
 * (login, completeLogin e logout). Um token vencido conta como ausente.
 */
@Injectable({ providedIn: 'root' })
export class AuthService {
  private readonly config = inject(ConfigService);

  private readonly tokenState = signal<string | null>(read(KEYS.access));
  private readonly idTokenState = signal<string | null>(read(KEYS.idToken));
  private readonly expiresAt = signal<number | null>(Number(read(KEYS.expiresAt)) || null);

  /** O access token vigente (null se não há ou se venceu). */
  readonly token = computed(() => {
    const token = this.tokenState();
    const expiresAt = this.expiresAt();
    return token !== null && (expiresAt === null || Date.now() < expiresAt) ? token : null;
  });
  readonly isAuthenticated = computed(() => this.token() !== null);

  setToken(token: string, expiresAtMs?: number, idToken?: string): void {
    this.tokenState.set(token);
    this.idTokenState.set(idToken ?? null);
    this.expiresAt.set(expiresAtMs ?? null);
    write(KEYS.access, token);
    write(KEYS.idToken, idToken ?? null);
    write(KEYS.expiresAt, expiresAtMs === undefined ? null : String(expiresAtMs));
  }

  clear(): void {
    this.tokenState.set(null);
    this.idTokenState.set(null);
    this.expiresAt.set(null);
    Object.values(KEYS).forEach((key) => write(key, null));
  }

  /** Leva o navegador ao provedor de identidade; ao voltar, a rota /auth/callback termina o login. */
  async login(returnUrl: string = currentUrl()): Promise<void> {
    window.location.assign(await buildLoginUrl(this.settings(), returnUrl, sessionStorage));
  }

  /** Chamada pela rota /auth/callback: troca o code pelos tokens e devolve para onde voltar. */
  async completeLogin(search: string = window.location.search): Promise<string> {
    const { tokens, returnUrl } = await completeLogin(this.settings(), search, sessionStorage);
    this.setToken(tokens.accessToken, tokens.expiresAt, tokens.idToken);
    return returnUrl;
  }

  /** Limpa a sessão local e, se o provedor tiver logout, encerra a sessão lá também. */
  async logout(): Promise<void> {
    const idToken = this.idTokenState() ?? undefined;
    this.clear();
    const url = await buildLogoutUrl(this.settings(), idToken, window.location.origin).catch(
      () => null,
    );
    window.location.assign(url ?? '/');
  }

  private settings(): OidcSettings {
    const { auth } = this.config.config();
    return {
      issuer: auth.issuer,
      clientId: auth.clientId,
      scope: auth.scope,
      redirectUri: `${window.location.origin}${CALLBACK_PATH}`,
    };
  }
}

function currentUrl(): string {
  return window.location.pathname + window.location.search;
}

function read(key: string): string | null {
  try {
    return sessionStorage.getItem(key);
  } catch {
    return null; // storage indisponível: o token vale só enquanto a página estiver aberta
  }
}

function write(key: string, value: string | null): void {
  try {
    if (value === null) sessionStorage.removeItem(key);
    else sessionStorage.setItem(key, value);
  } catch {
    // sem storage: mantém só em memória
  }
}
