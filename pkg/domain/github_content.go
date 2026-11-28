package domain

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
)

type GitHubContentFileParam struct {
	RepoOwner string
	RepoName  string
	Ref       string
	Path      string
	Private   bool
	// Public says the repository is known to be public, so a file raw.githubusercontent.com
	// answers 404 for doesn't exist, and the GitHub API isn't asked again.
	//
	// Without it, any failure there falls back to the API, which is what reaches a file in a
	// repository that turns out to need a token. That fallback costs an API request, which
	// a client without a token can make sixty times an hour. A reader asking about a file
	// that is often missing on purpose -- aqua-registry-g2 asks for a version's
	// registry.json first and for the package's template when there is none -- would spend
	// one on every miss.
	Public bool
}

// ErrGitHubContentNotFound is a file a public repository doesn't have. It is returned only
// for a GitHubContentFileParam with Public set, since only then does a 404 settle it.
var ErrGitHubContentNotFound = errors.New("the file isn't found in the repository")

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
	DownloadGitHubContentFile(ctx context.Context, logger *slog.Logger, param *GitHubContentFileParam) (*GitHubContentFile, error)
}
