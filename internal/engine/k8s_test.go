package engine_test

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/ikauedev/forge/internal/catalog"
	"github.com/ikauedev/forge/internal/engine"
	"github.com/ikauedev/forge/internal/names"
	"github.com/ikauedev/forge/internal/template"
	"github.com/ikauedev/forge/templates"
)

const (
	k8sOpt  = "k8s"
	k8sOn   = "kustomize"
	k8sApp  = "Orders Api"
	k8sName = "orders-api"
)

var (
	k8sEnvs      = []string{"dev", "hml", "prd"}
	k8sWorkflows = []string{"deploy-k8s.yml", "deploy-k8s-env.yml"}
	k8sBaseFiles = []string{
		"deployment.yaml", "service.yaml", "ingress.yaml", "hpa.yaml", "pdb.yaml",
		"networkpolicy.yaml", "serviceaccount.yaml", "kustomization.yaml",
	}
)

// genK8s gera o template tpl com a seleção sel e devolve o diretório do projeto.
func genK8s(t *testing.T, tpl, raw string, sel template.Selection) string {
	t.Helper()
	n, err := names.New(raw)
	if err != nil {
		t.Fatal(err)
	}
	src, err := fs.Sub(templates.FS, tpl)
	if err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(t.TempDir(), n.Kebab)
	if _, err := engine.Generate(src, engine.Options{Names: n, Package: "com.empresa", Dest: dest, Selection: sel}); err != nil {
		t.Fatal(err)
	}
	return dest
}

// k8sTemplates lista cada template com a seleção que liga o Kubernetes e o nome de referência dele.
func k8sTemplates() map[string]template.Selection {
	return map[string]template.Selection{
		templates.SpringAPI:  {k8sOpt: k8sOn},
		templates.QuarkusAPI: {k8sOpt: k8sOn},
		templates.AngularApp: {k8sOpt: k8sOn},
		templates.ReactApp:   {k8sOpt: k8sOn},
	}
}

func TestKubernetesIsOffByDefault(t *testing.T) {
	for _, tpl := range templates.Names {
		t.Run(tpl, func(t *testing.T) {
			dest := genK8s(t, tpl, k8sApp, nil)
			for _, rel := range []string{"k8s", "docs/KUBERNETES.md", ".github/workflows/deploy-k8s.yml", ".github/workflows/deploy-k8s-env.yml"} {
				if exists(dest, rel) {
					t.Errorf("%s não deveria existir com k8s=none", rel)
				}
			}
			if strings.Contains(read(t, dest, ".github/workflows/ci.yml"), "k8s") {
				t.Error("o ci.yml não deveria citar k8s com k8s=none")
			}
		})
	}
}

func TestKubernetesFilesAndNames(t *testing.T) {
	for tpl, sel := range k8sTemplates() {
		t.Run(tpl, func(t *testing.T) {
			dest := genK8s(t, tpl, k8sApp, sel)

			want := []string{"docs/KUBERNETES.md"}
			for _, f := range k8sBaseFiles {
				want = append(want, "k8s/base/"+f)
			}
			for _, e := range k8sEnvs {
				want = append(want, "k8s/overlays/"+e+"/kustomization.yaml", "k8s/overlays/"+e+"/namespace.yaml")
			}
			for _, w := range k8sWorkflows {
				want = append(want, ".github/workflows/"+w)
			}
			for _, rel := range want {
				if !exists(dest, rel) {
					t.Errorf("arquivo ausente: %s", rel)
				}
			}

			_ = filepath.WalkDir(filepath.Join(dest, "k8s"), func(p string, d os.DirEntry, err error) error {
				if err != nil || d.IsDir() {
					return err
				}
				rel, _ := filepath.Rel(dest, p)
				b, _ := os.ReadFile(p)
				text := string(b)
				if directiveLeft.MatchString(text) || markerLeft.MatchString(text) {
					t.Errorf("%s: sobrou diretiva ou marcador", rel)
				}
				if sentinelLeft.MatchString(text) || angularLeftover.MatchString(text) {
					t.Errorf("%s: sobrou sentinela", rel)
				}
				var doc map[string]any
				if err := yaml.Unmarshal(b, &doc); err != nil {
					t.Errorf("%s: YAML inválido: %v", rel, err)
				}
				return nil
			})

			// namespace de cada ambiente = <app>-<ambiente>, igual ao que o workflow de deploy espera
			for _, e := range k8sEnvs {
				ns := read(t, dest, "k8s/overlays/"+e+"/namespace.yaml")
				kz := read(t, dest, "k8s/overlays/"+e+"/kustomization.yaml")
				for _, text := range []string{ns, kz} {
					if !strings.Contains(text, k8sName+"-"+e) {
						t.Errorf("overlay %s sem o namespace %s-%s", e, k8sName, e)
					}
				}
			}
			checkK8sWorkflows(t, dest)
		})
	}
}

