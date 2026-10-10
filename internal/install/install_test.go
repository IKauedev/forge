package install_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/ikauedev/forge/internal/install"
)

type fakeSys struct {
	goos, home string
	env        map[string]string
	userPath   string
	writes     int
}

func (f *fakeSys) GOOS() string                  { return f.goos }
func (f *fakeSys) Home() (string, error)         { return f.home, nil }
func (f *fakeSys) Getenv(k string) string        { return f.env[k] }
func (f *fakeSys) ReadUserPath() (string, error) { return f.userPath, nil }
func (f *fakeSys) WriteUserPath(v string) error {
	f.userPath = v
	f.writes++
	return nil
}

func fakeExe(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "forge-src")
	if err := os.WriteFile(p, []byte(content), 0o755); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestDefaultDir(t *testing.T) {
	win := &fakeSys{goos: "windows", home: `C:\Users\ana`, env: map[string]string{"LOCALAPPDATA": `C:\Users\ana\AppData\Local`}}
	if got, _ := install.DefaultDir(win); got != filepath.Join(`C:\Users\ana\AppData\Local`, "Programs", "forge") {
		t.Errorf("windows: %s", got)
	}
	nix := &fakeSys{goos: "linux", home: "/home/ana", env: map[string]string{}}
	if got, _ := install.DefaultDir(nix); got != filepath.Join("/home/ana", ".local", "bin") {
		t.Errorf("linux: %s", got)
	}
}

func TestInstallOnWindowsCopiesAndAddsToUserPath(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "forge")
	sys := &fakeSys{goos: "windows", home: t.TempDir(), env: map[string]string{"PATH": `C:\Windows`}, userPath: `%USERPROFILE%\bin;C:\tools`}

	res, err := install.Install(sys, fakeExe(t, "v1"), dir)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Copied || !res.PathChanged {
		t.Errorf("deveria copiar e alterar o PATH: %+v", res)
	}
	b, err := os.ReadFile(filepath.Join(dir, "forge.exe"))
	if err != nil || string(b) != "v1" {
		t.Fatalf("binário não copiado: %v %q", err, b)
	}
	// preserva as entradas existentes (inclusive %VAR%) e acrescenta o diretório
	if sys.userPath != `%USERPROFILE%\bin;C:\tools;`+dir {
		t.Errorf("PATH do usuário = %q", sys.userPath)
	}

	// segunda execução: atualiza o binário sem duplicar a entrada
	res, err = install.Install(sys, fakeExe(t, "v2"), dir)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Copied || res.PathChanged || sys.writes != 1 {
		t.Errorf("não deveria alterar o PATH de novo: %+v (escritas=%d)", res, sys.writes)
	}
	if b, _ := os.ReadFile(filepath.Join(dir, "forge.exe")); string(b) != "v2" {
		t.Errorf("binário não atualizado: %q", b)
	}
	if _, err := os.Stat(filepath.Join(dir, "forge.exe.old")); err == nil {
		t.Error("sobrou o forge.exe.old")
	}
}

