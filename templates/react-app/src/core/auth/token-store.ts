const KEYS = { access: 'access_token', idToken: 'id_token', expiresAt: 'access_token_expires_at' };

type Listener = () => void;

const listeners = new Set<Listener>();
let token: string | null = read(KEYS.access);
let idToken: string | null = read(KEYS.idToken);
let expiresAt: number | null = Number(read(KEYS.expiresAt)) || null;

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

function notify(): void {
  listeners.forEach((listener) => listener());
}

function expired(): boolean {
  return expiresAt !== null && Date.now() >= expiresAt;
}

export const tokenStore = {
  /** O access token vigente; um token vencido conta como ausente. */
  get: (): string | null => (token !== null && expired() ? null : token),

  getIdToken: (): string | null => idToken,

  /** Chame depois do login (o fluxo OIDC de login.ts já faz isso). */
  set(value: string, expiresAtMs?: number, id?: string): void {
    token = value;
    idToken = id ?? null;
    expiresAt = expiresAtMs ?? null;
    write(KEYS.access, value);
    write(KEYS.idToken, idToken);
    write(KEYS.expiresAt, expiresAt === null ? null : String(expiresAt));
    notify();
  },

  clear(): void {
    token = idToken = null;
    expiresAt = null;
    Object.values(KEYS).forEach((key) => write(key, null));
    notify();
  },

  subscribe(listener: Listener): () => void {
    listeners.add(listener);
    return () => listeners.delete(listener);
  },
};
