package g2

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/aquaproj/aqua/v2/pkg/config/registry"
	"github.com/aquaproj/aqua/v2/pkg/domain"
)

// DefaultRepoOwner and DefaultRepoName are where aqua-registry-g2 lives.
const (
	DefaultRepoOwner = "aquaproj"
	DefaultRepoName  = "aqua-registry-g2"
)

// Registry is the content of a package version's registry.json.
//
// Nothing in it is a template: every field is already resolved for one environment,
// which is what lets aqua install from it without evaluating anything.
type Registry struct {
	Assets []*Asset `json:"assets"`
}

// Asset is everything needed to install the package on one environment.
type Asset struct {
	OS   string `json:"os"`
	Arch string `json:"arch"`
	// Variants tells apart entries that share an os and arch, such as a
	// linux/amd64 build for musl and one for glibc.
	Variants map[string]string `json:"variants,omitempty"`

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

	// The signing configuration is aqua's own, copied through from aqua-registry
	// rather than restated, so that verification at install time reads the same
	// shape whether the package came from a lock file or a registry.
	Cosign                     *registry.Cosign                     `json:"cosign,omitempty"`
	GitHubArtifactAttestations *registry.GitHubArtifactAttestations `json:"github_artifact_attestations,omitempty"`
	Minisign                   *registry.Minisign                   `json:"minisign,omitempty"`
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
	repoOwner string
	repoName  string
}

// New creates a Client. An empty owner or name falls back to aqua-registry-g2.
func New(dl Downloader, repoOwner, repoName string) *Client {
	if repoOwner == "" {
		repoOwner = DefaultRepoOwner
	}
	if repoName == "" {
		repoName = DefaultRepoName
	}
	return &Client{dl: dl, repoOwner: repoOwner, repoName: repoName}
}

// Get returns the registry.json of one package version.
//
// A version that hasn't been generated yet has no file, so this fails rather than
// returning something empty: the caller decides whether to fall back to another
// registry, and an empty result would look like a package supporting no environment.
func (c *Client) Get(ctx context.Context, logger *slog.Logger, pkgName, version string) (*Registry, error) {
	file, err := c.dl.DownloadGitHubContentFile(ctx, logger, &domain.GitHubContentFileParam{
		RepoOwner: c.repoOwner,
		RepoName:  c.repoName,
		// Each package has its own branch, so the ref carries the package and the
		// path carries the version.
		Ref:  BranchName(pkgName),
		Path: Path(version),
	})
	if err != nil {
		return nil, fmt.Errorf("download registry.json: %w", err)
	}
	defer file.Close()

	registry := &Registry{}
	if err := json.NewDecoder(file.Reader()).Decode(registry); err != nil {
		return nil, fmt.Errorf("read registry.json as JSON: %w", err)
	}
	if len(registry.Assets) == 0 {
		return nil, errNoAsset
	}
	return registry, nil
}
