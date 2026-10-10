// Package upgrade compara um projeto gerado com o que o forge geraria hoje com as mesmas respostas.
package upgrade

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ikauedev/forge/internal/engine"
	"github.com/ikauedev/forge/internal/names"
	"github.com/ikauedev/forge/internal/project"
	"github.com/ikauedev/forge/internal/template"
)

// skipped são diretórios de build/dependências que não fazem parte do que o forge gera.
var skipped = map[string]bool{
	"node_modules": true, "target": true, ".git": true, "dist": true, ".angular": true, ".terraform": true,
}

// FileDiff é um arquivo que existe nos dois lados com conteúdo diferente.
type FileDiff struct {
	Path    string
	Unified string
}

// Result separa o que mudou entre o projeto atual e a geração nova.
type Result struct {
	Added         []string // só na geração nova (arquivo novo no template)
	OnlyInProject []string // só no projeto (criado por você ou removido do template)
	Changed       []FileDiff
}

// Empty informa se não há nenhuma diferença.
func (r Result) Empty() bool {
	return len(r.Added) == 0 && len(r.OnlyInProject) == 0 && len(r.Changed) == 0
}

// Diff regenera o projeto descrito por cfg num diretório temporário a partir de src (raiz do template)
// e o compara com dir. Nada em dir é alterado.
func Diff(src fs.FS, dir string, cfg project.Config) (Result, error) {
	n, err := names.New(cfg.Project.Name)
	if err != nil {
		return Result{}, fmt.Errorf("nome do projeto no %s: %w", project.File, err)
	}
	tmp, err := os.MkdirTemp("", "forge-upgrade-")
	if err != nil {
		return Result{}, err
	}
	defer func() { _ = os.RemoveAll(tmp) }()

	fresh := filepath.Join(tmp, n.Kebab)
	sel := template.Selection(cfg.Options)
	if _, err := engine.Generate(src, engine.Options{Names: n, Package: cfg.Project.Package, Dest: fresh, Selection: sel}); err != nil {
		return Result{}, fmt.Errorf("regenerando: %w", err)
	}

	current, err := readTree(dir)
	if err != nil {
		return Result{}, err
	}
	generated, err := readTree(fresh)
	if err != nil {
		return Result{}, err
	}

	var r Result
	for _, p := range sortedKeys(generated) {
		cur, ok := current[p]
		switch {
		case !ok:
			r.Added = append(r.Added, p)
		case !bytes.Equal(cur, generated[p]):
			r.Changed = append(r.Changed, FileDiff{Path: p, Unified: unified(p, cur, generated[p])})
		}
	}
	for _, p := range sortedKeys(current) {
		if _, ok := generated[p]; !ok {
			r.OnlyInProject = append(r.OnlyInProject, p)
		}
	}
	return r, nil
}

// readTree lê todos os arquivos de root (caminhos com "/"), ignorando build e o .forge.yaml.
func readTree(root string) (map[string][]byte, error) {
	out := map[string][]byte{}
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if skipped[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if rel == project.File {
			return nil
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		out[rel] = b
		return nil
	})
	return out, err
}

func sortedKeys(m map[string][]byte) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// unified produz um diff no formato unificado (contexto de 3 linhas) entre old (projeto) e new (geração nova).
func unified(path string, oldB, newB []byte) string {
	if looksBinary(oldB) || looksBinary(newB) {
		return fmt.Sprintf("Arquivos binários %s diferem\n", path)
	}
	a, b := splitLines(string(oldB)), splitLines(string(newB))
	ops := diffOps(a, b)

	var sb strings.Builder
	fmt.Fprintf(&sb, "--- projeto/%s\n+++ novo/%s\n", path, path)
	const ctx = 3
	var changes []int
	for i, o := range ops {
		if o.kind != ' ' {
			changes = append(changes, i)
		}
	}
	for k := 0; k < len(changes); {
		first, last := changes[k], changes[k]
		k++
		// muda de hunk só quando há mais de 2*ctx linhas iguais entre duas mudanças
		for k < len(changes) && changes[k]-last-1 <= 2*ctx {
			last = changes[k]
			k++
		}
		start, end := max(first-ctx, 0), min(last+ctx+1, len(ops))

		oldStart, newStart, oldN, newN := 0, 0, 0, 0
		for _, o := range ops[:start] {
			if o.kind != '+' {
				oldStart++
			}
			if o.kind != '-' {
				newStart++
			}
		}
		for _, o := range ops[start:end] {
			if o.kind != '+' {
				oldN++
			}
			if o.kind != '-' {
				newN++
			}
		}
		fmt.Fprintf(&sb, "@@ -%d,%d +%d,%d @@\n", oldStart+1, oldN, newStart+1, newN)
		for _, o := range ops[start:end] {
			fmt.Fprintf(&sb, "%c%s\n", o.kind, o.line)
		}
	}
	return sb.String()
}

type op struct {
	kind byte // ' ', '-' (só no projeto), '+' (só na geração nova)
	line string
}

// diffOps calcula a menor sequência de edições por LCS. Os arquivos gerados são pequenos o bastante
// para a tabela O(n*m).
func diffOps(a, b []string) []op {
	n, m := len(a), len(b)
	lcs := make([][]int, n+1)
	for i := range lcs {
		lcs[i] = make([]int, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if a[i] == b[j] {
				lcs[i][j] = lcs[i+1][j+1] + 1
			} else {
				lcs[i][j] = max(lcs[i+1][j], lcs[i][j+1])
			}
		}
	}
	var ops []op
	i, j := 0, 0
	for i < n && j < m {
		switch {
		case a[i] == b[j]:
			ops = append(ops, op{' ', a[i]})
			i++
			j++
		case lcs[i+1][j] >= lcs[i][j+1]:
			ops = append(ops, op{'-', a[i]})
			i++
		default:
			ops = append(ops, op{'+', b[j]})
			j++
		}
	}
	for ; i < n; i++ {
		ops = append(ops, op{'-', a[i]})
	}
	for ; j < m; j++ {
		ops = append(ops, op{'+', b[j]})
	}
	return ops
}

func splitLines(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.TrimSuffix(s, "\n")
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

func looksBinary(b []byte) bool {
	probe := b
	if len(probe) > 8000 {
		probe = probe[:8000]
	}
	return bytes.IndexByte(probe, 0) >= 0
}
