package controller

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func (ctrl *Controller) downloadFromGitHub(ctx context.Context, owner, repoName, version, assetName string) (io.ReadCloser, error) {
	release, _, err := ctrl.GitHubRepositoryService.GetReleaseByTag(ctx, owner, repoName, version)
	if err != nil {
		return nil, fmt.Errorf("get the GitHub Release by Tag: %w", err)
	}
	var assetID int64
	for _, asset := range release.Assets {
		if asset.GetName() == assetName {
			assetID = asset.GetID()
			break
		}
	}
	if assetID == 0 {
		return nil, fmt.Errorf("the asset isn't found: %s", assetName)
	}
	body, redirectURL, err := ctrl.GitHubRepositoryService.DownloadReleaseAsset(ctx, owner, repoName, assetID, http.DefaultClient)
	if err != nil {
		return nil, fmt.Errorf("download the release asset (asset id: %d): %w", assetID, err)
	}
	if body != nil {
		return body, nil
	}
	b, err := ctrl.downloadFromURL(ctx, redirectURL)
	if err != nil {
		if b != nil {
			b.Close()
		}
		return nil, err
	}
	return b, nil
}
