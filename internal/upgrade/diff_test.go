package upgrade_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/ikauedev/forge/internal/engine"
	"github.com/ikauedev/forge/internal/names"
	"github.com/ikauedev/forge/internal/project"
	"github.com/ikauedev/forge/internal/template"
	"github.com/ikauedev/forge/internal/upgrade"
)

const manifest = `name: fixture
sentinel: starter-app
options:
  - name: mode
    values: [a, b]
    default: a
`

func fixture(files map[string]string) fs.FS {
	m := fstest.MapFS{"template.forge.yaml": {Data: []byte(manifest)}}
	for p, c := range files {
		m[p] = &fstest.MapFile{Data: []byte(c)}
	}
	return m
}

var v1 = map[string]string{
	"README.md":    "# Starter App\n\nlinha 1\nlinha 2\nlinha 3\nlinha 4\nlinha 5\nlinha 6\nlinha 7\nlinha 8\n",
	"src/main.txt": "name=starter-app\n",
	"mode.txt":     "# forge:if mode=a\nmodo A\n# forge:else\n#~ modo B\n# forge:end\n",
}

// generate cria um projeto com a versão 1 do template, como o forge new faria.
func generate(t *testing.T, opts map[string]string) (string, project.Config) {
	t.Helper()
	n, err := names.New("Store Front")
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(t.TempDir(), n.Kebab)
	if _, err := engine.Generate(fixture(v1), engine.Options{Names: n, Dest: dir, Selection: template.Selection(opts)}); err != nil {
		t.Fatal(err)
	}
	cfg := project.Config{Template: "fixture", Project: project.ProjectInfo{Name: "Store Front"}, Options: opts}
	if err := project.Write(dir, cfg); err != nil {
		t.Fatal(err)
	}
	return dir, cfg
}

func TestNoDifferencesWithSameTemplate(t *testing.T) {
	dir, cfg := generate(t, map[string]string{"mode": "a"})
	r, err := upgrade.Diff(fixture(v1), dir, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !r.Empty() {
		t.Errorf("não deveria haver diferenças: %+v", r)
	}
}

func TestShowsRealDifferencesBetweenTemplateVersions(t *testing.T) {
	dir, cfg := generate(t, map[string]string{"mode": "b"})
	// o usuário criou um arquivo próprio e editou outro
	if err := os.WriteFile(filepath.Join(dir, "minha-nota.txt"), []byte("meu\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// build e dependências nunca entram na comparação
	if err := os.MkdirAll(filepath.Join(dir, "node_modules", "x"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "node_modules", "x", "i.js"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	v2 := map[string]string{
		"README.md":    "# Starter App v2\n\nlinha 1\nlinha 2\nlinha 3\nlinha 4\nlinha 5\nlinha 6\nlinha 7\nlinha 8\nlinha 9\n",
		"src/main.txt": "name=starter-app\n",
		"mode.txt":     v1["mode.txt"],
		"docs/NEW.md":  "novo no template\n",
	}
	r, err := upgrade.Diff(fixture(v2), dir, cfg)
	if err != nil {
		t.Fatal(err)
	}

	if len(r.Added) != 1 || r.Added[0] != "docs/NEW.md" {
		t.Errorf("Added = %v", r.Added)
	}
	if len(r.OnlyInProject) != 1 || r.OnlyInProject[0] != "minha-nota.txt" {
		t.Errorf("OnlyInProject = %v", r.OnlyInProject)
	}
	if len(r.Changed) != 1 || r.Changed[0].Path != "README.md" {
		t.Fatalf("Changed = %+v", r.Changed)
	}
	u := r.Changed[0].Unified
	for _, want := range []string{"-# Store Front", "+# Store Front v2", "+linha 9"} {
		if !strings.Contains(u, want) {
			t.Errorf("diff sem %q:\n%s", want, u)
		}
	}
	// as duas mudanças ficam longe o bastante para virar dois hunks
	if strings.Count(u, "@@ -") != 2 {
		t.Errorf("esperava 2 hunks:\n%s", u)
	}
}

func TestUsesTheRecordedOptions(t *testing.T) {
	dir, cfg := generate(t, map[string]string{"mode": "b"})
	cfg.Options = map[string]string{"mode": "a"} // respostas diferentes das usadas na geração
	r, err := upgrade.Diff(fixture(v1), dir, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Changed) != 1 || r.Changed[0].Path != "mode.txt" {
		t.Errorf("Changed = %+v", r.Changed)
	}
}

func TestDoesNotModifyTheProject(t *testing.T) {
	dir, cfg := generate(t, map[string]string{"mode": "a"})
	before, _ := os.ReadFile(filepath.Join(dir, "README.md"))
	v2 := map[string]string{"README.md": "outro\n"}
	if _, err := upgrade.Diff(fixture(v2), dir, cfg); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(filepath.Join(dir, "README.md"))
	if string(before) != string(after) {
		t.Error("o projeto foi alterado")
	}
}
