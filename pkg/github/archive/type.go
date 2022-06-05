package archive

import (
	"context"
	"fmt"
	"io"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/download"
	"github.com/aquaproj/aqua/pkg/github"
)

type downloader struct {
	github github.ArchiveClient
	http   download.HTTPDownloader
}

type Downloader interface {
	Download(ctx context.Context, pkg *config.Package) (io.ReadCloser, error)
}

func New(gh github.ArchiveClient, httpDL download.HTTPDownloader) Downloader {
	return &downloader{
		github: gh,
		http:   httpDL,
	}
}

func (dl *downloader) Download(ctx context.Context, pkg *config.Package) (io.ReadCloser, error) {
	pkgInfo := pkg.PackageInfo
	pkgVersion := pkg.Package.Version
	if rc, err := dl.http.Download(ctx, fmt.Sprintf("https://github.com/%s/%s/archive/refs/tags/%s.tar.gz", pkgInfo.RepoOwner, pkgInfo.RepoName, pkgVersion)); err == nil {
		return rc, nil
	}
	// e.g. https://github.com/anqiansong/github-compare/archive/3972625c74bf6a5da00beb0e17e30e3e8d0c0950.zip
	if rc, err := dl.http.Download(ctx, fmt.Sprintf("https://github.com/%s/%s/archive/%s.tar.gz", pkgInfo.RepoOwner, pkgInfo.RepoName, pkgVersion)); err == nil {
		return rc, nil
	}
	u, _, err := dl.github.GetArchiveLink(ctx, pkgInfo.RepoOwner, pkgInfo.RepoName, github.Tarball, &github.RepositoryContentGetOptions{
		Ref: pkgVersion,
	}, true)
	if err != nil {
		return nil, fmt.Errorf("git an archive link with GitHub API: %w", err)
	}
	return dl.http.Download(ctx, u.String()) //nolint:wrapcheck
}
