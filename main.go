package main

import (
	"os"

	"github.com/josephlewis42/githubbeat/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
