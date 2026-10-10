// Package doctor verifica se a máquina tem o necessário para gerar e validar os projetos.
package doctor

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

// Status do resultado de uma checagem.
type Status int

const (
	OK   Status = iota // tudo certo
	Warn               // opcional ausente/antigo, ou algo que só atrapalha parte do fluxo
	Fail               // obrigatório ausente/antigo
	Skip               // não se aplica aqui
)

// Result é o resultado de uma checagem.
type Result struct {
	Name   string
	Status Status
	Detail string
	Hint   string
}

// Env isola o acesso ao sistema para os testes.
type Env interface {
	LookPath(file string) (string, error)
	// Output executa o comando e devolve stdout+stderr juntos.
	Output(name string, args ...string) (string, error)
	Getenv(key string) string
	Getwd() (string, error)
	GOOS() string
}

// OSEnv usa o sistema real.
type OSEnv struct{}

func (OSEnv) LookPath(f string) (string, error) { return exec.LookPath(f) }
func (OSEnv) Output(name string, args ...string) (string, error) {
	out, err := exec.Command(name, args...).CombinedOutput()
	return string(out), err
}
func (OSEnv) Getenv(k string) string { return os.Getenv(k) }
func (OSEnv) Getwd() (string, error) { return os.Getwd() }
func (OSEnv) GOOS() string           { return runtime.GOOS }

// Versões mínimas. Java vem do template (java.version=25). Terraform 1.10 introduziu o lock nativo
// do backend S3 e Node 22 é o piso do Angular atual; ambos a reconfirmar nas Fases 5 e 7.
const (
	MinJava      = 25
	MinTerraform = "1.10"
	MinNode      = 22
)

var versionRe = regexp.MustCompile(`(\d+)(?:\.(\d+))?(?:\.(\d+))?`)

// Run executa todas as checagens, nesta ordem.
func Run(env Env) []Result {
	return []Result{
		checkJava(env),
		checkDocker(env),
		checkGit(env),
		checkMavenWrapper(env),
		checkOptional(env, "terraform", []string{"version"}, "Terraform", MinTerraform, "necessário para validar a infra (Fases 5 e 7)"),
		checkOptional(env, "aws", []string{"--version"}, "AWS CLI", "", "necessário para operar a conta AWS"),
		checkOptional(env, "node", []string{"--version"}, "Node.js", strconv.Itoa(MinNode), "necessário para o template Angular"),
	}
}

// Failed conta as checagens obrigatórias que falharam.
func Failed(rs []Result) int {
	n := 0
	for _, r := range rs {
		if r.Status == Fail {
			n++
		}
	}
	return n
}

func javaBinary(env Env) (string, error) {
	if home := env.Getenv("JAVA_HOME"); home != "" {
		name := "java"
		if env.GOOS() == "windows" {
			name = "java.exe"
		}
		return filepath.Join(home, "bin", name), nil
	}
	return env.LookPath("java")
}

func checkJava(env Env) Result {
	r := Result{Name: "Java"}
	bin, err := javaBinary(env)
	if err != nil {
		r.Status, r.Detail = Fail, "não encontrado"
		r.Hint = fmt.Sprintf("instale um JDK %d+ e defina JAVA_HOME", MinJava)
		return r
	}
	out, err := env.Output(bin, "-version")
	if err != nil && out == "" {
		r.Status, r.Detail = Fail, "JAVA_HOME aponta para algo que não executa"
		r.Hint = "confira JAVA_HOME"
		return r
	}
	major, full := parseVersion(out)
	r.Detail = full
	if major < MinJava {
		r.Status = Fail
		r.Hint = fmt.Sprintf("os projetos gerados exigem JDK %d+ (defina JAVA_HOME)", MinJava)
		return r
	}
	r.Status = OK
	return r
}

