package pkgtype

import (
	"context"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/pkgtype/githubarchive"
	"github.com/aquaproj/aqua/pkg/pkgtype/githubcontent"
	"github.com/aquaproj/aqua/pkg/pkgtype/githubrelease"
	"github.com/aquaproj/aqua/pkg/pkgtype/gobuild"
	"github.com/aquaproj/aqua/pkg/pkgtype/goinstall"
	"github.com/aquaproj/aqua/pkg/pkgtype/http"
	"github.com/sirupsen/logrus"
)

type Package interface {
	Install(ctx context.Context, pkg *config.Package, pkgInfo *config.PackageInfo, logE *logrus.Entry) error
	CheckInstalled(pkg *config.Package, pkgInfo *config.PackageInfo) (bool, error)
	GetFiles(pkgInfo *config.PackageInfo) []*config.File
	Find(pkg *config.Package, pkgInfo *config.PackageInfo, exeName string, logE *logrus.Entry) (string, error)
	GetFilePath(pkg *config.Package, pkgInfo *config.PackageInfo, file *config.File) (string, error)
}

type Packages = map[string]Package

func New(githubArchive *githubarchive.Installer, githubContent *githubcontent.Installer, githubRelease *githubrelease.Installer, goBuild *gobuild.Installer, goInstall *goinstall.Installer, httpInstaller *http.Installer) Packages {
	return map[string]Package{
		githubarchive.PkgType: githubArchive,
		githubcontent.PkgType: githubContent,
		githubrelease.PkgType: githubRelease,
		gobuild.PkgType:       goBuild,
		goinstall.PkgType:     goInstall,
		http.PkgType:          httpInstaller,
	}
}
