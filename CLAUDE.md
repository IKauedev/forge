# forge

CLI em Go que gera projetos Spring Boot e Angular pré-configurados.
Plano completo em docs/PLAN.md. Execute UMA fase por vez.

## Estrutura
- cmd/forge/            entrypoint
- internal/template/    manifesto do template (opções, remoções)
- internal/project/     .forge.yaml do projeto gerado
- internal/cli/         comandos Cobra
- internal/names/       normalização do nome do projeto (kebab, snake, pascal...)
- internal/engine/      cópia, renomeação por sentinela, render de templates
- internal/catalog/     leitura do catalog/databases.yaml
- internal/doctor/      checagem de pré-requisitos
- templates/spring-api/ projeto de referência Spring (COMPILA sozinho)
- templates/quarkus-api/ projeto de referência Quarkus 3.40 LTS (COMPILA sozinho)
- templates/angular-app/ projeto de referência Angular (COMPILA sozinho)
- templates/react-app/  projeto de referência React + Vite (COMPILA sozinho)
- templates/*/k8s/       manifests Kustomize (base + overlays dev, hml, prd), opção `--k8s`
- catalog/              dados: bancos (databases.yaml) e migrations por dialeto
- testdata/golden/      saídas esperadas

## Comandos
- go build ./... && go vet ./... && go test ./...
- golangci-lint run
- Template Spring: cd templates/spring-api && ./mvnw -q verify
- Template Quarkus: cd templates/quarkus-api && ./mvnw -q verify (JDK 25 + Docker; apague target/ depois)
- Template Angular: cd templates/angular-app && npm ci && npm run lint && npm test -- --watch=false && npm run build
- Template React: cd templates/react-app && npm ci && npm run lint && npm test && npm run build
- Kubernetes (Kustomize): kubectl kustomize templates/<template>/k8s/overlays/<dev|hml|prd>
- Terraform: terraform fmt -check -recursive && terraform init -backend=false && terraform validate

## Regras
- Templates de referência DEVEM compilar e passar nos testes sem passar pelo forge.
- Arquivos de infra variáveis usam delimitadores [[ ]], nunca {{ }}.
- Antes de compilar o forge, garanta que templates/*/target não existe (`./mvnw clean`): o go:embed incluiria os artefatos.
  angular-app e react-app listam caminhos explícitos em `templates/templates.go` (não embutem node_modules/dist); arquivo novo na raiz deles precisa entrar na lista.
- Variações do template usam diretivas em comentários (`forge:if/else/end`) e ramos alternativos comentados com `~`
  (`//~`, `#~`, `<!--~ -->`), nunca `[[ ]]` em arquivos Java. O estado de referência do template é o de
  `reference:` em `template.forge.yaml` e precisa compilar sozinho. Não ponha diretivas dentro do bloco de imports
  (o engine poda imports sem uso) e use nomes totalmente qualificados em código que só existe fora da referência.
- Trechos que dependem do banco ficam em blocos `forge:slot <nome>` com o texto do Postgres (referência); o
  engine os troca por trechos gerados de `internal/catalog/slots.go`. Mudou um lado, mude o outro: o teste
  `TestPostgresSlotsMatchTemplateReference` acusa divergência. Bancos novos: docs/ADDING_A_DATABASE.md.
- Nunca renderizar arquivos binários (jar, png, fontes). Preservar bit de execução (mvnw, scripts).
- Todo nome gerado deve respeitar limites da AWS (ALB/target group 32, bucket 63, role IAM 64).
- Nome de banco, schema e usuário sempre em snake_case minúsculo.
- Não adicionar dependência Go nova sem justificar no commit.
- Antes de fixar versões (Spring Boot, Spring Cloud, Terraform, actions, imagens Docker),
  consulte a documentação oficial atual. Não confie em versões de memória.
- Compatibilidade Spring Boot x Spring Cloud: verificar a tabela oficial.
- Workflows do template: toda action de terceiros fixada por SHA de 40 hex (com a versão em comentário),
  `${{ }}` do GitHub fica intocado, nada de access keys (OIDC). Rode `actionlint` depois de mexer neles.
- Não rodar `terraform apply` nem criar recursos AWS. Apenas validate/plan.
- Commits pequenos, mensagens em Conventional Commits.
- Se uma tarefa estiver ambígua, pergunte antes de implementar.

## Definition of done (toda fase)
- Build, vet, testes e lint verdes.
- Documentação da fase atualizada (README ou docs/).
- Nenhum TODO sem issue ou nota em docs/PLAN.md.
