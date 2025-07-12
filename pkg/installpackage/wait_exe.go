package installpackage

import (
	"context"
	"fmt"
	"time"

	"github.com/aquaproj/aqua/v2/pkg/osfile"
	"github.com/aquaproj/aqua/v2/pkg/timer"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

func (is *Installer) WaitExe(ctx context.Context, logE *logrus.Entry, exePath string) error {
	for i := range 10 {
		logE.Debug("check if exec file exists")

		if fi, err := is.fs.Stat(exePath); err == nil {
			if osfile.IsOwnerExecutable(fi.Mode()) {
				break
			}
		}

		logE.WithFields(logrus.Fields{
			"retry_count": i + 1,
		}).Debug("command isn't found. wait for lazy install")
		err := timer.Wait(ctx, 10*time.Millisecond)
		if err != nil { //nolint:mnd
			return fmt.Errorf("wait: %w", logerr.WithFields(err, logE.Data))
		}
	}

	return nil
}
