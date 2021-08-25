package controller

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func (ctrl *Controller) downloadFromGitHub(ctx context.Context, owner, repoName, version, assetName string) (io.ReadCloser, error) {
	release, _, err := ctrl.GitHub.Repositories.GetReleaseByTag(ctx, owner, repoName, version)
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
	body, redirectURL, err := ctrl.GitHub.Repositories.DownloadReleaseAsset(ctx, owner, repoName, assetID, http.DefaultClient)
	if err != nil {
		return nil, fmt.Errorf("download the release asset (asset id: %d): %w", assetID, err)
	}
	if body != nil {
		return body, nil
	}
	b, err := ctrl.downloadFromRedirectURL(ctx, redirectURL)
	if err != nil {
		if b != nil {
			b.Close()
		}
		return nil, err
	}
	return b, nil
}

func (ctrl *Controller) downloadFromRedirectURL(ctx context.Context, redirectURL string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, redirectURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create a HTTP request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send the HTTP request: %w", err)
	}
	if resp.StatusCode >= 300 { //nolint:gomnd
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}
	return resp.Body, nil
}
