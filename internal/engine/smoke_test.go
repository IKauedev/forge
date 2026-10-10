//go:build smoke

package engine_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// TestSmokeVerify gera um projeto e roda ./mvnw verify nele.
// Requer JDK 25 (JAVA_HOME) e Docker. Execute: go test -tags smoke -timeout 20m ./internal/engine
func TestSmokeVerify(t *testing.T) {
	if os.Getenv("JAVA_HOME") == "" {
		t.Skip("JAVA_HOME não definido (precisa de JDK 25)")
	}
	dest := filepath.Join(t.TempDir(), "payments-api")
	generate(t, "Payments API", "com.empresa", dest)

	mvnw := "./mvnw"
	if runtime.GOOS == "windows" {
		mvnw = filepath.Join(dest, "mvnw.cmd")
	}
	cmd := exec.Command(mvnw, "-q", "verify")
	cmd.Dir = dest
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("mvnw verify falhou: %v\n%s", err, out)
	}
}
