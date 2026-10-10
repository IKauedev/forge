package engine

import (
	"fmt"
	"io/fs"

	"github.com/ikauedeveloper/forge/internal/catalog"
	"github.com/ikauedeveloper/forge/internal/template"
)

// migrationDir é onde o template guarda as migrations Flyway.
const migrationDir = "src/main/resources/db/migration/"

// LoadManifest lê o manifesto do template e completa as opções que vêm do catálogo de bancos.
func LoadManifest(src fs.FS) (*template.Manifest, error) {
	m, err := template.Load(src)
	if err != nil {
		return nil, err
	}
	for _, o := range m.Options {
		if !o.ExtendWithCatalog {
			continue
		}
		c, err := catalog.Load()
		if err != nil {
			return nil, err
		}
		if err := m.ExtendOption(o.Name, c.Names()); err != nil {
			return nil, err
		}
	}
	return m, nil
}

// databaseAssets devolve os trechos (slots) e as migrations do banco escolhido. Sem banco
// (db=none ou template sem a opção), não há nada a trocar.
func databaseAssets(sel template.Selection) (map[string]string, map[string][]byte, error) {
	db := sel["db"]
	if db == "" || db == catalog.None {
		return nil, nil, nil
	}
	c, err := catalog.Load()
	if err != nil {
		return nil, nil, err
	}
	d, ok := c.Get(db)
	if !ok {
		return nil, nil, fmt.Errorf("banco %q não existe no catálogo", db)
	}
	slots, err := d.Slots()
	if err != nil {
		return nil, nil, err
	}
	migrations, err := c.Migrations(db)
	if err != nil {
		return nil, nil, err
	}
	overrides := map[string][]byte{}
	for name, content := range migrations {
		overrides[migrationDir+name] = content
	}
	return slots, overrides, nil
}
