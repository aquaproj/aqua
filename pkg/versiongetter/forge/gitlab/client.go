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
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/versiongetter/forge"
)

// MaxPerPage is the largest page GitLab answers with.
const MaxPerPage = 100

// errNoRelease is returned when a project has published none.
var errNoRelease = errors.New("the project has no release")

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
		out = append(out, r.release(host))
	}
	return out, nil
}

// release is a release as this API answers with it.
type release struct {
	TagName     string `json:"tag_name"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Assets      struct {
		// Links are the files the release publishes. GitLab has sources beside
		// them -- the archives it makes of the tag -- which are not what a package
		// is installed from.
		Links []struct {
			Name string `json:"name"`
			// URL is where the file itself is, which is not always the instance:
			// a release link can point anywhere.
			URL string `json:"url"`
			// DirectAssetURL is where the instance serves the file. It is the
			// permanent release link when the link was created with a file path,
			// and wherever the file was uploaded when it wasn't.
			DirectAssetURL string `json:"direct_asset_url"`
		} `json:"links"`
	} `json:"assets"`
	// UpcomingRelease says the release is dated in the future. GitLab has no draft, and
	// this is the same answer to the same question: what it names isn't one to install
	// from yet.
	UpcomingRelease bool `json:"upcoming_release"`
	Links           struct {
		Self string `json:"self"`
	} `json:"_links"`
}

func (r *release) release(host string) *forge.Release {
	out := &forge.Release{
		TagName: r.TagName,
		Name:    r.Name,
		Body:    r.Description,
		HTMLURL: r.Links.Self,
		Draft:   r.UpcomingRelease,
		Assets:  make([]string, 0, len(r.Assets.Links)),
	}
	for _, link := range r.Assets.Links {
		if path, ok := served(host, link.URL, link.DirectAssetURL); ok {
			out.Assets = append(out.Assets, path)
		}
	}
	return out
}

// served is the path the instance serves the asset at, and whether it serves it at all.
//
// Two things have to be true, and the API answers both.
//
// The link has a permanent URL, /-/releases/<tag>/downloads/<path>, which GitLab builds
// from the file path the link was created with. That path is what the file is served at,
// and the API gives it only inside this URL: the name beside it is a label -- "package:
// RPM riscv64" is one -- and the two are the same string only where whoever published the
// release made them so.
//
// And the file is on the instance, which is where a link pointing anywhere else fails:
// gitlab-org/gitlab-runner publishes to S3, and asking the permanent link for one of
// those assets answers with the page GitLab shows before sending a reader to another
// site. A few hundred bytes of HTML is not a refusal a download notices, so such an asset
// is left out here. What a project publishing only those needs is an http package naming
// the file where it is.
func served(host, assetURL, directAssetURL string) (string, bool) {
	u, err := url.Parse(assetURL)
	if err != nil || u.Host != host {
		return "", false
	}
	_, afterRelease, found := strings.Cut(directAssetURL, "/-/releases/")
	if !found {
		return "", false
	}
	_, path, found := strings.Cut(afterRelease, "/downloads/")
	if !found || path == "" {
		return "", false
	}
	return path, true
}

// GetRelease returns one of a project's releases: the one the tag names, or the newest
// when the tag is empty.
//
// The newest is the first of the list rather than another endpoint's answer: GitLab orders
// releases by when they were released, and asking for one page of one is asking for that.
func (c *Client) GetRelease(ctx context.Context, host, project, tag string) (*forge.Release, error) {
	if tag == "" {
		releases, err := c.ListReleases(ctx, host, project, 1, 1)
		if err != nil {
			return nil, err
		}
		if len(releases) == 0 {
			return nil, errNoRelease
		}
		return releases[0], nil
	}
	endpoint := fmt.Sprintf("https://%s/api/v4/projects/%s/releases/%s",
		host, url.PathEscape(project), url.PathEscape(tag))
	out := &release{}
	if err := c.requester.GetJSON(ctx, endpoint, out); err != nil {
		return nil, fmt.Errorf("get the release: %w", err)
	}
	return out.release(host), nil
}

// GetDescription is what the project says it is, which is what a generated definition
// describes the package as.
func (c *Client) GetDescription(ctx context.Context, host, project string) (string, error) {
	endpoint := fmt.Sprintf("https://%s/api/v4/projects/%s", host, url.PathEscape(project))
	out := &struct {
		Description string `json:"description"`
	}{}
	if err := c.requester.GetJSON(ctx, endpoint, out); err != nil {
		return "", fmt.Errorf("get the project: %w", err)
	}
	return out.Description, nil
}
