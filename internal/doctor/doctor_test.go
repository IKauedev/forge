package doctor

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type fakeEnv struct {
	bins    map[string]string // binário -> saída de versão
	daemon  bool
	env     map[string]string
	wd      string
	goos    string
	execErr map[string]error
}

func (f fakeEnv) LookPath(file string) (string, error) {
	if _, ok := f.bins[file]; ok {
		return "/bin/" + file, nil
	}
	return "", errors.New("not found")
}

func (f fakeEnv) Output(name string, args ...string) (string, error) {
	base := filepath.Base(name)
	base = strings.TrimSuffix(base, ".exe")
	if base == "docker" && len(args) > 0 && args[0] == "info" {
		if f.daemon {
			return "29.0.0", nil
		}
		return "", errors.New("daemon down")
	}
	if out, ok := f.bins[base]; ok {
		return out, f.execErr[base]
	}
	return "", errors.New("unknown")
}

func (f fakeEnv) Getenv(k string) string { return f.env[k] }
func (f fakeEnv) Getwd() (string, error) { return f.wd, nil }
func (f fakeEnv) GOOS() string {
	if f.goos == "" {
		return "linux"
	}
	return f.goos
}

func healthy(t *testing.T) fakeEnv {
	return fakeEnv{
		bins: map[string]string{
			"java":      `openjdk version "25.0.3" 2026-04-21 LTS`,
			"docker":    "Docker version 29.5.3, build abc",
			"git":       "git version 2.50.1",
			"terraform": "Terraform v1.15.8",
			"aws":       "aws-cli/2.27.0 Python/3.13",
			"node":      "v26.3.0",
		},
		daemon: true,
		env:    map[string]string{},
		wd:     t.TempDir(),
	}
}

func byName(rs []Result, name string) Result {
	for _, r := range rs {
		if r.Name == name {
			return r
		}
	}
	return Result{}
}

func TestHealthyEnvironment(t *testing.T) {
	rs := Run(healthy(t))
	if Failed(rs) != 0 {
		t.Fatalf("não deveria falhar: %+v", rs)
	}
	if got := byName(rs, "Java"); got.Status != OK || got.Detail != "25.0.3" {
		t.Errorf("java: %+v", got)
	}
	if got := byName(rs, "Maven Wrapper"); got.Status != Skip {
		t.Errorf("fora de projeto deveria ser Skip: %+v", got)
	}
}

func TestMissingDockerIsFailure(t *testing.T) {
	env := healthy(t)
	delete(env.bins, "docker")
	rs := Run(env)
	if got := byName(rs, "Docker"); got.Status != Fail {
		t.Errorf("docker ausente deveria ser Fail: %+v", got)
	}
	if Failed(rs) != 1 {
		t.Errorf("deveria haver exatamente 1 falha, houve %d", Failed(rs))
	}
}

func TestDockerDaemonDownIsOnlyWarning(t *testing.T) {
	env := healthy(t)
	env.daemon = false
	rs := Run(env)
	if got := byName(rs, "Docker"); got.Status != Warn {
		t.Errorf("daemon parado deveria ser Warn: %+v", got)
	}
	if Failed(rs) != 0 {
		t.Error("daemon parado não deve reprovar")
	}
}

func TestOldJavaFails(t *testing.T) {
	env := healthy(t)
	env.bins["java"] = `openjdk version "17.0.13" 2024-10-15 LTS`
	if got := byName(Run(env), "Java"); got.Status != Fail || !strings.Contains(got.Hint, "25") {
		t.Errorf("java 17 deveria falhar pedindo 25: %+v", got)
	}
}

func TestJavaHomeTakesPrecedence(t *testing.T) {
	env := healthy(t)
	delete(env.bins, "java")
	home := t.TempDir()
	env.env["JAVA_HOME"] = home
	env.bins["java"] = `openjdk version "25.0.3"`
	if got := byName(Run(env), "Java"); got.Status != OK {
		t.Errorf("JAVA_HOME válido deveria passar: %+v", got)
	}
}

func TestOptionalToolsOnlyWarn(t *testing.T) {
	env := healthy(t)
	delete(env.bins, "terraform")
	delete(env.bins, "aws")
	env.bins["node"] = "v18.20.0"
	rs := Run(env)
	for _, n := range []string{"Terraform", "AWS CLI", "Node.js"} {
		if got := byName(rs, n); got.Status != Warn {
			t.Errorf("%s deveria ser Warn: %+v", n, got)
		}
	}
	if Failed(rs) != 0 {
		t.Error("opcionais não reprovam")
	}
}

