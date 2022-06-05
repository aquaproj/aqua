package pkgtype

import (
	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/pkgtype/githubarchive"
	"github.com/aquaproj/aqua/pkg/pkgtype/githubcontent"
	"github.com/aquaproj/aqua/pkg/pkgtype/githubrelease"
	"github.com/aquaproj/aqua/pkg/pkgtype/gobuild"
	"github.com/aquaproj/aqua/pkg/pkgtype/goinstall"
	"github.com/aquaproj/aqua/pkg/pkgtype/http"
)

func New(githubArchive *githubarchive.Installer, githubContent *githubcontent.Installer, githubRelease *githubrelease.Installer, goBuild *gobuild.Installer, goInstall *goinstall.Installer, httpInstaller *http.Installer) config.PackageTypes {
	return config.PackageTypes{
		githubarchive.PkgType: githubArchive,
		githubcontent.PkgType: githubContent,
		githubrelease.PkgType: githubRelease,
		gobuild.PkgType:       goBuild,
		goinstall.PkgType:     goInstall,
		http.PkgType:          httpInstaller,
	}
}
