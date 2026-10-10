package cli

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/ikauedev/forge/internal/doctor"
	"github.com/ikauedev/forge/internal/project"
	"github.com/ikauedev/forge/internal/selfupdate"
)

// scriptedUI responde ao menu com uma sequência de escolhas e confirmações.
type scriptedUI struct {
	choices  []string
	confirms []bool
	offered  [][]menuOption
}

func (s *scriptedUI) Choose(_ string, opts []menuOption) (string, error) {
	s.offered = append(s.offered, opts)
	if len(s.choices) == 0 {
		return actQuit, nil
	}
	c := s.choices[0]
	s.choices = s.choices[1:]
	return c, nil
}

func (s *scriptedUI) Confirm(string, bool) (bool, error) {
	if len(s.confirms) == 0 {
		return false, nil
	}
	c := s.confirms[0]
	s.confirms = s.confirms[1:]
	return c, nil
}

func keys(opts []menuOption) []string {
	var out []string
	for _, o := range opts {
		out = append(out, o.Key)
	}
	return out
}

func noUpdateCheck(t *testing.T) {
	t.Helper()
	t.Setenv("FORGE_NO_UPDATE_CHECK", "1")
}

func TestBareForgeWithoutTerminalPrintsHelp(t *testing.T) {
	withPrompt(t, &fakePrompter{}, false)
	out, err := run(t)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Available Commands") && !strings.Contains(out, "Comandos") {
		t.Errorf("deveria mostrar a ajuda:\n%s", out)
	}
	if strings.Contains(out, "Available Commands") || strings.Contains(out, "help for") {
		t.Errorf("a ajuda deveria estar em português:\n%s", out)
	}
	for _, want := range []string{"Uso:", "Comandos:", "new", "update", "install", "upgrade", "mostra a versão do forge"} {
		if !strings.Contains(out, want) {
			t.Errorf("a ajuda não lista %q", want)
		}
	}
}

func TestMenuShowsTheBannerAndRunsTheChosenAction(t *testing.T) {
	noUpdateCheck(t)
	ui := &scriptedUI{choices: []string{actList}, confirms: []bool{false}}
	old := newMenuUI
	newMenuUI = func() menuUI { return ui }
	t.Cleanup(func() { newMenuUI = old })
	withPrompt(t, &fakePrompter{}, true)

	out, err := run(t)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"|_|   \\___/", "forge ", "Templates (forge new --type <template>)", "quarkus-api"} {
		if !strings.Contains(out, want) {
			t.Errorf("saída sem %q:\n%s", want, out)
		}
	}
}

func TestMenuLoopsUntilTheUserLeaves(t *testing.T) {
	noUpdateCheck(t)
	ui := &scriptedUI{choices: []string{actList, actList, actQuit}, confirms: []bool{true, true}}
	cmd := NewRootCmd()
	var buf strings.Builder
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	if err := runMenu(cmd, ui); err != nil {
		t.Fatal(err)
	}
	if strings.Count(buf.String(), "Templates (forge new") != 2 || len(ui.offered) != 3 {
		t.Errorf("esperava 2 listagens e 3 menus (menus=%d)", len(ui.offered))
	}
}

func TestMenuOptionsAdaptToTheContext(t *testing.T) {
	base := state{version: "1.0.0", onPath: true}
	if got := keys(base.options()); strings.Join(got, ",") != "new,list,doctor,check,quit" {
		t.Errorf("base: %v", got)
	}

	inProject := base
	inProject.project = &project.Config{Template: "spring-api", Project: project.ProjectInfo{Name: "Orders"}}
	if got := keys(inProject.options()); !contains(got, actDiff) {
		t.Errorf("dentro de um projeto deveria oferecer o diff: %v", got)
	}
	if got := keys(base.options()); contains(got, actDiff) {
		t.Errorf("fora de um projeto não deve oferecer o diff: %v", got)
	}

	newer := base
	newer.newer = "1.2.0"
	got := keys(newer.options())
	if got[0] != actUpdate || contains(got, actCheck) {
		t.Errorf("com versão nova, atualizar vem primeiro e não repete a checagem: %v", got)
	}

	fresh := state{version: "dev", installable: true}
	if got := keys(fresh.options()); !contains(got, actInstall) {
		t.Errorf("sem estar no PATH deveria oferecer a instalação: %v", got)
	}
}

