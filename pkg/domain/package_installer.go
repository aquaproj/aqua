package domain

import (
	"context"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/config/aqua"
	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/sirupsen/logrus"
)

type PackageInstaller interface {
	InstallPackage(ctx context.Context, pkg *config.Package, logE *logrus.Entry) error
	InstallPackages(ctx context.Context, cfg *aqua.Config, registries map[string]*registry.Config, logE *logrus.Entry) error
	InstallProxy(ctx context.Context, logE *logrus.Entry) error
}
