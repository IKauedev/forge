# Como adicionar um banco de dados

Um banco novo exige **um bloco em `catalog/databases.yaml`** e **uma pasta de migrations**. Não é preciso
mexer em Go nem no manifesto do template: `--db`, `forge list`, o formulário interativo e os testes passam a
enxergar o banco sozinhos.

## Passo a passo

1. **Escolha as versões na documentação oficial** (imagem Docker com tag fixa, driver JDBC, módulo do Flyway
   e módulo do Testcontainers). Prefira versões LTS. Drivers e Flyway costumam vir do BOM do Spring Boot, então
   o `pom.xml` gerado não leva `<version>`.
2. **Acrescente um bloco** ao final de `catalog/databases.yaml`, copiando o mais parecido (`mysql` e
   `mariadb` são bons modelos para bancos com variáveis `*_DATABASE/_USER/_PASSWORD`):

   | Campo | Para que serve |
   |---|---|
   | `name` | valor de `--db` (minúsculo; `none` é reservado) |
   | `image`, `port`, `data_dir` | serviço do compose (tag fixa, nunca `latest`) |
   | `jdbc_scheme` | usado na URL `jdbc:<scheme>://db:<port>/<base>` do `compose.full.yaml` |
   | `driver.group/artifact` | dependência JDBC em runtime |
   | `flyway_module` | módulo do Flyway para o dialeto |
   | `testcontainers.artifact/class` | módulo e classe do container nos testes |
   | `env.database/user/password` | nomes das variáveis que a imagem lê |
   | `env.extra` | linhas `CHAVE: valor` extras (ex.: senha aleatória do root) |
   | `healthcheck` | comando do healthcheck; escape o `$` do compose como `$$` |
   | `aws.*` | engine, versão e family do RDS (usado pela infra Terraform) |

3. **Crie a migration** em `catalog/migrations/<name>/V1__create_addresses.sql` com o dialeto do banco
   (a tabela `addresses`: `cep` PK, `street`, `complement`, `neighborhood`, `city`, `state`).
4. **Rode os testes**:

   ```sh
   go test ./...
   ```

   Eles validam o catálogo (campos obrigatórios, nomes únicos, migration presente), geram o projeto com o banco
   novo em todas as combinações das outras opções e conferem que não sobra diretiva, marcador ou sentinela e
   que YAML e XML continuam válidos.
5. **Valide de verdade** com o smoke (precisa de JDK 25 e Docker). Acrescente o banco à `smokeMatrix` em
   `internal/engine/smoke_test.go` e rode:

   ```sh
   go test -tags smoke -timeout 60m -run TestSmokeVerify ./internal/engine
   ```

   Ele sobe o container com o healthcheck do compose (`up --wait`) e executa `./mvnw verify` com Testcontainers.

## O que o catálogo gera

O template de referência escreve o trecho do Postgres dentro de blocos `forge:slot`. Para os outros bancos o
engine troca o bloco por um trecho gerado de `internal/catalog/slots.go`:

| Slot | Arquivo | Conteúdo |
|---|---|---|
| `db-deps` | `pom.xml` | módulo do Flyway e driver JDBC |
| `db-test-deps` | `pom.xml` | módulo do Testcontainers |
| `db-compose-service` | `compose.yaml` | serviço `db` (imagem, env, porta, volume, healthcheck) |
| `db-full-url` | `compose.full.yaml` | `DB_URL` visto de dentro da rede do compose |
| `db-test-container` | `TestcontainersConfig.java` | bean do container com `@ServiceConnection` |

Além disso, as migrations de `catalog/migrations/<name>/` substituem a de referência em
`src/main/resources/db/migration/`.

`TestPostgresSlotsMatchTemplateReference` garante que os trechos gerados para o Postgres são idênticos ao texto
de referência do template. Se você mudar um trecho em `slots.go`, atualize o bloco correspondente no template
(e vice-versa).

## Quando o catálogo não basta

Se o banco precisar de algo que os campos não expressam (por exemplo, outro formato de URL JDBC, mais de um
arquivo no compose ou um dialeto Hibernate explícito), amplie o modelo do slot em `internal/catalog/slots.go`
e o campo correspondente em `Database`. Mantenha o texto de referência do template em sincronia: o teste acima
acusa qualquer divergência.
