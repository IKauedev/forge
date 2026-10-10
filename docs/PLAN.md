# forge: plano de ação para Claude Code

CLI em Go que gera projetos **Spring Boot** (e depois **Angular**) já opinativos: arquitetura hexagonal, Feign/WebClient, banco escolhível, Docker, docker-compose, pipeline GitHub Actions (OIDC) e Terraform para AWS.

> Regra de ouro: uma fase = uma sessão = um commit verde. Não misture fases.
> Para cada fase: plan mode primeiro, colar o prompt, só avançar com os critérios de aceite verdes, commit `feat(fase-N): ...` e `/clear`.

---

## 2. Decisões fixas (não renegociar durante a execução)

| Decisão | Escolha |
|---|---|
| Linguagem da CLI | Go (versão estável mais recente), módulo `github.com/<usuario>/forge` |
| CLI framework | Cobra; modo interativo com `charmbracelet/huh` |
| Templates | **Projetos de referência que compilam**, embutidos via `go:embed` |
| Variação entre projetos | Renomeação por **sentinelas** + templates só nos arquivos de infra |
| Delimitadores de template | `[[ ]]` (evita colisão com `${{ }}` do GitHub Actions) |
| Arquitetura Spring | Hexagonal (ports and adapters), validada com ArchUnit |
| Banco | Catálogo orientado a dados (`catalog/databases.yaml`) |
| Deploy padrão | Spring: ECS Fargate. Angular: S3 + CloudFront (OAC) |
| Autenticação CI → AWS | OIDC, **nunca** access keys em secrets |
| Escopo do MVP | Postgres, MySQL, MariaDB + ECS. O resto vem depois |

Sentinelas do projeto de referência (cada formato do nome tem a sua):

| Formato | Sentinela |
|---|---|
| kebab | `starter-api` |
| Pascal | `StarterApi` |
| flat (pacote) | `starterapi` |
| snake | `starter_api` |
| UPPER_SNAKE | `STARTER_API` |
| Title | `Starter Api` |

Pacote base de referência: `com.example.starterapi`.

---

## 4. Ajustes recomendados no Claude Code (opcionais)

- **Permissões**: liberar `go`, `./mvnw`, `npm`, `terraform fmt/validate`, `docker compose config` e `git` (exceto push). Bloquear `terraform apply/destroy` e `aws`.
- **Hook pós-edição**: `gofmt`/`goimports` em `.go` e `terraform fmt` em `.tf`.
- **Subagentes**: revisor (diff da fase vs. critérios de aceite) e verificador de versões (documentação oficial).
- **Slash commands** em `.claude/commands/`: `/fase N` e `/verificar`.

---

## 5. Fases

### Fase 0: Projeto de referência Spring (sem Go ainda)

**Objetivo:** um projeto Spring Boot real, hexagonal, que compila e testa.

**Prompt:**

```text
Leia CLAUDE.md e docs/PLAN.md (fase 0). Entre em modo plano primeiro.

Crie templates/spring-api/ como um projeto Spring Boot Maven completo, usando as
sentinelas: artifact starter-api, pacote com.example.starterapi, classe StarterApiApplication.

Requisitos:
- Java LTS atual, Maven Wrapper.
- Estrutura hexagonal: config/, domain/{model,port/in,port/out,service}, application/,
  adapter/in/web, adapter/out/{client,persistence}, shared/.
- Módulo de exemplo "address" (consulta de CEP via ViaCEP): porta AddressLookupPort,
  adapter Feign e adapter WebClient (escolha por propriedade clients.viacep.impl),
  DTO externo separado do domínio, mapper com MapStruct.
- Resilience4j (circuit breaker + retry) configurado por YAML.
- Persistência: JPA + Flyway + Postgres, com uma migration inicial mínima.
- @ControllerAdvice com ProblemDetail, correlation id (filtro + MDC + propagação nos clientes),
  logs JSON, Actuator com probes, graceful shutdown, springdoc-openapi só no perfil dev.
- Perfis: local, test, dev, prod.
- Testes: unitário do domínio, WebMvcTest, Testcontainers com @ServiceConnection,
  WireMock para os clientes HTTP, e ArchUnit garantindo que domain não depende de adapter nem de Spring.
- Spotless e JaCoCo no build.

Antes de fixar versões, verifique a compatibilidade Spring Boot x Spring Cloud OpenFeign na
documentação oficial. Mostre o plano e espere minha aprovação antes de implementar.
```

