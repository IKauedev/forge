package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ikauedev/forge/templates"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Lista os templates e suas opções",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			w := cmd.OutOrStdout()
			_, _ = fmt.Fprintln(w, "Templates (forge new --type <template>):")
			for _, name := range templates.Names {
				_, m, err := loadTemplate(name)
				if err != nil {
					return err
				}
				_, _ = fmt.Fprintf(w, "\n  %s  %s\n", m.Name, m.Description)
				if m.JavaPackage {
					_, _ = fmt.Fprintf(w, "    --package  pacote base Java (padrão %s)\n", defaultPackage)
				}
				for _, o := range m.Options {
					_, _ = fmt.Fprintf(w, "    --%-8s %s: %s (padrão %s)\n", o.Name, o.Description, strings.Join(o.Values, " | "), o.Default)
				}
			}
			_, _ = fmt.Fprintln(w, "\nOpções comuns:")
			_, _ = fmt.Fprintln(w, "  --out      diretório pai do projeto (padrão .)")
			_, _ = fmt.Fprintln(w, "  --yes      não perguntar; usar os padrões")
			_, _ = fmt.Fprintln(w, "  --no-git   não executar git init")
			return nil
		},
	}
}
