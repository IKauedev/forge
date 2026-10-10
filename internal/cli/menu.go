package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/ikauedev/forge/internal/install"
	"github.com/ikauedev/forge/internal/project"
	"github.com/ikauedev/forge/internal/selfupdate"
)

const banner = `  _____ ___  ____   ____ _____
 |  ___/ _ \|  _ \ / ___| ____|
 | |_ | | | | |_) | |  _|  _|
 |  _|| |_| |  _ <| |_| | |___
 |_|   \___/|_| \_\\____|_____|`

// Ações do menu.
const (
	actNew       = "new"
	actDiff      = "diff"
	actList      = "list"
	actDoctor    = "doctor"
	actUpdate    = "update"
	actCheck     = "check"
	actInstall   = "install"
	actQuit      = "quit"
	updateTTL    = 24 * time.Hour
	updateBudget = 2 * time.Second
)

type menuOption struct {
	Key, Label string
}

// menuUI é a interação com o usuário; os testes trocam por uma versão roteirizada.
type menuUI interface {
	Choose(title string, opts []menuOption) (string, error)
	Confirm(title string, def bool) (bool, error)
}

type huhUI struct{}

func (huhUI) Choose(title string, opts []menuOption) (string, error) {
	var choice string
	options := make([]huh.Option[string], len(opts))
	for i, o := range opts {
		options[i] = huh.NewOption(o.Label, o.Key)
	}
	err := huh.NewSelect[string]().Title(title).Options(options...).Value(&choice).Run()
	return choice, err
}

func (huhUI) Confirm(title string, def bool) (bool, error) {
	ok := def
	err := huh.NewConfirm().Title(title).Affirmative("Sim").Negative("Não").Value(&ok).Run()
	return ok, err
}

// confirmPrompt é trocado nos testes.
var confirmPrompt = func(title string, def bool) (bool, error) { return huhUI{}.Confirm(title, def) }

// state é o que o menu descobre sobre o contexto antes de mostrar as opções.
type state struct {
	version     string
	project     *project.Config // projeto do forge na pasta atual, se houver
	newer       string          // versão mais nova disponível, se houver
	onPath      bool            // existe um `forge` no PATH
	installable bool            // dá para instalar (ainda não está instalado no PATH)
}

func gatherState(ctx context.Context, cwd string) state {
	st := state{version: Version}
	if cfg, err := project.Read(cwd); err == nil {
		st.project = &cfg
	}
	_, lookErr := exec.LookPath("forge")
	st.onPath = lookErr == nil
	if !st.onPath {
		if _, ok := install.InstalledPath(installSystem); !ok {
			st.installable = true
		} else {
			st.installable = !install.InPath(installSystem, installDir())
		}
	}
	if os.Getenv("FORGE_NO_UPDATE_CHECK") == "" {
		ctx, cancel := context.WithTimeout(ctx, updateBudget)
		defer cancel()
		if hint, ok := selfupdate.CheckCached(ctx, newUpdateClient(), Version, updateCachePath(), time.Now(), updateTTL); ok && hint.Newer {
			st.newer = hint.Latest
		}
	}
	return st
}

func installDir() string {
	dir, _ := install.DefaultDir(installSystem)
	return dir
}

func updateCachePath() string {
	base, err := os.UserCacheDir()
	if err != nil {
		return filepath.Join(os.TempDir(), "forge-update.json")
	}
	return filepath.Join(base, "forge", "update.json")
}

func (st state) options() []menuOption {
	var opts []menuOption
	if st.newer != "" {
		opts = append(opts, menuOption{actUpdate, "⬆  Atualizar o forge para " + st.newer})
	}
	opts = append(opts, menuOption{actNew, "✨ Criar um novo projeto"})
	if st.project != nil {
		opts = append(opts, menuOption{actDiff, "🔍 Ver o que mudou no template desde que este projeto foi gerado"})
	}
	opts = append(opts,
		menuOption{actList, "📋 Ver templates e opções"},
		menuOption{actDoctor, "🩺 Verificar o ambiente (Java, Docker, Node...)"},
	)
	if st.newer == "" {
		opts = append(opts, menuOption{actCheck, "🔄 Procurar atualizações"})
	}
	if st.installable {
		opts = append(opts, menuOption{actInstall, "📦 Instalar o forge neste computador (PATH)"})
	}
	return append(opts, menuOption{actQuit, "Sair"})
}

func printBanner(w io.Writer, st state, colored bool) {
	paint := func(code, s string) string {
		if !colored {
			return s
		}
		return "\x1b[" + code + "m" + s + "\x1b[0m"
	}
	_, _ = fmt.Fprintln(w, paint("36", banner))
	_, _ = fmt.Fprintf(w, "\n %s  %s\n", paint("1", "forge "+displayVersion(st.version)), paint("90", "projetos Spring Boot, Quarkus, Angular e React prontos para produção"))
	if st.project != nil {
		_, _ = fmt.Fprintf(w, " %s\n", paint("32", fmt.Sprintf("Projeto detectado aqui: %s (%s)", st.project.Project.Name, st.project.Template)))
	}
	if st.newer != "" {
		_, _ = fmt.Fprintf(w, " %s\n", paint("33", fmt.Sprintf("Nova versão disponível: %s", st.newer)))
	}
	if !st.onPath {
		_, _ = fmt.Fprintf(w, " %s\n", paint("33", "O forge ainda não está no PATH — escolha \"Instalar\" para usar só `forge` em qualquer pasta."))
	}
	_, _ = fmt.Fprintln(w)
}

// runMenu mostra o menu inicial em loop até o usuário sair.
func runMenu(cmd *cobra.Command, ui menuUI) error {
	w := cmd.OutOrStdout()
	cwd, _ := os.Getwd()
	colored := useColor(cmd)

	for {
		st := gatherState(cmd.Context(), cwd)
		printBanner(w, st, colored)

		choice, err := ui.Choose("O que você quer fazer?", st.options())
		if err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				return nil
			}
			return err
		}
		var args []string
		switch choice {
		case actQuit:
			return nil
		case actNew:
			args = []string{"new"}
		case actDiff:
			args = []string{"upgrade", "--diff"}
		case actList:
			args = []string{"list"}
		case actDoctor:
			args = []string{"doctor"}
		case actUpdate, actCheck:
			args = []string{"update"}
		case actInstall:
			args = []string{"install"}
		}
		if err := runSub(cmd, args...); err != nil {
			var exit *ExitError
			if !errors.As(err, &exit) { // doctor com falha já imprimiu o relatório
				_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "erro:", err)
			}
		}
		_, _ = fmt.Fprintln(w)
		back, err := ui.Confirm("Voltar ao menu?", true)
		if err != nil || !back {
			return nil
		}
	}
}

// runSub executa outro comando do forge com a mesma entrada e saída do menu.
func runSub(cmd *cobra.Command, args ...string) error {
	root := NewRootCmd()
	root.SetOut(cmd.OutOrStdout())
	root.SetErr(cmd.ErrOrStderr())
	root.SetIn(cmd.InOrStdin())
	root.SetArgs(args)
	return root.ExecuteContext(cmd.Context())
}
