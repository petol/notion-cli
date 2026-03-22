package cmd

import (
	"github.com/petol/notion-cli/internal/notion"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(usersCmd)
}

var usersCmd = &cobra.Command{
	Use:   "users",
	Short: "List all workspace users",
	Args:  cobra.NoArgs,
	RunE:  runUsers,
}

func runUsers(_ *cobra.Command, _ []string) error {
	c, err := notion.New()
	if err != nil {
		writeError(err)
		return err
	}

	var all []notion.User
	cursor := ""
	for {
		resp, err := c.ListUsers(cursor)
		if err != nil {
			writeError(err)
			return err
		}
		all = append(all, resp.Results...)
		if !resp.HasMore {
			break
		}
		cursor = resp.NextCursor
	}

	outputJSON(map[string]any{
		"results": all,
		"count":   len(all),
	})
	return nil
}
