package g2

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aquaproj/aqua/v2/pkg/github"
)

// TreeGetter reads a git tree from a GitHub repository.
type TreeGetter interface {
	GetTree(ctx context.Context, owner, repo, sha string, recursive bool) (*github.Tree, *github.Response, error)
}

// VersionLister lists the versions aqua-registry-g2 holds for a package.
type VersionLister struct {
	git       TreeGetter
	repoOwner string
	repoName  string
}

// NewVersionLister creates a VersionLister. An empty owner or name falls back to
// aqua-registry-g2.
func NewVersionLister(git TreeGetter, repoOwner, repoName string) *VersionLister {
	if repoOwner == "" {
		repoOwner = DefaultRepoOwner
	}
	if repoName == "" {
		repoName = DefaultRepoName
	}
	return &VersionLister{git: git, repoOwner: repoOwner, repoName: repoName}
}

// NewDefaultVersionLister creates a VersionLister reading aqua-registry-g2.
func NewDefaultVersionLister(git TreeGetter) *VersionLister {
	return NewVersionLister(git, "", "")
}

// List returns every version of a package that g2 has generated a registry.json for.
//
// These are the versions aqua can lock, so they are the only ones worth offering when
// choosing one. A version that exists upstream but not here can be written into
// aqua.yaml, but "aqua lock update" would then fail on it.
//
// The versions are returned in the order the tree holds them, which is by name and so
// not by version: "v10.0.0" sorts before "v2.0.0". Ordering is the caller's job,
// which it can do because this returns all of them.
func (l *VersionLister) List(ctx context.Context, logger *slog.Logger, pkgName string) ([]string, error) {
	// The Git Data API is used rather than the Contents API, which stops at 1,000
	// entries in a directory and says so only by returning fewer. Truncation there
	// would drop versions from the middle of an order that isn't the version order,
	// so the newest could be among the ones lost.
	tree, resp, err := l.git.GetTree(ctx, l.repoOwner, l.repoName, BranchName(pkgName)+":"+VersionDir, false)
	if err != nil {
		return nil, fmt.Errorf("get the versions directory of a package branch: %w", err)
	}
	if resp != nil {
		logger.Debug("GitHub API Rate Limit info",
			"github_api_rate_limit", resp.Rate.Limit,
			"github_api_rate_remaining", resp.Rate.Remaining)
	}
	// This limit is 100,000 entries, so reaching it means something other than a
	// package's version history. Returning a partial list would quietly hide
	// versions, so it fails instead.
	if tree.GetTruncated() {
		return nil, errTreeTruncated
	}

	versions := make([]string, 0, len(tree.Entries))
	for _, entry := range tree.Entries {
		// Each version is a directory holding its registry.json. Anything else is
		// not a version.
		if entry.GetType() != "tree" {
			continue
		}
		versions = append(versions, entry.GetPath())
	}
	return versions, nil
}
