package g2

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"sync"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/domain"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

// DefaultRepoOwner and DefaultRepoName are where aqua-registry-g2 lives.
const (
	DefaultRepoOwner = "aquaproj"
	DefaultRepoName  = "aqua-registry-g2"
)

// DefaultBranch holds what the registry is as a whole -- the catalogue and the table of
// other names -- as opposed to a package's own branch, which holds that package.
const DefaultBranch = "main"

// Registry is the content of a package version's registry.json.
//
// Nothing in it is a template: every field is already resolved for one environment,
// which is what lets aqua install from it without evaluating anything.
type Registry struct {
	// PublishedAt is when the release this was generated from was published, as
	// RFC 3339. It is what the release says rather than when the file was written: a
	// reader deciding whether a version is old enough to install, or old enough to
	// stop offering, is asking about the release.
	//
	// It is empty for a version that has no release to ask -- a package whose versions
	// are tags -- and for a file generated before this was recorded.
	PublishedAt string `json:"published_at,omitempty"`

	Assets []*Asset `json:"assets"`
}

// Asset is everything needed to install the package on one environment.
type Asset struct {
	OS   string `json:"os"`
	Arch string `json:"arch"`
	// Variants tells apart entries that share an os and arch, such as a
	// linux/amd64 build for musl and one for glibc.
	Variants map[string]string `json:"variants,omitempty"`

	// LinkedLibc is the libc the executables in this asset are linked against:
	// "musl", "glibc", or "static" when they need none. It is read from the binary
	// rather than from the asset's name, and is absent when the asset holds nothing
	// that can be read that way, such as a script.
	//
	// Nothing selects on it today; Variants does that. It is recorded because which
	// build a machine should prefer is a choice worth leaving open: a statically
	// linked musl build runs anywhere, while a glibc one depends on the version of
	// glibc it was built against.
	LinkedLibc string `json:"linked_libc,omitempty"`

	Type      string `json:"type"`
	RepoOwner string `json:"repo_owner,omitempty"`
	RepoName  string `json:"repo_name,omitempty"`
	Asset     string `json:"asset,omitempty"`
	URL       string `json:"url,omitempty"`
	Format    string `json:"format,omitempty"`
	// Path is the Go module path of a go_install package. Crate and Cargo describe a
	// cargo one. A release says nothing about either, so ar2 copies them from
	// registry.yaml rather than inferring them.
	Path  string          `json:"path,omitempty"`
	Crate string          `json:"crate,omitempty"`
	Cargo *registry.Cargo `json:"cargo,omitempty"`

	Checksum          string `json:"checksum,omitempty"`
	ChecksumAlgorithm string `json:"checksum_algorithm,omitempty"`

	Files []*File `json:"files,omitempty"`

	// Private says the repository needs a token to download from. It is always
	// false in aqua-registry-g2, which mirrors the public registry, and is here so
	// that an entry and the file it is written from say the same things.
	Private bool `json:"private,omitempty"`

	// The signing configuration is aqua's own, copied through from aqua-registry
	// rather than restated, so that verification at install time reads the same
	// shape whether the package came from a lock file or a registry. The names of
	// the signature files are rendered for this environment like everything else
	// here: an entry that kept the templates would need the replacements kept with
	// it to expand them again.
	Cosign                     *registry.Cosign                     `json:"cosign,omitempty"`
	GitHubArtifactAttestations *registry.GitHubArtifactAttestations `json:"github_artifact_attestations,omitempty"`
	Minisign                   *registry.Minisign                   `json:"minisign,omitempty"`
	SLSAProvenance             *registry.SLSAProvenance             `json:"slsa_provenance,omitempty"`
}

// File is an executable inside the asset.
type File struct {
	Name string `json:"name"`
	Src  string `json:"src,omitempty"`
}

// Downloader fetches a file from a GitHub repository.
type Downloader interface {
	DownloadGitHubContentFile(ctx context.Context, logger *slog.Logger, param *domain.GitHubContentFileParam) (*domain.GitHubContentFile, error)
}

// Client reads aqua-registry-g2.
type Client struct {
	dl        Downloader
	cache     *Cache
	repoOwner string
	repoName  string
	// names is the table that resolves a name, read once. It is the registry's current
	// state rather than a fact about a version, so what is cached on disk is read
	// through its own age rather than kept for good the way registry.json is.
	names     *Names
	namesOnce sync.Once
	// catalogue is what the registry says it holds, read once, for the same reason.
	catalogue     *Index
	catalogueOnce sync.Once
}

