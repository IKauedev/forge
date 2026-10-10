package cli

import (
	"github.com/charmbracelet/huh"

	"github.com/ikauedev/forge/internal/names"
	"github.com/ikauedev/forge/internal/template"
)

// huhPrompter pergunta o que faltar com um formulário no terminal.
type huhPrompter struct{}

func (huhPrompter) Fill(m *template.Manifest, a *Answers) error {
	var fields []huh.Field

	if a.Name == "" {
		fields = append(fields, huh.NewInput().
			Title("Nome do projeto").
			Description("Ex.: Payments API, payments-api, PaymentsAPI").
			Value(&a.Name).
			Validate(func(s string) error {
				_, err := names.New(s)
				return err
			}))
	}
	if m.JavaPackage && !a.PackageSet {
		if a.Package == "" {
			a.Package = defaultPackage
		}
		fields = append(fields, huh.NewInput().
			Title("Pacote base Java").
			Value(&a.Package).
			Validate(names.ValidatePackage))
	}

	chosen := map[string]*string{}
	for _, o := range m.Options {
		if _, ok := a.Options[o.Name]; ok {
			continue
		}
		v := o.Default
		chosen[o.Name] = &v
		fields = append(fields, huh.NewSelect[string]().
			Title(o.Name).
			Description(o.Description).
			Options(huh.NewOptions(o.Values...)...).
			Value(&v))
	}
	if len(fields) == 0 {
		return nil
	}
	if err := huh.NewForm(huh.NewGroup(fields...)).Run(); err != nil {
		return err
	}
	for name, v := range chosen {
		a.Options[name] = *v
	}
	return nil
}
