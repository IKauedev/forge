export interface AppConfig {
  apiBaseUrl: string;
  auth: {
    enabled: boolean;
    /** Emissor OIDC (ex.: https://cognito-idp.us-east-1.amazonaws.com/us-east-1_abc). */
    issuer: string;
    /** Client ID do app (público, sem secret) cadastrado no provedor. */
    clientId: string;
    scope: string;
    /** Claim do JWT com as roles: "roles" (padrão do backend) ou "cognito:groups". */
    rolesClaim: string;
  };
}

let current: AppConfig | null = null;

/** Carrega /config.json (servido sem cache, ao lado do index.html). Chame antes de renderizar. */
export async function loadConfig(url = '/config.json'): Promise<AppConfig> {
  const response = await fetch(url, { cache: 'no-store' });
  if (!response.ok) {
    throw new Error(`Não foi possível carregar ${url} (HTTP ${response.status})`);
  }
  current = (await response.json()) as AppConfig;
  return current;
}

export function getConfig(): AppConfig {
  if (!current) {
    throw new Error('Configuração ainda não carregada (config.json).');
  }
  return current;
}

/** Só para testes. */
export function setConfig(config: AppConfig | null): void {
  current = config;
}
