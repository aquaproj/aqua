package controller

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/go-github/v39/github"
)

func getAssetIDFromAssets(assets []*github.ReleaseAsset, assetName string) (int64, error) {
	for _, asset := range assets {
		if asset.GetName() == assetName {
			return asset.GetID(), nil
		}
	}
	return 0, fmt.Errorf("the asset isn't found: %s", assetName)
}

func (downloader *pkgDownloader) downloadFromGitHubRelease(ctx context.Context, owner, repoName, version, assetName string) (io.ReadCloser, error) {
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
	b, err := downloader.downloadFromURL(ctx, redirectURL, http.DefaultClient)
	if err != nil {
		if b != nil {
			b.Close()
		}
		return nil, err
	}
	return b, nil
}

func (downloader *pkgDownloader) downloadGitHubContent(ctx context.Context, owner, repoName, version, assetName string) (io.ReadCloser, error) {
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
