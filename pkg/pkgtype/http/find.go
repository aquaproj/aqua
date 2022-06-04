package http

import (
	"fmt"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/sirupsen/logrus"
)

func (inst *Installer) Find(pkg *config.Package, pkgInfo *config.PackageInfo, exeName string, logE *logrus.Entry) (string, error) {
	for _, file := range pkgInfo.Files {
		if file.Name != exeName {
			continue
		}
		filePath, err := inst.GetFilePath(pkg, pkgInfo, file)
		if err != nil {
			return "", fmt.Errorf("get file_src: %w", err)
		}
		return filePath, nil
	}
	return "", nil
}
