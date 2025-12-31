package domain

import (
	"context"
	"log/slog"
)

type MockGitHubContentFileDownloader struct {
	File *GitHubContentFile
	Err  error
}

func (m *MockGitHubContentFileDownloader) DownloadGitHubContentFile(ctx context.Context, logger *slog.Logger, param *GitHubContentFileParam) (*GitHubContentFile, error) {
	return m.File, m.Err
}
