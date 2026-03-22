package cmd

import (
	"github.com/petol/notion-cli/internal/notion"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(searchCmd)
}

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search pages and databases",
	Args:  cobra.ExactArgs(1),
	RunE:  runSearch,
}

func runSearch(_ *cobra.Command, args []string) error {
	c, err := notion.New()
	if err != nil {
		writeError(err)
		return err
	}

	resp, err := c.Search(args[0])
	if err != nil {
		writeError(err)
		return err
	}

	outputJSON(resp)
	return nil
}
