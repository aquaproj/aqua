package githubrelease

import (
	"fmt"
	"path/filepath"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/sirupsen/logrus"
)

func (inst *Installer) Find(pkg *config.Package, pkgInfo *config.PackageInfo, exeName string, logE *logrus.Entry) (string, error) {
	assetName, err := inst.assetName(pkgInfo, pkg)
	if err != nil {
		return "", err
	}
	binDir := inst.getInstallDir(pkg, pkgInfo, assetName)
	for _, file := range pkgInfo.Files {
		if file.Name != exeName {
			continue
		}
		fileSrc, err := file.GetSrc(pkg, pkgInfo, inst.runtime)
		if err != nil {
			return "", fmt.Errorf("get file_src: %w", err)
		}
		return filepath.Join(binDir, fileSrc), nil
	}
	return "", nil
}
