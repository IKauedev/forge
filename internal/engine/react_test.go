package engine_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ikauedev/forge/templates"
)

func genReact(t *testing.T) (string, string) {
	t.Helper()
	dest, n := genFrontend(t, templates.ReactApp, "Store Front")
	return dest, n.Kebab
}

func TestReactGenerationRenamesEverything(t *testing.T) {
	dest, kebab := genReact(t)

	_ = filepath.WalkDir(dest, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		b, _ := os.ReadFile(p)
		if m := angularLeftover.FindString(string(b)); m != "" {
			t.Errorf("%s: sobrou sentinela %q", p, m)
		}
		return nil
	})

	if !strings.Contains(readFile(t, dest, "package.json"), `"name": "`+kebab+`"`) {
		t.Errorf("package.json deveria ter o nome %q", kebab)
	}
	if !strings.Contains(readFile(t, dest, "index.html"), "<title>Store Front</title>") {
		t.Error("index.html deveria ter o título do projeto")
	}
	if !strings.Contains(readFile(t, dest, "infra/envs/dev/main.tf"), `app = "`+kebab+`"`) {
		t.Error("a infra deveria usar o nome do projeto")
	}
	if !strings.Contains(readFile(t, dest, "src/layout/Shell.tsx"), "Store Front") {
		t.Error("o shell deveria mostrar o nome do projeto")
	}
	if _, err := os.Stat(filepath.Join(dest, "template.forge.yaml")); err == nil {
		t.Error("o manifesto não pode ir para o projeto gerado")
	}
}

// Os guards espelham as anotações do backend; o nome de cada um precisa continuar existindo.
func TestReactGuardsMirrorBackendAnnotations(t *testing.T) {
	dest, _ := genReact(t)
	guards := readFile(t, dest, "src/core/auth/guards.tsx")
	for _, want := range []string{"RequireAuth", "RequireRole", "RequireAnyRole", "RequireAdmin"} {
		if !strings.Contains(guards, "export function "+want) {
			t.Errorf("guards.tsx sem %s", want)
		}
	}
	readme := readFile(t, dest, "README.md")
	for _, ann := range []string{"@IsAuthenticated", "@HasRole", "@HasAnyRole", "@IsAdmin"} {
		if !strings.Contains(readme, ann) {
			t.Errorf("README sem a anotação %s", ann)
		}
	}
}

func TestReactRuntimeConfigAndDeploy(t *testing.T) {
	dest, _ := genReact(t)

	cfg := readFile(t, dest, "public/config.json")
	for _, want := range []string{"apiBaseUrl", "rolesClaim", `"enabled": false`} {
		if !strings.Contains(cfg, want) {
			t.Errorf("config.json sem %q", want)
		}
	}
	if !strings.Contains(readFile(t, dest, "Dockerfile"), "/app/dist ") {
		t.Error("o Dockerfile deveria copiar o dist do Vite")
	}
	if !strings.Contains(readFile(t, dest, ".github/workflows/deploy.yml"), "path: dist\n") {
		t.Error("o deploy deveria publicar o dist do Vite")
	}
	env := readFile(t, dest, ".github/workflows/deploy-env.yml")
	for _, want := range []string{"rolesClaim: $roles", "public,max-age=31536000,immutable", "cloudfront create-invalidation", "${{ vars.AWS_ROLE_ARN }}"} {
		if !strings.Contains(env, want) {
			t.Errorf("deploy-env.yml sem %q", want)
		}
	}
	if !strings.Contains(readFile(t, dest, "docker/40-config.sh"), "AUTH_ROLES_CLAIM") {
		t.Error("o container deveria aceitar AUTH_ROLES_CLAIM")
	}
	if strings.Contains(readFile(t, dest, ".github/workflows/ci.yml"), "--watch=false") {
		t.Error("vitest run não aceita --watch=false do Angular")
	}

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
}

func TestReactInfraRules(t *testing.T) {
	dest, _ := genReact(t)
	site := readFile(t, dest, "infra/modules/static-site/main.tf")
	for _, want := range []string{"aws_cloudfront_origin_access_control", "error_code            = 403", "error_code            = 404", "aws.us_east_1", "AWS:SourceArn"} {
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
	_ = filepath.WalkDir(filepath.Join(dest, "infra"), func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		switch filepath.Ext(p) {
		case ".tf", ".hcl", ".example", ".tfvars":
			b, _ := os.ReadFile(p)
			if m := literalSecret.FindString(string(b)); m != "" {
				t.Errorf("%s: possível segredo em texto: %q", p, m)
			}
		}
		return nil
	})

	if bin, err := exec.LookPath("terraform"); err == nil {
		cmd := exec.Command(bin, "fmt", "-check", "-recursive", "-diff", "-no-color")
		cmd.Dir = filepath.Join(dest, "infra")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("terraform fmt reprovou:\n%s", out)
		}
	}
}

// O .forge.yaml é gravado pelo forge (indentação própria) e o CI dos frontends roda prettier --check.
func TestFrontendsIgnoreForgeYamlInPrettier(t *testing.T) {
	for _, tpl := range []string{templates.AngularApp, templates.ReactApp} {
		dest, _ := genFrontend(t, tpl, "Store Front")
		if !strings.Contains(readFile(t, dest, ".prettierignore"), ".forge.yaml") {
			t.Errorf("%s: .prettierignore deveria ignorar o .forge.yaml", tpl)
		}
	}
}
