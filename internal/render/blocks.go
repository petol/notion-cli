package render

import (
	"fmt"
	"strings"

	"github.com/petol/notion-cli/internal/notion"
)

// BlocksToMarkdown converts a flat list of Notion blocks to a markdown string.
func BlocksToMarkdown(blocks []notion.Block) string {
	var sb strings.Builder
	for _, b := range blocks {
		renderBlock(&sb, b)
	}
	return sb.String()
}

func renderBlock(sb *strings.Builder, b notion.Block) {
	switch b.Type {
	case "paragraph":
		if b.Paragraph != nil {
			text := richTextToMarkdown(b.Paragraph.RichText)
			if text != "" {
				sb.WriteString(text + "\n\n")
			} else {
				sb.WriteString("\n")
			}
		}
	case "heading_1":
		if b.Heading1 != nil {
			sb.WriteString("# " + richTextToMarkdown(b.Heading1.RichText) + "\n\n")
		}
	case "heading_2":
		if b.Heading2 != nil {
			sb.WriteString("## " + richTextToMarkdown(b.Heading2.RichText) + "\n\n")
		}
	case "heading_3":
		if b.Heading3 != nil {
			sb.WriteString("### " + richTextToMarkdown(b.Heading3.RichText) + "\n\n")
		}
	case "bulleted_list_item":
		if b.BulletedListItem != nil {
			sb.WriteString("- " + richTextToMarkdown(b.BulletedListItem.RichText) + "\n")
		}
	case "numbered_list_item":
		if b.NumberedListItem != nil {
			sb.WriteString("1. " + richTextToMarkdown(b.NumberedListItem.RichText) + "\n")
		}
	case "to_do":
		if b.ToDo != nil {
			check := "[ ]"
			if b.ToDo.Checked {
				check = "[x]"
			}
			sb.WriteString("- " + check + " " + richTextToMarkdown(b.ToDo.RichText) + "\n")
		}
	case "code":
		if b.Code != nil {
			sb.WriteString("```" + b.Code.Language + "\n")
			sb.WriteString(richTextToMarkdown(b.Code.RichText) + "\n")
			sb.WriteString("```\n\n")
		}
	case "quote":
		if b.Quote != nil {
			for _, line := range strings.Split(richTextToMarkdown(b.Quote.RichText), "\n") {
				sb.WriteString("> " + line + "\n")
			}
			sb.WriteString("\n")
		}
	case "callout":
		if b.Callout != nil {
			sb.WriteString("> **Note:** " + richTextToMarkdown(b.Callout.RichText) + "\n\n")
		}
	case "toggle":
		if b.Toggle != nil {
			sb.WriteString(richTextToMarkdown(b.Toggle.RichText) + "\n\n")
		}
	case "divider":
		sb.WriteString("---\n\n")
	case "child_page":
		if b.ChildPage != nil {
			sb.WriteString(fmt.Sprintf("[%s](notion://page/%s)\n\n", b.ChildPage.Title, b.ID))
		}
	case "image":
		if b.Image != nil {
			caption := richTextToMarkdown(b.Image.Caption)
			url := ""
			if b.Image.External != nil {
				url = b.Image.External.URL
			} else if b.Image.File != nil {
				url = b.Image.File.URL
			}
			sb.WriteString(fmt.Sprintf("![%s](%s)\n\n", caption, url))
		}
	// Unknown block types are silently skipped — partial content is better than an error.
	}
}

// richTextToMarkdown concatenates rich text segments applying inline markdown formatting.
func richTextToMarkdown(rts []notion.RichText) string {
	var sb strings.Builder
	for _, rt := range rts {
		text := rt.PlainText
		if rt.Text != nil {
			text = rt.Text.Content
		}

		if rt.Annotations != nil {
			if rt.Annotations.Code {
				text = "`" + text + "`"
			}
			if rt.Annotations.Bold {
				text = "**" + text + "**"
			}
			if rt.Annotations.Italic {
				text = "_" + text + "_"
			}
			if rt.Annotations.Strikethrough {
				text = "~~" + text + "~~"
			}
		}

		if rt.Text != nil && rt.Text.Link != nil {
			text = "[" + text + "](" + rt.Text.Link.URL + ")"
		}

		sb.WriteString(text)
	}
	return sb.String()
}
