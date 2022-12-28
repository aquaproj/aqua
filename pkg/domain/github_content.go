package domain

import (
	"context"
	"io"

	"github.com/sirupsen/logrus"
)

type GitHubContentFileParam struct {
	RepoOwner string
	RepoName  string
	Ref       string
	Path      string
	Private   bool
}

type GitHubContentFile struct {
	ReadCloser io.ReadCloser
	String     string
}

type GitHubContentFileDownloader interface {
	DownloadGitHubContentFile(ctx context.Context, logE *logrus.Entry, param *GitHubContentFileParam) (*GitHubContentFile, error)
}
