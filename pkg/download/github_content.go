package download

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/aquaproj/aqua/v2/pkg/domain"
	"github.com/aquaproj/aqua/v2/pkg/github"
	"github.com/sirupsen/logrus"
)

type GitHubContentFileDownloader struct {
	github GitHubContentAPI
	ghescr GHESContentAPIResolver
	http   HTTPDownloader
}

type GitHubContentAPI interface {
	DownloadContents(ctx context.Context, owner, repo, filepath string, opts *github.RepositoryContentGetOptions) (io.ReadCloser, *github.Response, error)
}

type GHESContentAPIResolver interface { //nolint:iface
	Resolve(ctx context.Context, logE *logrus.Entry, baseURL string) (github.GitHub, error)
}

func NewGitHubContentFileDownloader(gh GitHubContentAPI, ghescr GHESContentAPIResolver, httpDL HTTPDownloader) *GitHubContentFileDownloader {
	return &GitHubContentFileDownloader{
		github: gh,
		ghescr: ghescr,
		http:   httpDL,
	}
}

func (dl *GitHubContentFileDownloader) DownloadGitHubContentFile(ctx context.Context, _ *logrus.Entry, param *domain.GitHubContentFileParam) (*domain.GitHubContentFile, error) {
	if param.GHESBaseURL == "" && !param.Private {
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

	var contentAPI GitHubContentAPI
	if param.GHESBaseURL != "" {
		ghAPI, err := dl.ghescr.Resolve(ctx, nil, param.GHESBaseURL)
		if err != nil {
			return nil, fmt.Errorf("resolve GHES client: %w", err)
		}
		contentAPI = ghAPI
	} else {
		contentAPI = dl.github
	}

	file, resp, err := contentAPI.DownloadContents(ctx, param.RepoOwner, param.RepoName, param.Path, &github.RepositoryContentGetOptions{
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
