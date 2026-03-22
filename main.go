package main

import (
	"os"

	"github.com/petol/notion-cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
