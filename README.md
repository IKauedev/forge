# forge

CLI em Go que gera projetos opinativos: **spring-api** (Spring Boot hexagonal, Feign/WebClient, Flyway, ArchUnit,
ECS Fargate), **quarkus-api** (Quarkus 3.40 LTS hexagonal, Panache, Fault Tolerance, Dev Services, ECS Fargate), **angular-app** (Angular standalone, config em runtime, S3 + CloudFront) e **react-app** (React + Vite
com guards por role que espelham o backend). Todos saem com pipeline GitHub Actions (OIDC) e Terraform.
Plano completo em [docs/PLAN.md](docs/PLAN.md).

## Uso

Rode `forge` sem argumentos, num terminal, para abrir o **menu interativo**: ele mostra o banner, detecta o
contexto (se você está dentro de um projeto do forge, se há versão nova, se o forge já está no PATH) e oferece só
o que faz sentido: criar projeto, ver o que mudou no template, listar templates, verificar o ambiente, atualizar e
instalar. Os comandos abaixo também funcionam direto, inclusive em scripts e no CI (sem terminal o `forge` só
mostra a ajuda).

```sh
forge                                          # menu interativo
forge doctor                                   # confere Java 25, Docker e Git
forge new                                      # formulário interativo
forge new "Payments API" --package com.empresa # só pergunta o que faltar
forge new payments-api --yes                   # sem perguntas (usa os padrões)
forge new orders --type quarkus-api --db mysql   # API Quarkus (LTS 3.40)
forge new storefront --type angular-app        # app Angular + infra S3/CloudFront
forge new webshop --type react-app             # app React (Vite) + infra S3/CloudFront
forge list                                     # templates e opções
forge upgrade --diff                           # o que mudou no template desde que o projeto foi gerado
cd payments-api && ./mvnw verify               # JDK 25 + Docker
```

### Opções de `forge new`

| Flag | Valores | Padrão | Efeito |
|---|---|---|---|
| `--type` | `spring-api`, `quarkus-api`, `angular-app`, `react-app` | `spring-api` | template a gerar |
| `--package` | pacote Java | `com.example` | pacote base e `groupId` (`spring-api` e `quarkus-api`) |
| `--out` | diretório | `.` | pai; o projeto vai para `<out>/<nome-kebab>` |
| `--vpc` | `new`, `existing` | `new` | rede do Terraform: cria a VPC ou usa uma existente (`spring-api` e `quarkus-api`) |
| `--db` | `postgres`, `mysql`, `mariadb`, `none` | `postgres` | banco do catálogo (imagem, driver, Flyway, Testcontainers e migration do dialeto); `none` remove JPA, Flyway e o serviço do compose |
| `--http` | `feign`, `webclient`, `both` | `both` | cliente HTTP do exemplo ViaCEP (com `both`, escolhe por `clients.viacep.impl`) |
| `--auth` | `none`, `jwt`, `cognito` | `jwt` | `none` remove a segurança; `cognito` lê roles de `cognito:groups` (só `spring-api`; no `quarkus-api` use `JWT_ROLES_CLAIM`) |
| `--cache` | `none`, `caffeine`, `redis` | `none` | `@Cacheable` na consulta; `redis` inclui serviço no compose e Testcontainers |
| `--yes`, `-y` | | | não pergunta; usa os padrões |
| `--no-git` | | | não executa `git init` |

Sem terminal (CI, pipes) o forge nunca abre o formulário: usa os padrões e exige o nome como argumento.
O nome aceita `Payments API`, `payments-api`, `PaymentsAPI` ou `payments_api` (máx. 24 caracteres em kebab,
para caber nos limites de nomes da AWS).

`--package`, `--db`, `--vpc`, `--http`, `--auth` e `--cache` valem para `spring-api`; o `quarkus-api` aceita
`--package`, `--db`, `--vpc` e `--auth none|jwt` (sem `--http`, `--cache` nem `cognito`). Usar uma opção que o
`--type` não tem é erro. `angular-app` e `react-app` não têm opções além das comuns.

Os bancos vêm de `catalog/databases.yaml` (`forge list` mostra os disponíveis).
Cada projeto gerado traz um `.forge.yaml` com as respostas (versão do forge, template, nome, pacote e opções).

### `forge doctor`
Obrigatórios (saída com código 1 se faltar): **Java 25+** (respeita `JAVA_HOME`), **Docker**, **Git**.
Opcionais (só avisam): Terraform, AWS CLI, Node.js. Dentro de um projeto, confere o bit de execução do `mvnw`.
Com o daemon do Docker parado o resultado é um aviso. Cores desligam com `NO_COLOR`.

### `forge upgrade --diff`
Dentro de um projeto (ou `forge upgrade <dir> --diff`), regenera o projeto num diretório temporário com as respostas
do `.forge.yaml` e mostra o diff unificado contra o projeto atual. Lista também os arquivos novos no template e os que
só existem no projeto. Nada é sobrescrito; por isso só o modo `--diff` existe.

## Instalação e atualização
```sh
go install github.com/ikauedev/forge/cmd/forge@latest   # ou baixe o binário da release
forge install                                            # instala no seu usuário e põe no PATH
forge update                                             # atualiza para a última release
```
- `forge install` copia o executável para `%LOCALAPPDATA%\Programsorge` (Windows) ou `~/.local/bin` (Linux e
  macOS) e registra a pasta no PATH do **usuário**; não precisa de administrador, preserva o PATH existente e é
  idempotente (rodar de novo atualiza a cópia). Abra um novo terminal depois. `forge uninstall` desfaz tudo.
