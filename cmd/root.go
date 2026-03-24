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
	Long: `Notion API CLI for LLM tooling. All output is JSON.

Commands:
  notion search '<query>'
  notion page get <page-id>
  notion page create --parent <id> --title 'Title'
  notion page append <page-id> --markdown '<markdown>'
  notion page update <page-id> --title 'New Title'
  notion page update <page-id> --props '<json>'
  notion page property <page-id> <property-id>
  notion block delete <block-id>
  notion block update <block-id> --text 'Updated text'
  notion block update <block-id> --json '<json>'
  notion db get <db-id>
  notion db query <db-id> [--filter <json>] [--sort <json>] [--limit <n>]
  notion users
  notion comments get <page-id>
  notion comments add <page-id> --text 'Comment'
  notion comments reply <discussion-id> --text 'Reply'

Appending blocks (page append):
  The --markdown flag converts markdown to Notion blocks. Literal \n and \t
  in the string are converted to real newlines/tabs. Supported syntax:

    # text       -> heading_1
    ## text      -> heading_2
    ### text     -> heading_3
    - text       -> bulleted_list_item
    1. text      -> numbered_list_item
    - [ ] text   -> to_do (unchecked)
    - [x] text   -> to_do (checked)
    fenced code  -> code block (language tag preserved)
    > text       -> quote
    ---          -> divider
    GFM table    -> table block (see below)
    plain text   -> paragraph

  Tables use GitHub-Flavored Markdown syntax. The separator row (|---|---|)
  marks the first row as a header. Columns are inferred from the header width.

    | Name  | Role  |
    |-------|-------|
    | Alice | Admin |
    | Bob   | User  |

  Example:
    notion page append <page-id> --markdown '# Title\n- Item one\n- Item two\n---\nDone.'`,
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
