// Package catalog lê o catálogo de bancos (catalog/databases.yaml) e gera os trechos do template.
package catalog

import (
	"bytes"
	"fmt"
	"io/fs"
	"path"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"

	data "github.com/ikauedeveloper/forge/catalog"
)

// None é o valor da opção db que dispensa banco.
const None = "none"

// Database descreve um banco suportado.
type Database struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Image       string `yaml:"image"`
	Port        int    `yaml:"port"`
	JDBCScheme  string `yaml:"jdbc_scheme"`
	Driver      struct {
		Group    string `yaml:"group"`
		Artifact string `yaml:"artifact"`
	} `yaml:"driver"`
	FlywayModule   string `yaml:"flyway_module"`
	Testcontainers struct {
		Artifact string `yaml:"artifact"`
		Class    string `yaml:"class"`
	} `yaml:"testcontainers"`
	Env struct {
		Database string   `yaml:"database"`
		User     string   `yaml:"user"`
		Password string   `yaml:"password"`
		Extra    []string `yaml:"extra"`
	} `yaml:"env"`
	DataDir     string `yaml:"data_dir"`
	Healthcheck string `yaml:"healthcheck"`
	AWS         struct {
		Engine  string `yaml:"engine"`
		Version string `yaml:"version"`
		Family  string `yaml:"family"`
		Port    int    `yaml:"port"`
	} `yaml:"aws"`
}

// Catalog é o conjunto de bancos, na ordem do YAML.
type Catalog struct {
	Databases []Database `yaml:"databases"`
	fsys      fs.FS
}

// Load lê o catálogo embutido.
func Load() (*Catalog, error) { return LoadFS(data.FS) }

// LoadFS lê um catálogo de fsys (databases.yaml e migrations/<nome>/). Usado nos testes.
func LoadFS(fsys fs.FS) (*Catalog, error) {
	raw, err := fs.ReadFile(fsys, "databases.yaml")
	if err != nil {
		return nil, err
	}
	var c Catalog
	if err := yaml.Unmarshal(raw, &c); err != nil {
		return nil, fmt.Errorf("databases.yaml inválido: %w", err)
	}
	c.fsys = fsys
	return &c, c.Validate()
}

// Validate confere campos obrigatórios, nomes únicos e a existência das migrations.
func (c *Catalog) Validate() error {
	if len(c.Databases) == 0 {
		return fmt.Errorf("catálogo vazio")
	}
	seen := map[string]bool{}
	for _, d := range c.Databases {
		if d.Name == "" {
			return fmt.Errorf("banco sem nome")
		}
		if d.Name == None {
			return fmt.Errorf("%q é reservado", None)
		}
		if seen[d.Name] {
			return fmt.Errorf("banco %q duplicado", d.Name)
		}
		seen[d.Name] = true

		required := map[string]string{
			"image": d.Image, "jdbc_scheme": d.JDBCScheme, "driver.group": d.Driver.Group,
			"driver.artifact": d.Driver.Artifact, "flyway_module": d.FlywayModule,
			"testcontainers.artifact": d.Testcontainers.Artifact, "testcontainers.class": d.Testcontainers.Class,
			"env.database": d.Env.Database, "env.user": d.Env.User, "env.password": d.Env.Password,
			"data_dir": d.DataDir, "healthcheck": d.Healthcheck,
		}
		for field, v := range required {
			if strings.TrimSpace(v) == "" {
				return fmt.Errorf("banco %q: campo %s é obrigatório", d.Name, field)
			}
		}
		if d.Port <= 0 {
			return fmt.Errorf("banco %q: port é obrigatório", d.Name)
		}
		files, err := c.migrationFiles(d.Name)
		if err != nil || len(files) == 0 {
			return fmt.Errorf("banco %q: falta migrations/%s/V1__*.sql", d.Name, d.Name)
		}
	}
	return nil
}

// Names devolve os nomes dos bancos na ordem do catálogo.
func (c *Catalog) Names() []string {
	out := make([]string, 0, len(c.Databases))
	for _, d := range c.Databases {
		out = append(out, d.Name)
	}
	return out
}

// Get devolve o banco pelo nome.
func (c *Catalog) Get(name string) (Database, bool) {
	for _, d := range c.Databases {
		if d.Name == name {
			return d, true
		}
	}
	return Database{}, false
}

func (c *Catalog) migrationFiles(name string) ([]string, error) {
	entries, err := fs.ReadDir(c.fsys, path.Join("migrations", name))
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			out = append(out, e.Name())
		}
	}
	return out, nil
}

// Migrations devolve as migrations do banco: nome do arquivo -> conteúdo.
func (c *Catalog) Migrations(name string) (map[string][]byte, error) {
	files, err := c.migrationFiles(name)
	if err != nil {
		return nil, err
	}
	out := map[string][]byte{}
	for _, f := range files {
		b, err := fs.ReadFile(c.fsys, path.Join("migrations", name, f))
		if err != nil {
			return nil, err
		}
		out[f] = b
	}
	return out, nil
}

// Slots gera os trechos que substituem os blocos "forge:slot" do template.
func (d Database) Slots() (map[string]string, error) {
	out := map[string]string{}
	for name, tpl := range slotTemplates {
		t, err := template.New(name).Delims("[[", "]]").Parse(tpl)
		if err != nil {
			return nil, fmt.Errorf("slot %s: %w", name, err)
		}
		var buf bytes.Buffer
		if err := t.Execute(&buf, d); err != nil {
			return nil, fmt.Errorf("slot %s: %w", name, err)
		}
		out[name] = buf.String()
	}
	return out, nil
}
