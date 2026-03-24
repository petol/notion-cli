package convert

import (
	"strings"

	"github.com/petol/notion-cli/internal/notion"
)

// MarkdownToBlocks converts a markdown string to a slice of Notion blocks.
// Supports: headings (#/##/###), bullet lists (-/*), numbered lists (1.),
// to-do items (- [ ] / - [x]), code fences (```), blockquotes (>),
// dividers (---), GFM tables (| col | col |), and paragraphs. Nested lists are flattened.
func MarkdownToBlocks(md string) []notion.Block {
	var blocks []notion.Block
	lines := strings.Split(strings.TrimSpace(md), "\n")

	i := 0
	for i < len(lines) {
		line := lines[i]

		// Table: collect all consecutive pipe-prefixed lines, check for separator row
		if strings.HasPrefix(strings.TrimSpace(line), "|") {
			var tableLines []string
			for i < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[i]), "|") {
				tableLines = append(tableLines, lines[i])
				i++
			}
			if len(tableLines) >= 2 && isTableSeparator(tableLines[1]) {
				blocks = append(blocks, makeTable(tableLines))
			} else {
				for _, tl := range tableLines {
					blocks = append(blocks, makeParagraph(tl))
				}
			}
			continue
		}

		// Code fence
		if strings.HasPrefix(line, "```") {
			lang := strings.TrimPrefix(line, "```")
			var codeLines []string
			i++
			for i < len(lines) && !strings.HasPrefix(lines[i], "```") {
				codeLines = append(codeLines, lines[i])
				i++
			}
			blocks = append(blocks, makeCodeBlock(strings.Join(codeLines, "\n"), lang))
			i++ // skip closing ```
			continue
		}

		// Headings
		if strings.HasPrefix(line, "### ") {
			blocks = append(blocks, makeHeading3(strings.TrimPrefix(line, "### ")))
		} else if strings.HasPrefix(line, "## ") {
			blocks = append(blocks, makeHeading2(strings.TrimPrefix(line, "## ")))
		} else if strings.HasPrefix(line, "# ") {
			blocks = append(blocks, makeHeading1(strings.TrimPrefix(line, "# ")))

		// Blockquote
		} else if strings.HasPrefix(line, "> ") {
			blocks = append(blocks, makeQuote(strings.TrimPrefix(line, "> ")))

		// Bullet or to-do
		} else if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			text := line[2:]
			if strings.HasPrefix(text, "[ ] ") {
				blocks = append(blocks, makeToDo(text[4:], false))
			} else if strings.HasPrefix(text, "[x] ") || strings.HasPrefix(text, "[X] ") {
				blocks = append(blocks, makeToDo(text[4:], true))
			} else {
				blocks = append(blocks, makeBullet(text))
			}

		// Numbered list (detect "N. " prefix where N is a digit)
		} else if len(line) >= 3 && line[0] >= '0' && line[0] <= '9' && line[1] == '.' && line[2] == ' ' {
			blocks = append(blocks, makeNumbered(line[3:]))

		// Divider
		} else if line == "---" || line == "***" || line == "___" {
			blocks = append(blocks, makeDivider())

		// Empty line: skip
		} else if strings.TrimSpace(line) == "" {
			i++
			continue

		// Paragraph
		} else {
			blocks = append(blocks, makeParagraph(line))
		}

		i++
	}
	return blocks
}

func makeRichText(content string) []notion.RichText {
	return []notion.RichText{
		{
			Type: "text",
			Text: &notion.TextContent{Content: content},
		},
	}
}

func makeParagraph(text string) notion.Block {
	return notion.Block{
		Object:    "block",
		Type:      "paragraph",
		Paragraph: &notion.RichTextBlock{RichText: makeRichText(text)},
	}
}

func makeHeading1(text string) notion.Block {
	return notion.Block{
		Object:   "block",
		Type:     "heading_1",
		Heading1: &notion.RichTextBlock{RichText: makeRichText(text)},
	}
}

func makeHeading2(text string) notion.Block {
	return notion.Block{
		Object:   "block",
		Type:     "heading_2",
		Heading2: &notion.RichTextBlock{RichText: makeRichText(text)},
	}
}

func makeHeading3(text string) notion.Block {
	return notion.Block{
		Object:   "block",
		Type:     "heading_3",
		Heading3: &notion.RichTextBlock{RichText: makeRichText(text)},
	}
}

func makeBullet(text string) notion.Block {
	return notion.Block{
		Object:           "block",
		Type:             "bulleted_list_item",
		BulletedListItem: &notion.RichTextBlock{RichText: makeRichText(text)},
	}
}

func makeNumbered(text string) notion.Block {
	return notion.Block{
		Object:           "block",
		Type:             "numbered_list_item",
		NumberedListItem: &notion.RichTextBlock{RichText: makeRichText(text)},
	}
}

func makeToDo(text string, checked bool) notion.Block {
	return notion.Block{
		Object: "block",
		Type:   "to_do",
		ToDo:   &notion.ToDoContent{RichText: makeRichText(text), Checked: checked},
	}
}

func makeCodeBlock(text, lang string) notion.Block {
	return notion.Block{
		Object: "block",
		Type:   "code",
		Code:   &notion.CodeContent{RichText: makeRichText(text), Language: lang},
	}
}

func makeQuote(text string) notion.Block {
	return notion.Block{
		Object: "block",
		Type:   "quote",
		Quote:  &notion.RichTextBlock{RichText: makeRichText(text)},
	}
}

func makeDivider() notion.Block {
	return notion.Block{
		Object: "block",
		Type:   "divider",
	}
}

// isTableSeparator returns true for lines like |---|:---|:---:| that separate
// a table header from its body.
func isTableSeparator(line string) bool {
	trimmed := strings.TrimSpace(line)
	if !strings.ContainsRune(trimmed, '-') {
		return false
	}
	for _, c := range trimmed {
		if c != '|' && c != '-' && c != ':' && c != ' ' {
			return false
		}
	}
	return true
}

// parseTableCells splits a markdown table row like "| a | b | c |" into cells.
func parseTableCells(line string) [][]notion.RichText {
	parts := strings.Split(line, "|")
	var cells [][]notion.RichText
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		cells = append(cells, makeRichText(p))
	}
	return cells
}

// makeTable builds a Notion table block from collected markdown table lines.
// lines[0] is the header row, lines[1] is the separator, lines[2:] are data rows.
func makeTable(lines []string) notion.Block {
	headerCells := parseTableCells(lines[0])
	width := len(headerCells)
	if width == 0 {
		width = 1
	}

	padCells := func(cells [][]notion.RichText) [][]notion.RichText {
		for len(cells) < width {
			cells = append(cells, makeRichText(""))
		}
		return cells[:width]
	}

	rows := []notion.Block{{
		Object:   "block",
		Type:     "table_row",
		TableRow: &notion.TableRowContent{Cells: padCells(headerCells)},
	}}
	for _, line := range lines[2:] {
		cells := padCells(parseTableCells(line))
		rows = append(rows, notion.Block{
			Object:   "block",
			Type:     "table_row",
			TableRow: &notion.TableRowContent{Cells: cells},
		})
	}

	return notion.Block{
		Object: "block",
		Type:   "table",
		Table: &notion.TableContent{
			TableWidth:      width,
			HasColumnHeader: true,
			Children:        rows,
		},
	}
}
