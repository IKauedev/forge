package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/ikauedev/forge/internal/cli"
)

func main() {
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
