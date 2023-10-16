package download

import (
	"context"
	"fmt"
	"io"

	"github.com/aquaproj/aqua/v2/pkg/github"
)

func (dl *Downloader) getReadCloserFromGitHubArchive(ctx context.Context, file *File) (io.ReadCloser, int64, error) {
	if rc, length, err := dl.http.Download(ctx, fmt.Sprintf("https://github.com/%s/%s/archive/refs/tags/%s.tar.gz", file.RepoOwner, file.RepoName, file.Version)); err == nil {
		return rc, length, nil
	}
	// e.g. https://github.com/anqiansong/github-compare/archive/3972625c74bf6a5da00beb0e17e30e3e8d0c0950.zip
	if rc, length, err := dl.http.Download(ctx, fmt.Sprintf("https://github.com/%s/%s/archive/%s.tar.gz", file.RepoOwner, file.RepoName, file.Version)); err == nil {
		return rc, length, nil
	}
	u, _, err := dl.github.GetArchiveLink(ctx, file.RepoOwner, file.RepoName, github.Tarball, &github.RepositoryContentGetOptions{
		Ref: file.Version,
	}, 2)
	if err != nil {
		return nil, 0, fmt.Errorf("git an archive link with GitHub API: %w", err)
	}
	return dl.http.Download(ctx, u.String()) //nolint:wrapcheck
}
