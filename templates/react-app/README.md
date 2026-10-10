# Starter App

Aplicação React (Vite + TypeScript, React Router, rotas lazy) gerada pelo forge.

## Desenvolvimento

```sh
npm ci
npm start            # http://localhost:4200
npm run lint
npm test
npm run build
```

## Estrutura

- `src/core/config/`: carrega o `config.json` antes de renderizar.
- `src/core/http/`: `apiFetch` (id de correlação, Bearer token, tratamento de 401 e `ApiError`).
- `src/core/auth/`: `tokenStore`, leitura das claims do JWT, `useAuth` e os guards por role.
- `src/layout/`: shell da aplicação. `src/shared/`: componentes reutilizáveis.
- `src/features/example/`: consulta de CEP, consome `GET /addresses/{cep}` da API.

## Configuração em runtime

A URL da API e a autenticação vêm de `config.json` (em `public/config.json` no desenvolvimento), carregado
antes de a aplicação iniciar. O mesmo build serve qualquer ambiente: o deploy e o container gravam o
`config.json` do ambiente. Não use `.env` por ambiente para isso.

```json
{
  "apiBaseUrl": "http://localhost:8080/api",
  "auth": { "enabled": false, "loginUrl": "/login", "rolesClaim": "roles" }
}
```

## Autenticação e guards por role

Com `auth.enabled = true`, as rotas protegidas exigem um token. Depois do login no seu provedor, chame
`tokenStore.set(accessToken)`. As roles vêm do claim `auth.rolesClaim` do JWT (`roles` por padrão;
`cognito:groups` no Cognito), em maiúsculas e sem `ROLE_`, como o backend as enxerga.

| Guard                         | Anotação do backend (spring-api) |
| ----------------------------- | -------------------------------- |
| `<RequireAuth>`               | `@IsAuthenticated`               |
| `<RequireRole role="X">`      | `@HasRole`                       |
| `<RequireAnyRole roles={[]}>` | `@HasAnyRole`                    |
| `<RequireAdmin>`              | `@IsAdmin`                       |

Os guards só controlam o que aparece na tela: o token é lido sem validar a assinatura. A autorização de
verdade é do backend.

## Docker

```sh
docker compose up --build   # http://localhost:4200
```

A imagem (Nginx sem privilégios, porta 8080) gera o `config.json` a partir de `API_BASE_URL`, `AUTH_ENABLED`,
`AUTH_LOGIN_URL` e `AUTH_ROLES_CLAIM`.

## Deploy e infraestrutura

- Pipeline: [docs/PIPELINE.md](docs/PIPELINE.md)
- Terraform (S3 privado + CloudFront com OAC): [docs/INFRA.md](docs/INFRA.md)
