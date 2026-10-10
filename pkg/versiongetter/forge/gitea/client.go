// Package gitea reads the releases a Forgejo or Gitea instance publishes.
//
// One client reads both: Forgejo serves the API it inherited from Gitea at /api/v1, and
// GitHub's API was the model for that without being the contract. A release says tag_name,
// published_at, draft and prerelease the same way, while a page is asked for with page and
// limit rather than per_page, and nothing says how many pages there are. That is why this
// is its own client rather than GitHub's pointed elsewhere.
//
// The two package types are separate although the client is one, so that the day Forgejo
// and Gitea differ is a change here rather than in every definition.
package gitea

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/aquaproj/aqua/v2/pkg/versiongetter/forge"
)

// MaxPerPage is the largest page Codeberg answers with: a larger limit is silently
// reduced to it, so asking for more only hides how many were left. An instance can be
// configured to answer with fewer, which is why a short page is not read as the last one.
const MaxPerPage = 50

// Client reads one or more instances. Which instance is read is an argument rather than
// state: a registry holds packages on several of them.
type Client struct {
	requester *forge.Requester
}

// New returns a client that reads instances over the given HTTP client.
func New(client *http.Client) *Client {
	return &Client{requester: forge.NewRequester(client)}
}

// MaxPerPage is the largest page to ask this API for.
func (c *Client) MaxPerPage() int {
	return MaxPerPage
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
func (c *Client) ListReleases(ctx context.Context, host, owner, repo string, page, limit int) ([]*forge.Release, error) {
	endpoint := fmt.Sprintf("https://%s/api/v1/repos/%s/%s/releases?page=%d&limit=%d",
		host, url.PathEscape(owner), url.PathEscape(repo), page, limit)
	releases := []*release{}
	if err := c.requester.GetJSON(ctx, endpoint, &releases); err != nil {
		return nil, fmt.Errorf("list releases: %w", err)
	}
	out := make([]*forge.Release, 0, len(releases))
	for _, r := range releases {
		out = append(out, r.release())
	}
	return out, nil
}

// release is a release as this API answers with it.
type release struct {
	TagName    string `json:"tag_name"`
	Name       string `json:"name"`
	Body       string `json:"body"`
	HTMLURL    string `json:"html_url"`
	Draft      bool   `json:"draft"`
	Prerelease bool   `json:"prerelease"`
	Assets     []struct {
		Name string `json:"name"`
	} `json:"assets"`
}

func (r *release) release() *forge.Release {
	out := &forge.Release{
		TagName:    r.TagName,
		Name:       r.Name,
		Body:       r.Body,
		HTMLURL:    r.HTMLURL,
		Draft:      r.Draft,
		Prerelease: r.Prerelease,
		Assets:     make([]string, 0, len(r.Assets)),
	}
	for _, a := range r.Assets {
		out.Assets = append(out.Assets, a.Name)
	}
	return out
}