**Notas de execução (2026-10-10)**
- Versões fixadas: Spring Boot 4.1.1, Spring Cloud 2025.1.3, Java 25, MapStruct 1.6.3, Testcontainers 2.0.5 (BOM do Boot), WireMock 3.13.2.
- Desvio: não existe starter Resilience4j para Spring Boot 4. O template usa o resilience nativo do Spring Framework 7 (`@Retryable` + `@EnableResilientMethods`, configurado por propriedades `clients.viacep.retry.*`). **Circuit breaker ficou de fora**; revisitar quando houver suporte oficial.

- Segurança: `spring-boot-starter-security-oauth2-resource-server`, JWT HS256 só em `local`/`test`, `issuer-uri` em `dev`/`prod`, meta-anotações em `shared.security` (decisão de 2026-10-10). O `--auth none|cognito` da Fase 6 vira variação desta base.

**Critérios de aceite**
- `./mvnw -q verify` verde (testes, Spotless, ArchUnit, JaCoCo).
- Trocar `clients.viacep.impl` entre `feign` e `webclient` funciona (teste cobrindo ambos).
- Nenhuma classe de `domain/` importa Spring.

---

### Fase 1: Docker, compose e Makefile

**Objetivo:** rodar o template localmente em container.

**Prompt:**

```text
Fase 1 do docs/PLAN.md. Modo plano primeiro.

Em templates/spring-api/ adicione:
- Dockerfile multi-stage (build com cache de Maven, extração de camadas do jar,
  runtime JRE, usuário não-root, MaxRAMPercentage, ExitOnOutOfMemoryError).
  Confira a sintaxe atual de extração de camadas do Spring Boot na documentação.
- .dockerignore.
- compose.yaml SOMENTE com infraestrutura (Postgres com healthcheck, volume, nomes derivados do
  projeto; Adminer em profile "tools").
- compose.full.yaml que adiciona a aplicação com depends_on: service_healthy.
- Integração spring-boot-docker-compose para o perfil local.
- .env.example (e .env no .gitignore).
- Makefile: up, down, run, test, image, full.
- README do template explicando como rodar.

Valide: docker compose config, build da imagem e subida do compose.full com healthcheck da app.
```

**Critérios de aceite**
- `make run` sobe o banco e a aplicação.
- `docker compose -f compose.yaml -f compose.full.yaml up` fica saudável.
- Imagem roda como usuário não-root.

---

### Fase 2: Engine Go e `forge new`

**Objetivo:** gerar o projeto com outro nome a partir do template embutido.

**Prompt:**

```text
Fase 2 do docs/PLAN.md. Modo plano primeiro.

Crie o módulo Go e implemente:
1. internal/names: NewNames(raw) -> {Kebab, Snake, Pascal, Flat, Upper, Title}, com validação
   (começa com letra, sem palavra reservada Java no pacote, limites AWS). Testes table-driven
   cobrindo "Payments API", "payments-api", "PaymentsAPI", "payments_api", entradas inválidas.
2. internal/engine: copia o template (go:embed) para o destino, renomeia conteúdo, arquivos e
   diretórios pelas sentinelas, ignora binários, preserva modo de arquivo (mvnw).
   A ordem importa: substitua primeiro as sentinelas mais longas/específicas.
3. internal/cli: comando `forge new <nome> --package com.empresa --out <dir>` com Cobra.
4. Pós-geração: git init e mensagem com próximos passos.
5. Teste golden: gerar "Payments API" e comparar com testdata/golden/.
6. Teste de fumaça (build tag "smoke"): gerar um projeto e rodar ./mvnw -q verify nele.

Não implemente banco escolhível nem modo interativo ainda.
```

