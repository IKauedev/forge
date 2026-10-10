package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Lista os templates e suas opções",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, m, err := loadSpringTemplate()
			if err != nil {
				return err
			}
			w := cmd.OutOrStdout()
			fmt.Fprintln(w, "Templates:")
			fmt.Fprintf(w, "  %s  %s\n\n", m.Name, m.Description)
			fmt.Fprintln(w, "Opções do forge new:")
			fmt.Fprintf(w, "  --package  pacote base Java (padrão %s)\n", defaultPackage)
			fmt.Fprintln(w, "  --out      diretório pai do projeto (padrão .)")
			for _, o := range m.Options {
				fmt.Fprintf(w, "  --%-8s %s: %s (padrão %s)\n", o.Name, o.Description, strings.Join(o.Values, " | "), o.Default)
			}
			fmt.Fprintln(w, "  --yes      não perguntar; usar os padrões")
			fmt.Fprintln(w, "  --no-git   não executar git init")
			return nil
		},
	}
}
