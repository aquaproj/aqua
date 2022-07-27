package download

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/aquaproj/aqua/pkg/github"
	"github.com/sirupsen/logrus"
)

func getAssetIDFromAssets(assets []*github.ReleaseAsset, assetName string) (int64, error) {
	for _, asset := range assets {
		if asset.GetName() == assetName {
			return asset.GetID(), nil
		}
	}
	return 0, fmt.Errorf("the asset isn't found: %s", assetName)
}

func (downloader *pkgDownloader) downloadFromGitHubRelease(ctx context.Context, owner, repoName, version, assetName string, logE *logrus.Entry) (io.ReadCloser, int64, error) {
	// I have tested if downloading assets from public repository's GitHub Releases anonymously is rate limited.
	// As a result of test, it seems not to be limited.
	// So at first aqua tries to download assets without GitHub API.
	// And if it failed, aqua tries again with GitHub API.
	// It avoids the rate limit of the access token.
	b, length, err := downloader.http.Download(ctx, "https://github.com/"+owner+"/"+repoName+"/releases/download/"+version+"/"+assetName)
	if err == nil {
		return b, length, nil
	}
	if b != nil {
		b.Close()
	}
	logE.WithError(err).WithFields(logrus.Fields{
		"repo_owner":    owner,
		"repo_name":     repoName,
		"asset_version": version,
		"asset_name":    assetName,
	}).Debug("failed to download an asset from GitHub Release without GitHub API. Try again with GitHub API")

	if downloader.github == nil {
		return nil, 0, errGitHubTokenIsRequired
	}

	release, _, err := downloader.github.GetReleaseByTag(ctx, owner, repoName, version)
	if err != nil {
		return nil, 0, fmt.Errorf("get the GitHub Release by Tag: %w", err)
	}
	assetID, err := getAssetIDFromAssets(release.Assets, assetName)
	if err != nil {
		return nil, 0, err
	}
	body, redirectURL, err := downloader.github.DownloadReleaseAsset(ctx, owner, repoName, assetID, http.DefaultClient)
	if err != nil {
		return nil, 0, fmt.Errorf("download the release asset (asset id: %d): %w", assetID, err)
	}
	if body != nil {
		// DownloadReleaseAsset doesn't return a http.Response, so the content length is zero.
		return body, 0, nil
	}
	b, length, err = downloader.http.Download(ctx, redirectURL)
	if err != nil {
		if b != nil {
			b.Close()
		}
		return nil, 0, fmt.Errorf("download asset from redirect URL: %w", err)
	}
	return b, length, nil
}
