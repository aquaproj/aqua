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

func (f *GitHubContentFile) Reader() io.Reader {
	if f.String != "" {
		return strings.NewReader(f.String)
	}
	return f.ReadCloser
}

func (f *GitHubContentFile) Byte() ([]byte, error) {
	if f.String != "" {
		return []byte(f.String), nil
	}
	cnt, err := io.ReadAll(f.ReadCloser)
	if err != nil {
		return nil, fmt.Errorf("read the registry configuration file: %w", err)
	}
	return cnt, nil
}

func (f *GitHubContentFile) Close() error {
	if f.ReadCloser != nil {
		return f.ReadCloser.Close() //nolint:wrapcheck
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

func (m *MockGitHubContentFileDownloader) DownloadGitHubContentFile(ctx context.Context, logE *logrus.Entry, param *GitHubContentFileParam) (*GitHubContentFile, error) {
	return m.File, m.Err
}
