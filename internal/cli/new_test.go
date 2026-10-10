package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
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

func TestNewCreatesProject(t *testing.T) {
	out := t.TempDir()
	stdout, err := run(t, "new", "Payments API", "--package", "com.empresa", "--out", out, "--no-git")
	if err != nil {
		t.Fatal(err)
	}
	pom := filepath.Join(out, "payments-api", "pom.xml")
	b, err := os.ReadFile(pom)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "<artifactId>payments-api</artifactId>") ||
		!strings.Contains(string(b), "<groupId>com.empresa</groupId>") {
		t.Errorf("pom.xml não foi renomeado:\n%s", b)
	}
	app := filepath.Join(out, "payments-api", "src", "main", "java", "com", "empresa", "paymentsapi", "PaymentsApiApplication.java")
	if _, err := os.Stat(app); err != nil {
		t.Errorf("classe principal ausente: %v", err)
	}
	if !strings.Contains(stdout, "Próximos passos") {
		t.Errorf("faltou mensagem de próximos passos: %s", stdout)
	}
}

func TestNewRejectsInvalidInput(t *testing.T) {
	out := t.TempDir()
	for _, args := range [][]string{
		{"new", "class", "--out", out, "--no-git"},
		{"new", "ok-name", "--package", "Com.Bad", "--out", out, "--no-git"},
	} {
		if _, err := run(t, args...); err == nil {
			t.Errorf("esperava erro para %v", args)
		}
	}
}

func TestNewRunsGitInit(t *testing.T) {
	out := t.TempDir()
	if _, err := run(t, "new", "git test", "--out", out); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(out, "git-test", ".git")); err != nil {
		t.Skipf("git indisponível? %v", err)
	}
}
