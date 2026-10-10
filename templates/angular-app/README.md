# Starter App

Aplicação Angular (standalone components, rotas lazy, signals) gerada pelo forge.

## Desenvolvimento

```sh
npm ci
npm start            # http://localhost:4200
npm run lint
npm test -- --watch=false
npm run build
```

## Estrutura

- `src/app/core/`: `ConfigService`, interceptors (correlation id, auth, erro), guard e `AuthService`.
- `src/app/shared/`: componentes reutilizáveis.
- `src/app/layout/`: shell da aplicação.
- `src/app/features/example/`: consulta de CEP, consome `GET /addresses/{cep}` da API.

## Configuração em runtime

A URL da API e a autenticação vêm de `config.json` (em `public/config.json` no desenvolvimento), carregado
antes de a aplicação iniciar. O mesmo build serve qualquer ambiente: o deploy e o container gravam o
`config.json` do ambiente. Não use `environment.ts` por ambiente.

```json
{
  "apiBaseUrl": "http://localhost:8080/api",
  "auth": { "enabled": false, "issuer": "", "clientId": "", "scope": "openid profile email" }
}
```

### Login (OIDC com PKCE)

Com `auth.enabled = true`, o `authGuard` leva o usuário ao seu provedor de identidade
(**Authorization Code + PKCE**, sem biblioteca; funciona com Cognito, Keycloak, Auth0, Entra ID...). O fluxo:
`authGuard` → `AuthService.login()` (descoberta em `<issuer>/.well-known/openid-configuration`, `state` e
`code_verifier` no `sessionStorage`) → provedor → `/auth/callback` (`AuthCallback`) troca o `code` pelos tokens e
volta para a página de origem. O botão **Sair** (`AuthService.logout()`) limpa a sessão e usa o
`end_session_endpoint`, se o provedor tiver. Um 401 da API limpa o token e refaz o login; um token vencido conta
como não autenticado.

Configure no provedor um app **público** (sem client secret) com:

- Callback URL: `https://<seu-site>/auth/callback` (e `http://localhost:4200/auth/callback` para desenvolver);
- Sign-out URL: a raiz do site;
- escopos `openid profile email` (ajuste em `auth.scope`).

No `config.json`: `issuer` (ex.: `https://cognito-idp.<regiao>.amazonaws.com/<user-pool-id>`) e `clientId`. O
backend deve aceitar o **access token** desse emissor (no `spring-api`/`quarkus-api`, `JWT_ISSUER_URI`).

## Docker

```sh
docker compose up --build   # http://localhost:4200
```

A imagem (Nginx sem privilégios, porta 8080) gera o `config.json` a partir de `API_BASE_URL`,
`AUTH_ENABLED`, `AUTH_ISSUER`, `AUTH_CLIENT_ID` e `AUTH_SCOPE`.

## Deploy e infraestrutura

- Pipeline: [docs/PIPELINE.md](docs/PIPELINE.md)
- Terraform (S3 privado + CloudFront com OAC): [docs/INFRA.md](docs/INFRA.md)
