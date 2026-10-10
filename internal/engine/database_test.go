package engine_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ikauedeveloper/forge/internal/catalog"
	"github.com/ikauedeveloper/forge/internal/engine"
	"github.com/ikauedeveloper/forge/internal/template"
	"github.com/ikauedeveloper/forge/templates"
)

const (
	migrationRel = "src/main/resources/db/migration/V1__create_addresses.sql"
	testcontCfg  = "src/test/java/com/empresa/ordersapi/TestcontainersConfig.java"
)

// Os trechos gerados para o banco de referência (postgres) precisam reproduzir exatamente o texto
// que está dentro dos blocos forge:slot do template. Sem isso, catálogo e template divergem em silêncio.
func TestPostgresSlotsMatchTemplateReference(t *testing.T) {
	c, err := catalog.Load()
	if err != nil {
		t.Fatal(err)
	}
	pg, _ := c.Get("postgres")
	slots, err := pg.Slots()
	if err != nil {
		t.Fatal(err)
	}
	sel, err := manifest(t).Resolve(nil)
	if err != nil {
		t.Fatal(err)
	}
	sel = template.Selection{"db": "postgres", "http": "both", "auth": "jwt", "cache": "redis"}

	src, _ := fs.Sub(templates.FS, templates.SpringAPI)
	used := map[string]bool{}
	err = fs.WalkDir(src, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || engine.IsBinary(p, nil) {
			return err
		}
		b, err := fs.ReadFile(src, p)
		if err != nil || !strings.Contains(string(b), "forge:slot") {
			return err
		}
		reference, err := engine.ProcessDirectivesWithSlots(p, string(b), sel, nil)
		if err != nil {
			return err
		}
		generated, err := engine.ProcessDirectivesWithSlots(p, string(b), sel, slots)
		if err != nil {
			return err
		}
		if reference != generated {
			t.Errorf("%s: o slot gerado para postgres difere do texto de referência.\n--- referência\n%s\n--- gerado\n%s", p, reference, generated)
		}
		for name := range slots {
			if strings.Contains(string(b), "forge:slot "+name) {
				used[name] = true
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for name := range slots {
		if !used[name] {
			t.Errorf("slot %q do catálogo não é usado em nenhum arquivo do template", name)
		}
	}
}

func TestReferenceMigrationMatchesCatalog(t *testing.T) {
	c, _ := catalog.Load()
	migs, err := c.Migrations("postgres")
	if err != nil {
		t.Fatal(err)
	}
	src, _ := fs.Sub(templates.FS, templates.SpringAPI)
	ref, err := fs.ReadFile(src, migrationRel)
	if err != nil {
		t.Fatal(err)
	}
	if string(migs["V1__create_addresses.sql"]) != string(ref) {
		t.Error("a migration de referência do template difere de catalog/migrations/postgres")
	}
}

func allFilesLower(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		if engine.IsBinary(p, b) {
			return nil
		}
		rel, _ := filepath.Rel(root, p)
		out[filepath.ToSlash(rel)] = strings.ToLower(string(b))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func TestDatabaseOption(t *testing.T) {
	type want struct {
		pom, compose, full, migration, container []string
		absent                                   string // texto que não pode aparecer em nenhum arquivo
	}
	cases := map[string]want{
		"postgres": {
			pom:       []string{"flyway-database-postgresql", "<groupid>org.postgresql</groupid>", "testcontainers-postgresql"},
			compose:   []string{"image: postgres:17-alpine", "postgres_db:", "pg_isready", "${db_port:-5432}:5432"},
			full:      []string{"jdbc:postgresql://db:5432/"},
			migration: []string{"create table addresses"},
			container: []string{"org.testcontainers.postgresql.postgresqlcontainer", "postgres:17-alpine"},
			absent:    "mysql",
		},
		"mysql": {
			pom:       []string{"<artifactid>flyway-mysql</artifactid>", "mysql-connector-j", "testcontainers-mysql"},
			compose:   []string{"image: mysql:8.4", "mysql_database:", `mysql_random_root_password: "yes"`, "mysqladmin ping", "${db_port:-3306}:3306", "/var/lib/mysql"},
			full:      []string{"jdbc:mysql://db:3306/"},
			migration: []string{"engine = innodb", "utf8mb4"},
			container: []string{"org.testcontainers.mysql.mysqlcontainer", "mysql:8.4"},
			absent:    "postgres",
		},
		"mariadb": {
			pom:       []string{"<artifactid>flyway-mysql</artifactid>", "mariadb-java-client", "testcontainers-mariadb"},
			compose:   []string{"image: mariadb:11.8", "mariadb_database:", `mariadb_random_root_password: "yes"`, "healthcheck.sh"},
			full:      []string{"jdbc:mariadb://db:3306/"},
			migration: []string{"engine = innodb", "utf8mb4_unicode_ci"},
			container: []string{"org.testcontainers.mariadb.mariadbcontainer", "mariadb:11.8"},
			absent:    "postgres",
		},
	}
	for name, w := range cases {
		t.Run(name, func(t *testing.T) {
			dest := gen(t, "db", name)
			files := allFilesLower(t, dest)
			expect := func(file string, parts []string) {
				for _, p := range parts {
					if !strings.Contains(files[file], p) {
						t.Errorf("%s deveria conter %q", file, p)
					}
				}
			}
			expect("pom.xml", w.pom)
			expect("compose.yaml", w.compose)
			expect("compose.full.yaml", w.full)
			expect(migrationRel, w.migration)
			expect(testcontCfg, w.container)
			for file, content := range files {
				if strings.Contains(content, w.absent) {
					t.Errorf("%s não deveria citar %q", file, w.absent)
				}
			}
		})
	}
}

func TestDatabaseNone(t *testing.T) {
	dest := gen(t, "db", "none", "cache", "none")
	files := allFilesLower(t, dest)
	for _, gone := range []string{
		"compose.yaml", testcontCfg, migrationRel,
		mainJava + "adapter/out/persistence/AddressPersistenceAdapter.java",
		testJava + "adapter/out/persistence/AddressPersistenceAdapterTest.java",
	} {
		if _, ok := files[gone]; ok {
			t.Errorf("db=none não deve gerar %s", gone)
		}
	}
	pom := files["pom.xml"]
	for _, absent := range []string{"jpa", "flyway", "postgres", "testcontainers", "mysql", "docker-compose"} {
		if strings.Contains(pom, absent) {
			t.Errorf("pom de db=none ainda cita %q", absent)
		}
	}
	for file, content := range files {
		for _, absent := range []string{"@entity", "org.flywaydb", "jakarta.persistence", "db_url", "depends_on", "adminer"} {
			if strings.HasSuffix(file, "ArchitectureTest.java") && absent == "jakarta.persistence" {
				continue // a regra do ArchUnit proíbe JPA no domínio e precisa citá-lo
			}
			if strings.Contains(content, absent) {
				t.Errorf("%s (db=none) ainda cita %q", file, absent)
			}
		}
	}
	if !strings.Contains(files[mainJava+"config/DomainConfig.java"], "noopaddressrepository") {
		t.Error("db=none deve ligar o repositório sem persistência")
	}
	if _, ok := files[mainJava+"config/NoOpAddressRepository.java"]; !ok {
		t.Error("db=none deve ter NoOpAddressRepository")
	}
	if strings.Contains(files["Makefile"], "docker compose up -d") ||
		!strings.Contains(files["Makefile"], "docker compose -f compose.full.yaml up --build") {
		t.Errorf("Makefile de db=none incorreto:\n%s", files["Makefile"])
	}
	if strings.Contains(files["src/main/resources/application.yaml"], "jpa") {
		t.Error("application.yaml de db=none ainda tem jpa")
	}

	// db=none com Redis: o compose.yaml fica só com o cache
	red := allFilesLower(t, gen(t, "db", "none", "cache", "redis"))
	compose := red["compose.yaml"]
	if !strings.Contains(compose, "redis:8-alpine") || strings.Contains(compose, "db:") || strings.Contains(compose, "adminer") {
		t.Errorf("compose de db=none+redis incorreto:\n%s", compose)
	}
	if _, ok := red[testcontCfg]; !ok || !strings.Contains(red["pom.xml"], "testcontainers") {
		t.Error("db=none+redis precisa de Testcontainers para o Redis")
	}
	if strings.Contains(red[testcontCfg], "postgres") {
		t.Error("TestcontainersConfig de db=none+redis não deve citar banco")
	}
}
