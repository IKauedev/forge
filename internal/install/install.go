// Package install coloca o binário do forge num diretório do usuário e o põe no PATH, sem exigir admin.
package install

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// System isola o que muda por sistema operacional, para os testes.
type System interface {
	GOOS() string
	Home() (string, error)
	Getenv(key string) string
	// ReadUserPath e WriteUserPath existem só no Windows: o PATH do usuário fica no registro.
	ReadUserPath() (string, error)
	WriteUserPath(value string) error
}

// Result descreve o que a instalação (ou remoção) fez.
type Result struct {
	Dir         string // diretório do binário
	Binary      string // caminho completo
	Copied      bool   // o binário foi copiado/atualizado
	PathChanged bool   // o PATH foi alterado (abra um novo terminal)
	ProfileFile string // arquivo de shell alterado (Unix)
	OnPath      bool   // o diretório já estava no PATH do processo atual
}

const (
	profileMarker = "# forge (adicionado por 'forge install')"
	binaryName    = "forge"
)

// BinaryName devolve o nome do executável no sistema (forge ou forge.exe).
func BinaryName(goos string) string {
	if goos == "windows" {
		return binaryName + ".exe"
	}
	return binaryName
}

// DefaultDir é o diretório de instalação por usuário:
// Windows %LOCALAPPDATA%\Programs\forge; Linux e macOS ~/.local/bin.
func DefaultDir(sys System) (string, error) {
	if sys.GOOS() == "windows" {
		if base := sys.Getenv("LOCALAPPDATA"); base != "" {
			return filepath.Join(base, "Programs", binaryName), nil
		}
		home, err := sys.Home()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, "AppData", "Local", "Programs", binaryName), nil
	}
	home, err := sys.Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "bin"), nil
}

// Install copia exe para dir (o padrão se vazio) e garante que dir esteja no PATH do usuário.
func Install(sys System, exe, dir string) (Result, error) {
	if dir == "" {
		d, err := DefaultDir(sys)
		if err != nil {
			return Result{}, err
		}
		dir = d
	}
	res := Result{Dir: dir, Binary: filepath.Join(dir, BinaryName(sys.GOOS()))}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return res, err
	}

	same, err := sameFile(exe, res.Binary)
	if err != nil {
		return res, err
	}
	if !same {
		if err := copyBinary(exe, res.Binary); err != nil {
			return res, fmt.Errorf("copiando para %s: %w", res.Binary, err)
		}
		res.Copied = true
	}

	res.OnPath = inPathList(sys.Getenv("PATH"), dir, sys.GOOS())
	if sys.GOOS() == "windows" {
		changed, err := addToWindowsPath(sys, dir)
		if err != nil {
			return res, fmt.Errorf("atualizando o PATH do usuário: %w", err)
		}
		res.PathChanged = changed
		return res, nil
	}
	if res.OnPath {
		return res, nil
	}
	file, changed, err := addToProfile(sys, dir)
	if err != nil {
		return res, fmt.Errorf("atualizando o perfil do shell: %w", err)
	}
	res.ProfileFile, res.PathChanged = file, changed
	return res, nil
}

// Uninstall remove o binário de dir (o padrão se vazio) e a entrada do PATH.
func Uninstall(sys System, dir string) (Result, error) {
	if dir == "" {
		d, err := DefaultDir(sys)
		if err != nil {
			return Result{}, err
		}
		dir = d
	}
	res := Result{Dir: dir, Binary: filepath.Join(dir, BinaryName(sys.GOOS()))}
	if err := os.Remove(res.Binary); err != nil && !errors.Is(err, os.ErrNotExist) {
		// No Windows o executável em uso não pode ser apagado: renomeia e o sistema limpa depois.
		if rerr := os.Rename(res.Binary, res.Binary+".old"); rerr != nil {
			return res, err
		}
	}
	if sys.GOOS() == "windows" {
		changed, err := removeFromWindowsPath(sys, dir)
		res.PathChanged = changed
		return res, err
	}
	file, changed, err := removeFromProfile(sys)
	res.ProfileFile, res.PathChanged = file, changed
	return res, err
}

// InstalledPath devolve onde o forge está instalado por este pacote, se estiver.
func InstalledPath(sys System) (string, bool) {
	dir, err := DefaultDir(sys)
	if err != nil {
		return "", false
	}
	p := filepath.Join(dir, BinaryName(sys.GOOS()))
	if _, err := os.Stat(p); err != nil {
		return "", false
	}
	return p, true
}

// InPath informa se dir está no PATH do processo atual.
func InPath(sys System, dir string) bool {
	return inPathList(sys.Getenv("PATH"), dir, sys.GOOS())
}

