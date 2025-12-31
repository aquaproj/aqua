package cp

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/aquaproj/aqua/v2/pkg/config"
)

func (c *Controller) copy(logger *slog.Logger, param *config.Param, exePath string, exeName string) error {
	p := filepath.Join(param.Dest, exeName)
	if c.runtime.GOOS == "windows" && filepath.Ext(exeName) == "" {
		p += ".exe"
	}
	logger.Info("coping a file",
		"exe_name", exeName,
		"dest", p)
	if err := c.packageInstaller.Copy(p, exePath); err != nil {
		return fmt.Errorf("copy a file: %w", err)
	}
	return nil
}
