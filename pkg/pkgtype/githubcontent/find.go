package githubcontent

import (
	"path/filepath"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/sirupsen/logrus"
)

func (inst *Installer) Find(pkg *config.Package, exeName string, logE *logrus.Entry) (string, error) {
	binDir := inst.getInstallDir(pkg)
	for _, file := range inst.GetFiles(pkg.PackageInfo) {
		if file.Name != exeName {
			continue
		}
		return filepath.Join(binDir, exeName), nil
	}
	return "", nil
}
