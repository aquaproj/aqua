// Package forgejo reads the releases a Forgejo instance publishes.
//
// Forgejo serves the API it inherited from Gitea at /api/v1, which GitHub's API was the
// model for without being the contract: a release says tag_name, published_at, draft and
// prerelease the same way, while a page is asked for with page and limit rather than
// per_page, and nothing says how many pages there are. That is why this is its own client
// rather than GitHub's pointed elsewhere.
package forgejo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

// MaxPerPage is the largest page Codeberg answers with: a larger limit is silently
// reduced to it, so asking for more only hides how many were left. An instance can be
// configured to answer with fewer, which is why a short page is not read as the last one.
const MaxPerPage = 50

// MaxPages is how many pages are read before giving up.
//
// The API says nothing about how many pages there are, so the end of the list is an empty
// page. An instance that answered the same page to every request would otherwise be read
// forever, and 50 pages of releases is already far more than a version is ever found in.
const MaxPages = 50

// Client reads one or more Forgejo instances. Which instance is read is an argument
// rather than state: a registry holds packages on several of them.
type Client struct {
	client *http.Client
}

// New returns a client that reads Forgejo instances over the given HTTP client.
func New(client *http.Client) *Client {
	return &Client{
		client: client,
	}
}

// Release is as much of a release as aqua reads. The instance says a good deal more.
type Release struct {
	TagName    string `json:"tag_name"`
	Name       string `json:"name"`
	Body       string `json:"body"`
	HTMLURL    string `json:"html_url"`
	Draft      bool   `json:"draft"`
	Prerelease bool   `json:"prerelease"`
}

// ListReleases returns one page of a repository's releases, newest first.
//
// Pages are 1-based, and an empty page is the end of the list: the API says nothing about
// how many pages there are, and a short page isn't the last one either, because an
// instance can be configured to answer with fewer items than were asked for.
//
// The instance is a host, which reaches one served at the root of that host over HTTPS. An
// instance under a sub-path, which Forgejo's ROOT_URL allows, would have to be named by a
// base URL instead; a field for that can be added beside this one when something needs it.
func (c *Client) ListReleases(ctx context.Context, host, owner, repo string, page, limit int) ([]*Release, error) {
	endpoint := fmt.Sprintf("https://%s/api/v1/repos/%s/%s/releases?page=%d&limit=%d",
		host, url.PathEscape(owner), url.PathEscape(repo), page, limit)
	b, err := c.doHTTPRequest(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("list releases: %w", slogerr.With(err,
			"api_endpoint", endpoint))
	}
	releases := []*Release{}
	if err := json.Unmarshal(b, &releases); err != nil {
		return nil, fmt.Errorf("decode the response body as JSON: %w", slogerr.With(err,
			"api_endpoint", endpoint))
	}
	return releases, nil
}

func (c *Client) doHTTPRequest(ctx context.Context, uri string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, fmt.Errorf("create a http request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send a http request: %w", err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read a response body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		// With what the instance said about it, which is where a rate limit, a
		// renamed repository and a private one are told apart.
		if message := said(b); message != "" {
			return nil, fmt.Errorf("unexpected status code: %d: %s", resp.StatusCode, message)
		}
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return b, nil
}

// said is the start of what an instance answered with, on one line, for an error to carry.
func said(b []byte) string {
	const maxLen = 200
	// By runes rather than bytes, so that a message in a language that doesn't write a
	// letter in one byte is neither cut in the middle of one nor sliced out of range.
	r := []rune(strings.Join(strings.Fields(string(b)), " "))
	if len(r) > maxLen {
		return string(r[:maxLen]) + "..."
	}
	return string(r)
}