// checkK8sWorkflows aplica aos workflows de deploy no cluster as mesmas regras dos demais:
// actions fixadas por SHA, OIDC (sem access key) e permissões mínimas.
func checkK8sWorkflows(t *testing.T, dest string) {
	t.Helper()
	for _, name := range k8sWorkflows {
		rel := ".github/workflows/" + name
		text := read(t, dest, rel)
		if staticAWSKeys.MatchString(text) {
			t.Errorf("%s: credencial estática da AWS", rel)
		}
		for _, m := range usesLine.FindAllStringSubmatch(text, -1) {
			ref := m[1]
			if strings.HasPrefix(ref, "./") {
				if ref != "./.github/workflows/deploy-k8s-env.yml" {
					t.Errorf("%s: workflow local inesperado %s", rel, ref)
				}
				continue
			}
			if !pinnedSHA.MatchString(ref + "\n") {
				t.Errorf("%s: action sem SHA fixo: %s", rel, ref)
			}
		}
		if sentinelLeft.MatchString(text) || angularLeftover.MatchString(text) {
			t.Errorf("%s: sobrou sentinela", rel)
		}
		var doc struct {
			Permissions map[string]string `yaml:"permissions"`
		}
		if err := yaml.Unmarshal([]byte(text), &doc); err != nil {
			t.Errorf("%s: YAML inválido: %v", rel, err)
			continue
		}
		if name == "deploy-k8s.yml" && doc.Permissions["contents"] != "read" {
			t.Errorf("%s: permissions de topo deve ser contents: read", rel)
		}
	}
	// os três ambientes, na ordem dev -> hml -> prd, e a trava que evita competir com o deploy do ECS
	text := read(t, dest, ".github/workflows/deploy-k8s.yml")
	for _, want := range []string{"environment: dev", "environment: hml", "environment: prd", "K8S_DEPLOY_ENABLED", "needs: [build-push, deploy-hml]"} {
		if !strings.Contains(text, want) {
			t.Errorf("deploy-k8s.yml sem %q", want)
		}
	}
	if !strings.Contains(read(t, dest, ".github/workflows/ci.yml"), "k8s-manifests") {
		t.Error("o ci.yml deveria validar os overlays com k8s=kustomize")
	}
}

// Para os backends: o que depende de db, auth e cache precisa aparecer (ou sumir) nos manifests.
func TestKubernetesBackendVariants(t *testing.T) {
	cat, err := catalog.Load()
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		tpl  string
		sels []template.Selection
	}{
		{templates.SpringAPI, k8sCombos(manifestOf(t, templates.SpringAPI), "vpc", "http")},
		{templates.QuarkusAPI, k8sCombos(manifestOf(t, templates.QuarkusAPI), "vpc")},
	}
	for _, c := range cases {
		for _, sel := range c.sels {
			t.Run(c.tpl+"/"+label(sel), func(t *testing.T) {
				dest := genK8s(t, c.tpl, k8sApp, sel)
				dep := read(t, dest, "k8s/base/deployment.yaml")
				np := read(t, dest, "k8s/base/networkpolicy.yaml")
				base := read(t, dest, "k8s/base/kustomization.yaml")
				overlays := ""
				for _, e := range k8sEnvs {
					overlays += read(t, dest, "k8s/overlays/"+e+"/kustomization.yaml")
				}
				all := dep + np + base + overlays

				if db := sel["db"]; db == "none" {
					for _, banned := range []string{"DB_HOST", "DB_URL", "DB_USERNAME", "DB_PASSWORD", k8sName + "-db"} {
						if strings.Contains(all, banned) {
							t.Errorf("db=none não deveria citar %s", banned)
						}
					}
				} else {
					d, ok := cat.Get(db)
					if !ok {
						t.Fatalf("banco %s fora do catálogo", db)
					}
					url := fmt.Sprintf("value: jdbc:%s://$(DB_HOST):%d/orders_api", d.JDBCScheme, d.Port)
					if !strings.Contains(dep, url) {
						t.Errorf("deployment sem %q:\n%s", url, dep)
					}
					if !strings.Contains(np, fmt.Sprintf("- port: %d", d.Port)) {
						t.Errorf("networkpolicy sem a porta %d do banco", d.Port)
					}
					for _, e := range k8sEnvs {
						if !strings.Contains(read(t, dest, "k8s/overlays/"+e+"/kustomization.yaml"), "DB_HOST="+k8sName+"-"+e) {
							t.Errorf("overlay %s sem DB_HOST próprio", e)
						}
					}
					if !strings.Contains(dep, "name: "+k8sName+"-db") {
						t.Error("deployment deveria ler usuário e senha do Secret do banco")
					}
				}

				if sel["auth"] == "none" {
					if strings.Contains(all, "JWT_ISSUER_URI") {
						t.Error("auth=none não deveria citar JWT_ISSUER_URI")
					}
				} else if !strings.Contains(base, "JWT_ISSUER_URI") {
					t.Error("auth ligada: faltou JWT_ISSUER_URI no ConfigMap")
				}

				redis := sel["cache"] == "redis"
				if got := strings.Contains(all, "REDIS_HOST"); got != redis {
					t.Errorf("REDIS_HOST presente=%v, esperado %v", got, redis)
				}
				if got := strings.Contains(np, "6379"); got != redis {
					t.Errorf("porta 6379 na NetworkPolicy presente=%v, esperado %v", got, redis)
				}
			})
		}
	}
}