- `forge update` consulta a última release no GitHub, baixa o binário do seu sistema, **confere o SHA-256** do
  `checksums.txt` e troca o executável (`--check` só informa; `--yes` não pergunta). Não precisa de credencial;
  `GITHUB_TOKEN`/`GH_TOKEN` (ou `gh auth login`) só servem para evitar o limite de requisições anônimas. O menu avisa
  de versão nova no máximo uma vez por dia (`FORGE_NO_UPDATE_CHECK=1` desliga o aviso).
- Homebrew: `brew install --cask ikauedev/tap/forge` (depois que o tap existir; ver Releases).

Binários para linux, macOS e Windows (amd64/arm64) e `checksums.txt` ficam em cada release.

## Pipeline gerada
Todo projeto traz `.github/workflows/` (`ci.yml`, `deploy.yml`, `deploy-env.yml`): CI em PR (verify, Hadolint,
Trivy) e deploy na `main` para o ECS via OIDC, dev automático e prod com aprovação. Sem access keys; as actions
ficam fixadas por SHA. Variáveis e Environments a configurar: `docs/PIPELINE.md` do projeto gerado.

No `angular-app` e no `react-app` o deploy sincroniza o S3 (assets com hash em cache imutável; `index.html` e `config.json` sem
cache), gera o `config.json` do ambiente e invalida o CloudFront.

## Infra gerada (Terraform)
Todo projeto traz `infra/` com `bootstrap/` (bucket de state, OIDC do GitHub e roles de deploy; roda uma vez, com
state local) e `envs/dev` e `envs/prod` (backend S3 com lock nativo, sem DynamoDB). O forge **nunca** executa
`terraform apply`. `spring-api`: ECR, ECS Fargate, ALB, autoscaling e RDS (segredo no Secrets Manager); `angular-app` e `react-app`:
S3 privado + CloudFront com OAC e fallback de SPA. Passo a passo em `docs/INFRA.md` do projeto gerado.

## Testes e CI

O CI do próprio forge (`.github/workflows/ci.yml`) roda `go test` em Linux e Windows, `golangci-lint`, os testes dos
dois templates de referência e um smoke que gera projetos de verdade e valida `actionlint`, `terraform validate` e o
build do Angular. A matriz pesada do Spring (JDK 25 + Docker) roda na `main`, no agendamento semanal e sob demanda.
O Dependabot cobre `go.mod`, Maven, npm, imagens Docker, actions e Terraform dos templates.

Release: `git tag v1.0.0 && git push origin v1.0.0` dispara `.github/workflows/release.yml` (GoReleaser). O tap do
Homebrew precisa do repositório `ikauedev/homebrew-tap` e do secret `HOMEBREW_TAP_TOKEN`.


```sh
go build ./... && go vet ./... && go test ./...
# smoke: gera 5 projetos (matriz representativa), valida o compose, sobe o banco e roda ./mvnw verify
# (JAVA_HOME=JDK 25, Docker ativo)
go test -tags smoke -timeout 60m ./internal/engine
# opcional: valida os workflows (go install github.com/rhysd/actionlint/cmd/actionlint@v1.7.12)
# regenerar os goldens depois de mudar o template
go test ./internal/engine -run TestGolden -update
```

## Como funciona

`templates/spring-api/`, `templates/quarkus-api/`, `templates/angular-app/` e `templates/react-app/` são projetos reais que compilam sozinhos. O `spring-api` compila
no estado de referência (`db=postgres`, `http=both`, `auth=jwt`, `cache=caffeine`, ver `template.forge.yaml`). O
engine copia o template embutido (`go:embed`) e:

1. remove arquivos conforme o manifesto (`remove:`);
2. aplica diretivas em comentários (`forge:if`, `forge:else`, `forge:end`) e destampa ramos alternativos
   escritos como `//~`, `#~` ou `<!--~ -->`; blocos `forge:slot` são trocados por trechos gerados do
   catálogo de bancos (`catalog/databases.yaml`);
3. renomeia conteúdo, arquivos e diretórios pelas sentinelas do template (`sentinel:` no manifesto): `starter-api`,
   `StarterApi`, `starterapi`, `starter_api`, `STARTER_API`, `Starter Api` e `com.example` no `spring-api` e no
   `quarkus-api`;
   as mesmas formas de `starter-app` no `angular-app` e no `react-app`.

O embed do `angular-app` e do `react-app` lista os caminhos explicitamente (`templates/templates.go`) para não incluir `node_modules/`.
O bit de execução do `mvnw` e de `*.sh` é restaurado por nome, pois o `embed.FS` não guarda permissões.
Para adicionar um banco: [docs/ADDING_A_DATABASE.md](docs/ADDING_A_DATABASE.md). Para adicionar uma opção: declare em `options:` do manifesto, use as diretivas nos arquivos e rode os testes
(o de 27 combinações garante que nada sobra e que YAML/XML continuam válidos).

## Licença

MIT (veja [LICENSE](LICENSE)). Os projetos gerados pelo forge não herdam a licença do forge.