func contains(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

func TestBannerMentionsProjectUpdateAndPath(t *testing.T) {
	var b strings.Builder
	printBanner(&b, state{
		version: "1.0.0",
		project: &project.Config{Template: "react-app", Project: project.ProjectInfo{Name: "Web Shop"}},
		newer:   "1.1.0",
	}, false)
	for _, want := range []string{"Projeto detectado aqui: Web Shop (react-app)", "Nova versão disponível: 1.1.0", "ainda não está no PATH"} {
		if !strings.Contains(b.String(), want) {
			t.Errorf("banner sem %q:\n%s", want, b.String())
		}
	}
	b.Reset()
	printBanner(&b, state{version: "1.0.0", onPath: true}, false)
	if strings.Contains(b.String(), "PATH") || strings.Contains(b.String(), "\x1b[") {
		t.Errorf("sem aviso de PATH e sem cores quando colored=false:\n%q", b.String())
	}
}

func TestGatherStateDetectsAForgeProject(t *testing.T) {
	noUpdateCheck(t)
	dir := t.TempDir()
	if err := project.Write(dir, project.Config{Template: "quarkus-api", Project: project.ProjectInfo{Name: "Orders"}}); err != nil {
		t.Fatal(err)
	}
	st := gatherState(context.Background(), dir)
	if st.project == nil || st.project.Template != "quarkus-api" {
		t.Errorf("projeto não detectado: %+v", st.project)
	}
	if st := gatherState(context.Background(), t.TempDir()); st.project != nil {
		t.Error("pasta vazia não é projeto")
	}
}

// ---- forge update ----

func fakeReleases(t *testing.T, tag string, binary string) *httptest.Server {
	t.Helper()
	name := selfupdate.AssetName(strings.TrimPrefix(tag, "v"), goosForTest(), goarchForTest())
	archive, sum := buildArchive(t, name, binary)
	mux := http.NewServeMux()
	var srv *httptest.Server
	mux.HandleFunc("/repos/ikauedev/forge/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, `{"tag_name":%q,"assets":[{"name":%q,"url":%q},{"name":"checksums.txt","url":%q}]}`,
			tag, name, srv.URL+"/a/archive", srv.URL+"/a/sums")
	})
	mux.HandleFunc("/a/archive", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write(archive) })
	mux.HandleFunc("/a/sums", func(w http.ResponseWriter, r *http.Request) { _, _ = fmt.Fprintf(w, "%s  %s\n", sum, name) })
	srv = httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func useUpdateServer(t *testing.T, srv *httptest.Server, version, exe string) {
	t.Helper()
	oldClient, oldExe, oldVer := newUpdateClient, executablePath, Version
	newUpdateClient = func() selfupdate.Client { return selfupdate.Client{API: srv.URL, Repo: Repo} }
	executablePath = func() (string, error) { return exe, nil }
	Version = version
	t.Cleanup(func() { newUpdateClient, executablePath, Version = oldClient, oldExe, oldVer })
}

func TestUpdateCheckOnlyReportsTheNewVersion(t *testing.T) {
	withPrompt(t, &fakePrompter{}, false)
	exe := filepath.Join(t.TempDir(), "forge")
	_ = os.WriteFile(exe, []byte("antigo"), 0o755)
	useUpdateServer(t, fakeReleases(t, "v1.2.0", "novo"), "1.0.0", exe)

	out, err := run(t, "update", "--check")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Nova versão disponível: 1.2.0") || !strings.Contains(out, "forge update") {
		t.Errorf("saída:\n%s", out)
	}
	if b, _ := os.ReadFile(exe); string(b) != "antigo" {
		t.Error("--check não pode alterar o binário")
	}
}

