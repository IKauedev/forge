package engine_test

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/ikauedev/forge/internal/engine"
	"github.com/ikauedev/forge/internal/names"
	"github.com/ikauedev/forge/internal/template"
	"github.com/ikauedev/forge/templates"
)

func quarkusManifest(t *testing.T) *template.Manifest {
	t.Helper()
	src, err := fs.Sub(templates.FS, templates.QuarkusAPI)
	if err != nil {
		t.Fatal(err)
	}
	m, err := engine.LoadManifest(src)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func genQuarkus(t *testing.T, raw string, sel template.Selection) string {
	t.Helper()
	n, err := names.New(raw)
	if err != nil {
		t.Fatal(err)
	}
	return genQuarkusAt(t, filepath.Join(t.TempDir(), n.Kebab), raw, sel)
}

func genQuarkusAt(t *testing.T, dest, raw string, sel template.Selection) string {
	t.Helper()
	n, err := names.New(raw)
	if err != nil {
		t.Fatal(err)
	}
	src, err := fs.Sub(templates.FS, templates.QuarkusAPI)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := engine.Generate(src, engine.Options{Names: n, Package: "com.empresa", Dest: dest, Selection: sel}); err != nil {
		t.Fatal(err)
	}
	return dest
}

func TestQuarkusReferenceIsAnOption(t *testing.T) {
	m := quarkusManifest(t)
	if _, err := m.Resolve(m.Reference); err != nil {
		t.Fatalf("seleção de referência inválida: %v", err)
	}
	if len(m.Reference) != len(m.Options) {
		t.Errorf("a referência deve definir todas as opções: %v", m.Reference)
	}
	if !m.JavaPackage || m.SentinelName() != "starter-api" {
		t.Error("quarkus-api usa pacote Java e a sentinela starter-api")
	}
}

var propertiesLine = regexp.MustCompile(`^\s*(#.*|%?[\w.\-"/:${}]+\s*=.*|)$`)

func TestQuarkusEveryCombinationIsCleanAndWellFormed(t *testing.T) {
	combos := allSelections(quarkusManifest(t))
	if len(combos) != 16 { // 4 bancos x 2 vpc x 2 auth
		t.Fatalf("esperava 16 combinações, há %d", len(combos))
	}
	for _, sel := range combos {
		t.Run(label(sel), func(t *testing.T) {
			dest := genQuarkus(t, "Orders Api", sel)
			checkWorkflows(t, dest, "orders-api")
			checkInfra(t, dest, sel)

			_ = filepath.WalkDir(dest, func(p string, d os.DirEntry, err error) error {
				if err != nil || d.IsDir() {
					return nil
				}
				rel, _ := filepath.Rel(dest, p)
				rel = filepath.ToSlash(rel)
				b, _ := os.ReadFile(p)
				if engine.IsBinary(rel, b) {
					return nil
				}
				text := string(b)
				if directiveLeft.MatchString(text) {
					t.Errorf("sobrou diretiva forge: em %s", rel)
				}
				if markerLeft.MatchString(text) {
					t.Errorf("sobrou marcador ~ em %s", rel)
				}
				if sentinelLeft.MatchString(rel) || sentinelLeft.MatchString(text) {
					t.Errorf("sobrou sentinela em %s", rel)
				}
				for _, banned := range []string{"springframework", "SPRING_PROFILES", "actuator", "REDIS_HOST"} {
					if strings.Contains(text, banned) && rel != "docs/PIPELINE.md" {
						t.Errorf("%s: resquício do template Spring (%s)", rel, banned)
					}
				}
				checkSyntax(t, rel, b)
				if strings.HasSuffix(rel, ".properties") {
					for i, line := range strings.Split(text, "\n") {
						if !propertiesLine.MatchString(strings.TrimRight(line, "\r")) {
							t.Errorf("%s:%d: linha inválida: %q", rel, i+1, line)
						}
					}
				}
				return nil
			})
		})
	}
}

func TestQuarkusDatabaseOption(t *testing.T) {
	type want struct {
		kind, image, extension, flyway, migration, compose, full string
		absent                                                   string
	}
	cases := map[string]want{
		"postgres": {"postgresql", "postgres:17-alpine", "quarkus-jdbc-postgresql", "flyway-database-postgresql", "create table addresses", "image: postgres:17-alpine", "jdbc:postgresql://db:5432/", "mysql"},
		"mysql":    {"mysql", "mysql:8.4", "quarkus-jdbc-mysql", "flyway-mysql", "engine = innodb", "image: mysql:8.4", "jdbc:mysql://db:3306/", "postgres"},
		"mariadb":  {"mariadb", "mariadb:11.8", "quarkus-jdbc-mariadb", "flyway-mysql", "engine = innodb", "image: mariadb:11.8", "jdbc:mariadb://db:3306/", "postgres"},
	}
	for db, w := range cases {
		t.Run(db, func(t *testing.T) {
			dest := genQuarkus(t, "Orders Api", template.Selection{"db": db, "vpc": "new", "auth": "jwt"})
			props := readFile(t, dest, "src/main/resources/application.properties")
			for _, want := range []string{"quarkus.datasource.db-kind=" + w.kind, "quarkus.datasource.devservices.image-name=" + w.image} {
				if !strings.Contains(props, want) {
					t.Errorf("application.properties sem %q", want)
				}
			}
			pom := readFile(t, dest, "pom.xml")
			for _, want := range []string{"<artifactId>" + w.extension + "</artifactId>", "<artifactId>" + w.flyway + "</artifactId>"} {
				if !strings.Contains(pom, want) {
					t.Errorf("pom.xml sem %q", want)
				}
			}
			if !strings.Contains(strings.ToLower(readFile(t, dest, "src/main/resources/db/migration/V1__create_addresses.sql")), w.migration) {
				t.Errorf("migration do dialeto %s ausente", db)
			}
			if !strings.Contains(readFile(t, dest, "compose.yaml"), w.compose) {
				t.Errorf("compose.yaml sem %q", w.compose)
			}
			if !strings.Contains(readFile(t, dest, "compose.full.yaml"), w.full) {
				t.Errorf("compose.full.yaml sem %q", w.full)
			}
			// o infra recebe engine/versão do catálogo (mesmo slot do Spring)
			if !strings.Contains(infraFile(dest, "envs/dev/main.tf"), w.kind) && db != "postgres" {
				t.Errorf("envs/dev/main.tf sem o engine %s", w.kind)
			}
			for rel, text := range allFilesLower(t, dest) {
				if strings.Contains(text, w.absent) && !strings.HasPrefix(rel, "docs/") && !strings.HasPrefix(rel, "infra/") {
					t.Errorf("%s menciona %q, de outro banco", rel, w.absent)
				}
			}
		})
	}
}

func TestQuarkusWithoutDatabase(t *testing.T) {
	dest := genQuarkus(t, "Orders Api", template.Selection{"db": "none", "vpc": "new", "auth": "jwt"})

	pom := readFile(t, dest, "pom.xml")
	for _, banned := range []string{"hibernate-orm", "flyway", "jdbc"} {
		if strings.Contains(pom, banned) {
			t.Errorf("pom.xml não deveria ter %s", banned)
		}
	}
	for _, rel := range []string{"compose.yaml", "src/main/resources/db", "src/main/java/com/empresa/ordersapi/adapter/out/persistence", "infra/modules/database"} {
		if exists(dest, rel) {
			t.Errorf("%s não deveria existir com db=none", rel)
		}
	}
	if strings.Contains(readFile(t, dest, "src/main/resources/application.properties"), "datasource") {
		t.Error("application.properties não deveria configurar datasource")
	}
	cfg := readFile(t, dest, "src/main/java/com/empresa/ordersapi/config/DomainConfig.java")
	if !strings.Contains(cfg, "new NoOpAddressRepository()") || strings.Contains(cfg, "~") {
		t.Error("DomainConfig deveria produzir o NoOpAddressRepository")
	}
	if !exists(dest, "src/main/java/com/empresa/ordersapi/config/NoOpAddressRepository.java") {
		t.Error("NoOpAddressRepository ausente")
	}
	if strings.Contains(readFile(t, dest, "Makefile"), "docker compose up -d") {
		t.Error("o Makefile não deveria ter up/down sem banco")
	}
}

func TestQuarkusAuthOption(t *testing.T) {
	with := genQuarkus(t, "Orders Api", template.Selection{"db": "postgres", "vpc": "new", "auth": "jwt"})
	without := genQuarkus(t, "Orders Api", template.Selection{"db": "postgres", "vpc": "new", "auth": "none"})

	if !strings.Contains(readFile(t, with, "pom.xml"), "quarkus-smallrye-jwt") ||
		!strings.Contains(readFile(t, with, "src/main/java/com/empresa/ordersapi/adapter/in/web/AddressResource.java"), "@Authenticated") ||
		!strings.Contains(readFile(t, with, "src/main/resources/application.properties"), "deny-unannotated-endpoints=true") {
		t.Error("auth=jwt deveria ligar o JWT, @Authenticated e deny-unannotated-endpoints")
	}
	props := readFile(t, with, "src/main/resources/application.properties")
	if !strings.Contains(props, "%prod.mp.jwt.verify.publickey.location=") {
		t.Error("em produção o JWT deve ser verificado pelo JWKS do emissor")
	}

	for _, rel := range []string{"adapter/in/web/MeResource.java", "adapter/in/web/AdminResource.java", "adapter/in/web/DevTokenResource.java", "shared/security"} {
		if exists(without, "src/main/java/com/empresa/ordersapi/"+rel) {
			t.Errorf("%s não deveria existir com auth=none", rel)
		}
	}
	text := readFile(t, without, "pom.xml") + readFile(t, without, "src/main/resources/application.properties") +
		readFile(t, without, "src/main/java/com/empresa/ordersapi/adapter/in/web/AddressResource.java")
	for _, banned := range []string{"jwt", "Authenticated", "security"} {
		if strings.Contains(strings.ToLower(text), strings.ToLower(banned)) {
			t.Errorf("auth=none ainda menciona %q", banned)
		}
	}
	if exists(without, "src/test/java/com/empresa/ordersapi/adapter/in/web/SecurityEndpointsTest.java") {
		t.Error("o teste de segurança não deveria existir com auth=none")
	}
	if strings.Contains(readFile(t, without, "src/test/java/com/empresa/ordersapi/adapter/in/web/AddressResourceTest.java"), "TestSecurity") {
		t.Error("AddressResourceTest não deveria usar @TestSecurity com auth=none")
	}
}

// Nunca deixar uma chave privada ou segredo real no projeto: o perfil dev usa só um segredo simétrico de desenvolvimento.
func TestQuarkusHasNoPrivateKeysAndDevTokenIsDevOnly(t *testing.T) {
	dest := genQuarkus(t, "Orders Api", template.Selection{"db": "postgres", "vpc": "new", "auth": "jwt"})
	_ = filepath.WalkDir(dest, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		b, _ := os.ReadFile(p)
		if strings.Contains(string(b), "BEGIN PRIVATE KEY") || strings.Contains(string(b), "BEGIN RSA PRIVATE KEY") {
			t.Errorf("%s: chave privada no projeto", p)
		}
		return nil
	})
	dev := readFile(t, dest, "src/main/java/com/empresa/ordersapi/adapter/in/web/DevTokenResource.java")
	if !strings.Contains(dev, `@IfBuildProfile(anyOf = {"dev", "test"})`) {
		t.Error("o /dev/token deve existir só nos perfis dev e test")
	}
	props := readFile(t, dest, "src/main/resources/application.properties")
	for _, line := range strings.Split(props, "\n") {
		if strings.HasPrefix(line, "%prod.") && strings.Contains(line, "dev-secret") {
			t.Errorf("o segredo de dev não pode ser usado em produção: %s", line)
		}
	}
}

