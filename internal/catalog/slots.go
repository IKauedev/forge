package catalog

// slotTemplates são os trechos gerados para cada "forge:slot <nome>" do template. O texto de
// referência (Postgres) vive nos arquivos do template; um teste garante que estes modelos, aplicados
// ao Postgres, reproduzem exatamente esse texto.
var slotTemplates = map[string]string{
	// pom.xml: Flyway e driver JDBC
	"db-deps": `    <dependency>
      <groupId>org.flywaydb</groupId>
      <artifactId>[[ .FlywayModule ]]</artifactId>
    </dependency>
    <dependency>
      <groupId>[[ .Driver.Group ]]</groupId>
      <artifactId>[[ .Driver.Artifact ]]</artifactId>
      <scope>runtime</scope>
    </dependency>
`,

	// pom.xml: módulo do Testcontainers
	"db-test-deps": `    <dependency>
      <groupId>org.testcontainers</groupId>
      <artifactId>[[ .Testcontainers.Artifact ]]</artifactId>
      <scope>test</scope>
    </dependency>
`,

	// compose.yaml: serviço do banco
	"db-compose-service": `  db:
    image: [[ .Image ]]
    environment:
      [[ .Env.Database ]]: ${DB_NAME:-starter_api}
      [[ .Env.User ]]: ${DB_USER:-starter_api}
      [[ .Env.Password ]]: ${DB_PASSWORD:-starter_api}
[[- range .Env.Extra ]]
      [[ . ]]
[[- end ]]
    ports:
      - "${DB_PORT:-[[ .Port ]]}:[[ .Port ]]"
    volumes:
      - db-data:[[ .DataDir ]]
    healthcheck:
      test: ["CMD-SHELL", "[[ .Healthcheck ]]"]
      interval: 5s
      timeout: 3s
      retries: 10
      start_period: 5s
`,

	// compose.full.yaml: URL JDBC vista de dentro da rede do compose
	"db-full-url": `      DB_URL: jdbc:[[ .JDBCScheme ]]://db:[[ .Port ]]/${DB_NAME:-starter_api}
`,

	// TestcontainersConfig.java: bean do container (nomes totalmente qualificados, sem import)
	"db-test-container": `  @Bean
  @ServiceConnection
  [[ .Testcontainers.Class ]] database() {
    return new [[ .Testcontainers.Class ]]("[[ .Image ]]");
  }
`,
}
