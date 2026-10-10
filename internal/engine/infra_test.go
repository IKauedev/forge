package engine_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/ikauedev/forge/internal/template"
)

// Atribuição de literal a password/secret em arquivos de infra (referências como ${...} ou var.x passam).
var literalSecret = regexp.MustCompile(`(?i)\b[a-z_]*(password|secret)[a-z_]*\s*=\s*"[^"$]`)

func infraFile(dest, rel string) string {
	b, err := os.ReadFile(filepath.Join(dest, "infra", filepath.FromSlash(rel)))
	if err != nil {
		return ""
	}
	return string(b)
}

func infraExists(dest, rel string) bool {
	_, err := os.Stat(filepath.Join(dest, "infra", filepath.FromSlash(rel)))
	return err == nil
}

func checkInfra(t *testing.T, dest string, sel template.Selection) {
	t.Helper()
	db, vpc, auth, cache := sel["db"], sel["vpc"], sel["auth"], sel["cache"]

	for _, rel := range []string{
		"bootstrap/main.tf", "bootstrap/.terraform.lock.hcl", "envs/dev/main.tf", "envs/dev/.terraform.lock.hcl",
		"envs/prod/main.tf", "envs/prod/.terraform.lock.hcl", "modules/ecr/main.tf", "modules/ecs-service/main.tf",
		".tflint.hcl", ".checkov.yaml",
	} {
		if !infraExists(dest, rel) {
			t.Errorf("infra/%s ausente", rel)
		}
	}
	if infraExists(dest, "modules/database") != (db != "none") {
		t.Errorf("modules/database deveria existir só com db!=none (db=%s)", db)
	}
	if infraExists(dest, "modules/network") != (vpc == "new") {
		t.Errorf("modules/network deveria existir só com vpc=new (vpc=%s)", vpc)
	}

	// nenhuma senha em texto em arquivos de infra
	_ = filepath.WalkDir(filepath.Join(dest, "infra"), func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		switch filepath.Ext(p) {
		case ".tf", ".hcl", ".example", ".tfvars":
			b, _ := os.ReadFile(p)
			if m := literalSecret.FindString(string(b)); m != "" {
				t.Errorf("%s: possível segredo em texto: %q", p, m)
			}
		}
		return nil
	})

	for _, env := range []string{"dev", "prod"} {
		main := infraFile(dest, "envs/"+env+"/main.tf")
		versions := infraFile(dest, "envs/"+env+"/versions.tf")
		vars := infraFile(dest, "envs/"+env+"/variables.tf")

		if !strings.Contains(versions, "use_lockfile = true") || strings.Contains(versions, "dynamodb_table") {
			t.Errorf("envs/%s: backend deve usar use_lockfile e não DynamoDB", env)
		}
		if strings.Contains(main, `module "database"`) != (db != "none") {
			t.Errorf("envs/%s: módulo database inconsistente com db=%s", env, db)
		}
		if strings.Contains(main, `module "network"`) != (vpc == "new") || strings.Contains(vars, `variable "vpc_id"`) != (vpc == "existing") {
			t.Errorf("envs/%s: rede inconsistente com vpc=%s", env, vpc)
		}
		if strings.Contains(main, "JWT_ISSUER_URI") != (auth != "none") || strings.Contains(vars, "jwt_issuer_uri") != (auth != "none") {
			t.Errorf("envs/%s: JWT_ISSUER_URI inconsistente com auth=%s", env, auth)
		}
		if strings.Contains(main, "REDIS_HOST") != (cache == "redis") || strings.Contains(vars, "redis_host") != (cache == "redis") {
			t.Errorf("envs/%s: REDIS_HOST inconsistente com cache=%s", env, cache)
		}
		if db != "none" && (!strings.Contains(main, "DB_PASSWORD") || !strings.Contains(main, ":password::")) {
			t.Errorf("envs/%s: senha do banco deve vir do Secrets Manager via secrets", env)
		}
	}

	if db != "none" {
		if !strings.Contains(infraFile(dest, "modules/database/main.tf"), "manage_master_user_password = true") {
			t.Error("o banco deve usar manage_master_user_password = true")
		}
		dev := infraFile(dest, "envs/dev/main.tf")
		for _, want := range []string{`jdbc_scheme`, `engine_version`, `parameter_group_family`, `log_exports`} {
			if !strings.Contains(dev, want) {
				t.Errorf("envs/dev/main.tf sem %s do catálogo", want)
			}
		}
	}
	if !strings.Contains(infraFile(dest, "modules/ecs-service/main.tf"), "ignore_changes = [task_definition, desired_count]") {
		t.Error("o service deve ignorar task_definition e desired_count (o pipeline os altera)")
	}
	// prod: HTTPS obrigatório, proteção contra exclusão
	if strings.Contains(infraFile(dest, "envs/prod/variables.tf"), `default     = ""`) {
		t.Error("prod não deve ter certificate_arn opcional")
	}
	if db != "none" && !strings.Contains(infraFile(dest, "envs/prod/main.tf"), "deletion_protection     = true") {
		t.Error("prod deve ter deletion_protection")
	}
}

func TestInfraDatabaseSlotMatchesCatalog(t *testing.T) {
	cases := map[string][]string{
		"postgres": {`engine                 = "postgres"`, `engine_version         = "17"`, `parameter_group_family = "postgres17"`, `port                   = 5432`, `jdbc_scheme            = "postgresql"`, `["postgresql", "upgrade"]`},
		"mysql":    {`engine                 = "mysql"`, `engine_version         = "8.4"`, `parameter_group_family = "mysql8.4"`, `port                   = 3306`, `jdbc_scheme            = "mysql"`, `["error", "slowquery"]`},
		"mariadb":  {`engine                 = "mariadb"`, `engine_version         = "11.8"`, `parameter_group_family = "mariadb11.8"`, `jdbc_scheme            = "mariadb"`},
	}
	for db, wants := range cases {
		dest := gen(t, "db", db)
		for _, env := range []string{"dev", "prod"} {
			main := infraFile(dest, "envs/"+env+"/main.tf")
			for _, want := range wants {
				if !strings.Contains(main, want) {
					t.Errorf("%s/%s: main.tf deveria conter %q", db, env, want)
				}
			}
		}
	}
}

// terraform fmt -check em toda combinação: as diretivas não podem deixar HCL mal formatado.
func TestInfraIsFormattedInEveryCombination(t *testing.T) {
	bin, err := exec.LookPath("terraform")
	if err != nil {
		t.Skip("terraform não está no PATH")
	}
	for _, sel := range allSelections(manifest(t)) {
		t.Run(label(sel), func(t *testing.T) {
			dest := filepath.Join(t.TempDir(), "orders-api")
			generateWith(t, "Orders API", "com.empresa", dest, sel)
			cmd := exec.Command(bin, "fmt", "-check", "-recursive", "-diff", "-no-color")
			cmd.Dir = filepath.Join(dest, "infra")
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("terraform fmt reprovou:\n%s", out)
			}
		})
	}
}