func sameFile(a, b string) (bool, error) {
	ia, err := os.Stat(a)
	if err != nil {
		return false, err
	}
	ib, err := os.Stat(b)
	if err != nil {
		return false, nil // destino ainda não existe
	}
	return os.SameFile(ia, ib), nil
}

func copyBinary(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	tmp := dst + ".new"
	out, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	// Um executável em uso (Windows) não pode ser sobrescrito, mas pode ser renomeado.
	if _, err := os.Stat(dst); err == nil {
		old := dst + ".old"
		_ = os.Remove(old)
		if err := os.Rename(dst, old); err != nil {
			_ = os.Remove(tmp)
			return err
		}
	}
	if err := os.Rename(tmp, dst); err != nil {
		return err
	}
	_ = os.Remove(dst + ".old")
	return nil
}

func inPathList(pathVar, dir, goos string) bool {
	sep := ":"
	if goos == "windows" {
		sep = ";"
	}
	want := normalize(dir, goos)
	for _, p := range strings.Split(pathVar, sep) {
		if p != "" && normalize(p, goos) == want {
			return true
		}
	}
	return false
}

func normalize(p, goos string) string {
	p = strings.TrimRight(filepath.Clean(p), `\/`)
	if goos == "windows" {
		return strings.ToLower(p)
	}
	return p
}

func addToWindowsPath(sys System, dir string) (bool, error) {
	current, err := sys.ReadUserPath()
	if err != nil {
		return false, err
	}
	if inPathList(current, dir, "windows") {
		return false, nil
	}
	next := dir
	if strings.TrimSpace(current) != "" {
		next = strings.TrimRight(current, ";") + ";" + dir
	}
	return true, sys.WriteUserPath(next)
}

func removeFromWindowsPath(sys System, dir string) (bool, error) {
	current, err := sys.ReadUserPath()
	if err != nil {
		return false, err
	}
	want := normalize(dir, "windows")
	var keep []string
	removed := false
	for _, p := range strings.Split(current, ";") {
		if p == "" {
			continue
		}
		if normalize(p, "windows") == want {
			removed = true
			continue
		}
		keep = append(keep, p)
	}
	if !removed {
		return false, nil
	}
	return true, sys.WriteUserPath(strings.Join(keep, ";"))
}

// profileFor escolhe o arquivo de inicialização do shell do usuário.
func profileFor(sys System) (path, line string, err error) {
	home, err := sys.Home()
	if err != nil {
		return "", "", err
	}
	exportLine := `export PATH="$HOME/.local/bin:$PATH"`
	switch filepath.Base(sys.Getenv("SHELL")) {
	case "zsh":
		return filepath.Join(home, ".zshrc"), exportLine, nil
	case "fish":
		return filepath.Join(home, ".config", "fish", "config.fish"), `fish_add_path -g "$HOME/.local/bin"`, nil
	case "bash":
		if sys.GOOS() == "darwin" {
			return filepath.Join(home, ".bash_profile"), exportLine, nil
		}
		return filepath.Join(home, ".bashrc"), exportLine, nil
	default:
		return filepath.Join(home, ".profile"), exportLine, nil
	}
}

func addToProfile(sys System, dir string) (string, bool, error) {
	home, err := sys.Home()
	if err != nil {
		return "", false, err
	}
	if filepath.Clean(dir) != filepath.Join(home, ".local", "bin") {
		return "", false, fmt.Errorf("para o diretório %q, adicione-o ao PATH manualmente", dir)
	}
	file, line, err := profileFor(sys)
	if err != nil {
		return "", false, err
	}
	existing, err := os.ReadFile(file)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return file, false, err
	}
	if strings.Contains(string(existing), profileMarker) {
		return file, false, nil
	}
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		return file, false, err
	}
	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return file, false, err
	}
	defer func() { _ = f.Close() }()
	prefix := "\n"
	if len(existing) == 0 {
		prefix = ""
	}
	_, err = fmt.Fprintf(f, "%s%s\n%s\n", prefix, profileMarker, line)
	return file, err == nil, err
}

func removeFromProfile(sys System) (string, bool, error) {
	file, _, err := profileFor(sys)
	if err != nil {
		return "", false, err
	}
	b, err := os.ReadFile(file)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return file, false, nil
		}
		return file, false, err
	}
	lines := strings.Split(string(b), "\n")
	var out []string
	removed := false
	for i := 0; i < len(lines); i++ {
		if lines[i] == profileMarker {
			removed = true
			i++ // pula também a linha do export
			continue
		}
		out = append(out, lines[i])
	}
	if !removed {
		return file, false, nil
	}
	return file, true, os.WriteFile(file, []byte(strings.Join(out, "\n")), 0o644)
}
