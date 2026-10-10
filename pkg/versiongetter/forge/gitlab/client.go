// Package gitlab reads the releases a GitLab instance publishes.
//
// GitLab's API is its own rather than GitHub's with another name on it: it is /api/v4, a
// repository is one escaped project path rather than an owner and a name, a page is asked
// for with page and per_page, and a release says released_at and upcoming_release where
// the others say published_at, draft and prerelease. So this is a second client, and what
// it answers is the release every forge here is read as.
//
// Only the version listing needs the API. Downloading needs no call at all: an asset of a
// release is served at a permanent link built from the project path, the tag and the
// asset's name.
package gitlab

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/aquaproj/aqua/v2/pkg/versiongetter/forge"
)

// MaxPerPage is the largest page GitLab answers with.
const MaxPerPage = 100

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

// ListReleases returns one page of a project's releases, newest first.
//
// The project is <namespace>/<project>, which GitLab reads as one id: the separators
// inside it are escaped rather than kept, and a project inside subgroups is named the same
// way.
//
// Pages are 1-based and an empty page is the end of the list, the same way the other
// instances are read.
func (c *Client) ListReleases(ctx context.Context, host, project string, page, limit int) ([]*forge.Release, error) {
	endpoint := fmt.Sprintf("https://%s/api/v4/projects/%s/releases?page=%d&per_page=%d",
		host, url.PathEscape(project), page, limit)
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
	TagName     string `json:"tag_name"`
	Name        string `json:"name"`
	Description string `json:"description"`
	// UpcomingRelease says the release is dated in the future. GitLab has no draft, and
	// this is the same answer to the same question: what it names isn't one to install
	// from yet.
	UpcomingRelease bool `json:"upcoming_release"`
	Links           struct {
		Self string `json:"self"`
	} `json:"_links"`
}

func (r *release) release() *forge.Release {
	return &forge.Release{
		TagName: r.TagName,
		Name:    r.Name,
		Body:    r.Description,
		HTMLURL: r.Links.Self,
		Draft:   r.UpcomingRelease,
	}
}
