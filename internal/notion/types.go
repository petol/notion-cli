package notion

import "encoding/json"

// RichText represents a Notion rich text object.
type RichText struct {
	Type        string       `json:"type"`
	PlainText   string       `json:"plain_text,omitempty"`
	Text        *TextContent `json:"text,omitempty"`
	Annotations *Annotations `json:"annotations,omitempty"`
	Href        *string      `json:"href,omitempty"`
}

type TextContent struct {
	Content string `json:"content"`
	Link    *Link  `json:"link,omitempty"`
}

type Annotations struct {
	Bold          bool `json:"bold"`
	Italic        bool `json:"italic"`
	Strikethrough bool `json:"strikethrough"`
	Code          bool `json:"code"`
}

type Link struct {
	URL string `json:"url"`
}

// Block represents a Notion block. The Type field determines which content
// field is populated. Using a flat struct avoids type assertions.
type Block struct {
	Object      string `json:"object,omitempty"`
	ID          string `json:"id,omitempty"`
	Type        string `json:"type"`
	HasChildren bool   `json:"has_children,omitempty"`

	Paragraph        *RichTextBlock  `json:"paragraph,omitempty"`
	Heading1         *RichTextBlock  `json:"heading_1,omitempty"`
	Heading2         *RichTextBlock  `json:"heading_2,omitempty"`
	Heading3         *RichTextBlock  `json:"heading_3,omitempty"`
	BulletedListItem *RichTextBlock  `json:"bulleted_list_item,omitempty"`
	NumberedListItem *RichTextBlock  `json:"numbered_list_item,omitempty"`
	ToDo             *ToDoContent    `json:"to_do,omitempty"`
	Code             *CodeContent    `json:"code,omitempty"`
	Quote            *RichTextBlock  `json:"quote,omitempty"`
	Callout          *RichTextBlock  `json:"callout,omitempty"`
	Toggle           *RichTextBlock  `json:"toggle,omitempty"`
	ChildPage        *ChildPageBlock `json:"child_page,omitempty"`
	Image            *FileBlock      `json:"image,omitempty"`
	// Divider has no content fields
}

type RichTextBlock struct {
	RichText []RichText `json:"rich_text"`
}

type ToDoContent struct {
	RichText []RichText `json:"rich_text"`
	Checked  bool       `json:"checked"`
}

type CodeContent struct {
	RichText []RichText `json:"rich_text"`
	Language string     `json:"language"`
}

type ChildPageBlock struct {
	Title string `json:"title"`
}

type FileBlock struct {
	Caption  []RichText    `json:"caption"`
	Type     string        `json:"type"`
	External *ExternalFile `json:"external,omitempty"`
	File     *HostedFile   `json:"file,omitempty"`
}

type ExternalFile struct {
	URL string `json:"url"`
}

type HostedFile struct {
	URL string `json:"url"`
}

// Page represents a Notion page.
type Page struct {
	Object         string                     `json:"object"`
	ID             string                     `json:"id"`
	URL            string                     `json:"url"`
	CreatedTime    string                     `json:"created_time"`
	LastEditedTime string                     `json:"last_edited_time"`
	Archived       bool                       `json:"archived"`
	Properties     map[string]json.RawMessage `json:"properties"`
}

// Database represents a Notion database.
type Database struct {
	Object         string                     `json:"object"`
	ID             string                     `json:"id"`
	URL            string                     `json:"url"`
	Title          []RichText                 `json:"title"`
	Properties     map[string]json.RawMessage `json:"properties"`
}

// Response wrappers

type BlockChildrenResponse struct {
	Results    []Block `json:"results"`
	HasMore    bool    `json:"has_more"`
	NextCursor string  `json:"next_cursor"`
}

type SearchResponse struct {
	Results    []json.RawMessage `json:"results"`
	HasMore    bool              `json:"has_more"`
	NextCursor string            `json:"next_cursor"`
}

type DatabaseQueryResponse struct {
	Results    []Page `json:"results"`
	HasMore    bool   `json:"has_more"`
	NextCursor string `json:"next_cursor"`
}

// User represents a Notion workspace user.
type User struct {
	Object    string      `json:"object"`
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	AvatarURL string      `json:"avatar_url"`
	Type      string      `json:"type"`
	Person    *UserPerson `json:"person,omitempty"`
	Bot       *UserBot    `json:"bot,omitempty"`
}

type UserPerson struct {
	Email string `json:"email"`
}

type UserBot struct {
	Owner     json.RawMessage `json:"owner,omitempty"`
	WorkspaceID string        `json:"workspace_id,omitempty"`
}

type UsersResponse struct {
	Results    []User `json:"results"`
	HasMore    bool   `json:"has_more"`
	NextCursor string `json:"next_cursor"`
}

// Comment represents a Notion comment.
type Comment struct {
	Object       string     `json:"object"`
	ID           string     `json:"id"`
	Parent       json.RawMessage `json:"parent"`
	DiscussionID string     `json:"discussion_id"`
	CreatedTime  string     `json:"created_time"`
	LastEditedTime string   `json:"last_edited_time"`
	CreatedBy    User       `json:"created_by"`
	RichText     []RichText `json:"rich_text"`
}

type CommentsResponse struct {
	Results    []Comment `json:"results"`
	HasMore    bool      `json:"has_more"`
	NextCursor string    `json:"next_cursor"`
}

// Request types

type Parent struct {
	Type       string `json:"type"`
	PageID     string `json:"page_id,omitempty"`
	DatabaseID string `json:"database_id,omitempty"`
}

type PageCreateRequest struct {
	Parent     Parent                     `json:"parent"`
	Properties map[string]json.RawMessage `json:"properties"`
}

type AppendBlocksRequest struct {
	Children []Block `json:"children"`
}

type SearchRequest struct {
	Query    string `json:"query"`
	PageSize int    `json:"page_size"`
}

type DatabaseQueryRequest struct {
	Filter      json.RawMessage `json:"filter,omitempty"`
	Sorts       json.RawMessage `json:"sorts,omitempty"`
	PageSize    int             `json:"page_size"`
	StartCursor string          `json:"start_cursor,omitempty"`
}

type PageUpdateRequest struct {
	Properties map[string]json.RawMessage `json:"properties,omitempty"`
	Archived   *bool                      `json:"archived,omitempty"`
	InTrash    *bool                      `json:"in_trash,omitempty"`
}

type CommentCreateRequest struct {
	Parent       json.RawMessage `json:"parent,omitempty"`
	DiscussionID string          `json:"discussion_id,omitempty"`
	RichText     []RichText      `json:"rich_text"`
}
