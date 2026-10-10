// Package cli define os comandos Cobra do forge.
package cli

import "github.com/spf13/cobra"

// NewRootCmd monta o comando raiz. Sem subcomando, em um terminal, abre o menu interativo.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "forge",
		Short: "Gera projetos Spring Boot, Quarkus, Angular e React pré-configurados",
		Long: `forge gera projetos prontos para produção (Spring Boot, Quarkus, Angular e React) com
testes, pipeline GitHub Actions (OIDC) e Terraform.

Sem argumentos, em um terminal, abre um menu interativo. Os comandos abaixo também funcionam direto,
inclusive em scripts e no CI.`,
		Version:       Version,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !isTerminal() {
				return cmd.Help()
			}
			return runMenu(cmd, newMenuUI())
		},
	}
	root.AddCommand(newNewCmd(), newListCmd(), newDoctorCmd(), newUpgradeCmd(), newUpdateCmd(), newInstallCmd(), newUninstallCmd())
	localize(root)
	return root
}

// newMenuUI é trocado nos testes.
var newMenuUI = func() menuUI { return huhUI{} }
