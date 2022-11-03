package cp

import (
	"fmt"
	"path/filepath"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/domain"
	"github.com/sirupsen/logrus"
)

func (ctrl *Controller) copy(logE *logrus.Entry, param *config.Param, findResult *domain.FindResult, exeName string) error {
	p := filepath.Join(param.Dest, exeName)
	if ctrl.runtime.GOOS == "windows" && filepath.Ext(exeName) == "" {
		p += ".exe"
	}
	logE.WithFields(logrus.Fields{
		"exe_name": exeName,
		"dest":     p,
	}).Info("coping a file")
	if err := ctrl.packageInstaller.Copy(p, findResult.ExePath); err != nil {
		return fmt.Errorf("copy a file: %w", err)
	}
	return nil
}
