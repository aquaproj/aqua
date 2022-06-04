package gobuild

import (
	"path/filepath"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/sirupsen/logrus"
)

func (inst *Installer) Find(pkg *config.Package, pkgInfo *config.PackageInfo, exeName string, logE *logrus.Entry) (string, error) {
	binDir := inst.getBinDir(pkg, pkgInfo)
	for _, file := range pkgInfo.Files {
		if file.Name != exeName {
			continue
		}
		return filepath.Join(binDir, exeName), nil
	}
	return "", nil
}
