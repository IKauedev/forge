package cli

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/ikauedeveloper/forge/internal/engine"
	"github.com/ikauedeveloper/forge/internal/names"
	"github.com/ikauedeveloper/forge/templates"
)

type newFlags struct {
	pkg   string
	out   string
	noGit bool
}

func newNewCmd() *cobra.Command {
	var f newFlags
	cmd := &cobra.Command{
		Use:   "new <nome>",
		Short: "Cria um novo projeto Spring Boot",
		Example: `  forge new "Payments API"
  forge new payments-api --package com.empresa --out ~/projetos`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runNew(cmd, args[0], f)
		},
	}
	cmd.Flags().StringVar(&f.pkg, "package", "com.example", "pacote base Java")
	cmd.Flags().StringVar(&f.out, "out", ".", "diretório pai; o projeto é criado em <out>/<nome-kebab>")
	cmd.Flags().BoolVar(&f.noGit, "no-git", false, "não executar git init")
	return cmd
}

func runNew(cmd *cobra.Command, raw string, f newFlags) error {
	n, err := names.New(raw)
	if err != nil {
		return err
	}
	if err := names.ValidatePackage(f.pkg); err != nil {
		return err
	}
	src, err := fs.Sub(templates.FS, templates.SpringAPI)
	if err != nil {
		return err
	}
	dest := filepath.Join(f.out, n.Kebab)
	files, err := engine.Generate(src, engine.Options{Names: n, Package: f.pkg, Dest: dest})
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()
	fmt.Fprintf(w, "Projeto %q criado em %s (%d arquivos)\n", n.Title, dest, len(files))

	if !f.noGit {
		if err := gitInit(dest); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "aviso: git init falhou: %v\n", err)
		}
	}

	fmt.Fprintf(w, "\nPróximos passos:\n  cd %s\n  ./mvnw verify   # requer JDK 25 e Docker (Testcontainers)\n", dest)
	return nil
}

func gitInit(dir string) error {
	c := exec.Command("git", "init", "-q")
	c.Dir = dir
	c.Stderr = os.Stderr
	return c.Run()
}
