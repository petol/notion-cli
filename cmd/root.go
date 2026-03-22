package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:          "notion",
	Short:        "Notion API CLI for LLM tooling",
	SilenceUsage: true,
}

func Execute() error {
	return rootCmd.Execute()
}

func outputJSON(v any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		writeError(err)
		os.Exit(1)
	}
}

func writeError(err error) {
	fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
}
