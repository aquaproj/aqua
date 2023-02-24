package domain

import (
	"context"
	"fmt"
	"io"
	"strings"

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

func (file *GitHubContentFile) Reader() io.Reader {
	if file.String != "" {
		return strings.NewReader(file.String)
	}
	return file.ReadCloser
}

func (file *GitHubContentFile) Byte() ([]byte, error) {
	if file.String != "" {
		return []byte(file.String), nil
	}
	cnt, err := io.ReadAll(file.ReadCloser)
	if err != nil {
		return nil, fmt.Errorf("read the registry configuration file: %w", err)
	}
	return cnt, nil
}

func (file *GitHubContentFile) Close() error {
	if file.ReadCloser != nil {
		return file.ReadCloser.Close() //nolint:wrapcheck
	}
	return nil
}

type GitHubContentFileDownloader interface {
	DownloadGitHubContentFile(ctx context.Context, logE *logrus.Entry, param *GitHubContentFileParam) (*GitHubContentFile, error)
}

type MockGitHubContentFileDownloader struct {
	File *GitHubContentFile
	Err  error
}

func (mock *MockGitHubContentFileDownloader) DownloadGitHubContentFile(ctx context.Context, logE *logrus.Entry, param *GitHubContentFileParam) (*GitHubContentFile, error) {
	return mock.File, mock.Err
}
