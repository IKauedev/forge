# Starter Api

API Spring Boot com arquitetura hexagonal (ports and adapters).

## Pré-requisitos
JDK 25 e Docker. Windows, Linux e macOS (o `Makefile` usa `mvnw.cmd` no Windows).

## Comandos

| Comando | O que faz |
|---|---|
<!-- forge:if db!=none||cache=redis -->
| `make up` / `make down` | sobe / derruba a infraestrutura (banco e cache, `compose.yaml`) |
| `make run` | sobe a infraestrutura e roda a app com o perfil `local` |
<!-- forge:else -->
<!--~ | `make run` | roda a app com o perfil `local` | -->
<!-- forge:end -->
| `make test` | `mvnw verify`: testes, Spotless, ArchUnit, JaCoCo |
| `make image` | constrói a imagem Docker `starter-api:local` |
| `make full` | infraestrutura + app em containers (`compose.full.yaml`) |

<!-- forge:if db!=none||cache=redis -->
No perfil `local`, o Spring Boot lê o `compose.yaml` e injeta a conexão dos serviços sozinho
(`spring-boot-docker-compose`), então não há datasource para configurar. Fora do perfil `local`
a integração fica desligada e o jar de produção nem a inclui.

<!-- forge:end -->
<!-- forge:if db!=none -->
O banco escolhido na geração está registrado em `.forge.yaml` (`options.db`). Copie `.env.example`
para `.env` para trocar nome, usuário, senha e portas. Adminer:
`docker compose --profile tools up -d` (http://localhost:8081).
<!-- forge:else -->
<!--~ Este projeto foi gerado com `--db none`: sem banco, JPA nem Flyway; a consulta de CEP não é persistida. -->
<!-- forge:end -->

## CI/CD
Workflows em `.github/workflows/` (CI em PR; deploy na `main` para ECS via OIDC). Variáveis, Environments e roles
da AWS: [docs/PIPELINE.md](docs/PIPELINE.md).

## Imagem Docker
Build multi-stage com cache do Maven, jar extraído em camadas (`jarmode=tools`), JRE Alpine,
usuário não-root, `MaxRAMPercentage=75` e `ExitOnOutOfMemoryError`.

## API
<!-- forge:if db!=none -->
`GET /api/addresses/{cep}` consulta o CEP (ViaCEP) e guarda o resultado no banco.
<!-- forge:else -->
<!--~ `GET /api/addresses/{cep}` consulta o CEP (ViaCEP). -->
<!-- forge:end -->
O cliente HTTP é escolhido por `clients.viacep.impl` (`feign` | `webclient`).
Readiness: `/actuator/health/readiness`. Swagger só com o perfil `dev`.

<!-- forge:if auth!=none -->
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

<!-- forge:if auth=cognito -->
<!--~ ### Amazon Cognito -->
<!--~ -->
<!--~ Este projeto foi gerado com `--auth cognito`: aponte `JWT_ISSUER_URI` para o user pool -->
<!--~ (`https://cognito-idp.<regiao>.amazonaws.com/<user-pool-id>`). Os grupos do usuário viram roles -->
<!--~ (`app.security.roles-claim: "cognito:groups"`) e o username sai de `cognito:username`. -->
<!--~ Crie o user pool e os grupos `ADMIN`/`USER` fora do projeto (console, CLI ou Terraform). -->
<!-- forge:end -->
<!-- forge:end -->

<!-- forge:if cache!=none -->
## Cache
`AddressApplicationService.lookup` usa `@Cacheable("addresses")`. Tipo do cache em `spring.cache.type`
<!-- forge:if cache=redis -->
<!--~ (`redis`; o `compose.yaml` sobe o Redis e, no perfil `dev`/`prod`, `REDIS_HOST` aponta para ele). -->
<!-- forge:else -->
(`caffeine`, em memória, `maximumSize=500,expireAfterWrite=10m`).
<!-- forge:end -->
<!-- forge:end -->
