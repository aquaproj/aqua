package domain

import (
	"context"

	"github.com/aquaproj/aqua/pkg/checksum"
	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/config/aqua"
	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/sirupsen/logrus"
)

type PackageInstaller interface {
	InstallPackage(ctx context.Context, logE *logrus.Entry, pkg *config.Package, checksums *checksum.Checksums) error
	InstallPackages(ctx context.Context, logE *logrus.Entry, cfg *aqua.Config, registries map[string]*registry.Config) error
	InstallProxy(ctx context.Context, logE *logrus.Entry) error
}
