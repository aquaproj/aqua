package download

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/aquaproj/aqua/v2/pkg/domain"
	"github.com/aquaproj/aqua/v2/pkg/github"
)

type GitHubContentFileDownloader struct {
	github GitHubContentAPI
	http   HTTPDownloader
}

type GitHubContentAPI interface {
	DownloadContents(ctx context.Context, owner, repo, filepath string, opts *github.RepositoryContentGetOptions) (io.ReadCloser, *github.Response, error)
}

func NewGitHubContentFileDownloader(gh GitHubContentAPI, httpDL HTTPDownloader) *GitHubContentFileDownloader {
	return &GitHubContentFileDownloader{
		github: gh,
		http:   httpDL,
	}
}

func (dl *GitHubContentFileDownloader) DownloadGitHubContentFile(ctx context.Context, _ *slog.Logger, param *domain.GitHubContentFileParam) (*domain.GitHubContentFile, error) {
	if !param.Private {
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
	}

	file, resp, err := dl.github.DownloadContents(ctx, param.RepoOwner, param.RepoName, param.Path, &github.RepositoryContentGetOptions{
		Ref: param.Ref,
	})
	if err != nil {
		return nil, fmt.Errorf("get a file by Get GitHub Content API: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		file.Close()
		return nil, fmt.Errorf("get a file by Get GitHub Content API: status code %d", resp.StatusCode)
	}
	return &domain.GitHubContentFile{
		ReadCloser: file,
	}, nil
}
