package download

import (
	"context"
	"fmt"

	"github.com/aquaproj/aqua/pkg/domain"
	"github.com/aquaproj/aqua/pkg/github"
	"github.com/sirupsen/logrus"
)

type GitHubContentFileDownloader struct {
	github GitHubContentAPI
	http   HTTPDownloader
}

type GitHubContentAPI interface {
	GetContents(ctx context.Context, repoOwner, repoName, path string, opt *github.RepositoryContentGetOptions) (*github.RepositoryContent, []*github.RepositoryContent, *github.Response, error)
}

func NewGitHubContentFileDownloader(gh GitHubContentAPI, httpDL HTTPDownloader) *GitHubContentFileDownloader {
	return &GitHubContentFileDownloader{
		github: gh,
		http:   httpDL,
	}
}

func (dl *GitHubContentFileDownloader) DownloadGitHubContentFile(ctx context.Context, logE *logrus.Entry, param *domain.GitHubContentFileParam) (*domain.GitHubContentFile, error) {
	// https://github.com/aquaproj/aqua/issues/391
	body, _, err := dl.http.Download(ctx, fmt.Sprintf(
		"https://raw.githubusercontent.com/%s/%s/%s/%s",
		param.RepoOwner, param.RepoName, param.Ref, param.Path,
	))
	if err == nil {
		return &domain.GitHubContentFile{
			ReadCloser: body,
		}, nil
	}
	if body != nil {
		body.Close()
	}

	file, _, _, err := dl.github.GetContents(ctx, param.RepoOwner, param.RepoName, param.Path, &github.RepositoryContentGetOptions{
		Ref: param.Ref,
	})
	if err != nil {
		return nil, fmt.Errorf("get a file by Get GitHub Content API: %w", err)
	}
	if file == nil {
		return nil, errGitHubContentMustBeFile
	}
	content, err := file.GetContent()
	if err != nil {
		return nil, fmt.Errorf("get a GitHub Content file content: %w", err)
	}
	return &domain.GitHubContentFile{
		String: content,
	}, nil
}
