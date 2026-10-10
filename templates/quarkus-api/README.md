# Starter Api

API Quarkus (linha LTS 3.40) com arquitetura hexagonal (ports and adapters), CDI e Java 25.

## Pré-requisitos
JDK 25 e Docker. Windows, Linux e macOS (o `Makefile` usa `mvnw.cmd` no Windows).

## Comandos

| Comando | O que faz |
|---|---|
| `make dev` | `quarkus:dev` com live reload e Dev UI (http://localhost:8080/q/dev) |
<!-- forge:if db!=none -->
| `make up` / `make down` | sobe / derruba um banco fixo (`compose.yaml`) no lugar do Dev Services |
<!-- forge:end -->
| `make test` | `mvnw verify`: testes, Spotless, ArchUnit e cobertura mínima de 60% (JaCoCo) |
| `make image` | constrói a imagem Docker `starter-api:local` |
| `make full` | aplicação (e banco) em containers (`compose.full.yaml`) |

<!-- forge:if db!=none -->
O banco escolhido na geração está registrado em `.forge.yaml` (`options.db`). Em `quarkus:dev` e nos testes o
**Dev Services** sobe um container descartável com a imagem de `quarkus.datasource.devservices.image-name`;
não há URL de banco para configurar. Em produção a conexão vem de `DB_URL`, `DB_USERNAME` e `DB_PASSWORD`.
Copie `.env.example` para `.env` para trocar nome, usuário, senha e portas do `compose.yaml`. Adminer:
`docker compose --profile tools up -d` (http://localhost:8081).
<!-- forge:else -->
<!--~ Este projeto foi gerado com `--db none`: sem Hibernate, Panache nem Flyway; a consulta de CEP não é persistida. -->
<!-- forge:end -->

## CI/CD
Workflows em `.github/workflows/` (CI em PR; deploy na `main` para ECS via OIDC). Variáveis, Environments e roles
da AWS: [docs/PIPELINE.md](docs/PIPELINE.md). Terraform: [docs/INFRA.md](docs/INFRA.md).

## Imagem Docker
Build multi-stage com cache do Maven, fast-jar do Quarkus em camadas, JRE Alpine, usuário não-root,
`MaxRAMPercentage=75` e `ExitOnOutOfMemoryError`. A task roda o perfil `prod` e o ALB usa
`/q/health/ready`.

## API
<!-- forge:if db!=none -->
`GET /api/addresses/{cep}` consulta o CEP (ViaCEP) e guarda o resultado no banco.
<!-- forge:else -->
<!--~ `GET /api/addresses/{cep}` consulta o CEP (ViaCEP). -->
<!-- forge:end -->
O cliente é um REST Client (`ViaCepRestClient`) protegido por SmallRye Fault Tolerance no adapter: `@Retry`
(2 tentativas) e `@CircuitBreaker`. Os valores se trocam em `application.properties`, por exemplo
`com.example.starterapi.adapter.out.client.ViaCepAdapter/findByCep/Retry/maxRetries=5`. Com o circuito aberto
ou o provedor fora do ar a API responde `503` em `application/problem+json`.

Health: `/q/health/live` e `/q/health/ready`. OpenAPI e Swagger UI (`/q/swagger-ui`) só em dev e test.
Todo log leva o `correlationId` (cabeçalho `X-Correlation-Id`, repassado ao ViaCEP); em produção os logs saem em JSON.

<!-- forge:if auth!=none -->
## Segurança (JWT)
Tokens JWT verificados pelo SmallRye JWT. Com `quarkus.security.jaxrs.deny-unannotated-endpoints=true` todo
endpoint precisa declarar o acesso (`@Authenticated`, `@RolesAllowed` ou `@PermitAll`); um esquecido responde
`403` em vez de ficar aberto. Os erros 401 (sem token ou token inválido/vencido, com `WWW-Authenticate: Bearer`) e 403
saem em `application/problem+json`, como os demais erros da API (`SecurityExceptionMapper`).

| Perfil | Origem do JWT |
|---|---|
| `dev`, `test` | segredo HS256 de desenvolvimento (`app.security.dev-secret`) + `POST /dev/token` |
| `prod` | JWKS do emissor: `JWT_ISSUER_URI` (+ `/.well-known/jwks.json`, ou `JWT_JWKS_URI`) |

```sh
TOKEN=$(curl -s -X POST "localhost:8080/dev/token?sub=ana&roles=ADMIN,USER&scopes=orders:read" | jq -r .access_token)
curl -H "Authorization: Bearer $TOKEN" localhost:8080/api/me
```

O claim de roles vem de `JWT_ROLES_CLAIM` (padrão `roles`; no Cognito, `cognito:groups`). O `/dev/token` é
removido do build de produção (`@IfBuildProfile`).

### Anotações de acesso
Use as anotações do Quarkus direto: `@Authenticated`, `@RolesAllowed(Roles.ADMIN)` e `@PermitAll`. **Não crie
anotações próprias que apenas empacotem `@RolesAllowed`** (meta-anotação ou `@Stereotype`): o Quarkus as ignora e o
endpoint fica aberto sem nenhum aviso. `CurrentUser.from(identity)` devolve subject, username, roles e scopes
para uso em adapters, e o `SecurityEndpointsTest` garante os 401/403 dos endpoints de exemplo.
<!-- forge:end -->
