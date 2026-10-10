# Slim Service

API Spring Boot com arquitetura hexagonal (ports and adapters).

## Pré-requisitos
JDK 25 e Docker. Windows, Linux e macOS (o `Makefile` usa `mvnw.cmd` no Windows).

## Comandos

| Comando | O que faz |
|---|---|
| `make run` | roda a app com o perfil `local` |
| `make test` | `mvnw verify`: testes, Spotless, ArchUnit, JaCoCo |
| `make image` | constrói a imagem Docker `slim-service:local` |
| `make full` | infraestrutura + app em containers (`compose.full.yaml`) |

Este projeto foi gerado com `--db none`: sem banco, JPA nem Flyway; a consulta de CEP não é persistida.

## CI/CD
Workflows em `.github/workflows/` (CI em PR; deploy na `main` para ECS via OIDC). Variáveis, Environments e roles
da AWS: [docs/PIPELINE.md](docs/PIPELINE.md).

## Imagem Docker
Build multi-stage com cache do Maven, jar extraído em camadas (`jarmode=tools`), JRE Alpine,
usuário não-root, `MaxRAMPercentage=75` e `ExitOnOutOfMemoryError`.

## API
`GET /api/addresses/{cep}` consulta o CEP (ViaCEP).
O cliente HTTP é escolhido por `clients.viacep.impl` (`feign` | `webclient`).
Readiness: `/actuator/health/readiness`. Swagger só com o perfil `dev`.


