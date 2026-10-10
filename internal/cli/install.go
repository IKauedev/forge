package cli

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/ikauedev/forge/internal/install"
)

// installSystem é trocado nos testes.
var installSystem install.System = install.OSSystem{}

func newInstallCmd() *cobra.Command {
	var dir string
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Instala o forge neste computador e o põe no PATH",
		Long: `Copia este executável para uma pasta do seu usuário e a adiciona ao PATH, para chamar apenas
` + "`forge`" + ` de qualquer pasta. Não precisa de administrador.

  Windows        %LOCALAPPDATA%\Programs\forge  (PATH do usuário)
  Linux e macOS  ~/.local/bin                    (linha no perfil do shell, se necessário)

Rodar de novo atualiza a cópia instalada.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			exe, err := executablePath()
			if err != nil {
				return err
			}
			res, err := install.Install(installSystem, exe, dir)
			if err != nil {
				return err
			}
			printInstall(cmd.OutOrStdout(), res)
			return nil
		},
	}
	cmd.Flags().StringVar(&dir, "dir", "", "pasta de instalação (padrão: a do seu sistema)")
	return cmd
}

func newUninstallCmd() *cobra.Command {
	var dir string
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Remove o forge instalado por `forge install` e tira a pasta do PATH",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := install.Uninstall(installSystem, dir)
			if err != nil {
				return err
			}
			w := cmd.OutOrStdout()
			_, _ = fmt.Fprintf(w, "Removido: %s\n", res.Binary)
			if res.PathChanged {
				_, _ = fmt.Fprintln(w, "PATH atualizado; abra um novo terminal.")
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&dir, "dir", "", "pasta da instalação (padrão: a do seu sistema)")
	return cmd
}

func printInstall(w io.Writer, res install.Result) {
	switch {
	case res.Copied:
		_, _ = fmt.Fprintf(w, "forge instalado em %s\n", res.Binary)
	default:
		_, _ = fmt.Fprintf(w, "forge já estava em %s\n", res.Binary)
	}
	switch {
	case res.PathChanged && res.ProfileFile != "":
		_, _ = fmt.Fprintf(w, "PATH configurado em %s.\nAbra um novo terminal (ou rode `source %s`) e use: forge\n", res.ProfileFile, res.ProfileFile)
	case res.PathChanged:
		_, _ = fmt.Fprintln(w, "PATH do usuário atualizado.\nAbra um NOVO terminal e use: forge")
	case res.OnPath:
		_, _ = fmt.Fprintln(w, "A pasta já está no PATH: use `forge` em qualquer lugar.")
	default:
		_, _ = fmt.Fprintln(w, "A pasta já estava registrada no PATH; abra um novo terminal se `forge` ainda não for encontrado.")
	}
}
