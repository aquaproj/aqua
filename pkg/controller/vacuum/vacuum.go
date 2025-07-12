package vacuum

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/vacuum"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

func (c *Controller) Vacuum(logE *logrus.Entry, param *config.Param) error {
	timestamps, err := c.vacuum.FindAll(logE)
	if err != nil {
		return fmt.Errorf("find timestamp files: %w", err)
	}

	timestampChecker := vacuum.NewTimestampChecker(time.Now(), param.VacuumDays)

	for pkgPath, timestamp := range timestamps {
		logE := logE.WithField("package_path", pkgPath)

		if !timestampChecker.Expired(timestamp) {
			continue
		}
		// remove the package
		p := filepath.Join(c.rootDir, pkgPath)
		err := c.fs.RemoveAll(p)
		if err != nil {
			return fmt.Errorf("remove a package: %w", logerr.WithFields(err, logrus.Fields{
				"package_path": p,
			}))
		}
		// remove the timestamp file
		err := c.vacuum.Remove(pkgPath)
		if err != nil {
			return fmt.Errorf("remove a timestamp file: %w", err)
		}

		logE.Info("removed the package")
	}

	return nil
}
