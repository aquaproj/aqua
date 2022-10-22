package domain

import (
	"context"

	"github.com/aquaproj/aqua/pkg/checksum"
	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/config/aqua"
	"github.com/aquaproj/aqua/pkg/config/registry"
	"github.com/aquaproj/aqua/pkg/config/security"
	"github.com/sirupsen/logrus"
)

type PackageInstaller interface {
	InstallPackage(ctx context.Context, logE *logrus.Entry, param *ParamInstallPackage) error
	InstallPackages(ctx context.Context, logE *logrus.Entry, param *ParamInstallPackages) error
	InstallProxy(ctx context.Context, logE *logrus.Entry) error
}

type ParamInstallPackages struct {
	ConfigFilePath string
	Config         *aqua.Config
	Registries     map[string]*registry.Config
	SkipLink       bool
	SecurityConfig *security.Config
}

type ParamInstallPackage struct {
	Pkg             *config.Package
	Checksums       *checksum.Checksums
	RequireChecksum bool
	SecurityConfig  *security.Config
}
