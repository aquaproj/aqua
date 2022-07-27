package download

import (
	"context"
	"fmt"
	"io"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/github"
)

func (downloader *pkgDownloader) getReadCloserFromGitHubArchive(ctx context.Context, pkg *config.Package) (io.ReadCloser, int64, error) {
	pkgInfo := pkg.PackageInfo
	if rc, length, err := downloader.http.Download(ctx, fmt.Sprintf("https://github.com/%s/%s/archive/refs/tags/%s.tar.gz", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Package.Version)); err == nil {
		return rc, length, nil
	}
	// e.g. https://github.com/anqiansong/github-compare/archive/3972625c74bf6a5da00beb0e17e30e3e8d0c0950.zip
	if rc, length, err := downloader.http.Download(ctx, fmt.Sprintf("https://github.com/%s/%s/archive/%s.tar.gz", pkgInfo.RepoOwner, pkgInfo.RepoName, pkg.Package.Version)); err == nil {
		return rc, length, nil
	}
	u, _, err := downloader.github.GetArchiveLink(ctx, pkgInfo.RepoOwner, pkgInfo.RepoName, github.Tarball, &github.RepositoryContentGetOptions{
		Ref: pkg.Package.Version,
	}, true)
	if err != nil {
		return nil, 0, fmt.Errorf("git an archive link with GitHub API: %w", err)
	}
	return downloader.http.Download(ctx, u.String()) //nolint:wrapcheck
}