func TestInstallWindowsIgnoresCaseAndTrailingSlash(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "Forge")
	sys := &fakeSys{goos: "windows", home: t.TempDir(), env: map[string]string{}, userPath: strings.ToUpper(dir) + `\`}
	res, err := install.Install(sys, fakeExe(t, "x"), dir)
	if err != nil {
		t.Fatal(err)
	}
	if res.PathChanged || sys.writes != 0 {
		t.Errorf("a entrada já existia: %+v", res)
	}
}

func TestInstallWhenAlreadyRunningFromTheInstalledCopy(t *testing.T) {
	dir := t.TempDir()
	exe := filepath.Join(dir, "forge.exe")
	if err := os.WriteFile(exe, []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
	sys := &fakeSys{goos: "windows", home: t.TempDir(), env: map[string]string{}}
	res, err := install.Install(sys, exe, dir)
	if err != nil {
		t.Fatal(err)
	}
	if res.Copied {
		t.Error("não deve copiar o binário sobre ele mesmo")
	}
	if b, _ := os.ReadFile(exe); string(b) != "x" {
		t.Error("o binário foi alterado")
	}
}

func TestInstallOnUnixAddsAProfileLineOnce(t *testing.T) {
	home := t.TempDir()
	sys := &fakeSys{goos: "linux", home: home, env: map[string]string{"SHELL": "/bin/zsh", "PATH": "/usr/bin"}}
	dir := filepath.Join(home, ".local", "bin")

	res, err := install.Install(sys, fakeExe(t, "v1"), "")
	if err != nil {
		t.Fatal(err)
	}
	if res.Dir != dir || !res.Copied || !res.PathChanged || filepath.Base(res.ProfileFile) != ".zshrc" {
		t.Fatalf("resultado inesperado: %+v", res)
	}
	if info, err := os.Stat(filepath.Join(dir, "forge")); err != nil || info.Mode()&0o100 == 0 && os.PathSeparator == '/' {
		t.Errorf("binário ausente ou sem execução: %v", err)
	}
	profile, _ := os.ReadFile(res.ProfileFile)
	if !strings.Contains(string(profile), `export PATH="$HOME/.local/bin:$PATH"`) {
		t.Errorf("perfil sem o export:\n%s", profile)
	}

	res, err = install.Install(sys, fakeExe(t, "v2"), "")
	if err != nil {
		t.Fatal(err)
	}
	if res.PathChanged {
		t.Error("não deveria alterar o perfil de novo")
	}
	profile, _ = os.ReadFile(res.ProfileFile)
	if strings.Count(string(profile), "export PATH") != 1 {
		t.Errorf("linha duplicada:\n%s", profile)
	}
}

func TestInstallOnUnixLeavesTheProfileAloneWhenAlreadyOnPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("simula Linux com um caminho do Windows (o C: quebra a lista separada por :); roda no CI Linux")
	}
	home := t.TempDir()
	dir := filepath.Join(home, ".local", "bin")
	sys := &fakeSys{goos: "linux", home: home, env: map[string]string{"SHELL": "/bin/bash", "PATH": "/usr/bin:" + dir}}

	res, err := install.Install(sys, fakeExe(t, "x"), "")
	if err != nil {
		t.Fatal(err)
	}
	if res.PathChanged || !res.OnPath {
		t.Errorf("já estava no PATH: %+v", res)
	}
	if _, err := os.Stat(filepath.Join(home, ".bashrc")); err == nil {
		t.Error("não deveria criar o .bashrc")
	}
}

func TestProfileChoiceByShell(t *testing.T) {
	cases := map[string]string{"/bin/bash": ".bashrc", "/usr/bin/zsh": ".zshrc", "/usr/bin/fish": "config.fish", "": ".profile"}
	for shell, want := range cases {
		home := t.TempDir()
		sys := &fakeSys{goos: "linux", home: home, env: map[string]string{"SHELL": shell}}
		res, err := install.Install(sys, fakeExe(t, "x"), "")
		if err != nil {
			t.Fatal(err)
		}
		if filepath.Base(res.ProfileFile) != want {
			t.Errorf("SHELL=%q: perfil %s, esperado %s", shell, res.ProfileFile, want)
		}
	}
}

func TestUninstall(t *testing.T) {
	// Windows
	dir := filepath.Join(t.TempDir(), "forge")
	win := &fakeSys{goos: "windows", home: t.TempDir(), env: map[string]string{}, userPath: `C:\tools`}
	if _, err := install.Install(win, fakeExe(t, "x"), dir); err != nil {
		t.Fatal(err)
	}
	res, err := install.Uninstall(win, dir)
	if err != nil {
		t.Fatal(err)
	}
	if !res.PathChanged || win.userPath != `C:\tools` {
		t.Errorf("PATH após remover = %q (%+v)", win.userPath, res)
	}
	if _, err := os.Stat(filepath.Join(dir, "forge.exe")); err == nil {
		t.Error("binário não removido")
	}

	// Unix
	home := t.TempDir()
	nix := &fakeSys{goos: "linux", home: home, env: map[string]string{"SHELL": "/bin/bash"}}
	rc := filepath.Join(home, ".bashrc")
	if err := os.WriteFile(rc, []byte("alias ll='ls -l'\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := install.Install(nix, fakeExe(t, "x"), ""); err != nil {
		t.Fatal(err)
	}
	if _, err := install.Uninstall(nix, ""); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(rc)
	if strings.Contains(string(b), "forge") || !strings.Contains(string(b), "alias ll") {
		t.Errorf("perfil após remover:\n%s", b)
	}
}
