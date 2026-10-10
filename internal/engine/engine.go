// Package engine copia um template de referência aplicando a renomeação por sentinelas.
package engine

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/ikauedeveloper/forge/internal/names"
	"github.com/ikauedeveloper/forge/internal/template"
)

// Valores de referência presentes nos templates.
const (
	refPackage = "com.example"
	refFlat    = "starterapi"
)

// Options configura uma geração.
type Options struct {
	Names   names.Names
	Package string // pacote base Java, ex.: com.empresa
	Dest    string // diretório de destino (será criado; deve estar vazio ou não existir)

	// Selection traz as escolhas das opções do template (http, auth, cache...).
	// O que faltar é completado com os padrões do manifesto.
	Selection template.Selection
}

var skippedDirs = map[string]bool{"target": true, ".git": true, "node_modules": true}

var binaryExts = map[string]bool{
	".jar": true, ".class": true, ".png": true, ".jpg": true, ".jpeg": true, ".gif": true,
	".ico": true, ".woff": true, ".woff2": true, ".ttf": true, ".otf": true, ".eot": true,
	".zip": true, ".gz": true, ".pdf": true,
}

// Replacer devolve o substituidor das sentinelas. A ordem importa: strings.Replacer dá
// prioridade ao par que aparece primeiro, então os mais longos/específicos vêm antes.
func Replacer(n names.Names, pkg string) *strings.Replacer {
	pkgPath := strings.ReplaceAll(pkg, ".", "/")
	return strings.NewReplacer(
		refPackage+"."+refFlat, pkg+"."+n.Flat,
		strings.ReplaceAll(refPackage, ".", "/")+"/"+refFlat, pkgPath+"/"+n.Flat,
		refPackage, pkg,
		strings.ReplaceAll(refPackage, ".", "/"), pkgPath,
		"STARTER_API", n.Upper,
		"StarterApi", n.Pascal,
		"starter-api", n.Kebab,
		"starter_api", n.Snake,
		"Starter Api", n.Title,
		refFlat, n.Flat,
	)
}

// IsExecutable informa se o arquivo gerado precisa do bit de execução. O embed.FS não preserva
// modos, então a regra é por nome.
func IsExecutable(name string) bool {
	base := path.Base(name)
	return base == "mvnw" || strings.HasSuffix(base, ".sh")
}

// IsBinary decide se o conteúdo deve ser copiado sem renderização.
func IsBinary(name string, content []byte) bool {
	if binaryExts[strings.ToLower(path.Ext(name))] {
		return true
	}
	probe := content
	if len(probe) > 8000 {
		probe = probe[:8000]
	}
	return bytes.IndexByte(probe, 0) >= 0
}

// Generate copia src (raiz do template) para opts.Dest e devolve os caminhos gerados.
func Generate(src fs.FS, opts Options) ([]string, error) {
	if err := names.ValidatePackage(opts.Package); err != nil {
		return nil, err
	}
	if err := ensureEmptyDest(opts.Dest); err != nil {
		return nil, err
	}
	manifest, err := LoadManifest(src)
	if err != nil {
		return nil, err
	}
	sel, err := manifest.Resolve(opts.Selection)
	if err != nil {
		return nil, err
	}
	removed, err := manifest.Removed(sel)
	if err != nil {
		return nil, err
	}
	rep := Replacer(opts.Names, opts.Package)
	slots, overrides, err := databaseAssets(sel)
	if err != nil {
		return nil, err
	}

	var generated []string
	emit := func(p string, content []byte) error {
		if !IsBinary(p, content) {
			processed, err := ProcessDirectivesWithSlots(p, string(content), sel, slots)
			if err != nil {
				return err
			}
			content = []byte(rep.Replace(processed))
		}
		rel := rep.Replace(p)
		target := filepath.Join(opts.Dest, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		mode := os.FileMode(0o644)
		if IsExecutable(p) {
			mode = 0o755
		}
		if err := os.WriteFile(target, content, mode); err != nil {
			return err
		}
		if err := os.Chmod(target, mode); err != nil {
			return err
		}
		generated = append(generated, rel)
		return nil
	}

	seen := map[string]bool{}
	err = fs.WalkDir(src, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if skippedDirs[d.Name()] || template.IsRemoved(p+"/", removed) {
				return fs.SkipDir
			}
			return nil
		}
		if p == template.ManifestFile || template.IsRemoved(p, removed) {
			return nil
		}
		content, err := fs.ReadFile(src, p)
		if err != nil {
			return err
		}
		if ov, ok := overrides[p]; ok {
			content, seen[p] = ov, true
		}
		return emit(p, content)
	})
	if err != nil {
		return nil, fmt.Errorf("gerando projeto: %w", err)
	}
	// arquivos do catálogo que o template de referência não tem (ex.: migrations extras)
	for p, content := range overrides {
		if !seen[p] && !template.IsRemoved(p, removed) {
			if err := emit(p, content); err != nil {
				return nil, fmt.Errorf("gerando projeto: %w", err)
			}
		}
	}
	return generated, nil
}

func ensureEmptyDest(dest string) error {
	entries, err := os.ReadDir(dest)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return nil
	case err != nil:
		return err
	case len(entries) > 0:
		return fmt.Errorf("o diretório %q já existe e não está vazio", dest)
	}
	return nil
}