func TestQuarkusInfraIsFormatted(t *testing.T) {
	bin, err := exec.LookPath("terraform")
	if err != nil {
		t.Skip("terraform não está no PATH")
	}
	for _, sel := range allSelections(quarkusManifest(t)) {
		t.Run(label(sel), func(t *testing.T) {
			dest := genQuarkus(t, "Orders Api", sel)
			cmd := exec.Command(bin, "fmt", "-check", "-recursive", "-diff", "-no-color")
			cmd.Dir = filepath.Join(dest, "infra")
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("terraform fmt reprovou:\n%s", out)
			}
		})
	}
}

func TestQuarkusRuntimeMatchesInfraConventions(t *testing.T) {
	dest := genQuarkus(t, "Orders Api", template.Selection{"db": "postgres", "vpc": "new", "auth": "jwt"})
	if !strings.Contains(readFile(t, dest, "infra/modules/ecs-service/variables.tf"), "/q/health/ready") {
		t.Error("o health check do ALB deve apontar para /q/health/ready")
	}
	if !strings.Contains(readFile(t, dest, "infra/envs/dev/main.tf"), `QUARKUS_PROFILE = "prod"`) {
		t.Error("a task deve rodar o perfil prod do Quarkus")
	}
	// as variáveis que a infra injeta são as que o application.properties lê
	props := readFile(t, dest, "src/main/resources/application.properties")
	for _, env := range []string{"DB_URL", "DB_USERNAME", "DB_PASSWORD", "JWT_ISSUER_URI"} {
		if !strings.Contains(props, "${"+env+"}") && !strings.Contains(props, "${"+env+":") {
			t.Errorf("application.properties não lê %s", env)
		}
		if !strings.Contains(readFile(t, dest, "infra/envs/dev/main.tf"), env) {
			t.Errorf("a infra não injeta %s", env)
		}
	}
	if !strings.Contains(readFile(t, dest, "Dockerfile"), "quarkus-run.jar") {
		t.Error("o Dockerfile deve executar o fast-jar do Quarkus")
	}
}
