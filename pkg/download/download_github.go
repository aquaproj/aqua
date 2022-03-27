package download

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/aquaproj/aqua/pkg/log"
	"github.com/google/go-github/v39/github"
	"github.com/sirupsen/logrus"
)

type GitHubRepositoryService interface {
	GetContents(ctx context.Context, repoOwner, repoName, path string, opt *github.RepositoryContentGetOptions) (*github.RepositoryContent, []*github.RepositoryContent, *github.Response, error)
	GetReleaseByTag(ctx context.Context, owner, repoName, version string) (*github.RepositoryRelease, *github.Response, error)
	DownloadReleaseAsset(ctx context.Context, owner, repoName string, assetID int64, httpClient *http.Client) (io.ReadCloser, string, error)
}

func getAssetIDFromAssets(assets []*github.ReleaseAsset, assetName string) (int64, error) {
	for _, asset := range assets {
		if asset.GetName() == assetName {
			return asset.GetID(), nil
		}
	}
	return 0, fmt.Errorf("the asset isn't found: %s", assetName)
}

func (downloader *PkgDownloader) getlogE() *logrus.Entry {
	if downloader.LogE == nil {
		return log.New()
	}
	return downloader.LogE()
}

func (downloader *PkgDownloader) downloadFromGitHubRelease(ctx context.Context, owner, repoName, version, assetName string) (io.ReadCloser, error) {
	// I have tested if downloading assets from public repository's GitHub Releases anonymously is rate limited.
	// As a result of test, it seems not to be limited.
	// So at first aqua tries to download assets without GitHub API.
	// And if it failed, aqua tries again with GitHub API.
	// It avoids the rate limit of the access token.
	b, err := FromURL(ctx, "https://github.com/"+owner+"/"+repoName+"/releases/download/"+version+"/"+assetName, http.DefaultClient)
	if err == nil {
		return b, nil
	}
	if b != nil {
		b.Close()
	}
	downloader.getlogE().WithError(err).WithFields(logrus.Fields{
		"repo_owner":    owner,
		"repo_name":     repoName,
		"asset_version": version,
		"asset_name":    assetName,
	}).Debug("failed to download an asset from GitHub Release without GitHub API. Try again with GitHub API")

	if downloader.GitHubRepositoryService == nil {
		return nil, errGitHubTokenIsRequired
	}

	release, _, err := downloader.GitHubRepositoryService.GetReleaseByTag(ctx, owner, repoName, version)
	if err != nil {
		return nil, fmt.Errorf("get the GitHub Release by Tag: %w", err)
	}
	assetID, err := getAssetIDFromAssets(release.Assets, assetName)
	if err != nil {
		return nil, err
	}
	body, redirectURL, err := downloader.GitHubRepositoryService.DownloadReleaseAsset(ctx, owner, repoName, assetID, http.DefaultClient)
	if err != nil {
		return nil, fmt.Errorf("download the release asset (asset id: %d): %w", assetID, err)
	}
	if body != nil {
		return body, nil
	}
	b, err = FromURL(ctx, redirectURL, http.DefaultClient)
	if err != nil {
		if b != nil {
			b.Close()
		}
		return nil, err
	}
	return b, nil
}

func (downloader *PkgDownloader) downloadGitHubContent(ctx context.Context, owner, repoName, version, assetName string) (io.ReadCloser, error) {
	// https://github.com/aquaproj/aqua/issues/391
	body, err := FromURL(ctx, "https://raw.githubusercontent.com/"+owner+"/"+repoName+"/"+version+"/"+assetName, http.DefaultClient)
	if err == nil {
		return body, nil
	}
	if body != nil {
		body.Close()
	}

	if downloader.GitHubRepositoryService == nil {
		return nil, errGitHubTokenIsRequired
	}

	file, _, _, err := downloader.GitHubRepositoryService.GetContents(ctx, owner, repoName, assetName, &github.RepositoryContentGetOptions{
		Ref: version,
	})
	if err != nil {
		return nil, fmt.Errorf("get the registry configuration file by Get GitHub Content API: %w", err)
	}
	if file == nil {
		return nil, errGitHubContentMustBeFile
	}
	content, err := file.GetContent()
	if err != nil {
		return nil, fmt.Errorf("get the registry configuration content: %w", err)
	}
	return io.NopCloser(strings.NewReader(content)), nil
}