**Critérios de aceite**
- `forge new "Payments API"` gera projeto que compila **sem edição manual**.
- Não sobra a string `starter` em nenhum arquivo (teste automatizado com grep).
- `go test ./...` e o smoke test verdes.

---

### Fase 3: Catálogo de bancos

**Objetivo:** escolher `--db postgres|mysql|mariadb|none` e ter tudo configurado.

**Prompt:**

```text
Fase 3 do docs/PLAN.md. Modo plano primeiro.

Implemente o catálogo de bancos orientado a dados:
- catalog/databases.yaml com postgres, mysql, mariadb (imagem fixada, porta, esquema JDBC,
  coordenadas Maven do driver, módulo Flyway, módulo Testcontainers, healthcheck, variáveis de
  ambiente, mapeamento AWS).
- Flag `forge new ... --db`.
- Renderização (delimitadores [[ ]]) de: pom.xml (driver, flyway, testcontainers),
  compose.yaml, application-*.yaml, migration inicial do dialeto escolhido (uma por dialeto,
  mínima). Não configurar dialect do Hibernate (autodetecção).
- --db none remove JPA, Flyway, adapters de persistência e o serviço de banco do compose.
- Verifique as versões atuais: módulos do Flyway por banco, imagens oficiais, driver MySQL.

Smoke test em matriz (postgres, mysql, mariadb, none): gerar, ./mvnw verify com Testcontainers
do banco correspondente, e docker compose config.
```

**Notas de execução (2026-10-10)**
- Imagens: `postgres:17-alpine`, `mysql:8.4` (LTS), `mariadb:11.8` (LTS; `11.4` não existe como tag). Drivers e Flyway vêm do BOM do Boot 4.1.1.
- O catálogo gera trechos (`forge:slot`) para pom, compose e `TestcontainersConfig`; o template de referência continua em Postgres. Serviço do compose renomeado para `db` e variáveis do `.env` neutras (`DB_*`).
- `--db none` remove JPA, Flyway, persistência, serviço do compose (e o `compose.yaml` inteiro quando não há Redis) e liga um repositório sem persistência.
- Mapeamento AWS (`aws:` no catálogo) fica para a Fase 5; confirmar lá a versão do RDS para MariaDB 11.8.

**Critérios de aceite**
- Os 4 valores de `--db` geram projetos que passam em `./mvnw verify`.
- Adicionar um banco novo exige só um bloco no YAML e uma migration (documentar em `docs/ADDING_A_DATABASE.md`).

---

### Fase 4: Pipeline GitHub Actions (Spring → AWS)

**Prompt:**

```text
Fase 4 do docs/PLAN.md. Modo plano primeiro.

Gere .github/workflows/ no template Spring:
- ci.yml (pull_request): build, testes, Spotless, ArchUnit, JaCoCo, Hadolint no Dockerfile,
  Trivy na imagem.
- deploy.yml: push na main -> build da imagem -> push no ECR -> deploy no ECS (ambiente dev);
  promoção para prod via GitHub Environment com aprovação manual.
- Autenticação por OIDC (aws-actions/configure-aws-credentials assumindo role). Sem access keys.
- permissions mínimas (id-token: write, contents: read), concurrency por ambiente.
- Task definition renderizada e amazon-ecs-deploy-task-definition com wait-for-service-stability.
- Variáveis do forge com [[ ]] e ${{ }} do GitHub intactos.
- Fixe as actions por versão estável atual (idealmente SHA). Confira as versões atuais.

Valide com actionlint. Documente em docs/ os secrets/variables que o usuário precisa configurar.
```