func TestMavenWrapperExecutableBit(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("bits POSIX")
	}
	env := healthy(t)
	if err := os.WriteFile(filepath.Join(env.wd, "mvnw"), []byte("#!/bin/sh"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := byName(Run(env), "Maven Wrapper"); got.Status != Fail {
		t.Errorf("mvnw sem +x deveria falhar: %+v", got)
	}
	if err := os.Chmod(filepath.Join(env.wd, "mvnw"), 0o755); err != nil {
		t.Fatal(err)
	}
	if got := byName(Run(env), "Maven Wrapper"); got.Status != OK {
		t.Errorf("mvnw com +x deveria passar: %+v", got)
	}
}

func TestVersionLess(t *testing.T) {
	for _, tt := range []struct {
		have, min string
		want      bool
	}{
		{"1.15.8", "1.10", false}, {"1.9.0", "1.10", true}, {"1.10.0", "1.10", false},
		{"18.20.0", "22", true}, {"26.3.0", "22", false}, {"", "22", true},
	} {
		if got := versionLess(tt.have, tt.min); got != tt.want {
			t.Errorf("versionLess(%q,%q)=%v", tt.have, tt.min, got)
		}
	}
}

func TestPrintColors(t *testing.T) {
	rs := []Result{{Name: "Docker", Status: Fail, Detail: "não encontrado", Hint: "instale"}}
	var plain, colored bytes.Buffer
	Print(&plain, rs, false)
	Print(&colored, rs, true)
	if strings.Contains(plain.String(), "\x1b[") {
		t.Error("sem cor não deve ter ANSI")
	}
	if !strings.Contains(colored.String(), "\x1b[31m") || !strings.Contains(plain.String(), "instale") {
		t.Errorf("saída colorida/hint incorretos:\n%s", colored.String())
	}
}

// Teste com o PATH real manipulado: só existe um "java" falso, sem docker nem git.
func TestRealPATHWithoutDocker(t *testing.T) {
	dir := t.TempDir()
	name, body := "java", "#!/bin/sh\necho 'openjdk version \"25.0.3\"' >&2\n"
	if runtime.GOOS == "windows" {
		name, body = "java.cmd", "@echo off\r\necho openjdk version \"25.0.3\" 1>&2\r\n"
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	t.Setenv("JAVA_HOME", "")

	rs := Run(OSEnv{})

	if got := byName(rs, "Java"); got.Status != OK {
		t.Errorf("java falso deveria passar: %+v", got)
	}
	if got := byName(rs, "Docker"); got.Status != Fail {
		t.Errorf("docker deveria faltar: %+v", got)
	}
	if Failed(rs) < 1 {
		t.Error("ambiente sem Docker deveria ter falhas")
	}
}

func TestForTemplateChecksOnlyWhatTheTemplateNeeds(t *testing.T) {
	env := healthy(t)
	for _, tpl := range []string{"spring-api", "quarkus-api"} {
		rs := ForTemplate(env, tpl)
		if len(rs) != 2 || rs[0].Name != "Java" || rs[1].Name != "Docker" {
			t.Errorf("%s: %+v", tpl, rs)
		}
	}
	for _, tpl := range []string{"angular-app", "react-app"} {
		rs := ForTemplate(env, tpl)
		if len(rs) != 1 || rs[0].Name != "Node.js" || rs[0].Status != OK {
			t.Errorf("%s: %+v", tpl, rs)
		}
	}
	if rs := ForTemplate(env, "outro"); rs != nil {
		t.Errorf("template desconhecido: %+v", rs)
	}
}

func TestForTemplateReportsWhatIsMissing(t *testing.T) {
	env := healthy(t)
	delete(env.bins, "java")
	delete(env.bins, "node")
	if r := byName(ForTemplate(env, "quarkus-api"), "Java"); r.Status != Fail {
		t.Errorf("Java ausente deveria falhar: %+v", r)
	}
	if r := byName(ForTemplate(env, "react-app"), "Node.js"); r.Status != Warn || r.Hint == "" {
		t.Errorf("Node ausente deveria avisar: %+v", r)
	}
}
