package notion

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	baseURL       = "https://api.notion.com/v1"
	notionVersion = "2022-06-28"
)

type Client struct {
	token string
	http  *http.Client
}

func New() (*Client, error) {
	token := os.Getenv("NOTION_API_KEY")
	if token == "" {
		return nil, errors.New("NOTION_API_KEY environment variable not set")
	}
	return &Client{
		token: token,
		http:  &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (c *Client) do(method, path string, body any, out any) error {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, baseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Notion-Version", notionVersion)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		// Try to extract Notion error message
		var notionErr struct {
			Message string `json:"message"`
			Code    string `json:"code"`
		}
		if jsonErr := json.Unmarshal(respBody, &notionErr); jsonErr == nil && notionErr.Message != "" {
			return fmt.Errorf("notion API error %d (%s): %s", resp.StatusCode, notionErr.Code, notionErr.Message)
		}
		return fmt.Errorf("notion API error %d: %s", resp.StatusCode, string(respBody))
	}

	if out != nil {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}

func (c *Client) Search(query string) (*SearchResponse, error) {
	var out SearchResponse
	err := c.do("POST", "/search", SearchRequest{Query: query, PageSize: 20}, &out)
	return &out, err
}

func (c *Client) GetPage(id string) (*Page, error) {
	var out Page
	err := c.do("GET", "/pages/"+id, nil, &out)
	return &out, err
}

func (c *Client) CreatePage(req PageCreateRequest) (*Page, error) {
	var out Page
	err := c.do("POST", "/pages", req, &out)
	return &out, err
}

func (c *Client) GetBlockChildren(id, cursor string) (*BlockChildrenResponse, error) {
	path := "/blocks/" + id + "/children?page_size=100"
	if cursor != "" {
		path += "&start_cursor=" + cursor
	}
	var out BlockChildrenResponse
	err := c.do("GET", path, nil, &out)
	return &out, err
}

func (c *Client) AppendBlockChildren(id string, blocks []Block) (*BlockChildrenResponse, error) {
	var out BlockChildrenResponse
	err := c.do("PATCH", "/blocks/"+id+"/children", AppendBlocksRequest{Children: blocks}, &out)
	return &out, err
}

func (c *Client) GetDatabase(id string) (*Database, error) {
	var out Database
	err := c.do("GET", "/databases/"+id, nil, &out)
	return &out, err
}

func (c *Client) QueryDatabase(id string, req DatabaseQueryRequest) (*DatabaseQueryResponse, error) {
	var out DatabaseQueryResponse
	err := c.do("POST", "/databases/"+id+"/query", req, &out)
	return &out, err
}

func (c *Client) UpdatePage(id string, req PageUpdateRequest) (*Page, error) {
	var out Page
	err := c.do("PATCH", "/pages/"+id, req, &out)
	return &out, err
}

func (c *Client) GetBlock(id string) (*Block, error) {
	var out Block
	err := c.do("GET", "/blocks/"+id, nil, &out)
	return &out, err
}

func (c *Client) UpdateBlock(id string, body any) (*Block, error) {
	var out Block
	err := c.do("PATCH", "/blocks/"+id, body, &out)
	return &out, err
}

func (c *Client) DeleteBlock(id string) error {
	return c.do("DELETE", "/blocks/"+id, nil, nil)
}

func (c *Client) ListUsers(cursor string) (*UsersResponse, error) {
	path := "/users?page_size=100"
	if cursor != "" {
		path += "&start_cursor=" + cursor
	}
	var out UsersResponse
	err := c.do("GET", path, nil, &out)
	return &out, err
}

func (c *Client) GetComments(blockID, cursor string) (*CommentsResponse, error) {
	path := "/comments?block_id=" + blockID + "&page_size=100"
	if cursor != "" {
		path += "&start_cursor=" + cursor
	}
	var out CommentsResponse
	err := c.do("GET", path, nil, &out)
	return &out, err
}

func (c *Client) CreateComment(req CommentCreateRequest) (*Comment, error) {
	var out Comment
	err := c.do("POST", "/comments", req, &out)
	return &out, err
}

func (c *Client) GetPageProperty(pageID, propertyID string) (json.RawMessage, error) {
	var out json.RawMessage
	err := c.do("GET", "/pages/"+pageID+"/properties/"+propertyID, nil, &out)
	return out, err
}
