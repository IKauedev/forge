export interface AppConfig {
  apiBaseUrl: string;
  auth: {
    enabled: boolean;
    loginUrl: string;
  };
}