**Notas de execução (2026-10-10)**
- Três workflows: `ci.yml`, `deploy.yml` e o reutilizável `deploy-env.yml`. A task definition vem do ECS (`describe-task-definition`) e só a imagem é trocada, então não há JSON no repositório. Convenção de nomes `<app>-<ambiente>` que a Fase 5 deve seguir (ECR `<app>`, cluster/serviço/família `<app>-<env>`, container `<app>`).
- Variáveis (não secrets): `AWS_REGION`, `AWS_ROLE_ARN` por Environment; opcionais `ECR_REPOSITORY`, `ECS_CLUSTER`, `ECS_SERVICE`, `ECS_TASK_FAMILY`. Documentadas em `docs/PIPELINE.md` do projeto.
- O Trivy local reprovou a imagem: Tomcat 11.0.24, Jackson 2.21.5 e Jackson 3.1.5 (HIGH/CRITICAL com correção). O pom sobrescreve `tomcat.version`, `jackson-2-bom.version` e `jackson-bom.version` até o Boot 4.1.x incorporar as correções.
- Validado: `actionlint` limpo, Hadolint limpo (só aviso informativo DL3066) e Trivy limpo na imagem. Não foi feito deploy real nem usadas credenciais AWS.

**Critérios de aceite**
- `actionlint` sem erros no projeto gerado.
- Nenhuma credencial estática nos workflows.

---

### Fase 5: Terraform do template Spring

**Prompt:**

```text
Fase 5 do docs/PLAN.md. Modo plano primeiro.

Em templates/spring-api/infra/ crie:
- bootstrap/: bucket de state, OIDC provider do GitHub, role de deploy (trust policy restrita ao
  repositório e à branch/ambiente). Roda uma vez com state local.
- modules/ecr, modules/ecs-service (cluster, task, service, ALB, logs, autoscaling básico),
  modules/database (RDS condicional ao --db, subnets privadas, SG só do ECS,
  manage_master_user_password = true, segredo injetado via bloco secrets da task),
  modules/network (opção --vpc new|existing).
- envs/dev e envs/prod com backend remoto (confira a opção atual de lock nativo no S3 antes
  de decidir entre S3 lock e DynamoDB).
- dev: instância pequena. prod: Multi-AZ, backup, deletion_protection.
- Nomes derivados do projeto, respeitando limites da AWS.

Valide: terraform fmt, init -backend=false e validate em cada diretório; tflint e checkov se
disponíveis. NÃO execute apply.
```

**Notas da execução**
- Nova opção `--vpc new|existing` (padrão `new`); `modules/network` só existe com `new` e `modules/database` só com `--db != none`. Slots `db-aws` vêm de `aws:` no catálogo.
- Backend S3 com `use_lockfile = true` (lock nativo), sem DynamoDB. Lockfiles `.terraform.lock.hcl` versionados; `.terraform/` ignorado (o `go:embed` o incluiria).
- Validado com Terraform 1.15.8: `fmt`, `init -backend=false` e `validate` em bootstrap, dev e prod. `tflint` e `checkov` não estão instalados nesta máquina, então não rodaram (configs em `infra/.tflint.hcl` e `infra/.checkov.yaml`). Nenhum `apply`.
- Doc do projeto gerado: `docs/INFRA.md`.

**Critérios de aceite**
- `terraform validate` verde em `bootstrap`, `dev` e `prod`, nas variações de `--db`.
- Nenhuma senha em variável ou arquivo de texto.

---

### Fase 6: Interativo, flags extras e `forge doctor`

**Prompt:**

```text
Fase 6 do docs/PLAN.md. Modo plano primeiro.

- Modo interativo com huh quando faltarem flags; `--yes` roda sem perguntas (para CI).
- Flags: --http feign|webclient|both, --auth none|jwt|cognito, --cache none|caffeine|redis.
- `forge list` (templates e opções) e `forge doctor` (java, maven wrapper, docker, terraform,
  aws cli, node: versões mínimas, saída colorida e exit code != 0 se faltar algo essencial).
- Escrever .forge.yaml no projeto gerado (versão do forge, respostas, features).
- Atualizar a matriz de smoke test (escolher combinações representativas, não todas).
```

