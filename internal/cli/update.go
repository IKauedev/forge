package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/ikauedev/forge/internal/selfupdate"
)

// Repo é o repositório das releases do forge.
const Repo = "ikauedev/forge"

// Variáveis trocadas nos testes.
var (
	newUpdateClient = func() selfupdate.Client {
		return selfupdate.Client{Repo: Repo, Token: githubToken()}
	}
	executablePath = currentExecutable
)

func currentExecutable() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	return exe, nil
}

// githubToken procura um token para repositórios privados: GITHUB_TOKEN, GH_TOKEN ou `gh auth token`.
func githubToken() string {
	for _, k := range []string{"GITHUB_TOKEN", "GH_TOKEN"} {
		if v := strings.TrimSpace(os.Getenv(k)); v != "" {
			return v
		}
	}
	if _, err := exec.LookPath("gh"); err != nil {
		return ""
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "gh", "auth", "token").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func newUpdateCmd() *cobra.Command {
	var check, yes bool
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Atualiza o forge para a última release",
		Long: `Consulta a última release no GitHub e, havendo versão nova, baixa o binário do seu sistema,
confere o SHA-256 (checksums.txt) e troca o executável atual.

Em repositório privado é preciso um token: GITHUB_TOKEN, GH_TOKEN ou uma sessão do ` + "`gh auth login`" + `.`,
		Example: `  forge update --check   # só diz se há versão nova
  forge update           # atualiza (pede confirmação)
  forge update --yes`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdate(cmd, check, yes)
		},
	}
	cmd.Flags().BoolVar(&check, "check", false, "só verifica se há versão nova; não altera nada")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "não pedir confirmação")
	return cmd
}

func runUpdate(cmd *cobra.Command, check, yes bool) error {
	w := cmd.OutOrStdout()
	exe, err := executablePath()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Minute)
	defer cancel()
	client := newUpdateClient()

	// Primeiro só consulta: a confirmação precisa saber de qual versão para qual.
	probe, err := client.Update(ctx, Version, exe, runtime.GOOS, runtime.GOARCH, true)
	if err != nil {
		return updateError(err)
	}
	if !probe.Newer {
		_, _ = fmt.Fprintf(w, "O forge %s já é a versão mais recente.\n", displayVersion(Version))
		return nil
	}
	_, _ = fmt.Fprintf(w, "Nova versão disponível: %s (instalada: %s).\n", probe.To, displayVersion(Version))
	if check {
		_, _ = fmt.Fprintln(w, "Rode `forge update` para atualizar.")
		return nil
	}
	if !yes && isTerminal() {
		ok, err := confirmPrompt(fmt.Sprintf("Atualizar de %s para %s?", displayVersion(Version), probe.To), true)
		if err != nil {
			return err
		}
		if !ok {
			_, _ = fmt.Fprintln(w, "Atualização cancelada.")
			return nil
		}
	}
	res, err := client.Update(ctx, Version, exe, runtime.GOOS, runtime.GOARCH, false)
	if err != nil {
		return updateError(err)
	}
	_, _ = fmt.Fprintf(w, "forge atualizado para %s em %s.\n", res.To, exe)
	return nil
}

func updateError(err error) error {
	if errors.Is(err, selfupdate.ErrNotFound) {
		return fmt.Errorf("%w\nEnquanto isso: go install github.com/%s/cmd/forge@latest", err, Repo)
	}
	return err
}

func displayVersion(v string) string {
	if v == selfupdate.DevVersion || v == "" {
		return "dev (compilado localmente)"
	}
	return v
}
