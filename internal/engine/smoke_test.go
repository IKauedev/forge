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
	{"http": "both", "auth": "jwt", "cache": "none"},   // padrões
	{"http": "feign", "auth": "none", "cache": "none"}, // enxuto
	{"http": "webclient", "auth": "cognito", "cache": "caffeine"},
	{"http": "feign", "auth": "jwt", "cache": "redis"},
}

// TestSmokeVerify gera cada projeto da matriz e roda ./mvnw verify nele.
// Requer JDK 25 (JAVA_HOME) e Docker. Execute: go test -tags smoke -timeout 40m ./internal/engine
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

			mvnw := "./mvnw"
			if runtime.GOOS == "windows" {
				mvnw = filepath.Join(dest, "mvnw.cmd")
			}
			cmd := exec.Command(mvnw, "-q", "verify")
			cmd.Dir = dest
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("mvnw verify falhou: %v\n%s", err, out)
			}
		})
	}
}