**Notas de execução (2026-10-10)**
- Feita antes das Fases 3, 4 e 5.
- Variação por diretivas em comentários + manifesto `template.forge.yaml` (ver README). `--auth cognito` = jwt com roles de `cognito:groups`; `--cache redis` inclui Redis no compose e nos testes.
- Padrões de `forge new`: `http=both`, `auth=jwt`, `cache=none`. Sem terminal, nunca pergunta (usa os padrões).
- Mínimos do doctor: Java 25, Terraform 1.10, Node 22; Terraform e Node a reconfirmar nas Fases 5 e 7.
- Matriz de smoke: `both/jwt/none`, `feign/none/none`, `webclient/cognito/caffeine`, `feign/jwt/redis`.

**Critérios de aceite**
- Fluxo interativo e não interativo geram saídas idênticas para as mesmas respostas (teste golden).
- `forge doctor` reporta corretamente um ambiente sem Docker (teste com PATH manipulado).

---

### Fase 7: Template Angular + pré-Terraform

**Prompt:**

```text
Fase 7 do docs/PLAN.md. Modo plano primeiro.

Crie templates/angular-app/ (sentinela starter-app) com a versão estável atual do Angular:
- Standalone components, rotas lazy, signals onde fizer sentido.
- core/ (interceptors de auth, erro e correlation id, guards, ConfigService), shared/, layout/,
  features/example/.
- Configuração em runtime: config.json carregado no bootstrap (APP_INITIALIZER ou equivalente
  atual), sem depender de environment.ts por ambiente.
- ESLint, Prettier, Husky, testes de serviço e de componente.
- infra/: bootstrap (state, OIDC, role), modules/static-site (S3 privado + CloudFront com OAC,
  fallback de SPA 403/404 -> /index.html, ACM em us-east-1 via provider alias, response headers
  policy), envs/dev e envs/prod.
- Workflows: ci.yml (npm ci, lint, test, build) e deploy.yml (s3 sync com cache imutável nos
  assets com hash, no-cache em index.html e config.json, invalidação do CloudFront, OIDC).
- Dockerfile opcional (build Node, serve Nginx) e compose para dev.

Valide: lint, test, build, terraform validate, actionlint. Integre ao `forge new --type angular-app`.
```

**Notas da execução**
- Angular 22.2.2 (versão estável do `npm view @angular/cli`; exige Node `^22.22.3 || ^24.15.0 || >=26`). O runner de testes do CLI é o Vitest. TypeScript 6, ESLint 10 com `angular-eslint`, Husky 9 + lint-staged, Prettier.
- Configuração em runtime: `provideAppInitializer` carrega `config.json` (`apiBaseUrl`, `auth`). O deploy gera o arquivo do ambiente a partir das variáveis do Environment; a imagem Docker gera a partir de `API_BASE_URL`/`AUTH_*`.
- Infra: `bootstrap` (state com lock nativo, OIDC, roles por Environment), `modules/static-site` (S3 privado + CloudFront com OAC, 403/404 → `/index.html` com 200, headers de segurança, ACM em us-east-1 via alias, Route 53 opcional) e `envs/dev|prod`. A role de deploy invalida `distribution/*` da conta porque o ID só existe após o apply (documentado em `docs/INFRA.md`).
- Workflows: `ci.yml`, `deploy.yml`, `deploy-env.yml`; actions fixadas por SHA (`setup-node` v7.1.0 e `download-artifact` v8.0.2 conferidos pela API do GitHub). `actionlint` limpo (via Docker).
- Integração: `forge new --type spring-api|angular-app`. O manifesto ganhou `sentinel`, `java_package` e `next_steps`; as flags de opção são a união dos templates e usar uma que não pertence ao `--type` é erro.
- `go:embed` do `angular-app` com caminhos explícitos para não incluir `node_modules/`; `TestAngularTemplateIsEmbedded` acusa arquivo novo fora da lista.
- Validado no projeto gerado (`forge new "Store Front" --type angular-app`): `npm ci`, lint, format, 11 testes e `ng build`; `terraform fmt`, `init -backend=false` e `validate` em bootstrap, dev e prod. A imagem Docker foi construída e testada (config em runtime e fallback de SPA em rota profunda → 200).
- Não executado: `tflint`/`checkov` (não instalados) e `terraform apply` (proibido).
- Pendências conhecidas: `golangci-lint` reporta 6 `errcheck` em `fmt.Fprint*` (`doctor`, `list`) que já existiam antes desta fase; não há `.golangci.yml`. O template Angular não traz fluxo de login (só `AuthService.setToken`); depende do provedor de identidade escolhido.

