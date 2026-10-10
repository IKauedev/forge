import { Injectable, computed, signal } from '@angular/core';

const TOKEN_KEY = 'access_token';

/**
 * Guarda o access token do usuário. A obtenção do token (OIDC, Cognito Hosted UI etc.) fica
 * fora do template: chame `setToken` depois do login no seu provedor.
 */
@Injectable({ providedIn: 'root' })
export class AuthService {
  private readonly tokenState = signal<string | null>(readToken());

  readonly token = this.tokenState.asReadonly();
  readonly isAuthenticated = computed(() => this.tokenState() !== null);

  setToken(token: string): void {
    this.tokenState.set(token);
    try {
      sessionStorage.setItem(TOKEN_KEY, token);
    } catch {
      // storage indisponível: o token vale só enquanto a página estiver aberta
    }
  }

  clear(): void {
    this.tokenState.set(null);
    try {
      sessionStorage.removeItem(TOKEN_KEY);
    } catch {
      // storage indisponível: nada a limpar
    }
  }
}

function readToken(): string | null {
  try {
    return sessionStorage.getItem(TOKEN_KEY);
  } catch {
    return null;
  }
}
