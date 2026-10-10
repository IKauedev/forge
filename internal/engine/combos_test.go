package engine_test

import (
	"encoding/xml"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/ikauedev/forge/internal/engine"
	"github.com/ikauedev/forge/internal/template"
	"github.com/ikauedev/forge/templates"
)

func manifest(t *testing.T) *template.Manifest {
	t.Helper()
	src, err := fs.Sub(templates.FS, templates.SpringAPI)
	if err != nil {
		t.Fatal(err)
	}
	m, err := engine.LoadManifest(src)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

// allSelections devolve o produto cartesiano de todas as opções do manifesto.
func allSelections(m *template.Manifest) []template.Selection {
	out := []template.Selection{{}}
	for _, o := range m.Options {
		var next []template.Selection
		for _, base := range out {
			for _, v := range o.Values {
				s := template.Selection{}
				for k, val := range base {
					s[k] = val
				}
				s[o.Name] = v
				next = append(next, s)
			}
		}
		out = next
	}
	return out
}

func label(sel template.Selection) string {
	var parts []string
	for _, k := range sel.SortedKeys() {
		parts = append(parts, k+"="+sel[k])
	}
	return strings.Join(parts, ",")
}

var (
	directiveLeft = regexp.MustCompile(`forge:(if|else|end)\b`)
	markerLeft    = regexp.MustCompile(`(?m)(//\s*~|#~|<!--~)`)
	sentinelLeft  = regexp.MustCompile(`(?i)starter[-_ ]?api`)
)

func TestManifestReferenceIsAnOption(t *testing.T) {
	m := manifest(t)
	if _, err := m.Resolve(m.Reference); err != nil {
		t.Fatalf("seleção de referência inválida: %v", err)
	}
	if len(m.Reference) != len(m.Options) {
		t.Errorf("a referência deve definir todas as opções: %v", m.Reference)
	}
}

func TestEveryCombinationIsCleanAndWellFormed(t *testing.T) {
	m := manifest(t)
	combos := allSelections(m)
	if len(combos) != 216 { // 4 bancos x 2 vpc x 3 http x 3 auth x 3 cache
		t.Fatalf("esperava 216 combinações, há %d", len(combos))
	}
	for _, sel := range combos {
		t.Run(label(sel), func(t *testing.T) {
			dest := filepath.Join(t.TempDir(), "orders-api")
			files := generateWith(t, "Orders API", "com.empresa", dest, sel)
			checkWorkflows(t, dest, "orders-api")
			checkInfra(t, dest, sel)

			for _, rel := range files {
				if rel == template.ManifestFile {
					t.Errorf("manifesto não deve ser copiado")
				}
				b, err := os.ReadFile(filepath.Join(dest, filepath.FromSlash(rel)))
				if err != nil {
					t.Fatal(err)
				}
				if engine.IsBinary(rel, b) {
					continue
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
				checkSyntax(t, rel, b)
			}
		})
	}
}

func checkSyntax(t *testing.T, rel string, b []byte) {
	t.Helper()
	switch filepath.Ext(rel) {
	case ".yaml", ".yml":
		var v any
		if err := yaml.Unmarshal(b, &v); err != nil {
			t.Errorf("YAML inválido em %s: %v", rel, err)
		}
	case ".xml":
		dec := xml.NewDecoder(strings.NewReader(string(b)))
		for {
			if _, err := dec.Token(); err != nil {
				if err != io.EOF {
					t.Errorf("XML inválido em %s: %v", rel, err)
				}
				break
			}
		}
	}
}

func read(t *testing.T, dest, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dest, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("esperava %s: %v", rel, err)
	}
	return string(b)
}

func exists(dest, rel string) bool {
	_, err := os.Stat(filepath.Join(dest, filepath.FromSlash(rel)))
	return err == nil
}

const (
	mainJava = "src/main/java/com/empresa/ordersapi/"
	testJava = "src/test/java/com/empresa/ordersapi/"
)

func gen(t *testing.T, kv ...string) string {
	t.Helper()
	sel := template.Selection{}
	for i := 0; i < len(kv); i += 2 {
		sel[kv[i]] = kv[i+1]
	}
	dest := filepath.Join(t.TempDir(), "orders-api")
	generateWith(t, "Orders API", "com.empresa", dest, sel)
	return dest
}

func TestHTTPOption(t *testing.T) {
	feign := gen(t, "http", "feign")
	if exists(feign, mainJava+"adapter/out/client/ViaCepWebClientAdapter.java") ||
		exists(feign, testJava+"adapter/out/client/ViaCepWebClientTest.java") {
		t.Error("http=feign não deve ter WebClient")
	}
	pom := read(t, feign, "pom.xml")
	if !strings.Contains(pom, "spring-cloud-starter-openfeign") || strings.Contains(pom, "spring-boot-starter-webclient") {
		t.Error("pom de http=feign com dependências erradas")
	}
	if strings.Contains(read(t, feign, mainJava+"config/ViaCepClientConfig.java"), "WebClientImpl") {
		t.Error("config de http=feign ainda cita WebClientImpl")
	}

	web := gen(t, "http", "webclient")
	for _, gone := range []string{
		mainJava + "adapter/out/client/ViaCepFeignAdapter.java",
		mainJava + "adapter/out/client/ViaCepFeignClient.java",
		mainJava + "adapter/out/client/CorrelationIdFeignInterceptor.java",
		testJava + "adapter/out/client/ViaCepFeignClientTest.java",
	} {
		if exists(web, gone) {
			t.Errorf("http=webclient não deve ter %s", gone)
		}
	}
	pom = read(t, web, "pom.xml")
	if strings.Contains(pom, "openfeign") || strings.Contains(pom, "spring-cloud") {
		t.Error("pom de http=webclient ainda tem Spring Cloud")
	}
	yml := read(t, web, "src/main/resources/application.yaml")
	if !strings.Contains(yml, "impl: webclient") || strings.Contains(yml, "openfeign") {
		t.Errorf("application.yaml de http=webclient incorreto:\n%s", yml)
	}

	both := gen(t, "http", "both")
	if !exists(both, mainJava+"adapter/out/client/ViaCepFeignAdapter.java") ||
		!exists(both, mainJava+"adapter/out/client/ViaCepWebClientAdapter.java") {
		t.Error("http=both deve ter os dois adapters")
	}
}

func TestAuthOption(t *testing.T) {
	none := gen(t, "auth", "none")
	for _, gone := range []string{
		mainJava + "shared/security/SecurityUtils.java",
		mainJava + "config/SecurityConfig.java",
		mainJava + "adapter/in/web/MeController.java",
		mainJava + "adapter/in/web/DevTokenController.java",
		testJava + "shared/security/SecurityAnnotationsTest.java",
	} {
		if exists(none, gone) {
			t.Errorf("auth=none não deve ter %s", gone)
		}
	}
	if strings.Contains(read(t, none, "pom.xml"), "security") {
		t.Error("pom de auth=none ainda tem Spring Security")
	}
	if strings.Contains(read(t, none, mainJava+"adapter/in/web/AddressController.java"), "IsAuthenticated") {
		t.Error("AddressController de auth=none ainda cita @IsAuthenticated")
	}
	if strings.Contains(read(t, none, "src/main/resources/application.yaml"), "roles-claim") {
		t.Error("application.yaml de auth=none ainda tem roles-claim")
	}
	if strings.Contains(read(t, none, ".env.example"), "JWT_ISSUER_URI") {
		t.Error(".env.example de auth=none ainda tem JWT_ISSUER_URI")
	}

	jwt := gen(t, "auth", "jwt")
	if !exists(jwt, mainJava+"shared/security/SecurityUtils.java") ||
		!strings.Contains(read(t, jwt, "src/main/resources/application.yaml"), "roles-claim: roles") {
		t.Error("auth=jwt incompleto")
	}

	cog := gen(t, "auth", "cognito")
	yml := read(t, cog, "src/main/resources/application.yaml")
	if !strings.Contains(yml, `roles-claim: "cognito:groups"`) || strings.Contains(yml, "roles-claim: roles") {
		t.Errorf("auth=cognito deveria usar cognito:groups:\n%s", yml)
	}
	if !strings.Contains(read(t, cog, ".env.example"), "cognito-idp") ||
		!strings.Contains(read(t, cog, "README.md"), "Amazon Cognito") {
		t.Error("auth=cognito sem documentação do user pool")
	}
}

func TestCacheOption(t *testing.T) {
	none := gen(t, "cache", "none")
	if exists(none, mainJava+"config/CacheConfig.java") {
		t.Error("cache=none não deve ter CacheConfig")
	}
	pom := read(t, none, "pom.xml")
	if strings.Contains(pom, "starter-cache") || strings.Contains(pom, "caffeine") || strings.Contains(pom, "redis") {
		t.Error("pom de cache=none com dependências de cache")
	}
	if strings.Contains(read(t, none, mainJava+"application/AddressApplicationService.java"), "Cacheable") {
		t.Error("serviço de cache=none ainda usa @Cacheable")
	}
	if strings.Contains(read(t, none, testJava+"adapter/out/client/AbstractViaCepClientTest.java"), "ServedFromCache") {
		t.Error("teste de cache presente em cache=none")
	}

	caf := gen(t, "cache", "caffeine")
	if !exists(caf, mainJava+"config/CacheConfig.java") ||
		!strings.Contains(read(t, caf, "pom.xml"), "<artifactId>caffeine</artifactId>") ||
		!strings.Contains(read(t, caf, "src/main/resources/application.yaml"), "type: caffeine") {
		t.Error("cache=caffeine incompleto")
	}

	red := gen(t, "cache", "redis")
	pom = read(t, red, "pom.xml")
	if !strings.Contains(pom, "spring-boot-starter-data-redis") || strings.Contains(pom, "<artifactId>caffeine</artifactId>") {
		t.Error("pom de cache=redis incorreto")
	}
	if !strings.Contains(read(t, red, "compose.yaml"), "image: redis:8-alpine") ||
		!strings.Contains(read(t, red, "compose.full.yaml"), "REDIS_HOST: redis") ||
		!strings.Contains(read(t, red, "src/main/resources/application-prod.yaml"), "host: ${REDIS_HOST}") ||
		!strings.Contains(read(t, red, "src/main/resources/application.yaml"), "type: redis") ||
		!strings.Contains(read(t, red, "src/test/java/com/empresa/ordersapi/TestcontainersConfig.java"), `name = "redis"`) {
		t.Error("cache=redis sem compose/config/testcontainers")
	}
	if strings.Contains(read(t, caf, "compose.yaml"), "redis") {
		t.Error("compose de caffeine não deve ter redis")
	}
}
