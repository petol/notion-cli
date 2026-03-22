package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/petol/notion-cli/internal/notion"
	"github.com/spf13/cobra"
)

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Database operations",
}

func init() {
	rootCmd.AddCommand(dbCmd)

	dbCmd.AddCommand(dbGetCmd)

	dbCmd.AddCommand(dbQueryCmd)
	dbQueryCmd.Flags().String("filter", "", "Filter JSON (Notion filter object)")
	dbQueryCmd.Flags().String("sort", "", "Sort JSON (array of Notion sort objects)")
	dbQueryCmd.Flags().Int("limit", 0, "Max results (0 = all)")
}

var dbGetCmd = &cobra.Command{
	Use:   "get <db-id>",
	Short: "Get database schema and properties",
	Args:  cobra.ExactArgs(1),
	RunE:  runDBGet,
}

func runDBGet(_ *cobra.Command, args []string) error {
	c, err := notion.New()
	if err != nil {
		writeError(err)
		return err
	}

	db, err := c.GetDatabase(args[0])
	if err != nil {
		writeError(err)
		return err
	}

	outputJSON(db)
	return nil
}

var dbQueryCmd = &cobra.Command{
	Use:   "query <db-id>",
	Short: "Query a database",
	Args:  cobra.ExactArgs(1),
	RunE:  runDBQuery,
}

func runDBQuery(cmd *cobra.Command, args []string) error {
	c, err := notion.New()
	if err != nil {
		writeError(err)
		return err
	}

	filterJSON, _ := cmd.Flags().GetString("filter")
	sortJSON, _ := cmd.Flags().GetString("sort")
	limit, _ := cmd.Flags().GetInt("limit")

	req := notion.DatabaseQueryRequest{PageSize: 100}

	if filterJSON != "" {
		if !json.Valid([]byte(filterJSON)) {
			err := fmt.Errorf("invalid --filter JSON")
			writeError(err)
			return err
		}
		req.Filter = json.RawMessage(filterJSON)
	}

	if sortJSON != "" {
		if !json.Valid([]byte(sortJSON)) {
			err := fmt.Errorf("invalid --sort JSON")
			writeError(err)
			return err
		}
		req.Sorts = json.RawMessage(sortJSON)
	}

	var allResults []notion.Page
	for {
		resp, err := c.QueryDatabase(args[0], req)
		if err != nil {
			writeError(err)
			return err
		}

		allResults = append(allResults, resp.Results...)

		if !resp.HasMore {
			break
		}
		if limit > 0 && len(allResults) >= limit {
			allResults = allResults[:limit]
			break
		}
		req.StartCursor = resp.NextCursor
	}

	outputJSON(map[string]any{
		"results": allResults,
		"count":   len(allResults),
	})
	return nil
}
