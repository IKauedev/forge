package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ikauedeveloper/forge/internal/doctor"
)

// ExitError faz o main sair com um código específico sem imprimir mensagem extra.
type ExitError struct {
	Code int
	Msg  string
}

func (e *ExitError) Error() string { return e.Msg }

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Confere se o ambiente tem o necessário (Java, Docker, Git...)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDoctor(cmd, doctor.OSEnv{}, useColor(cmd))
		},
	}
}

func runDoctor(cmd *cobra.Command, env doctor.Env, colored bool) error {
	w := cmd.OutOrStdout()
	results := doctor.Run(env)
	doctor.Print(w, results, colored)
	if n := doctor.Failed(results); n > 0 {
		fmt.Fprintf(w, "\n%d verificação(ões) obrigatória(s) falharam.\n", n)
		return &ExitError{Code: 1}
	}
	fmt.Fprintln(w, "\nAmbiente pronto.")
	return nil
}

// useColor liga as cores só em terminal e respeitando NO_COLOR.
func useColor(cmd *cobra.Command) bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	f, ok := cmd.OutOrStdout().(*os.File)
	if !ok {
		return false
	}
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}
