package versiongetter

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/fuzzyfinder"
	"github.com/aquaproj/aqua/v2/pkg/versiongetter/forge"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

// ForgeReleaseVersionGetter finds the versions of a package published as the releases of a
// repository on a forge instance.
//
// One getter for every such type, because what differs between the forges is how an
// instance is asked and not what the answer means: a client per API translates into the
// release this reads, and this picks the client the package's type is read by.
type ForgeReleaseVersionGetter struct {
	clients map[string]*forgeReader
}

// NewForgeRelease returns a version getter reading the given clients.
//
// A client may be nil, which is a getter that answers for the other forge and refuses the
// types the missing one reads. Nil as an interface, that is: a typed nil would be a client
// that is there and cannot be called.
func NewForgeRelease(gitea GiteaReleaseClient, gitlab GitLabReleaseClient) *ForgeReleaseVersionGetter {
	clients := map[string]*forgeReader{}
	if gitea != nil {
		reader := &forgeReader{list: gitea.ListReleases, maxPerPage: gitea.MaxPerPage()}
		clients[registry.PkgInfoTypeForgejoRelease] = reader
		clients[registry.PkgInfoTypeGiteaRelease] = reader
	}
	if gitlab != nil {
		clients[registry.PkgInfoTypeGitLabRelease] = &forgeReader{
			// The project is the owner and the name with a separator between them,
			// because that is the one id GitLab reads a project by.
			list: func(ctx context.Context, host, owner, repo string, page, limit int) ([]*forge.Release, error) {
				return gitlab.ListReleases(ctx, host, owner+"/"+repo, page, limit)
			},
			maxPerPage: gitlab.MaxPerPage(),
		}
	}
	return &ForgeReleaseVersionGetter{clients: clients}
}

// GiteaReleaseClient lists the releases of a repository on a Forgejo or Gitea instance,
// which that API names by an owner and a repository.
type GiteaReleaseClient interface {
	ListReleases(ctx context.Context, host, owner, repo string, page, limit int) ([]*forge.Release, error)
	MaxPerPage() int
}

// GitLabReleaseClient lists the releases of a project on a GitLab instance, which that API
// names by one project path: <namespace>/<project>, the namespace being where a project
// inside subgroups is.
type GitLabReleaseClient interface {
	ListReleases(ctx context.Context, host, project string, page, limit int) ([]*forge.Release, error)
	MaxPerPage() int
}

// forgeReader is one forge's client as the getter reads it: a page of releases by
// repository, and the page size that API caps at.
type forgeReader struct {
	list       func(ctx context.Context, host, owner, repo string, page, limit int) ([]*forge.Release, error)
	maxPerPage int
}

func convForgeRelease(release *forge.Release) *Release {
	v, prefix, _ := GetVersionAndPrefix(release.TagName)
	return &Release{
		Tag:           release.TagName,
		Version:       v,
		VersionPrefix: prefix,
		Prerelease:    release.Prerelease || (v != nil && v.Prerelease() != ""),
	}
}

// filterForgeRelease says whether a release is one to install from. A draft is left out
// as well as a prerelease: its assets are not served to anyone but the people who can
// publish it.
func filterForgeRelease(logger *slog.Logger, release *forge.Release, filters []*Filter) bool {
	if release.Draft {
		return false
	}
	return filterTagName(logger, release.TagName, release.Prerelease, filters)
}

// Get returns the newest version the filters allow.
func (g *ForgeReleaseVersionGetter) Get(ctx context.Context, logger *slog.Logger, pkg *registry.PackageInfo, filters []*Filter) (string, error) {
	client, err := g.client(pkg)
	if err != nil {
		return "", err
	}
	candidates := []*Release{}
	// Pages are 1-based, and an empty page is the end of the list. A short page is not:
	// an instance can be configured to answer with fewer items than were asked for, and
	// stopping there would hide every release after the first page.
	for page := 1; page <= forge.MaxPages; page++ {
		releases, err := client.list(ctx, pkg.GetHost(), pkg.RepoOwner, pkg.RepoName, page, client.maxPerPage)
		if err != nil {
			return "", fmt.Errorf("list releases: %w", err)
		}
		if len(releases) == 0 {
			return "", nil
		}
		for _, release := range releases {
			if filterForgeRelease(logger, release, filters) {
				candidates = append(candidates, convForgeRelease(release))
			}
		}
		if len(candidates) > 0 {
			return getLatestRelease(candidates).Tag, nil
		}
	}
	return "", nil
}

// List returns the versions the filters allow, newest first, for the version to be picked
// from.
func (g *ForgeReleaseVersionGetter) List(ctx context.Context, logger *slog.Logger, pkg *registry.PackageInfo, filters []*Filter, limit int) ([]*fuzzyfinder.Item, error) {
	client, err := g.client(pkg)
	if err != nil {
		return nil, err
	}
	perPage := forgeItemNumPerPage(limit, len(filters), client.maxPerPage)
	var items []*fuzzyfinder.Item
	tags := map[string]struct{}{}
	for page := 1; page <= forge.MaxPages; page++ {
		releases, err := client.list(ctx, pkg.GetHost(), pkg.RepoOwner, pkg.RepoName, page, perPage)
		if err != nil {
			return nil, fmt.Errorf("list releases: %w", err)
		}
		if len(releases) == 0 {
			return items, nil
		}
		for _, release := range releases {
			if _, ok := tags[release.TagName]; ok {
				continue
			}
			tags[release.TagName] = struct{}{}
			if !filterForgeRelease(logger, release, filters) {
				continue
			}
			v := &fuzzyfinder.Version{
				Name:        release.Name,
				Version:     release.TagName,
				Description: release.Body,
				URL:         release.HTMLURL,
			}
			items = append(items, &fuzzyfinder.Item{
				Item:    release.TagName,
				Preview: fuzzyfinder.PreviewVersion(v),
			})
		}
		if limit > 0 && len(items) >= limit { // Reach the limit
			return items[:limit], nil
		}
	}
	return items, nil
}

// client is what the package's releases are read by.
//
// An error rather than a nil client: a type nothing reads and an instance nothing names
// are both a question this can't answer, and a nil client would answer it by panicking.
func (g *ForgeReleaseVersionGetter) client(pkg *registry.PackageInfo) (*forgeReader, error) {
	if pkg.GetHost() == "" {
		return nil, errForgeHostRequired
	}
	client, ok := g.clients[pkg.Type]
	if !ok || client == nil {
		return nil, slogerr.With(errForgeClientMissing, "package_type", pkg.Type) //nolint:wrapcheck
	}
	return client, nil
}

// forgeItemNumPerPage asks for no more than the limit when every release asked for is one
// that can be shown, the same way itemNumPerPage does for GitHub. The cap is the API's
// rather than GitHub's, and each of them has its own.
func forgeItemNumPerPage(limit, filterNum, maxPerPage int) int {
	if limit > 0 && filterNum == 0 && maxPerPage > limit {
		return limit
	}
	return maxPerPage
}
