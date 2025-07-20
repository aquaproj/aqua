package download

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/aquaproj/aqua/v2/pkg/domain"
	"github.com/aquaproj/aqua/v2/pkg/github"
	"github.com/sirupsen/logrus"
)

// userAgentTransport wraps an http.RoundTripper to add a browser-like User-Agent
type userAgentTransport struct {
	base http.RoundTripper
}

func (t *userAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Add a User-Agent to avoid being blocked by cloud storage services
	// This is a workaround until the origin issue is resolved.
	// See: https://github.com/aquaproj/aqua/pull/4019#issuecomment-3092666269
	req.Header.Set("User-Agent", github.GetUserAgent())
	r, err := t.base.RoundTrip(req)
	if err != nil {
		return nil, fmt.Errorf("round trip: %w", err)
	}
	return r, nil
}

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

func (dl *GitHubReleaseDownloader) DownloadGitHubRelease(ctx context.Context, logE *logrus.Entry, param *domain.DownloadGitHubReleaseParam) (io.ReadCloser, int64, error) {
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
		logE.WithError(err).WithFields(logrus.Fields{
			"repo_owner":    param.RepoOwner,
			"repo_name":     param.RepoName,
			"asset_version": param.Version,
			"asset_name":    param.Asset,
		}).Debug("failed to download an asset from GitHub Release without GitHub API. Try again with GitHub API")
	}

	release, _, err := dl.github.GetReleaseByTag(ctx, param.RepoOwner, param.RepoName, param.Version)
	if err != nil {
		return nil, 0, fmt.Errorf("get the GitHub Release by Tag: %w", err)
	}
	assetID, err := getAssetIDFromAssets(release.Assets, param.Asset)
	if err != nil {
		return nil, 0, err
	}
	// Create a custom client with browser-like User-Agent for Azure/S3 compatibility
	client := &http.Client{
		Transport: &userAgentTransport{
			base: http.DefaultTransport,
		},
	}
	body, redirectURL, err := dl.github.DownloadReleaseAsset(ctx, param.RepoOwner, param.RepoName, assetID, client)
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
