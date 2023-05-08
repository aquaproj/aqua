package installpackage

import (
	"context"
	"fmt"
	"time"

	"github.com/aquaproj/aqua/v2/pkg/util"
	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

func (inst *InstallerImpl) WaitExe(ctx context.Context, logE *logrus.Entry, exePath string) error {
	for i := 0; i < 10; i++ {
		logE.Debug("check if exec file exists")
		if fi, err := inst.fs.Stat(exePath); err == nil {
			if util.IsOwnerExecutable(fi.Mode()) {
				break
			}
		}
		logE.WithFields(logrus.Fields{
			"retry_count": i + 1,
		}).Debug("command isn't found. wait for lazy install")
		if err := util.Wait(ctx, 10*time.Millisecond); err != nil { //nolint:gomnd
			return fmt.Errorf("wait: %w", logerr.WithFields(err, logE.Data))
		}
	}
	return nil
}
