// Package template lê o manifesto (template.forge.yaml) que descreve as opções de um template.
package template

import (
	"errors"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// ManifestFile é o nome do manifesto na raiz do template. Nunca é copiado para o projeto gerado.
const ManifestFile = "template.forge.yaml"

// Option descreve uma opção de variação (ex.: http, auth, cache).
type Option struct {
	Name        string   `yaml:"name"`
	Values      []string `yaml:"values"`
	Default     string   `yaml:"default"`
	Description string   `yaml:"description"`

	// ExtendWithCatalog indica que os valores vêm do catálogo de bancos (somados aos de Values,
	// como "none"). O engine preenche isso via ExtendOption.
	ExtendWithCatalog bool `yaml:"extend_with_catalog"`
}

// Removal lista caminhos (relativos ao template) removidos quando a condição é verdadeira.
// Um caminho terminado em "/" remove o diretório inteiro.
type Removal struct {
	When  string   `yaml:"when"`
	Paths []string `yaml:"paths"`
}

// Manifest é o conteúdo de template.forge.yaml.
type Manifest struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Options     []Option          `yaml:"options"`
	Reference   map[string]string `yaml:"reference"`
	Remove      []Removal         `yaml:"remove"`
}

// Selection é a escolha final de cada opção (todas preenchidas e validadas).
type Selection map[string]string

// Load lê o manifesto na raiz de src. Se ele não existir, devolve um manifesto vazio.
func Load(src fs.FS) (*Manifest, error) {
	data, err := fs.ReadFile(src, ManifestFile)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return &Manifest{}, nil
		}
		return nil, err
	}
	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("%s inválido: %w", ManifestFile, err)
	}
	return &m, m.Validate()
}

// ExtendOption põe values (ex.: nomes do catálogo) à frente dos valores já declarados na opção.
func (m *Manifest) ExtendOption(name string, values []string) error {
	for i, o := range m.Options {
		if o.Name == name {
			m.Options[i].Values = append(append([]string{}, values...), o.Values...)
			return m.Validate()
		}
	}
	return fmt.Errorf("opção %q não existe", name)
}

// Validate confere nomes, valores e padrões das opções.
func (m *Manifest) Validate() error {
	for _, o := range m.Options {
		if o.Name == "" || (len(o.Values) == 0 && !o.ExtendWithCatalog) {
			return fmt.Errorf("opção %q sem nome ou valores", o.Name)
		}
		if o.ExtendWithCatalog && len(o.Values) <= 1 {
			continue // ainda não estendida
		}
		if !contains(o.Values, o.Default) {
			return fmt.Errorf("opção %q: padrão %q não está em %v", o.Name, o.Default, o.Values)
		}
	}
	return nil
}

// Option devolve a opção pelo nome.
func (m *Manifest) Option(name string) (Option, bool) {
	for _, o := range m.Options {
		if o.Name == name {
			return o, true
		}
	}
	return Option{}, false
}

// Defaults devolve a seleção com o valor padrão de cada opção.
func (m *Manifest) Defaults() Selection {
	sel := Selection{}
	for _, o := range m.Options {
		sel[o.Name] = o.Default
	}
	return sel
}

// Resolve completa com os padrões o que faltar em choices e valida o resultado.
func (m *Manifest) Resolve(choices map[string]string) (Selection, error) {
	sel := m.Defaults()
	for k, v := range choices {
		if v == "" {
			continue
		}
		o, ok := m.Option(k)
		if !ok {
			return nil, fmt.Errorf("opção desconhecida %q", k)
		}
		if !contains(o.Values, v) {
			return nil, fmt.Errorf("valor %q inválido para --%s; use um de: %s", v, k, strings.Join(o.Values, ", "))
		}
		sel[k] = v
	}
	return sel, nil
}

// Removed devolve os caminhos removidos para a seleção dada.
func (m *Manifest) Removed(sel Selection) ([]string, error) {
	var out []string
	for _, r := range m.Remove {
		ok, err := Eval(r.When, sel)
		if err != nil {
			return nil, err
		}
		if ok {
			out = append(out, r.Paths...)
		}
	}
	return out, nil
}

// IsRemoved informa se p (caminho relativo, com "/") está coberto por algum caminho removido.
func IsRemoved(p string, removed []string) bool {
	for _, r := range removed {
		if strings.HasSuffix(r, "/") {
			if strings.HasPrefix(p, r) {
				return true
			}
		} else if p == r {
			return true
		}
	}
	return false
}

// Eval avalia condições como "auth!=none", "http=feign,both", termos ligados por "&&" e grupos
// ligados por "||" ("&&" tem precedência: "a=1 && b=2 || c=3").
func Eval(cond string, sel Selection) (bool, error) {
	for _, group := range strings.Split(cond, "||") {
		ok, err := evalAll(group, sel)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}

func evalAll(cond string, sel Selection) (bool, error) {
	for _, part := range strings.Split(cond, "&&") {
		part = strings.TrimSpace(part)
		if part == "" {
			return false, fmt.Errorf("condição vazia em %q", cond)
		}
		negate := false
		key, vals, found := strings.Cut(part, "!=")
		if found {
			negate = true
		} else {
			key, vals, found = strings.Cut(part, "=")
		}
		if !found {
			return false, fmt.Errorf("condição inválida %q (use chave=valor ou chave!=valor)", part)
		}
		key = strings.TrimSpace(key)
		current, ok := sel[key]
		if !ok {
			return false, fmt.Errorf("condição usa a opção desconhecida %q", key)
		}
		match := contains(strings.Split(strings.TrimSpace(vals), ","), current)
		if match == negate {
			return false, nil
		}
	}
	return true, nil
}

// SortedKeys devolve as chaves da seleção em ordem estável.
func (s Selection) SortedKeys() []string {
	keys := make([]string, 0, len(s))
	for k := range s {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func contains(list []string, v string) bool {
	for _, x := range list {
		if strings.TrimSpace(x) == v {
			return true
		}
	}
	return false
}
