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
	logger.With(
		"exe_name", exeName,
		"dest", p,
	).Info("coping a file")
	if err := c.packageInstaller.Copy(p, exePath); err != nil {
		return fmt.Errorf("copy a file: %w", err)
	}
	return nil
}