func TestUpdateReplacesTheExecutable(t *testing.T) {
	withPrompt(t, &fakePrompter{}, false)
	exe := filepath.Join(t.TempDir(), "forge")
	_ = os.WriteFile(exe, []byte("antigo"), 0o755)
	useUpdateServer(t, fakeReleases(t, "v1.2.0", "novo"), "1.0.0", exe)

	out, err := run(t, "update", "--yes")
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if !strings.Contains(out, "atualizado para 1.2.0") {
		t.Errorf("saída:\n%s", out)
	}
	if b, _ := os.ReadFile(exe); string(b) != "novo" {
		t.Errorf("binário = %q", b)
	}
}

func TestUpdateAsksForConfirmationInATerminal(t *testing.T) {
	withPrompt(t, &fakePrompter{}, true)
	oldConfirm := confirmPrompt
	confirmPrompt = func(string, bool) (bool, error) { return false, nil }
	t.Cleanup(func() { confirmPrompt = oldConfirm })
	exe := filepath.Join(t.TempDir(), "forge")
	_ = os.WriteFile(exe, []byte("antigo"), 0o755)
	useUpdateServer(t, fakeReleases(t, "v1.2.0", "novo"), "1.0.0", exe)

	out, err := run(t, "update")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "cancelada") {
		t.Errorf("saída:\n%s", out)
	}
	if b, _ := os.ReadFile(exe); string(b) != "antigo" {
		t.Error("recusou, mas o binário mudou")
	}
}

func TestUpdateWhenAlreadyCurrent(t *testing.T) {
	withPrompt(t, &fakePrompter{}, false)
	exe := filepath.Join(t.TempDir(), "forge")
	_ = os.WriteFile(exe, []byte("atual"), 0o755)
	useUpdateServer(t, fakeReleases(t, "v1.2.0", "x"), "1.2.0", exe)

	out, err := run(t, "update")
	if err != nil || !strings.Contains(out, "já é a versão mais recente") {
		t.Errorf("%v\n%s", err, out)
	}
}

func TestUpdateOnAPrivateRepoExplainsWhatToDo(t *testing.T) {
	withPrompt(t, &fakePrompter{}, false)
	srv := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(srv.Close)
	useUpdateServer(t, srv, "1.0.0", filepath.Join(t.TempDir(), "forge"))

	_, err := run(t, "update", "--check")
	if err == nil || !strings.Contains(err.Error(), "GITHUB_TOKEN") || !strings.Contains(err.Error(), "go install") {
		t.Errorf("erro = %v", err)
	}
}

// ---- forge install ----

type cliFakeSys struct {
	home, userPath string
	writes         int
}

func (cliFakeSys) GOOS() string             { return "windows" }
func (f *cliFakeSys) Home() (string, error) { return f.home, nil }
func (f *cliFakeSys) Getenv(k string) string {
	return map[string]string{"LOCALAPPDATA": filepath.Join(f.home, "Local")}[k]
}
func (f *cliFakeSys) ReadUserPath() (string, error) { return f.userPath, nil }
func (f *cliFakeSys) WriteUserPath(v string) error {
	f.userPath = v
	f.writes++
	return nil
}

func TestInstallAndUninstallCommands(t *testing.T) {
	withPrompt(t, &fakePrompter{}, false)
	sys := &cliFakeSys{home: t.TempDir(), userPath: `C:\tools`}
	oldSys, oldExe := installSystem, executablePath
	installSystem = sys
	exe := filepath.Join(t.TempDir(), "forge-src")
	_ = os.WriteFile(exe, []byte("bin"), 0o755)
	executablePath = func() (string, error) { return exe, nil }
	t.Cleanup(func() { installSystem, executablePath = oldSys, oldExe })
	dir := filepath.Join(sys.home, "Local", "Programs", "forge")

	out, err := run(t, "install")
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if !strings.Contains(out, "forge instalado em") || !strings.Contains(out, "NOVO terminal") {
		t.Errorf("saída:\n%s", out)
	}
	if _, err := os.Stat(filepath.Join(dir, "forge.exe")); err != nil {
		t.Error("binário não instalado")
	}
	if !strings.HasSuffix(sys.userPath, dir) {
		t.Errorf("PATH = %q", sys.userPath)
	}

	out, err = run(t, "install")
	if err != nil || sys.writes != 1 {
		t.Errorf("segunda instalação não deve mexer no PATH: %v (escritas=%d)\n%s", err, sys.writes, out)
	}

	out, err = run(t, "uninstall")
	if err != nil || !strings.Contains(out, "Removido") || sys.userPath != `C:\tools` {
		t.Errorf("%v\nPATH=%q\n%s", err, sys.userPath, out)
	}
}

