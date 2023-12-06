package domain

import (
	"context"

	"github.com/sirupsen/logrus"
)

type MockGitHubContentFileDownloader struct {
	File *GitHubContentFile
	Err  error
}

func (m *MockGitHubContentFileDownloader) DownloadGitHubContentFile(ctx context.Context, logE *logrus.Entry, param *GitHubContentFileParam) (*GitHubContentFile, error) {
	return m.File, m.Err
}
