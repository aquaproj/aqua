package domain

import (
	"context"
	"io"
	"log/slog"
)

type DownloadGitHubReleaseParam struct {
	RepoOwner string
	RepoName  string
	Version   string
	Asset     string
	Private   bool
}

type GitHubReleaseDownloader interface {
	DownloadGitHubRelease(ctx context.Context, logger *slog.Logger, param *DownloadGitHubReleaseParam) (io.ReadCloser, int64, error)
}

// AssetDigest represents a digest retrieved from GitHub API Release Asset.
type AssetDigest struct {
	Digest    string // SHA256 hex string (uppercase)
	Algorithm string // Hash algorithm, e.g. "sha256"
}

// ReleaseAssets represents cached release assets from GitHub API.
// This allows retrieving digests for multiple assets without repeated API calls.
type ReleaseAssets interface {
	GetDigest(assetName string) *AssetDigest
}
