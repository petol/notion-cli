package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/petol/notion-cli/internal/notion"
	"github.com/spf13/cobra"
)

var commentsCmd = &cobra.Command{
	Use:   "comments",
	Short: "Comment operations",
}

func init() {
	rootCmd.AddCommand(commentsCmd)

	commentsCmd.AddCommand(commentsGetCmd)

	commentsCmd.AddCommand(commentsAddCmd)
	commentsAddCmd.Flags().String("text", "", "Comment text (required)")
	_ = commentsAddCmd.MarkFlagRequired("text")

	commentsCmd.AddCommand(commentsReplyCmd)
	commentsReplyCmd.Flags().String("text", "", "Reply text (required)")
	_ = commentsReplyCmd.MarkFlagRequired("text")
}

// commentsGetCmd lists all comments on a page or block.
var commentsGetCmd = &cobra.Command{
	Use:   "get <page-or-block-id>",
	Short: "List comments on a page or block",
	Args:  cobra.ExactArgs(1),
	RunE:  runCommentsGet,
}

func runCommentsGet(_ *cobra.Command, args []string) error {
	c, err := notion.New()
	if err != nil {
		writeError(err)
		return err
	}

	var all []notion.Comment
	cursor := ""
	for len(all) < notion.MaxItems {
		resp, err := c.GetComments(args[0], cursor)
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

// commentsAddCmd adds a new comment (starts a new discussion thread) on a page.
var commentsAddCmd = &cobra.Command{
	Use:   "add <page-id>",
	Short: "Add a comment to a page (starts a new discussion thread)",
	Args:  cobra.ExactArgs(1),
	RunE:  runCommentsAdd,
}

func runCommentsAdd(cmd *cobra.Command, args []string) error {
	c, err := notion.New()
	if err != nil {
		writeError(err)
		return err
	}

	text, _ := cmd.Flags().GetString("text")

	parent, _ := json.Marshal(map[string]string{"type": "page_id", "page_id": args[0]})

	req := notion.CommentCreateRequest{
		Parent: parent,
		RichText: []notion.RichText{
			{
				Type:      "text",
				PlainText: text,
				Text:      &notion.TextContent{Content: text},
			},
		},
	}

	comment, err := c.CreateComment(req)
	if err != nil {
		writeError(err)
		return err
	}

	outputJSON(comment)
	return nil
}

// commentsReplyCmd replies to an existing discussion thread.
var commentsReplyCmd = &cobra.Command{
	Use:   "reply <discussion-id>",
	Short: "Reply to an existing discussion thread",
	Long: `Reply to an existing discussion thread by its discussion ID.
Use 'notion comments get <page-id>' to find discussion IDs.`,
	Args: cobra.ExactArgs(1),
	RunE: runCommentsReply,
}

func runCommentsReply(cmd *cobra.Command, args []string) error {
	c, err := notion.New()
	if err != nil {
		writeError(err)
		return err
	}

	text, _ := cmd.Flags().GetString("text")

	if args[0] == "" {
		err := fmt.Errorf("discussion-id is required")
		writeError(err)
		return err
	}

	req := notion.CommentCreateRequest{
		DiscussionID: args[0],
		RichText: []notion.RichText{
			{
				Type:      "text",
				PlainText: text,
				Text:      &notion.TextContent{Content: text},
			},
		},
	}

	comment, err := c.CreateComment(req)
	if err != nil {
		writeError(err)
		return err
	}

	outputJSON(comment)
	return nil
}
