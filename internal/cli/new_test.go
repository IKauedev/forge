package cli

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ikauedev/forge/internal/doctor"
	"github.com/ikauedev/forge/internal/project"
	"github.com/ikauedev/forge/internal/template"
)

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := NewRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

// fakePrompter responde como um usuário que escolheu sempre os mesmos valores.
type fakePrompter struct {
	name    string
	pkg     string
	options map[string]string
	called  bool
}

func (f *fakePrompter) Fill(_ *template.Manifest, a *Answers) error {
	f.called = true
	if a.Name == "" {
		a.Name = f.name
	}
	if !a.PackageSet {
		a.Package = f.pkg
	}
	for k, v := range f.options {
		if _, ok := a.Options[k]; !ok {
			a.Options[k] = v
		}
	}
	return nil
}

func withPrompt(t *testing.T, p Prompter, terminal bool) {
	t.Helper()
	oldP, oldT := newPrompter, isTerminal
	newPrompter, isTerminal = p, func() bool { return terminal }
	t.Cleanup(func() { newPrompter, isTerminal = oldP, oldT })
}

func readTree(t *testing.T, root string) map[string]string {
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
		rel, _ := filepath.Rel(root, p)
		out[filepath.ToSlash(rel)] = string(b)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func TestNewCreatesProject(t *testing.T) {
	out := t.TempDir()
	stdout, err := run(t, "new", "Payments API", "--package", "com.empresa", "--out", out, "--no-git")
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(out, "payments-api")
	pom, err := os.ReadFile(filepath.Join(root, "pom.xml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(pom), "<artifactId>payments-api</artifactId>") ||
		!strings.Contains(string(pom), "<groupId>com.empresa</groupId>") {
		t.Errorf("pom.xml não foi renomeado:\n%s", pom)
	}
	if _, err := os.Stat(filepath.Join(root, "src", "main", "java", "com", "empresa", "paymentsapi", "PaymentsApiApplication.java")); err != nil {
		t.Errorf("classe principal ausente: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "template.forge.yaml")); err == nil {
		t.Error("o manifesto não deve ser copiado para o projeto")
	}
	if !strings.Contains(stdout, "Próximos passos") || !strings.Contains(stdout, "auth = jwt") {
		t.Errorf("saída incompleta: %s", stdout)
	}
}

func TestNewWritesForgeYaml(t *testing.T) {
	out := t.TempDir()
	if _, err := run(t, "new", "Payments API", "--package", "com.empresa", "--out", out, "--no-git",
		"--db", "mysql", "--http", "feign", "--auth", "cognito", "--cache", "redis"); err != nil {
		t.Fatal(err)
	}
	cfg, err := project.Read(filepath.Join(out, "payments-api"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Template != "spring-api" || cfg.Project.Name != "Payments API" || cfg.Project.Package != "com.empresa" ||
		cfg.Forge.Version != Version {
		t.Errorf("config inesperada: %+v", cfg)
	}
	want := map[string]string{"db": "mysql", "http": "feign", "auth": "cognito", "cache": "redis"}
	for k, v := range want {
		if cfg.Options[k] != v {
			t.Errorf("option %s = %q, want %q", k, cfg.Options[k], v)
		}
	}
}

func TestInteractiveAndFlagsProduceIdenticalOutput(t *testing.T) {
	flagsOut, promptOut := t.TempDir(), t.TempDir()

	withPrompt(t, &fakePrompter{}, false)
	if _, err := run(t, "new", "Payments API", "--package", "com.empresa", "--out", flagsOut, "--no-git",
		"--db", "mariadb", "--http", "webclient", "--auth", "cognito", "--cache", "redis"); err != nil {
		t.Fatal(err)
	}

	fp := &fakePrompter{
		name: "Payments API", pkg: "com.empresa",
		options: map[string]string{"db": "mariadb", "http": "webclient", "auth": "cognito", "cache": "redis"},
	}
	withPrompt(t, fp, true)
	if _, err := run(t, "new", "--out", promptOut, "--no-git"); err != nil {
		t.Fatal(err)
	}
	if !fp.called {
		t.Fatal("o formulário deveria ter sido usado")
	}

	a, b := readTree(t, filepath.Join(flagsOut, "payments-api")), readTree(t, filepath.Join(promptOut, "payments-api"))
	if len(a) == 0 || len(a) != len(b) {
		t.Fatalf("árvores com tamanhos diferentes: %d vs %d", len(a), len(b))
	}
	for p, content := range a {
		if b[p] != content {
			t.Errorf("conteúdo diferente em %s", p)
		}
	}
}

func TestPromptOnlyAsksForWhatIsMissing(t *testing.T) {
	fp := &fakePrompter{pkg: "com.empresa", options: map[string]string{"db": "postgres", "vpc": "new", "http": "both", "auth": "jwt", "cache": "none", "k8s": "none"}}
	withPrompt(t, fp, true)
	out := t.TempDir()
	if _, err := run(t, "new", "complete", "--package", "com.empresa", "--out", out, "--no-git",
		"--db", "postgres", "--vpc", "new", "--http", "both", "--auth", "jwt", "--cache", "none", "--k8s=none"); err != nil {
		t.Fatal(err)
	}
	if fp.called {
		t.Error("com todas as respostas por flag não deve abrir o formulário")
	}
}

func TestYesNeverPromptsAndUsesDefaults(t *testing.T) {
	fp := &fakePrompter{}
	withPrompt(t, fp, true)
	out := t.TempDir()
	if _, err := run(t, "new", "defaults", "--yes", "--out", out, "--no-git"); err != nil {
		t.Fatal(err)
	}
	if fp.called {
		t.Error("--yes não pode perguntar")
	}
	cfg, err := project.Read(filepath.Join(out, "defaults"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Options["db"] != "postgres" || cfg.Options["http"] != "both" || cfg.Options["auth"] != "jwt" || cfg.Options["cache"] != "none" ||
		cfg.Project.Package != "com.example" {
		t.Errorf("padrões inesperados: %+v", cfg)
	}
}

func TestNameRequiredWithoutTerminal(t *testing.T) {
	withPrompt(t, &fakePrompter{}, false)
	if _, err := run(t, "new", "--out", t.TempDir(), "--no-git"); err == nil ||
		!strings.Contains(err.Error(), "nome") {
		t.Errorf("esperava erro pedindo o nome, veio %v", err)
	}
}

func TestNewRejectsInvalidInput(t *testing.T) {
	withPrompt(t, &fakePrompter{}, false)
	out := t.TempDir()
	for _, args := range [][]string{
		{"new", "class", "--out", out, "--no-git"},
		{"new", "ok-name", "--package", "Com.Bad", "--out", out, "--no-git"},
		{"new", "ok-name", "--http", "grpc", "--out", out, "--no-git"},
		{"new", "ok-name", "--auth", "saml", "--out", out, "--no-git"},
		{"new", "ok-name", "--cache", "memcached", "--out", out, "--no-git"},
		{"new", "ok-name", "--db", "oracle", "--out", out, "--no-git"},
	} {
		if _, err := run(t, args...); err == nil {
			t.Errorf("esperava erro para %v", args)
		}
	}
}

func TestNewRunsGitInit(t *testing.T) {
	withPrompt(t, &fakePrompter{}, false)
	out := t.TempDir()
	if _, err := run(t, "new", "git test", "--out", out); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(out, "git-test", ".git")); err != nil {
		t.Skipf("git indisponível? %v", err)
	}
}

func TestListShowsTemplateAndOptions(t *testing.T) {
	out, err := run(t, "list")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"spring-api", "--db", "postgres | mysql | mariadb | none", "--http", "feign | webclient | both", "--auth", "cognito", "--cache", "redis", "--yes"} {
		if !strings.Contains(out, want) {
			t.Errorf("list deveria mostrar %q:\n%s", want, out)
		}
	}
}

// emptyEnv simula uma máquina sem nada instalado.
type emptyEnv struct{}

func (emptyEnv) LookPath(string) (string, error)          { return "", errors.New("not found") }
func (emptyEnv) Output(string, ...string) (string, error) { return "", errors.New("not found") }
func (emptyEnv) Getenv(string) string                     { return "" }
func (emptyEnv) Getwd() (string, error)                   { return os.TempDir(), nil }
func (emptyEnv) GOOS() string                             { return "linux" }

func TestDoctorExitsNonZeroWhenRequiredMissing(t *testing.T) {
	var buf bytes.Buffer
	cmd := newDoctorCmd()
	cmd.SetOut(&buf)
	err := runDoctor(cmd, doctor.Env(emptyEnv{}), false)

	var exit *ExitError
	if !errors.As(err, &exit) || exit.Code != 1 {
		t.Fatalf("esperava ExitError com código 1, veio %v", err)
	}
	for _, want := range []string{"Docker", "Java", "obrigatória"} {
		if !strings.Contains(buf.String(), want) {
			t.Errorf("relatório deveria citar %q:\n%s", want, buf.String())
		}
	}
}

func TestNewAngularApp(t *testing.T) {
	withPrompt(t, &fakePrompter{}, false)
	out := t.TempDir()
	stdout, err := run(t, "new", "Store Front", "--type", "angular-app", "--out", out, "--no-git")
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(out, "store-front")
	for _, rel := range []string{"package.json", "src/app/app.ts", "infra/envs/dev/main.tf", ".forge.yaml"} {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(rel))); err != nil {
			t.Errorf("%s ausente", rel)
		}
	}
	if !strings.Contains(stdout, "npm ci") || strings.Contains(stdout, "mvnw") {
		t.Errorf("os próximos passos deveriam ser os do Angular:\n%s", stdout)
	}
	cfg, err := project.Read(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Template != "angular-app" || cfg.Project.Package != "" {
		t.Errorf(".forge.yaml inesperado: %+v", cfg)
	}
}

func TestNewRejectsOptionsOfAnotherTemplate(t *testing.T) {
	withPrompt(t, &fakePrompter{}, false)
	for _, args := range [][]string{
		{"new", "app", "--type", "angular-app", "--db", "mysql"},
		{"new", "app", "--type", "angular-app", "--package", "com.empresa"},
		{"new", "app", "--type", "react"},
	} {
		if _, err := run(t, append(args, "--out", t.TempDir(), "--no-git")...); err == nil {
			t.Errorf("%v deveria falhar", args)
		}
	}
}

func TestUpgradeDiffOnFreshAndEditedProject(t *testing.T) {
	withPrompt(t, &fakePrompter{}, false)
	out := t.TempDir()
	if _, err := run(t, "new", "Store Front", "--type", "angular-app", "--out", out, "--no-git"); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(out, "store-front")

	got, err := run(t, "upgrade", dir, "--diff")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "já está igual") {
		t.Errorf("projeto recém-gerado não deveria ter diferenças:\n%s", got)
	}

	readme := filepath.Join(dir, "README.md")
	if err := os.WriteFile(readme, []byte("# editado\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err = run(t, "upgrade", dir, "--diff")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "--- projeto/README.md") || !strings.Contains(got, "1 alterado(s)") {
		t.Errorf("diff do README ausente:\n%s", got)
	}
	if b, _ := os.ReadFile(readme); string(b) != "# editado\n" {
		t.Error("o upgrade não pode alterar o projeto")
	}
}

func TestUpgradeRequiresDiffAndAProject(t *testing.T) {
	if _, err := run(t, "upgrade", t.TempDir()); err == nil {
		t.Error("sem --diff deveria falhar")
	}
	if _, err := run(t, "upgrade", t.TempDir(), "--diff"); err == nil {
		t.Error("fora de um projeto do forge deveria falhar")
	}
}

func TestNewReactApp(t *testing.T) {
	withPrompt(t, &fakePrompter{}, false)
	out := t.TempDir()
	stdout, err := run(t, "new", "Store Front", "--type", "react-app", "--out", out, "--no-git")
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(out, "store-front")
	for _, rel := range []string{"package.json", "src/main.tsx", "src/core/auth/guards.tsx", "infra/envs/dev/main.tf"} {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(rel))); err != nil {
			t.Errorf("%s ausente", rel)
		}
	}
	if !strings.Contains(stdout, "npm ci") {
		t.Errorf("os próximos passos deveriam ser os do React:\n%s", stdout)
	}
	if _, err := run(t, "new", "app", "--type", "react-app", "--auth", "jwt", "--out", t.TempDir(), "--no-git"); err == nil {
		t.Error("--auth não se aplica ao react-app")
	}
}

func TestNewQuarkusApi(t *testing.T) {
	withPrompt(t, &fakePrompter{}, false)
	out := t.TempDir()
	stdout, err := run(t, "new", "Orders Api", "--type", "quarkus-api", "--db", "mysql", "--auth", "none",
		"--package", "com.empresa", "--out", out, "--no-git")
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(out, "orders-api")
	for _, rel := range []string{"pom.xml", "src/main/java/com/empresa/ordersapi/adapter/in/web/AddressResource.java", "infra/envs/dev/main.tf", ".forge.yaml"} {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(rel))); err != nil {
			t.Errorf("%s ausente", rel)
		}
	}
	if !strings.Contains(stdout, "quarkus:dev") || strings.Contains(stdout, "mvnw verify   # requer") {
		t.Errorf("os próximos passos deveriam ser os do Quarkus:\n%s", stdout)
	}
	cfg, err := project.Read(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Template != "quarkus-api" || cfg.Options["db"] != "mysql" || cfg.Options["auth"] != "none" || cfg.Project.Package != "com.empresa" {
		t.Errorf(".forge.yaml inesperado: %+v", cfg)
	}
}

func TestQuarkusRejectsOptionsItDoesNotHave(t *testing.T) {
	withPrompt(t, &fakePrompter{}, false)
	for _, args := range [][]string{
		{"new", "app", "--type", "quarkus-api", "--http", "feign"},   // sem cliente HTTP configurável
		{"new", "app", "--type", "quarkus-api", "--cache", "redis"},  // sem cache
		{"new", "app", "--type", "quarkus-api", "--auth", "cognito"}, // só none|jwt
	} {
		if _, err := run(t, append(args, "--out", t.TempDir(), "--no-git")...); err == nil {
			t.Errorf("%v deveria falhar", args)
		}
	}
}

func TestK8sFlagAloneEnablesKustomize(t *testing.T) {
	withPrompt(t, &fakePrompter{}, false)
	out := t.TempDir()
	if _, err := run(t, "new", "--k8s", "kube", "--yes", "--out", out, "--no-git"); err != nil {
		t.Fatal(err)
	}
	cfg, err := project.Read(filepath.Join(out, "kube"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Options["k8s"] != "kustomize" {
		t.Errorf("--k8s sozinho deveria ligar o kustomize: %+v", cfg.Options)
	}
	for _, env := range []string{"dev", "hml", "prd"} {
		if _, err := os.Stat(filepath.Join(out, "kube", "k8s", "overlays", env, "kustomization.yaml")); err != nil {
			t.Errorf("overlay %s ausente: %v", env, err)
		}
	}
}

func TestK8sIsOffByDefaultAndAcceptsNone(t *testing.T) {
	withPrompt(t, &fakePrompter{}, false)
	for _, args := range [][]string{{"plain"}, {"plain", "--k8s=none"}} {
		out := t.TempDir()
		if _, err := run(t, append([]string{"new"}, append(args, "--yes", "--out", out, "--no-git")...)...); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(filepath.Join(out, "plain", "k8s")); err == nil {
			t.Errorf("%v não deveria gerar k8s/", args)
		}
	}
}
