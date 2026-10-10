# forge

CLI em Go que gera projetos Spring Boot opinativos (arquitetura hexagonal, Feign/WebClient, Flyway, ArchUnit).
Plano completo em [docs/PLAN.md](docs/PLAN.md).

## Uso

```sh
go run ./cmd/forge new "Payments API" --package com.empresa --out ~/projetos
cd ~/projetos/payments-api && ./mvnw verify   # JDK 25 + Docker
```

| Flag | Padrão | Descrição |
|---|---|---|
| `--package` | `com.example` | pacote base Java |
| `--out` | `.` | diretório pai; o projeto vai para `<out>/<nome-kebab>` |
| `--no-git` | `false` | não executa `git init` |

O nome aceita `Payments API`, `payments-api`, `PaymentsAPI` ou `payments_api` (máx. 24 caracteres
em kebab, para caber nos limites de nomes da AWS).

## Testes

```sh
go build ./... && go vet ./... && go test ./...
# smoke: gera um projeto e roda ./mvnw verify (JAVA_HOME=JDK 25, Docker ativo)
go test -tags smoke -timeout 20m ./internal/engine
# regenerar o golden depois de mudar o template
go test ./internal/engine -run TestGolden -update
```

## Como funciona

`templates/spring-api/` é um projeto real que compila sozinho. O engine copia o template embutido
(`go:embed`) renomeando conteúdo, arquivos e diretórios pelas sentinelas (`starter-api`, `StarterApi`,
`starterapi`, `starter_api`, `STARTER_API`, `Starter Api`, `com.example`). O bit de execução do `mvnw`
e de `*.sh` é restaurado por nome, pois o `embed.FS` não guarda permissões.
