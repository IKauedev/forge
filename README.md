# forge

CLI em Go que gera projetos Spring Boot opinativos (arquitetura hexagonal, Feign/WebClient, Flyway, ArchUnit).
Plano completo em [docs/PLAN.md](docs/PLAN.md).

## Uso

```sh
forge doctor                                   # confere Java 25, Docker e Git
forge new                                      # formulário interativo
forge new "Payments API" --package com.empresa # só pergunta o que faltar
forge new payments-api --yes                   # sem perguntas (usa os padrões)
forge list                                     # templates e opções
cd payments-api && ./mvnw verify               # JDK 25 + Docker
```

### Opções de `forge new`

| Flag | Valores | Padrão | Efeito |
|---|---|---|---|
| `--package` | pacote Java | `com.example` | pacote base e `groupId` |
| `--out` | diretório | `.` | pai; o projeto vai para `<out>/<nome-kebab>` |
| `--db` | `postgres`, `mysql`, `mariadb`, `none` | `postgres` | banco do catálogo (imagem, driver, Flyway, Testcontainers e migration do dialeto); `none` remove JPA, Flyway e o serviço do compose |
| `--http` | `feign`, `webclient`, `both` | `both` | cliente HTTP do exemplo ViaCEP (com `both`, escolhe por `clients.viacep.impl`) |
| `--auth` | `none`, `jwt`, `cognito` | `jwt` | `none` remove o Spring Security; `cognito` lê roles de `cognito:groups` |
| `--cache` | `none`, `caffeine`, `redis` | `none` | `@Cacheable` na consulta; `redis` inclui serviço no compose e Testcontainers |
| `--yes`, `-y` | | | não pergunta; usa os padrões |
| `--no-git` | | | não executa `git init` |

Sem terminal (CI, pipes) o forge nunca abre o formulário: usa os padrões e exige o nome como argumento.
O nome aceita `Payments API`, `payments-api`, `PaymentsAPI` ou `payments_api` (máx. 24 caracteres em kebab,
para caber nos limites de nomes da AWS).

Os bancos vêm de `catalog/databases.yaml` (`forge list` mostra os disponíveis).
Cada projeto gerado traz um `.forge.yaml` com as respostas (versão do forge, template, nome, pacote e opções).

### `forge doctor`
Obrigatórios (saída com código 1 se faltar): **Java 25+** (respeita `JAVA_HOME`), **Docker**, **Git**.
Opcionais (só avisam): Terraform, AWS CLI, Node.js. Dentro de um projeto, confere o bit de execução do `mvnw`.
Com o daemon do Docker parado o resultado é um aviso. Cores desligam com `NO_COLOR`.

## Testes

```sh
go build ./... && go vet ./... && go test ./...
# smoke: gera 5 projetos (matriz representativa), valida o compose, sobe o banco e roda ./mvnw verify
# (JAVA_HOME=JDK 25, Docker ativo)
go test -tags smoke -timeout 60m ./internal/engine
# regenerar os goldens depois de mudar o template
go test ./internal/engine -run TestGolden -update
```

## Como funciona

`templates/spring-api/` é um projeto real que compila sozinho no estado de referência
(`db=postgres`, `http=both`, `auth=jwt`, `cache=caffeine`, ver `template.forge.yaml`). O engine copia o template embutido
(`go:embed`) e:

1. remove arquivos conforme o manifesto (`remove:`);
2. aplica diretivas em comentários (`forge:if`, `forge:else`, `forge:end`) e destampa ramos alternativos
   escritos como `//~`, `#~` ou `<!--~ -->`; blocos `forge:slot` são trocados por trechos gerados do
   catálogo de bancos (`catalog/databases.yaml`);
3. renomeia conteúdo, arquivos e diretórios pelas sentinelas (`starter-api`, `StarterApi`, `starterapi`,
   `starter_api`, `STARTER_API`, `Starter Api`, `com.example`).

O bit de execução do `mvnw` e de `*.sh` é restaurado por nome, pois o `embed.FS` não guarda permissões.
Para adicionar um banco: [docs/ADDING_A_DATABASE.md](docs/ADDING_A_DATABASE.md). Para adicionar uma opção: declare em `options:` do manifesto, use as diretivas nos arquivos e rode os testes
(o de 27 combinações garante que nada sobra e que YAML/XML continuam válidos).
