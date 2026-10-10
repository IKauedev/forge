// Package cli define os comandos Cobra do forge.
package cli

import "github.com/spf13/cobra"

// NewRootCmd monta o comando raiz.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "forge",
		Short:         "Gera projetos Spring Boot pré-configurados",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(newNewCmd())
	return root
}
