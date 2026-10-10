package main

import (
	"fmt"
	"os"

	"github.com/ikauedeveloper/forge/internal/cli"
)

func main() {
	if err := cli.NewRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		os.Exit(1)
	}
}
