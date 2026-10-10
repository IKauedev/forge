# forge

CLI em Go que gera projetos Spring Boot e Angular pré-configurados.
Plano completo em docs/PLAN.md. Execute UMA fase por vez.

## Estrutura
- cmd/forge/            entrypoint
- internal/cli/         comandos Cobra
- internal/names/       normalização do nome do projeto (kebab, snake, pascal...)
- internal/engine/      cópia, renomeação por sentinela, render de templates
- internal/catalog/     leitura do catalog/databases.yaml
- internal/doctor/      checagem de pré-requisitos
- templates/spring-api/ projeto de referência Spring (COMPILA sozinho)
- templates/angular-app/ projeto de referência Angular (COMPILA sozinho)
- catalog/              dados: bancos, features
- testdata/golden/      saídas esperadas

## Comandos
- go build ./... && go vet ./... && go test ./...
- golangci-lint run
- Template Spring: cd templates/spring-api && ./mvnw -q verify
- Template Angular: cd templates/angular-app && npm ci && npm run lint && npm test -- --watch=false && npm run build
- Terraform: terraform fmt -check -recursive && terraform init -backend=false && terraform validate

## Regras
- Templates de referência DEVEM compilar e passar nos testes sem passar pelo forge.
- Arquivos de infra variáveis usam delimitadores [[ ]], nunca {{ }}.
- Antes de compilar o forge, garanta que templates/*/target não existe (`./mvnw clean`): o go:embed incluiria os artefatos.
- Nunca renderizar arquivos binários (jar, png, fontes). Preservar bit de execução (mvnw, scripts).
- Todo nome gerado deve respeitar limites da AWS (ALB/target group 32, bucket 63, role IAM 64).
- Nome de banco, schema e usuário sempre em snake_case minúsculo.
- Não adicionar dependência Go nova sem justificar no commit.
- Antes de fixar versões (Spring Boot, Spring Cloud, Terraform, actions, imagens Docker),
  consulte a documentação oficial atual. Não confie em versões de memória.
- Compatibilidade Spring Boot x Spring Cloud: verificar a tabela oficial.
- Não rodar `terraform apply` nem criar recursos AWS. Apenas validate/plan.
- Commits pequenos, mensagens em Conventional Commits.
- Se uma tarefa estiver ambígua, pergunte antes de implementar.

## Definition of done (toda fase)
- Build, vet, testes e lint verdes.
- Documentação da fase atualizada (README ou docs/).
- Nenhum TODO sem issue ou nota em docs/PLAN.md.
