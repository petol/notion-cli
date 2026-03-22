package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/petol/notion-cli/internal/notion"
	"github.com/spf13/cobra"
)

var blockCmd = &cobra.Command{
	Use:   "block",
	Short: "Block operations",
}

func init() {
	rootCmd.AddCommand(blockCmd)

	blockCmd.AddCommand(blockDeleteCmd)

	blockCmd.AddCommand(blockUpdateCmd)
	blockUpdateCmd.Flags().String("text", "", "New plain text content (for text-based blocks)")
	blockUpdateCmd.Flags().String("json", "", "Full block content as JSON (type-specific field)")
}

var blockDeleteCmd = &cobra.Command{
	Use:   "delete <block-id>",
	Short: "Delete (trash) a block",
	Args:  cobra.ExactArgs(1),
	RunE:  runBlockDelete,
}

func runBlockDelete(_ *cobra.Command, args []string) error {
	c, err := notion.New()
	if err != nil {
		writeError(err)
		return err
	}

	if err := c.DeleteBlock(args[0]); err != nil {
		writeError(err)
		return err
	}

	outputJSON(map[string]any{"deleted": true, "id": args[0]})
	return nil
}

var blockUpdateCmd = &cobra.Command{
	Use:   "update <block-id>",
	Short: "Update a block's content",
	Long: `Update a block's content. Either --text (replaces rich_text for text-based blocks)
or --json (raw Notion type-specific content object) must be provided.

Example using --text:
  notion block update abc123 --text "Updated paragraph text"

Example using --json (for a to_do block):
  notion block update abc123 --json '{"rich_text":[{"type":"text","text":{"content":"Buy milk"}}],"checked":true}'`,
	Args: cobra.ExactArgs(1),
	RunE: runBlockUpdate,
}

func runBlockUpdate(cmd *cobra.Command, args []string) error {
	c, err := notion.New()
	if err != nil {
		writeError(err)
		return err
	}

	text, _ := cmd.Flags().GetString("text")
	text = unescapeNewlines(text)
	rawJSON, _ := cmd.Flags().GetString("json")

	if text == "" && rawJSON == "" {
		err := fmt.Errorf("one of --text or --json is required")
		writeError(err)
		return err
	}

	// Fetch the block to know its type
	block, err := c.GetBlock(args[0])
	if err != nil {
		writeError(err)
		return err
	}

	var body map[string]json.RawMessage

	if rawJSON != "" {
		// User supplied raw type-specific content
		var content json.RawMessage
		if err := json.Unmarshal([]byte(rawJSON), &content); err != nil {
			writeError(fmt.Errorf("invalid --json: %w", err))
			return err
		}
		body = map[string]json.RawMessage{block.Type: content}
	} else {
		// Build rich_text replacement from --text
		richText := []notion.RichText{
			{
				Type: "text",
				Text: &notion.TextContent{Content: text},
			},
		}
		rtJSON, _ := json.Marshal(map[string]any{"rich_text": richText})
		body = map[string]json.RawMessage{block.Type: rtJSON}
	}

	updated, err := c.UpdateBlock(args[0], body)
	if err != nil {
		writeError(err)
		return err
	}

	outputJSON(updated)
	return nil
}
