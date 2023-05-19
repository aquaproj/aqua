package generate

import (
	"context"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/sirupsen/logrus"
)

func (ctrl *Controller) getCargoVersion(ctx context.Context, logE *logrus.Entry, param *config.Param, pkg *FindingPackage) string {
	pkgInfo := pkg.PackageInfo
	if param.SelectVersion {
		versions, err := ctrl.cargoClient.ListVersions(ctx, *pkgInfo.Crate)
		if err != nil {
			logE.WithError(err).Warn("list versions")
			return ""
		}
		idx, err := ctrl.crateVersionSelector.Find(versions)
		if err != nil {
			return ""
		}
		return versions[idx]
	}
	version, err := ctrl.cargoClient.GetLatestVersion(ctx, *pkgInfo.Crate)
	if err != nil {
		logE.WithError(err).Warn("get a latest version")
		return ""
	}
	return version
}

type CrateVersionSelector interface {
	Find(versions []string) (int, error)
}

type MockCrateVersionSelector struct {
	Index int
	Err   error
}

func (mock *MockCrateVersionSelector) Find(versions []string) (int, error) {
	return mock.Index, mock.Err
}

type CrateVersionSelectorImpl struct{}

func NewCrateVersionSelectorImpl() *CrateVersionSelectorImpl {
	return &CrateVersionSelectorImpl{}
}

func (selector *CrateVersionSelectorImpl) Find(versions []string) (int, error) {
	return fuzzyfinder.Find(versions, func(i int) string { //nolint:wrapcheck
		return versions[i]
	})
}
