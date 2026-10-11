package cli

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/charmbracelet/x/term"
	"github.com/spf13/cobra"

	"github.com/ikauedev/forge/internal/doctor"
	"github.com/ikauedev/forge/internal/engine"
	"github.com/ikauedev/forge/internal/names"
	"github.com/ikauedev/forge/internal/project"
	"github.com/ikauedev/forge/internal/template"
	"github.com/ikauedev/forge/templates"
)

// Version é a versão do forge; o release define via -ldflags "-X .../internal/cli.Version=...".
var Version = "dev"

const (
	defaultPackage = "com.example"
	defaultType    = templates.SpringAPI

	// k8sOption liga os manifests Kubernetes; "--k8s" sem valor equivale a "--k8s kustomize".
	k8sOption  = "k8s"
	k8sEnabled = "kustomize"
)

// Answers reúne tudo o que define um projeto gerado.
type Answers struct {
	Type       string // template (spring-api, angular-app)
	Name       string
	Package    string
	PackageSet bool // o pacote veio de flag (não precisa perguntar)
	Options    map[string]string
}

// Prompter completa o que faltar em Answers perguntando ao usuário.
type Prompter interface {
	Fill(m *template.Manifest, a *Answers) error
}

// Variáveis trocadas nos testes.
var (
	newPrompter Prompter = huhPrompter{}
	isTerminal           = stdinIsTerminal
)

type newFlags struct {
	typ     string
	pkg     string
	out     string
	noGit   bool
	yes     bool
	options map[string]*string
}

func loadTemplate(name string) (fs.FS, *template.Manifest, error) {
	if !slices.Contains(templates.Names, name) {
		return nil, nil, fmt.Errorf("template %q desconhecido; use um de: %s", name, strings.Join(templates.Names, ", "))
	}
	src, err := fs.Sub(templates.FS, name)
	if err != nil {
		return nil, nil, err
	}
	m, err := engine.LoadManifest(src)
	return src, m, err
}

// optionFlags junta as opções de todos os templates, uma flag por nome. Qual vale é decidido em
// runNew, pelo --type escolhido.
func optionFlags() ([]template.Option, error) {
	var all []template.Option
	seen := map[string]bool{}
	for _, name := range templates.Names {
		_, m, err := loadTemplate(name)
		if err != nil {
			return nil, err
		}
		for _, o := range m.Options {
			if !seen[o.Name] {
				seen[o.Name] = true
				all = append(all, o)
			}
		}
	}
	return all, nil
}

func newNewCmd() *cobra.Command {
	opts, err := optionFlags()
	if err != nil {
		panic(err) // manifesto embutido inválido: erro de build do forge
	}
	f := newFlags{options: map[string]*string{}}
	cmd := &cobra.Command{
		Use:   "new [nome]",
		Short: "Cria um novo projeto (Spring Boot, Quarkus, Angular ou React)",
		Long: `Cria um projeto a partir de um template: spring-api (padrão), quarkus-api, angular-app ou react-app.

Em um terminal, o que faltar (template, nome, pacote e opções) é perguntado em um formulário; ao escolher o
template o forge avisa se falta algo na máquina para compilá-lo (Java, Docker, Node...).
Com --yes (ou sem terminal) usa os padrões para o que não for informado.

Opções por template:
  spring-api   --package --db --vpc --http --auth (none|jwt|cognito) --cache --k8s
  quarkus-api  --package --db --vpc --auth (none|jwt) --k8s
  angular-app  --k8s
  react-app    --k8s

--k8s gera k8s/ (Kustomize: base + overlays dev, hml e prd), docs/KUBERNETES.md e o pipeline de deploy no cluster.`,
		Example: `  forge new                                  # formulário completo
  forge new "Payments API"
  forge new payments-api --package com.empresa --http feign --auth cognito --cache redis
  forge new orders --type quarkus-api --db mysql
  forge new storefront --type react-app
  forge new orders --type quarkus-api --k8s    # com manifests Kubernetes para dev, hml e prd
  forge new payments-api --yes`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runNew(cmd, args, f)
		},
	}
	cmd.Flags().StringVar(&f.typ, "type", defaultType, "template: "+strings.Join(templates.Names, " | "))
	cmd.Flags().StringVar(&f.pkg, "package", defaultPackage, "pacote base Java (spring-api)")
	cmd.Flags().StringVar(&f.out, "out", ".", "diretório pai; o projeto é criado em <out>/<nome-kebab>")
	cmd.Flags().BoolVar(&f.noGit, "no-git", false, "não executar git init")
	cmd.Flags().BoolVarP(&f.yes, "yes", "y", false, "não perguntar nada; usar os padrões")
	for _, o := range opts {
		f.options[o.Name] = cmd.Flags().String(o.Name, "",
			fmt.Sprintf("%s (%s; padrão %s)", o.Description, strings.Join(o.Values, "|"), o.Default))
		if o.Name == k8sOption {
			cmd.Flags().Lookup(o.Name).NoOptDefVal = k8sEnabled // "--k8s" sozinho liga
		}
	}
	return cmd
}

// TypeChoice descreve um template para a pergunta "qual projeto?".
type TypeChoice struct {
	Name, Description string
}