// New creates a Client. An empty owner or name falls back to aqua-registry-g2, and a
// nil cache means every read goes to the network.
func New(dl Downloader, cache *Cache, repoOwner, repoName string) *Client {
	if repoOwner == "" {
		repoOwner = DefaultRepoOwner
	}
	if repoName == "" {
		repoName = DefaultRepoName
	}
	return &Client{dl: dl, cache: cache, repoOwner: repoOwner, repoName: repoName}
}

// Get returns the registry.json of one package version.
//
// A version that hasn't been generated yet has no file, so this fails rather than
// returning something empty: the caller decides whether to fall back to another
// registry, and an empty result would look like a package supporting no environment.
func (c *Client) Get(ctx context.Context, logger *slog.Logger, pkgName, version string) (*Registry, error) {
	cachePath := ""
	if c.cache != nil {
		// Keyed by the name that was asked for, so a file that has been read once is
		// found again without resolving anything.
		cachePath = c.cache.Path(c.repoOwner, c.repoName, pkgName, version)
		if b := c.cache.Read(cachePath); b != nil {
			if registry, err := parseRegistry(b); err == nil {
				logger.Debug("read registry.json from the cache", "cache_path", cachePath)
				return registry, nil
			}
			// A cached file that doesn't parse is a copy gone wrong, and the
			// original can be fetched again. Fetching replaces it.
			logger.Debug("the cached registry.json is broken", "cache_path", cachePath)
		}
	}

	b, err := c.download(ctx, logger, pkgName, version)
	if err != nil {
		return nil, err
	}
	registry, err := parseRegistry(b)
	if err != nil {
		return nil, err
	}

	// Only a file that parsed is cached, so a bad response isn't kept.
	if cachePath != "" {
		if err := c.cache.Write(cachePath, b); err != nil {
			// The fetch succeeded, so failing here would throw away a good
			// answer over a copy of it.
			slogerr.WithError(logger, err).Warn("cache registry.json", "cache_path", cachePath)
		}
	}
	return registry, nil
}

// BranchOf returns the branch holding the package.
//
// The branch is named after the package's id, so the name has to be resolved before anything
// can be asked for. The cached table answers first, which costs nothing; a name it doesn't
// know is a package that arrived after it was written -- or one this registry doesn't hold --
// and the registry's own table tells those apart and is cached so that the next run knows.
//
// A miss is the signal rather than an age, because a miss is what a stale table looks like
// from here: a name it knows is a name it knows, whenever it was written.
func (c *Client) BranchOf(ctx context.Context, logger *slog.Logger, pkgName string) (string, error) {
	if id, ok := c.cachedNames().ID(pkgName); ok {
		return IDBranchName(id), nil
	}
	if id, ok := c.fetchNames(ctx, logger).ID(pkgName); ok {
		return IDBranchName(id), nil
	}
	return "", fmt.Errorf("%w: %s", ErrNoPackageBranch, pkgName)
}

func (c *Client) download(ctx context.Context, logger *slog.Logger, pkgName, version string) ([]byte, error) {
	branch, err := c.BranchOf(ctx, logger, pkgName)
	if err != nil {
		return nil, err
	}
	file, err := c.dl.DownloadGitHubContentFile(ctx, logger, &domain.GitHubContentFileParam{
		RepoOwner: c.repoOwner,
		RepoName:  c.repoName,
		// Each package has its own branch, so the ref carries the package and the
		// path carries the version.
		Ref:  branch,
		Path: Path(version),
	})
	if err != nil {
		return nil, fmt.Errorf("download registry.json: %w", err)
	}
	defer file.Close()

	// The whole file is read rather than streamed into the decoder, because the
	// same bytes are what gets cached. registry.json describes one version of one
	// package, so it is small.
	b, err := io.ReadAll(file.Reader())
	if err != nil {
		return nil, fmt.Errorf("read registry.json: %w", err)
	}
	return b, nil
}

func parseRegistry(b []byte) (*Registry, error) {
	registry := &Registry{}
	if err := json.Unmarshal(b, registry); err != nil {
		return nil, fmt.Errorf("read registry.json as JSON: %w", err)
	}
	if len(registry.Assets) == 0 {
		return nil, errNoAsset
	}
	return registry, nil
}