// ---- forge new: escolha do template e avisos ----

type typedPrompter struct {
	*fakePrompter
	typ string
}

func (p typedPrompter) ChooseType(choices []TypeChoice) (string, error) {
	p.called = true
	return p.typ, nil
}

func TestNewAsksWhichTemplateInATerminalAndWarnsAboutMissingTools(t *testing.T) {
	fp := &fakePrompter{name: "Web Shop"}
	withPrompt(t, typedPrompter{fakePrompter: fp, typ: "react-app"}, true)
	oldCheck := toolCheck
	toolCheck = func(tpl string) []doctor.Result {
		return []doctor.Result{{Name: "Node.js", Status: doctor.Warn, Detail: "não encontrado", Hint: "instale o Node 22+ (" + tpl + ")"}}
	}
	t.Cleanup(func() { toolCheck = oldCheck })
	out := t.TempDir()

	stdout, err := run(t, "new", "--out", out, "--no-git")
	if err != nil {
		t.Fatalf("%v\n%s", err, stdout)
	}
	if _, err := os.Stat(filepath.Join(out, "web-shop", "src", "core", "auth", "guards.tsx")); err != nil {
		t.Errorf("deveria ter gerado o react-app escolhido no formulário: %v", err)
	}
	if !strings.Contains(stdout, "aviso: Node.js: não encontrado (instale o Node 22+ (react-app))") {
		t.Errorf("faltou o aviso de ferramenta:\n%s", stdout)
	}
}

func TestNewDoesNotAskTheTemplateWhenTypeIsGiven(t *testing.T) {
	withPrompt(t, typedPrompter{fakePrompter: &fakePrompter{name: "Api"}, typ: "react-app"}, true)
	out := t.TempDir()
	if _, err := run(t, "new", "Orders Api", "--type", "quarkus-api", "--db", "none", "--vpc", "new", "--auth", "none", "--package", "com.empresa", "--out", out, "--no-git"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(out, "orders-api", "pom.xml")); err != nil {
		t.Error("o --type explícito deve valer")
	}
}

func goosForTest() string   { return runtime.GOOS }
func goarchForTest() string { return runtime.GOARCH }

// buildArchive monta o .tar.gz (ou .zip, no Windows) da release com o executável e devolve o SHA-256.
func buildArchive(t *testing.T, name, binary string) ([]byte, string) {
	t.Helper()
	exeName := "forge"
	if strings.HasSuffix(name, ".zip") {
		exeName = "forge.exe"
	}
	var buf bytes.Buffer
	if strings.HasSuffix(name, ".zip") {
		zw := zip.NewWriter(&buf)
		w, _ := zw.Create(exeName)
		_, _ = w.Write([]byte(binary))
		_ = zw.Close()
	} else {
		gz := gzip.NewWriter(&buf)
		tw := tar.NewWriter(gz)
		_ = tw.WriteHeader(&tar.Header{Name: exeName, Mode: 0o755, Size: int64(len(binary)), Typeflag: tar.TypeReg})
		_, _ = tw.Write([]byte(binary))
		_ = tw.Close()
		_ = gz.Close()
	}
	sum := sha256.Sum256(buf.Bytes())
	return buf.Bytes(), hex.EncodeToString(sum[:])
}
