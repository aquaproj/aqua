package installpackage

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/aquaproj/aqua/v2/pkg/timer"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

func (is *Installer) WaitExe(ctx context.Context, logger *slog.Logger, exePath string) error {
	for i := range 10 {
		logger.Debug("check if exec file exists")
		if fi, err := is.fs.Stat(exePath); err == nil {
			if osfile.IsOwnerExecutable(fi.Mode()) {
				break
			}
		}
		logger.Debug("command isn't found. wait for lazy install",
			slog.Int("retry_count", i+1))
		if err := timer.Wait(ctx, 10*time.Millisecond); err != nil { //nolint:mnd
			return fmt.Errorf("wait: %w", slogerr.With(err))
		}
	}
	return nil
}
