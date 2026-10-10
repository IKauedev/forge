package engine_test

import (
	"flag"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/ikauedeveloper/forge/internal/engine"
	"github.com/ikauedeveloper/forge/internal/names"
	"github.com/ikauedeveloper/forge/templates"
)

var update = flag.Bool("update", false, "regenera testdata/golden")

const goldenDir = "../../testdata/golden/payments-api"

func generate(t *testing.T, raw, pkg, dest string) []string {
	t.Helper()
	n, err := names.New(raw)
	if err != nil {
		t.Fatal(err)
	}
	src, err := fs.Sub(templates.FS, templates.SpringAPI)
	if err != nil {
		t.Fatal(err)
	}
	files, err := engine.Generate(src, engine.Options{Names: n, Package: pkg, Dest: dest})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

func TestReplacerOrder(t *testing.T) {
	n, _ := names.New("Payments API")
	rep := engine.Replacer(n, "com.empresa")
	tests := map[string]string{
		"com.example.starterapi.domain":    "com.empresa.paymentsapi.domain",
		"com/example/starterapi/Foo.java":  "com/empresa/paymentsapi/Foo.java",
		"<groupId>com.example</groupId>":   "<groupId>com.empresa</groupId>",
		"StarterApiApplication":            "PaymentsApiApplication",
		"starter-api STARTER_API":          "payments-api PAYMENTS_API",
		"starter_api / Starter Api":        "payments_api / Payments Api",
		"spring-boot-starter-webmvc":       "spring-boot-starter-webmvc",
		"org.springframework.boot.starter": "org.springframework.boot.starter",
	}
	for in, want := range tests {
		if got := rep.Replace(in); got != want {
			t.Errorf("Replace(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNoSentinelLeftovers(t *testing.T) {
	dest := t.TempDir()
	files := generate(t, "Payments API", "com.empresa", filepath.Join(dest, "out"))
	leftover := regexp.MustCompile(`(?i)starter[-_ ]?api`)

	for _, rel := range files {
		if leftover.MatchString(rel) || strings.Contains(rel, "com/example") {
			t.Errorf("sobra sentinela no caminho %q", rel)
		}
		content, err := os.ReadFile(filepath.Join(dest, "out", filepath.FromSlash(rel)))
		if err != nil {
			t.Fatal(err)
		}
		if engine.IsBinary(rel, content) {
			continue
		}
		if leftover.Match(content) {
			t.Errorf("sobra sentinela no conteúdo de %q", rel)
		}
		if strings.Contains(string(content), "com.example") {
			t.Errorf("sobra com.example em %q", rel)
		}
	}
}

func TestExecutableBit(t *testing.T) {
	for name, want := range map[string]bool{
		"mvnw": true, "scripts/run.sh": true, "mvnw.cmd": false, "pom.xml": false,
	} {
		if got := engine.IsExecutable(name); got != want {
			t.Errorf("IsExecutable(%q) = %v, want %v", name, got, want)
		}
	}
	if runtime.GOOS == "windows" {
		t.Skip("bits de permissão POSIX não se aplicam no Windows")
	}
	dest := filepath.Join(t.TempDir(), "out")
	generate(t, "Payments API", "com.empresa", dest)
	info, err := os.Stat(filepath.Join(dest, "mvnw"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0o111 == 0 {
		t.Errorf("mvnw perdeu o bit de execução: %v", info.Mode())
	}
}

func TestBinaryFilesAreNotRendered(t *testing.T) {
	if !engine.IsBinary("a/logo.png", []byte("starter-api")) {
		t.Error("extensão .png deve ser binária")
	}
	if !engine.IsBinary("a/blob", []byte{'s', 0, 'x'}) {
		t.Error("conteúdo com NUL deve ser binário")
	}
	if engine.IsBinary("pom.xml", []byte("<project/>")) {
		t.Error("texto não deve ser binário")
	}
}

func TestDestMustBeEmpty(t *testing.T) {
	dest := t.TempDir()
	if err := os.WriteFile(filepath.Join(dest, "x"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	n, _ := names.New("Payments API")
	src, _ := fs.Sub(templates.FS, templates.SpringAPI)
	if _, err := engine.Generate(src, engine.Options{Names: n, Package: "com.empresa", Dest: dest}); err == nil {
		t.Fatal("esperava erro para destino não vazio")
	}
}

func TestSkipsBuildOutput(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "out")
	for _, rel := range generate(t, "Payments API", "com.empresa", dest) {
		if strings.HasPrefix(rel, "target/") {
			t.Errorf("target/ não deve ser gerado: %s", rel)
		}
	}
}

func TestGolden(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "payments-api")
	generate(t, "Payments API", "com.empresa", dest)

	if *update {
		if err := os.RemoveAll(goldenDir); err != nil {
			t.Fatal(err)
		}
		copyTree(t, dest, goldenDir)
		return
	}

	got, want := readTree(t, dest), readTree(t, goldenDir)
	for p, w := range want {
		g, ok := got[p]
		if !ok {
			t.Errorf("arquivo ausente no gerado: %s", p)
			continue
		}
		if g != w {
			t.Errorf("conteúdo diferente: %s", p)
		}
	}
	for p := range got {
		if _, ok := want[p]; !ok {
			t.Errorf("arquivo extra no gerado: %s", p)
		}
	}
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

func copyTree(t *testing.T, from, to string) {
	t.Helper()
	for rel, content := range readTree(t, from) {
		target := filepath.Join(to, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(target, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}
