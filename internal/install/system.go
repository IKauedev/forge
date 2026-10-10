package install

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// OSSystem usa o sistema real. No Windows o PATH do usuário é lido e gravado pelo PowerShell, sem
// dependência nova e preservando entradas como %USERPROFILE%.
type OSSystem struct{}

func (OSSystem) GOOS() string             { return runtime.GOOS }
func (OSSystem) Home() (string, error)    { return os.UserHomeDir() }
func (OSSystem) Getenv(key string) string { return os.Getenv(key) }

const readPathScript = `$k = Get-Item -Path 'HKCU:\Environment'; ` +
	`[Console]::Out.Write($k.GetValue('Path', '', 'DoNotExpandEnvironmentNames'))`

// O valor novo viaja por variável de ambiente (sem problema de aspas). ExpandString mantém os %VAR%,
// e o par set/remove de uma variável descartável faz o Windows avisar os programas abertos.
const writePathScript = `Set-ItemProperty -Path 'HKCU:\Environment' -Name 'Path' -Value $env:FORGE_NEW_PATH -Type ExpandString; ` +
	`[Environment]::SetEnvironmentVariable('FORGE_BROADCAST', '1', 'User'); ` +
	`[Environment]::SetEnvironmentVariable('FORGE_BROADCAST', $null, 'User')`

func powershell(script string, env ...string) (string, error) {
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	cmd.Env = append(os.Environ(), env...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("powershell: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}

func (OSSystem) ReadUserPath() (string, error) { return powershell(readPathScript) }

func (OSSystem) WriteUserPath(value string) error {
	_, err := powershell(writePathScript, "FORGE_NEW_PATH="+value)
	return err
}
