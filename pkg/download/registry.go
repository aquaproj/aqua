package download

import (
	"context"
	"fmt"
	"io"
	"net/http"

	githubSvc "github.com/aquaproj/aqua/pkg/github"
	"github.com/google/go-github/v39/github"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

type registryDownloader struct {
	github githubSvc.RepositoryService
}

func (downloader *registryDownloader) GetGitHubContentFile(ctx context.Context, repoOwner, repoName, ref, path string, logE *logrus.Entry) ([]byte, error) {
	// https://github.com/aquaproj/aqua/issues/391
	body, err := fromURL(ctx, "https://raw.githubusercontent.com/"+repoOwner+"/"+repoName+"/"+ref+"/"+path, http.DefaultClient)
	if body != nil {
		defer body.Close()
	}
	if err == nil {
		b, err := io.ReadAll(body)
		if err == nil {
			return b, nil
		}
	}

	logerr.WithError(logE, err).WithFields(logrus.Fields{
		"repo_owner": repoOwner,
		"repo_name":  repoName,
		"ref":        ref,
		"path":       path,
	}).Debug("failed to download a content from GitHub without GitHub API. Try again with GitHub API")

	if downloader.github == nil {
		return nil, errGitHubTokenIsRequired
	}

	file, _, _, err := downloader.github.GetContents(ctx, repoOwner, repoName, path, &github.RepositoryContentGetOptions{
		Ref: ref,
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

	return []byte(content), nil
}
