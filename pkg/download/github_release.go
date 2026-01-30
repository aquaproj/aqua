package download

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/aquaproj/aqua/v2/pkg/domain"
	"github.com/aquaproj/aqua/v2/pkg/github"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

type GitHubReleaseDownloader struct {
	github GitHubReleaseAPI
	http   HTTPDownloader
}

type GitHubReleaseAPI interface {
	GetReleaseByTag(ctx context.Context, owner, repoName, version string) (*github.RepositoryRelease, *github.Response, error)
	DownloadReleaseAsset(ctx context.Context, owner, repoName string, assetID int64, httpClient *http.Client) (io.ReadCloser, string, error)
}

func NewGitHubReleaseDownloader(gh GitHubReleaseAPI, httpDL HTTPDownloader) *GitHubReleaseDownloader {
	return &GitHubReleaseDownloader{
		github: gh,
		http:   httpDL,
	}
}

func (dl *GitHubReleaseDownloader) DownloadGitHubRelease(ctx context.Context, logger *slog.Logger, param *domain.DownloadGitHubReleaseParam) (io.ReadCloser, int64, error) {
	if !param.Private {
		// I have tested if downloading assets from public repository's GitHub Releases anonymously is rate limited.
		// As a result of test, it seems not to be limited.
		// So at first aqua tries to download assets without GitHub API.
		// And if it failed, aqua tries again with GitHub API.
		// It avoids the rate limit of the access token.
		b, length, err := dl.http.Download(ctx, fmt.Sprintf(
			"https://github.com/%s/%s/releases/download/%s/%s",
			param.RepoOwner, param.RepoName, param.Version, param.Asset))
		if err == nil {
			return b, length, nil
		}
		if b != nil {
			b.Close()
		}
		slogerr.WithError(logger, err).Debug("failed to download an asset from GitHub Release without GitHub API. Try again with GitHub API",
			"repo_owner", param.RepoOwner,
			"repo_name", param.RepoName,
			"asset_version", param.Version,
			"asset_name", param.Asset)
	}

	release, _, err := dl.github.GetReleaseByTag(ctx, param.RepoOwner, param.RepoName, param.Version)
	if err != nil {
		return nil, 0, fmt.Errorf("get the GitHub Release by Tag: %w", err)
	}
	assetID, err := getAssetIDFromAssets(release.Assets, param.Asset)
	if err != nil {
		return nil, 0, err
	}
	body, redirectURL, err := dl.github.DownloadReleaseAsset(ctx, param.RepoOwner, param.RepoName, assetID, http.DefaultClient)
	if err != nil {
		return nil, 0, fmt.Errorf("download the release asset (asset id: %d): %w", assetID, err)
	}
	if body != nil {
		// DownloadReleaseAsset doesn't return a http.Response, so the content length is zero.
		return body, 0, nil
	}
	b, length, err := dl.http.Download(ctx, redirectURL)
	if err != nil {
		if b != nil {
			b.Close()
		}
		return nil, 0, fmt.Errorf("download asset from redirect URL: %w", err)
	}
	return b, length, nil
}

func getAssetIDFromAssets(assets []*github.ReleaseAsset, assetName string) (int64, error) {
	for _, asset := range assets {
		if asset.GetName() == assetName {
			return asset.GetID(), nil
		}
	}
	return 0, fmt.Errorf("the asset isn't found: %s", assetName)
}

// releaseAssetsImpl implements domain.ReleaseAssets
type releaseAssetsImpl struct {
	assets map[string]*domain.AssetDigest // assetName -> digest
}

func (r *releaseAssetsImpl) GetDigest(assetName string) *domain.AssetDigest {
	return r.assets[assetName]
}

// GetReleaseAssets retrieves all asset digests from a GitHub Release.
// Returns nil without error if the release is not found or has no digest-enabled assets.
func (dl *GitHubReleaseDownloader) GetReleaseAssets(ctx context.Context, logger *slog.Logger, owner, repo, version string) (domain.ReleaseAssets, error) {
	release, _, err := dl.github.GetReleaseByTag(ctx, owner, repo, version)
	if err != nil {
		return nil, fmt.Errorf("get the GitHub Release by Tag: %w", err)
	}
	assets := make(map[string]*domain.AssetDigest, len(release.Assets))
	for _, asset := range release.Assets {
		name := asset.GetName()
		digest := asset.GetDigest()
		if digest == "" {
			continue
		}
		if parsed := parseDigest(logger, digest); parsed != nil {
			assets[name] = parsed
		}
	}
	if len(assets) == 0 {
		return nil, nil //nolint:nilnil
	}
	return &releaseAssetsImpl{assets: assets}, nil
}

// parseDigest parses a digest string in the format "algorithm:hex".
// Currently only sha256 is supported.
// Returns nil if the format is unsupported.
func parseDigest(logger *slog.Logger, digest string) *domain.AssetDigest {
	algorithm, hex, found := strings.Cut(digest, ":")
	if !found {
		logger.Debug("unsupported digest format", "digest", digest)
		return nil
	}
	if algorithm != "sha256" {
		logger.Debug("unsupported digest algorithm", "algorithm", algorithm)
		return nil
	}
	return &domain.AssetDigest{
		Digest:    strings.ToUpper(hex),
		Algorithm: algorithm,
	}
}