func manifestOf(t *testing.T, tpl string) *template.Manifest {
	t.Helper()
	src, err := fs.Sub(templates.FS, tpl)
	if err != nil {
		t.Fatal(err)
	}
	m, err := engine.LoadManifest(src)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

// k8sCombos cruza as opções do manifesto que mexem nos manifests (db, auth, cache), fixando as demais
// (fixed) no padrão e ligando o Kubernetes.
func k8sCombos(m *template.Manifest, fixed ...string) []template.Selection {
	skip := map[string]bool{}
	for _, f := range fixed {
		skip[f] = true
	}
	var out []template.Selection
	for _, sel := range allSelections(m) {
		keep := true
		for name := range skip {
			if o, ok := m.Option(name); ok && sel[name] != o.Default {
				keep = false
			}
		}
		if keep {
			sel[k8sOpt] = k8sOn
			out = append(out, sel)
		}
	}
	return out
}

// Os nomes dos recursos, labels e namespaces precisam caber nos limites do Kubernetes (63 caracteres)
// mesmo com o nome de projeto mais longo que o forge aceita.
func TestKubernetesNamesFitLimits(t *testing.T) {
	long := strings.Repeat("a", names.MaxKebabLen-1) + "b"
	nameLine := regexp.MustCompile(`(?m)^\s*(?:name|namespace|app\.kubernetes\.io/\S+|environment):\s*(\S+)\s*$`)
	for tpl, sel := range k8sTemplates() {
		t.Run(tpl, func(t *testing.T) {
			dest := genK8s(t, tpl, long, sel)
			_ = filepath.WalkDir(filepath.Join(dest, "k8s"), func(p string, d os.DirEntry, err error) error {
				if err != nil || d.IsDir() {
					return err
				}
				b, _ := os.ReadFile(p)
				for _, m := range nameLine.FindAllStringSubmatch(string(b), -1) {
					if len(m[1]) > 63 {
						t.Errorf("%s: %q passa de 63 caracteres", p, m[1])
					}
				}
				return nil
			})
		})
	}
}

// Com o kubectl disponível, todo overlay precisa renderizar: pega patch quebrado e recurso faltando.
func TestKubernetesOverlaysRender(t *testing.T) {
	kubectl, err := exec.LookPath("kubectl")
	if err != nil {
		t.Skip("kubectl não está no PATH")
	}
	variants := map[string][]template.Selection{
		templates.SpringAPI:  {{k8sOpt: k8sOn}, {k8sOpt: k8sOn, "db": "none", "auth": "none", "cache": "none"}, {k8sOpt: k8sOn, "db": "mysql", "auth": "cognito", "cache": "redis"}},
		templates.QuarkusAPI: {{k8sOpt: k8sOn}, {k8sOpt: k8sOn, "db": "none", "auth": "none"}},
		templates.AngularApp: {{k8sOpt: k8sOn}},
		templates.ReactApp:   {{k8sOpt: k8sOn}},
	}
	for tpl, sels := range variants {
		for _, sel := range sels {
			t.Run(tpl+"/"+label(sel), func(t *testing.T) {
				dest := genK8s(t, tpl, k8sApp, sel)
				for _, e := range k8sEnvs {
					cmd := exec.Command(kubectl, "kustomize", filepath.Join("k8s", "overlays", e))
					cmd.Dir = dest
					out, err := cmd.CombinedOutput()
					if err != nil {
						t.Fatalf("kubectl kustomize %s falhou: %v\n%s", e, err, out)
					}
					text := string(out)
					for _, want := range []string{"kind: Deployment", "kind: Ingress", "kind: HorizontalPodAutoscaler", "kind: PodDisruptionBudget", "kind: NetworkPolicy", "namespace: " + k8sName + "-" + e} {
						if !strings.Contains(text, want) {
							t.Errorf("%s: a renderização não tem %q", e, want)
						}
					}
				}
			})
		}
	}
}
