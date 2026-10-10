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
