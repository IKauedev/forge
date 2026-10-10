const TOKEN_KEY = 'access_token';

type Listener = () => void;

const listeners = new Set<Listener>();
let token: string | null = read();

function read(): string | null {
  try {
    return sessionStorage.getItem(TOKEN_KEY);
  } catch {
    return null; // storage indisponível: o token vale só enquanto a página estiver aberta
  }
}

function notify(): void {
  listeners.forEach((listener) => listener());
}

export const tokenStore = {
  get: (): string | null => token,

  /** Chame depois do login no seu provedor (OIDC, Cognito Hosted UI etc.). */
  set(value: string): void {
    token = value;
    try {
      sessionStorage.setItem(TOKEN_KEY, value);
    } catch {
      // sem storage: mantém só em memória
    }
    notify();
  },

  clear(): void {
    token = null;
    try {
      sessionStorage.removeItem(TOKEN_KEY);
    } catch {
      // sem storage: nada a limpar
    }
    notify();
  },

  subscribe(listener: Listener): () => void {
    listeners.add(listener);
    return () => listeners.delete(listener);
  },
};
