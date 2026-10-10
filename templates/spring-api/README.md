# Starter Api

API Spring Boot com arquitetura hexagonal (ports and adapters).

## Pré-requisitos
JDK 25 e Docker. Windows, Linux e macOS (o `Makefile` usa `mvnw.cmd` no Windows).

## Comandos

| Comando | O que faz |
|---|---|
| `make up` / `make down` | sobe / derruba o Postgres (compose.yaml) |
| `make run` | sobe o banco e roda a app com o perfil `local` |
| `make test` | `mvnw verify`: testes, Spotless, ArchUnit, JaCoCo |
| `make image` | constrói a imagem Docker `starter-api:local` |
| `make full` | banco + app em containers (`compose.full.yaml`) |

No perfil `local`, o Spring Boot lê o `compose.yaml` e injeta a conexão do banco sozinho
(`spring-boot-docker-compose`), então não há datasource para configurar. Fora do perfil `local`
a integração fica desligada e o jar de produção nem a inclui.

Copie `.env.example` para `.env` para trocar usuário, senha e portas. Adminer:
`docker compose --profile tools up -d` (http://localhost:8081).

## Imagem Docker
Build multi-stage com cache do Maven, jar extraído em camadas (`jarmode=tools`), JRE Alpine,
usuário não-root, `MaxRAMPercentage=75` e `ExitOnOutOfMemoryError`.

## API
`GET /api/addresses/{cep}` consulta o CEP (ViaCEP) e guarda o resultado no banco.
O cliente HTTP é escolhido por `clients.viacep.impl` (`feign` | `webclient`).
Readiness: `/actuator/health/readiness`. Swagger só com o perfil `dev`.

## Segurança (JWT)
API stateless como *resource server* OAuth2. Tudo exige token, exceto `/actuator/health/**`, `/actuator/info`
e a documentação OpenAPI. Erros 401/403 saem como `application/problem+json`.

| Perfil | Origem do JWT |
|---|---|
| `local`, `test` | segredo HS256 de desenvolvimento (`app.security.dev-secret`) + `POST /dev/token` (só no `local`) |
| `dev`, `prod` | `JWT_ISSUER_URI` (Cognito, Keycloak, Auth0...) via `issuer-uri` |

```sh
TOKEN=$(curl -s -X POST "localhost:8080/dev/token?sub=ana&roles=ADMIN,USER&scopes=orders:read" | jq -r .access_token)
curl -H "Authorization: Bearer $TOKEN" localhost:8080/api/me
```

O claim de roles é configurável em `app.security.roles-claim` (padrão `roles`; no Cognito, `cognito:groups`).

### Anotações personalizadas (`shared.security`)
| Anotação | Equivale a |
|---|---|
| `@IsAuthenticated` | `isAuthenticated()` |
| `@IsAdmin` | `hasRole('ADMIN')` |
| `@HasRole("MANAGER")` | `hasRole('MANAGER')` |
| `@HasAnyRole(roles = {"'USER'", "'ADMIN'"})` | `hasAnyRole('USER','ADMIN')` (aspas simples obrigatórias) |
| `@HasScope("orders:read")` | `hasAuthority('SCOPE_orders:read')` |

Valem em métodos e classes. `SecurityUtils.currentUser()` devolve o `CurrentUser` (subject, username, roles,
scopes) para uso em adapters. Evite `@PostAuthorize` em métodos que gravam no banco.
