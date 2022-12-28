package domain

import (
	"context"
	"io"

	"github.com/sirupsen/logrus"
)

type DownloadGitHubReleaseParam struct {
	RepoOwner string
	RepoName  string
	Version   string
	Asset     string
	Private   bool
}

type GitHubReleaseDownloader interface {
	DownloadGitHubRelease(ctx context.Context, logE *logrus.Entry, param *DownloadGitHubReleaseParam) (io.ReadCloser, int64, error)
}
