// Package project lê e grava o .forge.yaml, que registra como um projeto foi gerado.
package project

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// File é o nome do arquivo na raiz do projeto gerado.
const File = ".forge.yaml"

// Config é o conteúdo do .forge.yaml. Serve para regenerar o projeto (forge upgrade --diff).
type Config struct {
	Forge    ForgeInfo         `yaml:"forge"`
	Template string            `yaml:"template"`
	Project  ProjectInfo       `yaml:"project"`
	Options  map[string]string `yaml:"options"`
}

type ForgeInfo struct {
	Version string `yaml:"version"`
}

type ProjectInfo struct {
	Name    string `yaml:"name"`
	Package string `yaml:"package,omitempty"`
}

const header = "# Gerado pelo forge. Registra as respostas usadas para criar o projeto.\n"

// Write grava o .forge.yaml em dir.
func Write(dir string, c Config) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, File), append([]byte(header), data...), 0o644)
}

// Read lê o .forge.yaml de dir.
func Read(dir string) (Config, error) {
	data, err := os.ReadFile(filepath.Join(dir, File))
	if err != nil {
		return Config{}, err
	}
	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return Config{}, fmt.Errorf("%s inválido: %w", File, err)
	}
	return c, nil
}
