package engine_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

var (
	workflowFiles   = []string{"ci.yml", "deploy.yml", "deploy-env.yml"}
	staticAWSKeys   = regexp.MustCompile(`(?i)aws_access_key_id|aws_secret_access_key|aws-access-key-id|aws-secret-access-key|aws_session_token|secrets\.aws`)
	pinnedSHA       = regexp.MustCompile(`@[0-9a-f]{40}(\s|$)`)
	usesLine        = regexp.MustCompile(`(?m)^\s*(?:-\s+)?uses:\s*(\S+)`)
	oidcActionUses  = "aws-actions/configure-aws-credentials"
	localDeployUses = "./.github/workflows/deploy-env.yml"
)

// checkWorkflows valida as regras de segurança dos workflows de um projeto gerado.
func checkWorkflows(t *testing.T, dest, kebab string) {
	t.Helper()
	for _, name := range workflowFiles {
		rel := filepath.Join(".github", "workflows", name)
		b, err := os.ReadFile(filepath.Join(dest, rel))
		if err != nil {
			t.Errorf("workflow ausente: %s", rel)
			continue
		}
		text := string(b)

		if staticAWSKeys.MatchString(text) {
			t.Errorf("%s: credencial estática da AWS", rel)
		}
		for _, m := range usesLine.FindAllStringSubmatch(text, -1) {
			ref := m[1]
			if strings.HasPrefix(ref, "./") {
				continue
			}
			if !pinnedSHA.MatchString(ref + "\n") {
				t.Errorf("%s: action sem SHA fixo: %s", rel, ref)
			}
		}
		if regexp.MustCompile(`(?i)starter`).MatchString(text) {
			t.Errorf("%s: sobrou sentinela", rel)
		}

		var doc struct {
			Permissions map[string]string `yaml:"permissions"`
			Jobs        map[string]struct {
				Permissions map[string]string `yaml:"permissions"`
				Uses        string            `yaml:"uses"`
				Steps       []struct {
					Uses string `yaml:"uses"`
				} `yaml:"steps"`
			} `yaml:"jobs"`
		}
		if err := yaml.Unmarshal(b, &doc); err != nil {
			t.Errorf("%s: YAML inválido: %v", rel, err)
			continue
		}
		if name != "deploy-env.yml" && doc.Permissions["contents"] != "read" {
			t.Errorf("%s: permissions de topo deve ser contents: read", rel)
		}
		for jobName, job := range doc.Jobs {
			if job.Permissions["id-token"] != "write" {
				continue
			}
			usesOIDC := job.Uses == localDeployUses
			for _, s := range job.Steps {
				if strings.HasPrefix(s.Uses, oidcActionUses+"@") {
					usesOIDC = true
				}
			}
			if !usesOIDC {
				t.Errorf("%s: job %q pede id-token: write sem usar OIDC", rel, jobName)
			}
		}
	}

	// o nome do projeto entra nas convenções que a infra usa
	deploy, _ := os.ReadFile(filepath.Join(dest, ".github", "workflows", "deploy-env.yml"))
	if !strings.Contains(string(deploy), "'"+kebab+"'") || !strings.Contains(string(deploy), "CONTAINER_NAME: "+kebab) {
		t.Errorf("deploy-env.yml deveria usar o nome %q", kebab)
	}
}

func TestPipelineDocShipsWithProject(t *testing.T) {
	dest := gen(t)
	b, err := os.ReadFile(filepath.Join(dest, "docs", "PIPELINE.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"AWS_ROLE_ARN", "AWS_REGION", "Required reviewers", "orders-api-dev", "OIDC"} {
		if !strings.Contains(string(b), want) {
			t.Errorf("PIPELINE.md deveria citar %q", want)
		}
	}
	if regexp.MustCompile(`(?i)starter`).Match(b) {
		t.Error("PIPELINE.md com sentinela sobrando")
	}
}

// Roda o actionlint num projeto gerado, quando ele está instalado (go install github.com/rhysd/actionlint/cmd/actionlint@v1.7.12).
func TestActionlintOnGeneratedProject(t *testing.T) {
	bin, err := exec.LookPath("actionlint")
	if err != nil {
		t.Skip("actionlint não está no PATH")
	}
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git não está no PATH (o actionlint precisa de um repositório)")
	}
	dest := gen(t)
	if out, err := exec.Command("git", "-C", dest, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	cmd := exec.Command(bin)
	cmd.Dir = dest
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("actionlint reprovou os workflows:\n%s", out)
	}
}
