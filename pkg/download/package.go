package download

import (
	"github.com/aquaproj/aqua/pkg/runtime"
)

type PackageDownloader struct {
	github  RepositoriesService
	runtime *runtime.Runtime
	http    HTTPDownloader
}

func NewPackageDownloader(gh RepositoriesService, rt *runtime.Runtime, httpDownloader HTTPDownloader) *PackageDownloader {
	return &PackageDownloader{
		github:  gh,
		runtime: rt,
		http:    httpDownloader,
	}
}
