package content

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aquaproj/aqua/pkg/download"
	"github.com/aquaproj/aqua/pkg/github"
	"github.com/sirupsen/logrus"
)

type client struct {
	github github.RepositoryService
	http   download.HTTPDownloader
}

type Client interface {
	Download(ctx context.Context, owner, repoName, version, filePath string, logE *logrus.Entry) (io.ReadCloser, error)
}

func New(gh github.RepositoryService, http download.HTTPDownloader) Client {
	return &client{
		github: gh,
		http:   http,
	}
}

func (downloader *client) Download(ctx context.Context, owner, repoName, version, filePath string, logE *logrus.Entry) (io.ReadCloser, error) {
	// https://github.com/aquaproj/aqua/issues/391
	body, err := downloader.http.Download(ctx, "https://raw.githubusercontent.com/"+owner+"/"+repoName+"/"+version+"/"+filePath)
	if err == nil {
		return body, nil
	}
	if body != nil {
		body.Close()
	}

	file, _, _, err := downloader.github.GetContents(ctx, owner, repoName, filePath, &github.RepositoryContentGetOptions{
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
