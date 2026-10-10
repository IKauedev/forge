package catalog

import (
	"strings"
	"testing"
	"testing/fstest"
)

func TestEmbeddedCatalogIsValid(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if got, want := strings.Join(c.Names(), ","), "postgres,mysql,mariadb"; got != want {
		t.Errorf("bancos = %s, want %s", got, want)
	}
	for _, d := range c.Databases {
		m, err := c.Migrations(d.Name)
		if err != nil || len(m) == 0 {
			t.Errorf("%s sem migrations: %v", d.Name, err)
		}
		if d.AWS.Engine == "" || d.AWS.Family == "" || d.AWS.Port == 0 {
			t.Errorf("%s sem mapeamento AWS: %+v", d.Name, d.AWS)
		}
	}
}

const validDB = `databases:
  - name: demo
    image: demo:1
    port: 1234
    jdbc_scheme: demo
    driver: {group: g, artifact: a}
    flyway_module: flyway-demo
    testcontainers: {artifact: tc-demo, class: org.demo.DemoContainer}
    env: {database: D_DB, user: D_USER, password: D_PASS}
    data_dir: /data
    healthcheck: demo ping
`

func fsWith(yaml string, withMigration bool) fstest.MapFS {
	f := fstest.MapFS{"databases.yaml": {Data: []byte(yaml)}}
	if withMigration {
		f["migrations/demo/V1__x.sql"] = &fstest.MapFile{Data: []byte("select 1;")}
	}
	return f
}

func TestValidation(t *testing.T) {
	if _, err := LoadFS(fsWith(validDB, true)); err != nil {
		t.Fatalf("catálogo mínimo deveria ser válido: %v", err)
	}
	tests := []struct {
		name, yaml string
		migration  bool
		errPart    string
	}{
		{"sem migration", validDB, false, "migrations/demo"},
		{"sem imagem", strings.Replace(validDB, "    image: demo:1\n", "", 1), true, "image"},
		{"sem driver", strings.Replace(validDB, "    driver: {group: g, artifact: a}\n", "", 1), true, "driver"},
		{"sem porta", strings.Replace(validDB, "    port: 1234\n", "", 1), true, "port"},
		{"nome reservado", strings.Replace(validDB, "name: demo", "name: none", 1), true, "reservado"},
		{"duplicado", validDB + strings.TrimPrefix(validDB, "databases:\n"), true, "duplicado"},
		{"vazio", "databases: []\n", true, "vazio"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := LoadFS(fsWith(tt.yaml, tt.migration))
			if err == nil || !strings.Contains(err.Error(), tt.errPart) {
				t.Errorf("esperava erro com %q, veio %v", tt.errPart, err)
			}
		})
	}
}

func TestSlotsPerDatabase(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	expect := map[string]map[string][]string{
		"postgres": {
			"db-deps":            {"flyway-database-postgresql", "<groupId>org.postgresql</groupId>"},
			"db-test-deps":       {"testcontainers-postgresql"},
			"db-compose-service": {"image: postgres:17-alpine", "POSTGRES_DB:", "${DB_PORT:-5432}:5432", "pg_isready", "/var/lib/postgresql/data"},
			"db-full-url":        {"jdbc:postgresql://db:5432/"},
			"db-test-container":  {"org.testcontainers.postgresql.PostgreSQLContainer database()", `("postgres:17-alpine")`},
		},
		"mysql": {
			"db-deps":            {"flyway-mysql", "<groupId>com.mysql</groupId>", "mysql-connector-j"},
			"db-test-deps":       {"testcontainers-mysql"},
			"db-compose-service": {"image: mysql:8.4", "MYSQL_DATABASE:", `MYSQL_RANDOM_ROOT_PASSWORD: "yes"`, "${DB_PORT:-3306}:3306", "mysqladmin ping", "/var/lib/mysql"},
			"db-full-url":        {"jdbc:mysql://db:3306/"},
			"db-test-container":  {"org.testcontainers.mysql.MySQLContainer database()", `("mysql:8.4")`},
		},
		"mariadb": {
			"db-deps":            {"flyway-mysql", "org.mariadb.jdbc", "mariadb-java-client"},
			"db-test-deps":       {"testcontainers-mariadb"},
			"db-compose-service": {"image: mariadb:11.8", "MARIADB_USER:", `MARIADB_RANDOM_ROOT_PASSWORD: "yes"`, "healthcheck.sh"},
			"db-full-url":        {"jdbc:mariadb://db:3306/"},
			"db-test-container":  {"org.testcontainers.mariadb.MariaDBContainer database()", `("mariadb:11.8")`},
		},
	}
	for name, slots := range expect {
		d, ok := c.Get(name)
		if !ok {
			t.Fatalf("banco %s ausente", name)
		}
		got, err := d.Slots()
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != len(slotTemplates) {
			t.Errorf("%s: %d slots, esperava %d", name, len(got), len(slotTemplates))
		}
		for slot, parts := range slots {
			for _, part := range parts {
				if !strings.Contains(got[slot], part) {
					t.Errorf("%s/%s deveria conter %q:\n%s", name, slot, part, got[slot])
				}
			}
		}
	}
}

func TestComposeServiceWithoutExtraEnvHasNoBlankLines(t *testing.T) {
	c, _ := Load()
	pg, _ := c.Get("postgres")
	slots, err := pg.Slots()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(slots["db-compose-service"], "\n\n") {
		t.Errorf("linhas em branco inesperadas:\n%s", slots["db-compose-service"])
	}
}