// TypePrompter é implementado por quem sabe perguntar o template (o formulário real). É opcional:
// sem ele, vale o --type (padrão spring-api).
type TypePrompter interface {
	ChooseType(choices []TypeChoice) (string, error)
}

func typeChoices() ([]TypeChoice, error) {
	var out []TypeChoice
	for _, name := range templates.Names {
		_, m, err := loadTemplate(name)
		if err != nil {
			return nil, err
		}
		out = append(out, TypeChoice{Name: name, Description: m.Description})
	}
	return out, nil
}

// toolCheck é trocado nos testes.
var toolCheck = func(template string) []doctor.Result { return doctor.ForTemplate(doctor.OSEnv{}, template) }

// warnMissingTools avisa, já na escolha do template, o que falta na máquina para compilá-lo.
// Não impede a geração: o projeto sai igual e o usuário pode instalar depois.
func warnMissingTools(w io.Writer, template string) {
	for _, r := range toolCheck(template) {
		if r.Status != doctor.Fail && r.Status != doctor.Warn {
			continue
		}
		msg := fmt.Sprintf("aviso: %s: %s", r.Name, r.Detail)
		if r.Hint != "" {
			msg += " (" + r.Hint + ")"
		}
		_, _ = fmt.Fprintln(w, msg)
	}
}

func runNew(cmd *cobra.Command, args []string, f newFlags) error {
	interactive := !f.yes && isTerminal()
	if interactive && !cmd.Flags().Changed("type") {
		if tp, ok := newPrompter.(TypePrompter); ok {
			choices, err := typeChoices()
			if err != nil {
				return err
			}
			if f.typ, err = tp.ChooseType(choices); err != nil {
				return err
			}
		}
	}
	_, manifest, err := loadTemplate(f.typ)
	if err != nil {
		return err
	}

	a := Answers{Type: f.typ, Package: f.pkg, PackageSet: cmd.Flags().Changed("package"), Options: map[string]string{}}
	if len(args) == 1 {
		a.Name = args[0]
	}
	if a.PackageSet && !manifest.JavaPackage {
		return fmt.Errorf("--package não se aplica ao template %s", f.typ)
	}
	for name, v := range f.options {
		if *v == "" {
			continue
		}
		if _, ok := manifest.Option(name); !ok {
			return fmt.Errorf("--%s não se aplica ao template %s", name, f.typ)
		}
		a.Options[name] = *v
	}

	if interactive {
		warnMissingTools(cmd.ErrOrStderr(), f.typ)
	}

	missing := a.Name == "" || (manifest.JavaPackage && !a.PackageSet) || len(a.Options) < len(manifest.Options)
	if missing && !f.yes && isTerminal() {
		if err := newPrompter.Fill(manifest, &a); err != nil {
			return err
		}
	}
	if a.Name == "" {
		return fmt.Errorf("informe o nome do projeto (argumento ou modo interativo)")
	}
	return createProject(cmd.OutOrStdout(), cmd.ErrOrStderr(), a, f.out, f.noGit)
}

// createProject é o caminho único de geração, usado pelo fluxo interativo e pelo de flags.
func createProject(w, errW io.Writer, a Answers, out string, noGit bool) error {
	n, err := names.New(a.Name)
	if err != nil {
		return err
	}
	if a.Type == "" {
		a.Type = defaultType
	}
	src, manifest, err := loadTemplate(a.Type)
	if err != nil {
		return err
	}
	if manifest.JavaPackage {
		if err := names.ValidatePackage(a.Package); err != nil {
			return err
		}
	} else {
		a.Package = ""
	}
	sel, err := manifest.Resolve(a.Options)
	if err != nil {
		return err
	}

	dest := filepath.Join(out, n.Kebab)
	files, err := engine.Generate(src, engine.Options{Names: n, Package: a.Package, Dest: dest, Selection: sel})
	if err != nil {
		return err
	}
	cfg := project.Config{
		Forge:    project.ForgeInfo{Version: Version},
		Template: manifest.Name,
		Project:  project.ProjectInfo{Name: strings.TrimSpace(a.Name), Package: a.Package},
		Options:  sel,
	}
	if err := project.Write(dest, cfg); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(w, "Projeto %q criado em %s (%d arquivos)\n", n.Title, dest, len(files)+1)
	for _, k := range sel.SortedKeys() {
		_, _ = fmt.Fprintf(w, "  %s = %s\n", k, sel[k])
	}

	if !noGit {
		if err := gitInit(dest); err != nil {
			_, _ = fmt.Fprintf(errW, "aviso: git init falhou: %v\n", err)
		}
	}

	_, _ = fmt.Fprintf(w, "\nPróximos passos:\n  cd %s\n", dest)
	for _, step := range manifest.NextSteps {
		_, _ = fmt.Fprintf(w, "  %s\n", step)
	}
	return nil
}

func gitInit(dir string) error {
	c := exec.Command("git", "init", "-q")
	c.Dir = dir
	c.Stderr = os.Stderr
	return c.Run()
}

// stdinIsTerminal usa a checagem real de TTY: no Windows o dispositivo NUL parece "char device".
func stdinIsTerminal() bool {
	return term.IsTerminal(os.Stdin.Fd())
}