**Critérios de aceite**
- `ng build` e `terraform validate` verdes no projeto gerado.
- Refresh em rota profunda funciona (fallback de SPA presente no módulo).

---

### Fase 8: Release e qualidade

**Prompt:**

```text
Fase 8 do docs/PLAN.md. Modo plano primeiro.

- CI do próprio forge: go test, golangci-lint, matriz de smoke tests (gera e valida projetos).
- Renovate ou Dependabot cobrindo go.mod, Maven, npm, imagens Docker, actions e Terraform dos templates.
- GoReleaser: binários linux/darwin/windows (amd64/arm64), checksums, Homebrew tap, `go install`.
- README com GIF do fluxo (VHS ou asciinema), tabela de opções, arquitetura gerada e exemplo de
  repositório criado pelo forge com pipeline verde.
- `forge upgrade --diff`: regenera em diretório temporário com as respostas do .forge.yaml e
  mostra o diff contra o projeto atual, sem sobrescrever nada.
```

**Notas da execução**
- `forge upgrade --diff` (pacote `internal/upgrade`): regenera com as respostas do `.forge.yaml` num diretório temporário e compara. Diff unificado próprio (LCS, contexto de 3 linhas), sem dependência Go nova. Ignora `node_modules`, `target`, `dist`, `.angular` e `.terraform`. Testado com fixture de duas versões de template e pelo CLI.
- `.github/workflows/ci.yml` (Go em Linux e Windows, golangci-lint v2.14.0, templates, projetos gerados e matriz smoke do Spring), `release.yml` e `.github/dependabot.yml`. Actions fixadas por SHA conferido na API do GitHub; `actionlint` limpo. O `golangci-lint` passou a retornar 0 problemas (corrigidos os `errcheck`).
- `.goreleaser.yaml` (v2): linux/darwin/windows × amd64/arm64, checksums SHA-256 e `homebrew_casks` (`brews` está obsoleto). Validado com `goreleaser check` v2.18.3 num repo temporário com remote fictício; os 6 alvos compilam.
- Fora do que dá para fazer aqui: o repositório não tem `remote`, então a `v1.0.0` **não foi publicada**; o tap `ikauedev/homebrew-tap`, o secret `HOMEBREW_TAP_TOKEN` e um arquivo LICENSE (a licença é decisão sua; não declarei nenhuma) ainda precisam existir.
- O GIF do README (VHS/asciinema) não foi gerado e o exemplo de repositório com pipeline verde depende do repositório publicado.

**Critérios de aceite**
- Release `v1.0.0` publicada com binários.
- `forge upgrade --diff` mostra diferenças reais entre duas versões do template (teste com fixture).

---

### Fase 9: Template React (a detalhar em plan mode)

Vite + React + TypeScript (conferir versões atuais em `npm view` e na documentação oficial), ESLint, Vitest,
`config.json` em runtime, interceptor de auth/correlation id, guards por role espelhando as anotações do backend.
Reaproveita `infra/` (S3 + CloudFront com OAC) e workflows da Fase 7. Integra com `forge new --type react-app`.

