package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/ikauedev/forge/internal/cli"
	"github.com/ikauedev/forge/internal/selfupdate"
)

func main() {
	// Apaga o executável antigo deixado por um `forge update` (no Windows ele não sai enquanto roda).
	if exe, err := os.Executable(); err == nil {
		selfupdate.Cleanup(exe)
	}
	if err := cli.NewRootCmd().Execute(); err != nil {
		var exit *cli.ExitError
		if errors.As(err, &exit) {
			if exit.Msg != "" {
				fmt.Fprintln(os.Stderr, "erro:", exit.Msg)
			}
			os.Exit(exit.Code)
		}
		fmt.Fprintln(os.Stderr, "erro:", err)
		os.Exit(1)
	}
}
