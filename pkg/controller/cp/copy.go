package cp

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/aquaproj/aqua/pkg/config"
	"github.com/aquaproj/aqua/pkg/controller/which"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

func (ctrl *Controller) copy(logE *logrus.Entry, param *config.Param, findResult *which.FindResult, exeName string) error {
	p := filepath.Join(param.Dest, exeName)
	if ctrl.runtime.GOOS == "windows" && filepath.Ext(exeName) == "" {
		p += ".exe"
	}
	logE.WithFields(logrus.Fields{
		"exe_name": exeName,
		"dest":     p,
	}).Info("coping a file")
	dest, err := ctrl.fs.OpenFile(p, os.O_RDWR|os.O_CREATE|os.O_TRUNC, filePermission) //nolint:nosnakecase
	if err != nil {
		return fmt.Errorf("create a file: %w", logerr.WithFields(err, logE.Data))
	}
	defer dest.Close()
	src, err := ctrl.fs.Open(findResult.ExePath)
	if err != nil {
		return fmt.Errorf("open a file: %w", logerr.WithFields(err, logE.Data))
	}
	defer src.Close()
	if _, err := io.Copy(dest, src); err != nil {
		return fmt.Errorf("copy a file: %w", logerr.WithFields(err, logE.Data))
	}
	return nil
}
