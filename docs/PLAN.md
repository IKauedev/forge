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
- Feita antes das Fases 3, 4 e 5. Como `--db` ainda não existe, o `.forge.yaml` não tem `db` (entra na Fase 3).
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

**Critérios de aceite**
- Release `v1.0.0` publicada com binários.
- `forge upgrade --diff` mostra diferenças reais entre duas versões do template (teste com fixture).

---

### Fase 9: Template React (a detalhar em plan mode)

Vite + React + TypeScript (conferir versões atuais em `npm view` e na documentação oficial), ESLint, Vitest,
`config.json` em runtime, interceptor de auth/correlation id, guards por role espelhando as anotações do backend.
Reaproveita `infra/` (S3 + CloudFront com OAC) e workflows da Fase 7. Integra com `forge new --type react-app`.

**Critérios de aceite:** `npm run lint && npm test && npm run build` e `terraform validate` verdes no projeto gerado.

---

### Fase 10: Template Quarkus (a detalhar em plan mode)

Quarkus na linha LTS atual (3.33 na pesquisa de 2026-10-10; confirmar em quarkus.io/releases), hexagonal com CDI,
`quarkus-rest` + REST Client, Hibernate ORM com Panache + Flyway, Dev Services nos testes, SmallRye Fault Tolerance
(circuit breaker nativo), segurança JWT com `@RolesAllowed`/anotações próprias, Dockerfile, compose, ArchUnit.
Pipeline e Terraform reaproveitam os módulos das Fases 4 e 5. Integra com `forge new --type quarkus-api`.

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