**Notas da execução**
- Versões conferidas com `npm view`: Vite 8.3, React 19.3, React Router 8.4, Vitest 5, ESLint 10, TypeScript `~6.0.2`. O TypeScript 7.0 já existe, mas o `typescript-eslint` só aceita `<6.1`, então ficou no 6.0 (mesmo do scaffold oficial do Vite). O scaffold atual do Vite usa oxlint; mantive ESLint, como o plano pedia.
- Sem biblioteca de HTTP ou de estado: `apiFetch` usa `fetch` e faz o que o interceptor do Angular faz (id de correlação, Bearer, 401 → limpa token e vai ao login, `ApiError`). Config em runtime: `loadConfig()` roda antes de `createRoot`.
- Guards espelham o backend: `RequireAuth` (`@IsAuthenticated`), `RequireRole` (`@HasRole`), `RequireAnyRole` (`@HasAnyRole`) e `RequireAdmin` (`@IsAdmin`). As roles vêm do claim `auth.rolesClaim` do JWT (`roles`; `cognito:groups` no Cognito), normalizadas como no backend (sem `ROLE_`, maiúsculas). O token é lido sem validar a assinatura: só decide o que mostrar; a autorização é do backend.
- `infra/`, workflows, Docker e docs reaproveitam os do `angular-app` (S3 privado + CloudFront com OAC, fallback de SPA, OIDC). Diferenças: build em `dist/`, favicon SVG, `npm test` (Vitest) e a variável `AUTH_ROLES_CLAIM`.
- Achado que também vale para o Angular: o `.forge.yaml` gravado pelo forge reprovava o `prettier --check` do projeto gerado; os dois templates passaram a ignorá-lo no `.prettierignore` (teste `TestFrontendsIgnoreForgeYamlInPrettier`).
- Validado no projeto gerado (`forge new "Store Front" --type react-app`): `npm ci`, lint, `format:check`, 22 testes e `vite build`; `terraform fmt` e `validate` em bootstrap, dev e prod; `actionlint` limpo; imagem Docker construída (config em runtime com `AUTH_ROLES_CLAIM` e fallback de SPA → 200).
- CI do forge e Dependabot passaram a cobrir o `react-app`. `tflint`/`checkov` continuam sem rodar (não instalados).

**Critérios de aceite:** `npm run lint && npm test && npm run build` e `terraform validate` verdes no projeto gerado.

---

### Fase 10: Template Quarkus (a detalhar em plan mode)

Quarkus na linha LTS atual (3.33 na pesquisa de 2026-10-10; confirmar em quarkus.io/releases), hexagonal com CDI,
`quarkus-rest` + REST Client, Hibernate ORM com Panache + Flyway, Dev Services nos testes, SmallRye Fault Tolerance
(circuit breaker nativo), segurança JWT com `@RolesAllowed`/anotações próprias, Dockerfile, compose, ArchUnit.
Pipeline e Terraform reaproveitam os módulos das Fases 4 e 5. Integra com `forge new --type quarkus-api`.

