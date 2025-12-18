package goproxy

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type InfoPayload struct {
	Version string
	Time    time.Time
}

type Client struct {
	client *http.Client
}

func New(client *http.Client) *Client {
	return &Client{
		client: client,
	}
}

func (c *Client) doHTTPRequest(ctx context.Context, uri string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, fmt.Errorf("create a http request: %w", err)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send a http request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read a response body: %w", err)
	}
	return b, nil
}

func (c *Client) List(ctx context.Context, path string) ([]string, error) {
	b, err := c.doHTTPRequest(ctx, fmt.Sprintf("https://proxy.golang.org/%s/@v/list", path))
	if err != nil {
		return nil, err
	}
	s := strings.TrimSpace(string(b))
	if s != "" {
		return strings.Split(s, "\n"), nil
	}

	// Find the latest version (including pseudo-versions) if $module/@v/list is empty
	b, err = c.doHTTPRequest(ctx, fmt.Sprintf("https://proxy.golang.org/%s/@latest", path))
	if err != nil {
		return nil, err
	}
	if len(b) > 0 {
		payload := &InfoPayload{}
		if err := json.Unmarshal(b, &payload); err != nil {
			return nil, fmt.Errorf("decode the response body as JSON: %w", err)
		}
		return []string{payload.Version}, nil
	}

	return nil, nil
}
