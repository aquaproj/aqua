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

// ForgeTagVersionGetter finds the versions of a package as the tags of a repository on a
// forge instance.
//
// It is what github_tag is to github.com: a project that tags its versions without
// publishing a release there -- Zig, whose binaries are on its own site -- has no release
// to read, and the package downloading them is an http package that says where its
// versions are with version_source and host.
//
// Which client reads the tags is the version source rather than the package's type, because
// an http package is on no forge of its own.
type ForgeTagVersionGetter struct {
	clients map[string]*forgeTagReader
}

// NewForgeTag returns a version getter reading the given clients. A client may be nil, the
// same way NewForgeRelease's may.
func NewForgeTag(gitea GiteaTagClient, gitlab GitLabTagClient) *ForgeTagVersionGetter {
	clients := map[string]*forgeTagReader{}
	if gitea != nil {
		reader := &forgeTagReader{list: gitea.ListTags, maxPerPage: gitea.MaxPerPage()}
		clients[registry.VersionSourceForgejoTag] = reader
		clients[registry.VersionSourceGiteaTag] = reader
	}
	if gitlab != nil {
		clients[registry.VersionSourceGitLabTag] = &forgeTagReader{
			list: func(ctx context.Context, host, owner, repo string, page, limit int) ([]*forge.Tag, error) {
				return gitlab.ListTags(ctx, host, owner+"/"+repo, page, limit)
			},
			maxPerPage: gitlab.MaxPerPage(),
		}
	}
	return &ForgeTagVersionGetter{clients: clients}
}

// GiteaTagClient lists the tags of a repository on a Forgejo or Gitea instance.
type GiteaTagClient interface {
	ListTags(ctx context.Context, host, owner, repo string, page, limit int) ([]*forge.Tag, error)
	MaxPerPage() int
}

// GitLabTagClient lists the tags of a project on a GitLab instance, named by its project
// path the way GitLabReleaseClient names it.
type GitLabTagClient interface {
	ListTags(ctx context.Context, host, project string, page, limit int) ([]*forge.Tag, error)
	MaxPerPage() int
}

// forgeTagReader is one forge's client as the getter reads it.
type forgeTagReader struct {
	list       func(ctx context.Context, host, owner, repo string, page, limit int) ([]*forge.Tag, error)
	maxPerPage int
}

func convForgeTag(tag *forge.Tag) *Release {
	v, prefix, _ := GetVersionAndPrefix(tag.Name)
	return &Release{
		Tag:           tag.Name,
		Version:       v,
		VersionPrefix: prefix,
		Prerelease:    v != nil && v.Prerelease() != "",
	}
}

// filterForgeTag says whether a tag is a version the filters allow. A tag says nothing
// about being a prerelease; only a release does.
func filterForgeTag(logger *slog.Logger, tag *forge.Tag, filters []*Filter) bool {
	return filterTagName(logger, tag.Name, false, filters)
}

// Get returns the newest version the filters allow.
func (g *ForgeTagVersionGetter) Get(ctx context.Context, logger *slog.Logger, pkg *registry.PackageInfo, filters []*Filter) (string, error) {
	client, err := g.client(pkg)
	if err != nil {
		return "", err
	}
	return latestForgeVersion(func(page int) ([]*forge.Tag, error) {
		tags, err := client.list(ctx, pkg.GetHost(), pkg.RepoOwner, pkg.RepoName, page, client.maxPerPage)
		if err != nil {
			return nil, fmt.Errorf("list tags: %w", err)
		}
		return tags, nil
	}, func(tag *forge.Tag) (*Release, bool) {
		if !filterForgeTag(logger, tag, filters) {
			return nil, false
		}
		return convForgeTag(tag), true
	})
}

// List returns the versions the filters allow, newest first, for the version to be picked
// from.
func (g *ForgeTagVersionGetter) List(ctx context.Context, logger *slog.Logger, pkg *registry.PackageInfo, filters []*Filter, limit int) ([]*fuzzyfinder.Item, error) {
	client, err := g.client(pkg)
	if err != nil {
		return nil, err
	}
	perPage := forgeItemNumPerPage(limit, len(filters), client.maxPerPage)
	var versions []string
	names := map[string]struct{}{}
	for page := 1; page <= forge.MaxPages; page++ {
		tags, err := client.list(ctx, pkg.GetHost(), pkg.RepoOwner, pkg.RepoName, page, perPage)
		if err != nil {
			return nil, fmt.Errorf("list tags: %w", err)
		}
		if len(tags) == 0 {
			break
		}
		for _, tag := range tags {
			if _, ok := names[tag.Name]; ok {
				continue
			}
			names[tag.Name] = struct{}{}
			if filterForgeTag(logger, tag, filters) {
				versions = append(versions, tag.Name)
			}
		}
		if limit > 0 && len(versions) >= limit { // Reach the limit
			return fuzzyfinder.ConvertStringsToItems(versions[:limit]), nil
		}
	}
	return fuzzyfinder.ConvertStringsToItems(versions), nil
}

// client is what the package's tags are read by.
func (g *ForgeTagVersionGetter) client(pkg *registry.PackageInfo) (*forgeTagReader, error) {
	if pkg.GetHost() == "" {
		return nil, errForgeTagHostRequired
	}
	client, ok := g.clients[pkg.VersionSource]
	if !ok || client == nil {
		return nil, slogerr.With(errForgeClientMissing, "version_source", pkg.VersionSource) //nolint:wrapcheck
	}
	return client, nil
}
