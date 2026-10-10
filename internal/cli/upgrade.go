package cli

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ikauedev/forge/internal/project"
	"github.com/ikauedev/forge/internal/upgrade"
)

func newUpgradeCmd() *cobra.Command {
	var diff bool
	cmd := &cobra.Command{
		Use:   "upgrade [diretório]",
		Short: "Mostra o que mudou no template desde que o projeto foi gerado",
		Long: `Regenera o projeto num diretório temporário com as respostas do .forge.yaml e mostra o diff
contra o projeto atual. Nada é sobrescrito.

Arquivos que só existem no projeto (criados por você ou removidos do template) são apenas listados.`,
		Example: `  forge upgrade --diff
  forge upgrade ../payments-api --diff`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !diff {
				return errors.New("por enquanto só há o modo --diff (nada é sobrescrito)")
			}
			dir := "."
			if len(args) == 1 {
				dir = args[0]
			}
			return runUpgradeDiff(cmd, dir)
		},
	}
	cmd.Flags().BoolVar(&diff, "diff", false, "mostra as diferenças sem alterar o projeto")
	return cmd
}

func runUpgradeDiff(cmd *cobra.Command, dir string) error {
	cfg, err := project.Read(dir)
	if err != nil {
		return fmt.Errorf("não encontrei um projeto do forge em %q (%w)", dir, err)
	}
	src, _, err := loadTemplate(cfg.Template)
	if err != nil {
		return err
	}
	res, err := upgrade.Diff(src, dir, cfg)
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	if res.Empty() {
		_, _ = fmt.Fprintf(w, "O projeto já está igual ao que o forge %s gera para %s.\n", Version, cfg.Template)
		return nil
	}
	for _, c := range res.Changed {
		_, _ = fmt.Fprintln(w, c.Unified)
	}
	for _, p := range res.Added {
		_, _ = fmt.Fprintf(w, "novo no template: %s\n", p)
	}
	for _, p := range res.OnlyInProject {
		_, _ = fmt.Fprintf(w, "só no projeto:    %s\n", p)
	}
	_, _ = fmt.Fprintf(w, "\n%d alterado(s), %d novo(s) no template, %d só no projeto.\n",
		len(res.Changed), len(res.Added), len(res.OnlyInProject))
	return nil
}
