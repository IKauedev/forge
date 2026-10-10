//go:build smoke

package engine_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/ikauedeveloper/forge/internal/template"
)

// smokeMatrix cobre cada valor de cada opção ao menos uma vez, sem o produto cartesiano.
var smokeMatrix = []template.Selection{
	{"db": "postgres", "http": "both", "auth": "jwt", "cache": "none"}, // padrões
	{"db": "mysql", "http": "feign", "auth": "none", "cache": "none"},
	{"db": "mariadb", "http": "webclient", "auth": "cognito", "cache": "caffeine"},
	{"db": "none", "http": "feign", "auth": "jwt", "cache": "redis"},
	{"db": "none", "http": "webclient", "auth": "none", "cache": "none"}, // sem compose.yaml
}

func run(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s %v falhou: %v\n%s", name, args, err, out)
	}
}

func composeArgs(dest string) []string {
	if _, err := os.Stat(filepath.Join(dest, "compose.yaml")); err == nil {
		return []string{"compose", "-f", "compose.yaml", "-f", "compose.full.yaml"}
	}
	return []string{"compose", "-f", "compose.full.yaml"}
}

// TestSmokeVerify gera cada projeto da matriz, valida o compose, sobe o banco com healthcheck e
// roda ./mvnw verify. Requer JDK 25 (JAVA_HOME) e Docker.
// Execute: go test -tags smoke -timeout 60m ./internal/engine
func TestSmokeVerify(t *testing.T) {
	if os.Getenv("JAVA_HOME") == "" {
		t.Skip("JAVA_HOME não definido (precisa de JDK 25)")
	}
	for _, sel := range smokeMatrix {
		t.Run(label(sel), func(t *testing.T) {
			// t.TempDir() embute o nome do subteste ("auth=jwt,..."), que o cmd.exe não aceita no caminho.
			tmp, err := os.MkdirTemp("", "forge-smoke-")
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = os.RemoveAll(tmp) })
			dest := filepath.Join(tmp, "payments-api")
			generateWith(t, "Payments API", "com.empresa", dest, sel)

			run(t, dest, "docker", append(composeArgs(dest), "config", "-q")...)

			if _, err := exec.LookPath("actionlint"); err == nil {
				run(t, dest, "git", "init", "-q")
				run(t, dest, "actionlint")
			}

			if sel["db"] != "none" {
				up := append(composeArgs(dest), "up", "-d", "--wait", "--wait-timeout", "180", "db")
				t.Cleanup(func() { run(t, dest, "docker", append(composeArgs(dest), "down", "-v")...) })
				run(t, dest, "docker", up...)
				run(t, dest, "docker", append(composeArgs(dest), "down", "-v")...)
			}

			mvnw := "./mvnw"
			if runtime.GOOS == "windows" {
				mvnw = filepath.Join(dest, "mvnw.cmd")
			}
			run(t, dest, mvnw, "-q", "verify")
		})
	}
}
