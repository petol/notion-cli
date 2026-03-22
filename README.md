# notion-cli

A minimal CLI for the Notion API, designed for use as an LLM tool. Lightweight alternative to the Notion MCP server.

- Single static binary (pure Go, no dynamic linking)
- JSON output on stdout, plain-text errors on stderr
- Auth via `NOTION_API_KEY` environment variable

## Install

```bash
CGO_ENABLED=0 go build -o notion .
```

## Usage

```bash
export NOTION_API_KEY=secret_...
```

### Search

```bash
notion search "meeting notes"
```

### Pages

```bash
# Get page content (metadata + blocks rendered as markdown)
notion page get <page-id>

# Create a page
notion page create --parent <page-id> --title "My Page"
notion page create --parent <db-id> --parent-type database --title "New Entry"

# Append markdown content
notion page append <page-id> --markdown "# Section\nSome **bold** text"

# Update properties
notion page update <page-id> --title "New Title"
notion page update <page-id> --props '{"Status": {"select": {"name": "Done"}}}'
notion page update <page-id> --archived
notion page update <page-id> --restore

# Get a specific property item (useful for large relation/rollup fields)
notion page property <page-id> <property-id>
```

### Blocks

```bash
# Delete a block
notion block delete <block-id>

# Update block text content
notion block update <block-id> --text "Updated text"

# Update block with raw Notion content JSON
notion block update <block-id> --json '{"rich_text":[{"type":"text","text":{"content":"Buy milk"}}],"checked":true}'
```

### Databases

```bash
# Get database schema
notion db get <db-id>

# Query all entries
notion db query <db-id>

# Query with filter and sort
notion db query <db-id> \
  --filter '{"property":"Status","select":{"equals":"In Progress"}}' \
  --sort '[{"property":"Created","direction":"descending"}]' \
  --limit 50
```

### Users

```bash
notion users
```

### Comments

```bash
# List comments on a page or block
notion comments get <page-id>

# Add a comment (starts a new discussion thread)
notion comments add <page-id> --text "Looks good!"

# Reply to an existing discussion thread
notion comments reply <discussion-id> --text "Thanks!"
```

## Output

All commands output JSON to stdout. Pipe to `jq` for filtering:

```bash
notion page get <id> | jq '.content'
notion db query <id> | jq '.results[].properties.Name'
notion users | jq '.results[] | {name, email: .person.email}'
```

Errors are written to stderr as plain text. Exit code is non-zero on failure.