**Notas da execução**
- **Versão**: a LTS atual é a **3.40** (3.40.1, 2026-09-30, suporte até 2027-09-30; Java 17/21/25, recomendado 25), conferida em quarkus.io/releases e na API do code.quarkus.io. A 3.33 citada acima já não é a LTS.
- **Escopo escolhido**: `--db` (catálogo: postgres, mysql, mariadb, none), `--vpc new|existing` e `--auth none|jwt`. Sem `--http` (só REST Client) e sem `--cache`; sem `cognito` (use `JWT_ROLES_CLAIM=cognito:groups`).
- **Catálogo**: nenhum campo novo. O `db-kind` do Quarkus é o `jdbc_scheme` e a extensão é `quarkus-jdbc-<scheme>`; entraram só os slots `qk-db-deps` e `qk-db-config`. Compose, migrations e o slot AWS são os mesmos do Spring.
- **Arquitetura**: domínio idêntico ao do Spring (Java puro); `DomainConfig` registra o serviço de domínio com um produtor CDI (`@Typed`, para não competir com `AddressApplicationService`); adapters JAX-RS, REST Client (`ViaCepRestClient`) + `ViaCepAdapter` com `@Retry` e `@CircuitBreaker`, Panache, erros em `application/problem+json`.
- **Achado de segurança**: anotações próprias que agregam `@RolesAllowed` (meta-anotação ou `@Stereotype`) são **ignoradas** pelo Quarkus e deixam o endpoint aberto (anônimo recebeu 200). Por isso o template usa `@Authenticated`/`@RolesAllowed` direto, liga `deny-unannotated-endpoints` e tem testes de 401/403 (`SecurityEndpointsTest`). O PLAN previa "anotações próprias".
- **JWT**: produção verifica pelo JWKS do emissor (`JWT_ISSUER_URI/.well-known/jwks.json` ou `JWT_JWKS_URI`); dev/test usam HS256 por JWK simétrica (`smallrye.jwt.verify.algorithm=HS256`) com o `POST /dev/token`, que o `@IfBuildProfile` remove do build de produção. Nenhuma chave privada no repositório (teste).
- **Infra e pipeline**: reaproveitam os do Spring (ECS Fargate, RDS, OIDC), sem cache/Redis, `QUARKUS_PROFILE=prod` e health em `/q/health/ready`.
- **Validação**: `./mvnw verify` verde no template (32 testes) e nas variantes geradas mysql+jwt (32), mariadb+none (21), none+jwt (29) e none+none (18), todas com Dev Services; compose completo subiu em produção (JSON nos logs, readiness com banco, `/dev/token` e OpenAPI ausentes, cabeçalho de correlação até no 401); `terraform fmt/validate`, `actionlint` e Hadolint (só o aviso informativo DL3066, igual ao Spring) limpos.
- **Bug achado e corrigido nos dois Java**: o Javadoc de `LookupAddressUseCase` citava nomes totalmente qualificados que o google-java-format reflui conforme o tamanho do nome do projeto, e o `spotless:check` falhava em projetos gerados com nomes longos. Agora usa nomes simples (goldens do Spring atualizados).
- **Pendência**: o scan do Trivy na imagem Quarkus depende de baixar o banco de vulnerabilidades; ver o resultado na seção de resumo.

**Critérios de aceite:** `./mvnw verify` verde no template e no projeto gerado; `forge new --type` aceita os 4 tipos.

---

## 6. Backlog pós-v1.0

- MongoDB como segundo "perfil de persistência" (sem Flyway, Spring Data MongoDB, DocumentDB com aviso de compatibilidade parcial).
- `--messaging sqs` com LocalStack no compose.
- `--deploy eks` com manifests ou Helm chart.
- Módulos `forge add`: `observability` (OpenTelemetry), `shedlock`, `security-scan` (Checkov, SBOM).
- Templates customizados via repositório Git (`forge new --template github.com/org/repo`).
- Merge em três vias no `forge upgrade`.

---

## 7. Armadilhas e prevenção

| Armadilha | Prevenção |
|---|---|
| Fixar versões de memória | Regra no `CLAUDE.md` + "verifique na documentação oficial" em cada fase |
| Template que não compila | Critério de aceite sempre inclui build do template **antes** do forge tocar nele |
| Misturar fases | Uma sessão por fase, `/clear` entre elas, commit verde antes de avançar |
| Substituir sentinela errada (`starter` dentro de outra palavra) | Sentinelas distintas por formato e teste que procura sobras |
| Rodar `apply` por engano | Bloquear `terraform apply/destroy` nas permissões |
| Explosão de combinações | Matriz de smoke com combinações representativas |
| Quebrar bit de execução do `mvnw` | Teste que verifica o modo do arquivo no projeto gerado |

---

## 8. Checklist final de v1.0

- [ ] `forge new "Meu Projeto" --db postgres --deploy ecs` gera projeto que compila e passa em `verify`
- [ ] Todas as combinações da matriz verdes no CI
- [ ] `actionlint`, `terraform validate` e `tflint` verdes nos projetos gerados
- [ ] Nenhuma credencial estática em lugar nenhum
- [ ] README com GIF, instalação por `brew` e `go install`
- [ ] Renovate/Dependabot ativo
- [ ] Um repositório de exemplo gerado pela ferramenta com pipeline verde
