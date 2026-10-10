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
	"github.com/ikauedev/forge/templates"
)

var angularLeftover = regexp.MustCompile(`(?i)starter[-_ ]?app`)

func genAngular(t *testing.T, raw string) (string, names.Names) {
	t.Helper()
	return genFrontend(t, templates.AngularApp, raw)
}

func genFrontend(t *testing.T, template, raw string) (string, names.Names) {
	t.Helper()
	n, err := names.New(raw)
	if err != nil {
		t.Fatal(err)
	}
	src, err := fs.Sub(templates.FS, template)
	if err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(t.TempDir(), n.Kebab)
	if _, err := engine.Generate(src, engine.Options{Names: n, Dest: dest}); err != nil {
		t.Fatal(err)
	}
	return dest, n
}

// O embed usa padrões explícitos: tudo o que o template tem em disco (menos build/dependências) precisa estar nele.
func TestFrontendTemplatesAreEmbedded(t *testing.T) {
	skip := map[string]bool{"node_modules": true, "dist": true, ".angular": true, ".terraform": true}
	for _, tpl := range []string{templates.AngularApp, templates.ReactApp} {
		t.Run(tpl, func(t *testing.T) {
			root := filepath.Join("..", "..", "templates", tpl)
			err := filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() {
					if skip[d.Name()] {
						return filepath.SkipDir
					}
					return nil
				}
				rel, _ := filepath.Rel(root, p)
				if _, err := fs.Stat(templates.FS, tpl+"/"+filepath.ToSlash(rel)); err != nil {
					t.Errorf("%s não está embutido: ajuste os padrões em templates/templates.go", rel)
				}
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}

			err = fs.WalkDir(templates.FS, tpl, func(p string, d fs.DirEntry, err error) error {
				if err == nil && d.IsDir() && skip[d.Name()] {
					t.Errorf("%s não deveria estar embutido", p)
				}
				return err
			})
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestAngularGenerationRenamesEverything(t *testing.T) {
	dest, n := genAngular(t, "Store Front")

	_ = filepath.WalkDir(dest, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || engine.IsBinary(p, nil) {
			return nil
		}
		b, _ := os.ReadFile(p)
		if m := angularLeftover.FindString(string(b)); m != "" {
			t.Errorf("%s: sobrou sentinela %q", p, m)
		}
		return nil
	})

	pkg := readFile(t, dest, "package.json")
	if !strings.Contains(pkg, `"name": "`+n.Kebab+`"`) {
		t.Errorf("package.json deveria ter o nome %q", n.Kebab)
	}
	if !strings.Contains(readFile(t, dest, "src/index.html"), "<title>"+n.Title+"</title>") {
		t.Errorf("index.html deveria ter o título %q", n.Title)
	}
	if !strings.Contains(readFile(t, dest, "infra/envs/dev/main.tf"), `app = "`+n.Kebab+`"`) {
		t.Error("a infra deveria usar o nome do projeto")
	}
	if !strings.Contains(readFile(t, dest, "Dockerfile"), "/app/dist/"+n.Kebab+"/browser") {
		t.Error("o Dockerfile deveria apontar para o dist do projeto")
	}
	if !strings.Contains(readFile(t, dest, ".github/workflows/deploy.yml"), "path: dist/"+n.Kebab+"/browser") {
		t.Error("o deploy deveria publicar o dist do projeto")
	}
	if _, err := os.Stat(filepath.Join(dest, "src/app/core/config/config.service.ts")); err != nil {
		t.Error("ConfigService ausente")
	}
}

func readFile(t *testing.T, dest, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dest, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("lendo %s: %v", rel, err)
	}
	return string(b)
}

func TestAngularInfraRules(t *testing.T) {
	dest, _ := genAngular(t, "Store Front")

	for _, rel := range []string{
		"infra/bootstrap/main.tf", "infra/bootstrap/.terraform.lock.hcl",
		"infra/envs/dev/.terraform.lock.hcl", "infra/envs/prod/.terraform.lock.hcl",
		"infra/modules/static-site/main.tf", "infra/.tflint.hcl", "infra/.checkov.yaml",
	} {
		if _, err := os.Stat(filepath.Join(dest, filepath.FromSlash(rel))); err != nil {
			t.Errorf("%s ausente", rel)
		}
	}

	_ = filepath.WalkDir(filepath.Join(dest, "infra"), func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		switch filepath.Ext(p) {
		case ".tf", ".hcl", ".example", ".tfvars":
			if m := literalSecret.FindString(readFile(t, dest, mustRel(dest, p))); m != "" {
				t.Errorf("%s: possível segredo em texto: %q", p, m)
			}
		}
		return nil
	})

	site := readFile(t, dest, "infra/modules/static-site/main.tf")
	for _, want := range []string{
		"aws_cloudfront_origin_access_control", // OAC
		"error_code            = 403",          // fallback de SPA
		"error_code            = 404",
		`response_page_path    = "/index.html"`,
		"aws_cloudfront_response_headers_policy",
		"aws.us_east_1",                  // ACM em us-east-1
		"block_public_policy     = true", // bucket privado
		"AWS:SourceArn",                  // só a distribuição lê o bucket
	} {
		if !strings.Contains(site, want) {
			t.Errorf("static-site/main.tf sem %q", want)
		}
	}
	for _, env := range []string{"dev", "prod"} {
		v := readFile(t, dest, "infra/envs/"+env+"/versions.tf")
		if !strings.Contains(v, "use_lockfile = true") || strings.Contains(v, "dynamodb_table") {
			t.Errorf("envs/%s: backend deve usar use_lockfile e não DynamoDB", env)
		}
	}
	boot := readFile(t, dest, "infra/bootstrap/main.tf")
	if !strings.Contains(boot, `:environment:${each.key}`) || !strings.Contains(boot, "var.github_repository") {
		t.Error("a trust policy deve restringir repositório e Environment")
	}
}

