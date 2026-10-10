export interface AppConfig {
  apiBaseUrl: string;
  auth: {
    enabled: boolean;
    /** Emissor OIDC (ex.: https://cognito-idp.us-east-1.amazonaws.com/us-east-1_abc). */
    issuer: string;
    /** Client ID do app (público, sem secret) cadastrado no provedor. */
    clientId: string;
    scope: string;
  };
}
