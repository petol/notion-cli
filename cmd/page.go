package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/petol/notion-cli/internal/convert"
	"github.com/petol/notion-cli/internal/notion"
	"github.com/petol/notion-cli/internal/render"
	"github.com/spf13/cobra"
)

// unescapeNewlines converts literal \n and \t sequences (as produced by LLMs
// passing multiline content through shell command strings) into real newlines/tabs.
func unescapeNewlines(s string) string {
	s = strings.ReplaceAll(s, `\n`, "\n")
	s = strings.ReplaceAll(s, `\t`, "\t")
	return s
}

var pageCmd = &cobra.Command{
	Use:   "page",
	Short: "Page operations",
}

func init() {
	rootCmd.AddCommand(pageCmd)

	pageCmd.AddCommand(pageGetCmd)

	pageCmd.AddCommand(pageCreateCmd)
	pageCreateCmd.Flags().String("parent", "", "Parent page or database ID (required)")
	pageCreateCmd.Flags().String("title", "", "Page title (required)")
	pageCreateCmd.Flags().String("parent-type", "page", "Parent type: page or database")
	_ = pageCreateCmd.MarkFlagRequired("parent")
	_ = pageCreateCmd.MarkFlagRequired("title")

	pageCmd.AddCommand(pageAppendCmd)
	pageAppendCmd.Flags().String("markdown", "", "Markdown content to append (required)")
	_ = pageAppendCmd.MarkFlagRequired("markdown")

	pageCmd.AddCommand(pageUpdateCmd)
	pageUpdateCmd.Flags().String("title", "", "New page title")
	pageUpdateCmd.Flags().String("props", "", "Properties to update as JSON object")
	pageUpdateCmd.Flags().Bool("archived", false, "Archive the page")
	pageUpdateCmd.Flags().Bool("restore", false, "Restore the page from archive/trash")

	pageCmd.AddCommand(pagePropertyCmd)
}

// pageGetCmd fetches page metadata and all block content rendered as markdown.
var pageGetCmd = &cobra.Command{
	Use:   "get <page-id>",
	Short: "Get page metadata and block content",
	Args:  cobra.ExactArgs(1),
	RunE:  runPageGet,
}

func runPageGet(_ *cobra.Command, args []string) error {
	c, err := notion.New()
	if err != nil {
		writeError(err)
		return err
	}

	pageID := args[0]

	page, err := c.GetPage(pageID)
	if err != nil {
		writeError(err)
		return err
	}

	blocks, err := fetchAllBlocks(c, pageID)
	if err != nil {
		writeError(err)
		return err
	}

	outputJSON(map[string]any{
		"id":               page.ID,
		"url":              page.URL,
		"created_time":     page.CreatedTime,
		"last_edited_time": page.LastEditedTime,
		"archived":         page.Archived,
		"properties":       page.Properties,
		"content":          render.BlocksToMarkdown(blocks),
	})
	return nil
}

// fetchAllBlocks paginates through all block children for a given block/page ID.
func fetchAllBlocks(c *notion.Client, id string) ([]notion.Block, error) {
	var all []notion.Block
	cursor := ""
	for len(all) < notion.MaxItems {
		resp, err := c.GetBlockChildren(id, cursor)
		if err != nil {
			return nil, err
		}
		all = append(all, resp.Results...)
		if !resp.HasMore {
			break
		}
		cursor = resp.NextCursor
	}
	return all, nil
}

// pageCreateCmd creates a new Notion page.
var pageCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new page",
	RunE:  runPageCreate,
}

func runPageCreate(cmd *cobra.Command, _ []string) error {
	c, err := notion.New()
	if err != nil {
		writeError(err)
		return err
	}

	parentID, _ := cmd.Flags().GetString("parent")
	title, _ := cmd.Flags().GetString("title")
	parentType, _ := cmd.Flags().GetString("parent-type")

	parent := notion.Parent{}
	switch parentType {
	case "database":
		parent.Type = "database_id"
		parent.DatabaseID = parentID
	default:
		parent.Type = "page_id"
		parent.PageID = parentID
	}

	titleProp, _ := json.Marshal(map[string]any{
		"title": []map[string]any{
			{"type": "text", "text": map[string]string{"content": title}},
		},
	})

	req := notion.PageCreateRequest{
		Parent: parent,
		Properties: map[string]json.RawMessage{
			"title": titleProp,
		},
	}

	page, err := c.CreatePage(req)
	if err != nil {
		writeError(err)
		return err
	}

	outputJSON(page)
	return nil
}

// pageUpdateCmd updates page properties and/or archived state.
var pageUpdateCmd = &cobra.Command{
	Use:   "update <page-id>",
	Short: "Update page properties or archived state",
	Args:  cobra.ExactArgs(1),
	RunE:  runPageUpdate,
}

func runPageUpdate(cmd *cobra.Command, args []string) error {
	c, err := notion.New()
	if err != nil {
		writeError(err)
		return err
	}

	req := notion.PageUpdateRequest{
		Properties: map[string]json.RawMessage{},
	}

	title, _ := cmd.Flags().GetString("title")
	if title != "" {
		titleProp, _ := json.Marshal(map[string]any{
			"title": []map[string]any{
				{"type": "text", "text": map[string]string{"content": title}},
			},
		})
		req.Properties["title"] = titleProp
	}

	propsJSON, _ := cmd.Flags().GetString("props")
	if propsJSON != "" {
		var extra map[string]json.RawMessage
		if err := json.Unmarshal([]byte(propsJSON), &extra); err != nil {
			writeError(fmt.Errorf("invalid --props JSON: %w", err))
			return err
		}
		for k, v := range extra {
			req.Properties[k] = v
		}
	}

	if len(req.Properties) == 0 {
		req.Properties = nil
	}

	archived, _ := cmd.Flags().GetBool("archived")
	restore, _ := cmd.Flags().GetBool("restore")
	if archived {
		t := true
		req.Archived = &t
	} else if restore {
		f := false
		req.Archived = &f
		req.InTrash = &f
	}

	page, err := c.UpdatePage(args[0], req)
	if err != nil {
		writeError(err)
		return err
	}

	outputJSON(page)
	return nil
}

// pagePropertyCmd retrieves a single page property item (supports paginated relation/rollup).
var pagePropertyCmd = &cobra.Command{
	Use:   "property <page-id> <property-id>",
	Short: "Get a single page property item",
	Args:  cobra.ExactArgs(2),
	RunE:  runPageProperty,
}

func runPageProperty(_ *cobra.Command, args []string) error {
	c, err := notion.New()
	if err != nil {
		writeError(err)
		return err
	}

	result, err := c.GetPageProperty(args[0], args[1])
	if err != nil {
		writeError(err)
		return err
	}

	outputJSON(result)
	return nil
}

// pageAppendCmd appends markdown content as blocks to an existing page.
var pageAppendCmd = &cobra.Command{
	Use:   "append <page-id>",
	Short: "Append markdown content as blocks to a page",
	Args:  cobra.ExactArgs(1),
	RunE:  runPageAppend,
}

func runPageAppend(cmd *cobra.Command, args []string) error {
	c, err := notion.New()
	if err != nil {
		writeError(err)
		return err
	}

	md, _ := cmd.Flags().GetString("markdown")
	md = unescapeNewlines(md)
	blocks := convert.MarkdownToBlocks(md)

	resp, err := c.AppendBlockChildren(args[0], blocks)
	if err != nil {
		writeError(err)
		return err
	}

	outputJSON(resp)
	return nil
}
