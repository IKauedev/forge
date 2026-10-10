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
{ "apiBaseUrl": "http://localhost:8080/api", "auth": { "enabled": false, "loginUrl": "/login" } }
```

Com `auth.enabled = true`, o guard exige um token (`AuthService.setToken` depois do login no seu provedor).

## Docker

```sh
docker compose up --build   # http://localhost:4200
```

A imagem (Nginx sem privilégios, porta 8080) gera o `config.json` a partir de `API_BASE_URL`,
`AUTH_ENABLED` e `AUTH_LOGIN_URL`.

## Deploy e infraestrutura

- Pipeline: [docs/PIPELINE.md](docs/PIPELINE.md)
- Terraform (S3 privado + CloudFront com OAC): [docs/INFRA.md](docs/INFRA.md)
