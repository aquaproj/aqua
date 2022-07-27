package download

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aquaproj/aqua/pkg/github"
)

func (downloader *PackageDownloader) downloadGitHubContent(ctx context.Context, owner, repoName, version, assetName string) (io.ReadCloser, error) {
	// https://github.com/aquaproj/aqua/issues/391
	body, _, err := downloader.http.Download(ctx, "https://raw.githubusercontent.com/"+owner+"/"+repoName+"/"+version+"/"+assetName)
	if err == nil {
		return body, nil
	}
	if body != nil {
		body.Close()
	}

	if downloader.github == nil {
		return nil, errGitHubTokenIsRequired
	}

	file, _, _, err := downloader.github.GetContents(ctx, owner, repoName, assetName, &github.RepositoryContentGetOptions{
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
