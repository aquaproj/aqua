package domain

import (
	"context"
	"io"
	"log/slog"
)

type DownloadGitHubReleaseParam struct {
	RepoOwner string
	RepoName  string
	Version   string
	Asset     string
	Private   bool
}

type GitHubReleaseDownloader interface {
	DownloadGitHubRelease(ctx context.Context, logger *slog.Logger, param *DownloadGitHubReleaseParam) (io.ReadCloser, int64, error)
}
