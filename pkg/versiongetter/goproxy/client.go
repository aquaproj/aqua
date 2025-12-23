package goproxy

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

type InfoPayload struct {
	Version string
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

func (c *Client) List(ctx context.Context, logE *logrus.Entry, path string) ([]string, error) {
	listEndpoint := fmt.Sprintf("https://proxy.golang.org/%s/@v/list", path)
	b, err := c.doHTTPRequest(ctx, listEndpoint)
	if err != nil {
		logerr.WithError(logE, err).WithField("api_endpoint", listEndpoint).Debug("retrieve package versions")
	} else {
		s := strings.TrimSpace(string(b))
		if s != "" {
			return strings.Split(s, "\n"), nil
		}
	}

	// Find the latest version (including pseudo-versions) if $module/@v/list is empty
	latestEndpoint := fmt.Sprintf("https://proxy.golang.org/%s/@latest", path)
	fields := logrus.Fields{
		"api_endpoint": latestEndpoint,
	}
	logE = logE.WithFields(fields)
	b, err = c.doHTTPRequest(ctx, latestEndpoint)
	if err != nil {
		return nil, fmt.Errorf("retrieve the latest version: %w", logerr.WithFields(err, fields))
	}
	if len(b) == 0 {
		logE.Debug("the response body from go proxy is empty")
		return nil, nil
	}
	payload := &InfoPayload{}
	if err := json.Unmarshal(b, &payload); err != nil {
		return nil, fmt.Errorf("decode the response body as JSON: %w", logerr.WithFields(err, fields))
	}
	return []string{payload.Version}, nil
}