func checkDocker(env Env) Result {
	r := Result{Name: "Docker"}
	if _, err := env.LookPath("docker"); err != nil {
		r.Status, r.Detail = Fail, "não encontrado"
		r.Hint = "os testes (Testcontainers) e o compose precisam do Docker"
		return r
	}
	out, _ := env.Output("docker", "--version")
	_, r.Detail = parseVersion(out)
	if _, err := env.Output("docker", "info", "--format", "{{.ServerVersion}}"); err != nil {
		r.Status = Warn
		r.Detail += " (daemon parado)"
		r.Hint = "inicie o Docker antes de rodar testes e compose"
		return r
	}
	r.Status = OK
	return r
}

func checkGit(env Env) Result {
	r := Result{Name: "Git"}
	if _, err := env.LookPath("git"); err != nil {
		r.Status, r.Detail = Fail, "não encontrado"
		r.Hint = "o forge new executa git init"
		return r
	}
	out, _ := env.Output("git", "--version")
	_, r.Detail = parseVersion(out)
	r.Status = OK
	return r
}

func checkMavenWrapper(env Env) Result {
	r := Result{Name: "Maven Wrapper"}
	wd, err := env.Getwd()
	if err != nil {
		r.Status, r.Detail = Skip, "diretório atual indisponível"
		return r
	}
	info, err := os.Stat(filepath.Join(wd, "mvnw"))
	if err != nil {
		r.Status, r.Detail = Skip, "fora de um projeto (sem mvnw)"
		return r
	}
	if env.GOOS() != "windows" && info.Mode().Perm()&0o111 == 0 {
		r.Status, r.Detail = Fail, "mvnw sem permissão de execução"
		r.Hint = "chmod +x mvnw"
		return r
	}
	r.Status, r.Detail = OK, "mvnw presente"
	return r
}

func checkOptional(env Env, bin string, args []string, name, min, why string) Result {
	r := Result{Name: name}
	if _, err := env.LookPath(bin); err != nil {
		r.Status, r.Detail, r.Hint = Warn, "não encontrado", why
		return r
	}
	out, _ := env.Output(bin, args...)
	_, r.Detail = parseVersion(out)
	if min != "" && versionLess(r.Detail, min) {
		r.Status, r.Hint = Warn, fmt.Sprintf("versão mínima %s (%s)", min, why)
		return r
	}
	r.Status = OK
	return r
}

// parseVersion devolve a versão maior e o texto "X.Y.Z" da primeira ocorrência na saída.
func parseVersion(out string) (int, string) {
	m := versionRe.FindString(out)
	if m == "" {
		return 0, strings.TrimSpace(out)
	}
	major, _ := strconv.Atoi(versionRe.FindStringSubmatch(m)[1])
	return major, m
}

// versionLess compara "a.b.c" com o mínimo "a.b" (ou "a").
func versionLess(have, min string) bool {
	h := strings.Split(have, ".")
	m := strings.Split(min, ".")
	for i := range m {
		mi, _ := strconv.Atoi(m[i])
		hi := 0
		if i < len(h) {
			hi, _ = strconv.Atoi(h[i])
		}
		if hi != mi {
			return hi < mi
		}
	}
	return false
}

// Print escreve o relatório. colored liga as cores ANSI.
func Print(w io.Writer, rs []Result, colored bool) {
	paint := func(code, s string) string {
		if !colored {
			return s
		}
		return "\x1b[" + code + "m" + s + "\x1b[0m"
	}
	for _, r := range rs {
		var mark string
		switch r.Status {
		case OK:
			mark = paint("32", "✓")
		case Warn:
			mark = paint("33", "!")
		case Fail:
			mark = paint("31", "✗")
		default:
			mark = paint("90", "-")
		}
		_, _ = fmt.Fprintf(w, " %s %-14s %s\n", mark, r.Name, r.Detail)
		if r.Hint != "" && r.Status != OK {
			_, _ = fmt.Fprintf(w, "   %s\n", paint("90", r.Hint))
		}
	}
}