func mustRel(dest, p string) string {
	rel, err := filepath.Rel(dest, p)
	if err != nil {
		return p
	}
	return filepath.ToSlash(rel)
}

func TestAngularWorkflows(t *testing.T) {
	dest, _ := genAngular(t, "Store Front")

	for _, name := range workflowFiles {
		text := readFile(t, dest, ".github/workflows/"+name)
		if staticAWSKeys.MatchString(text) {
			t.Errorf("%s: credencial estática da AWS", name)
		}
		for _, m := range usesLine.FindAllStringSubmatch(text, -1) {
			if !strings.HasPrefix(m[1], "./") && !pinnedSHA.MatchString(m[1]+"\n") {
				t.Errorf("%s: action sem SHA fixo: %s", name, m[1])
			}
		}
	}
	env := readFile(t, dest, ".github/workflows/deploy-env.yml")
	for _, want := range []string{
		"public,max-age=31536000,immutable", // assets com hash
		"no-cache, no-store, must-revalidate",
		"cloudfront create-invalidation",
		"${{ vars.AWS_ROLE_ARN }}", // OIDC
	} {
		if !strings.Contains(env, want) {
			t.Errorf("deploy-env.yml sem %q", want)
		}
	}
}

func TestAngularShellScriptIsExecutable(t *testing.T) {
	if !engine.IsExecutable("docker/40-config.sh") {
		t.Error("docker/40-config.sh precisa do bit de execução")
	}
}

func TestAngularInfraIsFormatted(t *testing.T) {
	bin, err := exec.LookPath("terraform")
	if err != nil {
		t.Skip("terraform não está no PATH")
	}
	dest, _ := genAngular(t, "Store Front")
	cmd := exec.Command(bin, "fmt", "-check", "-recursive", "-diff", "-no-color")
	cmd.Dir = filepath.Join(dest, "infra")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("terraform fmt reprovou:\n%s", out)
	}
}
