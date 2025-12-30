package vacuum

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/aquaproj/aqua/v2/pkg/config"
	"github.com/aquaproj/aqua/v2/pkg/vacuum"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

func (c *Controller) Vacuum(logger *slog.Logger, param *config.Param) error {
	timestamps, err := c.vacuum.FindAll(logger)
	if err != nil {
		return fmt.Errorf("find timestamp files: %w", err)
	}
	timestampChecker := vacuum.NewTimestampChecker(time.Now(), param.VacuumDays)
	for pkgPath, timestamp := range timestamps {
		logger := logger.With("package_path", pkgPath)
		if !timestampChecker.Expired(timestamp) {
			continue
		}
		// remove the package
		p := filepath.Join(c.rootDir, pkgPath)
		if err := c.fs.RemoveAll(p); err != nil {
			return fmt.Errorf("remove a package: %w", slogerr.With(err,
				"package_path", p,
			))
		}
		// remove the timestamp file
		if err := c.vacuum.Remove(pkgPath); err != nil {
			return fmt.Errorf("remove a timestamp file: %w", err)
		}
		logger.Info("removed the package")
	}
	return nil
}
